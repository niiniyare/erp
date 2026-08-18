package finance

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"awo.so/awo/def"
)

// --- mock ActionRuntime ---

type mockRuntime struct {
	records   map[uuid.UUID]*def.EntityRecord
	txErr     error
	clockTime time.Time
}

func newMockRuntime(records ...*def.EntityRecord) *mockRuntime {
	m := &mockRuntime{
		records:   make(map[uuid.UUID]*def.EntityRecord),
		clockTime: time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC),
	}
	for _, r := range records {
		m.records[r.ID] = r
	}
	return m
}

func (m *mockRuntime) Repo(_ string) def.ActionEntityRepo   { return &mockRepo{rt: m} }
func (m *mockRuntime) Tx(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}
func (m *mockRuntime) Publish(_ context.Context, _ def.ActionEvent) error       { return nil }
func (m *mockRuntime) StartWorkflow(_ context.Context, _ def.ActionWorkflowSpec) (string, error) {
	return "wf-id", nil
}
func (m *mockRuntime) Notify(_ context.Context, _ def.ActionNotification) error  { return nil }
func (m *mockRuntime) InvalidateCache(_ context.Context, _ string) error         { return nil }
func (m *mockRuntime) Cache() def.ActionCache                                    { return nil }
func (m *mockRuntime) Clock() time.Time                                          { return m.clockTime }
func (m *mockRuntime) Logger() *slog.Logger                                      { return slog.Default() }
func (m *mockRuntime) TenantID() uuid.UUID                                       { return uuid.Nil }
func (m *mockRuntime) Actor() *def.Actor                                         { return &def.Actor{} }

// --- mock ActionEntityRepo ---

type mockRepo struct{ rt *mockRuntime }

func (r *mockRepo) EntityName() string { return "mock" }

func (r *mockRepo) Get(_ context.Context, id uuid.UUID) (*def.EntityRecord, error) {
	rec, ok := r.rt.records[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return rec, nil
}

func (r *mockRepo) Query(_ context.Context, _ def.ActionFilter, _ ...def.ActionQueryOpt) ([]*def.EntityRecord, error) {
	return nil, nil
}

func (r *mockRepo) Count(_ context.Context, _ def.ActionFilter) (int64, error) { return 0, nil }
func (r *mockRepo) Exists(_ context.Context, _ def.ActionFilter) (bool, error) { return false, nil }

func (r *mockRepo) Create(_ context.Context, data map[string]any) (*def.EntityRecord, error) {
	id := uuid.New()
	rec := &def.EntityRecord{ID: id, Data: data}
	r.rt.records[id] = rec
	return rec, nil
}

func (r *mockRepo) Update(_ context.Context, id uuid.UUID, patch map[string]any) (*def.EntityRecord, error) {
	rec, ok := r.rt.records[id]
	if !ok {
		return nil, errors.New("not found")
	}
	for k, v := range patch {
		rec.Data[k] = v
	}
	return rec, nil
}

func (r *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.rt.records, id)
	return nil
}

// --- helpers ---

func newRecord(status string) *def.EntityRecord {
	return &def.EntityRecord{
		ID:   uuid.New(),
		Data: map[string]any{"status": status},
	}
}

func makeCtx(rt def.ActionRuntime, id uuid.UUID) *def.ActionContext {
	return &def.ActionContext{
		Ctx:      context.Background(),
		RecordID: id,
		Runtime:  rt,
		Actor:    &def.Actor{},
	}
}

// --- transitionStatus tests ---

