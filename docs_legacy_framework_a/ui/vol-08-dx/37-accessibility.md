> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: "37 – Accessibility"
volume: "vol-08-dx"
chapter: 37
section: "Developer Experience"
status: "baseline via AMIS SDK; formal audit planned"
---

# Chapter 37 – Accessibility

## Table of Contents
- [37.1 AMIS SDK Baseline Accessibility](#371-amis-sdk-baseline-accessibility)
- [37.2 What DSL Authors Control](#372-what-dsl-authors-control)
- [37.3 Known Gaps](#373-known-gaps)
- [37.4 Planned: Formal Audit and WCAG 2.1 Compliance](#374-planned-formal-audit-and-wcag-21-compliance)

---

## 37.1 AMIS SDK Baseline Accessibility

AwoERP UI uses the AMIS SDK for all rendered components.  AMIS provides a
meaningful accessibility baseline out of the box:

- **Form labels** are rendered via the AMIS `label` property on each form
  control; the SDK emits the corresponding `<label for="...">` HTML association.
- **ARIA roles** are applied by the SDK to complex widgets: data tables get
  `role="grid"`, modals get `role="dialog"`, tabs get `role="tablist"`.
- **Keyboard navigation** is handled by the SDK for dropdowns, date pickers,
  and modal dialogs using standard browser conventions.
- **Focus management** on modal open/close is managed by AMIS dialog components.
- **Error announcements** — AMIS validation errors are rendered adjacent to form
  fields and are associated via `aria-describedby`.

DSL authors inherit these properties at no cost: any `ast.FormNode` with a
`label` property automatically produces an accessible form control.

---

## 37.2 What DSL Authors Control

DSL code influences accessibility through schema properties exposed in the AST.

### Descriptive labels

Every column, form field, and stat node should carry a human-readable `Label`.
Avoid leaving `Label` empty — AMIS will render the raw field name which is
machine-readable but not accessible.

```go
// Good
ast.TableColumnNode{Name: "inv_date", Label: "Invoice Date", Type: "date"}

// Avoid
ast.TableColumnNode{Name: "inv_date"} // label defaults to "inv_date"
```

### Button descriptions

`ActionNode` buttons exposed in toolbars should have a `Tooltip` or descriptive
`Label`.  Icon-only buttons are inaccessible without a tooltip.

```go
ast.ActionNode{
    Label:   "Export",
    Icon:    "fa fa-download",
    Tooltip: "Export current page to CSV",
}
```

### Placeholder text

Form fields used as search inputs should carry a `Placeholder` that describes
the expected input, not just "Search...":

```go
ast.InputTextNode{
    Name:        "q",
    Placeholder: "Search by invoice number or customer name",
}
```

---

## 37.3 Known Gaps

The following accessibility deficiencies are known and accepted pending the
formal audit:

| Area | Gap |
|---|---|
| Color contrast | Dark-mode token overrides have not been contrast-checked against WCAG AA thresholds. |
| Screen reader testing | No systematic screen reader test suite exists. |
| Custom stat blocks | `StatNode` custom CSS properties may override AMIS SDK focus-ring styles. |
| CRUD row actions | Icon-only row actions (edit / delete icons) lack visible labels on small screens. |
| Dynamic content | `InitAPI` data loads do not announce completion to screen readers. |

---

## 37.4 Planned: Formal Audit and WCAG 2.1 Compliance

> **PLANNED** — not yet scheduled.

| Milestone | Description |
|---|---|
| Accessibility audit | Engage an accessibility specialist to audit the rendered UI against WCAG 2.1 Level AA. |
| Contrast token review | Validate all dark/light mode CSS custom property values against AA contrast ratios. |
| Screen reader test suite | Automated axe-core checks in the CI pipeline against key page templates. |
| Skip navigation links | Add a "Skip to main content" link in `web/pages/index.html`. |
| Announce dynamic loads | Emit `aria-live` region updates when `InitAPI` data loads complete. |
| Keyboard-only workflow | Verify every primary workflow (create, edit, approve) is completable without a mouse. |

Until the formal audit completes, DSL authors should follow the guidance in
§37.2 and avoid icon-only interactive elements without tooltips.
