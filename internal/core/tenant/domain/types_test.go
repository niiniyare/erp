package domain_test

import (
	"testing"

	"awo.so/internal/core/tenant/domain"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TypesSuite struct {
	suite.Suite
}

func TestTypesSuite(t *testing.T) { suite.Run(t, new(TypesSuite)) }

// ---- PlanType constants ---------------------------------------------------

func (s *TypesSuite) TestPlanTypeConstants() {
	tests := []struct {
		name string
		plan domain.PlanType
		want string
	}{
		{"basic", domain.PlanBasic, "BASIC"},
		{"professional", domain.PlanProfessional, "PROFESSIONAL"},
		{"enterprise", domain.PlanEnterprise, "ENTERPRISE"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			require.Equal(s.T(), tc.want, string(tc.plan))
		})
	}
}

// ---- CompanySize ----------------------------------------------------------

func (s *TypesSuite) TestCompanySizeConstants() {
	tests := []struct {
		name string
		cs   domain.CompanySize
		want string
	}{
		{"startup", domain.CompanySizeStartup, "STARTUP"},
		{"small", domain.CompanySizeSmall, "SMALL"},
		{"medium", domain.CompanySizeMedium, "MEDIUM"},
		{"large", domain.CompanySizeLarge, "LARGE"},
		{"enterprise", domain.CompanySizeEnterprise, "ENTERPRISE"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			require.Equal(s.T(), tc.want, string(tc.cs))
		})
	}
}

func (s *TypesSuite) TestValidCompanySize() {
	tests := []struct {
		input string
		valid bool
	}{
		{"STARTUP", true},
		{"SMALL", true},
		{"MEDIUM", true},
		{"LARGE", true},
		{"ENTERPRISE", true},
		{"startup", false},
		{"Small", false},
		{"huge", false},
		{"", false},
		{"MICRO", false},
	}

	for _, tc := range tests {
		s.Run(tc.input, func() {
			require.Equal(s.T(), tc.valid, domain.ValidCompanySize(tc.input))
		})
	}
}

// ---- AccountingMethod -----------------------------------------------------

func (s *TypesSuite) TestAccountingMethodConstants() {
	require.Equal(s.T(), "ACCRUAL", string(domain.AccountingMethodAccrual))
	require.Equal(s.T(), "CASH", string(domain.AccountingMethodCash))
}

func (s *TypesSuite) TestValidAccountingMethod() {
	tests := []struct {
		input string
		valid bool
	}{
		{"ACCRUAL", true},
		{"CASH", true},
		{"accrual", false},
		{"cash", false},
		{"HYBRID", false},
		{"", false},
	}

	for _, tc := range tests {
		s.Run(tc.input, func() {
			require.Equal(s.T(), tc.valid, domain.ValidAccountingMethod(tc.input))
		})
	}
}

// ---- ReservedSubdomains ---------------------------------------------------

func (s *TypesSuite) TestReservedSubdomains() {
	expected := []string{
		"admin", "api", "www", "app", "cdn", "static", "assets", "mail",
		"ftp", "smtp", "dev", "staging", "prod", "production", "test", "localhost",
		"dashboard", "portal", "auth", "login", "signup", "register",
		"billing", "payment", "invoice", "support", "help", "docs", "status",
		"blog", "news", "about", "contact", "legal", "privacy", "terms",
	}

	for _, word := range expected {
		s.Run(word, func() {
			require.True(s.T(), domain.ReservedSubdomains[word],
				"%q should be reserved", word)
		})
	}

	// Non-reserved should be false
	require.False(s.T(), domain.ReservedSubdomains["acme"])
	require.False(s.T(), domain.ReservedSubdomains["mycompany"])
}
