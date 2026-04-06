# Pipeline & Workflow Engine — Implementation Task List

> Each task is atomic: one focused session, one PR-stage, one test run to verify.
> Work proceeds in order within each phase. A task is "done" only when its test criteria pass.
> After each task passes: stage the files → proceed to next task.

---

## Phase 1 — Core Pipeline Infrastructure

### P-001 · OperationContext

**Description**
Create `internal/pipeline/context.go`. Define the `OperationContext` struct with all fields from the spec, the `StageLog` struct, and all convenience methods: `GetData`, `SetData`, `SetFlag`, `Flag`, `Can`, `FeatureEnabled`, `SettingDecimal`, `SettingBool`, `SettingString`, `Suspend`, `RegisterTxHook`. Add `AcquireOperationContext` / `ReleaseOperationContext` using `sync.Pool` with pre-allocated maps (Data cap 32, Flags cap 16, Log cap 16). Add a `reset()` method that clears all fields for pool reuse.

**Files**
- `internal/pipeline/context.go`

**Expected Outcome**
- `OperationContext` compiles with all documented fields.
- Pool acquire/release correctly zeros the struct between uses.
- `Suspend()` sets `Suspended = true`, `SuspendReason`, `ResumePoint`.
- `RegisterTxHook()` appends to `TxHooks` slice.
- `FeatureEnabled(key)` delegates to `Session.Configuration.Flags[key]`.
- `SettingDecimal/Bool/String` return the default value when the key is absent.

**How to Test**
Write `internal/pipeline/context_test.go`:
```
TestOperationContext_SetAndGetData         → set "foo.bar" = 42, get returns 42, ok=true
TestOperationContext_GetData_Missing       → get "missing" returns nil, ok=false
TestOperationContext_SetFlag               → SetFlag("budget_exceeded", true); Flag() returns true
TestOperationContext_Flag_Default          → Flag on unset key returns false
TestOperationContext_Suspend               → after Suspend("reason","stage"): Suspended=true, fields set
TestOperationContext_RegisterTxHook        → two RegisterTxHook calls → len(TxHooks)==2
TestOperationContext_Pool_Reset            → Acquire, SetData, Release, Acquire again → Data is empty
TestOperationContext_SettingDecimal_Default → missing key returns provided default
```

---

### P-002 · Stage Interface + BaseStage + StageResult

**Description**
Create `internal/pipeline/stage.go`. Define the `Stage` interface exactly as specified. Define `BaseStage` struct embedding the common fields (name, operations, featureFlag, priority, required, runCondition, dependsOn) and implementing all interface methods except `Execute`. Define the `StageResult` struct. Define the `Simulatable` interface. Define `stageCheckpoint` (internal struct used by compensation: StageName + Output).

**Files**
- `internal/pipeline/stage.go`

**Expected Outcome**
- Any struct embedding `BaseStage` and implementing `Execute()` satisfies the `Stage` interface.
- `BaseStage.Operations()` returns the operations slice; `"*"` wildcard is a valid value.
- `BaseStage.FeatureFlag()` returns empty string when no flag set.
- `StageResult.Status` accepts `"completed"`, `"skipped"`, `"suspended"`, `"simulated"`, `"failed"`.

**How to Test**
Write `internal/pipeline/stage_test.go`:
```
TestBaseStage_InterfaceSatisfied      → compile-time check: var _ Stage = &testStage{}
TestBaseStage_Defaults               → zero-value BaseStage: Required()=false, RunCondition()=""
TestBaseStage_Name                   → Name() returns what was set in constructor
TestBaseStage_Operations_Wildcard    → Operations() = ["*"] is valid, returned correctly
TestStageResult_Fields               → StageResult with all fields set, verify no data loss
```

---

### P-003 · StageRegistry

**Description**
Create `internal/pipeline/stage_registry.go`. Define `StageRegistry` with a `sync.RWMutex`-protected map keyed by stage name. Implement `Register(...Stage)` (variadic, panics on duplicate name), `ForOperation(operationKey string) []Stage` (returns stages matching the operation key or `"*"` wildcard), `All() []Stage`.

**Files**
- `internal/pipeline/stage_registry.go`

**Expected Outcome**
- `Register` with duplicate name panics immediately (fast failure at startup).
- `ForOperation("ap.invoice.process")` returns stages whose `Operations()` contains `"ap.invoice.process"` or `"*"`.
- Thread-safe: concurrent reads are safe.

**How to Test**
Write `internal/pipeline/stage_registry_test.go`:
```
TestStageRegistry_Register_And_ForOperation    → register 3 stages, ForOperation returns correct subset
TestStageRegistry_Wildcard                     → stage with Operations=["*"] appears in any ForOperation call
TestStageRegistry_DuplicateNamePanics          → Register same name twice → panic
TestStageRegistry_All                          → All() returns all registered stages
TestStageRegistry_EmptyOperation               → ForOperation("unknown") returns empty slice
```

---

### P-004 · Hook Interface + HookRegistry

**Description**
Create `internal/pipeline/hook.go`. Define the `Hook` interface (Name, FeatureFlag, Priority, Operations, StageBefore, StageAfter, RunCondition, Execute). Define the `SimulatableHook` interface. Create `internal/pipeline/hook_registry.go`. Define `HookRegistry` with `Register(...Hook)` (panics on duplicate), `HooksForStage(operationKey, stageName string, session) (before []Hook, after []Hook)` — returns hooks whose `Operations()` matches and whose `StageBefore()`/`StageAfter()` contains the stageName or `"*"`, filtered by `FeatureFlag` against session flags, sorted by `Priority()`.

**Files**
- `internal/pipeline/hook.go`
- `internal/pipeline/hook_registry.go`

**Expected Outcome**
- `HooksForStage` correctly splits before/after hooks.
- Hooks with `StageBefore=["*"]` appear before every stage.
- Feature-flagged hooks are excluded when session flag is off.
- Hooks within same position (before/after) are sorted by `Priority()` ascending.

**How to Test**
Write `internal/pipeline/hook_registry_test.go`:
```
TestHookRegistry_BeforeHook               → hook with StageBefore=["gl.post"] returned for that stage
TestHookRegistry_AfterHook                → hook with StageAfter=["banking.pay"] returned correctly
TestHookRegistry_WildcardAfter            → StageAfter=["*"] appears after every stage
TestHookRegistry_FeatureFlag_Off          → hook with FeatureFlag="workflow" excluded when flag=false
TestHookRegistry_FeatureFlag_On           → same hook included when flag=true
TestHookRegistry_PriorityOrder            → two before-hooks on same stage returned sorted by Priority
TestHookRegistry_DuplicateNamePanics      → Register same name twice → panic
```

---

### P-005 · CompensationRegistry

**Description**
Create `internal/pipeline/compensation.go`. Define `CompensationFn = func(opCtx *OperationContext) error`. Define `CompensationRegistry` with `Register(stageName string, fn CompensationFn)`, `Get(stageName string) (CompensationFn, bool)`.

**Files**
- `internal/pipeline/compensation.go`

**Expected Outcome**
- `Get` on a registered stage returns the function and `true`.
- `Get` on an unregistered stage returns `nil, false`.
- Multiple registrations for the same stage name are allowed (last-write wins) — or panic, document the choice.

**How to Test**
Write `internal/pipeline/compensation_test.go`:
```
TestCompensationRegistry_RegisterAndGet   → register fn, Get returns it
TestCompensationRegistry_MissingKey       → Get on unknown stage returns nil, false
TestCompensationRegistry_Override         → register same stage twice, Get returns last-registered fn
```

---

### P-006 · TxHook Type

**Description**
Create `internal/pipeline/tx_hook.go`. Define the `TxHook` struct: `Name string`, `Priority int`, `Fn func(ctx context.Context, tx pgx.Tx, opCtx *OperationContext) error`. This is a data type only — execution is in the Pipeline.

**Files**
- `internal/pipeline/tx_hook.go`

**Expected Outcome**
- `TxHook` compiles and is usable from `OperationContext.RegisterTxHook(TxHook{...})`.
- `Fn` has the correct signature accepting `pgx.Tx`.

**How to Test**
Covered by P-001 `TestOperationContext_RegisterTxHook`. Add a compile-time check in `tx_hook_test.go`:
```
TestTxHook_CanBeConstructed   → build a TxHook literal, verify field access compiles
```

---

### P-007 · Pipeline.Execute()

**Description**
Create `internal/pipeline/pipeline.go`. Define the `Pipeline` struct. Implement `Execute(opCtx *OperationContext) (*OperationResult, error)` following the exact 6-step loop from the spec:
1. RunCondition gate (via `condition.Evaluator`)
2. Before hooks
3. Stage execute (or Simulate in dry-run)
4. After hooks
5. Emit `StageCompleted` event via `EventBus`
6. After all stages: run TxHooks

