package naming_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"awo.so/framework/definition"
	"awo.so/framework/naming"
)

// ── test double ──────────────────────────────────────────────────────────────

type stubRec struct{ data map[string]any }

func (r *stubRec) Get(f string) any    { return r.data[f] }
func (r *stubRec) Set(f string, v any) { r.data[f] = v }
func (r *stubRec) ID() uuid.UUID       { return uuid.Nil }
func (r *stubRec) TenantID() uuid.UUID { return uuid.Nil }
func (r *stubRec) EntityName() string  { return "invoice" }

var _ definition.MutableRecord = (*stubRec)(nil)

// mockExec simulates an atomic counter (no DB needed).
func mockExec(seq *int64) naming.ExecOneRow {
	return func(_ string, _ []any, dest []any) error {
		*seq++
		*(dest[0].(*int64)) = *seq
		return nil
	}
}

// ── suite ────────────────────────────────────────────────────────────────────

type NamingSuite struct{ suite.Suite }

func TestNamingSuite(t *testing.T) { suite.Run(t, new(NamingSuite)) }

// ── tests ────────────────────────────────────────────────────────────────────

func (s *NamingSuite) TestFormat_Basic() {
	ns := &definition.NamingSeriesDef{Prefix: "INV-", Padding: 5}
	s.Equal("INV-00001", naming.Format(ns, 1))
	s.Equal("INV-00042", naming.Format(ns, 42))
}

func (s *NamingSuite) TestFormat_PaddingExceeded_NoTruncation() {
	ns := &definition.NamingSeriesDef{Prefix: "INV-", Padding: 4}
	s.Equal("INV-100000", naming.Format(ns, 100000))
}

func (s *NamingSuite) TestFormat_DefaultPadding() {
	ns := &definition.NamingSeriesDef{Prefix: "ORD-"} // Padding=0 → 5
	s.Equal("ORD-00007", naming.Format(ns, 7))
}

func (s *NamingSuite) TestNext_Increments() {
	var seq int64
	exec := mockExec(&seq)
	tenantID := uuid.New()

	n1, err := naming.Next(exec, tenantID, "invoice")
	s.Require().NoError(err)
	s.Equal(int64(1), n1)

	n2, err := naming.Next(exec, tenantID, "invoice")
	s.Require().NoError(err)
	s.Equal(int64(2), n2)
}

func (s *NamingSuite) TestNext_PassesCorrectArgs() {
	tenantID := uuid.New()
	var gotArgs []any
	exec := func(_ string, args []any, dest []any) error {
		gotArgs = args
		*(dest[0].(*int64)) = 1
		return nil
	}
	_, err := naming.Next(exec, tenantID, "purchase_order")
	s.Require().NoError(err)
	s.Require().Len(gotArgs, 2)
	s.Equal(tenantID, gotArgs[0])
	s.Equal("purchase_order", gotArgs[1])
}

func (s *NamingSuite) TestStamp_NilSeries_Noop() {
	def := &definition.EntityDefinition{Name: "invoice", NamingSeries: nil}
	called := false
	exec := func(_ string, _ []any, _ []any) error {
		called = true
		return nil
	}
	r := &stubRec{data: map[string]any{}}
	s.Require().NoError(naming.Stamp(exec, uuid.New(), def, r))
	s.False(called, "exec must not be called when NamingSeries is nil")
}

func (s *NamingSuite) TestStamp_SetsTargetField() {
	def := &definition.EntityDefinition{
		Name:         "invoice",
		NamingSeries: &definition.NamingSeriesDef{Field: "name", Prefix: "INV-", Padding: 4},
	}
	var seq int64
	r := &stubRec{data: map[string]any{}}
	s.Require().NoError(naming.Stamp(mockExec(&seq), uuid.New(), def, r))
	s.Equal("INV-0001", r.data["name"])
}

func (s *NamingSuite) TestStamp_SequentialCalls_DifferentValues() {
	def := &definition.EntityDefinition{
		Name:         "invoice",
		NamingSeries: &definition.NamingSeriesDef{Field: "name", Prefix: "INV-", Padding: 3},
	}
	var seq int64
	exec := mockExec(&seq)
	tenantID := uuid.New()

	r1 := &stubRec{data: map[string]any{}}
	r2 := &stubRec{data: map[string]any{}}
	s.Require().NoError(naming.Stamp(exec, tenantID, def, r1))
	s.Require().NoError(naming.Stamp(exec, tenantID, def, r2))

	s.Equal("INV-001", r1.data["name"])
	s.Equal("INV-002", r2.data["name"])
	s.NotEqual(r1.data["name"], r2.data["name"])
}
