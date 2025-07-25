package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac/models"
)

// AttributeCollectionActivities contains all activities related to collecting attributes for ABAC evaluation
type AttributeCollectionActivities struct {
	// Dependencies will be injected
	// identityService identity.Service
	// resourceService resource.Service (when it exists)
	// geoService      geo.Service
	// deviceService   device.Service
}

// NewAttributeCollectionActivities creates a new instance of attribute collection activities
func NewAttributeCollectionActivities() *AttributeCollectionActivities {
	return &AttributeCollectionActivities{
		// Inject dependencies
	}
}

// CollectUserAttributes collects all user-related attributes from person, employee, and user tables
func (a *AttributeCollectionActivities) CollectUserAttributes(ctx context.Context, userID uuid.UUID) (*models.AttributeCollectionResult, error) {
	startTime := time.Now()
	
	result := &models.AttributeCollectionResult{
		UserAttributes:        make(map[string]interface{}),
		PersonAttributes:      make(map[string]interface{}),
		EmployeeAttributes:    make(map[string]interface{}),
		EnvironmentAttributes: make(map[string]interface{}),
	}

	// TODO: Replace with actual identity service calls
	// This is a placeholder implementation
	
	// Collect user attributes
	// user, err := a.identityService.GetUserByID(ctx, userID)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get user: %w", err)
	// }

	// Mock user attributes for now
	result.UserAttributes = map[string]interface{}{
		"user_id":     userID.String(),
		"user_type":   "INTERNAL",
		"is_active":   true,
		"mfa_enabled": false,
		// Add more user attributes from actual user data
	}

	// Collect person attributes if user is linked to a person
	// if user.PersonID != nil {
	//     person, err := a.identityService.GetPersonByID(ctx, *user.PersonID)
	//     if err == nil {
	//         result.PersonAttributes = person.SecurityAttributes
	//         result.PersonAttributes["person_type"] = person.PersonType
	//         result.PersonAttributes["first_name"] = person.FirstName
	//         result.PersonAttributes["last_name"] = person.LastName
	//     }
	// }

	// Mock person attributes
	result.PersonAttributes = map[string]interface{}{
		"person_type":     "EMPLOYEE",
		"security_level":  "STANDARD",
		"clearance_level": 3,
		"department":      "ENGINEERING",
	}

	// Collect employee attributes if user is linked to an employee
	// if user.EmployeeID != nil {
	//     employee, err := a.identityService.GetEmployeeByID(ctx, *user.EmployeeID)
	//     if err == nil {
	//         result.EmployeeAttributes = employee.AccessAttributes
	//         result.EmployeeAttributes["security_level"] = employee.SecurityLevel
	//         result.EmployeeAttributes["position_title"] = employee.PositionTitle
	//         result.EmployeeAttributes["employment_status"] = employee.Status
	//     }
	// }

	// Mock employee attributes
	result.EmployeeAttributes = map[string]interface{}{
		"security_level":     5,
		"position_title":     "Senior Software Engineer",
		"employment_status":  "ACTIVE",
		"hire_date":         "2022-01-15",
		"manager_id":        uuid.New().String(),
		"access_level":      "ELEVATED",
	}

	// Add computed attributes
	result.UserAttributes["combined_security_level"] = result.EmployeeAttributes["security_level"]
	result.UserAttributes["is_manager"] = false // TODO: Calculate based on subordinates
	
	result.CollectionTimeMS = time.Since(startTime).Milliseconds()
	return result, nil
}

// CollectResourceAttributes collects attributes related to the requested resource
func (a *AttributeCollectionActivities) CollectResourceAttributes(ctx context.Context, resourceName string, entityID uuid.UUID) (map[string]interface{}, error) {
	// TODO: Implement actual resource attribute collection
	// This would typically involve:
	// 1. Looking up resource definition in resources table
	// 2. Getting resource-specific metadata
	// 3. Calculating dynamic resource attributes

	// Mock resource attributes for now
	resourceAttributes := map[string]interface{}{
		"resource_name":        resourceName,
		"resource_type":        "API_ENDPOINT",
		"sensitivity_level":    "MEDIUM",
		"entity_id":           entityID.String(),
		"requires_approval":    false,
		"owner_department":     "ENGINEERING",
		"classification":       "INTERNAL",
		"access_pattern":       "READ_WRITE",
		"compliance_tags":      []string{"SOX", "GDPR"},
	}

	// Add resource-specific attributes based on resource type
	switch resourceName {
	case "financial_reports":
		resourceAttributes["sensitivity_level"] = "HIGH"
		resourceAttributes["requires_approval"] = true
		resourceAttributes["compliance_tags"] = []string{"SOX", "PCI"}
	case "user_management":
		resourceAttributes["sensitivity_level"] = "HIGH"
		resourceAttributes["requires_approval"] = true
		resourceAttributes["compliance_tags"] = []string{"GDPR", "CCPA"}
	case "system_logs":
		resourceAttributes["sensitivity_level"] = "MEDIUM"
		resourceAttributes["access_pattern"] = "READ_ONLY"
	}

	return resourceAttributes, nil
}

