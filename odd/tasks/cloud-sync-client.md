# cloud-sync-client

## Objective
The Expo app (this repo) stays offline-first with its local SQLite. A user can activate a paid subscription; only then does the app log in to the Chiro backend (`/home/kv/Projects/programming/chiro`, branch `feat/sync-subscription-gate`) and sync via `POST /api/sync`.

## Decisions (user)
- Subscription unlocks cloud sync only (no other premium features).
- Identity: email + password against the backend (no Google login on the backend).

## Scope
- Backend (repo `chiro`): gate `/api/sync` on an active Pro plan, never trust client `user_id`, fix `next_since` paging. Payment verification for `activate` is OUT of scope (stub stays; follow-up).
- App: API client (base URL env, JWT store, refresh), register/login UI, subscription status/activate, sync engine (since watermark, chunked push, paged pull), settings entry. Existing Google Drive sync stays untouched.

## Tasks
- [x] T1 Backend: sync requires active Pro (402/403 otherwise), force `user_id` from token, correct paging cursor. Tests. Route: delegated writer. Commit `995510c` (chiro, feat/sync-subscription-gate). Evidence: RED (undefined IsProActive/TablePage/NextSince, Normalize user_id test) then GREEN; `go test ./...` 109 passed; vet/build clean. Contract for the client: 402 = no active Pro; response `{pulled, serverNow, next_since, has_more}`; pull is `updated_at >= since`, loop while `has_more`; a table with >= limit rows at the same timestamp cannot advance (raise `?limit`, max 5000).
- [x] T2 App: API client + token storage + env `EXPO_PUBLIC_API_URL`. Route: delegated writer. Commit `45be6c0`. No unit-test runner in repo (no jest/vitest): verified with scratch node smoke script (single-flight refresh, 401 drop, timeout, isProActive), `pnpm typecheck` clean.
- [x] T3 App: login/register screen + subscription status/activate in settings. Route: delegated writer. Commit recorded in git log (`feat(cloud): add cloud account provider, screen and settings section`). Cloud strings live in `components/cloud/strings.ts` (es+en) because `utils/i18n.tsx` was outside the edit surface. GET /api/subscription has no `expires_at` yet (expiry shown only when present).
- [x] T4 App: sync engine against `/api/sync` (watermark, chunking, bump `updated_at` on writes, tombstones). Route: delegated writer. Pure logic in `utils/cloud/syncCore.ts` (RED with stubs: 0/17 pass; GREEN: 17/17 via scratch node script outside the repo), engine `utils/cloud/syncEngine.ts`, controller/hook/card in `components/cloud/`. `pnpm typecheck` clean. Not run on device/Electron/UI. Decisions: shared lock with Drive sync = `runSyncExclusive` (same queue as utils/sync.ts `serialized`); pull-apply uses new parameterized `buildUpsertStatement` (buildMergeStatement reads an attached `remote` DB, unusable for JSON rows); one table per push request (server merges tables of a request in map order, FKs); pull cursor re-reads a 1h overlap (client-clock cursor can miss rows another device pushes late with an old updated_at; real fix = server-side sequence). Audit: every write in utils/repo/* sets `updated_at` and deletes are soft (`deleted=1`), no hard DELETE exists; nothing to fix. Seed rows keep updated_at=0 (server wins).

## Open / follow-up
- Real payment verification on `POST /api/subscription/activate` (currently self-grants 30 days).
- Seed rows (`acc_cash`, `cat_food`) have `updated_at=0` in the app, so the server copy wins on first sync.

## Progress
T1-T4 done. Follow-ups: pulled rows are re-pushed once (echo, harmless under server LWW); switching accounts on one device pushes the previous account's local data to the new account (no wipe/warning yet); Drive-merge writes do not trigger a cloud sync until the next write/foreground.
