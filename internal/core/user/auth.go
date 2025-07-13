package user

// import (
// 	"context"
// 	"crypto/rand"
// 	"database/sql"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"log/slog"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"github.com/jackc/pgx/v5/pgtype"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	db "github.com/niiniyare/erp/db/sqlc"
// 	"golang.org/x/crypto/bcrypt"
// )
//
// // ================================================================================================
// // AUTHENTICATION SERVICE
// // ================================================================================================
//
// type AuthService struct {
// 	db     *db.Queries
// 	pool   *pgxpool.Pool
// 	logger *slog.Logger
// 	config AuthConfig
// }
//
// type AuthConfig struct {
// 	SessionDuration time.Duration
// 	BCryptCost      int
// 	MaxFailedLogins int
// 	LockoutDuration time.Duration
// 	TokenLength     int
// 	RequireMFA      bool
// }
//
// func NewAuthService(pool *pgxpool.Pool, logger *slog.Logger, config AuthConfig) *AuthService {
// 	if config.SessionDuration == 0 {
// 		config.SessionDuration = 8 * time.Hour
// 	}
// 	if config.BCryptCost == 0 {
// 		config.BCryptCost = bcrypt.DefaultCost
// 	}
// 	if config.MaxFailedLogins == 0 {
// 		config.MaxFailedLogins = 5
// 	}
// 	if config.LockoutDuration == 0 {
// 		config.LockoutDuration = 30 * time.Minute
// 	}
// 	if config.TokenLength == 0 {
// 		config.TokenLength = 32
// 	}
//
// 	return &AuthService{
// 		db:     db.New(pool),
// 		pool:   pool,
// 		logger: logger,
// 		config: config,
// 	}
// }
//
// type LoginRequest struct {
// 	Email        string                 `json:"email" validate:"required,email"`
// 	Password     string                 `json:"password" validate:"required"`
// 	TenantID     uuid.UUID              `json:"tenant_id" validate:"required"`
// 	IPAddress    string                 `json:"ip_address"`
// 	UserAgent    string                 `json:"user_agent"`
// 	DeviceInfo   map[string]interface{} `json:"device_info"`
// 	LocationInfo map[string]interface{} `json:"location_info"`
// 	MFACode      string                 `json:"mfa_code"`
// }
//
// type LoginResponse struct {
// 	User         *db.User  `json:"user"`
// 	SessionToken string    `json:"session_token"`
// 	RefreshToken string    `json:"refresh_token"`
// 	ExpiresAt    time.Time `json:"expires_at"`
// 	RequiresMFA  bool      `json:"requires_mfa"`
// }
//
// func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
// 	ctx = WithTenant(ctx, req.TenantID)
//
// 	// Set tenant context
// 	if err := s.db.SetTenantContext(ctx, req.TenantID.String()); err != nil {
// 		s.logger.Error("Failed to set tenant context", "error", err, "tenant_id", req.TenantID)
// 		return nil, ServiceError{ErrCodeInternalError, "Internal server error", err}
// 	}
//
// 	// Get user by email
// 	user, err := s.db.GetUserByEmail(ctx, db.GetUserByEmailParams{
// 		Email:    req.Email,
// 		TenantID: req.TenantID,
// 	})
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			s.logger.Warn("Login attempt with invalid email", "email", req.Email, "ip", req.IPAddress)
// 			return nil, ServiceError{ErrCodeUnauthorized, "Invalid credentials", nil}
// 		}
// 		s.logger.Error("Failed to get user by email", "error", err, "email", req.Email)
// 		return nil, ServiceError{ErrCodeInternalError, "Internal server error", err}
// 	}
//
// 	// Check account status
// 	if user.AccountStatus != "ACTIVE" {
// 		s.logger.Warn("Login attempt on inactive account", "user_id", user.ID, "status", user.AccountStatus)
// 		return nil, ServiceError{ErrCodeUnauthorized, "Account is not active", nil}
// 	}
//
// 	// Check if account is locked
// 	if user.LockoutUntil.Valid && user.LockoutUntil.Time.After(time.Now()) {
// 		s.logger.Warn("Login attempt on locked account", "user_id", user.ID, "lockout_until", user.LockoutUntil.Time)
// 		return nil, ServiceError{ErrCodeAccountLocked, fmt.Sprintf("Account is locked until %v", user.LockoutUntil.Time), nil}
// 	}
//
// 	// Verify password
// 	if !user.PasswordHash.Valid {
// 		s.logger.Error("User has no password hash", "user_id", user.ID)
// 		return nil, ServiceError{ErrCodeUnauthorized, "Invalid credentials", nil}
// 	}
//
// 	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
// 		// Increment failed login attempts
// 		if incrementErr := s.db.IncrementFailedLogins(ctx, user.ID); incrementErr != nil {
// 			s.logger.Error("Failed to increment failed login attempts", "error", incrementErr, "user_id", user.ID)
// 		}
//
// 		s.logger.Warn("Invalid password attempt", "user_id", user.ID, "ip", req.IPAddress)
// 		return nil, ServiceError{ErrCodeUnauthorized, "Invalid credentials", nil}
// 	}
//
// 	// Check MFA if required
// 	if user.MfaEnabled && req.MFACode == "" {
// 		return &LoginResponse{RequiresMFA: true}, nil
// 	}
//
// 	if user.MfaEnabled && req.MFACode != "" {
// 		// Validate MFA code (implementation depends on your MFA provider)
// 		if !s.validateMFACode(user.MfaSecret.String, req.MFACode) {
// 			s.logger.Warn("Invalid MFA code", "user_id", user.ID, "ip", req.IPAddress)
// 			return nil, ServiceError{ErrCodeUnauthorized, "Invalid MFA code", nil}
// 		}
// 	}
//
// 	// Generate session tokens
// 	sessionToken := s.generateSecureToken()
// 	refreshToken := s.generateSecureToken()
// 	expiresAt := time.Now().Add(s.config.SessionDuration)
//
// 	// Create session
// 	deviceInfoJSON, _ := json.Marshal(req.DeviceInfo)
// 	locationInfoJSON, _ := json.Marshal(req.LocationInfo)
//
// 	session, err := s.db.CreateSession(ctx, db.CreateSessionParams{
// 		UserID:       user.ID,
// 		SessionToken: sessionToken,
// 		RefreshToken: sql.NullString{String: refreshToken, Valid: true},
// 		IpAddress:    sql.NullString{String: req.IPAddress, Valid: req.IPAddress != ""},
// 		UserAgent:    sql.NullString{String: req.UserAgent, Valid: req.UserAgent != ""},
// 		DeviceInfo:   pgtype.JSONB{Bytes: deviceInfoJSON, Status: getJSONBStatus(deviceInfoJSON)},
// 		LocationInfo: pgtype.JSONB{Bytes: locationInfoJSON, Status: getJSONBStatus(locationInfoJSON)},
// 		ExpiresAt:    expiresAt,
// 	})
// 	if err != nil {
// 		s.logger.Error("Failed to create session", "error", err, "user_id", user.ID)
// 		return nil, ServiceError{ErrCodeInternalError, "Failed to create session", err}
// 	}
//
// 	// Update last login
// 	if err := s.db.UpdateUserLastLogin(ctx, user.ID); err != nil {
// 		s.logger.Warn("Failed to update last login", "error", err, "user_id", user.ID)
// 	}
//
// 	// Log successful login
// 	s.logSecurityEvent(ctx, "USER_LOGIN", "AUTH", "INFO", user.ID, req.IPAddress, "Successful login", 10)
//
// 	s.logger.Info("User logged in successfully", "user_id", user.ID, "session_id", session.ID)
//
// 	return &LoginResponse{
// 		User:         &user,
// 		SessionToken: sessionToken,
// 		RefreshToken: refreshToken,
// 		ExpiresAt:    expiresAt,
// 		RequiresMFA:  false,
// 	}, nil
// }
//
// func (s *AuthService) ValidateSession(ctx context.Context, token string) (*db.UserSession, *db.User, error) {
// 	session, err := s.db.GetSessionByToken(ctx, token)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil, ServiceError{ErrCodeSessionExpired, "Invalid or expired session", nil}
// 		}
// 		s.logger.Error("Failed to get session by token", "error", err)
// 		return nil, nil, ServiceError{ErrCodeInternalError, "Internal server error", err}
// 	}
//
// 	// Set tenant context
// 	if err := s.db.SetTenantContext(ctx, session.TenantID.String()); err != nil {
// 		return nil, nil, ServiceError{ErrCodeInternalError, "Internal server error", err}
// 	}
//
// 	// Get user
// 	user, err := s.db.GetUserByID(ctx, session.UserID)
// 	if err != nil {
// 		s.logger.Error("Failed to get user for session", "error", err, "session_id", session.ID)
// 		return nil, nil, ServiceError{ErrCodeInternalError, "Internal server error", err}
// 	}
//
// 	// Check user status
// 	if !user.IsActive || user.AccountStatus != "ACTIVE" {
// 		s.logger.Warn("Session validation failed - user inactive", "user_id", user.ID, "session_id", session.ID)
// 		return nil, nil, ServiceError{ErrCodeUnauthorized, "User account is not active", nil}
// 	}
//
// 	// Update last accessed time
// 	if err := s.db.UpdateSessionAccess(ctx, session.ID); err != nil {
// 		s.logger.Warn("Failed to update session access time", "error", err, "session_id", session.ID)
// 	}
//
// 	return &session, &user, nil
// }
//
// func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
// 	if err := s.db.InvalidateSession(ctx, sessionID); err != nil {
// 		s.logger.Error("Failed to invalidate session", "error", err, "session_id", sessionID)
// 		return ServiceError{ErrCodeInternalError, "Failed to logout", err}
// 	}
//
// 	// Log logout
// 	if userID, ok := GetUserID(ctx); ok {
// 		if tenantID, ok := GetTenantID(ctx); ok {
// 			ctx = WithTenant(ctx, tenantID)
// 			s.logSecurityEvent(ctx, "USER_LOGOUT", "AUTH", "INFO", userID, "", "User logged out", 5)
// 		}
// 	}
//
// 	s.logger.Info("User logged out", "session_id", sessionID)
// 	return nil
// }
//
// func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
// 	tenantID, ok := GetTenantID(ctx)
// 	if !ok {
// 		return ServiceError{ErrCodeInternalError, "Tenant context required", nil}
// 	}
//
// 	ctx = WithTenant(ctx, tenantID)
// 	if err := s.db.SetTenantContext(ctx, tenantID.String()); err != nil {
// 		return ServiceError{ErrCodeInternalError, "Internal server error", err}
// 	}
//
// 	// Get user
// 	user, err := s.db.GetUserByID(ctx, userID)
// 	if err != nil {
// 		return ServiceError{ErrCodeNotFound, "User not found", err}
// 	}
//
// 	// Verify old password
// 	if user.PasswordHash.Valid {
// 		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(oldPassword)); err != nil {
// 			s.logger.Warn("Invalid old password in change password attempt", "user_id", userID)
// 			return ServiceError{ErrCodeUnauthorized, "Invalid current password", nil}
// 		}
// 	}
//
// 	// Hash new password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.BCryptCost)
// 	if err != nil {
// 		s.logger.Error("Failed to hash new password", "error", err, "user_id", userID)
// 		return ServiceError{ErrCodeInternalError, "Failed to process password", err}
// 	}
//
// 	// Update password
// 	if err := s.db.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
// 		ID:           userID,
// 		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
// 	}); err != nil {
// 		s.logger.Error("Failed to update password", "error", err, "user_id", userID)
// 		return ServiceError{ErrCodeInternalError, "Failed to update password", err}
// 	}
//
// 	// Invalidate all sessions for security
// 	if err := s.db.InvalidateUserSessions(ctx, userID); err != nil {
// 		s.logger.Warn("Failed to invalidate user sessions after password change", "error", err, "user_id", userID)
// 	}
//
// 	// Log password change
// 	s.logSecurityEvent(ctx, "PASSWORD_CHANGED", "AUTH", "WARN", userID, "", "User changed password", 25)
//
// 	s.logger.Info("User password changed", "user_id", userID)
// 	return nil
// }
//
// func (s *AuthService) generateSecureToken() string {
// 	bytes := make([]byte, s.config.TokenLength)
// 	rand.Read(bytes)
// 	return hex.EncodeToString(bytes)
// }
//
// func (s *AuthService) validateMFACode(secret, code string) bool {
// 	// Implementation depends on your MFA provider (TOTP, SMS, etc.)
// 	// This is a placeholder
// 	return true
// }
//
// func (s *AuthService) logSecurityEvent(ctx context.Context, eventType, category, severity string, userID uuid.UUID, ipAddress, reason string, riskScore int32) {
// 	contextData := map[string]interface{}{
// 		"ip_address": ipAddress,
// 		"timestamp":  time.Now(),
// 	}
// 	contextJSON, _ := json.Marshal(contextData)
//
// 	_, err := s.db.LogAuditEvent(ctx, db.LogAuditEventParams{
// 		EventType:     eventType,
// 		EventCategory: category,
// 		Severity:      severity,
// 		UserID:        sql.NullString{String: userID.String(), Valid: true},
// 		Reason:        sql.NullString{String: reason, Valid: true},
// 		RiskScore:     sql.NullInt32{Int32: riskScore, Valid: true},
// 		Context:       pgtype.JSONB{Bytes: contextJSON, Status: pgtype.Present},
// 		IpAddress:     sql.NullString{String: ipAddress, Valid: ipAddress != ""},
// 	})
// 	if err != nil {
// 		s.logger.Error("Failed to log security event", "error", err)
// 	}
// }
