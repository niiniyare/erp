package model

import (
	"time"

	"github.com/google/uuid"
)

// ─── IDENTITY MODELS ─────────────────────────────────────────────────────────

// User represents a user account in the system
type User struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         uuid.UUID              `json:"tenant_id"`
	Email            string                 `json:"email"`
	PasswordHash     string                 `json:"-"` // Never serialize password hash
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	AccountStatus    UserAccountStatus      `json:"account_status"`
	EmailVerified    bool                   `json:"email_verified"`
	PhoneVerified    bool                   `json:"phone_verified"`
	MFAEnabled       bool                   `json:"mfa_enabled"`
	MFAMethod        *MFAMethod             `json:"mfa_method,omitempty"`
	MFASecret        *string                `json:"-"` // Never serialize MFA secret
	LastLoginAt      *time.Time             `json:"last_login_at,omitempty"`
	PasswordExpired  bool                   `json:"password_expired"`
	FailedLoginCount int                    `json:"failed_login_count"`
	LockedUntil      *time.Time             `json:"locked_until,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty"`
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	return u.AccountStatus == UserAccountStatusLocked ||
		(u.LockedUntil != nil && u.LockedUntil.After(time.Now()))
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.AccountStatus == UserAccountStatusActive && !u.IsLocked()
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// Person represents a person entity (can exist without a user account)
type Person struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	EntityID           uuid.UUID              `json:"entity_id"`
	PersonType         PersonType             `json:"person_type"`
	FirstName          string                 `json:"first_name"`
	LastName           string                 `json:"last_name"`
	MiddleName         *string                `json:"middle_name,omitempty"`
	Email              *string                `json:"email,omitempty"`
	PhoneNumber        *string                `json:"phone_number,omitempty"`
	BirthDate          time.Time              `json:"birth_date"`
	NationalID         *string                `json:"national_id,omitempty"`
	TaxID              *string                `json:"tax_id,omitempty"`
	Address            map[string]interface{} `json:"address,omitempty"`
	SecurityAttributes map[string]interface{} `json:"security_attributes,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	IsActive           bool                   `json:"is_active"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          *time.Time             `json:"deleted_at,omitempty"`
}

// FullName returns the person's full name
func (p *Person) FullName() string {
	return p.FirstName + " " + p.LastName
}

// Employee represents an employee (links to a person)
type Employee struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         uuid.UUID              `json:"tenant_id"`
	PersonID         uuid.UUID              `json:"person_id"`
	EmployeeNumber   string                 `json:"employee_number"`
	EntityID         uuid.UUID              `json:"entity_id"`
	PositionTitle    *string                `json:"position_title,omitempty"`
	DepartmentID     *uuid.UUID             `json:"department_id,omitempty"`
	ManagerID        *uuid.UUID             `json:"manager_id,omitempty"`
	HireDate         time.Time              `json:"hire_date"`
	TerminationDate  *time.Time             `json:"termination_date,omitempty"`
	SalaryInfo       map[string]interface{} `json:"salary_info,omitempty"`
	EmploymentStatus EmploymentStatus       `json:"employment_status"`
	WorkSchedule     map[string]interface{} `json:"work_schedule,omitempty"`
	SecurityLevel    int                    `json:"security_level"`
	AccessAttributes map[string]interface{} `json:"access_attributes,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty"`
}

// IsActive checks if the employee is currently active
func (e *Employee) IsActive() bool {
	return e.EmploymentStatus == EmploymentStatusActive
}

// Address represents a physical address
type Address struct {
	Street     string   `json:"street"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	PostalCode string   `json:"postal_code"`
	Country    string   `json:"country"`
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
}

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
	EntityID    *uuid.UUID             `json:"entity_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	UserID    uuid.UUID  `json:"user_id"`
	RoleID    uuid.UUID  `json:"role_id"`
	EntityID  *uuid.UUID `json:"entity_id,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// IsExpired checks if the role assignment has expired
func (ur *UserRole) IsExpired() bool {
	return ur.ExpiresAt != nil && ur.ExpiresAt.Before(time.Now())
}

// Session represents a user session
type Session struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenant_id"`
	UserID    uuid.UUID     `json:"user_id"`
	Token     string        `json:"token"`
	Status    SessionStatus `json:"status"`
	IPAddress string        `json:"ip_address"`
	UserAgent string        `json:"user_agent"`
	ExpiresAt time.Time     `json:"expires_at"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now()) || s.Status != SessionStatusActive
}

// ─── AUTHORIZATION MODELS ────────────────────────────────────────────────────

// Permission represents a permission in the system
type Permission struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	Action       string                 `json:"action"`
	EntityID     *uuid.UUID             `json:"entity_id,omitempty"`
	Conditions   []string               `json:"conditions,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// IsExpired checks if the permission has expired
func (p *Permission) IsExpired() bool {
	return p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now())
}

