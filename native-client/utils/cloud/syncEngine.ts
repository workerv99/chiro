/**
 * Cloud sync engine: pushes local changes to POST /api/sync and applies pulled rows with
 * last-writer-wins. Active only while `configureCloudSync` says it is enabled (Pro active
 * and a session). The pure parts live in syncCore.ts.
 *
 * Concurrency: one run at a time (single-flight, with one re-run if triggered meanwhile)
 * and every run goes through the same lock as the Google Drive sync (runSyncExclusive),
 * so the two never touch the DB concurrently.
 */
import AsyncStorage from '@react-native-async-storage/async-storage';
import { db } from '../db';
import { buildUpsertStatement, CREATE_TABLES, MIGRATIONS, SYNC_TABLES } from '../dbSchema';
import { runSyncExclusive } from '../sync';
import { CloudApiError } from './errors';
import { cloudApi } from './index';
import {
  chunkPushRows, normalizePulledRow, pickColumns, pullCursor, resolveState, runCycle, schemaColumns,
  shouldApply, SyncStalledError,
  type Row, type SyncResponse, type TableRows, type UserSyncState,
} from './syncCore';

export type CloudSyncStatus = 'disabled' | 'idle' | 'syncing' | 'error' | 'paywalled';

export interface CloudSyncState {
  status: CloudSyncStatus;
  lastSyncAt: number | null;
  /** Message of the last failure (status 'error'), otherwise null. */
  error: string | null;
}

const LAST_USER_KEY = 'cloud_sync_last_user';
const stateKey = (userId: string) => `cloud_sync_state:${userId}`;
const WRITE_DEBOUNCE_MS = 4000;
/** Columns of every synced table, from the app schema (same names as the server). */
const COLUMNS = schemaColumns(CREATE_TABLES, MIGRATIONS);

// ── Observable state ──────────────────────────────────────────────────────────

let state: CloudSyncState = { status: 'disabled', lastSyncAt: null, error: null };
const listeners = new Set<() => void>();

function setState(patch: Partial<CloudSyncState>) {
  state = { ...state, ...patch };
  listeners.forEach((l) => l());
}

export const getCloudSyncState = (): CloudSyncState => state;
export function subscribeCloudSync(fn: () => void): () => void {
  listeners.add(fn);
  return () => { listeners.delete(fn); };
}

// ── Configuration ─────────────────────────────────────────────────────────────

let enabled = false;
let userId: string | null = null;
let paywalled = false;
let onRemoteApplied: (() => void) | undefined;

/**
 * Called by the app whenever the account/subscription changes. Turning sync on (or
 * switching user) runs a sync right away; turning it off cancels pending work.
 */
export function configureCloudSync(cfg: { enabled: boolean; userId: string | null; onRemoteApplied?: () => void }): void {
  onRemoteApplied = cfg.onRemoteApplied;
  const nextEnabled = cfg.enabled && !!cfg.userId;
  const changed = nextEnabled !== enabled || (nextEnabled && cfg.userId !== userId);
  enabled = nextEnabled;
  userId = nextEnabled ? cfg.userId : null;
  if (!changed) return;

  if (writeTimer) { clearTimeout(writeTimer); writeTimer = null; }
  if (!enabled) {
    paywalled = false;
    setState({ status: 'disabled', error: null });
    return;
  }
  paywalled = false;
  setState({ status: 'idle', error: null });
  void loadState(userId!).then((s) => { if (userId) setState({ lastSyncAt: s.lastSyncAt }); }).catch(() => {});
  void cloudSyncNow();
}

// ── Triggers ──────────────────────────────────────────────────────────────────

let writeTimer: ReturnType<typeof setTimeout> | null = null;
let applyDepth = 0;

/** Debounced sync after a local write. Writes made by the pull-apply step itself are ignored. */
export function notifyCloudLocalWrite(): void {
  if (!enabled || paywalled || applyDepth > 0) return;
  if (writeTimer) clearTimeout(writeTimer);
  writeTimer = setTimeout(() => { writeTimer = null; void cloudSyncNow({ manual: false }); }, WRITE_DEBOUNCE_MS);
}

let running: Promise<void> | null = null;
let rerun = false;

/**
 * Syncs now. Resolves when finished and never rejects: the outcome is in the state
 * (idle / error / paywalled). Calls during a run share it and trigger one more pass.
 * Automatic triggers (`manual: false`) are skipped while paywalled (402), manual ones retry.
 */
export function cloudSyncNow(opts: { manual?: boolean } = {}): Promise<void> {
  if (!enabled || !userId) return Promise.resolve();
  const manual = opts.manual ?? true;
  if (paywalled && !manual) return Promise.resolve();
  if (manual) paywalled = false;
  if (running) { rerun = true; return running; }
  running = (async () => {
    do {
      rerun = false;
      await runOnce();
    } while (rerun && enabled && !paywalled && state.status !== 'error');
  })().finally(() => { running = null; });
  return running;
}