// EnrichEnvironmentContext enriches the context with environment-specific attributes
func (a *AttributeCollectionActivities) EnrichEnvironmentContext(ctx context.Context, baseContext map[string]interface{}) (*models.ContextEnrichmentResult, error) {
	startTime := time.Now()
	
	result := &models.ContextEnrichmentResult{
		TimeContext:    make(map[string]interface{}),
		DeviceContext:  make(map[string]interface{}),
		LocationContext: make(map[string]interface{}),
		NetworkContext: make(map[string]interface{}),
		RiskContext:    make(map[string]interface{}),
	}

	// Enrich time context
	now := time.Now()
	result.TimeContext = map[string]interface{}{
		"current_time":      now.Format(time.RFC3339),
		"hour_of_day":       now.Hour(),
		"day_of_week":       int(now.Weekday()),
		"is_business_hours": isBusinessHours(now),
		"is_weekend":        now.Weekday() == time.Saturday || now.Weekday() == time.Sunday,
		"timezone":          now.Location().String(),
	}

	// Enrich device context from base context
	if userAgent, ok := baseContext["user_agent"].(string); ok {
		deviceInfo := parseUserAgent(userAgent)
		result.DeviceContext = deviceInfo
	}

	// Enrich location context from IP address
	if ipAddress, ok := baseContext["ip_address"].(string); ok {
		locationInfo, err := a.getLocationFromIP(ipAddress)
		if err == nil {
			result.LocationContext = locationInfo
		}
	}

	// Enrich network context
	if ipAddress, ok := baseContext["ip_address"].(string); ok {
		networkInfo := a.analyzeNetworkContext(ipAddress)
		result.NetworkContext = networkInfo
	}

	// Calculate risk context
	riskScore := a.calculateRiskScore(baseContext, result)
	result.RiskContext = map[string]interface{}{
		"risk_score":       riskScore,
		"risk_level":       getRiskLevel(riskScore),
		"anomaly_detected": riskScore > 70,
		"risk_factors":     a.identifyRiskFactors(baseContext, result),
	}

	result.EnrichmentTimeMS = time.Since(startTime).Milliseconds()
	return result, nil
}

// Helper functions

func isBusinessHours(t time.Time) bool {
	hour := t.Hour()
	weekday := t.Weekday()
	
	// Business hours: Monday-Friday, 9 AM - 5 PM
	if weekday >= time.Monday && weekday <= time.Friday {
		return hour >= 9 && hour < 17
	}
	return false
}

func parseUserAgent(userAgent string) map[string]interface{} {
	// TODO: Implement proper user agent parsing
	// For now, return mock device info
	return map[string]interface{}{
		"device_type":      "desktop",
		"operating_system": "windows",
		"browser":          "chrome",
		"is_mobile":        false,
		"is_trusted":       false, // TODO: Check against trusted devices
	}
}

func (a *AttributeCollectionActivities) getLocationFromIP(ipAddress string) (map[string]interface{}, error) {
	// TODO: Implement actual IP geolocation
	// For now, return mock location data
	return map[string]interface{}{
		"country":       "US",
		"region":        "CA",
		"city":          "San Francisco",
		"latitude":      37.7749,
		"longitude":     -122.4194,
		"is_trusted":    true, // TODO: Check against trusted locations
		"is_corporate":  true, // TODO: Check if IP is in corporate ranges
	}, nil
}

func (a *AttributeCollectionActivities) analyzeNetworkContext(ipAddress string) map[string]interface{} {
	// TODO: Implement actual network analysis
	return map[string]interface{}{
		"ip_address":        ipAddress,
		"is_internal":       true,  // TODO: Check if IP is internal
		"is_vpn":           false, // TODO: Detect VPN usage
		"network_segment":   "corporate",
		"security_zone":     "trusted",
	}
}

func (a *AttributeCollectionActivities) calculateRiskScore(baseContext map[string]interface{}, enrichedContext *models.ContextEnrichmentResult) int {
	riskScore := 0
	
	// Increase risk for non-business hours
	if !enrichedContext.TimeContext["is_business_hours"].(bool) {
		riskScore += 20
	}
	
	// Increase risk for untrusted devices
	if deviceContext := enrichedContext.DeviceContext; len(deviceContext) > 0 {
		if trusted, ok := deviceContext["is_trusted"].(bool); ok && !trusted {
			riskScore += 30
		}
	}
	
	// Increase risk for external locations
	if locationContext := enrichedContext.LocationContext; len(locationContext) > 0 {
		if trusted, ok := locationContext["is_trusted"].(bool); ok && !trusted {
			riskScore += 25
		}
	}
	
	// Increase risk for external networks
	if networkContext := enrichedContext.NetworkContext; len(networkContext) > 0 {
		if internal, ok := networkContext["is_internal"].(bool); ok && !internal {
			riskScore += 35
		}
	}
	
	// Cap at 100
	if riskScore > 100 {
		riskScore = 100
	}
	
	return riskScore
}

func getRiskLevel(score int) string {
	switch {
	case score >= 80:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 40:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func (a *AttributeCollectionActivities) identifyRiskFactors(baseContext map[string]interface{}, enrichedContext *models.ContextEnrichmentResult) []string {
	var factors []string
	
	// Check for various risk factors
	if !enrichedContext.TimeContext["is_business_hours"].(bool) {
		factors = append(factors, "after_hours_access")
	}
	
	if enrichedContext.TimeContext["is_weekend"].(bool) {
		factors = append(factors, "weekend_access")
	}
	
	if deviceContext := enrichedContext.DeviceContext; len(deviceContext) > 0 {
		if trusted, ok := deviceContext["is_trusted"].(bool); ok && !trusted {
			factors = append(factors, "untrusted_device")
		}
	}
	
	if locationContext := enrichedContext.LocationContext; len(locationContext) > 0 {
		if trusted, ok := locationContext["is_trusted"].(bool); ok && !trusted {
			factors = append(factors, "untrusted_location")
		}
	}
	
	return factors
}