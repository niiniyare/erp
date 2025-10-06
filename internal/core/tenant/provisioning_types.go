package tenant

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Comprehensive provisioning request types

// ComprehensiveProvisionRequest describes a complete tenant provisioning request
type ComprehensiveProvisionRequest struct {
	// Basic tenant information
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Subdomain   string `json:"subdomain" validate:"required,min=3,max=50,alphanum"`
	Description string `json:"description,omitempty" validate:"max=500"`

	// Business information
	Industry    string `json:"industry,omitempty" validate:"omitempty,oneof=technology healthcare finance retail manufacturing education government other"`
	CompanySize string `json:"company_size,omitempty" validate:"omitempty,oneof=startup small medium large enterprise"`
	Country     string `json:"country" validate:"required,iso3166_1_alpha2"`

	// Subscription and plan
	PlanType     string `json:"plan_type" validate:"required,oneof=starter professional enterprise"`
	BillingCycle string `json:"billing_cycle,omitempty" validate:"omitempty,oneof=monthly yearly"`

	// Contact information
	Contact ContactInfo `json:"contact" validate:"required"`

	// Admin user for the tenant
	AdminUser AdminUserRequest `json:"admin_user" validate:"required"`

	// Initial configuration
	InitialSettings TenantInitialSettings `json:"initial_settings,omitempty"`

	// Features to enable
	EnabledModules []string `json:"enabled_modules,omitempty" validate:"omitempty,dive,oneof=finance inventory hr crm project_management reporting"`

	// Usage limits
	InitialLimits TenantLimits `json:"initial_limits,omitempty"`
}

// AdminUserRequest describes the initial admin user
type AdminUserRequest struct {
	Email            string `json:"email" validate:"required,email"`
	FirstName        string `json:"first_name" validate:"required,min=1,max=50"`
	LastName         string `json:"last_name" validate:"required,min=1,max=50"`
	Phone            string `json:"phone,omitempty" validate:"omitempty,e164"`
	Timezone         string `json:"timezone,omitempty"`
	Language         string `json:"language,omitempty" validate:"omitempty,len=2"`
	SendWelcomeEmail bool   `json:"send_welcome_email"`
}

// TenantInitialSettings describes initial tenant configuration
type TenantInitialSettings struct {
	Timezone         string `json:"timezone,omitempty"`
	Currency         string `json:"currency,omitempty" validate:"omitempty,len=3,uppercase"`
	FiscalYearStart  uint   `json:"fiscal_year_start,omitempty" validate:"omitempty,min=1,max=12"`
	DateFormat       string `json:"date_format,omitempty" validate:"omitempty,oneof=MM/DD/YYYY DD/MM/YYYY YYYY-MM-DD"`
	NumberFormat     string `json:"number_format,omitempty" validate:"omitempty,oneof=1,234.56 1.234,56 1 234.56"`
	Language         string `json:"language,omitempty" validate:"omitempty,len=2"`
	AccountingMethod string `json:"accounting_method,omitempty" validate:"omitempty,oneof=accrual cash"`
}

// ComprehensiveProvisionResult describes the result of tenant provisioning
type ComprehensiveProvisionResult struct {
	TenantID      uuid.UUID                  `json:"tenant_id"`
	AdminUser     *AdminUserResult           `json:"admin_user,omitempty"`
	Configuration *TenantConfigurationResult `json:"configuration,omitempty"`
	SetupStatus   TenantSetupStatus          `json:"setup_status"`
	AccessInfo    TenantAccessInfo           `json:"access_info"`
}

