package audit

import (
	"context"
	"encoding/json"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

type service struct {
	repo    Repository
	cache   cache.Service
	logger  logger.Logger
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewService creates a new audit service.
func NewService(
	repo Repository,
	cache cache.Service,
	logger logger.Logger,
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
) Service {
	return &service{
		repo:    repo,
		cache:   cache,
		logger:  logger,
		tracing: tracing,
		metrics: metrics,
	}
}

func (s *service) CreateAuditEvent(ctx context.Context, req CreateAuditEventRequest) (*AuditEvent, error) {
	// The repository will get tenant ID from the database session context (RLS)
	return s.repo.CreateAuditEvent(ctx, req)
}

func (s *service) GetAuditEvents(ctx context.Context, tenantID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	// TODO: Add caching
	return s.repo.GetAuditEvents(ctx, tenantID, filters)
}

func (s *service) GetAuditEventByID(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID) (*AuditEvent, error) {
	// TODO: Add caching
	return s.repo.GetAuditEventByID(ctx, tenantID, eventID)
}

func (s *service) GetUserAuditHistory(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	return s.repo.GetUserAuditHistory(ctx, tenantID, userID, filters)
}

func (s *service) GetHighRiskEvents(ctx context.Context, tenantID uuid.UUID, minRiskScore int, filters AuditEventFilters) ([]AuditEvent, error) {
	return s.repo.GetHighRiskEvents(ctx, tenantID, minRiskScore, filters)
}

func (s *service) GetFailedAccessAttempts(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, userID *uuid.UUID, ipAddress *string, limit, offset int) ([]AuditEvent, error) {
	// this is a string, but the repository expects a netip.Addr
	var ip *netip.Addr
	if ipAddress != nil {
		addr, err := netip.ParseAddr(*ipAddress)
		if err != nil {
			return nil, err
		}
		ip = &addr
	}
	return s.repo.GetFailedAccessAttempts(ctx, tenantID, startTime, endTime, userID, ip, limit, offset)
}

func (s *service) GetComplianceEvents(ctx context.Context, tenantID uuid.UUID, complianceFlag string, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error) {
	return s.repo.GetComplianceEvents(ctx, tenantID, complianceFlag, startTime, endTime, limit, offset)
}

func (s *service) GetAuditEventsByEntity(ctx context.Context, tenantID uuid.UUID, entityID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	return s.repo.GetAuditEventsByEntity(ctx, tenantID, entityID, filters)
}

func (s *service) GetAuditEventsByResource(ctx context.Context, tenantID uuid.UUID, resourceID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	return s.repo.GetAuditEventsByResource(ctx, tenantID, resourceID, filters)
}

func (s *service) GetAdminActions(ctx context.Context, tenantID uuid.UUID, adminUserID, targetUserID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error) {
	return s.repo.GetAdminActions(ctx, tenantID, adminUserID, targetUserID, startTime, endTime, limit, offset)
}

func (s *service) GetUserSessionEvents(ctx context.Context, tenantID uuid.UUID, sessionID uuid.UUID) ([]AuditEvent, error) {
	return s.repo.GetUserSessionEvents(ctx, tenantID, sessionID)
}

func (s *service) GetAuditStatsByCategory(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, severity *string) ([]AuditStatsByCategory, error) {
	return s.repo.GetAuditStatsByCategory(ctx, tenantID, startTime, endTime, severity)
}

func (s *service) GetUserRiskProfile(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime *time.Time) (*UserRiskProfile, error) {
	return s.repo.GetUserRiskProfile(ctx, tenantID, userID, startTime)
}

func (s *service) GetAuditLogHealth(ctx context.Context, tenantID uuid.UUID) (*AuditLogHealth, error) {
	return s.repo.GetAuditLogHealth(ctx, tenantID)
}

func (s *service) GetHourlyEventRates(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, limit int) ([]HourlyEventRate, error) {
	return s.repo.GetHourlyEventRates(ctx, tenantID, startTime, endTime, limit)
}

func (s *service) GetEventTimelineForUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime, endTime time.Time) ([]AuditEvent, error) {
	// This should return []EventTimeline, but the service interface returns []AuditEvent
	// I will need to discuss this with the user
	return nil, nil
}

func (s *service) GetRelatedEventsByContext(ctx context.Context, tenantID uuid.UUID, contextKey, contextSearch string, startTime, endTime time.Time, excludeEventID *uuid.UUID, limit int) ([]AuditEvent, error) {
	// This should return []RelatedEventsByContext, but the service interface returns []AuditEvent
	// I will need to discuss this with the user
	return nil, nil
}

func (s *service) UpdateEventRiskScore(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, riskScore int) error {
	return s.repo.UpdateEventRiskScore(ctx, tenantID, eventID, riskScore)
}

func (s *service) UpdateEventComplianceFlags(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, complianceFlags json.RawMessage) error {
	return s.repo.UpdateEventComplianceFlags(ctx, tenantID, eventID, complianceFlags)
}

func (s *service) DeleteOldAuditEvents(ctx context.Context, tenantID uuid.UUID, cutoffDate time.Time) error {
	return s.repo.DeleteOldAuditEvents(ctx, tenantID, cutoffDate)
}
