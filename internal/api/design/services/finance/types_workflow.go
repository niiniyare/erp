package finance

import (
	. "github.com/niiniyare/erp/internal/api/design/types"
	. "goa.design/goa/v3/dsl"
)

// Transaction Workflow Management Types

var TransactionStatusResult = Type("TransactionStatusResult", func() {
	Description("Detailed transaction status and workflow progress")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_number", String, "Transaction number")
	Attribute("status", String, "Current transaction status", func() {
		Enum("submitted", "awaiting_approval", "processing", "posted", "rejected", "cancelled")
	})
	Attribute("current_stage", String, "Current workflow stage", func() {
		Enum("validation", "manager_review", "cfo_approval", "posting", "balance_update", "completed")
	})
	Attribute("progress_percentage", Int32, "Completion percentage (0-100)")
	Attribute("estimated_completion", String, "Estimated completion time", func() {
		Format(FormatDateTime)
	})
	Attribute("workflow_history", ArrayOf(WorkflowStageResult), "Complete workflow history")
	Attribute("available_actions", ArrayOf(String), "Available actions for current user")
	Attribute("next_approver", ApproverResult, "Next approver information")

	Required("id", "transaction_number", "status", "current_stage", "progress_percentage", "available_actions")
})

var WorkflowStageResult = Type("WorkflowStageResult", func() {
	Description("Workflow stage information")

	Attribute("stage", String, "Stage name")
	Attribute("status", String, "Stage status", func() {
		Enum("pending", "in_progress", "completed", "skipped", "failed")
	})
	Attribute("started_at", String, "Stage start time", func() {
		Format(FormatDateTime)
	})
	Attribute("completed_at", String, "Stage completion time", func() {
		Format(FormatDateTime)
	})
	Attribute("actor", String, "User who performed the action", func() {
		Format(FormatUUID)
	})
	Attribute("assigned_to", String, "User assigned to this stage", func() {
		Format(FormatUUID)
	})
	Attribute("due_date", String, "Due date for this stage", func() {
		Format(FormatDateTime)
	})
	Attribute("validation_results", ValidationResult, "Validation results if applicable")
	Attribute("comments", String, "Stage comments or notes")

	Required("stage", "status")
})

var ApproverResult = Type("ApproverResult", func() {
	Description("Approver information")

	Attribute("id", String, "Approver user ID", func() {
		Format(FormatUUID)
	})
	Attribute("name", String, "Approver full name")
	Attribute("role", String, "Approver role")
	Attribute("email", String, "Approver email", func() {
		Format(FormatEmail)
	})
	Attribute("department", String, "Approver department")

	Required("id", "name", "role")
})

var ApprovalDecisionPayload = Type("ApprovalDecisionPayload", func() {
	Description("Payload for submitting approval decisions")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("decision", String, "Approval decision", func() {
		Enum("approved", "rejected", "request_changes")
	})
	Attribute("comments", String, "Approval comments", func() {
		MaxLength(1000)
	})
	Attribute("approver_id", String, "Approver user ID", func() {
		Format(FormatUUID)
	})
	Attribute("approval_level", String, "Level of approval", func() {
		Enum("manager", "director", "cfo", "system")
	})
	Attribute("escalation_reason", String, "Reason for escalation (optional)")

	Required("id", "decision", "approver_id", "approval_level")
})

var ApprovalDecisionResult = Type("ApprovalDecisionResult", func() {
	Description("Result of approval decision submission")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("status", String, "Updated transaction status")
	Attribute("current_stage", String, "Updated workflow stage")
	Attribute("approval_decision", ApprovalDecisionData, "Approval decision details")
	Attribute("estimated_completion", String, "Updated completion estimate", func() {
		Format(FormatDateTime)
	})
	Attribute("next_stage", String, "Next workflow stage")

	Required("id", "status", "current_stage", "approval_decision")
})

var ApprovalDecisionData = Type("ApprovalDecisionData", func() {
	Description("Approval decision data")

	Attribute("decision", String, "Decision made")
	Attribute("approver", ApproverResult, "Approver who made decision")
	Attribute("approved_at", String, "Decision timestamp", func() {
		Format(FormatDateTime)
	})
	Attribute("comments", String, "Approval comments")
})

