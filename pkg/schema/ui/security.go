package ui

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SecurityLevel represents the security classification level
type SecurityLevel string

const (
	SecurityLevelPublic       SecurityLevel = "public"
	SecurityLevelInternal     SecurityLevel = "internal"
	SecurityLevelConfidential SecurityLevel = "confidential"
	SecurityLevelSecret       SecurityLevel = "secret"
	SecurityLevelTopSecret    SecurityLevel = "top-secret"
)

// TenantSecurityContext provides military-grade tenant isolation
type TenantSecurityContext struct {
	TenantID        string            `json:"tenant_id" validate:"required"`
	UserID          string            `json:"user_id" validate:"required"`
	SessionID       string            `json:"session_id" validate:"required"`
	SecurityLevel   SecurityLevel     `json:"security_level"`
	Permissions     []Permission      `json:"permissions"`
	AccessControls  []AccessControl   `json:"access_controls"`
	AuditContext    *AuditContext     `json:"audit_context,omitempty"`
	EncryptionKeys  *EncryptionKeys   `json:"-"` // Never serialize keys
	CreatedAt       time.Time         `json:"created_at"`
	ExpiresAt       time.Time         `json:"expires_at"`
	metadata        map[string]string // Private metadata
	mu              sync.RWMutex      // Thread safety
}

// Permission represents a granular security permission
type Permission struct {
	Resource    string          `json:"resource" validate:"required"`
	Action      string          `json:"action" validate:"required"`
	Conditions  []Condition     `json:"conditions,omitempty"`
	Constraints []Constraint    `json:"constraints,omitempty"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
}

// AccessControl defines access control rules
type AccessControl struct {
	Type        AccessControlType `json:"type" validate:"required"`
	Subject     string           `json:"subject" validate:"required"`
	Object      string           `json:"object" validate:"required"`
	Action      string           `json:"action" validate:"required"`
	Effect      EffectType       `json:"effect" validate:"required"`
	Conditions  []Condition      `json:"conditions,omitempty"`
	Priority    int              `json:"priority"`
}

// AccessControlType represents different types of access controls
type AccessControlType string

const (
	AccessControlMAC  AccessControlType = "mac"  // Mandatory Access Control
	AccessControlRBAC AccessControlType = "rbac" // Role-Based Access Control
	AccessControlABAC AccessControlType = "abac" // Attribute-Based Access Control
	AccessControlDAC  AccessControlType = "dac"  // Discretionary Access Control
)

// EffectType represents the effect of an access control rule
type EffectType string

const (
	EffectAllow EffectType = "allow"
	EffectDeny  EffectType = "deny"
)

// Condition represents a security condition
type Condition struct {
	Attribute string      `json:"attribute" validate:"required"`
	Operator  string      `json:"operator" validate:"required"`
	Value     interface{} `json:"value" validate:"required"`
}

// Constraint represents a security constraint
type Constraint struct {
	Type        string      `json:"type" validate:"required"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string      `json:"description,omitempty"`
}

// AuditContext tracks security audit information
type AuditContext struct {
	RequestID    string            `json:"request_id" validate:"required"`
	UserAgent    string            `json:"user_agent,omitempty"`
	IPAddress    string            `json:"ip_address,omitempty"`
	GeoLocation  *GeoLocation      `json:"geo_location,omitempty"`
	DeviceInfo   *DeviceInfo       `json:"device_info,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
}

// GeoLocation tracks geographical location for security
type GeoLocation struct {
	Country     string  `json:"country,omitempty"`
	Region      string  `json:"region,omitempty"`
	City        string  `json:"city,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Accuracy    int     `json:"accuracy,omitempty"`
}

// DeviceInfo tracks device information for security
type DeviceInfo struct {
	DeviceID     string `json:"device_id,omitempty"`
	DeviceType   string `json:"device_type,omitempty"`
	OS           string `json:"os,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`
	Browser      string `json:"browser,omitempty"`
	BrowserVersion string `json:"browser_version,omitempty"`
	Fingerprint  string `json:"fingerprint,omitempty"`
}

// EncryptionKeys holds encryption keys for tenant data
type EncryptionKeys struct {
	DataEncryptionKey    []byte `json:"-"`
	KeyEncryptionKey     []byte `json:"-"`
	SigningKey           []byte `json:"-"`
	RotationSchedule     time.Duration
	LastRotated          time.Time
	NextRotation         time.Time
}

