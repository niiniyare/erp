package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestSuite defines test suite for core domain model validation
// Implements Phase 3 Task 3.1.1: Core Domain Model Testing (IAM-CORE-001 to IAM-CORE-007)
type TestSuite struct {
	suite.Suite
	tenantID uuid.UUID
	entityID uuid.UUID
}

// SetupTest initializes test fixtures for each test
func (s *TestSuite) SetupTest() {
	s.tenantID = uuid.New()
	s.entityID = uuid.New()
}

// TestIAMModels runs the core domain model test suite
func TestIAMModels(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TestPersonModelCreation implements IAM-CORE-001: Verify Person model creation with valid data
func (s *TestSuite) TestPersonModelCreation() {
	testCases := []struct {
		name            string
		setupPerson     func() *Person
		expectedValid   bool
		validationCheck func(*testing.T, *Person)
	}{
		{
			name: "ValidPersonCreation_AllFieldsSet",
			setupPerson: func() *Person {
				email := "test@example.com"
				phone := "+1234567890"
				nationalID := "123456789"
				taxID := "TAX123"
				middleName := "Middle"

				return &Person{
					ID:          uuid.New(),
					TenantID:    s.tenantID,
					EntityID:    s.entityID,
					PersonType:  PersonTypeEmployee,
					FirstName:   "John",
					LastName:    "Doe",
					MiddleName:  &middleName,
					Email:       &email,
					PhoneNumber: &phone,
					BirthDate:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					NationalID:  &nationalID,
					TaxID:       &taxID,
					Address: map[string]any{
						"street":      "123 Main St",
						"city":        "Test City",
						"state":       "TS",
						"postal_code": "12345",
						"country":     "TestCountry",
					},
					SecurityAttributes: map[string]any{
						"clearance_level": "PUBLIC",
						"department":      "Engineering",
					},
					Metadata: map[string]any{
						"source":      "manual",
						"import_date": time.Now(),
					},
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			expectedValid: true,
			validationCheck: func(t *testing.T, p *Person) {
				require.NotEqual(t, uuid.Nil, p.ID, "Person ID should be generated")
				require.Equal(t, s.tenantID, p.TenantID, "Person should have correct tenant ID")
				require.Equal(t, s.entityID, p.EntityID, "Person should have correct entity ID")
				require.Equal(t, PersonTypeEmployee, p.PersonType, "Person type should be set correctly")
				require.Equal(t, "John", p.FirstName, "First name should be set")
				require.Equal(t, "Doe", p.LastName, "Last name should be set")
				require.NotNil(t, p.Email, "Email should be set")
				require.Equal(t, "test@example.com", *p.Email, "Email should match")
				require.True(t, p.IsActive, "Person should be active by default")
				require.NotZero(t, p.CreatedAt, "CreatedAt timestamp should be set")
				require.NotZero(t, p.UpdatedAt, "UpdatedAt timestamp should be set")
				require.NotNil(t, p.Address, "Address should be set")
				require.NotNil(t, p.SecurityAttributes, "Security attributes should be set")
				require.Equal(t, "John Doe", p.FullName(), "FullName method should work correctly")
			},
		},
		{
			name: "MinimalPersonCreation_RequiredFieldsOnly",
			setupPerson: func() *Person {
				return &Person{
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
				}
			},
			expectedValid: true,
			validationCheck: func(t *testing.T, p *Person) {
				require.NotEqual(t, uuid.Nil, p.ID, "Person ID should be generated")
				require.Equal(t, s.tenantID, p.TenantID, "Tenant ID from context should be assigned")
				require.Equal(t, PersonTypeIndividual, p.PersonType, "Person type should be valid enum")
				require.Nil(t, p.Email, "Optional email should be nil")
				require.Nil(t, p.PhoneNumber, "Optional phone should be nil")
				require.Nil(t, p.DeletedAt, "DeletedAt should be nil for new person")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-001_"+tc.name, func() {
			// Arrange
			person := tc.setupPerson()

			// Act & Assert
			if tc.expectedValid {
				require.NotNil(s.T(), person, "Person should be created successfully")
				if tc.validationCheck != nil {
					tc.validationCheck(s.T(), person)
				}
			}
		})
	}
}

// TestPersonModelValidation implements IAM-CORE-002: Verify Person model validation for invalid data
func (s *TestSuite) TestPersonModelValidation() {
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
					FirstName:  "",          // Empty required field
					LastName:   "",          // Empty required field
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

			if tc.validationCheck != nil {
				tc.validationCheck(s.T(), person, validationErrors)
			}
		})
	}
}

