package identity

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=mock.go -package=identity


import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Repository defines the interface for identity data persistence.
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, req *CreateUserRequest, hashedPassword string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error // Soft delete
	GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error)
	GetUserPassword(ctx context.Context, userID uuid.UUID) (string, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error

	// Person operations
	CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error)
	GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error)

	// Employee operations
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error)

	AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
	RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error
}

// repository implements the Repository interface
type repository struct {
	store   db.Store
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// NewRepository creates a new user repository
func NewRepository(store db.Store, tracing tracing.TracingService, metrics metrics.MetricsProvider) Repository {
	return &repository{
		store:   store,
		tracing: tracing,
		metrics: metrics,
	}
}

// CreateUser creates a new user
func (r *repository) CreateUser(ctx context.Context, req *CreateUserRequest, hashedPassword string) (*User, error) {
	params, err := toSQLCCreateUserParams(req, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request to SQLC params: %w", err)
	}

	sqlcUser, err := r.store.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return fromSQLCUser(sqlcUser)
}

// GetUserByID retrieves a user by ID
func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	sqlcUser, err := r.store.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromSQLCUser(sqlcUser)
}

// GetUserByEmail retrieves a user by email
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	sqlcUser, err := r.store.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return fromSQLCUser(sqlcUser)
}

// GetUserByUsername retrieves a user by username
func (r *repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	sqlcUser, err := r.store.GetUserByUsername(ctx, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return fromSQLCUser(sqlcUser)
}

// UpdateUser updates an existing user
func (r *repository) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	params := db.UpdateUserParams{ID: id}
	// Populate params from req...

	sqlcUser, err := r.store.UpdateUser(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return fromSQLCUser(sqlcUser)
}

// DeleteUser soft-deletes a user
func (r *repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.store.SoftDeleteUser(ctx, id)
}

// GetUserWithDetails retrieves a user with person and employee details
func (r *repository) GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error) {
	sqlcProfile, err := r.store.GetCompleteUserProfile(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user with details: %w", err)
	}

	return fromSQLCCompleteUserProfile(sqlcProfile)
}

func (r *repository) GetUserPassword(ctx context.Context, userID uuid.UUID) (string, error) {
	hash, err := r.store.GetUserPasswordByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.ErrUserNotFound
		}
		return "", err
	}
	if hash == nil {
		return "", fmt.Errorf("password not found for user")
	}
	return *hash, nil
}

func (r *repository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	return r.store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: &newPasswordHash,
	})
}

// Person operations
func (r *repository) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error) {
	params, err := toSQLCCreatePersonParams(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}
	sqlcPerson, err := r.store.CreatePerson(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create person: %w", err)
	}
	return fromSQLCPerson(sqlcPerson)
}

func (r *repository) GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	sqlcPerson, err := r.store.GetPersonByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person not found")
		}
		return nil, fmt.Errorf("failed to get person: %w", err)
	}
	return fromSQLCPerson(sqlcPerson)
}

// Employee operations
func (r *repository) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	params, err := toSQLCCreateEmployeeParams(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}
	sqlcEmployee, err := r.store.CreateEmployee(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}
	return fromSQLCEmployee(sqlcEmployee)
}

func (r *repository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	sqlcEmployee, err := r.store.GetEmployeeByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	return fromSQLCEmployee(sqlcEmployee)
}

func (r *repository) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	_, err := r.store.AssignUserRole(ctx, db.AssignUserRoleParams{
		PUserID:   userID,
		PRoleID:   roleID,
		PEntityID: entityID,
	})
	return err
}

func (r *repository) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	return r.store.RevokeUserRole(ctx, db.RevokeUserRoleParams{
		PUserID:   userID,
		PRoleID:   roleID,
		PEntityID: entityID,
	})
}

// --- Conversion Helpers ---