Define `OperationResult` struct (Status, OperationID, StageLog, OutputData, Error, Compensated). Define `EventBus` interface with `Emit(ctx, event any)`. Define `OperationLogRepository` interface with `SaveLog`, `GetByOperationID`, `GetByIdempotencyKey`, `SetIdempotencyKey`, `MarkRejected`. Implement `abort()` (calls compensate, saves log synchronously, returns error). Log writes on success are non-blocking goroutines (5s timeout). `determineStatus(opCtx)` returns `"completed"` / `"pending_approval"` / `"failed"`.

**Files**
- `internal/pipeline/pipeline.go`
- `internal/pipeline/result.go`
- `internal/pipeline/event_bus.go`  (interface + noop impl)
- `internal/pipeline/log_repo.go`   (interface)

**Expected Outcome**
- Pipeline with zero stages executes without error, returns `"completed"`.
- A stage that returns an error and is `Required()=true` causes abort.
- A stage that returns an error and is `Required()=false` logs and continues.
- `opCtx.Suspended = true` after any stage causes pipeline to stop (no further stages).
- DryRun=true calls `Simulate()` on simulatable stages.
- Before hooks run before the stage; after hooks run after.
- TxHooks run after all stages when not suspended and not dry-run.

**How to Test**
Write `internal/pipeline/pipeline_test.go`:
```
TestPipeline_NoStages_Completes
TestPipeline_SingleStage_Completes
TestPipeline_RequiredStage_Failure_Aborts
TestPipeline_OptionalStage_Failure_Continues
TestPipeline_Suspended_StopsExecution
TestPipeline_DryRun_CallsSimulate
TestPipeline_DryRun_NoSimulatable_CallsExecute
TestPipeline_BeforeHook_RunsBefore
TestPipeline_AfterHook_RunsAfter
TestPipeline_TxHooks_RunAfterAllStages
TestPipeline_TxHooks_SkippedWhenSuspended
TestPipeline_TxHooks_SkippedInDryRun
TestPipeline_StageLog_RecordsAllStatuses
```

---

### P-008 · Retry with Backoff

**Description**
Add `retryWithBackoff(opCtx, stage Stage) (StageResult, error)` to `pipeline.go`. 3 retries, base delay 500ms, exponential backoff with jitter (delay/4). Define `isRetryable(err) bool` — initially checks for a `RetryableError` interface (`IsRetryable() bool`). Wire it into the Execute loop: if `isRetryable(err)` → retry before checking `Required()`.

**Files**
- `internal/pipeline/pipeline.go` (additions)
- `internal/pipeline/errors.go`

**Expected Outcome**
- A retryable error that succeeds on attempt 2 → stage marked completed, no abort.
- A retryable error that fails all 3 retries → treated as normal failure (required=abort, optional=continue).
- Non-retryable errors skip retry entirely.

**How to Test**
Add to `pipeline_test.go`:
```
TestPipeline_RetryableError_SucceedsOnRetry2
TestPipeline_RetryableError_ExhaustsAllRetries_Required_Aborts
TestPipeline_RetryableError_ExhaustsAllRetries_Optional_Continues
TestPipeline_NonRetryableError_NoRetry
```

---

### P-009 · LIFO Compensation

**Description**
Add `compensate(opCtx *OperationContext, completed []stageCheckpoint)` to `pipeline.go`. Iterates `completed` in reverse order, calls `compensationReg.Get(stageName)`, invokes `fn(opCtx)`. Compensation failures are emitted as `domain.CompensationFailed` events but do NOT stop other compensations. After all compensations, `opCtx.SetFlag("compensated", true)`. Update `abort()` to call `compensate()`.

**Files**
- `internal/pipeline/pipeline.go` (additions)

**Expected Outcome**
- Three stages complete, then failure: compensation runs in reverse order (stage3 → stage2 → stage1).
- A compensation function that fails does not prevent stage1 from being compensated.
- Stages without a registered compensation function are silently skipped.
- `opCtx.Flag("compensated")` is true after `abort()`.

**How to Test**
Add to `pipeline_test.go`:
```
TestPipeline_Compensation_LIFOOrder
TestPipeline_Compensation_FailureDoesNotStop
TestPipeline_Compensation_NoFnRegistered_SkipsGracefully
TestPipeline_Compensation_FlagSet
```

---

### P-010 · PipelineBuilder + Cache

**Description**
Create `internal/pipeline/builder.go`. Define `PipelineBuilder` struct with `stageRegistry`, `hookRegistry`, `compensationReg`, `conditionEvaluator`, `logRepo`, `events`, `cache`. Implement `Build(operationKey string, session *domain.ResolvedSession) (*Pipeline, error)`:
1. Compute `flagsHash` (SHA-256 of relevant flag keys + values, sorted)
2. Check `cache.Get(operationKey + ":" + flagsHash)`
3. If miss: filter stages by `FeatureFlag`, sort by `Priority`, collect hooks, build `Pipeline`, `cache.Set(key, pipeline, 5*time.Minute)`

Define `pipelineCache` with `sync.RWMutex`, TTL-based entries. Implement `InvalidateForTenant(tenantID uuid.UUID)` that deletes all cache entries whose key starts with the tenantID prefix.

**Files**
- `internal/pipeline/builder.go`
- `internal/pipeline/cache.go`

**Expected Outcome**
- Two `Build()` calls with identical session flags return the same `*Pipeline` pointer (cache hit).
- Changing a feature flag (new session) causes a cache miss and fresh build.
- `InvalidateForTenant` causes subsequent Build to miss cache.
- Stages are correctly filtered: flag=`"budget"` stage excluded when session flag is off.
- Stages sorted by Priority ascending.

**How to Test**
Write `internal/pipeline/builder_test.go`:
```
TestPipelineBuilder_CacheHit_ReturnsSamePointer
TestPipelineBuilder_CacheMiss_OnFlagChange
TestPipelineBuilder_InvalidateForTenant_ClearsCacheEntries
TestPipelineBuilder_StageFiltering_ByFeatureFlag
TestPipelineBuilder_StageOrdering_ByPriority
TestPipelineBuilder_NilFeatureFlag_AlwaysIncluded
TestPipelineBuilder_CacheTTL_Expiry
```

---

### P-011 · Idempotency + ExecuteIdempotent

**Description**
Add `ExecuteIdempotent(opCtx *OperationContext, idempotencyKey string) (*OperationResult, error)` to `Pipeline`. If key is empty, call `Execute` directly. Otherwise: call `repo.GetByIdempotencyKey()` first. Handle the four existing-status cases: `completed` → return cached result, `running` → return `ErrOperationInProgress`, `pending_approval` → return `ErrPendingApproval`, `failed` → return `ErrOperationFailed`. On miss: `repo.SetIdempotencyKey()`, then `Execute`. Define these sentinel errors in `errors.go`.

**Files**
- `internal/pipeline/pipeline.go` (additions)
- `internal/pipeline/errors.go` (additions)

**Expected Outcome**
- Submitting same idempotency key twice returns the same result the second time without re-executing.
- `ErrOperationInProgress` returned when status is `"running"`.
- Empty key bypasses idempotency check entirely.

**How to Test**
Add to `pipeline_test.go`:
```
TestPipeline_Idempotent_EmptyKey_ExecutesNormally
TestPipeline_Idempotent_CompletedOperation_ReturnsCached
TestPipeline_Idempotent_RunningOperation_ReturnsErrInProgress
TestPipeline_Idempotent_PendingApproval_ReturnsErrPendingApproval
TestPipeline_Idempotent_FailedOperation_ReturnsErrFailed
TestPipeline_Idempotent_NewKey_Executes
```

---

### P-012 · PipelineResumeService

**Description**
Create `internal/pipeline/resume.go`. Define `PipelineResumeService` with `pipelineBuilder`, `repo`, `sessionRepo`. Implement `Resume(ctx, operationID uuid.UUID, resumeData map[string]any) (*OperationResult, error)`:
1. `repo.GetByOperationID()` — error if status != `"pending_approval"`
2. Reconstruct `OperationContext` from persisted log fields (InputSnapshot, OutputData, FlagsSnapshot, StageLog, SessionSnapshot)
3. Inject `resumeData` as `"resume.{k}"` keys, set flag `"workflow_approved": true`, `Suspended = false`
4. Rebuild pipeline with `pipelineBuilder.Build()`
5. Call `p.SkipStages(completedNames)` and `p.StartFrom(log.ResumePoint)` on the pipeline
6. Call `p.Execute(opCtx)`

Add `SkipStages(names []string)` and `StartFrom(stageName string)` to `Pipeline`. `SkipStages` filters the stages slice before execution. `StartFrom` sets an internal cursor so execution starts at that stage name.

Define `SessionRepository` interface with `Reconstruct(ctx, snapshot json.RawMessage) *domain.ResolvedSession`.

