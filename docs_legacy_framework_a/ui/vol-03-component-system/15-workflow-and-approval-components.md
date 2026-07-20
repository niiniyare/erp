> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
volume: "03 — Component System"
chapter: 15
title: "Workflow and Approval Components"
audience: "Backend engineers building document screens with approval steps"
prerequisites:
  - "Vol 01 Ch 3 — Session context and permissions"
  - "Ch 12 — Screen composition"
  - "Ch 16 — ERP-Specific Components (for DocumentHeaderBlock)"
section: "vol-03-component-system"
related:
  - "[Ch 13 — Dashboard Framework](13-dashboard-framework.md)"
  - "[Ch 16 — ERP-Specific Components](16-erp-specific-components.md)"
---

# Chapter 15 — Workflow and Approval Components

## Table of Contents

1. [Approval pattern overview](#151-approval-pattern-overview)
2. [ApprovalWorkflowBlock](#152-approvalworkflowblock)
3. [Permission gating](#153-permission-gating)
4. [AMIS expression syntax for conditional controls](#154-amis-expression-syntax-for-conditional-controls)
5. [DisabledOn vs VisibleOn](#155-disabledon-vs-visibleon)
6. [Approval in the context of InvoiceScreen](#156-approval-in-the-context-of-invoicescreen)
7. [ActionNode for approve and reject buttons](#157-actionnode-for-approve-and-reject-buttons)
8. [Workflow state and UI state](#158-workflow-state-and-ui-state)

---

## 15.1 Approval pattern overview

Every ERP document that can be authorised — invoices, bills, purchase orders, journal entries, expense claims — follows the same approval pattern:

1. The page's `InitAPI` loads the document record and includes `can_approve` in the response.
2. `ApprovalWorkflowBlock` is added to the document body.
3. The block uses `can_approve` from page scope to enable or disable approval controls via AMIS expressions.
4. The session feature flag `approval_workflow` determines whether the block renders at all.

The approval section is always the last substantive block in a document form, placed after line items and totals but before optional attachment and notes sections.

```
DocumentHeaderBlock
PartyBlock
ProductServiceLineBlock
TaxSummaryBlock
TotalsSummaryBlock
ApprovalWorkflowBlock   ← last substantive block
[AttachmentsBlock]      ← optional
[InternalNotesBlock]    ← optional
```

---

## 15.2 ApprovalWorkflowBlock

`ApprovalWorkflowBlock` is defined in `internal/web/dsl/blocks/approval.go`. Its signature:

```go
func ApprovalWorkflowBlock(sess ui.UISessionContext) ast.Node
```

It accepts only the session — no config struct. All behaviour is derived from the session's feature flags and from the `can_approve` variable that the page's `InitAPI` places into page scope.

### Feature flag gate

If the tenant does not have `approval_workflow` enabled, the block renders a collapsed, read-only placeholder section instead of the full approval UI:

```go
if !sess.Flag("approval_workflow") {
    return ast.SectionNode{
        Title:     "Approval",
        Collapsed: true,
        Body: []ast.Node{
            ast.InputTextNode{
                Name:       "_approval_placeholder",
                Label:      "Approval workflow not enabled",
                DisabledOn: "true",
            },
        },
    }
}
```

This means the section is always present in the schema (useful for consistent layout), but visually collapsed and inert for tenants without the feature.

### Full approval section

When the feature is enabled, the block renders:

```go
ast.SectionNode{
    Title: "Approval",
    Body: []ast.Node{
        ast.SelectNode{
            Name:  "approval_status",
            Label: "Status",
            Options: []ast.SelectOption{
                {Label: "Pending",  Value: "pending"},
                {Label: "Approved", Value: "approved"},
                {Label: "Rejected", Value: "rejected"},
            },
            DisabledOn: "${!can_approve}",
        },
        ast.InputTextNode{
            Name:       "approval_note",
            Label:      "Note",
            DisabledOn: "${!can_approve}",
        },
    },
}
```

The status dropdown and the note field are both disabled when `can_approve` is falsy in page scope. When `can_approve` is true (i.e., the current user has approval rights for this document), both controls become interactive.

---

## 15.3 Permission gating

The `can_approve` variable originates from the backend, not from a client-side permission check. The `InitAPI` response must include it:

```json
{
  "id": "inv-001",
  "status": "pending",
  "can_approve": true,
  "line_items": [...],
  ...
}
```

The backend derives `can_approve` by evaluating the session's permission against the document, which may take into account:

- Whether the user holds the `approve` action on `finance.invoices`
- Whether the document belongs to an entity within the user's entity scope
- Whether the document is in an approvable state (e.g., `status == "pending"`)

**The block does not check `sess.Can(...)` for approval.** It only consumes `can_approve` from page scope. The reason is that approval eligibility is context-dependent (document state, entity scope) and cannot be determined purely from the session's static permission set.

```go
// Correct — block consumes can_approve from InitAPI response
blocks.ApprovalWorkflowBlock(sess)

// Wrong — never pre-check approval permission at screen level
if sess.Can("approve", "finance.invoices") {
    blocks.ApprovalWorkflowBlock(sess)
}
```

---

## 15.4 AMIS expression syntax for conditional controls

The `DisabledOn` field in `ApprovalWorkflowBlock` uses:

```
"${!can_approve}"
```

This is the correct AMIS expression syntax for negation. Do **not** write:

```
"!${can_approve}"   ← WRONG
```

### Why "!${can_approve}" is a bug

`"!${can_approve}"` is evaluated by the AMIS expression engine as a JavaScript template string first. When `can_approve` is `false`, the AMIS substitution produces the string `"!false"`. The AMIS `disabledOn` condition then evaluates `"!false"` as a truthy non-empty string, which always disables the control — even when the user has approval rights.

When `can_approve` is `true`, it produces `"!true"`, also a truthy non-empty string. The control is permanently disabled regardless of the actual value.

The correct form places the negation operator **inside** the AMIS expression:

```
"${!can_approve}"
```

Here the AMIS engine evaluates the expression `!can_approve` in JavaScript context, producing the boolean `false` when the user can approve (enabling the control) and `true` when they cannot (disabling it).

### General rule

All logic must be inside the `${}` brackets. The `${}` is a delimiter for a complete JavaScript expression, not a string interpolation point that you can wrap with operators.

| Correct | Incorrect |
|---|---|
| `"${!can_approve}"` | `"!${can_approve}"` |
| `"${status === 'draft'}"` | `"${status} === 'draft'"` |
| `"${qty && unit_price}"` | `"${qty} && ${unit_price}"` |
| `"${total > 0 ? 'red' : 'green'}"` | n/a (ternary must be inside) |

---

## 15.5 DisabledOn vs VisibleOn

AMIS provides two mechanisms for conditional rendering:

| Mechanism | Effect | When to use |
|---|---|---|
| `DisabledOn` | Control is visible but greyed-out and non-interactive | User should see the field exists but cannot edit it |
| `VisibleOn` | Control is completely absent from the DOM | Control is irrelevant and its presence would confuse the user |

### Approval section guidance

**Use `DisabledOn` for approval controls.** The approval status field and note field are always visible — non-approving users see the current approval state. Hiding them entirely would mean users cannot tell whether a document is pending or already approved.

**Use `VisibleOn` for structurally conditional blocks.** For example, a shipping address block is hidden when the document type is a service invoice (no physical delivery). In that case the field is not just inactive — it genuinely does not apply.

```go
// Approval controls: disable, don't hide
ast.SelectNode{
    Name:       "approval_status",
    DisabledOn: "${!can_approve}",   // visible but inactive for non-approvers
}

// Optional feature block: hide entirely when not applicable
ast.SectionNode{
    Title:     "Shipping Address",
    VisibleOn: "${document_type === 'physical'}",
}
```

### DisabledOn = "true"

The literal string `"true"` as a `DisabledOn` value unconditionally disables a control. This is used for read-only display fields, such as the computed subtotal column in line items:

```go
ast.InputNumberNode{
    Name:       "subtotal",
    Label:      "Subtotal",
    Precision:  2,
    DisabledOn: "true",  // always read-only
}
```

---

## 15.6 Approval in the context of InvoiceScreen

`InvoiceScreen` in `internal/web/dsl/screens/invoice.go` shows exactly how `ApprovalWorkflowBlock` fits into a complete document:

```go
func InvoiceScreen(sess ui.UISessionContext, cfg InvoiceScreenConfig) ast.Node {
    lineCfg := blocks.DefaultLineItemConfig()
    lineCfg.ShowDiscount = !cfg.IsPurchase
    lineCfg.ShowTaxRate  = true
    lineCfg.ReadOnly     = cfg.ReadOnly

    body := []ast.Node{
        blocks.DocumentHeaderBlock(sess, blocks.DocumentHeaderConfig{...}),
        blocks.PartyBlock(sess, partyCfg),
        blocks.AddressBlock(sess, blocks.AddressConfig{...}),
        blocks.ProductServiceLineBlock(sess, lineCfg),
        blocks.TaxSummaryBlock(sess),
        blocks.TotalsSummaryBlock(sess),
        blocks.ApprovalWorkflowBlock(sess),  // ← always last substantive block
    }
    if cfg.ShowPaymentTerms  { body = append(body, blocks.PaymentTermsBlock(sess, ...)) }
    if cfg.ShowAttachments   { body = append(body, blocks.AttachmentsBlock(sess)) }
    if cfg.ShowInternalNotes { body = append(body, blocks.InternalNotesBlock(sess)) }

    return ast.PageNode{
        Title: pageTitle(cfg.IsPurchase, "Invoice", "Bill"),
        InitAPI: &ast.APISpec{
            Method: "get",
            URL:    "/api/v1/finance/" + docType + "/:id",
            SendOn: "${params.id}",
            // Backend must return: can_approve, tenant_currency, totals[], tax_lines[]
        },
        Body: body,
    }
}
```

Key points from this implementation:

1. `ApprovalWorkflowBlock(sess)` is called unconditionally — the block itself decides whether to render approval controls or the feature-disabled placeholder.
2. The `InitAPI` comment documents that the backend is responsible for returning `can_approve` in the response. This is a contract between the screen and the backend handler.
3. `SendOn: "${params.id}"` skips the `InitAPI` call on the new-document route (where there is no `:id` param). This means `can_approve` will be absent from page scope on new documents — the block handles this gracefully because `!can_approve` evaluates to `true` (disabled) when the variable is undefined.

---

## 15.7 ActionNode for approve and reject buttons

The `ApprovalWorkflowBlock` as currently implemented uses a `SelectNode` for the approval status rather than dedicated approve/reject `ActionNode` buttons. This is a deliberate simplicity choice — one field rather than two buttons reduces layout complexity.

When explicit approve and reject buttons are needed (for example in a dedicated approval modal), use `ActionNode` with `ActionType: "ajax"` and a confirm dialog:

```go
// Approve button with confirmation dialog
ast.ActionNode{
    Label:      "Approve",
    ActionType: "ajax",
    Level:      "success",
    DisabledOn: "${!can_approve}",
    ConfirmText: "Approve this document? This action cannot be undone.",
    API: &ast.APISpec{
        Method: "post",
        URL:    "/api/v1/finance/invoices/${id}/approve",
    },
}

// Reject button with a dialog for rejection reason
ast.ActionNode{
    Label:      "Reject",
    ActionType: "dialog",
    Level:      "danger",
    DisabledOn: "${!can_approve}",
    Dialog: &ast.DialogNode{
        Title: "Reject Document",
        Body: []ast.Node{
            ast.InputTextNode{Name: "rejection_reason", Label: "Reason", Required: true},
        },
        Actions: []ast.ActionNode{
            {
                Label:      "Confirm Rejection",
                ActionType: "ajax",
                Level:      "danger",
                API: &ast.APISpec{
                    Method: "post",
                    URL:    "/api/v1/finance/invoices/${id}/reject",
                },
            },
        },
    },
}
```

Key constraints when using `ActionNode`:

- `ActionType: "dialog"` requires the `Dialog *ast.DialogNode` field to be set. Omitting it causes a ValidateStage panic.
- `ActionType: "drawer"` requires the `Drawer *ast.DrawerNode` field to be set.
- `ActionType: "ajax"` requires the `API *ast.APISpec` field to be set.
- `DisabledOn: "${!can_approve}"` is the same expression used in the select-based approach.

---

## 15.8 Workflow state and UI state

The `approval_status` field in the block and the `can_approve` variable from the API represent UI state. They are not the authoritative source of workflow state.

The authoritative workflow state lives in a Temporal workflow. When an invoice moves to "Approved" status, a Temporal workflow activity records the transition, emits an audit event, and potentially triggers downstream actions (e.g., releasing payment).

The relationship is:

```
Temporal workflow state (authoritative)
    ↓ reflected in API response
Backend handler returns can_approve + approval_status
    ↓ loaded by InitAPI
AMIS page scope holds can_approve
    ↓ read by ApprovalWorkflowBlock
DisabledOn="${!can_approve}" enables or disables controls
```

The UI does not drive the workflow — it reports to it. When a user sets `approval_status` to "Approved" and submits the form, the backend handler:
1. Validates that the user's session actually permits this action.
2. Signals or updates the Temporal workflow.
3. Persists the state change.
4. Returns the updated `can_approve` and `approval_status` in the response.

The AMIS page reloads its `InitAPI` after a successful form submit, which refreshes `can_approve` and re-evaluates the `DisabledOn` expressions.

> **Status: Planned — Not Yet Implemented**
> Direct Temporal workflow signal integration from approval actions is designed in the platform architecture but not yet wired to the approval block. Currently, approval status is persisted as a database field without Temporal workflow lifecycle management. Temporal integration for approval state machines is planned for a future milestone.

---

### Approval component checklist

- [ ] `ApprovalWorkflowBlock(sess)` called unconditionally — no outer permission guard
- [ ] Backend `InitAPI` response includes `can_approve` boolean
- [ ] `DisabledOn` uses `"${!can_approve}"` — not `"!${can_approve}"`
- [ ] Feature flag `approval_workflow` is enabled on the tenant for the full block to render
- [ ] `ApprovalWorkflowBlock` is the last substantive block before optional attachments/notes
- [ ] Explicit `ActionNode` approve/reject buttons use `DisabledOn: "${!can_approve}"` and include `ConfirmText` or a `Dialog` node
- [ ] Ajax action nodes have `API *ast.APISpec` set; dialog action nodes have `Dialog *ast.DialogNode` set
