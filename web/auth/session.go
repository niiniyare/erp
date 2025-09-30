package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// SessionManager handles secure session management with Redis storage
type SessionManager struct {
	redis        *redis.Client
	jwtValidator *JWTValidator
	logger       logger.Logger
	config       SessionConfig
}

// SessionConfig contains session management configuration
type SessionConfig struct {
	CookieName     string        // Name of the session cookie
	CookieDomain   string        // Cookie domain
	CookiePath     string        // Cookie path
	CookieSecure   bool          // Secure flag (HTTPS only)
	CookieHttpOnly bool          // HttpOnly flag
	CookieSameSite http.SameSite // SameSite policy
	SessionTTL     time.Duration // Session expiration time
	RefreshTTL     time.Duration // Refresh token expiration time
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id,omitempty"`
	ClientID     string    `json:"client_id,omitempty"`
	Role         string    `json:"role"`
	UIService    string    `json:"ui_service"` // "console", "workspace", "portal"
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccess   time.Time `json:"last_access"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceID     string    `json:"device_id,omitempty"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(redis *redis.Client, jwtValidator *JWTValidator, config SessionConfig, logger logger.Logger) *SessionManager {
	return &SessionManager{
		redis:        redis,
		jwtValidator: jwtValidator,
		logger:       logger,
		config:       config,
	}
}

// CreateSession creates a new session and returns session ID and tokens
func (sm *SessionManager) CreateSession(ctx context.Context, claims *JWTClaims, uiService, ipAddress, userAgent string) (*Session, error) {
	// Generate session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate access token
	accessToken, err := sm.jwtValidator.GenerateToken(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &JWTClaims{
		UserID:      claims.UserID,
		Role:        claims.Role,
		TenantID:    claims.TenantID,
		ClientID:    claims.ClientID,
		TokenType:   "refresh",
		SessionType: "web",
	}
	refreshToken, err := sm.jwtValidator.GenerateToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session object
	session := &Session{
		ID:           sessionID,
		UserID:       claims.UserID,
		TenantID:     claims.GetTenantID(),
		ClientID:     claims.GetClientID(),
		Role:         claims.Role,
		UIService:    uiService,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now(),
		LastAccess:   time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		DeviceID:     claims.DeviceID,
	}

	// Store session in Redis
	key := fmt.Sprintf("session:%s", sessionID)
	if err := sm.redis.HSet(ctx, key, map[string]interface{}{
		"user_id":       session.UserID,
		"tenant_id":     session.TenantID,
		"client_id":     session.ClientID,
		"role":          session.Role,
		"ui_service":    session.UIService,
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"created_at":    session.CreatedAt.Unix(),
		"last_access":   session.LastAccess.Unix(),
		"ip_address":    session.IPAddress,
		"user_agent":    session.UserAgent,
		"device_id":     session.DeviceID,
	}).Err(); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Set expiration
	if err := sm.redis.Expire(ctx, key, sm.config.SessionTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to set session expiration: %w", err)
	}

	sm.logger.Info("Session created", logger.Fields{
		"session_id": sessionID,
		"user_id":    claims.UserID,
		"ui_service": uiService,
		"role":       claims.Role,
	})

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	result, err := sm.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("session not found")
	}

	// Parse timestamps with proper error handling
	createdAt, err := parseTimestamp(result["created_at"])
	if err != nil {
		return nil, fmt.Errorf("invalid created_at timestamp: %w", err)
	}

	lastAccess, err := parseTimestamp(result["last_access"])
	if err != nil {
		return nil, fmt.Errorf("invalid last_access timestamp: %w", err)
	}

	session := &Session{
		ID:           sessionID,
		UserID:       result["user_id"],
		TenantID:     result["tenant_id"],
		ClientID:     result["client_id"],
		Role:         result["role"],
		UIService:    result["ui_service"],
		AccessToken:  result["access_token"],
		RefreshToken: result["refresh_token"],
		CreatedAt:    createdAt,
		LastAccess:   lastAccess,
		IPAddress:    result["ip_address"],
		UserAgent:    result["user_agent"],
		DeviceID:     result["device_id"],
	}

	// Update last access time
	sm.redis.HSet(ctx, key, "last_access", time.Now().Unix())
	sm.redis.Expire(ctx, key, sm.config.SessionTTL)

	return session, nil
}

// RefreshSession refreshes the access token using refresh token
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID, refreshToken string) (*Session, error) {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Validate refresh token
	if session.RefreshToken != refreshToken {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Validate refresh token JWT
	claims, err := sm.jwtValidator.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("not a refresh token")
	}

	// Create new access token with updated expiration
	newAccessClaims := &JWTClaims{
		UserID:      claims.UserID,
		Role:        claims.Role,
		TenantID:    claims.TenantID,
		ClientID:    claims.ClientID,
		TokenType:   "access",
		SessionType: "web",
	}

	newAccessToken, err := sm.jwtValidator.GenerateToken(newAccessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Update session
	session.AccessToken = newAccessToken
	session.LastAccess = time.Now()

	// Store updated session
	key := fmt.Sprintf("session:%s", sessionID)
	sm.redis.HSet(ctx, key, map[string]interface{}{
		"access_token": session.AccessToken,
		"last_access":  session.LastAccess.Unix(),
	})

	sm.logger.Info("Session refreshed", logger.Fields{
		"session_id": sessionID,
		"user_id":    claims.UserID,
	})

	return session, nil
}

// DeleteSession deletes a session
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sm.redis.Del(ctx, key).Err()
}

