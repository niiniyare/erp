package authn

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"go.opentelemetry.io/otel/attribute"

	"awo/internal/core/iam/model"
	"awo/internal/core/iam/repo"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// MFAService handles Multi-Factor Authentication operations
// NOTE: This service provides MFA support including TOTP, SMS, and Email verification
// TODO: Add hardware token support (YubiKey, FIDO2) for enhanced security
type MFAService interface {
	// Enrollment flows
	BeginMFAEnrollment(ctx context.Context, req *BeginMFAEnrollmentRequest) (*MFAEnrollmentResult, error)
	CompleteMFAEnrollment(ctx context.Context, req *CompleteMFAEnrollmentRequest) (*MFACompletionResult, error)
	CancelMFAEnrollment(ctx context.Context, userID uuid.UUID) error

	// Verification flows
	ValidateMFACode(ctx context.Context, req *ValidateMFACodeRequest) (*MFAValidationResult, error)
	SendMFACode(ctx context.Context, req *SendMFACodeRequest) (*MFACodeSentResult, error)

	// Management operations
	ListUserMFAMethods(ctx context.Context, userID uuid.UUID) ([]*MFAMethod, error)
	DisableMFAMethod(ctx context.Context, req *DisableMFAMethodRequest) error
	GenerateBackupCodes(ctx context.Context, userID uuid.UUID) (*BackupCodesResult, error)
	UseBackupCode(ctx context.Context, req *UseBackupCodeRequest) (*BackupCodeResult, error)

	// Administrative operations
	GetMFASettings(ctx context.Context, userID uuid.UUID) (*MFASettings, error)
	UpdateMFASettings(ctx context.Context, req *UpdateMFASettingsRequest) (*MFASettings, error)
	GetMFAStatistics(ctx context.Context, req *MFAStatisticsRequest) (*MFAStatistics, error)
}

// mfaService implements MFAService with multi-factor authentication support
type mfaService struct {
	userRepo        repo.UserRepository
	sessionRepo     repo.SessionRepository
	mfaRepo         MFARepository
	notificationSvc NotificationService

	// Security settings
	issuerName      string
	accountName     string
	secretLength    int
	backupCodeCount int

	// Infrastructure
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewMFAService creates a new MFA service instance with authentication support
func NewMFAService(
	userRepo repo.UserRepository,
	sessionRepo repo.SessionRepository,
	mfaRepo MFARepository,
	notificationSvc NotificationService,
	issuerName, accountName string,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) MFAService {
	return &mfaService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		mfaRepo:         mfaRepo,
		notificationSvc: notificationSvc,
		issuerName:      issuerName,
		accountName:     accountName,
		secretLength:    32, // 32 bytes for strong security
		backupCodeCount: 10, // Standard number of backup codes
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
	}
}

// BeginMFAEnrollment starts the MFA enrollment process for a user
func (s *mfaService) BeginMFAEnrollment(ctx context.Context, req *BeginMFAEnrollmentRequest) (*MFAEnrollmentResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "mfa.begin_enrollment",
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
			attribute.String("method", req.Method),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Beginning MFA enrollment",
		logger.Fields{
			"user_id": req.UserID,
			"method":  req.Method,
		})

	// Validate user exists and is active
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		s.tracer.RecordError(ctx, err)
		return nil, errors.NewBusinessError("USER_NOT_FOUND", "User not found for MFA enrollment")
	}

	if !user.IsActive() {
		return nil, errors.NewBusinessError("USER_INACTIVE", "Cannot enroll MFA for inactive user")
	}

	// Check if method is already enrolled
	existingMethods, err := s.mfaRepo.GetUserMFAMethods(ctx, req.UserID)
	if err != nil {
		s.tracer.RecordError(ctx, err)
		return nil, errors.NewBusinessError("MFA_METHOD_CHECK_FAILED", "Failed to check existing MFA methods")
	}

	for _, method := range existingMethods {
		if method.Method == req.Method && method.IsActive {
			return nil, errors.NewBusinessError("MFA_METHOD_ALREADY_ENROLLED",
				fmt.Sprintf("MFA method %s already enrolled", req.Method))
		}
	}

	// Create enrollment based on method type
	switch req.Method {
	case MFAMethodTOTP:
		return s.beginTOTPEnrollment(ctx, req)
	case MFAMethodSMS:
		return s.beginSMSEnrollment(ctx, req)
	case MFAMethodEmail:
		return s.beginEmailEnrollment(ctx, req)
	default:
		return nil, errors.NewBusinessError("INVALID_MFA_METHOD",
			fmt.Sprintf("Unsupported MFA method: %s", req.Method))
	}
}