**Files**
- `internal/pipeline/resume.go`
- `internal/pipeline/pipeline.go` (SkipStages, StartFrom additions)
- `internal/pipeline/session_repo.go` (interface)

**Expected Outcome**
- `Resume` on a `"pending_approval"` operation reconstructs context and continues execution.
- Already-completed stages are skipped; execution begins at `ResumePoint`.
- `Resume` on a `"completed"` or `"failed"` operation returns an error.
- `resumeData` values are accessible as `opCtx.Data["resume.approved_by"]` etc.

**How to Test**
Write `internal/pipeline/resume_test.go`:
```
TestResumeService_Resume_CompletesOperation
TestResumeService_Resume_SkipsCompletedStages
TestResumeService_Resume_StartsFromResumePoint
TestResumeService_Resume_InjectsResumeData
TestResumeService_Resume_NonPendingStatus_Errors
TestResumeService_Resume_SetsWorkflowApprovedFlag
```

---

## Phase 2 — Condition / Scripting Layer

### C-001 · Pipeline Condition Evaluator Setup

**Description**
Create `internal/pipeline/condition_setup.go`. Define `BuildPipelineConditionEvaluator(deps *ServiceDeps) *condition.Evaluator` that:
1. Creates evaluator with production `EvalOptions` (MaxDepth:10, MaxConditions:100, Timeout:500ms, CacheResults:true)
2. Registers all domain functions: `documentTotal`, `lineCount`, `fieldValue`, `isChildOf`, `isWithinPeriod`, `hasTag`, `isApprovedVendor`
3. Each function validates argument count, checks `session.Can()` via `sessionFromEvalCtx(ec)` before reading data

Define `ServiceDeps` struct with references to all services needed by registered functions (`DocumentSvc`, `EntitySvc`, `VendorSvc`, `ContactSvc`, `SchemaReg`). Define `sessionFromEvalCtx(ec *condition.EvalContext)` helper. Define `tenantScriptOptions` var (MaxDepth:3, MaxConditions:20, Timeout:200ms).

Also implement `buildEvalData(opCtx *OperationContext) map[string]any` — merges Data, Flags, `feature`, `setting`, `input`, `entity_id`, `tenant_id` into a flat map. Does NOT expose `Session.Permissions` or raw DB access.

**Files**
- `internal/pipeline/condition_setup.go`

**Expected Outcome**
- `documentTotal` with insufficient args returns a wrapped error, not a panic.
- `fieldValue` with no read permission on the resource returns `"access denied"` error.
- `buildEvalData` exposes `feature`, `setting`, `input` keys; does NOT expose permission maps.
- `tenantScriptOptions` has 200ms timeout (tighter than developer options).

**How to Test**
Write `internal/pipeline/condition_setup_test.go`:
```
TestBuildEvalData_ExposesDataAndFlags
TestBuildEvalData_ExposesFeatureAndSetting
TestBuildEvalData_DoesNotExposePermissions
TestBuildEvalData_InputMap
TestDocumentTotal_InsufficientArgs_Error
TestDocumentTotal_NoPermission_Error
TestDocumentTotal_ValidCall_ReturnsMockTotal
TestFieldValue_FieldNotGettable_Error
TestTenantScriptOptions_Timeout_200ms
```

---

### C-002 · ScriptStage

**Description**
Create `internal/pipeline/stages/script_stage.go`. Define `ScriptStageDefinition` struct and `ScriptStage` struct embedding `pipeline.BaseStage`. User scripts are always `Required() = false`. Implement `Execute()`:
1. Build `EvalContext` with `tenantScriptOptions`
2. Build condition group from `definition.Formula`
3. Evaluate; on error → `classifyScriptError(err)` → return `StageResult{Status:"skipped", Message:...}` (never propagate error)
4. Write `"script.{id}.result"` and `"script.{id}.rules_evaluated"` to `opCtx.Data`
5. Set `NextStageID` based on result → `OnTrue` or `OnFalse`

Implement `classifyScriptError` mapping condition package sentinel errors to human-readable messages.

**Files**
- `internal/pipeline/stages/script_stage.go`

**Expected Outcome**
- A valid formula that evaluates true → `Status:"completed"`, `NextStageID = OnTrue`.
- A valid formula that evaluates false → `Status:"completed"`, `NextStageID = OnFalse`.
- A timeout → `Status:"skipped"`, message contains `"timed out"`, no error returned.
- Exceeding max depth → `Status:"skipped"`, no error returned.
- Unknown function → `Status:"skipped"`, message contains `"not registered"`.

**How to Test**
Write `internal/pipeline/stages/script_stage_test.go`:
```
TestScriptStage_TrueFormula_OnTrue
TestScriptStage_FalseFormula_OnFalse
TestScriptStage_Timeout_Skipped_NotError
TestScriptStage_MaxDepthExceeded_Skipped
TestScriptStage_UnknownFunction_Skipped
TestScriptStage_WritesResultToOpCtxData
TestScriptStage_Required_AlwaysFalse
```

---

### C-003 · SchemaRegistry

**Description**
Create `internal/pipeline/schema_registry.go`. Define `DocumentSchema`, `GettableField`, `DocumentRelationship` structs exactly as specified. Define `SchemaRegistry` with `sync.RWMutex`. Implement `Register(*DocumentSchema)`, `All() []*DocumentSchema`, `Get(typeKey string) (*DocumentSchema, bool)`, `IsGettableField(typeKey, fieldName string, permissions map[string]bool) bool`, `GetContextKeysForOperation(operationKey string) []ContextKeyAutocomplete`.

`IsGettableField` checks: field exists in schema AND user has the required permission.
`GetContextKeysForOperation` returns the well-known `opCtx.Data` keys that stages for that operation will populate (registered at startup alongside schemas).

**Files**
- `internal/pipeline/schema_registry.go`

**Expected Outcome**
- `Register` stores schema; `Get("invoice")` returns it.
- `IsGettableField` returns false when field name doesn't exist in schema.
- `IsGettableField` returns false when required permission is missing from the map.
- `IsGettableField` returns true when field exists and permission is present.
- `GetContextKeysForOperation("ap.invoice.process")` returns keys including `"budget_check.available_amount"` (after budget stage registers its outputs).

**How to Test**
Write `internal/pipeline/schema_registry_test.go`:
```
TestSchemaRegistry_RegisterAndGet
TestSchemaRegistry_Get_Missing_ReturnsFalse
TestSchemaRegistry_IsGettableField_ValidFieldWithPermission
TestSchemaRegistry_IsGettableField_MissingPermission
TestSchemaRegistry_IsGettableField_UnknownField
TestSchemaRegistry_All_ReturnsAllSchemas
TestSchemaRegistry_GetContextKeys_KnownOperation
```

---

### C-004 · Autocomplete API Handler

**Description**
Create `internal/pipeline/handlers/autocomplete.go`. Implement `AutocompleteHandler(deps)` for `GET /api/pipeline/autocomplete?operation_key=...`. Implement `buildAutocompleteResponse(deps, session, operationKey)` which:
1. Iterates `SchemaRegistry.All()`, filters by `session.Can(readPermission, "read")`
2. For each allowed schema, filters fields by field-level permission
3. Builds `ResourceAutocomplete` with `ExampleExpr` for each field
4. Adds all registered functions
5. Adds `ContextKeys` from `SchemaRegistry.GetContextKeysForOperation(operationKey)`
6. Adds `FeatureKeys`, `SettingKeys` from session
7. Adds `standardOperators()`

Define all response types: `AutocompleteResponse`, `ResourceAutocomplete`, `FieldAutocomplete`, `FunctionAutocomplete`, `ContextKeyAutocomplete`, `RelationshipAutocomplete`, `OperatorDef`.

**Files**
- `internal/pipeline/handlers/autocomplete.go`
- `internal/pipeline/handlers/types.go`

**Expected Outcome**
- User with `ap.invoices.read` sees `"invoice"` resource but not `"gl_journal"` (requires `finance.journals.read`).
- User without any finance read permission sees zero resources from finance module.
- `functions` list always contains `documentTotal`, `lineCount`, `fieldValue`, etc.
- `context_keys` contains stage output keys appropriate for the operation.

**How to Test**
Write `internal/pipeline/handlers/autocomplete_test.go`:
```
TestAutocomplete_FiltersResourcesByPermission
TestAutocomplete_ExcludesResourceWithoutPermission
TestAutocomplete_AlwaysReturnsFunctions
TestAutocomplete_ContextKeysForOperation
TestAutocomplete_ExcludesFieldsWithHigherPermission
TestAutocomplete_FeatureAndSettingKeys
```

---

### C-005 · Expression Validation Handler

**Description**
Create `internal/pipeline/handlers/validate_expression.go`. Implement `ValidateExpressionHandler(deps)` for `POST /api/pipeline/validate-expression`. Body: `{formula, operation_key}`. Steps:
1. Parse formula via condition package
2. Validate syntax: `group.Rules[0].Validate()`
3. Check field/function access via `deps.ExpressionValidator.CheckAccess(formula, session, schemaReg)` → returns `[]string` warnings
4. Return `{valid, warnings, message}`

