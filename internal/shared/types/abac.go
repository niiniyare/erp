package types

// ─── ABAC DOMAIN ENUMS AND CONSTANTS ─────────────────────────────────────────

// PolicyEffect represents the effect of a policy decision
type PolicyEffect string

const (
	PolicyEffectAllow         PolicyEffect = "ALLOW"
	PolicyEffectDeny          PolicyEffect = "DENY"
	PolicyEffectNotApplicable PolicyEffect = "NOT_APPLICABLE"
)

var validPolicyEffects = map[PolicyEffect]struct{}{
	PolicyEffectAllow:         {},
	PolicyEffectDeny:          {},
	PolicyEffectNotApplicable: {},
}

func (pe PolicyEffect) IsValid() bool {
	_, ok := validPolicyEffects[pe]
	return ok
}

func (pe PolicyEffect) String() string {
	return string(pe)
}

func AllPolicyEffects() []PolicyEffect {
	return []PolicyEffect{
		PolicyEffectAllow,
		PolicyEffectDeny,
		PolicyEffectNotApplicable,
	}
}

// PolicyDecisionType represents the type of policy decision
type PolicyDecisionType string

const (
	PolicyDecisionAllow         PolicyDecisionType = "ALLOW"
	PolicyDecisionDeny          PolicyDecisionType = "DENY"
	PolicyDecisionNotApplicable PolicyDecisionType = "NOT_APPLICABLE"
)

var validPolicyDecisionTypes = map[PolicyDecisionType]struct{}{
	PolicyDecisionAllow:         {},
	PolicyDecisionDeny:          {},
	PolicyDecisionNotApplicable: {},
}

func (pdt PolicyDecisionType) IsValid() bool {
	_, ok := validPolicyDecisionTypes[pdt]
	return ok
}

func (pdt PolicyDecisionType) String() string {
	return string(pdt)
}

func AllPolicyDecisionTypes() []PolicyDecisionType {
	return []PolicyDecisionType{
		PolicyDecisionAllow,
		PolicyDecisionDeny,
		PolicyDecisionNotApplicable,
	}
}

// AttributeDataType represents the data type of an attribute
type AttributeDataType string

const (
	AttributeDataTypeString  AttributeDataType = "STRING"
	AttributeDataTypeNumber  AttributeDataType = "NUMBER"
	AttributeDataTypeBoolean AttributeDataType = "BOOLEAN"
	AttributeDataTypeDate    AttributeDataType = "DATE"
	AttributeDataTypeJSON    AttributeDataType = "JSON"
	AttributeDataTypeArray   AttributeDataType = "ARRAY"
	AttributeDataTypeEnum    AttributeDataType = "ENUM"
)

var validAttributeDataTypes = map[AttributeDataType]struct{}{
	AttributeDataTypeString:  {},
	AttributeDataTypeNumber:  {},
	AttributeDataTypeBoolean: {},
	AttributeDataTypeDate:    {},
	AttributeDataTypeJSON:    {},
	AttributeDataTypeArray:   {},
	AttributeDataTypeEnum:    {},
}

func (adt AttributeDataType) IsValid() bool {
	_, ok := validAttributeDataTypes[adt]
	return ok
}

func (adt AttributeDataType) String() string {
	return string(adt)
}

func AllAttributeDataTypes() []AttributeDataType {
	return []AttributeDataType{
		AttributeDataTypeString,
		AttributeDataTypeNumber,
		AttributeDataTypeBoolean,
		AttributeDataTypeDate,
		AttributeDataTypeJSON,
		AttributeDataTypeArray,
		AttributeDataTypeEnum,
	}
}

// AttributeCategory represents the category of an attribute
type AttributeCategory string

const (
	AttributeCategoryUser        AttributeCategory = "USER"
	AttributeCategoryResource    AttributeCategory = "RESOURCE"
	AttributeCategoryEnvironment AttributeCategory = "ENVIRONMENT"
	AttributeCategoryAction      AttributeCategory = "ACTION"
	AttributeCategoryEntity      AttributeCategory = "ENTITY"
	AttributeCategorySession     AttributeCategory = "SESSION"
)