// beginTOTPEnrollment handles TOTP (Time-based One-Time Password) enrollment
func (s *mfaService) beginTOTPEnrollment(ctx context.Context, req *BeginMFAEnrollmentRequest) (*MFAEnrollmentResult, error) {
	// Generate secure secret
	secret := make([]byte, s.secretLength)
	if _, err := rand.Read(secret); err != nil {
		return nil, errors.NewBusinessError("SECRET_GENERATION_FAILED", "Failed to generate TOTP secret")
	}

	secretString := base32.StdEncoding.EncodeToString(secret)

	// Get user for account name
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuerName,
		AccountName: user.Email,
		SecretSize:  uint(s.secretLength),
		Secret:      secret,
	})
	if err != nil {
		return nil, errors.NewBusinessError("TOTP_GENERATION_FAILED", "Failed to generate TOTP key")
	}

	// Store enrollment data temporarily
	enrollmentData := &MFAEnrollmentData{
		UserID:    req.UserID,
		Method:    MFAMethodTOTP,
		Secret:    secretString,
		Status:    MFAEnrollmentStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minute enrollment window
		CreatedAt: time.Now(),
		Metadata: map[string]any{
			"issuer":       s.issuerName,
			"account_name": user.Email,
		},
	}

	if err := s.mfaRepo.SaveEnrollmentData(ctx, enrollmentData); err != nil {
		return nil, errors.NewBusinessError("ENROLLMENT_DATA_SAVE_FAILED", "Failed to save enrollment data")
	}

	// Generate backup codes for TOTP enrollment
	backupCodes, err := s.generateBackupCodesInternal(ctx, req.UserID, 10)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to generate backup codes during TOTP enrollment",
			logger.Fields{"error": err.Error()})
	}

	s.metrics.IncrementCounter("mfa_enrollment_started",
		metrics.Fields{"method": MFAMethodTOTP})

	return &MFAEnrollmentResult{
		EnrollmentID: enrollmentData.ID,
		Method:       MFAMethodTOTP,
		Secret:       secretString,
		QRCodeURL:    key.URL(),
		BackupCodes:  backupCodes,
		ExpiresAt:    enrollmentData.ExpiresAt,
		Instructions: "Scan the QR code with your authenticator app and enter the verification code to complete enrollment",
	}, nil
}

// beginSMSEnrollment handles SMS-based MFA enrollment
func (s *mfaService) beginSMSEnrollment(ctx context.Context, req *BeginMFAEnrollmentRequest) (*MFAEnrollmentResult, error) {
	if req.PhoneNumber == nil || *req.PhoneNumber == "" {
		return nil, errors.NewBusinessError("PHONE_NUMBER_REQUIRED", "Phone number is required for SMS MFA")
	}

	// Validate phone number format
	if !s.isValidPhoneNumber(*req.PhoneNumber) {
		return nil, errors.NewBusinessError("INVALID_PHONE_NUMBER", "Invalid phone number format")
	}

	// Generate verification code
	verificationCode, err := s.generateVerificationCode(6)
	if err != nil {
		return nil, errors.NewBusinessError("CODE_GENERATION_FAILED", "Failed to generate verification code")
	}

	// Store enrollment data
	enrollmentData := &MFAEnrollmentData{
		UserID:      req.UserID,
		Method:      MFAMethodSMS,
		PhoneNumber: req.PhoneNumber,
		Status:      MFAEnrollmentStatusPending,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minute verification window
		CreatedAt:   time.Now(),
		Metadata: map[string]any{
			"verification_code": s.hashCode(verificationCode),
			"attempts":          0,
			"max_attempts":      3,
		},
	}

	if err := s.mfaRepo.SaveEnrollmentData(ctx, enrollmentData); err != nil {
		return nil, errors.NewBusinessError("ENROLLMENT_DATA_SAVE_FAILED", "Failed to save enrollment data")
	}

	// Send SMS verification code
	if err := s.notificationSvc.SendSMS(ctx, &SMSRequest{
		PhoneNumber: *req.PhoneNumber,
		Message:     fmt.Sprintf("Your %s verification code is: %s", s.issuerName, verificationCode),
	}); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send SMS verification code",
			logger.Fields{"error": err.Error()})
		return nil, errors.NewBusinessError("SMS_SEND_FAILED", "Failed to send verification code")
	}

	s.metrics.IncrementCounter("mfa_enrollment_started",
		metrics.Fields{"method": MFAMethodSMS})

	return &MFAEnrollmentResult{
		EnrollmentID: enrollmentData.ID,
		Method:       MFAMethodSMS,
		PhoneNumber:  req.PhoneNumber,
		ExpiresAt:    enrollmentData.ExpiresAt,
		Instructions: fmt.Sprintf("Enter the 6-digit code sent to %s", s.maskPhoneNumber(*req.PhoneNumber)),
	}, nil
}