Define `ExpressionValidator` interface with `CheckAccess(...) []string`.

**Files**
- `internal/pipeline/handlers/validate_expression.go`
- `internal/pipeline/expression_validator.go` (interface)

**Expected Outcome**
- Valid formula → `{valid: true, warnings: []}`.
- Syntax error → `{valid: false, errors: [...]}`.
- Valid formula referencing slow function → `{valid: true, warnings: ["Function 'hasTag' may be slow"]}`.
- Empty formula → `{valid: false}`.

**How to Test**
Write `internal/pipeline/handlers/validate_expression_test.go`:
```
TestValidateExpression_ValidFormula
TestValidateExpression_SyntaxError
TestValidateExpression_SlowFunctionWarning
TestValidateExpression_EmptyFormula_Invalid
```

---

## Phase 3 — Module Stages

### S-001 · ap.validate_invoice

**Description**
Create `internal/ap/pipeline_stages.go`. Define `ValidateInvoiceStage` (Priority:110, Required:true, Operations:`["ap.invoice.process", "ap.invoice.approve"]`). `Execute()` validates: invoice is not nil, `TotalAmount > 0`, `VendorID` is set, `InvoiceDate` is not zero, `Status == "submitted"`. Returns a descriptive error for each validation failure.

**Files**
- `internal/ap/pipeline_stages.go`

**Expected Outcome**
- Valid invoice → `StageResult{Status:"completed"}`.
- Nil input → error `"invoice is required"`.
- Zero amount → error `"total_amount must be greater than zero"`.
- Missing vendor → error `"vendor_id is required"`.

**How to Test**
Write `internal/ap/pipeline_stages_test.go`:
```
TestValidateInvoiceStage_ValidInvoice_Completes
TestValidateInvoiceStage_NilInput_Error
TestValidateInvoiceStage_ZeroAmount_Error
TestValidateInvoiceStage_MissingVendorID_Error
TestValidateInvoiceStage_ZeroDate_Error
TestValidateInvoiceStage_WrongStatus_Error
TestValidateInvoiceStage_Priority_110
TestValidateInvoiceStage_Required_True
```

---

### S-002 · ap.duplicate_check

**Description**
Add `DuplicateCheckStage` to `internal/ap/pipeline_stages.go` (Priority:120, Required:true). Depends on a `DuplicateCheckRepository` interface: `ExistsPosted(ctx, tenantID, vendorID, invoiceRef string, amount decimal.Decimal) (bool, error)`. If duplicate found → return `ErrDuplicateInvoice` (a retryable=false sentinel). Write `"ap.duplicate_check.is_duplicate": false` to opCtx.Data on pass.

**Files**
- `internal/ap/pipeline_stages.go` (additions)
- `internal/ap/repositories.go` (interface)

**Expected Outcome**
- No matching posted invoice → stage completes, `opCtx.Data["ap.duplicate_check.is_duplicate"] = false`.
- Matching posted invoice found → returns `ErrDuplicateInvoice`.
- Repository error → propagated as-is.

**How to Test**
Add to `internal/ap/pipeline_stages_test.go`:
```
TestDuplicateCheckStage_NoDuplicate_Completes
TestDuplicateCheckStage_DuplicateFound_Errors
TestDuplicateCheckStage_RepoError_Propagated
TestDuplicateCheckStage_WritesDataKey
```

---

### S-003 · ap.resolve_vendor

**Description**
Add `ResolveVendorStage` to `internal/ap/pipeline_stages.go` (Priority:210, Required:true). Calls `VendorRepository.GetByID(ctx, tenantID, vendorID)`. Writes `"ap.resolve_vendor.vendor_name"`, `"ap.resolve_vendor.payment_terms"`, `"ap.resolve_vendor.currency"`, `"ap.resolve_vendor.is_approved"` to opCtx.Data. If vendor not found → `ErrVendorNotFound`. If vendor is `"suspended"` → `ErrVendorSuspended`.

**Files**
- `internal/ap/pipeline_stages.go` (additions)

**Expected Outcome**
- Active vendor found → stage completes, all four data keys set.
- Vendor not found → `ErrVendorNotFound`.
- Suspended vendor → `ErrVendorSuspended`.

**How to Test**
```
TestResolveVendorStage_Found_WritesData
TestResolveVendorStage_NotFound_Error
TestResolveVendorStage_Suspended_Error
TestResolveVendorStage_Required_True
```

---

### S-004 · ap.resolve_gl_accounts

**Description**
Add `ResolveGLAccountsStage` to `internal/ap/pipeline_stages.go` (Priority:220, Required:true). Resolves GL accounts for each invoice line via `GLAccountRepository.ResolveForInvoice(ctx, tenantID, invoice)`. Writes `"ap.resolve_gl_accounts.lines"` (slice of resolved line+account pairs) and `"ap.resolve_gl_accounts.cost_center_id"` to opCtx.Data. Error if any line has no matching GL account.

**Files**
- `internal/ap/pipeline_stages.go` (additions)

**Expected Outcome**
- All lines resolved → completes, data keys written.
- Unresolvable GL account → returns descriptive error.

**How to Test**
```
TestResolveGLAccountsStage_AllResolved_Completes
TestResolveGLAccountsStage_UnresolvableAccount_Error
TestResolveGLAccountsStage_WritesDataKeys
```

---

### S-005 · budget.check

**Description**
Create `internal/budget/pipeline_stages.go`. Define `BudgetCheckStage` exactly as in the spec (Priority:310, Required:false, FeatureFlag:`"budget"`, `DependsOn:["ap.resolve_gl_accounts"]`). `RunCondition`: `toNumber(input.total_amount) > toNumber(setting["budget.minimum_check_amount"] ?? "0")`. Three control modes via `opCtx.SettingString("budget.control_mode", "warn")`:
- `"hard_block"` → return error
- `"soft_block"` → `opCtx.Suspend(...)`, return suspended result
- `"warn"` → complete with warning message

Write `"budget_check.approved"`, `"budget_check.available_amount"`, `"budget_check.utilisation_pct"` to Data. Set flag `"budget_exceeded"`. Implement `Simulate()` — calls budget service but returns `"simulated"` status without writing any state changes.

**Files**
- `internal/budget/pipeline_stages.go`

**Expected Outcome**
- Within budget → completes, `budget_exceeded = false`.
- Exceeded + `hard_block` → returns error with amounts in message.
- Exceeded + `soft_block` → `opCtx.Suspended = true`, `ResumePoint = "budget.check"`.
- Exceeded + `warn` → completes with warning message, `budget_exceeded = true`.
- `RunCondition` false (low amount) → stage is skipped by pipeline.
- `Simulate()` → `Status:"simulated"`, `budget_check.simulated: true` in outputs.

**How to Test**
Write `internal/budget/pipeline_stages_test.go`:
```
TestBudgetCheckStage_WithinBudget_Completes
TestBudgetCheckStage_Exceeded_HardBlock_Error
TestBudgetCheckStage_Exceeded_SoftBlock_Suspends
TestBudgetCheckStage_Exceeded_Warn_Completes
TestBudgetCheckStage_FlagSet_BudgetExceeded
TestBudgetCheckStage_Simulate_ReturnsMocked
TestBudgetCheckStage_FeatureFlag_Budget
TestBudgetCheckStage_Priority_310
TestBudgetCheckStage_Required_False
```

---

### S-006 · tax.resolve_codes

**Description**
Create `internal/tax/pipeline_stages.go`. Define `ResolveTaxCodesStage` (Priority:400, Required:false, FeatureFlag:`"tax"`). Calls `TaxService.ResolveCodesForInvoice(ctx, tenantID, invoice)`. Writes `"tax_calculation.tax_codes"` (slice) and `"tax_calculation.tax_regime"` to opCtx.Data.

**Files**
- `internal/tax/pipeline_stages.go`

**Expected Outcome**
- Tax codes resolved → completes, data keys written.
- Tax service error → propagated.

**How to Test**
```
TestResolveTaxCodesStage_Resolves_WritesData
TestResolveTaxCodesStage_ServiceError_Propagated
TestResolveTaxCodesStage_FeatureFlag_Tax
```

---

### S-007 · tax.calculate

**Description**
Add `CalculateTaxStage` to `internal/tax/pipeline_stages.go` (Priority:410, Required:false, FeatureFlag:`"tax"`, `DependsOn:["tax.resolve_codes"]`). Reads `"tax_calculation.tax_codes"` from opCtx.Data. Calls `TaxService.Calculate(ctx, params)`. Writes `"tax_calculation.tax_amount"`, `"tax_calculation.tax_lines"` to Data. Implement `Simulate()`.

**Files**
- `internal/tax/pipeline_stages.go` (additions)

