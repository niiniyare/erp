package settings

import (
	"context"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/settings/repository"
	"github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// SettingsService combines all settings-related services into a single interface
type SettingsService interface {
	service.ConfigurationService
	service.TemplateService
}

// settingsService implements SettingsService by embedding the individual services
type settingsService struct {
	service.ConfigurationService
	service.TemplateService
}

// NewSettingsService creates a new settings service that combines configuration and template services
func NewSettingsService(
	ctx context.Context,
	repo repository.ConfigurationRepository,
	auditService audit.Service,
	logger logger.Logger,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
) SettingsService {
	// Create individual services
	configService := service.NewConfigurationService(repo, auditService, logger, tracing, metrics)
	templateService := service.NewTemplateService(repo, auditService, logger, tracing, metrics)

	return &settingsService{
		ConfigurationService: configService,
		TemplateService:      templateService,
	}
}

// // Alternative approach: If you want more control over the service composition,
// // you can implement specific methods instead of embedding
//
// // SettingsServiceWithMethods provides an alternative implementation with explicit method delegation
// type SettingsServiceWithMethods struct {
// 	configService   service.ConfigurationService
// 	templateService service.TemplateService
// 	logger          logger.Logger
// }
//
// // NewSettingsServiceWithMethods creates a settings service with explicit method delegation
// func NewSettingsServiceWithMethods(
// 	ctx context.Context,
// 	repo repository.ConfigurationRepository,
// 	auditService audit.Service,
// 	logger logger.Logger,
// 	tracing tracing.TracingService,
// 	metrics metrics.MetricsProvider,
// ) *SettingsServiceWithMethods {
// 	return &SettingsServiceWithMethods{
// 		configService:   service.NewConfigurationService(repo, auditService, logger, tracing, metrics),
// 		templateService: service.NewTemplateService(repo, auditService, logger, tracing, metrics),
// 		logger:          logger,
// 	}
// }
//
// // Configuration Service Methods - delegate to configService
// func (s *SettingsServiceWithMethods) GetEffectiveConfiguration(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
// 	return s.configService.GetEffectiveConfiguration(ctx, entityID, module, key)
// }
//
// func (s *SettingsServiceWithMethods) ListEffectiveConfigurations(ctx context.Context, entityID *uuid.UUID, module domain.ModuleName) ([]*domain.Configuration, error) {
// 	return s.configService.ListEffectiveConfigurations(ctx, entityID, module)
// }
//
// func (s *SettingsServiceWithMethods) GetTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
// 	return s.configService.GetTenantConfiguration(ctx, module, key)
// }
//
// func (s *SettingsServiceWithMethods) UpdateTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error {
// 	return s.configService.UpdateTenantConfiguration(ctx, module, key, value, userID)
// }
//
// func (s *SettingsServiceWithMethods) DeleteTenantConfiguration(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error {
// 	return s.configService.DeleteTenantConfiguration(ctx, module, key, userID)
// }
//
// func (s *SettingsServiceWithMethods) GetEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey) (*domain.Configuration, error) {
// 	return s.configService.GetEntityConfiguration(ctx, entityID, module, key)
// }
//
// func (s *SettingsServiceWithMethods) UpdateEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue, userID uuid.UUID) error {
// 	return s.configService.UpdateEntityConfiguration(ctx, entityID, module, key, value, userID)
// }
//
// func (s *SettingsServiceWithMethods) DeleteEntityConfiguration(ctx context.Context, entityID uuid.UUID, module domain.ModuleName, key domain.ConfigKey, userID uuid.UUID) error {
// 	return s.configService.DeleteEntityConfiguration(ctx, entityID, module, key, userID)
// }
//
// func (s *SettingsServiceWithMethods) ValidateConfigurationValue(ctx context.Context, module domain.ModuleName, key domain.ConfigKey, value domain.ConfigValue) error {
// 	return s.configService.ValidateConfigurationValue(ctx, module, key, value)
// }
//
// func (s *SettingsServiceWithMethods) BulkUpdateConfigurations(ctx context.Context, updates []repository.BulkConfigurationUpdate) (*repository.BulkOperationResult, error) {
// 	return s.configService.BulkUpdateConfigurations(ctx, updates)
// }
//
// func (s *SettingsServiceWithMethods) SearchConfigurations(ctx context.Context, criteria repository.SearchCriteria) (*repository.SearchResult, error) {
// 	return s.configService.SearchConfigurations(ctx, criteria)
// }
//
// // Template Service Methods - delegate to templateService
// func (s *SettingsServiceWithMethods) CreateTemplate(ctx context.Context, template *domain.Template) error {
// 	return s.templateService.CreateTemplate(ctx, template)
// }
//
// func (s *SettingsServiceWithMethods) GetTemplate(ctx context.Context, templateID domain.TemplateID) (*domain.Template, error) {
// 	return s.templateService.GetTemplate(ctx, templateID)
// }
//
// func (s *SettingsServiceWithMethods) ListTemplates(ctx context.Context, filters repository.TemplateFilters) ([]*domain.Template, error) {
// 	return s.templateService.ListTemplates(ctx, filters)
// }
//
// func (s *SettingsServiceWithMethods) UpdateTemplate(ctx context.Context, template *domain.Template) error {
// 	return s.templateService.UpdateTemplate(ctx, template)
// }
//
// func (s *SettingsServiceWithMethods) DeactivateTemplate(ctx context.Context, templateID domain.TemplateID, userID uuid.UUID) error {
// 	return s.templateService.DeactivateTemplate(ctx, templateID, userID)
// }
//
// func (s *SettingsServiceWithMethods) ApplyTemplate(ctx context.Context, templateID domain.TemplateID, target repository.TemplateTarget, options repository.ApplyOptions, userID uuid.UUID) (*domain.ApplicationResult, error) {
// 	return s.templateService.ApplyTemplate(ctx, templateID, target, options, userID)
// }
//
// func (s *SettingsServiceWithMethods) ValidateTemplateApplication(ctx context.Context, templateID domain.TemplateID, target repository.TemplateTarget) (*service.TemplateValidationResult, error) {
// 	return s.templateService.ValidateTemplateApplication(ctx, templateID, target)
// }
//
// func (s *SettingsServiceWithMethods) GetTemplateApplicationHistory(ctx context.Context, entityID *uuid.UUID, limit, offset int) ([]*repository.TemplateApplication, error) {
// 	return s.templateService.GetTemplateApplicationHistory(ctx, entityID, limit, offset)
// }
//
// func (s *SettingsServiceWithMethods) GetTemplateUsageStats(ctx context.Context, templateID domain.TemplateID, startTime, endTime time.Time) (*service.TemplateUsageStats, error) {
// 	return s.templateService.GetTemplateUsageStats(ctx, templateID, startTime, endTime)
// }
//
// func (s *SettingsServiceWithMethods) GetMostUsedTemplates(ctx context.Context, limit int, startTime, endTime time.Time) ([]*service.TemplateUsageStats, error) {
// 	return s.templateService.GetMostUsedTemplates(ctx, limit, startTime, endTime)
// }
