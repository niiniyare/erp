package authn

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MFA Method Constants
const (
	MFAMethodTOTP  = "totp"
	MFAMethodSMS   = "sms"
	MFAMethodEmail = "email"
)

// MFA Enrollment Status Constants
const (
	MFAEnrollmentStatusPending   = "pending"
	MFAEnrollmentStatusCompleted = "completed"
	MFAEnrollmentStatusExpired   = "expired"
	MFAEnrollmentStatusCancelled = "cancelled"
)

// MFA Request Types
// NOTE: These types define the request/response structure for MFA operations
// TODO: Add validation tags and JSON marshaling support for API integration

// BeginMFAEnrollmentRequest initiates MFA enrollment process
type BeginMFAEnrollmentRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Method      string    `json:"method" validate:"required,oneof=totp sms email"`
	PhoneNumber *string   `json:"phone_number,omitempty" validate:"omitempty,phone"`
	DeviceName  *string   `json:"device_name,omitempty"`
}

// CompleteMFAEnrollmentRequest finalizes MFA enrollment with verification
type CompleteMFAEnrollmentRequest struct {
	EnrollmentID     uuid.UUID `json:"enrollment_id" validate:"required"`
	VerificationCode string    `json:"verification_code" validate:"required"`
	DeviceName       *string   `json:"device_name,omitempty"`
}

// ValidateMFACodeRequest validates MFA code during authentication
type ValidateMFACodeRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Code   string    `json:"code" validate:"required"`
	Method *string   `json:"method,omitempty"`
}

// SendMFACodeRequest sends MFA code via SMS or Email
type SendMFACodeRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Method      string    `json:"method" validate:"required,oneof=sms email"`
	PhoneNumber *string   `json:"phone_number,omitempty"`
	Email       *string   `json:"email,omitempty"`
}

// DisableMFAMethodRequest disables a specific MFA method
type DisableMFAMethodRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	MethodID uuid.UUID `json:"method_id" validate:"required"`
	Reason   *string   `json:"reason,omitempty"`
}

// UseBackupCodeRequest validates a backup code
type UseBackupCodeRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Code   string    `json:"code" validate:"required"`
}

// UpdateMFASettingsRequest updates user MFA settings
type UpdateMFASettingsRequest struct {
	UserID          uuid.UUID  `json:"user_id" validate:"required"`
	PrimaryMethodID *uuid.UUID `json:"primary_method_id,omitempty"`
	RequiredMethods []string   `json:"required_methods,omitempty"`
	GracePeriodDays *int       `json:"grace_period_days,omitempty"`
}

// MFAStatisticsRequest requests MFA usage statistics
type MFAStatisticsRequest struct {
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	StartDate time.Time  `json:"start_date" validate:"required"`
	EndDate   time.Time  `json:"end_date" validate:"required"`
	Period    string     `json:"period" validate:"required,oneof=day week month"`
}

// MFA Response Types

// MFAEnrollmentResult contains enrollment initiation details
type MFAEnrollmentResult struct {
	EnrollmentID uuid.UUID `json:"enrollment_id"`
	Method       string    `json:"method"`
	Secret       string    `json:"secret,omitempty"`       // For TOTP
	QRCodeURL    string    `json:"qr_code_url,omitempty"`  // For TOTP
	PhoneNumber  *string   `json:"phone_number,omitempty"` // For SMS
	Email        *string   `json:"email,omitempty"`        // For Email
	BackupCodes  []string  `json:"backup_codes,omitempty"` // Generated backup codes
	ExpiresAt    time.Time `json:"expires_at"`
	Instructions string    `json:"instructions"`
}

