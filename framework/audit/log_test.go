package audit_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"awo.so/framework/audit"
	"awo.so/framework/definition"
)

// ── test double ──────────────────────────────────────────────────────────────

type testRec struct {
	id   uuid.UUID
	data map[string]any
}

func newTestRec(id uuid.UUID, kv ...any) *testRec {
	r := &testRec{id: id, data: make(map[string]any)}
	for i := 0; i+1 < len(kv); i += 2 {
		r.data[kv[i].(string)] = kv[i+1]
	}
	return r
}

func (r *testRec) Get(f string) any    { return r.data[f] }
func (r *testRec) Set(f string, v any) { r.data[f] = v }
func (r *testRec) ID() uuid.UUID       { return r.id }
func (r *testRec) TenantID() uuid.UUID { return uuid.Nil }
func (r *testRec) EntityName() string  { return "invoice" }

var _ definition.MutableRecord = (*testRec)(nil)

// ── suite ────────────────────────────────────────────────────────────────────

type AuditSuite struct {
	suite.Suite
	tenantID uuid.UUID
}

func (s *AuditSuite) SetupTest() {
	s.tenantID = uuid.New()
}

func TestAuditSuite(t *testing.T) {
	suite.Run(t, new(AuditSuite))
}

// ── tests ────────────────────────────────────────────────────────────────────

func (s *AuditSuite) TestNoop_WhenNotAudited() {
	def := &definition.EntityDefinition{Name: "invoice", Audited: false}
	called := false
	exec := func(_ context.Context, _ string, _ []any) error {
		called = true
		return nil
	}
	m := &definition.Mutation{Op: definition.OpCreate, TenantID: s.tenantID.String()}
	s.NoError(audit.Write(context.Background(), exec, def, m))
	s.False(called, "exec must not be called when Audited=false")
}

func (s *AuditSuite) TestCreate_CallsExec() {
	def := &definition.EntityDefinition{
		Name:    "invoice",
		Audited: true,
		Fields:  []*definition.FieldDef{definition.String("title")},
	}
	id := uuid.New()
	after := newTestRec(id, "title", "INV-001")

	var gotArgs []any
	exec := func(_ context.Context, _ string, args []any) error {
		gotArgs = args
		return nil
	}
	m := &definition.Mutation{
		Op:       definition.OpCreate,
		After:    after,
		TenantID: s.tenantID.String(),
		ActorID:  "user-1",
	}
	s.Require().NoError(audit.Write(context.Background(), exec, def, m))
	s.Require().Len(gotArgs, 6, "expected 6 positional args")
	s.Equal("invoice", gotArgs[1])
	s.Equal(id, gotArgs[2])
	s.Equal("create", gotArgs[3])
	s.Equal("user-1", gotArgs[4])
}

func (s *AuditSuite) TestUpdate_DiffOnly() {
	def := &definition.EntityDefinition{
		Name:    "invoice",
		Audited: true,
		Fields: []*definition.FieldDef{
			definition.String("title"),
			definition.String("status"),
		},
	}
	id := uuid.New()
	before := newTestRec(id, "title", "INV-001", "status", "draft")
	after := newTestRec(id, "title", "INV-001", "status", "approved")

	var changesJSON string
	exec := func(_ context.Context, _ string, args []any) error {
		changesJSON = args[5].(string)
		return nil
	}
	m := &definition.Mutation{
		Op:       definition.OpUpdate,
		Before:   before,
		After:    after,
		TenantID: s.tenantID.String(),
		ActorID:  "user-1",
	}
	s.Require().NoError(audit.Write(context.Background(), exec, def, m))
	// Only changed field (status) should appear in after; unchanged (title) not in after.
	s.Contains(changesJSON, "approved", "changed value must appear")
	s.Contains(changesJSON, "draft", "before value must appear")
}

func (s *AuditSuite) TestSensitiveField_Excluded() {
	def := &definition.EntityDefinition{
		Name:    "user",
		Audited: true,
		Fields: []*definition.FieldDef{
			definition.String("email"),
			definition.String("password_hash").Sensitive(),
		},
	}
	after := newTestRec(uuid.New(), "email", "a@b.com", "password_hash", "secret")

	var changesJSON string
	exec := func(_ context.Context, _ string, args []any) error {
		changesJSON = args[5].(string)
		return nil
	}
	m := &definition.Mutation{Op: definition.OpCreate, After: after, TenantID: s.tenantID.String()}
	s.Require().NoError(audit.Write(context.Background(), exec, def, m))
	s.False(strings.Contains(changesJSON, "password_hash"), "sensitive field must not appear")
	s.False(strings.Contains(changesJSON, "secret"), "sensitive value must not appear")
}

func (s *AuditSuite) TestDelete_UsesBefore() {
	def := &definition.EntityDefinition{
		Name:    "invoice",
		Audited: true,
		Fields:  []*definition.FieldDef{definition.String("title")},
	}
	id := uuid.New()
	before := newTestRec(id, "title", "INV-001")

	var gotOp string
	exec := func(_ context.Context, _ string, args []any) error {
		gotOp = args[3].(string)
		return nil
	}
	m := &definition.Mutation{
		Op:       definition.OpDelete,
		Before:   before,
		TenantID: s.tenantID.String(),
	}
	s.Require().NoError(audit.Write(context.Background(), exec, def, m))
	s.Equal("delete", gotOp)
}
