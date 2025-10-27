package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// POLICY ENGINE & ABAC TYPES
// ============================================================================

// Policy represents an ABAC (Attribute-Based Access Control) policy for dynamic authorization.
// Supports complex conditions, obligations, and hierarchical policy evaluation.
var Policy = ResultType("application/vnd.erp.policy", func() {
	Description("ABAC policy definition with dynamic conditions, obligations, and hierarchical evaluation for fine-grained access control")
	
	Attributes(func() {
		Field(1, "id", String, "Unique policy identifier", func() {
			Format(FormatUUID)
			Example("policy-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for policy references")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})
		
		Field(3, "entity_id", String, "Entity scope for policy", func() {
			Format(FormatUUID)
			Example("entity-456e7890-e89b-12d3-a456-426614174000")
			Description("Optional entity-specific policy scope")
		})
		
		Field(4, "name", String, "Unique policy name", func() {
			Pattern("^[a-z][a-z0-9_]{2,99}$")
			MinLength(3)
			MaxLength(100)
			Example("high_value_transaction_approval")
			Description("Machine-readable policy identifier")
		})
		
		Field(5, "display_name", String, "Human-readable policy title", func() {
			MinLength(3)
			MaxLength(150)
			Example("High Value Transaction Approval Policy")
			Description("User-friendly name for UI display")
		})
		
		Field(6, "description", String, "Detailed policy description", func() {
			MaxLength(1000)
			Example("Requires manager approval for financial transactions exceeding $10,000 or involving sensitive accounts")
			Description("Comprehensive explanation of policy purpose and scope")
		})
		
		Field(7, "policy_type", String, "Classification of policy behavior", func() {
			Enum("ABAC", "RBAC", "HYBRID", "TIME_BASED", "LOCATION_BASED", "CONTEXT_AWARE", "RISK_BASED")
			Default("ABAC")
			Example("ABAC")
			Description("Determines evaluation engine and available features")
		})
		
		Field(8, "effect", String, "Policy decision outcome", func() {
			Enum("ALLOW", "DENY")
			Example("ALLOW")
			Description("Access decision when policy conditions are met")
		})
		
		Field(9, "priority", UInt, "Policy evaluation priority", func() {
			Minimum(1)
			Maximum(1000)
			Default(100)
			Example(200)
			Description("Higher numbers evaluated first (conflict resolution)")
		})
		
		Field(10, "category", String, "Policy functional category", func() {
			Enum("ACCESS", "DATA_FILTER", "FIELD_MASK", "AUDIT", "COMPLIANCE", 
				"WORKFLOW", "BUSINESS_RULE", "SECURITY")
			Default("ACCESS")
			Example("ACCESS")
			Description("Determines when and how policy is applied")
		})
		
		Field(11, "scope", String, "Policy application scope", func() {
			Enum("GLOBAL", "TENANT", "ENTITY", "DEPARTMENT", "PROJECT", "USER_SPECIFIC")
			Default("ENTITY")
			Example("ENTITY")
			Description("Boundary within which policy is effective")
		})
		
		Field(12, "target", MapOf(String, Any), "Policy target definition", func() {
			Example(map[string]any{
				"resource_type": "financial_transaction",
				"actions":      []string{"create", "update", "approve"},
				"attributes": map[string]any{
					"amount_threshold": 10000.00,
					"account_types":   []string{"CASH", "INVESTMENT"},
				},
			})
			Description("Defines what resources and actions this policy governs")
		})
		
		Field(13, "subjects", MapOf(String, Any), "Subject constraints", func() {
			Example(map[string]any{
				"user_types":    []string{"EMPLOYEE", "MANAGER"},
				"roles":        []string{"financial_analyst", "accountant"},
				"departments":  []string{"finance", "accounting"},
				"security_clearance_min": 3,
			})
			Description("Defines who this policy applies to")
		})
		
		Field(14, "conditions", MapOf(String, Any), "Dynamic evaluation conditions", func() {
			Example(map[string]any{
				"amount_conditions": map[string]any{
					"transaction.amount": map[string]any{
						"greater_than": 10000.00,
						"currency":     "USD",
					},
				},
				"time_conditions": map[string]any{
					"business_hours_only": true,
					"timezone":           "America/New_York",
					"excluded_days":      []string{"saturday", "sunday"},
				},
				"location_conditions": map[string]any{
					"allowed_countries": []string{"US", "CA"},
					"blocked_ips":      []string{"192.168.100.0/24"},
				},
				"context_conditions": map[string]any{
					"require_mfa":           true,
					"max_concurrent_sessions": 1,
					"device_trust_required": true,
				},
			})
			Description("Complex conditions evaluated at runtime")
		})
		
		Field(15, "rule_logic", String, "Policy rule evaluation logic", func() {
			Example("(user.department == 'finance' OR user.role == 'manager') AND transaction.amount > 10000 AND time.business_hours == true")
			Description("Logical expression for policy evaluation")
		})
		
		Field(16, "obligations", MapOf(String, Any), "Required actions when policy applies", func() {
			Example(map[string]any{
				"approval_workflow": map[string]any{
					"required":        true,
					"approver_roles": []string{"department_manager", "finance_director"},
					"escalation_hours": 24,
				},
				"audit_requirements": map[string]any{
					"detailed_logging": true,
					"notify_security": true,
					"retention_years": 7,
				},
				"notifications": map[string]any{
					"immediate_alert": true,
					"recipients":     []string{"compliance@company.com"},
					"escalation_chain": []string{"manager", "director", "cfo"},
				},
			})
			Description("Mandatory actions triggered when policy is invoked")
		})
		
		Field(17, "advice", MapOf(String, Any), "Optional actions and recommendations", func() {
			Example(map[string]any{
				"warnings": []string{
					"High value transaction detected",
					"Additional documentation may be required",
				},
				"recommendations": []string{
					"Consider splitting transaction into smaller amounts",
					"Verify vendor information before processing",
				},
				"ui_modifications": map[string]any{
					"require_comments": true,
					"show_risk_warning": true,
					"additional_fields": []string{"business_justification"},
				},
			})
			Description("Non-mandatory guidance provided to users")
		})
		
		Field(18, "exceptions", MapOf(String, Any), "Policy exception handling", func() {
			Example(map[string]any{
				"emergency_override": map[string]any{
					"enabled":       true,
					"required_role": "system_admin",
					"audit_level":   "CRITICAL",
				},
				"temporary_bypass": map[string]any{
					"max_duration_hours": 4,
					"requires_approval": true,
					"auto_expire":      true,
				},
			})
			Description("Controlled mechanisms for policy override")
		})
		
		Field(19, "is_active", Boolean, "Policy activation status", func() {
			Default(true)
			Example(true)
			Description("Whether policy is currently enforced")
		})
		
		Field(20, "is_system_policy", Boolean, "System-managed policy flag", func() {
			Default(false)
			Example(false)
			Description("Cannot be modified if true")
		})
		
		Field(21, "version", String, "Policy version identifier", func() {
			Pattern("^\\d+\\.\\d+\\.\\d+$")
			Default("1.0.0")
			Example("2.1.0")
			Description("Semantic version for policy changes")
		})
		
		Field(22, "effective_from", String, "Policy effective start date", func() {
			Format(FormatDateTime)
			Example("2023-12-01T00:00:00Z")
			Description("When policy becomes active")
		})
		
		Field(23, "expires_at", String, "Policy expiration date", func() {
			Format(FormatDateTime)
			Example("2024-11-30T23:59:59Z")
			Description("When policy automatically deactivates")
		})
		
		Field(24, "parent_policy_id", String, "Parent policy for inheritance", func() {
			Format(FormatUUID)
			Example("parent-policy-789abc12-def3-4567-890a-bcdef1234567")
			Description("Inherits conditions from parent policy")
		})
		
		Field(25, "tags", ArrayOf(String), "Policy classification tags", func() {
			Example([]string{"financial", "high_risk", "sox_compliance"})
			Description("Searchable tags for policy organization")
		})
		
		Field(26, "compliance_frameworks", ArrayOf(String), "Regulatory compliance mappings", func() {
			Example([]string{"SOX", "PCI_DSS", "GDPR", "HIPAA"})
			Description("Compliance requirements this policy addresses")
		})
		
		Field(27, "evaluation_count", UInt, "Number of policy evaluations", func() {
			Example(15247)
			Description("Statistical usage tracking")
		})
		
		Field(28, "allow_count", UInt, "Number of allow decisions", func() {
			Example(14892)
			Description("Successful policy matches")
		})
		
		Field(29, "deny_count", UInt, "Number of deny decisions", func() {
			Example(355)
			Description("Policy violations blocked")
		})
		
		Field(30, "last_evaluated_at", String, "Most recent evaluation timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T14:30:00Z")
			Description("Latest policy usage")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "name", "display_name", "policy_type", "effect", 
			"priority", "category", "target", "conditions", "is_active", "version", 
			"effective_from", "created_at")
	})
	
	View("default", func() {
		Description("Standard policy view for listings and management")
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("policy_type")
		Attribute("effect")
		Attribute("priority")
		Attribute("category")
		Attribute("is_active")
		Attribute("version")
		Attribute("evaluation_count")
		Attribute("created_at")
	})
	
	View("detailed", func() {
		Description("Complete policy view with all configuration details")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("entity_id")
		Attribute("name")
		Attribute("display_name")
		Attribute("description")
		Attribute("policy_type")
		Attribute("effect")
		Attribute("priority")
		Attribute("category")
		Attribute("scope")
		Attribute("target")
		Attribute("subjects")
		Attribute("conditions")
		Attribute("rule_logic")
		Attribute("obligations")
		Attribute("advice")
		Attribute("exceptions")
		Attribute("is_active")
		Attribute("version")
		Attribute("effective_from")
		Attribute("expires_at")
		Attribute("parent_policy_id")
		Attribute("tags")
		Attribute("compliance_frameworks")
		Attribute("evaluation_count")
		Attribute("allow_count")
		Attribute("deny_count")
		Attribute("last_evaluated_at")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("evaluation", func() {
		Description("Optimized view for policy evaluation engines")
		Attribute("id")
		Attribute("name")
		Attribute("policy_type")
		Attribute("effect")
		Attribute("priority")
		Attribute("target")
		Attribute("subjects")
		Attribute("conditions")
		Attribute("rule_logic")
		Attribute("obligations")
		Attribute("is_active")
		Attribute("effective_from")
		Attribute("expires_at")
	})
	
	View("compliance", func() {
		Description("Compliance reporting view")
		Attribute("id")
		Attribute("name")
		Attribute("display_name")
		Attribute("category")
		Attribute("compliance_frameworks")
		Attribute("version")
		Attribute("effective_from")
		Attribute("expires_at")
		Attribute("evaluation_count")
		Attribute("created_at")
	})
})

// PolicyEvaluation represents the result of policy engine evaluation.
var PolicyEvaluation = Type("PolicyEvaluation", func() {
	Description("Result of policy engine evaluation with detailed decision context")
	
	Field(1, "id", String, "Unique evaluation identifier", func() {
		Format(FormatUUID)
		Example("eval-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for evaluation record")
	})
	
	Field(2, "tenant_id", String, "Tenant context", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})
	
	Field(3, "user_id", String, "User being evaluated", func() {
		Format(FormatUUID)
		Example("user-456e7890-e89b-12d3-a456-426614174000")
		Description("Subject of the access request")
	})
	
	Field(4, "resource_type", String, "Type of resource being accessed", func() {
		Pattern("^[a-z][a-z0-9_]{2,49}$")
		Example("financial_transaction")
		Description("Category of protected resource")
	})
	
	Field(5, "resource_id", String, "Specific resource identifier", func() {
		Format(FormatUUID)
		Example("txn-789abc12-def3-4567-890a-bcdef1234567")
		Description("Unique identifier of requested resource")
	})
	
	Field(6, "action", String, "Requested action", func() {
		Enum("CREATE", "READ", "UPDATE", "DELETE", "EXECUTE", "APPROVE", "EXPORT", "IMPORT")
		Example("CREATE")
		Description("Operation being attempted")
	})
	
	Field(7, "decision", String, "Final access decision", func() {
		Enum("ALLOW", "DENY", "NOT_APPLICABLE", "ERROR")
		Example("ALLOW")
		Description("Overall policy evaluation result")
	})
	
	Field(8, "confidence_score", Float64, "Decision confidence level", func() {
		Minimum(0.0)
		Maximum(1.0)
		Example(0.95)
		Description("Algorithm confidence in decision (0-1)")
	})
	
	Field(9, "applicable_policies", ArrayOf(String), "Policies evaluated", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"policy-123e4567-e89b-12d3-a456-426614174000",
			"policy-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("List of policies that applied to this request")
	})
	
	Field(10, "policy_results", ArrayOf(Type("PolicyResult", func() {
		Field(1, "policy_id", String, "Policy identifier", func() {
			Format(FormatUUID)
		})
		Field(2, "policy_name", String, "Policy name", func() {
			Example("high_value_transaction_approval")
		})
		Field(3, "decision", String, "Policy-specific decision", func() {
			Enum("ALLOW", "DENY", "NOT_APPLICABLE")
		})
		Field(4, "reason", String, "Decision explanation", func() {
			MaxLength(500)
			Example("Transaction amount $15,000 exceeds $10,000 threshold")
		})
		Field(5, "conditions_met", MapOf(String, Boolean), "Condition evaluation results", func() {
			Example(map[string]any{
				"amount_threshold":  true,
				"business_hours":   true,
				"user_has_role":    true,
			})
		})
		Required("policy_id", "policy_name", "decision")
	})), "Individual policy evaluation results", func() {
		Description("Detailed results for each policy evaluated")
	})
	
	Field(11, "obligations", ArrayOf(String), "Required obligations", func() {
		Example([]string{
			"REQUIRE_APPROVAL",
			"DETAILED_AUDIT_LOG",
			"NOTIFY_COMPLIANCE_TEAM",
		})
		Description("Actions that must be performed")
	})
	
	Field(12, "advice", ArrayOf(String), "Advisory recommendations", func() {
		Example([]string{
			"Consider additional documentation",
			"Verify vendor information",
			"Review transaction details carefully",
		})
		Description("Non-mandatory guidance for users")
	})
	
	Field(13, "context", MapOf(String, Any), "Evaluation context", func() {
		Example(map[string]any{
			"request_time":     "2023-12-07T14:30:00Z",
			"user_location":    "San Francisco, CA",
			"device_trusted":   true,
			"session_duration": "02:15:30",
			"risk_factors": map[string]any{
				"unusual_time":     false,
				"high_amount":      true,
				"new_vendor":       false,
			},
		})
		Description("Contextual information used in evaluation")
	})
	
	Field(14, "evaluation_time_ms", UInt, "Processing time", func() {
		Example(45)
		Description("Time taken to evaluate policies in milliseconds")
	})
	
	Field(15, "cache_hit", Boolean, "Result from cache", func() {
		Default(false)
		Example(false)
		Description("Whether result was served from cache")
	})
	
	Field(16, "evaluation_engine", String, "Policy engine version", func() {
		Pattern("^v\\d+\\.\\d+\\.\\d+$")
		Example("v2.1.0")
		Description("Version of policy evaluation engine")
	})
	
	Field(17, "evaluated_at", String, "Evaluation timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-07T14:30:00Z")
		Description("When evaluation was performed")
	})
	
	Required("tenant_id", "user_id", "resource_type", "action", "decision", 
		"applicable_policies", "policy_results", "evaluated_at")
})

