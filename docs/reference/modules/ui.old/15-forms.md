[<-- Back to Index](README.md)

## Forms

### The Golden Rule

**Only show a field when the user needs to fill it in right now.**

A form with 8 visible fields that reveals 4 more when needed feels simpler than a form with 12 always-visible fields — even though the total field count is the same.

### Field Types and When to Use Them

| Field Type | Use When | Never Use When |
|---|---|---|
| `input-text` | Free text, names, descriptions | Selecting from a known set |
| `input-number` | Quantities, amounts, percentages | Codes that happen to be numeric |
| `select` | 5–15 mutually exclusive options | Fewer than 5 (use `radios`) |
| `radios` | 2–5 mutually exclusive options, all visible | More than 5 (use `select`) |
| `checkboxes` | Multiple selections from a known set | Single selection |
| `switch` | Binary on/off | Anything with meaningful intermediate states |
| `input-date` | Specific date selection | Relative dates ("last 30 days") |
| `input-date-range` | Period selection (payroll period, leave period) | Single-date selection |
| `select` (searchable) | Large collections (employees, GL accounts, countries) | Small sets where a dropdown is clearer |
| `input-file` | Document attachments, evidence | Anything that can be entered as structured data |
| `rich-text` | Long-form content (policies, descriptions) | Short text — use `input-text` |

### Layout Principles

**Single column is almost always right.** Two-column layouts cause users to skip fields, create ambiguous reading order, and break on mobile. Exception: related field pairs (start date / end date, min/max salary) can sit on the same row.

**Labels above inputs, not to the left.** Top labels work on all screen sizes and with any label length.

**Placeholder text is not a label.** Placeholder text disappears when the user starts typing. Always use a visible label. Use placeholder text only for example values.

**Mark optional fields as "optional".** Most fields should be required. Marking the rare optional ones is less noise than asterisking required fields.

### Conditional Fields

Use `visibleOn` to show fields only when relevant:

```json
[
  {
    "type":     "select",
    "name":     "customer_type",
    "label":    "Customer Type",
    "options":  [{"label":"Individual","value":"individual"},{"label":"Business","value":"business"}],
    "required": true,
    "value":    "individual"
  },
  {
    "type":      "input-text",
    "name":      "tax_id",
    "label":     "Tax ID / KRA PIN",
    "visibleOn": "${customer_type === 'business'}",
    "required":  true
  }
]
```

`clearValueOnHidden: true` clears the field value when it becomes hidden — prevents stale data being submitted.

### Validation

**Validate on blur (when user leaves a field), not only on submit:**

```json
{
  "type":     "input-number",
  "name":     "basic_salary",
  "label":    "Basic Salary",
  "required": true,
  "min":      16200,
  "validations": { "min": 16200 },
  "validationErrors": {
    "min": "Salary is below the minimum wage for Kenya (KES 16,200)."
  }
}
```

**Error messages must explain AND guide:**
```markdown
❌ "Invalid input"
✅ "The salary you entered (KES 8,000) is below the minimum wage for Kenya (KES 16,200). Enter KES 16,200 or more."
```

**Block on errors, warn on advisories.** Use red for errors (prevent save), amber for warnings (allow save):

```json
{
  "type":    "alert",
  "level":   "warning",
  "visibleOn": "${net_pay < gross_pay * 0.6}",
  "body":    "Net pay is less than 60% of gross. Verify deductions are correct before proceeding."
}
```

### Smart Defaults

Pre-fill every field with the best default. The user should only change fields that differ from the typical case:

```json
{
  "type":  "input-date",
  "name":  "period_start",
  "label": "Period Start",
  "value": "${DATETOSTR(STARTOF(NOW(), 'month'), 'YYYY-MM-DD')}"
}
```

Payroll run examples:
- Period defaults to current calendar month
- Currency defaults to tenant's base currency
- Pay date defaults to last working day of the period
- "All active employees" defaults to checked

### Multi-Step Forms (Wizards)

For complex sequential processes — running payroll, onboarding a new employee:

```markdown
Rules for wizards:
✅ Show all steps upfront so the user knows how long the process is
✅ Mark completed steps visually (checkmark, filled dot)
✅ Allow going back to any completed step without losing data
✅ Auto-save state between steps (session expiry → return to where they left off)
✅ Final step is always a review screen with a summary of all entered data
✅ Show a loading state on submit — payroll runs take seconds

Step indicator pattern:
  Step 1     Step 2     Step 3     Step 4     Step 5
    ●─────────●─────────○─────────○─────────○
  Period   Employees  Review    Approve   Post
  (done)   (active)   (next)
```

### Grouping Fields

Use `fieldset` or section titles to group related fields. Use titles, not card borders:

```json
{
  "type":  "fieldset",
  "title": "Contract Details",
  "body": [
    { "type": "select",       "name": "employment_type", "label": "Employment Type" },
    { "type": "input-date",   "name": "start_date",      "label": "Start Date" },
    { "type": "input-number", "name": "basic_salary",    "label": "Basic Salary" }
  ]
}
```

---