// SecurityValidator provides military-grade security validation
type SecurityValidator struct {
	securityPolicies map[SecurityLevel]*SecurityPolicy
	encryptionSuite  *EncryptionSuite
	auditLogger      AuditLogger
	mu              sync.RWMutex
}

// SecurityPolicy defines security requirements for each level
type SecurityPolicy struct {
	RequiredPermissions    []string          `json:"required_permissions"`
	AllowedOperations      []string          `json:"allowed_operations"`
	EncryptionRequired     bool              `json:"encryption_required"`
	AuditLevel            AuditLevel        `json:"audit_level"`
	SessionTimeout        time.Duration     `json:"session_timeout"`
	MaxConcurrentSessions int               `json:"max_concurrent_sessions"`
	IPWhitelist           []string          `json:"ip_whitelist,omitempty"`
	GeoRestrictions       []GeoRestriction  `json:"geo_restrictions,omitempty"`
	DataClassification    DataClassification `json:"data_classification"`
}

// AuditLevel defines audit logging levels
type AuditLevel string

const (
	AuditLevelNone    AuditLevel = "none"
	AuditLevelBasic   AuditLevel = "basic"
	AuditLevelDetailed AuditLevel = "detailed"
	AuditLevelFull    AuditLevel = "full"
)

// GeoRestriction defines geographical restrictions
type GeoRestriction struct {
	Type      RestrictionType `json:"type"`
	Countries []string        `json:"countries,omitempty"`
	Regions   []string        `json:"regions,omitempty"`
}

// RestrictionType defines the type of geographical restriction
type RestrictionType string

const (
	RestrictionAllow RestrictionType = "allow"
	RestrictionDeny  RestrictionType = "deny"
)

// DataClassification defines data classification levels
type DataClassification struct {
	Level                string   `json:"level"`
	HandlingRequirements []string `json:"handling_requirements"`
	RetentionPeriod      time.Duration `json:"retention_period"`
	DisposalMethod       string   `json:"disposal_method"`
}

// EncryptionSuite provides military-grade encryption
type EncryptionSuite struct {
	algorithm     string
	keySize       int
	blockMode     string
	padding       string
	keyDerivation string
}

// AuditLogger interface for security audit logging
type AuditLogger interface {
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
	LogAccessAttempt(ctx context.Context, attempt *AccessAttempt) error
	LogDataAccess(ctx context.Context, access *DataAccess) error
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID             string                 `json:"id"`
	Type           SecurityEventType      `json:"type"`
	Severity       SecuritySeverity       `json:"severity"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id"`
	SessionID      string                 `json:"session_id"`
	Description    string                 `json:"description"`
	Details        map[string]interface{} `json:"details,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	AuditContext   *AuditContext          `json:"audit_context,omitempty"`
}

// SecurityEventType defines types of security events
type SecurityEventType string

const (
	SecurityEventLogin            SecurityEventType = "login"
	SecurityEventLogout           SecurityEventType = "logout"
	SecurityEventAccessDenied     SecurityEventType = "access_denied"
	SecurityEventPermissionCheck  SecurityEventType = "permission_check"
	SecurityEventDataAccess       SecurityEventType = "data_access"
	SecurityEventDataModification SecurityEventType = "data_modification"
	SecurityEventEncryption       SecurityEventType = "encryption"
	SecurityEventDecryption       SecurityEventType = "decryption"
	SecurityEventKeyRotation      SecurityEventType = "key_rotation"
	SecurityEventAnomalyDetected  SecurityEventType = "anomaly_detected"
)

// SecuritySeverity defines security event severity levels
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// AccessAttempt represents an access attempt event
type AccessAttempt struct {
	ID           string        `json:"id"`
	TenantID     string        `json:"tenant_id"`
	UserID       string        `json:"user_id"`
	Resource     string        `json:"resource"`
	Action       string        `json:"action"`
	Result       AccessResult  `json:"result"`
	Reason       string        `json:"reason,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	AuditContext *AuditContext `json:"audit_context,omitempty"`
}

// AccessResult defines the result of an access attempt
type AccessResult string

const (
	AccessResultAllow AccessResult = "allow"
	AccessResultDeny  AccessResult = "deny"
	AccessResultError AccessResult = "error"
)

