package featureflag

//
// import (
// 	"context"
// 	"testing"
//
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
//
// 	"awo/internal/core/audit"
// )
//
// // SimpleAuditService for testing audit integration
// type SimpleAuditService struct {
// 	RecordedEvents []audit.AuditEvent
// }
//
// func (m *SimpleAuditService) Record(ctx context.Context, event audit.AuditEvent) error {
// 	m.RecordedEvents = append(m.RecordedEvents, event)
// 	return nil
// }
//
// // TestAuditHelperMethods tests the audit helper methods directly
// func TestAuditHelperMethods(t *testing.T) {
// 	ctx := context.Background()
// 	auditService := &SimpleAuditService{}
//
// 	service := &simpleServiceImpl{
// 		auditService: auditService,
// 	}
//
// 	t.Run("auditFlagCreated", func(t *testing.T) {
// 		testFlag := &FeatureFlag{
// 			ID:           uuid.New(),
// 			TenantID:     uuid.New(),
// 			Name:         "test-flag",
// 			Description:  "Test flag",
// 			FlagType:     FlagTypeBoolean,
// 			DefaultValue: true,
// 		}
//
// 		request := &CreateFeatureFlagRequest{
// 			Name:         "test-flag",
// 			Description:  "Test flag",
// 			FlagType:     FlagTypeBoolean,
// 			DefaultValue: true,
// 		}
//
// 		// Call audit method
// 		service.auditFlagCreated(ctx, testFlag, request)
//
// 		// Verify audit event was recorded
// 		assert.Len(t, auditService.RecordedEvents, 1)
// 		event := auditService.RecordedEvents[0]
//
// 		assert.Equal(t, AuditEventFeatureFlagCreated, event.EventType)
// 		assert.Equal(t, AuditCategoryFeatureManagement, event.EventCategory)
// 		assert.Equal(t, AuditSeverityInfo, event.Severity)
// 		assert.Equal(t, "CREATED", event.Decision)
// 		assert.Contains(t, event.Reason, "test-flag")
// 		assert.Equal(t, testFlag.ID, event.EntityID.UUID)
// 		assert.True(t, event.EntityID.Valid)
// 	})
//
// 	t.Run("auditFlagEvaluated", func(t *testing.T) {
// 		// Reset audit service
// 		auditService.RecordedEvents = nil
//
// 		testFlag := &FeatureFlag{
// 			ID:           uuid.New(),
// 			TenantID:     uuid.New(),
// 			Name:         "eval-flag",
// 			Description:  "Evaluation test flag",
// 			FlagType:     FlagTypeBoolean,
// 			DefaultValue: true,
// 		}
//
// 		evalCtx := &EvaluationContext{
// 			TenantID:    testFlag.TenantID,
// 			Environment: "test",
// 		}
//
// 		result := &EvaluationResult{
// 			FlagName: testFlag.Name,
// 			Value:    true,
// 			Enabled:  true,
// 			Reason:   string(ReasonDefaultValue),
// 		}
//
// 		// Call audit method
// 		service.auditFlagEvaluated(ctx, testFlag, evalCtx, result)
//
// 		// Verify audit event was recorded
// 		assert.Len(t, auditService.RecordedEvents, 1)
// 		event := auditService.RecordedEvents[0]
//
// 		assert.Equal(t, AuditEventFeatureFlagEvaluated, event.EventType)
// 		assert.Equal(t, AuditCategoryAccess, event.EventCategory)
// 		assert.Equal(t, AuditSeverityInfo, event.Severity)
// 		assert.Equal(t, "EVALUATED_true", event.Decision)
// 		assert.Contains(t, event.Reason, "eval-flag")
// 		assert.Equal(t, testFlag.ID, event.EntityID.UUID)
// 		assert.True(t, event.EntityID.Valid)
// 	})
//
// 	t.Run("auditFlagDeleted", func(t *testing.T) {
// 		// Reset audit service
// 		auditService.RecordedEvents = nil
//
// 		testFlag := &FeatureFlag{
// 			ID:           uuid.New(),
// 			TenantID:     uuid.New(),
// 			Name:         "delete-flag",
// 			Description:  "Flag to delete",
// 			FlagType:     FlagTypeBoolean,
// 			DefaultValue: false,
// 		}
//
// 		// Call audit method
// 		service.auditFlagDeleted(ctx, testFlag)
//
// 		// Verify audit event was recorded
// 		assert.Len(t, auditService.RecordedEvents, 1)
// 		event := auditService.RecordedEvents[0]
//
// 		assert.Equal(t, AuditEventFeatureFlagDeleted, event.EventType)
// 		assert.Equal(t, AuditCategoryFeatureManagement, event.EventCategory)
// 		assert.Equal(t, AuditSeverityHigh, event.Severity) // Deletions are high severity
// 		assert.Equal(t, "DELETED", event.Decision)
// 		assert.Contains(t, event.Reason, "delete-flag")
// 		assert.Equal(t, testFlag.ID, event.EntityID.UUID)
// 		assert.True(t, event.EntityID.Valid)
// 	})
// }
