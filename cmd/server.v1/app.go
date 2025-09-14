package main

import (
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/abac"
	"github.com/niiniyare/erp/internal/core/access/approval"
	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/access/execution"
	"github.com/niiniyare/erp/internal/core/access/request"
	"github.com/niiniyare/erp/internal/core/analytics"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/entity"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Services holds the application's core business services.
type Services struct {
	TenantService            tenant.Service
	EntityService            entity.Service
	IdentityService          identity.Service
	ABACService              abac.Service
	AuditService             audit.Service
	FeatureFlagService       featureflag.Service
	AdminFeatureFlagService  featureflag.AdminService
	AccessRequestService     request.AccessRequestService
	ConditionalAccessService conditional.ConditionalAccessService
	AnalyticsService         analytics.UserAnalyticsService
	NotificationService      notification.NotificationService
	ApproverService          approval.ApproverService
	ExecutionService         execution.AccessExecutionService
}

// application holds the application's dependencies.
type application struct {
	config          *config.Config
	logger          logger.Logger
	store           db.Store
	redis           cache.Service
	temporal        *temporal.Platform
	tracer          tracing.TracingService
	metrics         *metrics.MetricsService
	services        *Services
	financeServices *service.Services
}