var validAttributeCategories = map[AttributeCategory]struct{}{
	AttributeCategoryUser:        {},
	AttributeCategoryResource:    {},
	AttributeCategoryEnvironment: {},
	AttributeCategoryAction:      {},
	AttributeCategoryEntity:      {},
	AttributeCategorySession:     {},
}

func (ac AttributeCategory) IsValid() bool {
	_, ok := validAttributeCategories[ac]
	return ok
}

func (ac AttributeCategory) String() string {
	return string(ac)
}

func AllAttributeCategories() []AttributeCategory {
	return []AttributeCategory{
		AttributeCategoryUser,
		AttributeCategoryResource,
		AttributeCategoryEnvironment,
		AttributeCategoryAction,
		AttributeCategoryEntity,
		AttributeCategorySession,
	}
}

// PolicyCombiningAlgorithm represents how multiple policies are combined
type PolicyCombiningAlgorithm string

const (
	CombiningAlgorithmDenyOverrides          PolicyCombiningAlgorithm = "DENY_OVERRIDES"
	CombiningAlgorithmPermitOverrides        PolicyCombiningAlgorithm = "PERMIT_OVERRIDES"
	CombiningAlgorithmFirstApplicable        PolicyCombiningAlgorithm = "FIRST_APPLICABLE"
	CombiningAlgorithmOnlyOneApplicable      PolicyCombiningAlgorithm = "ONLY_ONE_APPLICABLE"
	CombiningAlgorithmOrderedDenyOverrides   PolicyCombiningAlgorithm = "ORDERED_DENY_OVERRIDES"
	CombiningAlgorithmOrderedPermitOverrides PolicyCombiningAlgorithm = "ORDERED_PERMIT_OVERRIDES"
)

var validPolicyCombiningAlgorithms = map[PolicyCombiningAlgorithm]struct{}{
	CombiningAlgorithmDenyOverrides:          {},
	CombiningAlgorithmPermitOverrides:        {},
	CombiningAlgorithmFirstApplicable:        {},
	CombiningAlgorithmOnlyOneApplicable:      {},
	CombiningAlgorithmOrderedDenyOverrides:   {},
	CombiningAlgorithmOrderedPermitOverrides: {},
}

func (pca PolicyCombiningAlgorithm) IsValid() bool {
	_, ok := validPolicyCombiningAlgorithms[pca]
	return ok
}

func (pca PolicyCombiningAlgorithm) String() string {
	return string(pca)
}

func AllPolicyCombiningAlgorithms() []PolicyCombiningAlgorithm {
	return []PolicyCombiningAlgorithm{
		CombiningAlgorithmDenyOverrides,
		CombiningAlgorithmPermitOverrides,
		CombiningAlgorithmFirstApplicable,
		CombiningAlgorithmOnlyOneApplicable,
		CombiningAlgorithmOrderedDenyOverrides,
		CombiningAlgorithmOrderedPermitOverrides,
	}
}

// PolicyType represents the type of ABAC policy
type PolicyType string

const (
	PolicyTypeAccess     PolicyType = "ACCESS"     // Controls access to resources
	PolicyTypeDelegation PolicyType = "DELEGATION" // Controls delegation of permissions
	PolicyTypeObligation PolicyType = "OBLIGATION" // Specifies obligations (logging, notifications)
	PolicyTypeRefrain    PolicyType = "REFRAIN"    // Prohibits certain actions
	PolicyTypeCondition  PolicyType = "CONDITION"  // Conditional access policies
)

var validPolicyTypes = map[PolicyType]struct{}{
	PolicyTypeAccess:     {},
	PolicyTypeDelegation: {},
	PolicyTypeObligation: {},
	PolicyTypeRefrain:    {},
	PolicyTypeCondition:  {},
}

func (pt PolicyType) IsValid() bool {
	_, ok := validPolicyTypes[pt]
	return ok
}

func (pt PolicyType) String() string {
	return string(pt)
}

func AllPolicyTypes() []PolicyType {
	return []PolicyType{
		PolicyTypeAccess,
		PolicyTypeDelegation,
		PolicyTypeObligation,
		PolicyTypeRefrain,
		PolicyTypeCondition,
	}
}

// PolicyCategory represents the category of ABAC policy
type PolicyCategory string