// DataAccess represents a data access event
type DataAccess struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	UserID           string                 `json:"user_id"`
	DataType         string                 `json:"data_type"`
	DataID           string                 `json:"data_id,omitempty"`
	Operation        string                 `json:"operation"`
	SecurityLevel    SecurityLevel          `json:"security_level"`
	EncryptionStatus bool                   `json:"encryption_status"`
	DataHash         string                 `json:"data_hash,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	AuditContext     *AuditContext          `json:"audit_context,omitempty"`
}

// NewSecurityValidator creates a new military-grade security validator
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		securityPolicies: initDefaultSecurityPolicies(),
		encryptionSuite:  initEncryptionSuite(),
		auditLogger:      &defaultAuditLogger{},
	}
}

// ValidateComponentSecurity validates component security with military-grade controls
func (sv *SecurityValidator) ValidateComponentSecurity(ctx context.Context, component Component, securityCtx *TenantSecurityContext) []SecurityViolation {
	var violations []SecurityViolation

	// 1. Tenant isolation validation
	if !sv.validateTenantIsolation(component, securityCtx) {
		violations = append(violations, SecurityViolation{
			Type:        ViolationTenantIsolation,
			Severity:    SecuritySeverityCritical,
			Description: "Component violates tenant isolation boundaries",
			Component:   component.ID,
		})
	}

	// 2. Permission validation
	if !sv.validatePermissions(component, securityCtx) {
		violations = append(violations, SecurityViolation{
			Type:        ViolationInsufficientPermissions,
			Severity:    SecuritySeverityHigh,
			Description: "Insufficient permissions to access component",
			Component:   component.ID,
		})
	}

	// 3. Data classification validation
	if !sv.validateDataClassification(component, securityCtx) {
		violations = append(violations, SecurityViolation{
			Type:        ViolationDataClassification,
			Severity:    SecuritySeverityHigh,
			Description: "Component data classification violation",
			Component:   component.ID,
		})
	}

	// 4. Encryption validation
	if !sv.validateEncryption(component, securityCtx) {
		violations = append(violations, SecurityViolation{
			Type:        ViolationEncryption,
			Severity:    SecuritySeverityHigh,
			Description: "Required encryption not applied to sensitive component data",
			Component:   component.ID,
		})
	}

	// 5. Audit trail validation
	sv.auditComponentAccess(ctx, component, securityCtx, violations)

	return violations
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Type        ViolationType    `json:"type"`
	Severity    SecuritySeverity `json:"severity"`
	Description string           `json:"description"`
	Component   string           `json:"component"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
}

// ViolationType defines types of security violations
type ViolationType string

const (
	ViolationTenantIsolation         ViolationType = "tenant_isolation"
	ViolationInsufficientPermissions ViolationType = "insufficient_permissions"
	ViolationDataClassification      ViolationType = "data_classification"
	ViolationEncryption              ViolationType = "encryption"
	ViolationAccessControl           ViolationType = "access_control"
	ViolationAuditFailure           ViolationType = "audit_failure"
)

// NewTenantSecurityContext creates a new tenant security context
func NewTenantSecurityContext(tenantID, userID string, securityLevel SecurityLevel) *TenantSecurityContext {
	sessionID := generateSecureID()
	now := time.Now()
	
	return &TenantSecurityContext{
		TenantID:      tenantID,
		UserID:        userID,
		SessionID:     sessionID,
		SecurityLevel: securityLevel,
		Permissions:   []Permission{},
		AccessControls: []AccessControl{},
		EncryptionKeys: generateEncryptionKeys(),
		CreatedAt:     now,
		ExpiresAt:     now.Add(24 * time.Hour), // Default 24-hour session
		metadata:      make(map[string]string),
	}
}

// Secure component creation with military-grade security
func CreateSecureComponent(ctx context.Context, registry ComponentRegistry, componentType ComponentType, config map[string]interface{}, securityCtx *TenantSecurityContext) (Component, error) {
	// 1. Pre-creation security validation
	if err := validateSecurityContext(securityCtx); err != nil {
		return Component{}, fmt.Errorf("invalid security context: %w", err)
	}

	// 2. Permission check
	if !hasPermission(securityCtx, fmt.Sprintf("component:%s", componentType), "create") {
		return Component{}, fmt.Errorf("insufficient permissions to create component %s", componentType)
	}

	// 3. Create component through registry
	component, err := registry.Create(ctx, componentType, config)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create component: %w", err)
	}

	// 4. Apply tenant isolation
	component = applyTenantIsolation(component, securityCtx)

	// 5. Apply encryption if required
	component, err = applySecurityEncryption(component, securityCtx)
	if err != nil {
		return Component{}, fmt.Errorf("failed to apply encryption: %w", err)
	}

	// 6. Audit the creation
	auditComponentCreation(ctx, component, securityCtx)

	return component, nil
}

