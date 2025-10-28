package activities

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/core/identity"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// AttributeCollectionActivities handles attribute collection for ABAC evaluation
type AttributeCollectionActivities struct {
	attributeRepo   repository.AttributeRepository
	identityService identity.Service
	tenantService   tenant.Service
	logger          logger.Logger
	metrics         metrics.MetricsProvider
	tracer          tracing.Service
}

// NewAttributeCollectionActivities creates a new AttributeCollectionActivities instance
func NewAttributeCollectionActivities(
	attributeRepo repository.AttributeRepository,
	identityService identity.Service,
	tenantService tenant.Service,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *AttributeCollectionActivities {
	return &AttributeCollectionActivities{
		attributeRepo:   attributeRepo,
		identityService: identityService,
		tenantService:   tenantService,
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
	}
}

// CollectUserAttributesInput represents input for user attribute collection
type CollectUserAttributesInput struct {
	UserID    uuid.UUID  `json:"user_id"`
	EntityID  *uuid.UUID `json:"entity_id,omitempty"`
	RequestID string     `json:"request_id"`
}

// CollectUserAttributesOutput represents output of user attribute collection
type CollectUserAttributesOutput struct {
	Attributes map[string]any `json:"attributes"`
	Source     string         `json:"source"`
	Timestamp  time.Time      `json:"timestamp"`
}

// CollectUserAttributes collects user attributes from various sources
func (a *AttributeCollectionActivities) CollectUserAttributes(ctx context.Context, input *CollectUserAttributesInput) (*CollectUserAttributesOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.CollectUserAttributes",
		tracing.WithAttributes(
			attribute.String("user_id", input.UserID.String()),
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Starting user attribute collection",
		logger.Fields{
			"user_id":    input.UserID,
			"request_id": input.RequestID,
		})

	attributes := make(map[string]any)

	// Get user basic attributes
	userAttrs, err := a.collectBasicUserAttributes(ctx, input.UserID)
	if err != nil {
		a.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		a.metrics.IncrementCounter("user_attribute_collection_failed", metrics.Fields{"reason": "basic_attributes_failed"})
		return nil, errors.NewBusinessErrorWithContext(ctx, "USER_ATTRIBUTE_COLLECTION_FAILED", "Failed to collect basic user attributes").WithDetail("error", err.Error())
	}

	// Merge basic user attributes
	for k, v := range userAttrs {
		attributes[fmt.Sprintf("user.%s", k)] = v
	}

	// Get person attributes if user is linked to a person
	personAttrs, err := a.collectPersonAttributes(ctx, input.UserID)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to collect person attributes",
			logger.Fields{"user_id": input.UserID, "error": err.Error()})
		// Don't fail the entire collection for person attributes
	} else {
		for k, v := range personAttrs {
			attributes[fmt.Sprintf("person.%s", k)] = v
		}
	}

	// Get employee attributes if user is an employee
	employeeAttrs, err := a.collectEmployeeAttributes(ctx, input.UserID)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to collect employee attributes",
			logger.Fields{"user_id": input.UserID, "error": err.Error()})
		// Don't fail the entire collection for employee attributes
	} else {
		for k, v := range employeeAttrs {
			attributes[fmt.Sprintf("employee.%s", k)] = v
		}
	}

	// Get role-based attributes
	roleAttrs, err := a.collectRoleAttributes(ctx, input.UserID, input.EntityID)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to collect role attributes",
			logger.Fields{"user_id": input.UserID, "error": err.Error()})
	} else {
		for k, v := range roleAttrs {
			attributes[fmt.Sprintf("role.%s", k)] = v
		}
	}

	// Get stored attribute values
	storedAttrs, err := a.collectStoredAttributes(ctx, input.UserID, types.AttributeCategoryUser)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to collect stored attributes",
			logger.Fields{"user_id": input.UserID, "error": err.Error()})
	} else {
		for k, v := range storedAttrs {
			attributes[fmt.Sprintf("custom.%s", k)] = v
		}
	}

	duration := time.Since(startTime)
	span.SetAttributes(attribute.Int("attributes_count", len(attributes)))

	a.metrics.ObserveHistogram("abac_user_attribute_collection_duration_seconds", duration.Seconds(),
		metrics.Fields{"success": "true"})
	a.metrics.SetGauge("abac_user_attributes_collected", float64(len(attributes)),
		metrics.Fields{"user_id": input.UserID.String()})

	a.logger.InfoContext(ctx, "User attribute collection completed",
		logger.Fields{
			"user_id":          input.UserID,
			"attributes_count": len(attributes),
			"duration_ms":      duration.Milliseconds(),
			"request_id":       input.RequestID,
		})

	return &CollectUserAttributesOutput{
		Attributes: attributes,
		Source:     "identity_service",
		Timestamp:  time.Now(),
	}, nil
}

