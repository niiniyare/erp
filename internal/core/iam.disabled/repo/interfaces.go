package repo

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"

	db "awo/db/sqlc"
	"awo/internal/core/iam/model"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
	ListByStatus(ctx context.Context, status model.UserAccountStatus) ([]*model.User, error)
	Count(ctx context.Context) (int64, error)

	// Authentication specific
	GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, hash string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error

	// Account management
	LockAccount(ctx context.Context, userID uuid.UUID, lockedUntil *time.Time, reason string) error
	UnlockAccount(ctx context.Context, userID uuid.UUID) error
	UpdateFailedLoginCount(ctx context.Context, userID uuid.UUID, count int) error

	// MFA operations
	SetMFASecret(ctx context.Context, userID uuid.UUID, secret string) error
	GetMFASecret(ctx context.Context, userID uuid.UUID) (string, error)
	EnableMFA(ctx context.Context, userID uuid.UUID, method model.MFAMethod) error
	DisableMFA(ctx context.Context, userID uuid.UUID) error
}

// PersonRepository defines the interface for person data operations
type PersonRepository interface {
	Create(ctx context.Context, person *model.Person) (*model.Person, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Person, error)
	Update(ctx context.Context, person *model.Person) (*model.Person, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Person, error)
	Count(ctx context.Context) (int64, error)
	GetByEmail(ctx context.Context, email string) (*model.Person, error)
}

// EmployeeRepository defines the interface for employee data operations
type EmployeeRepository interface {
	Create(ctx context.Context, employee *model.Employee) (*model.Employee, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) (*model.Employee, error)
	GetByEmployeeNumber(ctx context.Context, employeeNumber string) (*model.Employee, error)
	Update(ctx context.Context, employee *model.Employee) (*model.Employee, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Employee, error)
	ListByStatus(ctx context.Context, status model.EmploymentStatus) ([]*model.Employee, error)
	ListByDepartment(ctx context.Context, department string) ([]*model.Employee, error)
	Count(ctx context.Context) (int64, error)
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) (*model.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	Update(ctx context.Context, role *model.Role) (*model.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Role, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*model.Role, error)
	Count(ctx context.Context) (int64, error)

	// Role hierarchy
	GetChildRoles(ctx context.Context, parentID uuid.UUID) ([]*model.Role, error)
	GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*model.Role, error)
}

// UserRoleRepository defines the interface for user role assignment operations
type UserRoleRepository interface {
	Assign(ctx context.Context, userRole *model.UserRole) error
	Remove(ctx context.Context, userID, roleID uuid.UUID, entityID *uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.UserRole, error)
	GetUserRolesByEntity(ctx context.Context, userID, entityID uuid.UUID) ([]*model.UserRole, error)
	GetRoleUsers(ctx context.Context, roleID uuid.UUID) ([]*model.UserRole, error)
	IsUserInRole(ctx context.Context, userID, roleID uuid.UUID, entityID *uuid.UUID) (bool, error)

	// Bulk operations
	AssignBulk(ctx context.Context, userRoles []*model.UserRole) error
	RemoveBulk(ctx context.Context, userRoleIDs []uuid.UUID) error
}

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) (*model.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	GetByToken(ctx context.Context, token string) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) (*model.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Session management
	InvalidateSession(ctx context.Context, sessionID uuid.UUID) error
	InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error
	InvalidateExpiredSessions(ctx context.Context) error
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error)
	GetUserSessionCount(ctx context.Context, userID uuid.UUID) (int, error)
}

// PolicyRepository defines the interface for policy data operations
type PolicyRepository interface {
	Create(ctx context.Context, policy *model.Policy) (*model.Policy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Policy, error)
	GetByName(ctx context.Context, name string) (*model.Policy, error)
	Update(ctx context.Context, policy *model.Policy) (*model.Policy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Policy, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*model.Policy, error)
	ListEnabled(ctx context.Context) ([]*model.Policy, error)
	Count(ctx context.Context) (int64, error)

	// Policy evaluation
	GetPoliciesForEvaluation(ctx context.Context, resourceType, action string, entityID *uuid.UUID) ([]*model.Policy, error)
	GetPoliciesByPriority(ctx context.Context, entityID *uuid.UUID) ([]*model.Policy, error)

	// Policy search
	SearchPolicies(ctx context.Context, query string, limit, offset int) ([]*model.Policy, error)
}

// AttributeRepository defines the interface for attribute data operations
type AttributeRepository interface {
	Create(ctx context.Context, attr *model.Attribute) (*model.Attribute, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Attribute, error)
	GetByName(ctx context.Context, name string) (*model.Attribute, error)
	Update(ctx context.Context, attr *model.Attribute) (*model.Attribute, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Attribute, error)
	ListByCategory(ctx context.Context, category model.AttributeCategory) ([]*model.Attribute, error)
	ListByType(ctx context.Context, dataType model.AttributeDataType) ([]*model.Attribute, error)
	Count(ctx context.Context) (int64, error)
}

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	Create(ctx context.Context, permission *model.Permission) (*model.Permission, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	Update(ctx context.Context, permission *model.Permission) (*model.Permission, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Permission queries
	GetUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*model.Permission, error)
	GetResourcePermissions(ctx context.Context, resourceType string, resourceID *uuid.UUID) ([]*model.Permission, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, resourceType, action string, resourceID *uuid.UUID) (bool, error)

	// Grant/Revoke operations
	GrantPermission(ctx context.Context, userID uuid.UUID, resourceType, action string, resourceID, entityID *uuid.UUID, expiresAt *time.Time) error
	RevokePermission(ctx context.Context, userID uuid.UUID, resourceType, action string, resourceID, entityID *uuid.UUID) error

	// Cleanup
	RemoveExpiredPermissions(ctx context.Context) error
}

