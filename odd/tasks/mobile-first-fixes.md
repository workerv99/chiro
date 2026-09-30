# Mobile-first UX fixes

## Objective
Fix the gaps found by the mobile-first audit of `web/src`, so the app works well on phones (~375px).

## Problem
Breakpoints are already mobile-first. The real-device ergonomics are not:
- inputs trigger iOS zoom
- touch targets are under 44px
- fixed bottom elements do not handle the safe area
- `vh` units are used instead of `dvh`
- the Dialog cannot scroll

## Scope
Frontend only (`web/src`). No backend changes, no visual redesign.

## Constraints
- Mobile-first: base classes for small screens, `sm:`/`md:` (min-width) for larger ones. Never use `max-width` overrides.
- Touch targets must be at least 44x44px (`h-11`/`size-11`/`min-h-11`).
- Icons come from lucide only. No emoji or glyphs.
- Branch: `fix/mobile-first-ux`. The working tree holds pre-existing uncommitted work (ports, payment history, loans/[id] page). Stage only this feature's hunks.

## Tasks
- [x] T1: Form controls use 16px text. `ui/input.svelte`, native `<select>`/`<textarea>` (ExpenseModal, config, others) switch from `text-sm` to `text-base`.
- [x] T2: Touch targets are at least 44px.
  - `ui/button.svelte`: the default, `sm` and `icon` sizes.
  - `+layout.svelte`: nav tabs and logout.
  - `UndoToast.svelte`: undo and close.
  - `loans/[id]/+page.svelte:519`: the icon edit button.
  - `config/+page.svelte:280`: the color input.
  - Also fixed `ThemeSwitch.svelte` toggle buttons and the landing page login link (both were `h-9`).
- [x] T3: Safe area and dynamic viewport units.
  - FABs (dashboard, budgets, loans), UndoToast, CookieBanner and the main `pb-28` add `env(safe-area-inset-bottom)`.
  - `min-h-screen` becomes `min-h-dvh` (login, +layout, root +page).
  - `max-h-[92vh]` becomes `dvh` (config sheet).
  - The sticky nav header is `sticky`, not `fixed`, so it was left alone (no safe-area-inset-top needed).
- [x] T4: Dialog behaves as a bottom sheet on mobile and a centered dialog from `sm:`, with `max-h-[85dvh] overflow-y-auto`. This reuses the config sheet pattern.
- [x] T5: Low-priority polish.
  - Admin `text-[10px]` badges become `text-xs`.
  - Truncated strings get a `title` attribute (admin, config, dashboard, stats, loans/person).
  - Note: `loans/[id]/+page.svelte:463` also has a `text-[10px]` hint but was left as-is — out of the audit's explicit scope (admin only) and that file carries unrelated pre-existing edits.

## Acceptance criteria
- `npm run check`, `npm run lint` and `npm run build` pass. No new errors compared with the base.
- At a 375px viewport, the controls reviewed have tap areas of at least 44px, inputs have a 16px font size, and the ExpenseModal can scroll.

## Checks / TDD
- TDD: strict mode is enabled in the session config, but `web/` has no test runner (no vitest/playwright test setup). RED/GREEN therefore does not apply. The functional checks are `check`, `lint`, `build` and a Playwright visual check at 375px.

## Routing
- Delegated direct: one writer. The trigger is 2+ non-trivial files across about 15 files.

## Progress
- Audit done (engram obs #82).
- T1 done: commit f776eec.
- T2 done: commit fecc8f4.
- T3 done: commit b8a114f.
- T4 done: commit b11d52c.
- T5 done: commit fee8db1.
- Verification: `npm run check` 0 errors/0 warnings; `npm run lint` 28 errors, identical to base (no new errors); `npm run build` succeeded.

## Review assessment
- `gentle-ai review assess --base-ref main --committed-only --untracked-scope=exclude` returned: risk medium (executable_change), 153 changed lines, `review_due: false` (`under_budget`). No native review is due yet.
- Parent spot check: `npm run build` passes. The loans/[id] commit holds only this feature's hunks, and the unrelated edits stay unstaged.

## Next step
- Manual visual check at a 375px viewport, especially the Dialog bottom sheet and the FABs above the home indicator.
- One leftover `text-[10px]` at `loans/[id]/+page.svelte:463`, out of scope.
- Push and PR are the user's call.
