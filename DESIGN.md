---
name: Chiro
description: Personal expense and loan tracker. Quiet, numeric, fast to use on a phone.
colors:
  ink: "#09090b"
  graphite: "#18181b"
  paper: "#ffffff"
  mist: "#f4f4f5"
  hairline: "#e4e4e7"
  slate-text: "#71717a"
  ink-dark-surface: "#09090b"
  graphite-dark-surface: "#27272a"
  silver-text: "#a1a1aa"
  focus-ring-dark: "#d4d4d8"
  paid-green: "#22c55e"
  pending-amber: "#f59e0b"
  overdue-red: "#ef4444"
  idle-slate: "#94a3b8"
  destructive-deep: "#7f1d1d"
typography:
  title:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 700
    lineHeight: 1.33
  headline:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 700
    lineHeight: 1.55
  body:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.43
  field:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 700
    lineHeight: 1.33
    letterSpacing: "0.05em"
rounded:
  sm: "8px"
  md: "10px"
  lg: "12px"
  xl: "16px"
  full: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
components:
  button-primary:
    backgroundColor: "{colors.graphite}"
    textColor: "{colors.paper}"
    rounded: "{rounded.md}"
    height: "44px"
    padding: "8px 16px"
  button-outline:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    height: "44px"
    padding: "8px 16px"
  card:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.lg}"
    padding: "16px"
  input:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    height: "44px"
    padding: "8px 12px"
  badge:
    backgroundColor: "{colors.graphite}"
    textColor: "{colors.paper}"
    rounded: "{rounded.full}"
    padding: "2px 10px"
  nav-item-active:
    backgroundColor: "#18181b1a"
    textColor: "{colors.graphite}"
    rounded: "{rounded.full}"
    height: "44px"
---

# Design System: Chiro

## 1. Overview

**Creative North Star: "The Pocket Ledger"**

Chiro is a ledger you carry in a pocket. The amount is the loudest thing on any screen; everything around it stays quiet so a balance, an installment or a total reads at a glance. The system is monochrome and flat: near-black ink on white (or the reverse in dark mode), hairline borders, one solid button per screen, and color reserved for the state of money (paid, pending, overdue).

Density is moderate and mobile first. Controls are 44px tall, content sits in a single centered column no wider than 760px, and the phone gets a bottom tab bar while larger screens get a top bar. The system is built from the shadcn zinc palette, so it is coherent and predictable, but it carries no brand color yet; identity today comes from restraint and from numbers set bold, not from a signature hue.

This system explicitly rejects the default shadcn template look with no brand, cold corporate banking (navy and gold), neon fintech (gradients, glows and saturated color on dark), and the overloaded dashboard with many cards and charts competing before the content.

**Key Characteristics:**
- Flat surfaces separated by 1px hairlines; shadow only on what floats.
- Near-black graphite as the single action color; no secondary brand accent.
- Color means state: green paid, amber pending, red overdue, slate idle.
- System sans typography, with weight and size carrying the hierarchy.
- 44px minimum touch height on every control.
- Light and dark themes share one token set.

## 2. Colors

A zinc-neutral palette with graphite as the action color and four semantic state colors.

