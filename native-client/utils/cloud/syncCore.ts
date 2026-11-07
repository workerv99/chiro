/**
 * Pure logic of the cloud sync engine (no React Native, no DB, no network), so it can
 * be exercised from a plain Node script. The engine in syncEngine.ts wires it to the
 * local SQLite layer and to POST /api/sync.
 */

export type Row = Record<string, unknown>;
export type TableRows = Record<string, Row[]>;

export interface SyncRequestBody { tables: TableRows; since: number }
export interface SyncResponse {
  pulled?: TableRows;
  serverNow?: number;
  next_since?: number;
  has_more?: boolean;
}

/** Server body cap is 1 MiB; stay well under it. */
export const MAX_CHUNK_ROWS = 200;
export const MAX_CHUNK_BYTES = 700_000;
/** Fixed bytes of the JSON envelope around one chunk (keys, braces, since). */
const ENVELOPE_BYTES = 128;

export const DEFAULT_PULL_LIMIT = 500;
export const MAX_PULL_LIMIT = 5000;
/**
 * Pull `since` is client-clock based, so rows another device pushed late (with an
 * older updated_at) could fall behind the cursor. Re-pulling a window is cheap
 * because already-known rows are skipped by the LWW check.
 */
export const PULL_OVERLAP_MS = 60 * 60 * 1000;

// ── Columns ───────────────────────────────────────────────────────────────────

/** Column names per table, from the CREATE TABLE statements plus ALTER ... ADD COLUMN. */
export function schemaColumns(createTables: string[], migrations: string[]): Record<string, string[]> {
  const out: Record<string, string[]> = {};
  for (const sql of createTables) {
    const m = /CREATE TABLE IF NOT EXISTS\s+(\w+)\s*\(([\s\S]*)\)\s*;?\s*$/i.exec(sql.trim());
    if (!m) continue;
    const cols: string[] = [];
    for (const raw of m[2].split('\n')) {
      const line = raw.trim();
      if (!line || /^(PRIMARY|CONSTRAINT|UNIQUE|CHECK|FOREIGN)\b/i.test(line)) continue;
      const id = /^(\w+)\s/.exec(line);
      if (id) cols.push(id[1]);
    }
    out[m[1]] = cols;
  }
  for (const sql of migrations) {
    const m = /^ALTER TABLE\s+(\w+)\s+ADD COLUMN\s+(\w+)/i.exec(sql.trim());
    if (m && out[m[1]] && !out[m[1]].includes(m[2])) out[m[1]].push(m[2]);
  }
  return out;
}

/** Keeps only known columns (the server answers 400 on unknown ones) and drops `undefined`. */
export function pickColumns(row: Row, allowed: readonly string[]): Row {
  const out: Row = {};
  for (const c of allowed) {
    const v = row[c];
    if (v !== undefined) out[c] = v;
  }
  return out;
}

const DATE_COLUMNS = new Set(['date', 'due_date', 'paid_date', 'next_date', 'first_due_date']);

/**
 * Server rows (Postgres) to the app's SQLite shapes: DATE columns come back as
 * ISO timestamps ("2026-01-31T00:00:00Z") but the app stores "YYYY-MM-DD"; booleans become 0/1.
 */
export function normalizePulledRow(row: Row, allowed: readonly string[]): Row {
  const out = pickColumns(row, allowed);
  for (const [k, v] of Object.entries(out)) {
    if (typeof v === 'boolean') out[k] = v ? 1 : 0;
    else if (typeof v === 'string' && DATE_COLUMNS.has(k) && /^\d{4}-\d{2}-\d{2}T/.test(v)) out[k] = v.slice(0, 10);
  }
  return out;
}

/** Last-writer-wins: a pulled row replaces the local one only if strictly newer. */
export function shouldApply(localUpdatedAt: number | null | undefined, remoteUpdatedAt: unknown): boolean {
  const remote = Number(remoteUpdatedAt);
  if (!Number.isFinite(remote)) return false;
  if (localUpdatedAt == null) return true;
  return remote > Number(localUpdatedAt);
}

// ── Chunking ──────────────────────────────────────────────────────────────────

/** UTF-8 byte length without TextEncoder/Buffer (not guaranteed on every runtime). */
export function utf8Length(s: string): number {
  let n = 0;
  for (let i = 0; i < s.length; i++) {
    const c = s.charCodeAt(i);
    if (c < 0x80) n += 1;
    else if (c < 0x800) n += 2;
    else if (c >= 0xd800 && c <= 0xdbff) { n += 4; i++; } // surrogate pair
    else n += 3;
  }
  return n;
}

export interface ChunkLimits { maxRows: number; maxBytes: number }