// beginEmailEnrollment handles email-based MFA enrollment
func (s *mfaService) beginEmailEnrollment(ctx context.Context, req *BeginMFAEnrollmentRequest) (*MFAEnrollmentResult, error) {
	// Get user's email
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Generate verification code
	verificationCode, err := s.generateVerificationCode(8)
	if err != nil {
		return nil, errors.NewBusinessError("CODE_GENERATION_FAILED", "Failed to generate verification code")
	}

	// Store enrollment data
	enrollmentData := &MFAEnrollmentData{
		UserID:    req.UserID,
		Method:    MFAMethodEmail,
		Email:     &user.Email,
		Status:    MFAEnrollmentStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minute verification window
		CreatedAt: time.Now(),
		Metadata: map[string]any{
			"verification_code": s.hashCode(verificationCode),
			"attempts":          0,
			"max_attempts":      3,
		},
	}

	if err := s.mfaRepo.SaveEnrollmentData(ctx, enrollmentData); err != nil {
		return nil, errors.NewBusinessError("ENROLLMENT_DATA_SAVE_FAILED", "Failed to save enrollment data")
	}

	// Send email verification code
	if err := s.notificationSvc.SendEmail(ctx, &EmailRequest{
		To:      user.Email,
		Subject: fmt.Sprintf("%s - MFA Setup Verification", s.issuerName),
		Body:    fmt.Sprintf("Your MFA setup verification code is: %s\n\nThis code expires in 15 minutes.", verificationCode),
	}); err != nil {
		s.logger.ErrorContext(ctx, "Failed to send email verification code",
			logger.Fields{"error": err.Error()})
		return nil, errors.NewBusinessError("EMAIL_SEND_FAILED", "Failed to send verification code")
	}

	s.metrics.IncrementCounter("mfa_enrollment_started",
		metrics.Fields{"method": MFAMethodEmail})

	return &MFAEnrollmentResult{
		EnrollmentID: enrollmentData.ID,
		Method:       MFAMethodEmail,
		Email:        &user.Email,
		ExpiresAt:    enrollmentData.ExpiresAt,
		Instructions: fmt.Sprintf("Enter the 8-digit code sent to %s", s.maskEmail(user.Email)),
	}, nil
}