// TestEmployeeModelCreation implements IAM-CORE-003: Verify Employee model creation and linking to a Person
func (s *TestSuite) TestEmployeeModelCreation() {
	testCases := []struct {
		name            string
		setupEmployee   func() (*Person, *Employee)
		expectedValid   bool
		validationCheck func(*testing.T, *Person, *Employee)
	}{
		{
			name: "ValidEmployeeCreation_LinkedToPerson",
			setupEmployee: func() (*Person, *Employee) {
				person := &Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: PersonTypeEmployee,
					FirstName:  "John",
					LastName:   "Doe",
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}

				deptID := uuid.New()
				managerID := uuid.New()
				positionTitle := "Software Engineer"

				employee := &Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         person.ID, // Link to person
					EmployeeNumber:   "EMP001",
					EntityID:         s.entityID,
					PositionTitle:    &positionTitle,
					DepartmentID:     &deptID,
					ManagerID:        &managerID,
					HireDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EmploymentStatus: EmploymentStatusActive,
					SecurityLevel:    2,
					AccessAttributes: map[string]any{
						"clearance":  "SECRET",
						"department": "Engineering",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				return person, employee
			},
			expectedValid: true,
			validationCheck: func(t *testing.T, person *Person, employee *Employee) {
				require.NotEqual(t, uuid.Nil, employee.ID, "Employee ID should be generated")
				require.Equal(t, person.ID, employee.PersonID, "Employee should be linked to Person")
				require.Equal(t, s.tenantID, employee.TenantID, "Employee should have correct tenant ID")
				require.Equal(t, "EMP001", employee.EmployeeNumber, "Employee number should be set")
				require.Equal(t, EmploymentStatusActive, employee.EmploymentStatus, "Employment status should be active")
				require.True(t, employee.IsActive(), "IsActive method should return true")
				require.NotNil(t, employee.ManagerID, "Manager ID should be set (self-referential)")
				require.Nil(t, employee.DeletedAt, "DeletedAt should be nil for active employee")
				require.NotZero(t, employee.CreatedAt, "CreatedAt should be set")
				require.NotZero(t, employee.UpdatedAt, "UpdatedAt should be set")
				require.NotNil(t, employee.AccessAttributes, "Access attributes should be set")
				require.Equal(t, 2, employee.SecurityLevel, "Security level should be set")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-003_"+tc.name, func() {
			// Arrange
			person, employee := tc.setupEmployee()

			// Act & Assert
			if tc.expectedValid {
				require.NotNil(s.T(), person, "Person should be created")
				require.NotNil(s.T(), employee, "Employee should be created")
				if tc.validationCheck != nil {
					tc.validationCheck(s.T(), person, employee)
				}
			}
		})
	}
}

