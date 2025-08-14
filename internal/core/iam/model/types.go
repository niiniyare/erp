package model

// ─── IAM DOMAIN ENUMS AND CONSTANTS ─────────────────────────────────────────

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

// RequestType represents the type of access request
type RequestType string

const (
	RequestTypeRoleAssignment  RequestType = "ROLE_ASSIGNMENT"
	RequestTypePermissionGrant RequestType = "PERMISSION_GRANT"
	RequestTypeResourceAccess  RequestType = "RESOURCE_ACCESS"
	RequestTypeElevation       RequestType = "ELEVATION"
)

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
	ApprovalStatusExpired  ApprovalStatus = "EXPIRED"
	ApprovalStatusRevoked  ApprovalStatus = "REVOKED"
)

// UserAccountStatus represents the status of a user account
type UserAccountStatus string

const (
	UserAccountStatusActive    UserAccountStatus = "ACTIVE"
	UserAccountStatusInactive  UserAccountStatus = "INACTIVE"
	UserAccountStatusLocked    UserAccountStatus = "LOCKED"
	UserAccountStatusSuspended UserAccountStatus = "SUSPENDED"
	UserAccountStatusPending   UserAccountStatus = "PENDING"
)

// EmploymentStatus represents the employment status of an employee
type EmploymentStatus string

const (
	EmploymentStatusActive     EmploymentStatus = "ACTIVE"
	EmploymentStatusInactive   EmploymentStatus = "INACTIVE"
	EmploymentStatusTerminated EmploymentStatus = "TERMINATED"
	EmploymentStatusOnLeave    EmploymentStatus = "ON_LEAVE"
	EmploymentStatusContractor EmploymentStatus = "CONTRACTOR"
)

// MFAMethod represents multi-factor authentication methods
type MFAMethod string

const (
	MFAMethodTOTP  MFAMethod = "TOTP"
	MFAMethodSMS   MFAMethod = "SMS"
	MFAMethodEmail MFAMethod = "EMAIL"
	MFAMethodPush  MFAMethod = "PUSH"
)

// SessionStatus represents the status of a user session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "ACTIVE"
	SessionStatusExpired  SessionStatus = "EXPIRED"
	SessionStatusRevoked  SessionStatus = "REVOKED"
	SessionStatusInactive SessionStatus = "INACTIVE"
)
