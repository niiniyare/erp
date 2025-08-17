package model

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TypeTestSuite is the test suite for IAM domain types and enums
type TypeTestSuite struct {
	suite.Suite
}

// TestIAMTypes runs the test suite for types
func TestIAMTypes(t *testing.T) {
	suite.Run(t, new(TypeTestSuite))
}

// TestPolicyEffectEnum covers policy effect validation
func (s *TypeTestSuite) TestPolicyEffectEnum() {
	testCases := []struct {
		name        string
		effect      PolicyEffect
		expectValid bool
		stringValue string
	}{
		{
			name:        "Valid ALLOW Effect",
			effect:      PolicyEffectAllow,
			expectValid: true,
			stringValue: "ALLOW",
		},
		{
			name:        "Valid DENY Effect",
			effect:      PolicyEffectDeny,
			expectValid: true,
			stringValue: "DENY",
		},
		{
			name:        "Valid NOT_APPLICABLE Effect",
			effect:      PolicyEffectNotApplicable,
			expectValid: true,
			stringValue: "NOT_APPLICABLE",
		},
		{
			name:        "Invalid Effect",
			effect:      PolicyEffect("INVALID"),
			expectValid: false,
			stringValue: "INVALID",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.expectValid, tc.effect.IsValid())
			s.Require().Equal(tc.stringValue, tc.effect.String())
		})
	}
}

// TestUserAccountStatus covers user account status validation
func (s *TypeTestSuite) TestUserAccountStatus() {
	testCases := []struct {
		name   string
		status UserAccountStatus
		valid  bool
	}{
		{"Active Status", UserAccountStatusActive, true},
		{"Inactive Status", UserAccountStatusInactive, true},
		{"Locked Status", UserAccountStatusLocked, true},
		{"Suspended Status", UserAccountStatusSuspended, true},
		{"Pending Status", UserAccountStatusPending, true},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().NotEmpty(string(tc.status))
			// Test that status constants are properly defined
			switch tc.status {
			case UserAccountStatusActive:
				s.Require().Equal("ACTIVE", string(tc.status))
			case UserAccountStatusInactive:
				s.Require().Equal("INACTIVE", string(tc.status))
			case UserAccountStatusLocked:
				s.Require().Equal("LOCKED", string(tc.status))
			case UserAccountStatusSuspended:
				s.Require().Equal("SUSPENDED", string(tc.status))
			case UserAccountStatusPending:
				s.Require().Equal("PENDING", string(tc.status))
			}
		})
	}
}

// TestEmploymentStatus covers employment status validation
func (s *TypeTestSuite) TestEmploymentStatus() {
	testCases := []struct {
		name   string
		status EmploymentStatus
		valid  bool
	}{
		{"Active Employment", EmploymentStatusActive, true},
		{"Inactive Employment", EmploymentStatusInactive, true},
		{"Terminated Employment", EmploymentStatusTerminated, true},
		{"On Leave Employment", EmploymentStatusOnLeave, true},
		{"Contractor Employment", EmploymentStatusContractor, true},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().NotEmpty(string(tc.status))
			// Test that status constants are properly defined
			switch tc.status {
			case EmploymentStatusActive:
				s.Require().Equal("ACTIVE", string(tc.status))
			case EmploymentStatusInactive:
				s.Require().Equal("INACTIVE", string(tc.status))
			case EmploymentStatusTerminated:
				s.Require().Equal("TERMINATED", string(tc.status))
			case EmploymentStatusOnLeave:
				s.Require().Equal("ON_LEAVE", string(tc.status))
			case EmploymentStatusContractor:
				s.Require().Equal("CONTRACTOR", string(tc.status))
			}
		})
	}
}

// TestPersonType covers person type validation
func (s *TypeTestSuite) TestPersonType() {
	testCases := []struct {
		name        string
		personType  PersonType
		stringValue string
	}{
		{"Individual Type", PersonTypeIndividual, "INDIVIDUAL"},
		{"Employee Type", PersonTypeEmployee, "EMPLOYEE"},
		{"Contact Type", PersonTypeContact, "CONTACT"},
		{"Customer Type", PersonTypeCustomer, "CUSTOMER"},
		{"Vendor Type", PersonTypeVendor, "VENDOR"},
		{"Contractor Type", PersonTypeContractor, "CONTRACTOR"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.personType))
			s.Require().NotEmpty(string(tc.personType))
		})
	}
}

// TestRoleType covers role type validation
func (s *TypeTestSuite) TestRoleType() {
	testCases := []struct {
		name        string
		roleType    RoleType
		stringValue string
	}{
		{"System Role Type", RoleTypeSystem, "SYSTEM"},
		{"Tenant Role Type", RoleTypeTenant, "TENANT"},
		{"Entity Role Type", RoleTypeEntity, "ENTITY"},
		{"Custom Role Type", RoleTypeCustom, "CUSTOM"},
		{"Functional Role Type", RoleTypeFunctional, "FUNCTIONAL"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.roleType))
			s.Require().NotEmpty(string(tc.roleType))
		})
	}
}

// TestMFAMethod covers MFA method validation
func (s *TypeTestSuite) TestMFAMethod() {
	testCases := []struct {
		name        string
		method      MFAMethod
		stringValue string
	}{
		{"TOTP Method", MFAMethodTOTP, "TOTP"},
		{"SMS Method", MFAMethodSMS, "SMS"},
		{"Email Method", MFAMethodEmail, "EMAIL"},
		{"Push Method", MFAMethodPush, "PUSH"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.method))
			s.Require().NotEmpty(string(tc.method))
		})
	}
}

