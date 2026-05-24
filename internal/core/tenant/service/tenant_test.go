package service_test

import (
	"context"
	"testing"
	"time"

	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/core/tenant/service"
	"awo.so/internal/platform/cache"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ---------------------------------------------------------------------------
// TenantServiceSuite
// ---------------------------------------------------------------------------

type TenantServiceSuite struct {
	suite.Suite
	repo  *mockRepo
	mock  *gomock.Controller
	cache *cache.MockService
	svc   *service.TenantService
}

func TestTenantServiceSuite(t *testing.T) { suite.Run(t, new(TenantServiceSuite)) }

func (s *TenantServiceSuite) SetupTest() {
	s.repo = newMockRepo()
	ctrl := gomock.NewController(s.T())
	s.cache = cache.NewMockService(ctrl)
	s.svc = service.NewTenantService(s.repo, s.cache, noopTracer{}, noopLogger{})
}

// helper to build a minimal valid create request
func validCreateReq() domain.CreateTenantRequest {
	return domain.CreateTenantRequest{
		Name:         "Acme Corp",
		Email:        "admin@acme.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}
}

// helper to seed a tenant into the mock repo and return it
func (s *TenantServiceSuite) seedTenant(status domain.TenantStatus) *domain.Tenant {
	now := time.Now()
	t := &domain.Tenant{
		ID:           uuid.New(),
		Slug:         "seeded",
		Name:         "Seeded Tenant",
		Email:        "seed@test.com",
		Status:       status,
		Timezone:     "UTC",
		CurrencyCode: "USD",
		Metadata:     make(map[string]any),
		Settings:     make(map[string]any),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.repo.seed(t)
	return t
}

// ---- Create ---------------------------------------------------------------

func (s *TenantServiceSuite) TestCreate_Success() {
	req := validCreateReq()
	t, err := s.svc.Create(context.Background(), req)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), t)
	require.Equal(s.T(), "Acme Corp", t.Name)
	require.Equal(s.T(), "admin@acme.com", t.Email)
	require.NotEqual(s.T(), uuid.Nil, t.ID)
}

func (s *TenantServiceSuite) TestCreate_ValidationErrors() {
	tests := []struct {
		name    string
		mutate  func(*domain.CreateTenantRequest)
		wantErr string
	}{
		{
			name:    "empty name",
			mutate:  func(r *domain.CreateTenantRequest) { r.Name = "" },
			wantErr: "invalid request",
		},
		{
			name:    "empty email",
			mutate:  func(r *domain.CreateTenantRequest) { r.Email = "" },
			wantErr: "invalid request",
		},
		{
			name: "invalid company size",
			mutate: func(r *domain.CreateTenantRequest) {
				bad := "tiny"
				r.CompanySize = &bad
			},
			wantErr: "invalid company size",
		},
		{
			name: "reserved subdomain",
			mutate: func(r *domain.CreateTenantRequest) {
				sub := "admin"
				r.Subdomain = &sub
			},
			wantErr: "invalid subdomain",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			req := validCreateReq()
			tc.mutate(&req)
			_, err := s.svc.Create(context.Background(), req)
			require.Error(s.T(), err)
			require.Contains(s.T(), err.Error(), tc.wantErr)
		})
	}
}

func (s *TenantServiceSuite) TestCreate_SubdomainTaken() {
	sub := "taken"
	existing := s.seedTenant(domain.StatusActive)
	existing.Subdomain = &sub

	req := validCreateReq()
	req.Subdomain = &sub

	_, err := s.svc.Create(context.Background(), req)
	require.ErrorIs(s.T(), err, domain.ErrSubdomainTaken)
}

func (s *TenantServiceSuite) TestCreate_ValidCompanySize() {
	sizes := []string{"STARTUP", "SMALL", "MEDIUM", "LARGE", "ENTERPRISE"}
	for _, sz := range sizes {
		s.Run(sz, func() {
			s.SetupTest() // fresh repo
			req := validCreateReq()
			size := sz
			req.CompanySize = &size
			t, err := s.svc.Create(context.Background(), req)
			require.NoError(s.T(), err)
			require.NotNil(s.T(), t)
		})
	}
}

// ---- GetByID --------------------------------------------------------------

func (s *TenantServiceSuite) TestGetByID_Found() {
	seeded := s.seedTenant(domain.StatusActive)

	t, err := s.svc.GetByID(context.Background(), seeded.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), seeded.ID, t.ID)
}

func (s *TenantServiceSuite) TestGetByID_NotFound() {
	_, err := s.svc.GetByID(context.Background(), uuid.New())
	require.ErrorIs(s.T(), err, domain.ErrTenantNotFound)
}

// ---- GetBySubdomain -------------------------------------------------------

func (s *TenantServiceSuite) TestGetBySubdomain_Found() {
	sub := "acme-corp"
	seeded := s.seedTenant(domain.StatusActive)
	seeded.Subdomain = &sub

	t, err := s.svc.GetBySubdomain(context.Background(), "acme-corp")
	require.NoError(s.T(), err)
	require.Equal(s.T(), seeded.ID, t.ID)
}

func (s *TenantServiceSuite) TestGetBySubdomain_NotFound() {
	_, err := s.svc.GetBySubdomain(context.Background(), "nonexistent")
	require.ErrorIs(s.T(), err, domain.ErrTenantNotFound)
}

// ---- Update ---------------------------------------------------------------

