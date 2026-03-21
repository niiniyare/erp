package authz

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo/internal/core/abac/models"
	"awo/internal/core/access"
	"awo/internal/core/iam/model"
	"awo/internal/shared/types"
)

// ─── POLICY DECISION TYPE CONVERSIONS ─────────────────────────────────────

// convertPolicyDecisionType converts shared types.PolicyDecisionType to IAM model.PolicyDecisionType
func convertPolicyDecisionType(decision types.PolicyDecisionType) model.PolicyDecisionType {
	switch decision {
	case types.PolicyDecisionAllow:
		return model.PolicyDecisionAllow
	case types.PolicyDecisionDeny:
		return model.PolicyDecisionDeny
	case types.PolicyDecisionNotApplicable:
		return model.PolicyDecisionNotApplicable
	default:
		return model.PolicyDecisionDeny // Default to deny for safety
	}
}

// convertPolicyDecisionTypeToShared converts IAM model.PolicyDecisionType to shared types.PolicyDecisionType
func convertPolicyDecisionTypeToShared(decision model.PolicyDecisionType) types.PolicyDecisionType {
	switch decision {
	case model.PolicyDecisionAllow:
		return types.PolicyDecisionAllow
	case model.PolicyDecisionDeny:
		return types.PolicyDecisionDeny
	case model.PolicyDecisionNotApplicable:
		return types.PolicyDecisionNotApplicable
	default:
		return types.PolicyDecisionDeny // Default to deny for safety
	}
}

// ─── POLICY EFFECT CONVERSIONS ─────────────────────────────────────────────

// convertPolicyEffectType converts shared types.PolicyEffect to IAM model.PolicyEffect
func convertPolicyEffectType(effect types.PolicyEffect) model.PolicyEffect {
	switch effect {
	case types.PolicyEffectAllow:
		return model.PolicyEffectAllow
	case types.PolicyEffectDeny:
		return model.PolicyEffectDeny
	default:
		return model.PolicyEffectDeny // Default to deny for safety
	}
}

// convertPolicyEffectTypeToShared converts IAM model.PolicyEffect to shared types.PolicyEffect
func convertPolicyEffectTypeToShared(effect model.PolicyEffect) types.PolicyEffect {
	switch effect {
	case model.PolicyEffectAllow:
		return types.PolicyEffectAllow
	case model.PolicyEffectDeny:
		return types.PolicyEffectDeny
	default:
		return types.PolicyEffectDeny // Default to deny for safety
	}
}

// ─── POLICY DECISION CONVERSIONS ───────────────────────────────────────────

// convertPolicyDecisions converts ABAC policy decisions to IAM policy decisions
func convertPolicyDecisions(abacDecisions []*models.PolicyDecision) []*model.PolicyDecision {
	if abacDecisions == nil {
		return nil
	}

	iamDecisions := make([]*model.PolicyDecision, len(abacDecisions))
	for i, abacDecision := range abacDecisions {
		iamDecisions[i] = &model.PolicyDecision{
			PolicyID: abacDecision.PolicyID,
			Decision: convertPolicyDecisionType(abacDecision.Decision),
			Effect:   model.PolicyEffectAllow, // Default since ABAC PolicyDecision doesn't have Effect field
			Reason:   abacDecision.Reason,
		}
	}

	return iamDecisions
}

// ─── ACCESS REQUEST CONVERSIONS ────────────────────────────────────────────

// convertAccessRequestToIAMModel converts access service model to IAM model
func convertAccessRequestToIAMModel(accessReq *access.AccessRequest) *model.AccessRequest {
	if accessReq == nil {
		return nil
	}

	return &model.AccessRequest{
		ID:               accessReq.ID,
		TenantID:         accessReq.TenantID,
		RequesterID:      accessReq.RequesterID,
		TargetUserID:     accessReq.TargetUserID,
		EntityID:         &accessReq.EntityID,
		RequestType:      model.RequestType(string(accessReq.RequestType)),
		ResourceType:     inferResourceType(accessReq.RequestType),
		ResourceID:       accessReq.ResourceID,
		Action:           inferAction(accessReq.RequestType),
		Justification:    accessReq.Justification,
		Priority:         "medium", // Default priority
		ApprovalStatus:   convertApprovalStatus(string(accessReq.ApprovalStatus)),
		ApprovedBy:       accessReq.ApprovedBy,
		ApprovedAt:       accessReq.ApprovedAt,
		ApprovalComments: accessReq.ApprovalComments,
		ExpiresAt:        accessReq.ExpiresAt,
		AutoRevoke:       false, // Default value
		Metadata:         map[string]any{},
		CreatedAt:        accessReq.CreatedAt,
		UpdatedAt:        accessReq.UpdatedAt,
	}
}

