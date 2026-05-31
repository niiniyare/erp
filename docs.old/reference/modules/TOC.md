---

## Part 1 — Gap Analysis

### Summary Verdict

The v1 TOC was well-structured for a **mobile SDK / SDUI framework product** but would mislead developers, architects, and implementers of a **production ERP platform**. The gaps fall into five severity classes below.

---

### Critical Gaps (entirely absent)

**ERP Platform Kernel** — the foundational cross-cutting services that every ERP module depends on have no documentation home: business document framework, document numbering, print/PDF framework, attachment framework, notification framework, search framework, and rule engine. These are not module features — they are platform infrastructure that must be documented at the architectural level before any module documentation makes sense.

**Metadata-Driven Architecture** — the platform's defining characteristic (metadata drives documents, forms, views, validations, and workflows rather than hardcoded screens) has zero coverage. There is no section explaining the metadata model, document definitions, field definitions, validation definitions, or how the DSL, AST, and metadata layers relate to each other. This is the conceptual keystone of the entire architecture.

**Go Clean Architecture** — the backend codebase uses a strict Domain → Application → Repository → Infrastructure → API layering. None of these layers are documented. There is no coverage of domain aggregates, application services, repository contracts, infrastructure adapters, transaction management, or CQRS decisions. Engineers joining the project have no architectural map.

**IAM Depth** — Casbin v2 + CEL is in the stack but the authorization documentation covers only generic RBAC/ABAC concepts. Missing entirely: the Casbin policy model for AwoERP, permission inheritance through the org/department/project hierarchy, separation of duties, delegation chains, and multi-level approval authorization.

**Organizational Hierarchy** — AwoERP has a multi-level entity hierarchy (Tenant → Organization → Department → Project → Cost Centre). This hierarchy drives permissions, data scoping, approval routing, and reporting. It is completely undocumented.

**Platform Background Services** — twelve named platform services (Tenant, IAM, User, Feature Flag, Notification, Workflow, File, Search, Audit, Report, Integration, Settings) each warranting a service contract, API reference, and integration guide are scattered or absent entirely.

---

### Serious Gaps (present but severely underweighted)

**Multi-Tenancy** — the v1 TOC has multi-tenancy spread thinly across three portals. PostgreSQL Row Level Security, the `WithTenant()` pattern, tenant provisioning automation, tenant migrations, and cross-tenant administration are mentioned in passing but have no dedicated architecture section.

**ERP Module Standard Template** — v1 modules have inconsistent structure. A production ERP documentation suite requires every module to follow a canonical template (Business Overview → Concepts → Entities → Document Types → Workflows → Permissions → APIs → Mobile Screens → Reports → Integrations → Extension Points). None of the 15 modules in v1 fully follow this.

**Fuel Station Domain** — listed as a strategic domain by the client but documented at the same depth as a generic ERP module. Missing: price management, LPG sub-module, lubricants, fleet account billing, ATG integration, forecourt controller integration, regulatory compliance, and the wetstock reconciliation engine.

**Approval and Workflow Framework** — approval workflows appear in two places (Core Concepts and Domains) with no unified framework document explaining the multi-level approval model, delegation, escalation, timeout handling, and the Temporal workflow backing.

**UI Platform** — the v1 TOC is mobile-only. The platform has (or will have) a web renderer sharing the same DSL, AST, and schema contracts. A unified UI Platform portal is needed covering the shared compiler, the web renderer (likely Flutter Web or a separate web runtime), and the mobile renderer as two deployment targets of one UI pipeline.

---

### Moderate Gaps (exists but thin)

**Developer Experience** — coding standards, repository conventions, ADR process, PR process, contribution guides, and architecture governance are entirely absent. Engineers cannot onboard without these.

**Event Architecture** — Temporal is in the stack, domain events are implied by the audit log, but there is no section covering the event catalogue, event schemas, event-driven integration patterns, or CQRS read-model strategy.

**Report and Print Framework** — reporting appears in the domain portal as a module but the platform-level reporting engine (report definitions, data sources, rendering pipeline, scheduling, export formats) and the print framework (PDF generation, print templates, barcode/QR support) have no platform-level home.

**Search Framework** — platform-wide search (full-text, filtered, cross-module) is absent.

**Settings and Configuration Service** — tenant-level, organization-level, and user-level settings with inheritance and override semantics are undocumented.

---

### Minor Gaps

ADR index exists but ADR authoring process and governance are not documented. Schema Registry gets an API reference but no operational runbook. Integration Service has no connector framework or adapter pattern documentation. Release management exists but is shallow on the branching model and hotfix process.

---

## Part 2 — Recommended Structural Changes

**Reorganize from 7 portals to 9 portals** as proposed, with the following refinements:

Portal 3 (Platform Architecture) should be split into two distinct concerns: the ERP Platform Kernel (cross-cutting services, frameworks, metadata architecture) and the Go Backend Architecture (clean architecture layers, service contracts, infrastructure). These are different audiences — Solution Architects read the first, Backend Engineers read the second.

Portal 7 (ERP Modules) should enforce a mandatory standard module template across all 15+ modules. Every module gets the same 12-section structure. Fuel Station becomes a super-module with sub-modules (Wetstock, Meter Reading, LPG) documented under it rather than as peers.

Portal 5 (UI Platform) should replace the mobile-only renderer documentation with a unified UI Platform covering the shared pipeline and two renderer targets (web, mobile).

Add a new top-level section in Portal 3 for **Metadata Architecture** — this is the conceptual center of the entire platform and must be prominent.

**Consolidate** the scattered approval/workflow content into a single Workflow & Approval Framework section within Portal 3 rather than splitting it across Core Concepts and Domains.

**Promote** IAM to a full portal sub-section within Portal 6 (Operations & Security) covering the complete Casbin model, organizational hierarchy, permission inheritance, delegation, and separation of duties.

---

## Part 3 — Revised Documentation Architecture

Nine portals. Each portal has a primary audience, a secondary audience, and a single coherent concern.

| Portal | Name | Primary Audience | Secondary Audience |
|--------|------|------------------|--------------------|
| 1 | Product & Business | Executives, PMs | ERP Implementers |
| 2 | ERP Concepts | ERP Implementers, Solution Architects | Product Managers |
| 3 | Platform Architecture | Solution Architects | Backend Engineers, Mobile Engineers |
| 4 | Backend Engineering | Backend Engineers | DevOps, Integration Developers |
| 5 | UI Platform | UI DSL Developers, Mobile Engineers | Solution Architects |
| 6 | Operations & Security | DevOps, Security Teams | Backend Engineers |
| 7 | ERP Modules | ERP Implementers, Integration Developers | Backend Engineers |
| 8 | SDK & Extension Framework | Third-Party Developers | Integration Developers |
| 9 | Reference | All Engineers | Support Teams |

---

## Part 4 — Revised Folder Structure

