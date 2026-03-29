//go:build unit
// +build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

// TransactionRepositoryUnitTestSuite defines unit test suite using mocks
type TransactionRepositoryUnitTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockStore  *db.MockStore
	mockTracer *tracing.MockService
	repo       *transactionRepository
	ctx        context.Context
	tenantID   uuid.UUID
}

// SetupTest runs before each test
func (s *TransactionRepositoryUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockStore = db.NewMockStore(s.ctrl)
	s.mockTracer = tracing.NewMockService(s.ctrl)
	s.repo = NewTransactionRepository(s.mockStore, s.mockTracer)
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)
}

// TearDownTest runs after each test
func (s *TransactionRepositoryUnitTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test runner
func TestTransactionRepositoryUnitTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionRepositoryUnitTestSuite))
}

// TestCreateTransactionSuccess tests successful transaction creation
func (s *TransactionRepositoryUnitTestSuite) TestCreateTransactionSuccess() {
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("TXN-001"),
		Description:       "Test transaction",
		Amount:            decimal.NewFromFloat(100.50),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	// Expected SQLC parameters
	expectedParams := db.CreateTransactionParams{
		ID:                 transaction.ID,
		EntityID:           transaction.EntityID,
		TransactionDate:    transaction.TransactionDate,
		TransactionType:    db.TransactionTypeEnumMANUAL,
		TransactionStatus:  db.TransactionStatusEnumPENDING,
		ReferenceNumber:    transaction.ReferenceNumber,
		Description:        transaction.Description,
		Amount:             decimalToPgNumeric(&transaction.Amount),
		CurrencyCode:       transaction.CurrencyCode,
		ExchangeRate:       decimalToPgNumeric(&transaction.ExchangeRate),
		IsRecurring:        transaction.IsRecurring,
		RecurringFrequency: mapDomainRecurringFrequencyToNullEnum(transaction.RecurringFrequency),
		RecurringEndDate:   transaction.RecurringEndDate,
		CreatedBy:          &transaction.CreatedBy,
	}

	// Mock successful creation
	returnedTransaction := db.FinanceTransaction{
		ID:                transaction.ID,
		TenantID:          s.tenantID,
		EntityID:          transaction.EntityID,
		TransactionDate:   transaction.TransactionDate,
		TransactionType:   db.TransactionTypeEnumMANUAL,
		TransactionStatus: db.TransactionStatusEnumPENDING,
		ReferenceNumber:   transaction.ReferenceNumber,
		Description:       transaction.Description,
		Amount:            decimalToPgNumeric(&transaction.Amount),
		CurrencyCode:      transaction.CurrencyCode,
		IsRecurring:       transaction.IsRecurring,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         &transaction.CreatedBy,
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			// Mock the inner store call
			s.mockStore.EXPECT().
				CreateTransaction(gomock.Any(), gomock.Eq(expectedParams)).
				Return(&returnedTransaction, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Create(s.ctx, transaction)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().Equal(s.tenantID, transaction.TenantID)
	s.Assert().NotZero(transaction.CreatedAt)
	s.Assert().NotZero(transaction.UpdatedAt)
}

// TestCreateTransactionDuplicateReference tests handling of duplicate reference numbers
func (s *TransactionRepositoryUnitTestSuite) TestCreateTransactionDuplicateReference() {
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPending,
		ReferenceNumber:   stringPtr("DUP-REF-001"),
		Description:       "Test transaction",
		Amount:            decimal.NewFromFloat(100.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedBy:         uuid.New(),
	}

	// Mock duplicate key violation
	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				CreateTransaction(gomock.Any(), gomock.Any()).
				Return(nil, &mockPgError{code: "23505"}). // Unique violation
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Create(s.ctx, transaction)

	// Assert results
	s.Assert().Error(err)
	s.Assert().Equal(domain.ErrTransactionReferenceExists, err)
}

// TestGetByIDSuccess tests successful transaction retrieval by ID
func (s *TransactionRepositoryUnitTestSuite) TestGetByIDSuccess() {
	transactionID := uuid.New()
	expectedTransaction := db.FinanceTransaction{
		ID:                transactionID,
		TenantID:          s.tenantID,
		TransactionDate:   time.Now(),
		TransactionType:   db.TransactionTypeEnumMANUAL,
		TransactionStatus: db.TransactionStatusEnumPENDING,
		ReferenceNumber:   stringPtr("TXN-001"),
		Description:       "Test transaction",
		Amount:            decimalToPgNumeric(decimalPtr(decimal.NewFromFloat(100.50))),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         &uuid.UUID{},
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				GetTransactionByID(gomock.Any(), transactionID).
				Return(&expectedTransaction, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	result, err := s.repo.GetByID(s.ctx, transactionID)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().NotNil(result)
	s.Assert().Equal(transactionID, result.ID)
	s.Assert().Equal("Test transaction", result.Description)
	s.Assert().Equal(domain.TransactionTypeManual, result.TransactionType)
	s.Assert().Equal(domain.TransactionStatusPending, result.TransactionStatus)
}

// TestGetByIDNotFound tests handling of non-existent transaction
func (s *TransactionRepositoryUnitTestSuite) TestGetByIDNotFound() {
	transactionID := uuid.New()

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				GetTransactionByID(gomock.Any(), transactionID).
				Return(nil, db.ErrNoRows).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	result, err := s.repo.GetByID(s.ctx, transactionID)

	// Assert results
	s.Assert().Error(err)
	s.Assert().Nil(result)
	s.Assert().Equal(domain.ErrTransactionNotFound, err)
}

// TestGetByReferenceSuccess tests successful transaction retrieval by reference number
func (s *TransactionRepositoryUnitTestSuite) TestGetByReferenceSuccess() {
	referenceNumber := "REF-001"
	expectedTransaction := db.FinanceTransaction{
		ID:                uuid.New(),
		TenantID:          s.tenantID,
		TransactionDate:   time.Now(),
		TransactionType:   db.TransactionTypeEnumMANUAL,
		TransactionStatus: db.TransactionStatusEnumPENDING,
		ReferenceNumber:   &referenceNumber,
		Description:       "Test transaction",
		Amount:            decimalToPgNumeric(decimalPtr(decimal.NewFromFloat(100.50))),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				GetTransactionByReference(gomock.Any(), &referenceNumber).
				Return(&expectedTransaction, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	result, err := s.repo.GetByReference(s.ctx, referenceNumber)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().NotNil(result)
	s.Assert().Equal(referenceNumber, *result.ReferenceNumber)
	s.Assert().Equal("Test transaction", result.Description)
}

// TestUpdateTransactionSuccess tests successful transaction update
func (s *TransactionRepositoryUnitTestSuite) TestUpdateTransactionSuccess() {
	transaction := &domain.Transaction{
		ID:                uuid.New(),
		TenantID:          s.tenantID,
		TransactionDate:   time.Now(),
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusPosted,
		ReferenceNumber:   stringPtr("UPD-001"),
		Description:       "Updated transaction description",
		Amount:            decimal.NewFromFloat(150.00),
		CurrencyCode:      "USD",
		IsRecurring:       false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         uuid.New(),
	}

	expectedParams := db.UpdateTransactionParams{
		ID:                 transaction.ID,
		TransactionDate:    transaction.TransactionDate,
		TransactionStatus:  db.TransactionStatusEnumPOSTED,
		Description:        transaction.Description,
		Amount:             decimalToPgNumeric(&transaction.Amount),
		ExchangeRate:       decimalToPgNumeric(&transaction.ExchangeRate),
		IsRecurring:        transaction.IsRecurring,
		RecurringFrequency: mapDomainRecurringFrequencyToNullEnum(transaction.RecurringFrequency),
		RecurringEndDate:   transaction.RecurringEndDate,
		UpdatedBy:          transaction.UpdatedBy,
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				UpdateTransaction(gomock.Any(), gomock.Eq(expectedParams)).
				Return(nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Update(s.ctx, transaction)

	// Assert results
	s.Assert().NoError(err)
}

// TestApproveTransactionSuccess tests successful transaction approval
func (s *TransactionRepositoryUnitTestSuite) TestApproveTransactionSuccess() {
	transactionID := uuid.New()
	approvedBy := uuid.New()
	approvedAt := time.Now()
	notes := "Approved for processing"

	expectedParams := db.ApproveTransactionParams{
		ID: transactionID,
		ApprovalStatus: db.NullApprovalStatusEnum{
			ApprovalStatusEnum: db.ApprovalStatusEnumAPPROVED,
			Valid:              true,
		},
		ApprovedBy:    &approvedBy,
		ApprovedAt:    &approvedAt,
		ApprovalNotes: &notes,
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				ApproveTransaction(gomock.Any(), gomock.Eq(expectedParams)).
				Return(nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Approve(s.ctx, transactionID, approvedBy, approvedAt, &notes)

	// Assert results
	s.Assert().NoError(err)
}

// TestDeleteTransactionSuccess tests successful transaction soft delete
func (s *TransactionRepositoryUnitTestSuite) TestDeleteTransactionSuccess() {
	transactionID := uuid.New()

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				SoftDeleteTransaction(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	err := s.repo.Delete(s.ctx, transactionID)

	// Assert results
	s.Assert().NoError(err)
}

// TestListTransactionsSuccess tests successful transaction listing
func (s *TransactionRepositoryUnitTestSuite) TestListTransactionsSuccess() {
	filter := &domain.TransactionFilter{
		Status: &domain.TransactionStatusPosted,
		Limit:  intPtr(10),
		Offset: intPtr(0),
	}

	expectedTransactions := []*db.FinanceTransaction{
		{
			ID:                uuid.New(),
			TenantID:          s.tenantID,
			TransactionDate:   time.Now(),
			TransactionType:   db.TransactionTypeEnumMANUAL,
			TransactionStatus: db.TransactionStatusEnumPOSTED,
			Description:       "Transaction 1",
			Amount:            decimalToPgNumeric(decimalPtr(decimal.NewFromFloat(100.00))),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			TenantID:          s.tenantID,
			TransactionDate:   time.Now(),
			TransactionType:   db.TransactionTypeEnumSYSTEM,
			TransactionStatus: db.TransactionStatusEnumPOSTED,
			Description:       "Transaction 2",
			Amount:            decimalToPgNumeric(decimalPtr(decimal.NewFromFloat(200.00))),
			CurrencyCode:      "USD",
			IsRecurring:       false,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
	}

	s.mockStore.EXPECT().
		WithTenant(gomock.Any(), s.tenantID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, tenantID uuid.UUID, fn func(context.Context, db.Store) error) error {
			s.mockStore.EXPECT().
				ListTransactions(gomock.Any(), gomock.Any()).
				Return(expectedTransactions, nil).
				Times(1)

			return fn(ctx, s.mockStore)
		}).
		Times(1)

	// Execute test
	results, err := s.repo.List(s.ctx, filter)

	// Assert results
	s.Assert().NoError(err)
	s.Assert().Len(results, 2)
	s.Assert().Equal("Transaction 1", results[0].Description)
	s.Assert().Equal("Transaction 2", results[1].Description)
}

// Helper functions (reuse from main file and add missing ones)
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