// Helper functions to extract information from AccessRequest
func getTargetUserID(accessReq *access.AccessRequest) uuid.UUID {
	if accessReq.TargetUserID != nil {
		return *accessReq.TargetUserID
	}
	return accessReq.RequesterID
}

func inferResourceType(requestType access.RequestType) string {
	switch requestType {
	case access.RequestType("ROLE_ASSIGNMENT"):
		return "role"
	case access.RequestType("PERMISSION_GRANT"):
		return "permission"
	case access.RequestType("RESOURCE_ACCESS"):
		return "resource"
	case access.RequestType("ELEVATION"):
		return "privilege"
	default:
		return "resource"
	}
}

func inferAction(requestType access.RequestType) string {
	switch requestType {
	case access.RequestType("ROLE_ASSIGNMENT"):
		return "assign"
	case access.RequestType("PERMISSION_GRANT"):
		return "grant"
	case access.RequestType("RESOURCE_ACCESS"):
		return "access"
	case access.RequestType("ELEVATION"):
		return "elevate"
	default:
		return "access"
	}
}

// convertApprovalStatus converts access service status to IAM model status
func convertApprovalStatus(status string) model.ApprovalStatus {
	switch status {
	case "PENDING":
		return model.ApprovalStatusPending
	case "APPROVED":
		return model.ApprovalStatusApproved
	case "REJECTED":
		return model.ApprovalStatusRejected
	case "EXPIRED":
		return model.ApprovalStatusExpired
	case "REVOKED":
		return model.ApprovalStatusRevoked
	default:
		return model.ApprovalStatusPending // Default to pending
	}
}

// ─── APPROVAL WORKFLOW CONVERSIONS ─────────────────────────────────────────

// convertApprovalWorkflowToIAMModel converts access service model to IAM model
func convertApprovalWorkflowToIAMModel(accessWorkflow *access.ApprovalWorkflow) *model.ApprovalWorkflow {
	if accessWorkflow == nil {
		return nil
	}

	return &model.ApprovalWorkflow{
		ID:          accessWorkflow.ID,
		TenantID:    accessWorkflow.TenantID,
		Name:        accessWorkflow.Name,
		Description: accessWorkflow.Description,
		Steps:       convertApprovalStepsToIAM(accessWorkflow.Steps),
		Enabled:     accessWorkflow.IsActive,
		Metadata:    accessWorkflow.Metadata,
		CreatedAt:   accessWorkflow.CreatedAt,
		UpdatedAt:   accessWorkflow.UpdatedAt,
	}
}

// convertApprovalStepsToIAM converts access service steps to IAM model steps
func convertApprovalStepsToIAM(accessSteps []*access.ApprovalStep) []*model.ApprovalStep {
	if accessSteps == nil {
		return nil
	}

	iamSteps := make([]*model.ApprovalStep, len(accessSteps))
	for i, accessStep := range accessSteps {
		iamSteps[i] = &model.ApprovalStep{
			ID:            accessStep.ID.String(),
			Name:          accessStep.Name,
			Order:         accessStep.Order,
			RequiredVotes: accessStep.RequiredCount,
			ApproverUsers: accessStep.ApproverIDs,
			TimeoutHours:  &accessStep.TimeoutMinutes, // Convert minutes to hours - simplified
			AutoApprove:   false,                      // Default value
		}
	}

	return iamSteps
}

// convertApprovalStepsFromIAM converts IAM model steps to access service steps
func convertApprovalStepsFromIAM(iamSteps []*model.ApprovalStep) []*access.ApprovalStep {
	if iamSteps == nil {
		return nil
	}

	accessSteps := make([]*access.ApprovalStep, len(iamSteps))
	for i, iamStep := range iamSteps {
		stepID, _ := uuid.Parse(iamStep.ID) // Parse string ID to UUID
		timeoutMinutes := 60                // Default timeout
		if iamStep.TimeoutHours != nil {
			timeoutMinutes = *iamStep.TimeoutHours * 60
		}

		accessSteps[i] = &access.ApprovalStep{
			ID:             stepID,
			Order:          iamStep.Order,
			Name:           iamStep.Name,
			Description:    "",     // Not available in IAM model
			ApproverType:   "user", // Default type
			ApproverIDs:    iamStep.ApproverUsers,
			RequiredCount:  iamStep.RequiredVotes,
			TimeoutMinutes: timeoutMinutes,
			Conditions:     []*access.ApprovalCondition{}, // Empty for now
			Metadata:       map[string]any{},              // Empty for now
		}
	}

	return accessSteps
}