// AccessRequestRepository defines the interface for access request data operations
type AccessRequestRepository interface {
	Create(ctx context.Context, request *model.AccessRequest) (*model.AccessRequest, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.AccessRequest, error)
	Update(ctx context.Context, request *model.AccessRequest) (*model.AccessRequest, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	List(ctx context.Context, limit, offset int) ([]*model.AccessRequest, error)
	ListByRequester(ctx context.Context, requesterID uuid.UUID) ([]*model.AccessRequest, error)
	ListByStatus(ctx context.Context, status model.ApprovalStatus) ([]*model.AccessRequest, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*model.AccessRequest, error)
	ListPendingApprovals(ctx context.Context, approverID uuid.UUID) ([]*model.AccessRequest, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status model.ApprovalStatus) (int64, error)

	// Status management
	UpdateStatus(ctx context.Context, requestID uuid.UUID, status model.ApprovalStatus, approvedBy *uuid.UUID, comments *string) error

	// Cleanup
	RemoveExpiredRequests(ctx context.Context) error
}

// ApprovalWorkflowRepository defines the interface for approval workflow data operations
type ApprovalWorkflowRepository interface {
	Create(ctx context.Context, workflow *model.ApprovalWorkflow) (*model.ApprovalWorkflow, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalWorkflow, error)
	GetByName(ctx context.Context, name string) (*model.ApprovalWorkflow, error)
	Update(ctx context.Context, workflow *model.ApprovalWorkflow) (*model.ApprovalWorkflow, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.ApprovalWorkflow, error)
	ListEnabled(ctx context.Context) ([]*model.ApprovalWorkflow, error)
	Count(ctx context.Context) (int64, error)
}

// ConditionalAccessRepository defines the interface for conditional access policy operations
type ConditionalAccessRepository interface {
	Create(ctx context.Context, policy *model.ConditionalAccessPolicy) (*model.ConditionalAccessPolicy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ConditionalAccessPolicy, error)
	Update(ctx context.Context, policy *model.ConditionalAccessPolicy) (*model.ConditionalAccessPolicy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.ConditionalAccessPolicy, error)
	ListEnabled(ctx context.Context) ([]*model.ConditionalAccessPolicy, error)
	ListByPriority(ctx context.Context) ([]*model.ConditionalAccessPolicy, error)
	Count(ctx context.Context) (int64, error)
}

// PolicyTemplateRepository defines the interface for policy template operations
type PolicyTemplateRepository interface {
	Create(ctx context.Context, template *model.PolicyTemplate) (*model.PolicyTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.PolicyTemplate, error)
	GetByName(ctx context.Context, name string) (*model.PolicyTemplate, error)
	Update(ctx context.Context, template *model.PolicyTemplate) (*model.PolicyTemplate, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.PolicyTemplate, error)
	ListByCategory(ctx context.Context, category string) ([]*model.PolicyTemplate, error)
	Count(ctx context.Context) (int64, error)
}

// PolicyVersionRepository defines the interface for policy version operations
type PolicyVersionRepository interface {
	Create(ctx context.Context, version *model.PolicyVersion) (*model.PolicyVersion, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.PolicyVersion, error)
	GetByPolicyAndVersion(ctx context.Context, policyID uuid.UUID, version int) (*model.PolicyVersion, error)
	Update(ctx context.Context, version *model.PolicyVersion) (*model.PolicyVersion, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPolicy(ctx context.Context, policyID uuid.UUID) ([]*model.PolicyVersion, error)
	GetActiveVersion(ctx context.Context, policyID uuid.UUID) (*model.PolicyVersion, error)
	SetActiveVersion(ctx context.Context, policyID uuid.UUID, version int) error
}

// UserAnalyticsRepository defines the interface for user analytics operations
type UserAnalyticsRepository interface {
	Create(ctx context.Context, analytics *model.UserAnalytics) (*model.UserAnalytics, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.UserAnalytics, error)
	Update(ctx context.Context, analytics *model.UserAnalytics) (*model.UserAnalytics, error)
	IncrementLoginCount(ctx context.Context, userID uuid.UUID) error
	IncrementFailedLoginCount(ctx context.Context, userID uuid.UUID) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error
	UpdateLastActivity(ctx context.Context, userID uuid.UUID, activityTime time.Time) error

	// Activity tracking
	RecordActivity(ctx context.Context, activity *model.UserActivity) error
	GetUserActivities(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.UserActivity, error)
	GetActivitiesByType(ctx context.Context, userID uuid.UUID, activityType string) ([]*model.UserActivity, error)
	GetActivitiesByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*model.UserActivity, error)
}

// IAMRepository aggregates all repository interfaces and provides access to the Store
type IAMRepository interface {
	// Core repositories
	Users() UserRepository
	Persons() PersonRepository
	Employees() EmployeeRepository
	Roles() RoleRepository
	UserRoles() UserRoleRepository
	Sessions() SessionRepository

	// Authorization repositories
	Policies() PolicyRepository
	Attributes() AttributeRepository
	Permissions() PermissionRepository
	AccessRequests() AccessRequestRepository
	ApprovalWorkflows() ApprovalWorkflowRepository
	ConditionalAccess() ConditionalAccessRepository

	// Policy management repositories
	PolicyTemplates() PolicyTemplateRepository
	PolicyVersions() PolicyVersionRepository

	// Analytics repositories
	UserAnalytics() UserAnalyticsRepository

	// Transaction support using existing Store interface
	Store() db.Store
}
