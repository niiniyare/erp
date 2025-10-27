package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// AUDIT & SECURITY MONITORING TYPES
// ============================================================================

// AuditLog represents a comprehensive audit trail entry for security monitoring.
// Captures all user actions, system events, and access decisions with full context.
var AuditLog = ResultType("application/vnd.erp.audit", func() {
	Description("Comprehensive audit log entry for security monitoring, compliance tracking, and forensic analysis")

	Attributes(func() {
		Field(1, "id", String, "Unique audit log identifier", func() {
			Format(FormatUUID)
			Example("audit-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for audit record")
		})

		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})

		Field(3, "event_type", String, "Specific type of audited event", func() {
			Pattern("^[A-Z][A-Z0-9_]{2,49}$")
			Example("USER_LOGIN_SUCCESS")
			Description("Standardized event classification")
		})

		Field(4, "event_category", String, "High-level event categorization", func() {
			Enum("ACCESS", "ADMIN", "DATA", "AUTH", "SYSTEM", "COMPLIANCE", "SECURITY", "BUSINESS")
			Example("AUTH")
			Description("Broad classification for filtering and reporting")
		})

		Field(5, "severity", String, "Event severity level for alerting", func() {
			Enum("LOW", "INFO", "WARN", "HIGH", "CRITICAL")
			Default("INFO")
			Example("HIGH")
			Description("Risk level requiring different response procedures")
		})

		Field(6, "user_id", String, "User who triggered the event", func() {
			Format(FormatUUID)
			Example("user-456e7890-e89b-12d3-a456-426614174000")
			Description("Identity of the acting user (if applicable)")
		})

		Field(7, "entity_id", String, "Entity context for the event", func() {
			Format(FormatUUID)
			Example("entity-789abc12-def3-4567-890a-bcdef1234567")
			Description("Organizational entity where event occurred")
		})

		Field(8, "resource_type", String, "Type of resource being accessed", func() {
			Pattern("^[a-z][a-z0-9_]{2,49}$")
			Example("financial_account")
			Description("Category of resource affected by the event")
		})

		Field(9, "resource_id", String, "Specific resource identifier", func() {
			Format(FormatUUID)
			Example("resource-abc123de-f456-7890-abc1-23def4567890")
			Description("Unique identifier of the affected resource")
		})

		Field(10, "action_type", String, "Type of action performed", func() {
			Enum("CREATE", "READ", "UPDATE", "DELETE", "EXECUTE", "APPROVE", "EXPORT",
				"IMPORT", "LOGIN", "LOGOUT", "ACCESS_GRANTED", "ACCESS_DENIED")
			Example("READ")
			Description("Standard action classification")
		})

		Field(11, "decision", String, "Access control decision result", func() {
			Enum("ALLOW", "DENY", "ERROR", "NOT_APPLICABLE")
			Example("ALLOW")
			Description("Result of access control evaluation")
		})

		Field(12, "reason", String, "Detailed reason for the decision", func() {
			MaxLength(500)
			Example("User has required role: financial_analyst with permission: accounts.read")
			Description("Explanation of why access was granted or denied")
		})

		Field(13, "risk_score", UInt, "Calculated risk score (0-100)", func() {
			Maximum(100)
			Example(25)
			Description("Algorithmic risk assessment for the event")
		})

		Field(14, "session_id", String, "User session identifier", func() {
			Format(FormatUUID)
			Example("session-def456gh-i789-0123-def4-56gh78901234")
			Description("Session context for the event")
		})

		Field(15, "ip_address", String, "Source IP address", func() {
			Pattern("^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$")
			Example("192.168.1.100")
			Description("Client IP address for geolocation and security analysis")
		})

		Field(16, "user_agent", String, "Client user agent string", func() {
			MaxLength(500)
			Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
			Description("Browser/application identification")
		})

		Field(17, "geo_location", MapOf(String, Any), "Geographic location data", func() {
			Example(map[string]any{
				"country":   "United States",
				"region":    "California",
				"city":      "San Francisco",
				"latitude":  37.7749,
				"longitude": -122.4194,
				"timezone":  "America/Los_Angeles",
			})
			Description("Geolocation information for security analysis")
		})

		Field(18, "device_info", MapOf(String, Any), "Device and browser information", func() {
			Example(map[string]any{
				"device_type": "desktop",
				"os":          "Windows 10",
				"browser":     "Chrome 119.0",
				"screen_res":  "1920x1080",
				"fingerprint": "a1b2c3d4e5f6",
			})
			Description("Device characteristics for fraud detection")
		})

		Field(19, "request_id", String, "HTTP request correlation ID", func() {
			Format(FormatUUID)
			Example("req-123e4567-e89b-12d3-a456-426614174000")
			Description("Links audit event to specific API request")
		})

		Field(20, "api_endpoint", String, "API endpoint accessed", func() {
			Pattern("^(GET|POST|PUT|PATCH|DELETE)\\s+/[a-z0-9/\\-_{}]*$")
			Example("GET /api/v1/finance/accounts/123")
			Description("HTTP method and URL path")
		})

		Field(21, "response_status", UInt, "HTTP response status code", func() {
			Minimum(100)
			Maximum(599)
			Example(200)
			Description("HTTP status code returned")
		})

		Field(22, "response_time_ms", UInt, "Request processing time", func() {
			Example(145)
			Description("Time taken to process request in milliseconds")
		})

		Field(23, "data_before", MapOf(String, Any), "Data state before change", func() {
			Example(map[string]any{
				"status": "PENDING",
				"amount": "1000.00",
			})
			Description("Previous state for data modification events")
		})

		Field(24, "data_after", MapOf(String, Any), "Data state after change", func() {
			Example(map[string]any{
				"status": "APPROVED",
				"amount": "1000.00",
			})
			Description("New state after data modification")
		})

		Field(25, "context", MapOf(String, Any), "Additional event context", func() {
			Example(map[string]any{
				"policy_evaluated":   "high_value_transaction",
				"approval_required":  true,
				"notification_sent":  true,
				"transaction_amount": 5000.00,
			})
			Description("Flexible context data for event analysis")
		})

		Field(26, "tags", ArrayOf(String), "Event classification tags", func() {
			Example([]string{"financial", "high_value", "requires_review"})
			Description("Searchable tags for event categorization")
		})

		Field(27, "compliance_flags", ArrayOf(String), "Compliance framework markers", func() {
			Example([]string{"SOX", "PCI_DSS", "GDPR"})
			Description("Regulatory compliance requirements")
		})

		Field(28, "anomaly_indicators", MapOf(String, Any), "Anomaly detection results", func() {
			Example(map[string]any{
				"unusual_location":   true,
				"off_hours_access":   false,
				"anomaly_score":      0.75,
				"baseline_deviation": 2.3,
			})
			Description("Machine learning anomaly detection flags")
		})

		Field(29, "investigation_status", String, "Security investigation state", func() {
			Enum("NONE", "FLAGGED", "INVESTIGATING", "RESOLVED", "FALSE_POSITIVE")
			Default("NONE")
			Example("NONE")
			Description("Current state of security investigation")
		})

		Field(30, "retention_period_days", UInt, "Data retention requirement", func() {
			Minimum(30)
			Maximum(2555) // ~7 years
			Default(2555)
			Example(2555)
			Description("Number of days to retain this audit record")
		})

		Field(31, "occurred_at", String, "Event occurrence timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T10:30:00Z")
			Description("Precise time when event occurred")
		})

		// Standard audit fields
		AuditFields()

		Required("id", "tenant_id", "event_type", "event_category", "severity",
			"occurred_at", "created_at")
	})

	View("default", func() {
		Description("Standard audit view for general monitoring dashboards")
		Attribute("id")
		Attribute("event_type")
		Attribute("event_category")
		Attribute("severity")
		Attribute("user_id")
		Attribute("resource_type")
		Attribute("action_type")
		Attribute("decision")
		Attribute("ip_address")
		Attribute("occurred_at")
	})

	View("detailed", func() {
		Description("Complete audit view for security investigations")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("event_type")
		Attribute("event_category")
		Attribute("severity")
		Attribute("user_id")
		Attribute("entity_id")
		Attribute("resource_type")
		Attribute("resource_id")
		Attribute("action_type")
		Attribute("decision")
		Attribute("reason")
		Attribute("risk_score")
		Attribute("session_id")
		Attribute("ip_address")
		Attribute("user_agent")
		Attribute("geo_location")
		Attribute("device_info")
		Attribute("request_id")
		Attribute("api_endpoint")
		Attribute("response_status")
		Attribute("response_time_ms")
		Attribute("data_before")
		Attribute("data_after")
		Attribute("context")
		Attribute("tags")
		Attribute("compliance_flags")
		Attribute("anomaly_indicators")
		Attribute("investigation_status")
		Attribute("occurred_at")
		Attribute("created_at")
	})

	View("security", func() {
		Description("Security-focused view for threat analysis")
		Attribute("id")
		Attribute("event_type")
		Attribute("severity")
		Attribute("user_id")
		Attribute("decision")
		Attribute("risk_score")
		Attribute("ip_address")
		Attribute("geo_location")
		Attribute("anomaly_indicators")
		Attribute("investigation_status")
		Attribute("occurred_at")
	})

	View("compliance", func() {
		Description("Compliance reporting view")
		Attribute("id")
		Attribute("event_type")
		Attribute("event_category")
		Attribute("user_id")
		Attribute("resource_type")
		Attribute("action_type")
		Attribute("compliance_flags")
		Attribute("retention_period_days")
		Attribute("occurred_at")
	})
})