// CompleteMFAEnrollment completes the MFA enrollment process
func (s *mfaService) CompleteMFAEnrollment(ctx context.Context, req *CompleteMFAEnrollmentRequest) (*MFACompletionResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "mfa.complete_enrollment",
		tracing.WithAttributes(
			attribute.String("enrollment_id", req.EnrollmentID.String()),
		))
	defer span.End()

	s.logger.InfoContext(ctx, "Completing MFA enrollment",
		logger.Fields{
			"enrollment_id": req.EnrollmentID,
		})

	// Get enrollment data
	enrollmentData, err := s.mfaRepo.GetEnrollmentData(ctx, req.EnrollmentID)
	if err != nil {
		return nil, errors.NewBusinessError("ENROLLMENT_NOT_FOUND", "MFA enrollment not found")
	}

	// Check enrollment status and expiration
	if enrollmentData.Status != MFAEnrollmentStatusPending {
		return nil, errors.NewBusinessError("ENROLLMENT_NOT_PENDING", "MFA enrollment is not in pending status")
	}

	if time.Now().After(enrollmentData.ExpiresAt) {
		// Clean up expired enrollment
		s.mfaRepo.DeleteEnrollmentData(ctx, req.EnrollmentID)
		return nil, errors.NewBusinessError("ENROLLMENT_EXPIRED", "MFA enrollment has expired")
	}

	// Verify the code based on method
	valid := false
	switch enrollmentData.Method {
	case MFAMethodTOTP:
		valid, err = s.verifyTOTPCode(enrollmentData.Secret, req.VerificationCode)
	case MFAMethodSMS, MFAMethodEmail:
		valid, err = s.verifyStoredCode(enrollmentData, req.VerificationCode)
	default:
		return nil, errors.NewBusinessError("INVALID_MFA_METHOD", "Invalid MFA method")
	}

	if err != nil {
		return nil, err
	}

	if !valid {
		// Increment attempt count
		attempts := enrollmentData.Metadata["attempts"].(int) + 1
		maxAttempts := enrollmentData.Metadata["max_attempts"].(int)

		if attempts >= maxAttempts {
			// Too many failed attempts, cancel enrollment
			s.mfaRepo.DeleteEnrollmentData(ctx, req.EnrollmentID)
			s.metrics.IncrementCounter("mfa_enrollment_failed",
				metrics.Fields{"method": enrollmentData.Method, "reason": "max_attempts"})
			return nil, errors.NewBusinessError("MAX_ATTEMPTS_EXCEEDED", "Maximum verification attempts exceeded")
		}

		// Update attempt count
		enrollmentData.Metadata["attempts"] = attempts
		s.mfaRepo.UpdateEnrollmentData(ctx, enrollmentData)

		s.metrics.IncrementCounter("mfa_enrollment_invalid_code",
			metrics.Fields{"method": enrollmentData.Method})
		return nil, errors.NewBusinessError("INVALID_VERIFICATION_CODE", "Invalid verification code")
	}

	// Create MFA method record
	mfaMethod := &MFAMethod{
		ID:           uuid.New(),
		UserID:       enrollmentData.UserID,
		Method:       enrollmentData.Method,
		Secret:       enrollmentData.Secret,
		PhoneNumber:  enrollmentData.PhoneNumber,
		Email:        enrollmentData.Email,
		IsActive:     true,
		IsPrimary:    false, // User can set this later
		EnrolledAt:   time.Now(),
		LastUsedAt:   nil,
		FailureCount: 0,
		Metadata:     enrollmentData.Metadata,
	}

	// Check if this is the first MFA method (make it primary)
	existingMethods, err := s.mfaRepo.GetUserMFAMethods(ctx, enrollmentData.UserID)
	if err != nil {
		return nil, err
	}

	if len(existingMethods) == 0 {
		mfaMethod.IsPrimary = true
	}

	// Save MFA method
	if err := s.mfaRepo.SaveMFAMethod(ctx, mfaMethod); err != nil {
		return nil, errors.NewBusinessError("MFA_METHOD_SAVE_FAILED", "Failed to save MFA method")
	}

	// Generate backup codes if not already generated
	var backupCodes []string
	if enrollmentData.Method == MFAMethodTOTP {
		// Check if backup codes already exist
		existingCodes, err := s.mfaRepo.GetBackupCodes(ctx, enrollmentData.UserID)
		if err == nil && len(existingCodes) > 0 {
			// Use existing backup codes
			backupCodes = make([]string, len(existingCodes))
			for i, code := range existingCodes {
				backupCodes[i] = code.Code
			}
		} else {
			// Generate new backup codes
			backupCodes, err = s.generateBackupCodesInternal(ctx, enrollmentData.UserID, s.backupCodeCount)
			if err != nil {
				s.logger.WarnContext(ctx, "Failed to generate backup codes",
					logger.Fields{"error": err.Error()})
			}
		}
	}

	// Clean up enrollment data
	if err := s.mfaRepo.DeleteEnrollmentData(ctx, req.EnrollmentID); err != nil {
		s.logger.WarnContext(ctx, "Failed to clean up enrollment data",
			logger.Fields{"error": err.Error()})
	}

	// Update user MFA status
	if err := s.userRepo.EnableMFA(ctx, enrollmentData.UserID, model.MFAMethod(enrollmentData.Method)); err != nil {
		s.logger.WarnContext(ctx, "Failed to update user MFA status",
			logger.Fields{"error": err.Error()})
	}

	s.metrics.IncrementCounter("mfa_enrollment_completed",
		metrics.Fields{"method": enrollmentData.Method})

	s.logger.InfoContext(ctx, "MFA enrollment completed successfully",
		logger.Fields{
			"user_id": enrollmentData.UserID,
			"method":  enrollmentData.Method,
		})

	return &MFACompletionResult{
		MethodID:    mfaMethod.ID,
		Method:      mfaMethod.Method,
		BackupCodes: backupCodes,
		IsPrimary:   mfaMethod.IsPrimary,
		EnrolledAt:  mfaMethod.EnrolledAt,
		Success:     true,
		Message:     fmt.Sprintf("MFA method %s enrolled successfully", mfaMethod.Method),
	}, nil
}

