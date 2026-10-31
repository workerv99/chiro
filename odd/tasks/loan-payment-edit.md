# loan-payment-edit

## Objective
On the loan detail page, list the payments made (date, amount, installments they were applied to) and let the user edit the amount (and date) of any specific payment.

## Problem
Backend already returns payments + allocations (`GET /api/loans/{id}/payments`) but the UI never lists them; only "undo last payment" exists. No payment edit exists.

## Scope (authorized)
- Backend: new service function + endpoint to edit a payment's amount/date, re-allocating across installments and recomputing the loan.
- Frontend: "Pagos realizados" section + edit dialog in `web/src/routes/loans/[id]/+page.svelte` and API helper.
- Tests: Go service test (cascade_test.go style), web vitest if applicable.

## Tasks
- [x] T1 Backend: EditPayment service + handler + route + Go tests
- [x] T2 Frontend: payments list + edit dialog + store/api helper (no web test: no meaningful pure unit)

## Route declaration
Delegated writer (2+ non-trivial files, preparation reading).

## Evidence
- T1: RED observed (undefined: ReallocatePayment), then GREEN. `go build ./... && go vet ./... && go test ./pkg/svc/...`: build/vet clean, 40 passed.
  Design: pure `ReallocatePayment` (cascade.go) reverses the payment's allocations in memory, then reuses `AllocateCascade` from the starting installment; later payments are untouched. `EditPayment` (loans.go) applies it in one ExecAll tx. Route PUT /api/loans/{id}/payments/{paymentId}; 400 Spanish on validation / over-balance, 404 on missing payment. No DB-backed test (none exists in repo); the tx/SQL layer is not exercised by tests.
- T2: `npm run -s check`: 0 errors 0 warnings; `npm run -s build`: ok; `npx vitest run`: 10 passed (no new web test).
- Pending: commits by parent; manual UI check not done.
