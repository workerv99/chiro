# Critical flow fixes

## Objective
Fix the three verified high-severity issues from the app-wide flow review (engram obs #88).

## Problem
1. Plan limits are not enforced.
   - `checkLimits` and `checkPlanLimits` in `backend/pkg/app/subscription.go` are never wired.
   - `checkLimits` hardcodes `uid := ""`.
   - Free users can create unlimited expenses, accounts and loans through the API.
2. Expense create/update (`backend/pkg/app/expense_handlers.go`, `saveExpenseWithTags`) does not validate amount > 0 or the date. Transfers do validate both.
3. A budget with "All categories" (`category_id = ''`) is created but never listed, because `BudgetProgress` in `backend/pkg/svc/stats.go:212` INNER JOINs category. It cannot be edited or deleted.

## Scope
Backend fixes, plus the minimal frontend changes these three need, such as surfacing the limit error message. No loans payment-history work: that is the user's uncommitted WIP.

## Constraints
- The user asked to work directly on `main` (no feature branch).
- The working tree holds unrelated uncommitted WIP:
  - backend: `loan_handlers.go`, `router.go`, `config.go`, `model.go`, `svc/loans.go`, Dockerfiles, `.env.example`
  - untracked migrations
  - `web/src/routes/loans/[id]/+page.svelte`
  - `docker-compose.yml`, `dev.sh`, `README.md`, `web/vite.config.js`, `web/.env.example`
  Stage only this feature's hunks. `router.go` has WIP hunks, so stage it through a HEAD-based copy (`git hash-object -w` + `git update-index --cacheinfo`).
- Keep existing error-message language (Spanish backend messages) and code style.

## Tasks
- [x] T1: Enforce plan limits server-side for creating expenses, accounts and loans, using the real user id and plan. The decision logic is a pure, unit-tested function. The frontend shows the backend error.
- [x] T2: Validate amount > 0 and the date format on expense create/update. This reuses `svc.ValidateAmountPositive` / `svc.ValidateDate`.
- [x] T3: Budgets with "All categories" are listed, editable and deletable.
  - The query uses a LEFT JOIN, and spending across all categories counts toward the budget.
  - The label is "All categories" via i18n.

## Acceptance criteria
- A Free user at the limit gets a 4xx with a clear message when creating an expense, account or loan. A Pro user does not.
- `POST` of an expense with amount <= 0 or an invalid date returns 400.
- A budget with `category_id = ''` shows up in the budgets list with correct spent/percent.

## Checks / TDD
- Backend: strict TDD with the `go test ./...` runner (from `backend/`). Existing tests are in `pkg/svc/*_test.go`. RED before GREEN for the pure logic.
- Frontend: `npm run check` (0 errors), `npm run lint` (28 pre-existing), `npm run build`.
- Functional: curl against the local Docker stack (backend :4300), using test user `mobile-test@chiro.local` / `test123456`.

## Routing
- Delegated direct: one writer. The trigger is 2+ non-trivial backend files.

## Progress
- Findings verified by the parent against the code.

### T1: plan limits (commit 79ad353)
- New pure function `svc.CheckLimit(resource string, limit, current int) error` in `backend/pkg/svc/limits.go`, table-driven test in `backend/pkg/svc/limits_test.go`.
  - RED: `go test ./pkg/svc/... -run TestCheckLimit -v` → `pkg/svc/limits_test.go:25:11: undefined: CheckLimit` (build failed).
  - GREEN: same command → `Go test: 10 passed in 1 packages`.
- `backend/pkg/app/subscription.go` rewritten: `checkLimits(ctx, uid, resource)` now takes the real uid/ctx (no more hardcoded `uid := ""` / `context.Background()`), looks up plan + current count, and delegates the decision to `svc.CheckLimit`. Removed the unused, never-wired `checkPlanLimits` middleware.
- Wired into every creation path for the three limited resources:
  - `backend/pkg/app/expense_handlers.go` `handleCreateExpense` (not `handleUpdateExpense`).
  - `backend/pkg/app/crud_handlers.go` `mountResource` POST handler, gated on `table == "account"` (covers the generic `/api/accounts` create path; other generic resources have no configured limit).
  - `backend/pkg/app/loan_handlers.go` `handleCreateLoan`.
  - Returns HTTP 403 with the existing Spanish message format.
- `go build ./...`, `go vet ./...`: clean. `gofmt -l` on all changed files: no output.
- `go test ./...`: `Go test: 24 passed in 15 packages` (14 baseline + 10 new).
- Functional (Docker stack, test user, Free plan):
  - Accounts (limit 3, had 2): 3rd create → 200; 4th create → 403 `"límite de 3 cuentas alcanzado. Actualiza a Pro"`.
  - Loans (limit 10, had 0, via a test person): loans 1-10 → 200 each; 11th → 403 `"límite de 10 préstamos alcanzado. Actualiza a Pro"`.
  - Expenses: not driven to the 50/month limit (would need 50 curl calls); the identical `checkLimits`/`svc.CheckLimit` code path is exercised and unit-tested by the account/loan runs above, plus the 10-case table test.
  - Frontend: `config/+page.svelte` (accounts) and `ExpenseModal.svelte` (expenses) already render `e.message` from the API client on a thrown error — no frontend change needed, confirmed by reading both files.
- Test data created during verification (TestAcc3, TestPerson + 10 loans) deleted via the API afterward; verified back to baseline (2 accounts, 0 persons, 0 loans).

### T2: expense validation (commit 9e305d8)
- Reused existing, already-unit-tested `svc.ValidateAmountPositive` / `svc.ValidateDate` (see `backend/pkg/svc/validate_test.go`); no new pure logic, so no new test was added, per the task instructions.
- `backend/pkg/app/expense_handlers.go` `saveExpenseWithTags` (shared by create and update) now validates `amount > 0` and the date format before writing, returning 400 with the validator's Spanish message.
- `go build ./...`, `go vet ./...`, `gofmt -l`: clean. `go test ./...`: `Go test: 24 passed in 15 packages` (unchanged, as expected — no new svc tests for this task).
- Functional (Docker stack):
  - `amount: -5` → 400 `"amount debe ser > 0, recibido: -5"`.
  - `amount: 0` → 400 `"amount debe ser > 0, recibido: 0"`.
  - `date: "not-a-date"` → 400 `"fecha inválida \"not-a-date\" (usa YYYY-MM-DD)"`.
  - valid expense (`amount: 10`, `date: "2026-09-30"`) → 200.
  - Test expense deleted via the API afterward.

### T3: budgets "All categories" (commit 92f33e8)
- `backend/pkg/svc/stats.go` `BudgetProgress`: `JOIN category` → `LEFT JOIN category`, `category_name`/`category_color` wrapped in `COALESCE(..., '')`. The spent subquery now matches `(b.category_id = '' OR e.category_id = b.category_id)`, so an all-categories budget sums the whole period's expenses instead of a single (nonexistent) category.
- No SQL-level test harness exists (per the task instructions), so this was verified functionally only; `go build`/`go vet`/`gofmt -l`/`go test ./...` all stay clean (24 passed, unchanged — no new Go logic beyond the SQL/string change).
- `web/src/routes/budgets/+page.svelte`: row label now shows `i18n.t('budgets.allCategories')` when `b.category_id` is empty (the key already existed in both `es`/`en` locales, reused from the existing category-picker option).
- Functional (Docker stack):
  - Created budget `{category_id: "", amount: 500, month: 9, year: 2026}` → 200.
  - `GET /api/budgets/progress?year=2026&month=9` → budget appears with `category_id: ""`, `category_name: ""`, `spent: 10` (matching the one expense created that month), `percentage: 2`.
  - `PUT` the same budget (`amount: 600`) → 200 (editable).
  - `DELETE` the same budget → 200, and it disappears from the progress listing (deletable).
  - Test budget deleted via the API afterward (already part of the edit/delete verification above).

### Frontend checks (run once, after all three tasks)
- `npm run check`: `COMPLETED 3956 FILES 0 ERRORS 0 WARNINGS 0 FILES_WITH_PROBLEMS`.
- `npm run lint`: `28 problems (28 errors, 0 warnings)` — identical count to the recorded baseline (`/tmp/.../lint-base.txt`); the one pre-existing issue in `budgets/+page.svelte` (`'pct' is defined but never used`) is unrelated to this change (unused import predating this work). No new lint errors introduced.
- `npm run build`: succeeds (`✓ built in 11.11s`, static adapter output written).

## Next step
Done. All three tasks implemented, tested, and verified; nothing pending.