// CancelMFAEnrollment cancels an ongoing MFA enrollment
func (s *mfaService) CancelMFAEnrollment(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.tracer.StartSpan(ctx, "mfa.cancel_enrollment",
		tracing.WithAttributes(
			attribute.String("user_id", userID.String()),
		))
	defer span.End()

	// Find and delete any pending enrollments for the user
	enrollments, err := s.mfaRepo.GetUserPendingEnrollments(ctx, userID)
	if err != nil {
		return err
	}

	for _, enrollment := range enrollments {
		if err := s.mfaRepo.DeleteEnrollmentData(ctx, enrollment.ID); err != nil {
			s.logger.WarnContext(ctx, "Failed to delete enrollment data",
				logger.Fields{"enrollment_id": enrollment.ID, "error": err.Error()})
		}
	}

	s.metrics.IncrementCounter("mfa_enrollment_cancelled",
		metrics.Fields{"user_id": userID.String()})

	return nil
}

// ValidateMFACode validates an MFA code for authentication
func (s *mfaService) ValidateMFACode(ctx context.Context, req *ValidateMFACodeRequest) (*MFAValidationResult, error) {
	ctx, span := s.tracer.StartSpan(ctx, "mfa.validate_code",
		tracing.WithAttributes(
			attribute.String("user_id", req.UserID.String()),
		))
	defer span.End()

	// Get user's MFA methods
	methods, err := s.mfaRepo.GetUserMFAMethods(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if len(methods) == 0 {
		return &MFAValidationResult{
			Valid:       false,
			Method:      "",
			ValidatedAt: time.Time{},
			Error:       "No MFA methods enrolled",
		}, nil
	}

	// Try validation with each method
	for _, method := range methods {
		if !method.IsActive {
			continue
		}

		valid, err := s.validateCodeForMethod(ctx, method, req.Code)
		if err != nil {
			s.logger.WarnContext(ctx, "Error validating MFA code",
				logger.Fields{"method": method.Method, "error": err.Error()})
			continue
		}

		if valid {
			// Update last used timestamp
			method.LastUsedAt = &[]time.Time{time.Now()}[0]
			method.FailureCount = 0
			s.mfaRepo.UpdateMFAMethod(ctx, method)

			s.metrics.IncrementCounter("mfa_validation_success",
				metrics.Fields{"method": method.Method})

			return &MFAValidationResult{
				Valid:       true,
				Method:      method.Method,
				ValidatedAt: time.Now(),
			}, nil
		} else {
			// Increment failure count
			method.FailureCount++
			s.mfaRepo.UpdateMFAMethod(ctx, method)
		}
	}

	// Try backup codes if regular methods fail
	if s.validateBackupCode(ctx, req.UserID, req.Code) {
		s.metrics.IncrementCounter("mfa_validation_success",
			metrics.Fields{"method": "backup_code"})

		return &MFAValidationResult{
			Valid:       true,
			Method:      "backup_code",
			ValidatedAt: time.Now(),
		}, nil
	}

	s.metrics.IncrementCounter("mfa_validation_failed",
		metrics.Fields{"user_id": req.UserID.String()})

	return &MFAValidationResult{
		Valid:       false,
		Method:      "",
		ValidatedAt: time.Time{},
		Error:       "Invalid MFA code",
	}, nil
}

// Helper methods for code validation
func (s *mfaService) validateCodeForMethod(ctx context.Context, method *MFAMethod, code string) (bool, error) {
	switch method.Method {
	case MFAMethodTOTP:
		return s.verifyTOTPCode(method.Secret, code)
	case MFAMethodSMS, MFAMethodEmail:
		// For SMS/Email, we need to check if there's a pending validation
		return s.validatePendingCode(ctx, method, code)
	default:
		return false, errors.NewBusinessError("INVALID_MFA_METHOD", "Invalid MFA method")
	}
}

func (s *mfaService) verifyTOTPCode(secret, code string) (bool, error) {
	return totp.Validate(code, secret), nil
}

func (s *mfaService) verifyStoredCode(enrollmentData *MFAEnrollmentData, code string) (bool, error) {
	storedHash, ok := enrollmentData.Metadata["verification_code"].(string)
	if !ok {
		return false, errors.NewBusinessError("INVALID_ENROLLMENT_DATA", "Invalid enrollment data")
	}

	return s.hashCode(code) == storedHash, nil
}

func (s *mfaService) validatePendingCode(ctx context.Context, method *MFAMethod, code string) (bool, error) {
	// This would check for recently sent codes in a temporary storage
	// For now, return false as this requires additional implementation
	// TODO: Implement pending code validation for SMS/Email MFA
	return false, nil
}

func (s *mfaService) validateBackupCode(ctx context.Context, userID uuid.UUID, code string) bool {
	backupCodes, err := s.mfaRepo.GetBackupCodes(ctx, userID)
	if err != nil {
		return false
	}

	for _, backupCode := range backupCodes {
		if !backupCode.Used && backupCode.Code == code {
			// Mark backup code as used
			backupCode.Used = true
			backupCode.UsedAt = &[]time.Time{time.Now()}[0]
			s.mfaRepo.UpdateBackupCode(ctx, backupCode)
			return true
		}
	}

	return false
}

// Helper methods
func (s *mfaService) generateVerificationCode(length int) (string, error) {
	const charset = "0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

func (s *mfaService) generateBackupCodesInternal(ctx context.Context, userID uuid.UUID, count int) ([]string, error) {
	var codes []string
	var backupCodes []*BackupCode

	for i := 0; i < count; i++ {
		code, err := s.generateVerificationCode(8)
		if err != nil {
			return nil, err
		}

		backupCode := &BackupCode{
			ID:        uuid.New(),
			UserID:    userID,
			Code:      code,
			Used:      false,
			CreatedAt: time.Now(),
		}

		backupCodes = append(backupCodes, backupCode)
		codes = append(codes, code)
	}

	if err := s.mfaRepo.SaveBackupCodes(ctx, backupCodes); err != nil {
		return nil, err
	}

	return codes, nil
}

func (s *mfaService) hashCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func (s *mfaService) isValidPhoneNumber(phone string) bool {
	// Basic phone number validation - in production, use a proper phone validation library
	// TODO: Integrate with libphonenumber for phone validation
	return len(phone) >= 10 && len(phone) <= 15
}

func (s *mfaService) maskPhoneNumber(phone string) string {
	if len(phone) < 4 {
		return "***"
	}
	return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
}

func (s *mfaService) maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***.com"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		username = strings.Repeat("*", len(username))
	} else {
		username = username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]
	}

	return username + "@" + domain
}

