# Loan review follow-ups

## Objective
Address the non-blocking advisory findings from the native review of the loan payment-history work (lineage `review-1fd23e5d15d3cb3b`, approved and acknowledged on 2026-10-03). They are separate later work and do not reopen that review.

## Source and limits
Only each finding's id, location and severity were available when this file was written; the reviewers' full reasoning was not read. Each task must start by re-reading the cited code and confirming the finding is real before changing anything.

## Scope
Backend `backend/pkg/svc/loans.go`, `backend/pkg/app/loan_handlers.go`; frontend `web/src/routes/loans/[id]/+page.svelte`; `docker-compose.yml`. Implementation is NOT authorized yet: this document only tracks the work.

## Tasks
Warnings first.
- [ ] T1 (WARNING, R4-cascade-no-lock, `loans.go:117`): the cascade payment reads and updates installments without row locks, so two simultaneous payments on one loan can overwrite each other. Lock the loan's installments (e.g. `SELECT ... FOR UPDATE`) inside the transaction; cover with a test that exercises the pure allocation plus a concurrency check where feasible.
- [ ] T2 (WARNING, R3-pay-history-nonidempotent, `loans.go:70-90`): `PayInstallment` writes payment history non-idempotently; a retry can duplicate the row. Make the write idempotent or skip history when nothing changed.
- [ ] T3 (WARNING, R3-unpay-no-history-fallback, `loans.go:306-312`): for loans without history, unpay falls back to reverting a single installment instead of the whole last payment. Decide the right behavior for legacy data (the 003 backfill may already cover it) and test it.
- [ ] T4 (WARNING, R2-unpay-dup, `loans.go:204-260`): duplicated logic in the unpay functions; extract the shared part.
- [ ] T5 (WARNING, R2-unpay-hack, `+page.svelte:204-208`): `unpay()` sends `schedule[0].installment_id` just to identify the loan. Give the endpoint a loan-based route or document the contract, and drop the workaround.
- [ ] T6 (SUGGESTION, R3-handler-untested, `loan_handlers.go:476-555`): the installments handler that builds `remaining`, `is_partial` and `original_amount` has no test. Extract the computation into a pure function and test it.
- [ ] T7 (SUGGESTION, R2-redundant-counts, `loans.go:160-162`): redundant counts; simplify.
- [ ] T8 (SUGGESTION, R2-pdf-layout, `+page.svelte:225-235`): PDF header layout constants are hard to follow; name them.
- [ ] T9 (SUGGESTION, R2-port-churn, `docker-compose.yml:11`): the dev-port change touches many files; consider centralizing the ports in one env file.

## Acceptance criteria
- Each task re-confirms its finding against the current code before changing it; findings that no longer apply are closed with a note.
- Behavior changes follow strict TDD (RED then GREEN).

## Checks / TDD
- Strict TDD enabled, runner `go test ./...` from `backend/` (focused: `go test ./pkg/svc/`).
- Frontend: `npm run build` from `web/`.
- Functional: Playwright against the local stack with the credentials in memory `chiro-local-dev-credentials`, on loan `loan_a10a79f3-fb0c-484e-9967-ddb49f5515d5`; restore data afterwards.

## Routing
- To decide when implementation is authorized: T1-T3 likely one delegated writer (same file, shared transaction code); T4-T9 are small and can be done inline or in a second pass.

## Progress
- Created; nothing implemented.