func TestTransitionStatus_Success(t *testing.T) {
	rec := newRecord("draft")
	rt := newMockRuntime(rec)
	h := transitionStatus("finance_fiscal_year", "activate", "draft", "active")

	result, err := h(makeCtx(rt, rec.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if rt.records[rec.ID].Data["status"] != "active" {
		t.Errorf("expected status=active, got %v", rt.records[rec.ID].Data["status"])
	}
}

func TestTransitionStatus_WrongState(t *testing.T) {
	rec := newRecord("posted") // not draft
	rt := newMockRuntime(rec)
	h := transitionStatus("finance_journal_entry", "submit", "draft", "submitted")

	_, err := h(makeCtx(rt, rec.ID))
	if err == nil {
		t.Fatal("expected error for wrong state")
	}
}

func TestTransitionStatus_NotFound(t *testing.T) {
	rt := newMockRuntime()
	h := transitionStatus("finance_journal_entry", "submit", "draft", "submitted")

	_, err := h(makeCtx(rt, uuid.New()))
	if err == nil {
		t.Fatal("expected error for missing record")
	}
}

// --- transitionStatusWithTimestamp tests ---

func TestTransitionStatusWithTimestamp_SetsField(t *testing.T) {
	rec := newRecord("open")
	rec.Data["closed_at"] = nil
	rt := newMockRuntime(rec)
	h := transitionStatusWithTimestamp("finance_accounting_period", "close", "open", "closed", "closed_at")

	_, err := h(makeCtx(rt, rec.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated := rt.records[rec.ID]
	if updated.Data["status"] != "closed" {
		t.Errorf("expected status=closed, got %v", updated.Data["status"])
	}
	if updated.Data["closed_at"] == nil {
		t.Error("closed_at should be set")
	}
}

// --- cancelAction tests ---

func TestCancelAction_FromDraft(t *testing.T) {
	rec := newRecord("draft")
	rt := newMockRuntime(rec)
	h := cancelAction("finance_payment", "processed", "reconciled")

	result, err := h(makeCtx(rt, rec.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if rt.records[rec.ID].Data["status"] != "cancelled" {
		t.Errorf("expected status=cancelled")
	}
}

func TestCancelAction_TerminalState_Rejected(t *testing.T) {
	rec := newRecord("reconciled")
	rt := newMockRuntime(rec)
	h := cancelAction("finance_payment", "processed", "reconciled")

	_, err := h(makeCtx(rt, rec.ID))
	if err == nil {
		t.Fatal("expected error cancelling terminal status")
	}
}

// --- postJournalEntry tests ---

func TestPostJournalEntry_FromSubmitted(t *testing.T) {
	rec := newRecord("submitted")
	rt := newMockRuntime(rec)
	h := postJournalEntry()

	result, err := h(makeCtx(rt, rec.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Message == "" {
		t.Error("expected non-empty message")
	}
	if rt.records[rec.ID].Data["status"] != "posted" {
		t.Errorf("expected status=posted")
	}
}

func TestPostJournalEntry_WrongState(t *testing.T) {
	rec := newRecord("draft")
	rt := newMockRuntime(rec)
	h := postJournalEntry()

	_, err := h(makeCtx(rt, rec.ID))
	if err == nil {
		t.Fatal("expected error posting non-submitted entry")
	}
}

// --- reverseJournalEntry tests ---

func TestReverseJournalEntry_FromPosted(t *testing.T) {
	rec := newRecord("posted")
	rec.Data["journal"] = uuid.New()
	rec.Data["reference"] = "JE-2026-00001"
	rt := newMockRuntime(rec)
	h := reverseJournalEntry()

	result, err := h(makeCtx(rt, rec.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	// Original should be reversed.
	if rt.records[rec.ID].Data["status"] != "reversed" {
		t.Errorf("original entry should be reversed, got %v", rt.records[rec.ID].Data["status"])
	}
	// A reversal record should have been created.
	if len(rt.records) < 2 {
		t.Error("expected reversal record to be created")
	}
}

func TestReverseJournalEntry_NotPosted_Rejected(t *testing.T) {
	rec := newRecord("submitted")
	rt := newMockRuntime(rec)
	h := reverseJournalEntry()

	_, err := h(makeCtx(rt, rec.ID))
	if err == nil {
		t.Fatal("expected error reversing non-posted entry")
	}
}

// --- closeFiscalYear tests ---

func TestCloseFiscalYear_FromActive(t *testing.T) {
	fy := newRecord("active")
	rt := newMockRuntime(fy)
	h := closeFiscalYear()

	result, err := h(makeCtx(rt, fy.ID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if rt.records[fy.ID].Data["status"] != "closed" {
		t.Errorf("expected fiscal year status=closed")
	}
}

func TestCloseFiscalYear_WrongState(t *testing.T) {
	fy := newRecord("draft")
	rt := newMockRuntime(fy)
	h := closeFiscalYear()

	_, err := h(makeCtx(rt, fy.ID))
	if err == nil {
		t.Fatal("expected error closing non-active fiscal year")
	}
}
