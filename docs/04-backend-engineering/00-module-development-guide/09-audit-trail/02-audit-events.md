---
title: Audit Events Reference
portal: 4 — Backend Engineering
section: 00-module-development-guide/09-audit-trail
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-audit-overview.md
    title: Audit Trail Overview
---

# Audit Events Reference

Complete list of audit events for the contracts module with field values.

## contract.create

```go
audit.Event{
	TenantID:     contract.TenantID,
	EntityID:     contract.EntityID,
	ActorID:      contract.CreatedBy,
	Action:       "contract.create",
	ResourceType: "contract",
	ResourceID:   contract.ID.String(),
	Before:       nil,
	After:        contract,
	Metadata: map[string]string{
		"contract_number": contract.ContractNumber,
		"status":          string(contract.Status),
	},
	OccurredAt: time.Now(),
}
```

## contract.update

```go
audit.Event{
	TenantID:     updated.TenantID,
	EntityID:     updated.EntityID,
	ActorID:      req.UpdatedBy,
	Action:       "contract.update",
	ResourceType: "contract",
	ResourceID:   updated.ID.String(),
	Before:       currentContract,   // state before update
	After:        updated,           // state after update
	OccurredAt:   time.Now(),
}
```

## contract.submit / contract.approve / contract.activate / contract.suspend / contract.terminate

Pattern is identical for all status transitions — only `Action` and the relevant metadata fields differ:

```go
audit.Event{
	TenantID:     updated.TenantID,
	EntityID:     updated.EntityID,
	ActorID:      actorID,
	Action:       "contract.submit",   // or .approve, .activate, etc.
	ResourceType: "contract",
	ResourceID:   updated.ID.String(),
	Before:       currentContract,
	After:        updated,
	Metadata: map[string]string{
		"from_status": string(currentContract.Status),
		"to_status":   string(updated.Status),
	},
	OccurredAt: time.Now(),
}
```

## contract.delete

```go
audit.Event{
	TenantID:     contract.TenantID,
	EntityID:     contract.EntityID,
	ActorID:      deletedBy,
	Action:       "contract.delete",
	ResourceType: "contract",
	ResourceID:   contract.ID.String(),
	Before:       contract,   // final state before deletion
	After:        nil,
	OccurredAt:   time.Now(),
}
```

## contract_line.create / contract_line.update / contract_line.delete

```go
audit.Event{
	TenantID:     line.TenantID,
	EntityID:     contract.EntityID,   // parent contract's entity
	ActorID:      actorID,
	Action:       "contract_line.create",
	ResourceType: "contract_line",
	ResourceID:   line.ID.String(),
	Before:       nil,
	After:        line,
	Metadata: map[string]string{
		"contract_id":     line.ContractID.String(),
		"contract_number": contract.ContractNumber,
		"line_number":     strconv.Itoa(line.LineNumber),
	},
	OccurredAt: time.Now(),
}
```

## Serialisation Note

`Before` and `After` are `any`. The audit service JSON-serialises them. Ensure domain entities are JSON-serialisable (all exported fields, no unexported value objects). If a field type is not JSON-serialisable (e.g., `decimal.Decimal`), use the `json:` tag to control serialisation:

```go
type Contract struct {
	// ...
	TotalValue ContractValue `json:"-"`           // exclude — use string form instead
	TotalValueStr string     `json:"total_value"` // serialise as string
}
```

Or provide a custom `MarshalJSON` method on the domain entity.