// AdminUserResult describes admin user creation result
type AdminUserResult struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Status           string    `json:"status"`
	WelcomeEmailSent bool      `json:"welcome_email_sent"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TenantSetupStatus describes tenant setup completion status
type TenantSetupStatus struct {
	OverallStatus  string   `json:"overall_status"`
	CompletedSteps []string `json:"completed_steps"`
	FailedSteps    []string `json:"failed_steps"`
	NextSteps      []string `json:"next_steps"`
}

// TenantAccessInfo describes access information for new tenant
type TenantAccessInfo struct {
	TenantURL        string `json:"tenant_url"`
	AdminPortalURL   string `json:"admin_portal_url"`
	APIBaseURL       string `json:"api_base_url"`
	DocumentationURL string `json:"documentation_url,omitempty"`
}

// Tenant action request types

// SuspendTenantRequest describes a tenant suspension request
type SuspendTenantRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	Reason      string    `json:"reason" validate:"required,min=10,max=500"`
	ActorID     uuid.UUID `json:"actor_id" validate:"required"`
	NotifyUsers bool      `json:"notify_users"`
}

// ReactivateTenantRequest describes a tenant reactivation request
type ReactivateTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Reason   string    `json:"reason" validate:"required,min=10,max=500"`
	ActorID  uuid.UUID `json:"actor_id" validate:"required"`
}

// ArchiveTenantRequest describes a tenant archiving request
type ArchiveTenantRequest struct {
	TenantID          uuid.UUID `json:"tenant_id" validate:"required"`
	Reason            string    `json:"reason" validate:"required,min=10,max=500"`
	ActorID           uuid.UUID `json:"actor_id" validate:"required"`
	DataRetentionDays uint      `json:"data_retention_days" validate:"max=365"`
	ImmediateDeletion bool      `json:"immediate_deletion"`
}

// TenantActionResult describes the result of tenant actions
type TenantActionResult struct {
	TenantID      uuid.UUID  `json:"tenant_id"`
	Action        string     `json:"action"`
	Status        string     `json:"status"`
	Message       string     `json:"message"`
	EffectiveDate time.Time  `json:"effective_date"`
	AuditLogID    *uuid.UUID `json:"audit_log_id,omitempty"`
}

// Configuration management types

// UpdateTenantConfigurationRequest describes configuration update request
type UpdateTenantConfigurationRequest struct {
	TenantID                uuid.UUID                `json:"tenant_id" validate:"required"`
	Limits                  *TenantLimits            `json:"limits,omitempty"`
	EnabledFeatures         []string                 `json:"enabled_features,omitempty"`
	SecurityPolicies        *SecurityPolicies        `json:"security_policies,omitempty"`
	NotificationPreferences *NotificationPreferences `json:"notification_preferences,omitempty"`
	Reason                  string                   `json:"reason" validate:"required,min=10,max=500"`
	ActorID                 uuid.UUID                `json:"actor_id" validate:"required"`
}

// TenantConfigurationResult describes tenant configuration
type TenantConfigurationResult struct {
	Limits                  TenantLimits             `json:"limits"`
	EnabledFeatures         []string                 `json:"enabled_features"`
	SecurityPolicies        *SecurityPolicies        `json:"security_policies,omitempty"`
	IntegrationSettings     *IntegrationSettings     `json:"integration_settings,omitempty"`
	NotificationPreferences *NotificationPreferences `json:"notification_preferences,omitempty"`
	UpdatedAt               time.Time                `json:"updated_at"`
	UpdatedBy               *uuid.UUID               `json:"updated_by,omitempty"`
}

// SecurityPolicies describes security configuration
type SecurityPolicies struct {
	PasswordPolicy  *PasswordPolicy `json:"password_policy,omitempty"`
	SessionTimeout  uint            `json:"session_timeout"`
	RequireMFA      bool            `json:"require_mfa"`
	AllowedIPRanges []string        `json:"allowed_ip_ranges,omitempty"`
}

// PasswordPolicy describes password requirements
type PasswordPolicy struct {
	MinLength        uint `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
}

// IntegrationSettings describes third-party integration settings
type IntegrationSettings struct {
	WebhookEndpoints []string        `json:"webhook_endpoints,omitempty"`
	APIRateLimits    map[string]uint `json:"api_rate_limits,omitempty"`
}

// NotificationPreferences describes notification settings
type NotificationPreferences struct {
	EmailNotifications   bool     `json:"email_notifications"`
	WebhookNotifications bool     `json:"webhook_notifications"`
	NotificationTypes    []string `json:"notification_types,omitempty"`
}

// Usage analytics types

// UsageAnalyticsRequest describes analytics request
type UsageAnalyticsRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Period   string    `json:"period" validate:"required,oneof=current_month last_month last_3_months last_year"`
	Metrics  []string  `json:"metrics,omitempty" validate:"omitempty,dive,oneof=users storage api_calls transactions revenue"`
}

// TenantUsageAnalyticsResult describes usage analytics
type TenantUsageAnalyticsResult struct {
	TenantID         uuid.UUID            `json:"tenant_id"`
	Period           string               `json:"period"`
	UserMetrics      *UserMetrics         `json:"user_metrics,omitempty"`
	StorageMetrics   *StorageMetrics      `json:"storage_metrics,omitempty"`
	APIMetrics       *APIMetrics          `json:"api_metrics,omitempty"`
	FinancialMetrics *FinancialMetrics    `json:"financial_metrics,omitempty"`
	FeatureUsage     *FeatureUsageMetrics `json:"feature_usage,omitempty"`
}

// UserMetrics describes user activity metrics
type UserMetrics struct {
	TotalUsers         uint `json:"total_users"`
	ActiveUsers        uint `json:"active_users"`
	NewUsers           uint `json:"new_users"`
	AvgSessionDuration uint `json:"avg_session_duration"`
}