// SetSessionCookie sets the session cookie in HTTP response
func (sm *SessionManager) SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    sessionID,
		Domain:   sm.config.CookieDomain,
		Path:     sm.config.CookiePath,
		MaxAge:   int(sm.config.SessionTTL.Seconds()),
		Secure:   sm.config.CookieSecure,
		HttpOnly: sm.config.CookieHttpOnly,
		SameSite: sm.config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.config.CookieName,
		Value:    "",
		Domain:   sm.config.CookieDomain,
		Path:     sm.config.CookiePath,
		MaxAge:   -1,
		Secure:   sm.config.CookieSecure,
		HttpOnly: sm.config.CookieHttpOnly,
		SameSite: sm.config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromRequest extracts session ID from request cookie
func (sm *SessionManager) GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sm.config.CookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found: %w", err)
	}
	return cookie.Value, nil
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// parseInt64 safely parses string to int64
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	// Simple conversion for this example - in production use strconv.ParseInt
	return 0
}

// Login endpoints for each UI service
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	UIService  string `json:"ui_service"` // "console", "workspace", "portal"
	RememberMe bool   `json:"remember_me"`
	DeviceID   string `json:"device_id,omitempty"`
}

type LoginResponse struct {
	Success     bool   `json:"success"`
	RedirectURL string `json:"redirect_url"`
	Message     string `json:"message,omitempty"`
}

// AuthenticationService handles login/logout for all UI services
type AuthenticationService struct {
	sessionManager *SessionManager
	userService    UserService // Your existing user service
	logger         logger.Logger
}

// UserService interface for user authentication
type UserService interface {
	AuthenticateUser(ctx context.Context, email, password string) (*JWTClaims, error)
}

// Login handles login for all UI services
func (as *AuthenticationService) Login(w http.ResponseWriter, r *http.Request, req *LoginRequest) (*LoginResponse, error) {
	ctx := r.Context()

	// Authenticate user
	claims, err := as.userService.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Invalid email or password",
		}, nil
	}

	// Validate UI access
	if !claims.HasUIAccess(req.UIService) {
		return &LoginResponse{
			Success: false,
			Message: "Access denied to this interface",
		}, nil
	}

	// Create session
	session, err := as.sessionManager.CreateSession(
		ctx,
		claims,
		req.UIService,
		getIPAddress(r),
		r.UserAgent(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set session cookie
	as.sessionManager.SetSessionCookie(w, session.ID)

	// Determine redirect URL based on UI service
	redirectURL := getRedirectURL(req.UIService, claims.Role)

	return &LoginResponse{
		Success:     true,
		RedirectURL: redirectURL,
	}, nil
}

// Logout handles logout for all UI services
func (as *AuthenticationService) Logout(w http.ResponseWriter, r *http.Request) error {
	sessionID, err := as.sessionManager.GetSessionFromRequest(r)
	if err != nil {
		// No session to logout
		return nil
	}

	// Delete session
	if err := as.sessionManager.DeleteSession(r.Context(), sessionID); err != nil {
		as.logger.Error("Failed to delete session", logger.Fields{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	// Clear session cookie
	as.sessionManager.ClearSessionCookie(w)

	return nil
}

// Helper function to parse timestamp strings
func parseTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, fmt.Errorf("timestamp is empty")
	}

	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(sec, 0), nil
}

// Helper functions
func getIPAddress(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func getRedirectURL(uiService, role string) string {
	switch uiService {
	case "console":
		return "/console/dashboard"
	case "workspace":
		return "/workspace/dashboard"
	case "portal":
		return "/portal/dashboard"
	default:
		return "/"
	}
}