async function runOnce(): Promise<void> {
  const uid = userId;
  if (!enabled || !uid) return;
  setState({ status: 'syncing', error: null });
  try {
    const finished = await runSyncExclusive(() => syncCycle(uid));
    if (enabled && userId === uid) setState({ status: 'idle', lastSyncAt: finished.lastSyncAt, error: null });
  } catch (e) {
    if (e instanceof Error && e.message === 'SYNC_ABORTED') return;
    if (e instanceof CloudApiError && e.isPaymentRequired) {
      paywalled = true;
      setState({ status: 'paywalled', error: null });
    } else {
      const message = e instanceof SyncStalledError ? 'SYNC_STALLED' : e instanceof Error ? e.message : String(e);
      console.warn('[cloud-sync] failed:', message);
      setState({ status: 'error', error: message });
    }
  }
}

// ── One cycle ─────────────────────────────────────────────────────────────────

async function loadState(uid: string): Promise<UserSyncState> {
  const [lastUser, raw] = await Promise.all([AsyncStorage.getItem(LAST_USER_KEY), AsyncStorage.getItem(stateKey(uid))]);
  let stored: UserSyncState | null = null;
  try { stored = raw ? (JSON.parse(raw) as UserSyncState) : null; } catch { stored = null; }
  return resolveState(uid, lastUser, stored).state;
}

async function readPushRows(lastPushedAt: number): Promise<{ tables: TableRows; maxUpdatedAt: number }> {
  const tables: TableRows = {};
  let maxUpdatedAt = lastPushedAt;
  for (const { table } of SYNC_TABLES) {
    const rows = await db.getAllAsync<Row>(`SELECT * FROM ${table} WHERE updated_at > ? ORDER BY updated_at ASC`, lastPushedAt);
    if (rows.length === 0) continue;
    tables[table] = rows.map((r) => {
      const u = Number(r.updated_at);
      if (u > maxUpdatedAt) maxUpdatedAt = u;
      return pickColumns(r, COLUMNS[table]);
    });
  }
  return { tables, maxUpdatedAt };
}

/** Applies pulled rows in one transaction; returns how many actually changed local data. */
async function applyPulled(pulled: TableRows): Promise<number> {
  applyDepth++;
  try {
    let changed = 0;
    await db.withTransactionAsync(async (tx) => {
      for (const { table, pk } of SYNC_TABLES) {
        for (const raw of pulled[table] ?? []) {
          const row = normalizePulledRow(raw, COLUMNS[table]);
          if (pk.some((k) => row[k] == null)) continue;
          const local = await tx.getFirstAsync<{ updated_at: number }>(
            `SELECT updated_at FROM ${table} WHERE ${pk.map((k) => `${k}=?`).join(' AND ')}`,
            ...pk.map((k) => row[k]),
          );
          if (!shouldApply(local?.updated_at, row.updated_at)) continue;
          const cols = COLUMNS[table].filter((c) => c in row);
          await tx.runAsync(buildUpsertStatement(table, pk, cols), ...cols.map((c) => row[c]));
          changed++;
        }
      }
    });
    return changed;
  } finally {
    // The write interceptor reacts after the transaction promise settles; keep ignoring until then.
    setTimeout(() => { applyDepth--; }, 0);
  }
}

async function syncCycle(uid: string): Promise<UserSyncState> {
  const lastUser = await AsyncStorage.getItem(LAST_USER_KEY);
  const state0 = await loadState(uid);
  if (lastUser !== uid) {
    // Different account than the one that synced last: watermarks were reset by loadState.
    await AsyncStorage.setItem(stateKey(uid), JSON.stringify(state0));
    await AsyncStorage.setItem(LAST_USER_KEY, uid);
  }

  const push = await readPushRows(state0.lastPushedAt);
  const chunks = chunkPushRows(push.tables, SYNC_TABLES.map((t) => t.table));
  let appliedRows = 0;

  const result = await runCycle({
    chunks,
    since: pullCursor(state0.pullSince),
    request: (body, limit) => {
      if (!enabled || userId !== uid) throw new Error('SYNC_ABORTED');
      return cloudApi.request<SyncResponse>(`/api/sync?limit=${limit}`, { method: 'POST', body });
    },
    applyPulled: async (pulled) => { appliedRows += await applyPulled(pulled); },
  });

  const next: UserSyncState = {
    lastPushedAt: push.maxUpdatedAt,
    // The overlap window may report a lower cursor than the stored one: never go backwards.
    pullSince: Math.max(state0.pullSince, result.cursor),
    lastSyncAt: Date.now(),
  };
  await AsyncStorage.setItem(stateKey(uid), JSON.stringify(next));
  if (appliedRows > 0) onRemoteApplied?.();
  return next;
}
