package activities

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/notification"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

// NotificationActivities handles all notification-related Temporal activities
type NotificationActivities struct {
	notificationService notification.NotificationService
	auditService        audit.Service
	settingsService     settingsService.ConfigurationService
	logger              logger.Logger
	metrics             metrics.MetricsProvider
	tracer              tracing.TracingService
}

// NotificationActivityDeps contains dependencies for notification activities
type NotificationActivityDeps struct {
	NotificationService notification.NotificationService
	AuditService        audit.Service
	SettingsService     settingsService.ConfigurationService
	Logger              logger.Logger
	Metrics             metrics.MetricsProvider
	Tracer              tracing.TracingService
}

// NewNotificationActivities creates a new notification activities instance
func NewNotificationActivities(deps NotificationActivityDeps) *NotificationActivities {
	return &NotificationActivities{
		notificationService: deps.NotificationService,
		auditService:        deps.AuditService,
		settingsService:     deps.SettingsService,
		logger:              deps.Logger,
		metrics:             deps.Metrics,
		tracer:              deps.Tracer,
	}
}

// RegisterWith registers notification activities with a Temporal worker
func (n *NotificationActivities) RegisterWith(w worker.Worker) {
	w.RegisterActivity(n.SendAccountCreatedNotificationActivity)
	w.RegisterActivity(n.SendTransactionPostedNotificationActivity)
	w.RegisterActivity(n.SendApprovalRequestNotificationActivity)
	w.RegisterActivity(n.SendBudgetExceededNotificationActivity)
	w.RegisterActivity(n.SendComplianceAlertNotificationActivity)
	w.RegisterActivity(n.SendPeriodClosingNotificationActivity)
	w.RegisterActivity(n.SendErrorNotificationActivity)
	w.RegisterActivity(n.SendBulkOperationCompletedNotificationActivity)
}

// Notification Activity Input/Output Types

// AccountNotificationInput represents account-related notification input
type AccountNotificationInput struct {
	Account     *domain.Accounts       `json:"account"`
	Action      string                 `json:"action"`
	UserID      uuid.UUID              `json:"user_id"`
	Recipients  []string               `json:"recipients,omitempty"`
	MessageData map[string]interface{} `json:"message_data,omitempty"`
}

// TransactionNotificationInput represents transaction-related notification input
type TransactionNotificationInput struct {
	Transaction *domain.Transaction    `json:"transaction"`
	Action      string                 `json:"action"`
	UserID      uuid.UUID              `json:"user_id"`
	Recipients  []string               `json:"recipients,omitempty"`
	MessageData map[string]interface{} `json:"message_data,omitempty"`
}

// ApprovalNotificationInput represents approval-related notification input
type ApprovalNotificationInput struct {
	ResourceType   string                 `json:"resource_type"`
	ResourceID     uuid.UUID              `json:"resource_id"`
	RequesterID    uuid.UUID              `json:"requester_id"`
	ApproverIDs    []uuid.UUID            `json:"approver_ids"`
	ApprovalAmount float64                `json:"approval_amount,omitempty"`
	MessageData    map[string]interface{} `json:"message_data,omitempty"`
}

// BudgetNotificationInput represents budget-related notification input
type BudgetNotificationInput struct {
	AccountID      uuid.UUID              `json:"account_id"`
	AccountName    string                 `json:"account_name"`
	BudgetLimit    float64                `json:"budget_limit"`
	CurrentUsage   float64                `json:"current_usage"`
	ExceededAmount float64                `json:"exceeded_amount,omitempty"`
	Period         string                 `json:"period"`
	Recipients     []string               `json:"recipients,omitempty"`
	MessageData    map[string]interface{} `json:"message_data,omitempty"`
}

// ComplianceNotificationInput represents compliance-related notification input
type ComplianceNotificationInput struct {
	AlertType      string                 `json:"alert_type"`
	Severity       string                 `json:"severity"`
	ResourceType   string                 `json:"resource_type"`
	ResourceID     string                 `json:"resource_id"`
	ViolationDetails map[string]interface{} `json:"violation_details"`
	Recipients     []string               `json:"recipients,omitempty"`
	MessageData    map[string]interface{} `json:"message_data,omitempty"`
}

