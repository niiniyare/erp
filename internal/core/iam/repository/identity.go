// Package repository provides IAM infrastructure adapters (DB + cache).
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
	"awo.so/internal/platform/cache"
	sharedErrors "awo.so/internal/shared/errors"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

const userCacheTTL = 30 * time.Minute

// ─── Port (interface) ─────────────────────────────────────────────────────────

// UserRepository defines the persistence port for identity data.
// Implementations handle both DB and cache — callers never touch cache directly.
type UserRepository interface {
	// Read — cache-aside: cache hit → return; miss → DB → populate cache
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserWithDetails(ctx context.Context, id uuid.UUID) (*domain.UserWithDetails, error)
	GetUserPassword(ctx context.Context, userID uuid.UUID) (string, error)

	// Write — invalidate cache after mutation
	CreateUser(ctx context.Context, req *domain.CreateUserRequest, hashedPassword string) (*domain.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error

	// Brute-force protection
	IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error
	ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error
	LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	// Person / Employee
	CreatePerson(ctx context.Context, req *domain.CreatePersonRequest) (*domain.Person, error)
	GetPersonByID(ctx context.Context, id uuid.UUID) (*domain.Person, error)
	CreateEmployee(ctx context.Context, req *domain.CreateEmployeeRequest) (*domain.Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error)

	// Collections
	ListUsers(ctx context.Context, req *domain.ListUsersRequest) ([]*domain.User, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, error)

	// Role assignments (user-level; Casbin-level is in AuthzRepository)
	AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
}

// ─── Adapter (implementation) ─────────────────────────────────────────────────

type userRepository struct {
	store   db.Store
	cache   cache.Service
	tracing tracing.Service
	metrics metrics.MetricsProvider
}

// NewUserRepository constructs a cache-backed Postgres UserRepository.
func NewUserRepository(
	store db.Store,
	cacheSvc cache.Service,
	tracer tracing.Service,
	m metrics.MetricsProvider,
) UserRepository {
	return &userRepository{
		store:   store,
		cache:   cacheSvc,
		tracing: tracer,
		metrics: m,
	}
}

// ─── Read operations (cache-aside) ────────────────────────────────────────────

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	key := fmt.Sprintf("user:id:%s", id)
	var u domain.User
	if err := r.cache.Get(ctx, key, &u); err == nil {
		return &u, nil
	}
	row, err := r.store.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sharedErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("iam repo: get user by id: %w", err)
	}
	user, err := fromSQLCUser(row)
	if err != nil {
		return nil, err
	}
	r.cacheUser(ctx, user)
	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	key := fmt.Sprintf("user:email:%s", email)
	var u domain.User
	if err := r.cache.Get(ctx, key, &u); err == nil {
		return &u, nil
	}
	row, err := r.store.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sharedErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("iam repo: get user by email: %w", err)
	}
	user, err := fromSQLCUser(row)
	if err != nil {
		return nil, err
	}
	r.cacheUser(ctx, user)
	return user, nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	key := fmt.Sprintf("user:username:%s", username)
	var u domain.User
	if err := r.cache.Get(ctx, key, &u); err == nil {
		return &u, nil
	}
	row, err := r.store.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sharedErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("iam repo: get user by username: %w", err)
	}
	user, err := fromSQLCUser(row)
	if err != nil {
		return nil, err
	}
	r.cacheUser(ctx, user)
	return user, nil
}

func (r *userRepository) GetUserWithDetails(ctx context.Context, id uuid.UUID) (*domain.UserWithDetails, error) {
	row, err := r.store.GetCompleteUserProfile(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sharedErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("iam repo: get user with details: %w", err)
	}
	return fromSQLCCompleteUserProfile(row)
}

func (r *userRepository) GetUserPassword(ctx context.Context, userID uuid.UUID) (string, error) {
	hash, err := r.store.GetUserPasswordByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", sharedErrors.ErrUserNotFound
		}
		return "", fmt.Errorf("iam repo: get user password: %w", err)
	}
	if hash == nil {
		return "", fmt.Errorf("iam repo: password not set for user")
	}
	return *hash, nil
}

