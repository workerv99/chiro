# Mobile bottom navigation

## Objective
Replace the squeezed top nav on phones with a thumb-reachable bottom tab bar. Keep the top horizontal nav from `sm:` up.

## Problem
`web/src/routes/+layout.svelte` renders a desktop-style top nav on every viewport:
- 6-7 text-only items at ~45px wide with `text-xs`
- "Log out" wraps to two lines
- logout sits among the tabs, where it is easy to hit by accident
- the top of the screen is the hardest area to reach one-handed

## Scope
Frontend only (`web/src`). The decisions below were approved by the user.
- Mobile (base): a fixed bottom tab bar with 5 tabs (Expenses, Stats, Budgets, Loans, Settings). Each tab has a lucide icon and a short label, and respects `env(safe-area-inset-bottom)`.
- Mobile: Logout and the Admin link (admin role only) move into the Settings page.
- `sm:` and up: the top horizontal nav, as today (logo + text tabs + Admin + Logout).
- Fixed elements (FABs, UndoToast, CookieBanner) must not overlap the bottom bar on mobile.

## Constraints
- Mobile-first: base = mobile, `sm:` (min-width) for larger screens. No max-width queries.
- Touch targets must be at least 44x44px.
- Icons come from `lucide-svelte` (already a dependency). No emoji or glyphs.
- i18n: reuse the existing keys (`tabs.*`, `common.logout`). Any new string needs both ES and EN.
- Stage only this feature's files. The working tree holds unrelated uncommitted work (backend, ports, payment history, `loans/[id]/+page.svelte`).

## Tasks
- [x] T1: The layout shows a bottom tab bar on mobile and the top nav from `sm:`.
- [x] T2: Logout and the Admin link appear in Settings on mobile.
- [x] T3: FABs, UndoToast and CookieBanner sit above the bottom bar on mobile. Main content padding clears the bar.

## Acceptance criteria
- At 375px: bottom bar visible, 5 tabs of at least 44px, active tab marked (`aria-current`), no horizontal overflow, nothing overlapping.
- At 1024px: top nav as before, bottom bar hidden.
- Logout reachable on mobile through Settings.

## Checks / TDD
- TDD: strict mode is on in the session config, but `web/` has no test runner. Functional checks instead:
  - `npm run check`: 0 errors at baseline.
  - `npm run lint`: 28 pre-existing errors at baseline.
  - `npm run build`
  - Playwright at 375px and 1024px.

## Routing
- Delegated direct: one writer. The trigger is 2+ non-trivial files (layout, config page, fixed-position components).

## Progress
- Branch `feat/mobile-bottom-nav` created from main (84608d8).
- T1 done: bottom tab bar (Receipt/ChartPie/PiggyBank/HandCoins/Settings icons), top nav now `hidden sm:flex`, `isActive()` uses startsWith for nested routes, main `pb` grown to clear the bar on mobile. Commit ba0596b2.
- T2 done: config page's Session card gained an `sm:hidden` block with an Admin link (Shield icon, admin role only) and a destructive-styled Logout button (LogOut icon); existing `sm:` logout button kept for desktop. Commit 5838889e.
- T3 done: dashboard/budgets/loans FABs got `sm:bottom-6` (mobile offset of 5rem already clears the 3.5rem bar by 1.5rem margin, verified by calc, no change needed there); CookieBanner raised to `bottom-[calc(4.5rem+safe-area)]` on mobile / `sm:bottom-[calc(1.5rem+safe-area)]` so it clears the bar; UndoToast's existing 6rem offset already clears the bar, left unchanged. Commit 226ce39f.
- Verification (from `web/`): `npm run check` → 0 errors (baseline: 0). `npm run lint` → 28 errors, same files/lines/messages as the pre-existing baseline, no new errors. `npm run build` → succeeded. Playwright visual checks at 375px/1024px deferred to the parent per its instructions (dev server / Playwright not run here).

## Next step
Parent to do visual verification at 375px and 1024px; otherwise the feature is complete and ready for review/PR.