**Expected Outcome**
- Codes present → calculates, writes tax_amount.
- Missing `tax_calculation.tax_codes` key → returns error `"tax codes not resolved"`.
- `Simulate()` → `Status:"simulated"`, outputs contain `expected_tax_amount`.

**How to Test**
```
TestCalculateTaxStage_Calculates_WritesAmount
TestCalculateTaxStage_MissingCodesKey_Error
TestCalculateTaxStage_DependsOn_ResolveCodes
TestCalculateTaxStage_Simulate_SimulatedStatus
```

---

### S-008 · tax.apply_withholding

**Description**
Add `WithholdingTaxStage` to `internal/tax/pipeline_stages.go` (Priority:430, Required:false, FeatureFlag:`"tax.withholding"`). Reads `"tax_calculation.tax_lines"`. Calls `TaxService.ApplyWithholding(ctx, params)`. Writes `"tax_calculation.withholding_amount"` to Data.

**Files**
- `internal/tax/pipeline_stages.go` (additions)

**Expected Outcome**
- Applied → completes, `withholding_amount` written.
- FeatureFlag `"tax.withholding"` gates the stage.

**How to Test**
```
TestWithholdingTaxStage_Applies_WritesAmount
TestWithholdingTaxStage_FeatureFlag_TaxWithholding
```

---

### S-009 · gl.post_transaction + TxHooks

**Description**
Create `internal/finance/pipeline_stages.go`. Define `GLPostingStage` (Priority:610, Required:true, Operations:`["*"]` financial operations). `Execute()`:
1. Calls `GLService.PostWithTransaction(ctx, params)` → returns `txResult` (TransactionID + open `pgx.Tx`)
2. Writes `"gl_posting.transaction_id"` and `"gl_posting.tx"` to Data
3. Registers TxHook `"gl.seed_domain_event"` (priority:10) — INSERTs into `domain_events` outbox
4. If `FeatureEnabled("budget")`: registers TxHook `"gl.update_budget_actuals"` (priority:20)
5. If `FeatureEnabled("workflow")`: registers TxHook `"gl.seed_workflow_trigger"` (priority:30) — INSERTs into `workflow_trigger_queue`

Compensation function (registered in wire.go, not in stage): call `GLService.ReverseTransaction()`.

**Files**
- `internal/finance/pipeline_stages.go`

**Expected Outcome**
- After `Execute()`: `"gl_posting.transaction_id"` is set in Data.
- `opCtx.TxHooks` length is 1 (base), 2 (+ budget), or 3 (+ budget + workflow) depending on flags.
- TxHook `"gl.seed_domain_event"` is always registered (no feature flag guard).
- `"gl.update_budget_actuals"` only registered when `budget` flag is on.
- `"gl.seed_workflow_trigger"` only registered when `workflow` flag is on.

**How to Test**
Write `internal/finance/pipeline_stages_test.go`:
```
TestGLPostingStage_RegistersTxHook_DomainEvent_Always
TestGLPostingStage_RegistersTxHook_BudgetActuals_WhenFlagOn
TestGLPostingStage_RegistersTxHook_BudgetActuals_NotWhenFlagOff
TestGLPostingStage_RegistersTxHook_WorkflowTrigger_WhenFlagOn
TestGLPostingStage_RegistersTxHook_WorkflowTrigger_NotWhenFlagOff
TestGLPostingStage_SetsTransactionIDInData
TestGLPostingStage_ServiceError_Propagated
TestGLPostingStage_Required_True
```

---

### S-010 · TxHook Atomic Execution (Pipeline.runTxHooks)

**Description**
Add `runTxHooks(opCtx *OperationContext) error` to `pipeline.go`. Reads `"gl_posting.tx"` from Data as `pgx.Tx`. If missing, returns nil (no open transaction — skip gracefully). Sorts hooks by Priority. Executes each in order; on any error → `tx.Rollback()` → return wrapped error. On all success → `tx.Commit()`. Wire this into the end of `Execute()` when `!DryRun && !Suspended && len(TxHooks) > 0`.

**Files**
- `internal/pipeline/pipeline.go` (additions)

**Expected Outcome**
- All TxHooks succeed → transaction committed.
- First TxHook fails → transaction rolled back, error returned, remaining hooks not executed.
- Second TxHook fails → transaction rolled back, error returned.
- No `"gl_posting.tx"` key in Data → `runTxHooks` returns nil silently.

**How to Test**
Add to `pipeline_test.go`:
```
TestRunTxHooks_AllSucceed_Commits
TestRunTxHooks_FirstFails_Rollback_NoFurtherHooks
TestRunTxHooks_SecondFails_Rollback
TestRunTxHooks_NoTxInData_ReturnsNil
TestRunTxHooks_SortedByPriority
TestRunTxHooks_SkippedInDryRun
TestRunTxHooks_SkippedWhenSuspended
```

---

### S-011 · ar.apply_advance

**Description**
Add `ApplyAdvancePaymentStage` to `internal/ar/pipeline_stages.go` (Priority:620, Required:false, FeatureFlag:`"ar.advances"`, Operations:`["ap.invoice.process"]`). Checks if vendor has any open advance payments via `ARService.GetOpenAdvances(ctx, tenantID, vendorID)`. If found, applies the advance against the invoice. Writes `"ar.apply_advance.applied_amount"` and `"ar.apply_advance.advance_id"` to Data. Compensation: `ARService.ReverseAdvanceApplication()`.

**Files**
- `internal/ar/pipeline_stages.go`

**Expected Outcome**
- No open advances → completes, no data written (or zero amount written).
- Advance found → completes, `applied_amount` written.

**How to Test**
```
TestApplyAdvanceStage_NoAdvances_Completes
TestApplyAdvanceStage_AdvanceApplied_WritesAmount
TestApplyAdvanceStage_FeatureFlag_ARAdvances
```

---

### S-012 · dms.archive + banking.schedule_payment

**Description**
Create `internal/dms/pipeline_stages.go` — `ArchiveDocumentStage` (Priority:710, Required:false, FeatureFlag:`"dms"`, Operations:`["*"]`). Calls `DMSService.Archive(ctx, params)`. Writes `"dms_archive.document_id"` to Data. Compensation: `DMSService.VoidDocument()`.

Create `internal/banking/pipeline_stages.go` — `SchedulePaymentStage` (Priority:720, Required:false, FeatureFlag:`"banking"`, Operations:`["ap.invoice.process"]`). Calls `BankingService.SchedulePayment(ctx, params)`. Writes `"payment_scheduling.payment_id"` and `"payment_scheduling.payment_date"` to Data. Compensation: `BankingService.CancelScheduledPayment()`.

**Files**
- `internal/dms/pipeline_stages.go`
- `internal/banking/pipeline_stages.go`

**Expected Outcome**
- Both stages complete when services succeed, data keys written.
- Both feature-flagged correctly.

**How to Test**
```
TestArchiveDocumentStage_Archives_WritesDocID
TestArchiveDocumentStage_FeatureFlag_DMS
TestSchedulePaymentStage_Schedules_WritesPaymentID
TestSchedulePaymentStage_FeatureFlag_Banking
```

---

### S-013 · audit.log_operation

**Description**
Create `internal/audit/pipeline_stages.go` — `AuditLogStage` (Priority:910, Required:true, Operations:`["*"]`, no FeatureFlag). Calls `AuditService.LogOperation(ctx, params)` with full opCtx summary: operationID, tenantID, userID, operationKey, status, stage count, duration. This stage runs last and always.

**Files**
- `internal/audit/pipeline_stages.go`

**Expected Outcome**
- Completes, `AuditService.LogOperation` called with correct params.
- Required:true so failure aborts (audit is mandatory).
- No FeatureFlag — always included in every pipeline.

**How to Test**
```
TestAuditLogStage_LogsOperation
TestAuditLogStage_Required_True
TestAuditLogStage_NoFeatureFlag
TestAuditLogStage_Priority_910
```

---

## Phase 4 — Hook Implementations

### H-001 · audit.log_hook

**Description**
Create `internal/audit/pipeline_hooks.go` — `AuditLogHook`. `StageAfter()=["*"]`, `Operations()=["*"]`, no FeatureFlag. Fires after every stage on every operation. `Execute()` calls `AuditService.LogStageEvent(ctx, stageName, status, duration, tenantID)`.

**Files**
- `internal/audit/pipeline_hooks.go`

**Expected Outcome**
- Fires after every stage regardless of module or feature flags.
- `AuditService.LogStageEvent` called with correct stage name and status.

**How to Test**
```
TestAuditLogHook_StageAfterWildcard
TestAuditLogHook_OperationsWildcard
TestAuditLogHook_NoFeatureFlag
TestAuditLogHook_Execute_CallsAuditService
```

---

### H-002 · notify.posting_complete

