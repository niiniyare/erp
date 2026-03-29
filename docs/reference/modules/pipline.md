# Pipeline & Workflow Engine — Complete Technical Guide

> **Packages:** `awo/internal/pipeline` · `awo/internal/workflow` · `awo/pkg/condition`  
> **Audience:** Platform engineers, module developers, tenant administrators, and business analysts.  
> **Scope:** End-to-end design of the operation pipeline, the customer workflow engine, DB-transaction-aware side workflows, tenant scripting with the condition package, autocomplete for user-defined expressions, and UI integration patterns.

---

## Table of Contents

1. [Architecture Philosophy](#1-architecture-philosophy)
2. [Core Concepts Glossary](#2-core-concepts-glossary)
3. [OperationContext — The Carrier](#3-operationcontext--the-carrier)
4. [Stage — The Unit of Work](#4-stage--the-unit-of-work)
5. [Pipeline & PipelineBuilder](#5-pipeline--pipelinebuilder)
6. [Hook System](#6-hook-system)
7. [Stage & Hook Registries](#7-stage--hook-registries)
8. [DB Transaction Hooks — Side Workflows](#8-db-transaction-hooks--side-workflows)
9. [Failure Recovery & Compensation](#9-failure-recovery--compensation)
10. [Performance Optimisation](#10-performance-optimisation)
11. [API Handler Design](#11-api-handler-design)
12. [Dry-Run Mode](#12-dry-run-mode)
13. [The pkg/condition Package](#13-the-pkgcondition-package)
14. [Tenant Scripting Engine](#14-tenant-scripting-engine)
15. [Autocomplete for User-Defined Expressions](#15-autocomplete-for-user-defined-expressions)
16. [Customer Workflow Engine](#16-customer-workflow-engine)
17. [Pipeline ↔ Workflow Engine Integration](#17-pipeline--workflow-engine-integration)
18. [Canonical Example: AP Invoice Processing](#18-canonical-example-ap-invoice-processing)
19. [All Module Stage Contributions](#19-all-module-stage-contributions)
20. [Testing Strategy](#20-testing-strategy)
21. [UI Design](#21-ui-design)
22. [Troubleshooting](#22-troubleshooting)

---

## 1. Architecture Philosophy

### 1.1 The Two Systems and Their Relationship

Awo ERP has two distinct but complementary systems for orchestrating complex operations:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PIPELINE ARCHITECTURE                                                       │
│  awo/internal/pipeline                                                       │
│                                                                              │
│  Purpose: Execute synchronous ERP operations that involve multiple           │
│           services, vary by tenant configuration, and need compensation      │
│           if they fail midway.                                               │
│                                                                              │
│  Characteristics:                                                            │
│  • Executes within a single HTTP request lifetime (< 30s typical)           │
│  • Stages are Go code; extended by module developers                         │
│  • Configuration-driven (feature flags + tenant settings)                   │
│  • Automatic compensation (saga pattern) on failure                          │
│  • DB transactions can spawn side workflows                                  │
│  • Tenant scripting via pkg/condition                                        │
│                                                                              │
│  Examples: Post AP invoice, approve GL journal, receive GRN, post payroll   │
└────────────────────────────────────┬────────────────────────────────────────┘
                                     │ WorkflowApprovalHook suspends pipeline
                                     │ and starts a Workflow Engine instance
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  CUSTOMER WORKFLOW ENGINE                                                    │
│  awo/internal/workflow                                                       │
│                                                                              │
│  Purpose: Execute long-running, human-in-the-loop business processes         │
│           that can span days or weeks and survive system restarts.           │
│                                                                              │
│  Characteristics:                                                            │
│  • Runs via Temporal (durable execution, survives restarts)                 │
│  • Declarative JSON definition; built by business users in visual designer  │
│  • Human tasks (approvals, reviews, decisions) with deadline + escalation   │
│  • Tenant-defined; no code deployment needed to add a workflow               │
│  • When complete: signals the Pipeline to resume                             │
│                                                                              │
│  Examples: Multi-level invoice approval, vendor onboarding, audit review    │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 The Unified Key Insight

Both systems obey the same contract: **the tenant's configuration defines what runs**. A tenant without the `budget` module enabled never sees a budget check stage. A tenant without the `workflow` flag never gets an approval gate. The session's pre-computed feature flags and settings (built at login — zero DB hits per request) drive which stages and hooks are included in every pipeline execution.

### 1.3 How They Connect

```
AP Invoice POST
       │
       ▼
PipelineBuilder.Build("ap.invoice.process", session)
       │
       │  Reads: StageRegistry + HookRegistry filtered by session.Flags
       ▼
Configured Pipeline executes stages 100→900
       │
       │  At priority 500-599 band: WorkflowApprovalHook fires
       │  → condition package evaluates: amount > threshold?
       │  → YES: WorkflowEngine.StartWorkflow() via Temporal
       │  → opCtx.Suspend() persists state to operation_logs
       │
       ▼
API returns 202 { operation_id, status: "pending_approval" }
       │
       │  Days later: approver acts in My Tasks UI
       │
       ▼
Temporal signal → workflow completes → PipelineResumeService.Resume()
       │
       ▼
Pipeline continues from gl.post_transaction → completes
       │
       ▼
WebSocket event → UI updates live → 200 completed
```

---

## 2. Core Concepts Glossary

| Term | Definition |
|---|---|
| **OperationContext** | The mutable carrier object passed through every stage and hook. Contains session, input, shared data map, flags map, and execution log. |
| **Stage** | A single independently-testable unit of work. Declares its own feature flag gate, priority, and compensation function. |
| **Pipeline** | An ordered list of stages configured at runtime based on feature flags. Never hardcoded. |
| **PipelineBuilder** | Constructs a pipeline for a given operation key by filtering the stage registry through the session's feature flags. |
| **StageRegistry** | Global catalogue of all stages across all modules. Modules register at startup. |
| **Hook** | An extension point that fires before or after a named stage. Used by Workflow Engine, Audit, Notifications, E-invoice. |
| **HookRegistry** | Global catalogue of all hooks. Same registration pattern as stages. |
| **CompensationFn** | The undo function for a stage that has side effects. Runs LIFO when the pipeline aborts after a required stage fails. |
| **OperationLog** | Persisted record of every pipeline execution: which stages ran, outputs, timing, suspension state. |
| **ScriptStage** | A pipeline stage whose logic is a tenant-authored formula evaluated by the condition package. |
| **WorkflowTemplate** | A declarative JSON definition of a business process. Built by tenants in the visual designer. |
| **WorkflowInstance** | A running or completed execution of a workflow template, backed by a Temporal workflow. |
| **UserTask** | A human-in-the-loop task (approval, review, decision) created by the Workflow Engine and managed in the My Tasks UI. |
| **TxHook** | A function registered against a DB transaction that fires within the same transaction boundary, allowing the pipeline to trigger side workflows atomically. |

---

## 3. OperationContext — The Carrier

### 3.1 Design

The `OperationContext` is not the standard `context.Context`. It is a domain-level object that carries everything any stage or hook might need. It is passed by pointer so stages can write to the shared state:

```go
// awo/internal/pipeline/context.go

type OperationContext struct {
    // ── Identity ──────────────────────────────────────────────────────────
    Ctx        context.Context
    Session    *domain.ResolvedSession
    TenantID   uuid.UUID
    UserID     uuid.UUID
    EntityID   uuid.UUID

    // ── Operation identity ────────────────────────────────────────────────
    OperationID   uuid.UUID
    OperationKey  string      // "ap.invoice.process"
    Resource      string      // "ap.invoice"
    Action        string      // "process"

    // ── Mode ─────────────────────────────────────────────────────────────
    DryRun bool   // when true: stages simulate without writing to DB

    // ── Input ────────────────────────────────────────────────────────────
    Input any     // typed by caller; stages cast to expected type

    // ── Shared mutable state ──────────────────────────────────────────────
    // Keys follow "{stage_name}.{key}" convention to avoid collisions:
    //   "tax_calculation.tax_amount"        → decimal.Decimal
    //   "budget_check.available_amount"     → decimal.Decimal
    //   "workflow.approval_id"              → uuid.UUID
    //   "gl_posting.transaction_id"         → uuid.UUID
    Data map[string]any

    // ── Decision flags ────────────────────────────────────────────────────
    Flags map[string]bool
    // "requires_approval": true, "tax_exempt": false, "budget_exceeded": true

    // ── Suspension ────────────────────────────────────────────────────────
    Suspended     bool
    SuspendReason string   // "pending_workflow_approval:approval_uuid"
    ResumePoint   string   // stage name to resume from after approval

    // ── Execution log ─────────────────────────────────────────────────────
    Log         []StageLog
    StartedAt   time.Time
    CompletedAt *time.Time

    // ── DB transaction hooks ──────────────────────────────────────────────
    // Hooks registered by stages that fire within the same DB transaction.
    // See Section 8 for details.
    TxHooks []TxHook
}

type StageLog struct {
    StageName  string
    Status     string       // "ran" | "skipped" | "suspended" | "failed" | "simulated"
    SkipReason string
    StartedAt  time.Time
    Duration   time.Duration
    Output     map[string]any
    Error      string
}

// Convenience accessors
func (o *OperationContext) GetData(key string) (any, bool)
func (o *OperationContext) SetData(key string, value any)
func (o *OperationContext) SetFlag(key string, value bool)
func (o *OperationContext) Flag(key string) bool
func (o *OperationContext) Can(resource, action string) bool
func (o *OperationContext) FeatureEnabled(key string) bool
func (o *OperationContext) SettingDecimal(key string, def decimal.Decimal) decimal.Decimal
func (o *OperationContext) SettingBool(key string, def bool) bool
func (o *OperationContext) SettingString(key string, def string) string
func (o *OperationContext) Suspend(reason, resumeAt string)
func (o *OperationContext) RegisterTxHook(hook TxHook)
```

### 3.2 Stage Communication Protocol

Stages communicate exclusively through `OperationContext.Data`. No direct service-to-service calls within a pipeline:

```go
// DON'T — couples stages directly
func (s *GLPostingStage) Execute(opCtx *OperationContext) (StageResult, error) {
    tax, _ := taxService.GetCalculated(ctx, invoiceID) // wrong
}

// DO — read from shared context
func (s *GLPostingStage) Execute(opCtx *OperationContext) (StageResult, error) {
    taxAmount, _ := opCtx.GetData("tax_calculation.tax_amount")
}
```

**Data key convention — standard prefixes:**

```
"tax_calculation.*"          → tax module outputs
"budget_check.*"             → budget module outputs
"three_way_match.*"          → procurement module outputs
"workflow.*"                 → workflow engine outputs
"gl_posting.*"               → finance GL outputs
"dms_archive.*"              → DMS outputs
"payment_scheduling.*"       → banking outputs
"script.{script_id}.*"       → tenant script stage outputs
```

### 3.3 Object Pooling

OperationContext objects are pooled to reduce GC pressure under high concurrency:

```go
var opCtxPool = sync.Pool{
    New: func() any {
        return &OperationContext{
            Data:  make(map[string]any, 32),
            Flags: make(map[string]bool, 16),
            Log:   make([]StageLog, 0, 16),
        }
    },
}

func AcquireOperationContext() *OperationContext { return opCtxPool.Get().(*OperationContext) }
func ReleaseOperationContext(ctx *OperationContext) { ctx.reset(); opCtxPool.Put(ctx) }
```

---

## 4. Stage — The Unit of Work

### 4.1 Stage Interface

```go
// awo/internal/pipeline/stage.go

type Stage interface {
    // Unique identifier: "{module}.{description}"
    // e.g. "ap.validate_invoice", "tax.calculate", "budget.check"
    Name() string

    // Operation keys this stage applies to. "*" = all operations.
    // e.g. []string{"ap.invoice.process", "ap.invoice.approve"}
    Operations() []string

    // Feature flag that gates this stage. Empty = always included.
    // The PipelineBuilder excludes the stage entirely if the flag is off.
    FeatureFlag() string

    // Execution order. Standard bands:
    //   100–199: Validation
    //   200–299: Enrichment / data fetching
    //   300–399: Business rule checks
    //   400–499: Tax & regulatory
    //   500–599: Workflow & approval gates
    //   600–699: Core financial posting
    //   700–799: Secondary financial & storage
    //   800–899: Events & notifications
    //   900–999: Audit & compliance
    Priority() int

    // Required = true: pipeline aborts and compensates if Execute returns error.
    // Required = false: error is logged, execution continues.
    Required() bool

    // RunCondition returns an expr-lang formula evaluated via pkg/condition
    // before Execute is called. Empty string = always run.
    // Formula has access to opCtx.Data and opCtx.Flags.
    RunCondition() string

    // DependsOn returns stage names whose output this stage reads.
    // The pipeline uses this to identify which stages in the same priority
    // band can run in parallel.
    DependsOn() []string

    Execute(opCtx *OperationContext) (StageResult, error)
}

// Stages with side effects also implement Simulatable for dry-run mode.
type Simulatable interface {
    Simulate(opCtx *OperationContext) (StageResult, error)
}

type StageResult struct {
    Status      string         // "completed" | "skipped" | "suspended" | "simulated" | "failed"
    Outputs     map[string]any
    Message     string
    NextStageID string         // for condition/script stages that branch
}
```

### 4.2 Priority Bands and Parallel Execution

Within a priority band, stages that share no `DependsOn()` can run concurrently:

```
Band 300-399 (Business rule checks):
  BudgetCheckStage    DependsOn: ["ap.resolve_gl_accounts"]
  SanctionsCheckStage DependsOn: ["ap.resolve_vendor"]
  
  → Neither depends on the other → run in parallel within band 300-399

Band 400-499 (Tax):
  TaxResolveCodes     DependsOn: ["ap.resolve_gl_accounts"]
  TaxCalculate        DependsOn: ["tax.resolve_codes"]   ← depends on previous tax stage
  
  → TaxCalculate must wait for TaxResolveCodes → sequential within band
```

### 4.3 Stage Example — Budget Check

```go
// awo/internal/budget/pipeline_stages.go

type BudgetCheckStage struct {
    pipeline.BaseStage
    budgetSvc BudgetService
    evaluator *condition.Evaluator
}

func NewBudgetCheckStage(svc BudgetService) *BudgetCheckStage {
    return &BudgetCheckStage{
        BaseStage: pipeline.BaseStage{
            name:        "budget.check",
            operations:  []string{"ap.invoice.process", "ap.invoice.approve"},
            featureFlag: "budget",
            priority:    310,
            required:    false,
            // RunCondition: only check if invoice amount is above configured minimum
            runCondition: `toNumber(input.total_amount) > toNumber(setting["budget.minimum_check_amount"] ?? "0")`,
        },
        budgetSvc: svc,
        evaluator: condition.NewEvaluator(nil, condition.EvalOptions{Timeout: 100*time.Millisecond}),
    }
}

func (s *BudgetCheckStage) DependsOn() []string {
    return []string{"ap.resolve_gl_accounts"}  // reads cost center
}

func (s *BudgetCheckStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    invoice := opCtx.Input.(*ap.Invoice)
    mode    := opCtx.SettingString("budget.control_mode", "warn")

    result, err := s.budgetSvc.Check(opCtx.Ctx, budget.CheckParams{
        TenantID:     opCtx.TenantID,
        CostCenterID: invoice.CostCenterID,
        Amount:       invoice.TotalAmount,
        PeriodDate:   invoice.InvoiceDate,
    })
    if err != nil { return pipeline.StageResult{}, err }

    opCtx.SetData("budget_check.approved",          result.WithinBudget)
    opCtx.SetData("budget_check.available_amount",  result.AvailableAmount)
    opCtx.SetData("budget_check.utilisation_pct",   result.UtilisationPct)
    opCtx.SetFlag("budget_exceeded",                !result.WithinBudget)

    if !result.WithinBudget {
        switch mode {
        case "hard_block":
            return pipeline.StageResult{}, fmt.Errorf(
                "budget exceeded: %.2f available, %.2f requested",
                result.AvailableAmount.InexactFloat64(),
                invoice.TotalAmount.InexactFloat64())
        case "soft_block":
            opCtx.Suspend("budget_exceeded:override_required", "budget.check")
            return pipeline.StageResult{Status: "suspended", Message: "Budget exceeded — awaiting override"}, nil
        case "warn":
            return pipeline.StageResult{
                Status:  "completed",
                Message: fmt.Sprintf("Warning: budget %.1f%% utilised", result.UtilisationPct.InexactFloat64()),
                Outputs: map[string]any{"budget_exceeded": true},
            }, nil
        }
    }
    return pipeline.StageResult{Status: "completed"}, nil
}

// Dry-run: show expected budget check result without writing anything
func (s *BudgetCheckStage) Simulate(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    invoice := opCtx.Input.(*ap.Invoice)
    result, _ := s.budgetSvc.Check(opCtx.Ctx, budget.CheckParams{
        TenantID:     opCtx.TenantID,
        CostCenterID: invoice.CostCenterID,
        Amount:       invoice.TotalAmount,
        PeriodDate:   invoice.InvoiceDate,
    })
    return pipeline.StageResult{
        Status:  "simulated",
        Message: fmt.Sprintf("Budget check: %.1f%% utilised, %s available",
            result.UtilisationPct.InexactFloat64(),
            result.AvailableAmount.String()),
        Outputs: map[string]any{
            "budget_check.simulated":        true,
            "budget_check.within_budget":    result.WithinBudget,
            "budget_check.available_amount": result.AvailableAmount,
        },
    }, nil
}
```

---

## 5. Pipeline & PipelineBuilder

### 5.1 Pipeline Execution

```go
// awo/internal/pipeline/pipeline.go

type Pipeline struct {
    operationKey        string
    stages              []Stage
    hooksBefore         map[string][]Hook
    hooksAfter          map[string][]Hook
    compensationReg     *CompensationRegistry
    conditionEvaluator  *condition.Evaluator
    repo                OperationLogRepository
    events              EventBus
}

func (p *Pipeline) Execute(opCtx *OperationContext) (*OperationResult, error) {
    opCtx.StartedAt = time.Now()
    var completedStages []stageCheckpoint

    for _, stage := range p.stages {
        if opCtx.Suspended { break }

        stageLog := StageLog{StageName: stage.Name(), StartedAt: time.Now()}

        // ── 1. RunCondition gate (pkg/condition evaluation) ───────────────
        if cond := stage.RunCondition(); cond != "" {
            evalCtx := condition.NewEvalContext(buildEvalData(opCtx), condition.DefaultEvalOptions())
            builder  := condition.NewBuilder(condition.ConjunctionAnd)
            builder.AddFormula(cond)
            pass, err := p.conditionEvaluator.Evaluate(opCtx.Ctx, builder.Build(), evalCtx)
            if err != nil || !pass {
                stageLog.Status     = "skipped"
                stageLog.SkipReason = "run condition evaluated to false"
                opCtx.Log = append(opCtx.Log, stageLog)
                p.events.Emit(opCtx.Ctx, domain.StageSkipped{
                    OperationID: opCtx.OperationID, StageName: stage.Name(),
                })
                continue
            }
        }

        // ── 2. Before hooks ───────────────────────────────────────────────
        if err := p.runHooks(opCtx, p.hooksBefore[stage.Name()]); err != nil {
            if stage.Required() { return p.abort(opCtx, completedStages, err) }
        }
        if opCtx.Suspended { break }

        // ── 3. Execute stage (or simulate in dry-run) ─────────────────────
        var result StageResult
        var err    error

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

        if err != nil {
            // Retry retryable errors before giving up
            if isRetryable(err) {
                result, err = p.retryWithBackoff(opCtx, stage)
            }
            if err != nil {
                stageLog.Status = "failed"
                stageLog.Error  = err.Error()
                opCtx.Log = append(opCtx.Log, stageLog)
                if stage.Required() {
                    return p.abort(opCtx, completedStages, err)
                }
                continue
            }
        }

        stageLog.Status  = result.Status
        stageLog.Output  = result.Outputs
        stageLog.Message = result.Message
        opCtx.Log = append(opCtx.Log, stageLog)

        // Push onto compensation stack (only if stage ran, not simulated/skipped)
        if result.Status == "completed" {
            completedStages = append(completedStages, stageCheckpoint{
                StageName: stage.Name(), Output: result.Outputs,
            })
        }

        if opCtx.Suspended { break }

        // ── 4. After hooks ────────────────────────────────────────────────
        p.runHooks(opCtx, p.hooksAfter[stage.Name()])

        // ── 5. Broadcast progress event (WebSocket consumers) ─────────────
        p.events.Emit(opCtx.Ctx, domain.StageCompleted{
            OperationID: opCtx.OperationID,
            TenantID:    opCtx.TenantID,
            StageName:   stage.Name(),
            Status:      result.Status,
            DurationMs:  int(stageLog.Duration.Milliseconds()),
        })
    }

    // ── 6. Run DB transaction hooks (atomically, see Section 8) ──────────
    if !opCtx.DryRun && !opCtx.Suspended && len(opCtx.TxHooks) > 0 {
        if err := p.runTxHooks(opCtx); err != nil {
            // TxHook failure is non-fatal for the pipeline — log and continue
            p.events.Emit(opCtx.Ctx, domain.TxHookFailed{
                OperationID: opCtx.OperationID, Error: err.Error(),
            })
        }
    }

    now := time.Now()
    opCtx.CompletedAt = &now
    status := p.determineStatus(opCtx)
    result := p.buildResult(opCtx, status)

    // Persist asynchronously (non-blocking)
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        p.repo.SaveLog(ctx, result)
    }()

    return result, nil
}

// abort triggers LIFO compensation and marks operation as failed
func (p *Pipeline) abort(opCtx *OperationContext,
    completed []stageCheckpoint, err error) (*OperationResult, error) {

    p.compensate(opCtx, completed)
    result := p.buildResult(opCtx, "failed")
    result.Error = err.Error()

    // Persist synchronously on failure (audit trail before anything else)
    p.repo.SaveLog(opCtx.Ctx, result)
    return result, err
}
```

### 5.2 PipelineBuilder with Cache

```go
// awo/internal/pipeline/builder.go

type PipelineBuilder struct {
    stageRegistry      *StageRegistry
    hookRegistry       *HookRegistry
    compensationReg    *CompensationRegistry
    conditionEvaluator *condition.Evaluator
    logRepo            OperationLogRepository
    events             EventBus
    cache              *pipelineCache  // key: "{op_key}:{flags_sha256}"
}

func (b *PipelineBuilder) Build(operationKey string,
    session *domain.ResolvedSession) (*Pipeline, error) {

    flagsHash := b.hashRelevantFlags(operationKey, session)
    cacheKey  := operationKey + ":" + flagsHash

    if cached, ok := b.cache.Get(cacheKey); ok {
        return cached, nil  // reuse pre-built pipeline
    }

    // Build fresh pipeline
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

    // Collect hooks
    hooksBefore, hooksAfter := b.collectHooks(operationKey, session)

    pipeline := &Pipeline{
        operationKey:       operationKey,
        stages:             activeStages,
        hooksBefore:        hooksBefore,
        hooksAfter:         hooksAfter,
        compensationReg:    b.compensationReg,
        conditionEvaluator: b.conditionEvaluator,
        repo:               b.logRepo,
        events:             b.events,
    }

    b.cache.Set(cacheKey, pipeline, 5*time.Minute)
    return pipeline, nil
}

// Invalidate cache when tenant flags change
func (b *PipelineBuilder) InvalidateForTenant(tenantID uuid.UUID) {
    b.cache.DeleteByPrefix(tenantID.String())
}
```

---

## 6. Hook System

### 6.1 Hook Interface

```go
// awo/internal/pipeline/hook.go

type Hook interface {
    Name()         string
    FeatureFlag()  string
    Priority()     int
    Operations()   []string   // "*" = all operations
    StageBefore()  []string   // stage names to fire before
    StageAfter()   []string   // stage names to fire after
    RunCondition() string     // expr-lang condition (empty = always)
    Execute(opCtx *OperationContext) error
}

// Hooks with simulation support
type SimulatableHook interface {
    SimulateExecute(opCtx *OperationContext) error
}
```

### 6.2 Standard Hooks

```go
// WorkflowApprovalHook — fires before GL posting, suspends if approval needed
func (h *WorkflowApprovalHook) StageBefore()  []string { return []string{"gl.post_transaction"} }
func (h *WorkflowApprovalHook) FeatureFlag()   string   { return "workflow" }
func (h *WorkflowApprovalHook) RunCondition()  string   {
    // Only fire if amount exceeds the configured threshold
    // (uses pkg/condition — formula evaluated against opCtx.Data)
    return `toNumber(input.total_amount) > toNumber(setting["workflow.approval_threshold_amount"] ?? "0") || budget_exceeded == true`
}

// AuditLogHook — fires after every stage on every operation
func (h *AuditLogHook) StageAfter()  []string { return []string{"*"} }
func (h *AuditLogHook) Operations()  []string { return []string{"*"} }
func (h *AuditLogHook) FeatureFlag() string   { return "" } // always active

// EInvoiceHook — fires after GL posting, submits to tax authority
func (h *EInvoiceHook) StageAfter()  []string { return []string{"gl.post_transaction"} }
func (h *EInvoiceHook) FeatureFlag() string   { return "tax.e_invoice" }

// NotificationHook — fires after payment scheduling
func (h *NotificationHook) StageAfter()  []string { return []string{"banking.schedule_payment"} }
func (h *NotificationHook) FeatureFlag() string   { return "notifications" }
```

---

## 7. Stage & Hook Registries

### 7.1 Registration at Startup (wire.go)

All modules register their stages, hooks, and compensation functions at startup. This is the only place with full knowledge of all modules:

```go
// awo/cmd/server/wire.go

func InitializeApp(cfg *config.Config) (*App, error) {
    // ── Build all services ───────────────────────────────────────────────
    budgetSvc, taxSvc, workflowSvc, glSvc, ... := buildServices(cfg)

    // ── Stage registry ───────────────────────────────────────────────────
    stageRegistry := pipeline.NewStageRegistry()
    stageRegistry.Register(
        // AP
        ap.NewValidateInvoiceStage(),
        ap.NewDuplicateCheckStage(),
        ap.NewResolveVendorStage(),
        ap.NewResolveGLAccountsStage(),
        // Procurement
        procurement.NewThreeWayMatchStage(procurementSvc),
        // Budget
        budget.NewBudgetCheckStage(budgetSvc),
        // Tax
        tax.NewResolveTaxCodesStage(taxSvc),
        tax.NewCalculateTaxStage(taxSvc),
        tax.NewWithholdingTaxStage(taxSvc),
        // GL
        finance.NewGLPostingStage(glSvc),
        finance.NewAROffsetStage(arSvc),
        // DMS, Banking, Audit
        dms.NewArchiveDocumentStage(dmsSvc),
        banking.NewSchedulePaymentStage(bankingSvc),
        audit.NewAuditLogStage(auditSvc),
    )

    // ── Hook registry ────────────────────────────────────────────────────
    hookRegistry := pipeline.NewHookRegistry()
    hookRegistry.Register(
        workflow.NewWorkflowApprovalHook(workflowSvc),
        audit.NewAuditLogHook(auditSvc),
        notify.NewPostingNotificationHook(notifySvc),
        tax.NewEInvoiceHook(taxSvc),
        compliance.NewSanctionsCheckHook(complianceSvc),
    )

    // ── Compensation registry ────────────────────────────────────────────
    compensationReg := pipeline.NewCompensationRegistry()
    compensationReg.Register("gl.post_transaction", func(opCtx *pipeline.OperationContext) error {
        txID, ok := opCtx.GetData("gl_posting.transaction_id")
        if !ok { return nil }
        return glSvc.ReverseTransaction(opCtx.Ctx, txID.(uuid.UUID), "Pipeline compensation")
    })
    compensationReg.Register("banking.schedule_payment", func(opCtx *pipeline.OperationContext) error {
        paymentID, ok := opCtx.GetData("payment_scheduling.payment_id")
        if !ok { return nil }
        return bankingSvc.CancelScheduledPayment(opCtx.Ctx, paymentID.(uuid.UUID))
    })
    compensationReg.Register("dms.archive", func(opCtx *pipeline.OperationContext) error {
        docID, ok := opCtx.GetData("dms_archive.document_id")
        if !ok { return nil }
        return dmsSvc.VoidDocument(opCtx.Ctx, docID.(uuid.UUID), "Pipeline rollback")
    })

    // ── Build pipeline infrastructure ────────────────────────────────────
    pipelineBuilder := pipeline.NewPipelineBuilder(
        stageRegistry, hookRegistry, compensationReg, logRepo, eventBus)

    return &App{PipelineBuilder: pipelineBuilder, ...}, nil
}
```

---

## 8. DB Transaction Hooks — Side Workflows

### 8.1 The Problem

When the GL posting stage writes a journal entry to the database, you often want to trigger secondary processes as a guaranteed consequence of that write — not as a best-effort async fire-and-forget, but atomically within the same database transaction. If the GL entry is rolled back, the secondary process should never have started. If the secondary process fails to register, the GL entry should be rolled back.

Examples:
- Every posted GL transaction should automatically seed a `domain_events` outbox row that the Workflow Engine picks up
- A purchase order posting should trigger a `vendor_payment_schedule` creation in the same transaction
- An inventory GRN should atomically create a `wetstock_reconciliation_pending` row

This is the **Transactional Outbox Pattern** made explicit and extensible.

### 8.2 TxHook Design

```go
// awo/internal/pipeline/tx_hook.go

// TxHook is a function that runs inside the same database transaction
// as the operation's primary write. It receives the transaction-aware
// DB connection and the current OperationContext.
//
// If a TxHook returns an error, the entire transaction (including the
// primary write) is rolled back.
//
// TxHooks are registered by stages during their Execute() call, not at startup.
// They fire after all stages complete but before the transaction commits.
type TxHook struct {
    Name     string
    Priority int    // lower = executes first within same transaction
    Fn       func(ctx context.Context, tx pgx.Tx, opCtx *OperationContext) error
}

// Registration — called by a stage during Execute()
func (o *OperationContext) RegisterTxHook(hook TxHook) {
    o.TxHooks = append(o.TxHooks, hook)
}
```

### 8.3 How the GL Posting Stage Uses TxHooks

The GL posting stage executes its write AND registers a TxHook to create the domain event outbox row — both inside the same transaction:

```go
// awo/internal/finance/pipeline_stages.go

func (s *GLPostingStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    invoice := opCtx.Input.(*ap.Invoice)

    // The GL service runs its INSERT inside a DB transaction that it holds open
    // and passes back to the pipeline for TxHook execution
    txResult, err := s.glSvc.PostWithTransaction(opCtx.Ctx, finance.PostParams{
        TenantID:  opCtx.TenantID,
        Invoice:   invoice,
        PostedBy:  opCtx.UserID,
    })
    if err != nil { return pipeline.StageResult{}, err }

    opCtx.SetData("gl_posting.transaction_id", txResult.TransactionID)
    opCtx.SetData("gl_posting.tx",             txResult.Tx)  // live DB transaction

    // Register TxHook: seed domain event outbox WITHIN the same DB transaction
    opCtx.RegisterTxHook(pipeline.TxHook{
        Name:     "gl.seed_domain_event",
        Priority: 10,
        Fn: func(ctx context.Context, tx pgx.Tx, opCtx *pipeline.OperationContext) error {
            txID := opCtx.Data["gl_posting.transaction_id"].(uuid.UUID)
            // INSERT into domain_events outbox — same transaction as GL entry
            _, err := tx.Exec(ctx, `
                INSERT INTO domain_events (event_type, payload, tenant_id)
                VALUES ($1, $2, $3)
            `, "gl.transaction.posted",
               buildGLEventPayload(txID, opCtx),
               opCtx.TenantID)
            return err
        },
    })

    // Register TxHook: if budget module enabled, update budget actuals atomically
    if opCtx.FeatureEnabled("budget") {
        opCtx.RegisterTxHook(pipeline.TxHook{
            Name:     "gl.update_budget_actuals",
            Priority: 20,
            Fn: func(ctx context.Context, tx pgx.Tx, opCtx *pipeline.OperationContext) error {
                return s.budgetSvc.UpdateActualsInTx(ctx, tx, txID, invoice)
            },
        })
    }

    // Register TxHook: seed workflow trigger outbox if workflow engine is enabled
    if opCtx.FeatureEnabled("workflow") {
        opCtx.RegisterTxHook(pipeline.TxHook{
            Name:     "gl.seed_workflow_trigger",
            Priority: 30,
            Fn: func(ctx context.Context, tx pgx.Tx, opCtx *pipeline.OperationContext) error {
                // This row is picked up by the workflow engine's trigger listener
                // If GL rolls back, this row is also rolled back — no orphaned workflow
                _, err := tx.Exec(ctx, `
                    INSERT INTO workflow_trigger_queue
                        (event_name, tenant_id, payload, created_at)
                    VALUES
                        ('gl.transaction.posted', $1, $2, NOW())
                `, opCtx.TenantID, buildWorkflowPayload(opCtx))
                return err
            },
        })
    }

    return pipeline.StageResult{
        Status: "completed",
        Outputs: map[string]any{
            "gl_posting.transaction_id": txResult.TransactionID,
        },
    }, nil
}
```

### 8.4 Pipeline Runs TxHooks Within the Open Transaction

After all stages complete, the pipeline flushes TxHooks in priority order, then commits:

```go
// awo/internal/pipeline/pipeline.go

func (p *Pipeline) runTxHooks(opCtx *OperationContext) error {
    // Get the open transaction from the stage that created it
    tx, ok := opCtx.Data["gl_posting.tx"].(pgx.Tx)
    if !ok { return nil }   // no open transaction to hook into

    // Sort by priority
    hooks := opCtx.TxHooks
    sort.Slice(hooks, func(i, j int) bool {
        return hooks[i].Priority < hooks[j].Priority
    })

    // Execute each hook within the same transaction
    for _, hook := range hooks {
        if err := hook.Fn(opCtx.Ctx, tx, opCtx); err != nil {
            // Rollback the entire transaction if any hook fails
            tx.Rollback(opCtx.Ctx)
            return fmt.Errorf("TxHook %s failed: %w — transaction rolled back", hook.Name, err)
        }
    }

    // All hooks succeeded — commit
    return tx.Commit(opCtx.Ctx)
}
```

### 8.5 What This Enables

```
Without TxHooks (naive approach):
  1. GL INSERT committed ✓
  2. domain_events INSERT fails ✗
     → GL is posted but no event fired → Workflow Engine never notified
     → Inconsistent state

With TxHooks (atomic approach):
  1. GL INSERT (transaction open, not yet committed)
  2. TxHook: domain_events INSERT (same transaction)
  3. TxHook: workflow_trigger_queue INSERT (same transaction)
  4. All succeed → COMMIT → everything visible atomically
     OR
  4. Any fails → ROLLBACK → nothing is visible — clean state
```

**Workflow trigger outbox table:**

```sql
CREATE TABLE workflow_trigger_queue (
  id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id   uuid        NOT NULL REFERENCES tenants(id),
  event_name  text        NOT NULL,    -- matches workflow_triggers.event_name
  payload     jsonb       NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  processed_at timestamptz,            -- NULL = not yet processed
  error_count int         NOT NULL DEFAULT 0
);

CREATE INDEX idx_workflow_trigger_queue_pending
  ON workflow_trigger_queue(tenant_id, created_at)
  WHERE processed_at IS NULL;
```

The Workflow Engine runs a lightweight poller that picks up rows from `workflow_trigger_queue`, matches them to active `workflow_triggers`, starts the appropriate Temporal workflows, and marks them processed.

---

## 9. Failure Recovery & Compensation

### 9.1 Failure Categories

| Category | Example | Recovery |
|---|---|---|
| ValidationFailure | Duplicate invoice detected | Abort immediately — nothing was written |
| BusinessRuleFailure | Budget hard block | Abort — GL not touched, safe |
| ServiceFailure | Tax service HTTP 503 | Retry with exponential backoff |
| IntegrationFailure | GL posting DB conflict | Retry; if exhausted → compensate prior stages |
| InfrastructureFailure | DB connection lost mid-pipeline | Resume from checkpoint after recovery |

### 9.2 Compensation Algorithm — LIFO

```go
func (p *Pipeline) compensate(opCtx *OperationContext, completed []stageCheckpoint) {
    // LIFO: last completed stage is compensated first
    for i := len(completed) - 1; i >= 0; i-- {
        checkpoint := completed[i]
        fn, exists := p.compensationReg.Get(checkpoint.StageName)
        if !exists { continue }  // no compensation = no side effects

        if err := fn(opCtx); err != nil {
            // Compensation failure is logged but does NOT stop other compensations
            p.events.Emit(opCtx.Ctx, domain.CompensationFailed{
                OperationID: opCtx.OperationID,
                StageName:   checkpoint.StageName,
                Error:       err.Error(),
            })
            // Human intervention required for this stage
            // Other stages continue to be compensated
        }
    }
    opCtx.SetFlag("compensated", true)
}
```

**What each stage's compensation does:**

| Stage | Compensation |
|---|---|
| `gl.post_transaction` | Post reversal transaction (mirror of original) |
| `banking.schedule_payment` | Cancel scheduled payment before execution date |
| `dms.archive` | Mark document as voided (cannot delete — audit requirement) |
| `budget.check` (if reservation made) | Release budget reservation |
| `ar.apply_advance` | Reverse advance payment application |
| Read-only stages (validate, tax calc, duplicate check) | No compensation registered — no side effects |

### 9.3 Retry with Backoff

```go
func (p *Pipeline) retryWithBackoff(opCtx *OperationContext, stage Stage) (StageResult, error) {
    maxRetries := 3
    baseDelay  := 500 * time.Millisecond
    for attempt := 1; attempt <= maxRetries; attempt++ {
        // Exponential backoff with jitter
        delay  := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
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

```go
func (p *Pipeline) ExecuteIdempotent(opCtx *OperationContext,
    idempotencyKey string) (*OperationResult, error) {

    existing, err := p.repo.GetByIdempotencyKey(opCtx.Ctx, opCtx.TenantID, idempotencyKey)
    if err == nil {
        switch existing.Status {
        case "completed":      return existing.Result, nil
        case "running":        return nil, ErrOperationInProgress
        case "pending_approval": return existing.Result, ErrPendingApproval
        case "failed":         return existing.Result, ErrOperationFailed
        }
    }
    p.repo.SetIdempotencyKey(opCtx.Ctx, opCtx.OperationID, idempotencyKey)
    return p.Execute(opCtx)
}
```

### 9.5 Resume After Suspension

```go
// awo/internal/pipeline/resume.go

func (s *PipelineResumeService) Resume(ctx context.Context,
    operationID uuid.UUID, resumeData map[string]any) (*OperationResult, error) {

    log, _ := s.repo.GetByOperationID(ctx, operationID)
    if log.Status != "pending_approval" {
        return nil, fmt.Errorf("cannot resume operation with status: %s", log.Status)
    }

    // Reconstruct OperationContext from persisted state
    opCtx := &OperationContext{
        Ctx:          ctx,
        Session:      s.sessionRepo.Reconstruct(ctx, log.SessionSnapshot),
        OperationID:  operationID,
        OperationKey: log.OperationKey,
        Input:        log.InputSnapshot,
        Data:         log.OutputData,      // accumulated stage outputs
        Flags:        log.FlagsSnapshot,
        Log:          log.StageLog,
    }

    // Inject approval decision into context
    for k, v := range resumeData {
        opCtx.SetData("resume."+k, v)
    }
    opCtx.SetFlag("workflow_approved", true)
    opCtx.Suspended = false

    // Rebuild pipeline and skip already-completed stages
    p, _ := s.pipelineBuilder.Build(log.OperationKey, opCtx.Session)
    completedNames := extractCompletedStageNames(opCtx.Log)
    p.SkipStages(completedNames)
    p.StartFrom(log.ResumePoint)

    return p.Execute(opCtx)
}
```

---

## 10. Performance Optimisation

### 10.1 Pipeline Cache (5-Minute TTL by Flags Hash)

```go
type pipelineCache struct {
    mu    sync.RWMutex
    items map[string]*cachedEntry
}

type cachedEntry struct {
    pipeline *Pipeline
    cachedAt time.Time
    ttl      time.Duration
}

func (c *pipelineCache) Get(key string) (*Pipeline, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    e, ok := c.items[key]
    if !ok || time.Since(e.cachedAt) > e.ttl { return nil, false }
    return e.pipeline, true
}
```

Cache invalidated when `FlagService.Set()` is called — the next pipeline build is fresh, then cached again.

### 10.2 Parallel Execution Within Priority Bands

Independent stages (no shared `DependsOn()` entries) in the same priority band run concurrently. Tax calculation and budget check are both in band 300-499 and have independent inputs — they run in parallel, cutting latency by ~40% on full enterprise tenants.

### 10.3 Operation Log Writes

Stage logs are accumulated in-memory and written in a single UPDATE at pipeline completion. Compensation logs are written synchronously immediately (audit requirement). All other log writes are non-blocking goroutines with a 5-second timeout.

---

## 11. API Handler Design

### 11.1 Sync/Async Split

```go
const pipelineSyncDeadline = 8 * time.Second

func ProcessInvoiceHandler(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        idempotencyKey := c.Get("Idempotency-Key")  // optional
        // ... build opCtx ...

        resultCh := make(chan *pipeline.OperationResult, 1)
        errCh    := make(chan error, 1)
        go func() {
            result, err := deps.APService.ProcessInvoice(c.Context(), req, idempotencyKey)
            if err != nil { errCh <- err; return }
            resultCh <- result
        }()

        select {
        case result := <-resultCh:
            return renderPipelineResult(c, result)
        case err := <-errCh:
            return renderPipelineError(c, err)
        case <-time.After(pipelineSyncDeadline):
            return c.Status(202).JSON(fiber.Map{
                "status":            "running",
                "operation_id":      opCtx.OperationID,
                "poll_url":          "/api/operations/" + opCtx.OperationID.String() + "/status",
                "poll_interval_ms":  2000,
            })
        }
    }
}
```

### 11.2 Status Polling and WebSocket Live

```go
// GET /api/operations/:id/status — polling endpoint
// GET /api/operations/:id/live   — WebSocket for real-time stage events

// Standard error response schema
{
  "status":        "failed",
  "operation_id":  "a1b2c3d4-...",
  "error": {
    "code":                  "BUDGET_EXCEEDED",
    "message":               "Invoice amount KES 750,000 exceeds budget KES 500,000",
    "stage":                 "budget.check",
    "recoverable":           false,
    "user_action_required":  "Request a budget override from your Finance Controller"
  },
  "compensated":     true,
  "compensation_log": [...],
  "stage_log":       [...]
}
```

---

## 12. Dry-Run Mode

### 12.1 Behaviour

When `opCtx.DryRun = true`, stages with side effects call `Simulate()` instead of `Execute()`. Read-only stages (validation, tax calculation, duplicate check) run normally — they have no side effects and their real output is useful for the preview.

```go
// Dry-run API endpoint
// POST /api/operations/dry-run
{
  "operation_key": "ap.invoice.process",
  "input": { "invoice_id": "...", ... }
}

// Response
{
  "dry_run":                  true,
  "would_succeed":            true,
  "would_require_approval":   true,
  "expected_approvers":       ["Finance Controller"],
  "stages_that_would_run":    ["ap.validate_invoice", "budget.check", "tax.calculate", "gl.post_transaction"],
  "stages_that_would_skip":   ["procurement.three_way_match"],
  "expected_gl_entries":      [...],
  "expected_payment_date":    "2025-02-28",
  "expected_tax_amount":      "12480.00",
  "stage_log":                [...]
}
```

### 12.2 Workflow Hook in Dry-Run

The `WorkflowApprovalHook` in dry-run mode checks if approval would be required and writes that to context — but does NOT call `opCtx.Suspend()`. The dry-run continues through the full pipeline to show all expected outputs:

```go
func (h *WorkflowApprovalHook) SimulateExecute(opCtx *pipeline.OperationContext) error {
    config, _ := h.workflowSvc.GetRuleFor(opCtx.Ctx, ...)
    opCtx.SetData("workflow.approval_required_in_real_run", config.RequiresApproval)
    opCtx.SetData("workflow.expected_approvers",            config.Approvers)
    // Note: does NOT call opCtx.Suspend() in dry-run mode
    return nil
}
```

---

## 13. The pkg/condition Package

### 13.1 Role in the Pipeline

The `pkg/condition` package provides safe, sandboxed boolean expression evaluation. It is used in four places:

| Use | Who writes the expression | Evaluated by |
|---|---|---|
| Stage `RunCondition()` | Go developer (module stage) | Pipeline before each stage |
| Hook `RunCondition()` | Go developer (hook author) | Pipeline before each hook |
| `ScriptStage` formula | Tenant user (in UI) | ScriptStage.Execute() |
| Workflow condition step | Tenant user (visual builder) | Workflow Engine condition executor |

### 13.2 Custom Function Registration

The `pkg/condition` package allows registering named functions that can be called from both Go-authored conditions AND tenant-authored UI expressions. This is the core extension mechanism:

```go
// awo/internal/pipeline/condition_setup.go

// BuildPipelineConditionEvaluator creates the shared evaluator used by the pipeline
// and pre-registers all domain functions available to tenant expressions.
func BuildPipelineConditionEvaluator(deps *ServiceDeps) *condition.Evaluator {
    opts := condition.EvalOptions{
        MaxDepth:      10,
        MaxConditions: 100,
        Timeout:       500 * time.Millisecond,
        RegexTimeout:  100 * time.Millisecond,
        CacheResults:  true,
    }
    eval := condition.NewEvaluator(nil, opts)

    // ── Financial functions ───────────────────────────────────────────────
    // These are callable from tenant expressions in the UI

    // documentTotal(document_type, document_id) → decimal
    // Returns the grand total of any document the user has access to.
    // UI users call this as: documentTotal("invoice", input.invoice_id) > 50000
    eval.RegisterFunction("documentTotal",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            if len(args) < 2 { return nil, fmt.Errorf("documentTotal requires (doc_type, doc_id)") }
            docType, _ := args[0].(string)
            docID, _    := args[1].(string)
            // Enforce: user must have read permission on this document type
            session := sessionFromEvalCtx(ec)
            if !session.Can(docType+"."+docType+"s", "read") {
                return nil, fmt.Errorf("access denied: cannot read %s", docType)
            }
            return deps.DocumentSvc.GetTotal(ctx, docType, uuid.MustParse(docID))
        })

    // lineCount(document_type, document_id) → int
    eval.RegisterFunction("lineCount",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            docType, _ := args[0].(string)
            docID, _   := args[1].(string)
            return deps.DocumentSvc.GetLineCount(ctx, docType, uuid.MustParse(docID))
        })

    // fieldValue(document_type, document_id, field_name) → any
    // Returns the value of a specific field from a document.
    // UI users call this as: fieldValue("purchase_order", input.po_id, "vendor_name") == "Acme"
    eval.RegisterFunction("fieldValue",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            if len(args) < 3 { return nil, fmt.Errorf("fieldValue requires (doc_type, doc_id, field)") }
            docType,   _ := args[0].(string)
            docID,     _ := args[1].(string)
            fieldName, _ := args[2].(string)
            session := sessionFromEvalCtx(ec)
            // Verify field is in the gettable fields for this document type
            if !deps.SchemaReg.IsGettableField(docType, fieldName, session.Permissions) {
                return nil, fmt.Errorf("field %s.%s is not accessible", docType, fieldName)
            }
            return deps.DocumentSvc.GetField(ctx, docType, uuid.MustParse(docID), fieldName)
        })

    // ── Entity functions ──────────────────────────────────────────────────

    // isChildOf(entity_id, parent_entity_id) → bool
    eval.RegisterFunction("isChildOf",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            childID,  _ := args[0].(string)
            parentID, _ := args[1].(string)
            return deps.EntitySvc.IsDescendantOf(ctx, uuid.MustParse(childID), uuid.MustParse(parentID))
        })

    // ── Date functions ────────────────────────────────────────────────────

    // isWithinPeriod(date_str, period_start, period_end) → bool
    eval.RegisterFunction("isWithinPeriod",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            dateStr, _  := args[0].(string)
            startStr, _ := args[1].(string)
            endStr, _   := args[2].(string)
            date,  _    := time.Parse("2006-01-02", dateStr)
            start, _    := time.Parse("2006-01-02", startStr)
            end, _      := time.Parse("2006-01-02", endStr)
            return !date.Before(start) && !date.After(end), nil
        })

    // ── Vendor / Contact functions ────────────────────────────────────────

    // hasTag(entity_type, entity_id, tag) → bool
    eval.RegisterFunction("hasTag",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            entityType, _ := args[0].(string)
            entityID, _   := args[1].(string)
            tag, _        := args[2].(string)
            return deps.ContactSvc.HasTag(ctx, entityType, uuid.MustParse(entityID), tag)
        })

    // isApprovedVendor(vendor_id) → bool
    eval.RegisterFunction("isApprovedVendor",
        func(ctx context.Context, args []any, ec *condition.EvalContext) (any, error) {
            vendorID, _ := args[0].(string)
            return deps.VendorSvc.IsApproved(ctx, uuid.MustParse(vendorID))
        })

    return eval
}
```

### 13.3 Security Limits for Tenant Scripts

```go
// Tighter limits for user-authored scripts vs. Go-authored stage conditions
var tenantScriptOptions = condition.EvalOptions{
    MaxDepth:      3,
    MaxConditions: 20,
    Timeout:       200 * time.Millisecond,
    RegexTimeout:  50 * time.Millisecond,
    CacheResults:  true,
}
```

**Access control in registered functions:** Every function that reads data checks `session.Can()` before executing. The session is threaded through `EvalContext` via a context key. A tenant user cannot call `documentTotal("payroll", ...)` if they do not have `payroll.payrolls.read` permission.

**Formula cache note:** Because each `ScriptStage` gets its own `Evaluator` instance handling exactly one formula, the unbounded cache growth issue from the condition docs is neutralised. Each evaluator caches exactly one compiled formula.

---

## 14. Tenant Scripting Engine

### 14.1 The ScriptStage

A `ScriptStage` is a pipeline stage whose entire logic is a tenant-authored formula. It is stored in the database and loaded at runtime:

```go
// awo/internal/pipeline/stages/script_stage.go

type ScriptStage struct {
    pipeline.BaseStage
    definition ScriptStageDefinition
    evaluator  *condition.Evaluator
}

type ScriptStageDefinition struct {
    ID          string
    TenantID    uuid.UUID
    Name        string
    Formula     string         // expr-lang formula
    OnTrue      string         // next stage name when formula = true
    OnFalse     string         // next stage name when formula = false
    FeatureFlag string
    Priority    int
    Operations  []string
}

func NewScriptStage(def ScriptStageDefinition,
    eval *condition.Evaluator) *ScriptStage {
    return &ScriptStage{
        BaseStage: pipeline.BaseStage{
            name:        "script." + def.ID,
            operations:  def.Operations,
            featureFlag: def.FeatureFlag,
            priority:    def.Priority,
            required:    false,  // user scripts are never required — safety policy
        },
        definition: def,
        evaluator:  eval,
    }
}

func (s *ScriptStage) Execute(opCtx *pipeline.OperationContext) (pipeline.StageResult, error) {
    evalCtx := condition.NewEvalContext(buildEvalData(opCtx), tenantScriptOptions)

    builder := condition.NewBuilder(condition.ConjunctionAnd)
    builder.AddFormula(s.definition.Formula)
    group := builder.Build()

    result, err := s.evaluator.Evaluate(opCtx.Ctx, group, evalCtx)
    if err != nil {
        // Never crash the pipeline on a script error — log and skip
        return pipeline.StageResult{
            Status:  "skipped",
            Message: "Script error: " + s.classifyScriptError(err),
        }, nil
    }

    metrics := evalCtx.GetMetrics()
    opCtx.SetData("script."+s.definition.ID+".result",          result)
    opCtx.SetData("script."+s.definition.ID+".rules_evaluated", metrics.RulesEvaluated)

    nextStage := s.definition.OnFalse
    if result { nextStage = s.definition.OnTrue }

    return pipeline.StageResult{
        Status:      "completed",
        NextStageID: nextStage,
        Outputs: map[string]any{
            "result":          result,
            "rules_evaluated": metrics.RulesEvaluated,
            "duration_ms":     metrics.Duration.Milliseconds(),
        },
    }, nil
}

func (s *ScriptStage) classifyScriptError(err error) string {
    switch err {
    case condition.ErrEvaluationTimeout:   return "script timed out (200ms limit)"
    case condition.ErrMaxDepthExceeded:    return "script nesting too deep (max 3 levels)"
    case condition.ErrResourceLimitExceeded: return "script too complex (max 20 conditions)"
    case condition.ErrFieldNotFound:        return "referenced field not found in context"
    case condition.ErrFunctionNotFound:     return "called function is not registered"
    default:                               return err.Error()
    }
}
```

### 14.2 What Tenant Expressions Can Access

The `buildEvalData(opCtx)` function constructs the data map available to tenant expressions:

```go
func buildEvalData(opCtx *pipeline.OperationContext) map[string]any {
    data := make(map[string]any)

    // Stage outputs (with {stage}.{key} prefix)
    for k, v := range opCtx.Data { data[k] = v }

    // Boolean flags set by stages
    for k, v := range opCtx.Flags { data[k] = v }

    // Feature flags (read-only view of own tenant's flags)
    data["feature"] = opCtx.Session.Configuration.Flags

    // Tenant settings (own tenant only)
    data["setting"] = opCtx.Session.Configuration.Settings

    // Original input (invoice, PO, journal, etc.)
    if input, ok := opCtx.Input.(map[string]any); ok {
        data["input"] = input
    }

    // Entity context
    data["entity_id"] = opCtx.EntityID.String()
    data["tenant_id"] = opCtx.TenantID.String()

    // NOT exposed: session.Permissions, other tenants' data, raw DB access
    return data
}
```

**Example tenant expressions:**

```javascript
// Route based on invoice total from the document (calls registered function)
documentTotal("invoice", input.invoice_id) > 100000

// Check if vendor is approved and within budget
isApprovedVendor(input.vendor_id) && budget_check.available_amount > input.total_amount

// Check document has required line items
lineCount("invoice", input.invoice_id) > 0 && fieldValue("invoice", input.invoice_id, "currency") == "KES"

// Date-based routing
isWithinPeriod(input.invoice_date, "2025-01-01", "2025-12-31")

// Entity hierarchy check
isChildOf(input.entity_id, setting["finance.restricted_entities_parent"])

// Combined condition
feature["tax"] && input.total_amount > toNumber(setting["tax.minimum_taxable_amount"])
```

---

## 15. Autocomplete for User-Defined Expressions

### 15.1 The Problem

When a tenant user is writing a workflow condition or script formula in the UI, they need to know:
- Which **resources** (document types) they can reference
- Which **fields** on those resources are **gettable** for decision-making
- Which **functions** are available to call
- What the **current OperationContext data keys** are (from prior stages)

This information must respect the user's **permissions** — they should only see resources and fields they have `read` access to.

### 15.2 The Schema Registry

A `SchemaRegistry` is the authoritative catalogue of all document types, their gettable fields, and the permissions required to access each:

```go
// awo/internal/pipeline/schema_registry.go

// SchemaRegistry is the authoritative catalogue of all document types
// and their gettable fields for use in tenant expressions.
// Registered at startup by each module.
type SchemaRegistry struct {
    mu       sync.RWMutex
    schemas  map[string]*DocumentSchema  // key: "invoice", "purchase_order", etc.
}

type DocumentSchema struct {
    // The document type key used in functions: "invoice", "purchase_order", etc.
    TypeKey      string
    Label        string       // "AP Invoice"
    Module       string       // "ap"

    // Permission required to access this document type in expressions
    ReadPermission string     // e.g. "ap.invoices.read"

    // Gettable fields — only these can be retrieved via fieldValue()
    Fields []GettableField

    // Relationships that can be queried (document lines, related documents)
    Relationships []DocumentRelationship
}

type GettableField struct {
    Name        string       // "total_amount", "vendor_name", "currency", "status"
    Label       string       // "Total Amount", "Vendor Name", "Currency", "Status"
    Type        string       // "decimal", "string", "date", "enum", "uuid"
    Description string
    EnumValues  []string     // for type == "enum"

    // Which permission is needed to see this field in autocomplete
    // Usually same as parent ReadPermission, but sensitive fields may need more
    RequiredPermission string
}

type DocumentRelationship struct {
    Name        string       // "lines", "tax_lines", "attachments"
    Label       string       // "Invoice Lines"
    TargetType  string       // "invoice_line"
    Cardinality string       // "one" | "many"
    Fields      []GettableField  // fields available on the related document
}
```

### 15.3 Module Schema Registration at Startup

Each module registers its document schemas in wire.go:

```go
// wire.go

schemaRegistry := pipeline.NewSchemaRegistry()

// AP module registers invoice schema
schemaRegistry.Register(&pipeline.DocumentSchema{
    TypeKey:        "invoice",
    Label:          "AP Invoice",
    Module:         "ap",
    ReadPermission: "ap.invoices.read",
    Fields: []pipeline.GettableField{
        {Name: "total_amount",    Label: "Total Amount",    Type: "decimal",
         Description: "Invoice grand total including tax"},
        {Name: "subtotal",        Label: "Subtotal",        Type: "decimal",
         Description: "Net amount before tax"},
        {Name: "tax_amount",      Label: "Tax Amount",      Type: "decimal"},
        {Name: "currency",        Label: "Currency",        Type: "enum",
         EnumValues: []string{"KES", "USD", "EUR", "GBP"}},
        {Name: "status",          Label: "Status",          Type: "enum",
         EnumValues: []string{"draft", "submitted", "approved", "posted", "paid"}},
        {Name: "vendor_name",     Label: "Vendor Name",     Type: "string"},
        {Name: "vendor_id",       Label: "Vendor ID",       Type: "uuid"},
        {Name: "invoice_date",    Label: "Invoice Date",    Type: "date"},
        {Name: "due_date",        Label: "Due Date",        Type: "date"},
        {Name: "entity_id",       Label: "Entity",          Type: "uuid"},
        {Name: "po_number",       Label: "PO Number",       Type: "string",
         Description: "Purchase order reference"},
    },
    Relationships: []pipeline.DocumentRelationship{
        {
            Name:        "lines",
            Label:       "Invoice Lines",
            TargetType:  "invoice_line",
            Cardinality: "many",
            Fields: []pipeline.GettableField{
                {Name: "amount",      Label: "Line Amount",    Type: "decimal"},
                {Name: "quantity",    Label: "Quantity",       Type: "decimal"},
                {Name: "unit_price",  Label: "Unit Price",     Type: "decimal"},
                {Name: "description", Label: "Description",    Type: "string"},
                {Name: "item_code",   Label: "Item Code",      Type: "string"},
                {Name: "gl_account",  Label: "GL Account",     Type: "string"},
                {Name: "cost_center", Label: "Cost Centre",    Type: "string"},
                {Name: "tax_rate",    Label: "Tax Rate",       Type: "decimal"},
            },
        },
    },
})

// Procurement module registers purchase order schema
schemaRegistry.Register(&pipeline.DocumentSchema{
    TypeKey:        "purchase_order",
    Label:          "Purchase Order",
    Module:         "procurement",
    ReadPermission: "procurement.purchase_orders.read",
    Fields: []pipeline.GettableField{
        {Name: "total_amount",    Label: "Total Amount",   Type: "decimal"},
        {Name: "approved_amount", Label: "Approved Amount",Type: "decimal"},
        {Name: "status",          Label: "Status",         Type: "enum",
         EnumValues: []string{"draft", "approved", "sent", "partially_received", "received", "closed"}},
        {Name: "vendor_id",       Label: "Vendor",         Type: "uuid"},
        {Name: "order_date",      Label: "Order Date",     Type: "date"},
        {Name: "expected_date",   Label: "Expected Delivery", Type: "date"},
    },
    Relationships: []pipeline.DocumentRelationship{
        {
            Name:       "lines",
            Label:      "PO Lines",
            TargetType: "po_line",
            Cardinality: "many",
            Fields: []pipeline.GettableField{
                {Name: "quantity_ordered",  Label: "Qty Ordered",   Type: "decimal"},
                {Name: "quantity_received", Label: "Qty Received",  Type: "decimal"},
                {Name: "unit_price",        Label: "Unit Price",    Type: "decimal"},
                {Name: "item_code",         Label: "Item Code",     Type: "string"},
            },
        },
    },
})

// Finance module registers GL journal schema
schemaRegistry.Register(&pipeline.DocumentSchema{
    TypeKey:        "gl_journal",
    Label:          "GL Journal",
    Module:         "finance",
    ReadPermission: "finance.journals.read",
    Fields: []pipeline.GettableField{
        {Name: "total_debit",    Label: "Total Debit",   Type: "decimal"},
        {Name: "total_credit",   Label: "Total Credit",  Type: "decimal"},
        {Name: "journal_date",   Label: "Journal Date",  Type: "date"},
        {Name: "period",         Label: "Accounting Period", Type: "string"},
        {Name: "status",         Label: "Status",        Type: "enum",
         EnumValues: []string{"draft", "pending_approval", "approved", "posted"}},
        {Name: "entity_id",      Label: "Entity",        Type: "uuid"},
    },
    Relationships: []pipeline.DocumentRelationship{
        {
            Name:       "entries",
            Label:      "Journal Entries",
            TargetType: "journal_entry",
            Cardinality: "many",
            Fields: []pipeline.GettableField{
                {Name: "account_code",  Label: "Account Code",  Type: "string"},
                {Name: "debit_amount",  Label: "Debit Amount",  Type: "decimal"},
                {Name: "credit_amount", Label: "Credit Amount", Type: "decimal"},
                {Name: "description",   Label: "Description",   Type: "string"},
                {Name: "cost_center",   Label: "Cost Centre",   Type: "string"},
            },
        },
    },
})
```

### 15.4 The Autocomplete API

The UI calls this endpoint when the user opens the script editor:

```go
// GET /api/pipeline/autocomplete
// Returns all resources, fields, functions, and context keys the user can access.
// Filtered by the user's current permissions from their session.

func AutocompleteHandler(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        session := middleware.ContextSession(c)
        operationKey := c.Query("operation_key")  // e.g. "ap.invoice.process"

        response := buildAutocompleteResponse(deps, session, operationKey)
        return c.JSON(response)
    }
}

type AutocompleteResponse struct {
    // Document types and their fields (filtered by permissions)
    Resources []ResourceAutocomplete `json:"resources"`
    // Registered functions callable from expressions
    Functions []FunctionAutocomplete `json:"functions"`
    // Current OperationContext data keys (from prior stages)
    ContextKeys []ContextKeyAutocomplete `json:"context_keys"`
    // Feature flags available (boolean variables)
    FeatureKeys []string `json:"feature_keys"`
    // Tenant settings keys (for use in expressions)
    SettingKeys []string `json:"setting_keys"`
    // Operators reference
    Operators []OperatorDef `json:"operators"`
}

type ResourceAutocomplete struct {
    TypeKey      string              `json:"type_key"`      // "invoice"
    Label        string              `json:"label"`         // "AP Invoice"
    Description  string              `json:"description"`
    Fields       []FieldAutocomplete `json:"fields"`
    Relationships []RelationshipAutocomplete `json:"relationships"`
}

type FieldAutocomplete struct {
    Name        string   `json:"name"`         // "total_amount"
    Label       string   `json:"label"`        // "Total Amount"
    Type        string   `json:"type"`         // "decimal"
    Description string   `json:"description"`
    EnumValues  []string `json:"enum_values"`  // for enum fields
    ExampleExpr string   `json:"example_expr"` // "fieldValue(\"invoice\", input.invoice_id, \"total_amount\")"
}

type RelationshipAutocomplete struct {
    Name        string              `json:"name"`
    Label       string              `json:"label"`
    TargetType  string              `json:"target_type"`
    Cardinality string              `json:"cardinality"`
    Fields      []FieldAutocomplete `json:"fields"`
    ExampleExpr string              `json:"example_expr"` // "lineCount(\"invoice\", input.invoice_id)"
}

type FunctionAutocomplete struct {
    Name        string            `json:"name"`
    Label       string            `json:"label"`
    Description string            `json:"description"`
    Signature   string            `json:"signature"`  // "documentTotal(doc_type, doc_id)"
    ReturnType  string            `json:"return_type"`
    Parameters  []ParamAutocomplete `json:"parameters"`
    Examples    []string          `json:"examples"`
}

type ContextKeyAutocomplete struct {
    Key         string `json:"key"`          // "budget_check.available_amount"
    Label       string `json:"label"`        // "Budget: Available Amount"
    Type        string `json:"type"`         // "decimal"
    StageSource string `json:"stage_source"` // "budget.check"
    Description string `json:"description"`
}

func buildAutocompleteResponse(deps *app.Deps,
    session *domain.ResolvedSession,
    operationKey string) AutocompleteResponse {

    resp := AutocompleteResponse{}

    // ── Resources: filter by user permissions ────────────────────────────
    for _, schema := range deps.SchemaRegistry.All() {
        if !session.Can(schema.ReadPermission[:lastDot(schema.ReadPermission)],
                        "read") {
            continue  // user cannot read this resource type
        }

        resource := ResourceAutocomplete{
            TypeKey: schema.TypeKey,
            Label:   schema.Label,
        }

        for _, field := range schema.Fields {
            // Check field-level permission if more restrictive than resource
            reqPerm := field.RequiredPermission
            if reqPerm == "" { reqPerm = schema.ReadPermission }
            if !session.Can(reqPerm[:lastDot(reqPerm)], "read") { continue }

            resource.Fields = append(resource.Fields, FieldAutocomplete{
                Name:        field.Name,
                Label:       field.Label,
                Type:        field.Type,
                Description: field.Description,
                EnumValues:  field.EnumValues,
                ExampleExpr: fmt.Sprintf(`fieldValue("%s", input.%s_id, "%s")`,
                    schema.TypeKey, schema.TypeKey, field.Name),
            })
        }

        for _, rel := range schema.Relationships {
            ra := RelationshipAutocomplete{
                Name:        rel.Name,
                Label:       rel.Label,
                TargetType:  rel.TargetType,
                Cardinality: rel.Cardinality,
            }
            for _, f := range rel.Fields {
                ra.Fields = append(ra.Fields, FieldAutocomplete{
                    Name: f.Name, Label: f.Label, Type: f.Type,
                })
            }
            if rel.Cardinality == "many" {
                ra.ExampleExpr = fmt.Sprintf(`lineCount("%s", input.%s_id)`,
                    schema.TypeKey, schema.TypeKey)
            }
            resource.Relationships = append(resource.Relationships, ra)
        }

        resp.Resources = append(resp.Resources, resource)
    }

    // ── Functions ─────────────────────────────────────────────────────────
    resp.Functions = []FunctionAutocomplete{
        {
            Name: "documentTotal", Label: "Document Grand Total",
            Description: "Returns the grand total of a document",
            Signature:   "documentTotal(doc_type, doc_id)",
            ReturnType:  "decimal",
            Parameters: []ParamAutocomplete{
                {Name: "doc_type", Type: "string", Description: "e.g. \"invoice\", \"purchase_order\""},
                {Name: "doc_id",   Type: "uuid",   Description: "Document UUID"},
            },
            Examples: []string{
                `documentTotal("invoice", input.invoice_id) > 100000`,
                `documentTotal("purchase_order", input.po_id) > 50000`,
            },
        },
        {
            Name: "lineCount", Label: "Document Line Count",
            Description: "Returns the number of lines in a document",
            Signature:   "lineCount(doc_type, doc_id)",
            ReturnType:  "int",
            Examples:    []string{`lineCount("invoice", input.invoice_id) > 0`},
        },
        {
            Name: "fieldValue", Label: "Get Field Value",
            Description: "Returns the value of a specific field from a document",
            Signature:   "fieldValue(doc_type, doc_id, field_name)",
            ReturnType:  "any",
            Examples: []string{
                `fieldValue("invoice", input.invoice_id, "currency") == "KES"`,
                `fieldValue("invoice", input.invoice_id, "status") == "approved"`,
            },
        },
        {
            Name: "isApprovedVendor", Label: "Is Approved Vendor",
            Signature:  "isApprovedVendor(vendor_id)",
            ReturnType: "bool",
            Examples:   []string{`isApprovedVendor(input.vendor_id)`},
        },
        {
            Name: "isChildOf", Label: "Is Child Entity",
            Description: "Returns true if entity_id is a descendant of parent_id",
            Signature:   "isChildOf(entity_id, parent_entity_id)",
            ReturnType:  "bool",
            Examples:    []string{`isChildOf(input.entity_id, "00000000-0000-0000-0000-000000000001")`},
        },
        {
            Name: "isWithinPeriod", Label: "Is Within Period",
            Signature:  "isWithinPeriod(date, period_start, period_end)",
            ReturnType: "bool",
            Examples:   []string{`isWithinPeriod(input.invoice_date, "2025-01-01", "2025-12-31")`},
        },
        {
            Name: "hasTag", Label: "Has Tag",
            Signature:  "hasTag(entity_type, entity_id, tag_name)",
            ReturnType: "bool",
            Examples:   []string{`hasTag("vendor", input.vendor_id, "preferred")`},
        },
    }

    // ── Context keys from prior pipeline stages ───────────────────────────
    // For the given operation_key, what keys will be in opCtx.Data by this stage?
    resp.ContextKeys = deps.SchemaRegistry.GetContextKeysForOperation(operationKey)

    // ── Feature and setting keys ──────────────────────────────────────────
    for k := range session.Configuration.Flags    { resp.FeatureKeys = append(resp.FeatureKeys, k) }
    for k := range session.Configuration.Settings { resp.SettingKeys = append(resp.SettingKeys, k) }

    // ── Operators reference ───────────────────────────────────────────────
    resp.Operators = standardOperators()

    return resp
}
```

### 15.5 UI Script Editor with Autocomplete

The script editor is an amis `editor` component with the Monaco editor. The autocomplete data from the API is used to configure Monaco's completion provider:

```json
{
  "type": "form",
  "title": "Script Condition Editor",
  "body": [
    {
      "type": "service",
      "api": "/api/pipeline/autocomplete?operation_key=${operation_key}",
      "body": [
        {
          "type": "panel",
          "title": "Available Variables",
          "collapsable": true,
          "collapsed": true,
          "body": {
            "type": "tabs",
            "tabs": [
              {
                "title": "Documents",
                "body": {
                  "type": "each",
                  "name": "resources",
                  "items": {
                    "type": "collapse",
                    "header": "${label}",
                    "body": {
                      "type": "table",
                      "source": "${fields}",
                      "columns": [
                        {"name": "name",  "label": "Field"},
                        {"name": "type",  "label": "Type"},
                        {"name": "label", "label": "Description"},
                        {
                          "name": "example_expr",
                          "label": "Copy",
                          "type": "button",
                          "label": "Copy",
                          "actionType": "copy",
                          "content": "${example_expr}"
                        }
                      ]
                    }
                  }
                }
              },
              {
                "title": "Functions",
                "body": {
                  "type": "each",
                  "name": "functions",
                  "items": {
                    "type": "card",
                    "className": "m-b-sm",
                    "header": "${label} — ${signature}",
                    "body": [
                      {"type": "tpl", "tpl": "<p class='text-muted'>${description}</p>"},
                      {
                        "type": "each",
                        "name": "examples",
                        "items": {
                          "type": "tpl",
                          "tpl": "<code class='block p-xs bg-light m-b-xs'>${item}</code>"
                        }
                      }
                    ]
                  }
                }
              },
              {
                "title": "Context Keys",
                "body": {
                  "type": "table",
                  "source": "${context_keys}",
                  "columns": [
                    {"name": "key",         "label": "Key"},
                    {"name": "label",       "label": "Description"},
                    {"name": "type",        "label": "Type"},
                    {"name": "stage_source","label": "Set by Stage"}
                  ]
                }
              }
            ]
          }
        },
        {
          "type": "editor",
          "name": "formula",
          "label": "Condition Formula",
          "language": "javascript",
          "height": 200,
          "placeholder": "e.g. documentTotal(\"invoice\", input.invoice_id) > 50000",
          "options": {
            "minimap": {"enabled": false},
            "fontSize": 14,
            "wordWrap": "on"
          },
          "completionItems": "${functions.map(f => ({label: f.name, kind: 'Function', insertText: f.signature, documentation: f.description})).concat(context_keys.map(k => ({label: k.key, kind: 'Variable', documentation: k.label})))"
        },
        {
          "type": "button",
          "label": "Validate Expression",
          "level": "info",
          "actionType": "ajax",
          "api": {
            "method": "post",
            "url": "/api/pipeline/validate-expression",
            "data": { "formula": "${formula}", "operation_key": "${operation_key}" }
          },
          "messages": {
            "success": "Expression is valid",
            "failed":  "Expression has errors — check the formula"
          }
        }
      ]
    }
  ]
}
```

### 15.6 Expression Validation Endpoint

```go
// POST /api/pipeline/validate-expression
func ValidateExpressionHandler(deps *app.Deps) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req struct {
            Formula      string `json:"formula"`
            OperationKey string `json:"operation_key"`
        }
        c.BodyParser(&req)

        session := middleware.ContextSession(c)

        // 1. Parse the formula using condition package
        builder := condition.NewBuilder(condition.ConjunctionAnd)
        builder.AddFormula(req.Formula)
        group := builder.Build()

        if err := group.Rules[0].Validate(); err != nil {
            return c.Status(200).JSON(fiber.Map{
                "valid":   false,
                "errors":  []string{err.Error()},
                "message": "Invalid formula syntax",
            })
        }

        // 2. Check that referenced fields and functions are accessible
        warnings := deps.ExpressionValidator.CheckAccess(
            req.Formula, session, deps.SchemaRegistry)

        return c.Status(200).JSON(fiber.Map{
            "valid":    true,
            "warnings": warnings,  // e.g. "Function 'hasTag' may be slow"
            "message":  "Expression is valid",
        })
    }
}
```

### 15.7 How the UI Queries Document Relationships

When a tenant writes an expression that references document lines or related fields, the UI provides a **field explorer** that queries the schema registry:

```
GET /api/pipeline/document-schema/:type_key
→ Returns the full schema for a document type (fields + relationships)
→ Filtered by user permissions

GET /api/pipeline/document-schema/:type_key/relationships/:rel_name
→ Returns the fields available on a relationship
→ e.g. /api/pipeline/document-schema/invoice/relationships/lines
   → returns all fields on invoice_line records
```

**How the UI uses it to show the field explorer:**

```json
{
  "type": "dialog",
  "title": "Field Explorer",
  "body": {
    "type": "service",
    "api": "/api/pipeline/document-schema/${selectedDocType}",
    "body": [
      {
        "type": "tree",
        "source": "${buildFieldTree(data)}",
        "onSelect": "setFormulaFragment(getExpressionFor(item))"
      }
    ]
  }
}
```

The tree shows:

```
 AP Invoice (invoice)
  ├──  Total Amount           → fieldValue("invoice", input.invoice_id, "total_amount")
  ├──  Invoice Date           → fieldValue("invoice", input.invoice_id, "invoice_date")
  ├── ️  Status                → fieldValue("invoice", input.invoice_id, "status")
  └──  Invoice Lines (many)
        ├──  Line Amount      → lineCount("invoice", input.invoice_id)
        ├──  Item Code        → (use in loop conditions)
        └──  GL Account       → (use in loop conditions)
```

Clicking any field inserts the correct expression into the formula editor at the cursor position.

---

## 16. Customer Workflow Engine

### 16.1 Overview

The Workflow Engine is a separate module that handles long-running, human-in-the-loop business processes. Business users design workflows in a visual builder. Temporal provides durable execution (workflows survive restarts). The engine connects to the Pipeline via the `WorkflowApprovalHook`.

### 16.2 Database Schema

```sql
-- Workflow templates: customer-designed process definitions
CREATE TABLE workflow_templates (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           uuid        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name                text        NOT NULL,
  description         text,
  category            text,       -- 'approval' | 'notification' | 'data_processing' | 'compliance'
  trigger_type        text        NOT NULL DEFAULT 'manual',
  -- 'manual' | 'event' | 'scheduled' | 'webhook'
  trigger_event       text,       -- "ap.invoice.process.completed", "gl.transaction.posted"
  trigger_conditions  jsonb       DEFAULT '{}',
  definition          jsonb       NOT NULL,   -- see Section 16.3
  amis_builder_state  jsonb,                  -- visual builder UI state
  is_active           bool        NOT NULL DEFAULT true,
  is_draft            bool        NOT NULL DEFAULT false,
  version             int         NOT NULL DEFAULT 1,
  owner_id            uuid        NOT NULL REFERENCES users(id),
  execution_count     int         NOT NULL DEFAULT 0,
  success_count       int         NOT NULL DEFAULT 0,
  failure_count       int         NOT NULL DEFAULT 0,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

-- Workflow instances: individual executions
CREATE TABLE workflow_instances (
  id                   uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  template_id          uuid        NOT NULL REFERENCES workflow_templates(id),
  tenant_id            uuid        NOT NULL REFERENCES tenants(id),
  temporal_workflow_id text        NOT NULL UNIQUE,
  temporal_run_id      text        NOT NULL,
  trigger_type         text        NOT NULL,
  triggered_by         uuid        REFERENCES users(id),
  trigger_data         jsonb,
  status               text        NOT NULL DEFAULT 'pending',
  -- 'pending' | 'running' | 'paused' | 'completed' | 'failed' | 'cancelled'
  current_step_id      text,
  steps_completed      int         NOT NULL DEFAULT 0,
  total_steps          int         NOT NULL,
  completion_percentage int        NOT NULL DEFAULT 0,
  started_at           timestamptz NOT NULL DEFAULT now(),
  completed_at         timestamptz,
  context_data         jsonb       NOT NULL DEFAULT '{}',
  output_data          jsonb,
  error_message        text,
  -- Link to Pipeline operation (when spawned by WorkflowApprovalHook)
  pipeline_operation_id uuid       REFERENCES operation_logs(operation_id),
  related_record_type  text,       -- "invoice", "purchase_order", etc.
  related_record_id    uuid,
  priority             text        NOT NULL DEFAULT 'normal'
);

-- Workflow step executions
CREATE TABLE workflow_step_executions (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  instance_id         uuid        NOT NULL REFERENCES workflow_instances(id),
  tenant_id           uuid        NOT NULL REFERENCES tenants(id),
  step_id             text        NOT NULL,
  step_name           text        NOT NULL,
  step_type           text        NOT NULL,
  step_order          int         NOT NULL,
  status              text        NOT NULL DEFAULT 'pending',
  started_at          timestamptz NOT NULL DEFAULT now(),
  completed_at        timestamptz,
  duration_ms         int,
  input_data          jsonb,
  output_data         jsonb,
  error_message       text,
  assigned_to_user_id uuid        REFERENCES users(id),
  assigned_to_role    text,
  assigned_to_entity_id uuid      REFERENCES entities(id),
  completed_by        uuid        REFERENCES users(id),
  completion_comment  text
);

-- User tasks (approvals, reviews, decisions)
CREATE TABLE workflow_user_tasks (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  step_execution_id   uuid        NOT NULL REFERENCES workflow_step_executions(id),
  instance_id         uuid        NOT NULL REFERENCES workflow_instances(id),
  tenant_id           uuid        NOT NULL REFERENCES tenants(id),
  title               text        NOT NULL,
  description         text,
  task_type           text        NOT NULL DEFAULT 'approval',
  -- 'approval' | 'review' | 'input' | 'decision' | 'acknowledgment'
  priority            text        NOT NULL DEFAULT 'normal',
  -- Assignment (at least one must be set)
  assigned_to_user_id uuid        REFERENCES users(id),
  assigned_to_role    text,
  assigned_to_entity_id uuid      REFERENCES entities(id),
  assigned_at         timestamptz NOT NULL DEFAULT now(),
  -- Deadlines
  due_at              timestamptz,
  escalated_at        timestamptz,
  escalated_to        uuid        REFERENCES users(id),
  escalation_level    int         NOT NULL DEFAULT 0,
  -- Decision
  status              text        NOT NULL DEFAULT 'pending',
  -- 'pending' | 'in_progress' | 'completed' | 'cancelled' | 'escalated'
  completed_at        timestamptz,
  completed_by        uuid        REFERENCES users(id),
  decision            text,
  -- 'approved' | 'rejected' | 'delegated' | 'cancelled'
  comment             text,
  form_data           jsonb,
  -- UI configuration
  form_schema         jsonb,      -- AMIS form schema for data collection
  action_buttons      jsonb,      -- custom button configurations
  context_data        jsonb,      -- data displayed to help user decide
  -- Permissions
  can_delegate        bool        NOT NULL DEFAULT true,
  can_reassign        bool        NOT NULL DEFAULT false,
  requires_comment    bool        NOT NULL DEFAULT false,
  created_at          timestamptz NOT NULL DEFAULT now()
);

-- Workflow trigger queue (populated by TxHooks, consumed by workflow engine)
CREATE TABLE workflow_trigger_queue (
  id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    uuid        NOT NULL REFERENCES tenants(id),
  event_name   text        NOT NULL,
  payload      jsonb       NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  processed_at timestamptz,
  error_count  int         NOT NULL DEFAULT 0,
  last_error   text
);

-- Enable RLS on all workflow tables
ALTER TABLE workflow_templates       ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_instances       ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_step_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_user_tasks      ENABLE ROW LEVEL SECURITY;
ALTER TABLE workflow_trigger_queue   ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON workflow_templates       FOR ALL TO awo_app USING (tenant_id = current_setting('app.tenant_id')::uuid);
CREATE POLICY tenant_isolation ON workflow_instances       FOR ALL TO awo_app USING (tenant_id = current_setting('app.tenant_id')::uuid);
CREATE POLICY tenant_isolation ON workflow_step_executions FOR ALL TO awo_app USING (tenant_id = current_setting('app.tenant_id')::uuid);
CREATE POLICY tenant_isolation ON workflow_user_tasks      FOR ALL TO awo_app USING (tenant_id = current_setting('app.tenant_id')::uuid);
CREATE POLICY tenant_isolation ON workflow_trigger_queue   FOR ALL TO awo_app USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- Operation logs (Pipeline persistence)
CREATE TABLE operation_logs (
  id                uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         uuid        NOT NULL REFERENCES tenants(id),
  operation_id      uuid        NOT NULL UNIQUE,
  operation_key     text        NOT NULL,
  resource          text        NOT NULL,
  action            text        NOT NULL,
  status            text        NOT NULL,
  -- 'running' | 'completed' | 'pending_approval' | 'failed' | 'compensated'
  idempotency_key   text,
  input_snapshot    jsonb       NOT NULL,
  session_snapshot  jsonb       NOT NULL,
  stage_log         jsonb       NOT NULL DEFAULT '[]',
  output_data       jsonb       NOT NULL DEFAULT '{}',
  flags_snapshot    jsonb       NOT NULL DEFAULT '{}',
  suspend_reason    text,
  resume_point      text,
  failed_stage      text,
  error_message     text,
  compensated       bool        NOT NULL DEFAULT false,
  compensation_log  jsonb       NOT NULL DEFAULT '[]',
  http_status       int,
  http_response     jsonb,
  initiated_by      uuid        NOT NULL REFERENCES users(id),
  started_at        timestamptz NOT NULL,
  completed_at      timestamptz,
  duration_ms       int
);

CREATE UNIQUE INDEX idx_operation_logs_idempotency
  ON operation_logs(tenant_id, idempotency_key)
  WHERE idempotency_key IS NOT NULL;

CREATE INDEX idx_operation_logs_pending
  ON operation_logs(tenant_id)
  WHERE status = 'pending_approval';

ALTER TABLE operation_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON operation_logs FOR ALL TO awo_app
  USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### 16.3 Workflow Definition Language

Workflows are declarative JSON stored in `workflow_templates.definition`. The condition package evaluates step conditions:

```json
{
  "name": "Invoice Approval Workflow",
  "version": 1,
  "variables": [
    {"name": "invoice",        "type": "object",  "required": true},
    {"name": "approval_chain", "type": "array",   "default": []}
  ],
  "steps": [
    {
      "id": "validate",
      "type": "validation",
      "name": "Validate Invoice",
      "config": {
        "rules": [
          {"field": "$.invoice.total_amount", "operator": ">",      "value": 0},
          {"field": "$.invoice.vendor_id",    "operator": "exists"}
        ]
      },
      "on_success": "check_amount_threshold",
      "on_failure": "notify_error"
    },
    {
      "id": "check_amount_threshold",
      "type": "condition",
      "name": "Route by Amount",
      "config": {
        "cases": [
          {"condition": "$.invoice.total_amount >= 500000", "next_step": "require_cfo"},
          {"condition": "$.invoice.total_amount >= 100000", "next_step": "require_director"},
          {"condition": "$.invoice.total_amount >= 10000",  "next_step": "require_manager"}
        ],
        "default": "auto_approve"
      }
    },
    {
      "id": "require_manager",
      "type": "user_task",
      "name": "Manager Approval",
      "config": {
        "title": "Approve Invoice ${invoice.invoice_number}",
        "description": "Invoice from ${invoice.vendor_name} for ${invoice.total_amount}",
        "priority": "normal",
        "assign_to_role": "manager",
        "entity_scope": "$.invoice.entity_id",
        "timeout": "48h",
        "timeout_action": "escalate",
        "escalate_to_role": "director",
        "form_fields": [
          {"name": "comment", "type": "textarea", "label": "Comment", "required": false}
        ],
        "actions": [
          {"id": "approved",  "label": "Approve", "style": "primary"},
          {"id": "rejected",  "label": "Reject",  "style": "danger"},
          {"id": "delegated", "label": "Delegate","style": "default"}
        ]
      },
      "on_approved": "notify_approved",
      "on_rejected": "notify_rejected"
    },
    {
      "id": "notify_approved",
      "type": "notification",
      "name": "Notify Approval",
      "config": {
        "template": "invoice_approved",
        "channel": "email",
        "recipients": ["$.invoice.created_by_email"]
      },
      "on_success": "end"
    }
  ]
}
```

### 16.4 Temporal Integration

The Workflow Engine uses Temporal for durable execution. Key patterns:

```go
// CustomWorkflowExecution: the Temporal workflow function
func CustomWorkflowExecution(ctx workflow.Context, definition WorkflowDefinition,
    instanceID uuid.UUID, input map[string]any) error {

    // Execute steps sequentially with support for:
    // - User tasks (wait for signal)
    // - Parallel branches (child workflows)
    // - Wait steps (timers)
    // - Condition routing
    // - Loop iteration

    currentStepID := definition.Steps[0].ID
    for currentStepID != "" && currentStepID != "end" {
        step := findStep(definition.Steps, currentStepID)

        // For user tasks: create task, wait for Temporal signal
        if step.Type == "user_task" {
            result, err := executeUserTaskStep(ctx, instanceID, *step, wfContext)
            // ... handles approve/reject/delegate/timeout/escalation
        }
        // ... other step types
    }
}
```

The user task flow:
1. Temporal creates `workflow_user_tasks` row
2. Sends notification to assignee
3. Waits for signal `"task-{id}-completed"`
4. User approves in My Tasks UI → `POST /api/workflows/tasks/{id}/complete`
5. API sends Temporal signal → workflow resumes

---

## 17. Pipeline ↔ Workflow Engine Integration

### 17.1 The Connection — WorkflowApprovalHook

```go
// awo/internal/workflow/pipeline_hook.go

type WorkflowApprovalHook struct {
    workflowEngine *workflow.Engine
    approvalTemplateID uuid.UUID  // the pre-built "Approval Gate" template
}

func (h *WorkflowApprovalHook) StageBefore()  []string { return []string{"gl.post_transaction"} }
func (h *WorkflowApprovalHook) FeatureFlag()  string   { return "workflow" }
func (h *WorkflowApprovalHook) Operations()   []string { return []string{"*"} }
func (h *WorkflowApprovalHook) RunCondition() string   {
    // pkg/condition evaluates this:
    // Fire only when amount exceeds threshold OR budget was exceeded
    return `toNumber(input.total_amount) > toNumber(setting["workflow.approval_threshold_amount"] ?? "0") || budget_exceeded == true`
}

func (h *WorkflowApprovalHook) Execute(opCtx *pipeline.OperationContext) error {
    config, err := h.workflowEngine.GetApprovalRule(opCtx.Ctx, workflow.RuleQuery{
        TenantID:    opCtx.TenantID,
        OperationKey: opCtx.OperationKey,
        UserID:      opCtx.UserID,
        EntityID:    opCtx.EntityID,
        Amount:      getAmount(opCtx),
        ExtraFlags:  opCtx.Flags,
    })
    if err != nil { return err }
    if !config.RequiresApproval { return nil }

    // Start a Temporal-backed approval workflow
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

    // Suspend the pipeline — Temporal will call the resume URL when done
    opCtx.Suspend(
        "pending_workflow_approval:"+instance.ID.String(),
        "gl.post_transaction",  // resume from here
    )
    return nil
}

// SimulateExecute: dry-run shows approval requirement without suspending
func (h *WorkflowApprovalHook) SimulateExecute(opCtx *pipeline.OperationContext) error {
    config, _ := h.workflowEngine.GetApprovalRule(opCtx.Ctx, ...)
    opCtx.SetData("workflow.approval_required_in_real_run", config.RequiresApproval)
    opCtx.SetData("workflow.expected_approvers",            config.Approvers)
    // Does NOT call opCtx.Suspend() — dry-run continues through full pipeline
    return nil
}
```

### 17.2 Workflow Completion → Pipeline Resume

When the Temporal workflow completes (approver acts on task):

```go
// In workflow engine's Temporal activity:
func (a *Activities) CompleteWorkflowInstanceActivity(ctx context.Context,
    instanceID uuid.UUID, output map[string]any) error {

    // Update workflow_instances status
    a.store.CompleteWorkflowInstance(ctx, instanceID, output)

    // If this was an approval gate for a Pipeline operation, resume it
    if opIDStr, ok := output["operation_id"].(string); ok {
        opID, _ := uuid.Parse(opIDStr)
        decision := output["decision"].(string)

        if decision == "approved" {
            a.pipelineResumeService.Resume(ctx, opID, map[string]any{
                "approved_by": output["completed_by"],
                "decision":    decision,
                "comment":     output["comment"],
            })
        } else {
            // Rejected — mark pipeline as failed without compensation
            // (nothing was written — GL was gated by this hook)
            a.pipelineLogRepo.MarkRejected(ctx, opID,
                "Rejected by " + output["completed_by_name"].(string))
        }
    }
    return nil
}
```

### 17.3 DB Transaction Hook Seeds Workflow Trigger

When the GL posting stage commits, the `workflow_trigger_queue` row (seeded atomically via TxHook) is picked up by the Workflow Engine's trigger listener:

```go
// awo/internal/workflow/trigger_listener.go

// Polls workflow_trigger_queue for unprocessed rows
func (l *TriggerListener) Run(ctx context.Context) {
    ticker := time.NewTicker(2 * time.Second)
    for {
        select {
        case <-ctx.Done(): return
        case <-ticker.C:
            l.processQueue(ctx)
        }
    }
}

func (l *TriggerListener) processQueue(ctx context.Context) {
    rows, _ := l.repo.GetUnprocessed(ctx, 50)
    for _, row := range rows {
        // Find matching active workflow triggers
        triggers, _ := l.triggerRepo.FindByEvent(ctx, row.TenantID, row.EventName)
        for _, trigger := range triggers {
            // Evaluate trigger conditions using pkg/condition
            if l.evaluateTriggerConditions(trigger, row.Payload) {
                l.workflowEngine.StartWorkflow(ctx, trigger.TemplateID, row.Payload)
            }
        }
        l.repo.MarkProcessed(ctx, row.ID)
    }
}
```

---

## 18. Canonical Example: AP Invoice Processing

### 18.1 Complete Pipeline Definition

```
Operation: "ap.invoice.process"

Stage                           Module        Flag Gate          Priority  Required
─────────────────────────────────────────────────────────────────────────────────────
ap.validate_invoice             ap            —                  110       Yes
ap.duplicate_check              ap            —                  120       Yes
ap.resolve_vendor               ap            —                  210       Yes
ap.resolve_gl_accounts          ap            —                  220       Yes
procurement.three_way_match     procurement   "procurement"      300       No
budget.check                    budget        "budget"           310       No
compliance.sanctions_check      compliance    "compliance"       390       No
tax.resolve_codes               tax           "tax"              400       No
tax.calculate                   tax           "tax"              410       No
tax.apply_withholding           tax           "tax.withholding"  430       No
  ↕ HOOK before gl.post_transaction: workflow.approval_gate ("workflow")
  ↕ HOOK before gl.post_transaction: compliance.pre_posting_check ("compliance")
gl.post_transaction             finance       —                  610       Yes
  ↕ TxHook: gl.seed_domain_event (always)
  ↕ TxHook: gl.update_budget_actuals ("budget")
  ↕ TxHook: gl.seed_workflow_trigger ("workflow")
ar.apply_advance_payment        ar            "ar.advances"      620       No
dms.archive_document            dms           "dms"              710       No
banking.schedule_payment        banking       "banking"          720       No
  ↕ HOOK after banking.schedule_payment: notify.posting_complete (—)
  ↕ HOOK after banking.schedule_payment: tax.e_invoice_submit ("tax.e_invoice")
audit.log_operation             audit         —                  910       Yes
─────────────────────────────────────────────────────────────────────────────────────
```

### 18.2 Tenant Configurations

**Basic Finance tenant (3 modules):**
```
Active stages: validate → duplicate_check → resolve_vendor → resolve_gl_accounts → gl.post_transaction → audit
Active hooks: audit_log (always)
Pipeline stages: 6    Pipeline hooks: 1
```

**Full Enterprise tenant (all modules):**
```
Active stages: all 14 listed above
Active hooks: workflow.approval_gate, compliance.pre_posting, audit_log,
              notify.posting_complete, tax.e_invoice_submit
Pipeline stages: 14    Pipeline hooks: 5
```

Both call the same `APService.ProcessInvoice()`. The pipeline built for each tenant is different. The service code is identical.

### 18.3 Service Entry Point

```go
// awo/internal/ap/service.go

func (s *APService) ProcessInvoice(ctx context.Context,
    params ProcessInvoiceParams, idempotencyKey string) (*ProcessInvoiceResult, error) {

    session := domain.SessionFromContext(ctx)
    if !session.Can("ap.invoices", "process") {
        return nil, domain.ErrForbidden
    }

    invoice, err := s.invoiceRepo.Get(ctx, params.InvoiceID)
    if err != nil { return nil, err }

    p, err := s.pipelineBuilder.Build("ap.invoice.process", session)
    if err != nil { return nil, err }

    opCtx := pipeline.AcquireOperationContext()
    defer pipeline.ReleaseOperationContext(opCtx)
    opCtx.Ctx          = ctx
    opCtx.Session      = session
    opCtx.TenantID     = session.TenantID
    opCtx.UserID       = session.UserID
    opCtx.EntityID     = session.EntityID
    opCtx.OperationID  = uuid.New()
    opCtx.OperationKey = "ap.invoice.process"
    opCtx.Resource     = "ap.invoice"
    opCtx.Action       = "process"
    opCtx.Input        = invoice

    result, err := p.ExecuteIdempotent(opCtx, idempotencyKey)
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

---

## 19. All Module Stage Contributions

| Stage / Hook | Module | Operation(s) | Flag Gate | Priority | Required |
|---|---|---|---|---|---|
| `ap.validate_invoice` | AP | `ap.invoice.process` | — | 110 | Yes |
| `ap.duplicate_check` | AP | `ap.invoice.*` | — | 120 | Yes |
| `ap.resolve_vendor` | AP | `ap.invoice.*` | — | 210 | Yes |
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
| `gl.post_transaction` | Finance | all financial operations | — | 610 | Yes |
| `ar.apply_advance` | AR | `ap.invoice.process` | `ar.advances` | 620 | No |
| `ar.apply_credit_note` | AR | `ap.invoice.process` | `ar` | 630 | No |
| `inventory.update_grn` | Inventory | `ap.invoice.process` | `inventory` | 730 | No |
| `dms.archive` | DMS | `*` | `dms` | 710 | No |
| `banking.schedule_payment` | Banking | `ap.invoice.process` | `banking` | 720 | No |
| `audit.log_operation` | Audit | `*` | — | 910 | Yes |
| **HOOKS** | | | | | |
| `workflow.approval_gate` | Workflow | `*` | `workflow` | before `gl.post_transaction` | — |
| `audit.log_hook` | Audit | `*` | — | after `*` | — |
| `notify.posting` | Notifications | `*` | `notifications` | after `banking.schedule_payment` | — |
| `tax.e_invoice_submit` | Tax | `ar.invoice.*` | `tax.e_invoice` | after `gl.post_transaction` | — |
| `compliance.pre_posting` | Compliance | `*` | `compliance` | before `gl.post_transaction` | — |

---

## 20. Testing Strategy

### 20.1 Stage Tests — Complete Isolation

```go
// Test a stage in complete isolation — no pipeline needed
func TestBudgetCheckStage_HardBlock(t *testing.T) {
    stage := budget.NewBudgetCheckStage(&mockBudgetService{
        result: budget.CheckResult{WithinBudget: false, AvailableAmount: dec("50000")},
    })

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

func TestBudgetCheckStage_DrRun_SimulatesCorrectly(t *testing.T) {
    stage := budget.NewBudgetCheckStage(mockSvc)
    opCtx := buildTestOpCtx(t)
    opCtx.DryRun = true

    result, err := stage.Simulate(opCtx)
    assert.NoError(t, err)
    assert.Equal(t, "simulated", result.Status)
    assert.Contains(t, result.Outputs, "budget_check.available_amount")
}
```

### 20.2 TxHook Tests

```go
func TestGLPostingStage_RegistersTxHooks(t *testing.T) {
    stage := finance.NewGLPostingStage(mockGLSvc)
    opCtx := buildTestOpCtx(t)
    opCtx.Session = buildSessionWithFlags(t, map[string]bool{"budget": true, "workflow": true})

    _, err := stage.Execute(opCtx)
    assert.NoError(t, err)

    // Verify three TxHooks were registered
    assert.Len(t, opCtx.TxHooks, 3)
    hookNames := extractHookNames(opCtx.TxHooks)
    assert.Contains(t, hookNames, "gl.seed_domain_event")
    assert.Contains(t, hookNames, "gl.update_budget_actuals")
    assert.Contains(t, hookNames, "gl.seed_workflow_trigger")
}

func TestTxHooks_RollbackOnHookFailure(t *testing.T) {
    // Setup: GL succeeds but domain_events INSERT fails
    pipeline := buildTestPipeline(t, withStageSequence{
        finance.NewGLPostingStage(mockGLSvc),
    })
    opCtx := buildTestOpCtx(t)
    opCtx.RegisterTxHook(pipeline.TxHook{
        Name: "gl.seed_domain_event",
        Fn: func(ctx context.Context, tx pgx.Tx, opCtx *pipeline.OperationContext) error {
            return fmt.Errorf("domain_events table is locked")
        },
    })

    result, err := pipeline.Execute(opCtx)
    // The transaction should have been rolled back — GL entry does not exist
    assert.ErrorContains(t, err, "domain_events table is locked")
    assertGLEntryDoesNotExist(t, opCtx.Data["gl_posting.transaction_id"])
}
```

### 20.3 Pipeline Integration Tests

```go
func TestPipeline_APInvoice_FullEnterpriseFlow(t *testing.T) {
    registry := buildFullRegistryWithAllModules(t)
    session  := buildEnterpriseSession(t) // all flags enabled

    builder := pipeline.NewPipelineBuilder(registry, ...)
    p, _    := builder.Build("ap.invoice.process", session)

    // Verify stage count and order
    assert.Len(t, p.Stages(), 14)
    assert.Equal(t, "ap.validate_invoice", p.Stages()[0].Name())
    assert.Equal(t, "audit.log_operation", p.Stages()[len(p.Stages())-1].Name())
}

func TestPipeline_Resume_AfterWorkflowApproval(t *testing.T) {
    // 1. Execute until suspended
    result, _ := p.Execute(opCtx)
    assert.Equal(t, "pending_approval", result.Status)
    assert.True(t, opCtx.Suspended)

    // 2. Resume after approval
    resumeSvc := pipeline.NewResumeService(builder, logRepo)
    err := resumeSvc.Resume(ctx, opCtx.OperationID, map[string]any{
        "approved_by": cfoUserID, "decision": "approved",
    })
    assert.NoError(t, err)

    // 3. Verify GL was posted
    reloaded, _ := logRepo.GetByOperationID(ctx, opCtx.OperationID)
    assert.Equal(t, "completed", reloaded.Status)
    assert.NotNil(t, reloaded.OutputData["gl_posting.transaction_id"])
}
```

### 20.4 Expression Validation Tests

```go
func TestScriptStage_ValidFormula(t *testing.T) {
    stage := pipeline.NewScriptStage(pipeline.ScriptStageDefinition{
        Formula: `documentTotal("invoice", input.invoice_id) > 50000`,
        OnTrue:  "high_value_path",
        OnFalse: "standard_path",
    }, testEvaluator)

    opCtx := buildTestOpCtx(t)
    opCtx.Data["input"] = map[string]any{"invoice_id": uuid.New().String()}

    result, err := stage.Execute(opCtx)
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.Status)
}

func TestScriptStage_TimeoutHandledGracefully(t *testing.T) {
    // Formula that takes too long
    stage := pipeline.NewScriptStage(pipeline.ScriptStageDefinition{
        Formula: generateSlowFormula(100), // 100 nested conditions
    }, condition.NewEvaluator(nil, condition.EvalOptions{Timeout: 1*time.Millisecond}))

    result, err := stage.Execute(buildTestOpCtx(t))
    assert.NoError(t, err)  // script errors are non-fatal
    assert.Equal(t, "skipped", result.Status)
    assert.Contains(t, result.Message, "timed out")
}
```

### 20.5 Autocomplete API Tests

```go
func TestAutocompleteAPI_FiltersbyPermissions(t *testing.T) {
    // User with only AP read permission
    session := buildSessionWithPermissions(t, map[string]bool{
        "ap.invoices.read": true,
        // finance.journals.read: NOT set
    })

    resp := callAutocomplete(t, session, "ap.invoice.process")

    resourceKeys := extractResourceKeys(resp.Resources)
    assert.Contains(t,    resourceKeys, "invoice")       // user has ap.invoices.read
    assert.NotContains(t, resourceKeys, "gl_journal")    // no finance.journals.read
}

func TestAutocompleteAPI_ReturnsContextKeysForOperation(t *testing.T) {
    resp := callAutocomplete(t, session, "ap.invoice.process")

    contextKeyNames := extractContextKeyNames(resp.ContextKeys)
    assert.Contains(t, contextKeyNames, "budget_check.available_amount")
    assert.Contains(t, contextKeyNames, "tax_calculation.tax_amount")
    assert.Contains(t, contextKeyNames, "three_way_match.matched_po_id")
}
```

---

## 21. UI Design

### 21.1 Live Pipeline Progress

```json
{
  "type": "page",
  "title": "Operation Progress",
  "body": {
    "type": "service",
    "api": "/api/operations/${operationId}",
    "ws":  "/api/operations/${operationId}/live",
    "body": [
      {
        "type": "progress",
        "value": "${completedStages / totalStages * 100}",
        "showLabel": true
      },
      {
        "type": "each",
        "name": "stage_log",
        "items": {
          "type": "flex",
          "className": "m-b-xs p-xs border-l-4 rounded",
          "style": {
            "borderColor": "${status === 'ran' ? '#52c41a' : status === 'failed' ? '#f5222d' : status === 'skipped' ? '#d9d9d9' : '#1890ff'}"
          },
          "items": [
            {"type": "icon", "icon": "${status === 'ran' ? 'fa fa-check-circle' : status === 'failed' ? 'fa fa-times-circle' : 'fa fa-minus-circle'}"},
            {"type": "tpl", "tpl": "<strong class='m-l-sm'>${stage_name}</strong> <span class='text-muted text-xs'>${duration_ms}ms</span>"},
            {"type": "tpl", "tpl": "<div class='text-muted text-xs'>${skip_reason || message || ''}</div>"}
          ]
        }
      },
      {
        "type": "panel",
        "title": "Error & Compensation",
        "visibleOn": "${status === 'failed'}",
        "body": [
          {"type": "tpl", "tpl": "<p class='text-danger'>${error.message}</p>"},
          {"type": "tpl", "tpl": "<p class='text-info'>${error.user_action_required}</p>"},
          {
            "type": "button",
            "label": "Retry",
            "visibleOn": "${error.recoverable}",
            "actionType": "ajax",
            "api": {"method": "post", "url": "/api/operations/${operation_id}/retry"}
          }
        ]
      }
    ]
  }
}
```

### 21.2 My Tasks Inbox

```json
{
  "type": "page",
  "title": "My Tasks",
  "body": {
    "type": "crud",
    "api": {"method": "get", "url": "/api/workflows/tasks/my-tasks"},
    "interval": 30000,
    "columns": [
      {"name": "priority", "label": "Priority", "type": "mapping",
       "map": {"urgent": "<span class='label label-danger'>URGENT</span>", "high": "<span class='label label-warning'>HIGH</span>", "normal": "<span class='label label-info'>NORMAL</span>"}},
      {"name": "title",    "label": "Task",     "type": "text"},
      {"name": "due_at",   "label": "Due",      "type": "datetime", "format": "fromNow",
       "classNameExpr": "${DATETOTIME(due_at) < NOW() ? 'text-danger font-bold' : ''}"},
      {
        "type": "operation",
        "label": "Actions",
        "buttons": [{
          "type": "button", "label": "Review", "level": "primary", "size": "sm",
          "actionType": "dialog",
          "dialog": {
            "title": "${title}", "size": "lg",
            "body": {
              "type": "service",
              "api": "/api/workflows/tasks/${id}",
              "body": [
                {"type": "alert",    "level": "info",  "body": "${data.description}"},
                {"type": "panel",    "title": "Context", "body": {"type": "json", "source": "${data.context_data}"}},
                {
                  "type": "form",
                  "api": {"method": "post", "url": "/api/workflows/tasks/${id}/complete"},
                  "body": [
                    "${data.form_schema}",
                    {"type": "textarea", "name": "comment", "label": "Comment",
                     "required": "${data.requires_comment}"},
                    {"type": "hidden", "name": "decision", "value": ""}
                  ],
                  "actions": [
                    {"type": "button", "label": "Approve", "level": "primary", "actionType": "submit",
                     "onClick": "this.props.formStore.setValues({decision: 'approved'})"},
                    {"type": "button", "label": "Reject",  "level": "danger",  "actionType": "submit",
                     "onClick": "this.props.formStore.setValues({decision: 'rejected'})"}
                  ]
                }
              ]
            }
          }
        }]
      }
    ]
  }
}
```

### 21.3 Script Editor with Full Autocomplete

```json
{
  "type": "page",
  "title": "Workflow Condition Editor",
  "body": {
    "type": "service",
    "api": "/api/pipeline/autocomplete?operation_key=${operation_key}",
    "body": [
      {
        "type": "grid",
        "columns": [
          {
            "columnClassName": "w-72",
            "body": {
              "type": "panel",
              "title": "Variables & Functions",
              "body": {
                "type": "tabs",
                "tabs": [
                  {
                    "title": "Documents",
                    "body": {
                      "type": "each", "name": "resources",
                      "items": {
                        "type": "collapse",
                        "header": "${label} (${type_key})",
                        "body": {
                          "type": "each", "name": "fields",
                          "items": {
                            "type": "flex",
                            "className": "m-b-xs cursor-pointer hover:bg-gray-50 p-xs rounded",
                            "onClick": "insertAtCursor(example_expr)",
                            "items": [
                              {"type": "tpl", "tpl": "<span class='badge badge-${type === \"decimal\" ? \"success\" : type === \"string\" ? \"info\" : \"default\"}'>${type}</span>"},
                              {"type": "tpl", "tpl": "<span class='m-l-xs'>${label}</span>"}
                            ]
                          }
                        }
                      }
                    }
                  },
                  {
                    "title": "Functions",
                    "body": {
                      "type": "each", "name": "functions",
                      "items": {
                        "type": "card", "className": "m-b-xs cursor-pointer",
                        "onClick": "insertAtCursor(signature)",
                        "header": "${name}",
                        "body": {"type": "tpl", "tpl": "<code class='text-xs'>${signature}</code><p class='text-muted text-xs m-t-xs'>${description}</p>"}
                      }
                    }
                  },
                  {
                    "title": "Context",
                    "body": {
                      "type": "each", "name": "context_keys",
                      "items": {
                        "type": "flex",
                        "className": "m-b-xs cursor-pointer hover:bg-gray-50 p-xs rounded",
                        "onClick": "insertAtCursor(key)",
                        "items": [
                          {"type": "tpl", "tpl": "<code class='text-xs text-blue-600'>${key}</code>"},
                          {"type": "tpl", "tpl": "<span class='text-muted text-xs m-l-xs'>${label}</span>"}
                        ]
                      }
                    }
                  }
                ]
              }
            }
          },
          {
            "body": [
              {
                "type": "editor",
                "name": "formula",
                "label": "Condition Formula",
                "language": "javascript",
                "height": 250,
                "placeholder": "documentTotal(\"invoice\", input.invoice_id) > 50000",
                "options": {"minimap": {"enabled": false}, "fontSize": 14}
              },
              {
                "type": "button-toolbar",
                "buttons": [
                  {
                    "type": "button", "label": "Validate", "level": "info",
                    "actionType": "ajax",
                    "api": {
                      "method": "post", "url": "/api/pipeline/validate-expression",
                      "data": {"formula": "${formula}", "operation_key": "${operation_key}"}
                    }
                  },
                  {
                    "type": "button", "label": "Dry-Run Preview", "level": "default",
                    "actionType": "dialog",
                    "dialog": {
                      "title": "Dry-Run Preview",
                      "body": {
                        "type": "service",
                        "api": {"method": "post", "url": "/api/operations/dry-run",
                                "data": {"operation_key": "${operation_key}", "input": "${$$}"}},
                        "body": [
                          {"type": "alert", "level": "${would_succeed ? 'success' : 'danger'}",
                           "body": "${would_succeed ? 'Would succeed' : 'Would fail at: ' + stage_log.find(s => s.status === 'failed').stage_name}"},
                          {"type": "tpl",   "visibleOn": "${would_require_approval}",
                           "tpl": "<div class='alert alert-warning'>Would require approval from: ${expected_approvers.join(', ')}</div>"},
                          {"type": "json",  "name": "expected_gl_entries", "source": "${expected_gl_entries}"}
                        ]
                      }
                    }
                  }
                ]
              }
            ]
          }
        ]
      }
    ]
  }
}
```

---

## 22. Troubleshooting

### 22.1 Pipeline Failure Diagnostic SQL

```sql
-- Full operation log for a failed operation
SELECT
  stage_log->>'stage_name'  AS stage,
  stage_log->>'status'      AS status,
  stage_log->>'skip_reason' AS skip_reason,
  stage_log->>'message'     AS message,
  stage_log->>'error'       AS error,
  stage_log->>'duration_ms' AS duration_ms
FROM operation_logs,
  jsonb_array_elements(stage_log) AS stage_log
WHERE operation_id = '<uuid>'
ORDER BY stage_log->>'started_at';

-- All pending approvals
SELECT ol.operation_id, ol.operation_key, ol.suspend_reason, ol.started_at,
       u.email AS initiated_by
FROM operation_logs ol JOIN users u ON u.id = ol.initiated_by
WHERE ol.tenant_id = '<tenant_uuid>' AND ol.status = 'pending_approval';

-- Stage skipped — check if flag is off
SELECT ffd.flag_key, COALESCE(tff.enabled, ffd.default_value) AS effective
FROM feature_flag_definitions ffd
LEFT JOIN tenant_feature_flags tff ON tff.flag_id = ffd.id AND tff.tenant_id = '<uuid>'
WHERE ffd.flag_key = 'budget';
```

### 22.2 TxHook Failure

```sql
-- Find failed workflow trigger queue rows
SELECT * FROM workflow_trigger_queue
WHERE tenant_id = '<uuid>' AND processed_at IS NULL AND error_count > 0
ORDER BY created_at;

-- Manually re-process a stuck trigger
UPDATE workflow_trigger_queue SET error_count = 0 WHERE id = '<uuid>';
```

### 22.3 Common Error Reference

| Error | Cause | Fix |
|---|---|---|
| `stage X failed after 3 retries` | Persistent service failure | Check downstream service health |
| `TxHook Y failed — transaction rolled back` | Hook error within open transaction | Check `workflow_trigger_queue` for constraint violations |
| `operation cannot be resumed (status: completed)` | Double-resume call | Idempotency key collision; check `Idempotency-Key` header |
| `Script timed out (200ms limit)` | Complex tenant formula | Simplify expression; break into multiple script stages |
| `Field X not accessible` in expression | Missing permission | Grant user `{resource}.read` permission |
| `Function Y not registered` in expression | Tenant called unknown function | Check registered functions in autocomplete API |
| Workflow approval task not appearing in My Tasks | Role assignment mismatch | Verify `assigned_to_role` in task matches user's role |
| Pipeline cache stale after flag change | Cache not invalidated | Verify `FlagService.Set()` calls `pipelineBuilder.InvalidateForTenant()` |
