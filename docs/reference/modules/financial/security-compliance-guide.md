# AWO ERP Financial Module - Security & Compliance Guide

**Version**: 1.0  
**Date**: January 2025  
**Classification**: Internal Technical Documentation  

---

## Table of Contents

1. [Security Architecture Overview](#security-architecture-overview)
2. [ABAC Integration for Financial Operations](#abac-integration-for-financial-operations)
3. [Audit & Compliance Framework](#audit-compliance-framework)
4. [Data Protection & Privacy](#data-protection-privacy)
5. [Regulatory Compliance](#regulatory-compliance)
6. [Security Controls & Monitoring](#security-controls-monitoring)
7. [Threat Modeling & Risk Assessment](#threat-modeling-risk-assessment)
8. [Security Testing & Validation](#security-testing-validation)
9. [Incident Response & Recovery](#incident-response-recovery)
10. [Security Best Practices](#security-best-practices)

---

## Security Architecture Overview

### **Defense in Depth Strategy**

The AWO ERP Financial Module implements a  defense-in-depth security architecture that protects financial data at multiple layers:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Application Security Layer                   │
│  • Input validation and sanitization                           │
│  • Business logic authorization                                │
│  • Rate limiting and DDoS protection                          │
│  • Session management and timeout                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  ABAC Authorization Layer                       │
│  • Attribute-based access control                             │
│  • Context-aware decision making                              │
│  • Policy evaluation engine                                   │
│  • Real-time risk assessment                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Data Access Security Layer                   │
│  • Row-level security (RLS)                                   │
│  • Column-level encryption                                    │
│  • Database audit trails                                      │
│  • Connection security and pooling                            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                 Infrastructure Security Layer                  │
│  • Network segmentation and firewalls                         │
│  • TLS encryption in transit                                  │
│  • Key management and rotation                                │
│  • Infrastructure monitoring                                  │
└─────────────────────────────────────────────────────────────────┘
```

### **Security Principles**

#### **1. Zero Trust Architecture**
```go
// Every request must be authenticated and authorized
func (s *FinancialService) ProcessTransaction(ctx context.Context, req TransactionRequest) error {
    // 1. Authentication verification
    userID, err := s.auth.ValidateToken(ctx, req.Token)
    if err != nil {
        return errors.Unauthorized("invalid authentication")
    }
    
    // 2. Tenant context validation
    tenantID, err := s.tenant.ValidateContext(ctx, req.TenantID)
    if err != nil {
        return errors.Forbidden("invalid tenant context")
    }
    
    // 3. ABAC authorization
    authorized, err := s.abac.Authorize(ctx, abac.Request{
        Subject:  abac.UserSubject(userID),
        Resource: abac.TransactionResource(req.Type),
        Action:   "create",
        Context:  abac.TenantContext(tenantID),
        Attributes: map[string]interface{}{
            "transaction_amount": req.Amount,
            "transaction_type":   req.Type,
            "risk_level":         s.calculateRiskLevel(req),
        },
    })
    if err != nil || !authorized {
        return errors.Forbidden("insufficient permissions")
    }
    
    // 4. Business logic execution with audit
    return s.executeTransaction(ctx, req, userID, tenantID)
}
```

#### **2. Principle of Least Privilege**
```go
// Financial ABAC policies implement granular permissions
var FinancialPolicies = []abac.Policy{
    {
        ID: "transaction-create-basic",
        Subjects: []abac.Condition{
            abac.RoleCondition("accountant"),
            abac.DepartmentCondition("finance"),
        },
        Resources: []abac.Condition{
            abac.ResourceTypeCondition("transaction"),
        },
        Actions: []abac.Condition{
            abac.ActionCondition("create"),
        },
        Context: []abac.Condition{
            abac.TenantCondition(),
            abac.AmountCondition("<=", 10000), // Limit based on amount
            abac.TimeCondition("business_hours"),
            abac.LocationCondition("authorized_location"),
        },
        Effect: abac.EffectAllow,
    },
}
```

#### **3. Defense Against Financial Fraud**
```go
// @internal/core/finance/security/fraud_detection.go

type FraudDetectionService struct {
    riskEngine    RiskEngine
    anomalyDetector AnomalyDetector
    audit         audit.Service
    alerting      AlertingService
}

func (s *FraudDetectionService) EvaluateTransaction(ctx context.Context, transaction *Transaction) (*RiskAssessment, error) {
    assessment := &RiskAssessment{
        TransactionID: transaction.ID,
        TenantID:      transaction.TenantID,
        EvaluatedAt:   time.Now(),
        RiskScore:     0,
        RiskFactors:   []RiskFactor{},
    }
    
    // Amount-based risk assessment
    if transaction.TotalAmount.GreaterThan(decimal.NewFromInt(100000)) {
        assessment.RiskScore += 50
        assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
            Type:        "high_amount",
            Description: "Transaction amount exceeds normal threshold",
            Score:       50,
        })
    }
    
    // Time-based anomaly detection
    if s.isAfterHours(transaction.CreatedAt) {
        assessment.RiskScore += 25
        assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
            Type:        "after_hours",
            Description: "Transaction created outside business hours",
            Score:       25,
        })
    }
    
    // User behavior analysis
    userPattern, err := s.anomalyDetector.AnalyzeUserBehavior(ctx, transaction.CreatedBy)
    if err != nil {
        return nil, err
    }
    
    if userPattern.IsAnomalous {
        assessment.RiskScore += userPattern.AnomalyScore
        assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
            Type:        "behavioral_anomaly",
            Description: userPattern.Description,
            Score:       userPattern.AnomalyScore,
        })
    }
    
    // Velocity checks - rapid successive transactions
    recentCount, err := s.getRecentTransactionCount(ctx, transaction.CreatedBy, time.Hour)
    if err != nil {
        return nil, err
    }
    
    if recentCount > 10 {
        assessment.RiskScore += 30
        assessment.RiskFactors = append(assessment.RiskFactors, RiskFactor{
            Type:        "high_velocity",
            Description: fmt.Sprintf("User created %d transactions in the last hour", recentCount),
            Score:       30,
        })
    }
    
    // Determine risk level
    assessment.RiskLevel = s.categorizeRisk(assessment.RiskScore)
    
    // Trigger alerts for high-risk transactions
    if assessment.RiskLevel >= RiskLevelHigh {
        go s.triggerSecurityAlert(ctx, transaction, assessment)
    }
    
    return assessment, nil
}

func (s *FraudDetectionService) triggerSecurityAlert(ctx context.Context, transaction *Transaction, assessment *RiskAssessment) {
    alert := SecurityAlert{
        ID:            uuid.New(),
        Type:          "high_risk_transaction",
        Severity:      "high",
        TenantID:      transaction.TenantID,
        TransactionID: transaction.ID,
        RiskScore:     assessment.RiskScore,
        RiskFactors:   assessment.RiskFactors,
        CreatedAt:     time.Now(),
        RequiresReview: true,
    }
    
    // Send to security operations center
    if err := s.alerting.SendAlert(ctx, alert); err != nil {
        // Log but don't fail the transaction
        log.Error("Failed to send security alert", "error", err, "alert_id", alert.ID)
    }
    
    // Create audit entry
    s.audit.LogSecurityEvent(ctx, audit.Event{
        Type:        "security.high_risk_transaction",
        TenantID:    transaction.TenantID,
        UserID:      transaction.CreatedBy,
        ResourceID:  string(transaction.ID),
        RiskScore:   audit.RiskCritical,
        Data: map[string]interface{}{
            "risk_assessment": assessment,
            "alert_id":        alert.ID,
        },
        Timestamp: time.Now(),
    })
}
```

---

## ABAC Integration for Financial Operations

### **Financial Policy Framework**

#### **Account Management Policies**
```go
// @internal/core/finance/policies/account_policies.go

var AccountManagementPolicies = []abac.Policy{
    {
        ID:          "account-create-standard",
        Name:        "Standard Account Creation",
        Description: "Allows accountants to create standard accounts during business hours",
        Effect:      abac.EffectAllow,
        Version:     "1.0",
        
        Subjects: []abac.Condition{
            abac.RoleCondition("accountant", "senior_accountant"),
            abac.DepartmentCondition("finance", "accounting"),
            abac.NOT(abac.UserStatusCondition("suspended", "terminated")),
        },
        
        Resources: []abac.Condition{
            abac.ResourceTypeCondition("account"),
            abac.NOT(abac.AccountTypeCondition("bank", "cash")), // Exclude sensitive accounts
        },
        
        Actions: []abac.Condition{
            abac.ActionCondition("create", "update"),
        },
        
        Context: []abac.Condition{
            abac.TenantCondition(),
            abac.TimeCondition("business_hours"),
            abac.LocationCondition("office_network", "vpn_network"),
            abac.NOT(abac.DateCondition("fiscal_year_end")), // Restrict during year-end
        },
        
        Obligations: []abac.Obligation{
            {
                Type: "audit_enhanced",
                Config: map[string]interface{}{
                    "audit_level": "detailed",
                    "real_time":   true,
                },
            },
        },
    },
    
    {
        ID:          "account-create-sensitive",
        Name:        "Sensitive Account Creation",
        Description: "Strict controls for bank and cash account creation",
        Effect:      abac.EffectAllow,
        Version:     "1.0",
        
        Subjects: []abac.Condition{
            abac.RoleCondition("finance_manager", "controller", "cfo"),
            abac.CertificationCondition("financial_controls_certified"),
            abac.MFACondition("required"),
        },
        
        Resources: []abac.Condition{
            abac.ResourceTypeCondition("account"),
            abac.AccountTypeCondition("bank", "cash"),
        },
        
        Actions: []abac.Condition{
            abac.ActionCondition("create", "update"),
        },
        
        Context: []abac.Condition{
            abac.TenantCondition(),
            abac.TimeCondition("business_hours"),
            abac.LocationCondition("secure_office"),
            abac.ApprovalCondition("dual_approval_required"),
            abac.ComplianceCondition("sox_environment"),
        },
        
        Obligations: []abac.Obligation{
            {
                Type: "notification",
                Config: map[string]interface{}{
                    "notify_roles": []string{"cfo", "auditor"},
                    "priority":     "high",
                    "immediate":    true,
                },
            },
            {
                Type: "approval_required",
                Config: map[string]interface{}{
                    "approval_type": "dual_authorization",
                    "approver_roles": []string{"cfo", "controller"},
                    "timeout_hours": 24,
                },
            },
        },
    },
}
```

#### **Transaction Authorization Policies**
```go
// Transaction policies with dynamic risk assessment
var TransactionAuthorizationPolicies = []abac.Policy{
    {
        ID:   "transaction-post-risk-based",
        Name: "Risk-Based Transaction Posting",
        Description: "Transaction posting authorization based on risk assessment",
        Effect: abac.EffectAllow,
        Version: "1.0",
        
        Subjects: []abac.Condition{
            abac.OR(
                // Low risk: accountants can post small transactions
                abac.AND(
                    abac.RoleCondition("accountant"),
                    abac.AmountCondition("<=", 5000),
                    abac.RiskCondition("<=", "low"),
                ),
                // Medium risk: senior accountants
                abac.AND(
                    abac.RoleCondition("senior_accountant", "finance_manager"),
                    abac.AmountCondition("<=", 50000),
                    abac.RiskCondition("<=", "medium"),
                ),
                // High risk: only senior management
                abac.AND(
                    abac.RoleCondition("controller", "cfo"),
                    abac.RiskCondition("<=", "high"),
                    abac.ApprovalCondition("executive_approval"),
                ),
            ),
        },
        
        Resources: []abac.Condition{
            abac.ResourceTypeCondition("transaction"),
        },
        
        Actions: []abac.Condition{
            abac.ActionCondition("post"),
        },
        
        Context: []abac.Condition{
            abac.TenantCondition(),
            abac.TimeCondition("business_hours"),
            abac.NOT(abac.UserCondition("same_as_creator")), // Segregation of duties
            abac.ComplianceCondition("internal_controls_active"),
        },
        
        Obligations: []abac.Obligation{
            {
                Type: "segregation_of_duties",
                Config: map[string]interface{}{
                    "require_different_user": true,
                    "minimum_separation_hours": 1,
                },
            },
            {
                Type: "dual_authorization",
                Config: map[string]interface{}{
                    "required_for_amounts_over": 25000,
                    "timeout_hours": 48,
                },
            },
        },
    },
}
```

### **Dynamic Risk Assessment**

#### **Risk Context Provider**
```go
// @internal/core/finance/abac/risk_context_provider.go

type FinancialRiskContextProvider struct {
    fraudDetection FraudDetectionService
    userAnalytics  UserAnalyticsService
    transactionRepo TransactionRepository
}

func (p *FinancialRiskContextProvider) EnrichContext(ctx context.Context, request abac.Request) (map[string]interface{}, error) {
    enrichedContext := make(map[string]interface{})
    
    // Basic context from request
    for k, v := range request.Context {
        enrichedContext[k] = v
    }
    
    // Add risk assessment if transaction-related
    if request.Resource.Type == "transaction" {
        riskAssessment, err := p.assessTransactionRisk(ctx, request)
        if err != nil {
            return enrichedContext, err
        }
        
        enrichedContext["risk_level"] = riskAssessment.RiskLevel
        enrichedContext["risk_score"] = riskAssessment.RiskScore
        enrichedContext["risk_factors"] = riskAssessment.RiskFactors
    }
    
    // Add user behavior context
    if userID, ok := request.Subject.Attributes["user_id"].(string); ok {
        behaviorProfile, err := p.userAnalytics.GetBehaviorProfile(ctx, userID)
        if err == nil {
            enrichedContext["user_behavior_score"] = behaviorProfile.TrustScore
            enrichedContext["recent_anomalies"] = behaviorProfile.RecentAnomalies
            enrichedContext["access_patterns"] = behaviorProfile.AccessPatterns
        }
    }
    
    // Add temporal context
    now := time.Now()
    enrichedContext["hour_of_day"] = now.Hour()
    enrichedContext["day_of_week"] = now.Weekday().String()
    enrichedContext["is_business_hours"] = p.isBusinessHours(now)
    enrichedContext["is_weekend"] = now.Weekday() == time.Saturday || now.Weekday() == time.Sunday
    
    // Add compliance context
    enrichedContext["sox_period"] = p.isSOXReportingPeriod(now)
    enrichedContext["audit_period"] = p.isAuditPeriod(now)
    enrichedContext["fiscal_year_end"] = p.isFiscalYearEnd(now)
    
    return enrichedContext, nil
}

func (p *FinancialRiskContextProvider) assessTransactionRisk(ctx context.Context, request abac.Request) (*RiskAssessment, error) {
    // Extract transaction details from request
    amount, _ := request.Resource.Attributes["amount"].(decimal.Decimal)
    transactionType, _ := request.Resource.Attributes["type"].(string)
    
    // Create mock transaction for risk assessment
    mockTransaction := &Transaction{
        TotalAmount: amount,
        Type:        TransactionType(transactionType),
        CreatedAt:   time.Now(),
        CreatedBy:   identity.UserID(request.Subject.ID),
    }
    
    return p.fraudDetection.EvaluateTransaction(ctx, mockTransaction)
}
```

### **Policy Evaluation Engine**

#### **Custom Financial Condition Evaluators**
```go
// @internal/core/finance/abac/condition_evaluators.go

// Amount condition evaluator for financial policies
func AmountConditionEvaluator(condition abac.Condition, context abac.EvaluationContext) (bool, error) {
    operator, ok := condition.Config["operator"].(string)
    if !ok {
        return false, fmt.Errorf("amount condition missing operator")
    }
    
    thresholdStr, ok := condition.Config["threshold"].(string)
    if !ok {
        return false, fmt.Errorf("amount condition missing threshold")
    }
    
    threshold, err := decimal.NewFromString(thresholdStr)
    if err != nil {
        return false, fmt.Errorf("invalid threshold amount: %w", err)
    }
    
    // Get amount from resource attributes
    amountAttr, exists := context.Resource.Attributes["amount"]
    if !exists {
        return false, fmt.Errorf("amount not found in resource attributes")
    }
    
    amount, ok := amountAttr.(decimal.Decimal)
    if !ok {
        // Try to parse from string or float
        if amountStr, ok := amountAttr.(string); ok {
            amount, err = decimal.NewFromString(amountStr)
            if err != nil {
                return false, fmt.Errorf("invalid amount format: %w", err)
            }
        } else if amountFloat, ok := amountAttr.(float64); ok {
            amount = decimal.NewFromFloat(amountFloat)
        } else {
            return false, fmt.Errorf("unsupported amount type")
        }
    }
    
    // Evaluate condition
    switch operator {
    case "<=":
        return amount.LessThanOrEqual(threshold), nil
    case "<":
        return amount.LessThan(threshold), nil
    case ">=":
        return amount.GreaterThanOrEqual(threshold), nil
    case ">":
        return amount.GreaterThan(threshold), nil
    case "=":
        return amount.Equal(threshold), nil
    case "!=":
        return !amount.Equal(threshold), nil
    default:
        return false, fmt.Errorf("unsupported operator: %s", operator)
    }
}

// Risk level condition evaluator
func RiskConditionEvaluator(condition abac.Condition, context abac.EvaluationContext) (bool, error) {
    operator, ok := condition.Config["operator"].(string)
    if !ok {
        return false, fmt.Errorf("risk condition missing operator")
    }
    
    thresholdStr, ok := condition.Config["threshold"].(string)
    if !ok {
        return false, fmt.Errorf("risk condition missing threshold")
    }
    
    // Get risk level from context
    riskLevelAttr, exists := context.Environment["risk_level"]
    if !exists {
        return false, fmt.Errorf("risk level not found in context")
    }
    
    riskLevel, ok := riskLevelAttr.(string)
    if !ok {
        return false, fmt.Errorf("invalid risk level type")
    }
    
    // Convert risk levels to numeric values for comparison
    riskValues := map[string]int{
        "low":      1,
        "medium":   2,
        "high":     3,
        "critical": 4,
    }
    
    currentRisk, ok := riskValues[riskLevel]
    if !ok {
        return false, fmt.Errorf("unknown risk level: %s", riskLevel)
    }
    
    thresholdRisk, ok := riskValues[thresholdStr]
    if !ok {
        return false, fmt.Errorf("unknown threshold risk level: %s", thresholdStr)
    }
    
    // Evaluate condition
    switch operator {
    case "<=":
        return currentRisk <= thresholdRisk, nil
    case "<":
        return currentRisk < thresholdRisk, nil
    case ">=":
        return currentRisk >= thresholdRisk, nil
    case ">":
        return currentRisk > thresholdRisk, nil
    case "=":
        return currentRisk == thresholdRisk, nil
    case "!=":
        return currentRisk != thresholdRisk, nil
    default:
        return false, fmt.Errorf("unsupported operator: %s", operator)
    }
}

// Segregation of duties evaluator
func SegregationOfDutiesEvaluator(condition abac.Condition, context abac.EvaluationContext) (bool, error) {
    requireDifferentUser, ok := condition.Config["require_different_user"].(bool)
    if !ok || !requireDifferentUser {
        return true, nil // No segregation requirement
    }
    
    // Get current user ID
    currentUserID, ok := context.Subject.Attributes["user_id"].(string)
    if !ok {
        return false, fmt.Errorf("user ID not found in subject")
    }
    
    // Get resource creator ID
    creatorID, exists := context.Resource.Attributes["created_by"]
    if !exists {
        return true, nil // No creator information, allow
    }
    
    creatorIDStr, ok := creatorID.(string)
    if !ok {
        return false, fmt.Errorf("invalid creator ID type")
    }
    
    // Users must be different
    return currentUserID != creatorIDStr, nil
}

// Business hours evaluator
func BusinessHoursEvaluator(condition abac.Condition, context abac.EvaluationContext) (bool, error) {
    requiredCondition, ok := condition.Config["required"].(string)
    if !ok {
        requiredCondition = "business_hours"
    }
    
    // Get current time
    now := time.Now()
    
    switch requiredCondition {
    case "business_hours":
        return isBusinessHours(now), nil
    case "after_hours":
        return !isBusinessHours(now), nil
    case "weekend":
        return now.Weekday() == time.Saturday || now.Weekday() == time.Sunday, nil
    case "weekday":
        return now.Weekday() != time.Saturday && now.Weekday() != time.Sunday, nil
    default:
        return false, fmt.Errorf("unknown time condition: %s", requiredCondition)
    }
}

func isBusinessHours(t time.Time) bool {
    hour := t.Hour()
    weekday := t.Weekday()
    
    // Monday to Friday, 8 AM to 6 PM
    return weekday >= time.Monday && weekday <= time.Friday && hour >= 8 && hour < 18
}
```

---

## Audit & Compliance Framework

### **Comprehensive Audit Logging**

#### **Financial Audit Service**
```go
// @internal/core/finance/audit/financial_audit_service.go

type FinancialAuditService struct {
    baseAudit    audit.Service
    storage      AuditStorageService
    compliance   ComplianceService
    alerting     AlertingService
    encryption   EncryptionService
}

func (s *FinancialAuditService) LogFinancialEvent(ctx context.Context, event FinancialAuditEvent) error {
    // Enrich event with compliance context
    enrichedEvent := s.enrichEventForCompliance(event)
    
    // Encrypt sensitive financial data
    if err := s.encryptSensitiveData(&enrichedEvent); err != nil {
        return fmt.Errorf("failed to encrypt audit data: %w", err)
    }
    
    // Store in immutable audit log
    if err := s.storage.Store(ctx, enrichedEvent); err != nil {
        return fmt.Errorf("failed to store audit event: %w", err)
    }
    
    // Check for compliance violations
    if violations := s.compliance.CheckViolations(enrichedEvent); len(violations) > 0 {
        s.handleComplianceViolations(ctx, enrichedEvent, violations)
    }
    
    // Real-time monitoring for high-risk events
    if enrichedEvent.RiskScore >= audit.RiskHigh {
        go s.triggerRealTimeAlert(ctx, enrichedEvent)
    }
    
    return nil
}

type FinancialAuditEvent struct {
    // Core audit fields
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`
    TenantID     string                 `json:"tenant_id"`
    UserID       string                 `json:"user_id"`
    ResourceID   string                 `json:"resource_id"`
    ResourceType string                 `json:"resource_type"`
    Action       string                 `json:"action"`
    Timestamp    time.Time              `json:"timestamp"`
    
    // Financial-specific fields
    Amount       *decimal.Decimal       `json:"amount,omitempty"`
    Currency     string                 `json:"currency,omitempty"`
    AccountCodes []string               `json:"account_codes,omitempty"`
    
    // Compliance fields
    ComplianceFrameworks []string       `json:"compliance_frameworks"`
    RetentionPeriod      time.Duration  `json:"retention_period"`
    DataClassification   string         `json:"data_classification"`
    
    // Risk assessment
    RiskScore    audit.RiskLevel        `json:"risk_score"`
    RiskFactors  []RiskFactor          `json:"risk_factors,omitempty"`
    
    // Context and metadata
    SessionID    string                 `json:"session_id,omitempty"`
    IPAddress    string                 `json:"ip_address,omitempty"`
    UserAgent    string                 `json:"user_agent,omitempty"`
    Location     *GeographicLocation    `json:"location,omitempty"`
    
    // Audit trail integrity
    PreviousEventHash string           `json:"previous_event_hash"`
    EventHash         string           `json:"event_hash"`
    DigitalSignature  string           `json:"digital_signature"`
    
    // Sensitive data (encrypted)
    EncryptedData map[string]interface{} `json:"encrypted_data,omitempty"`
}

func (s *FinancialAuditService) LogAccountOperation(ctx context.Context, operation AccountOperation) error {
    event := FinancialAuditEvent{
        ID:           uuid.New().String(),
        Type:         fmt.Sprintf("finance.account.%s", operation.Type),
        TenantID:     operation.TenantID.String(),
        UserID:       operation.UserID.String(),
        ResourceID:   operation.AccountID.String(),
        ResourceType: "account",
        Action:       operation.Type,
        Timestamp:    time.Now(),
        
        AccountCodes:         []string{string(operation.AccountCode)},
        ComplianceFrameworks: []string{"SOX", "GAAP", "IFRS"},
        RetentionPeriod:      7 * 365 * 24 * time.Hour, // 7 years for SOX
        DataClassification:   "financial_confidential",
        RiskScore:           s.assessAccountOperationRisk(operation),
        
        SessionID: extractSessionID(ctx),
        IPAddress: extractIPAddress(ctx),
        UserAgent: extractUserAgent(ctx),
        Location:  extractLocation(ctx),
    }
    
    // Add operation-specific data
    event.EncryptedData = map[string]interface{}{
        "account_name":        operation.AccountName,
        "account_type":        operation.AccountType,
        "previous_values":     operation.PreviousValues,
        "new_values":          operation.NewValues,
        "business_justification": operation.Justification,
    }
    
    return s.LogFinancialEvent(ctx, event)
}

func (s *FinancialAuditService) LogTransactionEvent(ctx context.Context, transaction *Transaction, eventType string) error {
    event := FinancialAuditEvent{
        ID:           uuid.New().String(),
        Type:         fmt.Sprintf("finance.transaction.%s", eventType),
        TenantID:     transaction.TenantID.String(),
        UserID:       transaction.CreatedBy.String(),
        ResourceID:   transaction.ID.String(),
        ResourceType: "transaction",
        Action:       eventType,
        Timestamp:    time.Now(),
        
        Amount:               &transaction.TotalAmount,
        Currency:             string(transaction.Currency),
        ComplianceFrameworks: []string{"SOX", "GAAP", "SARBANES_OXLEY"},
        RetentionPeriod:      7 * 365 * 24 * time.Hour,
        DataClassification:   "financial_confidential",
        RiskScore:           s.assessTransactionRisk(transaction),
        
        SessionID: extractSessionID(ctx),
        IPAddress: extractIPAddress(ctx),
        UserAgent: extractUserAgent(ctx),
    }
    
    // Extract account codes from entries
    accountCodes := make([]string, len(transaction.Entries))
    for i, entry := range transaction.Entries {
        accountCodes[i] = string(entry.AccountID)
    }
    event.AccountCodes = accountCodes
    
    // Encrypt transaction details
    event.EncryptedData = map[string]interface{}{
        "transaction_number": transaction.Number,
        "transaction_type":   transaction.Type,
        "posting_date":       transaction.PostingDate,
        "due_date":           transaction.DueDate,
        "exchange_rate":      transaction.ExchangeRate,
        "reference_number":   transaction.ReferenceNumber,
        "description":        transaction.Description,
        "entries":            s.encryptTransactionEntries(transaction.Entries),
    }
    
    // Add segregation of duties information
    if eventType == "posted" && transaction.ApprovedBy != nil {
        event.EncryptedData["posted_by"] = transaction.ApprovedBy.String()
        event.EncryptedData["segregation_of_duties"] = transaction.CreatedBy.String() != transaction.ApprovedBy.String()
    }
    
    return s.LogFinancialEvent(ctx, event)
}

func (s *FinancialAuditService) enrichEventForCompliance(event FinancialAuditEvent) FinancialAuditEvent {
    // Add compliance-specific enrichment
    event.PreviousEventHash = s.getLastEventHash(event.TenantID)
    event.EventHash = s.calculateEventHash(event)
    event.DigitalSignature = s.signEvent(event)
    
    // Enrich with regulatory context
    if s.isSOXApplicable(event.TenantID) {
        event.ComplianceFrameworks = append(event.ComplianceFrameworks, "SOX")
    }
    
    if s.isGDPRApplicable(event.TenantID) {
        event.ComplianceFrameworks = append(event.ComplianceFrameworks, "GDPR")
    }
    
    return event
}
```

### **Compliance Monitoring & Reporting**

#### **SOX Compliance Service**
```go
// @internal/core/finance/compliance/sox_compliance.go

type SOXComplianceService struct {
    audit           audit.Service
    policyEngine    PolicyEngine
    controlTesting  ControlTestingService
    reportGenerator ReportGenerator
}

// SOX Section 404 - Internal Controls Assessment
func (s *SOXComplianceService) AssessInternalControls(ctx context.Context, tenantID tenant.ID, period Period) (*SOXAssessment, error) {
    assessment := &SOXAssessment{
        TenantID:      tenantID,
        Period:        period,
        AssessmentDate: time.Now(),
        Controls:      []ControlAssessment{},
        Deficiencies:  []ControlDeficiency{},
        Status:        "in_progress",
    }
    
    // Assess key financial controls
    controls := []string{
        "segregation_of_duties",
        "authorization_limits",
        "transaction_approval_workflow",
        "account_reconciliation",
        "journal_entry_controls",
        "financial_reporting_controls",
    }
    
    for _, controlName := range controls {
        controlAssessment, err := s.assessControl(ctx, tenantID, controlName, period)
        if err != nil {
            return nil, fmt.Errorf("failed to assess control %s: %w", controlName, err)
        }
        
        assessment.Controls = append(assessment.Controls, *controlAssessment)
        
        if controlAssessment.Effectiveness != "effective" {
            assessment.Deficiencies = append(assessment.Deficiencies, ControlDeficiency{
                ControlName:     controlName,
                DeficiencyType:  controlAssessment.DeficiencyType,
                Severity:        controlAssessment.Severity,
                Description:     controlAssessment.Description,
                RemediationPlan: controlAssessment.RemediationPlan,
            })
        }
    }
    
    // Determine overall assessment
    assessment.Status = s.determineOverallAssessment(assessment.Controls)
    
    return assessment, nil
}

func (s *SOXComplianceService) assessControl(ctx context.Context, tenantID tenant.ID, controlName string, period Period) (*ControlAssessment, error) {
    switch controlName {
    case "segregation_of_duties":
        return s.assessSegregationOfDuties(ctx, tenantID, period)
    case "authorization_limits":
        return s.assessAuthorizationLimits(ctx, tenantID, period)
    case "transaction_approval_workflow":
        return s.assessApprovalWorkflows(ctx, tenantID, period)
    default:
        return nil, fmt.Errorf("unknown control: %s", controlName)
    }
}

func (s *SOXComplianceService) assessSegregationOfDuties(ctx context.Context, tenantID tenant.ID, period Period) (*ControlAssessment, error) {
    assessment := &ControlAssessment{
        ControlName:   "segregation_of_duties",
        TestDate:      time.Now(),
        Effectiveness: "effective",
    }
    
    // Query transactions where creator and approver are the same
    violationQuery := `
        SELECT COUNT(*) as violation_count
        FROM finance_transactions 
        WHERE tenant_id = $1 
        AND posting_date BETWEEN $2 AND $3
        AND created_by = approved_by
        AND status = 'posted'
        AND total_amount > 1000 -- Significant amounts
    `
    
    var violationCount int
    err := s.audit.QueryDB(ctx, violationQuery, &violationCount, tenantID, period.StartDate, period.EndDate)
    if err != nil {
        return nil, err
    }
    
    if violationCount > 0 {
        assessment.Effectiveness = "ineffective"
        assessment.DeficiencyType = "material_weakness"
        assessment.Severity = "high"
        assessment.Description = fmt.Sprintf("Found %d transactions where the same person created and approved the transaction", violationCount)
        assessment.RemediationPlan = "Implement automated controls to prevent same-person approval for transactions over $1,000"
    }
    
    // Test authorization matrix compliance
    matrixViolations, err := s.testAuthorizationMatrix(ctx, tenantID, period)
    if err != nil {
        return nil, err
    }
    
    if len(matrixViolations) > 0 {
        if assessment.Effectiveness == "effective" {
            assessment.Effectiveness = "deficient"
            assessment.DeficiencyType = "significant_deficiency"
            assessment.Severity = "medium"
        }
        assessment.Description += fmt.Sprintf("; Found %d authorization matrix violations", len(matrixViolations))
    }
    
    return assessment, nil
}

// GDPR Compliance for Financial Data
func (s *SOXComplianceService) AssessGDPRCompliance(ctx context.Context, tenantID tenant.ID) (*GDPRAssessment, error) {
    assessment := &GDPRAssessment{
        TenantID:       tenantID,
        AssessmentDate: time.Now(),
        DataCategories: []DataCategoryAssessment{},
        Rights:         []DataSubjectRightAssessment{},
    }
    
    // Assess financial data categories
    categories := []string{
        "financial_transactions",
        "account_balances",
        "payment_information",
        "bank_account_details",
        "credit_information",
    }
    
    for _, category := range categories {
        categoryAssessment := s.assessDataCategory(ctx, tenantID, category)
        assessment.DataCategories = append(assessment.DataCategories, categoryAssessment)
    }
    
    // Assess data subject rights implementation
    rights := []string{
        "right_to_access",
        "right_to_rectification",
        "right_to_erasure",
        "right_to_portability",
        "right_to_restrict_processing",
    }
    
    for _, right := range rights {
        rightAssessment := s.assessDataSubjectRight(ctx, tenantID, right)
        assessment.Rights = append(assessment.Rights, rightAssessment)
    }
    
    return assessment, nil
}
```

### **Audit Trail Integrity**

#### **Immutable Audit Chain**
```go
// @internal/core/finance/audit/audit_chain.go

type AuditChain struct {
    storage    AuditStorageService
    crypto     CryptographicService
    blockchain BlockchainService // Optional: for ultimate immutability
}

type AuditBlock struct {
    Index         uint64                 `json:"index"`
    TenantID      string                 `json:"tenant_id"`
    Timestamp     time.Time              `json:"timestamp"`
    Events        []FinancialAuditEvent  `json:"events"`
    PreviousHash  string                 `json:"previous_hash"`
    MerkleRoot    string                 `json:"merkle_root"`
    Hash          string                 `json:"hash"`
    Signature     string                 `json:"signature"`
}

func (c *AuditChain) AddEvent(ctx context.Context, event FinancialAuditEvent) error {
    // Get the latest block for this tenant
    latestBlock, err := c.storage.GetLatestBlock(ctx, event.TenantID)
    if err != nil && !errors.IsNotFound(err) {
        return fmt.Errorf("failed to get latest block: %w", err)
    }
    
    var newBlock *AuditBlock
    
    if latestBlock == nil {
        // Genesis block for this tenant
        newBlock = &AuditBlock{
            Index:        0,
            TenantID:     event.TenantID,
            Timestamp:    time.Now(),
            Events:       []FinancialAuditEvent{event},
            PreviousHash: "0", // Genesis block
        }
    } else {
        // Check if we should create a new block (time-based or event count)
        if c.shouldCreateNewBlock(latestBlock) {
            newBlock = &AuditBlock{
                Index:        latestBlock.Index + 1,
                TenantID:     event.TenantID,
                Timestamp:    time.Now(),
                Events:       []FinancialAuditEvent{event},
                PreviousHash: latestBlock.Hash,
            }
        } else {
            // Add to existing block
            latestBlock.Events = append(latestBlock.Events, event)
            newBlock = latestBlock
        }
    }
    
    // Calculate Merkle root of all events
    newBlock.MerkleRoot = c.calculateMerkleRoot(newBlock.Events)
    
    // Calculate block hash
    newBlock.Hash = c.calculateBlockHash(newBlock)
    
    // Sign the block
    newBlock.Signature = c.crypto.SignBlock(newBlock)
    
    // Store the block
    if err := c.storage.StoreBlock(ctx, newBlock); err != nil {
        return fmt.Errorf("failed to store audit block: %w", err)
    }
    
    // Optionally commit to blockchain for ultimate immutability
    if c.blockchain != nil {
        go func() {
            if err := c.blockchain.CommitBlock(context.Background(), newBlock); err != nil {
                // Log error but don't fail the operation
                log.Error("Failed to commit block to blockchain", "error", err, "block_hash", newBlock.Hash)
            }
        }()
    }
    
    return nil
}

func (c *AuditChain) VerifyIntegrity(ctx context.Context, tenantID string, fromTime, toTime time.Time) (*IntegrityReport, error) {
    blocks, err := c.storage.GetBlocksInRange(ctx, tenantID, fromTime, toTime)
    if err != nil {
        return nil, fmt.Errorf("failed to get blocks: %w", err)
    }
    
    report := &IntegrityReport{
        TenantID:      tenantID,
        StartTime:     fromTime,
        EndTime:       toTime,
        TotalBlocks:   len(blocks),
        TotalEvents:   0,
        Violations:    []IntegrityViolation{},
        IsValid:       true,
    }
    
    for i, block := range blocks {
        // Count events
        report.TotalEvents += len(block.Events)
        
        // Verify block hash
        calculatedHash := c.calculateBlockHash(block)
        if calculatedHash != block.Hash {
            report.IsValid = false
            report.Violations = append(report.Violations, IntegrityViolation{
                Type:        "hash_mismatch",
                BlockIndex:  block.Index,
                Description: fmt.Sprintf("Block hash mismatch: expected %s, got %s", calculatedHash, block.Hash),
            })
        }
        
        // Verify signature
        if !c.crypto.VerifyBlockSignature(block) {
            report.IsValid = false
            report.Violations = append(report.Violations, IntegrityViolation{
                Type:        "signature_invalid",
                BlockIndex:  block.Index,
                Description: "Block signature verification failed",
            })
        }
        
        // Verify chain integrity
        if i > 0 && block.PreviousHash != blocks[i-1].Hash {
            report.IsValid = false
            report.Violations = append(report.Violations, IntegrityViolation{
                Type:        "chain_broken",
                BlockIndex:  block.Index,
                Description: "Previous hash doesn't match preceding block",
            })
        }
        
        // Verify Merkle root
        calculatedMerkleRoot := c.calculateMerkleRoot(block.Events)
        if calculatedMerkleRoot != block.MerkleRoot {
            report.IsValid = false
            report.Violations = append(report.Violations, IntegrityViolation{
                Type:        "merkle_mismatch",
                BlockIndex:  block.Index,
                Description: "Merkle root verification failed",
            })
        }
    }
    
    return report, nil
}

func (c *AuditChain) calculateMerkleRoot(events []FinancialAuditEvent) string {
    if len(events) == 0 {
        return ""
    }
    
    // Create leaf hashes
    hashes := make([]string, len(events))
    for i, event := range events {
        hashes[i] = c.crypto.HashEvent(event)
    }
    
    // Build Merkle tree
    for len(hashes) > 1 {
        var newLevel []string
        for i := 0; i < len(hashes); i += 2 {
            if i+1 < len(hashes) {
                combined := hashes[i] + hashes[i+1]
                newLevel = append(newLevel, c.crypto.Hash(combined))
            } else {
                newLevel = append(newLevel, hashes[i])
            }
        }
        hashes = newLevel
    }
    
    return hashes[0]
}
```

This security and compliance guide demonstrates the sophisticated approach to protecting financial data in the AWO ERP system. The framework ensures that all financial operations are properly authorized, ly audited, and compliant with major regulatory requirements while maintaining the highest levels of data integrity and security.