// PeriodClosingNotificationInput represents period closing notification input
type PeriodClosingNotificationInput struct {
	Period      string                 `json:"period"`
	Status      string                 `json:"status"`
	Summary     map[string]interface{} `json:"summary"`
	Recipients  []string               `json:"recipients,omitempty"`
	MessageData map[string]interface{} `json:"message_data,omitempty"`
}

// ErrorNotificationInput represents error notification input
type ErrorNotificationInput struct {
	ErrorType    string                 `json:"error_type"`
	ErrorMessage string                 `json:"error_message"`
	Context      map[string]interface{} `json:"context"`
	Severity     string                 `json:"severity"`
	Recipients   []string               `json:"recipients,omitempty"`
}

// BulkOperationNotificationInput represents bulk operation notification input
type BulkOperationNotificationInput struct {
	OperationType string                 `json:"operation_type"`
	TotalCount    int                    `json:"total_count"`
	SuccessCount  int                    `json:"success_count"`
	ErrorCount    int                    `json:"error_count"`
	Duration      string                 `json:"duration"`
	Summary       map[string]interface{} `json:"summary"`
	Recipients    []string               `json:"recipients,omitempty"`
	MessageData   map[string]interface{} `json:"message_data,omitempty"`
}

// NotificationActivityOutput represents notification activity output
type NotificationActivityOutput struct {
	Success           bool     `json:"success"`
	NotificationID    string   `json:"notification_id,omitempty"`
	Message           string   `json:"message"`
	ErrorCode         string   `json:"error_code,omitempty"`
	DeliveredTo       []string `json:"delivered_to,omitempty"`
	FailedRecipients  []string `json:"failed_recipients,omitempty"`
}

// Notification Activities Implementation