// TestSessionStatus covers session status validation
func (s *TypeTestSuite) TestSessionStatus() {
	testCases := []struct {
		name        string
		status      SessionStatus
		stringValue string
	}{
		{"Active Session", SessionStatusActive, "ACTIVE"},
		{"Expired Session", SessionStatusExpired, "EXPIRED"},
		{"Revoked Session", SessionStatusRevoked, "REVOKED"},
		{"Inactive Session", SessionStatusInactive, "INACTIVE"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.status))
			s.Require().NotEmpty(string(tc.status))
		})
	}
}

// TestAttributeDataType covers attribute data type validation
func (s *TypeTestSuite) TestAttributeDataType() {
	testCases := []struct {
		name        string
		dataType    AttributeDataType
		stringValue string
	}{
		{"String Data Type", AttributeDataTypeString, "STRING"},
		{"Number Data Type", AttributeDataTypeNumber, "NUMBER"},
		{"Boolean Data Type", AttributeDataTypeBoolean, "BOOLEAN"},
		{"Date Data Type", AttributeDataTypeDate, "DATE"},
		{"JSON Data Type", AttributeDataTypeJSON, "JSON"},
		{"Array Data Type", AttributeDataTypeArray, "ARRAY"},
		{"Enum Data Type", AttributeDataTypeEnum, "ENUM"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.dataType))
			s.Require().NotEmpty(string(tc.dataType))
		})
	}
}

// TestRequestType covers access request type validation
func (s *TypeTestSuite) TestRequestType() {
	testCases := []struct {
		name        string
		reqType     RequestType
		stringValue string
	}{
		{"Role Assignment Type", RequestTypeRoleAssignment, "ROLE_ASSIGNMENT"},
		{"Permission Grant Type", RequestTypePermissionGrant, "PERMISSION_GRANT"},
		{"Resource Access Type", RequestTypeResourceAccess, "RESOURCE_ACCESS"},
		{"Elevation Type", RequestTypeElevation, "ELEVATION"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.reqType))
			s.Require().NotEmpty(string(tc.reqType))
		})
	}
}

// TestApprovalStatus covers approval status validation
func (s *TypeTestSuite) TestApprovalStatus() {
	testCases := []struct {
		name        string
		status      ApprovalStatus
		stringValue string
	}{
		{"Pending Status", ApprovalStatusPending, "PENDING"},
		{"Approved Status", ApprovalStatusApproved, "APPROVED"},
		{"Rejected Status", ApprovalStatusRejected, "REJECTED"},
		{"Expired Status", ApprovalStatusExpired, "EXPIRED"},
		{"Revoked Status", ApprovalStatusRevoked, "REVOKED"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.status))
			s.Require().NotEmpty(string(tc.status))
		})
	}
}

// TestPolicyCombiningAlgorithm covers policy combining algorithm validation
func (s *TypeTestSuite) TestPolicyCombiningAlgorithm() {
	testCases := []struct {
		name        string
		algorithm   PolicyCombiningAlgorithm
		stringValue string
	}{
		{"Deny Overrides", CombiningAlgorithmDenyOverrides, "DENY_OVERRIDES"},
		{"Permit Overrides", CombiningAlgorithmPermitOverrides, "PERMIT_OVERRIDES"},
		{"First Applicable", CombiningAlgorithmFirstApplicable, "FIRST_APPLICABLE"},
		{"Only One Applicable", CombiningAlgorithmOnlyOneApplicable, "ONLY_ONE_APPLICABLE"},
		{"Ordered Deny Overrides", CombiningAlgorithmOrderedDenyOverrides, "ORDERED_DENY_OVERRIDES"},
		{"Ordered Permit Overrides", CombiningAlgorithmOrderedPermitOverrides, "ORDERED_PERMIT_OVERRIDES"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.algorithm))
			s.Require().NotEmpty(string(tc.algorithm))
		})
	}
}

// TestAttributeCategory covers attribute category validation
func (s *TypeTestSuite) TestAttributeCategory() {
	testCases := []struct {
		name        string
		category    AttributeCategory
		stringValue string
	}{
		{"User Category", AttributeCategoryUser, "USER"},
		{"Resource Category", AttributeCategoryResource, "RESOURCE"},
		{"Environment Category", AttributeCategoryEnvironment, "ENVIRONMENT"},
		{"Action Category", AttributeCategoryAction, "ACTION"},
		{"Entity Category", AttributeCategoryEntity, "ENTITY"},
		{"Session Category", AttributeCategorySession, "SESSION"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.stringValue, string(tc.category))
			s.Require().NotEmpty(string(tc.category))
		})
	}
}

// TestPolicyDecisionType covers policy decision type validation
func (s *TypeTestSuite) TestPolicyDecisionType() {
	testCases := []struct {
		name        string
		decision    PolicyDecisionType
		expectValid bool
		stringValue string
	}{
		{
			name:        "Valid ALLOW Decision",
			decision:    PolicyDecisionAllow,
			expectValid: true,
			stringValue: "ALLOW",
		},
		{
			name:        "Valid DENY Decision",
			decision:    PolicyDecisionDeny,
			expectValid: true,
			stringValue: "DENY",
		},
		{
			name:        "Valid NOT_APPLICABLE Decision",
			decision:    PolicyDecisionNotApplicable,
			expectValid: true,
			stringValue: "NOT_APPLICABLE",
		},
		{
			name:        "Invalid Decision",
			decision:    PolicyDecisionType("INVALID"),
			expectValid: false,
			stringValue: "INVALID",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Require().Equal(tc.expectValid, tc.decision.IsValid())
			s.Require().Equal(tc.stringValue, tc.decision.String())
		})
	}
}