// Additional MFA service methods (ListUserMFAMethods, DisableMFAMethod, etc.)
// NOTE: These methods provide MFA management capabilities
// TODO: Add rate limiting for MFA validation attempts
// TODO: Implement geolocation-based MFA requirements

// ListUserMFAMethods returns all MFA methods for a user
func (s *mfaService) ListUserMFAMethods(ctx context.Context, userID uuid.UUID) ([]*MFAMethod, error) {
	return s.mfaRepo.GetUserMFAMethods(ctx, userID)
}

// DisableMFAMethod disables a specific MFA method for a user
func (s *mfaService) DisableMFAMethod(ctx context.Context, req *DisableMFAMethodRequest) error {
	method, err := s.mfaRepo.GetMFAMethod(ctx, req.MethodID)
	if err != nil {
		return err
	}

	if method.UserID != req.UserID {
		return errors.NewBusinessError("UNAUTHORIZED", "Cannot disable MFA method for different user")
	}

	method.IsActive = false
	method.DisabledAt = &[]time.Time{time.Now()}[0]

	return s.mfaRepo.UpdateMFAMethod(ctx, method)
}

// GenerateBackupCodes generates new backup codes for a user
func (s *mfaService) GenerateBackupCodes(ctx context.Context, userID uuid.UUID) (*BackupCodesResult, error) {
	// Invalidate existing backup codes
	if err := s.mfaRepo.InvalidateBackupCodes(ctx, userID); err != nil {
		return nil, err
	}

	// Generate new codes
	codes, err := s.generateBackupCodesInternal(ctx, userID, s.backupCodeCount)
	if err != nil {
		return nil, err
	}

	return &BackupCodesResult{
		UserID:      userID,
		Codes:       codes,
		GeneratedAt: time.Now(),
		Count:       len(codes),
	}, nil
}