// ─── Write operations (invalidate cache) ─────────────────────────────────────

func (r *userRepository) CreateUser(ctx context.Context, req *domain.CreateUserRequest, hashedPassword string) (*domain.User, error) {
	params, err := toSQLCCreateUserParams(req, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("iam repo: build create user params: %w", err)
	}
	row, err := r.store.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("iam repo: create user: %w", err)
	}
	user, err := fromSQLCUser(row)
	if err != nil {
		return nil, err
	}
	r.invalidateUser(ctx, user.ID, user.Email, user.Username)
	return user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error) {
	params := db.UpdateUserParams{ID: id}
	row, err := r.store.UpdateUser(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sharedErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("iam repo: update user: %w", err)
	}
	user, err := fromSQLCUser(row)
	if err != nil {
		return nil, err
	}
	r.invalidateUser(ctx, user.ID, user.Email, user.Username)
	r.cacheUser(ctx, user)
	return user, nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := r.store.SoftDeleteUser(ctx, id); err != nil {
		return fmt.Errorf("iam repo: delete user: %w", err)
	}
	r.invalidateUser(ctx, id, "", "")
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	return r.store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: &newPasswordHash,
	})
}

// ─── Brute-force protection ───────────────────────────────────────────────────

func (r *userRepository) IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "iam.repo.IncrementFailedAttempts")
	defer span.End()
	if err := r.store.IncrementFailedLogins(ctx, userID); err != nil {
		return fmt.Errorf("iam repo: increment failed attempts: %w", err)
	}
	r.invalidateUser(ctx, userID, "", "")
	return nil
}

func (r *userRepository) ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "iam.repo.ResetFailedAttempts")
	defer span.End()
	if err := r.store.UnlockUser(ctx, userID); err != nil {
		return fmt.Errorf("iam repo: reset failed attempts: %w", err)
	}
	r.invalidateUser(ctx, userID, "", "")
	return nil
}

func (r *userRepository) LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error {
	ctx, span := r.tracing.StartSpan(ctx, "iam.repo.LockAccount")
	defer span.End()
	if err := r.store.LockAccount(ctx, db.LockAccountParams{
		ID:           userID,
		LockoutUntil: sql.NullTime{Time: until, Valid: true},
	}); err != nil {
		return fmt.Errorf("iam repo: lock account: %w", err)
	}
	r.invalidateUser(ctx, userID, "", "")
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	ctx, span := r.tracing.StartSpan(ctx, "iam.repo.UpdateLastLogin")
	defer span.End()
	if err := r.store.UpdateUserLastLogin(ctx, userID); err != nil {
		return fmt.Errorf("iam repo: update last login: %w", err)
	}
	r.invalidateUser(ctx, userID, "", "")
	return nil
}

// ─── Person / Employee ────────────────────────────────────────────────────────

func (r *userRepository) CreatePerson(ctx context.Context, req *domain.CreatePersonRequest) (*domain.Person, error) {
	params, err := toSQLCCreatePersonParams(req)
	if err != nil {
		return nil, fmt.Errorf("iam repo: build create person params: %w", err)
	}
	row, err := r.store.CreatePerson(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("iam repo: create person: %w", err)
	}
	return fromSQLCPerson(row)
}

func (r *userRepository) GetPersonByID(ctx context.Context, id uuid.UUID) (*domain.Person, error) {
	row, err := r.store.GetPersonByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("iam repo: person not found")
		}
		return nil, fmt.Errorf("iam repo: get person: %w", err)
	}
	return fromSQLCPerson(row)
}

func (r *userRepository) CreateEmployee(ctx context.Context, req *domain.CreateEmployeeRequest) (*domain.Employee, error) {
	params, err := toSQLCCreateEmployeeParams(req)
	if err != nil {
		return nil, fmt.Errorf("iam repo: build create employee params: %w", err)
	}
	row, err := r.store.CreateEmployee(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("iam repo: create employee: %w", err)
	}
	return fromSQLCEmployee(row)
}

