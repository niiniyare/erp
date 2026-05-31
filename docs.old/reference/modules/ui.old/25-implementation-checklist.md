[<-- Back to Index](README.md)

## Implementation Checklist

Use this before marking any screen or feature as done. Not every item applies to every screen — use judgement.

---

### Before You Start

- [ ] I know which persona is the **primary user** of this screen (see §13)
- [ ] I know which surface type this is: Dashboard, List, or Detail (see §14)
- [ ] I have checked whether an existing AMIS component already does what I need — I did not build a custom component for something AMIS provides

---

### Schema & Architecture

- [ ] The schema is the only source of truth for UI — no JS conditionals driving visibility
- [ ] Permissions are injected into page `data` from Go and read via `visibleOn` / `disabledOn`
- [ ] No hardcoded colour values (`#hex`) in the schema — only semantic CSS variables or AMIS colour names
- [ ] `crud` component has `syncLocation: true` — filter/page state persists in the URL
- [ ] All `chart` components have `"backgroundColor": "transparent"` in their config

---

### Information Architecture

- [ ] The screen answers all four questions above the fold: Where am I? What is the state? What can I do? What is the detail?
- [ ] Content is in the correct zone (navigation / context / workspace / status)
- [ ] Progressive disclosure applied — not everything is visible at Level 1

---

### Data Presentation (Tables)

- [ ] Default column selection is appropriate for the primary persona — not everything shown by default
- [ ] Column alignment follows data type rules (numbers right, text left, status center)
- [ ] Default sort is the most useful for the user's task — not `created_at desc` for everything
- [ ] Both empty states are designed: "no records exist" and "filters exclude everything"
- [ ] Bulk actions available where users manage multiple records of the same type

---

### Forms

- [ ] Single-column layout (no two-column form layouts)
- [ ] Labels are above inputs, not placeholder-only
- [ ] Optional fields are marked "optional" — not required fields marked with `*`
- [ ] Conditional fields use `visibleOn` + `clearValueOnHidden: true`
- [ ] Validation runs on blur — form-level AND field-level
- [ ] Error messages explain what is wrong AND how to fix it
- [ ] Smart defaults are pre-filled — user only changes what differs from the typical case
- [ ] Multi-step wizard has a review step as its final step

---

### Dark Mode

- [ ] Tested explicitly in dark mode — not assumed to work
- [ ] Any new overlay, dropdown, or portal component has an explicit `html.dark` CSS rule
- [ ] Charts render with transparent background — no white flash in dark mode
- [ ] Any custom HTML in `tpl` components uses CSS variables, not hardcoded colours

---

### Mobile

- [ ] Primary action is accessible without scrolling on a 375px screen
- [ ] Table does not require horizontal scrolling to see the actions column on mobile
- [ ] Forms use single-column layout — already required, but verify on mobile viewport
- [ ] Dialogs/drawers are tested on mobile — use `size: "full"` for complex content

---

### Navigation & State

- [ ] Filters persist in URL (`syncLocation: true`) — sharing URL preserves filter state
- [ ] Active menu item highlights correctly after navigate
- [ ] Breadcrumb reflects current location correctly

---

### Accessibility

- [ ] Colour contrast meets 4.5:1 for normal text, 3:1 for large text
- [ ] Status information is not conveyed by colour alone (text label always present)
- [ ] All interactive elements are keyboard-reachable in logical tab order
- [ ] Form inputs have explicit labels (not just placeholder)
- [ ] Touch targets are at least 44×44px — 48×48px preferred, 60×80px for kiosk

---

### Notifications & Feedback

- [ ] The correct notification type is used: Inline Alert / Toast / Notification Panel / Push (see §19)
- [ ] Success toast appears after mutations — user knows the action worked
- [ ] Error messages persist — not auto-dismissed toasts for errors

---

### API & Backend Contract

- [ ] List endpoint returns `{ items: [...], count: N }` under `data`
- [ ] Validation errors return `{ status: 422, errors: { field: "message" } }`
- [ ] Mutation response returns `{ status: 0, msg: "...", data: { id: "..." } }` so AMIS can redirect
- [ ] 401 returns AMIS envelope `{ status: 1, msg: "..." }` — not raw HTTP 401 without body
- [ ] DELETE responds 204 or `{ status: 0 }` — not an HTML error page

---