```
docs/
├── README.md                            # Audience routing matrix + portal index
├── GLOSSARY.md                          # Platform-wide terminology
├── CHANGELOG.md
│
├── 01-product/                          # Portal 1 — Product & Business
├── 02-erp-concepts/                     # Portal 2 — ERP Concepts
├── 03-platform-architecture/
│
├── 01-architecture-overview/
│   ├── 01-architecture-principles.md
│   ├── 02-system-context-diagram.md
│   ├── 03-platform-component-map.md
│   ├── 04-deployment-topology.md
│   ├── 05-request-lifecycle-end-to-end.md
│   ├── 06-data-flow-diagrams.md
│   ├── 07-technology-stack-rationale.md
│   └── 08-adr-index.md
│       └── adr/
│           ├── ADR-001-go-backend-stack.md
│           ├── ADR-002-sdui-architecture.md
│           ├── ADR-003-flutter-renderer.md
│           ├── ADR-004-metadata-driven-design.md
│           ├── ADR-005-temporal-for-workflows.md
│           ├── ADR-006-casbin-cel-authorization.md
│           ├── ADR-007-postgresql-rls-tenancy.md
│           ├── ADR-008-offline-first-mobile.md
│           ├── ADR-009-schema-registry.md
│           ├── ADR-010-cqrs-decisions.md
│           ├── ADR-011-event-architecture.md
│           └── ADR-012-pipeline-over-monolithic-service.md
│
├── 02-erp-platform-kernel/
│   ├── 01-kernel-overview.md
│   ├── 02-business-document-framework/
│   │   ├── 01-document-framework-overview.md
│   │   ├── 02-document-type-registry.md
│   │   ├── 03-document-state-machine.md
│   │   ├── 04-document-versioning.md
│   │   ├── 05-document-locking.md
│   │   ├── 06-document-reversal-and-cancellation.md
│   │   └── 07-cross-document-references.md
│   ├── 03-numbering-framework/
│   │   ├── 01-numbering-overview.md
│   │   ├── 02-sequence-definitions.md
│   │   ├── 03-numbering-patterns.md
│   │   ├── 04-multi-branch-numbering.md
│   │   └── 05-numbering-reset-policies.md
│   ├── 04-audit-framework/
│   │   ├── 01-audit-framework-overview.md
│   │   ├── 02-audit-event-model.md
│   │   ├── 03-field-level-change-tracking.md
│   │   ├── 04-audit-storage-and-retention.md
│   │   ├── 05-audit-query-api.md
│   │   └── 06-compliance-audit-reports.md
│   ├── 05-notification-framework/
│   │   ├── 01-notification-framework-overview.md
│   │   ├── 02-notification-channels.md
│   │   ├── 03-notification-templates.md
│   │   ├── 04-notification-routing-rules.md
│   │   ├── 05-notification-preferences.md
│   │   ├── 06-in-app-notification-model.md
│   │   └── 07-notification-delivery-guarantees.md
│   ├── 06-attachment-framework/
│   │   ├── 01-attachment-framework-overview.md
│   │   ├── 02-attachment-model.md
│   │   ├── 03-storage-backends.md
│   │   ├── 04-attachment-lifecycle.md
│   │   ├── 05-access-control-for-attachments.md
│   │   └── 06-mobile-attachment-handling.md
│   ├── 07-search-framework/
│   │   ├── 01-search-framework-overview.md
│   │   ├── 02-full-text-search.md
│   │   ├── 03-faceted-search.md
│   │   ├── 04-cross-module-search.md
│   │   ├── 05-search-indexing-pipeline.md
│   │   └── 06-search-api.md
│   ├── 08-print-framework/
│   │   ├── 01-print-framework-overview.md
│   │   ├── 02-print-template-model.md
│   │   ├── 03-pdf-generation-pipeline.md
│   │   ├── 04-barcode-and-qr-support.md
│   │   ├── 05-print-preview.md
│   │   └── 06-mobile-print-support.md
│   ├── 09-reporting-framework/
│   │   ├── 01-reporting-framework-overview.md
│   │   ├── 02-report-definition-model.md
│   │   ├── 03-report-data-source-model.md
│   │   ├── 04-report-rendering-pipeline.md
│   │   ├── 05-report-scheduling-engine.md
│   │   ├── 06-report-export-formats.md
│   │   ├── 07-report-distribution.md
│   │   └── 08-custom-report-authoring.md
│   └── 10-rule-engine/
│       ├── 01-rule-engine-overview.md
│       ├── 02-rule-definition-model.md
│       ├── 03-cel-expression-engine.md
│       ├── 04-rule-evaluation-lifecycle.md
│       ├── 05-rule-chaining.md
│       ├── 06-business-rule-authoring.md
│       └── 07-rule-testing.md
│
├─── 03-pipeline-architecture/              ◄ NEW — sourced from the new guide
│    │
│    ├── 00-pipeline-overview.md
│    │     # The two systems and their relationship (§1.1)
│    │     # The unified key insight: session flags drive pipeline shape (§1.2)
│    │     # End-to-end request flow diagram (§1.3)
│    │     # Packages: awo/internal/pipeline · awo/internal/workflow · awo/pkg/condition
│   │
│   ├── 01-core-concepts-glossary.md
│   │     # All 15 terms: OperationContext, Stage, Pipeline, PipelineBuilder,
│   │     # StageRegistry, Hook, HookRegistry, CompensationFn, OperationLog,
│   │     # ScriptStage, WorkflowTemplate, WorkflowInstance, UserTask, TxHook (§2)
│   │
│   ├── 02-operation-context/
│   │   ├── 01-operationcontext-design.md
│   │   │     # Full struct layout: Identity, OperationID, Mode, Input,
│   │   │     # Data map, Flags map, Suspension fields, Log, TxHooks (§3.1)
│   │   ├── 02-stage-communication-protocol.md
│   │   │     # Data key convention, standard key prefixes (§3.2)
│   │   │     # The rule: stages communicate only through opCtx.Data
│   │   ├── 03-object-pooling.md
│   │   │     # sync.Pool pattern, AcquireOperationContext / ReleaseOperationContext (§3.3)
│   │   └── 04-convenience-accessors-reference.md
│   │         # All helper methods: GetData, SetData, SetFlag, Flag, Can,
│   │         # FeatureEnabled, SettingDecimal/Bool/String, Suspend, RegisterTxHook
│   │
│   ├── 03-stage-system/
│   │   ├── 01-stage-interface.md
│   │   │     # Full Stage interface: Name, Operations, FeatureFlag, Priority,
│   │   │     # Required, RunCondition, DependsOn, Execute (§4.1)
│   │   │     # Simulatable interface for dry-run (§4.1)
│   │   │     # StageResult struct
│   │   ├── 02-priority-band-system.md
│   │   │     # The 9 priority bands (100–999) and their semantic meaning (§4.2)
│   │   │     # Parallel execution within bands via DependsOn analysis
│   │   │     # Band assignment guide for module developers
│   │   ├── 03-parallel-execution.md
│   │   │     # How independent stages in the same band run concurrently
│   │   │     # DependsOn() as the parallelism declaration
│   │   │     # Worked example: BudgetCheck ∥ SanctionsCheck in band 300–399
│   │   ├── 04-run-condition-gate.md
│   │   │     # RunCondition() as a pkg/condition formula
│   │   │     # Evaluation timing: before hooks, before Execute
│   │   │     # Skip vs. failure semantics
│   │   ├── 05-required-vs-optional-stages.md
│   │   │     # Required=true → pipeline aborts + compensates on error
│   │   │     # Required=false → error logged, execution continues
│   │   │     # Policy: tenant ScriptStages are never Required (safety rule)
│   │   └── 06-writing-a-stage-guide.md
│   │         # Complete worked example: BudgetCheckStage (§4.3)
│   │         # How to implement Execute(), Simulate(), DependsOn()
│   │         # How to read from and write to opCtx.Data
│   │         # How to read tenant settings via opCtx.SettingString/Decimal/Bool
│   │
│   ├── 04-pipeline-and-builder/
│   │   ├── 01-pipeline-execution-lifecycle.md
│   │   │     # Full execution loop: RunCondition → Before hooks → Execute →
│   │   │     # After hooks → StageCompleted event → TxHooks → persist log (§5.1)
│   │   │     # Retry-with-backoff within execution
│   │   │     # Suspension detection and early exit
│   │   ├── 02-pipeline-builder.md
│   │   │     # Build(operationKey, session) algorithm (§5.2)
│   │   │     # Feature flag filtering during build
│   │   │     # Stage sort by priority
│   │   │     # Hook collection
│   │   ├── 03-pipeline-cache.md
│   │   │     # Cache key = "{op_key}:{flags_sha256}" (§10.1)
│   │   │     # 5-minute TTL
│   │   │     # Invalidation on FlagService.Set() → InvalidateForTenant()
│   │   │     # Why caching is safe: pipeline shape is pure function of flags
│   │   └── 04-idempotency.md
│   │         # ExecuteIdempotent() pattern (§9.4)
│   │         # idempotency_key in operation_logs
│   │         # Status-based deduplication: completed / running / pending / failed
│   │         # Idempotency-Key HTTP header convention
│   │
│   ├── 05-hook-system/
│   │   ├── 01-hook-interface.md
│   │   │     # Hook interface: Name, FeatureFlag, Priority, Operations,
│   │   │     # StageBefore, StageAfter, RunCondition, Execute (§6.1)
│   │   │     # SimulatableHook interface
│   │   ├── 02-standard-hooks-reference.md
│   │   │     # WorkflowApprovalHook — fires before gl.post_transaction (§6.2)
│   │   │     # AuditLogHook — fires after every stage
│   │   │     # EInvoiceHook — fires after GL posting
│   │   │     # NotificationHook — fires after payment scheduling
│   │   │     # CompliancePrePostingHook
│   │   └── 03-writing-a-hook-guide.md
│   │         # How to implement Execute() and SimulateExecute()
│   │         # Choosing StageBefore vs. StageAfter
│   │         # RunCondition for hook-level gating
│   │         # Hook registration pattern in wire.go
│   │
│   ├── 06-stage-and-hook-registries/
│   │   ├── 01-registry-design.md
│   │   │     # StageRegistry and HookRegistry as global catalogues
│   │   │     # Registration timing: startup only, via wire.go (§7)
│   │   ├── 02-registration-in-wire.md
│   │   │     # Full wire.go registration example (§7.1)
│   │   │     # Stage registry → Hook registry → Compensation registry → PipelineBuilder
│   │   ├── 03-compensation-registry.md
│   │   │     # CompensationRegistry design
│   │   │     # Which stages register compensations and why
│   │   │     # Standard compensations: GL reversal, payment cancel, DMS void,
│   │   │       budget reservation release
│   │   └── 04-module-stage-contributions-catalogue.md
│   │         # Master table of all stages and hooks across all modules (§19)
│   │         # Columns: Stage/Hook, Module, Operations, Flag Gate, Priority, Required
│   │         # Kept current as modules are added
│   │
│   ├── 07-db-transaction-hooks/
│   │   ├── 01-the-problem-and-solution.md
│   │   │     # Why atomicity matters: GL + domain event + workflow trigger (§8.1)
│   │   │     # The naive approach and its failure mode
│   │   │     # Transactional Outbox Pattern made explicit
│   │   ├── 02-txhook-design.md
│   │   │     # TxHook struct: Name, Priority, Fn signature (§8.2)
│   │   │     # RegisterTxHook() on OperationContext
│   │   │     # TxHook vs. after-hook: the atomicity guarantee
│   │   ├── 03-gl-posting-txhook-example.md
│   │   │     # How GLPostingStage registers three TxHooks (§8.3):
│   │   │     #   gl.seed_domain_event (always)
│   │   │     #   gl.update_budget_actuals (budget flag)
│   │   │     #   gl.seed_workflow_trigger (workflow flag)
│   │   │     # The open pgx.Tx passed via opCtx.Data
│   │   ├── 04-pipeline-txhook-execution.md
│   │   │     # runTxHooks(): priority sort → execute each → COMMIT or ROLLBACK (§8.4)
│   │   │     # Failure semantics: any hook failure rolls back entire transaction
│   │   ├── 05-workflow-trigger-queue.md
│   │   │     # workflow_trigger_queue table schema (§8.5)
│   │   │     # Polling mechanism and processing lifecycle
│   │   │     # Index strategy for pending rows
│   │   └── 06-what-txhooks-enable.md
│   │         # Atomicity guarantee diagram: success path vs. rollback path
│   │         # Use cases: budget actuals, domain events, workflow triggers,
│   │           inventory movements, wetstock reconciliation pending
│   │
│   ├── 08-failure-recovery-and-compensation/
│   │   ├── 01-failure-categories.md
│   │   │     # Five categories: Validation, BusinessRule, Service,
│   │   │     # Integration, Infrastructure (§9.1)
│   │   │     # Recovery strategy per category
│   │   ├── 02-compensation-algorithm.md
│   │   │     # LIFO execution order (§9.2)
│   │   │     # Non-fatal compensation failures: log, continue, flag human intervention
│   │   │     # Standard compensation table per stage
│   │   ├── 03-retry-with-backoff.md
│   │   │     # Exponential backoff with jitter (§9.3)
│   │   │     # isRetryable() classification
│   │   │     # Max 3 retries → exhaustion → compensate
│   │   └── 04-resume-after-suspension.md
│   │         # PipelineResumeService.Resume() (§9.5)
│   │         # Reconstructing OperationContext from persisted log
│   │         # SkipStages(completed) + StartFrom(resumePoint)
│   │         # Resume data injection: approved_by, decision, comment
│   │
│   ├── 09-dry-run-mode/
│   │   ├── 01-dry-run-overview.md
│   │   │     # opCtx.DryRun = true semantics (§12.1)
│   │   │     # Simulatable interface: Execute vs. Simulate dispatch
│   │   │     # Read-only stages run normally in dry-run
│   │   ├── 02-dry-run-api.md
│   │   │     # POST /api/operations/dry-run request/response schema (§12.1)
│   │   │     # would_succeed, would_require_approval, expected_approvers,
│   │   │       expected_gl_entries, expected_tax_amount, expected_payment_date
│   │   └── 03-workflow-hook-in-dry-run.md
│   │         # WorkflowApprovalHook.SimulateExecute() (§12.2)
│   │         # Sets approval_required_in_real_run + expected_approvers
│   │         # Does NOT call opCtx.Suspend() in dry-run
│   │
│   ├── 10-condition-package/              ◄ pkg/condition — entirely new section
│   │   ├── 01-condition-package-overview.md
│   │   │     # Role in the pipeline: four usage contexts (§13.1)
│   │   │     # Stage RunCondition / Hook RunCondition / ScriptStage / Workflow step
│   │   │     # Package path: awo/pkg/condition
│   │   ├── 02-evaluator-and-builder.md
│   │   │     # condition.NewEvaluator(nil, EvalOptions)
│   │   │     # condition.NewBuilder(ConjunctionAnd)
│   │   │     # builder.AddFormula(expr) → group
│   │   │     # evaluator.Evaluate(ctx, group, evalCtx) → bool, error
│   │   ├── 03-eval-options.md
│   │   │     # MaxDepth, MaxConditions, Timeout, RegexTimeout, CacheResults
│   │   │     # Platform defaults vs. tenant script limits (§13.3)
│   │   │     # Why tenant scripts use tighter limits
│   │   ├── 04-eval-context-and-data.md
│   │   │     # condition.NewEvalContext(data, options)
│   │   │     # buildEvalData(opCtx): what tenant expressions can access (§14.2)
│   │   │     #   stage outputs (Data map), flags, feature flags, settings, input, entity
│   │   │     # What is NOT exposed: permissions, other tenants, raw DB
│   │   ├── 05-custom-function-registration.md
│   │   │     # eval.RegisterFunction(name, fn) pattern (§13.2)
│   │   │     # Platform-registered functions:
│   │   │     #   documentTotal(doc_type, doc_id)
│   │   │     #   lineCount(doc_type, doc_id)
│   │   │     #   fieldValue(doc_type, doc_id, field_name)
│   │   │     #   isChildOf(entity_id, parent_entity_id)
│   │   │     #   isWithinPeriod(date, start, end)
│   │   │     #   hasTag(entity_type, entity_id, tag)
│   │   │     #   isApprovedVendor(vendor_id)
│   │   ├── 06-security-model.md
│   │   │     # Per-function session.Can() enforcement (§13.3)
│   │   │     # Session threaded through EvalContext via context key
│   │   │     # Formula cache: one evaluator per ScriptStage = no unbounded growth
│   │   └── 07-expression-syntax-reference.md
│   │         # Operator reference: ==, !=, >, <, >=, <=, &&, ||, !
│   │         # Built-in: toNumber(), toString()
│   │         # Null coalescing: ?? operator
│   │         # Map access: setting["key"], feature["key"]
│   │         # Example tenant expressions with annotations (§14.2)
│   │
│   ├── 11-tenant-scripting-engine/        ◄ ScriptStage — entirely new section
│   │   ├── 01-script-stage-overview.md
│   │   │     # What a ScriptStage is: formula-logic pipeline stage (§14.1)
│   │   │     # Stored in DB: ScriptStageDefinition schema
│   │   │     # Loaded at runtime, registered in StageRegistry per tenant
│   │   ├── 02-script-stage-definition.md
│   │   │     # ScriptStageDefinition fields: ID, TenantID, Name, Formula,
│   │   │     # OnTrue, OnFalse, FeatureFlag, Priority, Operations
│   │   ├── 03-script-stage-execution.md
│   │   │     # Execute() flow: NewEvalContext → builder.AddFormula → Evaluate (§14.1)
│   │   │     # Non-fatal error policy: script errors produce "skipped" not "failed"
│   │   │     # NextStageID branching via OnTrue/OnFalse
│   │   │     # Metrics output: rules_evaluated, duration_ms
│   │   ├── 04-error-classification.md
│   │   │     # classifyScriptError() mapping (§14.1):
│   │   │     # ErrEvaluationTimeout → "timed out (200ms limit)"
│   │   │     # ErrMaxDepthExceeded → "nesting too deep"
│   │   │     # ErrResourceLimitExceeded → "too complex"
│   │   │     # ErrFieldNotFound / ErrFunctionNotFound
│   │   └── 05-authoring-tenant-scripts.md
│   │         # What expressions can reference (§14.2)
│   │         # Worked examples: routing by amount, vendor status, date range,
│   │           entity hierarchy, feature-gated logic
│   │         # Security limits enforced by tenantScriptOptions
│   │         # Testing scripts with the dry-run mode
│   │
│   ├── 12-schema-registry-and-autocomplete/   ◄ entirely new section
│   │   ├── 01-schema-registry-overview.md
│   │   │     # SchemaRegistry as authoritative document type catalogue (§15.2)
│   │   │     # DocumentSchema: TypeKey, Label, Module, ReadPermission,
│   │   │       Fields [], Relationships []
│   │   │     # GettableField: Name, Label, Type, Description, EnumValues,
│   │   │       RequiredPermission
│   │   │     # DocumentRelationship: Name, Label, TargetType, Cardinality, Fields
│   │   ├── 02-module-schema-registration.md
│   │   │     # Registration in wire.go: schemaRegistry.Register(&DocumentSchema{}) (§15.3)
│   │   │     # AP Invoice schema (full field list with types)
│   │   │     # Purchase Order schema
│   │   │     # GL Journal schema
│   │   │     # Pattern for new module registration
│   │   ├── 03-autocomplete-api.md
│   │   │     # GET /api/pipeline/autocomplete?operation_key= (§15.4)
│   │   │     # AutocompleteResponse: Resources, Functions, ContextKeys,
│   │   │       FeatureKeys, SettingKeys, Operators
│   │   │     # Permission filtering: session.Can() applied per resource and field
│   │   │     # buildAutocompleteResponse() algorithm
│   │   ├── 04-context-keys-for-operation.md
│   │   │     # GetContextKeysForOperation(operationKey) — what keys prior stages
│   │   │       will have written by the time a ScriptStage runs
│   │   │     # ContextKeyAutocomplete: Key, Label, Type, StageSource
│   │   ├── 05-expression-validation-api.md
│   │   │     # POST /api/pipeline/validate-expression (§15.6)
│   │   │     # Two-phase validation: syntax (condition.Rule.Validate()) +
│   │   │       access check (ExpressionValidator.CheckAccess())
│   │   │     # Warning vs. error semantics
│   │   ├── 06-document-schema-api.md
│   │   │     # GET /api/pipeline/document-schema/:type_key (§15.7)
│   │   │     # GET /api/pipeline/document-schema/:type_key/relationships/:rel_name
│   │   │     # Permission-filtered field lists
│   │   └── 07-ui-script-editor-integration.md
│   │         # Monaco editor + completionItems from autocomplete API (§15.5)
│   │         # Field explorer tree (§15.7): click to insert expression fragment
│   │         # Validate button → /validate-expression
│   │         # Dry-run preview button → /dry-run
│   │
│   ├── 13-api-handler-design/
│   │   ├── 01-sync-async-split.md
│   │   │     # pipelineSyncDeadline = 8 seconds (§11.1)
│   │   │     # Goroutine + select pattern: result / error / timeout
│   │   │     # 202 response with poll_url and poll_interval_ms
│   │   ├── 02-status-polling-endpoint.md
│   │   │     # GET /api/operations/:id/status
│   │   │     # Standard error response schema with code, message, stage,
│   │   │       recoverable, user_action_required
│   │   ├── 03-websocket-live-updates.md
│   │   │     # GET /api/operations/:id/live — WebSocket
│   │   │     # StageCompleted event schema
│   │   │     # UI live progress rendering with amis ws: field
│   │   └── 04-idempotency-key-header.md
│   │         # Idempotency-Key HTTP header usage
│   │         # Client responsibility for key generation
│   │         # Server deduplication on retry
│   │
│   ├── 14-performance-optimisation/
│   │   ├── 01-pipeline-cache.md
│   │   │     # pipelineCache: sync.RWMutex + TTL map (§10)
│   │   │     # Cache key construction
│   │   │     # Invalidation strategy
│   │   ├── 02-parallel-band-execution.md
│   │   │     # Concurrency within priority bands
│   │   │     # ~40% latency reduction on full enterprise tenants
│   │   └── 03-operation-log-write-strategy.md
│   │         # In-memory accumulation → single UPDATE at completion
│   │         # Compensation logs: synchronous (audit requirement)
│   │         # All other log writes: non-blocking goroutines with 5s timeout
│   │
│   └── 15-canonical-example-ap-invoice/
│       ├── 01-complete-pipeline-definition.md
│       │     # Full stage table for ap.invoice.process (§18.1):
│       │     # All 14 stages + 5 hooks with module, flag, priority, required
│       ├── 02-tenant-configuration-comparison.md
│       │     # Basic Finance tenant: 6 stages, 1 hook
│       │     # Full Enterprise tenant: 14 stages, 5 hooks
│       │     # Same service entry point — different pipeline shape
│       ├── 03-service-entry-point.md
│       │     # APService.ProcessInvoice() complete implementation (§18.3)
│       │     # Permission check → invoice fetch → PipelineBuilder.Build →
│       │       AcquireOperationContext → ExecuteIdempotent → result mapping
│       └── 04-end-to-end-flow-walkthrough.md
│             # Narrative walkthrough: POST → build pipeline → execute stages →
│               suspend for approval → 202 response → approver acts →
│               Temporal signal → Resume → GL posted → WebSocket update
│
├── 04-metadata-architecture/
│   ├── 01-metadata-architecture-overview.md
│   ├── 02-metadata-driven-development.md
│   ├── 03-metadata-registry.md
│   ├── 04-document-definitions/
│   │   ├── 01-document-definition-model.md
│   │   ├── 02-document-definition-schema.md
│   │   ├── 03-document-type-inheritance.md
│   │   └── 04-document-definition-authoring.md
│   ├── 05-field-definitions/
│   │   ├── 01-field-definition-model.md
│   │   ├── 02-field-types-catalogue.md
│   │   ├── 03-field-constraints.md
│   │   └── 04-computed-fields.md
│   ├── 06-form-definitions/
│   │   ├── 01-form-definition-model.md
│   │   ├── 02-form-layout-definitions.md
│   │   ├── 03-form-visibility-rules.md
│   │   └── 04-form-definition-to-dsl-compilation.md
│   ├── 07-view-definitions/
│   │   ├── 01-view-definition-model.md
│   │   ├── 02-list-view-definitions.md
│   │   ├── 03-detail-view-definitions.md
│   │   └── 04-view-permissions.md
│   ├── 08-validation-definitions/
│   │   ├── 01-validation-definition-model.md
│   │   ├── 02-field-validators.md
│   │   ├── 03-cross-field-validators.md
│   │   ├── 04-async-validators.md
│   │   └── 05-validation-error-messages.md
│   ├── 09-workflow-definitions/
│   │   ├── 01-workflow-definition-model.md
│   │   ├── 02-approval-step-definitions.md
│   │   ├── 03-condition-definitions.md
│   │   └── 04-workflow-definition-to-temporal.md
│   ├── 10-dsl-metadata-model/
│   │   ├── 01-dsl-metadata-relationship.md
│   │   ├── 02-metadata-to-ast-compilation.md
│   │   ├── 03-metadata-driven-screen-generation.md
│   │   └── 04-metadata-override-and-customization.md
│   └── 11-metadata-governance/
│       ├── 01-metadata-versioning.md
│       ├── 02-metadata-migrations.md
│       └── 03-metadata-testing.md
│
├── 05-multi-tenancy/
│   ├── 01-multi-tenancy-overview.md
│   ├── 02-tenancy-model.md
│   ├── 03-row-level-security/
│   │   ├── 01-rls-architecture.md
│   │   ├── 02-rls-policy-design.md
│   │   ├── 03-tenant-context-propagation.md
│   │   ├── 04-with-tenant-pattern.md
│   │   ├── 05-rls-testing-strategy.md
│   │   └── 06-rls-performance-considerations.md
│   ├── 04-tenant-provisioning/
│   │   ├── 01-provisioning-workflow.md
│   │   ├── 02-tenant-schema-initialization.md
│   │   ├── 03-tenant-seed-data.md
│   │   └── 04-provisioning-automation.md
│   ├── 05-tenant-lifecycle/
│   │   ├── 01-lifecycle-states.md
│   │   ├── 02-activation-and-suspension.md
│   │   ├── 03-archival-and-deletion.md
│   │   └── 04-data-export-on-offboarding.md
│   ├── 06-tenant-configuration/
│   │   ├── 01-tenant-settings-model.md
│   │   ├── 02-module-activation.md
│   │   ├── 03-feature-flag-management.md
│   │   └── 04-locale-and-timezone.md
│   ├── 07-tenant-migrations/
│   │   ├── 01-migration-strategy.md
│   │   ├── 02-zero-downtime-migrations.md
│   │   └── 03-tenant-specific-migrations.md
│   └── 08-cross-tenant-administration/
│       ├── 01-superadmin-model.md
│       ├── 02-cross-tenant-operations.md
│       └── 03-tenant-usage-reporting.md
│
├── 06-iam-architecture/
│   ├── 01-iam-architecture-overview.md
│   ├── 02-identity-model/
│   │   ├── 01-user-identity-model.md
│   │   ├── 02-service-account-model.md
│   │   ├── 03-api-key-model.md
│   │   ├── 04-session-model.md
│   │   └── 05-sso-integration-model.md
│   ├── 03-authorization-model/
│   │   ├── 01-casbin-policy-architecture.md
│   │   ├── 02-rbac-model.md
│   │   ├── 03-abac-model.md
│   │   ├── 04-cel-policy-expressions.md
│   │   ├── 05-permission-string-format.md
│   │   ├── 06-permission-inheritance-by-hierarchy.md
│   │   ├── 07-resource-scoping.md
│   │   └── 08-field-level-permissions.md
│   ├── 04-organizational-authorization/
│   │   ├── 01-org-hierarchy-permissions.md
│   │   ├── 02-department-scoped-access.md
│   │   ├── 03-project-scoped-access.md
│   │   └── 04-branch-scoped-access.md
│   ├── 05-delegation-and-proxy/
│   │   ├── 01-delegation-model.md
│   │   ├── 02-delegation-limits.md
│   │   ├── 03-time-bound-delegation.md
│   │   └── 04-delegation-audit-trail.md
│   ├── 06-separation-of-duties/
│   │   ├── 01-sod-model.md
│   │   ├── 02-conflicting-role-matrix.md
│   │   ├── 03-sod-enforcement.md
│   │   └── 04-sod-violation-alerts.md
│   ├── 07-authentication-architecture/
│   │   ├── 01-authentication-flows.md
│   │   ├── 02-mfa-architecture.md
│   │   ├── 03-oauth2-and-oidc.md
│   │   ├── 04-mobile-authentication.md
│   │   └── 05-session-management.md
│   └── 08-permission-aware-rendering/
│       ├── 01-server-side-permission-injection.md
│       ├── 02-ui-element-visibility-by-permission.md
│       ├── 03-action-availability-by-permission.md
│       └── 04-field-masking-by-permission.md
│
├── 07-workflow-and-approval-framework/
│   ├── 01-framework-overview.md
│   ├── 02-temporal-integration/
│   │   ├── 01-temporal-architecture-in-awoerp.md
│   │   ├── 02-workflow-worker-topology.md
│   │   ├── 03-workflow-registration.md
│   │   └── 04-temporal-namespace-strategy.md
│   ├── 03-approval-framework/
│   │   ├── 01-approval-framework-architecture.md
│   │   ├── 02-approval-rule-engine.md
│   │   ├── 03-approval-chain-model.md
│   │   ├── 04-multi-level-approval-sequences.md
│   │   ├── 05-parallel-approvals.md
│   │   ├── 06-conditional-routing.md
│   │   ├── 07-approval-authority-limits.md
│   │   ├── 08-delegation-in-approvals.md
│   │   ├── 09-escalation-and-timeout.md
│   │   ├── 10-approval-notifications.md
│   │   └── 11-approval-audit-trail.md
│   ├── 04-customer-workflow-engine/        ◄ sourced from guide §16
│   │   ├── 01-workflow-engine-overview.md
│   │   │     # Purpose: long-running, human-in-the-loop, durable (§16.1)
│   │   │     # Characteristics: Temporal-backed, declarative JSON, tenant-designed
│   │   ├── 02-workflow-database-schema.md
│   │   │     # workflow_templates, workflow_instances, workflow_step_executions,
│   │   │     # workflow_user_tasks, workflow_trigger_queue (§16.2)
│   │   │     # Full DDL with RLS policies
│   │   ├── 03-workflow-definition-language.md
│   │   │     # Declarative JSON schema (§16.3)
│   │   │     # Step types: validation, condition, user_task, notification,
│   │   │       wait, parallel, loop, script, webhook
│   │   │     # Variables block, condition cases, on_success/on_failure routing
│   │   │     # Worked example: Invoice Approval Workflow
│   │   ├── 04-step-types-reference.md
│   │   │     # validation — rules engine against variables
│   │   │     # condition — pkg/condition expression routing
│   │   │     # user_task — human approval/review/decision/input/acknowledgment
│   │   │     # notification — channel + template + recipients
│   │   │     # wait — timer / external event signal
│   │   │     # parallel — child workflow branches
│   │   │     # script — tenant formula step
│   │   ├── 05-user-task-model.md
│   │   │     # workflow_user_tasks schema in depth
│   │   │     # Task types: approval, review, input, decision, acknowledgment
│   │   │     # Assignment: user_id, role, entity scope
│   │   │     # Deadline, escalation, delegation model
│   │   │     # form_schema: AMIS form for data collection
│   │   │     # action_buttons: custom decision buttons
│   │   ├── 06-temporal-workflow-function.md
│   │   │     # CustomWorkflowExecution Temporal workflow function (§16.4)
│   │   │     # User task flow: create row → notify → wait for signal → resume
│   │   │     # Signal name pattern: "task-{id}-completed"
│   │   └── 07-trigger-listener.md
│   │         # TriggerListener poller (§17.3)
│   │         # workflow_trigger_queue → match triggers → evaluate conditions →
│   │           start Temporal workflow → mark processed
│   ├── 05-pipeline-workflow-integration/   ◄ sourced from guide §17
│   │   ├── 01-workflowapprovalhook.md
│   │   │     # Full hook implementation (§17.1)
│   │   │     # RunCondition: amount > threshold || budget_exceeded
│   │   │     # GetApprovalRule(), StartWorkflow(), opCtx.Suspend()
│   │   │     # SimulateExecute() for dry-run: no suspension
│   │   ├── 02-workflow-completion-to-pipeline-resume.md
│   │   │     # CompleteWorkflowInstanceActivity (§17.2)
│   │   │     # approved → PipelineResumeService.Resume()
│   │   │     # rejected → MarkRejected() (no compensation — GL was gated)
│   │   └── 03-txhook-seeds-workflow-trigger.md
│   │         # How the GL TxHook and TriggerListener connect (§17.3)
│   │         # Atomicity guarantee: GL commit = workflow trigger commit
│   ├── 06-workflow-definition-runtime/
│   │   ├── 01-workflow-definition-to-temporal.md
│   │   ├── 02-activity-catalogue.md
│   │   ├── 03-workflow-state-and-ui-sync.md
│   │   ├── 04-workflow-error-handling.md
│   │   └── 05-workflow-observability.md
│   └── 07-business-process-automation/
│       ├── 01-event-triggered-workflows.md
│       ├── 02-scheduled-workflows.md
│       ├── 03-cross-document-workflows.md
│       └── 04-integration-workflows.md
│
├── 08-sdui-architecture/
│   ├── 01-sdui-design-philosophy.md
│   ├── 02-ui-pipeline-overview.md
│   ├── 03-metadata-to-ui-pipeline.md
│   ├── 04-ast-to-json-pipeline.md
│   ├── 05-schema-versioning-strategy.md
│   ├── 06-rendering-contract.md
│   ├── 07-permission-and-flag-injection.md
│   ├── 08-data-binding-architecture.md
│   ├── 09-offline-ui-architecture.md
│   └── 10-sdui-security-model.md
│
└── 09-observability-architecture/
    ├── 01-observability-strategy.md
    ├── 02-logging-architecture.md
    ├── 03-metrics-architecture.md
    ├── 04-tracing-architecture.md
    ├── 05-alerting-architecture.md
    └── 06-mobile-observability-architecture.md│
├── 04-backend-engineering/              # Portal 4 — Backend Engineering
│   ├── 01-go-architecture/
│   ├── 02-platform-services/
│   ├── 03-database/
│   ├── 04-api-design/
│   ├── 05-event-architecture/
│   ├── 06-temporal-workflows/
│   ├── 07-testing/
    └── 10-pipeline-testing-patterns.md    ◄ NEW 
          # Stage isolation testing (§20.1): no pipeline needed, opCtx built directly
        # TxHook testing (§20.2): verify hooks registered, rollback on failure
        # Pipeline integration testing (§20.3): stage count, order, resume
        # Expression validation testing (§20.4): valid formula, timeout handled
        # Autocomplete API testing (§20.5): permission filtering, context keys│   └── 08-developer-standards/
│
├── 05-ui-platform/                      # Portal 5 — UI Platform
│   ├── 01-ui-platform-overview/
│   ├── 02-ui-dsl-specification/
│   ├── 03-ast-design/
│   ├── 04-compiler/
│   ├── 05-schema-contracts/
│   ├── 06-component-system/
│   ├── 07-form-system/
│   ├── 08-action-system/
│   ├── 09-navigation-system/
│   ├── 10-web-renderer/
│   └── 11-mobile-renderer/
│
├── 06-operations-security/              # Portal 6 — Operations & Security
│   ├── 01-deployment/
│   ├── 02-iam-operations/
│   ├── 03-security/
│   ├── 04-observability/
│   ├── 05-performance/
│   └── 06-runbooks/
│        ├── RB-009-pipeline-stage-failed-after-retries.md    ◄ NEW
│        ├──  RB-010-txhook-failure-transaction-rolled-back.md  ◄ NEW
│        ├── RB-011-operation-stuck-in-pending-approval.md     ◄ NEW
│        └── RB-012-script-stage-timeout-spike.md              ◄ NEW
│
├── 07-erp-modules/                      # Portal 7 — ERP Modules
│   ├── _template/                       # Standard module template (12 sections)
│   ├── 01-finance/
│   ├── 02-crm/
│   ├── 03-inventory/
│   ├── 04-procurement/
│   ├── 05-sales/
│   ├── 06-manufacturing/
│   ├── 07-hr/
│   ├── 08-projects/
│   ├── 09-asset-management/
│   ├── 10-fuel-station/                 # Super-module
│   │   ├── 10a-station-operations/
│   │   ├── 10b-wetstock/
│   │   ├── 10c-meter-reading/
│   │   └── 10d-lpg/
│   ├── 11-approval-workflows/
│   ├── 12-dashboards/
│   └── 13-reporting/
│
├── 08-sdk-extensions/                   # Portal 8 — SDK & Extension Framework
│   ├── 01-extension-model/
│   ├── 02-go-server-sdk/
│   ├── 03-flutter-client-sdk/
│   ├── 04-dsl-sdk/
│   ├── 05-plugin-system/
│   ├── 06-custom-components/
│   ├── 07-custom-actions/
│   └── 08-integration-adapters/
│
├── 09-reference/                        # Portal 9 — Reference
│   ├── 01-api-reference/
│   ├── 02-schema-catalogue/
│   ├── 03-permission-catalogue/
│   ├── 04-error-codes/
│   ├── 05-event-catalogue/
│   ├── 06-migration-guides/
│   ├── 07-release-management/
│   └── 08-troubleshooting/
│
└── assets/
    ├── diagrams/
    ├── adr/                             # Architecture Decision Records
    └── images/
```

