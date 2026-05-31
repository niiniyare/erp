[<-- Back to Index](README.md)

## Accessibility

Accessibility in Awo ERP is a physical constraint — a forecourt operator with gloves, a finance manager with RSI navigating by keyboard, a line manager in bright sunlight on a low-quality screen.

### WCAG 2.1 AA Baseline

Every screen must pass WCAG 2.1 AA. The most frequently violated rules in ERP software:

**Colour Contrast**

```markdown
Normal text (< 18pt):              4.5:1 minimum
Large text (≥ 18pt or ≥ 14pt bold): 3:1 minimum
UI components, icons:               3:1 minimum
Kiosk/forecourt mode:               7:1 minimum (AAA) — direct sunlight degrades contrast
```

This is why tenant branding validates primary colours against contrast requirements. A tenant cannot choose a primary colour that makes buttons unreadable.

**Colour Is Never the Only Indicator**

Status badges must use colour AND text label. Chart lines must use colour AND pattern. Error states must use red border AND an icon. A user who cannot distinguish red from green must still understand status.

```json
{
  "name":     "status",
  "label":    "Status",
  "type":     "tag",
  "colorMap": { "confirmed": "success", "cancelled": "error" }
}
```

The `tag` type renders both the colour background AND the text label — correct by default.

**Keyboard Navigation**

Every interactive element must be reachable and operable by keyboard:
- Tab order follows visual reading order (left-to-right, top-to-bottom)
- Focus rings are visible and high-contrast — never hidden with `outline: none`
- Modal dialogs trap focus — Tab cycles only within the open dialog
- Dropdowns respond to Arrow keys and Enter
- AMIS handles most of this automatically for its own components

**Semantic Structure**

- Every page has exactly one `<h1>` (the page title)
- Form inputs have explicit labels, not just placeholder text
- Tables have proper `<thead>` and `<th>` elements
- Buttons have descriptive text or `aria-label` — not just an icon

### Kiosk / Forecourt Requirements

Beyond standard WCAG:

```markdown
Touch target size:
  Standard: minimum 48×48px for all interactive elements
  Kiosk mode: 60×80px for primary actions (allow for gloves)
  Spacing between adjacent targets: ≥8px (prevent accidental activation)

Session behaviour:
  Kiosk sessions do NOT time out during active use
  Inactivity timeout (default 5 min): return to locked start screen
  → The operator can re-authenticate without losing transaction state
  → Never log them out mid-transaction

Error messages in kiosk mode:
  ❌ "HTTP 503 Service Unavailable"
  ✅ "Could not save. Try again." with a large retry button

Sunlight readability:
  Use AAA contrast (7:1) in kiosk density mode
  Test on a bright screen in direct light if possible
```

### AMIS and Accessibility

AMIS components are generally keyboard-accessible and have reasonable ARIA attributes. Areas to verify:

```markdown
VERIFY for each new page:
✅ Tab order is logical within forms
✅ Dialog focus trap works (Tab stays inside open dialog)
✅ Dynamic content changes (toast appearing, table reloading) are announced
✅ Custom tpl HTML does not create inaccessible elements
✅ Images and icons have descriptive alt text or aria-label
✅ Charts have text alternatives (data table or description)
```

### Screen Reader Support

Dynamic content changes must be announced. AMIS handles this for its own components via ARIA live regions. For custom content in `tpl` or `html` components, add them manually:

```json
{
  "type":       "tpl",
  "tpl":        "<div role='status' aria-live='polite'>${status_message}</div>"
}
```

---
