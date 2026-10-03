# Loan installment fixes

## Objective
Fix three bugs found in the browser test of the loan detail flow (engram "Loans browser test findings").

## Problem
1. "Deshacer ultimo pago" reverts only the last fully-paid installment. After paying 50 over installments #1 (32.50) and #2 (17.50), unpay leaves 17.50 on #2.
2. Edit-installment dialog (and loan edit) shows the date one day earlier because `new Date('YYYY-MM-DD')` parses as UTC. Saving without noticing shifts the date.
3. `ComputeEffectiveAmounts` (`backend/pkg/svc/schedule.go`) carries the remaining of an overdue partial installment into the next installment's amount while the partial one still owes it. The calendar double counts (#3 showed 47.50 incl. 15 carried while #2 still owed 15).

## Decision (task 3)
Remove the carry-over: every installment keeps its own amount, and a partial installment keeps its own `remaining`. The cascade already distributes overpayments, so the carry is redundant and double counts. The "Incluye ... arrastrado" and "se descuenta de la siguiente cuota" UI/PDF texts go away.

## Scope
Backend `svc/schedule.go`, `svc/loans.go` (unpay), `app/loan_handlers.go` if needed; frontend `web/src/routes/loans/[id]/+page.svelte` and `web/src/lib/format.js`.

## Constraints
- Working on `main` (the user asked for direct work on main earlier in this project).
- The tree holds the user's uncommitted loans payment-history WIP in the same files (`loans.go`, `loan_handlers.go`, `model.go`, `+page.svelte`, migrations 002/003, `cascade.go`). Do not revert or reformat it. Commit only by hunks where separable; otherwise report.
- Spanish backend messages; English identifiers/comments.
- No emojis in UI; existing icon library only.

## Tasks
- [x] T1: Unpay reverts the whole last payment (all installments and the payment_history rows it touched), not just the last fully-paid installment.
- [x] T2: Date-only strings are parsed without timezone shift in the loan and installment edit dialogs.
- [x] T3: Remove the overdue-partial carry-over; calendar sums match the headline total.

## Acceptance criteria
- Pay 50 on a 32.50 installment, unpay: all installments back to `paid_amount=0`, payment history empty.
- Open edit installment on an installment due 28/11: dialog shows 28/11; saving unchanged keeps `2026-11-28`.
- With #2 partially paid and overdue, sum of per-installment remaining equals loan pending total; no installment amount is inflated.

## Checks / TDD
- Strict TDD enabled (source: project config), runner `go test ./...` from `backend/` (focused: `go test ./pkg/svc/`). RED before GREEN.
- Frontend: `npm run build` from `web/`.
- Functional: Playwright against local stack (backend :4300, `npm run dev` in `web/`), login in memory `chiro-local-dev-credentials`; use loan `loan_a10a79f3-fb0c-484e-9967-ddb49f5515d5` and restore data after.

## Routing
- Delegated direct: one writer (3 tasks across 2+ non-trivial files).

## Progress
- T1 (implemented, verified): new `svc.UnpayLastPayment` + pure `svc.ReversePaid` (cascade.go); the unpay handler now reverts the loan's most recent non-deleted payment (all its allocations), soft-deletes the history row, clears paid_date at 0, recomputes loan paid; falls back to per-installment unpay when the loan has no history. Frontend `unpay()` passes the first installment id (route unchanged) and the button shows when `payments.length > 0`.
  - RED: `go test ./pkg/svc/` -> `TestReversePaid` 4 failures (stub returned -1) and compile failure before the stub.
  - GREEN: `go test ./pkg/svc/` -> `ok chiro/pkg/svc`.
  - Functional (loan_a10a79f3..., API): cascade 50 -> #1 paid 32.5, #2 paid 17.5 (remaining 15); unpay -> all paid_amount 0, paid_date NULL, payment_history row deleted=1, 0 allocations, loan is_paid 0. Data restored (history row removed, updated_at restored).
- T2 (implemented, build OK): `isoParts` helper in `web/src/lib/format.js`; used in `openEdit` and `openEditInstallment` instead of `new Date(str)`. Not browser-tested. `npm run build` OK.
- T3 (implemented, verified): `ComputeEffectiveAmounts` removed (no-carry means identity); handler uses installments directly, `original_amount` = `amount`. Removed "Incluye ... arrastrado" and "se descuenta de la siguiente cuota" texts in page and PDF detail. RED: `TestComputeEffectiveAmounts_OverduePartialNoCarry` -> `inst 2: got amount 150, want 100`; GREEN after change; carry tests removed with the function.
- Browser check (Playwright, 375px), parent-run: pay 50,00 -> #2 default 15.00; calendar #3 stays 32.50 (15 + 4x32.50 = 145 = pending total); edit dialog shows 28/11 for #4; UI unpay -> all installments back to 0. T1/T2/T3 confirmed; test data restored.
- Parent re-ran `go build ./... && go test ./...` (32 passed) and `npm run build` (OK).
- Commits: pending user decision (fixes sit on top of uncommitted payment-history WIP in the same files).
- `go build ./... && go vet ./pkg/svc ./pkg/app && go test ./...` pass.