---

## Part 5 — Revised Full Hierarchical Table of Contents

---

### Portal 1 — Product & Business

```
01-product/
├── 01-introduction/
│   ├── 01-what-is-awoerp.md
│   ├── 02-platform-principles.md
│   ├── 03-sdui-primer-for-business.md
│   ├── 04-metadata-driven-erp-primer.md
│   ├── 05-key-capabilities.md
│   ├── 06-platform-editions-and-licensing.md
│   ├── 07-supported-erp-domains.md
│   └── 08-terminology-glossary.md
│
├── 02-business-value/
│   ├── 01-executive-summary.md
│   ├── 02-value-proposition.md
│   ├── 03-total-cost-of-ownership.md
│   ├── 04-no-app-update-advantage.md
│   ├── 05-multi-tenant-saas-model.md
│   └── 06-competitive-positioning.md
│
├── 03-product-roadmap/
│   ├── 01-roadmap-overview.md
│   ├── 02-module-availability-matrix.md
│   ├── 03-platform-maturity-model.md
│   └── 04-release-cadence.md
│
└── 04-getting-started/
    ├── 01-prerequisites.md
    ├── 02-platform-quick-tour.md
    ├── 03-first-tenant-onboarding.md
    ├── 04-first-module-activation.md
    ├── 05-mobile-app-first-run.md
    └── 06-sandbox-and-demo-environment.md
```

