package tenant_test

//
// import (
// 	"context"
// 	"testing"
//
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )
//
// // MockTenantRepository is a mock implementation of TenantRepository
// type MockTenantRepository struct {
// 	mock.Mock
// }
//
// func (m *MockTenantRepository) Create(ctx context.Context, tenant *Tenant) error {
// 	args := m.Called(ctx, tenant)
// 	return args.Error(0)
// }
//
// func (m *MockTenantRepository) GetByID(ctx context.Context, id string) (*Tenant, error) {
// 	args := m.Called(ctx, id)
// 	return args.Get(0).(*Tenant), args.Error(1)
// }
//
// func (m *MockTenantRepository) Update(ctx context.Context, tenant *Tenant) error {
// 	args := m.Called(ctx, tenant)
// 	return args.Error(0)
// }
//
// func (m *MockTenantRepository) Delete(ctx context.Context, id string) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }
//
// func TestTenantService_CreateTenant(t *testing.T) {
// 	mockRepo := new(MockTenantRepository)
// 	service := NewTenantService(mockRepo)
//
// 	tenant := &Tenant{
// 		Name:   "Test Tenant",
// 		Domain: "test.example.com",
// 	}
//
// 	mockRepo.On("Create", mock.Anything, tenant).Return(nil)
//
// 	err := service.CreateTenant(context.Background(), tenant)
//
// 	assert.NoError(t, err)
// 	mockRepo.AssertExpectations(t)
// }
//
// func TestTenantService_GetTenant(t *testing.T) {
// 	mockRepo := new(MockTenantRepository)
// 	service := NewTenantService(mockRepo)
//
// 	expectedTenant := &Tenant{
// 		ID:     "123",
// 		Name:   "Test Tenant",
// 		Domain: "test.example.com",
// 	}
//
// 	mockRepo.On("GetByID", mock.Anything, "123").Return(expectedTenant, nil)
//
// 	tenant, err := service.GetTenant(context.Background(), "123")
//
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedTenant, tenant)
// 	mockRepo.AssertExpectations(t)
// }
//
