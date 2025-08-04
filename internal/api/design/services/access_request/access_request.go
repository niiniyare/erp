package design

import (
	. "goa.design/goa/v3/dsl"
)

// AccessRequestService defines the access request management service
var _ = Service("access_request", func() {
	Description("Access request management service for ABAC authorization")

	// Global error definitions
	Error("bad_request", ErpError, "Bad request")
	Error("not_found", ErpError, "Resource not found")
	Error("access_request_not_found", String, "Access request not found")
	Error("access_request_already_processed", String, "Access request already processed")
	Error("invalid_access_request_status", String, "Invalid access request status")
	Error("access_denied", String, "Access denied")
	Error("conditional_access_failed", String, "Conditional access evaluation failed")

	// Core Access Request operations
	Method("create", func() {
		Description("Create a new access request")
		
		Payload(CreateAccessRequestPayload)
		Result(AccessRequestResult)
		
		HTTP(func() {
			POST("/api/v1/access-requests")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
			Response("access_denied", StatusForbidden)
		})
	})

	Method("get", func() {
		Description("Get access request by ID")
		
		Payload(func() {
			Attribute("id", String, "Access request ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("id")
		})
		Result(AccessRequestResult)
		
		HTTP(func() {
			GET("/api/v1/access-requests/{id}")
			Response(StatusOK)
			Response("access_request_not_found", StatusNotFound)
		})
	})

	Method("process", func() {
		Description("Process an access request (approve/deny)")
		
		Payload(ProcessAccessRequestPayload)
		Result(AccessRequestResult)
		
		HTTP(func() {
			POST("/api/v1/access-requests/{id}/process")
			Response(StatusOK)
			Response("access_request_not_found", StatusNotFound)
			Response("access_request_already_processed", StatusConflict)
			Response("invalid_access_request_status", StatusBadRequest)
		})
	})

	Method("revoke", func() {
		Description("Revoke an access request")
		
		Payload(func() {
			Attribute("id", String, "Access request ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("id")
		})
		
		HTTP(func() {
			DELETE("/api/v1/access-requests/{id}")
			Response(StatusNoContent)
			Response("access_request_not_found", StatusNotFound)
		})
	})

	Method("list", func() {
		Description("List access requests with filtering")
		
		Payload(ListAccessRequestsPayload)
		Result(AccessRequestListResult)
		
		HTTP(func() {
			GET("/api/v1/access-requests")
			Param("status:status")
			Param("requester_id:requester_id")
			Param("entity_id:entity_id")
			Param("limit:limit")
			Param("offset:offset")
			Response(StatusOK)
		})
	})

	Method("stats", func() {
		Description("Get access request statistics")
		
		Result(AccessRequestStatsResult)
		
		HTTP(func() {
			GET("/api/v1/access-requests/stats")
			Response(StatusOK)
		})
	})

	// Conditional Access operations
	Method("evaluate_conditional_access", func() {
		Description("Evaluate conditional access rules")
		
		Payload(ConditionalAccessPayload)
		Result(ConditionalAccessResult)
		
		HTTP(func() {
			POST("/api/v1/conditional-access/evaluate")
			Response(StatusOK)
			Response("conditional_access_failed", StatusForbidden)
		})
	})

	Method("create_conditional_rule", func() {
		Description("Create a conditional access rule")
		
		Payload(CreateConditionalRulePayload)
		Result(ConditionalRuleResult)
		
		HTTP(func() {
			POST("/api/v1/conditional-access/rules")
			Response(StatusCreated)
			Response("bad_request", StatusBadRequest)
		})
	})

	// Analytics operations
	Method("user_behavior_analytics", func() {
		Description("Get user behavior analytics")
		
		Payload(func() {
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("user_id")
		})
		Result(UserBehaviorResult)
		
		HTTP(func() {
			GET("/api/v1/analytics/users/{user_id}/behavior")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("user_risk_assessment", func() {
		Description("Get user risk assessment")
		
		Payload(func() {
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("user_id")
		})
		Result(UserRiskResult)
		
		HTTP(func() {
			GET("/api/v1/analytics/users/{user_id}/risk")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("user_insights", func() {
		Description("Get personalized user insights")
		
		Payload(func() {
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("user_id")
		})
		Result(UserInsightsResult)
		
		HTTP(func() {
			GET("/api/v1/analytics/users/{user_id}/insights")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("detect_anomalies", func() {
		Description("Detect user behavior anomalies")
		
		Payload(func() {
			Attribute("user_id", String, "User ID", func() {
				Format(FormatUUID)
				Example("123e4567-e89b-12d3-a456-426614174000")
			})
			Required("user_id")
		})
		Result(AnomalyDetectionResult)
		
		HTTP(func() {
			POST("/api/v1/analytics/users/{user_id}/detect-anomalies")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
})

// Payload types
var CreateAccessRequestPayload = Type("CreateAccessRequestPayload", func() {
	Description("Payload for creating access request")
	
	Attribute("requester_id", String, "ID of the user requesting access", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "ID of the entity being requested", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("resource_type", String, "Type of resource", func() {
		Enum("document", "system", "application", "data", "facility")
		Example("document")
	})
	Attribute("access_level", String, "Requested access level", func() {
		Enum("read", "write", "admin", "full")
		Example("read")
	})
	Attribute("reason", String, "Reason for access request", func() {
		MaxLength(1000)
		Example("Need access to review quarterly reports")
	})
	Attribute("duration_hours", Int, "Requested access duration in hours", func() {
		Minimum(1)
		Maximum(8760) // 1 year
		Example(24)
	})
	Attribute("metadata", MapOf(String, Any), "Additional request metadata")
	
	Required("requester_id", "entity_id", "resource_type", "access_level", "reason")
})

var ProcessAccessRequestPayload = Type("ProcessAccessRequestPayload", func() {
	Description("Payload for processing access request")
	
	Attribute("id", String, "Access request ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("action", String, "Action to take", func() {
		Enum("approve", "deny", "pending")
		Example("approve")
	})
	Attribute("reviewer_id", String, "ID of the reviewer", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("comments", String, "Review comments", func() {
		MaxLength(1000)
		Example("Approved for project work")
	})
	
	Required("id", "action", "reviewer_id")
})

var ListAccessRequestsPayload = Type("ListAccessRequestsPayload", func() {
	Description("Payload for listing access requests")
	
	Attribute("status", String, "Filter by status", func() {
		Enum("pending", "approved", "denied", "expired", "revoked")
		Example("pending")
	})
	Attribute("requester_id", String, "Filter by requester ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Filter by entity ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("limit", Int, "Number of requests to return", func() {
		Default(50)
		Minimum(1)
		Maximum(1000)
		Example(50)
	})
	Attribute("offset", Int, "Number of requests to skip", func() {
		Default(0)
		Minimum(0)
		Example(0)
	})
})

var ConditionalAccessPayload = Type("ConditionalAccessPayload", func() {
	Description("Payload for conditional access evaluation")
	
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("resource_id", String, "Resource ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("action", String, "Requested action", func() {
		Enum("read", "write", "delete", "execute", "admin")
		Example("read")
	})
	Attribute("context", ConditionalAccessContext, "Access context")
	
	Required("user_id", "resource_id", "action", "context")
})

var CreateConditionalRulePayload = Type("CreateConditionalRulePayload", func() {
	Description("Payload for creating conditional access rule")
	
	Attribute("name", String, "Rule name", func() {
		MinLength(1)
		MaxLength(255)
		Example("Office Hours Access")
	})
	Attribute("description", String, "Rule description", func() {
		MaxLength(1000)
		Example("Allow access only during office hours")
	})
	Attribute("conditions", ArrayOf(String), "Rule conditions", func() {
		Example([]string{"time_range:09:00-17:00", "location:office"})
	})
	Attribute("actions", ArrayOf(String), "Actions to take", func() {
		Example([]string{"allow", "require_mfa"})
	})
	Attribute("priority", Int, "Rule priority (higher = more important)", func() {
		Default(100)
		Minimum(1)
		Maximum(1000)
		Example(100)
	})
	Attribute("enabled", Boolean, "Whether rule is enabled", func() {
		Default(true)
		Example(true)
	})
	
	Required("name", "conditions", "actions")
})

// Result types
var AccessRequestResult = Type("AccessRequestResult", func() {
	Description("Access request information")
	
	Attribute("id", String, "Access request ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("requester_id", String, "Requester user ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("entity_id", String, "Entity ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("resource_type", String, "Resource type", func() {
		Example("document")
	})
	Attribute("access_level", String, "Access level", func() {
		Example("read")
	})
	Attribute("status", String, "Request status", func() {
		Example("pending")
	})
	Attribute("reason", String, "Request reason", func() {
		Example("Need access to review quarterly reports")
	})
	Attribute("duration_hours", Int, "Duration in hours", func() {
		Example(24)
	})
	Attribute("reviewer_id", String, "Reviewer ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("reviewed_at", String, "Review timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("comments", String, "Review comments", func() {
		Example("Approved for project work")
	})
	Attribute("expires_at", String, "Expiration timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-08T10:30:00Z")
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	
	Required("id", "requester_id", "entity_id", "resource_type", "access_level", "status", "reason", "duration_hours", "created_at", "updated_at")
})

var AccessRequestListResult = Type("AccessRequestListResult", func() {
	Description("List of access requests")
	
	Attribute("requests", ArrayOf(AccessRequestResult), "List of access requests")
	Attribute("total", Int, "Total number of requests", func() {
		Example(150)
	})
	Attribute("limit", Int, "Requested limit", func() {
		Example(50)
	})
	Attribute("offset", Int, "Requested offset", func() {
		Example(0)
	})
	
	Required("requests", "total", "limit", "offset")
})

var AccessRequestStatsResult = Type("AccessRequestStatsResult", func() {
	Description("Access request statistics")
	
	Attribute("total_requests", Int, "Total number of requests", func() {
		Example(1250)
	})
	Attribute("pending_requests", Int, "Number of pending requests", func() {
		Example(45)
	})
	Attribute("approved_requests", Int, "Number of approved requests", func() {
		Example(980)
	})
	Attribute("denied_requests", Int, "Number of denied requests", func() {
		Example(200)
	})
	Attribute("expired_requests", Int, "Number of expired requests", func() {
		Example(25)
	})
	Attribute("average_approval_time_hours", Float64, "Average approval time in hours", func() {
		Example(4.5)
	})
	Attribute("approval_rate", Float64, "Approval rate (0-1)", func() {
		Example(0.82)
	})
	
	Required("total_requests", "pending_requests", "approved_requests", "denied_requests", "expired_requests", "average_approval_time_hours", "approval_rate")
})

var ConditionalAccessContext = Type("ConditionalAccessContext", func() {
	Description("Context for conditional access evaluation")
	
	Attribute("ip_address", String, "Client IP address", func() {
		Example("192.168.1.100")
	})
	Attribute("user_agent", String, "Client user agent", func() {
		Example("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	})
	Attribute("location", String, "User location", func() {
		Example("office")
	})
	Attribute("device_id", String, "Device identifier", func() {
		Example("device-123")
	})
	Attribute("time_of_day", String, "Time of access", func() {
		Example("14:30:00")
	})
	Attribute("day_of_week", String, "Day of the week", func() {
		Enum("monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday")
		Example("monday")
	})
	Attribute("risk_score", Float64, "Calculated risk score", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.25)
	})
	
	Required("ip_address", "time_of_day", "day_of_week")
})

var ConditionalAccessResult = Type("ConditionalAccessResult", func() {
	Description("Result of conditional access evaluation")
	
	Attribute("decision", String, "Access decision", func() {
		Enum("allow", "deny", "challenge")
		Example("allow")
	})
	Attribute("reason", String, "Reason for decision", func() {
		Example("User is accessing from trusted location during office hours")
	})
	Attribute("required_actions", ArrayOf(String), "Actions required before access", func() {
		Example([]string{"mfa_verification", "device_registration"})
	})
	Attribute("confidence_score", Float64, "Confidence in decision", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.95)
	})
	Attribute("valid_until", String, "When decision expires", func() {
		Format(FormatDateTime)
		Example("2023-12-07T18:00:00Z")
	})
	
	Required("decision", "reason", "confidence_score")
})

var ConditionalRuleResult = Type("ConditionalRuleResult", func() {
	Description("Conditional access rule")
	
	Attribute("id", String, "Rule ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("name", String, "Rule name", func() {
		Example("Office Hours Access")
	})
	Attribute("description", String, "Rule description", func() {
		Example("Allow access only during office hours")
	})
	Attribute("conditions", ArrayOf(String), "Rule conditions", func() {
		Example([]string{"time_range:09:00-17:00", "location:office"})
	})
	Attribute("actions", ArrayOf(String), "Actions to take", func() {
		Example([]string{"allow", "require_mfa"})
	})
	Attribute("priority", Int, "Rule priority", func() {
		Example(100)
	})
	Attribute("enabled", Boolean, "Whether rule is enabled", func() {
		Example(true)
	})
	Attribute("created_at", String, "Creation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	
	Required("id", "name", "conditions", "actions", "priority", "enabled", "created_at", "updated_at")
})

// Analytics result types
var UserBehaviorResult = Type("UserBehaviorResult", func() {
	Description("User behavior analytics")
	
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("login_frequency", Int, "Logins per day average", func() {
		Example(5)
	})
	Attribute("peak_hours", ArrayOf(Int), "Peak activity hours", func() {
		Example([]int{9, 10, 14, 15})
	})
	Attribute("common_locations", ArrayOf(String), "Common access locations", func() {
		Example([]string{"office", "home"})
	})
	Attribute("device_usage", MapOf(String, Int), "Device usage patterns", func() {
		Example(map[string]interface{}{"desktop": 80, "mobile": 20})
	})
	Attribute("access_patterns", ArrayOf(String), "Common access patterns", func() {
		Example([]string{"morning_reports", "afternoon_updates"})
	})
	
	Required("user_id", "login_frequency", "peak_hours", "common_locations")
})

var UserRiskResult = Type("UserRiskResult", func() {
	Description("User risk assessment")
	
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("risk_score", Float64, "Overall risk score", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.25)
	})
	Attribute("risk_level", String, "Risk level category", func() {
		Enum("low", "medium", "high", "critical")
		Example("low")
	})
	Attribute("risk_factors", ArrayOf(String), "Contributing risk factors", func() {
		Example([]string{"unusual_location", "off_hours_access"})
	})
	Attribute("recommendations", ArrayOf(String), "Security recommendations", func() {
		Example([]string{"enable_mfa", "review_permissions"})
	})
	Attribute("last_assessment", String, "Last assessment timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	
	Required("user_id", "risk_score", "risk_level", "last_assessment")
})

var UserInsightsResult = Type("UserInsightsResult", func() {
	Description("Personalized user insights")
	
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("productivity_score", Float64, "Productivity score", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.85)
	})
	Attribute("usage_trends", ArrayOf(String), "Usage trend insights", func() {
		Example([]string{"increased_weekend_usage", "consistent_morning_pattern"})
	})
	Attribute("optimization_suggestions", ArrayOf(String), "Suggestions for optimization", func() {
		Example([]string{"consolidate_morning_tasks", "use_mobile_app_more"})
	})
	Attribute("security_alerts", ArrayOf(String), "Security-related alerts", func() {
		Example([]string{"new_device_detected", "unusual_data_access"})
	})
	
	Required("user_id", "productivity_score", "usage_trends")
})

var AnomalyDetectionResult = Type("AnomalyDetectionResult", func() {
	Description("Anomaly detection results")
	
	Attribute("user_id", String, "User ID", func() {
		Format(FormatUUID)
		Example("123e4567-e89b-12d3-a456-426614174000")
	})
	Attribute("anomalies_detected", Int, "Number of anomalies detected", func() {
		Example(3)
	})
	Attribute("anomalies", ArrayOf(AnomalyResult), "List of detected anomalies")
	Attribute("overall_risk", String, "Overall risk assessment", func() {
		Enum("low", "medium", "high", "critical")
		Example("medium")
	})
	Attribute("detection_timestamp", String, "When detection was performed", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	
	Required("user_id", "anomalies_detected", "anomalies", "overall_risk", "detection_timestamp")
})

var AnomalyResult = Type("AnomalyResult", func() {
	Description("Individual anomaly detection result")
	
	Attribute("type", String, "Type of anomaly", func() {
		Enum("unusual_location", "off_hours_access", "unusual_data_volume", "new_device", "permission_escalation")
		Example("unusual_location")
	})
	Attribute("severity", String, "Anomaly severity", func() {
		Enum("low", "medium", "high", "critical")
		Example("medium")
	})
	Attribute("description", String, "Description of the anomaly", func() {
		Example("User accessed system from new geographic location")
	})
	Attribute("confidence", Float64, "Confidence in anomaly detection", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.85)
	})
	Attribute("timestamp", String, "When anomaly occurred", func() {
		Format(FormatDateTime)
		Example("2023-12-07T10:30:00Z")
	})
	Attribute("metadata", MapOf(String, Any), "Additional anomaly metadata")
	
	Required("type", "severity", "description", "confidence", "timestamp")
})

// Common error type
var ErpError = Type("ErpError", func() {
	Description("Common error response")
	
	Attribute("message", String, "Error message", func() {
		Example("Invalid request format")
	})
	Attribute("code", String, "Error code", func() {
		Example("INVALID_REQUEST")
	})
	Attribute("details", String, "Additional error details", func() {
		Example("Field 'user_id' is required")
	})
	
	Required("message", "code")
})