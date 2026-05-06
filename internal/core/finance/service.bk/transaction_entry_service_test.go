package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// ============================================================================
// Suite
// ============================================================================

type EntryServiceSuite struct {
	suite.Suite
	req      *require.Assertions
	repo     *mockTransactionRepo // reused from transaction_service_test.go
	acctRepo *mockAccountsRepo
	svc      service.TransactionEntryService
	tenantID uuid.UUID
	ctx      context.Context
}

func TestEntryServiceSuite(t *testing.T) {
	suite.Run(t, new(EntryServiceSuite))
}

func (s *EntryServiceSuite) SetupTest() {
	s.req = require.New(s.T())
	s.tenantID = uuid.New()
	s.ctx = shared.WithTenantID(context.Background(), s.tenantID)

	s.repo = new(mockTransactionRepo)
	s.acctRepo = new(mockAccountsRepo)

	s.svc = service.NewTransactionEntryService(
		s.repo,
		s.acctRepo,
		tracing.NewNoOpService(),
		metrics.NewNoOpMetricsProvider(),
	)
}

func (s *EntryServiceSuite) TearDownTest() {
	s.repo.AssertExpectations(s.T())
	s.acctRepo.AssertExpectations(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

func (s *EntryServiceSuite) validEntry(txnID, acctID uuid.UUID) *domain.TransactionEntry {
	return &domain.TransactionEntry{
		ID:            uuid.New(),
		TenantID:      s.tenantID,
		TransactionID: txnID,
		EntryNumber:   1,
		AccountID:     acctID,
		DebitAmount:   decimal.NewFromInt(1000),
		CreditAmount:  decimal.Zero,
		Description:   "Test debit entry",
		ExchangeRate:  decimal.NewFromInt(1),
	}
}

// ============================================================================
// BUG-01: UpdateEntry must return post-update state, not pre-update state
// ============================================================================

func (s *EntryServiceSuite) TestUpdateEntry_ReturnsPostUpdateState() {
	entryID := uuid.New()
	txnID := uuid.New()
	acctID := uuid.New()

	preUpdate := s.validEntry(txnID, acctID)
	preUpdate.ID = entryID
	preUpdate.Description = "old description"

	postUpdate := *preUpdate
	postUpdate.Description = "new description"
	postUpdate.DebitAmount = decimal.NewFromInt(2000)

	req := domain.TransactionEntry{
		ID:            entryID,
		TenantID:      s.tenantID,
		TransactionID: txnID,
		EntryNumber:   1,
		AccountID:     acctID,
		DebitAmount:   decimal.NewFromInt(2000),
		CreditAmount:  decimal.Zero,
		Description:   "new description",
		ExchangeRate:  decimal.NewFromInt(1),
	}

	// Initial fetch (pre-update exists check)
	s.repo.On("GetEntryByID", s.ctx, entryID).Return(preUpdate, nil).Once()
	// Persist update
	s.repo.On("UpdateEntry", s.ctx, &req).Return(nil)
	// Re-fetch after update — returns fresh state
	s.repo.On("GetEntryByID", s.ctx, entryID).Return(&postUpdate, nil).Once()

	result, err := s.svc.UpdateEntry(s.ctx, entryID, req)
	s.req.NoError(err)
	s.req.NotNil(result)
	s.req.Equal("new description", result.Description, "must return post-update description, not pre-update")
	s.req.Equal(decimal.NewFromInt(2000), result.DebitAmount, "must return post-update amount")
}

// ============================================================================
// BUG-02: DeleteEntry must be rejected when parent transaction is POSTED
// ============================================================================

func (s *EntryServiceSuite) TestDeleteEntry_PostedTransaction_Rejected() {
	entryID := uuid.New()
	txnID := uuid.New()
	acctID := uuid.New()

	entry := s.validEntry(txnID, acctID)
	entry.ID = entryID
	entry.Reconciled = false // not reconciled — only parent status matters

	postedTxn := &domain.Transaction{
		ID:                txnID,
		TenantID:          s.tenantID,
		TransactionStatus: domain.TransactionStatusPosted,
	}

	s.repo.On("GetEntryByID", s.ctx, entryID).Return(entry, nil)
	s.repo.On("GetByID", s.ctx, txnID).Return(postedTxn, nil)

	err := s.svc.DeleteEntry(s.ctx, entryID)
	s.req.Error(err, "deleting an entry from a POSTED transaction must return an error")
}

// ============================================================================
// BUG-02b: DeleteEntry is allowed when parent transaction is DRAFT
// ============================================================================

func (s *EntryServiceSuite) TestDeleteEntry_DraftTransaction_Allowed() {
	entryID := uuid.New()
	txnID := uuid.New()
	acctID := uuid.New()

	entry := s.validEntry(txnID, acctID)
	entry.ID = entryID
	entry.Reconciled = false

	draftTxn := &domain.Transaction{
		ID:                txnID,
		TenantID:          s.tenantID,
		TransactionStatus: domain.TransactionStatusDraft,
	}

	s.repo.On("GetEntryByID", s.ctx, entryID).Return(entry, nil)
	s.repo.On("GetByID", s.ctx, txnID).Return(draftTxn, nil)
	s.repo.On("DeleteEntry", s.ctx, entryID).Return(nil)

	err := s.svc.DeleteEntry(s.ctx, entryID)
	s.req.NoError(err)
}

// ============================================================================
// BUG-02c: DeleteEntry must also be rejected when parent is REVERSED
// ============================================================================

func (s *EntryServiceSuite) TestDeleteEntry_ReversedTransaction_Rejected() {
	entryID := uuid.New()
	txnID := uuid.New()
	acctID := uuid.New()

	entry := s.validEntry(txnID, acctID)
	entry.ID = entryID
	entry.Reconciled = false

	reversedTxn := &domain.Transaction{
		ID:                txnID,
		TenantID:          s.tenantID,
		TransactionStatus: domain.TransactionStatusReversed,
	}

	s.repo.On("GetEntryByID", s.ctx, entryID).Return(entry, nil)
	s.repo.On("GetByID", s.ctx, txnID).Return(reversedTxn, nil)

	err := s.svc.DeleteEntry(s.ctx, entryID)
	s.req.Error(err, "deleting an entry from a REVERSED transaction must return an error")
}

// ============================================================================
// DeleteEntry: reconciled entries are always blocked regardless of txn status
// ============================================================================

func (s *EntryServiceSuite) TestDeleteEntry_ReconciledEntry_Rejected() {
	entryID := uuid.New()
	txnID := uuid.New()
	acctID := uuid.New()

	entry := s.validEntry(txnID, acctID)
	entry.ID = entryID
	entry.Reconciled = true // reconciled — blocked before parent status check

	s.repo.On("GetEntryByID", s.ctx, entryID).Return(entry, nil)

	err := s.svc.DeleteEntry(s.ctx, entryID)
	s.req.Error(err, "deleting a reconciled entry must return an error")
	// GetByID (parent txn check) must NOT be called — reconciled guard fires first
	s.repo.AssertNotCalled(s.T(), "GetByID", mock.Anything, mock.Anything)
}

// ============================================================================
// GetEntriesByTransactionID — wraps slice correctly
// ============================================================================

func (s *EntryServiceSuite) TestGetEntriesByTransactionID_ReturnsCopies() {
	txnID := uuid.New()
	acctID := uuid.New()

	raw := []domain.TransactionEntry{
		{
			ID:            uuid.New(),
			TenantID:      s.tenantID,
			TransactionID: txnID,
			EntryNumber:   1,
			AccountID:     acctID,
			DebitAmount:   decimal.NewFromInt(500),
			CreditAmount:  decimal.Zero,
			ExchangeRate:  decimal.NewFromInt(1),
		},
	}

	s.repo.On("GetEntriesByTransaction", s.ctx, txnID).Return(raw, nil)

	result, err := s.svc.GetEntriesByTransactionID(s.ctx, txnID)
	s.req.NoError(err)
	s.req.Len(result, 1)
	s.req.Equal(raw[0].ID, result[0].ID)
}

// ============================================================================
// UpdateEntry — entry not found returns error
// ============================================================================

func (s *EntryServiceSuite) TestUpdateEntry_NotFound_ReturnsError() {
	entryID := uuid.New()
	notFoundErr := errors.New("entry not found")

	s.repo.On("GetEntryByID", s.ctx, entryID).Return((*domain.TransactionEntry)(nil), notFoundErr)

	_, err := s.svc.UpdateEntry(s.ctx, entryID, domain.TransactionEntry{})
	s.req.Error(err)
}
