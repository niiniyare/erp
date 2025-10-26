package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// ACCESS REQUEST & APPROVAL WORKFLOW TYPES
// ============================================================================

// AccessRequest represents a request for elevated access or permissions.
// Supports workflow-based approval processes with detailed justification and audit trails.
var AccessRequest = ResultType("application/vnd.erp.access_request", func() {
	Description("Access request for elevated permissions with workflow-based approval process and comprehensive audit trail")
	
	Attributes(func() {
		Field(1, "id", String, "Unique access request identifier", func() {
			Format(FormatUUID)
			Example("access-req-123e4567-e89b-12d3-a456-426614174000")
			Description("Primary key for access request")
		})
		
		Field(2, "tenant_id", String, "Associated tenant identifier", func() {
			Format(FormatUUID)
			Example("550e8400-e29b-41d4-a716-446655440000")
			Description("Tenant isolation boundary")
		})
		
		Field(3, "requester_id", String, "User requesting access", func() {
			Format(FormatUUID)
			Example("user-456e7890-e89b-12d3-a456-426614174000")
			Description("User who initiated the access request")
		})
		
		Field(4, "target_user_id", String, "User receiving the access", func() {
			Format(FormatUUID)
			Example("user-789abc12-def3-4567-890a-bcdef1234567")
			Description("User who will receive the requested access (may be same as requester)")
		})
		
		Field(5, "entity_id", String, "Entity context for request", func() {
			Format(FormatUUID)
			Example("entity-def456gh-i789-0123-def4-56gh78901234")
			Description("Organizational entity where access is needed")
		})
		
		Field(6, "request_type", String, "Type of access being requested", func() {
			Enum("ROLE_ASSIGNMENT", "PERMISSION_GRANT", "RESOURCE_ACCESS", "ELEVATION", 
				"TEMPORARY_ACCESS", "EMERGENCY_ACCESS", "DELEGATION", "PROJECT_ACCESS")
			Example("ROLE_ASSIGNMENT")
			Description("Category of access request")
		})
		
		Field(7, "urgency_level", String, "Request urgency classification", func() {
			Enum("LOW", "NORMAL", "HIGH", "URGENT", "EMERGENCY")
			Default("NORMAL")
			Example("HIGH")
			Description("Priority level affecting approval SLA")
		})
		
		Field(8, "role_id", String, "Role being requested", func() {
			Format(FormatUUID)
			Example("role-123e4567-e89b-12d3-a456-426614174000")
			Description("Specific role for role assignment requests")
		})
		
		Field(9, "permission_ids", ArrayOf(String), "Specific permissions requested", func() {
			Elem(func() {
				Format(FormatUUID)
			})
			Example([]string{
				"perm-456e7890-e89b-12d3-a456-426614174000",
				"perm-789abc12-def3-4567-890a-bcdef1234567",
			})
			Description("List of individual permissions for permission grant requests")
		})
		
		Field(10, "resource_ids", ArrayOf(String), "Resources requiring access", func() {
			Elem(func() {
				Format(FormatUUID)
			})
			Example([]string{
				"account-123e4567-e89b-12d3-a456-426614174000",
				"project-456e7890-e89b-12d3-a456-426614174000",
			})
			Description("Specific resources for resource access requests")
		})
		
		Field(11, "justification", String, "Business justification for request", func() {
			MinLength(10)
			MaxLength(2000)
			Example("Need temporary financial analyst access to complete year-end audit procedures. Current project requires review of account balances and transaction details.")
			Description("Detailed explanation of why access is needed")
		})
		
		Field(12, "business_reason", String, "High-level business case", func() {
			MaxLength(500)
			Example("Year-end financial audit compliance requirement")
			Description("Brief business context for the request")
		})
		
		Field(13, "duration_hours", UInt, "Requested access duration", func() {
			Minimum(1)
			Maximum(8760) // 1 year
			Example(72)
			Description("How long access is needed (in hours)")
		})
		
		Field(14, "is_temporary", Boolean, "Temporary access flag", func() {
			Default(true)
			Example(true)
			Description("Whether access should auto-expire")
		})
		
		Field(15, "start_date", String, "Requested access start time", func() {
			Format(FormatDateTime)
			Example("2023-12-07T09:00:00Z")
			Description("When access should become effective")
		})
		
		Field(16, "end_date", String, "Requested access end time", func() {
			Format(FormatDateTime)
			Example("2023-12-10T17:00:00Z")
			Description("When access should automatically expire")
		})
		
		Field(17, "approval_status", String, "Current approval workflow status", func() {
			Enum("PENDING", "UNDER_REVIEW", "APPROVED", "REJECTED", "EXPIRED", 
				"REVOKED", "CANCELLED", "ESCALATED")
			Default("PENDING")
			Example("APPROVED")
			Description("Current state in approval workflow")
		})
		
		Field(18, "workflow_id", String, "Associated approval workflow", func() {
			Format(FormatUUID)
			Example("workflow-abc123de-f456-7890-abc1-23def4567890")
			Description("Workflow template governing approval process")
		})
		
		Field(19, "approval_steps", ArrayOf(Type("ApprovalStep", func() {
			Field(1, "step_number", UInt, "Step sequence", func() {
				Minimum(1)
				Example(1)
			})
			Field(2, "approver_role", String, "Required approver role", func() {
				Example("department_manager")
			})
			Field(3, "approver_id", String, "Assigned approver", func() {
				Format(FormatUUID)
				Example("user-manager-123e4567-e89b-12d3-a456-426614174000")
			})
			Field(4, "status", String, "Step status", func() {
				Enum("PENDING", "APPROVED", "REJECTED", "SKIPPED", "ESCALATED")
				Example("APPROVED")
			})
			Field(5, "decision_date", String, "Decision timestamp", func() {
				Format(FormatDateTime)
				Example("2023-12-07T14:30:00Z")
			})
			Field(6, "comments", String, "Approver comments", func() {
				MaxLength(1000)
				Example("Approved for year-end audit procedures. Monitor access usage.")
			})
			Field(7, "due_date", String, "Decision due date", func() {
				Format(FormatDateTime)
				Example("2023-12-08T17:00:00Z")
			})
			Required("step_number", "approver_role", "status")
		})), "Approval workflow steps", func() {
			Description("Detailed approval workflow progress")
		})
		
		Field(20, "current_approver_id", String, "Current pending approver", func() {
			Format(FormatUUID)
			Example("approver-def456gh-i789-0123-def4-56gh78901234")
			Description("User who needs to make the next approval decision")
		})
		
		Field(21, "approved_by", String, "Final approver user ID", func() {
			Format(FormatUUID)
			Example("manager-123e4567-e89b-12d3-a456-426614174000")
			Description("User who gave final approval")
		})
		
		Field(22, "approved_at", String, "Final approval timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T15:30:00Z")
			Description("When request was finally approved")
		})
		
		Field(23, "rejection_reason", String, "Reason for rejection", func() {
			MaxLength(1000)
			Example("Requested permissions exceed user's role requirements. Consider requesting specific read-only access instead.")
			Description("Detailed explanation of why request was denied")
		})
		
		Field(24, "rejection_by", String, "User who rejected request", func() {
			Format(FormatUUID)
			Example("security-admin-456e7890-e89b-12d3-a456-426614174000")
			Description("User who denied the request")
		})
		
		Field(25, "rejected_at", String, "Rejection timestamp", func() {
			Format(FormatDateTime)
			Example("2023-12-07T16:00:00Z")
			Description("When request was rejected")
		})
		
		Field(26, "escalation_level", UInt, "Current escalation level", func() {
			Maximum(5)
			Default(0)
			Example(1)
			Description("Number of times request has been escalated")
		})
		
		Field(27, "escalated_to", String, "Escalation target user", func() {
			Format(FormatUUID)
			Example("director-789abc12-def3-4567-890a-bcdef1234567")
			Description("Higher authority for escalated approval")
		})
		
		Field(28, "sla_deadline", String, "Service level agreement deadline", func() {
			Format(FormatDateTime)
			Example("2023-12-09T17:00:00Z")
			Description("Target resolution time based on urgency")
		})
		
		Field(29, "auto_approve_eligible", Boolean, "Eligible for automatic approval", func() {
			Default(false)
			Example(false)
			Description("Whether request meets auto-approval criteria")
		})
		
		Field(30, "risk_assessment", MapOf(String, Any), "Security risk evaluation", func() {
			Example(map[string]any{
				"risk_score":        75,
				"risk_level":       "MEDIUM",
				"risk_factors": []string{
					"High privilege level",
					"Cross-department access",
				},
				"mitigation_required": true,
			})
			Description("Automated and manual risk assessment results")
		})
		
		Field(31, "compliance_review", MapOf(String, Any), "Compliance framework review", func() {
			Example(map[string]any{
				"frameworks_applicable": []string{"SOX", "PCI_DSS"},
				"additional_controls_required": true,
				"compliance_officer_review": true,
				"documentation_requirements": []string{
					"Business justification form",
					"Manager sign-off",
				},
			})
			Description("Regulatory compliance considerations")
		})
		
		Field(32, "conditions", ArrayOf(String), "Approval conditions", func() {
			Example([]string{
				"Access limited to business hours only",
				"Must complete additional security training",
				"Supervisor notification required for each use",
			})
			Description("Conditions attached to approved access")
		})
		
		Field(33, "monitoring_requirements", MapOf(String, Any), "Enhanced monitoring settings", func() {
			Example(map[string]any{
				"detailed_audit_log":    true,
				"real_time_alerts":     true,
				"usage_reporting":      "DAILY",
				"anomaly_detection":    true,
				"supervisor_notifications": true,
			})
			Description("Additional monitoring for granted access")
		})
		
		Field(34, "access_granted_at", String, "When access was actually granted", func() {
			Format(FormatDateTime)
			Example("2023-12-07T16:30:00Z")
			Description("Timestamp when permissions were activated")
		})
		
		Field(35, "access_revoked_at", String, "When access was revoked", func() {
			Format(FormatDateTime)
			Example("2023-12-10T17:00:00Z")
			Description("Timestamp when permissions were removed")
		})
		
		Field(36, "usage_statistics", MapOf(String, Any), "Access usage tracking", func() {
			Example(map[string]any{
				"total_logins":        15,
				"total_actions":       47,
				"last_activity":      "2023-12-10T15:30:00Z",
				"resources_accessed": 8,
				"unusual_activity":   false,
			})
			Description("Statistics on how granted access was used")
		})
		
		Field(37, "attachments", ArrayOf(String), "Supporting documentation", func() {
			Elem(func() {
				Format(FormatUUID)
			})
			Example([]string{
				"doc-business-case-123e4567-e89b-12d3-a456-426614174000",
				"doc-manager-approval-456e7890-e89b-12d3-a456-426614174000",
			})
			Description("Supporting documents for the request")
		})
		
		Field(38, "tags", ArrayOf(String), "Request classification tags", func() {
			Example([]string{"audit", "financial", "temporary", "high_priority"})
			Description("Searchable tags for request categorization")
		})
		
		Field(39, "external_ticket_id", String, "External ticketing system reference", func() {
			MaxLength(100)
			Example("TICKET-2023-ACCESS-5678")
			Description("Reference to external service desk ticket")
		})
		
		// Audit fields from common.go
		AuditFields()
		
		Required("id", "tenant_id", "requester_id", "target_user_id", "entity_id", 
			"request_type", "justification", "duration_hours", "approval_status", 
			"created_at")
	})
	
	View("default", func() {
		Description("Standard access request view for listings and dashboards")
		Attribute("id")
		Attribute("requester_id")
		Attribute("target_user_id")
		Attribute("request_type")
		Attribute("urgency_level")
		Attribute("approval_status")
		Attribute("business_reason")
		Attribute("duration_hours")
		Attribute("sla_deadline")
		Attribute("created_at")
	})
	
	View("detailed", func() {
		Description("Complete access request view with full workflow details")
		Attribute("id")
		Attribute("tenant_id")
		Attribute("requester_id")
		Attribute("target_user_id")
		Attribute("entity_id")
		Attribute("request_type")
		Attribute("urgency_level")
		Attribute("role_id")
		Attribute("permission_ids")
		Attribute("resource_ids")
		Attribute("justification")
		Attribute("business_reason")
		Attribute("duration_hours")
		Attribute("is_temporary")
		Attribute("start_date")
		Attribute("end_date")
		Attribute("approval_status")
		Attribute("workflow_id")
		Attribute("approval_steps")
		Attribute("current_approver_id")
		Attribute("approved_by")
		Attribute("approved_at")
		Attribute("rejection_reason")
		Attribute("escalation_level")
		Attribute("sla_deadline")
		Attribute("risk_assessment")
		Attribute("compliance_review")
		Attribute("conditions")
		Attribute("monitoring_requirements")
		Attribute("access_granted_at")
		Attribute("access_revoked_at")
		Attribute("usage_statistics")
		Attribute("attachments")
		Attribute("tags")
		Attribute("external_ticket_id")
		Attribute("created_at")
		Attribute("updated_at")
		Attribute("created_by")
		Attribute("updated_by")
	})
	
	View("approval", func() {
		Description("Approval workflow view for managers and reviewers")
		Attribute("id")
		Attribute("requester_id")
		Attribute("target_user_id")
		Attribute("request_type")
		Attribute("urgency_level")
		Attribute("justification")
		Attribute("business_reason")
		Attribute("duration_hours")
		Attribute("approval_status")
		Attribute("approval_steps")
		Attribute("current_approver_id")
		Attribute("sla_deadline")
		Attribute("risk_assessment")
		Attribute("conditions")
		Attribute("created_at")
	})
	
	View("summary", func() {
		Description("Minimal view for references and quick lookups")
		Attribute("id")
		Attribute("request_type")
		Attribute("approval_status")
		Attribute("urgency_level")
		Attribute("business_reason")
		Attribute("created_at")
	})
})