func (s *TenantServiceSuite) TestUpdate_Success() {
	seeded := s.seedTenant(domain.StatusActive)
	newName := "Updated Name"

	t, err := s.svc.Update(context.Background(), seeded.ID, domain.UpdateTenantRequest{
		Name: &newName,
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "Updated Name", t.Name)
}

func (s *TenantServiceSuite) TestUpdate_ValidationErrors() {
	seeded := s.seedTenant(domain.StatusActive)

	tests := []struct {
		name string
		req  domain.UpdateTenantRequest
	}{
		{
			name: "empty name",
			req:  domain.UpdateTenantRequest{Name: strPtr("")},
		},
		{
			name: "name too long",
			req:  domain.UpdateTenantRequest{Name: strPtr(longStr(256))},
		},
		{
			name: "subdomain too long",
			req:  domain.UpdateTenantRequest{Subdomain: strPtr(longStr(64))},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := s.svc.Update(context.Background(), seeded.ID, tc.req)
			require.Error(s.T(), err)
		})
	}
}

func (s *TenantServiceSuite) TestUpdate_SubdomainTaken() {
	sub := "taken"
	existing := s.seedTenant(domain.StatusActive)
	existing.Subdomain = &sub

	other := s.seedTenant(domain.StatusActive)

	_, err := s.svc.Update(context.Background(), other.ID, domain.UpdateTenantRequest{
		Subdomain: &sub,
	})
	require.ErrorIs(s.T(), err, domain.ErrSubdomainTaken)
}

// ---- Activate / Suspend / Archive -----------------------------------------

func (s *TenantServiceSuite) TestActivate() {
	tests := []struct {
		name    string
		initial domain.TenantStatus
		wantErr error
	}{
		{"pending→active", domain.StatusPending, nil},
		{"suspended→active", domain.StatusSuspended, nil},
		{"already active", domain.StatusActive, domain.ErrAlreadyActive},
		{"archived blocked", domain.StatusArchived, domain.ErrCannotActivateArchivedTenant},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			seeded := s.seedTenant(tc.initial)

			err := s.svc.Activate(context.Background(), seeded.ID)

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

func (s *TenantServiceSuite) TestSuspend() {
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
			s.SetupTest()
			seeded := s.seedTenant(tc.initial)

			err := s.svc.Suspend(context.Background(), seeded.ID, "test reason")

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

func (s *TenantServiceSuite) TestArchive() {
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
			s.SetupTest()
			seeded := s.seedTenant(tc.initial)

			err := s.svc.Archive(context.Background(), seeded.ID)

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

// ---- Delete ---------------------------------------------------------------

func (s *TenantServiceSuite) TestDelete() {
	seeded := s.seedTenant(domain.StatusActive)
	err := s.svc.Delete(context.Background(), seeded.ID)
	require.NoError(s.T(), err)
}

func (s *TenantServiceSuite) TestDelete_NotFound() {
	err := s.svc.Delete(context.Background(), uuid.New())
	require.ErrorIs(s.T(), err, domain.ErrTenantNotFound)
}

// ---- List -----------------------------------------------------------------

func (s *TenantServiceSuite) TestList_PaginationDefaults() {
	// Seed 5 tenants
	for i := 0; i < 5; i++ {
		s.seedTenant(domain.StatusActive)
	}

	tests := []struct {
		name      string
		filter    domain.TenantFilter
		wantLimit int32
	}{
		{"zero limit defaults to 20", domain.TenantFilter{Limit: 0}, 20},
		{"negative limit defaults to 20", domain.TenantFilter{Limit: -1}, 20},
		{"over 100 defaults to 20", domain.TenantFilter{Limit: 200}, 20},
		{"valid limit 3", domain.TenantFilter{Limit: 3}, 3},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tenants, total, err := s.svc.List(context.Background(), tc.filter)
			require.NoError(s.T(), err)
			require.Equal(s.T(), int64(5), total)
			if int32(len(tenants)) > tc.wantLimit {
				s.Fail("returned more than limit")
			}
		})
	}
}

// ---- ResolveTenantID ------------------------------------------------------

func (s *TenantServiceSuite) TestResolveTenantID() {
	sub := "resolve-me"
	seeded := s.seedTenant(domain.StatusActive)
	seeded.Subdomain = &sub

	id, err := s.svc.ResolveTenantID(context.Background(), "resolve-me")
	require.NoError(s.T(), err)
	require.Equal(s.T(), seeded.ID, id)
}

func (s *TenantServiceSuite) TestResolveTenantID_NotFound() {
	_, err := s.svc.ResolveTenantID(context.Background(), "ghost")
	require.Error(s.T(), err)
}

// ---- ValidateTenantAccess -------------------------------------------------

func (s *TenantServiceSuite) TestValidateTenantAccess() {
	tests := []struct {
		name    string
		status  domain.TenantStatus
		wantErr error
	}{
		{"active ok", domain.StatusActive, nil},
		{"suspended blocked", domain.StatusSuspended, domain.ErrTenantSuspended},
		{"pending blocked", domain.StatusPending, domain.ErrTenantSuspended},
		{"archived blocked", domain.StatusArchived, domain.ErrTenantSuspended},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			seeded := s.seedTenant(tc.status)

			err := s.svc.ValidateTenantAccess(context.Background(), seeded.ID)

			if tc.wantErr != nil {
				require.ErrorIs(s.T(), err, tc.wantErr)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}

// ---- ExistsTenant ---------------------------------------------------------

func (s *TenantServiceSuite) TestExistsTenant() {
	seeded := s.seedTenant(domain.StatusActive)

	exists, err := s.svc.ExistsTenant(context.Background(), seeded.ID)
	require.NoError(s.T(), err)
	require.True(s.T(), exists)

	exists, err = s.svc.ExistsTenant(context.Background(), uuid.New())
	require.NoError(s.T(), err)
	require.False(s.T(), exists)
}

// ---- helpers --------------------------------------------------------------

func strPtr(s string) *string { return &s }

func longStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
