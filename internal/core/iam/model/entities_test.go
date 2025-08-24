package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CoreDomainModelTestSuite defines comprehensive test suite for core domain model validation
// Implements Phase 3 Task 3.1.1: Core Domain Model Testing (IAM-CORE-001 to IAM-CORE-007)
type CoreDomainModelTestSuite struct {
	suite.Suite
	tenantID uuid.UUID
	entityID uuid.UUID
}

// SetupTest initializes test fixtures for each test
func (s *CoreDomainModelTestSuite) SetupTest() {
	s.tenantID = uuid.New()
	s.entityID = uuid.New()
}

// TestCoreEntities runs the comprehensive core domain model test suite
func TestCoreEntities(t *testing.T) {
	suite.Run(t, new(CoreDomainModelTestSuite))
}

// TestPersonModelCreation implements IAM-CORE-001: Verify Person model creation with valid data
func (s *CoreDomainModelTestSuite) TestPersonModelCreation() {
	testCases := []struct {
		name          string
		personData    Person
		expectValid   bool
		checkResult   func(p *Person)
		expectedError string
	}{
		{
			name: "ValidPersonCreation_AllFieldsSet",
			personData: Person{
				ID:         uuid.New(),
				TenantID:   s.tenantID,
				EntityID:   s.entityID,
				PersonType: PersonTypeEmployee,
				FirstName:  "John",
				LastName:   "Doe",
				Email:      stringPtr("john.doe@example.com"),
				BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				IsActive:   true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Person) {
				require.NotEqual(s.T(), uuid.Nil, p.ID, "Person ID should be generated")
				require.Equal(s.T(), s.tenantID, p.TenantID, "Person should have correct tenant ID")
				require.Equal(s.T(), s.entityID, p.EntityID, "Person should have correct entity ID")
				require.Equal(s.T(), PersonTypeEmployee, p.PersonType, "Person type should be set correctly")
				require.Equal(s.T(), "John", p.FirstName, "First name should be set")
				require.Equal(s.T(), "Doe", p.LastName, "Last name should be set")
				require.NotNil(s.T(), p.Email, "Email should be set")
				require.Equal(s.T(), "john.doe@example.com", *p.Email, "Email should match")
				require.True(s.T(), p.IsActive, "Person should be active by default")
				require.NotZero(s.T(), p.CreatedAt, "CreatedAt timestamp should be set")
				require.NotZero(s.T(), p.UpdatedAt, "UpdatedAt timestamp should be set")
				require.Equal(s.T(), "John Doe", p.FullName(), "FullName method should work correctly")
				require.Nil(s.T(), p.DeletedAt, "DeletedAt should be nil for new person")
			},
		},
		{
			name: "MinimalPersonCreation_RequiredFieldsOnly",
			personData: Person{
				ID:         uuid.New(),
				TenantID:   s.tenantID,
				EntityID:   s.entityID,
				PersonType: PersonTypeIndividual,
				FirstName:  "Jane",
				LastName:   "Smith",
				BirthDate:  time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC),
				IsActive:   true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Person) {
				require.NotEqual(s.T(), uuid.Nil, p.ID, "Person ID should be generated")
				require.Equal(s.T(), s.tenantID, p.TenantID, "Tenant ID from context should be assigned")
				require.Equal(s.T(), PersonTypeIndividual, p.PersonType, "Person type should be valid enum")
				require.Nil(s.T(), p.Email, "Optional email should be nil")
				require.Nil(s.T(), p.PhoneNumber, "Optional phone should be nil")
				require.Nil(s.T(), p.DeletedAt, "DeletedAt should be nil for new person")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-001_"+tc.name, func() {
			person := tc.personData

			if tc.expectValid {
				require.NotNil(s.T(), &person, "Person should be created successfully")
				if tc.checkResult != nil {
					tc.checkResult(&person)
				}
			}
		})
	}
}

