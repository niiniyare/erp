package workflows

import (
	"time"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/notification"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ComplianceWorkflows contains all compliance-related Temporal workflows
type ComplianceWorkflows struct {
	deps ComplianceWorkflowDeps
}

// ComplianceWorkflowDeps contains dependencies for compliance workflows
type ComplianceWorkflowDeps struct {
	Services            *service.Services
	IAMService          iam.Service
	AuditService        audit.Service
	FeatureFlagService  featureflag.Service
	SettingsService     settingsService.ConfigurationService
	NotificationService notification.NotificationService
	CacheService        cache.Service
	Logger              loggerPkg.Logger
	Metrics             metrics.MetricsProvider
	Tracer              tracing.TracingService
}

// NewComplianceWorkflows creates new compliance workflows
func NewComplianceWorkflows(deps ComplianceWorkflowDeps) *ComplianceWorkflows {
	return &ComplianceWorkflows{
		deps: deps,
	}
}

// ComplianceAuditWorkflow handles compliance audit processes
func (cw *ComplianceWorkflows) ComplianceAuditWorkflow(ctx workflow.Context, input domain.ComplianceAuditWorkflowInput) (*domain.ComplianceAuditWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting compliance audit workflow", "audit_type", input.AuditType, "period", input.AuditPeriod)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour * 2, // Compliance audits can take longer
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
			InitialInterval: time.Minute * 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.ComplianceAuditWorkflowResult{
		AuditType: input.AuditType,
		Period:    input.AuditPeriod,
		StartTime: workflow.Now(ctx),
	}

	// Step 1: Initialize audit
	var auditInit domain.AuditInitializationResult
	err := workflow.ExecuteActivity(ctx, "InitializeComplianceAudit", domain.AuditInitializationInput{
		AuditType:       input.AuditType,
		AuditPeriod:     input.AuditPeriod,
		AuditScope:      input.AuditScope,
		InitiatedBy:     input.InitiatedBy,
		ComplianceRules: input.ComplianceRules,
	}).Get(ctx, &auditInit)

	if err != nil {
		logger.Error("Audit initialization failed", "error", err)
		result.Status = domain.AuditStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.AuditID = auditInit.AuditID

	// Step 2: Collect audit data
	var dataCollection domain.AuditDataCollectionResult
	err = workflow.ExecuteActivity(ctx, "CollectAuditData", domain.AuditDataCollectionInput{
		AuditID:    auditInit.AuditID,
		AuditType:  input.AuditType,
		AuditScope: input.AuditScope,
		PeriodFrom: input.PeriodFrom,
		PeriodTo:   input.PeriodTo,
	}).Get(ctx, &dataCollection)

	if err != nil {
		logger.Error("Audit data collection failed", "error", err)
		result.Status = domain.AuditStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 3: Run compliance checks
	var complianceResults domain.ComplianceCheckResults
	err = workflow.ExecuteActivity(ctx, "RunComplianceChecks", domain.ComplianceCheckInput{
		AuditID:         auditInit.AuditID,
		ComplianceRules: input.ComplianceRules,
		AuditData:       dataCollection.CollectedData,
	}).Get(ctx, &complianceResults)

	if err != nil {
		logger.Error("Compliance checks failed", "error", err)
		result.Status = domain.AuditStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.ComplianceScore = complianceResults.ComplianceScore
	result.ViolationCount = len(complianceResults.Violations)
	result.Violations = complianceResults.Violations

	// Step 4: Generate audit report
	var auditReport domain.AuditReportResult
	err = workflow.ExecuteActivity(ctx, "GenerateComplianceReport", domain.AuditReportInput{
		AuditID:         auditInit.AuditID,
		AuditType:       input.AuditType,
		ComplianceScore: complianceResults.ComplianceScore,
		Violations:      complianceResults.Violations,
		Recommendations: complianceResults.Recommendations,
	}).Get(ctx, &auditReport)

	if err != nil {
		logger.Error("Audit report generation failed", "error", err)
		result.Status = domain.AuditStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.ReportID = auditReport.ReportID
	result.ReportPath = auditReport.ReportPath

	// Step 5: Send notifications for violations
	if len(complianceResults.Violations) > 0 {
		err = workflow.ExecuteActivity(ctx, "SendComplianceAlert", domain.ComplianceAlertInput{
			AuditID:    auditInit.AuditID,
			AlertType:  domain.AlertTypeViolation,
			Violations: complianceResults.Violations,
			Severity:   complianceResults.MaxSeverity,
			Recipients: input.NotificationRecipients,
		}).Get(ctx, nil)

		if err != nil {
			logger.Warn("Failed to send compliance alert", "error", err)
		}
	}

	result.Status = domain.AuditStatusCompleted
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Compliance audit workflow completed",
		"audit_id", result.AuditID,
		"compliance_score", result.ComplianceScore,
		"violation_count", result.ViolationCount,
		"duration", result.Duration)

	return result, nil
}

// FraudDetectionWorkflow handles fraud detection and analysis
func (cw *ComplianceWorkflows) FraudDetectionWorkflow(ctx workflow.Context, input domain.FraudDetectionWorkflowInput) (*domain.FraudDetectionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting fraud detection workflow", "detection_type", input.DetectionType)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 30,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: time.Second * 10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	result := &domain.FraudDetectionWorkflowResult{
		DetectionType: input.DetectionType,
		StartTime:     workflow.Now(ctx),
		Alerts:        make([]domain.FraudAlert, 0),
	}

	// Step 1: Initialize fraud detection
	var detection domain.FraudDetectionInitResult
	err := workflow.ExecuteActivity(ctx, "InitializeFraudDetection", domain.FraudDetectionInitInput{
		DetectionType:    input.DetectionType,
		DetectionRules:   input.DetectionRules,
		MonitoringPeriod: input.MonitoringPeriod,
		InitiatedBy:      input.InitiatedBy,
	}).Get(ctx, &detection)

	if err != nil {
		logger.Error("Fraud detection initialization failed", "error", err)
		result.Status = domain.FraudDetectionStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.DetectionID = detection.DetectionID

	// Step 2: Analyze transactions for fraud patterns
	var patternAnalysis domain.FraudPatternAnalysisResult
	err = workflow.ExecuteActivity(ctx, "AnalyzeFraudPatterns", domain.FraudPatternAnalysisInput{
		DetectionID:     detection.DetectionID,
		DetectionRules:  input.DetectionRules,
		TransactionData: input.TransactionData,
		HistoricalData:  input.HistoricalData,
	}).Get(ctx, &patternAnalysis)

	if err != nil {
		logger.Error("Fraud pattern analysis failed", "error", err)
		result.Status = domain.FraudDetectionStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 3: Generate fraud risk scores
	var riskScoring domain.FraudRiskScoringResult
	err = workflow.ExecuteActivity(ctx, "CalculateFraudRiskScores", domain.FraudRiskScoringInput{
		DetectionID: detection.DetectionID,
		Patterns:    patternAnalysis.DetectedPatterns,
		RiskFactors: patternAnalysis.RiskFactors,
	}).Get(ctx, &riskScoring)

	if err != nil {
		logger.Error("Fraud risk scoring failed", "error", err)
		result.Status = domain.FraudDetectionStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	// Step 4: Evaluate high-risk transactions
	for _, transaction := range riskScoring.HighRiskTransactions {
		if transaction.RiskScore >= input.AlertThreshold {
			// Create fraud alert
			alert := domain.FraudAlert{
				TransactionID: transaction.TransactionID,
				AlertType:     domain.FraudAlertTypeHighRisk,
				RiskScore:     transaction.RiskScore,
				RiskFactors:   transaction.RiskFactors,
				DetectedAt:    workflow.Now(ctx),
				Status:        domain.FraudAlertStatusActive,
			}
			result.Alerts = append(result.Alerts, alert)

			// Send immediate alert for high-risk transactions
			if transaction.RiskScore >= input.CriticalThreshold {
				err = workflow.ExecuteActivity(ctx, "SendFraudAlert", domain.FraudAlertNotificationInput{
					DetectionID: detection.DetectionID,
					Alert:       alert,
					Recipients:  input.AlertRecipients,
					Priority:    domain.AlertPriorityHigh,
				}).Get(ctx, nil)

				if err != nil {
					logger.Warn("Failed to send critical fraud alert", "error", err, "transaction_id", transaction.TransactionID)
				}
			}
		}
	}

	// Step 5: Generate detection report
	var detectionReport domain.FraudDetectionReportResult
	err = workflow.ExecuteActivity(ctx, "GenerateFraudDetectionReport", domain.FraudDetectionReportInput{
		DetectionID:          detection.DetectionID,
		DetectionType:        input.DetectionType,
		AnalyzedTransactions: len(input.TransactionData),
		HighRiskTransactions: len(riskScoring.HighRiskTransactions),
		Alerts:               result.Alerts,
		RiskFactors:          riskScoring.TopRiskFactors,
	}).Get(ctx, &detectionReport)

	if err != nil {
		logger.Error("Fraud detection report generation failed", "error", err)
		result.Status = domain.FraudDetectionStatusFailed
		result.Error = err.Error()
		return result, nil
	}

	result.ReportID = detectionReport.ReportID
	result.ReportPath = detectionReport.ReportPath
	result.RiskFactors = riskScoring.TopRiskFactors
	result.AnalyzedTransactions = len(input.TransactionData)
	result.HighRiskTransactions = len(riskScoring.HighRiskTransactions)

	result.Status = domain.FraudDetectionStatusCompleted
	result.CompletedTime = workflow.Now(ctx)
	result.Duration = result.CompletedTime.Sub(result.StartTime)

	logger.Info("Fraud detection workflow completed",
		"detection_id", result.DetectionID,
		"analyzed_transactions", result.AnalyzedTransactions,
		"high_risk_transactions", result.HighRiskTransactions,
		"alert_count", len(result.Alerts),
		"duration", result.Duration)

	return result, nil
}
