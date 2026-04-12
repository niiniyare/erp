package pipeline

import (
	"testing"
	"time"

	iamdomain "awo.so/internal/core/iam/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestSession(flags map[string]bool, settings map[string]string) *iamdomain.ResolvedSession {
	if flags == nil {
		flags = map[string]bool{}
	}
	if settings == nil {
		settings = map[string]string{}
	}
	return &iamdomain.ResolvedSession{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Permissions: map[string]bool{
			"ap.invoices.read":    true,
			"ap.invoices.create":  true,
			"ap.invoices.process": true,
		},
		Configuration: iamdomain.Configuration{
			Flags:    flags,
			Settings: settings,
		},
	}
}

func newTestOpCtx() *OperationContext {
	return &OperationContext{
		Data:    make(map[string]any, 8),
		Flags:   make(map[string]bool, 8),
		Log:     make([]StageLog, 0, 4),
		TxHooks: make([]TxHook, 0, 2),
		Session: newTestSession(nil, nil),
	}
}

// ── GetData / SetData ─────────────────────────────────────────────────────────

func TestOperationContext_SetAndGetData(t *testing.T) {
	o := newTestOpCtx()
	o.SetData("foo.bar", 42)

	v, ok := o.GetData("foo.bar")
	assert.True(t, ok)
	assert.Equal(t, 42, v)
}

