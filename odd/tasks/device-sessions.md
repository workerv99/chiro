# Feature: device-sessions

## Objective
Stop users being logged out every 24h and give them per-device session control (list + revoke).

## Problem
- JWT TTL is 24h. `/api/auth/refresh` exists (accepts tokens expired <7d) but the web client never calls it; `api()` clears the token on any 401.
- Tokens carry no device identity: no way to list or revoke sessions. `logout()` is client-only.

## Design (accepted by user: "las dos")
- Migration `007_sessions.sql`: `sessions(session_id TEXT PK, user_id FK ON DELETE CASCADE, user_agent, ip_address, created_at, last_used_at, revoked_at TIMESTAMPTZ)`.
- JWT gets `sid` claim. Login/register create a session; refresh keeps the same sid, bumps `last_used_at`, rejects revoked sessions.
- Legacy tokens without `sid` stay valid until they expire (max 24h); refreshing one creates a session.
- Middleware path (`requireActive`) rejects tokens whose session is revoked/missing (when `sid` present).
- Endpoints (protected): `GET /api/auth/sessions` (marks current), `DELETE /api/auth/sessions/{id}`, `POST /api/auth/logout` (revokes current).
- Web: `api()` on 401 does single-flight refresh + one retry; logout calls server; Config page gets a "Devices" card (lucide icons, mobile-first, 44px targets).

## Constraints
- Artifacts in English (code, comments, UI copy keys); i18n has es + en entries.
- Conventional commits, NO Co-Authored-By / AI attribution (user rule).
- ~400 authored changed lines per task is a heuristic only.
- TDD: strict mode enabled. Go runner: `cd backend && go test ./...`. Web has no test runner (only `npm run check` / `lint`); web tasks verified with svelte-check + lint + browser run, disclosed.

## Tasks
- [x] T1 Migration 007 + auth.Manager `sid` claim (Issue/Parse/ParseClaims return sid) with Go unit tests. Route: delegated writer (backend).
- [x] T2 Session store + handlers (create on login/register, refresh w/ revoke check, list, revoke, logout) + middleware revoke check. Route: delegated writer (backend).
- [x] T3 Web: silent refresh in `api()`, startup refresh, server logout. Route: delegated writer (frontend).
- [ ] T4 Web: Devices card in Config + i18n es/en. Route: delegated writer (frontend).

## Delivery
Strategy: ask-on-risk. Branch: `feat/device-sessions`. Forecast ~450-550 authored lines; slice into 2 PRs if it exceeds ~400 (backend T1-T2, web T3-T4).

## Progress
- Mapping done. T1 (05355f6) and T2 done (backend). T3 done (web: single-flight refresh in api(), startup refresh, best-effort server logout). Next: T4.

## Verification evidence
- T1 RED: `go test ./pkg/auth` failed to compile (Issue arity, Parse 3 returns, no Claims.SessionID). GREEN after impl: `ok chiro/pkg/auth`. `go build ./...` and `go vet ./...` clean. Migration 007 is not exercised by tests (no DB test infra).
- T2 RED: `go test ./pkg/app` failed to compile (truncateUserAgent, shouldTouchSession, sessionIsCurrent undefined). GREEN after impl; `go test ./...`, `go vet ./...`, `go build ./...` clean. Not covered (needs DB): SQL in session create/list/revoke/logout/refresh and the requireActive JOIN query; handlers verified by compilation only.
- T3: `cd web && npm run check` 0 errors/0 warnings. `npm run lint` 27 errors, identical to base (stash comparison; all pre-existing parse errors in ui/*.svelte and unused vars), none in changed files. No browser run for T3 (needs running backend); logic verified by reading only.

## Next step
T4 Devices card.