// collectBasicUserAttributes collects basic user information
func (a *AttributeCollectionActivities) collectBasicUserAttributes(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	user, err := a.identityService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	attributes := map[string]any{
		"id":         user.ID.String(),
		"email":      user.Email,
		"status":     user.AccountStatus,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	// Add optional fields if present
	// Note: FirstName is not available in User model, it's in Person model
	// We would need to fetch the Person record separately if needed
	// For now, we'll skip this field
	// Note: LastName and PhoneNumber are in Person model, not User model
	// These would need to be fetched separately if needed
	if user.LastLoginAt != nil {
		attributes["last_login_at"] = *user.LastLoginAt
		attributes["days_since_last_login"] = time.Since(*user.LastLoginAt).Hours() / 24
	}

	return attributes, nil
}

// collectPersonAttributes collects person-specific attributes
func (a *AttributeCollectionActivities) collectPersonAttributes(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	// Get user by ID to check if it has person-related data
	_, err := a.identityService.GetUserByID(ctx, userID)
	if err != nil {
		return map[string]any{}, nil // Not found or error
	}

	// This is a placeholder - adjust based on actual person model
	attributes := map[string]any{
		"is_person": true,
		"type":      "person",
	}

	return attributes, nil
}

// collectEmployeeAttributes collects employee-specific attributes
func (a *AttributeCollectionActivities) collectEmployeeAttributes(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	// Get user by ID to check if it has employee-related data
	_, err := a.identityService.GetUserByID(ctx, userID)
	if err != nil {
		return map[string]any{}, nil // Not found or error
	}

	// This is a placeholder - adjust based on actual employee model
	attributes := map[string]any{
		"is_employee": true,
		"type":        "employee",
	}

	return attributes, nil
}

// collectRoleAttributes collects role-based attributes
func (a *AttributeCollectionActivities) collectRoleAttributes(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (map[string]any, error) {
	// This is a placeholder for role collection
	// You'll need to implement based on your role/permission system
	attributes := map[string]any{
		"has_roles": false,
	}

	return attributes, nil
}

// collectStoredAttributes collects custom stored attributes
func (a *AttributeCollectionActivities) collectStoredAttributes(ctx context.Context, entityID uuid.UUID, category types.AttributeCategory) (map[string]any, error) {
	attributeValues, err := a.attributeRepo.GetAttributeValuesByEntity(ctx, entityID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored attributes: %w", err)
	}

	attributes := make(map[string]any)
	for _, attrValue := range attributeValues {
		// You'll need to implement attribute definition lookup to get the name
		attributes[fmt.Sprintf("attr_%s", attrValue.DefinitionID.String())] = attrValue.Value
	}

	return attributes, nil
}

// CollectResourceAttributesInput represents input for resource attribute collection
type CollectResourceAttributesInput struct {
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	RequestID    string     `json:"request_id"`
}

// CollectResourceAttributesOutput represents output of resource attribute collection
type CollectResourceAttributesOutput struct {
	Attributes map[string]any `json:"attributes"`
	Source     string         `json:"source"`
	Timestamp  time.Time      `json:"timestamp"`
}

// CollectResourceAttributes collects resource attributes
func (a *AttributeCollectionActivities) CollectResourceAttributes(ctx context.Context, input *CollectResourceAttributesInput) (*CollectResourceAttributesOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.CollectResourceAttributes",
		tracing.WithAttributes(
			attribute.String("resource_type", input.ResourceType),
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Starting resource attribute collection",
		logger.Fields{
			"resource_type": input.ResourceType,
			"resource_id":   input.ResourceID,
			"request_id":    input.RequestID,
		})

	attributes := make(map[string]any)

	// Add basic resource attributes
	attributes["type"] = input.ResourceType
	if input.ResourceID != nil {
		attributes["id"] = input.ResourceID.String()
	}
	if input.EntityID != nil {
		attributes["entity_id"] = input.EntityID.String()
	}

	// Get tenant context
	currentTenant, err := a.tenantService.GetCurrentTenant(ctx)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to get current tenant for resource attributes",
			logger.Fields{"error": err.Error()})
	} else {
		attributes["tenant_id"] = currentTenant.ID.String()
		attributes["tenant_name"] = currentTenant.Name
		attributes["tenant_status"] = currentTenant.Status
	}

	// Collect resource-specific attributes based on type
	resourceSpecificAttrs, err := a.collectResourceSpecificAttributes(ctx, input)
	if err != nil {
		a.logger.WarnContext(ctx, "Failed to collect resource-specific attributes",
			logger.Fields{"resource_type": input.ResourceType, "error": err.Error()})
	} else {
		for k, v := range resourceSpecificAttrs {
			attributes[k] = v
		}
	}

	// Get stored resource attributes
	if input.ResourceID != nil {
		storedAttrs, err := a.collectStoredAttributes(ctx, *input.ResourceID, types.AttributeCategoryResource)
		if err != nil {
			a.logger.WarnContext(ctx, "Failed to collect stored resource attributes",
				logger.Fields{"resource_id": input.ResourceID, "error": err.Error()})
		} else {
			for k, v := range storedAttrs {
				attributes[fmt.Sprintf("custom.%s", k)] = v
			}
		}
	}

	duration := time.Since(startTime)
	span.SetAttributes(attribute.Int("attributes_count", len(attributes)))

	a.metrics.ObserveHistogram("abac_resource_attribute_collection_duration_seconds", duration.Seconds(),
		metrics.Fields{"resource_type": input.ResourceType, "success": "true"})

	a.logger.InfoContext(ctx, "Resource attribute collection completed",
		logger.Fields{
			"resource_type":    input.ResourceType,
			"attributes_count": len(attributes),
			"duration_ms":      duration.Milliseconds(),
			"request_id":       input.RequestID,
		})

	return &CollectResourceAttributesOutput{
		Attributes: attributes,
		Source:     "resource_service",
		Timestamp:  time.Now(),
	}, nil
}