// PolicySet represents a collection of related policies for grouped management.
var PolicySet = Type("PolicySet", func() {
	Description("Collection of related policies for grouped management and deployment")
	
	Field(1, "id", String, "Unique policy set identifier", func() {
		Format(FormatUUID)
		Example("policyset-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for policy set")
	})
	
	Field(2, "tenant_id", String, "Associated tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})
	
	Field(3, "name", String, "Policy set name", func() {
		Pattern("^[a-z][a-z0-9_]{2,99}$")
		MinLength(3)
		MaxLength(100)
		Example("financial_controls_q4_2023")
		Description("Unique identifier for policy set")
	})
	
	Field(4, "display_name", String, "Human-readable title", func() {
		MinLength(3)
		MaxLength(150)
		Example("Financial Controls Q4 2023")
		Description("User-friendly name for UI display")
	})
	
	Field(5, "description", String, "Policy set description", func() {
		MaxLength(1000)
		Example("Comprehensive financial controls for Q4 2023 compliance requirements")
		Description("Detailed explanation of policy set purpose")
	})
	
	Field(6, "category", String, "Policy set category", func() {
		Enum("COMPLIANCE", "SECURITY", "BUSINESS_RULES", "OPERATIONAL", "REGULATORY")
		Example("COMPLIANCE")
		Description("Functional classification")
	})
	
	Field(7, "policy_ids", ArrayOf(String), "Contained policies", func() {
		Elem(func() {
			Format(FormatUUID)
		})
		Example([]string{
			"policy-123e4567-e89b-12d3-a456-426614174000",
			"policy-456e7890-e89b-12d3-a456-426614174000",
		})
		Description("List of policies in this set")
	})
	
	Field(8, "deployment_status", String, "Deployment state", func() {
		Enum("DRAFT", "TESTING", "DEPLOYED", "DEPRECATED")
		Default("DRAFT")
		Example("DEPLOYED")
		Description("Current deployment status")
	})
	
	Field(9, "version", String, "Policy set version", func() {
		Pattern("^\\d+\\.\\d+\\.\\d+$")
		Default("1.0.0")
		Example("1.2.0")
		Description("Semantic version for change tracking")
	})
	
	Field(10, "effective_from", String, "Effective start date", func() {
		Format(FormatDateTime)
		Example("2023-12-01T00:00:00Z")
		Description("When policy set becomes active")
	})
	
	Field(11, "expires_at", String, "Expiration date", func() {
		Format(FormatDateTime)
		Example("2024-03-31T23:59:59Z")
		Description("When policy set automatically deactivates")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "tenant_id", "name", "display_name", "category", "policy_ids", 
		"deployment_status", "version", "effective_from", "created_at")
})