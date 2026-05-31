# Operation Pipeline — Complete Reference

> **Package:** `awo/internal/pipeline`
> **Audience:** Platform engineers and module developers.
> **Scope:** A configuration-aware, feature-flag-driven pipeline engine for complex multi-service operations. Covers the full lifecycle: architecture, stage design, hook system, workflow integration, saga pattern, midway failure handling and compensation, performance optimisation, API handler design, UI patterns, tenant scripting with `pkg/condition`, dry-run mode, and a full synthesis with the Customer Workflow Engine.

---

## Table of Contents

1. [The Problem This Solves](#1-the-problem-this-solves)
   - 1.1 The Bloated Function Anti-Pattern
   - 1.2 What the Pipeline Architecture Provides
2. [Core Concepts](#2-core-concepts)
   - 2.1 Concepts at a Glance
   - 2.2 Architecture Diagram
3. [OperationContext — The Carrier](#3-operationcontext--the-carrier)
   - 3.1 Design
   - 3.2 Stage Communication Protocol
4. [Stage — The Unit of Work](#4-stage--the-unit-of-work)
   - 4.1 Stage Interface
   - 4.2 Base Stage Implementation
   - 4.3 Concrete Stage Example — Budget Check
   - 4.4 Stage Priority Bands
5. [Pipeline & PipelineBuilder](#5-pipeline--pipelinebuilder)
   - 5.1 The Pipeline
   - 5.2 PipelineBuilder
   - 5.3 Operation Log Persistence
6. [Hook System — The Extension Point](#6-hook-system--the-extension-point)
   - 6.1 Hook Interface
   - 6.2 How Hooks Enable Workflow Without Core Changes
   - 6.3 Other Hook Examples
7. [Stage & Hook Registries](#7-stage--hook-registries)
   - 7.1 Registry Design
   - 7.2 Registration at Startup (wire.go)
8. [Canonical Example: AP Invoice Processing](#8-canonical-example-ap-invoice-processing)
   - 8.1 The Full Pipeline Definition
   - 8.2 Service Entry Point
   - 8.3 What Each Tenant Sees
9. [Midway Failure — What Actually Happens](#9-midway-failure--what-actually-happens)
   - 9.1 Failure Taxonomy
   - 9.2 The Compensation Registry
   - 9.3 Saga Rollback Algorithm
   - 9.4 Idempotency Keys
   - 9.5 Partial Completion Recovery
10. [Long-Running Operations — Saga Pattern](#10-long-running-operations--saga-pattern)
    - 10.1 When an Operation Spans Days
    - 10.2 The Saga Coordinator
    - 10.3 Compensation (Undo) Logic
11. [Workflow Engine Integration](#11-workflow-engine-integration)
    - 11.1 Workflow as a Pure Hook Consumer
    - 11.2 Pipeline Resume After Approval
    - 11.3 Approval Task Lifecycle
12. [Performance Optimisation](#12-performance-optimisation)
    - 12.1 Pre-Built Pipeline Cache
    - 12.2 Parallel Stage Execution
    - 12.3 Stage Result Memoisation
    - 12.4 Context Object Pooling
    - 12.5 Database Access Patterns
13. [API Handler Design](#13-api-handler-design)
    - 13.1 Synchronous vs Asynchronous Response
    - 13.2 Status Polling Endpoint
    - 13.3 WebSocket Real-time Updates
    - 13.4 Idempotency at the HTTP Layer
    - 13.5 Error Response Schema
14. [UI Design — Operation Dashboard](#14-ui-design--operation-dashboard)
    - 14.1 Pipeline Visualiser (Live)
    - 14.2 Dry-Run Preview UI
15. [Tenant Scripting — The condition Package](#15-tenant-scripting--the-condition-package)
    - 15.1 Where condition Fits in the Pipeline
    - 15.2 Script Stage — Tenant-Defined Logic
    - 15.3 Condition-Gated Stage Execution
    - 15.4 Custom Hook Conditions
    - 15.5 Security Sandbox
    - 15.6 The condition Package API Reference
16. [Dry-Run Mode](#16-dry-run-mode)
    - 16.1 How Dry-Run Works
    - 16.2 Stage Dry-Run Contract
    - 16.3 Dry-Run API
17. [Customer Workflow Engine — Synthesis](#17-customer-workflow-engine--synthesis)
    - 17.1 Two Complementary Systems
    - 17.2 What We Borrow from the Workflow Engine
    - 17.3 The Unified Trigger Model
    - 17.4 Workflow Engine as a Pipeline Stage
    - 17.5 Shared Schema Between Both Systems
    - 17.6 What Stays Separate
18. [All Planned Modules & Their Stage Contributions](#18-all-planned-modules--their-stage-contributions)
19. [Adding a New Stage or Module](#19-adding-a-new-stage-or-module)
    - 19.1 The Contract for New Module Developers
    - 19.2 Example: Adding the DMS Module Later
    - 19.3 Adding the Future Workflow Module
20. [Testing Strategy](#20-testing-strategy)
    - 20.1 Stage Tests (Isolated — No Pipeline)
    - 20.2 Pipeline Tests (Integration)
    - 20.3 Context Isolation Tests
21. [Troubleshooting](#21-troubleshooting)
    - 21.1 Why Was a Stage Skipped?
    - 21.2 Pipeline Suspended — Find All Pending
    - 21.3 Stage Registered But Not Running
    - 21.4 Hook Not Firing
    - 21.5 Common Error Reference

---

## 1. The Problem This Solves

### 1.1 The Bloated Function Anti-Pattern

A naive implementation of posting an AP invoice looks like this:

```go
func (s *APService) ProcessInvoice(ctx context.Context, inv *Invoice) error {
    session := domain.SessionFromContext(ctx)

    // Duplicate check
    if err := s.checkDuplicate(ctx, inv); err != nil { return err }

    // Three-way matching
    if session.FeatureEnabled("procurement") {
        if err := s.threeWayMatch(ctx, inv); err != nil { return err }
    }

    // Budget check
    if session.FeatureEnabled("budget") {
        if err := s.budgetSvc.Check(ctx, inv); err != nil { return err }
    }

    // Tax calculation
    if session.FeatureEnabled("tax") {
        tax, err := s.taxSvc.Calculate(ctx, inv)
        if err != nil { return err }
        inv.TaxAmount = tax
    }

    // Workflow approval
    if session.FeatureEnabled("workflow") {
        approved, err := s.workflowSvc.Submit(ctx, inv)
        if err != nil { return err }
        if !approved { return ErrPendingApproval }
    }

    // GL posting (always)
    if err := s.glSvc.Post(ctx, inv); err != nil { return err }

    // DMS archival
    if session.FeatureEnabled("dms") {
        s.dmsSvc.Archive(ctx, inv)
    }

    // Banking/payment scheduling
    if session.FeatureEnabled("banking") {
        s.bankingSvc.Schedule(ctx, inv)
    }

    // Notifications
    if session.FeatureEnabled("notifications") {
        s.notifySvc.Send(ctx, inv)
    }

    // Audit
    s.auditSvc.Log(ctx, inv)

    return nil
}
```

This function is broken in multiple ways:

- **Untestable in isolation.** Every test must mock all services, even ones being tested for other things.
- **Cannot be extended without modifying the function.** Adding the future e-invoicing module requires touching `ProcessInvoice`.
- **Workflow approval is wrong.** `ProcessInvoice` returns an error when an approval is pending, losing all the work done before it. The invoice must be suspended mid-flight, not aborted.
- **Order matters but is implicit.** Why does tax come before workflow? Why does GL come before DMS? The ordering rule is in the developer's head, not in the code.
- **No audit trail of the pipeline itself.** You know the invoice was processed but not which stages ran, which were skipped, and what each decided.
- **Future modules have no hook point.** A workflow module added in six months has no way to insert itself without modifying every complex operation in the system.

### 1.2 What the Pipeline Architecture Provides

```
Every complex business operation becomes:

  1. BuildPipeline(operation, session)
       → reads feature flags, settings, user preferences
       → returns a pipeline with only the applicable stages in the correct order

  2. pipeline.Execute(operationContext)
       → each stage reads from and writes to shared context
       → each stage is skipped, run, or deferred independently
       → hooks fire before and after each stage
       → the workflow module inserts itself via hooks without any core changes
       → the full execution trace is captured in the operation log

  3. Result
       → typed result with stage outcomes
       → complete audit trail
       → if pending approval: operation is persisted and resumed later
```

---

## 2. Core Concepts

### 2.1 Concepts at a Glance

```
OperationContext   The carrier. Holds session (flags, settings, permissions),
                   shared mutable state between stages, and the execution log.
                   Every stage reads from it and writes to it.

Stage              A single, independently testable unit of work.
                   Declares its own feature flag gate, priority, and whether
                   it is required or optional.

Pipeline           An ordered slice of stages, configured at runtime based on
                   which stages are registered and which flags are enabled.

PipelineBuilder    Constructs a pipeline for a given operation by filtering
                   the registry of all known stages through the session's flags.

StageRegistry      The global catalogue of all stages across all modules.
                   Modules register their stages here at startup (wire.go).

Hook               An extension point that fires before or after a named stage.
                   Future modules register hooks here to insert behavior without
                   modifying the core operation.

HookRegistry       The global catalogue of all hooks. Modules register hooks
                   at startup. The pipeline calls all matching hooks in priority order.

OperationLog       The complete record of what happened: which stages ran, which
                   were skipped, what each decided, and how long each took.
                   Persisted to the database for audit and debugging.
```

### 2.2 Architecture Diagram

```
Module startup (wire.go)
  │
  ├── StageRegistry.Register(BudgetCheckStage{})       ← budget module
  ├── StageRegistry.Register(TaxCalculationStage{})    ← tax module
  ├── HookRegistry.Register(WorkflowApprovalHook{})   ← workflow module (future)
  ├── HookRegistry.Register(AuditLogHook{})            ← audit module
  └── HookRegistry.Register(NotificationHook{})        ← notification module

                          ↓
At operation time:
  PipelineBuilder.Build("ap.invoice.process", session)
    │
    │ reads StageRegistry → filters by operation + feature flags
    │ reads HookRegistry  → filters by feature flags
    │
    ▼
  Configured Pipeline:
    [ValidateInvoice]      always       → no flag gate
    [DuplicateCheck]       always       → no flag gate
    [ThreeWayMatch]        optional     → gate: "procurement"
    [BudgetCheck]          optional     → gate: "budget"
    [TaxCalculation]       optional     → gate: "tax"
       ↕ Hook: "before:gl_posting"     → WorkflowApprovalHook (gate: "workflow")
    [GLPosting]            always       → no flag gate
    [DMSArchive]           optional     → gate: "dms"
    [PaymentScheduling]    optional     → gate: "banking"
       ↕ Hook: "after:payment_scheduling" → NotificationHook (always)
    [AuditLog]             always       → no flag gate
    │
    ▼
  OperationResult {
    Status:  "completed" | "pending_approval" | "failed"
    Outputs: map[string]any  (typed data produced by stages)
    Log:     []StageLog      (full audit trail)
  }
```

---

## 3. OperationContext — The Carrier

### 3.1 Design

The `OperationContext` is not the standard `context.Context`. It is a domain-level object that carries everything any stage might need. It is passed by pointer so stages can write to the shared state:

```go
// awo/internal/pipeline/context.go

// OperationContext carries all inputs, shared state, and the execution log
// for a single pipeline run. Passed by pointer through every stage and hook.
type OperationContext struct {
    // ── Identity ─────────────────────────────────────────────────────────
    Ctx        context.Context          // standard context (cancellation, deadlines)
    Session    *domain.ResolvedSession  // permissions, flags, settings, preferences
    TenantID   uuid.UUID
    UserID     uuid.UUID
    EntityID   uuid.UUID

    // ── Operation identity ───────────────────────────────────────────────
    OperationID   uuid.UUID   // unique ID for this execution (for saga resume)
    OperationKey  string      // "ap.invoice.process"
    Resource      string      // "ap.invoice"
    Action        string      // "process"

    // ── Input ────────────────────────────────────────────────────────────
    Input any   // typed by the caller; stages cast to what they expect

    // ── Shared mutable state between stages ──────────────────────────────
    // Stages read from and write to this map.
    // Keys follow {stage_name}.{key} convention to avoid collision:
    //   "tax_calculation.tax_amount" → decimal.Decimal
    //   "budget_check.approved_up_to" → decimal.Decimal
    //   "three_way_match.po_id" → uuid.UUID
    //   "workflow.approval_id" → uuid.UUID
    Data map[string]any

    // ── Decision flags ───────────────────────────────────────────────────
    // Set by stages to influence subsequent stage behavior.
    // Stages read these to make decisions without coupling to other stages.
    Flags map[string]bool
    // e.g. "requires_approval": true, "tax_exempt": false, "within_budget": true

    // ── Suspension (for workflow approval, async operations) ─────────────
    Suspended     bool
    SuspendReason string  // "pending_workflow_approval:approval_uuid"
    ResumePoint   string  // stage name to resume from after approval

    // ── Dry-run mode ─────────────────────────────────────────────────────
    DryRun bool  // when true, all stages operate in simulation mode

    // ── Execution log ────────────────────────────────────────────────────
    Log         []StageLog
    StartedAt   time.Time
    CompletedAt *time.Time
}

type StageLog struct {
    StageName   string
    Status      string    // "ran" | "skipped" | "suspended" | "failed"
    SkipReason  string    // populated when status = "skipped"
    StartedAt   time.Time
    Duration    time.Duration
    Output      map[string]any  // what this stage produced (for debugging)
    Error       string
}

// Convenience accessors
func (o *OperationContext) GetData(key string) (any, bool) {
    v, ok := o.Data[key]
    return v, ok
}

func (o *OperationContext) SetData(key string, value any) {
    if o.Data == nil { o.Data = make(map[string]any) }
    o.Data[key] = value
}

func (o *OperationContext) SetFlag(key string, value bool) {
    if o.Flags == nil { o.Flags = make(map[string]bool) }
    o.Flags[key] = value
}

func (o *OperationContext) Flag(key string) bool {
    return o.Flags[key]
}

// Suspend halts the pipeline at this point. It will be resumed from ResumeAt
// when the suspension condition is resolved (e.g., workflow approved).
func (o *OperationContext) Suspend(reason, resumeAt string) {
    o.Suspended = true
    o.SuspendReason = reason
    o.ResumePoint = resumeAt
}

// Can checks if the current session has a permission.
func (o *OperationContext) Can(resource, action string) bool {
    return o.Session.Can(resource, action)
}

// FeatureEnabled checks if a feature flag is on for this tenant.
func (o *OperationContext) FeatureEnabled(key string) bool {
    return o.Session.FeatureEnabled(key)
}

// Setting returns a tenant setting value with a typed helper.
func (o *OperationContext) SettingDecimal(key string, def decimal.Decimal) decimal.Decimal {
    return o.Session.SettingDecimal(key, def)
}

func (o *OperationContext) SettingBool(key string, def bool) bool {
    return o.Session.SettingBool(key, def)
}

func (o *OperationContext) SettingString(key string, def string) string {
    return o.Session.SettingString(key, def)
}
```

### 3.2 Stage Communication Protocol

Stages communicate exclusively through the `OperationContext`. Direct service-to-service calls within a pipeline are forbidden:

```
DON'T (tight coupling between stages):
  func (s *GLPostingStage) Execute(ctx *OperationContext) (StageResult, error) {
      // directly reading from another service
      tax, _ := taxService.GetCalculated(ctx.Ctx, invoiceID)
      ...
  }

DO (read from shared context):
  func (s *GLPostingStage) Execute(ctx *OperationContext) (StageResult, error) {
      // read what the tax stage wrote to context
      taxAmount, _ := ctx.GetData("tax_calculation.tax_amount")
      ...
  }
```

**Data key convention:**

```
"{stage_name}.{output_key}"

Examples:
  "tax_calculation.tax_amount"       → decimal.Decimal
  "tax_calculation.tax_code"         → string (e.g. "VAT16")
  "tax_calculation.is_exempt"        → bool
  "budget_check.approved"            → bool
  "budget_check.available_amount"    → decimal.Decimal
  "three_way_match.matched_po_id"    → uuid.UUID
  "three_way_match.quantity_variance"→ decimal.Decimal
  "workflow.approval_id"             → uuid.UUID
  "workflow.approved_by"             → uuid.UUID
  "gl_posting.transaction_id"        → uuid.UUID
  "dms_archive.document_id"          → uuid.UUID
  "payment_scheduling.payment_id"    → uuid.UUID
```

---

## 4. Stage — The Unit of Work

### 4.1 Stage Interface

```go
// awo/internal/pipeline/stage.go

// Stage is a single independently-testable unit of work in a pipeline.
// Each stage declares its own configuration: which operation it applies to,
// which feature flag gates it, its priority in the execution order, and
// whether it is required or can be skipped.
type Stage interface {
    // Name returns the unique identifier for this stage across all modules.
    // Format: "{module}.{stage_description}"
    // Examples: "ap.validate_invoice", "tax.calculate", "budget.check"
    Name() string

    // Operations returns the list of operation keys this stage applies to.
    // "*" means it applies to all operations (e.g., audit log stage).
    // Format: "{module}.{resource}.{action}"
    Operations() []string

    // FeatureFlag returns the flag key that gates this stage.
    // Empty string means the stage always runs (no gate).
    // The pipeline skips this stage if FeatureFlag() returns a key
    // and the session does not have that flag enabled.
    FeatureFlag() string

    // Priority determines execution order within the pipeline.
    // Lower numbers execute first.
    // Standard priority bands:
    //   100-199: Validation
    //   200-299: Enrichment (fetch related data, resolve references)
    //   300-399: Business rule checks (budget, compliance, limits)
    //   400-499: Tax and regulatory
    //   500-599: Workflow and approval gates
    //   600-699: Core financial posting (GL, AR, AP)
    //   700-799: Secondary financial (DMS, banking, AR offset)
    //   800-899: Notifications and events
    //   900-999: Audit and compliance logging
    Priority() int

    // Required returns whether this stage must succeed for the operation
    // to proceed. If true and Execute returns an error, the pipeline aborts.
    // If false and Execute returns an error, the error is logged and
    // execution continues (best-effort stage).
    Required() bool

    // DependsOn returns stage names this stage reads data from.
    // Empty slice means the stage has no intra-pipeline data dependencies.
    // The pipeline uses this to identify which stages in the same priority
    // band can run in parallel.
    DependsOn() []string

    // RunCondition returns a condition formula (expr-lang) that must evaluate
    // to true for this stage to execute. Empty string = always run.
    // The condition has access to the OperationContext's Data and Flags.
    RunCondition() string

    // Execute performs the stage's work.
    // Returns a StageResult describing what the stage did and any outputs.
    // The stage writes its outputs to opCtx.Data using the key convention.
    Execute(opCtx *OperationContext) (StageResult, error)
}

type StageResult struct {
    Status      string      // "completed" | "skipped" | "suspended" | "deferred" | "simulated"
    Outputs     map[string]any
    Message     string
    NextStageID string      // used by ScriptStage to route execution
}
```

### 4.2 Base Stage Implementation

A `BaseStage` provides sensible defaults so concrete stages only override what they need:

```go
// awo/internal/pipeline/stage.go

type BaseStage struct {
    name        string
    operations  []string
    featureFlag string
    priority    int
    required    bool
}

func (b *BaseStage) Name() string         { return b.name }
func (b *BaseStage) Operations() []string { return b.operations }
func (b *BaseStage) FeatureFlag() string  { return b.featureFlag }
func (b *BaseStage) Priority() int        { return b.priority }
func (b *BaseStage) Required() bool       { return b.required }
func (b *BaseStage) DependsOn() []string  { return []string{} }
func (b *BaseStage) RunCondition() string { return "" }
```

### 4.3 Concrete Stage Example — Budget Check

```go
// awo/internal/budget/pipeline_stages.go

// BudgetCheckStage checks that the invoice amount does not exceed
// the available budget for the relevant cost center.
// Gate: "budget" feature flag.
// Priority: 310 (after validation and enrichment, before tax and workflow).
type BudgetCheckStage struct {
    BaseStage
    budgetSvc BudgetService
}

func NewBudgetCheckStage(svc BudgetService) *BudgetCheckStage {
    return &BudgetCheckStage{
        BaseStage: BaseStage{
            name:        "budget.check",
            operations:  []string{"ap.invoice.process", "ap.invoice.approve"},
            featureFlag: "budget",     // skipped if budget module not enabled
            priority:    310,
            required:    false,        // budget warning does not block posting
            // behaviour is driven by settings, not by Required()
        },
        budgetSvc: svc,
    }
}

func (s *BudgetCheckStage) RunCondition() string {
    // Only run budget check if the invoice amount exceeds a meaningful threshold
    // (tenant can configure this threshold in settings)
    return `setting["budget.minimum_check_amount"] == "" || input.total_amount > toNumber(setting["budget.minimum_check_amount"])`
}

func (s *BudgetCheckStage) Execute(opCtx *OperationContext) (StageResult, error) {
    invoice, ok := opCtx.Input.(*ap.Invoice)
    if !ok {
        return StageResult{}, fmt.Errorf("budget.check: unexpected input type")
    }

    // Behaviour driven by tenant settings (not by the stage being present):
    // budget.control_mode = "none" | "warn" | "soft_block" | "hard_block"
    mode := opCtx.SettingString("budget.control_mode", "warn")

    result, err := s.budgetSvc.Check(opCtx.Ctx, budget.CheckParams{
        TenantID:     opCtx.TenantID,
        CostCenterID: invoice.CostCenterID,
        Amount:       invoice.TotalAmount,
        PeriodDate:   invoice.InvoiceDate,
    })
    if err != nil { return StageResult{}, err }

    // Write outputs for downstream stages (GL posting, notifications)
    opCtx.SetData("budget_check.approved", result.WithinBudget)
    opCtx.SetData("budget_check.available_amount", result.AvailableAmount)
    opCtx.SetData("budget_check.utilisation_pct", result.UtilisationPct)
    opCtx.SetFlag("budget_exceeded", !result.WithinBudget)

    if !result.WithinBudget {
        switch mode {
        case "hard_block":
            // Return error — pipeline aborts here
            return StageResult{}, fmt.Errorf(
                "budget exceeded: %.2f available, %.2f requested",
                result.AvailableAmount.InexactFloat64(),
                invoice.TotalAmount.InexactFloat64())
        case "soft_block":
            // Suspend — requires override approval from budget manager
            opCtx.Suspend(
                "budget_exceeded:override_required",
                "budget.check",   // resume from here after override approval
            )
            return StageResult{
                Status:  "suspended",
                Message: "Budget exceeded — awaiting override approval",
            }, nil
        case "warn":
            // Continue — just flag it; downstream stages (notification) will alert
            return StageResult{
                Status:  "completed",
                Message: fmt.Sprintf("Warning: budget %.1f%% utilised", result.UtilisationPct.InexactFloat64()),
                Outputs: map[string]any{"budget_exceeded": true},
            }, nil
        }
    }

    return StageResult{
        Status:  "completed",
        Message: fmt.Sprintf("Budget OK — %.1f%% utilised", result.UtilisationPct.InexactFloat64()),
    }, nil
}
```

### 4.4 Stage Priority Bands

```
100–199  Validation
         ap.validate_invoice (110)
         ap.duplicate_check  (120)
         inventory.validate_grn (110)

200–299  Enrichment / Data Fetching
         ap.resolve_vendor (210)
         ap.resolve_purchase_order (220)
         tax.resolve_tax_codes (250)

300–399  Business Rule Checks
         procurement.three_way_match (300)
         budget.check (310)
         ap.credit_limit_check (320)
         compliance.sanctions_check (390)

400–499  Tax & Regulatory
         tax.calculate (410)
         tax.validate_e_invoice (420)
         tax.apply_withholding (430)

500–599  Workflow & Approval Gates
         workflow.submit_for_approval (510)
         ← pipeline may suspend here

600–699  Core Financial Posting
         gl.post_transaction (610)
         ar.apply_advance_payment (620)
         ar.apply_credit_note (630)

700–799  Secondary Financial & Storage
         dms.archive_document (710)
         banking.schedule_payment (720)
         inventory.update_grn_status (730)

800–899  Events & Notifications
         notifications.send (810)
         events.emit_domain_event (820)

900–999  Audit & Compliance
         audit.log_operation (910)
         compliance.regulatory_report (950)
```

---

## 5. Pipeline & PipelineBuilder

### 5.1 The Pipeline

```go
// awo/internal/pipeline/pipeline.go

type Pipeline struct {
    operationKey string
    stages       []Stage
    hooksBefore  map[string][]Hook  // "budget.check" → []Hook
    hooksAfter   map[string][]Hook
    repo         OperationLogRepository
}

func (p *Pipeline) Execute(opCtx *OperationContext) (*OperationResult, error) {
    opCtx.StartedAt = time.Now()
    var completedStages []stageCheckpoint   // stack for saga compensation

    for _, stage := range p.stages {
        if opCtx.Suspended { break }

        stageLog := StageLog{StageName: stage.Name(), StartedAt: time.Now()}

        // ── Feature flag gate ─────────────────────────────────────────────
        if flag := stage.FeatureFlag(); flag != "" {
            if !opCtx.FeatureEnabled(flag) {
                stageLog.Status = "skipped"
                stageLog.SkipReason = fmt.Sprintf("feature '%s' not enabled", flag)
                opCtx.Log = append(opCtx.Log, stageLog)
                continue
            }
        }

        // ── Run condition gate ────────────────────────────────────────────
        if cond := stage.RunCondition(); cond != "" {
            evalCtx := condition.NewEvalContext(buildEvalData(opCtx), condition.DefaultEvalOptions())
            builder := condition.NewBuilder(condition.ConjunctionAnd)
            builder.AddFormula(cond)
            result, err := evaluator.Evaluate(opCtx.Ctx, builder.Build(), evalCtx)
            if err != nil || !result {
                stageLog.Status = "skipped"
                stageLog.SkipReason = "run condition evaluated to false"
                opCtx.Log = append(opCtx.Log, stageLog)
                continue
            }
        }

        // ── Before hooks ──────────────────────────────────────────────────
        for _, hook := range p.hooksBefore[stage.Name()] {
            if err := hook.Execute(opCtx); err != nil {
                if stage.Required() { return nil, fmt.Errorf("before hook %s: %w", hook.Name(), err) }
            }
            if opCtx.Suspended { break }
        }
        if opCtx.Suspended { break }

        // ── Execute stage (dry-run aware) ─────────────────────────────────
        var result StageResult
        var err error
        if opCtx.DryRun {
            if sim, ok := stage.(Simulatable); ok {
                result, err = sim.Simulate(opCtx)
            } else {
                result, err = stage.Execute(opCtx)  // read-only stages run normally
            }
        } else {
            result, err = stage.Execute(opCtx)
        }

        stageLog.Duration = time.Since(stageLog.StartedAt)

        if err != nil && stage.Required() {
            // Classify the failure
            failClass := classifyError(err)
            if failClass == FailureRetryable {
                result, err = p.retryStage(opCtx, stage)
            }
            if err != nil {
                stageLog.Status = "failed"
                stageLog.Error = err.Error()
                opCtx.Log = append(opCtx.Log, stageLog)
                p.compensate(opCtx, completedStages)
                p.repo.SaveLog(opCtx.Ctx, p.buildResult(opCtx, "failed"))
                return p.buildResult(opCtx, "failed"), err
            }
        } else if err != nil {
            // Non-required stage failed — log and continue
            stageLog.Status = "failed"
            stageLog.Error = err.Error()
            opCtx.Log = append(opCtx.Log, stageLog)
            continue
        }

        stageLog.Status = result.Status
        stageLog.Output = result.Outputs
        stageLog.Message = result.Message
        opCtx.Log = append(opCtx.Log, stageLog)

        // Push onto checkpoint stack for possible compensation
        completedStages = append(completedStages, stageCheckpoint{
            StageName: stage.Name(),
            Output:    result.Outputs,
        })

        if opCtx.Suspended { break }

        // ── After hooks ───────────────────────────────────────────────────
        for _, hook := range p.hooksAfter[stage.Name()] {
            if err := hook.Execute(opCtx); err != nil {
                p.logHookError(opCtx, hook.Name(), err)
            }
        }
    }

    now := time.Now()
    opCtx.CompletedAt = &now

    result := p.buildResult(opCtx, p.determineStatus(opCtx))
    p.repo.SaveLog(opCtx.Ctx, result)
    return result, nil
}

type stageCheckpoint struct {
    StageName string
    Output    map[string]any
}

func (p *Pipeline) determineStatus(opCtx *OperationContext) string {
    if opCtx.Suspended { return "pending_approval" }
    for _, log := range opCtx.Log {
        if log.Status == "failed" { return "failed" }
    }
    return "completed"
}
```

### 5.2 PipelineBuilder

```go
// awo/internal/pipeline/builder.go

type PipelineBuilder struct {
    stageRegistry *StageRegistry
    hookRegistry  *HookRegistry
    logRepo       OperationLogRepository
    cache         *pipelineCache
}

type pipelineCache struct {
    mu    sync.RWMutex
    items map[string]*cachedPipeline   // key: "{op_key}:{flags_hash}"
    ttl   time.Duration                // default: 5 minutes
}

type cachedPipeline struct {
    pipeline  *Pipeline
    cachedAt  time.Time
}

func (b *PipelineBuilder) Build(
    operationKey string,
    session *domain.ResolvedSession) (*Pipeline, error) {

    // Check cache first
    flagsHash := b.hashRelevantFlags(operationKey, session)
    cacheKey  := operationKey + ":" + flagsHash
    if cached, ok := b.cache.Get(cacheKey); ok {
        return cached, nil
    }

    // ── Collect applicable stages ─────────────────────────────────────────
    allStages := b.stageRegistry.ForOperation(operationKey)

    var activeStages []Stage
    for _, stage := range allStages {
        if flag := stage.FeatureFlag(); flag != "" {
            if !session.FeatureEnabled(flag) { continue }
        }
        activeStages = append(activeStages, stage)
    }

    sort.Slice(activeStages, func(i, j int) bool {
        return activeStages[i].Priority() < activeStages[j].Priority()
    })

    // ── Collect applicable hooks ──────────────────────────────────────────
    allHooks := b.hookRegistry.ForOperation(operationKey)
    hooksBefore := make(map[string][]Hook)
    hooksAfter  := make(map[string][]Hook)

    for _, hook := range allHooks {
        if flag := hook.FeatureFlag(); flag != "" {
            if !session.FeatureEnabled(flag) { continue }
        }
        for _, stageName := range hook.StageBefore() {
            hooksBefore[stageName] = append(hooksBefore[stageName], hook)
        }
        for _, stageName := range hook.StageAfter() {
            hooksAfter[stageName] = append(hooksAfter[stageName], hook)
        }
    }

    for stageName := range hooksBefore {
        sort.Slice(hooksBefore[stageName], func(i, j int) bool {
            return hooksBefore[stageName][i].Priority() < hooksBefore[stageName][j].Priority()
        })
    }

    pipeline := &Pipeline{
        operationKey: operationKey,
        stages:       activeStages,
        hooksBefore:  hooksBefore,
        hooksAfter:   hooksAfter,
        repo:         b.logRepo,
    }

    b.cache.Set(cacheKey, pipeline)
    return pipeline, nil
}

// InvalidateForTenant clears cached pipelines when tenant flags change.
// Called automatically by FlagService.Set().
func (b *PipelineBuilder) InvalidateForTenant(tenantID uuid.UUID) {
    b.cache.DeleteByPrefix(tenantID.String() + ":")
}
```

### 5.3 Operation Log Persistence

Every pipeline execution is persisted for audit and saga resumption:

```sql
CREATE TABLE operation_logs (
  id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        uuid        NOT NULL REFERENCES tenants(id),
  operation_id     uuid        NOT NULL UNIQUE,
  operation_key    text        NOT NULL,
  resource         text        NOT NULL,
  action           text        NOT NULL,
  status           text        NOT NULL,
  -- 'completed' | 'pending_approval' | 'failed' | 'resumed'
  input_snapshot   jsonb       NOT NULL,
  stage_log        jsonb       NOT NULL,
  output_data      jsonb       NOT NULL DEFAULT '{}',
  suspend_reason   text,
  resume_point     text,
  idempotency_key  text,
  session_snapshot jsonb,
  flags_snapshot   jsonb,
  initiated_by     uuid        NOT NULL REFERENCES users(id),
  started_at       timestamptz NOT NULL,
  completed_at     timestamptz,
  duration_ms      int
);

-- For saga resumption queries
CREATE INDEX idx_operation_logs_pending ON operation_logs(tenant_id, operation_key)
    WHERE status = 'pending_approval';

-- For audit trail queries
CREATE INDEX idx_operation_logs_tenant_date ON operation_logs(tenant_id, started_at DESC);

-- For idempotency
CREATE UNIQUE INDEX idx_operation_logs_idempotency
    ON operation_logs(tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
```

---

## 6. Hook System — The Extension Point

### 6.1 Hook Interface

```go
// awo/internal/pipeline/hook.go

// Hook is an extension point that fires before or after a named stage.
// Future modules (workflow, DMS, e-invoicing, compliance) register hooks
// without modifying core operation code.
type Hook interface {
    Name() string
    FeatureFlag() string
    Priority() int
    StageBefore() []string
    StageAfter() []string
    Operations() []string
    Execute(opCtx *OperationContext) error
}
```

### 6.2 How Hooks Enable Workflow Without Core Changes

The workflow module, when enabled, registers a single hook that fires before GL posting on any financial operation. The core AP, inventory, and GL code never changes:

```go
// awo/internal/workflow/pipeline_hook.go

// WorkflowApprovalHook intercepts financial operations before they post
// to the GL. If the operation requires approval (based on workflow rules),
// it suspends the pipeline and creates a pending approval task.
type WorkflowApprovalHook struct {
    workflowSvc WorkflowService
    evaluator   *condition.Evaluator
}

func (h *WorkflowApprovalHook) Name() string        { return "workflow.approval_gate" }
func (h *WorkflowApprovalHook) FeatureFlag() string  { return "workflow" }
func (h *WorkflowApprovalHook) Priority() int        { return 100 }
func (h *WorkflowApprovalHook) Operations() []string { return []string{"*"} }
func (h *WorkflowApprovalHook) StageBefore() []string { return []string{"gl.post_transaction"} }
func (h *WorkflowApprovalHook) StageAfter()  []string { return []string{} }

func (h *WorkflowApprovalHook) RunCondition() string {
    return `
        toNumber(gl_posting_amount) > toNumber(setting["workflow.approval_threshold_amount"])
        || budget_exceeded == true
        || input.entity_type == "subsidiary"
    `
}

func (h *WorkflowApprovalHook) Execute(opCtx *OperationContext) error {
    config, err := h.workflowSvc.GetRuleFor(opCtx.Ctx, workflow.RuleQuery{
        TenantID:    opCtx.TenantID,
        Operation:   opCtx.OperationKey,
        UserID:      opCtx.UserID,
        EntityID:    opCtx.EntityID,
        Amount:      getAmount(opCtx),
        ExtraFlags:  opCtx.Flags,
    })
    if err != nil { return err }

    if !config.RequiresApproval { return nil }

    approval, err := h.workflowSvc.CreateApprovalTask(opCtx.Ctx, workflow.TaskParams{
        TenantID:         opCtx.TenantID,
        OperationID:      opCtx.OperationID,
        OperationKey:     opCtx.OperationKey,
        ResourceSnapshot: opCtx.Input,
        RequestedBy:      opCtx.UserID,
        Approvers:        config.Approvers,
        DueIn:            config.ApprovalDeadline,
        CallbackURL:      "/internal/pipeline/resume/" + opCtx.OperationID.String(),
    })
    if err != nil { return err }

    opCtx.SetData("workflow.approval_id", approval.ID)
    opCtx.SetData("workflow.approvers", config.Approvers)
    opCtx.Suspend(
        "pending_workflow_approval:" + approval.ID.String(),
        "gl.post_transaction",
    )
    return nil
}
```

### 6.3 Other Hook Examples

```go
// Notification hook — fires after GL posting on any financial operation
type PostingNotificationHook struct {
    notifySvc NotificationService
}
func (h *PostingNotificationHook) StageBefore() []string { return []string{} }
func (h *PostingNotificationHook) StageAfter()  []string { return []string{"gl.post_transaction"} }
func (h *PostingNotificationHook) Execute(opCtx *OperationContext) error {
    txID, _ := opCtx.GetData("gl_posting.transaction_id")
    budgetExceeded := opCtx.Flag("budget_exceeded")
    return h.notifySvc.Dispatch(opCtx.Ctx, ...)
}

// Audit hook — fires after EVERY stage on EVERY operation
type AuditLogHook struct {
    auditSvc AuditService
}
func (h *AuditLogHook) Operations() []string  { return []string{"*"} }
func (h *AuditLogHook) StageBefore() []string { return []string{} }
func (h *AuditLogHook) StageAfter()  []string { return []string{"*"} }
func (h *AuditLogHook) Execute(opCtx *OperationContext) error {
    if len(opCtx.Log) == 0 { return nil }
    last := opCtx.Log[len(opCtx.Log)-1]
    return h.auditSvc.RecordStage(opCtx.Ctx, audit.StageRecord{
        OperationID: opCtx.OperationID,
        TenantID:    opCtx.TenantID,
        Stage:       last.StageName,
        Status:      last.Status,
        ActorUserID: opCtx.UserID,
        OccurredAt:  last.StartedAt,
    })
}

// E-invoice hook — fires after GL posting if e-invoicing is enabled
type EInvoiceSubmissionHook struct {
    eInvoiceSvc EInvoiceService
}
func (h *EInvoiceSubmissionHook) FeatureFlag() string   { return "tax.e_invoice" }
func (h *EInvoiceSubmissionHook) StageBefore() []string { return []string{} }
func (h *EInvoiceSubmissionHook) StageAfter()  []string { return []string{"gl.post_transaction"} }
func (h *EInvoiceSubmissionHook) Execute(opCtx *OperationContext) error {
    txID, _ := opCtx.GetData("gl_posting.transaction_id")
    return h.eInvoiceSvc.Submit(opCtx.Ctx, txID.(uuid.UUID))
}
```

---

## 7. Stage & Hook Registries

### 7.1 Registry Design

```go
// awo/internal/pipeline/registry.go

type StageRegistry struct {
    mu     sync.RWMutex
    stages []Stage
}

func (r *StageRegistry) Register(stages ...Stage) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.stages = append(r.stages, stages...)
}

func (r *StageRegistry) ForOperation(operationKey string) []Stage {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var result []Stage
    for _, stage := range r.stages {
        for _, op := range stage.Operations() {
            if op == "*" || op == operationKey {
                result = append(result, stage)
                break
            }
        }
    }
    return result
}

// HookRegistry follows the same pattern
type HookRegistry struct {
    mu    sync.RWMutex
    hooks []Hook
}

func (r *HookRegistry) Register(hooks ...Hook) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.hooks = append(r.hooks, hooks...)
}

func (r *HookRegistry) ForOperation(operationKey string) []Hook {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var result []Hook
    for _, hook := range r.hooks {
        for _, op := range hook.Operations() {
            if op == "*" || op == operationKey {
                result = append(result, hook)
                break
            }
        }
    }
    return result
}
```

### 7.2 Registration at Startup (wire.go)

All modules register their stages and hooks in the application wiring. This is the only place where module dependencies are visible — the pipeline itself knows nothing about individual modules:

```go
// awo/cmd/server/wire.go

func InitializeApp(cfg *config.Config) (*App, error) {

    // ── Build services ────────────────────────────────────────────────────
    budgetSvc      := budget.NewService(...)
    taxSvc         := tax.NewService(...)
    workflowSvc    := workflow.NewService(...)
    glSvc          := finance.NewGLService(...)
    dmsSvc         := dms.NewService(...)
    bankingSvc     := banking.NewService(...)
    notifySvc      := notify.NewService(...)
    auditSvc       := audit.NewService(...)
    procurementSvc := procurement.NewService(...)

    // ── Register stages ───────────────────────────────────────────────────
    stageRegistry := pipeline.NewStageRegistry()
    stageRegistry.Register(
        ap.NewValidateInvoiceStage(),
        ap.NewDuplicateCheckStage(),
        ap.NewResolveVendorStage(),
        ap.NewResolveGLAccountsStage(),
        procurement.NewThreeWayMatchStage(procurementSvc),
        budget.NewBudgetCheckStage(budgetSvc),
        tax.NewResolveTaxCodesStage(taxSvc),
        tax.NewCalculateTaxStage(taxSvc),
        tax.NewWithholdingTaxStage(taxSvc),
        finance.NewGLPostingStage(glSvc),
        finance.NewAROffsetStage(arSvc),
        dms.NewArchiveDocumentStage(dmsSvc),
        banking.NewSchedulePaymentStage(bankingSvc),
        audit.NewAuditLogStage(auditSvc),
    )

    // ── Register hooks ────────────────────────────────────────────────────
    hookRegistry := pipeline.NewHookRegistry()
    hookRegistry.Register(
        workflow.NewWorkflowApprovalHook(workflowSvc),
        notify.NewPostingNotificationHook(notifySvc),
        audit.NewAuditLogHook(auditSvc),
        tax.NewEInvoiceSubmissionHook(taxSvc),
        compliance.NewSanctionsCheckHook(complianceSvc),
    )

    // ── Register compensations ────────────────────────────────────────────
    compensationRegistry := pipeline.NewCompensationRegistry()

    compensationRegistry.Register("gl.post_transaction", func(opCtx *pipeline.OperationContext) error {
        txID, ok := opCtx.GetData("gl_posting.transaction_id")
        if !ok { return nil }
        return glSvc.ReverseTransaction(opCtx.Ctx, txID.(uuid.UUID),
            "Pipeline compensation: "+opCtx.OperationKey)
    })
    compensationRegistry.Register("banking.schedule_payment", func(opCtx *pipeline.OperationContext) error {
        paymentID, ok := opCtx.GetData("payment_scheduling.payment_id")
        if !ok { return nil }
        return bankingSvc.CancelScheduledPayment(opCtx.Ctx, paymentID.(uuid.UUID))
    })
    compensationRegistry.Register("dms.archive", func(opCtx *pipeline.OperationContext) error {
        docID, ok := opCtx.GetData("dms_archive.document_id")
        if !ok { return nil }
        return dmsSvc.VoidDocument(opCtx.Ctx, docID.(uuid.UUID), "Pipeline rollback")
    })
    compensationRegistry.Register("budget.check", func(opCtx *pipeline.OperationContext) error {
        if !opCtx.Flag("budget_reserved") { return nil }
        return budgetSvc.ReleaseReservation(opCtx.Ctx, opCtx.OperationID)
    })
    // Stages with no side effects need no compensation:
    // ap.validate_invoice, ap.duplicate_check, tax.calculate → read-only

    // ── Build pipeline infrastructure ─────────────────────────────────────
    pipelineBuilder := pipeline.NewPipelineBuilder(stageRegistry, hookRegistry, logRepo)

    apService := ap.NewService(pipelineBuilder, ...)
    inventoryService := inventory.NewService(pipelineBuilder, ...)

    return &App{...}, nil
}
```

---

## 8. Canonical Example: AP Invoice Processing

### 8.1 The Full Pipeline Definition

```
Operation: "ap.invoice.process"

Stage                          Module        Flag Gate          Priority  Required
─────────────────────────────────────────────────────────────────────────────────
ap.validate_invoice            ap            —                  110       Yes
ap.duplicate_check             ap            —                  120       Yes
ap.resolve_vendor              ap            —                  210       Yes
ap.resolve_gl_accounts         ap            —                  220       Yes
procurement.three_way_match    procurement   "procurement"      300       No
budget.check                   budget        "budget"           310       No
compliance.sanctions_check     compliance    "compliance"       390       No
tax.resolve_tax_codes          tax           "tax"              400       No
tax.calculate                  tax           "tax"              410       No
tax.apply_withholding          tax           "tax.withholding"  430       No
  ↕ HOOK before gl.post_transaction: workflow.approval_gate ("workflow")
gl.post_transaction            finance       —                  610       Yes
ar.apply_advance_payment       ar            "ar.advances"      620       No
dms.archive_document           dms           "dms"              710       No
banking.schedule_payment       banking       "banking"          720       No
  ↕ HOOK after banking.schedule_payment: notify.send_all (—)
audit.log_operation            audit         —                  910       Yes
─────────────────────────────────────────────────────────────────────────────────
```

### 8.2 Service Entry Point

The AP service no longer contains any feature-flag logic. It just builds and executes the pipeline:

```go
// awo/internal/ap/service.go

type APService struct {
    pipelineBuilder *pipeline.PipelineBuilder
    invoiceRepo     InvoiceRepository
}

func (s *APService) ProcessInvoice(ctx context.Context,
    params ProcessInvoiceParams) (*ProcessInvoiceResult, error) {

    session := domain.SessionFromContext(ctx)

    if !session.Can("ap.invoices", "process") {
        return nil, domain.ErrForbidden
    }

    invoice, err := s.invoiceRepo.Get(ctx, params.InvoiceID)
    if err != nil { return nil, err }

    p, err := s.pipelineBuilder.Build("ap.invoice.process", session)
    if err != nil { return nil, err }

    opCtx := &pipeline.OperationContext{
        Ctx:          ctx,
        Session:      session,
        TenantID:     session.TenantID,
        UserID:       session.UserID,
        EntityID:     session.EntityID,
        OperationID:  uuid.New(),
        OperationKey: "ap.invoice.process",
        Resource:     "ap.invoice",
        Action:       "process",
        Input:        invoice,
        Data:         make(map[string]any),
        Flags:        make(map[string]bool),
    }

    result, err := p.Execute(opCtx)
    if err != nil { return nil, err }

    return &ProcessInvoiceResult{
        Status:      result.Status,
        InvoiceID:   params.InvoiceID,
        GLTxID:      getUUID(opCtx.Data, "gl_posting.transaction_id"),
        ApprovalID:  getUUID(opCtx.Data, "workflow.approval_id"),
        OperationID: opCtx.OperationID,
        Log:         opCtx.Log,
    }, nil
}
```

### 8.3 What Each Tenant Sees

**Tenant with only Finance module:**
```
Pipeline stages active:
  ✓ ap.validate_invoice
  ✓ ap.duplicate_check
  ✓ ap.resolve_vendor
  ✓ ap.resolve_gl_accounts
  ✗ procurement.three_way_match   (flag "procurement" = false)
  ✗ budget.check                  (flag "budget" = false)
  ✗ compliance.sanctions_check    (flag "compliance" = false)
  ✗ tax.resolve_tax_codes         (flag "tax" = false)
  ✓ gl.post_transaction
  ✗ ar.apply_advance_payment      (flag "ar.advances" = false)
  ✗ dms.archive_document          (flag "dms" = false)
  ✗ banking.schedule_payment      (flag "banking" = false)
  ✓ audit.log_operation

Result: simple 5-stage pipeline for a basic tenant.
```

**Tenant with full Enterprise suite:**
```
Pipeline stages active:
  ✓ ap.validate_invoice
  ✓ ap.duplicate_check
  ✓ ap.resolve_vendor
  ✓ ap.resolve_gl_accounts
  ✓ procurement.three_way_match
  ✓ budget.check
  ✓ compliance.sanctions_check
  ✓ tax.resolve_tax_codes
  ✓ tax.calculate
  ✓ tax.apply_withholding
    ↕ Hook: workflow.approval_gate  ← suspends if amount > threshold
  ✓ gl.post_transaction
  ✓ ar.apply_advance_payment
  ✓ dms.archive_document
  ✓ banking.schedule_payment
    ↕ Hook: notify.send_all
  ✓ audit.log_operation

Result: 15-stage pipeline with 2 hooks for an enterprise tenant.
```

Both tenants call the **same `ProcessInvoice` function**. The pipeline is different. The core service code is identical.

---

## 9. Midway Failure — What Actually Happens

### 9.1 Failure Taxonomy

Not all failures are the same. The pipeline distinguishes five categories, each handled differently:

```
Category              Example                           Recovery
─────────────────────────────────────────────────────────────────────
ValidationFailure     Duplicate invoice detected        Abort immediately; no compensation needed
BusinessRuleFailure   Budget exceeded (hard block)      Abort; no GL was touched; safe
ServiceFailure        Tax service HTTP 503              Retry with backoff; resume from this stage
IntegrationFailure    GL posting DB conflict            Retry; if exhausted → compensate prior stages
InfrastructureFailure DB connection lost mid-pipeline   Resume from checkpoint after recovery
```

The pipeline does not treat all errors as fatal. `stage.Required() = false` means a service failure logs and continues. `stage.Required() = true` means the pipeline evaluates the failure category and decides: retry, suspend, compensate, or abort.

### 9.2 The Compensation Registry

Every stage that produces a side effect that cannot be automatically rolled back by a database transaction must register a compensating function. Compensation is the saga pattern applied at the stage level:

```go
// awo/internal/pipeline/compensation.go

// CompensationFn undoes what a stage did.
// It receives the original OperationContext at the time the stage ran,
// including the stage's output in opCtx.Data["{stage_name}.*"].
type CompensationFn func(opCtx *OperationContext) error

// CompensationRegistry maps stage names to their undo functions.
// Registered at startup alongside stages and hooks.
type CompensationRegistry struct {
    mu      sync.RWMutex
    entries map[string]CompensationFn
}

func (r *CompensationRegistry) Register(stageName string, fn CompensationFn) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.entries[stageName] = fn
}

func (r *CompensationRegistry) Get(stageName string) (CompensationFn, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    fn, ok := r.entries[stageName]
    return fn, ok
}
```

### 9.3 Saga Rollback Algorithm

When a required stage fails and retry is exhausted, the pipeline runs compensation in **reverse order** for every stage that successfully completed:

```go
// compensate runs compensation functions in reverse order (LIFO)
func (p *Pipeline) compensate(opCtx *OperationContext, completed []stageCheckpoint) {
    for i := len(completed) - 1; i >= 0; i-- {
        checkpoint := completed[i]
        fn, exists := p.compensationRegistry.Get(checkpoint.StageName)
        if !exists { continue }

        p.logCompensation(opCtx, checkpoint.StageName, "starting")
        if err := fn(opCtx); err != nil {
            // Compensation failure is logged but does NOT stop other compensations.
            // A human must resolve compensation failures manually.
            p.logCompensation(opCtx, checkpoint.StageName, "FAILED: "+err.Error())
            p.events.Emit(opCtx.Ctx, domain.CompensationFailed{
                OperationID: opCtx.OperationID,
                StageName:   checkpoint.StageName,
                Error:       err.Error(),
            })
            // Continue compensating other stages — do not abort
        } else {
            p.logCompensation(opCtx, checkpoint.StageName, "completed")
        }
    }
    opCtx.SetFlag("compensated", true)
}
```

**Retry with exponential backoff:**

```go
func (p *Pipeline) retryStage(opCtx *OperationContext, stage Stage) (StageResult, error) {
    maxRetries := 3
    baseDelay  := 500 * time.Millisecond

    for attempt := 1; attempt <= maxRetries; attempt++ {
        delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
        jitter := time.Duration(rand.Int63n(int64(delay / 4)))
        time.Sleep(delay + jitter)

        result, err := stage.Execute(opCtx)
        if err == nil { return result, nil }

        p.logRetry(opCtx, stage.Name(), attempt, err)
    }
    return StageResult{}, fmt.Errorf("stage %s failed after %d retries", stage.Name(), maxRetries)
}
```

### 9.4 Idempotency Keys

Every pipeline execution has an `OperationID` (uuid). If the same operation is submitted twice (network retry, user double-click), the second call detects the existing in-progress or completed operation and returns its result without re-executing:

```go
// awo/internal/pipeline/idempotency.go

func (p *Pipeline) ExecuteIdempotent(opCtx *OperationContext,
    idempotencyKey string) (*OperationResult, error) {

    existing, err := p.repo.GetByIdempotencyKey(opCtx.Ctx, opCtx.TenantID, idempotencyKey)
    if err == nil {
        switch existing.Status {
        case "completed":
            return existing.Result, nil
        case "running":
            return nil, ErrOperationInProgress
        case "failed":
            return existing.Result, ErrOperationFailed
        case "pending_approval":
            return existing.Result, ErrPendingApproval
        }
    }

    p.repo.SetIdempotencyKey(opCtx.Ctx, opCtx.OperationID, idempotencyKey)
    return p.Execute(opCtx)
}
```

### 9.5 Partial Completion Recovery

For pipelines suspended on workflow approval, the state is checkpointed. When resumed, the pipeline skips already-completed stages and reconstructs the `OperationContext` from the persisted log:

```go
// awo/internal/pipeline/resume.go

func (s *PipelineResumeService) Resume(ctx context.Context,
    operationID uuid.UUID, resumeData map[string]any) (*OperationResult, error) {

    log, _ := s.repo.GetByOperationID(ctx, operationID)
    if log.Status != "pending_approval" {
        return nil, fmt.Errorf("operation %s cannot be resumed (status: %s)", operationID, log.Status)
    }

    opCtx := &OperationContext{
        Ctx:          ctx,
        Session:      s.sessionRepo.Reconstruct(ctx, log.SessionSnapshot),
        OperationID:  operationID,
        OperationKey: log.OperationKey,
        Input:        log.InputSnapshot,
        Data:         log.OutputData,
        Flags:        log.FlagsSnapshot,
        Log:          log.StageLog,
    }

    for k, v := range resumeData {
        opCtx.SetData("resume."+k, v)
    }
    opCtx.SetFlag("workflow_approved", true)
    opCtx.Suspended = false

    session := opCtx.Session
    p, _ := s.pipelineBuilder.Build(log.OperationKey, session)

    completedStageNames := extractCompletedStageNames(opCtx.Log)
    p.SkipStages(completedStageNames)
    p.StartFrom(log.ResumePoint)

    return p.Execute(opCtx)
}
```

---

## 10. Long-Running Operations — Saga Pattern

### 10.1 When an Operation Spans Days

Some operations cannot complete in a single HTTP request:

```
AP Invoice Processing with workflow approval:
  T+0:     Invoice received → pipeline runs up to WorkflowApprovalHook → suspended
  T+0:     Approval notification sent to Finance Manager
  T+2d:    Finance Manager approves → pipeline resumes → GL posted → payment scheduled
  T+7d:    Payment execution → bank transfer sent

Procurement → Inventory → AP (three-way match):
  T+0:     PO raised → sent to vendor
  T+5d:    Goods received → GRN created → inventory updated
  T+7d:    Vendor invoice arrives → three-way match → AP invoice processed
```

### 10.2 The Saga Coordinator

For operations that span multiple services across time, a saga coordinator tracks the state machine:

```go
// awo/internal/pipeline/saga.go

type Saga struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    SagaType    string    // "procurement.full_cycle" | "ap.invoice.process"
    Status      string    // "running" | "waiting" | "completed" | "compensating" | "failed"
    Steps       []SagaStep
    CurrentStep int
    Context     map[string]any
    StartedAt   time.Time
}

type SagaStep struct {
    Name        string
    Status      string      // "pending" | "running" | "completed" | "failed" | "compensated"
    OperationID *uuid.UUID
    CompletedAt *time.Time
    Compensation string     // what to call if this step needs to be undone
}
```

### 10.3 Compensation (Undo) Logic

If a later step fails, earlier steps must be compensated (rolled back):

```
Procurement Full Cycle Saga:
  Step 1: CreatePurchaseOrder       Compensation: CancelPurchaseOrder
  Step 2: ReceiveGoods (GRN)        Compensation: ReverseGRN
  Step 3: MatchAndPostInvoice       Compensation: VoidInvoice
  Step 4: SchedulePayment           Compensation: CancelPayment

If Step 3 fails (three-way match discrepancy):
  → Compensate Step 2: ReverseGRN (stock goes back)
  → Compensate Step 1: CancelPurchaseOrder (optional — may keep PO open)
```

---

## 11. Workflow Engine Integration

### 11.1 Workflow as a Pure Hook Consumer

The workflow engine does not need any special integration points in the core modules. It registers one hook and one resume endpoint:

```go
// awo/internal/workflow/domain/rule.go

// WorkflowRule defines when an operation requires approval and who approves it.
type WorkflowRule struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    OperationKey string
    Name         string

    // Conditions (ALL must be true for rule to apply)
    Conditions   []RuleCondition
    // e.g. {field: "amount", op: "gte", value: "100000"}
    //      {field: "entity_id", op: "in", value: ["nairobi_uuid"]}
    //      {field: "budget_exceeded", op: "eq", value: "true"}

    ApprovalType string    // "any_of" | "all_of" | "sequential"
    Approvers    []Approver
    Deadline     time.Duration
    EscalatesTo  []Approver
    Priority     int
}

type RuleCondition struct {
    Field    string   // "amount" | "budget_exceeded" | "entity_id"
    Operator string   // "eq" | "neq" | "gte" | "lte" | "in" | "contains"
    Value    string
}
```

### 11.2 Pipeline Resume After Approval

When an approver approves the task, the workflow module calls the resume endpoint. The pipeline reloads the persisted `OperationContext` and continues from where it was suspended:

```go
// awo/internal/pipeline/resume.go

func (s *PipelineResumeService) Resume(ctx context.Context,
    operationID uuid.UUID, approvalData map[string]any) error {

    log, err := s.logRepo.GetByOperationID(ctx, operationID)
    if err != nil { return err }

    if log.Status != "pending_approval" {
        return fmt.Errorf("operation %s is not pending approval (status: %s)",
            operationID, log.Status)
    }

    opCtx, err := s.reconstructContext(ctx, log)
    if err != nil { return err }

    for k, v := range approvalData {
        opCtx.SetData("workflow.approval_"+k, v)
    }
    opCtx.SetFlag("workflow_approved", true)
    opCtx.Suspended = false
    opCtx.SuspendReason = ""

    session := opCtx.Session
    p, err := s.pipelineBuilder.Build(log.OperationKey, session)
    if err != nil { return err }

    completedStages := s.extractCompletedStages(opCtx.Log)
    p.SkipCompletedStages(completedStages)

    _, err = p.Execute(opCtx)
    return err
}
```

### 11.3 Approval Task Lifecycle

```sql
CREATE TABLE workflow_approval_tasks (
  id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        uuid        NOT NULL REFERENCES tenants(id),
  operation_id     uuid        NOT NULL,
  operation_key    text        NOT NULL,
  status           text        NOT NULL DEFAULT 'pending',
  -- 'pending' | 'approved' | 'rejected' | 'escalated' | 'expired'
  requested_by     uuid        NOT NULL REFERENCES users(id),
  requested_at     timestamptz NOT NULL DEFAULT now(),
  due_at           timestamptz NOT NULL,
  approval_type    text        NOT NULL,  -- 'any_of' | 'all_of' | 'sequential'
  approvers        jsonb       NOT NULL,
  current_step     int         NOT NULL DEFAULT 0,
  decided_at       timestamptz,
  decided_by       uuid        REFERENCES users(id),
  decision         text,
  decision_notes   text,
  resource_snapshot jsonb      NOT NULL
);
```

---

## 12. Performance Optimisation

### 12.1 Pre-Built Pipeline Cache

Building a pipeline on every request involves iterating all registered stages, filtering by operation and flags, and sorting by priority. For high-frequency operations this cost adds up. Pipelines are cached by `(operation_key, flags_hash)`:

```go
// Cache invalidation is triggered by FlagService.Set():
func (s *FlagService) Set(ctx context.Context, params domain.SetFlagParams) error {
    if err := s.repo.UpsertTenantFlag(ctx, params); err != nil { return err }
    // Invalidate pipeline cache so next request builds with updated flags
    s.pipelineBuilder.InvalidateForTenant(params.TenantID)
    // Invalidate sessions (existing behaviour)
    s.sessionRepo.InvalidateByTenant(ctx, params.TenantID)
    return nil
}
```

### 12.2 Parallel Stage Execution

Some stages within a band have no data dependency on each other and can run concurrently. The pipeline detects this from a stage's `DependsOn()` declaration:

```go
// Example — Tax and Budget can run in parallel (neither reads the other's output):
func (s *TaxCalculationStage) DependsOn() []string {
    return []string{"ap.resolve_gl_accounts"}
}
func (s *BudgetCheckStage) DependsOn() []string {
    return []string{"ap.resolve_gl_accounts"}
}
// Both depend on gl_accounts but not on each other → can run in parallel
```

**Parallel execution within a priority band:**

```go
func (p *Pipeline) executeBand(opCtx *OperationContext, stages []Stage) error {
    parallelGroups := p.groupIndependent(stages, opCtx)

    for _, group := range parallelGroups {
        if len(group) == 1 {
            p.runStage(opCtx, group[0])
            continue
        }

        var wg sync.WaitGroup
        errs := make([]error, len(group))

        for i, stage := range group {
            wg.Add(1)
            go func(idx int, s Stage) {
                defer wg.Done()
                // Each goroutine writes to its own key prefix — no collision
                errs[idx] = p.runStage(opCtx, s)
            }(i, stage)
        }
        wg.Wait()

        for _, err := range errs {
            if err != nil { return err }
        }
    }
    return nil
}
```

### 12.3 Stage Result Memoisation

For idempotent read-only stages that are called multiple times within the same request (e.g., a vendor lookup called by both `ap.resolve_vendor` and `compliance.sanctions_check`), memoisation avoids redundant DB calls:

```go
// A stage opts into memoisation by implementing Cacheable
type Cacheable interface {
    CacheKey(opCtx *OperationContext) string  // returns "" to disable caching
    CacheTTL() time.Duration
}

// In the pipeline executor:
if cacheable, ok := stage.(Cacheable); ok {
    key := cacheable.CacheKey(opCtx)
    if key != "" {
        if cached, found := opCtx.stageResultCache[key]; found {
            for k, v := range cached { opCtx.SetData(k, v) }
            continue
        }
    }
}
// After stage runs: store in opCtx.stageResultCache
```

### 12.4 Context Object Pooling

`OperationContext` objects are allocated per request. Under high concurrency this creates GC pressure. Pool them:

```go
// awo/internal/pipeline/context.go

var opCtxPool = sync.Pool{
    New: func() any {
        return &OperationContext{
            Data:  make(map[string]any, 32),
            Flags: make(map[string]bool, 16),
            Log:   make([]StageLog, 0, 16),
        }
    },
}

func AcquireOperationContext() *OperationContext {
    ctx := opCtxPool.Get().(*OperationContext)
    ctx.reset()
    return ctx
}

func ReleaseOperationContext(ctx *OperationContext) {
    ctx.reset()
    opCtxPool.Put(ctx)
}

func (o *OperationContext) reset() {
    o.OperationID   = uuid.Nil
    o.Suspended     = false
    o.SuspendReason = ""
    o.ResumePoint   = ""
    o.DryRun        = false
    for k := range o.Data  { delete(o.Data, k) }
    for k := range o.Flags { delete(o.Flags, k) }
    o.Log = o.Log[:0]
}
```

### 12.5 Database Access Patterns

**Operation log writes are non-blocking:**

```go
// Fire-and-forget log write — does not block the pipeline response
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    p.repo.SaveLog(ctx, result)
}()
```

**Stage log entries are batched.** The individual `StageLog` entries accumulated in `opCtx.Log` are written once in a single UPDATE at the end of the pipeline, not after every stage:

```sql
-- Single write at pipeline completion: the full stage_log JSONB array
UPDATE operation_logs SET stage_log = $1, status = $2, completed_at = $3
WHERE operation_id = $4;
```

**For compensation events specifically**, write immediately and synchronously — you want the audit trail before anything else:

```go
func (p *Pipeline) logCompensation(opCtx *OperationContext, stageName, status string) {
    p.repo.AppendCompensationLog(opCtx.Ctx, domain.CompensationLog{
        OperationID: opCtx.OperationID,
        StageName:   stageName,
        Status:      status,
        OccurredAt:  time.Now(),
    })
}
```

---

## 13. API Handler Design

### 13.1 Synchronous vs Asynchronous Response

The pipeline can complete in milliseconds (minimal tenant, no workflow approval) or be suspended for days (full enterprise, multi-level approval). The HTTP layer must handle both:

```
Short pipelines (< 2 seconds expected):
  → Run synchronously
  → HTTP 200 with full result

Long pipelines or unknown duration:
  → Run synchronously up to a configurable deadline (e.g., 8 seconds)
  → If not complete: persist state, return HTTP 202 Accepted with operation_id
  → Client polls GET /operations/{id}/status

Suspended pipelines (pending approval):
  → Always return HTTP 202 immediately after workflow hook fires
  → Never make the HTTP client wait for human approval
```

```go
// awo/internal/api/handlers/operation_handler.go

const pipelineSyncDeadline = 8 * time.Second

func ProcessInvoiceHandler(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)

        var req ap.ProcessInvoiceRequest
        if err := c.BodyParser(&req); err != nil {
            return fiber.NewError(400, "invalid request body")
        }

        idempotencyKey := c.Get("Idempotency-Key")

        opCtx := pipeline.AcquireOperationContext()
        defer pipeline.ReleaseOperationContext(opCtx)
        opCtx.Populate(c.Context(), session, "ap.invoice.process", req)

        resultCh := make(chan *pipeline.OperationResult, 1)
        errCh    := make(chan error, 1)

        go func() {
            result, err := deps.APService.ProcessInvoice(c.Context(), req, idempotencyKey)
            if err != nil { errCh <- err; return }
            resultCh <- result
        }()

        select {
        case result := <-resultCh:
            return handlePipelineResult(c, result)
        case err := <-errCh:
            return handlePipelineError(c, err)
        case <-time.After(pipelineSyncDeadline):
            return c.Status(202).JSON(fiber.Map{
                "status":           "running",
                "operation_id":     opCtx.OperationID,
                "message":          "Operation is processing. Poll /operations/" + opCtx.OperationID.String() + "/status for updates.",
                "poll_interval_ms": 2000,
            })
        }
    }
}

func handlePipelineResult(c *fiber.Ctx, result *pipeline.OperationResult) error {
    switch result.Status {
    case "completed":
        return c.Status(200).JSON(fiber.Map{
            "status":       "completed",
            "operation_id": result.OperationID,
            "data":         result.Outputs,
            "duration_ms":  result.DurationMs,
        })
    case "pending_approval":
        return c.Status(202).JSON(fiber.Map{
            "status":       "pending_approval",
            "operation_id": result.OperationID,
            "approval_id":  result.Outputs["workflow.approval_id"],
            "message":      "Operation requires approval before completing.",
            "poll_url":     "/operations/" + result.OperationID.String() + "/status",
        })
    case "failed":
        return c.Status(422).JSON(fiber.Map{
            "status":       "failed",
            "operation_id": result.OperationID,
            "error":        result.Error,
            "failed_stage": result.FailedStage,
            "compensated":  result.Compensated,
            "stage_log":    result.Log,
        })
    }
    return nil
}
```

### 13.2 Status Polling Endpoint

```go
// GET /operations/:id/status
func OperationStatusHandler(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        operationID, err := uuid.Parse(c.Params("id"))
        if err != nil { return fiber.NewError(400, "invalid operation ID") }

        session := middleware.ContextSession(c)
        log, err := deps.OperationLogRepo.GetByOperationID(c.Context(), operationID)
        if err != nil { return fiber.NewError(404, "operation not found") }

        if log.TenantID != session.TenantID && !session.IsPlatform() {
            return fiber.NewError(403, "access denied")
        }

        resp := fiber.Map{
            "operation_id": log.OperationID,
            "status":       log.Status,
            "started_at":   log.StartedAt,
            "completed_at": log.CompletedAt,
            "duration_ms":  log.DurationMs,
            "stage_log":    log.StageLog,
        }

        if log.Status == "pending_approval" {
            resp["suspend_reason"]    = log.SuspendReason
            resp["approval_id"]       = log.OutputData["workflow.approval_id"]
            resp["poll_url"]          = "/operations/" + operationID.String() + "/status"
            resp["poll_interval_ms"]  = 5000
        }
        if log.Status == "completed" {
            resp["data"] = log.OutputData
        }
        if log.Status == "failed" {
            resp["error"]       = log.ErrorMessage
            resp["failed_stage"]= log.FailedStage
            resp["compensated"] = log.OutputData["compensated"]
        }

        return c.JSON(resp)
    }
}
```

### 13.3 WebSocket Real-time Updates

For the UI to show live stage progress without polling, a WebSocket channel broadcasts stage completion events:

```go
// awo/internal/api/handlers/operation_ws.go

// GET /operations/:id/live (WebSocket upgrade)
func OperationLiveHandler(deps *app.Deps) fiber.Handler {
    return websocket.New(func(c *websocket.Conn) {
        operationID, _ := uuid.Parse(c.Params("id"))

        sub := deps.EventBus.Subscribe(
            "operation.stage.completed:"+operationID.String(),
            "operation.completed:"+operationID.String(),
            "operation.failed:"+operationID.String(),
            "operation.suspended:"+operationID.String(),
        )
        defer deps.EventBus.Unsubscribe(sub)

        for {
            select {
            case event := <-sub.Channel():
                if err := c.WriteJSON(event); err != nil { return }
                if isTerminal(event.Type) { return }
            case <-time.After(30 * time.Second):
                c.WriteMessage(websocket.PingMessage, nil)
            }
        }
    })
}
```

**Events emitted by the pipeline onto the event bus:**

```go
// In pipeline.Execute(), after each stage:
p.events.Emit(opCtx.Ctx, domain.StageCompleted{
    OperationID: opCtx.OperationID,
    TenantID:    opCtx.TenantID,
    StageName:   stage.Name(),
    Status:      result.Status,
    DurationMs:  int(stageLog.Duration.Milliseconds()),
    OccurredAt:  time.Now(),
})
```

### 13.4 Idempotency at the HTTP Layer

```go
// Middleware: validate and track Idempotency-Key header
func IdempotencyMiddleware(repo OperationLogRepository) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := c.Get("Idempotency-Key")
        if key == "" { return c.Next() }

        if len(key) > 128 {
            return fiber.NewError(400, "Idempotency-Key must be ≤ 128 characters")
        }

        existing, err := repo.GetByIdempotencyKey(c.Context(), getTenantID(c), key)
        if err == nil {
            c.Set("Idempotency-Key-Status", "replayed")
            return c.Status(existing.HTTPStatus).JSON(existing.HTTPResponseBody)
        }

        c.Locals("idempotency_key", key)
        return c.Next()
    }
}
```

### 13.5 Error Response Schema

Every pipeline error response follows the same structure so the UI can render it consistently:

```json
{
  "status": "failed",
  "operation_id": "a1b2c3d4-...",
  "error": {
    "code": "BUDGET_EXCEEDED",
    "message": "Invoice amount KES 750,000 exceeds available budget KES 500,000",
    "stage": "budget.check",
    "recoverable": false,
    "user_action_required": "Request a budget override from your Finance Controller"
  },
  "compensated": true,
  "compensation_log": [
    {"stage": "gl.post_transaction", "status": "compensated"},
    {"stage": "dms.archive",         "status": "compensated"}
  ],
  "stage_log": [
    {"stage": "ap.validate_invoice", "status": "ran",    "duration_ms": 12},
    {"stage": "budget.check",        "status": "failed", "duration_ms": 45}
  ]
}
```

---

## 14. UI Design — Operation Dashboard

### 14.1 Pipeline Visualiser (Live)

The operation detail page shows a live pipeline visualisation. Each stage is a card. Status, duration, and output are revealed as stages complete:

```json
{
  "type": "page",
  "title": "Operation Progress",
  "body": {
    "type": "service",
    "api": "/api/operations/${operationId}",
    "ws": "/api/operations/${operationId}/live",
    "body": [
      {
        "type": "panel",
        "title": "${operation_key} — ${status}",
        "headerClassName": "${status === 'completed' ? 'bg-success' : status === 'failed' ? 'bg-danger' : 'bg-info'} text-white",
        "body": [
          {
            "type": "progress",
            "value": "${completedStages / totalStages * 100}",
            "showLabel": true,
            "className": "m-b-md"
          },
          {
            "type": "each",
            "name": "stage_log",
            "items": {
              "type": "flex",
              "className": "m-b-sm p-sm border rounded",
              "style": {
                "borderColor": "${status === 'ran' ? '#52c41a' : status === 'failed' ? '#f5222d' : status === 'skipped' ? '#d9d9d9' : '#1890ff'}",
                "borderLeftWidth": "4px"
              },
              "items": [
                {
                  "type": "icon",
                  "icon": "${status === 'ran' ? 'fa fa-check-circle' : status === 'failed' ? 'fa fa-times-circle' : status === 'skipped' ? 'fa fa-minus-circle' : 'fa fa-spinner fa-spin'}",
                  "className": "text-xl m-r-sm",
                  "style": {"color": "${status === 'ran' ? '#52c41a' : status === 'failed' ? '#f5222d' : '#d9d9d9'}"}
                },
                {
                  "type": "container",
                  "className": "flex-1",
                  "body": [
                    {"type": "tpl", "tpl": "<strong>${stage_name}</strong> <span class='text-muted text-sm m-l-sm'>${duration_ms}ms</span>"},
                    {"type": "tpl", "tpl": "<div class='text-sm text-muted'>${skip_reason || message || ''}</div>"}
                  ]
                }
              ]
            }
          }
        ]
      },
      {
        "type": "alert",
        "level": "warning",
        "visibleOn": "${status === 'pending_approval'}",
        "body": "This operation is waiting for approval. You will be notified when it is approved or rejected.",
        "showIcon": true
      },
      {
        "type": "panel",
        "title": "Error Details",
        "visibleOn": "${status === 'failed'}",
        "className": "border-danger",
        "body": [
          {"type": "tpl", "tpl": "<p class='text-danger'><strong>${error.message}</strong></p>"},
          {"type": "tpl", "tpl": "<p class='text-muted'>${error.user_action_required}</p>"},
          {
            "type": "button-toolbar",
            "buttons": [
              {
                "type": "button",
                "label": "Retry Operation",
                "level": "warning",
                "visibleOn": "${error.recoverable}",
                "actionType": "ajax",
                "api": {"method": "post", "url": "/api/operations/${operation_id}/retry"}
              },
              {
                "type": "button",
                "label": "View Compensation Log",
                "level": "default",
                "visibleOn": "${compensated}",
                "actionType": "dialog",
                "dialog": {
                  "title": "Compensation Log",
                  "body": {
                    "type": "each",
                    "name": "compensation_log",
                    "items": {
                      "type": "tpl",
                      "tpl": "<div class='p-xs'><i class='fa ${status === \"compensated\" ? \"fa-undo text-info\" : \"fa-exclamation-triangle text-danger\"}'></i> ${stage} — ${status}</div>"
                    }
                  }
                }
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### 14.2 Dry-Run Preview UI

Before submitting a real operation, a tenant can preview which stages will run:

```json
{
  "type": "dialog",
  "title": "Preview — What will happen",
  "size": "lg",
  "body": {
    "type": "service",
    "api": {
      "method": "post",
      "url": "/api/operations/dry-run",
      "data": {"operation_key": "ap.invoice.process", "input": "${$$}"}
    },
    "body": [
      {
        "type": "alert",
        "level": "info",
        "body": "This is a preview. No data will be saved. Stages marked as simulated will show expected behaviour.",
        "showIcon": true,
        "className": "m-b-md"
      },
      {
        "type": "each",
        "name": "stages",
        "items": {
          "type": "flex",
          "className": "m-b-xs p-xs border rounded",
          "items": [
            {
              "type": "icon",
              "icon": "${will_run ? 'fa fa-play-circle text-success' : 'fa fa-ban text-muted'}",
              "className": "m-r-sm"
            },
            {
              "type": "container",
              "className": "flex-1",
              "body": [
                {"type": "tpl", "tpl": "<strong>${stage_name}</strong> ${feature_flag ? '<span class=\"label label-info m-l-xs\">' + feature_flag + '</span>' : ''}"},
                {"type": "tpl", "tpl": "<div class='text-muted text-xs'>${skip_reason || 'Will execute'}</div>"},
                {
                  "type": "tpl",
                  "visibleOn": "${dry_run_output}",
                  "tpl": "<div class='text-sm m-t-xs p-xs bg-light rounded'>Expected: ${dry_run_output}</div>"
                }
              ]
            }
          ]
        }
      },
      {
        "type": "panel",
        "title": "Expected Outcome",
        "body": {
          "type": "property",
          "items": [
            {"label": "Stages that will run",    "content": "${stages_will_run}"},
            {"label": "Stages that will skip",   "content": "${stages_skipped}"},
            {"label": "Approval required",        "content": "${will_require_approval ? 'Yes' : 'No'}"},
            {"label": "Estimated duration",       "content": "${estimated_duration_ms}ms"}
          ]
        }
      }
    ]
  }
}
```

---

## 15. Tenant Scripting — The condition Package

### 15.1 Where condition Fits in the Pipeline

The `pkg/condition` package is directly usable in three places in the pipeline:

```
1. Script Stage (NEW)
   A stage whose Execute logic is a tenant-authored condition/formula
   evaluated against the OperationContext.

2. Condition-Gated Stage Execution (ENHANCES Stage interface)
   Every stage can declare a condition (not just a feature flag) that
   must be true for it to run. More expressive than a binary flag.

3. Hook Conditions (ENHANCES Hook interface)
   A hook fires only when its declared condition evaluates to true
   against the OperationContext at the time the hook fires.
```

### 15.2 Script Stage — Tenant-Defined Logic

A `ScriptStage` is a pipeline stage whose logic is entirely defined by the tenant in the UI and stored in the database. It uses the `condition` package to evaluate formulas safely:

```go
// awo/internal/pipeline/stages/script_stage.go

type ScriptStage struct {
    pipeline.BaseStage
    definition ScriptStageDefinition
    evaluator  *condition.Evaluator
}

type ScriptStageDefinition struct {
    ID          string
    Name        string
    Formula     string
    OnTrue      string   // next stage name if formula is true
    OnFalse     string   // next stage name if formula is false
    FeatureFlag string
    Priority    int
    Operations  []string
}

func NewScriptStage(def ScriptStageDefinition) *ScriptStage {
    return &ScriptStage{
        BaseStage: pipeline.BaseStage{
            name:        "script." + def.ID,
            operations:  def.Operations,
            featureFlag: def.FeatureFlag,
            priority:    def.Priority,
            required:    false,  // user scripts never block required stages
        },
        definition: def,
        // Each ScriptStage gets its own evaluator — one formula per evaluator instance.
        // Formula cache grows to exactly 1 entry then never grows again.
        evaluator: condition.NewEvaluator(nil, tenantScriptOptions),
    }
}

func (s *ScriptStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    evalCtx := condition.NewEvalContext(
        s.buildEvalData(opCtx),
        condition.DefaultEvalOptions(),
    )
    s.registerDomainFunctions(evalCtx, opCtx)

    builder := condition.NewBuilder(condition.ConjunctionAnd)
    builder.AddFormula(s.definition.Formula)
    group := builder.Build()

    if err := group.Rules[0].Validate(); err != nil {
        return pipeline.StageResult{
            Status:  "failed",
            Message: "Script validation error: " + err.Error(),
        }, nil
    }

    result, err := s.evaluator.Evaluate(opCtx.Ctx, group, evalCtx)
    if err != nil {
        switch err {
        case condition.ErrEvaluationTimeout:
            return pipeline.StageResult{Status: "failed", Message: "Script timed out (200ms limit)"}, nil
        case condition.ErrMaxDepthExceeded:
            return pipeline.StageResult{Status: "failed", Message: "Script too deeply nested"}, nil
        default:
            return pipeline.StageResult{Status: "failed", Message: "Script error: " + err.Error()}, nil
        }
    }

    metrics := evalCtx.GetMetrics()
    nextStage := s.definition.OnFalse
    if result { nextStage = s.definition.OnTrue }

    opCtx.SetData("script."+s.definition.ID+".result", result)
    opCtx.SetData("script."+s.definition.ID+".rules_evaluated", metrics.RulesEvaluated)

    return pipeline.StageResult{
        Status:      "completed",
        NextStageID: nextStage,
        Outputs: map[string]any{
            "script_result":   result,
            "rules_evaluated": metrics.RulesEvaluated,
            "duration_ms":     metrics.Duration.Milliseconds(),
        },
    }, nil
}

// buildEvalData flattens the OperationContext into the eval data map
func (s *ScriptStage) buildEvalData(opCtx *pipeline.OperationContext) map[string]any {
    data := make(map[string]any)
    for k, v := range opCtx.Data  { data[k] = v }
    for k, v := range opCtx.Flags { data[k] = v }
    data["feature"]   = opCtx.Session.Configuration.Flags
    data["setting"]   = opCtx.Session.Configuration.Settings
    data["entity_id"] = opCtx.EntityID.String()
    if input, ok := opCtx.Input.(map[string]any); ok {
        data["input"] = input
    }
    return data
}

// registerDomainFunctions makes ERP-specific functions available in scripts
func (s *ScriptStage) registerDomainFunctions(evalCtx *condition.EvalContext,
    opCtx *pipeline.OperationContext) {

    evalCtx.RegisterFunction("isWithinFXVariance",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            return true, nil
        })
    evalCtx.RegisterFunction("isWithinPeriod",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            return true, nil
        })
    evalCtx.RegisterFunction("isChildOf",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            return false, nil
        })
}
```

### 15.3 Condition-Gated Stage Execution

Every stage can declare a `RunCondition()` that is evaluated before the stage executes. This is more expressive than a binary feature flag:

```go
// In TaxCalculationStage:
func (s *TaxCalculationStage) RunCondition() string {
    // Skip tax calculation if vendor is marked as tax-exempt
    return `!(input.vendor_tax_exempt == true)`
}
```

The pipeline evaluates `RunCondition()` using the `condition` package before calling `Execute()`. See the `Pipeline.Execute()` implementation in section 5.1 for the full evaluation flow.

### 15.4 Custom Hook Conditions

Hooks can also carry conditions. The workflow approval hook should only fire when the operation amount exceeds the configured approval threshold:

```go
func (h *WorkflowApprovalHook) RunCondition() string {
    return `
        toNumber(gl_posting_amount) > toNumber(setting["workflow.approval_threshold_amount"])
        || budget_exceeded == true
        || input.entity_type == "subsidiary"
    `
}
```

### 15.5 Security Sandbox

Tenant scripts run in the condition package's sandbox. The following limits are enforced specifically for tenant-authored scripts (tighter than internal condition evaluation):

```go
var tenantScriptOptions = condition.EvalOptions{
    MaxDepth:      3,
    MaxConditions: 20,
    Timeout:       200 * time.Millisecond,
    CacheResults:  true,
    RegexTimeout:  50 * time.Millisecond,
}
```

**What tenant scripts CAN access:**
- All fields in `opCtx.Data` (stage outputs with `{stage}.{key}` prefix)
- `opCtx.Flags` (boolean signals set by stages)
- `feature.*` (their own feature flags)
- `setting.*` (their own tenant settings)
- `input.*` (the original operation input)
- Domain functions registered explicitly: `isWithinPeriod`, `isChildOf`, `isWithinFXVariance`

**What tenant scripts CANNOT access:**
- Raw DB queries (no database function registered)
- Other tenants' data (no cross-tenant function registered)
- File system, network, environment (expr-lang is not Turing-complete; no syscalls)
- Session internals (no `session.*` namespace exposed — permissions are not exposed to scripts)

### 15.6 The condition Package — Suitability Assessment

The `pkg/condition` package is **highly suitable** for the pipeline. The table below maps the condition package's features to pipeline use cases:

| condition Feature | Pipeline Use Case | Verdict |
|---|---|---|
| 17+ comparison operators | Stage `RunCondition()` expressions | ✅ Direct use |
| `ConjunctionAnd` / `ConjunctionOr` builder | Multi-clause stage gating | ✅ Direct use |
| `AddFormula(expr-lang)` | Tenant script stages | ✅ Core use |
| Custom functions | Domain-specific functions (isChildOf, etc.) | ✅ Register at startup |
| ABAC / field comparison | `AddFieldComparison("amount", "threshold")` in budget gating | ✅ Direct use |
| Dot-notation nested access | `opCtx.Data` contains nested maps → `budget_check.available_amount` | ✅ Works |
| Resource limits + timeout | Tenant script safety | ✅ Essential |
| Thread-safe `EvalContext` | Parallel stage evaluation | ✅ Required |
| Metrics (`RulesEvaluated`, `Duration`) | Logged in StageLog for debug | ✅ Valuable |
| Formula cache | High-frequency condition evaluation | ✅ Cache evaluators, not formulas |
| `ErrEvaluationTimeout` | Gracefully skip timed-out tenant script | ✅ Handle explicitly |

**One known limitation:** the formula cache has no eviction bound. Since tenant scripts may produce many unique formulas, create one `Evaluator` per `ScriptStage` definition (not one shared global evaluator). Each stage's evaluator handles only that stage's single formula — no unbounded growth.

---

## 16. Dry-Run Mode

### 16.1 How Dry-Run Works

A dry-run executes the full pipeline with all stages and hooks but wraps every external call in a mock that returns a plausible simulated result without writing anything to the database or calling external services.

The `DryRun` flag is set on the `OperationContext` before calling `pipeline.Execute()`. Every stage and hook that has side effects checks this flag and returns a simulated result instead.

```go
// Dry-run mode check in every stage that has side effects
func (s *GLPostingStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    if opCtx.DryRun {
        entries := s.buildEntries(opCtx)
        return pipeline.StageResult{
            Status:  "simulated",
            Message: fmt.Sprintf("Would post %d GL entries totalling %s", len(entries), totalAmount(entries)),
            Outputs: map[string]any{
                "gl_posting.simulated":    true,
                "gl_posting.entry_count":  len(entries),
                "gl_posting.total_amount": totalAmount(entries),
                "gl_posting.entries":      entries,
            },
        }, nil
    }
    // Real execution...
}
```

### 16.2 Stage Dry-Run Contract

Every stage must implement the `Simulatable` interface if it has side effects. Stages that are read-only (validation, duplicate check, tax calculation) can simply run normally in dry-run mode — they already have no side effects.

```go
// Stages with side effects implement Simulatable
type Simulatable interface {
    Simulate(opCtx *OperationContext) (StageResult, error)
}

// Pipeline checks Simulatable in dry-run mode
if opCtx.DryRun {
    if sim, ok := stage.(Simulatable); ok {
        result, err = sim.Simulate(opCtx)
    } else {
        result, err = stage.Execute(opCtx)  // read-only stages run normally
    }
}
```

**Implementations:**

```go
// GL posting simulation
func (s *GLPostingStage) Simulate(opCtx *OperationContext) (StageResult, error) {
    entries := s.buildEntries(opCtx)
    return StageResult{
        Status: "simulated",
        Outputs: map[string]any{
            "gl_posting.simulated":    true,
            "gl_posting.entries":      entries,
            "gl_posting.total_debit":  sumDebits(entries),
            "gl_posting.total_credit": sumCredits(entries),
        },
    }, nil
}

// Banking simulation
func (s *PaymentSchedulingStage) Simulate(opCtx *OperationContext) (StageResult, error) {
    invoice := opCtx.Input.(*ap.Invoice)
    dueDate := invoice.DueDate.Format("2006-01-02")
    return StageResult{
        Status: "simulated",
        Outputs: map[string]any{
            "payment_scheduling.simulated":     true,
            "payment_scheduling.amount":        invoice.TotalAmount,
            "payment_scheduling.scheduled_for": dueDate,
        },
    }, nil
}

// Workflow approval hook simulation
func (h *WorkflowApprovalHook) SimulateExecute(opCtx *OperationContext) error {
    config, _ := h.workflowSvc.GetRuleFor(opCtx.Ctx, workflow.RuleQuery{
        TenantID:  opCtx.TenantID,
        Operation: opCtx.OperationKey,
        Amount:    getAmount(opCtx),
    })
    opCtx.SetData("workflow.approval_required_in_real_run", config.RequiresApproval)
    opCtx.SetData("workflow.approvers_in_real_run", config.Approvers)
    // Does NOT call Suspend() — dry-run continues through the full pipeline
    return nil
}
```

### 16.3 Dry-Run API

```go
// POST /api/operations/dry-run
func DryRunHandler(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req DryRunRequest
        if err := c.BodyParser(&req); err != nil {
            return fiber.NewError(400, "invalid request body")
        }

        session := middleware.ContextSession(c)
        p, _ := deps.PipelineBuilder.Build(req.OperationKey, session)

        opCtx := pipeline.AcquireOperationContext()
        defer pipeline.ReleaseOperationContext(opCtx)
        opCtx.Populate(c.Context(), session, req.OperationKey, req.Input)
        opCtx.DryRun = true

        result, err := p.Execute(opCtx)
        if err != nil {
            return c.Status(200).JSON(fiber.Map{
                "dry_run":      true,
                "would_fail":   true,
                "error":        err.Error(),
                "failed_stage": result.FailedStage,
                "stage_log":    result.Log,
            })
        }

        return c.JSON(fiber.Map{
            "dry_run":                true,
            "would_succeed":          result.Status == "completed" || result.Status == "pending_approval",
            "would_require_approval": result.Outputs["workflow.approval_required_in_real_run"],
            "stages_that_would_run":  stagesWithStatus(result.Log, "ran", "simulated"),
            "stages_that_would_skip": stagesWithStatus(result.Log, "skipped"),
            "expected_gl_entries":    result.Outputs["gl_posting.entries"],
            "expected_payment_date":  result.Outputs["payment_scheduling.scheduled_for"],
            "stage_log":              result.Log,
        })
    }
}
```

---

## 17. Customer Workflow Engine — Synthesis

### 17.1 Two Complementary Systems

The Pipeline Architecture and the Customer Workflow Engine solve different problems:

```
Pipeline Architecture               Customer Workflow Engine
────────────────────────────────    ────────────────────────────────────
Synchronous ERP operations          Long-running human workflows
Coded in Go, driven by modules      Declarative JSON, designed by users
Executes in seconds                 Can run for days or weeks
Automatic compensation on failure   Temporal's durable execution
Single HTTP request lifetime        Survives process restarts
Configuration-driven (flags)        Definition-driven (visual builder)
Technical users extend it           Business users build with it
```

They are not alternatives. They are complementary. The Pipeline runs an AP invoice posting. The Workflow Engine handles the approval that gates that posting. They connect via the `WorkflowApprovalHook`.

### 17.2 What We Borrow from the Workflow Engine

**Borrow completely — no conflict:**

1. **The visual workflow builder UI (AMIS schemas).** The AMIS designer, template list, instance tracker, and task inbox are adopted verbatim. They define the UX for tenants creating approval workflows.

2. **The database schema** for `workflow_templates`, `workflow_instances`, `workflow_step_executions`, `workflow_user_tasks`, `workflow_triggers`, `workflow_variables`, `workflow_action_definitions`. These tables are the persistence layer for the Workflow Engine; they do not conflict with `operation_logs` which is the Pipeline's persistence layer.

3. **The Temporal integration.** Temporal worker, `CustomWorkflowExecution` workflow, activities, signal/resume pattern — all adopted for the Workflow Engine's long-running operations. The Pipeline Architecture does not use Temporal (it runs synchronously).

4. **The JSONPath expression language** for workflow step configurations. This is the tenant-facing expression language for the visual builder. It is separate from the condition package which is the Go-level evaluation engine.

5. **Pre-built workflow templates** (Invoice Approval, Employee Onboarding, Document Review). These are seeded into `workflow_templates` and are immediately available.

6. **The step type vocabulary** (validation, condition, user_task, action, notification, parallel, loop, wait). These become the building blocks tenants use in the visual designer.

7. **The task assignment strategies** (by user, by role, by role-in-entity). These map directly to the IAM module's entity hierarchy.

**Borrow with adaptation:**

8. **Step executors become Pipeline Stages.** The `ValidationExecutor`, `ConditionExecutor`, `UserTaskExecutor`, etc. from the Workflow Engine are adapted as Pipeline Stages (`WorkflowValidationStage`, `WorkflowUserTaskStage`). This allows the same executor logic to be used both in Pipeline operations and in Temporal activities.

**Do not borrow — keep separate:**

9. **The Workflow Engine's own service.go and routing.** The Workflow Engine is a separate module (`awo/internal/workflow`). It is NOT integrated into the Pipeline's `PipelineBuilder`. It has its own routes and handler layer.

10. **The Temporal worker for the Pipeline.** The Pipeline Architecture is synchronous and does not need Temporal. Temporal is only for the Workflow Engine's long-running human-in-the-loop flows.

### 17.3 The Unified Trigger Model

Both systems need to react to ERP events. The unified trigger model:

```
ERP Event (e.g., invoice.created)
        │
        ▼
Domain Event Bus
        │
        ├──► Pipeline:  APService.ProcessInvoice()
        │    Runs synchronously, completes in seconds
        │    If approval needed → emits WorkflowApprovalHook → suspends
        │
        └──► Workflow Engine trigger check
             If trigger_event = "invoice.created" AND conditions match
             → starts WorkflowInstance (Temporal)
             → this is for ADDITIONAL business process (e.g., new vendor onboarding)
             not the same as the approval gate inside ProcessInvoice
```

A single ERP event can trigger both a Pipeline operation AND a Workflow Engine instance. They are independent and complementary.

### 17.4 Workflow Engine as a Pipeline Stage

The `WorkflowApprovalHook` is the connection point. When the hook fires, it starts a Temporal workflow via the Workflow Engine's `StartWorkflow()` API and suspends the pipeline. When the Temporal workflow completes, it calls the pipeline's resume endpoint:

```go
// awo/internal/workflow/pipeline_hook.go

func (h *WorkflowApprovalHook) Execute(opCtx *pipeline.OperationContext) error {
    if !config.RequiresApproval { return nil }

    instance, err := h.workflowEngine.StartWorkflow(opCtx.Ctx,
        h.approvalTemplateID,
        map[string]any{
            "operation_id":      opCtx.OperationID,
            "operation_key":     opCtx.OperationKey,
            "approvers":         config.Approvers,
            "resume_url":        "/internal/pipeline/resume/" + opCtx.OperationID.String(),
            "resource_snapshot": opCtx.Input,
            "deadline":          config.ApprovalDeadline,
        })
    if err != nil { return err }

    opCtx.SetData("workflow.instance_id", instance.ID)
    opCtx.SetData("workflow.approval_id",  instance.ID)

    opCtx.Suspend("pending_workflow_approval:"+instance.ID.String(), "gl.post_transaction")
    return nil
}
```

**The Temporal workflow calls back:**

```go
// In the Workflow Engine's completion activity:
func (a *Activities) CompleteWorkflowInstanceActivity(ctx context.Context,
    instanceID uuid.UUID, output map[string]any) error {

    if operationID, ok := output["operation_id"].(string); ok {
        opUUID, _ := uuid.Parse(operationID)
        decision := output["decision"].(string)
        a.pipelineResumeService.Resume(ctx, opUUID, map[string]any{
            "approved_by": output["completed_by"],
            "decision":    decision,
            "comment":     output["comment"],
        })
    }
    return nil
}
```

### 17.5 Shared Schema Between Both Systems

Both systems write to and read from shared tables:

```
workflow_user_tasks ← written by WorkflowApprovalHook (via Workflow Engine)
                    ← read by Task Inbox UI (My Tasks page)
                    ← completed by user → signals Temporal → Pipeline resumes

operation_logs     ← written by Pipeline
                   ← read by Operation Status UI
                   ← resume endpoint reads to reconstruct OperationContext

entities           ← used by both for entity-scoped task assignment
users              ← used by both for assignment, audit, completion tracking
```

### 17.6 What Stays Separate

```
Pipeline Architecture               Customer Workflow Engine
────────────────────────────────    ────────────────────────────────────
operation_logs                      workflow_templates
pipeline.Stage interface            workflow_instances
pipeline.Hook interface             workflow_step_executions
StageRegistry                       workflow_triggers
HookRegistry                        workflow_variables
CompensationRegistry                workflow_action_definitions
PipelineBuilder                     Temporal worker + activities
OperationContext                    WorkflowDefinition JSON
awo/internal/pipeline/*             awo/internal/workflow/*
```

Neither system imports the other directly. They communicate through:
1. The shared database tables (`workflow_user_tasks`, `entities`, `users`)
2. The event bus (`operation.suspended` → Workflow Engine starts; `workflow.completed` → Pipeline resumes)
3. The HTTP resume endpoint (`/internal/pipeline/resume/:id`)

**Full integration flow quick reference:**

```
tenant enables "workflow" flag
        ↓
WorkflowApprovalHook registered in HookRegistry
        ↓
APService.ProcessInvoice() called
        ↓
PipelineBuilder builds pipeline (includes hook)
        ↓
Pipeline executes up to "gl.post_transaction"
        ↓
WorkflowApprovalHook fires BEFORE gl.post_transaction
  → condition package evaluates: amount > threshold?
  → YES: WorkflowEngine.StartWorkflow(approval_template)
  → opCtx.Suspend()
        ↓
Pipeline persists state to operation_logs
        ↓
API returns 202 { status: "pending_approval", operation_id }
        ↓
Temporal picks up approval workflow
  → creates workflow_user_tasks (assigned to approvers)
  → sends notification
  → waits for signal
        ↓
Approver completes task in UI (My Tasks page)
        ↓
Temporal signal → workflow completes
  → CompleteWorkflowInstanceActivity calls PipelineResumeService.Resume()
        ↓
Pipeline resumes from "gl.post_transaction"
  → GLPostingStage executes (for real this time)
  → DMS, Banking, Audit stages run
  → Pipeline completes
        ↓
operation_logs status → "completed"
WebSocket event broadcast → UI updates live
```

---

## 18. All Planned Modules & Their Stage Contributions

Each module registers stages and hooks at startup. This table is the catalogue of what exists and what flag gates each contribution:

| Stage / Hook | Module | Operation(s) | Flag Gate | Priority | Required |
|---|---|---|---|---|---|
| `ap.validate_invoice` | AP | `ap.invoice.process` | — | 110 | Yes |
| `ap.duplicate_check` | AP | `ap.invoice.process` | — | 120 | Yes |
| `ap.resolve_vendor` | AP | `ap.invoice.process` | — | 210 | Yes |
| `ap.resolve_gl_accounts` | AP | `ap.invoice.*` | — | 220 | Yes |
| `procurement.three_way_match` | Procurement | `ap.invoice.process` | `procurement` | 300 | No |
| `procurement.validate_po` | Procurement | `procurement.po.*` | `procurement` | 110 | Yes |
| `budget.check` | Budget | `ap.invoice.*`, `gl.journal.*` | `budget` | 310 | No |
| `budget.reserve` | Budget | `procurement.po.create` | `budget` | 320 | No |
| `compliance.sanctions_check` | Compliance | `ap.invoice.*` | `compliance` | 390 | No |
| `tax.resolve_codes` | Tax | `ap.invoice.*`, `ar.invoice.*` | `tax` | 400 | No |
| `tax.calculate` | Tax | `ap.invoice.*`, `ar.invoice.*` | `tax` | 410 | No |
| `tax.withholding` | Tax | `ap.invoice.process` | `tax.withholding` | 430 | No |
| `tax.e_invoice_validate` | Tax | `ar.invoice.post` | `tax.e_invoice` | 420 | No |
| `gl.post_transaction` | Finance | `ap.invoice.*`, `ar.invoice.*`, `inventory.*` | — | 610 | Yes |
| `ar.apply_advance` | AR | `ap.invoice.process` | `ar.advances` | 620 | No |
| `ar.apply_credit_note` | AR | `ap.invoice.process` | `ar` | 630 | No |
| `inventory.update_grn` | Inventory | `ap.invoice.process` | `inventory` | 730 | No |
| `dms.archive` | DMS | `*` | `dms` | 710 | No |
| `banking.schedule_payment` | Banking | `ap.invoice.process` | `banking` | 720 | No |
| `banking.reconcile` | Banking | `banking.statement.import` | `banking` | 610 | Yes |
| `audit.log_operation` | Audit | `*` | — | 910 | Yes |
| `compliance.regulatory_report` | Compliance | `ap.invoice.*` | `compliance` | 950 | No |
| **HOOKS** | | | | | |
| `workflow.approval_gate` | Workflow | `*` | `workflow` | 100 | — |
| `audit.log_hook` | Audit | `*` | — | 999 | — |
| `notify.posting` | Notifications | `*` | `notifications` | 100 | — |
| `tax.e_invoice_submit` | Tax | `ar.invoice.*` | `tax.e_invoice` | 50 | — |
| `compliance.access_log` | Compliance | `*` | `compliance` | 999 | — |
| `reporting.refresh_kpi` | Reporting | `gl.journal.post`, `ap.invoice.process` | `reporting` | 100 | — |

---

## 19. Adding a New Stage or Module

### 19.1 The Contract for New Module Developers

Adding a new module requires:

1. **Implement the `Stage` interface** for each pipeline step the module contributes.
2. **Implement the `Hook` interface** for any extension points (before/after existing stages).
3. **Register in `wire.go`** — one line per stage, one line per hook.
4. **Write tests for each stage in isolation** — no pipeline required.
5. **No changes to existing service code.** If you need to change `APService.ProcessInvoice` to add your module, the design has been violated.

### 19.2 Example: Adding the DMS Module Later

```go
// awo/internal/dms/pipeline_stages.go

type ArchiveDocumentStage struct {
    pipeline.BaseStage
    dmsSvc DMSService
}

func NewArchiveDocumentStage(svc DMSService) *ArchiveDocumentStage {
    return &ArchiveDocumentStage{
        BaseStage: pipeline.BaseStage{
            name:        "dms.archive",
            operations:  []string{"ap.invoice.process", "ar.invoice.post", "gl.journal.post"},
            featureFlag: "dms",
            priority:    710,
            required:    false,
        },
        dmsSvc: svc,
    }
}

func (s *ArchiveDocumentStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    glTxID, ok := opCtx.GetData("gl_posting.transaction_id")
    if !ok {
        return pipeline.StageResult{Status: "skipped", Message: "no GL transaction to archive"}, nil
    }

    docID, err := s.dmsSvc.Archive(opCtx.Ctx, dms.ArchiveParams{
        TenantID:     opCtx.TenantID,
        DocumentType: opCtx.Resource,
        SourceID:     glTxID.(uuid.UUID),
        Document:     opCtx.Input,
        Tags:         []string{opCtx.OperationKey, opCtx.EntityID.String()},
    })
    if err != nil {
        return pipeline.StageResult{
            Status:  "failed",
            Message: "DMS archive failed: " + err.Error(),
        }, nil  // return nil error to continue pipeline
    }

    opCtx.SetData("dms_archive.document_id", docID)
    return pipeline.StageResult{
        Status:  "completed",
        Outputs: map[string]any{"document_id": docID},
    }, nil
}

// Registration in wire.go — ONE NEW LINE:
// stageRegistry.Register(dms.NewArchiveDocumentStage(dmsSvc))
```

Nothing else changes. The DMS stage is now part of every applicable pipeline for every tenant with the `dms` flag enabled.

### 19.3 Adding the Future Workflow Module

```go
// wire.go — register the hook when workflow module is ready:
// hookRegistry.Register(workflow.NewWorkflowApprovalHook(workflowSvc))

// That's it. No changes to APService, InventoryService, or GLService.
```

---

## 20. Testing Strategy

### 20.1 Stage Tests (Isolated — No Pipeline)

Each stage is independently testable. You do not need to construct a full pipeline:

```go
// awo/internal/budget/pipeline_stages_test.go

func TestBudgetCheckStage_HardBlock_WhenExceeded(t *testing.T) {
    budgetSvc := &mockBudgetService{
        result: budget.CheckResult{WithinBudget: false, AvailableAmount: dec("50000")},
    }
    stage := budget.NewBudgetCheckStage(budgetSvc)

    session := buildTestSession(t, map[string]bool{"budget": true},
        map[string]string{"budget.control_mode": "hard_block"})

    opCtx := &pipeline.OperationContext{
        Session: session,
        Input:   &ap.Invoice{TotalAmount: dec("75000"), CostCenterID: uuid.New()},
        Data:    make(map[string]any),
        Flags:   make(map[string]bool),
    }

    _, err := stage.Execute(opCtx)
    assert.ErrorContains(t, err, "budget exceeded")
}

func TestBudgetCheckStage_Warn_WhenExceeded(t *testing.T) {
    session := buildTestSession(t, map[string]bool{"budget": true},
        map[string]string{"budget.control_mode": "warn"})
    opCtx := buildOpCtx(t, session)

    result, err := stage.Execute(opCtx)
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.Status)
    assert.True(t, opCtx.Flags["budget_exceeded"])
    assert.False(t, opCtx.Suspended)
}
```

### 20.2 Pipeline Tests (Integration)

```go
// awo/internal/pipeline/pipeline_test.go

func TestPipeline_APInvoiceProcess_MinimalTenant(t *testing.T) {
    registry := buildTestRegistry(t,
        ap.NewValidateInvoiceStage(),
        ap.NewDuplicateCheckStage(),
        finance.NewGLPostingStage(mockGL),
        audit.NewAuditLogStage(mockAudit),
    )

    session := buildSessionWithFlags(t, map[string]bool{
        "finance": true,
    })

    builder := pipeline.NewPipelineBuilder(registry, emptyHooks, logRepo)
    p, _ := builder.Build("ap.invoice.process", session)

    assert.Len(t, p.Stages(), 4)
}

func TestPipeline_APInvoiceProcess_SuspendsOnWorkflowApproval(t *testing.T) {
    workflowHook := &mockWorkflowHook{requiresApproval: true, approvalID: uuid.New()}
    hookRegistry := buildTestHookRegistry(t, workflowHook)

    session := buildSessionWithFlags(t, map[string]bool{
        "finance": true, "workflow": true,
    })
    p, _ := builder.Build("ap.invoice.process", session)
    result, err := p.Execute(opCtx)

    assert.NoError(t, err)
    assert.Equal(t, "pending_approval", result.Status)
    assert.True(t, opCtx.Suspended)
    assert.NotEmpty(t, opCtx.SuspendReason)
    assert.False(t, hasStageLog(opCtx, "gl.post_transaction"))
}

func TestPipeline_Resume_AfterApproval(t *testing.T) {
    // 1. Suspend the pipeline
    result, _ := p.Execute(opCtx)
    assert.Equal(t, "pending_approval", result.Status)

    // 2. Resume after approval
    resumeSvc := pipeline.NewResumeService(builder, logRepo)
    err := resumeSvc.Resume(ctx, opCtx.OperationID, map[string]any{
        "approved_by": cfoUserID, "decision": "approved",
    })
    assert.NoError(t, err)

    // 3. Verify GL was posted after resume
    reloaded, _ := logRepo.GetByOperationID(ctx, opCtx.OperationID)
    assert.Equal(t, "completed", reloaded.Status)
    assert.NotNil(t, reloaded.OutputData["gl_posting.transaction_id"])
}
```

### 20.3 Context Isolation Tests

```go
func TestOperationContext_StageDataIsolation(t *testing.T) {
    opCtx := &pipeline.OperationContext{Data: make(map[string]any)}

    opCtx.SetData("stage_a.result", "hello")

    val, ok := opCtx.GetData("stage_a.result")
    assert.True(t, ok)
    assert.Equal(t, "hello", val)

    _, ok = opCtx.GetData("stage_a.internal_only")
    assert.False(t, ok)
}
```

---

## 21. Troubleshooting

### 21.1 Why Was a Stage Skipped?

```sql
SELECT
  stage_log->>'StageName'  AS stage,
  stage_log->>'Status'     AS status,
  stage_log->>'SkipReason' AS skip_reason,
  stage_log->>'Duration'   AS duration_ns
FROM operation_logs,
  jsonb_array_elements(stage_log) AS stage_log
WHERE operation_id = '<operation_uuid>'
ORDER BY stage_log->>'StartedAt';
```

### 21.2 Pipeline Suspended — Find All Pending

```sql
SELECT
  ol.operation_id,
  ol.operation_key,
  ol.suspend_reason,
  ol.resume_point,
  ol.started_at,
  u.email AS initiated_by
FROM operation_logs ol
JOIN users u ON u.id = ol.initiated_by
WHERE ol.tenant_id = '<tenant_uuid>'
  AND ol.status = 'pending_approval'
ORDER BY ol.started_at;
```

### 21.3 Stage Registered But Not Running

```
Cause: Stage is registered but its Operations() list does not include
the operation being executed.

Fix: Add the operation key to the stage's Operations() slice.

Cause: Stage has a FeatureFlag that is off for this tenant.
The PipelineBuilder excludes it at build time.

Check:
  SELECT flag_key, COALESCE(tff.enabled, ffd.default_value)
  FROM feature_flag_definitions ffd
  LEFT JOIN tenant_feature_flags tff ON tff.flag_id = ffd.id
  WHERE ffd.flag_key = '<your_flag>';

Cause: Stage's RunCondition() evaluated to false for this execution.
Check the stage_log entry — SkipReason will say "run condition evaluated to false".
```

### 21.4 Hook Not Firing

```
Cause: Hook's StageBefore()/StageAfter() contains a stage name that
does not match any stage in the pipeline (typo or stage not registered).

Debug: Print the pipeline stages at build time in non-production:
  fmt.Println(p.StageNames())

Cause: Hook's FeatureFlag is off for the tenant.

Cause: Hook's Operations() does not include the current operation key.
```

### 21.5 Common Error Reference

| Error | Cause | Fix |
|---|---|---|
| `unexpected input type` | Stage cast of opCtx.Input to wrong type | Verify the service passes the correct input type |
| `operation not pending approval` | Resume called on wrong operation | Check operation status before resuming |
| `stage not found for operation` | Stage.Operations() mismatch | Add operation key to stage registration |
| `nil pointer in opCtx.Data` | Stage reads key before it is written | Check stage execution order; verify earlier stage ran |
| `pipeline: no required stages` | All stages excluded by flags | At minimum, validate + GL post stages must have no flag gate |
| `Script timed out (200ms limit)` | Tenant script exceeds execution budget | Simplify the formula; check for regex backtracking |
| `Script too deeply nested` | Tenant script exceeds `MaxDepth: 3` | Flatten condition groups; reduce nesting |
| `compensation FAILED` | Compensating function itself errored | Requires manual resolution; check `CompensationFailed` events |
