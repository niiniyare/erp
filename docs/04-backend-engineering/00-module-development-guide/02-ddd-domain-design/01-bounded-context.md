---
title: Bounded Context
portal: 4 — Backend Engineering
section: 00-module-development-guide/02-ddd-domain-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-entity-design.md
    title: Entity Design
  - path: ./03-state-machine.md
    title: State Machine
---

# Bounded Context

Before writing code, define the module's bounded context. This document explains what a bounded context is in AwoERP terms and how to define one for the contracts module.

## What Is a Bounded Context

A bounded context is the explicit boundary within which a domain model applies. Inside the boundary, every term has one precise meaning. Outside the boundary, the same word may mean something different.

In AwoERP, a bounded context maps directly to:

- One directory: `internal/core/<module>/`
- One Go module package path: `awo.so/internal/core/<module>`
- One migration prefix: `NNN_*`
- One SQLC query file: `db/queries/<module>.sql`

Other modules that need data from this context call the module's service interface — they never import sub-packages or query the database directly.

## Contracts Bounded Context

### Context Name

**Procurement Contracts** — manages the full lifecycle of agreements between the tenant organisation and external vendors or internal business units.

### Ubiquitous Language

These terms have precise meanings inside this context. Use them exactly in code, comments, API docs, and UI labels.

| Term | Meaning |
|------|---------|
| **Contract** | A formal agreement with a defined value, parties, start/end dates, and lifecycle status |
| **Contract Line** | A single deliverable or scope item within a contract; has its own unit price and quantity |
| **Contract Value** | The total monetary commitment in the tenant's base currency |
| **Contract Number** | Human-readable unique reference (e.g., `CONT-2025-0042`) assigned at creation |
| **Status** | The current lifecycle state; one of `draft`, `submitted`, `under_review`, `approved`, `active`, `suspended`, `terminated` |
| **Approval Threshold** | Tenant-configured monetary limit above which contracts require senior approval |
| **Version** | Optimistic lock counter; incremented on every write; used to detect concurrent edits |
| **Entity** | The organisational unit (branch, department, subsidiary) that owns the contract |

### Terms That Mean Something Different Elsewhere

| Term | This Context | Other Contexts |
|------|-------------|----------------|
| `entity_id` | Which organisational unit owns this contract | Entire entity graph (tenant module) |
| `status` | Contract lifecycle state | User account status (IAM), tenant status (tenant module) |
| `value` | Monetary contract total | Generic field in configuration module |

### What This Context Owns

- Contract lifecycle (all status transitions)
- Contract lines (child records)
- Contract approval rules (threshold checks)
- Contract documents (attachments, if implemented)

### What This Context Does NOT Own

- Vendor records → Vendor Management module
- User identity → IAM module
- Organisational hierarchy → Tenant/Entity module
- Payment processing → Finance module
- Notifications delivery → Notifications platform service
- Audit log storage → Audit platform service

### Cross-Context References

This context references other contexts by their IDs only — never by importing their internal packages.

```go
type Contract struct {
    // ...
    VendorID  uuid.UUID  // references vendor module — stored as FK, never joined in domain
    CreatedBy uuid.UUID  // references IAM user — stored as FK
    // ...
}
```

If a handler needs the vendor name for display, it calls the vendor service separately and assembles the response DTO in the handler layer, not in the domain.

## Defining Your Module's Context

Work through these questions before writing any code:

1. **What is the single noun at the centre of this context?** (The primary entity.)
2. **What lifecycle does that noun go through?** (Draw the state machine.)
3. **What child records does it have?** (These become child tables.)
4. **What does it reference from other contexts?** (These become UUID foreign keys.)
5. **What actions does a user perform on it?** (These become service methods and permission strings.)
6. **What changes in this context that other contexts care about?** (These become domain events.)

Answer these before `domain/<noun>.go` exists. The answers drive every file you write.