---

### Portal 2 — ERP Concepts

```
02-erp-concepts/
├── 01-erp-fundamentals/
│   ├── 01-erp-overview.md
│   ├── 02-erp-domain-model.md
│   ├── 03-master-data-management.md
│   ├── 04-transactional-data-model.md
│   ├── 05-business-document-concepts.md
│   ├── 06-document-lifecycle.md
│   ├── 07-document-numbering.md
│   ├── 08-period-management.md
│   └── 09-currency-and-localization.md
│
├── 02-organizational-model/
│   ├── 01-tenant-model.md
│   ├── 02-organization-model.md
│   ├── 03-department-model.md
│   ├── 04-project-and-cost-centre-model.md
│   ├── 05-branch-and-location-model.md
│   ├── 06-hierarchy-traversal.md
│   └── 07-inter-company-transactions.md
│
├── 03-identity-and-access/
│   ├── 01-identity-concepts.md
│   ├── 02-user-types-and-personas.md
│   ├── 03-roles-and-permissions.md
│   ├── 04-permission-inheritance.md
│   ├── 05-delegation-and-proxy.md
│   ├── 06-separation-of-duties.md
│   └── 07-approval-authority-model.md
│
├── 04-workflow-and-approval-concepts/
│   ├── 01-workflow-concepts.md
│   ├── 02-approval-concepts.md
│   ├── 03-multi-level-approvals.md
│   ├── 04-escalation-and-delegation.md
│   ├── 05-approval-authority-limits.md
│   └── 06-workflow-audit-trail.md
│
├── 05-metadata-driven-concepts/
│   ├── 01-what-is-metadata-driven.md
│   ├── 02-document-definitions.md
│   ├── 03-field-definitions.md
│   ├── 04-form-definitions.md
│   ├── 05-view-definitions.md
│   ├── 06-workflow-definitions.md
│   ├── 07-validation-definitions.md
│   ├── 08-business-rule-definitions.md
│   └── 09-metadata-vs-code-tradeoffs.md
│
└── 06-platform-services-overview/
    ├── 01-platform-services-map.md
    ├── 02-tenant-service-overview.md
    ├── 03-iam-service-overview.md
    ├── 04-notification-service-overview.md
    ├── 05-workflow-service-overview.md
    ├── 06-file-service-overview.md
    ├── 07-search-service-overview.md
    ├── 08-audit-service-overview.md
    ├── 09-report-service-overview.md
    ├── 10-integration-service-overview.md
    └── 11-settings-service-overview.md
```

