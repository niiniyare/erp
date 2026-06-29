package domain_test

import (
	"testing"

	"awo.so/internal/core/tenant/domain"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// StatusSuite — TenantStatus state machine, Valid, Parse
// ---------------------------------------------------------------------------

type StatusSuite struct {
	suite.Suite
}

func TestStatusSuite(t *testing.T) { suite.Run(t, new(StatusSuite)) }

// ---- Constants -----------------------------------------------------------

func (s *StatusSuite) TestStatusConstants() {
	tests := []struct {
		name    string
		status  domain.TenantStatus
		wantStr string
	}{
		{"active", domain.StatusActive, "ACTIVE"},
		{"suspended", domain.StatusSuspended, "SUSPENDED"},
		{"pending", domain.StatusPending, "PENDING"},
		{"archived", domain.StatusArchived, "ARCHIVED"},
		{"trial alias equals pending", domain.StatusTrial, "PENDING"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			require.Equal(s.T(), tc.wantStr, string(tc.status))
			require.Equal(s.T(), tc.wantStr, tc.status.String())
		})
	}
}

// ---- Valid ---------------------------------------------------------------

func (s *StatusSuite) TestValid() {
	tests := []struct {
		status domain.TenantStatus
		want   bool
	}{
		{domain.StatusActive, true},
		{domain.StatusSuspended, true},
		{domain.StatusPending, true},
		{domain.StatusArchived, true},
		{domain.TenantStatus("INVALID"), false},
		{domain.TenantStatus("active"), false},
		{domain.TenantStatus(""), false},
	}

	for _, tc := range tests {
		s.Run(string(tc.status), func() {
			require.Equal(s.T(), tc.want, tc.status.Valid())
		})
	}
}

// ---- CanTransitionTo — valid transitions ---------------------------------

func (s *StatusSuite) TestCanTransitionTo_Valid() {
	tests := []struct {
		name   string
		from   domain.TenantStatus
		to     domain.TenantStatus
		expect bool
	}{
		// PENDING → ACTIVE only
		{"pending→active", domain.StatusPending, domain.StatusActive, true},

		// ACTIVE → SUSPENDED, ARCHIVED
		{"active→suspended", domain.StatusActive, domain.StatusSuspended, true},
		{"active→archived", domain.StatusActive, domain.StatusArchived, true},

		// SUSPENDED → ACTIVE, ARCHIVED
		{"suspended→active", domain.StatusSuspended, domain.StatusActive, true},
		{"suspended→archived", domain.StatusSuspended, domain.StatusArchived, true},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			require.True(s.T(), tc.from.CanTransitionTo(tc.to))
		})
	}
}

// ---- CanTransitionTo — blocked transitions --------------------------------

func (s *StatusSuite) TestCanTransitionTo_Blocked() {
	tests := []struct {
		name string
		from domain.TenantStatus
		to   domain.TenantStatus
	}{
		// PENDING cannot go anywhere except ACTIVE
		{"pending→suspended", domain.StatusPending, domain.StatusSuspended},
		{"pending→archived", domain.StatusPending, domain.StatusArchived},

		// ARCHIVED is terminal
		{"archived→active", domain.StatusArchived, domain.StatusActive},
		{"archived→suspended", domain.StatusArchived, domain.StatusSuspended},
		{"archived→pending", domain.StatusArchived, domain.StatusPending},

		// Cannot go back to PENDING
		{"active→pending", domain.StatusActive, domain.StatusPending},
		{"suspended→pending", domain.StatusSuspended, domain.StatusPending},

		// Self transitions
		{"active→active", domain.StatusActive, domain.StatusActive},
		{"pending→pending", domain.StatusPending, domain.StatusPending},
		{"suspended→suspended", domain.StatusSuspended, domain.StatusSuspended},
		{"archived→archived", domain.StatusArchived, domain.StatusArchived},

		// Unknown source
		{"unknown→active", domain.TenantStatus("UNKNOWN"), domain.StatusActive},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			require.False(s.T(), tc.from.CanTransitionTo(tc.to))
		})
	}
}

// ---- ParseTenantStatus ---------------------------------------------------

func (s *StatusSuite) TestParseTenantStatus() {
	tests := []struct {
		input   string
		want    domain.TenantStatus
		wantErr bool
	}{
		{"ACTIVE", domain.StatusActive, false},
		{"PENDING", domain.StatusPending, false},
		{"SUSPENDED", domain.StatusSuspended, false},
		{"ARCHIVED", domain.StatusArchived, false},
		{"active", "", true},
		{"Active", "", true},
		{"INVALID", "", true},
		{"", "", true},
		{"DEACTIVATED", "", true},
	}

	for _, tc := range tests {
		s.Run(tc.input, func() {
			got, err := domain.ParseTenantStatus(tc.input)
			if tc.wantErr {
				require.Error(s.T(), err)
				require.Empty(s.T(), got)
			} else {
				require.NoError(s.T(), err)
				require.Equal(s.T(), tc.want, got)
			}
		})
	}
}
