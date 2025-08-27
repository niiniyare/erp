package authn

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MFARepository defines the interface for MFA data persistence operations
// NOTE: This repository handles all MFA-related database operations with multi-tenant support
// TODO: Add bulk operations support for enterprise-scale MFA management
type MFARepository interface {
	// MFA Method Management
	// NOTE: These methods manage enrolled MFA devices/methods for users
	SaveMFAMethod(ctx context.Context, method *MFAMethod) error
	GetMFAMethod(ctx context.Context, methodID uuid.UUID) (*MFAMethod, error)
	GetUserMFAMethods(ctx context.Context, userID uuid.UUID) ([]*MFAMethod, error)
	UpdateMFAMethod(ctx context.Context, method *MFAMethod) error
	DeleteMFAMethod(ctx context.Context, methodID uuid.UUID) error
	
	// MFA Method Status Operations
	// NOTE: These operations handle activation/deactivation of MFA methods
	ActivateMFAMethod(ctx context.Context, methodID uuid.UUID) error
	DeactivateMFAMethod(ctx context.Context, methodID uuid.UUID) error
	SetPrimaryMFAMethod(ctx context.Context, userID, methodID uuid.UUID) error
	
	// Enrollment Data Management
	// NOTE: Enrollment data is temporary and cleaned up after enrollment completion/expiration
	SaveEnrollmentData(ctx context.Context, data *MFAEnrollmentData) error
	GetEnrollmentData(ctx context.Context, enrollmentID uuid.UUID) (*MFAEnrollmentData, error)
	GetUserPendingEnrollments(ctx context.Context, userID uuid.UUID) ([]*MFAEnrollmentData, error)
	UpdateEnrollmentData(ctx context.Context, data *MFAEnrollmentData) error
	DeleteEnrollmentData(ctx context.Context, enrollmentID uuid.UUID) error
	
	// Enrollment Cleanup Operations
	// NOTE: These operations handle cleanup of expired or completed enrollments
	CleanupExpiredEnrollments(ctx context.Context) (int, error)
	DeleteUserEnrollments(ctx context.Context, userID uuid.UUID) error
	
	// Backup Code Management
	// NOTE: Backup codes provide emergency access when MFA devices are unavailable
	SaveBackupCodes(ctx context.Context, codes []*BackupCode) error
	GetBackupCodes(ctx context.Context, userID uuid.UUID) ([]*BackupCode, error)
	UpdateBackupCode(ctx context.Context, code *BackupCode) error
	InvalidateBackupCodes(ctx context.Context, userID uuid.UUID) error
	GetBackupCode(ctx context.Context, userID uuid.UUID, code string) (*BackupCode, error)
	
	// MFA Statistics and Analytics
	// NOTE: These operations provide insights into MFA usage and security metrics
	GetMFAMethodCount(ctx context.Context) (map[string]int, error)
	GetUserMFAStatus(ctx context.Context, userID uuid.UUID) (*MFAUserStatus, error)
	GetMFAUsageStats(ctx context.Context, req *MFAUsageStatsRequest) (*MFAUsageStats, error)
	
	// Security and Audit Operations
	// NOTE: These operations track MFA security events and failed attempts
	RecordMFAAttempt(ctx context.Context, attempt *MFAAttempt) error
	GetMFAAttempts(ctx context.Context, userID uuid.UUID, since time.Time) ([]*MFAAttempt, error)
	CleanupOldMFAAttempts(ctx context.Context, olderThan time.Time) (int, error)
}

// Supporting Types for Repository Operations

// MFAUserStatus represents a user's overall MFA configuration status
type MFAUserStatus struct {
	UserID              uuid.UUID    `json:"user_id"`
	HasMFAEnabled       bool         `json:"has_mfa_enabled"`
	ActiveMethodCount   int          `json:"active_method_count"`
	PrimaryMethod       *MFAMethod   `json:"primary_method,omitempty"`
	LastValidationAt    *time.Time   `json:"last_validation_at,omitempty"`
	BackupCodesCount    int          `json:"backup_codes_count"`
	PendingEnrollments  int          `json:"pending_enrollments"`
	RecentFailureCount  int          `json:"recent_failure_count"`
}

// MFAUsageStatsRequest contains parameters for usage statistics queries
type MFAUsageStatsRequest struct {
	TenantID      *uuid.UUID `json:"tenant_id,omitempty"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	GroupBy       string     `json:"group_by"` // "day", "week", "month"
	IncludeUsers  []uuid.UUID `json:"include_users,omitempty"`
	ExcludeUsers  []uuid.UUID `json:"exclude_users,omitempty"`
	MethodFilter  []string   `json:"method_filter,omitempty"`
}

// MFAUsageStats contains aggregated MFA usage statistics
type MFAUsageStats struct {
	Period                string                    `json:"period"`
	TotalUsers            int                       `json:"total_users"`
	MFAEnabledUsers       int                       `json:"mfa_enabled_users"`
	MethodDistribution    map[string]int           `json:"method_distribution"`
	SuccessfulValidations int                       `json:"successful_validations"`
	FailedValidations     int                       `json:"failed_validations"`
	SuccessRate           float64                   `json:"success_rate"`
	NewEnrollments        int                       `json:"new_enrollments"`
	MethodUsageByDay      map[string]map[string]int `json:"method_usage_by_day"`
	TopFailureReasons     map[string]int           `json:"top_failure_reasons"`
	GeneratedAt           time.Time                 `json:"generated_at"`
}

// MFAAttempt represents an MFA validation attempt (for security monitoring)
// NOTE: This tracks both successful and failed MFA attempts for security analysis
// TODO: Add geolocation and device fingerprinting data for enhanced security
type MFAAttempt struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	MethodID    *uuid.UUID `json:"method_id,omitempty" db:"method_id"`
	Method      string     `json:"method" db:"method"`
	Success     bool       `json:"success" db:"success"`
	FailureReason *string  `json:"failure_reason,omitempty" db:"failure_reason"`
	IPAddress   *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string    `json:"user_agent,omitempty" db:"user_agent"`
	AttemptedAt time.Time  `json:"attempted_at" db:"attempted_at"`
	Metadata    map[string]any `json:"metadata,omitempty" db:"metadata"`
}