func fromSQLCUser(sqlcUser *db.User) (*User, error) {
	var userAttributes, settings map[string]any
	if err := json.Unmarshal(sqlcUser.UserAttributes, &userAttributes); err != nil && len(sqlcUser.UserAttributes) > 0 {
		return nil, err
	}
	if err := json.Unmarshal(sqlcUser.Settings, &settings); err != nil && len(sqlcUser.Settings) > 0 {
		return nil, err
	}

	var username string
	if sqlcUser.Username != nil {
		username = *sqlcUser.Username
	}

	var accountStatus AccountStatus
	if sqlcUser.AccountStatus != nil {
		accountStatus = AccountStatus(*sqlcUser.AccountStatus)
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

	return &User{
		ID:                    sqlcUser.ID,
		TenantID:              sqlcUser.TenantID,
		EntityID:              sqlcUser.EntityID,
		PersonID:              sqlcUser.PersonID,
		EmployeeID:            sqlcUser.EmployeeID,
		Username:              username,
		Email:                 sqlcUser.Email,
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

func fromSQLCPerson(sqlcPerson *db.Person) (*Person, error) {
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

	return &Person{
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

func fromSQLCEmployee(sqlcEmployee *db.Employee) (*Employee, error) {
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

	var employmentStatus EmploymentStatus
	if sqlcEmployee.EmploymentStatus != nil {
		employmentStatus = EmploymentStatus(*sqlcEmployee.EmploymentStatus)
	}

	var securityLevel int32
	if sqlcEmployee.SecurityLevel != nil {
		securityLevel = *sqlcEmployee.SecurityLevel
	}

	return &Employee{
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

func fromSQLCCompleteUserProfile(row *db.GetCompleteUserProfileRow) (*UserWithDetails, error) {
	user, err := fromSQLCUser(&db.User{
		ID:                    row.ID,
		TenantID:              row.TenantID,
		EntityID:              row.EntityID,
		PersonID:              row.PersonID,
		EmployeeID:            row.EmployeeID,
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
		return nil, fmt.Errorf("failed to convert user from profile: %w", err)
	}

	var person *Person
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
			return nil, fmt.Errorf("failed to convert person from profile: %w", err)
		}
	}

	var employee *Employee
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
			return nil, fmt.Errorf("failed to convert employee from profile: %w", err)
		}
	}

	return &UserWithDetails{
		User:     *user,
		Person:   person,
		Employee: employee,
	}, nil
}

func toSQLCCreateUserParams(req *CreateUserRequest, hashedPassword string) (db.CreateUserParams, error) {
	var userAttributes, settings []byte
	var err error
	if req.UserAttributes != nil {
		userAttributes, err = json.Marshal(req.UserAttributes)
		if err != nil {
			return db.CreateUserParams{}, err
		}
	}
	if req.Settings != nil {
		settings, err = json.Marshal(req.Settings)
		if err != nil {
			return db.CreateUserParams{}, err
		}
	}

	return db.CreateUserParams{
		EntityID:              req.EntityID,
		PersonID:              req.PersonID,
		EmployeeID:            req.EmployeeID,
		Username:              &req.Username,
		Email:                 req.Email,
		PasswordHash:          &hashedPassword,
		UserType:              req.UserType,
		AccountStatus:         &req.AccountStatus,
		SessionTimeoutMinutes: &req.SessionTimeoutMinutes,
		MfaEnabled:            &req.MfaEnabled,
		UserAttributes:        userAttributes,
		Settings:              settings,
	}, nil
}

func toSQLCCreatePersonParams(req *CreatePersonRequest) (db.CreatePersonParams, error) {
	var securityAttributes, metadata []byte
	var err error
	if req.SecurityAttributes != nil {
		securityAttributes, err = json.Marshal(req.SecurityAttributes)
		if err != nil {
			return db.CreatePersonParams{}, err
		}
	}
	if req.Metadata != nil {
		metadata, err = json.Marshal(req.Metadata)
		if err != nil {
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

func toSQLCCreateEmployeeParams(req *CreateEmployeeRequest) (db.CreateEmployeeParams, error) {
	var salaryInfo, workSchedule, accessAttributes []byte
	var err error
	if req.SalaryInfo != nil {
		salaryInfo, err = json.Marshal(req.SalaryInfo)
		if err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}
	if req.WorkSchedule != nil {
		workSchedule, err = json.Marshal(req.WorkSchedule)
		if err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}
	if req.AccessAttributes != nil {
		accessAttributes, err = json.Marshal(req.AccessAttributes)
		if err != nil {
			return db.CreateEmployeeParams{}, err
		}
	}

	statusString := string(req.Status)
	return db.CreateEmployeeParams{
		PersonID:         req.PersonID,
		EmployeeNumber:   req.EmployeeNumber,
		EntityID:         req.EntityID,
		PositionTitle:    req.PositionTitle,
		DepartmentID:     req.DepartmentID,
		ManagerID:        req.ManagerID,
		HireDate:         req.HireDate,
		SalaryInfo:       salaryInfo,
		EmploymentStatus: &statusString,
		WorkSchedule:     workSchedule,
		SecurityLevel:    &req.SecurityLevel,
		AccessAttributes: accessAttributes,
	}, nil
}