func (r *userRepository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.Employee, error) {
	row, err := r.store.GetEmployeeByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("iam repo: employee not found")
		}
		return nil, fmt.Errorf("iam repo: get employee: %w", err)
	}
	return fromSQLCEmployee(row)
}

// ─── Collections ─────────────────────────────────────────────────────────────

func (r *userRepository) ListUsers(ctx context.Context, req *domain.ListUsersRequest) ([]*domain.User, error) {
	limit := int32(req.Limit)
	offset := int32(req.Offset)
	if limit == 0 {
		limit = 50
	}
	rows, err := r.store.ListUsers(ctx, db.ListUsersParams{
		UserType:      req.UserType,
		AccountStatus: req.AccountStatus,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, fmt.Errorf("iam repo: list users: %w", err)
	}
	out := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		u, err := fromSQLCUser(row)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func (r *userRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*domain.User, error) {
	rows, err := r.store.SearchUsersAdvanced(ctx, db.SearchUsersAdvancedParams{
		Query:  &query,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("iam repo: search users: %w", err)
	}
	out := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		u, err := fromSQLCUser(row)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

// ─── Role assignments ─────────────────────────────────────────────────────────

func (r *userRepository) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	_, err := r.store.AssignUserRole(ctx, db.AssignUserRoleParams{
		PUserID:     userID,
		PRoleID:     roleID,
		PEntityID:   entityID,
		PAssignedBy: uuid.Nil,
	})
	return err
}

func (r *userRepository) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return r.store.RevokeUserRole(ctx, db.RevokeUserRoleParams{
		PUserID:   userID,
		PRoleID:   roleID,
		PEntityID: entityID,
	})
}

// ─── Cache helpers (internal) ─────────────────────────────────────────────────

func (r *userRepository) cacheUser(ctx context.Context, u *domain.User) {
	_ = r.cache.Set(ctx, fmt.Sprintf("user:id:%s", u.ID), u, userCacheTTL)
	_ = r.cache.Set(ctx, fmt.Sprintf("user:email:%s", u.Email), u, userCacheTTL)
	if u.Username != "" {
		_ = r.cache.Set(ctx, fmt.Sprintf("user:username:%s", u.Username), u, userCacheTTL)
	}
}

func (r *userRepository) invalidateUser(ctx context.Context, id uuid.UUID, email, username string) {
	_ = r.cache.Delete(ctx, fmt.Sprintf("user:id:%s", id))
	if email != "" {
		_ = r.cache.Delete(ctx, fmt.Sprintf("user:email:%s", email))
	}
	if username != "" {
		_ = r.cache.Delete(ctx, fmt.Sprintf("user:username:%s", username))
	}
}

// ─── SQLC conversion helpers ──────────────────────────────────────────────────

func fromSQLCUser(sqlcUser *db.User) (*domain.User, error) {
	var userAttributes, settings map[string]any
	if err := json.Unmarshal(sqlcUser.UserAttributes, &userAttributes); err != nil && len(sqlcUser.UserAttributes) > 0 {
		return nil, err
	}
	if err := json.Unmarshal(sqlcUser.Settings, &settings); err != nil && len(sqlcUser.Settings) > 0 {
		return nil, err
	}

	var accountStatus domain.AccountStatus
	if sqlcUser.AccountStatus != nil {
		accountStatus = domain.AccountStatus(*sqlcUser.AccountStatus)
	}

	var failedLoginAttempts int32
	if sqlcUser.FailedLoginAttempts != nil {
		failedLoginAttempts = *sqlcUser.FailedLoginAttempts
	}

	var sessionTimeoutMinutes int32
	if sqlcUser.SessionTimeoutMinutes != nil {
		sessionTimeoutMinutes = *sqlcUser.SessionTimeoutMinutes
	}

	var mfaEnabled bool
	if sqlcUser.MfaEnabled != nil {
		mfaEnabled = *sqlcUser.MfaEnabled
	}

	return &domain.User{
		ID:                    sqlcUser.ID,
		TenantID:              sqlcUser.TenantID,
		EntityID:              sqlcUser.EntityID,
		PersonID:              sqlcUser.PersonID,
		EmployeeID:            sqlcUser.EmployeeID,
		PrincipalID:           sqlcUser.PrincipalID,
		Username:              sqlcUser.Username,
		Email:                 sqlcUser.Email,
		DisplayName:           sqlcUser.DisplayName,
		UserType:              sqlcUser.UserType,
		AccountStatus:         accountStatus,
		IsActive:              sqlcUser.IsActive,
		LastLoginAt:           &sqlcUser.LastLoginAt.Time,
		PasswordChangedAt:     &sqlcUser.PasswordChangedAt.Time,
		FailedLoginAttempts:   failedLoginAttempts,
		LockoutUntil:          &sqlcUser.LockoutUntil.Time,
		SessionTimeoutMinutes: sessionTimeoutMinutes,
		MfaEnabled:            mfaEnabled,
		UserAttributes:        userAttributes,
		Settings:              settings,
		CreatedAt:             sqlcUser.CreatedAt,
		UpdatedAt:             sqlcUser.UpdatedAt,
		DeletedAt:             &sqlcUser.DeletedAt.Time,
	}, nil
}

func fromSQLCPerson(sqlcPerson *db.Person) (*domain.Person, error) {
	var securityAttributes, metadata map[string]any
	if err := json.Unmarshal(sqlcPerson.SecurityAttributes, &securityAttributes); err != nil && len(sqlcPerson.SecurityAttributes) > 0 {
		return nil, err
	}
	if err := json.Unmarshal(sqlcPerson.Metadata, &metadata); err != nil && len(sqlcPerson.Metadata) > 0 {
		return nil, err
	}

	var email string
	if sqlcPerson.Email != nil {
		email = *sqlcPerson.Email
	}

	return &domain.Person{
		ID:                 sqlcPerson.ID,
		TenantID:           sqlcPerson.TenantID,
		EntityID:           sqlcPerson.EntityID,
		PersonType:         sqlcPerson.PersonType,
		FirstName:          sqlcPerson.FirstName,
		LastName:           sqlcPerson.LastName,
		MiddleName:         sqlcPerson.MiddleName,
		Email:              email,
		Phone:              sqlcPerson.Phone,
		BirthDate:          &sqlcPerson.BirthDate,
		NationalID:         sqlcPerson.NationalID,
		TaxID:              sqlcPerson.TaxID,
		Address:            sqlcPerson.Address,
		SecurityAttributes: securityAttributes,
		Metadata:           metadata,
		IsActive:           sqlcPerson.IsActive,
		CreatedAt:          sqlcPerson.CreatedAt,
		UpdatedAt:          sqlcPerson.UpdatedAt,
		DeletedAt:          &sqlcPerson.DeletedAt.Time,
	}, nil
}

func fromSQLCEmployee(sqlcEmployee *db.Employee) (*domain.Employee, error) {
	var salaryInfo, workSchedule, accessAttributes map[string]any
	if err := json.Unmarshal(sqlcEmployee.SalaryInfo, &salaryInfo); err != nil && len(sqlcEmployee.SalaryInfo) > 0 {
		return nil, err
	}
	if err := json.Unmarshal(sqlcEmployee.WorkSchedule, &workSchedule); err != nil && len(sqlcEmployee.WorkSchedule) > 0 {
		return nil, err
	}
	if err := json.Unmarshal(sqlcEmployee.AccessAttributes, &accessAttributes); err != nil && len(sqlcEmployee.AccessAttributes) > 0 {
		return nil, err
	}

	var employmentStatus domain.EmploymentStatus
	if sqlcEmployee.EmploymentStatus != nil {
		employmentStatus = domain.EmploymentStatus(*sqlcEmployee.EmploymentStatus)
	}

	var securityLevel int32
	if sqlcEmployee.SecurityLevel != nil {
		securityLevel = *sqlcEmployee.SecurityLevel
	}

	return &domain.Employee{
		ID:               sqlcEmployee.ID,
		TenantID:         sqlcEmployee.TenantID,
		PersonID:         sqlcEmployee.PersonID,
		EmployeeNumber:   sqlcEmployee.EmployeeNumber,
		EntityID:         sqlcEmployee.EntityID,
		PositionTitle:    sqlcEmployee.PositionTitle,
		DepartmentID:     sqlcEmployee.DepartmentID,
		ManagerID:        sqlcEmployee.ManagerID,
		HireDate:         sqlcEmployee.HireDate,
		TerminationDate:  &sqlcEmployee.TerminationDate,
		SalaryInfo:       salaryInfo,
		Status:           employmentStatus,
		WorkSchedule:     workSchedule,
		SecurityLevel:    securityLevel,
		AccessAttributes: accessAttributes,
		CreatedAt:        sqlcEmployee.CreatedAt,
		UpdatedAt:        sqlcEmployee.UpdatedAt,
		DeletedAt:        &sqlcEmployee.DeletedAt.Time,
	}, nil
}

func fromSQLCCompleteUserProfile(row *db.GetCompleteUserProfileRow) (*domain.UserWithDetails, error) {
	user, err := fromSQLCUser(&db.User{
		ID:                    row.ID,
		TenantID:              row.TenantID,
		EntityID:              row.EntityID,
		PersonID:              row.PersonID,
		EmployeeID:            row.EmployeeID,
		DisplayName:           row.DisplayName,
		PrincipalID:           row.PrincipalID,
		Username:              row.Username,
		Email:                 row.Email,
		PasswordHash:          row.PasswordHash,
		UserType:              row.UserType,
		AccountStatus:         row.AccountStatus,
		IsActive:              row.IsActive,
		LastLoginAt:           row.LastLoginAt,
		PasswordChangedAt:     row.PasswordChangedAt,
		FailedLoginAttempts:   row.FailedLoginAttempts,
		LockoutUntil:          row.LockoutUntil,
		SessionTimeoutMinutes: row.SessionTimeoutMinutes,
		MfaEnabled:            row.MfaEnabled,
		MfaSecret:             row.MfaSecret,
		UserAttributes:        row.UserAttributes,
		Settings:              row.Settings,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("iam repo: convert user from profile: %w", err)
	}

	var person *domain.Person
	if row.ID_2 != nil {
		person, err = fromSQLCPerson(&db.Person{
			ID:                 *row.ID_2,
			TenantID:           *row.TenantID_2,
			EntityID:           *row.EntityID_2,
			PersonType:         *row.PersonType,
			FirstName:          *row.FirstName,
			LastName:           *row.LastName,
			MiddleName:         row.MiddleName,
			Email:              row.Email_2,
			Phone:              row.Phone,
			BirthDate:          row.BirthDate,
			NationalID:         row.NationalID,
			TaxID:              row.TaxID,
			Address:            row.Address,
			SecurityAttributes: row.SecurityAttributes,
			Metadata:           row.Metadata,
			IsActive:           *row.IsActive_2,
			CreatedAt:          row.CreatedAt_2.Time,
			UpdatedAt:          row.UpdatedAt_2.Time,
			DeletedAt:          row.DeletedAt_2,
		})
		if err != nil {
			return nil, fmt.Errorf("iam repo: convert person from profile: %w", err)
		}
	}

	var employee *domain.Employee
	if row.ID_3 != nil {
		employee, err = fromSQLCEmployee(&db.Employee{
			ID:               *row.ID_3,
			TenantID:         *row.TenantID_3,
			PersonID:         *row.PersonID_2,
			EmployeeNumber:   *row.EmployeeNumber,
			EntityID:         *row.EntityID_3,
			PositionTitle:    row.PositionTitle,
			DepartmentID:     row.DepartmentID,
			ManagerID:        row.ManagerID,
			HireDate:         row.HireDate,
			TerminationDate:  row.TerminationDate,
			SalaryInfo:       row.SalaryInfo,
			EmploymentStatus: row.EmploymentStatus,
			WorkSchedule:     row.WorkSchedule,
			SecurityLevel:    row.SecurityLevel,
			AccessAttributes: row.AccessAttributes,
			CreatedAt:        row.CreatedAt_3.Time,
			UpdatedAt:        row.UpdatedAt_3.Time,
			DeletedAt:        row.DeletedAt_3,
		})
		if err != nil {
			return nil, fmt.Errorf("iam repo: convert employee from profile: %w", err)
		}
	}

	return &domain.UserWithDetails{User: *user, Person: person, Employee: employee}, nil
}

func toSQLCCreateUserParams(req *domain.CreateUserRequest, hashedPassword string) (db.CreateUserParams, error) {
	var userAttributes, settings []byte
	var err error
	if req.UserAttributes != nil {
		if userAttributes, err = json.Marshal(req.UserAttributes); err != nil {
			return db.CreateUserParams{}, err
		}
	}
	if req.Settings != nil {
		if settings, err = json.Marshal(req.Settings); err != nil {
			return db.CreateUserParams{}, err
		}
	}
	return db.CreateUserParams{
		EntityID:              req.EntityID,
		PersonID:              req.PersonID,
		EmployeeID:            req.EmployeeID,
		Username:              req.Username,
		Email:                 req.Email,
		DisplayName:           req.DisplayName,
		PasswordHash:          &hashedPassword,
		UserType:              req.UserType,
		AccountStatus:         &req.AccountStatus,
		SessionTimeoutMinutes: &req.SessionTimeoutMinutes,
		MfaEnabled:            &req.MfaEnabled,
		UserAttributes:        userAttributes,
		Settings:              settings,
	}, nil
}

func toSQLCCreatePersonParams(req *domain.CreatePersonRequest) (db.CreatePersonParams, error) {
	var securityAttributes, metadata []byte
	var err error
	if req.SecurityAttributes != nil {
		if securityAttributes, err = json.Marshal(req.SecurityAttributes); err != nil {
			return db.CreatePersonParams{}, err
		}
	}
	if req.Metadata != nil {
		if metadata, err = json.Marshal(req.Metadata); err != nil {
			return db.CreatePersonParams{}, err
		}
	}
	return db.CreatePersonParams{
		EntityID:           req.EntityID,
		PersonType:         req.PersonType,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		MiddleName:         req.MiddleName,
		Email:              req.Email,
		Phone:              req.Phone,
		NationalID:         req.NationalID,
		TaxID:              req.TaxID,
		Address:            req.Address,
		SecurityAttributes: securityAttributes,
		Metadata:           metadata,
	}, nil
}

func toSQLCCreateEmployeeParams(req *domain.CreateEmployeeRequest) (db.CreateEmployeeParams, error) {
	var salaryInfo, workSchedule, accessAttributes []byte
	var err error
	if req.SalaryInfo != nil {
		if salaryInfo, err = json.Marshal(req.SalaryInfo); err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}
	if req.WorkSchedule != nil {
		if workSchedule, err = json.Marshal(req.WorkSchedule); err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}
	if req.AccessAttributes != nil {
		if accessAttributes, err = json.Marshal(req.AccessAttributes); err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}
	statusStr := string(req.Status)
	return db.CreateEmployeeParams{
		PersonID:         req.PersonID,
		EmployeeNumber:   req.EmployeeNumber,
		EntityID:         req.EntityID,
		PositionTitle:    req.PositionTitle,
		DepartmentID:     req.DepartmentID,
		ManagerID:        req.ManagerID,
		HireDate:         req.HireDate,
		SalaryInfo:       salaryInfo,
		EmploymentStatus: &statusStr,
		WorkSchedule:     workSchedule,
		SecurityLevel:    &req.SecurityLevel,
		AccessAttributes: accessAttributes,
	}, nil
}