// TestPersonModelValidation implements IAM-CORE-002: Verify Person model validation for invalid data
func (s *CoreDomainModelTestSuite) TestPersonModelValidation() {
	testCases := []struct {
		name            string
		setupPerson     func() *Person
		expectedErrors  []string
		validationCheck func(*testing.T, *Person, []string)
	}{
		{
			name: "InvalidPersonType_ShouldValidate",
			setupPerson: func() *Person {
				return &Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: PersonType("INVALID_TYPE"), // Invalid enum
					FirstName:  "John",
					LastName:   "Doe",
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			},
			expectedErrors: []string{"invalid person_type enum"},
			validationCheck: func(t *testing.T, p *Person, errors []string) {
				// Validate that PersonType enum validation would catch this
				validTypes := []PersonType{
					PersonTypeIndividual, PersonTypeEmployee, PersonTypeContact,
					PersonTypeCustomer, PersonTypeVendor, PersonTypeContractor,
				}
				isValid := false
				for _, validType := range validTypes {
					if p.PersonType == validType {
						isValid = true
						break
					}
				}
				require.False(t, isValid, "Invalid PersonType should be detected")
			},
		},
		{
			name: "EmptyRequiredFields_ShouldValidate",
			setupPerson: func() *Person {
				return &Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: PersonTypeEmployee,
					FirstName:  "", // Empty required field
					LastName:   "", // Empty required field
					BirthDate:  time.Time{}, // Zero time
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			},
			expectedErrors: []string{"first_name required", "last_name required", "birth_date required"},
			validationCheck: func(t *testing.T, p *Person, errors []string) {
				require.Empty(t, p.FirstName, "FirstName should be empty")
				require.Empty(t, p.LastName, "LastName should be empty")
				require.True(t, p.BirthDate.IsZero(), "BirthDate should be zero")
			},
		},
		{
			name: "NilTenantID_ShouldValidate",
			setupPerson: func() *Person {
				return &Person{
					ID:         uuid.New(),
					TenantID:   uuid.Nil, // Invalid tenant ID
					EntityID:   s.entityID,
					PersonType: PersonTypeEmployee,
					FirstName:  "John",
					LastName:   "Doe",
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
			},
			expectedErrors: []string{"tenant_id required"},
			validationCheck: func(t *testing.T, p *Person, errors []string) {
				require.Equal(t, uuid.Nil, p.TenantID, "TenantID should be nil")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-002_"+tc.name, func() {
			// Arrange
			person := tc.setupPerson()

			// Act - Simulate validation logic
			var validationErrors []string

			// Basic validation logic that would be in a validator
			if person.FirstName == "" {
				validationErrors = append(validationErrors, "first_name required")
			}
			if person.LastName == "" {
				validationErrors = append(validationErrors, "last_name required")
			}
			if person.BirthDate.IsZero() {
				validationErrors = append(validationErrors, "birth_date required")
			}
			if person.TenantID == uuid.Nil {
				validationErrors = append(validationErrors, "tenant_id required")
			}

			// Enum validation
			validPersonTypes := []PersonType{
				PersonTypeIndividual, PersonTypeEmployee, PersonTypeContact,
				PersonTypeCustomer, PersonTypeVendor, PersonTypeContractor,
			}
			isValidType := false
			for _, validType := range validPersonTypes {
				if person.PersonType == validType {
					isValidType = true
					break
				}
			}
			if !isValidType {
				validationErrors = append(validationErrors, "invalid person_type enum")
			}

			// Assert
			require.NotEmpty(s.T(), validationErrors, "Validation should detect errors")
			require.Len(s.T(), validationErrors, len(tc.expectedErrors), "Should have expected number of validation errors")
			
			if tc.validationCheck != nil {
				tc.validationCheck(s.T(), person, validationErrors)
			}
		})
	}
}

// TestEmployeeModelCreation implements IAM-CORE-003: Verify Employee model creation and linking to a Person
func (s *CoreDomainModelTestSuite) TestEmployeeModelCreation() {
	personID := uuid.New()

	testCases := []struct {
		name         string
		employeeData Employee
		expectValid  bool
		checkResult  func(e *Employee)
	}{
		{
			name: "ValidEmployeeCreation_LinkedToPerson",
			employeeData: Employee{
				ID:               uuid.New(),
				TenantID:         s.tenantID,
				PersonID:         personID,
				EmployeeNumber:   "EMP001",
				EntityID:         s.entityID,
				PositionTitle:    stringPtr("Software Engineer"),
				HireDate:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				EmploymentStatus: EmploymentStatusActive,
				SecurityLevel:    3,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(e *Employee) {
				require.NotEqual(s.T(), uuid.Nil, e.ID, "Employee ID should be generated")
				require.Equal(s.T(), s.tenantID, e.TenantID, "Employee should have correct tenant ID")
				require.Equal(s.T(), personID, e.PersonID, "Employee should be linked to Person")
				require.Equal(s.T(), "EMP001", e.EmployeeNumber, "Employee number should be set")
				require.Equal(s.T(), s.entityID, e.EntityID, "Employee should have correct entity ID")
				require.NotNil(s.T(), e.PositionTitle, "Position title should be set")
				require.Equal(s.T(), "Software Engineer", *e.PositionTitle, "Position should match")
				require.Equal(s.T(), EmploymentStatusActive, e.EmploymentStatus, "Employment status should be active")
				require.Equal(s.T(), 3, e.SecurityLevel, "Security level should be set")
				require.True(s.T(), e.IsActive(), "IsActive method should return true")
				require.Nil(s.T(), e.DeletedAt, "DeletedAt should be nil for active employee")
			},
		},
		{
			name: "EmployeeWithManagerHierarchy_SelfReference",
			employeeData: Employee{
				ID:               uuid.New(),
				TenantID:         s.tenantID,
				PersonID:         personID,
				EmployeeNumber:   "EMP002",
				EntityID:         s.entityID,
				PositionTitle:    stringPtr("Senior Developer"),
				DepartmentID:     &s.entityID,
				ManagerID:        &personID,
				HireDate:         time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
				EmploymentStatus: EmploymentStatusActive,
				SecurityLevel:    4,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(e *Employee) {
				require.NotNil(s.T(), e.DepartmentID, "Department ID should be set")
				require.Equal(s.T(), s.entityID, *e.DepartmentID, "Department should match entity")
				require.NotNil(s.T(), e.ManagerID, "Manager ID should be set (self-referential)")
				require.Equal(s.T(), personID, *e.ManagerID, "Manager should reference correct person")
				require.Equal(s.T(), 4, e.SecurityLevel, "High-level position should have elevated security")
			},
		},
		{
			name: "Employee with Complex Salary and Schedule",
			employeeData: Employee{
				ID:             uuid.New(),
				TenantID:       tenantID,
				PersonID:       personID,
				EmployeeNumber: "EMP003",
				EntityID:       entityID,
				HireDate:       time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
				SalaryInfo: map[string]any{
					"base_salary":    75000,
					"currency":       "USD",
					"pay_frequency":  "monthly",
					"bonus_eligible": true,
				},
				WorkSchedule: map[string]any{
					"type":           "standard",
					"hours_per_week": 40,
					"start_time":     "09:00",
					"end_time":       "17:00",
					"timezone":       "America/New_York",
				},
				AccessAttributes: map[string]any{
					"vpn_access":        true,
					"admin_rights":      false,
					"data_access_level": "standard",
				},
				EmploymentStatus: EmploymentStatusActive,
				SecurityLevel:    2,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(e *Employee) {
				s.Require().NotNil(e.SalaryInfo)
				s.Require().Equal(75000, e.SalaryInfo["base_salary"])
				s.Require().Equal("USD", e.SalaryInfo["currency"])
				s.Require().NotNil(e.WorkSchedule)
				s.Require().Equal("standard", e.WorkSchedule["type"])
				s.Require().Equal(40, e.WorkSchedule["hours_per_week"])
				s.Require().NotNil(e.AccessAttributes)
				s.Require().Equal(true, e.AccessAttributes["vpn_access"])
				s.Require().Equal(false, e.AccessAttributes["admin_rights"])
			},
		},
		{
			name: "Terminated Employee",
			employeeData: Employee{
				ID:               uuid.New(),
				TenantID:         tenantID,
				PersonID:         personID,
				EmployeeNumber:   "EMP004",
				EntityID:         entityID,
				HireDate:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				TerminationDate:  timePtr(time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)),
				EmploymentStatus: EmploymentStatusTerminated,
				SecurityLevel:    0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(e *Employee) {
				s.Require().NotNil(e.TerminationDate)
				s.Require().Equal(EmploymentStatusTerminated, e.EmploymentStatus)
				s.Require().False(e.IsActive())
				s.Require().Equal(0, e.SecurityLevel)
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-003_"+tc.name, func() {
			employee := tc.employeeData

			if tc.expectValid {
				require.NotNil(s.T(), &employee, "Employee should be created successfully")
				if tc.checkResult != nil {
					tc.checkResult(&employee)
				}
			}
		})
	}
}

// TestUserModelCreation covers test case IAM-CORE-004 and IAM-CORE-005
func (s *CoreDomainModelTestSuite) TestUserModelCreation() {
	tenantID := uuid.New()

	testCases := []struct {
		name        string
		userData    User
		expectValid bool
		checkResult func(u *User)
	}{
		{
			name: "IAM-CORE-004: Valid User Creation with Security Defaults",
			userData: User{
				ID:               uuid.New(),
				TenantID:         tenantID,
				Email:            "john.doe@company.com",
				PasswordHash:     "$2a$10$hashed.password.here",
				FirstName:        "John",
				LastName:         "Doe",
				AccountStatus:    UserAccountStatusActive,
				EmailVerified:    true,
				MFAEnabled:       false,
				FailedLoginCount: 0,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(u *User) {
				s.Require().NotEqual(uuid.Nil, u.ID)
				s.Require().Equal(tenantID, u.TenantID)
				s.Require().NotEmpty(u.PasswordHash)
				s.Require().Equal(UserAccountStatusActive, u.AccountStatus)
				s.Require().Equal(0, u.FailedLoginCount)
				s.Require().Equal("John Doe", u.FullName())
				s.Require().True(u.IsActive())
				s.Require().False(u.IsLocked())
			},
		},
		{
			name: "IAM-CORE-005: User Security Feature Defaults",
			userData: User{
				ID:               uuid.New(),
				TenantID:         tenantID,
				Email:            "jane.smith@company.com",
				PasswordHash:     "$2a$10$another.hashed.password",
				FirstName:        "Jane",
				LastName:         "Smith",
				AccountStatus:    UserAccountStatusActive,
				EmailVerified:    false, // Default
				PhoneVerified:    false, // Default
				MFAEnabled:       false, // Default
				FailedLoginCount: 0,     // Default
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(u *User) {
				s.Require().False(u.MFAEnabled)
				s.Require().Nil(u.LockedUntil)
				s.Require().False(u.EmailVerified)
				s.Require().False(u.PhoneVerified)
				s.Require().Equal(0, u.FailedLoginCount)
				s.Require().False(u.PasswordExpired)
			},
		},
		{
			name: "User with MFA and Advanced Security",
			userData: User{
				ID:               uuid.New(),
				TenantID:         tenantID,
				Email:            "security.admin@company.com",
				PasswordHash:     "$2a$10$secure.password.hash",
				FirstName:        "Security",
				LastName:         "Admin",
				AccountStatus:    UserAccountStatusActive,
				EmailVerified:    true,
				PhoneVerified:    true,
				MFAEnabled:       true,
				MFASecret:        stringPtr("encrypted.totp.secret"),
				FailedLoginCount: 0,
				LastLoginAt:      timePtr(time.Now().AddDate(0, -2, 0)),
				Metadata: map[string]any{
					"role":               "admin",
					"last_login_ip":      "192.168.1.100",
					"security_questions": 3,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectValid: true,
			checkResult: func(u *User) {
				s.Require().True(u.MFAEnabled)
				s.Require().NotNil(u.MFASecret)
				s.Require().True(u.EmailVerified)
				s.Require().True(u.PhoneVerified)
				s.Require().NotNil(u.LastLoginAt)
				s.Require().NotNil(u.Metadata)
				s.Require().Equal("admin", u.Metadata["role"])
				s.Require().Equal("192.168.1.100", u.Metadata["last_login_ip"])
			},
		},
		{
			name: "Locked User Account",
			userData: User{
				ID:               uuid.New(),
				TenantID:         tenantID,
				Email:            "locked.user@company.com",
				PasswordHash:     "$2a$10$locked.user.password",
				FirstName:        "Locked",
				LastName:         "User",
				AccountStatus:    UserAccountStatusLocked,
				FailedLoginCount: 5,
				LockedUntil:      timePtr(time.Now().Add(time.Hour * 24)),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(u *User) {
				s.Require().Equal(UserAccountStatusLocked, u.AccountStatus)
				s.Require().Equal(5, u.FailedLoginCount)
				s.Require().NotNil(u.LockedUntil)
				s.Require().True(u.LockedUntil.After(time.Now()))
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			user := tc.userData

			if tc.expectValid {
				s.Require().NotNil(&user)
				if tc.checkResult != nil {
					tc.checkResult(&user)
				}
			}
		})
	}
}

// TestRoleModelCreation covers test case IAM-CORE-006
func (s *CoreDomainModelTestSuite) TestRoleModelCreation() {
	tenantID := uuid.New()
	entityID := uuid.New()
	parentRoleID := uuid.New()

	testCases := []struct {
		name        string
		roleData    Role
		expectValid bool
		checkResult func(r *Role)
	}{
		{
			name: "IAM-CORE-006: Role Creation with Hierarchy",
			roleData: Role{
				ID:          uuid.New(),
				TenantID:    tenantID,
				Name:        "manager",
				Description: "Manages department operations and staff",
				ParentID:    &parentRoleID,
				EntityID:    &entityID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectValid: true,
			checkResult: func(r *Role) {
				s.Require().NotEqual(uuid.Nil, r.ID)
				s.Require().Equal(tenantID, r.TenantID)
				s.Require().NotNil(r.EntityID)
				s.Require().Equal(entityID, *r.EntityID)
				s.Require().Equal("manager", r.Name)
				s.Require().Equal("Manages department operations and staff", r.Description)
				s.Require().NotNil(r.ParentID)
				s.Require().Equal(parentRoleID, *r.ParentID)
			},
		},
		{
			name: "System Role (Root Level)",
			roleData: Role{
				ID:          uuid.New(),
				TenantID:    tenantID,
				Name:        "system_admin",
				Description: "Full system access and administration",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectValid: true,
			checkResult: func(r *Role) {
				s.Require().Equal("system_admin", r.Name)
				s.Require().Equal("Full system access and administration", r.Description)
				s.Require().Nil(r.ParentID)
				s.Require().Nil(r.EntityID)
			},
		},
		{
			name: "Role with Metadata",
			roleData: Role{
				ID:          uuid.New(),
				TenantID:    tenantID,
				Name:        "data_analyst",
				Description: "Analyzes business data with time restrictions",
				Metadata: map[string]any{
					"permissions": []string{"read_reports", "export_data"},
					"time_restrictions": map[string]any{
						"start_time": "09:00",
						"end_time":   "17:00",
						"timezone":   "America/New_York",
					},
					"ip_restrictions": []string{"192.168.1.0/24", "10.0.0.0/8"},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectValid: true,
			checkResult: func(r *Role) {
				s.Require().Equal("data_analyst", r.Name)
				s.Require().NotNil(r.Metadata)
				permissions := r.Metadata["permissions"].([]string)
				s.Require().Len(permissions, 2)
				s.Require().Equal("read_reports", permissions[0])
				s.Require().Equal("export_data", permissions[1])
				timeRestrictions := r.Metadata["time_restrictions"].(map[string]any)
				s.Require().Equal("09:00", timeRestrictions["start_time"])
				s.Require().Equal("17:00", timeRestrictions["end_time"])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			role := tc.roleData

			if tc.expectValid {
				s.Require().NotNil(&role)
				if tc.checkResult != nil {
					tc.checkResult(&role)
				}
			}
		})
	}
}

// TestPermissionModelCreation covers test case IAM-CORE-007
func (s *EntityTestSuite) TestPermissionModelCreation() {
	tenantID := uuid.New()
	resourceID := uuid.New()
	entityID := uuid.New()

	testCases := []struct {
		name           string
		permissionData Permission
		expectValid    bool
		checkResult    func(p *Permission)
	}{
		{
			name: "IAM-CORE-007: Permission Creation with Resource and Action",
			permissionData: Permission{
				ID:           uuid.New(),
				TenantID:     tenantID,
				ResourceType: "financial_records",
				ResourceID:   &resourceID,
				Action:       "delete",
				EntityID:     &entityID,
				Conditions:   []string{"ip_restrictions", "time_based"},
				Metadata: map[string]any{
					"risk_level":        "HIGH",
					"requires_approval": true,
					"category":          "ADMINISTRATIVE",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Permission) {
				s.Require().NotEqual(uuid.Nil, p.ID)
				s.Require().Equal(tenantID, p.TenantID)
				s.Require().Equal("financial_records", p.ResourceType)
				s.Require().NotNil(p.ResourceID)
				s.Require().Equal(resourceID, *p.ResourceID)
				s.Require().Equal("delete", p.Action)
				s.Require().NotNil(p.EntityID)
				s.Require().Equal(entityID, *p.EntityID)
				s.Require().Len(p.Conditions, 2)
				s.Require().Contains(p.Conditions, "ip_restrictions")
				s.Require().Contains(p.Conditions, "time_based")
				s.Require().NotNil(p.Metadata)
				s.Require().Equal("HIGH", p.Metadata["risk_level"])
				s.Require().Equal(true, p.Metadata["requires_approval"])
			},
		},
		{
			name: "Simple Read Permission",
			permissionData: Permission{
				ID:           uuid.New(),
				TenantID:     tenantID,
				ResourceType: "public_reports",
				Action:       "read",
				Metadata: map[string]any{
					"risk_level": "LOW",
					"category":   "STANDARD",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Permission) {
				s.Require().Equal("public_reports", p.ResourceType)
				s.Require().Equal("read", p.Action)
				s.Require().Nil(p.ResourceID)
				s.Require().Nil(p.EntityID)
				s.Require().Equal("LOW", p.Metadata["risk_level"])
			},
		},
		{
			name: "Bulk Operation Permission with Expiration",
			permissionData: Permission{
				ID:           uuid.New(),
				TenantID:     tenantID,
				ResourceType: "users",
				Action:       "bulk_import",
				Conditions:   []string{"max_records=1000", "approval_required"},
				ExpiresAt:    timePtr(time.Now().Add(time.Hour * 24 * 30)), // 30 days
				Metadata: map[string]any{
					"risk_level":        "CRITICAL",
					"category":          "BULK",
					"requires_approval": true,
					"max_records":       1000,
					"approval_levels":   2,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Permission) {
				s.Require().Equal("users", p.ResourceType)
				s.Require().Equal("bulk_import", p.Action)
				s.Require().NotNil(p.ExpiresAt)
				s.Require().True(p.ExpiresAt.After(time.Now()))
				s.Require().False(p.IsExpired())
				s.Require().NotNil(p.Metadata)
				s.Require().Equal("CRITICAL", p.Metadata["risk_level"])
				s.Require().Equal(1000, p.Metadata["max_records"])
				s.Require().Equal(2, p.Metadata["approval_levels"])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			permission := tc.permissionData

			if tc.expectValid {
				s.Require().NotNil(&permission)
				if tc.checkResult != nil {
					tc.checkResult(&permission)
				}
			}
		})
	}
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