// MFACompletionResult contains enrollment completion details
type MFACompletionResult struct {
	MethodID    uuid.UUID `json:"method_id"`
	Method      string    `json:"method"`
	BackupCodes []string  `json:"backup_codes,omitempty"`
	IsPrimary   bool      `json:"is_primary"`
	EnrolledAt  time.Time `json:"enrolled_at"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
}

// Note: MFAValidationResult is defined in service.go

// MFACodeSentResult contains code sending outcome
type MFACodeSentResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	SentAt    time.Time `json:"sent_at,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// BackupCodeResult contains backup code validation result
type BackupCodeResult struct {
	Valid   bool      `json:"valid"`
	UsedAt  time.Time `json:"used_at,omitempty"`
	Message string    `json:"message"`
}

// BackupCodesResult contains generated backup codes
type BackupCodesResult struct {
	UserID      uuid.UUID `json:"user_id"`
	Codes       []string  `json:"codes"`
	GeneratedAt time.Time `json:"generated_at"`
	Count       int       `json:"count"`
}

// MFASettings contains user MFA configuration
type MFASettings struct {
	UserID               uuid.UUID    `json:"user_id"`
	MFAEnabled           bool         `json:"mfa_enabled"`
	EnrolledMethods      []*MFAMethod `json:"enrolled_methods"`
	PrimaryMethodID      *uuid.UUID   `json:"primary_method_id,omitempty"`
	BackupCodesRemaining int          `json:"backup_codes_remaining"`
	LastValidationAt     *time.Time   `json:"last_validation_at,omitempty"`
	RequiredForAccount   bool         `json:"required_for_account"`
	GracePeriodEndsAt    *time.Time   `json:"grace_period_ends_at,omitempty"`
}

// MFAStatistics contains MFA usage analytics
type MFAStatistics struct {
	TotalUsers            int            `json:"total_users"`
	MFAEnabledUsers       int            `json:"mfa_enabled_users"`
	MethodDistribution    map[string]int `json:"method_distribution"`
	ValidationSuccessRate float64        `json:"validation_success_rate"`
	Period                string         `json:"period"`
	GeneratedAt           time.Time      `json:"generated_at"`
}

// MFA Data Models

// MFAMethod represents an enrolled MFA method for a user
// NOTE: This model supports multiple MFA methods per user with activation states
// TODO: Add metadata field for method-specific configuration (rate limits, etc.)
type MFAMethod struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	UserID       uuid.UUID      `json:"user_id" db:"user_id"`
	Method       string         `json:"method" db:"method"`
	Secret       string         `json:"-" db:"secret"` // Never expose in JSON
	PhoneNumber  *string        `json:"phone_number,omitempty" db:"phone_number"`
	Email        *string        `json:"email,omitempty" db:"email"`
	IsActive     bool           `json:"is_active" db:"is_active"`
	IsPrimary    bool           `json:"is_primary" db:"is_primary"`
	EnrolledAt   time.Time      `json:"enrolled_at" db:"enrolled_at"`
	LastUsedAt   *time.Time     `json:"last_used_at,omitempty" db:"last_used_at"`
	DisabledAt   *time.Time     `json:"disabled_at,omitempty" db:"disabled_at"`
	FailureCount int            `json:"failure_count" db:"failure_count"`
	Metadata     map[string]any `json:"metadata,omitempty" db:"metadata"`
}

// MFAEnrollmentData represents temporary enrollment data
// NOTE: This data is stored temporarily during enrollment process and cleaned up after completion
// TODO: Add rate limiting metadata to prevent enrollment abuse
type MFAEnrollmentData struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	UserID      uuid.UUID      `json:"user_id" db:"user_id"`
	Method      string         `json:"method" db:"method"`
	Secret      string         `json:"-" db:"secret"` // Never expose in JSON
	PhoneNumber *string        `json:"phone_number,omitempty" db:"phone_number"`
	Email       *string        `json:"email,omitempty" db:"email"`
	Status      string         `json:"status" db:"status"`
	ExpiresAt   time.Time      `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	Metadata    map[string]any `json:"metadata,omitempty" db:"metadata"`
}

// BackupCode represents a recovery code for MFA bypass
// NOTE: Backup codes are single-use and provide emergency access when MFA devices are unavailable
// TODO: Add expiration policy for backup codes (e.g., expire after 1 year)
type BackupCode struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Code      string     `json:"code" db:"code"`
	Used      bool       `json:"used" db:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// Notification Service Types (for MFA integration)

// NotificationService defines notification capabilities needed for MFA
// NOTE: This interface extends the existing notification service for MFA-specific needs
// TODO: Add template-based messaging with localization support
type NotificationService interface {
	SendSMS(ctx context.Context, req *SMSRequest) error
	SendEmail(ctx context.Context, req *EmailRequest) error
}

// SMSRequest contains SMS message parameters
type SMSRequest struct {
	PhoneNumber string            `json:"phone_number" validate:"required,phone"`
	Message     string            `json:"message" validate:"required,max=160"`
	Template    *string           `json:"template,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

// EmailRequest contains email message parameters
type EmailRequest struct {
	To        string            `json:"to" validate:"required,email"`
	Subject   string            `json:"subject" validate:"required"`
	Body      string            `json:"body" validate:"required"`
	Template  *string           `json:"template,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}