func TestOperationContext_GetData_Missing(t *testing.T) {
	o := newTestOpCtx()

	v, ok := o.GetData("missing.key")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestOperationContext_SetData_OverwritesExisting(t *testing.T) {
	o := newTestOpCtx()
	o.SetData("k", "first")
	o.SetData("k", "second")

	v, ok := o.GetData("k")
	assert.True(t, ok)
	assert.Equal(t, "second", v)
}

// ── Flags ─────────────────────────────────────────────────────────────────────

func TestOperationContext_SetFlag(t *testing.T) {
	o := newTestOpCtx()
	o.SetFlag("budget_exceeded", true)

	assert.True(t, o.Flag("budget_exceeded"))
}

func TestOperationContext_Flag_Default(t *testing.T) {
	o := newTestOpCtx()

	assert.False(t, o.Flag("not_set"))
}

func TestOperationContext_SetFlag_False(t *testing.T) {
	o := newTestOpCtx()
	o.SetFlag("approved", true)
	o.SetFlag("approved", false)

	assert.False(t, o.Flag("approved"))
}

// ── Suspend ───────────────────────────────────────────────────────────────────

func TestOperationContext_Suspend(t *testing.T) {
	o := newTestOpCtx()
	require.False(t, o.Suspended)

	o.Suspend("pending_workflow_approval:abc-123", "gl.post_transaction")

	assert.True(t, o.Suspended)
	assert.Equal(t, "pending_workflow_approval:abc-123", o.SuspendReason)
	assert.Equal(t, "gl.post_transaction", o.ResumePoint)
}

// ── RegisterTxHook ────────────────────────────────────────────────────────────

func TestOperationContext_RegisterTxHook(t *testing.T) {
	o := newTestOpCtx()
	require.Len(t, o.TxHooks, 0)

	o.RegisterTxHook(TxHook{Name: "hook-a", Priority: 10})
	o.RegisterTxHook(TxHook{Name: "hook-b", Priority: 20})

	assert.Len(t, o.TxHooks, 2)
	assert.Equal(t, "hook-a", o.TxHooks[0].Name)
	assert.Equal(t, "hook-b", o.TxHooks[1].Name)
}

// ── Can ───────────────────────────────────────────────────────────────────────

func TestOperationContext_Can_PermissionPresent(t *testing.T) {
	o := newTestOpCtx()
	assert.True(t, o.Can("ap.invoices", "read"))
}

func TestOperationContext_Can_PermissionAbsent(t *testing.T) {
	o := newTestOpCtx()
	assert.False(t, o.Can("finance.journals", "post"))
}

func TestOperationContext_Can_NilSession(t *testing.T) {
	o := newTestOpCtx()
	o.Session = nil
	assert.False(t, o.Can("ap.invoices", "read"))
}

// ── FeatureEnabled ────────────────────────────────────────────────────────────

func TestOperationContext_FeatureEnabled_On(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(map[string]bool{"budget": true}, nil)

	assert.True(t, o.FeatureEnabled("budget"))
}

func TestOperationContext_FeatureEnabled_Off(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(map[string]bool{"budget": false}, nil)

	assert.False(t, o.FeatureEnabled("budget"))
}

func TestOperationContext_FeatureEnabled_Missing(t *testing.T) {
	o := newTestOpCtx()
	assert.False(t, o.FeatureEnabled("nonexistent_flag"))
}

func TestOperationContext_FeatureEnabled_NilSession(t *testing.T) {
	o := newTestOpCtx()
	o.Session = nil
	assert.False(t, o.FeatureEnabled("budget"))
}

// ── SettingDecimal ────────────────────────────────────────────────────────────

func TestOperationContext_SettingDecimal_Present(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(nil, map[string]string{"budget.minimum_check_amount": "5000.50"})

	got := o.SettingDecimal("budget.minimum_check_amount", decimal.Zero)
	assert.True(t, decimal.NewFromFloat(5000.50).Equal(got))
}

func TestOperationContext_SettingDecimal_Default(t *testing.T) {
	o := newTestOpCtx()
	def := decimal.NewFromInt(100)

	got := o.SettingDecimal("missing.key", def)
	assert.True(t, def.Equal(got))
}

func TestOperationContext_SettingDecimal_InvalidValue_ReturnsDefault(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(nil, map[string]string{"bad.key": "not-a-number"})
	def := decimal.NewFromInt(99)

	got := o.SettingDecimal("bad.key", def)
	assert.True(t, def.Equal(got))
}

func TestOperationContext_SettingDecimal_NilSession(t *testing.T) {
	o := newTestOpCtx()
	o.Session = nil
	def := decimal.NewFromInt(42)

	got := o.SettingDecimal("any.key", def)
	assert.True(t, def.Equal(got))
}

// ── SettingBool ───────────────────────────────────────────────────────────────

func TestOperationContext_SettingBool_True(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(nil, map[string]string{"feature.enabled": "true"})

	assert.True(t, o.SettingBool("feature.enabled", false))
}

func TestOperationContext_SettingBool_False(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(nil, map[string]string{"feature.enabled": "false"})

	assert.False(t, o.SettingBool("feature.enabled", true))
}

func TestOperationContext_SettingBool_Default(t *testing.T) {
	o := newTestOpCtx()
	assert.True(t, o.SettingBool("missing", true))
	assert.False(t, o.SettingBool("missing", false))
}

// ── SettingString ─────────────────────────────────────────────────────────────

func TestOperationContext_SettingString_Present(t *testing.T) {
	o := newTestOpCtx()
	o.Session = newTestSession(nil, map[string]string{"budget.control_mode": "hard_block"})

	assert.Equal(t, "hard_block", o.SettingString("budget.control_mode", "warn"))
}

func TestOperationContext_SettingString_Default(t *testing.T) {
	o := newTestOpCtx()
	assert.Equal(t, "warn", o.SettingString("missing.key", "warn"))
}

// ── Object pool ───────────────────────────────────────────────────────────────

func TestOperationContext_Pool_Reset(t *testing.T) {
	o := AcquireOperationContext()

	// Populate fields
	o.OperationKey = "ap.invoice.process"
	o.SetData("gl_posting.transaction_id", uuid.New())
	o.SetFlag("budget_exceeded", true)
	o.Suspended = true
	o.SuspendReason = "some reason"
	now := time.Now()
	o.StartedAt = now
	o.Log = append(o.Log, StageLog{StageName: "test", Status: "completed"})
	o.RegisterTxHook(TxHook{Name: "hook-1"})

	ReleaseOperationContext(o)

	// Re-acquire — may or may not be the same pointer; either way must be zero
	o2 := AcquireOperationContext()
	defer ReleaseOperationContext(o2)

	assert.Empty(t, o2.OperationKey)
	assert.Empty(t, o2.Data)
	assert.Empty(t, o2.Flags)
	assert.Empty(t, o2.Log)
	assert.Empty(t, o2.TxHooks)
	assert.False(t, o2.Suspended)
	assert.Empty(t, o2.SuspendReason)
	assert.True(t, o2.StartedAt.IsZero())
	assert.Nil(t, o2.CompletedAt)
	assert.Nil(t, o2.Session)
	assert.Nil(t, o2.Input)
	assert.False(t, o2.DryRun)
}

func TestOperationContext_Pool_MapsPreAllocated(t *testing.T) {
	// Maps should be non-nil on fresh acquire so stages can write immediately
	o := AcquireOperationContext()
	defer ReleaseOperationContext(o)

	require.NotNil(t, o.Data)
	require.NotNil(t, o.Flags)
	require.NotNil(t, o.Log)

	// Should not panic on write
	assert.NotPanics(t, func() {
		o.SetData("key", "value")
		o.SetFlag("flag", true)
	})
}
