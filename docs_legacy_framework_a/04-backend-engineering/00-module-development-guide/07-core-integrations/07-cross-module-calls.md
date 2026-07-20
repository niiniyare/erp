> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Cross-Module Calls
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-tenant-integration.md
    title: Tenant Integration
  - path: ../02-ddd-domain-design/01-bounded-context.md
    title: Bounded Context
---

# Cross-Module Calls

When the contracts module needs data from another module (vendor details, entity names, user display names), it calls that module's service interface. It never queries the foreign module's database tables directly.

## The Rule

> Import the module facade (`awo.so/internal/core/<module>`). Call its service interface. Never import `awo.so/internal/core/<module>/domain` or `awo.so/internal/core/<module>/repository`.

```go
// CORRECT — import the facade
import vendor "awo.so/internal/core/vendors"

// handler receives vendor service injected via Wire
vendorSvc vendor.VendorService

// Call through the interface
v, err := h.vendorSvc.GetByID(ctx, contractVendorID, tenantID)

// WRONG — importing sub-packages directly
import "awo.so/internal/core/vendors/repository"  // never do this
import "awo.so/internal/core/vendors/domain"       // only if truly needed for type reference
```

## Pattern: Enrich Response in Handler

Cross-module data is fetched in the handler layer when assembling response DTOs, not inside the service or repository.

```go
// handlers/handler.go — List handler
func (h *Handler) List(c *fiber.Ctx) error {
	sess := c.Locals(iam.LocalsKeySession).(*iam.ResolvedSession)

	// 1. Get contracts from contracts service
	result, err := h.service.List(ctx, params, principal)
	if err != nil {
		return mapError(err)
	}

	// 2. Enrich with vendor names (cross-module call)
	vendorIDs := extractVendorIDs(result.Items)
	vendors, err := h.vendorSvc.GetByIDs(ctx, vendorIDs, sess.TenantID)
	if err != nil {
		// Non-critical: log and continue with empty vendor names
		h.logger.Warn().Err(err).Msg("failed to fetch vendor names for contract list")
		vendors = map[uuid.UUID]string{}
	}

	// 3. Map to response DTO with enrichment
	return c.JSON(mapListToResponse(result, vendors))
}
```

Enrichment failures (vendor service down) are non-fatal for read operations. The response degrades gracefully — contract data is still returned, vendor name is empty or shows the UUID.

## Avoiding Circular Imports

Modules must not import each other's sub-packages in cycles. The dependency graph should be a DAG (directed acyclic graph):

```
contracts → vendor module (OK)
vendor module → contracts (NOT OK — creates cycle)
```

If two modules need each other's data, extract shared types to a `shared` package or use event-driven integration (one module publishes events, the other subscribes) to break the cycle.

## Service Interface Injection (Wire)

Cross-module service dependencies are injected via Wire:

```go
// internal/core/contracts/service/contract.go
type contractService struct {
	// ...
	vendorSvc vendor.VendorService  // injected cross-module dependency
}

func NewContractService(
	// ...
	vendorSvc vendor.VendorService,
) ContractService {
```

```go
// Module facade wire set
var ContractsServiceSet = wire.NewSet(
	service.NewContractService,
	// vendorSvc is provided by the VendorModule's wire set elsewhere in the graph
)
```

## When NOT to Call Cross-Module Services

**For pure display enrichment**, pass IDs in the response and let the frontend fetch display data. This is simpler and avoids coupling:

```json
{
  "id": "...",
  "vendor_id": "...",
  "vendor_name": null  // frontend fetches separately if needed
}
```

**For high-volume list queries**, batch cross-module calls (`GetByIDs`) rather than one call per record. A list of 50 contracts should trigger one vendor service call with 50 IDs, not 50 individual calls.

**For consistency-critical operations**, verify the foreign resource exists in the service before creating the record. For non-critical display data, verify in the handler or not at all.
