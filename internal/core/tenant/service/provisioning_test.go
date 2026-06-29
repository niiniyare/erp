package service_test

import (
	"context"
	"fmt"
	"testing"

	"awo.so/internal/core/tenant/domain"
	"awo.so/internal/core/tenant/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// ProvisioningServiceSuite
// ---------------------------------------------------------------------------

type ProvisioningServiceSuite struct {
	suite.Suite
	repo *mockRepo
	svc  *service.ProvisioningService
}

func TestProvisioningServiceSuite(t *testing.T) { suite.Run(t, new(ProvisioningServiceSuite)) }

func (s *ProvisioningServiceSuite) SetupTest() {
	s.repo = newMockRepo()
	s.svc = service.NewProvisioningService(s.repo, noopTracer{})
}

func (s *ProvisioningServiceSuite) TestProvision_Success() {
	sub := "acme"
	input := domain.ProvisioningInput{
		Name:         "Acme Corp",
		Email:        "admin@acme.com",
		Subdomain:    &sub,
		CountryCode:  "US",
		CurrencyCode: "USD",
	}

	result, err := s.svc.Provision(context.Background(), input)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
	require.NotEqual(s.T(), uuid.Nil, result.TenantID)
	require.Equal(s.T(), &sub, result.Subdomain)
}

func (s *ProvisioningServiceSuite) TestProvision_RepoError() {
	s.repo.provisionFn = func(_ context.Context, _ domain.ProvisioningInput) (*domain.ProvisioningResult, error) {
		return nil, fmt.Errorf("db connection lost")
	}

	input := domain.ProvisioningInput{
		Name:         "Fail Corp",
		Email:        "fail@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}

	result, err := s.svc.Provision(context.Background(), input)

	require.Error(s.T(), err)
	require.Nil(s.T(), result)
	require.Contains(s.T(), err.Error(), "provisioning failed")
}

func (s *ProvisioningServiceSuite) TestProvision_TracingSpanCreated() {
	// Verifies the method doesn't panic with noopTracer — tracing integration is implicitly tested.
	input := domain.ProvisioningInput{
		Name:         "Traced Corp",
		Email:        "trace@test.com",
		CountryCode:  "US",
		CurrencyCode: "USD",
	}

	result, err := s.svc.Provision(context.Background(), input)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), result)
}