// collectResourceSpecificAttributes collects attributes specific to resource type
func (a *AttributeCollectionActivities) collectResourceSpecificAttributes(ctx context.Context, input *CollectResourceAttributesInput) (map[string]any, error) {
	attributes := make(map[string]any)

	switch input.ResourceType {
	case "user":
		attributes["category"] = "identity"
		attributes["access_level"] = "personal"

	case "document":
		attributes["category"] = "content"
		attributes["access_level"] = "restricted"

	case "report":
		attributes["category"] = "analytics"
		attributes["access_level"] = "business"

	case "system":
		attributes["category"] = "infrastructure"
		attributes["access_level"] = "administrative"

	default:
		attributes["category"] = "general"
		attributes["access_level"] = "standard"
	}

	return attributes, nil
}

// BuildEnvironmentContextInput represents input for environment context building
type BuildEnvironmentContextInput struct {
	IPAddress   string         `json:"ip_address,omitempty"`
	UserAgent   string         `json:"user_agent,omitempty"`
	SessionData map[string]any `json:"session_data,omitempty"`
	RequestTime time.Time      `json:"request_time"`
	RequestID   string         `json:"request_id"`
}

// BuildEnvironmentContextOutput represents output of environment context building
type BuildEnvironmentContextOutput struct {
	Context   map[string]any `json:"context"`
	Source    string         `json:"source"`
	Timestamp time.Time      `json:"timestamp"`
}