**Description**
Create `internal/notify/pipeline_hooks.go` — `PostingNotificationHook`. `StageAfter()=["banking.schedule_payment"]`, FeatureFlag:`"notifications"`. `Execute()` calls `NotificationService.SendPostingComplete(ctx, params)` with invoice and payment details from opCtx.Data.

**Files**
- `internal/notify/pipeline_hooks.go`

**Expected Outcome**
- Fires after `banking.schedule_payment`.
- Gated by `notifications` feature flag.
- `NotificationService.SendPostingComplete` called.

**How to Test**
```
TestNotificationHook_FiresAfterBankingStage
TestNotificationHook_FeatureFlag_Notifications
TestNotificationHook_Execute_CallsNotifyService
```

---

### H-003 · tax.e_invoice_submit

**Description**
Add `EInvoiceHook` to `internal/tax/pipeline_hooks.go`. `StageAfter()=["gl.post_transaction"]`, FeatureFlag:`"tax.e_invoice"`, Operations:`["ar.invoice.*"]`. `Execute()` calls `TaxService.SubmitEInvoice(ctx, params)` with GL transaction ID from opCtx.Data. Writes `"e_invoice.submission_id"` to Data.

**Files**
- `internal/tax/pipeline_hooks.go`

**Expected Outcome**
- Only fires after `gl.post_transaction` on AR invoice operations.
- Gated by `tax.e_invoice` flag.
- Submission ID written to opCtx.Data.

**How to Test**
```
TestEInvoiceHook_FiresAfterGLPost
TestEInvoiceHook_FeatureFlag_TaxEInvoice
TestEInvoiceHook_WritesSubmissionID
```

---

### H-004 · compliance.pre_posting + compliance.sanctions_check (stage)

**Description**
Create `internal/compliance/pipeline_stages.go` — `SanctionsCheckStage` (Priority:390, Required:false, FeatureFlag:`"compliance"`, `DependsOn:["ap.resolve_vendor"]`). Calls `ComplianceService.CheckSanctions(ctx, vendorID)`. Sets flag `"sanctions_hit"` in opCtx. If hit found → returns error.

Create `internal/compliance/pipeline_hooks.go` — `PrePostingCheckHook`. `StageBefore()=["gl.post_transaction"]`, FeatureFlag:`"compliance"`. Reads `"sanctions_hit"` flag; if true → returns error blocking GL posting.

**Files**
- `internal/compliance/pipeline_stages.go`
- `internal/compliance/pipeline_hooks.go`

**Expected Outcome**
- Clean vendor → stage completes, `sanctions_hit=false`.
- Sanctioned vendor → stage returns error.
- `PrePostingCheckHook` blocks GL posting if `sanctions_hit=true`.

**How to Test**
```
TestSanctionsCheckStage_Clean_Completes
TestSanctionsCheckStage_Hit_Error
TestSanctionsCheckStage_SetsSanctionsHitFlag
TestPrePostingCheckHook_SanctionsHit_BlocksGL
TestPrePostingCheckHook_NoHit_Passes
```

---

### H-005 · workflow.approval_gate

**Description**
Create `internal/workflow/pipeline_hook.go` — `WorkflowApprovalHook`. `StageBefore()=["gl.post_transaction"]`, FeatureFlag:`"workflow"`, Operations:`["*"]`. `RunCondition`:
```
toNumber(input.total_amount) > toNumber(setting["workflow.approval_threshold_amount"] ?? "0") || budget_exceeded == true
```
`Execute()`:
1. `workflowEngine.GetApprovalRule(ctx, ...)` — if no approval required, return nil
2. `workflowEngine.StartWorkflow(ctx, templateID, payload)` → get instance
3. Write `"workflow.instance_id"` and `"workflow.approval_id"` to Data
4. `opCtx.Suspend("pending_workflow_approval:"+instanceID, "gl.post_transaction")`

`SimulateExecute()`: calls `GetApprovalRule`, writes `"workflow.approval_required_in_real_run"` and `"workflow.expected_approvers"` to Data — does NOT call `Suspend()`.

**Files**
- `internal/workflow/pipeline_hook.go`

**Expected Outcome**
- Amount above threshold → workflow started, `opCtx.Suspended = true`.
- Amount below threshold AND `budget_exceeded=false` → hook fires but no suspension.
- `budget_exceeded=true` overrides amount threshold → always suspends.
- `SimulateExecute()` sets approval metadata but does NOT suspend.
- In dry-run: full pipeline continues past this hook.

**How to Test**
Write `internal/workflow/pipeline_hook_test.go`:
```
TestWorkflowApprovalHook_AboveThreshold_Suspends
TestWorkflowApprovalHook_BelowThreshold_NoSuspend
TestWorkflowApprovalHook_BudgetExceeded_Suspends
TestWorkflowApprovalHook_NoApprovalRequired_NoSuspend
TestWorkflowApprovalHook_SimulateExecute_NoSuspend
TestWorkflowApprovalHook_SimulateExecute_WritesApprovalMetadata
TestWorkflowApprovalHook_SetsWorkflowDataKeys
```

---

## Phase 5 — Workflow Engine

### W-001 · DB Migrations

**Description**
Create migration files in `db/migration/`:
- `001_XXX_create_operation_logs.sql` — `operation_logs` table with all columns, indexes, RLS policy
- `001_XXX_create_workflow_tables.sql` — `workflow_templates`, `workflow_instances`, `workflow_step_executions`, `workflow_user_tasks`, `workflow_trigger_queue` tables, all indexes, all RLS policies

Use exact schemas from the spec. Add `DOWN` migration for each.

**Files**
- `db/migration/XXXXXX_create_operation_logs.up.sql`
- `db/migration/XXXXXX_create_operation_logs.down.sql`
- `db/migration/XXXXXX_create_workflow_tables.up.sql`
- `db/migration/XXXXXX_create_workflow_tables.down.sql`

**Expected Outcome**
- `migrate up` creates all tables with correct column types, constraints, and indexes.
- `migrate down` drops them cleanly.
- RLS policies use `current_setting('app.tenant_id')::uuid` correctly.
- Unique index on `operation_logs(tenant_id, idempotency_key) WHERE idempotency_key IS NOT NULL`.
- Partial index on `operation_logs(tenant_id) WHERE status = 'pending_approval'`.

**How to Test**
Run migrations against a test database:
```
psql: migrate up → all tables and indexes created (check with \dt and \di)
psql: migrate down → all tables dropped cleanly
psql: verify RLS enabled on each workflow table
psql: verify indexes on workflow_trigger_queue(tenant_id, created_at) WHERE processed_at IS NULL
```

---

### W-002 · OperationLog Repository

**Description**
Create `internal/pipeline/log_repo_pg.go` — concrete PostgreSQL implementation of `OperationLogRepository`:
- `SaveLog(ctx, *OperationResult) error` — upsert on `operation_id`
- `GetByOperationID(ctx, uuid.UUID) (*OperationResult, error)`
- `GetByIdempotencyKey(ctx, tenantID, key string) (*OperationResult, error)`
- `SetIdempotencyKey(ctx, operationID uuid.UUID, key string) error`
- `MarkRejected(ctx, operationID uuid.UUID, reason string) error`

All queries use tenant isolation from `pgx` connection with `app.tenant_id` setting.

**Files**
- `internal/pipeline/log_repo_pg.go`

**Expected Outcome**
- `SaveLog` followed by `GetByOperationID` returns same data.
- `GetByIdempotencyKey` returns error when key doesn't exist.
- `MarkRejected` sets `status="failed"` and `error_message`.
- All queries respect tenant isolation.

**How to Test**
Write `internal/pipeline/log_repo_pg_test.go` (requires test DB):
```
TestOperationLogRepo_SaveAndGet
TestOperationLogRepo_GetByIdempotencyKey_Found
TestOperationLogRepo_GetByIdempotencyKey_NotFound
TestOperationLogRepo_MarkRejected
TestOperationLogRepo_TenantIsolation
```

---

### W-003 · Workflow Template + Instance Repositories

**Description**
Create `internal/workflow/repositories.go`. Define interfaces:
- `WorkflowTemplateRepository`: `GetByID`, `ListActive(tenantID)`, `Create`, `Update`
- `WorkflowInstanceRepository`: `Create`, `GetByID`, `UpdateStatus`, `CompleteInstance`, `GetByOperationID`
- `UserTaskRepository`: `Create`, `GetByID`, `ListForUser(userID)`, `Complete(taskID, decision, comment, formData)`, `Escalate(taskID, escalatedTo)`
- `TriggerQueueRepository`: `GetUnprocessed(ctx, limit) []TriggerQueueRow`, `MarkProcessed(ctx, id)`, `IncrementErrorCount(ctx, id, errMsg)`

Create `internal/workflow/log_repo_pg.go` — PostgreSQL implementations of all four.

**Files**
- `internal/workflow/repositories.go`
- `internal/workflow/repo_pg.go`

