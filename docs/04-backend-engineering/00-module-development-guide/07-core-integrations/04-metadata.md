---
title: Metadata Integration
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-tenant-integration.md
    title: Tenant Integration
  - path: ./05-audit-integration.md
    title: Audit Integration
---

# Metadata Integration

The metadata module provides entity hierarchy management and dynamic attributes. Business modules interact with it to get entity information and to attach domain-specific attributes to entities.

## entity_id on Every Record

Every business record carries an `entity_id` that references the organisational entity (branch, department, subsidiary) it belongs to. This is Layer 2 of tenant isolation.

```go
// On create: entity_id comes from session scope
EntityID: uuid.MustParse(sess.EntityScope.EntityID)

// On list: filter by entity_id based on scope
if sess.EntityScope.Type != iam.EntityScopeAll {
	params.EntityID = &entityID
}
```

## Getting Entity Display Name

When a list response needs to show the entity name alongside contract records, the handler fetches entity data from the entity service after the list query:

```go
// handler — enrich list response with entity names
contracts, err := h.service.List(ctx, params, principal)
// ...

// Collect unique entity IDs
entityIDs := extractEntityIDs(contracts.Items)

// Fetch entity names in one batch call
entities, err := h.entityService.GetByIDs(ctx, entityIDs, sess.TenantID)
// ...

// Build response with entity names
resp := mapContractsToResponse(contracts.Items, entities)
```

Never join across module boundaries in SQL — fetch IDs in the contracts query, then fetch display data from the entity service separately.

## Dynamic Attributes

The metadata module supports tenant-defined attributes on entities. If the contracts module allows tenants to add custom fields to contracts (e.g., a "project code" field), use the attribute system:

```go
// Reading dynamic attributes for a contract
attrs, err := h.metadataService.GetAttributes(ctx, metadata.GetAttributesRequest{
	TenantID:     tenantID,
	ResourceType: "contract",
	ResourceID:   contractID.String(),
})
```

Custom attributes are rendered in the UI via the attributes panel — the schema does not need to hardcode them. The metadata module handles storage and retrieval.

## Entity State Checks

Some modules need to check whether an entity is active before creating records within it:

```go
// Verify the entity is active before creating a contract within it
entity, err := h.entityService.GetByID(ctx, entityID, tenantID)
if err != nil || entity.Status != "active" {
	return fiber.NewError(fiber.StatusBadRequest, "cannot create contract: entity is not active")
}
```

This check belongs in the handler or service layer — not in the repository.