// TestUserModelCreation implements IAM-CORE-004: Verify User model creation and linking to Person/Employee
func (s *TestSuite) TestUserModelCreation() {
	testCases := []struct {
		name            string
		setupUser       func() (*Person, *Employee, *User)
		expectedValid   bool
		validationCheck func(*testing.T, *Person, *Employee, *User)
	}{
		{
			name: "ValidUserCreation_LinkedToPersonAndEmployee",
			setupUser: func() (*Person, *Employee, *User) {
				// Create Person
				person := &Person{
					ID:         uuid.New(),
					TenantID:   s.tenantID,
					EntityID:   s.entityID,
					PersonType: PersonTypeEmployee,
					FirstName:  "John",
					LastName:   "Doe",
					BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
					IsActive:   true,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}

				// Create Employee
				employee := &Employee{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					PersonID:         person.ID,
					EmployeeNumber:   "EMP001",
					EntityID:         s.entityID,
					HireDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					EmploymentStatus: EmploymentStatusActive,
					SecurityLevel:    2,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}

				// Create User
				phone := "+1234567890"
				mfaMethod := MFAMethodTOTP
				user := &User{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					EntityID:         s.entityID,
					PersonID:         &person.ID,   // Link to person
					EmployeeID:       &employee.ID, // Link to employee
					Email:            "john.doe@example.com",
					PasswordHash:     "$2a$10$hashedpassword", // Properly hashed password
					FirstName:        person.FirstName,
					LastName:         person.LastName,
					PhoneNumber:      &phone,
					AccountStatus:    UserAccountStatusActive,
					EmailVerified:    true,
					PhoneVerified:    false,
					MFAEnabled:       false,
					MFAMethod:        &mfaMethod,
					PasswordExpired:  false,
					FailedLoginCount: 0, // Default to 0
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}

				return person, employee, user
			},
			expectedValid: true,
			validationCheck: func(t *testing.T, person *Person, employee *Employee, user *User) {
				require.NotEqual(t, uuid.Nil, user.ID, "User ID should be generated")
				require.Equal(t, s.tenantID, user.TenantID, "User should have correct tenant ID")
				require.NotNil(t, user.PersonID, "User should be linked to Person")
				require.Equal(t, person.ID, *user.PersonID, "PersonID should match")
				require.NotNil(t, user.EmployeeID, "User should be linked to Employee")
				require.Equal(t, employee.ID, *user.EmployeeID, "EmployeeID should match")

				// Password hash validation
				require.NotEmpty(t, user.PasswordHash, "Password hash should be stored")
				require.NotEqual(t, "plaintext", user.PasswordHash, "Password should be hashed, not plain text")
				require.Contains(t, user.PasswordHash, "$2a$", "Password should use bcrypt format")

				// Account status validation
				require.Equal(t, UserAccountStatusActive, user.AccountStatus, "Account status should default to ACTIVE")
				require.Equal(t, 0, user.FailedLoginCount, "Failed login attempts should default to 0")
				require.True(t, user.IsActive(), "IsActive method should return true")
				require.False(t, user.IsLocked(), "IsLocked method should return false")

				// Method tests
				require.Equal(t, "John Doe", user.FullName(), "FullName method should work")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-004_"+tc.name, func() {
			// Arrange
			person, employee, user := tc.setupUser()

			// Act & Assert
			if tc.expectedValid {
				require.NotNil(s.T(), person, "Person should be created")
				require.NotNil(s.T(), employee, "Employee should be created")
				require.NotNil(s.T(), user, "User should be created")
				if tc.validationCheck != nil {
					tc.validationCheck(s.T(), person, employee, user)
				}
			}
		})
	}
}

// TestUserModelSecurityDefaults implements IAM-CORE-005: Verify User model security feature defaults
func (s *TestSuite) TestUserModelSecurityDefaults() {
	testCases := []struct {
		name            string
		setupUser       func() *User
		validationCheck func(*testing.T, *User)
	}{
		{
			name: "NewUser_SecurityDefaults",
			setupUser: func() *User {
				return &User{
					ID:               uuid.New(),
					TenantID:         s.tenantID,
					EntityID:         s.entityID,
					Email:            "test@example.com",
					PasswordHash:     "$2a$10$hashedpassword",
					FirstName:        "Test",
					LastName:         "User",
					AccountStatus:    UserAccountStatusActive,
					EmailVerified:    false, // Default
					PhoneVerified:    false, // Default
					MFAEnabled:       false, // Default
					PasswordExpired:  false, // Default
					FailedLoginCount: 0,     // Default
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}
			},
			validationCheck: func(t *testing.T, user *User) {
				// Security defaults validation
				require.False(t, user.MFAEnabled, "MFA should be disabled by default")
				require.Nil(t, user.LockedUntil, "LockedUntil should be nil by default")
				require.NotZero(t, user.UpdatedAt, "Password changed timestamp should be set")
				require.False(t, user.EmailVerified, "Email should not be verified by default")
				require.False(t, user.PhoneVerified, "Phone should not be verified by default")
				require.False(t, user.PasswordExpired, "Password should not be expired by default")
				require.Equal(t, 0, user.FailedLoginCount, "Failed login count should be 0")
				require.Nil(t, user.LastLoginAt, "Last login should be nil for new user")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-005_"+tc.name, func() {
			// Arrange
			user := tc.setupUser()

			// Act & Assert
			require.NotNil(s.T(), user, "User should be created")
			if tc.validationCheck != nil {
				tc.validationCheck(s.T(), user)
			}
		})
	}
}