**Expected Outcome**
- Template CRUD round-trips correctly.
- `ListForUser(userID)` returns only tasks assigned to that user that are `"pending"`.
- `MarkProcessed` sets `processed_at`.

**How to Test**
Write `internal/workflow/repo_pg_test.go` (requires test DB):
```
TestWorkflowTemplateRepo_CreateAndGet
TestWorkflowInstanceRepo_CreateAndGetByID
TestUserTaskRepo_CreateAndListForUser
TestUserTaskRepo_Complete_SetsDecision
TestTriggerQueueRepo_GetUnprocessed
TestTriggerQueueRepo_MarkProcessed
```

---

### W-004 · Workflow Engine Service

**Description**
Create `internal/workflow/engine.go`. Define `Engine` struct. Implement:
- `StartWorkflow(ctx, templateID uuid.UUID, input map[string]any) (*WorkflowInstance, error)` — creates `workflow_instances` row, starts Temporal workflow via `temporalClient.ExecuteWorkflow()`
- `GetApprovalRule(ctx, RuleQuery) (*ApprovalConfig, error)` — checks `workflow_triggers` table for matching active triggers, evaluates `trigger_conditions` via `pkg/condition`, returns `ApprovalConfig{RequiresApproval bool, Approvers []string, ApprovalDeadline time.Duration}`

Define `ApprovalConfig`, `RuleQuery` structs.

**Files**
- `internal/workflow/engine.go`
- `internal/workflow/types.go`

**Expected Outcome**
- `StartWorkflow` creates DB row and returns instance with ID.
- `GetApprovalRule` returns `RequiresApproval=false` when no matching trigger exists.
- `GetApprovalRule` returns `RequiresApproval=true` with approvers when trigger condition matches.
- `StartWorkflow` with Temporal client unavailable returns error.

**How to Test**
Write `internal/workflow/engine_test.go` (mock Temporal):
```
TestWorkflowEngine_StartWorkflow_CreatesInstance
TestWorkflowEngine_StartWorkflow_StartsTemporalWorkflow
TestWorkflowEngine_GetApprovalRule_NoTrigger_NoApproval
TestWorkflowEngine_GetApprovalRule_TriggerMatches_RequiresApproval
TestWorkflowEngine_GetApprovalRule_TriggerConditionFalse_NoApproval
```

---

### W-005 · Temporal Workflow Function

**Description**
Create `internal/workflow/temporal_workflow.go`. Implement `CustomWorkflowExecution(ctx workflow.Context, definition WorkflowDefinition, instanceID uuid.UUID, input map[string]any) error`. The workflow iterates steps:
- `"validation"` step: checks rules against input fields
- `"condition"` step: evaluates cases in order, routes to matching `next_step`
- `"user_task"` step: calls `executeUserTaskStep()` — creates task row, sends notification activity, waits for Temporal signal `"task-{id}-completed"`, handles timeout → escalation
- `"notification"` step: sends notification via activity
- `"end"` step: completes

On workflow completion: if `pipeline_operation_id` is in input → call `CompleteWorkflowInstanceActivity` → triggers `PipelineResumeService.Resume()`.

**Files**
- `internal/workflow/temporal_workflow.go`
- `internal/workflow/temporal_activities.go`

**Expected Outcome**
- Validation step passes → routes to next step.
- Validation step fails → routes to `on_failure` step.
- Condition step evaluates cases in order, picks first matching.
- User task step waits for signal; signal with `decision="approved"` → `on_approved` path.
- User task timeout → task escalated.
- Workflow completion with `operation_id` in input → `PipelineResumeService.Resume()` called.

**How to Test**
Write `internal/workflow/temporal_workflow_test.go` (Temporal test server):
```
TestWorkflow_ValidationStep_Pass_RoutesToNext
TestWorkflow_ValidationStep_Fail_RoutesToOnFailure
TestWorkflow_ConditionStep_RoutesCorrectly
TestWorkflow_UserTaskStep_Approved_RoutesToOnApproved
TestWorkflow_UserTaskStep_Rejected_RoutesToOnRejected
TestWorkflow_UserTaskStep_Timeout_Escalates
TestWorkflow_OnCompletion_ResumesPipeline
```

---

### W-006 · Trigger Listener

**Description**
Create `internal/workflow/trigger_listener.go`. Define `TriggerListener` with `triggerRepo`, `instanceRepo`, `workflowEngine`, `evaluator`. Implement `Run(ctx)` — polls `workflow_trigger_queue` every 2 seconds. `processQueue(ctx)`:
1. `GetUnprocessed(ctx, limit=50)`
2. For each row: `triggerRepo.FindByEvent(ctx, tenantID, eventName)`
3. For each matching trigger: evaluate `trigger_conditions` via `pkg/condition`
4. If passes: `workflowEngine.StartWorkflow(ctx, trigger.TemplateID, row.Payload)`
5. `MarkProcessed(ctx, row.ID)` on success; `IncrementErrorCount` on failure

**Files**
- `internal/workflow/trigger_listener.go`

**Expected Outcome**
- Unprocessed rows are picked up and matching workflows started.
- Rows with `trigger_conditions` that evaluate false are skipped (not started).
- Rows that fail to start workflow have `error_count` incremented, not marked processed.
- `Run()` respects context cancellation.

**How to Test**
Write `internal/workflow/trigger_listener_test.go`:
```
TestTriggerListener_PicksUpUnprocessed
TestTriggerListener_ConditionFalse_SkipsWorkflow
TestTriggerListener_StartWorkflowError_IncrementsErrorCount
TestTriggerListener_MarksProcessedOnSuccess
TestTriggerListener_ContextCancellation_Stops
```

---

### W-007 · My Tasks API Handlers

**Description**
Create `internal/workflow/handlers/tasks.go`. Implement:
- `GET /api/workflows/tasks/my-tasks` — lists `workflow_user_tasks` for `session.UserID` with status `"pending"` or `"in_progress"`, sorted by priority then due_at
- `GET /api/workflows/tasks/:id` — returns full task with `context_data`, `form_schema`, `action_buttons`
- `POST /api/workflows/tasks/:id/complete` — validates ownership, sets decision/comment/formData, sends Temporal signal `"task-{id}-completed"`, returns 200

**Files**
- `internal/workflow/handlers/tasks.go`

**Expected Outcome**
- `my-tasks` returns only tasks assigned to the logged-in user.
- `complete` with wrong user → 403.
- `complete` with valid decision → task status set to `"completed"`, Temporal signal sent.
- Overdue tasks still returned in list (not filtered out).

**How to Test**
Write `internal/workflow/handlers/tasks_test.go`:
```
TestMyTasksHandler_ReturnsOnlyAssignedTasks
TestGetTask_ReturnsFullTaskWithContext
TestCompleteTask_ValidDecision_Succeeds
TestCompleteTask_WrongUser_403
TestCompleteTask_SendsTemporalSignal
TestCompleteTask_AlreadyCompleted_409
```

---

## Phase 6 — API Layer

### A-001 · ProcessInvoice Handler (Sync/Async Split)

**Description**
Create `internal/ap/handlers/process.go`. Implement `ProcessInvoiceHandler(deps)` for `POST /api/ap/invoices/:id/process`. Uses 8-second sync deadline:
- Start pipeline in goroutine
- `select` on result/error/timeout channels
- On timeout → return `202 {status:"running", operation_id, poll_url, poll_interval_ms:2000}`
- On result → `renderPipelineResult(c, result)` — 200 completed, 202 pending_approval, 422 failed
- On error → `renderPipelineError(c, err)` — map pipeline errors to HTTP status codes

Reads `Idempotency-Key` header, passes to `ExecuteIdempotent`. Reads `?dry_run=true` query param.

**Files**
- `internal/ap/handlers/process.go`
- `internal/ap/handlers/render.go`

**Expected Outcome**
- Fast pipeline (< 8s) → 200 with result synchronously.
- Slow pipeline (> 8s) → 202 with `operation_id` for polling.
- Suspended pipeline → 202 with `status: "pending_approval"`.
- Failed pipeline → 422 with structured error including `stage`, `code`, `user_action_required`.
- Dry-run → `dry_run: true` in response.

**How to Test**
Write `internal/ap/handlers/process_test.go`:
```
TestProcessInvoiceHandler_FastPipeline_Returns200
TestProcessInvoiceHandler_SlowPipeline_Returns202WithPollURL
TestProcessInvoiceHandler_PendingApproval_Returns202
TestProcessInvoiceHandler_Failed_Returns422WithStructuredError
TestProcessInvoiceHandler_DryRun_ReturnsSimulatedResult
TestProcessInvoiceHandler_IdempotencyKey_ReturnsExistingResult
```

---

### A-002 · Operation Status + WebSocket Live Events