// SendMFACode sends an MFA code via SMS or Email (placeholder implementation)
func (s *mfaService) SendMFACode(ctx context.Context, req *SendMFACodeRequest) (*MFACodeSentResult, error) {
	// TODO: Implement MFA code sending for SMS/Email methods
	// This would integrate with the notification service to send codes on-demand
	return &MFACodeSentResult{
		Success:   false,
		Message:   "MFA code sending not implemented",
		SentAt:    time.Time{},
		ExpiresAt: time.Time{},
	}, nil
}

// UseBackupCode validates and uses a backup code
func (s *mfaService) UseBackupCode(ctx context.Context, req *UseBackupCodeRequest) (*BackupCodeResult, error) {
	if s.validateBackupCode(ctx, req.UserID, req.Code) {
		return &BackupCodeResult{
			Valid:   true,
			UsedAt:  time.Now(),
			Message: "Backup code validated successfully",
		}, nil
	}

	return &BackupCodeResult{
		Valid:   false,
		Message: "Invalid or already used backup code",
	}, nil
}

// GetMFASettings returns MFA settings for a user
func (s *mfaService) GetMFASettings(ctx context.Context, userID uuid.UUID) (*MFASettings, error) {
	methods, err := s.mfaRepo.GetUserMFAMethods(ctx, userID)
	if err != nil {
		return nil, err
	}

	backupCodes, err := s.mfaRepo.GetBackupCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	unusedCodes := 0
	for _, code := range backupCodes {
		if !code.Used {
			unusedCodes++
		}
	}

	return &MFASettings{
		UserID:               userID,
		MFAEnabled:           len(methods) > 0,
		EnrolledMethods:      methods,
		PrimaryMethodID:      s.getPrimaryMethodID(methods),
		BackupCodesRemaining: unusedCodes,
		LastValidationAt:     s.getLastValidationTime(methods),
		RequiredForAccount:   false, // This would be configurable per tenant/organization
	}, nil
}

// UpdateMFASettings updates MFA settings for a user
func (s *mfaService) UpdateMFASettings(ctx context.Context, req *UpdateMFASettingsRequest) (*MFASettings, error) {
	// TODO: Implement MFA settings updates (primary method, requirements, etc.)
	return s.GetMFASettings(ctx, req.UserID)
}

// GetMFAStatistics returns MFA statistics for reporting
func (s *mfaService) GetMFAStatistics(ctx context.Context, req *MFAStatisticsRequest) (*MFAStatistics, error) {
	// TODO: Implement MFA statistics collection
	return &MFAStatistics{
		TotalUsers:            0,
		MFAEnabledUsers:       0,
		MethodDistribution:    make(map[string]int),
		ValidationSuccessRate: 0,
		Period:                req.Period,
		GeneratedAt:           time.Now(),
	}, nil
}

// Helper methods for settings
func (s *mfaService) getPrimaryMethodID(methods []*MFAMethod) *uuid.UUID {
	for _, method := range methods {
		if method.IsPrimary && method.IsActive {
			return &method.ID
		}
	}
	return nil
}

func (s *mfaService) getLastValidationTime(methods []*MFAMethod) *time.Time {
	var latest *time.Time
	for _, method := range methods {
		if method.LastUsedAt != nil {
			if latest == nil || method.LastUsedAt.After(*latest) {
				latest = method.LastUsedAt
			}
		}
	}
	return latest
}