// StorageMetrics describes storage usage metrics
type StorageMetrics struct {
	TotalUsedMB    uint64  `json:"total_used_mb"`
	DocumentsCount uint    `json:"documents_count"`
	MediaUsedMB    uint64  `json:"media_used_mb"`
	GrowthRate     float64 `json:"growth_rate"`
}

// APIMetrics describes API usage metrics
type APIMetrics struct {
	TotalCalls      uint64  `json:"total_calls"`
	SuccessfulCalls uint64  `json:"successful_calls"`
	ErrorRate       float64 `json:"error_rate"`
	AvgResponseTime uint    `json:"avg_response_time"`
}

// FinancialMetrics describes financial activity metrics
type FinancialMetrics struct {
	TransactionsCount    uint  `json:"transactions_count"`
	TotalAmount          Money `json:"total_amount"`
	AvgTransactionAmount Money `json:"avg_transaction_amount"`
}

// FeatureUsageMetrics describes feature usage breakdown
type FeatureUsageMetrics struct {
	FinanceUsage   uint `json:"finance_usage"`
	InventoryUsage uint `json:"inventory_usage"`
	HRUsage        uint `json:"hr_usage"`
	ReportingUsage uint `json:"reporting_usage"`
}

// Money represents monetary values
type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// Bulk operations types

// BulkTenantOperationRequest describes bulk operation request
type BulkTenantOperationRequest struct {
	Operation  string         `json:"operation" validate:"required,oneof=suspend reactivate archive update_limits"`
	TenantIDs  []uuid.UUID    `json:"tenant_ids" validate:"required,min=1,max=100"`
	Parameters map[string]any `json:"parameters,omitempty"`
	ActorID    uuid.UUID      `json:"actor_id" validate:"required"`
}

// BulkOperationResult describes bulk operation results
type BulkOperationResult struct {
	OperationID uuid.UUID                 `json:"operation_id"`
	Status      string                    `json:"status"`
	Total       uint                      `json:"total"`
	Successful  uint                      `json:"successful"`
	Failed      uint                      `json:"failed"`
	Results     []BulkOperationItemResult `json:"results,omitempty"`
	StartedAt   time.Time                 `json:"started_at"`
	CompletedAt *time.Time                `json:"completed_at,omitempty"`
}

// BulkOperationItemResult describes individual item result
type BulkOperationItemResult struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Status   string    `json:"status"`
	Message  string    `json:"message,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// Audit logging types

// AuditLogRequest describes audit log query request
type AuditLogRequest struct {
	TenantID     uuid.UUID  `json:"tenant_id" validate:"required"`
	ActionFilter []string   `json:"action_filter,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	Page         uint       `json:"page" validate:"min=1"`
	PageSize     uint       `json:"page_size" validate:"min=1,max=100"`
}

// AuditLogResult describes audit log results
type AuditLogResult struct {
	Data       []TenantAuditLogEntry `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
}

// TenantAuditLogEntry describes audit log entry
type TenantAuditLogEntry struct {
	ID          uuid.UUID      `json:"id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Action      string         `json:"action"`
	ActorID     uuid.UUID      `json:"actor_id"`
	ActorName   string         `json:"actor_name,omitempty"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	IPAddress   string         `json:"ip_address,omitempty"`
}

// LogTenantActionRequest describes audit logging request
type LogTenantActionRequest struct {
	ID          uuid.UUID      `json:"id,omitempty"`
	TenantID    uuid.UUID      `json:"tenant_id" validate:"required"`
	Action      string         `json:"action" validate:"required"`
	ActorID     uuid.UUID      `json:"actor_id"`
	Description string         `json:"description" validate:"required"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	IPAddress   string         `json:"ip_address,omitempty"`
}

// Advanced listing types

