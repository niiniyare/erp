package pipeline

import (
	"context"
	"sync"
	"time"

	iamdomain "awo.so/internal/core/iam/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// StageLog records the outcome of a single stage execution within a pipeline run.
type StageLog struct {
	StageName string
	// Status mirrors StageResult.Status: "completed" | "skipped" | "suspended" | "failed" | "simulated"
	Status     string
	SkipReason string
	Message    string
	StartedAt  time.Time
	Duration   time.Duration
	Output     map[string]any
	Error      string
}

// OperationContext is the mutable carrier object threaded through every stage
// and hook in a pipeline execution. It is passed by pointer so stages can
// write to shared state.
//
// Data key convention: "{stage_name}.{key}" — e.g. "budget_check.available_amount".
// This avoids collisions between modules without a registry.
//
// Instances are pooled (see AcquireOperationContext / ReleaseOperationContext)
// to reduce GC pressure under high concurrency.
type OperationContext struct {
	// ── Identity ──────────────────────────────────────────────────────────
	Ctx      context.Context
	Session  *iamdomain.ResolvedSession
	TenantID uuid.UUID
	UserID   uuid.UUID
	EntityID uuid.UUID

	// ── Operation identity ────────────────────────────────────────────────
	OperationID  uuid.UUID
	OperationKey string // "ap.invoice.process"
	Resource     string // "ap.invoice"
	Action       string // "process"

	// ── Mode ─────────────────────────────────────────────────────────────
	// DryRun: when true stages simulate without writing to DB.
	DryRun bool

	// ── Input ────────────────────────────────────────────────────────────
	// Input is typed by the caller; stages cast to the expected concrete type.
	Input any

	// ── Shared mutable state ─────────────────────────────────────────────
	// Keys follow "{stage_name}.{key}" convention to avoid collisions.
	//   "tax_calculation.tax_amount"     → decimal.Decimal
	//   "budget_check.available_amount"  → decimal.Decimal
	//   "workflow.approval_id"           → uuid.UUID
	//   "gl_posting.transaction_id"      → uuid.UUID
	Data map[string]any

	// ── Decision flags ────────────────────────────────────────────────────
	// Boolean outcomes set by stages for downstream consumers.
	//   "requires_approval" | "tax_exempt" | "budget_exceeded"
	Flags map[string]bool

	// ── Suspension ────────────────────────────────────────────────────────
	Suspended     bool
	SuspendReason string // "pending_workflow_approval:<uuid>"
	ResumePoint   string // stage name to resume from after un-suspension

	// ── Execution log ─────────────────────────────────────────────────────
	Log         []StageLog
	StartedAt   time.Time
	CompletedAt *time.Time

	// ── DB transaction hooks ──────────────────────────────────────────────
	// Registered by stages during Execute(); fired atomically within the
	// same DB transaction after all stages complete.
	TxHooks []TxHook
}

// ── Shared state accessors ────────────────────────────────────────────────────

// GetData returns the value stored under key and whether it was present.
func (o *OperationContext) GetData(key string) (any, bool) {
	v, ok := o.Data[key]
	return v, ok
}

// SetData stores value under key in the shared data map.
func (o *OperationContext) SetData(key string, value any) {
	o.Data[key] = value
}

// SetFlag sets the named boolean decision flag.
func (o *OperationContext) SetFlag(key string, value bool) {
	o.Flags[key] = value
}

// Flag returns the value of the named boolean flag.
// Returns false when the key has not been set.
func (o *OperationContext) Flag(key string) bool {
	return o.Flags[key]
}

// ── Delegation helpers ────────────────────────────────────────────────────────

// FeatureEnabled returns whether the named feature flag is on for this tenant.
// Returns false when the session is nil.
func (o *OperationContext) FeatureEnabled(key string) bool {
	if o.Session == nil {
		return false
	}
	return o.Session.FeatureEnabled(key)
}

// SettingDecimal returns the named tenant setting as a decimal.Decimal.
// Returns def when the key is absent or cannot be parsed.
// The IAM session stores settings as strings parsed to float64; this method
// converts the float64 to decimal.Decimal to preserve the pipeline's numeric API.
func (o *OperationContext) SettingDecimal(key string, def decimal.Decimal) decimal.Decimal {
	if o.Session == nil {
		return def
	}
	f := o.Session.SettingDecimal(key, def.InexactFloat64())
	return decimal.NewFromFloat(f)
}

// SettingBool returns the named tenant setting as a bool.
// Returns def when the key is absent.
func (o *OperationContext) SettingBool(key string, def bool) bool {
	if o.Session == nil {
		return def
	}
	return o.Session.SettingBool(key, def)
}

// SettingString returns the named tenant setting as a string.
// Returns def when the key is absent.
func (o *OperationContext) SettingString(key string, def string) string {
	if o.Session == nil {
		return def
	}
	return o.Session.SettingString(key, def)
}

// ── Suspension ────────────────────────────────────────────────────────────────

// Suspend marks the operation as suspended.
// reason describes why (e.g. "pending_workflow_approval:<id>").
// resumeAt is the stage name the pipeline should restart from when resumed.
func (o *OperationContext) Suspend(reason, resumeAt string) {
	o.Suspended = true
	o.SuspendReason = reason
	o.ResumePoint = resumeAt
}

// ── TxHook registration ───────────────────────────────────────────────────────

// RegisterTxHook appends a hook to be executed within the open DB transaction
// after all pipeline stages complete.
func (o *OperationContext) RegisterTxHook(hook TxHook) {
	o.TxHooks = append(o.TxHooks, hook)
}

// ── Object pool ───────────────────────────────────────────────────────────────

var opCtxPool = sync.Pool{
	New: func() any {
		return &OperationContext{
			Data:    make(map[string]any, 32),
			Flags:   make(map[string]bool, 16),
			Log:     make([]StageLog, 0, 16),
			TxHooks: make([]TxHook, 0, 4),
		}
	},
}

// AcquireOperationContext retrieves a zeroed OperationContext from the pool.
func AcquireOperationContext() *OperationContext {
	return opCtxPool.Get().(*OperationContext)
}

// ReleaseOperationContext resets the context and returns it to the pool.
// Callers must not use the context after calling this.
func ReleaseOperationContext(o *OperationContext) {
	o.reset()
	opCtxPool.Put(o)
}

// reset clears all fields so a pooled instance is safe to reuse.
// Map allocations are retained to avoid unnecessary GC churn.
func (o *OperationContext) reset() {
	o.Ctx = nil
	o.Session = nil
	o.TenantID = uuid.Nil
	o.UserID = uuid.Nil
	o.EntityID = uuid.Nil
	o.OperationID = uuid.Nil
	o.OperationKey = ""
	o.Resource = ""
	o.Action = ""
	o.DryRun = false
	o.Input = nil

	for k := range o.Data {
		delete(o.Data, k)
	}
	for k := range o.Flags {
		delete(o.Flags, k)
	}

	o.Log = o.Log[:0]
	o.TxHooks = o.TxHooks[:0]

	o.Suspended = false
	o.SuspendReason = ""
	o.ResumePoint = ""
	o.StartedAt = time.Time{}
	o.CompletedAt = nil
}