// convertApprovalConditionsToIAM converts access service conditions to IAM model conditions
func convertApprovalConditionsToIAM(accessConditions []*access.ApprovalCondition) []*ApprovalCondition {
	if accessConditions == nil {
		return nil
	}

	iamConditions := make([]*ApprovalCondition, len(accessConditions))
	for i, accessCondition := range accessConditions {
		iamConditions[i] = &ApprovalCondition{
			Type:        accessCondition.Type,
			Operator:    accessCondition.Operator,
			Value:       accessCondition.Value,
			Description: accessCondition.Description,
		}
	}

	return iamConditions
}

// convertApprovalConditionsFromIAM converts IAM model conditions to access service conditions
func convertApprovalConditionsFromIAM(iamConditions []*ApprovalCondition) []*access.ApprovalCondition {
	if iamConditions == nil {
		return nil
	}

	accessConditions := make([]*access.ApprovalCondition, len(iamConditions))
	for i, iamCondition := range iamConditions {
		accessConditions[i] = &access.ApprovalCondition{
			Type:        iamCondition.Type,
			Operator:    iamCondition.Operator,
			Value:       iamCondition.Value,
			Description: iamCondition.Description,
		}
	}

	return accessConditions
}

// ApprovalCondition Local type definitions for missing types in IAM model
type ApprovalCondition struct {
	Type        string `json:"type"`
	Operator    string `json:"operator"`
	Value       any    `json:"value"`
	Description string `json:"description"`
}

// Use IAM model types directly
type (
	DeviceInfo      = model.DeviceContext
	LocationInfo    = model.GeolocationContext
	SecurityContext struct {
		ThreatLevel         string         `json:"threat_level"`
		AuthenticationLevel string         `json:"authentication_level"`
		EncryptionLevel     string         `json:"encryption_level"`
		SecurityFlags       []string       `json:"security_flags"`
		RiskScore           float64        `json:"risk_score"`
		Attributes          map[string]any `json:"attributes"`
	}
)

type TimeContext struct {
	RequestTime   time.Time      `json:"request_time"`
	TimeZone      string         `json:"time_zone"`
	BusinessHours *BusinessHours `json:"business_hours,omitempty"`
	IsWeekend     bool           `json:"is_weekend"`
	IsHoliday     bool           `json:"is_holiday"`
}

type BusinessHours struct {
	Start string   `json:"start"`
	End   string   `json:"end"`
	Days  []string `json:"days"`
}

// ─── CONDITIONAL ACCESS CONVERSIONS ────────────────────────────────────────

// convertAccessContextFromIAM converts IAM model context to access service context
func convertAccessContextFromIAM(iamContext *model.AccessContext) *access.AccessContext {
	if iamContext == nil {
		return nil
	}

	return &access.AccessContext{
		IPAddress:    iamContext.IPAddress,
		UserAgent:    iamContext.UserAgent,
		DeviceInfo:   convertDeviceInfoFromIAM(iamContext.Device),
		LocationInfo: convertLocationInfoFromIAM(iamContext.Location),
		SecurityContext: &access.SecurityContext{
			ThreatLevel:         iamContext.RiskLevel,
			AuthenticationLevel: "standard",
			EncryptionLevel:     "standard",
			SecurityFlags:       []string{},
			RiskScore:           0.0,
			Attributes:          iamContext.Attributes,
		},
		TimeContext: &access.TimeContext{
			RequestTime:   iamContext.Time,
			TimeZone:      "UTC",
			BusinessHours: nil,
			IsWeekend:     false,
			IsHoliday:     false,
		},
		Metadata: iamContext.Attributes,
	}
}

// convertDeviceInfoFromIAM converts device info from IAM to access service format
func convertDeviceInfoFromIAM(iamDevice *model.DeviceContext) *access.DeviceInfo {
	if iamDevice == nil {
		return nil
	}

	return &access.DeviceInfo{
		DeviceID:       iamDevice.DeviceID,
		DeviceType:     iamDevice.Platform,
		OS:             iamDevice.Platform,
		OSVersion:      "", // Not available in DeviceContext
		Browser:        iamDevice.Browser,
		BrowserVersion: "", // Not available in DeviceContext
		IsTrusted:      iamDevice.IsTrusted,
		IsManaged:      iamDevice.IsManaged,
	}
}