### Primary
- **Graphite** (#18181b): The action color. Solid primary buttons, the active nav pill (at 10% opacity) and focus rings in light mode. In dark mode this role inverts to a near-white (#fafafa).

### Neutral
- **Ink** (#09090b): Body text in light mode and the page background in dark mode.
- **Paper** (#ffffff): Page and card background in light mode.
- **Mist** (#f4f4f5): Muted, secondary and hover surfaces.
- **Hairline** (#e4e4e7): Borders and input strokes in light mode.
- **Slate Text** (#71717a): Secondary text and inactive nav labels in light mode.
- **Graphite Dark Surface** (#27272a): Muted surfaces, borders and inputs in dark mode.
- **Silver Text** (#a1a1aa): Secondary text in dark mode.

### State
- **Paid Green** (#22c55e): Settled loans and installments, positive balances.
- **Pending Amber** (#f59e0b): Open loans with money still due.
- **Overdue Red** (#ef4444): Overdue installments, expenses and destructive actions.
- **Idle Slate** (#94a3b8): Installments not yet due, neutral markers.

### Named Rules
**The State-Only Color Rule.** Outside the graphite action color, a color on screen must mean the state of money. Decorative color is prohibited.

**The One Action Rule.** One solid graphite button per screen. Every other action is outline, ghost or link.

**The Token Rule.** Status colors currently appear as literal hex values in several pages. New work must reference a token, not repeat a hex.

## 3. Typography

**Display Font:** none (the system sans stack is used throughout)
**Body Font:** ui-sans-serif, system-ui, sans-serif
**Label/Mono Font:** none

**Character:** Neutral and native. The system sans disappears into the task, and hierarchy comes from weight and size alone. The app does not load a web font.

### Hierarchy
- **Title** (700, 1.5rem at md and up, 1.25rem on phones, 1.33): Page headings such as "Loans" or a lender's name.
- **Headline** (700, 1.125rem, 1.55): Key amounts on summary cards.
- **Body** (400 to 600, 0.875rem, 1.43): List rows, descriptions and metadata. Prose stays under 65 to 75ch.
- **Field** (400, 1rem, 1.5): Text inside inputs and selects. Kept at 16px so mobile browsers do not zoom on focus.
- **Label** (700, 0.75rem, 0.05em tracking, uppercase): Card captions such as "REMAINING". Short labels only.

### Named Rules
**The Amount Leads Rule.** On any row or card the amount is the heaviest element after the page title.

**The 16px Field Rule.** Form fields never go below 1rem.

## 4. Elevation

The system is flat by default. Surfaces are told apart by a 1px hairline border, not by shadow. Shadow appears only on elements that sit above the page.

### Shadow Vocabulary
- **Card rest** (`box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05)`): The faint shadow-sm on cards, barely visible and used only to lift the border.
- **Floating** (`box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)`): Dialogs and the floating add button.

### Named Rules
**The Flat-By-Default Rule.** Surfaces are flat at rest. Shadow belongs to dialogs and the floating action button only.

**The Backdrop Blur Caveat.** The navigation bars use a translucent background with backdrop blur so content can scroll under them. This is functional, not decorative; do not extend blur to cards or dialogs.

## 5. Components

Firm and compact: 10 to 12px corners, 44px heights, one solid button per screen.

### Buttons
- **Shape:** Gently rounded (10px). Always 44px tall, including the small and large sizes.
- **Primary:** Graphite fill, near-white text, 8px by 16px padding, 14px medium weight.
- **Hover / Focus:** Primary dims to 90% on hover. Focus shows a 2px ring in the ring color with a 2px offset.
- **Outline:** 1px hairline border on the page background; hover fills with mist.
- **Ghost / Link:** No fill or border; hover fills with mist, link underlines.
- **Destructive:** Red fill, reserved for irreversible actions.

### Chips
- **Style:** Full-pill badge, 2px by 10px padding, 12px semibold. Variants: graphite fill (default), muted fill (secondary), red fill (destructive), bordered (outline).

### Cards / Containers
- **Corner Style:** 12px.
- **Background:** Card color (white or ink).
- **Shadow Strategy:** The faint card-rest shadow only, see Elevation.
- **Border:** 1px hairline.
- **Internal Padding:** 16px, tighter at 12px for compact stat cards.

### Inputs / Fields
- **Style:** 44px tall, 10px corners, 1px hairline stroke, page background, 12px horizontal padding, 16px text.
- **Focus:** 2px ring with 2px offset, no glow.
- **Error / Disabled:** Disabled drops to 50% opacity and blocks the cursor; errors appear as a red line of text at the end of the form.

### Navigation
- **Phone:** A fixed bottom bar with five tabs (Expenses, Stats, Budgets, Loans, Settings). Each tab stacks a 20px icon over an 11px label, with a 56px minimum height and safe-area padding. The active tab is shown only by a color change to graphite.
- **Desktop and tablet (sm and up):** A sticky top bar with a small "C" mark, text-only 12px semibold pills 44px tall, and "Log out" pushed to the right. The active pill is graphite at 10% opacity with graphite text.
- **Both:** Translucent background with backdrop blur and a 1px hairline edge.

### Loan Row (signature component)
A full-width row inside a card, 12 to 16px padding. A 10px status dot (green paid, amber open) leads, then a bold truncated title over a muted line holding the start date and installment count. The right side holds the remaining amount in red, or a green check with "Paid".

## 6. Do's and Don'ts

### Do:
- **Do** keep one solid graphite button per screen and make every other action outline, ghost or link.
- **Do** use color only for the state of money: #22c55e paid, #f59e0b pending, #ef4444 overdue, #94a3b8 idle.
- **Do** give every interactive element at least 44px of height.
- **Do** make the amount the heaviest element on a row or card, after the page title.
- **Do** keep form fields at 1rem so phones do not zoom on focus.
- **Do** carry state with text or an icon as well as color (a green check and the word "Paid").
- **Do** reference color tokens instead of repeating hex values in components.

### Don't:
- **Don't** let Chiro read as the default shadcn template: all zinc, black buttons and no identity.
- **Don't** use cold corporate banking styling: navy and gold, formal and distant.
- **Don't** use neon fintech styling: gradients, glows or saturated color on dark backgrounds.
- **Don't** build an overloaded dashboard: many cards and charts competing before the content.
- **Don't** use `border-left` or `border-right` wider than 1px as a colored stripe on cards, rows or alerts.
- **Don't** use gradient text, decorative glass cards, or shadows on resting surfaces.
- **Don't** convey paid, pending or overdue by color alone.
- **Don't** add a second accent color for decoration; state colors are the only color.