---

### Portal 3 — Platform Architecture

```
03-platform-architecture/
│
├── 01-architecture-overview/
│   ├── 01-architecture-principles.md
│   ├── 02-system-context-diagram.md
│   ├── 03-platform-component-map.md
│   ├── 04-deployment-topology.md
│   ├── 05-request-lifecycle-end-to-end.md
│   ├── 06-data-flow-diagrams.md
│   ├── 07-technology-stack-rationale.md
│   └── 08-adr-index.md
│       └── adr/
│           ├── ADR-001-go-backend-stack.md
│           ├── ADR-002-sdui-architecture.md
│           ├── ADR-003-flutter-renderer.md
│           ├── ADR-004-metadata-driven-design.md
│           ├── ADR-005-temporal-for-workflows.md
│           ├── ADR-006-casbin-cel-authorization.md
│           ├── ADR-007-postgresql-rls-tenancy.md
│           ├── ADR-008-offline-first-mobile.md
│           ├── ADR-009-schema-registry.md
│           ├── ADR-010-cqrs-decisions.md
│           └── ADR-011-event-architecture.md
│
├── 02-erp-platform-kernel/
│   ├── 01-kernel-overview.md
│   ├── 02-business-document-framework/
│   │   ├── 01-document-framework-overview.md
│   │   ├── 02-document-type-registry.md
│   │   ├── 03-document-state-machine.md
│   │   ├── 04-document-versioning.md
│   │   ├── 05-document-locking.md
│   │   ├── 06-document-reversal-and-cancellation.md
│   │   └── 07-cross-document-references.md
│   ├── 03-numbering-framework/
│   │   ├── 01-numbering-overview.md
│   │   ├── 02-sequence-definitions.md
│   │   ├── 03-numbering-patterns.md
│   │   ├── 04-multi-branch-numbering.md
│   │   └── 05-numbering-reset-policies.md
│   ├── 04-audit-framework/
│   │   ├── 01-audit-framework-overview.md
│   │   ├── 02-audit-event-model.md
│   │   ├── 03-field-level-change-tracking.md
│   │   ├── 04-audit-storage-and-retention.md
│   │   ├── 05-audit-query-api.md
│   │   └── 06-compliance-audit-reports.md
│   ├── 05-notification-framework/
│   │   ├── 01-notification-framework-overview.md
│   │   ├── 02-notification-channels.md
│   │   ├── 03-notification-templates.md
│   │   ├── 04-notification-routing-rules.md
│   │   ├── 05-notification-preferences.md
│   │   ├── 06-in-app-notification-model.md
│   │   └── 07-notification-delivery-guarantees.md
│   ├── 06-attachment-framework/
│   │   ├── 01-attachment-framework-overview.md
│   │   ├── 02-attachment-model.md
│   │   ├── 03-storage-backends.md
│   │   ├── 04-attachment-lifecycle.md
│   │   ├── 05-access-control-for-attachments.md
│   │   └── 06-mobile-attachment-handling.md
│   ├── 07-search-framework/
│   │   ├── 01-search-framework-overview.md
│   │   ├── 02-full-text-search.md
│   │   ├── 03-faceted-search.md
│   │   ├── 04-cross-module-search.md
│   │   ├── 05-search-indexing-pipeline.md
│   │   └── 06-search-api.md
│   ├── 08-print-framework/
│   │   ├── 01-print-framework-overview.md
│   │   ├── 02-print-template-model.md
│   │   ├── 03-pdf-generation-pipeline.md
│   │   ├── 04-barcode-and-qr-support.md
│   │   ├── 05-print-preview.md
│   │   └── 06-mobile-print-support.md
│   ├── 09-reporting-framework/
│   │   ├── 01-reporting-framework-overview.md
│   │   ├── 02-report-definition-model.md
│   │   ├── 03-report-data-source-model.md
│   │   ├── 04-report-rendering-pipeline.md
│   │   ├── 05-report-scheduling-engine.md
│   │   ├── 06-report-export-formats.md
│   │   ├── 07-report-distribution.md
│   │   └── 08-custom-report-authoring.md
│   └── 10-rule-engine/
│       ├── 01-rule-engine-overview.md
│       ├── 02-rule-definition-model.md
│       ├── 03-cel-expression-engine.md
│       ├── 04-rule-evaluation-lifecycle.md
│       ├── 05-rule-chaining.md
│       ├── 06-business-rule-authoring.md
│       └── 07-rule-testing.md
│
├── 03-metadata-architecture/
│   ├── 01-metadata-architecture-overview.md
│   ├── 02-metadata-driven-development.md
│   ├── 03-metadata-registry.md
│   ├── 04-document-definitions/
│   │   ├── 01-document-definition-model.md
│   │   ├── 02-document-definition-schema.md
│   │   ├── 03-document-type-inheritance.md
│   │   └── 04-document-definition-authoring.md
│   ├── 05-field-definitions/
│   │   ├── 01-field-definition-model.md
│   │   ├── 02-field-types-catalogue.md
│   │   ├── 03-field-constraints.md
│   │   └── 04-computed-fields.md
│   ├── 06-form-definitions/
│   │   ├── 01-form-definition-model.md
│   │   ├── 02-form-layout-definitions.md
│   │   ├── 03-form-visibility-rules.md
│   │   └── 04-form-definition-to-dsl-compilation.md
│   ├── 07-view-definitions/
│   │   ├── 01-view-definition-model.md
│   │   ├── 02-list-view-definitions.md
│   │   ├── 03-detail-view-definitions.md
│   │   └── 04-view-permissions.md
│   ├── 08-validation-definitions/
│   │   ├── 01-validation-definition-model.md
│   │   ├── 02-field-validators.md
│   │   ├── 03-cross-field-validators.md
│   │   ├── 04-async-validators.md
│   │   └── 05-validation-error-messages.md
│   ├── 09-workflow-definitions/
│   │   ├── 01-workflow-definition-model.md
│   │   ├── 02-approval-step-definitions.md
│   │   ├── 03-condition-definitions.md
│   │   └── 04-workflow-definition-to-temporal.md
│   ├── 10-dsl-metadata-model/
│   │   ├── 01-dsl-metadata-relationship.md
│   │   ├── 02-metadata-to-ast-compilation.md
│   │   ├── 03-metadata-driven-screen-generation.md
│   │   └── 04-metadata-override-and-customization.md
│   └── 11-metadata-governance/
│       ├── 01-metadata-versioning.md
│       ├── 02-metadata-migrations.md
│       └── 03-metadata-testing.md
│
├── 04-multi-tenancy/
│   ├── 01-multi-tenancy-overview.md
│   ├── 02-tenancy-model.md
│   ├── 03-row-level-security/
│   │   ├── 01-rls-architecture.md
│   │   ├── 02-rls-policy-design.md
│   │   ├── 03-tenant-context-propagation.md
│   │   ├── 04-with-tenant-pattern.md
│   │   ├── 05-rls-testing-strategy.md
│   │   └── 06-rls-performance-considerations.md
│   ├── 04-tenant-provisioning/
│   │   ├── 01-provisioning-workflow.md
│   │   ├── 02-tenant-schema-initialization.md
│   │   ├── 03-tenant-seed-data.md
│   │   └── 04-provisioning-automation.md
│   ├── 05-tenant-lifecycle/
│   │   ├── 01-lifecycle-states.md
│   │   ├── 02-activation-and-suspension.md
│   │   ├── 03-archival-and-deletion.md
│   │   └── 04-data-export-on-offboarding.md
│   ├── 06-tenant-configuration/
│   │   ├── 01-tenant-settings-model.md
│   │   ├── 02-module-activation.md
│   │   ├── 03-feature-flag-management.md
│   │   └── 04-locale-and-timezone.md
│   ├── 07-tenant-migrations/
│   │   ├── 01-migration-strategy.md
│   │   ├── 02-zero-downtime-migrations.md
│   │   └── 03-tenant-specific-migrations.md
│   └── 08-cross-tenant-administration/
│       ├── 01-superadmin-model.md
│       ├── 02-cross-tenant-operations.md
│       └── 03-tenant-usage-reporting.md
│
├── 05-iam-architecture/
│   ├── 01-iam-architecture-overview.md
│   ├── 02-identity-model/
│   │   ├── 01-user-identity-model.md
│   │   ├── 02-service-account-model.md
│   │   ├── 03-api-key-model.md
│   │   ├── 04-session-model.md
│   │   └── 05-sso-integration-model.md
│   ├── 03-authorization-model/
│   │   ├── 01-casbin-policy-architecture.md
│   │   ├── 02-rbac-model.md
│   │   ├── 03-abac-model.md
│   │   ├── 04-cel-policy-expressions.md
│   │   ├── 05-permission-string-format.md
│   │   ├── 06-permission-inheritance-by-hierarchy.md
│   │   ├── 07-resource-scoping.md
│   │   └── 08-field-level-permissions.md
│   ├── 04-organizational-authorization/
│   │   ├── 01-org-hierarchy-permissions.md
│   │   ├── 02-department-scoped-access.md
│   │   ├── 03-project-scoped-access.md
│   │   └── 04-branch-scoped-access.md
│   ├── 05-delegation-and-proxy/
│   │   ├── 01-delegation-model.md
│   │   ├── 02-delegation-limits.md
│   │   ├── 03-time-bound-delegation.md
│   │   └── 04-delegation-audit-trail.md
│   ├── 06-separation-of-duties/
│   │   ├── 01-sod-model.md
│   │   ├── 02-conflicting-role-matrix.md
│   │   ├── 03-sod-enforcement.md
│   │   └── 04-sod-violation-alerts.md
│   ├── 07-authentication-architecture/
│   │   ├── 01-authentication-flows.md
│   │   ├── 02-mfa-architecture.md
│   │   ├── 03-oauth2-and-oidc.md
│   │   ├── 04-mobile-authentication.md
│   │   └── 05-session-management.md
│   └── 08-permission-aware-rendering/
│       ├── 01-server-side-permission-injection.md
│       ├── 02-ui-element-visibility-by-permission.md
│       ├── 03-action-availability-by-permission.md
│       └── 04-field-masking-by-permission.md
│
├── 06-workflow-and-approval-framework/
│   ├── 01-framework-overview.md
│   ├── 02-temporal-integration/
│   │   ├── 01-temporal-architecture-in-awoerp.md
│   │   ├── 02-workflow-worker-topology.md
│   │   ├── 03-workflow-registration.md
│   │   └── 04-temporal-namespace-strategy.md
│   ├── 03-approval-framework/
│   │   ├── 01-approval-framework-architecture.md
│   │   ├── 02-approval-rule-engine.md
│   │   ├── 03-approval-chain-model.md
│   │   ├── 04-multi-level-approval-sequences.md
│   │   ├── 05-parallel-approvals.md
│   │   ├── 06-conditional-routing.md
│   │   ├── 07-approval-authority-limits.md
│   │   ├── 08-delegation-in-approvals.md
│   │   ├── 09-escalation-and-timeout.md
│   │   ├── 10-approval-notifications.md
│   │   └── 11-approval-audit-trail.md
│   ├── 04-workflow-definition-runtime/
│   │   ├── 01-workflow-definition-to-temporal.md
│   │   ├── 02-activity-catalogue.md
│   │   ├── 03-workflow-state-and-ui-sync.md
│   │   ├── 04-workflow-error-handling.md
│   │   └── 05-workflow-observability.md
│   └── 05-business-process-automation/
│       ├── 01-event-triggered-workflows.md
│       ├── 02-scheduled-workflows.md
│       ├── 03-cross-document-workflows.md
│       └── 04-integration-workflows.md
│
├── 07-sdui-architecture/
│   ├── 01-sdui-design-philosophy.md
│   ├── 02-ui-pipeline-overview.md
│   ├── 03-metadata-to-ui-pipeline.md
│   ├── 04-ast-to-json-pipeline.md
│   ├── 05-schema-versioning-strategy.md
│   ├── 06-rendering-contract.md
│   ├── 07-permission-and-flag-injection.md
│   ├── 08-data-binding-architecture.md
│   ├── 09-offline-ui-architecture.md
│   └── 10-sdui-security-model.md
│
└── 08-observability-architecture/
    ├── 01-observability-strategy.md
    ├── 02-logging-architecture.md
    ├── 03-metrics-architecture.md
    ├── 04-tracing-architecture.md
    ├── 05-alerting-architecture.md
    └── 06-mobile-observability-architecture.md
```