// SecurityEvent represents high-priority security events requiring immediate attention.
var SecurityEvent = Type("SecurityEvent", func() {
	Description("High-priority security event requiring immediate attention and response")

	Field(1, "id", String, "Unique security event identifier", func() {
		Format(FormatUUID)
		Example("sec-event-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for security event")
	})

	Field(2, "audit_log_id", String, "Associated audit log entry", func() {
		Format(FormatUUID)
		Example("audit-456e7890-e89b-12d3-a456-426614174000")
		Description("Reference to detailed audit log")
	})

	Field(3, "event_severity", String, "Security event severity", func() {
		Enum("INFORMATIONAL", "LOW", "MEDIUM", "HIGH", "CRITICAL")
		Example("HIGH")
		Description("Security impact level")
	})

	Field(4, "threat_type", String, "Type of security threat", func() {
		Enum("BRUTE_FORCE", "PRIVILEGE_ESCALATION", "DATA_EXFILTRATION",
			"UNAUTHORIZED_ACCESS", "SUSPICIOUS_ACTIVITY", "POLICY_VIOLATION")
		Example("UNAUTHORIZED_ACCESS")
		Description("Classification of security threat")
	})

	Field(5, "affected_resources", ArrayOf(String), "Resources impacted by event", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"account-123e4567-e89b-12d3-a456-426614174000",
			"user-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("List of resources potentially compromised")
	})

	Field(6, "indicators_of_compromise", ArrayOf(String), "IOCs identified", func() {
		Example([]string{
			"Multiple failed login attempts",
			"Access from blacklisted IP",
			"Unusual data access pattern",
		})
		Description("Security indicators that triggered the alert")
	})

	Field(7, "response_status", String, "Incident response status", func() {
		Enum("NEW", "ACKNOWLEDGED", "INVESTIGATING", "CONTAINED", "RESOLVED", "FALSE_POSITIVE")
		Default("NEW")
		Example("INVESTIGATING")
		Description("Current state of security response")
	})

	Field(8, "assigned_to", String, "Security analyst assigned", func() {
		Format(FormatUUID)
		Example("analyst-789abc12-def3-4567-890a-bcdef1234567")
		Description("Security team member handling the incident")
	})

	Field(9, "remediation_actions", ArrayOf(String), "Actions taken to address threat", func() {
		Example([]string{
			"Account temporarily suspended",
			"IP address blocked",
			"Additional monitoring enabled",
		})
		Description("Steps taken to mitigate the security event")
	})

	Field(10, "estimated_impact", String, "Potential business impact", func() {
		Enum("NONE", "MINIMAL", "MODERATE", "SIGNIFICANT", "SEVERE")
		Example("MODERATE")
		Description("Assessment of potential damage")
	})

	// Audit fields
	AuditFields()

	Required("id", "audit_log_id", "event_severity", "threat_type", "response_status", "created_at")
})