// convertLocationInfoFromIAM converts location info from IAM to access service format
func convertLocationInfoFromIAM(iamLocation *model.GeolocationContext) *access.LocationInfo {
	if iamLocation == nil {
		return nil
	}

	return &access.LocationInfo{
		Country:    iamLocation.Country,
		Region:     iamLocation.Region,
		City:       iamLocation.City,
		Latitude:   &iamLocation.Latitude,
		Longitude:  &iamLocation.Longitude,
		TimeZone:   "",               // Not available in GeolocationContext
		IsTrusted:  false,            // Not available in GeolocationContext
		RiskLevel:  "",               // Not available in GeolocationContext
		Attributes: map[string]any{}, // Not available in GeolocationContext
	}
}

// convertSecurityContextFromIAM converts security context from IAM to access service format
func convertSecurityContextFromIAM(iamSecurity *SecurityContext) *access.SecurityContext {
	if iamSecurity == nil {
		return nil
	}

	return &access.SecurityContext{
		ThreatLevel:         iamSecurity.ThreatLevel,
		AuthenticationLevel: iamSecurity.AuthenticationLevel,
		EncryptionLevel:     iamSecurity.EncryptionLevel,
		SecurityFlags:       iamSecurity.SecurityFlags,
		RiskScore:           iamSecurity.RiskScore,
		Attributes:          iamSecurity.Attributes,
	}
}

// convertTimeContextFromIAM converts time context from IAM to access service format
func convertTimeContextFromIAM(iamTime *TimeContext) *access.TimeContext {
	if iamTime == nil {
		return nil
	}

	return &access.TimeContext{
		RequestTime:   iamTime.RequestTime,
		TimeZone:      iamTime.TimeZone,
		BusinessHours: convertBusinessHoursFromIAM(iamTime.BusinessHours),
		IsWeekend:     iamTime.IsWeekend,
		IsHoliday:     iamTime.IsHoliday,
	}
}

// convertBusinessHoursFromIAM converts business hours from IAM to access service format
func convertBusinessHoursFromIAM(iamHours *BusinessHours) *access.BusinessHours {
	if iamHours == nil {
		return nil
	}

	return &access.BusinessHours{
		Start: iamHours.Start,
		End:   iamHours.End,
		Days:  iamHours.Days,
	}
}

// convertAccessConditionsToIAM converts access service conditions to IAM model conditions
func convertAccessConditionsToIAM(accessConditions []*access.AccessCondition) []*model.AccessCondition {
	if accessConditions == nil {
		return nil
	}

	iamConditions := make([]*model.AccessCondition, len(accessConditions))
	for i, accessCondition := range accessConditions {
		iamConditions[i] = &model.AccessCondition{
			Type:        accessCondition.Type,
			Description: accessCondition.Description,
			Required:    accessCondition.IsMandatory,
			Parameters:  accessCondition.Metadata,
		}
	}

	return iamConditions
}

// ─── CONDITIONAL ACCESS POLICY CONVERSIONS ────────────────────────────────

// convertConditionalAccessPolicyToIAMModel converts access service policy to IAM model
func convertConditionalAccessPolicyToIAMModel(accessPolicy *access.ConditionalAccessPolicy) *model.ConditionalAccessPolicy {
	if accessPolicy == nil {
		return nil
	}

	description := ""
	if accessPolicy.Description != nil {
		description = *accessPolicy.Description
	}

	return &model.ConditionalAccessPolicy{
		ID:          accessPolicy.ID,
		TenantID:    accessPolicy.TenantID,
		Name:        accessPolicy.Name,
		Description: description,
		Conditions:  convertGenericConditionsToIAM(accessPolicy.Conditions),
		Actions:     convertGenericActionsToIAM(accessPolicy.Actions),
		Priority:    accessPolicy.Priority,
		Enabled:     accessPolicy.IsActive,
		Metadata:    map[string]any{"source": "conditional_access"},
		CreatedAt:   accessPolicy.CreatedAt,
		UpdatedAt:   accessPolicy.UpdatedAt,
	}
}

// convertGenericConditionsToIAM converts generic conditions map to IAM model conditions
func convertGenericConditionsToIAM(conditions map[string]any) []*model.PolicyCondition {
	if conditions == nil {
		return nil
	}

	iamConditions := make([]*model.PolicyCondition, 0, len(conditions))
	for key, value := range conditions {
		iamConditions = append(iamConditions, &model.PolicyCondition{
			Expression: fmt.Sprintf("%s = %v", key, value),
			Attributes: map[string]any{key: value},
		})
	}

	return iamConditions
}