---

### Portal 4 — Backend Engineering

```
04-backend-engineering/
│
├── 01-go-architecture/
│   ├── 01-clean-architecture-overview.md
│   ├── 02-domain-layer/
│   │   ├── 01-domain-layer-overview.md
│   │   ├── 02-entities-and-aggregates.md
│   │   ├── 03-value-objects.md
│   │   ├── 04-domain-events.md
│   │   ├── 05-domain-errors.md
│   │   └── 06-domain-services.md
│   ├── 03-application-layer/
│   │   ├── 01-application-layer-overview.md
│   │   ├── 02-service-contracts.md
│   │   ├── 03-command-handlers.md
│   │   ├── 04-query-handlers.md
│   │   ├── 05-application-errors.md
│   │   └── 06-transaction-management.md
│   ├── 04-repository-layer/
│   │   ├── 01-repository-pattern.md
│   │   ├── 02-repository-interfaces.md
│   │   ├── 03-sqlc-adapter-pattern.md
│   │   ├── 04-query-composition.md
│   │   └── 05-pagination-and-filtering.md
│   ├── 05-infrastructure-layer/
│   │   ├── 01-infrastructure-layer-overview.md
│   │   ├── 02-database-infrastructure.md
│   │   ├── 03-cache-infrastructure.md
│   │   ├── 04-file-storage-infrastructure.md
│   │   ├── 05-messaging-infrastructure.md
│   │   └── 06-external-service-adapters.md
│   ├── 06-api-layer/
│   │   ├── 01-api-layer-overview.md
│   │   ├── 02-goa-dsl-service-definitions.md
│   │   ├── 03-handler-patterns.md
│   │   ├── 04-middleware-chain.md
│   │   ├── 05-request-validation.md
│   │   └── 06-response-serialization.md
│   ├── 07-cqrs-and-read-models/
│   │   ├── 01-cqrs-decision-rationale.md
│   │   ├── 02-command-side-patterns.md
│   │   ├── 03-query-side-patterns.md
│   │   └── 04-read-model-projections.md
│   └── 08-wire-dependency-injection/
│       ├── 01-wire-overview.md
│       ├── 02-provider-sets-per-domain.md
│       ├── 03-dependencies-struct-pattern.md
│       ├── 04-wiring-a-new-domain.md
│       └── 05-wire-troubleshooting.md
│
├── 02-platform-services/
│   ├── 01-tenant-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-provisioning-implementation.md
│   │   ├── 03-lifecycle-management.md
│   │   └── 04-tenant-service-api.md
│   ├── 02-iam-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-casbin-configuration.md
│   │   ├── 03-permission-loading.md
│   │   ├── 04-session-management.md
│   │   └── 05-iam-service-api.md
│   ├── 03-user-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-user-registration-and-invite.md
│   │   ├── 03-profile-management.md
│   │   └── 04-user-service-api.md
│   ├── 04-feature-flag-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-flag-evaluation-implementation.md
│   │   ├── 03-targeting-rules.md
│   │   └── 04-feature-flag-service-api.md
│   ├── 05-notification-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-channel-adapters.md
│   │   ├── 03-template-engine.md
│   │   ├── 04-delivery-guarantees.md
│   │   └── 05-notification-service-api.md
│   ├── 06-workflow-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-temporal-client-usage.md
│   │   ├── 03-workflow-registration.md
│   │   └── 04-workflow-service-api.md
│   ├── 07-file-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-storage-backend-adapters.md
│   │   ├── 03-access-control.md
│   │   ├── 04-virus-scanning.md
│   │   └── 05-file-service-api.md
│   ├── 08-search-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-indexing-pipeline.md
│   │   ├── 03-query-engine.md
│   │   └── 04-search-service-api.md
│   ├── 09-audit-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-event-capture.md
│   │   ├── 03-change-tracking-implementation.md
│   │   └── 04-audit-service-api.md
│   ├── 10-report-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-report-execution-engine.md
│   │   ├── 03-pdf-rendering.md
│   │   └── 04-report-service-api.md
│   ├── 11-integration-service/
│   │   ├── 01-service-contract.md
│   │   ├── 02-connector-framework.md
│   │   ├── 03-webhook-engine.md
│   │   ├── 04-inbound-adapter-pattern.md
│   │   ├── 05-outbound-adapter-pattern.md
│   │   └── 06-integration-service-api.md
│   └── 12-settings-service/
│       ├── 01-service-contract.md
│       ├── 02-settings-hierarchy.md
│       ├── 03-settings-inheritance.md
│       ├── 04-settings-override-model.md
│       └── 05-settings-service-api.md
│
├── 03-database/
│   ├── 01-schema-conventions.md
│   ├── 02-migration-guide.md
│   ├── 03-sqlc-query-authoring.md
│   ├── 04-rls-policy-authoring.md
│   ├── 05-tenant-context-propagation.md
│   ├── 06-indexing-strategy.md
│   ├── 07-query-optimization.md
│   ├── 08-transaction-patterns.md
│   ├── 09-connection-pool-configuration.md
│   └── 10-data-seeding.md
│
├── 04-api-design/
│   ├── 01-api-design-principles.md
│   ├── 02-goa-dsl-guide.md
│   ├── 03-rest-conventions.md
│   ├── 04-sdui-api-endpoints.md
│   ├── 05-pagination-and-filtering.md
│   ├── 06-api-versioning-strategy.md
│   ├── 07-error-response-format.md
│   ├── 08-rate-limiting.md
│   └── 09-api-security-headers.md
│
├── 05-event-architecture/
│   ├── 01-event-architecture-overview.md
│   ├── 02-domain-event-model.md
│   ├── 03-event-publishing-pattern.md
│   ├── 04-event-consumption-pattern.md
│   ├── 05-event-catalogue-authoring.md
│   ├── 06-event-schema-versioning.md
│   ├── 07-outbox-pattern.md
│   └── 08-event-driven-integration.md
│
├── 06-temporal-workflows/
│   ├── 01-workflow-authoring-guide.md
│   ├── 02-activity-authoring-guide.md
│   ├── 03-workflow-state-management.md
│   ├── 04-workflow-testing.md
│   ├── 05-workflow-deployment.md
│   ├── 06-workflow-versioning.md
│   └── 07-temporal-debugging.md
│
├── 07-testing/
│   ├── 01-testing-strategy.md
│   ├── 02-unit-testing-patterns.md
│   ├── 03-integration-testing.md
│   ├── 04-api-testing.md
│   ├── 05-rls-testing.md
│   ├── 06-workflow-testing.md
│   ├── 07-mock-generation.md
│   ├── 08-test-data-management.md
│   └── 09-performance-testing.md
│
└── 08-developer-standards/
    ├── 01-coding-standards.md
    ├── 02-repository-conventions.md
    ├── 03-naming-conventions.md
    ├── 04-error-handling-standards.md
    ├── 05-logging-standards.md
    ├── 06-adr-authoring-process.md
    ├── 07-contribution-guide.md
    ├── 08-pull-request-process.md
    ├── 09-code-review-standards.md
    ├── 10-release-process.md
    ├── 11-testing-standards.md
    └── 12-architecture-governance.md
```

---

### Portal 5 — UI Platform