// BuildEnvironmentContext builds environment context for policy evaluation
func (a *AttributeCollectionActivities) BuildEnvironmentContext(ctx context.Context, input *BuildEnvironmentContextInput) (*BuildEnvironmentContextOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, "abac.activities.BuildEnvironmentContext",
		tracing.WithAttributes(
			attribute.String("request_id", input.RequestID),
		))
	defer span.End()

	startTime := time.Now()

	a.logger.InfoContext(ctx, "Starting environment context building",
		logger.Fields{
			"ip_address": input.IPAddress,
			"request_id": input.RequestID,
		})

	context := make(map[string]any)

	// Time-based attributes
	timeAttrs := a.buildTimeAttributes(input.RequestTime)
	for k, v := range timeAttrs {
		context[fmt.Sprintf("time.%s", k)] = v
	}

	// Network-based attributes
	if input.IPAddress != "" {
		networkAttrs := a.buildNetworkAttributes(input.IPAddress)
		for k, v := range networkAttrs {
			context[fmt.Sprintf("network.%s", k)] = v
		}
	}

	// Device-based attributes
	if input.UserAgent != "" {
		deviceAttrs := a.buildDeviceAttributes(input.UserAgent)
		for k, v := range deviceAttrs {
			context[fmt.Sprintf("device.%s", k)] = v
		}
	}

	// Session-based attributes
	if input.SessionData != nil {
		sessionAttrs := a.buildSessionAttributes(input.SessionData)
		for k, v := range sessionAttrs {
			context[fmt.Sprintf("session.%s", k)] = v
		}
	}

	// Risk-based attributes
	riskAttrs := a.buildRiskAttributes(ctx, input)
	for k, v := range riskAttrs {
		context[fmt.Sprintf("risk.%s", k)] = v
	}

	duration := time.Since(startTime)
	span.SetAttributes(attribute.Int("context_attributes_count", len(context)))

	a.metrics.ObserveHistogram("abac_environment_context_building_duration_seconds", duration.Seconds(),
		metrics.Fields{"success": "true"})

	a.logger.InfoContext(ctx, "Environment context building completed",
		logger.Fields{
			"context_attributes": len(context),
			"duration_ms":        duration.Milliseconds(),
			"request_id":         input.RequestID,
		})

	return &BuildEnvironmentContextOutput{
		Context:   context,
		Source:    "environment_collector",
		Timestamp: time.Now(),
	}, nil
}

// buildTimeAttributes builds time-based context attributes
func (a *AttributeCollectionActivities) buildTimeAttributes(requestTime time.Time) map[string]any {
	now := requestTime
	if requestTime.IsZero() {
		now = time.Now()
	}

	return map[string]any{
		"current_time":      now,
		"hour":              now.Hour(),
		"day_of_week":       now.Weekday().String(),
		"day_of_month":      now.Day(),
		"month":             now.Month().String(),
		"year":              now.Year(),
		"is_weekend":        now.Weekday() == time.Saturday || now.Weekday() == time.Sunday,
		"is_business_hours": now.Hour() >= 9 && now.Hour() < 17 && now.Weekday() != time.Saturday && now.Weekday() != time.Sunday,
		"quarter":           (now.Month()-1)/3 + 1,
		"unix_timestamp":    now.Unix(),
	}
}

// buildNetworkAttributes builds network-based context attributes
func (a *AttributeCollectionActivities) buildNetworkAttributes(ipAddress string) map[string]any {
	attributes := map[string]any{
		"ip_address": ipAddress,
	}

	// Parse IP address
	ip := net.ParseIP(ipAddress)
	if ip != nil {
		attributes["is_ipv4"] = ip.To4() != nil
		attributes["is_ipv6"] = ip.To4() == nil
		attributes["is_private"] = a.isPrivateIP(ip)
		attributes["is_loopback"] = ip.IsLoopback()
	}

	// Simple geolocation (you might want to use a proper geolocation service)
	geoAttrs := a.buildSimpleGeolocation(ipAddress)
	for k, v := range geoAttrs {
		attributes[k] = v
	}

	return attributes
}