// AdvancedListRequest describes advanced listing request
type AdvancedListRequest struct {
	StatusFilter    []string   `json:"status_filter,omitempty"`
	PlanFilter      []string   `json:"plan_filter,omitempty"`
	CreatedAfter    *time.Time `json:"created_after,omitempty"`
	CreatedBefore   *time.Time `json:"created_before,omitempty"`
	Search          string     `json:"search,omitempty"`
	IncludeArchived bool       `json:"include_archived"`
	Page            uint       `json:"page" validate:"min=1"`
	PageSize        uint       `json:"page_size" validate:"min=1,max=100"`
	SortBy          string     `json:"sort_by,omitempty"`
	SortOrder       string     `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// AdvancedListResult describes advanced listing results
type AdvancedListResult struct {
	Data       []DetailedTenantResult `json:"data"`
	Pagination PaginationMeta         `json:"pagination"`
	Summary    TenantSummaryStats     `json:"summary"`
}

// DetailedTenantResult extends basic tenant with administrative info
type DetailedTenantResult struct {
	Tenant              `json:",inline"`
	Configuration       *TenantConfigurationResult `json:"configuration,omitempty"`
	UsageStats          *TenantUsageStatsResult    `json:"usage_stats,omitempty"`
	SubscriptionDetails *SubscriptionDetails       `json:"subscription_details,omitempty"`
	BillingInfo         *BillingInfo               `json:"billing_info,omitempty"`
	SecuritySettings    *SecuritySettings          `json:"security_settings,omitempty"`
	AuditSummary        *AuditSummary              `json:"audit_summary,omitempty"`
}

// TenantUsageStatsResult describes current usage statistics
type TenantUsageStatsResult struct {
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
	ActiveUsers           uint      `json:"active_users"`
	TotalEntities         uint      `json:"total_entities"`
	StorageUsedMB         uint64    `json:"storage_used_mb"`
	APICalls              uint64    `json:"api_calls"`
	TransactionsProcessed uint64    `json:"transactions_processed"`
}

// Supporting types for detailed results
type SubscriptionDetails struct {
	SubscriptionID    string     `json:"subscription_id,omitempty"`
	PaymentMethod     string     `json:"payment_method,omitempty"`
	LastPaymentDate   *time.Time `json:"last_payment_date,omitempty"`
	LastPaymentAmount *Money     `json:"last_payment_amount,omitempty"`
	Plan              string     `json:"plan"`
	Status            string     `json:"status"`
	BillingCycle      string     `json:"billing_cycle"`
	NextBillingDate   *time.Time `json:"next_billing_date,omitempty"`
	TrialEndsAt       *time.Time `json:"trial_ends_at,omitempty"`
}

type BillingInfo struct {
	BillingAddress    *Address `json:"billing_address,omitempty"`
	PaymentStatus     string   `json:"payment_status"`
	OutstandingAmount *Money   `json:"outstanding_amount,omitempty"`
	CreditBalance     *Money   `json:"credit_balance,omitempty"`
}

type SecuritySettings struct {
	LastSecurityScan *time.Time `json:"last_security_scan,omitempty"`
	SecurityScore    uint       `json:"security_score"`
	EnabledFeatures  []string   `json:"enabled_features,omitempty"`
}

type AuditSummary struct {
	LastLogin     *time.Time `json:"last_login,omitempty"`
	RecentChanges uint       `json:"recent_changes"`
	LastBackup    *time.Time `json:"last_backup,omitempty"`
}

// TenantSummaryStats describes summary statistics
type TenantSummaryStats struct {
	TotalTenants      uint    `json:"total_tenants"`
	ActiveTenants     uint    `json:"active_tenants"`
	SuspendedTenants  uint    `json:"suspended_tenants"`
	PendingTenants    uint    `json:"pending_tenants"`
	ArchivedTenants   uint    `json:"archived_tenants"`
	TotalRevenue      *Money  `json:"total_revenue,omitempty"`
	AvgUsersPerTenant float64 `json:"avg_users_per_tenant"`
}

// PaginationMeta describes pagination metadata
type PaginationMeta struct {
	CurrentPage uint `json:"current_page"`
	PageSize    uint `json:"page_size"`
	TotalItems  uint `json:"total_items"`
	TotalPages  uint `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// ContactInfo describes contact information
type ContactInfo struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone,omitempty" validate:"omitempty,e164"`
	Title string `json:"title,omitempty" validate:"max=100"`
}

// Address describes physical address
type Address struct {
	StreetLine1   string   `json:"street_line_1" validate:"required,max=100"`
	StreetLine2   string   `json:"street_line_2,omitempty" validate:"max=100"`
	City          string   `json:"city" validate:"required,max=100"`
	StateProvince string   `json:"state_province,omitempty" validate:"max=100"`
	PostalCode    string   `json:"postal_code,omitempty" validate:"max=20"`
	Country       string   `json:"country" validate:"required,iso3166_1_alpha2"`
	Latitude      *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude     *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
}

// Service interfaces for dependencies

// AuditLogger defines audit logging interface
type AuditLogger interface {
	LogAction(ctx context.Context, req LogTenantActionRequest) error
}

// NotificationSender defines notification sending interface
type NotificationSender interface {
	SendWelcomeEmail(ctx context.Context, tenantID uuid.UUID, email, firstName string) error
	NotifyTenantUsers(ctx context.Context, tenantID uuid.UUID, eventType string, data map[string]any) error
}