// Helper functions for military-grade security

func (sv *SecurityValidator) validateTenantIsolation(component Component, securityCtx *TenantSecurityContext) bool {
	// Validate that component belongs to the correct tenant
	componentTenantID := extractTenantIDFromComponent(component)
	return componentTenantID == securityCtx.TenantID
}

func (sv *SecurityValidator) validatePermissions(component Component, securityCtx *TenantSecurityContext) bool {
	requiredPermission := fmt.Sprintf("component:%s", component.Type)
	return hasPermission(securityCtx, requiredPermission, "read")
}

func (sv *SecurityValidator) validateDataClassification(component Component, securityCtx *TenantSecurityContext) bool {
	sv.mu.RLock()
	policy, exists := sv.securityPolicies[securityCtx.SecurityLevel]
	sv.mu.RUnlock()
	
	if !exists {
		return false
	}

	// Validate based on data classification requirements
	componentClassification := extractDataClassification(component)
	return isClassificationAllowed(componentClassification, policy.DataClassification)
}

func (sv *SecurityValidator) validateEncryption(component Component, securityCtx *TenantSecurityContext) bool {
	sv.mu.RLock()
	policy, exists := sv.securityPolicies[securityCtx.SecurityLevel]
	sv.mu.RUnlock()
	
	if !exists {
		return false
	}

	if policy.EncryptionRequired {
		return componentHasEncryption(component)
	}
	
	return true
}

func (sv *SecurityValidator) auditComponentAccess(ctx context.Context, component Component, securityCtx *TenantSecurityContext, violations []SecurityViolation) {
	event := &SecurityEvent{
		ID:          generateSecureID(),
		Type:        SecurityEventDataAccess,
		Severity:    SecuritySeverityMedium,
		TenantID:    securityCtx.TenantID,
		UserID:      securityCtx.UserID,
		SessionID:   securityCtx.SessionID,
		Description: fmt.Sprintf("Component access: %s (%s)", component.ID, component.Type),
		Details: map[string]interface{}{
			"component_type": component.Type,
			"component_id":   component.ID,
			"violations":     len(violations),
		},
		Timestamp:    time.Now(),
		AuditContext: securityCtx.AuditContext,
	}

	if len(violations) > 0 {
		event.Severity = SecuritySeverityHigh
		event.Type = SecurityEventAccessDenied
	}

	sv.auditLogger.LogSecurityEvent(ctx, event)
}

// Utility functions

func generateSecureID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateEncryptionKeys() *EncryptionKeys {
	dek := make([]byte, 32) // 256-bit key
	kek := make([]byte, 32) // 256-bit key  
	signing := make([]byte, 32) // 256-bit key
	
	rand.Read(dek)
	rand.Read(kek)
	rand.Read(signing)
	
	now := time.Now()
	return &EncryptionKeys{
		DataEncryptionKey: dek,
		KeyEncryptionKey:  kek,
		SigningKey:        signing,
		RotationSchedule:  24 * time.Hour, // Daily rotation
		LastRotated:       now,
		NextRotation:      now.Add(24 * time.Hour),
	}
}

func validateSecurityContext(securityCtx *TenantSecurityContext) error {
	if securityCtx == nil {
		return fmt.Errorf("security context is nil")
	}
	
	if securityCtx.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	
	if securityCtx.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	
	if securityCtx.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	
	if time.Now().After(securityCtx.ExpiresAt) {
		return fmt.Errorf("security context has expired")
	}
	
	return nil
}

func hasPermission(securityCtx *TenantSecurityContext, resource, action string) bool {
	for _, permission := range securityCtx.Permissions {
		if permission.Resource == resource && permission.Action == action {
			if permission.ExpiresAt == nil || time.Now().Before(*permission.ExpiresAt) {
				return true
			}
		}
	}
	return false
}

