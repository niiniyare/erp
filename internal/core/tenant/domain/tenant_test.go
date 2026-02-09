package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/tenant/domain"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// TenantSuite — NewTenant, Options, Business Methods, Predicates
// ---------------------------------------------------------------------------

type TenantSuite struct {
	suite.Suite
}

func TestTenantSuite(t *testing.T) { suite.Run(t, new(TenantSuite)) }

// ---- NewTenant creation ---------------------------------------------------

func (s *TenantSuite) TestNewTenant_Valid() {
	tests := []struct {
		name      string
		tenantNm  string
		email     string
		wantSlug  string
		wantEmail string
	}{
		{
			name:      "simple name",
			tenantNm:  "Acme Corporation",
			email:     "admin@acme.com",
			wantSlug:  "acme-corporation",
			wantEmail: "admin@acme.com",
		},
		{
			name:      "email normalised to lowercase",
			tenantNm:  "Beta Ltd",
			email:     "Admin@Beta.COM",
			wantSlug:  "beta-ltd",
			wantEmail: "admin@beta.com",
		},
		{
			name:      "name trimmed",
			tenantNm:  "  Gamma Inc  ",
			email:     "info@gamma.io",
			wantSlug:  "gamma-inc",
			wantEmail: "info@gamma.io",
		},
		{
			name:      "special characters in name",
			tenantNm:  "Smith & Sons Ltd.",
			email:     "contact@smith.co.uk",
			wantSlug:  "smith-and-sons-ltd",
			wantEmail: "contact@smith.co.uk",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t, err := domain.NewTenant(tc.tenantNm, tc.email)
			require.NoError(s.T(), err)
			require.NotNil(s.T(), t)

			require.NotEqual(s.T(), uuid.Nil, t.ID)
			require.Equal(s.T(), tc.wantSlug, t.Slug)
			require.Equal(s.T(), strings.TrimSpace(tc.tenantNm), t.Name)
			require.Equal(s.T(), tc.wantEmail, t.Email)
			require.Equal(s.T(), domain.StatusPending, t.Status)
			require.Equal(s.T(), "UTC", t.Timezone)
			require.Equal(s.T(), "USD", t.CurrencyCode)
			require.NotNil(s.T(), t.Metadata)
			require.NotNil(s.T(), t.Settings)
			require.False(s.T(), t.CreatedAt.IsZero())
			require.False(s.T(), t.UpdatedAt.IsZero())
			require.Nil(s.T(), t.DeletedAt)
		})
	}
}

func (s *TenantSuite) TestNewTenant_ValidationErrors() {
	tests := []struct {
		name    string
		tName   string
		email   string
		wantErr error
	}{
		{"empty name", "", "a@b.com", domain.ErrTenantNameRequired},
		{"whitespace name", "   ", "a@b.com", domain.ErrTenantNameRequired},
		{"empty email", "Acme", "", domain.ErrTenantEmailRequired},
		{"whitespace email", "Acme", "   ", domain.ErrTenantEmailRequired},
		{"no at sign", "Acme", "invalid-email", domain.ErrInvalidEmail},
		{"no dot", "Acme", "user@domain", domain.ErrInvalidEmail},
		{"at only", "Acme", "@", domain.ErrInvalidEmail},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t, err := domain.NewTenant(tc.tName, tc.email)
			require.ErrorIs(s.T(), err, tc.wantErr)
			require.Nil(s.T(), t)
		})
	}
}

// ---- Functional options ---------------------------------------------------

func (s *TenantSuite) TestWithSubdomain() {
	tests := []struct {
		name      string
		subdomain string
		wantErr   error
		wantVal   string
	}{
		{"valid", "acme-corp", nil, "acme-corp"},
		{"numeric", "tenant123", nil, "tenant123"},
		{"single char", "a", nil, "a"},
		{"uppercase normalized", "ACME", nil, "acme"},
		{"reserved admin", "admin", domain.ErrInvalidSubdomain, ""},
		{"reserved api", "api", domain.ErrInvalidSubdomain, ""},
		{"reserved www", "www", domain.ErrInvalidSubdomain, ""},
		{"reserved dashboard", "dashboard", domain.ErrInvalidSubdomain, ""},
		{"leading hyphen", "-bad", domain.ErrInvalidSubdomain, ""},
		{"trailing hyphen", "bad-", domain.ErrInvalidSubdomain, ""},
		{"dot", "bad.sub", domain.ErrInvalidSubdomain, ""},
		{"space", "bad sub", domain.ErrInvalidSubdomain, ""},
		{"too long", strings.Repeat("a", 64), domain.ErrInvalidSubdomain, ""},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t, err := domain.NewTenant("X", "x@x.com", domain.WithSubdomain(tc.subdomain))
			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
				require.Nil(s.T(), t)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), t.Subdomain)
				require.Equal(s.T(), tc.wantVal, *t.Subdomain)
			}
		})
	}
}

func (s *TenantSuite) TestWithLimits_CompanySizeValidation() {
	tests := []struct {
		name    string
		size    string
		wantErr error
	}{
		{"STARTUP", "STARTUP", nil},
		{"SMALL", "SMALL", nil},
		{"MEDIUM", "MEDIUM", nil},
		{"LARGE", "LARGE", nil},
		{"ENTERPRISE", "ENTERPRISE", nil},
		{"lowercase rejected", "small", domain.ErrInvalidCompanySize},
		{"title case rejected", "Small", domain.ErrInvalidCompanySize},
		{"unknown rejected", "huge", domain.ErrInvalidCompanySize},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			size := tc.size
			t, err := domain.NewTenant("X", "x@x.com", domain.WithLimits(nil, &size))
			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
				require.NotNil(s.T(), t.CompanySize)
				require.Equal(s.T(), tc.size, *t.CompanySize)
			}
		})
	}
}

