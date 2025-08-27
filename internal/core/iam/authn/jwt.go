package authn

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/errors"
)

const (
	// JWT Configuration
	defaultAccessTokenExpiry  = 15 * time.Minute
	defaultRefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
	issuer                    = "erp-system"

	// JWT Claims
	claimTenantID   = "tenant_id"
	claimEntityID   = "entity_id"
	claimUserID     = "user_id"
	claimEmail      = "email"
	claimRoles      = "roles"
	claimMFAEnabled = "mfa_enabled"
	claimType       = "type"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	accessSecret  string
	refreshSecret string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(accessSecret, refreshSecret string) *JWTManager {
	return &JWTManager{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

// TokenType represents the type of JWT token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	EntityID   uuid.UUID `json:"entity_id"`
	Email      string    `json:"email"`
	Roles      []string  `json:"roles"`
	MFAEnabled bool      `json:"mfa_enabled"`
	TokenType  TokenType `json:"type"`
	jwt.RegisteredClaims
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (j *JWTManager) GenerateTokenPair(user *model.User, roles []string) (accessToken, refreshToken string, expiresAt time.Time, err error) {
	now := time.Now()

	// Generate access token
	accessToken, accessExpiry, err := j.generateToken(user, roles, TokenTypeAccess, now)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err = j.generateToken(user, roles, TokenTypeRefresh, now)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, accessExpiry, nil
}

// generateToken generates a JWT token for the specified user and type
func (j *JWTManager) generateToken(user *model.User, roles []string, tokenType TokenType, now time.Time) (string, time.Time, error) {
	var expiry time.Duration
	var secret string

	switch tokenType {
	case TokenTypeAccess:
		expiry = defaultAccessTokenExpiry
		secret = j.accessSecret
	case TokenTypeRefresh:
		expiry = defaultRefreshTokenExpiry
		secret = j.refreshSecret
	default:
		return "", time.Time{}, fmt.Errorf("invalid token type: %s", tokenType)
	}

	expiresAt := now.Add(expiry)

	// Create claims
	claims := &TokenClaims{
		UserID:     user.ID,
		TenantID:   user.TenantID,
		EntityID:   user.EntityID,
		Email:      user.Email,
		Roles:      roles,
		MFAEnabled: user.MFAEnabled,
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (j *JWTManager) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	return j.validateToken(tokenString, TokenTypeAccess, j.accessSecret)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	return j.validateToken(tokenString, TokenTypeRefresh, j.refreshSecret)
}

// validateToken validates a JWT token and returns the claims
func (j *JWTManager) validateToken(tokenString string, expectedType TokenType, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.NewBusinessError("INVALID_TOKEN", "Token validation failed")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.NewBusinessError("INVALID_TOKEN", "Invalid token claims")
	}

	// Verify token type
	if claims.TokenType != expectedType {
		return nil, errors.NewBusinessError("INVALID_TOKEN", fmt.Sprintf("Expected %s token, got %s", expectedType, claims.TokenType))
	}

	// Verify issuer
	if claims.Issuer != issuer {
		return nil, errors.NewBusinessError("INVALID_TOKEN", "Invalid token issuer")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.NewBusinessError("TOKEN_EXPIRED", "Token has expired")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTManager) RefreshAccessToken(refreshToken string) (newAccessToken string, expiresAt time.Time, err error) {
	// Validate refresh token
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now()

	// Generate new access token with same claims
	newClaims := &TokenClaims{
		UserID:     claims.UserID,
		TenantID:   claims.TenantID,
		EntityID:   claims.EntityID,
		Email:      claims.Email,
		Roles:      claims.Roles,
		MFAEnabled: claims.MFAEnabled,
		TokenType:  TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultAccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   claims.UserID.String(),
			ID:        uuid.New().String(),
		},
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err := token.SignedString([]byte(j.accessSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign new access token: %w", err)
	}

	return tokenString, newClaims.ExpiresAt.Time, nil
}

// ExtractClaimsFromToken extracts claims without validation (for logging/debugging)
func ExtractClaimsFromToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't validate signature for extraction
		return []byte("dummy"), nil
	})

	if err != nil {
		// Check if it's a malformed token error
		if err == jwt.ErrTokenMalformed {
			return nil, fmt.Errorf("malformed token: %w", err)
		}
		// For other parsing errors, try to extract claims if token was parsed
		if token != nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				return claims, nil
			}
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("failed to extract claims")
}

// GetTokenExpiration returns the expiration time for a token type
func GetTokenExpiration(tokenType TokenType) time.Duration {
	switch tokenType {
	case TokenTypeAccess:
		return defaultAccessTokenExpiry
	case TokenTypeRefresh:
		return defaultRefreshTokenExpiry
	default:
		return 0
	}
}