**Description**
Create `internal/pipeline/handlers/status.go`:
- `GET /api/operations/:id/status` — reads `operation_logs` by operationID, returns `{status, stage_log, output_data, error}`. 404 if not found or wrong tenant.
- `GET /api/operations/:id/live` — WebSocket upgrade; subscribes to `EventBus` for `StageCompleted` and `TxHookFailed` events on this operationID; sends each as JSON frame; closes when operation completes.

**Files**
- `internal/pipeline/handlers/status.go`

**Expected Outcome**
- Status endpoint returns correct status for running/completed/failed operations.
- Status for a different tenant's operation → 404 (not 403).
- WebSocket receives `StageCompleted` events in real-time during execution.
- WebSocket closes when pipeline reaches terminal status.

**How to Test**
Write `internal/pipeline/handlers/status_test.go`:
```
TestStatusHandler_CompletedOperation_Returns200
TestStatusHandler_NotFound_Returns404
TestStatusHandler_WrongTenant_Returns404
TestStatusHandler_PendingApproval_Returns202Body
TestWebSocketLive_ReceivesStageEvents
TestWebSocketLive_ClosesOnCompletion
```

---

### A-003 · Dry-Run Endpoint

**Description**
Create `internal/pipeline/handlers/dry_run.go`. Implement `DryRunHandler(deps)` for `POST /api/operations/dry-run`. Parses `{operation_key, input}`, sets `opCtx.DryRun = true`, runs pipeline, returns:
```json
{
  "dry_run": true,
  "would_succeed": bool,
  "would_require_approval": bool,
  "expected_approvers": [],
  "stages_that_would_run": [],
  "stages_that_would_skip": [],
  "expected_gl_entries": [],
  "expected_payment_date": "",
  "expected_tax_amount": "",
  "stage_log": []
}
```

**Files**
- `internal/pipeline/handlers/dry_run.go`

**Expected Outcome**
- `would_succeed` is true when no stage returns error in simulate mode.
- `would_require_approval` matches `opCtx.Data["workflow.approval_required_in_real_run"]`.
- `stages_that_would_skip` lists stages whose `RunCondition` was false.
- `expected_tax_amount` comes from `"tax_calculation.tax_amount"` in opCtx.Data.
- Nothing written to database.

**How to Test**
```
TestDryRunHandler_ReturnsSimulatedResult
TestDryRunHandler_WouldRequireApproval
TestDryRunHandler_NothingWrittenToDB
TestDryRunHandler_SkippedStagesListed
```

---

### A-004 · Internal Pipeline Resume Endpoint

**Description**
Create `internal/pipeline/handlers/resume.go`. Implement `POST /internal/pipeline/resume/:operation_id` — internal endpoint (not exposed to clients; only callable from Temporal activities). Calls `PipelineResumeService.Resume(ctx, operationID, body)`. Returns 200 on success. Secured with internal service token (header `X-Internal-Token`).

**Files**
- `internal/pipeline/handlers/resume.go`

**Expected Outcome**
- Valid token + pending operation → resumes pipeline, 200.
- Invalid token → 401.
- Already-completed operation → 409.

**How to Test**
```
TestResumeHandler_ValidToken_Resumes
TestResumeHandler_InvalidToken_401
TestResumeHandler_AlreadyCompleted_409
```

---

## Phase 7 — Wire + Integration

### I-001 · wire.go — Full Registration

**Description**
Update `cmd/server/wire.go` (or create it). Register all stages, hooks, and compensation functions:

**Stages (in order):**
`ValidateInvoiceStage`, `DuplicateCheckStage`, `ResolveVendorStage`, `ResolveGLAccountsStage`, `ThreeWayMatchStage` (procurement — stub if not yet impl), `BudgetCheckStage`, `SanctionsCheckStage`, `ResolveTaxCodesStage`, `CalculateTaxStage`, `WithholdingTaxStage`, `GLPostingStage`, `ApplyAdvancePaymentStage`, `ArchiveDocumentStage`, `SchedulePaymentStage`, `AuditLogStage`

**Hooks:**
`WorkflowApprovalHook`, `AuditLogHook`, `PostingNotificationHook`, `EInvoiceHook`, `PrePostingCheckHook`

**Compensation functions:**
- `"gl.post_transaction"` → `GLService.ReverseTransaction`
- `"banking.schedule_payment"` → `BankingService.CancelScheduledPayment`
- `"dms.archive"` → `DMSService.VoidDocument`
- `"ar.apply_advance"` → `ARService.ReverseAdvanceApplication`

Register all document schemas on `SchemaRegistry`.

**Files**
- `cmd/server/wire.go`

**Expected Outcome**
- Server starts without panic.
- `PipelineBuilder.Build("ap.invoice.process", basicSession)` returns pipeline with 6 stages.
- `PipelineBuilder.Build("ap.invoice.process", enterpriseSession)` returns pipeline with 14 stages.
- All compensation functions are registered.

**How to Test**
Write `cmd/server/wire_test.go`:
```
TestWire_ServerStartsWithoutPanic
TestWire_BasicTenant_6Stages
TestWire_EnterpriseTenant_14Stages
TestWire_AllCompensationFnsRegistered
TestWire_AllSchemasRegistered
```

---

### I-002 · Integration Test — Basic Finance Tenant

**Description**
Write `internal/pipeline/integration/basic_tenant_test.go`. Uses a real PostgreSQL test database. Creates a minimal session with only core flags (`ap`, `finance`, `audit` enabled, all others off). Creates a test invoice. Runs `APService.ProcessInvoice()`. Verifies end-to-end.

**Files**
- `internal/pipeline/integration/basic_tenant_test.go`

**Expected Outcome**
- Pipeline runs 6 stages: validate → duplicate_check → resolve_vendor → resolve_gl_accounts → gl.post_transaction → audit.
- No budget, tax, procurement, workflow stages run.
- `operation_logs` row created with status `"completed"`.
- GL transaction row created in DB.
- `domain_events` outbox row seeded atomically by TxHook.
- Return value has `Status:"completed"` and non-nil `GLTxID`.

**How to Test**
The test IS the test:
```
TestBasicTenant_APInvoice_FullFlow_Completes
TestBasicTenant_APInvoice_OperationLogPersisted
TestBasicTenant_APInvoice_GLTransactionCreated
TestBasicTenant_APInvoice_DomainEventSeeded
TestBasicTenant_APInvoice_NoWorkflowStagesRun
```

---

### I-003 · Integration Test — Full Enterprise Tenant

**Description**
Write `internal/pipeline/integration/enterprise_tenant_test.go`. Uses real PostgreSQL test database + mock Temporal client. Creates session with all flags enabled. Tests:
1. Small invoice → no approval required → completes immediately
2. Large invoice → `WorkflowApprovalHook` fires → pipeline suspends → simulates Temporal signal → `PipelineResumeService.Resume()` → pipeline completes
3. Budget exceeded invoice → `BudgetCheckStage` flags it → hook suspends → same resume flow

**Files**
- `internal/pipeline/integration/enterprise_tenant_test.go`

**Expected Outcome**
- 14 stages run for full enterprise session.
- Large invoice: `operation_logs.status = "pending_approval"` after first execution.
- After resume: `operation_logs.status = "completed"`, GL transaction exists.
- `workflow_trigger_queue` row created atomically by TxHook and later processed by `TriggerListener`.
- Budget exceeded + `warn` mode: completes with warning in stage log.

**How to Test**
The test IS the test:
```
TestEnterpriseTenant_SmallInvoice_CompletesDirectly
TestEnterpriseTenant_LargeInvoice_SuspendsForApproval
TestEnterpriseTenant_LargeInvoice_Resume_Completes
TestEnterpriseTenant_BudgetExceeded_Warn_Completes
TestEnterpriseTenant_WorkflowTriggerQueueSeeded
TestEnterpriseTenant_AllStagesRunInCorrectOrder
TestEnterpriseTenant_CompensationOnGLFailure
```

---

## Task Summary

| Phase | Tasks | IDs |
|---|---|---|
| Core Pipeline | 12 | P-001 → P-012 |
| Condition/Scripting | 5 | C-001 → C-005 |
| Module Stages | 13 | S-001 → S-013 |
| Hooks | 5 | H-001 → H-005 |
| Workflow Engine | 7 | W-001 → W-007 |
| API Layer | 4 | A-001 → A-004 |
| Wire + Integration | 3 | I-001 → I-003 |
| **Total** | **49** | |

---

## Conventions

- **Naming**: stage files → `{module}/pipeline_stages.go`, hook files → `{module}/pipeline_hooks.go`
- **Data keys**: always `"{stage_name}.{key}"` — never bare keys
- **Error propagation**: repo errors propagate up; domain sentinel errors are defined in `{module}/errors.go`
- **No feature flag = always included**: stages with empty `FeatureFlag()` run for all tenants
- **Tests first**: write the test stubs from the "How to Test" section before implementing; all listed tests must pass before a task is checked off
- **Stage at completion**: after all tests pass, `git add` only the files listed in the task's **Files** section