var ChangeRequestPayload = Type("ChangeRequestPayload", func() {
	Description("Payload for requesting transaction changes")

	Attribute("id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("requested_by", String, "User requesting changes", func() {
		Format(FormatUUID)
	})
	Attribute("reason", String, "Reason for change request", func() {
		MinLength(1)
		MaxLength(500)
	})
	Attribute("required_changes", ArrayOf(RequiredChangeItem), "List of required changes")
	Attribute("due_date", String, "Due date for changes", func() {
		Format(FormatDateTime)
	})
	Attribute("priority", String, "Change request priority", func() {
		Enum("low", "medium", "high", "urgent")
		Default("medium")
	})

	Required("id", "requested_by", "reason", "required_changes")
})

var RequiredChangeItem = Type("RequiredChangeItem", func() {
	Description("Individual change requirement")

	Attribute("field", String, "Field that needs to be changed")
	Attribute("current_value", String, "Current field value")
	Attribute("suggested_value", String, "Suggested new value")
	Attribute("reason", String, "Reason for this specific change")
	Attribute("is_mandatory", Boolean, "Whether this change is mandatory", func() {
		Default(true)
	})

	Required("field", "reason")
})

var ChangeRequestResult = Type("ChangeRequestResult", func() {
	Description("Result of change request submission")

	Attribute("change_request_id", String, "Change request ID", func() {
		Format(FormatUUID)
	})
	Attribute("transaction_id", String, "Original transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("status", String, "Change request status", func() {
		Enum("pending", "in_progress", "completed", "rejected")
	})
	Attribute("assigned_to", String, "User assigned to make changes", func() {
		Format(FormatUUID)
	})
	Attribute("due_date", String, "Due date for changes", func() {
		Format(FormatDateTime)
	})

	// Use common audit fields
	AuditFields()

	Required("change_request_id", "transaction_id", "status")
})

var TransactionWorkflowResult = Type("TransactionWorkflowResult", func() {
	Description("Complete workflow information for a transaction")

	Attribute("transaction_id", String, "Transaction ID", func() {
		Format(FormatUUID)
	})
	Attribute("workflow_template", String, "Workflow template used")
	Attribute("stages", ArrayOf(WorkflowStageResult), "All workflow stages")
	Attribute("available_actions", MapOf(String, WorkflowActionResult), "Available actions by stage")
	Attribute("escalation_rules", ArrayOf(EscalationRuleResult), "Applicable escalation rules")
	Attribute("sla_metrics", SLAMetrics, "SLA tracking metrics")

	Required("transaction_id", "workflow_template", "stages", "available_actions")
})

var WorkflowActionResult = Type("WorkflowActionResult", func() {
	Description("Available workflow action")

	Attribute("action", String, "Action name")
	Attribute("label", String, "Human-readable label")
	Attribute("description", String, "Action description")
	Attribute("requires_comment", Boolean, "Whether action requires a comment")
	Attribute("permissions_required", ArrayOf(String), "Required permissions")

	Required("action", "label")
})

var EscalationRuleResult = Type("EscalationRuleResult", func() {
	Description("Escalation rule information")

	Attribute("rule_name", String, "Escalation rule name")
	Attribute("trigger_condition", String, "Condition that triggers escalation")
	Attribute("escalate_to", String, "User to escalate to", func() {
		Format(FormatUUID)
	})
	Attribute("escalation_delay_hours", Int32, "Hours before escalation")
	Attribute("is_active", Boolean, "Whether rule is currently active")

	Required("rule_name", "trigger_condition")
})

var SLAMetrics = Type("SLAMetrics", func() {
	Description("Service Level Agreement metrics")

	Attribute("target_completion_hours", Int32, "Target completion time in hours")
	Attribute("elapsed_hours", Int32, "Elapsed time in hours")
	Attribute("remaining_hours", Int32, "Remaining time in hours")
	Attribute("is_overdue", Boolean, "Whether workflow is overdue")
})
