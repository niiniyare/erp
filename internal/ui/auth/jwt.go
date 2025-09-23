package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the structure of JWT tokens for all UI services
type JWTClaims struct {
	jwt.RegisteredClaims
	
	// User Information
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	
	// Role and Permissions
	Role        string   `json:"role"`          // "admin", "tenant_user", "client"
	Scopes      []string `json:"scopes"`       // ["admin:read", "tenant:write", "client:read"]
	Permissions []string `json:"permissions"`  // ABAC permissions
	
	// Context Information
	TenantID    *string `json:"tenant_id,omitempty"`    // Nil for admin, set for tenant users
	ClientID    *string `json:"client_id,omitempty"`    // Set for portal users
	
	// UI-Specific Claims
	UIAccess    []string `json:"ui_access"`              // ["console", "workspace", "portal"]
	SessionType string   `json:"session_type"`           // "web", "api", "widget"
	
	// Security
	TokenType   string `json:"token_type"`              // "access", "refresh"
	DeviceID    string `json:"device_id,omitempty"`     // For device tracking
}

// JWTValidator handles JWT token validation and claims extraction
type JWTValidator struct {
	publicKey     *rsa.PublicKey
	privateKey    *rsa.PrivateKey
	issuer        string
	audience      string
	tokenDuration time.Duration
}

// NewJWTValidator creates a new JWT validator with RSA keys
func NewJWTValidator(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey, issuer, audience string, tokenDuration time.Duration) *JWTValidator {
	return &JWTValidator{
		publicKey:     publicKey,
		privateKey:    privateKey,
		issuer:        issuer,
		audience:      audience,
		tokenDuration: tokenDuration,
	}
}

// ValidateToken validates a JWT token and returns claims
func (v *JWTValidator) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Validate issuer and audience
	if claims.Issuer != v.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	if claims.Audience[0] != v.audience {
		return nil, fmt.Errorf("invalid audience")
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// GenerateToken creates a new JWT token with specified claims
func (v *JWTValidator) GenerateToken(claims *JWTClaims) (string, error) {
	// Set standard claims
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    v.issuer,
		Audience:  []string{v.audience},
		Subject:   claims.UserID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(v.tokenDuration)),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(v.privateKey)
}

// HasScope checks if the claims contain a specific scope
func (c *JWTClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasUIAccess checks if the claims allow access to a specific UI
func (c *JWTClaims) HasUIAccess(ui string) bool {
	for _, access := range c.UIAccess {
		if access == ui {
			return true
		}
	}
	return false
}

// GetTenantID safely returns the tenant ID (handles nil pointer)
func (c *JWTClaims) GetTenantID() string {
	if c.TenantID != nil {
		return *c.TenantID
	}
	return ""
}

// GetClientID safely returns the client ID (handles nil pointer)
func (c *JWTClaims) GetClientID() string {
	if c.ClientID != nil {
		return *c.ClientID
	}
	return ""
}

// IsAdmin checks if the user has admin privileges
func (c *JWTClaims) IsAdmin() bool {
	return c.Role == "admin" && c.HasScope("admin:read")
}

// IsTenantUser checks if the user is a tenant user
func (c *JWTClaims) IsTenantUser() bool {
	return c.Role == "tenant_user" && c.HasScope("tenant:read")
}

// IsClient checks if the user is a client user
func (c *JWTClaims) IsClient() bool {
	return c.Role == "client" && c.HasScope("client:read")
}

// Example JWT claims for different user types:

// CreateAdminClaims creates JWT claims for an admin user
func CreateAdminClaims(userID, email, firstName, lastName string) *JWTClaims {
	return &JWTClaims{
		UserID:      userID,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Role:        "admin",
		Scopes:      []string{"admin:read", "admin:write", "console:access"},
		UIAccess:    []string{"console", "workspace", "portal"}, // Admin can access all
		SessionType: "web",
		TokenType:   "access",
	}
}

// CreateTenantUserClaims creates JWT claims for a tenant user
func CreateTenantUserClaims(userID, email, firstName, lastName, tenantID string, permissions []string) *JWTClaims {
	return &JWTClaims{
		UserID:      userID,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Role:        "tenant_user",
		Scopes:      []string{"tenant:read", "tenant:write", "workspace:access"},
		Permissions: permissions,
		TenantID:    &tenantID,
		UIAccess:    []string{"workspace"},
		SessionType: "web",
		TokenType:   "access",
	}
}

// CreateClientClaims creates JWT claims for a client user
func CreateClientClaims(userID, email, firstName, lastName, clientID string) *JWTClaims {
	return &JWTClaims{
		UserID:      userID,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		Role:        "client",
		Scopes:      []string{"client:read", "portal:access"},
		ClientID:    &clientID,
		UIAccess:    []string{"portal"},
		SessionType: "web",
		TokenType:   "access",
	}
}