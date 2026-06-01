---
title: Querying the Audit Trail
portal: 4 — Backend Engineering
section: 00-module-development-guide/09-audit-trail
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-audit-overview.md
    title: Audit Trail Overview
---

# Querying the Audit Trail

The audit trail is stored in the platform `audit_log` table and queried via the audit service. Business modules expose audit history in their detail views.

## Fetching Audit History for a Resource

```go
// In the handler, to fetch audit history for a contract:
history, err := h.auditSvc.ListByResource(c.Context(), audit.ListByResourceParams{
	TenantID:     sess.TenantID,
	ResourceType: "contract",
	ResourceID:   contractID.String(),
	PageSize:     20,
	PageOffset:   0,
})
```

## Audit History in Response DTO

Include audit history in the contract detail response:

```go
type ContractDetailResponse struct {
	Contract ContractResponse      `json:"contract"`
	Lines    []ContractLineResponse `json:"lines"`
	History  []AuditEntryResponse  `json:"history"`
}

type AuditEntryResponse struct {
	ID         string            `json:"id"`
	Action     string            `json:"action"`
	ActorID    string            `json:"actor_id"`
	ActorName  string            `json:"actor_name"`  // enriched from user service
	OccurredAt string            `json:"occurred_at"` // ISO 8601
	Before     json.RawMessage   `json:"before,omitempty"`
	After      json.RawMessage   `json:"after,omitempty"`
}
```

## amis Audit Panel

The amis schema for the contract detail view includes an audit history panel:

```json
{
	"type": "panel",
	"title": "Change History",
	"body": {
		"type": "crud",
		"api": "/api/v1/contracts/${id}/history",
		"columns": [
			{"name": "occurred_at", "label": "When", "type": "datetime"},
			{"name": "actor_name",  "label": "By",   "type": "text"},
			{"name": "action",      "label": "Action", "type": "mapping",
			 "map": {
				"contract.create":    "Created",
				"contract.submit":    "Submitted",
				"contract.approve":   "Approved",
				"contract.activate":  "Activated",
				"contract.terminate": "Terminated"
			 }
			}
		]
	}
}
```

## Audit Route

Register the audit history endpoint in `routes.go`:

```go
// GET /api/v1/contracts/:id/history
contracts.Get("/:id/history",
	deps.AuthorizeMiddleware("contracts.contract.read"),
	h.GetHistory,
)
```

```go
// handler method
func (h *Handler) GetHistory(c *fiber.Ctx) error {
	contractID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid contract id")
	}
	sess := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)

	history, err := h.auditSvc.ListByResource(c.Context(), audit.ListByResourceParams{
		TenantID:     sess.TenantID,
		ResourceType: "contract",
		ResourceID:   contractID.String(),
		PageSize:     parsePageSize(c, 50),
		PageOffset:   parsePageOffset(c),
	})
	if err != nil {
		return mapError(err)
	}

	return c.JSON(mapHistoryToResponse(history))
}
```