const (
	PolicyCategorySecurity    PolicyCategory = "SECURITY"    // Security-related policies
	PolicyCategoryCompliance  PolicyCategory = "COMPLIANCE"  // Compliance and regulatory policies
	PolicyCategoryBusiness    PolicyCategory = "BUSINESS"    // Business logic policies
	PolicyCategoryOperational PolicyCategory = "OPERATIONAL" // Operational policies
	PolicyCategorySystem      PolicyCategory = "SYSTEM"      // System-level policies
)

var validPolicyCategories = map[PolicyCategory]struct{}{
	PolicyCategorySecurity:    {},
	PolicyCategoryCompliance:  {},
	PolicyCategoryBusiness:    {},
	PolicyCategoryOperational: {},
	PolicyCategorySystem:      {},
}

func (pc PolicyCategory) IsValid() bool {
	_, ok := validPolicyCategories[pc]
	return ok
}

func (pc PolicyCategory) String() string {
	return string(pc)
}

func AllPolicyCategories() []PolicyCategory {
	return []PolicyCategory{
		PolicyCategorySecurity,
		PolicyCategoryCompliance,
		PolicyCategoryBusiness,
		PolicyCategoryOperational,
		PolicyCategorySystem,
	}
}

// EvaluationStatus represents the status of a policy evaluation
type EvaluationStatus string

const (
	EvaluationStatusSuccess    EvaluationStatus = "SUCCESS"    // Evaluation completed successfully
	EvaluationStatusFailed     EvaluationStatus = "FAILED"     // Evaluation failed due to error
	EvaluationStatusTimeout    EvaluationStatus = "TIMEOUT"    // Evaluation timed out
	EvaluationStatusIncomplete EvaluationStatus = "INCOMPLETE" // Evaluation incomplete (missing attributes)
)

var validEvaluationStatuses = map[EvaluationStatus]struct{}{
	EvaluationStatusSuccess:    {},
	EvaluationStatusFailed:     {},
	EvaluationStatusTimeout:    {},
	EvaluationStatusIncomplete: {},
}

func (es EvaluationStatus) IsValid() bool {
	_, ok := validEvaluationStatuses[es]
	return ok
}

func (es EvaluationStatus) String() string {
	return string(es)
}

func AllEvaluationStatuses() []EvaluationStatus {
	return []EvaluationStatus{
		EvaluationStatusSuccess,
		EvaluationStatusFailed,
		EvaluationStatusTimeout,
		EvaluationStatusIncomplete,
	}
}

// AttributeSource represents the source of an attribute value
type AttributeSource string

const (
	AttributeSourceIdentity    AttributeSource = "IDENTITY"    // From identity service
	AttributeSourceSession     AttributeSource = "SESSION"     // From session context
	AttributeSourceEnvironment AttributeSource = "ENVIRONMENT" // From environment context
	AttributeSourceResource    AttributeSource = "RESOURCE"    // From resource metadata
	AttributeSourceExternal    AttributeSource = "EXTERNAL"    // From external systems
	AttributeSourceComputed    AttributeSource = "COMPUTED"    // Computed/derived
	AttributeSourceCache       AttributeSource = "CACHE"       // From cache
)

var validAttributeSources = map[AttributeSource]struct{}{
	AttributeSourceIdentity:    {},
	AttributeSourceSession:     {},
	AttributeSourceEnvironment: {},
	AttributeSourceResource:    {},
	AttributeSourceExternal:    {},
	AttributeSourceComputed:    {},
	AttributeSourceCache:       {},
}

func (as AttributeSource) IsValid() bool {
	_, ok := validAttributeSources[as]
	return ok
}

func (as AttributeSource) String() string {
	return string(as)
}

func AllAttributeSources() []AttributeSource {
	return []AttributeSource{
		AttributeSourceIdentity,
		AttributeSourceSession,
		AttributeSourceEnvironment,
		AttributeSourceResource,
		AttributeSourceExternal,
		AttributeSourceComputed,
		AttributeSourceCache,
	}
}

// CombiningAlgorithm is an alias for PolicyCombiningAlgorithm
type CombiningAlgorithm = PolicyCombiningAlgorithm