// Policy represents an ABAC policy
type Policy struct {
	ID          uuid.UUID                `json:"id"`
	TenantID    uuid.UUID                `json:"tenant_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Version     int                      `json:"version"`
	Effect      PolicyEffect             `json:"effect"`
	Target      *PolicyTarget            `json:"target,omitempty"`
	Condition   *PolicyCondition         `json:"condition,omitempty"`
	Rules       []*PolicyRule            `json:"rules,omitempty"`
	Priority    int                      `json:"priority"`
	Enabled     bool                     `json:"enabled"`
	Algorithm   PolicyCombiningAlgorithm `json:"algorithm"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	DeletedAt   *time.Time               `json:"deleted_at,omitempty"`
}

// PolicyTarget defines what the policy applies to
type PolicyTarget struct {
	Subjects  []*PolicySubject  `json:"subjects,omitempty"`
	Resources []*PolicyResource `json:"resources,omitempty"`
	Actions   []string          `json:"actions,omitempty"`
	EntityID  *uuid.UUID        `json:"entity_id,omitempty"`
}

// PolicySubject represents a subject in a policy
type PolicySubject struct {
	Type       string                 `json:"type"` // user, role, group
	ID         *uuid.UUID             `json:"id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyResource represents a resource in a policy
type PolicyResource struct {
	Type       string                 `json:"type"`
	ID         *uuid.UUID             `json:"id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyCondition represents a condition in a policy
type PolicyCondition struct {
	Expression string                 `json:"expression"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyRule represents a rule within a policy
type PolicyRule struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Effect      PolicyEffect           `json:"effect"`
	Condition   *PolicyCondition       `json:"condition,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyDecision represents the result of evaluating a policy
type PolicyDecision struct {
	PolicyID    uuid.UUID              `json:"policy_id"`
	PolicyName  string                 `json:"policy_name"`
	Decision    PolicyDecisionType     `json:"decision"`
	Effect      PolicyEffect           `json:"effect"`
	Reason      string                 `json:"reason"`
	Obligations []*PolicyObligation    `json:"obligations,omitempty"`
	Advice      []*PolicyAdvice        `json:"advice,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
}

// PolicyObligation represents an obligation that must be fulfilled
type PolicyObligation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyAdvice represents advice for the decision
type PolicyAdvice struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// Attribute represents an ABAC attribute
type Attribute struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	Name         string                 `json:"name"`
	Type         AttributeDataType      `json:"type"`
	Category     AttributeCategory      `json:"category"`
	Description  string                 `json:"description"`
	Required     bool                   `json:"required"`
	Multivalued  bool                   `json:"multivalued"`
	DefaultValue interface{}            `json:"default_value,omitempty"`
	Constraints  *AttributeConstraints  `json:"constraints,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty"`
}

// AttributeConstraints defines constraints for an attribute
type AttributeConstraints struct {
	MinValue      *float64 `json:"min_value,omitempty"`
	MaxValue      *float64 `json:"max_value,omitempty"`
	MinLength     *int     `json:"min_length,omitempty"`
	MaxLength     *int     `json:"max_length,omitempty"`
	Pattern       *string  `json:"pattern,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
}

// ─── ACCESS REQUEST MODELS ───────────────────────────────────────────────────

// AccessRequest represents an access request
type AccessRequest struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         uuid.UUID              `json:"tenant_id"`
	RequesterID      uuid.UUID              `json:"requester_id"`
	TargetUserID     *uuid.UUID             `json:"target_user_id,omitempty"`
	EntityID         *uuid.UUID             `json:"entity_id,omitempty"`
	RequestType      RequestType            `json:"request_type"`
	ResourceType     string                 `json:"resource_type"`
	ResourceID       *uuid.UUID             `json:"resource_id,omitempty"`
	Action           string                 `json:"action"`
	RoleID           *uuid.UUID             `json:"role_id,omitempty"`
	PermissionID     *uuid.UUID             `json:"permission_id,omitempty"`
	Justification    string                 `json:"justification"`
	BusinessReason   *string                `json:"business_reason,omitempty"`
	Duration         *time.Duration         `json:"duration,omitempty"`
	Priority         string                 `json:"priority"`
	ApprovalStatus   ApprovalStatus         `json:"approval_status"`
	ApprovedBy       *uuid.UUID             `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time             `json:"approved_at,omitempty"`
	ApprovalComments *string                `json:"approval_comments,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	AutoRevoke       bool                   `json:"auto_revoke"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// IsExpired checks if the access request has expired
func (ar *AccessRequest) IsExpired() bool {
	return ar.ExpiresAt != nil && ar.ExpiresAt.Before(time.Now())
}

// CanBeApproved checks if the request is in a state that allows approval
func (ar *AccessRequest) CanBeApproved() bool {
	return ar.ApprovalStatus == ApprovalStatusPending && !ar.IsExpired()
}

// ApprovalWorkflow represents an approval workflow
type ApprovalWorkflow struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []*ApprovalStep        `json:"steps"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ApprovalStep represents a step in an approval workflow
type ApprovalStep struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Order         int         `json:"order"`
	RequiredVotes int         `json:"required_votes"`
	ApproverRoles []string    `json:"approver_roles,omitempty"`
	ApproverUsers []uuid.UUID `json:"approver_users,omitempty"`
	TimeoutHours  *int        `json:"timeout_hours,omitempty"`
	AutoApprove   bool        `json:"auto_approve"`
}

// ConditionalAccessPolicy represents a conditional access policy
type ConditionalAccessPolicy struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Conditions  []*PolicyCondition     `json:"conditions"`
	Actions     []*PolicyAction        `json:"actions"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PolicyAction represents an action in a conditional access policy
type PolicyAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AccessContext represents the context for conditional access evaluation
type AccessContext struct {
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Location   *GeolocationContext    `json:"location,omitempty"`
	Device     *DeviceContext         `json:"device,omitempty"`
	Time       time.Time              `json:"time"`
	RiskLevel  string                 `json:"risk_level"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// GeolocationContext represents geolocation information
type GeolocationContext struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// DeviceContext represents device information
type DeviceContext struct {
	DeviceID  string `json:"device_id"`
	Platform  string `json:"platform"`
	Browser   string `json:"browser"`
	IsTrusted bool   `json:"is_trusted"`
	IsManaged bool   `json:"is_managed"`
}

// AccessCondition represents a condition for access
type AccessCondition struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ─── POLICY TEMPLATE MODELS ──────────────────────────────────────────────────

// PolicyTemplate represents a policy template
type PolicyTemplate struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Template    *PolicyTemplateSpec    `json:"template"`
	Parameters  []*TemplateParameter   `json:"parameters,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PolicyTemplateSpec defines the template specification
type PolicyTemplateSpec struct {
	Effect    PolicyEffect      `json:"effect"`
	Target    *PolicyTarget     `json:"target,omitempty"`
	Condition *PolicyCondition  `json:"condition,omitempty"`
	Rules     []*PolicyRule     `json:"rules,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
}

// TemplateParameter represents a parameter in a policy template
type TemplateParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
}

// PolicyVersion represents a version of a policy
type PolicyVersion struct {
	ID        uuid.UUID              `json:"id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	PolicyID  uuid.UUID              `json:"policy_id"`
	Version   int                    `json:"version"`
	Changes   string                 `json:"changes"`
	Target    *PolicyTarget          `json:"target,omitempty"`
	Condition *PolicyCondition       `json:"condition,omitempty"`
	Rules     []*PolicyRule          `json:"rules,omitempty"`
	IsActive  bool                   `json:"is_active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// ─── ANALYTICS MODELS ────────────────────────────────────────────────────────

// UserActivity represents user activity tracking
type UserActivity struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UserID       uuid.UUID              `json:"user_id"`
	ActivityType string                 `json:"activity_type"`
	Resource     string                 `json:"resource"`
	Action       string                 `json:"action"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Timestamp    time.Time              `json:"timestamp"`
}

// UserAnalytics represents user analytics data
type UserAnalytics struct {
	UserID                 uuid.UUID  `json:"user_id"`
	TenantID               uuid.UUID  `json:"tenant_id"`
	LoginCount             int64      `json:"login_count"`
	LastLoginAt            *time.Time `json:"last_login_at,omitempty"`
	FailedLoginCount       int64      `json:"failed_login_count"`
	SessionCount           int64      `json:"session_count"`
	AverageSessionDuration int64      `json:"average_session_duration_minutes"`
	AccessRequestCount     int64      `json:"access_request_count"`
	PermissionUsageCount   int64      `json:"permission_usage_count"`
	LastActivityAt         *time.Time `json:"last_activity_at,omitempty"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// ─── POLICY ANALYSIS MODELS ──────────────────────────────────────────────────

// PolicyConflict represents a conflict between policies
type PolicyConflict struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	PolicyIDs   []uuid.UUID `json:"policy_ids"`
	Suggestion  string      `json:"suggestion"`
}

// PolicyWarning represents a warning about a policy
type PolicyWarning struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	PolicyID    uuid.UUID `json:"policy_id"`
	Suggestion  string    `json:"suggestion"`
}

// PolicySuggestion represents a suggestion for policy improvement
type PolicySuggestion struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	PolicyID    uuid.UUID `json:"policy_id"`
	Action      string    `json:"action"`
}

// PolicyTestCase represents a test case for policy testing
type PolicyTestCase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Subject     *PolicySubject         `json:"subject"`
	Resource    *PolicyResource        `json:"resource"`
	Action      string                 `json:"action"`
	Environment *PolicyEnvironment     `json:"environment,omitempty"`
	Expected    PolicyDecisionType     `json:"expected"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyEnvironment represents the environment context for policy evaluation
type PolicyEnvironment struct {
	Time       time.Time              `json:"time"`
	IPAddress  string                 `json:"ip_address"`
	Location   *GeolocationContext    `json:"location,omitempty"`
	Device     *DeviceContext         `json:"device,omitempty"`
	RiskLevel  string                 `json:"risk_level"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// PolicyTestResult represents the result of a policy test
type PolicyTestResult struct {
	TestCaseID string             `json:"test_case_id"`
	Passed     bool               `json:"passed"`
	Actual     PolicyDecisionType `json:"actual"`
	Expected   PolicyDecisionType `json:"expected"`
	Reason     string             `json:"reason,omitempty"`
	Duration   time.Duration      `json:"duration"`
}