// AuditSearchQuery represents complex audit log search capabilities.
var AuditSearchQuery = Type("AuditSearchQuery", func() {
	Description("Advanced search parameters for audit log analysis and investigation")

	Field(1, "tenant_id", String, "Tenant scope filter", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Limit search to specific tenant")
	})

	Field(2, "event_types", ArrayOf(String), "Event types to include", func() {
		Example([]string{"USER_LOGIN_SUCCESS", "DATA_ACCESS", "PERMISSION_DENIED"})
		Description("Filter by specific event types")
	})

	Field(3, "event_categories", ArrayOf(String), "Event categories to include", func() {
		Example([]string{"AUTH", "DATA", "SECURITY"})
		Description("Filter by high-level categories")
	})

	Field(4, "severity_levels", ArrayOf(String), "Severity levels to include", func() {
		Example([]string{"HIGH", "CRITICAL"})
		Description("Filter by severity thresholds")
	})

	Field(5, "user_ids", ArrayOf(String), "Specific users to track", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"user-123e4567-e89b-12d3-a456-426614174000",
			"user-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("Focus on specific user activities")
	})

	Field(6, "time_range", TimeRange, "Time period for search", func() {
		Description("Limit results to specific time window")
	})

	Field(7, "ip_addresses", ArrayOf(String), "Source IP addresses", func() {
		Example([]string{"192.168.1.100", "10.0.0.50"})
		Description("Filter by source IP addresses")
	})

	Field(8, "risk_score_min", UInt, "Minimum risk score", func() {
		Maximum(100)
		Example(70)
		Description("Include events above risk threshold")
	})

	Field(9, "include_anomalies", Boolean, "Include anomaly-flagged events", func() {
		Default(false)
		Example(true)
		Description("Focus on events flagged by anomaly detection")
	})

	Field(10, "compliance_frameworks", ArrayOf(String), "Compliance-related events", func() {
		Example([]string{"SOX", "PCI_DSS", "GDPR"})
		Description("Filter by regulatory compliance requirements")
	})

	Field(11, "aggregation_level", String, "Result aggregation level", func() {
		Enum("RAW", "HOURLY", "DAILY", "WEEKLY", "MONTHLY")
		Default("RAW")
		Example("DAILY")
		Description("Level of detail for results")
	})

	Required("tenant_id")
})