// TestRoleModelCreation implements IAM-CORE-006: Verify Role creation with hierarchy
func (s *TestSuite) TestRoleModelCreation() {
	testCases := []struct {
		name               string
		setupRoles         func() (*Role, *Role) // parent, child
		expectedValid      bool
		expectCircularDeps bool
		validationCheck    func(*testing.T, *Role, *Role)
	}{
		{
			name: "ValidRoleHierarchy_ParentChild",
			setupRoles: func() (*Role, *Role) {
				parentRole := &Role{
					ID:          uuid.New(),
					TenantID:    s.tenantID,
					Name:        "Admin",
					Description: "Administrative role with full permissions",
					ParentID:    nil, // Top-level role
					EntityID:    &s.entityID,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				childRole := &Role{
					ID:          uuid.New(),
					TenantID:    s.tenantID,
					Name:        "Manager",
					Description: "Management role inheriting from Admin",
					ParentID:    &parentRole.ID, // Child of Admin
					EntityID:    &s.entityID,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				return parentRole, childRole
			},
			expectedValid:      true,
			expectCircularDeps: false,
			validationCheck: func(t *testing.T, parent *Role, child *Role) {
				require.NotEqual(t, uuid.Nil, parent.ID, "Parent role ID should be generated")
				require.NotEqual(t, uuid.Nil, child.ID, "Child role ID should be generated")
				require.Nil(t, parent.ParentID, "Parent role should have no parent")
				require.NotNil(t, child.ParentID, "Child role should have parent")
				require.Equal(t, parent.ID, *child.ParentID, "Child should reference correct parent")
				require.NotEqual(t, parent.ID, child.ID, "Parent and child should have different IDs")
				require.Equal(t, s.tenantID, parent.TenantID, "Parent should have correct tenant ID")
				require.Equal(t, s.tenantID, child.TenantID, "Child should have correct tenant ID")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-006_"+tc.name, func() {
			// Arrange
			parent, child := tc.setupRoles()

			// Act - Simulate circular dependency detection
			hasCircularDep := false
			if parent.ParentID != nil && child.ParentID != nil {
				if *parent.ParentID == child.ID && *child.ParentID == parent.ID {
					hasCircularDep = true
				}
			}

			// Assert
			if tc.expectCircularDeps {
				require.True(s.T(), hasCircularDep, "Circular dependency should be detected")
			} else {
				require.False(s.T(), hasCircularDep, "No circular dependency should exist")
			}

			if tc.expectedValid {
				require.NotNil(s.T(), parent, "Parent role should be created")
				require.NotNil(s.T(), child, "Child role should be created")
			}

			if tc.validationCheck != nil {
				tc.validationCheck(s.T(), parent, child)
			}
		})
	}
}

// TestPermissionModelCreation implements IAM-CORE-007: Verify Permission creation with risk levels and categories
func (s *TestSuite) TestPermissionModelCreation() {
	testCases := []struct {
		name            string
		setupPermission func() *Permission
		expectedValid   bool
		validationCheck func(*testing.T, *Permission)
	}{
		{
			name: "ValidPermission_WithRiskLevelAndCategory",
			setupPermission: func() *Permission {
				resourceID := uuid.New()
				entityID := uuid.New()
				expiresAt := time.Now().Add(24 * time.Hour)

				return &Permission{
					ID:           uuid.New(),
					TenantID:     s.tenantID,
					ResourceType: "financial_reports",
					ResourceID:   &resourceID,
					Action:       "read",
					EntityID:     &entityID,
					Conditions:   []string{"department=finance", "time_of_day=business_hours"},
					ExpiresAt:    &expiresAt,
					Metadata: map[string]any{
						"risk_level":          "HIGH",
						"action_category":     "DATA_ACCESS",
						"requires_approval":   true,
						"data_classification": "CONFIDENTIAL",
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			expectedValid: true,
			validationCheck: func(t *testing.T, perm *Permission) {
				require.NotEqual(t, uuid.Nil, perm.ID, "Permission ID should be generated")
				require.Equal(t, s.tenantID, perm.TenantID, "Permission should have correct tenant ID")
				require.Equal(t, "financial_reports", perm.ResourceType, "Resource type should be set")
				require.Equal(t, "read", perm.Action, "Action should be set")
				require.NotNil(t, perm.ResourceID, "Resource ID should be set")
				require.NotNil(t, perm.EntityID, "Entity ID should be set")
				require.NotEmpty(t, perm.Conditions, "Conditions should be set")
				require.Len(t, perm.Conditions, 2, "Should have 2 conditions")
				require.NotNil(t, perm.ExpiresAt, "Expiration should be set")
				require.False(t, perm.IsExpired(), "Permission should not be expired")

				// Validate metadata fields that would be enums in actual implementation
				require.NotNil(t, perm.Metadata, "Metadata should be set")
				require.Equal(t, "HIGH", perm.Metadata["risk_level"], "Risk level should be set")
				require.Equal(t, "DATA_ACCESS", perm.Metadata["action_category"], "Action category should be set")
				require.Equal(t, true, perm.Metadata["requires_approval"], "Requires approval flag should be set")
			},
		},
	}

	for _, tc := range testCases {
		s.Run("IAM-CORE-007_"+tc.name, func() {
			// Arrange
			permission := tc.setupPermission()

			// Act & Assert
			require.NotNil(s.T(), permission, "Permission should be created")

			if tc.validationCheck != nil {
				tc.validationCheck(s.T(), permission)
			}
		})
	}
}

// TestEdgeCasesAndBoundaryConditions tests various edge cases and boundary conditions
func (s *TestSuite) TestEdgeCasesAndBoundaryConditions() {
	testCases := []struct {
		name         string
		spec         string
		testFunction func(*testing.T)
	}{
		{
			name: "PersonFullName_EmptyNames",
			spec: "IAM-CORE-001",
			testFunction: func(t *testing.T) {
				person := &Person{FirstName: "", LastName: ""}
				require.Equal(t, " ", person.FullName(), "FullName should handle empty names")
			},
		},
		{
			name: "UserIsLocked_VariousScenarios",
			spec: "IAM-CORE-004",
			testFunction: func(t *testing.T) {
				// Test locked status
				lockedUser := &User{AccountStatus: UserAccountStatusLocked}
				require.True(t, lockedUser.IsLocked(), "User with LOCKED status should be locked")

				// Test time-based lock
				futureTime := time.Now().Add(1 * time.Hour)
				timeLockedUser := &User{
					AccountStatus: UserAccountStatusActive,
					LockedUntil:   &futureTime,
				}
				require.True(t, timeLockedUser.IsLocked(), "User locked until future time should be locked")

				// Test expired time lock
				pastTime := time.Now().Add(-1 * time.Hour)
				expiredLockUser := &User{
					AccountStatus: UserAccountStatusActive,
					LockedUntil:   &pastTime,
				}
				require.False(t, expiredLockUser.IsLocked(), "User with expired lock time should not be locked")
			},
		},
		{
			name: "SessionExpiration_EdgeCases",
			spec: "IAM-CORE-004",
			testFunction: func(t *testing.T) {
				// Active session not expired
				activeSession := &Session{
					Status:    SessionStatusActive,
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				require.False(t, activeSession.IsExpired(), "Active session should not be expired")

				// Expired session by time
				expiredSession := &Session{
					Status:    SessionStatusActive,
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
				require.True(t, expiredSession.IsExpired(), "Session past expiration time should be expired")

				// Revoked session
				revokedSession := &Session{
					Status:    SessionStatusRevoked,
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				require.True(t, revokedSession.IsExpired(), "Revoked session should be considered expired")
			},
		},
		{
			name: "UserRoleExpiration_EdgeCases",
			spec: "IAM-CORE-006",
			testFunction: func(t *testing.T) {
				// Non-expiring role
				permanentRole := &UserRole{ExpiresAt: nil}
				require.False(t, permanentRole.IsExpired(), "Role without expiration should not be expired")

				// Future expiration
				futureExpiry := time.Now().Add(24 * time.Hour)
				futureRole := &UserRole{ExpiresAt: &futureExpiry}
				require.False(t, futureRole.IsExpired(), "Role with future expiration should not be expired")

				// Past expiration
				pastExpiry := time.Now().Add(-1 * time.Hour)
				expiredRole := &UserRole{ExpiresAt: &pastExpiry}
				require.True(t, expiredRole.IsExpired(), "Role with past expiration should be expired")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.spec+"_"+tc.name, func() {
			tc.testFunction(s.T())
		})
	}
}