/**
 * Splits rows into request bodies of one table each, following `order` (parents
 * before children, as in SYNC_TABLES; the server merges tables of one request in
 * undefined order, which could trip foreign keys). Each chunk has at most
 * `maxRows` rows and about `maxBytes` serialized; a single oversized row still
 * travels alone.
 */
export function chunkPushRows(
  tables: TableRows,
  order: readonly string[],
  limits: ChunkLimits = { maxRows: MAX_CHUNK_ROWS, maxBytes: MAX_CHUNK_BYTES },
): TableRows[] {
  const chunks: TableRows[] = [];
  for (const table of order) {
    const rows = tables[table];
    if (!rows || rows.length === 0) continue;
    let cur: Row[] = [];
    let bytes = ENVELOPE_BYTES + utf8Length(table);
    for (const row of rows) {
      const size = utf8Length(JSON.stringify(row)) + 1;
      if (cur.length > 0 && (cur.length >= limits.maxRows || bytes + size > limits.maxBytes)) {
        chunks.push({ [table]: cur });
        cur = [];
        bytes = ENVELOPE_BYTES + utf8Length(table);
      }
      cur.push(row);
      bytes += size;
    }
    if (cur.length > 0) chunks.push({ [table]: cur });
  }
  return chunks;
}

// ── Watermarks ────────────────────────────────────────────────────────────────

export interface UserSyncState {
  /** Highest local updated_at already pushed. Rows with updated_at <= this are skipped. */
  lastPushedAt: number;
  /** Pull cursor (next_since of the last complete pull). */
  pullSince: number;
  lastSyncAt: number | null;
}

export const EMPTY_STATE: UserSyncState = { lastPushedAt: 0, pullSince: 0, lastSyncAt: null };

/**
 * Picks the watermarks to use for `userId`. When a different user synced last,
 * the local DB may hold that user's data, so we start over (full push, full pull).
 */
export function resolveState(
  userId: string,
  lastSyncedUserId: string | null,
  stored: UserSyncState | null,
): { state: UserSyncState; reset: boolean } {
  if (lastSyncedUserId !== userId || !stored) return { state: { ...EMPTY_STATE }, reset: lastSyncedUserId !== userId };
  return { state: stored, reset: false };
}

/** Cursor used for the first request: step back by the overlap window, never below 0. */
export function pullCursor(pullSince: number, overlapMs: number = PULL_OVERLAP_MS): number {
  return pullSince <= 0 ? 0 : Math.max(0, pullSince - overlapMs);
}

// ── Sync cycle ────────────────────────────────────────────────────────────────

export class SyncStalledError extends Error {
  constructor() { super('SYNC_STALLED'); this.name = 'SyncStalledError'; }
}

export interface CycleOptions {
  /** Push bodies, one request each (see chunkPushRows). Empty means pull only. */
  chunks: TableRows[];
  since: number;
  limit?: number;
  request: (body: SyncRequestBody, limit: number) => Promise<SyncResponse>;
  applyPulled: (pulled: TableRows) => Promise<void>;
  maxRequests?: number;
}

export interface CycleResult { cursor: number; requests: number; pushedRows: number; limit: number }

/**
 * Runs the request loop: each chunk is pushed with the current cursor, pulled pages are
 * applied as they arrive, and the loop continues while `has_more` using `next_since`.
 * If the server says there is more but the cursor cannot advance (a table with >= limit
 * rows sharing one timestamp), the pull limit is doubled up to the server maximum.
 */
export async function runCycle(opts: CycleOptions): Promise<CycleResult> {
  const queue = [...opts.chunks];
  let cursor = opts.since;
  let limit = opts.limit ?? DEFAULT_PULL_LIMIT;
  const maxRequests = opts.maxRequests ?? 1000;
  let requests = 0;
  let pushedRows = 0;

  for (;;) {
    if (requests >= maxRequests) throw new SyncStalledError();
    const tables = queue.shift() ?? {};
    const res = await opts.request({ tables, since: cursor }, limit);
    requests++;
    for (const rows of Object.values(tables)) pushedRows += rows.length;

    if (res.pulled && Object.values(res.pulled).some((r) => r.length > 0)) await opts.applyPulled(res.pulled);

    const next = typeof res.next_since === 'number' ? res.next_since : cursor;
    if (res.has_more) {
      if (next <= cursor) {
        if (limit >= MAX_PULL_LIMIT) throw new SyncStalledError();
        limit = Math.min(limit * 2, MAX_PULL_LIMIT);
      }
      cursor = Math.max(cursor, next);
      continue;
    }
    cursor = Math.max(cursor, next);
    if (queue.length === 0) break;
  }
  return { cursor, requests, pushedRows, limit };
}
