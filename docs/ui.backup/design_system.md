# ERP System UI Design System

**Version:** 2.0  
**Last Updated:** October 2025  
**Status:** Active Development  
**Maintainer:** Design & Engineering Teams

---

## Table of Contents

1. [Executive Overview](#executive-overview)
2. [Design Philosophy](#design-philosophy)
3. [Design Tokens](#design-tokens)
4. [Typography System](#typography-system)
5. [Color & Status System](#color--status-system)
6. [Spacing & Layout Grid](#spacing--layout-grid)
7. [Component Library](#component-library)
8. [Data Presentation](#data-presentation)
9. [Document Workflows & States](#document-workflows--states)
10. [Navigation Architecture](#navigation-architecture)
11. [Forms & Data Entry](#forms--data-entry)
12. [Responsive Design Strategy](#responsive-design-strategy)
13. [Interactions & Motion](#interactions--motion)
14. [Accessibility & Inclusive Design](#accessibility--inclusive-design)
15. [Dark Mode & Themes](#dark-mode--themes)
16. [Performance Standards](#performance-standards)
17. [Implementation & Quality Assurance](#implementation--quality-assurance)

---

## Executive Overview

This design system powers an enterprise resource planning (ERP) interface serving operational managers, data entry clerks, finance officers, warehouse staff, and system administrators. The system prioritizes clarity and efficiency over visual flourish, with all decisions grounded in measurable user outcomes and accessibility standards.

**Core Metrics:**
- Page load time target: < 2.5s on 4G, < 1.5s on LTE
- Lighthouse score: 90+
- WCAG 2.1 AA compliance: 100%
- Mobile adoption: Fully responsive at 320px–1920px
- Support for extended workflows with undo/revision history

---

## Design Philosophy

### Five Core Principles

**1. Clarity First, Aesthetics Second**

The interface exists to serve workflows, not to impress. Every visual element must earn its place by improving comprehension or reducing errors. Users should complete tasks without documentation or support intervention. When clarity and aesthetics conflict, clarity always wins.

*Implementation:*
- Use whitespace generously to separate concerns
- Avoid decorative elements or animations that don't communicate state
- Test all critical paths with target users before shipping
- Provide inline guidance for ambiguous fields

**2. Consistency Reduces Friction**

Users learn patterns once and apply them everywhere. Inconsistent behavior creates confusion, increases support tickets, and slows task completion. Every interaction should follow predictable, documented rules.

*Implementation:*
- Reuse proven component patterns across modules
- Document all interaction patterns and variations
- Enforce design token usage through tooling
- Audit new features against consistency standards before release

**3. Speed Enables Productivity**

Fast load times, minimal clicks, and responsive feedback directly impact user throughput. Each second of delay costs business efficiency. Optimization is a feature, not a luxury.

*Implementation:*
- Measure and track performance metrics continuously
- Implement lazy loading for non-critical content
- Minimize bundle sizes through code splitting and tree shaking
- Provide keyboard shortcuts for frequent actions
- Support bulk operations for batch tasks

**4. Transparency Builds Trust**

Users need constant feedback. They should know what's happening, why it's happening, and what happens next. Silence creates anxiety and errors. System behavior should be predictable and explainable.

*Implementation:*
- Show loading states with estimated wait times
- Provide clear error messages with actionable solutions
- Display approval workflows and document relationships
- Maintain audit trails for all modifications
- Offer undo/revision history for data changes

**5. Flexibility Without Fragmentation**

Different user roles have different needs, but the system must remain coherent. Customization should be possible without breaking core usability or creating maintenance nightmares.

*Implementation:*
- Use role-based access control (RBAC) to show/hide features appropriately
- Allow dashboard personalization without requiring it
- Support custom fields through a standardized extension mechanism
- Provide configuration UI for power users without exposing it to all

---

## Design Tokens

All visual properties derive from a centralized token system. This ensures consistency, simplifies theming, and makes updates manageable.

### Color Tokens

**Primary Brand Colors**

| Token | Value | Contrast Ratio | Usage |
|-------|-------|---|---------|
| `primary-base` | #4a9b9b | - | Interactive states, focus indicators |
| `primary-dark` | #1a3a3a | 11.2:1 (AAA) | Sidebars, dark backgrounds |
| `primary-light` | #ecf7f7 | 1.1:1 (AA) | Hover states, subtle backgrounds |

**Neutral Scale** (for text, borders, backgrounds)

| Tone | Value | WCAG Level | Usage |
|------|-------|-----------|-------|
| `text-primary` | #1a1a1a | AAA | Body text, headings |
| `text-secondary` | #374151 | AAA | Secondary text, labels |
| `text-tertiary` | #6b7280 | AA | Disabled states, metadata |
| `border-light` | #e5e7eb | AA | Subtle dividers |
| `border-medium` | #d1d5db | AA | Standard borders |
| `background-page` | #f5f6f8 | AA | Page background |
| `background-surface` | #ffffff | AAA | Cards, panels |

**Semantic Status Colors** (with distinct symbols to support colorblindness)

| Status | Base Hex | Background | Icon | Text Color | Accessibility |
|--------|----------|-----------|------|-----------|---|
| Success | #10b981 | #dcfce7 | ✓ | #166534 | 5.8:1 AAA |
| Warning | #f59e0b | #fef3c7 | ⏱ | #92400e | 5.4:1 AAA |
| Info | #3b82f6 | #dbeafe | ℹ | #1e40af | 5.1:1 AAA |
| Error | #ef4444 | #fee2e2 | ✕ | #991b1b | 5.2:1 AAA |
| Neutral | #6b7280 | #f3f4f6 | — | #374151 | 7:1 AAA |

**Implementation (CSS)**

```css
:root {
  /* Colors */
  --color-primary: #4a9b9b;
  --color-primary-dark: #1a3a3a;
  --color-primary-light: #ecf7f7;
  --color-text-primary: #1a1a1a;
  --color-text-secondary: #374151;
  --color-text-tertiary: #6b7280;
  --color-border-light: #e5e7eb;
  --color-border-medium: #d1d5db;
  --color-bg-page: #f5f6f8;
  --color-bg-surface: #ffffff;
  --color-success: #10b981;
  --color-success-bg: #dcfce7;
  --color-warning: #f59e0b;
  --color-warning-bg: #fef3c7;
  --color-error: #ef4444;
  --color-error-bg: #fee2e2;
  --color-info: #3b82f6;
  --color-info-bg: #dbeafe;
  
  /* Typography */
  --font-family-ui: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-family-mono: 'IBM Plex Mono', 'Courier New', monospace;
  --font-size-xs: 11px;
  --font-size-sm: 12px;
  --font-size-base: 13px;
  --font-size-lg: 14px;
  --font-size-xl: 16px;
  --font-size-2xl: 20px;
  --font-size-3xl: 24px;
  --font-weight-regular: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  
  /* Spacing */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 12px;
  --space-lg: 16px;
  --space-xl: 24px;
  --space-2xl: 32px;
  --space-3xl: 48px;
  
  /* Transitions */
  --transition-quick: 150ms ease-out;
  --transition-base: 200ms ease-in-out;
  --transition-slow: 300ms ease-in-out;
  --transition-slower: 500ms ease-in-out;
  
  /* Shadows */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 20px 25px rgba(0, 0, 0, 0.15);
  
  /* Border radius */
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
  --radius-full: 9999px;
  
  /* Layout */
  --sidebar-width: 240px;
  --header-height: 56px;
  --max-content-width: 1400px;
  --grid-columns: 12;
}

/* Dark mode theme */
@media (prefers-color-scheme: dark) {
  :root {
    --color-text-primary: #f5f5f5;
    --color-text-secondary: #d1d1d1;
    --color-text-tertiary: #909090;
    --color-bg-page: #0a0a0a;
    --color-bg-surface: #1a1a1a;
    --color-border-light: #333333;
    --color-border-medium: #2d2d2d;
    --color-primary-light: #1a4f4f;
    /* Status colors remain same for consistency */
  }
}
```

### Spacing Scale

All spacing uses multiples of 4px. This creates a consistent rhythm and simplifies responsive adjustments.

| Token | Value | Primary Use Cases |
|-------|-------|---|
| `xs` | 4px | Tight spacing between related elements |
| `sm` | 8px | Internal component padding, gaps |
| `md` | 12px | Form rows, section spacing |
| `lg` | 16px | Container padding, card spacing |
| `xl` | 24px | Major section separation |
| `2xl` | 32px | Page-level spacing |
| `3xl` | 48px | Large section breaks, hero regions |

**Example Usage**

```css
.card {
  padding: var(--space-lg);
  margin-bottom: var(--space-xl);
}

.form-group {
  margin-bottom: var(--space-md);
}

.list-item {
  padding: var(--space-md) var(--space-lg);
  gap: var(--space-sm);
}
```

---

## Typography System

### Font Stack

**UI Font Stack** (system fonts for performance)

```css
font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 
             'Helvetica Neue', sans-serif;
```

This stack renders natively on each platform, ensuring optimal rendering and zero web font load time.

**Monospace Font Stack** (for codes, IDs, data)

```css
font-family: 'IBM Plex Mono', 'Courier New', monospace;
```

Used for document IDs, SKUs, currency amounts, and technical values where digit alignment matters.

### Type Scale

Hierarchy creates visual structure and guides user attention. Overuse of different sizes creates chaos; stick to this scale.

| Level | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| Display | 24px | 600 | 1.3 | Page titles, module headers, primary calls-to-action |
| Heading 1 | 20px | 600 | 1.3 | Section headings, document names |
| Heading 2 | 16px | 600 | 1.4 | Subsection titles, card headers |
| Heading 3 | 14px | 600 | 1.4 | Form section labels, list item titles |
| Body Large | 14px | 400 | 1.5 | Primary content, descriptions, long-form text |
| Body | 13px | 400 | 1.5 | Form labels, table content, standard text |
| Body Small | 12px | 400 | 1.6 | Helper text, metadata, timestamps, captions |
| Mono | 13px | 400 | 1.4 | Document IDs, codes, amounts, technical values |

**Line Height Rationale:** Tighter leading (1.3) for headings improves scan-ability; looser leading (1.5–1.6) for body text improves readability during extended reading sessions.

---

## Color & Status System

### Status Lifecycle

Every transaction document follows this workflow:

```
Draft → Submitted → Approved* → Completed
   ↓                  ↓
 Delete           Cancel (if allowed)
                     ↓
                   Reversed
                     
Rejected → Amended → (returns to Draft)
```

### Status Definitions & UI Treatment

**Draft** — Document created but not processed

```
Color: Gray (#6b7280)
Background: #f3f4f6
Icon: 📝
Badge: "Draft"
Permissions: Full edit, delete, submit
UI Treatment:
  - All fields editable
  - Delete button visible and enabled
  - Save button prominent
  - No approval chain shown
```

**Submitted** — Awaiting approval

```
Color: Blue (#3b82f6)
Background: #dbeafe
Icon: ✓
Badge: "Awaiting Approval" or document-specific status
Permissions: View-only, comments, request info, amend
UI Treatment:
  - Form fields disabled with full opacity (not grayed out)
  - Approval chain visible with current step highlighted
  - Show approver name, step duration if overdue
  - "Amend" button available to return to draft
  - Approver card appears with approve/reject buttons
```

**Approved** — Ready to process

```
Color: Green (#10b981)
Background: #dcfce7
Icon: ✓✓
Badge: "Approved"
Permissions: Process, cancel (if not started), view
UI Treatment:
  - Form fields read-only
  - Show completion actions based on document type
  - Historical approval chain in collapsed view
  - "Process" button replaces "Submit"
```

**Completed** — Fully processed (goods received, payment made, etc.)

```
Color: Blue (#3b82f6)
Background: #dbeafe
Icon: ✓
Badge: "Completed" / "Received" / "Invoiced" (context-dependent)
Permissions: View-only, reverse (if policy allows), export
UI Treatment:
  - All fields read-only
  - Show transaction reference numbers
  - Timestamp of completion
  - Reverse/undo option with confirmation
  - Archive option
```

**Rejected** — Requires revision

```
Color: Red (#ef4444)
Background: #fee2e2
Icon: ✕
Badge: "Rejected"
Permissions: Edit, resubmit, delete, view rejection reason
UI Treatment:
  - Show rejection reason in prominent alert
  - Allow returning to draft to fix issues
  - Highlight previously approved sections for context
  - "Resubmit" button instead of "Submit"
```

**Cancelled** — Voided, no longer valid

```
Color: Gray (#6b7280)
Background: #f3f4f6
Icon: ✕
Badge: "Cancelled"
Permissions: View-only, amend (creates new), archive
UI Treatment:
  - Strikethrough text or reduced opacity for content
  - Show cancellation reason if provided
  - Option to create amended version
```

### Colorblindness Support

Status is never communicated by color alone. Every status includes:

1. **Icon** — Distinct symbols (✓, ✕, ⏱, —) that don't rely on color
2. **Text** — Clear status label
3. **Context** — Position and surrounding elements

This approach supports:
- Red-green colorblindness (deuteranopia, protanopia)
- Blue-yellow colorblindness (tritanopia)
- Complete colorblindness (achromatopsia)

**Testing:** Use a colorblind simulator (e.g., Chrome DevTools simulation) before shipping.

---

## Spacing & Layout Grid

### Grid System

**Base Grid:** 4px  
**Column System:** 12-column responsive grid  
**Container Max-Width:** 1400px  
**Gutter:** 16px (8px on each side)

```css
.layout-container {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: var(--space-lg); /* 16px */
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: 0 var(--space-lg);
}

/* Example: 3-column layout */
.sidebar { grid-column: 1 / 4; }
.main { grid-column: 4 / -1; }

/* Example: Full-width content */
.full-width { grid-column: 1 / -1; }
```

### Common Spacing Patterns

**Page-Level Spacing**

```
Page padding: 24–32px (2xl–3xl)
Section separation: 32px (2xl)
Subsection spacing: 24px (xl)
```

**Form Spacing**

```
Form group (label + input): 16px vertical gap (lg)
Input height: 36–44px
Form section padding: 20px
Section title margin-bottom: 16px
Error message margin-top: 4px (xs)
```

**Card/Panel Spacing**

```
Card padding: 16px (lg) for standard, 20px for emphasis
Card gap between items: 8–12px (sm–md)
Title margin-bottom: 12px (md)
Action gap: 8px (sm)
```

**Table/List Spacing**

```
Row height: 40px minimum
Cell padding: 10–12px
Row gap: 1px border
Header padding: 10px 12px
```

---

## Component Library

### 1. Buttons

**Anatomy**

```
┌─ padding-left ────┬─ icon ─┬─ gap ─┬─ text ─┬─ padding-right ─┐
└──────────────────┴────────┴───────┴────────┴────────────────┘
```

**Size Variants**

| Size | Padding | Font Size | Min Width | Use Case |
|------|---------|-----------|-----------|----------|
| SM | 6px 12px | 12px | 56px | Table actions, secondary |
| MD | 8px 16px | 12px | 64px | Standard, default |
| LG | 12px 20px | 13px | 80px | Primary actions, modals |

**Style Variants**

```css
/* Primary (CTA) */
.btn-primary {
  background: var(--color-primary);
  color: white;
  border: none;
  
  &:hover { background: #3d7d7d; }
  &:active { background: #2d5555; }
  &:disabled { background: var(--color-border-medium); }
}

/* Secondary */
.btn-secondary {
  background: white;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-medium);
  
  &:hover { background: #f9fafb; }
  &:active { background: #e5e7eb; }
  &:disabled { color: var(--color-text-tertiary); border-color: var(--color-border-medium); }
}

/* Success (Approve, Save) */
.btn-success {
  background: var(--color-success);
  color: white;
  
  &:hover { background: #059669; }
  &:active { background: #047857; }
}

/* Danger (Delete, Reject) */
.btn-danger {
  background: var(--color-error);
  color: white;
  
  &:hover { background: #dc2626; }
  &:active { background: #b91c1c; }
}

/* Ghost (Minimal) */
.btn-ghost {
  background: transparent;
  color: var(--color-primary);
  border: none;
  
  &:hover { background: var(--color-primary-light); }
  &:active { background: rgba(74, 155, 155, 0.2); }
}
```

**Common States**

```
Default: Base style
Hover: Color shift, subtle shadow
Active/Pressed: Scale down 0.98, held state
Focus: 3px solid outline with 2px offset
Disabled: Grayed out, cursor: not-allowed
Loading: Spinner replaces text, button disabled
```

### 2. Form Controls

**Input Field States**

```css
.input {
  padding: 8px 12px;
  border: 1px solid var(--color-border-medium);
  border-radius: var(--radius-md);
  font: var(--font-size-base) / 1.5 var(--font-family-ui);
  background: var(--color-bg-surface);
  transition: var(--transition-base);
  height: 36px;
  
  &:hover { border-color: #9ca3af; }
  &:focus {
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px var(--color-primary-light);
    outline: none;
  }
  
  &:disabled {
    background: var(--color-bg-page);
    border-color: var(--color-border-light);
    color: var(--color-text-tertiary);
    cursor: not-allowed;
  }
  
  &.error {
    border-color: var(--color-error);
    background: var(--color-error-bg);
  }
  
  &.success {
    border-color: var(--color-success);
    padding-right: 40px; /* Space for icon */
  }
}
```

**Form Labels**

```css
.label {
  display: block;
  font: var(--font-weight-medium) var(--font-size-sm) / 1.5 var(--font-family-ui);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-xs);
  letter-spacing: -0.2px;
  
  .required {
    color: var(--color-error);
    font-weight: bold;
    margin-left: 2px;
  }
}

.label.optional {
  &::after {
    content: ' (Optional)';
    color: var(--color-text-tertiary);
    font-weight: normal;
    font-size: var(--font-size-xs);
  }
}
```

**Helper Text & Validation**

```css
.helper-text {
  font: var(--font-size-xs) var(--font-family-ui);
  color: var(--color-text-tertiary);
  margin-top: 4px;
}

.error-message {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font: var(--font-weight-medium) var(--font-size-xs) var(--font-family-ui);
  color: #991b1b;
  background: var(--color-error-bg);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  margin-top: var(--space-xs);
  
  &::before { content: '⚠'; }
}

.success-message {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font: var(--font-weight-medium) var(--font-size-xs) var(--font-family-ui);
  color: #166534;
  background: var(--color-success-bg);
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
  margin-top: var(--space-xs);
  
  &::before { content: '✓'; }
}
```

**Checkboxes & Radio Buttons**

```css
.checkbox,
.radio {
  width: 16px;
  height: 16px;
  border: 2px solid var(--color-border-medium);
  border-radius: var(--radius-sm);
  background: var(--color-bg-surface);
  cursor: pointer;
  transition: var(--transition-quick);
  appearance: none;
  
  &:hover { border-color: var(--color-primary); }
  
  &:focus {
    outline: 2px solid var(--color-primary-light);
    outline-offset: 2px;
  }
  
  &:checked {
    background: var(--color-primary);
    border-color: var(--color-primary);
    
    &::after { content: '✓'; color: white; }
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.radio { border-radius: 50%; }
```

### 3. Status Badges

**Badge Component**

```css
.badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  font: var(--font-weight-medium) var(--font-size-xs) var(--font-family-ui);
  height: 24px;
  white-space: nowrap;
  
  &.success {
    background: var(--color-success-bg);
    color: #166534;
    
    &::before { content: '✓'; }
  }
  
  &.warning {
    background: var(--color-warning-bg);
    color: #92400e;
    
    &::before { content: '⏱'; }
  }
  
  &.error {
    background: var(--color-error-bg);
    color: #991b1b;
    
    &::before { content: '✕'; }
  }
  
  &.info {
    background: var(--color-info-bg);
    color: #1e40af;
    
    &::before { content: 'ℹ'; }
  }
  
  &.neutral {
    background: #f3f4f6;
    color: var(--color-text-secondary);
    
    &::before { content: '—'; }
  }
}

/* Document status badge (larger, more prominent) */
.badge-document {
  padding: 6px 12px;
  font: var(--font-weight-semibold) var(--font-size-sm) var(--font-family-ui);
  border: 1px solid currentColor;
  opacity: 0.9;
  box-shadow: var(--shadow-sm);
}
```

### 4. Tables

**Structure & Styling**

```css
table {
  width: 100%;
  border-collapse: collapse;
  background: var(--color-bg-surface);
}

thead {
  background: #f9fafb;
  
  th {
    padding: 10px 12px;
    text-align: left;
    font: var(--font-weight-semibold) var(--font-size-sm) var(--font-family-ui);
    color: var(--color-text-tertiary);
    border-bottom: 1px solid var(--color-border-light);
    letter-spacing: -0.2px;
    user-select: none;
    
    &[data-sortable] {
      cursor: pointer;
      
      &::after { content: '↕'; opacity: 0; margin-left: 4px; }
      
      &:hover::after { opacity: 0.5; }
      &.sort-asc::after { content: '▲'; opacity: 1; }
      &.sort-desc::after { content: '▼'; opacity: 1; }
    }
  }
}

tbody tr {
  border-bottom: 1px solid #f3f4f6;
  transition: var(--transition-quick);
  
  &:hover { background: #f9fafb; }
  
  &.selected {
    background: var(--color-primary-light);
    border-left: 3px solid var(--color-primary);
  }
}

td {
  padding: 10px 12px;
  font: var(--font-size-sm) var(--font-family-ui);
  color: var(--color-text-primary);
  height: 40px;
  
  &[data-type="number"] {
    text-align: right;
    font-family: var(--font-family-mono);
  }
  
  &[data-type="date"] {
    font-family: var(--font-family-mono);
  }
}

/* Sticky header on scroll */
@media (prefers-reduced-motion: no-preference) {
  thead {
    position: sticky;
    top: 0;
    z-index: 10;
  }
}
```

**Responsive Table Handling**

For screens smaller than 768px, tables switch to card layout or horizontal scrolling with sticky first column:

```css
@media (max-width: 767px) {
  table {
    display: block;
    overflow-x: auto;
    
    thead { display: none; }
    
    tbody,
    tr,
    td {
      display: block;
    }
    
    tr {
      border: 1px solid var(--color-border-light);
      border-radius: var(--radius-md);
      margin-bottom: var(--space-md);
      padding: var(--space-md);
    }
    
    td {
      &::before {
        content: attr(data-label);
        font-weight: var(--font-weight-semibold);
        color: var(--color-text-secondary);
        display: block;
        margin-bottom: var(--space-xs);
        font-size: var(--font-size-xs);
      }
    }
  }
}
```

### 5. Cards & Panels

```css
.card {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  box-shadow: var(--shadow-sm);
  transition: var(--transition-base);
  
  &:hover {
    box-shadow: var(--shadow-md);
    border-color: var(--color-border-medium);
  }
  
  &.clickable {
    cursor: pointer;
    
    &:active { transform: scale(0.99); }
  }
}

/* Related document card (left border accent) */
.card-document {
  border-left: 4px solid currentColor;
  
  &.purchase { --accent-color: #3b82f6; }
  &.sales { --accent-color: #8b5cf6; }
  &.finance { --accent-color: #f59e0b; }
  &.inventory { --accent-color: #10b981; }
  
  border-left-color: var(--accent-color);
}

.card-header {
  font: var(--font-weight-semibold) var(--font-size-lg) var(--font-family-ui);
  color: var(--color-text-primary);
  margin-bottom: var(--space-md);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-body {
  font: var(--font-size-base) var(--font-family-ui);
  color: var(--color-text-primary);
}

.card-footer {
  margin-top: var(--space-lg);
  padding-top: var(--space-lg);
  border-top: 1px solid var(--color-border-light);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-md);
}
```

### 6. Modals & Dialogs

```css
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  opacity: 0;
  animation: fadeIn var(--transition-base) forwards;
}

@keyframes fadeIn { to { opacity: 1; } }

.modal {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0.95);
  background: var(--color-bg-surface);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  z-index: 1001;
  animation: slideUp var(--transition-base) forwards;
  
  @keyframes slideUp {
    to { transform: translate(-50%, -50%) scale(1); }
  }
}

.modal-header {
  padding: var(--space-xl) var(--space-2xl);
  border-bottom: 1px solid var(--color-border-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
  
  h2 {
    font: var(--font-weight-semibold) var(--font-size-xl) var(--font-family-ui);
  }
  
  .close {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0.5;
    transition: opacity var(--transition-quick);
    
    &:hover { opacity: 1; }
  }
}

.modal-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-xl) var(--space-2xl);
}

.modal-footer {
  padding: var(--space-lg) var(--space-2xl);
  border-top: 1px solid var(--color-border-light);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-md);
}
```

### 7. Approval Workflow Visualization

**Approval Chain Display**

```css
.approval-chain {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
  padding: var(--space-lg);
  background: #f9fafb;
  border-radius: var(--radius-md);
  overflow-x: auto;
}

.approval-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  min-width: 100px;
  
  .step-icon {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    color: white;
    
    &.completed { background: var(--color-success); }
    &.current { background: var(--color-primary); animation: pulse 2s infinite; }
    &.pending { background: var(--color-border-medium); }
    &.rejected { background: var(--color-error); }
  }
  
  .step-label {
    font: var(--font-size-xs) var(--font-family-ui);
    text-align: center;
    color: var(--color-text-secondary);
  }
  
  .step-detail {
    font: var(--font-size-xs) var(--font-family-ui);
    text-align: center;
    color: var(--color-text-tertiary);
  }
}

.approval-connector {
  width: var(--space-lg);
  height: 2px;
  background: var(--color-border-light);
  
  &.completed { background: var(--color-success); }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
```

**Current Approver Card**

```css
.approver-card {
  background: var(--color-info-bg);
  border-left: 4px solid var(--color-info);
  border-radius: var(--radius-md);
  padding: var(--space-lg);
  margin: var(--space-lg) 0;
  
  .approver-name {
    font: var(--font-weight-semibold) var(--font-size-base) var(--font-family-ui);
    color: var(--color-text-primary);
  }
  
  .approver-role {
    font: var(--font-size-xs) var(--font-family-ui);
    color: var(--color-text-secondary);
    margin-bottom: var(--space-md);
  }
  
  .time-pending {
    font: var(--font-size-xs) var(--font-family-ui);
    margin-bottom: var(--space-md);
    
    &.overdue {
      color: var(--color-error);
      font-weight: var(--font-weight-semibold);
    }
  }
  
  .actions {
    display: flex;
    gap: var(--space-md);
  }
  
  textarea {
    width: 100%;
    margin-top: var(--space-md);
    padding: var(--space-md);
    border: 1px solid var(--color-border-medium);
    border-radius: var(--radius-md);
    font: var(--font-size-sm) var(--font-family-ui);
    min-height: 60px;
  }
}
```

---

## Data Presentation

### Lists & Collections

**List Header**

```css
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-lg);
  background: var(--color-bg-surface);
  border-bottom: 1px solid var(--color-border-light);
  
  .list-title {
    font: var(--font-weight-semibold) var(--font-size-lg) var(--font-family-ui);
    color: var(--color-text-primary);
    
    .result-count {
      color: var(--color-text-tertiary);
      font-weight: normal;
      font-size: var(--font-size-sm);
    }
  }
  
  .list-controls {
    display: flex;
    gap: var(--space-md);
    align-items: center;
  }
}
```

**Filter Bar**

```css
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
  padding: var(--space-lg);
  background: var(--color-bg-page);
  border-bottom: 1px solid var(--color-border-light);
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  background: #f3f4f6;
  border: 1px solid var(--color-border-medium);
  border-radius: var(--radius-full);
  font: var(--font-size-sm) var(--font-family-ui);
  cursor: pointer;
  transition: var(--transition-quick);
  
  &:hover { background: #e5e7eb; }
  
  &.active {
    background: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
  }
  
  .remove {
    opacity: 0.6;
    cursor: pointer;
    
    &:hover { opacity: 1; }
  }
}
```

**Bulk Selection & Actions**

```css
.bulk-action-bar {
  position: sticky;
  top: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-lg);
  background: var(--color-primary-light);
  border-bottom: 1px solid var(--color-primary);
  z-index: 100;
  animation: slideDown 0.2s ease-out;
  
  @keyframes slideDown { from { transform: translateY(-100%); } }
  
  .selection-info {
    font: var(--font-weight-medium) var(--font-size-sm) var(--font-family-ui);
    color: var(--color-text-primary);
  }
  
  .bulk-actions {
    display: flex;
    gap: var(--space-md);
  }
}
```

**Empty State**

```css
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-3xl) var(--space-lg);
  min-height: 300px;
  text-align: center;
  
  .empty-icon {
    font-size: 64px;
    opacity: 0.2;
    margin-bottom: var(--space-xl);
  }
  
  .empty-title {
    font: var(--font-weight-semibold) var(--font-size-xl) var(--font-family-ui);
    color: var(--color-text-primary);
    margin-bottom: var(--space-sm);
  }
  
  .empty-description {
    font: var(--font-size-base) var(--font-family-ui);
    color: var(--color-text-tertiary);
    max-width: 400px;
    margin-bottom: var(--space-xl);
  }
  
  .empty-action {
    display: flex;
    gap: var(--space-md);
  }
}
```

### Pagination

```css
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-lg);
  background: #f9fafb;
  border-top: 1px solid var(--color-border-light);
  
  .pagination-info {
    font: var(--font-size-sm) var(--font-family-ui);
    color: var(--color-text-tertiary);
  }
  
  .pagination-controls {
    display: flex;
    gap: var(--space-sm);
    align-items: center;
  }
  
  button {
    width: 28px;
    height: 28px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--color-border-medium);
    background: white;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: var(--font-size-xs);
    transition: var(--transition-quick);
    
    &:hover { background: #f3f4f6; }
    &.active { background: var(--color-primary); color: white; border-color: var(--color-primary); }
    &:disabled { opacity: 0.5; cursor: not-allowed; }
  }
  
  .page-size-selector {
    font: var(--font-size-sm) var(--font-family-ui);
    padding: 4px 8px;
    border: 1px solid var(--color-border-medium);
    border-radius: var(--radius-sm);
    background: white;
  }
}
```

---

## Document Workflows & States

### Workflow Patterns by Document Type

Different documents follow different paths. Define the exact flow for each:

**Purchase Order Flow**

```
Draft → Submitted → Approved → Received → Invoiced → Paid
              ↓
           Rejected → Amended
```

**Sales Order Flow**

```
Draft → Submitted → Approved → Picked → Shipped → Invoiced
              ↓
           Rejected → Amended
```

**Expense Claim Flow**

```
Draft → Submitted → Approved → Reimbursed
              ↓
           Rejected → Amended
```

### Action State Machine

**From Draft State**

- Submit (if all required fields filled)
- Save (Ctrl+S)
- Delete
- Discard changes

**From Submitted State**

- Approve (if user has approver role)
- Reject (if user has approver role)
- Request Info (from approver)
- Amend (anyone, creates new version as Draft)
- Print
- Download

**From Approved State**

- Process/Receive/Complete (transition to Completed)
- Cancel (if not partially processed)
- Amend
- Print

**From Completed State**

- Reverse (if policy allows)
- Amend (create new version)
- Archive
- Print
- Export

**From Rejected State**

- Edit (returns to Draft)
- Resubmit
- Delete
- View rejection reason

### Error Scenarios & Recovery

**Scenario: Form submission fails**

```
1. Show error message at top of form
2. Highlight all invalid fields
3. Disable submit button
4. Show specific guidance on each field
5. Auto-scroll to first error
6. Allow user to fix and retry
```

**Scenario: Network timeout during approval**

```
1. Show alert: "Network connection lost. Your changes may not have been saved."
2. Provide "Retry" button
3. Provide "Work offline" option (if applicable)
4. Never silently fail
```

**Scenario: Document locked (concurrent edit)**

```
1. Show notification: "This document is being edited by [User]. Changes you make will not be saved."
2. Offer "Request control" or "View their changes"
3. Disable submit/save buttons
4. Allow read-only view
```

---

## Navigation Architecture

### Information Architecture Hierarchy

```
ERP System
├── Dashboard
│   ├── My Approvals
│   ├── My Tasks
│   └── Key Metrics
├── Procurement
│   ├── Purchase Orders
│   ├── Vendors
│   ├── Receipts
│   └── Settings
├── Sales
│   ├── Sales Orders
│   ├── Customers
│   ├── Shipments
│   └── Settings
├── Inventory
│   ├── Items
│   ├── Warehouses
│   ├── Stock Transfers
│   └── Settings
├── Finance
│   ├── Invoices
│   ├── Payments
│   ├── Accounts
│   └── Reporting
└── Administration
    ├── Users
    ├── Roles & Permissions
    ├── System Settings
    └── Audit Log
```

### Sidebar Navigation

```css
.sidebar {
  width: var(--sidebar-width);
  background: var(--color-primary-dark);
  color: white;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow-y: auto;
  padding: var(--space-xl) var(--space-lg);
  position: fixed;
  left: 0;
  top: 0;
  z-index: 900;
  
  @media (max-width: 1023px) {
    transform: translateX(-100%);
    transition: transform var(--transition-base);
    
    &.open { transform: translateX(0); }
  }
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  font: var(--font-weight-semibold) var(--font-size-lg) var(--font-family-ui);
  margin-bottom: var(--space-xl);
  padding-bottom: var(--space-lg);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-section {
  margin-bottom: var(--space-xl);
  
  .section-title {
    font: var(--font-weight-semibold) var(--font-size-xs) var(--font-family-ui);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: rgba(255, 255, 255, 0.5);
    margin-bottom: var(--space-md);
  }
}

.nav-item {
  padding: var(--space-md) var(--space-md);
  margin-bottom: var(--space-xs);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: var(--transition-quick);
  display: flex;
  align-items: center;
  gap: var(--space-md);
  
  &:hover { background: rgba(255, 255, 255, 0.1); }
  
  &.active {
    background: var(--color-primary);
    font-weight: var(--font-weight-medium);
  }
  
  &.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.nav-badge {
  margin-left: auto;
  background: var(--color-error);
  color: white;
  padding: 2px 6px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}
```

### Top Header/Navigation Bar

```css
.header {
  position: fixed;
  top: 0;
  left: var(--sidebar-width);
  right: 0;
  height: var(--header-height);
  background: var(--color-bg-surface);
  border-bottom: 1px solid var(--color-border-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--space-xl);
  z-index: 50;
  box-shadow: var(--shadow-sm);
  
  @media (max-width: 1023px) {
    left: 0;
  }
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
}

.hamburger {
  display: none;
  width: 32px;
  height: 32px;
  cursor: pointer;
  
  @media (max-width: 1023px) {
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

.search-box {
  width: 280px;
  position: relative;
  
  input {
    width: 100%;
    padding: var(--space-sm) var(--space-md);
    padding-left: 32px;
    background: var(--color-bg-page);
    border: 1px solid var(--color-border-light);
    border-radius: var(--radius-md);
    font: var(--font-size-base) var(--font-family-ui);
    
    &:focus {
      background: white;
      border-color: var(--color-primary);
    }
  }
  
  .search-icon {
    position: absolute;
    left: var(--space-md);
    top: 50%;
    transform: translateY(-50%);
    color: var(--color-text-tertiary);
    pointer-events: none;
  }
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
}

.header-icons {
  display: flex;
  gap: var(--space-md);
  
  button {
    width: 40px;
    height: 40px;
    background: none;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--color-text-tertiary);
    border-radius: var(--radius-md);
    transition: var(--transition-quick);
    position: relative;
    
    &:hover {
      color: var(--color-text-primary);
      background: var(--color-bg-page);
    }
    
    .badge {
      position: absolute;
      top: 8px;
      right: 8px;
      background: var(--color-error);
      color: white;
      width: 20px;
      height: 20px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 10px;
      font-weight: var(--font-weight-semibold);
    }
  }
}

.user-menu {
  position: relative;
  
  button {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--color-primary-light);
    border: 2px solid var(--color-primary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: var(--font-weight-semibold);
    color: var(--color-primary);
  }
  
  .dropdown {
    position: absolute;
    top: calc(100% + var(--space-sm));
    right: 0;
    background: var(--color-bg-surface);
    border: 1px solid var(--color-border-light);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    min-width: 200px;
    z-index: 1000;
    display: none;
    
    &.open { display: block; }
    
    a, button {
      display: block;
      width: 100%;
      padding: var(--space-md) var(--space-lg);
      text-align: left;
      border: none;
      background: none;
      cursor: pointer;
      font: var(--font-size-sm) var(--font-family-ui);
      color: var(--color-text-primary);
      transition: var(--transition-quick);
      
      &:hover { background: var(--color-bg-page); }
      
      &:first-child { border-top-left-radius: var(--radius-md); border-top-right-radius: var(--radius-md); }
      &:last-child { border-bottom-left-radius: var(--radius-md); border-bottom-right-radius: var(--radius-md); }
    }
    
    hr {
      margin: var(--space-xs) 0;
      border: none;
      border-top: 1px solid var(--color-border-light);
    }
  }
}
```

### Breadcrumb Navigation

```css
.breadcrumb {
  display: flex;
  align-items: center;
  padding: var(--space-md) var(--space-xl);
  background: #f9fafb;
  border-bottom: 1px solid var(--color-border-light);
  font: var(--font-size-sm) var(--font-family-ui);
  color: var(--color-text-tertiary);
  gap: var(--space-sm);
  
  a {
    color: var(--color-primary);
    text-decoration: none;
    transition: color var(--transition-quick);
    
    &:hover { color: var(--color-text-primary); text-decoration: underline; }
  }
  
  .current {
    color: var(--color-text-primary);
    font-weight: var(--font-weight-medium);
  }
  
  .separator {
    color: var(--color-text-tertiary);
  }
}
```

---

## Forms & Data Entry

### Form Patterns

**Single Column (Default)**

Use for simple workflows with 5–8 fields.

```css
.form-single {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-lg);
  max-width: 600px;
}
```

**Two Column**

Use for related field pairs (dates, amounts, names).

```css
.form-two-column {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-lg);
  
  @media (max-width: 767px) {
    grid-template-columns: 1fr;
  }
}
```

**Multi-Section with Collapsible Groups**

```css
.form-section {
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  overflow: hidden;
  margin-bottom: var(--space-lg);
  
  .section-header {
    padding: var(--space-lg);
    background: #f9fafb;
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: var(--font-weight-semibold);
    transition: var(--transition-quick);
    user-select: none;
    
    &:hover { background: #f3f4f6; }
    
    .toggle-icon {
      transition: transform var(--transition-quick);
      
      &.open { transform: rotate(180deg); }
    }
  }
  
  .section-content {
    padding: var(--space-lg);
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-lg);
    max-height: 0;
    overflow: hidden;
    transition: max-height var(--transition-base), opacity var(--transition-base);
    opacity: 0;
    
    &.open {
      max-height: 1000px;
      opacity: 1;
    }
  }
}
```

### Progressive Disclosure

**Basic View (Default)**

Show only the 5–8 most critical fields. Most users will never need more.

```css
.form-basic { /* Visible by default */ }
```

**Advanced View (Expandable)**

```css
.form-advanced {
  max-height: 0;
  overflow: hidden;
  opacity: 0;
  transition: max-height var(--transition-slow), opacity var(--transition-base);
  
  &.open {
    max-height: 2000px;
    opacity: 1;
  }
}

.show-advanced-toggle {
  display: block;
  font: var(--font-weight-medium) var(--font-size-sm) var(--font-family-ui);
  color: var(--color-primary);
  cursor: pointer;
  margin-top: var(--space-lg);
  
  &::before { content: '▼ '; }
  
  &.open::before { content: '▲ '; }
}
```

### Real-Time Validation

```css
.form-group {
  position: relative;
}

.form-group.validating .input {
  border-color: var(--color-warning);
}

.form-group.valid .input {
  border-color: var(--color-success);
}

.form-group.invalid .input {
  border-color: var(--color-error);
  background: var(--color-error-bg);
}

.validation-icon {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 16px;
  opacity: 0;
  transition: opacity var(--transition-quick);
  
  .valid & { color: var(--color-success); opacity: 1; }
  .invalid & { color: var(--color-error); opacity: 1; }
}
```

### Field-Specific Guidance

**Currency Input**

```css
.input-currency {
  text-align: right;
  font-family: var(--font-family-mono);
  
  &::before { content: '$'; position: absolute; left: 12px; }
  padding-left: 24px;
}
```

**Date Input**

```
Format: MMM DD, YYYY (Oct 15, 2025)
Display: Calendar icon on right
Keyboard support: YYYY-MM-DD when typing
Mobile: Native date picker
```

**Textarea for Long Content**

```css
.textarea {
  min-height: 80px;
  max-height: 400px;
  resize: vertical;
  
  &[data-limit] {
    position: relative;
    
    &::after {
      content: attr(data-chars-remaining);
      position: absolute;
      bottom: 8px;
      right: 12px;
      font-size: var(--font-size-xs);
      color: var(--color-text-tertiary);
    }
  }
}
```

**Dropdown with Search**

```css
.select-searchable {
  position: relative;
  
  .select-input {
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .search-field {
    padding: var(--space-md);
    border-bottom: 1px solid var(--color-border-light);
  }
  
  .options {
    max-height: 300px;
    overflow-y: auto;
    
    .option {
      padding: var(--space-md) var(--space-lg);
      cursor: pointer;
      transition: var(--transition-quick);
      
      &:hover { background: var(--color-bg-page); }
      &.selected { background: var(--color-primary-light); color: var(--color-primary); }
    }
  }
  
  .no-results {
    padding: var(--space-xl);
    text-align: center;
    color: var(--color-text-tertiary);
  }
}
```

---

## Responsive Design Strategy

### Breakpoints

```css
/* Mobile: 320–479px */
@media (max-width: 479px) { }

/* Tablet: 480–767px */
@media (min-width: 480px) and (max-width: 767px) { }

/* Desktop: 768–1023px */
@media (min-width: 768px) and (max-width: 1023px) { }

/* Large Desktop: 1024px+ */
@media (min-width: 1024px) { }

/* Extra Large: 1400px+ */
@media (min-width: 1400px) { }
```

### Mobile Design (< 480px)

```css
/* Sidebar: Hidden by default */
.sidebar { display: none; }

/* Header simplified */
.search-box { display: none; }

/* Single column layouts */
.form-two-column,
.list-two-column {
  grid-template-columns: 1fr;
}

/* Reduced padding */
.card,
.modal { padding: var(--space-md); }

/* Full-width modals */
.modal {
  width: 95%;
  max-height: 95vh;
}

/* Bottom navigation for core actions */
.mobile-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 56px;
  background: var(--color-bg-surface);
  border-top: 1px solid var(--color-border-light);
  display: flex;
  justify-content: space-around;
  align-items: center;
  z-index: 50;
  
  .nav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    font-size: var(--font-size-xs);
    cursor: pointer;
  }
}

/* Adjust body for bottom nav */
body { padding-bottom: 56px; }

/* Touch targets: minimum 44px × 44px */
button, a, [role="button"] {
  min-height: 44px;
  min-width: 44px;
}

/* Larger font for inputs (prevents auto-zoom on iOS) */
input, textarea, select {
  font-size: 16px !important;
}
```

### Tablet Design (480–767px)

```css
/* Sidebar collapsible but visible as icons */
.sidebar {
  width: 64px;
  padding: var(--space-md);
  
  .nav-label { display: none; }
  .nav-icon { font-size: 24px; }
}

.header { left: 64px; }

/* Two-column layouts work */
.form-two-column,
.list-two-column {
  grid-template-columns: 1fr 1fr;
}

/* Adjust spacing */
.card { padding: var(--space-md); }
.form-group { margin-bottom: var(--space-md); }

/* Table scrolls horizontally */
table { display: block; overflow-x: auto; }
```

### Desktop (768px+)

```css
.sidebar { width: var(--sidebar-width); }
.header { left: var(--sidebar-width); }

/* Full layouts */
.form-three-column { grid-template-columns: 1fr 1fr 1fr; }

/* Hover effects visible */
.card:hover { box-shadow: var(--shadow-md); }
```

### Large Desktop (1024px+)

```css
.sidebar { display: block; }

/* Two-panel layout for list/detail */
.list-detail-layout {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: var(--space-lg);
}

/* Generous spacing */
.container { padding: var(--space-2xl); }
```

---

## Interactions & Motion

### Timing & Easing

All animations use carefully tuned timing functions to feel responsive without being jarring.

**Easing Functions**

```css
/* Ease Out: Quick start, smooth deceleration (UI feedback) */
--easing-out: cubic-bezier(0, 0, 0.2, 1);

/* Ease In-Out: Smooth acceleration and deceleration (page transitions) */
--easing-in-out: cubic-bezier(0.4, 0, 0.2, 1);

/* Ease In: Subtle acceleration (entrance animations) */
--easing-in: cubic-bezier(0.4, 0, 1, 1);

/* Linear: For continuous motion (spinners, progress bars) */
--easing-linear: linear;
```

**Duration Reference**

```
Micro-interactions (hover, focus): 150ms
State changes (click, collapse): 200ms
Navigation (page transitions): 300ms
Full screen changes (modal open/close): 400ms
Loading states (spinners): 1000ms (full rotation)
```

### Hover Effects (Desktop Only)

Never use hover effects on mobile—they don't exist on touchscreens and create confusion.

```css
@media (hover: hover) {
  button:hover {
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
    transition: var(--transition-base);
  }
  
  .card:hover {
    border-color: var(--color-border-medium);
    box-shadow: var(--shadow-md);
  }
  
  a:hover {
    text-decoration: underline;
  }
  
  /* Icon hover opacity change */
  .icon-button:hover {
    opacity: 1;
  }
}

@media (hover: none) {
  /* Touch devices: No hover effects */
  button:hover { transform: none; box-shadow: none; }
}
```

### Focus States

Focus states must always be visible. Never remove focus indicators.

```css
/* Universal focus style */
*:focus-visible {
  outline: 3px solid var(--color-primary);
  outline-offset: 2px;
}

/* Custom focus for specific elements */
button:focus-visible,
a:focus-visible,
input:focus-visible {
  outline: 3px solid var(--color-primary);
  outline-offset: 2px;
}

/* Remove default outline and provide custom one */
button:focus { outline: none; }
button:focus-visible { outline: 3px solid var(--color-primary); }
```

### Click/Press Effects

```css
/* Button press animation */
button:active {
  transform: scale(0.97);
  transition: transform 100ms ease-in-out;
}

/* Checkbox/radio selection */
.checkbox:checked,
.radio:checked {
  animation: scaleIn 150ms ease-out;
}

@keyframes scaleIn {
  from { transform: scale(0.9); }
  to { transform: scale(1); }
}
```

### Loading States

**Spinner**

```css
.spinner {
  width: 24px;
  height: 24px;
  border: 3px solid var(--color-border-light);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Prevent motion on user preference */
@media (prefers-reduced-motion: reduce) {
  .spinner { animation: none; }
}
```

**Skeleton Screen**

```css
.skeleton {
  background: linear-gradient(
    90deg,
    var(--color-bg-page) 0%,
    var(--color-border-light) 50%,
    var(--color-bg-page) 100%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

**Progress Bar**

```css
.progress-bar {
  height: 3px;
  background: var(--color-border-light);
  border-radius: var(--radius-full);
  overflow: hidden;
  
  .fill {
    height: 100%;
    background: var(--color-primary);
    width: 0%;
    transition: width 0.3s ease;
  }
  
  &.indeterminate .fill {
    animation: indeterminate 1.5s infinite;
  }
}

@keyframes indeterminate {
  0% { width: 0%; }
  50% { width: 100%; }
  100% { width: 100%; }
}
```

### Page & Modal Transitions

**Fade In**

```css
.fade-in {
  opacity: 0;
  animation: fadeIn var(--transition-base) forwards;
}

@keyframes fadeIn {
  to { opacity: 1; }
}
```

**Slide Up (Modal Open)**

```css
.slide-up {
  transform: translateY(20px);
  opacity: 0;
  animation: slideUp var(--transition-base) forwards;
}

@keyframes slideUp {
  to {
    transform: translateY(0);
    opacity: 1;
  }
}
```

**List Item Entrance**

```css
.list-item {
  opacity: 0;
  animation: slideIn var(--transition-base) forwards;
  
  @for $i from 1 through 10 {
    &:nth-child(#{$i}) {
      animation-delay: $i * 50ms;
    }
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
```

### Respect Prefers-Reduced-Motion

Always honor user motion preferences. This is accessibility, not optional.

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

### Micro-Interactions

**Copy to Clipboard**

```javascript
function copyToClipboard(text) {
  navigator.clipboard.writeText(text);
  
  // Visual feedback
  const button = event.target;
  const originalText = button.textContent;
  button.textContent = '✓ Copied';
  
  setTimeout(() => {
    button.textContent = originalText;
  }, 2000);
}
```

**Delete Confirmation with Undo**

```javascript
function deleteItem(id) {
  const item = document.getElementById(id);
  item.classList.add('deleting');
  
  // Show undo option
  showNotification('Item deleted', {
    action: 'Undo',
    callback: () => undoDelete(id)
  });
  
  // Auto-delete after 8 seconds
  setTimeout(() => {
    if (item.classList.contains('deleting')) {
      item.remove();
    }
  }, 8000);
}
```

**Form Validation Success**

```css
.field.valid::after {
  content: '✓';
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-success);
  font-weight: bold;
  animation: popIn 300ms ease-out;
}

@keyframes popIn {
  0% {
    transform: translateY(-50%) scale(0);
  }
  50% {
    transform: translateY(-50%) scale(1.2);
  }
  100% {
    transform: translateY(-50%) scale(1);
  }
}
```

---

## Accessibility & Inclusive Design

### WCAG 2.1 AA Compliance

All components must meet WCAG 2.1 AA standards. AAA compliance is preferred where practical.

**Color Contrast Audit**

Every color combination is tested:

```
Dark Text (#1a1a1a) on White: 16:1 ✓ AAA
Secondary Text (#374151) on White: 11:1 ✓ AAA
Tertiary Text (#6b7280) on White: 7:1 ✓ AA
Primary Button Text on Teal: 5:1 ✓ AAA
Status Colors all: 5+:1 ✓ AAA
```

Use a contrast checker (e.g., WebAIM) before every color change.

### Keyboard Navigation

**Tab Order**

Maintain logical tab order matching reading order (left-to-right, top-to-bottom):

```html
<!-- Correct -->
<form>
  <input type="text" /> <!-- First tab -->
  <input type="email" /> <!-- Second tab -->
  <button>Submit</button> <!-- Third tab -->
</form>

<!-- Wrong: Skip links backward -->
<div>
  <button tabindex="10">Save</button>
  <input type="text" tabindex="1" />
</div>
```

**Keyboard Shortcuts**

```
Global:
  ? — Show keyboard help
  / — Focus search
  Esc — Close modal/menu
  
Document Actions:
  Ctrl+S (Cmd+S) — Save
  Ctrl+Enter — Submit/Approve
  Ctrl+Shift+D — Delete
  
Navigation:
  J — Next item
  K — Previous item
  H — Go home
  G+I — Go to inbox
  
Dark Mode:
  Ctrl+Shift+T — Toggle dark mode
```

### Screen Reader Support

**Semantic HTML**

```html
<!-- Use semantic elements, not divs -->
<button>Save</button>
<nav>Navigation</nav>
<main>Content</main>
<section>Section</section>
<h1>Page Title</h1>

<!-- ARIA attributes where needed -->
<button aria-label="Close dialog" onclick="closeDialog()">✕</button>
<div aria-live="polite">Status updates appear here</div>
<span aria-hidden="true">Decorative icon</span>
```

**Form Labels**

```html
<!-- Always associate labels with inputs -->
<label for="vendor-name">Vendor Name</label>
<input id="vendor-name" type="text" />

<!-- Or use aria-labelledby for complex cases -->
<div id="group-label">Billing Address</div>
<div aria-labelledby="group-label">
  <input type="text" placeholder="Street" />
  <input type="text" placeholder="City" />
</div>
```

**Error Announcements**

```html
<!-- Errors announced immediately -->
<input id="amount" type="number" aria-required="true" />
<div id="amount-error" role="alert" aria-live="assertive">
  Please enter a valid amount
</div>
```

**Lists and Tables**

```html
<!-- Proper table structure -->
<table>
  <thead>
    <tr>
      <th scope="col">Item Code</th>
      <th scope="col">Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>SKU-001</td>
      <td>Product Name</td>
    </tr>
  </tbody>
</table>

<!-- Skip navigation -->
<a href="#main-content" class="skip-link">Skip to main content</a>
```

### Color Accessibility

**Never use color alone to convey status:**

```css
/* ❌ Bad: Only color */
<span style="color: green;">Approved</span>

/* ✓ Good: Color + icon + text */
<span class="badge success">
  <span class="icon">✓</span>
  <span>Approved</span>
</span>
```

**Test with colorblind simulator:**

Use Chrome DevTools rendering options to simulate:
- Protanopia (red-blind)
- Deuteranopia (green-blind)
- Tritanopia (blue-blind)
- Achromatopsia (complete colorblindness)

### Motion & Vestibular

Respect user motion preferences:

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## Dark Mode & Themes

### Dark Mode Implementation

Dark mode is offered as an opt-in feature. Implement using CSS custom properties for easy theming.

```css
/* Light mode (default) */
:root {
  --color-text-primary: #1a1a1a;
  --color-bg-surface: #ffffff;
  --color-bg-page: #f5f6f8;
  /* ... other tokens */
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  :root {
    --color-text-primary: #f5f5f5;
    --color-text-secondary: #d1d1d1;
    --color-bg-surface: #1a1a1a;
    --color-bg-page: #0a0a0a;
    --color-border-medium: #2d2d2d;
    --color-primary-light: #1a4f4f;
    
    /* Status colors remain consistent */
    --color-success: #10b981;
    --color-error: #ef4444;
    --color-warning: #f59e0b;
  }
}

/* Manual dark mode toggle */
body.dark-mode {
  --color-text-primary: #f5f5f5;
  --color-bg-surface: #1a1a1a;
  /* ... */
}
```

**Contrast in Dark Mode**

Verify all status colors maintain sufficient contrast against dark backgrounds:

```
White text on #1a1a1a background: 16:1 ✓ AAA
Status colors on dark bg: All 5+:1 ✓ AAA
```

---

## Performance Standards

### Load Time Targets

| Metric | 4G | LTE | 3G |
|--------|-----|-----|-----|
| First Contentful Paint | < 1.5s | < 2s | < 4s |
| Largest Contentful Paint | < 2s | < 2.5s | < 5s |
| Time to Interactive | < 3.5s | < 4.5s | < 8s |

### Core Web Vitals Targets

- **LCP (Largest Contentful Paint)**: < 2.5s
- **FID (First Input Delay)**: < 100ms
- **CLS (Cumulative Layout Shift)**: < 0.1

### Bundle Size Limits

| Bundle | Size Limit | Gzipped |
|--------|-----------|---------|
| JS (core) | 300KB | 100KB |
| CSS | 100KB | 30KB |
| Total | 400KB | 130KB |

### Optimization Strategies

**Code Splitting**

```javascript
// Lazy load modules by route
const ProcurementModule = React.lazy(() => import('./modules/procurement'));
const SalesModule = React.lazy(() => import('./modules/sales'));

<Suspense fallback={<Loading />}>
  <Route path="/procurement" component={ProcurementModule} />
  <Route path="/sales" component={SalesModule} />
</Suspense>
```

**Image Optimization**

```html
<!-- Responsive images -->
<picture>
  <source srcset="image-1024.webp 1024w, image-512.webp 512w" type="image/webp" />
  <source srcset="image-1024.jpg 1024w, image-512.jpg 512w" type="image/jpeg" />
  <img src="image-512.jpg" alt="Description" />
</picture>

<!-- Lazy load below the fold -->
<img src="placeholder.svg" data-src="image.webp" loading="lazy" alt="" />
```

**Network Optimization**

```
- Enable gzip compression (server-side)
- Use CDN for static assets
- Implement service worker for offline support
- Cache API responses (5–30 min TTL)
- Prefetch critical routes
```

---

## Implementation & Quality Assurance

### Component Documentation Template

Every component must document:

```markdown
## [Component Name]

**Purpose:** Brief explanation of when/why to use this component

**Variants:** List all visual variations

**States:** Default, hover, active, disabled, loading, error

**Props/Attributes:** All configurable options with types

**Accessibility:** ARIA attributes, keyboard support, screen reader testing

**Usage Example:**
\`\`\`html
<component variant="primary" state="active">Label</component>
\`\`\`

**Common Mistakes:** Common implementation errors to avoid

**Mobile Considerations:** Any responsive adjustments needed
```

### Testing Checklist

**Visual Regression**
- [ ] Screenshot comparisons at all breakpoints
- [ ] All component states tested
- [ ] Dark mode rendering verified
- [ ] Cross-browser compatibility (Chrome, Firefox, Safari, Edge)

**Accessibility**
- [ ] Keyboard navigation works without mouse
- [ ] All focus states visible
- [ ] Color contrast meets AA (5.5:1 text)
- [ ] Screen reader tested (NVDA, JAWS, VoiceOver)
- [ ] Forms properly labeled
- [ ] Error messages announced
- [ ] Zoom tested to 200%

**Performance**
- [ ] Lighthouse score 90+
- [ ] Core Web Vitals passing
- [ ] Bundle size within limits
- [ ] Load time < 2.5s on 4G
- [ ] No layout shift on interaction

**Functionality**
- [ ] All interactive elements respond
- [ ] Form validation working
- [ ] Error handling graceful
- [ ] Success states clear
- [ ] Undo/revision history works

### Code Standards

**CSS Organization**

```css
/* 1. Reset & Variables */
:root { /* tokens */ }

/* 2. Base Styles */
html, body { }
h1, h2, h3 { }
button, input { }

/* 3. Layout */
.container { }
.sidebar { }

/* 4. Components */
.button { }
.card { }
.modal { }

/* 5. Utilities */
.text-center { }
.mt-lg { }

/* 6. Responsive */
@media (max-width: 767px) { }

/* 7. Dark Mode */
@media (prefers-color-scheme: dark) { }

/* 8. Animation */
@keyframes { }
```

**JavaScript Patterns**

```javascript
// Use event delegation for dynamic content
document.addEventListener('click', (e) => {
  if (e.target.matches('[data-action="approve"]')) {
    approveDocument(e.target.dataset.documentId);
  }
});

// Debounce expensive operations
const debouncedSearch = debounce((query) => {
  filterList(query);
}, 300);

// Use data attributes for behavior
<button data-action="delete" data-id="123">Delete</button>

// Always provide fallback for errors
try {
  submitForm();
} catch (error) {
  showError('Failed to submit. Please try again.');
}
```

### Design System Maintenance

**Update Process**

1. Propose change with rationale
2. Review with design + engineering leads
3. Update documentation
4. Audit existing usage
5. Plan migration if breaking change
6. Update all affected components
7. Deploy and monitor

**Versioning**

- Patch (1.0.x): Bug fixes, minor adjustments
- Minor (1.x.0): New components, non-breaking changes
- Major (x.0.0): Breaking changes, significant redesigns

**Review Cadence**

- Monthly: Audit new issues, collect feedback
- Quarterly: Major theme/token review
- Annually: Full design system audit

---

## Final Checklist

Before shipping any feature:

- [ ] Follows design system tokens
- [ ] Components used from library (not custom)
- [ ] Responsive at all breakpoints
- [ ] WCAG AA accessibility verified
- [ ] Performance standards met
- [ ] Dark mode compatible
- [ ] Documented and tested
- [ ] No breaking changes without migration plan
- [ ] Approved by design system team

---

**Version History**

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | October 2025 | Complete rewrite: improved clarity, added dark mode, performance standards, detailed implementation guidance |
| 1.0 | October 2025 | Initial specification |

**Document Owner:** Design & Engineering Teams  
**Last Updated:** October 2025  
**Next Review:** January 2026