```
05-ui-platform/
│
├── 01-ui-platform-overview/
│   ├── 01-unified-ui-platform.md
│   ├── 02-shared-pipeline-architecture.md
│   ├── 03-web-vs-mobile-renderer.md
│   ├── 04-sdui-design-contract.md
│   └── 05-ui-platform-versioning.md
│
├── 02-ui-dsl-specification/
│   ├── 01-dsl-overview.md
│   ├── 02-dsl-design-goals.md
│   ├── 03-type-system.md
│   ├── 04-expression-language.md
│   ├── 05-binding-syntax.md
│   ├── 06-event-syntax.md
│   ├── 07-conditional-rendering.md
│   ├── 08-iteration-syntax.md
│   ├── 09-slot-and-composition.md
│   ├── 10-layout-primitives.md
│   ├── 11-theming-and-tokens.md
│   ├── 12-i18n-and-l10n.md
│   ├── 13-accessibility-directives.md
│   ├── 14-dsl-grammar-reference.md
│   └── 15-dsl-changelog.md
│
├── 03-ast-design/
│   ├── 01-ast-overview.md
│   ├── 02-node-types.md
│   ├── 03-widget-nodes.md
│   ├── 04-layout-nodes.md
│   ├── 05-form-nodes.md
│   ├── 06-action-nodes.md
│   ├── 07-data-binding-nodes.md
│   ├── 08-navigation-nodes.md
│   ├── 09-conditional-nodes.md
│   ├── 10-iterator-nodes.md
│   ├── 11-slot-nodes.md
│   └── 12-ast-serialization-format.md
│
├── 04-compiler/
│   ├── 01-compiler-overview.md
│   ├── 02-lexer-and-parser.md
│   ├── 03-semantic-analysis.md
│   ├── 04-type-checking.md
│   ├── 05-optimization-passes.md
│   ├── 06-code-generation.md
│   ├── 07-source-maps.md
│   ├── 08-incremental-compilation.md
│   ├── 09-compiler-plugins.md
│   ├── 10-compiler-errors-reference.md
│   └── 11-compiler-cli-reference.md
│
├── 05-schema-contracts/
│   ├── 01-schema-overview.md
│   ├── 02-page-schema.md
│   ├── 03-widget-schema.md
│   ├── 04-form-schema.md
│   ├── 05-action-schema.md
│   ├── 06-navigation-schema.md
│   ├── 07-data-source-schema.md
│   ├── 08-theme-schema.md
│   ├── 09-permission-schema.md
│   ├── 10-feature-flag-schema.md
│   ├── 11-offline-hint-schema.md
│   ├── 12-schema-versioning.md
│   ├── 13-schema-validation-rules.md
│   └── 14-schema-registry-api.md
│
├── 06-component-system/
│   ├── 01-component-model.md
│   ├── 02-component-registry.md
│   ├── 03-component-lifecycle.md
│   ├── 04-built-in-widgets-reference/
│   │   ├── 01-layout-widgets.md
│   │   ├── 02-display-widgets.md
│   │   ├── 03-input-widgets.md
│   │   ├── 04-navigation-widgets.md
│   │   ├── 05-data-widgets.md
│   │   ├── 06-chart-widgets.md
│   │   ├── 07-media-widgets.md
│   │   └── 08-utility-widgets.md
│   ├── 05-composite-components.md
│   ├── 06-custom-component-authoring.md
│   ├── 07-component-versioning.md
│   └── 08-component-accessibility.md
│
├── 07-form-system/
│   ├── 01-form-architecture.md
│   ├── 02-field-types-reference.md
│   ├── 03-validation-engine.md
│   ├── 04-cross-field-rules.md
│   ├── 05-async-validation.md
│   ├── 06-form-state-model.md
│   ├── 07-multi-step-forms.md
│   ├── 08-dynamic-forms.md
│   ├── 09-form-submission-lifecycle.md
│   ├── 10-form-error-handling.md
│   └── 11-form-accessibility.md
│
├── 08-action-system/
│   ├── 01-action-model.md
│   ├── 02-built-in-actions-reference.md
│   ├── 03-action-chaining.md
│   ├── 04-conditional-actions.md
│   ├── 05-async-actions.md
│   ├── 06-optimistic-actions.md
│   ├── 07-action-error-handling.md
│   ├── 08-custom-action-handlers.md
│   └── 09-action-audit-trail.md
│
├── 09-navigation-system/
│   ├── 01-navigation-model.md
│   ├── 02-screen-registry.md
│   ├── 03-route-definitions.md
│   ├── 04-deep-linking.md
│   ├── 05-navigation-guards.md
│   ├── 06-navigation-history.md
│   ├── 07-tabs-and-drawers.md
│   └── 08-navigation-analytics.md
│
├── 10-web-renderer/
│   ├── 01-web-renderer-overview.md
│   ├── 02-web-renderer-architecture.md
│   ├── 03-component-registry-web.md
│   ├── 04-state-management-web.md
│   ├── 05-web-navigation-integration.md
│   ├── 06-web-offline-support.md
│   ├── 07-web-accessibility.md
│   ├── 08-web-renderer-performance.md
│   └── 09-web-renderer-testing.md
│
└── 11-mobile-renderer/
    ├── 01-mobile-renderer-overview.md
    ├── 02-flutter-renderer-architecture.md
    ├── 03-component-registry-mobile.md
    ├── 04-state-management-mobile.md
    ├── 05-offline-and-sync/
    │   ├── 01-offline-first-strategy.md
    │   ├── 02-local-storage-model.md
    │   ├── 03-sync-engine.md
    │   ├── 04-conflict-resolution-ui.md
    │   └── 05-offline-ui-definitions.md
    ├── 06-platform-integration/
    │   ├── 01-push-notifications.md
    │   ├── 02-biometric-authentication.md
    │   ├── 03-camera-and-scanner.md
    │   ├── 04-file-handling.md
    │   └── 05-device-permissions.md
    ├── 07-mobile-build-and-release/
    │   ├── 01-build-flavors.md
    │   ├── 02-signing-and-distribution.md
    │   └── 03-over-the-air-updates.md
    └── 08-mobile-testing/
        ├── 01-widget-testing.md
        ├── 02-integration-testing.md
        ├── 03-golden-tests.md
        └── 04-e2e-testing.md
```

---

### Portal 6 — Operations & Security

```
06-operations-security/
│
├── 01-deployment/
│   ├── 01-deployment-overview.md
│   ├── 02-infrastructure-requirements.md
│   ├── 03-containerization/
│   │   ├── 01-docker-setup.md
│   │   ├── 02-docker-compose-guide.md
│   │   └── 03-image-hardening.md
│   ├── 04-kubernetes/
│   │   ├── 01-kubernetes-deployment.md
│   │   ├── 02-helm-chart-reference.md
│   │   ├── 03-resource-sizing.md
│   │   ├── 04-horizontal-pod-autoscaling.md
│   │   ├── 05-ingress-configuration.md
│   │   └── 06-namespace-strategy.md
│   ├── 05-data-tier-deployment/
│   │   ├── 01-postgresql-setup.md
│   │   ├── 02-connection-pooling.md
│   │   ├── 03-redis-setup.md
│   │   ├── 04-backup-and-restore.md
│   │   └── 05-replication-setup.md
│   ├── 06-temporal-cluster-setup.md
│   ├── 07-schema-registry-deployment.md
│   ├── 08-ci-cd-pipeline/
│   │   ├── 01-pipeline-overview.md
│   │   ├── 02-build-pipeline.md
│   │   ├── 03-test-pipeline.md
│   │   ├── 04-mobile-build-pipeline.md
│   │   ├── 05-deployment-pipeline.md
│   │   └── 06-rollback-procedures.md
│   └── 09-environment-management/
│       ├── 01-environment-tiers.md
│       ├── 02-configuration-management.md
│       ├── 03-secrets-management.md
│       └── 04-environment-promotion.md
│
├── 02-iam-operations/
│   ├── 01-iam-operations-overview.md
│   ├── 02-casbin-policy-management.md
│   ├── 03-role-and-permission-administration.md
│   ├── 04-sso-configuration.md
│   ├── 05-mfa-administration.md
│   ├── 06-api-key-management.md
│   ├── 07-session-administration.md
│   └── 08-iam-audit-reporting.md
│
├── 03-security/
│   ├── 01-security-overview.md
│   ├── 02-security-hardening-guide.md
│   ├── 03-network-security.md
│   ├── 04-tls-certificate-management.md
│   ├── 05-secrets-rotation.md
│   ├── 06-vulnerability-management.md
│   ├── 07-penetration-testing-guide.md
│   ├── 08-compliance/
│   │   ├── 01-data-privacy-guide.md
│   │   ├── 02-gdpr-compliance.md
│   │   ├── 03-audit-trail-requirements.md
│   │   └── 04-data-retention-policy.md
│   ├── 09-mobile-security/
│   │   ├── 01-mobile-app-hardening.md
│   │   ├── 02-certificate-pinning.md
│   │   ├── 03-local-data-encryption.md
│   │   └── 04-jailbreak-detection.md
│   └── 10-incident-response/
│       ├── 01-incident-response-plan.md
│       ├── 02-security-incident-classification.md
│       └── 03-breach-notification-procedure.md
│
├── 04-observability/
│   ├── 01-observability-setup.md
│   ├── 02-logging/
│   │   ├── 01-log-configuration.md
│   │   ├── 02-log-aggregation.md
│   │   ├── 03-log-querying.md
│   │   └── 04-log-retention-policy.md
│   ├── 03-metrics/
│   │   ├── 01-prometheus-setup.md
│   │   ├── 02-metrics-reference.md
│   │   ├── 03-grafana-dashboards.md
│   │   └── 04-custom-metrics.md
│   ├── 04-tracing/
│   │   ├── 01-opentelemetry-setup.md
│   │   ├── 02-trace-sampling-strategy.md
│   │   └── 03-distributed-trace-analysis.md
│   └── 05-alerting/
│       ├── 01-alert-rules-reference.md
│       ├── 02-alert-routing.md
│       ├── 03-on-call-setup.md
│       └── 04-alert-runbook-index.md
│   └── 06-pipeline-observability.md/
│       ├── 06-pipeline-observability.md    ◄ NEW
      # StageCompleted / StageSkipped / TxHookFailed domain events
      # operation_logs table as the primary audit surface
      # Diagnostic SQL queries (§22.1): full stage log, pending approvals,
        flag check for skipped stage
      # workflow_trigger_queue failure SQL (§22.2)
      # Common error reference table (§22.3):
        stage failed after retries / TxHook failed / double-resume /
        script timeout / field not accessible / function not registered /
        task not appearing / cache stale
│
├── 05-performance/
│   ├── 01-performance-baseline.md
│   ├── 02-load-testing-guide.md
│   ├── 03-backend-performance-tuning.md
│   ├── 04-database-performance-tuning.md
│   ├── 05-mobile-performance-tuning.md
│   ├── 06-sdui-rendering-performance.md
│   └── 07-sync-performance-tuning.md
│
└── 06-runbooks/
    ├── 01-runbook-index.md
    ├── 02-backend-runbooks/
    │   ├── RB-001-api-high-latency.md
    │   ├── RB-002-database-connection-exhaustion.md
    │   ├── RB-003-redis-outage.md
    │   ├── RB-004-temporal-worker-down.md
    │   ├── RB-005-schema-registry-unavailable.md
    │   ├── RB-006-tenant-provisioning-failure.md
    │   ├── RB-007-rls-bypass-detected.md
    │   └── RB-008-workflow-stuck.md
    ├── 03-mobile-runbooks/
    │   ├── RB-101-sdui-schema-failure.md
    │   ├── RB-102-sync-engine-stuck.md
    │   ├── RB-103-offline-queue-overflow.md
    │   └── RB-104-mobile-crash-spike.md
    └── 04-business-continuity/
        ├── 01-disaster-recovery-plan.md
        ├── 02-backup-verification-procedure.md
        └── 03-failover-procedure.md
```

---

### Portal 7 — ERP Modules

Every module follows the mandatory 12-section template defined in `_template/`. Fuel Station is a super-module with four sub-modules.