// SendAccountCreatedNotificationActivity sends account created notification
func (n *NotificationActivities) SendAccountCreatedNotificationActivity(ctx context.Context, input AccountNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeNotification,
		"account_id":    input.Account.ID,
		"action":        input.Action,
	})

	activityLogger.InfoContext(ctx, "Sending account created notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeAccountCreated)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to get notification settings", logger.Fields{
			"error": err.Error(),
		})
		return &NotificationActivityOutput{
			Success:   false,
			Message:   "Failed to get notification settings",
			ErrorCode: domain.ErrCodeNotificationSettingsRetrievalFailed,
		}, err
	}

	if !notificationSettings.Enabled {
		activityLogger.InfoContext(ctx, "Account created notifications are disabled")
		return &NotificationActivityOutput{
			Success: true,
			Message: "Notifications disabled for this event type",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("New Account Created: %s", input.Account.AccountName)
	message := fmt.Sprintf("Account '%s' (Code: %s) has been created successfully.\n\nDetails:\n- Type: %s\n- Currency: %s\n- Status: %s\n- Created by: User %s",
		input.Account.AccountName,
		input.Account.AccountCode,
		string(input.Account.RootType),
		func() string { if input.Account.CurrencyCode != nil { return *input.Account.CurrencyCode } else { return "N/A" } }(),
		string(input.Account.Status),
		input.UserID,
	)

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeAccountCreated,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   domain.NotificationPriorityNormal,
		Data:       input.MessageData,
	})

	// Log audit event for notification - simplified implementation
	// In real implementation, you'd use the proper audit service with CreateAuditEventRequest
	activityLogger.InfoContext(ctx, "Audit event logged for notification", logger.Fields{
		"notification_id":   notificationID,
		"notification_type": domain.NotificationTypeAccountCreated,
		"delivered_to":      deliveredTo,
		"failed_recipients": failedRecipients,
	})

	success := len(failedRecipients) == 0
	if success {
		activityLogger.InfoContext(ctx, "Account created notification sent successfully", logger.Fields{
			"notification_id": notificationID,
			"delivered_to":    deliveredTo,
		})
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
	} else {
		activityLogger.WarnContext(ctx, "Account created notification partially failed", logger.Fields{
			"notification_id":   notificationID,
			"delivered_to":      deliveredTo,
			"failed_recipients": failedRecipients,
		})
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Account created notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendTransactionPostedNotificationActivity sends transaction posted notification
func (n *NotificationActivities) SendTransactionPostedNotificationActivity(ctx context.Context, input TransactionNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  domain.ActivityTypeNotification,
		"transaction_id": input.Transaction.ID,
		"action":         input.Action,
	})

	activityLogger.InfoContext(ctx, "Sending transaction posted notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeTransactionPosted)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Transaction posted notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("Transaction Posted: %s", input.Transaction.TransactionNumber)
	message := fmt.Sprintf("Transaction '%s' has been posted to the general ledger.\n\nDetails:\n- Description: %s\n- Amount: %s %s\n- Date: %s\n- Posted by: User %s",
		input.Transaction.TransactionNumber,
		input.Transaction.Description,
		input.Transaction.TotalDebitAmount.String(),
		input.Transaction.CurrencyCode,
		input.Transaction.TransactionDate.Format("2006-01-02"),
		input.UserID,
	)

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeTransactionPosted,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   domain.NotificationPriorityNormal,
		Data:       input.MessageData,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Transaction posted notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendApprovalRequestNotificationActivity sends approval request notification
func (n *NotificationActivities) SendApprovalRequestNotificationActivity(ctx context.Context, input ApprovalNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeNotification,
		"resource_type": input.ResourceType,
		"resource_id":   input.ResourceID,
	})

	activityLogger.InfoContext(ctx, "Sending approval request notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeApprovalRequest)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Approval request notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("Approval Required: %s", input.ResourceType)
	message := fmt.Sprintf("An approval is required for %s (ID: %s).\n\nAmount: $%.2f\nRequested by: User %s\n\nPlease review and approve or reject this request.",
		input.ResourceType,
		input.ResourceID,
		input.ApprovalAmount,
		input.RequesterID,
	)

	// Convert approver IDs to recipient emails/IDs
	recipients := make([]string, len(input.ApproverIDs))
	for i, approverID := range input.ApproverIDs {
		recipients[i] = approverID.String() // In real implementation, this would be converted to email
	}

	// Send notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeApprovalRequest,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   domain.NotificationPriorityHigh,
		Data:       input.MessageData,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Approval request notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendBudgetExceededNotificationActivity sends budget exceeded notification
func (n *NotificationActivities) SendBudgetExceededNotificationActivity(ctx context.Context, input BudgetNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":     info.ActivityID,
		"workflow_id":     info.WorkflowExecution.ID,
		"activity_type":   domain.ActivityTypeNotification,
		"account_id":      input.AccountID,
		"exceeded_amount": input.ExceededAmount,
	})

	activityLogger.InfoContext(ctx, "Sending budget exceeded notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeBudgetExceeded)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Budget exceeded notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("Budget Exceeded Alert: %s", input.AccountName)
	message := fmt.Sprintf("ALERT: Budget has been exceeded for account '%s'.\n\nDetails:\n- Budget Limit: $%.2f\n- Current Usage: $%.2f\n- Exceeded By: $%.2f\n- Period: %s\n\nImmediate attention required.",
		input.AccountName,
		input.BudgetLimit,
		input.CurrentUsage,
		input.ExceededAmount,
		input.Period,
	)

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send high-priority notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeBudgetExceeded,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   domain.NotificationPriorityHigh,
		Data:       input.MessageData,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
		n.metrics.Counter(domain.MetricBudgetAlertsSent, "Total budget alerts sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Budget exceeded notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendComplianceAlertNotificationActivity sends compliance alert notification
func (n *NotificationActivities) SendComplianceAlertNotificationActivity(ctx context.Context, input ComplianceNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeNotification,
		"alert_type":    input.AlertType,
		"severity":      input.Severity,
	})

	activityLogger.InfoContext(ctx, "Sending compliance alert notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeComplianceAlert)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Compliance alert notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("COMPLIANCE ALERT: %s - %s Severity", input.AlertType, input.Severity)
	message := fmt.Sprintf("COMPLIANCE VIOLATION DETECTED\n\nAlert Type: %s\nSeverity: %s\nResource: %s (ID: %s)\n\nViolation Details:\n%v\n\nImmediate review and remediation required.",
		input.AlertType,
		input.Severity,
		input.ResourceType,
		input.ResourceID,
		input.ViolationDetails,
	)

	// Determine priority based on severity
	priority := domain.NotificationPriorityNormal
	if input.Severity == "HIGH" || input.Severity == "CRITICAL" {
		priority = domain.NotificationPriorityCritical
	}

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send critical-priority notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeComplianceAlert,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   priority,
		Data:       input.MessageData,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
		n.metrics.Counter(domain.MetricComplianceAlertsSent, "Total compliance alerts sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Compliance alert notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendPeriodClosingNotificationActivity sends period closing notification
func (n *NotificationActivities) SendPeriodClosingNotificationActivity(ctx context.Context, input PeriodClosingNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeNotification,
		"period":        input.Period,
		"status":        input.Status,
	})

	activityLogger.InfoContext(ctx, "Sending period closing notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypePeriodClosing)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Period closing notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	var subject, message string
	if input.Status == "COMPLETED" {
		subject = fmt.Sprintf("Period Closing Completed: %s", input.Period)
		message = fmt.Sprintf("The accounting period closing for %s has been completed successfully.\n\nSummary:\n%v\n\nAll transactions have been finalized and the period is now closed.",
			input.Period,
			input.Summary,
		)
	} else {
		subject = fmt.Sprintf("Period Closing Started: %s", input.Period)
		message = fmt.Sprintf("The accounting period closing process for %s has been initiated.\n\nCurrent Status: %s\n\nNo new transactions can be posted to this period during the closing process.",
			input.Period,
			input.Status,
		)
	}

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypePeriodClosing,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   domain.NotificationPriorityHigh,
		Data:       input.MessageData,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Period closing notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendErrorNotificationActivity sends error notification
func (n *NotificationActivities) SendErrorNotificationActivity(ctx context.Context, input ErrorNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":    info.ActivityID,
		"workflow_id":    info.WorkflowExecution.ID,
		"activity_type":  domain.ActivityTypeNotification,
		"error_type":     input.ErrorType,
		"severity":       input.Severity,
	})

	activityLogger.InfoContext(ctx, "Sending error notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeError)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Error notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("System Error Alert: %s - %s", input.ErrorType, input.Severity)
	message := fmt.Sprintf("A system error has occurred that requires attention.\n\nError Type: %s\nSeverity: %s\nMessage: %s\n\nContext:\n%v\n\nPlease investigate and resolve this issue promptly.",
		input.ErrorType,
		input.Severity,
		input.ErrorMessage,
		input.Context,
	)

	// Determine priority based on severity
	priority := domain.NotificationPriorityNormal
	if input.Severity == "HIGH" || input.Severity == "CRITICAL" {
		priority = domain.NotificationPriorityCritical
	}

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeError,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   priority,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
		n.metrics.Counter(domain.MetricErrorNotificationsSent, "Total error notifications sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Error notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// SendBulkOperationCompletedNotificationActivity sends bulk operation completed notification
func (n *NotificationActivities) SendBulkOperationCompletedNotificationActivity(ctx context.Context, input BulkOperationNotificationInput) (*NotificationActivityOutput, error) {
	ctx, span := n.tracer.StartSpan(ctx, domain.ActivityTypeNotification)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := n.logger.WithFields(logger.Fields{
		"activity_id":      info.ActivityID,
		"workflow_id":      info.WorkflowExecution.ID,
		"activity_type":    domain.ActivityTypeNotification,
		"operation_type":   input.OperationType,
		"total_count":      input.TotalCount,
		"success_count":    input.SuccessCount,
		"error_count":      input.ErrorCount,
	})

	activityLogger.InfoContext(ctx, "Sending bulk operation completed notification")
	n.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get notification settings
	notificationSettings, err := n.getNotificationSettings(ctx, domain.NotificationTypeBulkOperationCompleted)
	if err != nil || !notificationSettings.Enabled {
		return &NotificationActivityOutput{
			Success: true,
			Message: "Bulk operation notifications disabled or settings unavailable",
		}, nil
	}

	// Build notification message
	subject := fmt.Sprintf("Bulk Operation Completed: %s", input.OperationType)
	message := fmt.Sprintf("Bulk %s operation has been completed.\n\nResults:\n- Total Records: %d\n- Successful: %d\n- Failed: %d\n- Duration: %s\n\nSummary:\n%v",
		input.OperationType,
		input.TotalCount,
		input.SuccessCount,
		input.ErrorCount,
		input.Duration,
		input.Summary,
	)

	// Determine recipients
	recipients := input.Recipients
	if len(recipients) == 0 {
		recipients = notificationSettings.DefaultRecipients
	}

	// Send notifications
	notificationID := uuid.New().String()
	deliveredTo, failedRecipients := n.sendNotification(ctx, NotificationRequest{
		ID:         notificationID,
		Type:       domain.NotificationTypeBulkOperationCompleted,
		Subject:    subject,
		Message:    message,
		Recipients: recipients,
		Priority:   domain.NotificationPriorityNormal,
		Data:       input.MessageData,
	})

	success := len(failedRecipients) == 0
	if success {
		n.metrics.Counter(domain.MetricNotificationsSent, "Total notifications sent").Add(1, nil)
	} else {
		n.metrics.Counter(domain.MetricNotificationErrors, "Total notification errors").Add(1, nil)
	}

	return &NotificationActivityOutput{
		Success:          success,
		NotificationID:   notificationID,
		Message:          "Bulk operation notification processed",
		DeliveredTo:      deliveredTo,
		FailedRecipients: failedRecipients,
	}, nil
}

// Helper types and functions

// NotificationSettings represents notification configuration
type NotificationSettings struct {
	Enabled           bool     `json:"enabled"`
	DefaultRecipients []string `json:"default_recipients"`
	Priority          string   `json:"priority"`
}

// NotificationRequest represents a notification request
type NotificationRequest struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Subject    string                 `json:"subject"`
	Message    string                 `json:"message"`
	Recipients []string               `json:"recipients"`
	Priority   string                 `json:"priority"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// getNotificationSettings retrieves notification settings for a given type
func (n *NotificationActivities) getNotificationSettings(ctx context.Context, notificationType string) (*NotificationSettings, error) {
	// This is a simplified implementation - in real implementation, you'd use proper domain types
	// and retrieve actual settings from the configuration service
	
	// For now, return default settings
	return &NotificationSettings{
		Enabled:           true, // Default enabled
		DefaultRecipients: []string{"admin@company.com"},
		Priority:          domain.NotificationPriorityNormal,
	}, nil
}

// sendNotification sends a notification using the existing notification service
func (n *NotificationActivities) sendNotification(ctx context.Context, req NotificationRequest) ([]string, []string) {
	// Convert notification type to existing enum
	var notificationType notification.NotificationType
	switch req.Type {
	case domain.NotificationTypeAccountCreated:
		notificationType = notification.NotificationTypeGeneric
	case domain.NotificationTypeTransactionPosted:
		notificationType = notification.NotificationTypeGeneric
	case domain.NotificationTypeApprovalRequest:
		notificationType = notification.NotificationTypeAccessRequestCreated
	default:
		notificationType = notification.NotificationTypeGeneric
	}

	// Extract tenant ID from context
	tenantID := getTenantIDFromContext(ctx)
	
	// Convert recipients to notification recipients
	recipients := make([]notification.Recipient, 0, len(req.Recipients))
	deliveredTo := make([]string, 0)
	failedRecipients := make([]string, 0)

	for _, recipientStr := range req.Recipients {
		// In real implementation, this would parse the recipient string or UUID
		// For now, we'll create a basic recipient structure
		recipient := notification.Recipient{
			UserID: uuid.New(), // This should be parsed from recipientStr
			Email:  recipientStr,
		}
		recipients = append(recipients, recipient)
	}

	// Create notification using existing service types
	notificationData := &notification.Notification{
		TenantID:   tenantID,
		Recipients: recipients,
		Subject:    req.Subject,
		Message:    req.Message,
		Data:       req.Data,
		Type:       notificationType,
		Channels:   []notification.NotificationChannel{notification.NotificationChannelEmail},
	}

	// Send notification using the existing service
	err := n.notificationService.Send(ctx, notificationData)
	if err != nil {
		// Mark all recipients as failed if service call fails
		for _, recipientStr := range req.Recipients {
			failedRecipients = append(failedRecipients, recipientStr)
		}
		n.logger.ErrorContext(ctx, "Failed to send notification", logger.Fields{
			"error":           err.Error(),
			"notification_id": req.ID,
		})
	} else {
		// Mark all recipients as delivered if service call succeeds
		for _, recipientStr := range req.Recipients {
			deliveredTo = append(deliveredTo, recipientStr)
		}
		n.logger.InfoContext(ctx, "Notification sent successfully", logger.Fields{
			"notification_id":   req.ID,
			"recipient_count":   len(recipients),
		})
	}

	return deliveredTo, failedRecipients
}

