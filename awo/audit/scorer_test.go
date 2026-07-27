package audit

import (
	"testing"

	"github.com/google/uuid"

	"awo.so/awo/def"
)

func TestRiskScorer_Score_Bounds(t *testing.T) {
	t.Parallel()

	rs := NewRiskScorer()
	rec := &AuditRecord{
		TenantID:      uuid.New(),
		EntityName:    "finance_invoice",
		Operation:     OperationDelete,
		EventCategory: CategoryAdmin,
		Actor:         &def.Actor{UserID: uuid.New()},
	}
	score := rs.Score(rec)
	if score < 0 || score > 100 {
		t.Errorf("Score() = %d, must be in [0, 100]", score)
	}
}

func TestRiskScorer_Score_DeleteAdmin(t *testing.T) {
	t.Parallel()

	rs := NewRiskScorer()
	rec := &AuditRecord{
		EntityName:    "platform_tenant",
		Operation:     OperationDelete, // base = 30
		EventCategory: CategoryAdmin,   // premium = +30
		Actor:         &def.Actor{UserID: uuid.New()},
	}
	got := rs.Score(rec)
	// 30 + 30 = 60, no sensitive entity premium.
	if got != 60 {
		t.Errorf("Score(delete/admin) = %d, want 60", got)
	}
}

func TestRiskScorer_Score_Logout(t *testing.T) {
	t.Parallel()

	rs := NewRiskScorer()
	rec := &AuditRecord{
		EntityName:    "iam_session",
		Operation:     OperationLogout, // base = 0
		EventCategory: CategoryAuth,    // premium = +10
		Actor:         &def.Actor{UserID: uuid.New()},
	}
	got := rs.Score(rec)
	if got != 10 {
		t.Errorf("Score(logout/auth) = %d, want 10", got)
	}
}

func TestRiskScorer_Score_Cap(t *testing.T) {
	t.Parallel()

	rs := NewRiskScorer()
	// delete(30) + SECURITY(30) = 60; capped at 100 even with sensitive premium.
	rec := &AuditRecord{
		EntityName:    "finance_invoice",
		Operation:     OperationDelete,
		EventCategory: CategorySecurity,
		Actor:         &def.Actor{UserID: uuid.New()},
	}
	got := rs.Score(rec)
	if got > 100 {
		t.Errorf("Score must be capped at 100, got %d", got)
	}
}