func (s *TenantSuite) TestWithTimezoneAndCurrency() {
	t, err := domain.NewTenant("X", "x@x.com",
		domain.WithTimezone("Africa/Nairobi"),
		domain.WithCurrency("KES"),
	)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "Africa/Nairobi", t.Timezone)
	require.Equal(s.T(), "KES", t.CurrencyCode)
}

func (s *TenantSuite) TestWithStatus() {
	t, err := domain.NewTenant("X", "x@x.com", domain.WithStatus(domain.StatusActive))
	require.NoError(s.T(), err)
	require.Equal(s.T(), domain.StatusActive, t.Status)
}

// ---- Business methods (Activate, Suspend, Archive) ------------------------

func (s *TenantSuite) TestActivate() {
	tests := []struct {
		name       string
		initial    domain.TenantStatus
		wantErr    error
		wantStatus domain.TenantStatus
	}{
		{"pending→active", domain.StatusPending, nil, domain.StatusActive},
		{"suspended→active", domain.StatusSuspended, nil, domain.StatusActive},
		{"already active", domain.StatusActive, domain.ErrAlreadyActive, domain.StatusActive},
		{"archived blocked", domain.StatusArchived, domain.ErrCannotActivateArchivedTenant, domain.StatusArchived},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t := &domain.Tenant{Status: tc.initial, Metadata: make(map[string]any)}
			before := t.UpdatedAt

			err := t.Activate()

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
				require.Equal(s.T(), tc.initial, t.Status)
			} else {
				require.NoError(s.T(), err)
				require.Equal(s.T(), tc.wantStatus, t.Status)
				require.True(s.T(), t.UpdatedAt.After(before) || t.UpdatedAt.Equal(before))
			}
		})
	}
}

func (s *TenantSuite) TestSuspend() {
	tests := []struct {
		name    string
		initial domain.TenantStatus
		wantErr error
	}{
		{"active→suspended", domain.StatusActive, nil},
		{"already suspended", domain.StatusSuspended, domain.ErrAlreadySuspended},
		{"archived blocked", domain.StatusArchived, domain.ErrCannotSuspendArchivedTenant},
		{"pending blocked", domain.StatusPending, domain.ErrInvalidTransition},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t := &domain.Tenant{Status: tc.initial, Metadata: make(map[string]any)}

			err := t.Suspend("payment_failure")

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
				require.Equal(s.T(), domain.StatusSuspended, t.Status)
				require.Equal(s.T(), "payment_failure", t.Metadata["suspension_reason"])
				require.NotNil(s.T(), t.Metadata["suspended_at"])
			}
		})
	}
}

func (s *TenantSuite) TestSuspend_InitializesMetadata() {
	t := &domain.Tenant{Status: domain.StatusActive, Metadata: nil}
	require.NoError(s.T(), t.Suspend("test"))
	require.NotNil(s.T(), t.Metadata)
	require.Equal(s.T(), "test", t.Metadata["suspension_reason"])
}

func (s *TenantSuite) TestArchive() {
	tests := []struct {
		name    string
		initial domain.TenantStatus
		wantErr error
	}{
		{"active→archived", domain.StatusActive, nil},
		{"suspended→archived", domain.StatusSuspended, nil},
		{"already archived", domain.StatusArchived, domain.ErrAlreadyArchived},
		{"pending blocked", domain.StatusPending, domain.ErrInvalidTransition},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t := &domain.Tenant{Status: tc.initial}

			err := t.Archive()

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
				require.Equal(s.T(), domain.StatusArchived, t.Status)
				require.NotNil(s.T(), t.DeletedAt)
				require.True(s.T(), t.IsSoftDeleted())
			}
		})
	}
}

// ---- UpdateActivity -------------------------------------------------------

func (s *TenantSuite) TestUpdateActivity() {
	t := &domain.Tenant{}
	require.Nil(s.T(), t.LastActivityAt)

	t.UpdateActivity()

	require.NotNil(s.T(), t.LastActivityAt)
	require.WithinDuration(s.T(), time.Now(), *t.LastActivityAt, time.Second)
	require.WithinDuration(s.T(), time.Now(), t.UpdatedAt, time.Second)
}

// ---- Predicates -----------------------------------------------------------

func (s *TenantSuite) TestPredicates() {
	tests := []struct {
		name        string
		status      domain.TenantStatus
		deletedAt   *time.Time
		isActive    bool
		isSuspended bool
		isArchived  bool
		isSoftDel   bool
	}{
		{"active", domain.StatusActive, nil, true, false, false, false},
		{"suspended", domain.StatusSuspended, nil, false, true, false, false},
		{"pending", domain.StatusPending, nil, false, false, false, false},
		{"archived", domain.StatusArchived, timePtr(time.Now()), false, false, true, true},
		{"active but soft deleted", domain.StatusActive, timePtr(time.Now()), true, false, false, true},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			t := &domain.Tenant{Status: tc.status, DeletedAt: tc.deletedAt}
			require.Equal(s.T(), tc.isActive, t.IsActive())
			require.Equal(s.T(), tc.isSuspended, t.IsSuspended())
			require.Equal(s.T(), tc.isArchived, t.IsArchived())
			require.Equal(s.T(), tc.isSoftDel, t.IsSoftDeleted())
		})
	}
}

func timePtr(t time.Time) *time.Time { return &t }