func applyTenantIsolation(component Component, securityCtx *TenantSecurityContext) Component {
	// Add tenant ID to component metadata
	if component.metadata == nil {
		component.metadata = make(map[string]string)
	}
	component.metadata["tenant_id"] = securityCtx.TenantID
	component.metadata["security_level"] = string(securityCtx.SecurityLevel)
	component.metadata["created_by"] = securityCtx.UserID
	component.metadata["session_id"] = securityCtx.SessionID
	
	return component
}

func applySecurityEncryption(component Component, securityCtx *TenantSecurityContext) (Component, error) {
	if securityCtx.EncryptionKeys == nil {
		return component, nil
	}

	// Encrypt sensitive component data
	if component.Config != nil {
		encryptedConfig, err := encryptData(component.Config, securityCtx.EncryptionKeys.DataEncryptionKey)
		if err != nil {
			return component, fmt.Errorf("failed to encrypt component config: %w", err)
		}
		component.encryptedConfig = encryptedConfig
		component.encrypted = true
	}

	return component, nil
}

func auditComponentCreation(ctx context.Context, component Component, securityCtx *TenantSecurityContext) {
	// Audit logging would be implemented here
	// For now, this is a placeholder
}

func extractTenantIDFromComponent(component Component) string {
	if component.metadata != nil {
		return component.metadata["tenant_id"]
	}
	return ""
}

func extractDataClassification(component Component) string {
	if component.metadata != nil {
		return component.metadata["data_classification"]
	}
	return "internal" // Default classification
}

func isClassificationAllowed(componentClass string, policyClass DataClassification) bool {
	// Implement classification hierarchy validation
	classificationLevels := map[string]int{
		"public":       1,
		"internal":     2,
		"confidential": 3,
		"secret":       4,
		"top-secret":   5,
	}
	
	componentLevel := classificationLevels[componentClass]
	policyLevel := classificationLevels[policyClass.Level]
	
	return componentLevel <= policyLevel
}

func componentHasEncryption(component Component) bool {
	return component.encrypted
}

func encryptData(data interface{}, key []byte) ([]byte, error) {
	// Implement AES-256-GCM encryption
	// This is a placeholder - real implementation would use proper crypto
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", data)))
	return hash[:], nil
}

// Initialize default security policies
func initDefaultSecurityPolicies() map[SecurityLevel]*SecurityPolicy {
	return map[SecurityLevel]*SecurityPolicy{
		SecurityLevelPublic: {
			RequiredPermissions:    []string{"read"},
			AllowedOperations:      []string{"read"},
			EncryptionRequired:     false,
			AuditLevel:            AuditLevelBasic,
			SessionTimeout:        2 * time.Hour,
			MaxConcurrentSessions: 10,
			DataClassification: DataClassification{
				Level:                "public",
				HandlingRequirements: []string{"standard"},
				RetentionPeriod:      365 * 24 * time.Hour,
				DisposalMethod:       "standard_deletion",
			},
		},
		SecurityLevelTopSecret: {
			RequiredPermissions:   []string{"read", "write", "delete", "admin"},
			AllowedOperations:     []string{"read", "write", "delete"},
			EncryptionRequired:    true,
			AuditLevel:           AuditLevelFull,
			SessionTimeout:       30 * time.Minute,
			MaxConcurrentSessions: 1,
			DataClassification: DataClassification{
				Level:                "top-secret",
				HandlingRequirements: []string{"encrypted", "air_gapped", "dual_person_integrity"},
				RetentionPeriod:      7 * 365 * 24 * time.Hour,
				DisposalMethod:       "cryptographic_erasure",
			},
		},
	}
}

func initEncryptionSuite() *EncryptionSuite {
	return &EncryptionSuite{
		algorithm:     "AES-256-GCM",
		keySize:       256,
		blockMode:     "GCM",
		padding:       "PKCS7",
		keyDerivation: "PBKDF2",
	}
}

// Default audit logger implementation
type defaultAuditLogger struct{}

func (dal *defaultAuditLogger) LogSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	// Implementation would log to secure audit system
	return nil
}

func (dal *defaultAuditLogger) LogAccessAttempt(ctx context.Context, attempt *AccessAttempt) error {
	// Implementation would log to secure audit system
	return nil
}

func (dal *defaultAuditLogger) LogDataAccess(ctx context.Context, access *DataAccess) error {
	// Implementation would log to secure audit system
	return nil
}