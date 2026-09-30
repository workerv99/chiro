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
- [ ] T3: Safe area and dynamic viewport units.
  - FABs (dashboard, budgets, loans), UndoToast, CookieBanner and the main `pb-28` add `env(safe-area-inset-bottom)`.
  - `min-h-screen` becomes `min-h-dvh`.
  - `max-h-[92vh]` becomes `dvh`.
- [ ] T4: Dialog behaves as a bottom sheet on mobile and a centered dialog from `sm:`, with `max-h-[85dvh] overflow-y-auto`. This reuses the config sheet pattern.
- [ ] T5: Low-priority polish.
  - Admin `text-[10px]` badges become `text-xs`.
  - Truncated strings get a `title` attribute.

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

## Next step
T3.