```
07-erp-modules/
│
├── _template/
│   ├── 00-module-template-guide.md
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 01-finance/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md           # Chart of Accounts, Journal, Invoice, Payment, etc.
│   ├── 05-workflows.md                # Invoice approval, payment release, period close
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md                  # P&L, Balance Sheet, Cash Flow, AR Aging, AP Aging
│   ├── 11-integrations.md             # Accounting system connectors, bank feeds
│   └── 12-extension-points.md
│
├── 02-crm/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 03-inventory/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 04-procurement/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 05-sales/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 06-manufacturing/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 07-hr/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 08-projects/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 09-asset-management/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 10-fuel-station/                    # Strategic super-module
│   ├── 00-fuel-station-platform-overview.md
│   │
│   ├── 10a-station-operations/
│   │   ├── 01-business-overview.md
│   │   ├── 02-concepts.md
│   │   ├── 03-entities.md             # Station, Forecourt, Pump, Nozzle, Tank, Price Board
│   │   ├── 04-document-types.md       # Shift report, delivery note, stock transfer
│   │   ├── 05-workflows.md            # Shift open/close, delivery receive, price change
│   │   ├── 06-permissions.md
│   │   ├── 07-pump-management.md
│   │   ├── 08-price-management.md
│   │   ├── 09-shift-management.md
│   │   ├── 10-cash-management.md
│   │   ├── 11-fleet-accounts.md
│   │   ├── 12-lubricants.md
│   │   ├── 13-apis.md
│   │   ├── 14-mobile-screens.md
│   │   ├── 15-ui-dsl-patterns.md
│   │   ├── 16-reports.md              # Daily sales, shift reconciliation, variance
│   │   ├── 17-integrations.md         # POS systems, fleet card processors, forecourt controllers
│   │   └── 18-extension-points.md
│   │
│   ├── 10b-wetstock/
│   │   ├── 01-business-overview.md
│   │   ├── 02-concepts.md
│   │   ├── 03-entities.md             # Tank, Dip, Delivery, Variance
│   │   ├── 04-document-types.md
│   │   ├── 05-workflows.md
│   │   ├── 06-permissions.md
│   │   ├── 07-tank-management.md
│   │   ├── 08-dip-and-reconciliation.md
│   │   ├── 09-delivery-management.md
│   │   ├── 10-variance-analysis.md
│   │   ├── 11-regulatory-compliance.md
│   │   ├── 12-atg-integration.md
│   │   ├── 13-forecourt-controller-integration.md
│   │   ├── 14-apis.md
│   │   ├── 15-mobile-screens.md
│   │   ├── 16-ui-dsl-patterns.md
│   │   ├── 17-reports.md
│   │   └── 18-extension-points.md
│   │
│   ├── 10c-meter-reading/
│   │   ├── 01-business-overview.md
│   │   ├── 02-concepts.md
│   │   ├── 03-entities.md
│   │   ├── 04-document-types.md
│   │   ├── 05-workflows.md
│   │   ├── 06-permissions.md
│   │   ├── 07-meter-registration.md
│   │   ├── 08-reading-schedules.md
│   │   ├── 09-reading-entry-and-validation.md
│   │   ├── 10-variance-alerts.md
│   │   ├── 11-apis.md
│   │   ├── 12-mobile-screens.md
│   │   ├── 13-ui-dsl-patterns.md
│   │   ├── 14-reports.md
│   │   └── 15-extension-points.md
│   │
│   └── 10d-lpg/
│       ├── 01-business-overview.md
│       ├── 02-concepts.md
│       ├── 03-entities.md             # Cylinder, Manifold, Bulk Tank, Customer
│       ├── 04-document-types.md       # Cylinder exchange, bulk delivery, customer account
│       ├── 05-workflows.md
│       ├── 06-permissions.md
│       ├── 07-cylinder-management.md
│       ├── 08-bulk-lpg-management.md
│       ├── 09-customer-accounts.md
│       ├── 10-delivery-management.md
│       ├── 11-apis.md
│       ├── 12-mobile-screens.md
│       ├── 13-ui-dsl-patterns.md
│       ├── 14-reports.md
│       └── 15-extension-points.md
│
├── 11-approval-workflows/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-workflows.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-reports.md
│   ├── 11-integrations.md
│   └── 12-extension-points.md
│
├── 12-dashboards/
│   ├── 01-business-overview.md
│   ├── 02-concepts.md
│   ├── 03-entities.md
│   ├── 04-document-types.md
│   ├── 05-widgets-catalogue.md
│   ├── 06-permissions.md
│   ├── 07-apis.md
│   ├── 08-mobile-screens.md
│   ├── 09-ui-dsl-patterns.md
│   ├── 10-role-based-dashboard-defaults.md
│   ├── 11-user-customization.md
│   └── 12-extension-points.md
│
└── 13-reporting/
    ├── 01-business-overview.md
    ├── 02-concepts.md
    ├── 03-entities.md
    ├── 04-standard-report-catalogue.md
    ├── 05-workflows.md
    ├── 06-permissions.md
    ├── 07-apis.md
    ├── 08-mobile-report-viewer.md
    ├── 09-ui-dsl-patterns.md
    ├── 10-custom-report-authoring.md
    ├── 11-report-scheduling.md
    └── 12-extension-points.md
```

---

### Portal 8 — SDK & Extension Framework

```
08-sdk-extensions/
│
├── 01-extension-model/
│   ├── 01-extension-model-overview.md
│   ├── 02-extension-points-catalogue.md
│   ├── 03-extension-security-model.md
│   └── 04-extension-lifecycle.md
│
├── 02-go-server-sdk/
│   ├── 01-installation.md
│   ├── 02-getting-started.md
│   ├── 03-ui-dsl-builder-api.md
│   ├── 04-schema-compiler-api.md
│   ├── 05-authorization-api.md
│   ├── 06-audit-api.md
│   ├── 07-workflow-api.md
│   └── 08-go-sdk-reference.md
│
├── 03-flutter-client-sdk/
│   ├── 01-installation.md
│   ├── 02-getting-started.md
│   ├── 03-renderer-api.md
│   ├── 04-component-registry-api.md
│   ├── 05-action-handler-api.md
│   ├── 06-sync-engine-api.md
│   └── 07-flutter-sdk-reference.md
│
├── 04-dsl-sdk/
│   ├── 01-compiler-api.md
│   ├── 02-linter-api.md
│   ├── 03-preview-server-api.md
│   └── 04-dsl-ide-extensions.md
│
├── 05-plugin-system/
│   ├── 01-plugin-architecture.md
│   ├── 02-plugin-authoring-guide.md
│   ├── 03-plugin-manifest.md
│   ├── 04-plugin-lifecycle.md
│   ├── 05-plugin-distribution.md
│   └── 06-plugin-security-model.md
│
├── 06-custom-components/
│   ├── 01-custom-component-sdk.md
│   ├── 02-component-manifest.md
│   ├── 03-component-testing.md
│   └── 04-component-publishing.md
│
├── 07-custom-actions/
│   ├── 01-custom-action-sdk.md
│   ├── 02-action-registration.md
│   └── 03-action-testing.md
│
└── 08-integration-adapters/
    ├── 01-adapter-framework-overview.md
    ├── 02-inbound-adapter-sdk.md
    ├── 03-outbound-adapter-sdk.md
    ├── 04-webhook-handler-sdk.md
    ├── 05-authentication-for-adapters.md
    └── 06-adapter-testing.md
```

---

### Portal 9 — Reference

```
09-reference/
│
├── 01-api-reference/
│   ├── 01-api-overview.md
│   ├── 02-authentication.md
│   ├── 03-tenant-api.md
│   ├── 04-iam-api.md
│   ├── 05-user-api.md
│   ├── 06-schema-registry-api.md
│   ├── 07-feature-flag-api.md
│   ├── 08-workflow-api.md
│   ├── 09-audit-api.md
│   ├── 10-notification-api.md
│   ├── 11-file-api.md
│   ├── 12-search-api.md
│   ├── 13-report-api.md
│   ├── 14-settings-api.md
│   └── 15-domain-apis/
│       ├── finance-api.md
│       ├── crm-api.md
│       ├── inventory-api.md
│       ├── procurement-api.md
│       ├── sales-api.md
│       ├── manufacturing-api.md
│       ├── hr-api.md
│       ├── projects-api.md
│       ├── asset-management-api.md
│       ├── fuel-station-api.md
│       ├── wetstock-api.md
│       ├── meter-reading-api.md
│       └── lpg-api.md
│
├── 02-schema-catalogue/
│   ├── 01-schema-catalogue-index.md
│   ├── 02-page-schema-v1.json
│   ├── 03-widget-schema-v1.json
│   ├── 04-form-schema-v1.json
│   ├── 05-action-schema-v1.json
│   ├── 06-navigation-schema-v1.json
│   └── 07-theme-schema-v1.json
│
├── 03-permission-catalogue/
│   ├── 01-permission-catalogue-index.md
│   ├── 02-platform-permissions.md
│   ├── 03-finance-permissions.md
│   ├── 04-inventory-permissions.md
│   ├── 05-procurement-permissions.md
│   ├── 06-sales-permissions.md
│   ├── 07-hr-permissions.md
│   ├── 08-fuel-station-permissions.md
│   └── 09-admin-permissions.md
│
├── 04-error-codes/
│   ├── 01-error-code-format.md
│   ├── 02-platform-error-codes.md
│   ├── 03-domain-error-codes.md
│   └── 04-mobile-error-codes.md
│
├── 05-event-catalogue/
│   ├── 01-event-catalogue-index.md
│   ├── 02-platform-events.md
│   ├── 03-finance-events.md
│   ├── 04-inventory-events.md
│   ├── 05-fuel-station-events.md
│   └── 06-workflow-events.md
│   └── 07-pipeline-events.md    ◄ NEW
          ##  StageCompleted, StageSkipped, StageCompleted, TxHookFailed,
              CompensationFailed, OperationSuspended, OperationResumed,
              OperationCompleted, OperationFailed│
├── 06-migration-guides/
│   ├── 01-migration-overview.md
│   ├── 02-schema-migration-guides.md
│   ├── 03-api-migration-guides.md
│   ├── 04-mobile-sdk-migration-guides.md
│   ├── 05-database-migration-procedures.md
│   └── 06-dsl-migration-guides.md
│
├── 07-release-management/
│   ├── 01-release-process.md
│   ├── 02-versioning-policy.md
│   ├── 03-branching-model.md
│   ├── 04-hotfix-process.md
│   ├── 05-deprecation-policy.md
│   ├── 06-changelog.md
│   └── 07-known-issues.md
│
└── 08-troubleshooting/
    ├── 01-troubleshooting-guide.md
    ├── 02-common-sdui-issues.md
    ├── 03-common-sync-issues.md
    ├── 04-common-auth-issues.md
    ├── 05-common-mobile-issues.md
    ├── 06-common-workflow-issues.md
    ├── 07-diagnostic-tools.md
    ├── 08-log-analysis-guide.md
    └── 09-support-escalation-guide.md
```

---

## Summary of Changes from v1 → v2

| Dimension | v1 | v2 |
|-----------|----|----|
| Portals | 7 | 9 |
| Documents (approx.) | ~430 | ~680 |
| ERP Platform Kernel | absent | 10 framework sections in Portal 3 |
| Metadata Architecture | absent | 11 sections in Portal 3 |
| Go Clean Architecture | absent | 8 sections in Portal 4 |
| Platform Services | scattered | 12 dedicated service contracts in Portal 4 |
| IAM depth | generic | Full Casbin model, org hierarchy, SoD, delegation in Portal 3 |
| Multi-tenancy | thin | RLS, provisioning, lifecycle, migrations, cross-tenant admin in Portal 3 |
| UI Platform | mobile-only | Unified pipeline + Web Renderer + Mobile Renderer in Portal 5 |
| ERP Module template | inconsistent | Mandatory 12-section template enforced across all modules |
| Fuel Station | single module | Strategic super-module with 4 sub-modules (16 sub-sections) |
| Developer Standards | absent | 12 sections in Portal 4 |
| Event Architecture | absent | 8 sections in Portal 4 |
| Workflow Framework | scattered | Consolidated 5-section framework in Portal 3 |

Awaiting your approval before any documentation content is written.
