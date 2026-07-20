> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Worked Example — Contracts Module
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[What This Guide Covers](../01-overview/01-what-this-guide-covers.md)"
  - "[MDG Quick Checklist](../01-overview/03-mdg-checklist.md)"
  - "[Domain Layer](02-domain-layer.md)"
---

# Worked Example — Contracts Module

This section is the capstone of the Module Development Guide. It walks through the complete implementation of the Contracts module from scratch, following every step of the §01 development sequence in order.

The goal is not to repeat the conceptual explanations from earlier sections — it is to show **exactly what files to create, in what order, and what each file contains**, so you can reproduce the same pattern for any new module.

## Module Spec

| Item | Value |
|------|-------|
| Module key | `contracts` |
| Module group | `011` |
| Go package | `awo.so/internal/core/contracts` |
| DB prefix | `011001_`, `011002_`, ... |
| Permission prefix | `contracts.` |
| Task queue | `contracts` |
| Feature flag | `contracts.enabled` |
| Entity type | `contract` |

## Sections in This Worked Example

| # | Section | What It Shows |
|---|---------|--------------|
| 01 | This overview | Module spec, complete file list |
| 02 | Domain layer | All domain files created |
| 03 | Database layer | All migration SQL |
| 04 | SQLC layer | Complete query file |
| 05 | Repository layer | Interface + adapter |
| 06 | Service layer | Full service implementation |
| 07 | Handler layer | Full handler + DTOs |
| 08 | Wire registration | Contracts Wire setup |
| 09 | Tests | Domain + repo + service test examples |
| 10 | Launch checklist | Final verification |

## Complete File Tree

```
internal/core/contracts/
├── domain/
│   ├── contract.go                    # Contract entity, ContractLine
│   ├── status.go                      # ContractStatus type + state machine
│   ├── value_objects.go               # ContractValue, ContractNumber
│   ├── errors.go                      # 7 sentinel errors
│   ├── events.go                      # 5 domain events
│   └── notification_categories.go    # Notification category constants
├── repository/
│   ├── interface.go                   # ContractRepository interface
│   ├── params.go                      # All Params structs
│   ├── contract_sqlc.go               # SQLC adapter implementation
│   └── error_mapping.go               # DB error → domain error
├── service/
│   ├── interface.go                   # ContractService interface
│   ├── requests.go                    # Request/Result types
│   ├── contract_service.go            # Implementation
│   ├── write_ops.go                   # Create, Update, Delete
│   ├── read_ops.go                    # GetByID, List
│   ├── transition_ops.go              # Submit, Approve, Reject, ...
│   └── notifications.go               # Notification builders
├── handler/
│   ├── contract_handler.go            # Handler struct + all methods
│   ├── routes.go                      # ContractRouteRegistrar
│   ├── error_mapping.go               # domain error → HTTP status
│   └── dto/
│       ├── create_request.go
│       ├── update_request.go
│       ├── transition_request.go
│       └── contract_response.go
├── pipeline/
│   ├── import_pipeline.go
│   └── export_pipeline.go
├── workflows/
│   ├── approval_workflow.go
│   └── expiry_workflow.go
├── activities/
│   └── contract_activities.go
├── worker/
│   └── worker.go
└── contracts.go                       # ProviderSet, WorkerSet

db/
├── migration/
│   ├── 011001_create_contracts.up.sql
│   ├── 011001_create_contracts.down.sql
│   ├── 011002_create_contract_lines.up.sql
│   ├── 011002_create_contract_lines.down.sql
│   ├── 011003_create_contracts_views.up.sql
│   └── 011003_create_contracts_views.down.sql
└── queries/
    └── contracts.sql

internal/core/contracts/
└── (test files alongside their subjects)
    ├── domain/contract_test.go
    ├── domain/value_objects_test.go
    ├── repository/contract_integration_test.go
    └── service/contract_service_test.go
```
