package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, tenant *Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Tenant), args.Error(1)
}

func (m *MockRepository) GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	args := m.Called(ctx, subdomain)
	return args.Get(0).(*Tenant), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id uuid.UUID, updates UpdateTenantRequest) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, offset, limit int) ([]*Tenant, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*Tenant), args.Error(1)
}

func (m *MockRepository) Exists(ctx context.Context, subdomain string) (bool, error) {
	args := m.Called(ctx, subdomain)
	return args.Bool(0), args.Error(1)
}

func TestTenantService(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := cache.NewMockCache()
	service := NewService(mockRepo, mockCache)

	t.Run("CreateTenant", func(t *testing.T) {
		req := CreateTenantRequest{
			Name:      "Test Tenant",
			Subdomain: "test",
			PlanType:  PlanTypeBasic,
		}

		mockRepo.On("Exists", mock.Anything, req.Subdomain).Return(false, nil)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		tenant, err := service.CreateTenant(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, tenant)
		assert.Equal(t, req.Name, tenant.Name)
		assert.Equal(t, req.Subdomain, tenant.Subdomain)

		mockRepo.AssertExpectations(t)
	})
}

