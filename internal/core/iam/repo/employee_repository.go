package repo

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// employeeRepository implements EmployeeRepository interface
type employeeRepository struct {
	store   db.Store
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(
	store db.Store,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) EmployeeRepository {
	return &employeeRepository{
		store:   store,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// Create creates a new employee
func (r *employeeRepository) Create(ctx context.Context, employee *model.Employee) (*model.Employee, error) {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.Create")
	defer span.End()

	var createdEmployee *model.Employee
	err := r.store.WithTenant(ctx, employee.TenantID, func(ctx context.Context, store db.Store) error {
		// Convert domain model to SQLC params
		salaryInfoJSON, _ := json.Marshal(employee.SalaryInfo)
		workScheduleJSON, _ := json.Marshal(employee.WorkSchedule)
		accessAttrsJSON, _ := json.Marshal(employee.AccessAttributes)

		params := db.CreateEmployeeParams{
			PersonID:         employee.PersonID,
			EmployeeNumber:   employee.EmployeeNumber,
			EntityID:         employee.EntityID,
			PositionTitle:    employee.PositionTitle,
			DepartmentID:     employee.DepartmentID,
			ManagerID:        employee.ManagerID,
			HireDate:         employee.HireDate,
			SalaryInfo:       salaryInfoJSON,
			EmploymentStatus: stringPtr(string(employee.EmploymentStatus)),
			WorkSchedule:     workScheduleJSON,
			SecurityLevel: func() *int32 {
				// Validate bounds to prevent integer overflow
				if employee.SecurityLevel > 2147483647 || employee.SecurityLevel < -2147483648 {
					return nil // Return nil for invalid values rather than overflow
				}
				// #nosec G115 - Safe conversion after bounds check
				return int32Ptr(int32(employee.SecurityLevel))
			}(),
			AccessAttributes: accessAttrsJSON,
		}

		// Create employee using SQLC
		dbEmployee, err := store.CreateEmployee(ctx, params)
		if err != nil {
			// Handle database-specific errors
			if isDuplicateKeyError(err) {
				return errors.NewBusinessError("EMPLOYEE_NUMBER_EXISTS", "Employee with this number already exists")
			}
			return fmt.Errorf("failed to create employee: %w", err)
		}

		// Convert back to domain model
		createdEmployee = convertEmployeeToDomain(dbEmployee)

		r.metrics.IncrementCounter("employee_repository_create_success", nil)
		return nil
	})

	if err != nil {
		r.tracer.RecordError(ctx, err, tracing.WithErrorStatus())
		r.metrics.IncrementCounter("employee_repository_create_error", nil)
		return nil, err
	}

	r.logger.InfoContext(ctx, "Employee created successfully",
		logger.Fields{"employee_id": createdEmployee.ID, "tenant_id": employee.TenantID})

	return createdEmployee, nil
}

// GetByID retrieves an employee by ID
func (r *employeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error) {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.GetByID")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("GetByID not implemented")
}

// GetByEmployeeNumber retrieves an employee by employee number
func (r *employeeRepository) GetByEmployeeNumber(ctx context.Context, employeeNumber string) (*model.Employee, error) {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.GetByEmployeeNumber")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("GetByEmployeeNumber not implemented")
}

// GetByPersonID retrieves an employee by person ID
func (r *employeeRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) (*model.Employee, error) {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.GetByPersonID")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("GetByPersonID not implemented")
}

// Update updates an existing employee
func (r *employeeRepository) Update(ctx context.Context, employee *model.Employee) (*model.Employee, error) {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.Update")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("Update not implemented")
}

// Delete soft deletes an employee
func (r *employeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.Delete")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return fmt.Errorf("Delete not implemented")
}

// List lists employees with pagination
func (r *employeeRepository) List(ctx context.Context, limit, offset int) ([]*model.Employee, error) {
	ctx, span := r.tracer.StartSpan(ctx, "employee_repository.List")
	defer span.End()

	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("List not implemented")
}

// ListByStatus lists employees by employment status
func (r *employeeRepository) ListByStatus(ctx context.Context, status model.EmploymentStatus) ([]*model.Employee, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("ListByStatus not implemented")
}

// ListByDepartment lists employees by department
func (r *employeeRepository) ListByDepartment(ctx context.Context, department string) ([]*model.Employee, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("ListByDepartment not implemented")
}

// ListByManager lists employees by manager
func (r *employeeRepository) ListByManager(ctx context.Context, managerID uuid.UUID) ([]*model.Employee, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("ListByManager not implemented")
}

// Count returns total number of employees
func (r *employeeRepository) Count(ctx context.Context) (int64, error) {
	// TODO: Implement when SQLC query is available
	return 0, fmt.Errorf("Count not implemented")
}

// GetDirectReports retrieves direct reports for a manager
func (r *employeeRepository) GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]*model.Employee, error) {
	// TODO: Implement when SQLC query is available
	return nil, fmt.Errorf("GetDirectReports not implemented")
}

// UpdateEmploymentStatus updates employee employment status
func (r *employeeRepository) UpdateEmploymentStatus(ctx context.Context, employeeID uuid.UUID, status model.EmploymentStatus) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("UpdateEmploymentStatus not implemented")
}

// UpdateManager updates employee's manager
func (r *employeeRepository) UpdateManager(ctx context.Context, employeeID uuid.UUID, managerID *uuid.UUID) error {
	// TODO: Implement when SQLC query is available
	return fmt.Errorf("UpdateManager not implemented")
}

// Helper function to convert SQLC Employee to domain model
func convertEmployeeToDomain(dbEmployee *db.Employee) *model.Employee {
	employee := &model.Employee{
		ID:               dbEmployee.ID,
		TenantID:         dbEmployee.TenantID,
		PersonID:         dbEmployee.PersonID,
		EmployeeNumber:   dbEmployee.EmployeeNumber,
		EntityID:         dbEmployee.EntityID,
		PositionTitle:    dbEmployee.PositionTitle,
		DepartmentID:     dbEmployee.DepartmentID,
		ManagerID:        dbEmployee.ManagerID,
		HireDate:         dbEmployee.HireDate,
		TerminationDate:  convertTimePointer(dbEmployee.TerminationDate),
		EmploymentStatus: model.EmploymentStatus(getStringValue(dbEmployee.EmploymentStatus)),
		SecurityLevel:    int(getInt32Value(dbEmployee.SecurityLevel)),
		CreatedAt:        dbEmployee.CreatedAt,
		UpdatedAt:        dbEmployee.UpdatedAt,
	}

	// Unmarshal JSON fields - silently ignore errors for non-critical fields
	if dbEmployee.SalaryInfo != nil {
		_ = json.Unmarshal(dbEmployee.SalaryInfo, &employee.SalaryInfo)
	}
	if dbEmployee.WorkSchedule != nil {
		_ = json.Unmarshal(dbEmployee.WorkSchedule, &employee.WorkSchedule)
	}
	if dbEmployee.AccessAttributes != nil {
		_ = json.Unmarshal(dbEmployee.AccessAttributes, &employee.AccessAttributes)
	}

	return employee
}

// Helper function to convert sql.NullTime to *time.Time
func convertTimePointer(nullTime any) *time.Time {
	// TODO: Implement proper conversion based on the actual type
	return nil
}