// buildDeviceAttributes builds device-based context attributes
func (a *AttributeCollectionActivities) buildDeviceAttributes(userAgent string) map[string]any {
	attributes := map[string]any{
		"user_agent": userAgent,
	}

	// Simple user agent parsing (you might want to use a proper user agent parser)
	lowerUA := strings.ToLower(userAgent)

	// Browser detection
	if strings.Contains(lowerUA, "chrome") {
		attributes["browser"] = "chrome"
	} else if strings.Contains(lowerUA, "firefox") {
		attributes["browser"] = "firefox"
	} else if strings.Contains(lowerUA, "safari") {
		attributes["browser"] = "safari"
	} else if strings.Contains(lowerUA, "edge") {
		attributes["browser"] = "edge"
	} else {
		attributes["browser"] = "unknown"
	}

	// OS detection
	if strings.Contains(lowerUA, "windows") {
		attributes["os"] = "windows"
	} else if strings.Contains(lowerUA, "mac") {
		attributes["os"] = "macos"
	} else if strings.Contains(lowerUA, "linux") {
		attributes["os"] = "linux"
	} else if strings.Contains(lowerUA, "android") {
		attributes["os"] = "android"
	} else if strings.Contains(lowerUA, "ios") {
		attributes["os"] = "ios"
	} else {
		attributes["os"] = "unknown"
	}

	// Device type detection
	if strings.Contains(lowerUA, "mobile") || strings.Contains(lowerUA, "android") || strings.Contains(lowerUA, "iphone") {
		attributes["device_type"] = "mobile"
	} else if strings.Contains(lowerUA, "tablet") || strings.Contains(lowerUA, "ipad") {
		attributes["device_type"] = "tablet"
	} else {
		attributes["device_type"] = "desktop"
	}

	return attributes
}

// buildSessionAttributes builds session-based context attributes
func (a *AttributeCollectionActivities) buildSessionAttributes(sessionData map[string]any) map[string]any {
	attributes := make(map[string]any)

	for k, v := range sessionData {
		switch k {
		case "session_id", "csrf_token", "auth_token":
			// Don't include sensitive session data in attributes
			continue
		default:
			attributes[k] = v
		}
	}

	// Add session metadata
	if sessionID, ok := sessionData["session_id"]; ok {
		attributes["has_session"] = true
		attributes["session_id_length"] = len(fmt.Sprintf("%v", sessionID))
	}

	return attributes
}

// buildRiskAttributes builds risk-based context attributes
func (a *AttributeCollectionActivities) buildRiskAttributes(ctx context.Context, input *BuildEnvironmentContextInput) map[string]any {
	attributes := map[string]any{
		"risk_score": 0.0,
		"risk_level": "low",
	}

	var riskFactors []string
	riskScore := 0.0

	// Check for suspicious IP
	if input.IPAddress != "" {
		ip := net.ParseIP(input.IPAddress)
		if ip != nil && !a.isPrivateIP(ip) {
			riskFactors = append(riskFactors, "external_ip")
			riskScore += 0.1
		}
	}

	// Check for unusual access time
	now := input.RequestTime
	if now.IsZero() {
		now = time.Now()
	}
	if now.Hour() < 6 || now.Hour() > 22 {
		riskFactors = append(riskFactors, "unusual_time")
		riskScore += 0.1
	}

	// Check for weekend access
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		riskFactors = append(riskFactors, "weekend_access")
		riskScore += 0.05
	}

	// Determine risk level
	riskLevel := "low"
	if riskScore > 0.5 {
		riskLevel = "high"
	} else if riskScore > 0.2 {
		riskLevel = "medium"
	}

	attributes["risk_score"] = riskScore
	attributes["risk_level"] = riskLevel
	attributes["risk_factors"] = riskFactors

	return attributes
}

// isPrivateIP checks if an IP address is private
func (a *AttributeCollectionActivities) isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// buildSimpleGeolocation provides basic geolocation (placeholder implementation)
func (a *AttributeCollectionActivities) buildSimpleGeolocation(ipAddress string) map[string]any {
	// This is a very basic implementation
	// In production, you would use a proper geolocation service like MaxMind GeoIP2
	attributes := map[string]any{
		"country":  "unknown",
		"region":   "unknown",
		"city":     "unknown",
		"timezone": "UTC",
	}

	// Simple private IP detection
	ip := net.ParseIP(ipAddress)
	if ip != nil && a.isPrivateIP(ip) {
		attributes["country"] = "private"
		attributes["region"] = "private"
		attributes["city"] = "private"
		attributes["timezone"] = "local"
	}

	return attributes
}