// convertGenericConditionsFromIAM converts IAM model policy conditions to generic map
func convertGenericConditionsFromIAM(iamConditions []*model.PolicyCondition) map[string]any {
	if iamConditions == nil {
		return nil
	}

	conditions := make(map[string]any)
	for _, iamCondition := range iamConditions {
		// Extract conditions from attributes
		for key, value := range iamCondition.Attributes {
			conditions[key] = value
		}
		// If no attributes, use expression as a generic condition
		if len(iamCondition.Attributes) == 0 {
			conditions["expression"] = iamCondition.Expression
		}
	}

	return conditions
}

// convertGenericActionsToIAM converts generic actions map to IAM model actions
func convertGenericActionsToIAM(actions map[string]any) []*model.PolicyAction {
	if actions == nil {
		return nil
	}

	iamActions := make([]*model.PolicyAction, 0, len(actions))
	for key, value := range actions {
		iamActions = append(iamActions, &model.PolicyAction{
			Type:        key,
			Description: fmt.Sprintf("Action: %s", key),
			Parameters:  map[string]any{"value": value},
		})
	}

	return iamActions
}

// convertGenericActionsFromIAM converts IAM model policy actions to generic map
func convertGenericActionsFromIAM(iamActions []*model.PolicyAction) map[string]any {
	if iamActions == nil {
		return nil
	}

	actions := make(map[string]any)
	for _, iamAction := range iamActions {
		if value, exists := iamAction.Parameters["value"]; exists {
			actions[iamAction.Type] = value
		} else {
			actions[iamAction.Type] = iamAction.Description
		}
	}

	return actions
}

// ─── PERMISSION CONVERSIONS ────────────────────────────────────────────────

// convertPermissionToIAMModel converts access service permission to IAM model
func convertPermissionToIAMModel(accessPerm *access.Permission) *model.Permission {
	if accessPerm == nil {
		return nil
	}

	return &model.Permission{
		ID:           accessPerm.ID,
		TenantID:     accessPerm.TenantID,
		ResourceType: accessPerm.ResourceType,
		ResourceID:   accessPerm.ResourceID,
		Action:       accessPerm.Action,
		EntityID:     accessPerm.EntityID,
		Conditions:   accessPerm.Conditions,
		ExpiresAt:    accessPerm.ExpiresAt,
		Metadata:     accessPerm.Metadata,
		CreatedAt:    accessPerm.CreatedAt,
		UpdatedAt:    accessPerm.UpdatedAt,
	}
}

// convertPermissionEffect converts access service permission effect to IAM model effect
func convertPermissionEffect(effect string) model.PolicyEffect {
	switch effect {
	case "ALLOW":
		return model.PolicyEffectAllow
	case "DENY":
		return model.PolicyEffectDeny
	default:
		return model.PolicyEffectDeny // Default to deny for safety
	}
}

// convertPolicyConditionsFromIAM converts IAM model policy conditions to access service policy conditions
func convertPolicyConditionsFromIAM(iamConditions []*model.PolicyCondition) []*access.PolicyCondition {
	if iamConditions == nil {
		return nil
	}

	accessConditions := make([]*access.PolicyCondition, len(iamConditions))
	for i, iamCondition := range iamConditions {
		accessConditions[i] = &access.PolicyCondition{
			Type:       "expression",
			Operator:   "equals",
			Value:      iamCondition.Expression,
			Expression: iamCondition.Expression,
			Metadata:   iamCondition.Attributes,
		}
	}

	return accessConditions
}

// convertPolicyActionsFromIAM converts IAM model policy actions to access service policy actions
func convertPolicyActionsFromIAM(iamActions []*model.PolicyAction) []*access.PolicyAction {
	if iamActions == nil {
		return nil
	}

	accessActions := make([]*access.PolicyAction, len(iamActions))
	for i, iamAction := range iamActions {
		value := iamAction.Description
		if paramValue, exists := iamAction.Parameters["value"]; exists {
			if paramStr, ok := paramValue.(string); ok {
				value = paramStr
			} else {
				value = fmt.Sprintf("%v", paramValue)
			}
		}

		accessActions[i] = &access.PolicyAction{
			Type:        iamAction.Type,
			Value:       value,
			Description: iamAction.Description,
			Parameters:  iamAction.Parameters,
			Metadata:    map[string]any{},
		}
	}

	return accessActions
}
