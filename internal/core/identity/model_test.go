//go:build unit
// +build unit

package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   AccountStatus
		expected bool
	}{
		{name: "Active Status", status: AccountStatusActive, expected: true},
		{name: "Inactive Status", status: AccountStatusInactive, expected: true},
		{name: "Locked Status", status: AccountStatusLocked, expected: true},
		{name: "Suspended Status", status: AccountStatusSuspended, expected: true},
		{name: "Invalid Status", status: "INVALID", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestAccountStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   AccountStatus
		expected string
	}{
		{name: "Active String", status: AccountStatusActive, expected: "ACTIVE"},
		{name: "Inactive String", status: AccountStatusInactive, expected: "INACTIVE"},
		{name: "Locked String", status: AccountStatusLocked, expected: "LOCKED"},
		{name: "Suspended String", status: AccountStatusSuspended, expected: "SUSPENDED"},
		{name: "Custom String", status: "CUSTOM", expected: "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestEmploymentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   EmploymentStatus
		expected bool
	}{
		{name: "Active Status", status: EmploymentStatusActive, expected: true},
		{name: "Inactive Status", status: EmploymentStatusInactive, expected: true},
		{name: "Terminated Status", status: EmploymentStatusTerminated, expected: true},
		{name: "On Leave Status", status: EmploymentStatusOnLeave, expected: true},
		{name: "Suspended Status", status: EmploymentStatusSuspended, expected: true},
		{name: "Invalid Status", status: "INVALID", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestEmploymentStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   EmploymentStatus
		expected string
	}{
		{name: "Active String", status: EmploymentStatusActive, expected: "ACTIVE"},
		{name: "Inactive String", status: EmploymentStatusInactive, expected: "INACTIVE"},
		{name: "Terminated String", status: EmploymentStatusTerminated, expected: "TERMINATED"},
		{name: "On Leave String", status: EmploymentStatusOnLeave, expected: "ON_LEAVE"},
		{name: "Suspended String", status: EmploymentStatusSuspended, expected: "SUSPENDED"},
		{name: "Custom String", status: "CUSTOM", expected: "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestPerson_GetFullName(t *testing.T) {
	tests := []struct {
		name     string
		person   Person
		expected string
	}{
		{name: "Without Middle Name", person: Person{FirstName: "John", LastName: "Doe"}, expected: "John Doe"},
		{name: "With Middle Name", person: Person{FirstName: "Jane", MiddleName: stringPtr("M"), LastName: "Doe"}, expected: "Jane M Doe"},
		{name: "Empty Middle Name", person: Person{FirstName: "Peter", MiddleName: stringPtr(""), LastName: "Jones"}, expected: "Peter Jones"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.person.GetFullName())
		})
	}
}

func TestIsValidUserType(t *testing.T) {
	tests := []struct {
		name     string
		userType string
		expected bool
	}{
		{name: "Valid ADMIN", userType: "ADMIN", expected: true},
		{name: "Valid INTERNAL", userType: "INTERNAL", expected: true},
		{name: "Valid CUSTOMER", userType: "CUSTOMER", expected: true},
		{name: "Valid VENDOR", userType: "VENDOR", expected: true},
		{name: "Invalid Type", userType: "GUEST", expected: false},
		{name: "Empty Type", userType: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidUserType(tt.userType))
		})
	}
}

func TestIsValidAccountStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{name: "Valid ACTIVE", status: "ACTIVE", expected: true},
		{name: "Valid INACTIVE", status: "INACTIVE", expected: true},
		{name: "Valid LOCKED", status: "LOCKED", expected: true},
		{name: "Valid SUSPENDED", status: "SUSPENDED", expected: true},
		{name: "Invalid Status", status: "BLOCKED", expected: false},
		{name: "Empty Status", status: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidAccountStatus(tt.status))
		})
	}
}

func TestIsValidEmploymentStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{name: "Valid ACTIVE", status: "ACTIVE", expected: true},
		{name: "Valid TERMINATED", status: "TERMINATED", expected: true},
		{name: "Valid ON_LEAVE", status: "ON_LEAVE", expected: true},
		{name: "Valid SUSPENDED", status: "SUSPENDED", expected: true},
		{name: "Invalid Status", status: "FIRED", expected: false},
		{name: "Empty Status", status: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidEmploymentStatus(tt.status))
		})
	}
}

// Helper function to get a pointer to a string
func stringPtr(s string) *string {
	return &s
}