// AccessRequestTemplate represents predefined templates for common access requests.
var AccessRequestTemplate = Type("AccessRequestTemplate", func() {
	Description("Predefined template for common access request patterns to streamline the request process")
	
	Field(1, "id", String, "Unique template identifier", func() {
		Format(FormatUUID)
		Example("template-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for template")
	})
	
	Field(2, "tenant_id", String, "Associated tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})
	
	Field(3, "name", String, "Template name", func() {
		Pattern("^[a-z][a-z0-9_]{2,99}$")
		MinLength(3)
		MaxLength(100)
		Example("financial_analyst_temporary")
		Description("Machine-readable template identifier")
	})
	
	Field(4, "display_name", String, "Human-readable template title", func() {
		MinLength(3)
		MaxLength(150)
		Example("Temporary Financial Analyst Access")
		Description("User-friendly name for UI display")
	})
	
	Field(5, "description", String, "Template description", func() {
		MaxLength(500)
		Example("Standard template for temporary financial analyst access during audit periods")
		Description("Explanation of template purpose and scope")
	})
	
	Field(6, "category", String, "Template category", func() {
		Enum("ROLE_BASED", "PROJECT_BASED", "EMERGENCY", "COMPLIANCE", "TEMPORARY", "DELEGATION")
		Example("TEMPORARY")
		Description("Classification for template organization")
	})
	
	Field(7, "request_type", String, "Default request type", func() {
		Enum("ROLE_ASSIGNMENT", "PERMISSION_GRANT", "RESOURCE_ACCESS", "ELEVATION", 
			"TEMPORARY_ACCESS", "EMERGENCY_ACCESS", "DELEGATION", "PROJECT_ACCESS")
		Example("ROLE_ASSIGNMENT")
		Description("Type of access request this template creates")
	})
	
	Field(8, "default_urgency", String, "Default urgency level", func() {
		Enum("LOW", "NORMAL", "HIGH", "URGENT", "EMERGENCY")
		Default("NORMAL")
		Example("HIGH")
		Description("Recommended urgency for this type of request")
	})
	
	Field(9, "default_duration_hours", UInt, "Default access duration", func() {
		Minimum(1)
		Maximum(8760)
		Example(72)
		Description("Recommended duration for this access type")
	})
	
	Field(10, "pre_approved_roles", ArrayOf(String), "Roles that can auto-approve", func() {
		Example([]string{"department_manager", "security_admin"})
		Description("Roles with authority to approve this template")
	})
	
	Field(11, "required_justification_fields", ArrayOf(String), "Required justification elements", func() {
		Example([]string{"business_case", "project_reference", "duration_rationale"})
		Description("Fields that must be completed in justification")
	})
	
	Field(12, "approval_workflow_id", String, "Associated approval workflow", func() {
		Format(FormatUUID)
		Example("workflow-456e7890-e89b-12d3-a456-426614174000")
		Description("Default workflow for requests using this template")
	})
	
	Field(13, "is_active", Boolean, "Template availability", func() {
		Default(true)
		Example(true)
		Description("Whether template is available for use")
	})
	
	Field(14, "usage_count", UInt, "Number of times template was used", func() {
		Example(247)
		Description("Statistical usage tracking")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "tenant_id", "name", "display_name", "category", "request_type", 
		"default_duration_hours", "is_active", "created_at")
})

// AccessReview represents periodic review of granted access for compliance.
var AccessReview = Type("AccessReview", func() {
	Description("Periodic review of user access rights for compliance and security governance")
	
	Field(1, "id", String, "Unique review identifier", func() {
		Format(FormatUUID)
		Example("review-123e4567-e89b-12d3-a456-426614174000")
		Description("Primary key for access review")
	})
	
	Field(2, "tenant_id", String, "Associated tenant identifier", func() {
		Format(FormatUUID)
		Example("550e8400-e29b-41d4-a716-446655440000")
		Description("Tenant isolation boundary")
	})
	
	Field(3, "review_type", String, "Type of access review", func() {
		Enum("PERIODIC", "TRIGGERED", "COMPLIANCE", "ROLE_CHANGE", "DEPARTURE", "AUDIT")
		Example("PERIODIC")
		Description("Reason for conducting the review")
	})
	
	Field(4, "scope", String, "Review scope", func() {
		Enum("USER_SPECIFIC", "ROLE_BASED", "ENTITY_WIDE", "TENANT_WIDE", "PROJECT_BASED")
		Example("ENTITY_WIDE")
		Description("Breadth of access being reviewed")
	})
	
	Field(5, "target_user_id", String, "User whose access is being reviewed", func() {
		Format(FormatUUID)
		Example("user-456e7890-e89b-12d3-a456-426614174000")
		Description("Specific user for user-focused reviews")
	})
	
	Field(6, "entity_id", String, "Entity scope for review", func() {
		Format(FormatUUID)
		Example("entity-789abc12-def3-4567-890a-bcdef1234567")
		Description("Organizational boundary for the review")
	})
	
	Field(7, "reviewer_id", String, "Assigned reviewer", func() {
		Format(FormatUUID)
		Example("manager-def456gh-i789-0123-def4-56gh78901234")
		Description("User responsible for conducting the review")
	})
	
	Field(8, "review_status", String, "Current review status", func() {
		Enum("SCHEDULED", "IN_PROGRESS", "COMPLETED", "OVERDUE", "CANCELLED")
		Default("SCHEDULED")
		Example("IN_PROGRESS")
		Description("Current state of the review process")
	})
	
	Field(9, "due_date", String, "Review completion deadline", func() {
		Format(FormatDateTime)
		Example("2023-12-15T17:00:00Z")
		Description("Target completion date for review")
	})
	
	Field(10, "access_items", ArrayOf(Type("AccessReviewItem", func() {
		Field(1, "item_id", String, "Review item identifier", func() {
			Format(FormatUUID)
		})
		Field(2, "access_type", String, "Type of access", func() {
			Enum("ROLE", "PERMISSION", "RESOURCE", "SYSTEM")
		})
		Field(3, "access_name", String, "Name of access", func() {
			Example("Financial Analyst Role")
		})
		Field(4, "current_status", String, "Current access status", func() {
			Enum("ACTIVE", "INACTIVE", "EXPIRED", "SUSPENDED")
		})
		Field(5, "review_decision", String, "Reviewer decision", func() {
			Enum("APPROVE", "REVOKE", "MODIFY", "PENDING")
		})
		Field(6, "justification", String, "Decision justification", func() {
			MaxLength(500)
		})
		Field(7, "last_used", String, "Last usage timestamp", func() {
			Format(FormatDateTime)
		})
		Required("item_id", "access_type", "access_name", "current_status")
	})), "Items being reviewed", func() {
		Description("Individual access rights under review")
	})
	
	Field(11, "findings", ArrayOf(String), "Review findings and recommendations", func() {
		Example([]string{
			"User has excessive permissions for current role",
			"Some access has not been used in 90+ days",
			"Missing approval documentation for recent role assignment",
		})
		Description("Issues identified during the review")
	})
	
	Field(12, "recommendations", ArrayOf(String), "Recommended actions", func() {
		Example([]string{
			"Remove unused financial reporting permissions",
			"Document business justification for elevated access",
			"Implement quarterly review schedule",
		})
		Description("Suggested improvements or actions")
	})
	
	Field(13, "completed_at", String, "Review completion timestamp", func() {
		Format(FormatDateTime)
		Example("2023-12-14T16:30:00Z")
		Description("When review was completed")
	})
	
	// Audit fields
	AuditFields()
	
	Required("id", "tenant_id", "review_type", "scope", "reviewer_id", "review_status", 
		"due_date", "access_items", "created_at")
})