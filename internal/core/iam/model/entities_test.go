package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// EntityTestSuite is the test suite for IAM domain models
type EntityTestSuite struct {
	suite.Suite
}

// TestIAMEntityModels runs the test suite
func TestIAMEntityModels(t *testing.T) {
	suite.Run(t, new(EntityTestSuite))
}

// TestPersonModelCreation covers test case IAM-CORE-001
func (s *EntityTestSuite) TestPersonModelCreation() {
	testCases := []struct {
		name          string
		personData    Person
		expectValid   bool
		checkResult   func(p *Person)
		expectedError string
	}{
		{
			name: "IAM-CORE-001: Valid Person Creation",
			personData: Person{
				ID:         uuid.New(),
				TenantID:   uuid.New(),
				EntityID:   uuid.New(),
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
				s.Require().NotEqual(uuid.Nil, p.ID)
				s.Require().NotEqual(uuid.Nil, p.TenantID)
				s.Require().NotEqual(uuid.Nil, p.EntityID)
				s.Require().Equal(PersonTypeEmployee, p.PersonType)
				s.Require().Equal("John", p.FirstName)
				s.Require().Equal("Doe", p.LastName)
				s.Require().NotNil(p.Email)
				s.Require().Equal("john.doe@example.com", *p.Email)
				s.Require().True(p.IsActive)
				s.Require().Equal("John Doe", p.FullName())
			},
		},
		{
			name: "Person with Optional Fields",
			personData: Person{
				ID:          uuid.New(),
				TenantID:    uuid.New(),
				EntityID:    uuid.New(),
				PersonType:  PersonTypeCustomer,
				FirstName:   "Jane",
				LastName:    "Smith",
				MiddleName:  stringPtr("Marie"),
				Email:       stringPtr("jane.smith@example.com"),
				PhoneNumber: stringPtr("+1-555-0123"),
				BirthDate:   time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
				NationalID:  stringPtr("123-45-6789"),
				TaxID:       stringPtr("TAX123456"),
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Person) {
				s.Require().Equal(PersonTypeCustomer, p.PersonType)
				s.Require().NotNil(p.MiddleName)
				s.Require().Equal("Marie", *p.MiddleName)
				s.Require().NotNil(p.PhoneNumber)
				s.Require().Equal("+1-555-0123", *p.PhoneNumber)
				s.Require().NotNil(p.NationalID)
				s.Require().Equal("123-45-6789", *p.NationalID)
				s.Require().NotNil(p.TaxID)
				s.Require().Equal("TAX123456", *p.TaxID)
			},
		},
		{
			name: "Person with Complex Data Fields",
			personData: Person{
				ID:         uuid.New(),
				TenantID:   uuid.New(),
				EntityID:   uuid.New(),
				PersonType: PersonTypeContractor,
				FirstName:  "Bob",
				LastName:   "Johnson",
				BirthDate:  time.Date(1980, 12, 25, 0, 0, 0, 0, time.UTC),
				Address: map[string]any{
					"street":      "123 Main St",
					"city":        "Anytown",
					"state":       "CA",
					"postal_code": "12345",
					"country":     "US",
				},
				SecurityAttributes: map[string]any{
					"clearance_level": "confidential",
					"department":      "engineering",
					"location":        "HQ",
				},
				Metadata: map[string]any{
					"hire_source":       "referral",
					"emergency_contact": "spouse",
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectValid: true,
			checkResult: func(p *Person) {
				s.Require().Equal(PersonTypeContractor, p.PersonType)
				s.Require().NotNil(p.Address)
				s.Require().Equal("123 Main St", p.Address["street"])
				s.Require().Equal("CA", p.Address["state"])
				s.Require().NotNil(p.SecurityAttributes)
				s.Require().Equal("confidential", p.SecurityAttributes["clearance_level"])
				s.Require().Equal("engineering", p.SecurityAttributes["department"])
				s.Require().NotNil(p.Metadata)
				s.Require().Equal("referral", p.Metadata["hire_source"])
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			person := tc.personData

			if tc.expectValid {
				s.Require().NotNil(&person)
				if tc.checkResult != nil {
					tc.checkResult(&person)
				}
			}
		})
	}
}

// TestEmployeeModelCreation covers test case IAM-CORE-003
func (s *EntityTestSuite) TestEmployeeModelCreation() {
	personID := uuid.New()
	entityID := uuid.New()
	tenantID := uuid.New()

	testCases := []struct {
		name         string
		employeeData Employee
		expectValid  bool
		checkResult  func(e *Employee)
	}{
		{
			name: "IAM-CORE-003: Valid Employee Creation with Person Link",
			employeeData: Employee{
				ID:               uuid.New(),
				TenantID:         tenantID,
				PersonID:         personID,
				EmployeeNumber:   "EMP001",
				EntityID:         entityID,
				PositionTitle:    stringPtr("Software Engineer"),
				HireDate:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				EmploymentStatus: EmploymentStatusActive,
				SecurityLevel:    3,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(e *Employee) {
				s.Require().NotEqual(uuid.Nil, e.ID)
				s.Require().Equal(tenantID, e.TenantID)
				s.Require().Equal(personID, e.PersonID)
				s.Require().Equal("EMP001", e.EmployeeNumber)
				s.Require().Equal(entityID, e.EntityID)
				s.Require().NotNil(e.PositionTitle)
				s.Require().Equal("Software Engineer", *e.PositionTitle)
				s.Require().Equal(EmploymentStatusActive, e.EmploymentStatus)
				s.Require().Equal(3, e.SecurityLevel)
				s.Require().True(e.IsActive())
				s.Require().Nil(e.DeletedAt)
			},
		},
		{
			name: "Employee with Manager Hierarchy",
			employeeData: Employee{
				ID:               uuid.New(),
				TenantID:         tenantID,
				PersonID:         personID,
				EmployeeNumber:   "EMP002",
				EntityID:         entityID,
				PositionTitle:    stringPtr("Senior Developer"),
				DepartmentID:     &entityID,
				ManagerID:        &personID,
				HireDate:         time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
				EmploymentStatus: EmploymentStatusActive,
				SecurityLevel:    4,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectValid: true,
			checkResult: func(e *Employee) {
				s.Require().NotNil(e.DepartmentID)
				s.Require().Equal(entityID, *e.DepartmentID)
				s.Require().NotNil(e.ManagerID)
				s.Require().Equal(personID, *e.ManagerID)
				s.Require().Equal(4, e.SecurityLevel)
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
		s.Run(tc.name, func() {
			employee := tc.employeeData

			if tc.expectValid {
				s.Require().NotNil(&employee)
				if tc.checkResult != nil {
					tc.checkResult(&employee)
				}
			}
		})
	}
}

// TestUserModelCreation covers test case IAM-CORE-004 and IAM-CORE-005
func (s *EntityTestSuite) TestUserModelCreation() {
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
func (s *EntityTestSuite) TestRoleModelCreation() {
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
