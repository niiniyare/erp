package repo

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// iamRepository implements IAMRepository interface
type iamRepository struct {
	store   db.Store
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService

	// Repository implementations
	userRepo        UserRepository
	personRepo      PersonRepository
	employeeRepo    EmployeeRepository
	roleRepo        RoleRepository
	userRoleRepo    UserRoleRepository
	sessionRepo     SessionRepository
	policyRepo      PolicyRepository
	attributeRepo   AttributeRepository
	permissionRepo  PermissionRepository
	accessReqRepo   AccessRequestRepository
	approvalRepo    ApprovalWorkflowRepository
	conditionalRepo ConditionalAccessRepository
	policyTplRepo   PolicyTemplateRepository
	policyVerRepo   PolicyVersionRepository
	analyticsRepo   UserAnalyticsRepository
}

// NewIAMRepository creates a new IAM repository implementation
func NewIAMRepository(
	store db.Store,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) IAMRepository {
	repo := &iamRepository{
		store:   store,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}

	// Initialize repository implementations
	repo.userRepo = NewUserRepository(store, logger, metrics, tracer)
	repo.personRepo = NewPersonRepository(store, logger, metrics, tracer)
	repo.employeeRepo = NewEmployeeRepository(store, logger, metrics, tracer)
	repo.roleRepo = NewRoleRepository(store, logger, metrics, tracer)
	// TODO: Implement remaining repository constructors
	// repo.userRoleRepo = NewUserRoleRepository(store, logger, metrics, tracer)
	// repo.sessionRepo = NewSessionRepository(store, logger, metrics, tracer)
	// repo.policyRepo = NewPolicyRepository(store, logger, metrics, tracer)
	// repo.attributeRepo = NewAttributeRepository(store, logger, metrics, tracer)
	// repo.permissionRepo = NewPermissionRepository(store, logger, metrics, tracer)
	// repo.accessReqRepo = NewAccessRequestRepository(store, logger, metrics, tracer)
	// repo.approvalRepo = NewApprovalWorkflowRepository(store, logger, metrics, tracer)
	// repo.conditionalRepo = NewConditionalAccessRepository(store, logger, metrics, tracer)
	// repo.policyTplRepo = NewPolicyTemplateRepository(store, logger, metrics, tracer)
	// repo.policyVerRepo = NewPolicyVersionRepository(store, logger, metrics, tracer)
	// repo.analyticsRepo = NewUserAnalyticsRepository(store, logger, metrics, tracer)

	return repo
}

// Core repositories
func (r *iamRepository) Users() UserRepository         { return r.userRepo }
func (r *iamRepository) Persons() PersonRepository     { return r.personRepo }
func (r *iamRepository) Employees() EmployeeRepository { return r.employeeRepo }
func (r *iamRepository) Roles() RoleRepository         { return r.roleRepo }
func (r *iamRepository) UserRoles() UserRoleRepository { return nil } // TODO: implement
func (r *iamRepository) Sessions() SessionRepository   { return nil } // TODO: implement

// Authorization repositories
func (r *iamRepository) Policies() PolicyRepository                     { return nil } // TODO: implement
func (r *iamRepository) Attributes() AttributeRepository                { return nil } // TODO: implement
func (r *iamRepository) Permissions() PermissionRepository              { return nil } // TODO: implement
func (r *iamRepository) AccessRequests() AccessRequestRepository        { return nil } // TODO: implement
func (r *iamRepository) ApprovalWorkflows() ApprovalWorkflowRepository  { return nil } // TODO: implement
func (r *iamRepository) ConditionalAccess() ConditionalAccessRepository { return nil } // TODO: implement

// Policy management repositories
func (r *iamRepository) PolicyTemplates() PolicyTemplateRepository { return nil } // TODO: implement
func (r *iamRepository) PolicyVersions() PolicyVersionRepository   { return nil } // TODO: implement

// Analytics repositories
func (r *iamRepository) UserAnalytics() UserAnalyticsRepository { return nil } // TODO: implement

// Store access
func (r *iamRepository) Store() db.Store { return r.store }
