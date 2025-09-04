# AWO ERP Financial Module - Testing Strategy & Quality Assurance

**Version**: 1.0  
**Date**: January 2025  
**Status**: Technical Specification  

---

## Table of Contents

1. [Testing Strategy Overview](#testing-strategy-overview)
2. [Test Pyramid Implementation](#test-pyramid-implementation)
3. [Unit Testing Framework](#unit-testing-framework)
4. [Integration Testing](#integration-testing)
5. [Financial Domain Testing](#financial-domain-testing)
6. [Security & Compliance Testing](#security-compliance-testing)
7. [Performance Testing](#performance-testing)
8. [End-to-End Testing](#end-to-end-testing)
9. [Test Data Management](#test-data-management)
10. [Quality Gates & Metrics](#quality-gates-metrics)

---

## Testing Strategy Overview

### **Quality Assurance Philosophy**

The AWO ERP Financial Module testing strategy is built on the principle that **financial accuracy is non-negotiable**. Every test is designed to ensure:

- **Correctness**: All financial calculations are mathematically accurate
- **Completeness**: Every business rule is validated
- **Compliance**: Regulatory requirements are met
- **Security**: Authorization and audit requirements are enforced
- **Performance**: System meets SLA requirements under load

### **Testing Pyramid Architecture**

```
                    ┌─────────────────────┐
                    │   E2E Tests (5%)    │
                    │  Full User Workflows │
                    └─────────────────────┘
                   ┌─────────────────────────┐
                   │ Integration Tests (25%) │
                   │  Service & DB Testing   │
                   └─────────────────────────┘
              ┌─────────────────────────────────────┐
              │        Unit Tests (70%)             │
              │  Domain Logic & Business Rules      │
              └─────────────────────────────────────┘
```

### **Test Coverage Requirements**

| Layer | Coverage Target | Focus Areas |
|-------|----------------|-------------|
| **Unit Tests** | 90%+ | Domain logic, business rules, calculations |
| **Integration Tests** | 80%+ | Service interactions, database operations |
| **API Tests** | 95%+ | Contract validation, error handling |
| **E2E Tests** | Key workflows | Critical user journeys |

---

## Test Pyramid Implementation

### **Unit Tests (70% of Test Suite)**

#### **Domain Model Testing**
```go
// @internal/core/finance/domain/account_test.go
package domain_test

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/niiniyare/erp/internal/core/finance/domain"
    "github.com/niiniyare/erp/internal/core/tenant"
)

func TestAccount_Validation(t *testing.T) {
    tests := []struct {
        name        string
        account     *domain.Account
        expectError bool
        errorType   error
    }{
        {
            name: "valid account",
            account: &domain.Account{
                TenantID:    tenant.ID("tenant-1"),
                Code:        domain.AccountCode("1000"),
                Name:        "Cash",
                AccountType: domain.AccountTypeCash,
                RootType:    domain.RootTypeAsset,
                Currency:    domain.Currency("USD"),
                IsActive:    true,
            },
            expectError: false,
        },
        {
            name: "missing account code",
            account: &domain.Account{
                TenantID:    tenant.ID("tenant-1"),
                Code:        domain.AccountCode(""),
                Name:        "Cash",
                AccountType: domain.AccountTypeCash,
                RootType:    domain.RootTypeAsset,
                Currency:    domain.Currency("USD"),
            },
            expectError: true,
            errorType:   domain.ErrAccountCodeRequired,
        },
        {
            name: "invalid account type for root type",
            account: &domain.Account{
                TenantID:    tenant.ID("tenant-1"),
                Code:        domain.AccountCode("1000"),
                Name:        "Cash",
                AccountType: domain.AccountTypeIncome, // Wrong for Asset root type
                RootType:    domain.RootTypeAsset,
                Currency:    domain.Currency("USD"),
            },
            expectError: true,
            errorType:   domain.ErrInvalidAccountTypeForRootType,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.account.Validate()
            
            if tt.expectError {
                require.Error(t, err)
                if tt.errorType != nil {
                    assert.ErrorIs(t, err, tt.errorType)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestAccount_BusinessRules(t *testing.T) {
    t.Run("group account cannot have transactions", func(t *testing.T) {
        account := &domain.Account{
            IsGroup: true,
        }
        
        assert.False(t, account.CanHaveBalance())
        assert.False(t, account.CanHaveTransactions())
    })
    
    t.Run("debit nature accounts", func(t *testing.T) {
        tests := []struct {
            rootType     domain.RootType
            expectDebit  bool
        }{
            {domain.RootTypeAsset, true},
            {domain.RootTypeExpense, true},
            {domain.RootTypeLiability, false},
            {domain.RootTypeIncome, false},
            {domain.RootTypeEquity, false},
        }
        
        for _, tt := range tests {
            account := &domain.Account{RootType: tt.rootType}
            assert.Equal(t, tt.expectDebit, account.IsDebitNature())
        }
    })
}
```

#### **Transaction Testing**
```go
// @internal/core/finance/domain/transaction_test.go
func TestTransaction_DoubleEntryValidation(t *testing.T) {
    t.Run("balanced transaction is valid", func(t *testing.T) {
        transaction := &domain.Transaction{
            TenantID: tenant.ID("tenant-1"),
            Number:   domain.TransactionNumber("TXN-001"),
            Type:     domain.TransactionTypeManual,
            Entries: []domain.TransactionEntry{
                {
                    AccountID:    domain.AccountID("acc-1"),
                    DebitAmount:  decimal.NewFromFloat(1000.00),
                    CreditAmount: decimal.Zero,
                },
                {
                    AccountID:    domain.AccountID("acc-2"),
                    DebitAmount:  decimal.Zero,
                    CreditAmount: decimal.NewFromFloat(1000.00),
                },
            },
        }
        
        assert.True(t, transaction.IsBalanced())
        assert.NoError(t, transaction.Validate())
    })
    
    t.Run("unbalanced transaction is invalid", func(t *testing.T) {
        transaction := &domain.Transaction{
            TenantID: tenant.ID("tenant-1"),
            Number:   domain.TransactionNumber("TXN-002"),
            Type:     domain.TransactionTypeManual,
            Entries: []domain.TransactionEntry{
                {
                    AccountID:    domain.AccountID("acc-1"),
                    DebitAmount:  decimal.NewFromFloat(1000.00),
                    CreditAmount: decimal.Zero,
                },
                {
                    AccountID:    domain.AccountID("acc-2"),
                    DebitAmount:  decimal.Zero,
                    CreditAmount: decimal.NewFromFloat(900.00), // Unbalanced
                },
            },
        }
        
        assert.False(t, transaction.IsBalanced())
        assert.ErrorIs(t, transaction.Validate(), domain.ErrTransactionMustBeBalanced)
    })
}

func TestTransaction_StateTransitions(t *testing.T) {
    transaction := &domain.Transaction{
        Status: domain.TransactionStatusSubmitted,
        Entries: []domain.TransactionEntry{
            {DebitAmount: decimal.NewFromFloat(100), CreditAmount: decimal.Zero},
            {DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(100)},
        },
    }
    
    t.Run("can post submitted transaction", func(t *testing.T) {
        userID := identity.UserID("user-1")
        
        err := transaction.Post(userID)
        require.NoError(t, err)
        
        assert.Equal(t, domain.TransactionStatusPosted, transaction.Status)
        assert.Equal(t, &userID, transaction.ApprovedBy)
        assert.NotNil(t, transaction.ApprovedAt)
    })
    
    t.Run("cannot post already posted transaction", func(t *testing.T) {
        err := transaction.Post(identity.UserID("user-2"))
        assert.ErrorIs(t, err, domain.ErrCanOnlyPostSubmittedTransactions)
    })
}
```

### **Integration Tests (25% of Test Suite)**

#### **Service Layer Integration**
```go
// @internal/core/finance/service/account_service_integration_test.go
package service_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "github.com/niiniyare/erp/internal/core/finance/service"
    "github.com/niiniyare/erp/internal/core/finance/repository"
    "github.com/niiniyare/erp/test/testutil"
)

type AccountServiceIntegrationTestSuite struct {
    suite.Suite
    ctx     context.Context
    service service.AccountService
    store   *sqlc.Store
    cleanup func()
}

func (s *AccountServiceIntegrationTestSuite) SetupSuite() {
    s.ctx = context.Background()
    
    // Setup test database
    s.store, s.cleanup = testutil.SetupTestDB(s.T())
    
    // Setup dependencies
    cache := testutil.NewMockCache()
    logger := testutil.NewTestLogger()
    abac := testutil.NewMockABAC()
    audit := testutil.NewMockAudit()
    
    // Create service
    repo := repository.NewAccountRepository(s.store, cache, logger)
    s.service = service.NewAccountService(repo, abac, audit, logger, nil, nil)
}

func (s *AccountServiceIntegrationTestSuite) TearDownSuite() {
    s.cleanup()
}

func (s *AccountServiceIntegrationTestSuite) TestCreateAccount_Success() {
    tenantID := testutil.CreateTestTenant(s.T(), s.store)
    userID := testutil.CreateTestUser(s.T(), s.store, tenantID)
    
    cmd := service.CreateAccountCommand{
        TenantID:    tenantID,
        Code:        domain.AccountCode("1000"),
        Name:        "Cash",
        AccountType: domain.AccountTypeCash,
        RootType:    domain.RootTypeAsset,
        Currency:    domain.Currency("USD"),
        UserID:      userID,
    }
    
    account, err := s.service.CreateAccount(s.ctx, cmd)
    s.Require().NoError(err)
    s.Require().NotNil(account)
    
    s.Equal(cmd.Code, account.Code)
    s.Equal(cmd.Name, account.Name)
    s.Equal(cmd.AccountType, account.AccountType)
    s.Equal(cmd.RootType, account.RootType)
    s.True(account.IsActive)
}

func (s *AccountServiceIntegrationTestSuite) TestCreateAccount_DuplicateCode() {
    tenantID := testutil.CreateTestTenant(s.T(), s.store)
    userID := testutil.CreateTestUser(s.T(), s.store, tenantID)
    
    // Create first account
    cmd1 := service.CreateAccountCommand{
        TenantID:    tenantID,
        Code:        domain.AccountCode("1000"),
        Name:        "Cash",
        AccountType: domain.AccountTypeCash,
        RootType:    domain.RootTypeAsset,
        Currency:    domain.Currency("USD"),
        UserID:      userID,
    }
    
    _, err := s.service.CreateAccount(s.ctx, cmd1)
    s.Require().NoError(err)
    
    // Try to create duplicate
    cmd2 := cmd1
    cmd2.Name = "Different Cash Account"
    
    _, err = s.service.CreateAccount(s.ctx, cmd2)
    s.Require().Error(err)
    s.True(errors.IsConflict(err))
}

func (s *AccountServiceIntegrationTestSuite) TestCreateAccount_TenantIsolation() {
    tenant1 := testutil.CreateTestTenant(s.T(), s.store)
    tenant2 := testutil.CreateTestTenant(s.T(), s.store)
    user1 := testutil.CreateTestUser(s.T(), s.store, tenant1)
    user2 := testutil.CreateTestUser(s.T(), s.store, tenant2)
    
    // Create account in tenant1
    cmd := service.CreateAccountCommand{
        TenantID:    tenant1,
        Code:        domain.AccountCode("1000"),
        Name:        "Cash Tenant 1",
        AccountType: domain.AccountTypeCash,
        RootType:    domain.RootTypeAsset,
        Currency:    domain.Currency("USD"),
        UserID:      user1,
    }
    
    account1, err := s.service.CreateAccount(s.ctx, cmd)
    s.Require().NoError(err)
    
    // Create same account code in tenant2 (should succeed)
    cmd.TenantID = tenant2
    cmd.Name = "Cash Tenant 2"
    cmd.UserID = user2
    
    account2, err := s.service.CreateAccount(s.ctx, cmd)
    s.Require().NoError(err)
    
    s.NotEqual(account1.TenantID, account2.TenantID)
    
    // Verify tenant1 user cannot access tenant2 account
    _, err = s.service.GetAccount(
        testutil.WithUserContext(s.ctx, user1),
        tenant1,
        account2.ID,
    )
    s.Require().Error(err)
    s.True(errors.IsNotFound(err))
}

func TestAccountServiceIntegration(t *testing.T) {
    suite.Run(t, new(AccountServiceIntegrationTestSuite))
}
```

#### **Repository Integration Testing**
```go
// @internal/core/finance/repository/account_repository_integration_test.go

func (s *AccountRepositoryIntegrationTestSuite) TestAccountHierarchy() {
    tenantID := testutil.CreateTestTenant(s.T(), s.store)
    
    // Create account hierarchy
    // Assets (Group)
    //   ├── Current Assets (Group)
    //   │   ├── Cash
    //   │   └── Bank Account
    //   └── Fixed Assets (Group)
    //       └── Equipment
    
    assets, err := s.repo.Create(s.ctx, &domain.Account{
        TenantID:    tenantID,
        Code:        "1000",
        Name:        "Assets",
        AccountType: domain.AccountTypeReceivable,
        RootType:    domain.RootTypeAsset,
        Currency:    "USD",
        IsGroup:     true,
        IsActive:    true,
    })
    s.Require().NoError(err)
    
    currentAssets, err := s.repo.Create(s.ctx, &domain.Account{
        TenantID:        tenantID,
        Code:            "1100",
        Name:            "Current Assets",
        ParentAccountID: &assets.ID,
        AccountType:     domain.AccountTypeReceivable,
        RootType:        domain.RootTypeAsset,
        Currency:        "USD",
        IsGroup:         true,
        IsActive:        true,
    })
    s.Require().NoError(err)
    
    cash, err := s.repo.Create(s.ctx, &domain.Account{
        TenantID:        tenantID,
        Code:            "1110",
        Name:            "Cash",
        ParentAccountID: &currentAssets.ID,
        AccountType:     domain.AccountTypeCash,
        RootType:        domain.RootTypeAsset,
        Currency:        "USD",
        IsActive:        true,
    })
    s.Require().NoError(err)
    
    // Test hierarchy queries
    hierarchy, err := s.repo.GetHierarchy(s.ctx, tenantID)
    s.Require().NoError(err)
    s.Len(hierarchy, 3) // All accounts in hierarchy order
    
    // Test children query
    children, err := s.repo.GetChildren(s.ctx, tenantID, currentAssets.ID)
    s.Require().NoError(err)
    s.Len(children, 1)
    s.Equal(cash.ID, children[0].ID)
    
    // Test nested set model values
    s.NotNil(cash.Left)
    s.NotNil(cash.Right)
    s.Equal(2, cash.Depth) // Assets(0) -> Current Assets(1) -> Cash(2)
}

func (s *AccountRepositoryIntegrationTestSuite) TestRLSIsolation() {
    tenant1 := testutil.CreateTestTenant(s.T(), s.store)
    tenant2 := testutil.CreateTestTenant(s.T(), s.store)
    
    // Create account in tenant1
    account1, err := s.repo.Create(testutil.WithTenant(s.ctx, tenant1), &domain.Account{
        TenantID:    tenant1,
        Code:        "1000",
        Name:        "Cash Tenant 1",
        AccountType: domain.AccountTypeCash,
        RootType:    domain.RootTypeAsset,
        Currency:    "USD",
        IsActive:    true,
    })
    s.Require().NoError(err)
    
    // Try to access from tenant2 context (should fail due to RLS)
    _, err = s.repo.GetByID(testutil.WithTenant(s.ctx, tenant2), tenant2, account1.ID)
    s.Require().Error(err)
    s.True(errors.IsNotFound(err))
    
    // Verify access works with correct tenant context
    retrieved, err := s.repo.GetByID(testutil.WithTenant(s.ctx, tenant1), tenant1, account1.ID)
    s.Require().NoError(err)
    s.Equal(account1.ID, retrieved.ID)
}
```

---

## Financial Domain Testing

### **Accounting Rules Testing**

#### **Double-Entry Bookkeeping Validation**
```go
// @internal/core/finance/domain/accounting_rules_test.go

func TestDoubleEntryRules(t *testing.T) {
    tests := []struct {
        name        string
        entries     []domain.TransactionEntry
        expectValid bool
    }{
        {
            name: "simple balanced transaction",
            entries: []domain.TransactionEntry{
                {AccountID: "cash", DebitAmount: decimal.NewFromFloat(1000), CreditAmount: decimal.Zero},
                {AccountID: "revenue", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(1000)},
            },
            expectValid: true,
        },
        {
            name: "complex multi-entry transaction",
            entries: []domain.TransactionEntry{
                {AccountID: "cash", DebitAmount: decimal.NewFromFloat(1000), CreditAmount: decimal.Zero},
                {AccountID: "accounts_receivable", DebitAmount: decimal.NewFromFloat(500), CreditAmount: decimal.Zero},
                {AccountID: "revenue", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(1200)},
                {AccountID: "discount", DebitAmount: decimal.NewFromFloat(300), CreditAmount: decimal.Zero},
            },
            expectValid: true,
        },
        {
            name: "unbalanced transaction",
            entries: []domain.TransactionEntry{
                {AccountID: "cash", DebitAmount: decimal.NewFromFloat(1000), CreditAmount: decimal.Zero},
                {AccountID: "revenue", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(900)},
            },
            expectValid: false,
        },
        {
            name: "entry with both debit and credit (invalid)",
            entries: []domain.TransactionEntry{
                {AccountID: "cash", DebitAmount: decimal.NewFromFloat(1000), CreditAmount: decimal.NewFromFloat(100)},
                {AccountID: "revenue", DebitAmount: decimal.Zero, CreditAmount: decimal.NewFromFloat(1100)},
            },
            expectValid: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            transaction := &domain.Transaction{
                TenantID: tenant.ID("test-tenant"),
                Number:   domain.TransactionNumber("TEST-001"),
                Type:     domain.TransactionTypeManual,
                Entries:  tt.entries,
            }
            
            isBalanced := transaction.IsBalanced()
            err := transaction.Validate()
            
            if tt.expectValid {
                assert.True(t, isBalanced, "Transaction should be balanced")
                assert.NoError(t, err, "Transaction should be valid")
            } else {
                if !isBalanced {
                    assert.ErrorIs(t, err, domain.ErrTransactionMustBeBalanced)
                } else {
                    assert.Error(t, err, "Transaction should have validation errors")
                }
            }
        })
    }
}
```

#### **Currency and Exchange Rate Testing**
```go
func TestMoneyOperations(t *testing.T) {
    t.Run("same currency operations", func(t *testing.T) {
        money1 := domain.Money{Amount: decimal.NewFromFloat(100.00), Currency: "USD"}
        money2 := domain.Money{Amount: decimal.NewFromFloat(50.00), Currency: "USD"}
        
        result, err := money1.Add(money2)
        require.NoError(t, err)
        
        expected := domain.Money{Amount: decimal.NewFromFloat(150.00), Currency: "USD"}
        assert.True(t, result.Equals(expected))
    })
    
    t.Run("different currency operations fail", func(t *testing.T) {
        money1 := domain.Money{Amount: decimal.NewFromFloat(100.00), Currency: "USD"}
        money2 := domain.Money{Amount: decimal.NewFromFloat(50.00), Currency: "EUR"}
        
        _, err := money1.Add(money2)
        assert.ErrorIs(t, err, domain.ErrCurrencyMismatch)
    })
    
    t.Run("currency conversion", func(t *testing.T) {
        money := domain.Money{Amount: decimal.NewFromFloat(100.00), Currency: "USD"}
        rate := domain.ExchangeRate{
            FromCurrency: "USD",
            ToCurrency:   "EUR",
            Rate:         decimal.NewFromFloat(0.85),
            Date:         time.Now(),
        }
        
        converted, err := money.ConvertTo("EUR", rate)
        require.NoError(t, err)
        
        expectedAmount := decimal.NewFromFloat(85.00)
        assert.True(t, converted.Amount.Equal(expectedAmount))
        assert.Equal(t, domain.Currency("EUR"), converted.Currency)
    })
}
```

### **Business Rule Testing**

#### **Account Balance Calculations**
```go
func TestAccountBalanceCalculations(t *testing.T) {
    testCases := []struct {
        name            string
        accountType     domain.AccountType
        rootType        domain.RootType
        transactions    []TransactionForBalance
        expectedBalance decimal.Decimal
    }{
        {
            name:        "asset account with debits and credits",
            accountType: domain.AccountTypeCash,
            rootType:    domain.RootTypeAsset,
            transactions: []TransactionForBalance{
                {Type: "debit", Amount: decimal.NewFromFloat(1000.00)},
                {Type: "credit", Amount: decimal.NewFromFloat(300.00)},
                {Type: "debit", Amount: decimal.NewFromFloat(500.00)},
            },
            expectedBalance: decimal.NewFromFloat(1200.00), // 1000 - 300 + 500
        },
        {
            name:        "liability account with debits and credits",
            accountType: domain.AccountTypePayable,
            rootType:    domain.RootTypeLiability,
            transactions: []TransactionForBalance{
                {Type: "credit", Amount: decimal.NewFromFloat(2000.00)},
                {Type: "debit", Amount: decimal.NewFromFloat(500.00)},
                {Type: "credit", Amount: decimal.NewFromFloat(300.00)},
            },
            expectedBalance: decimal.NewFromFloat(1800.00), // -2000 + 500 - 300 (negative because it's credit balance)
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            account := &domain.Account{
                AccountType: tc.accountType,
                RootType:    tc.rootType,
            }
            
            balance := calculateTestBalance(account, tc.transactions)
            assert.True(t, balance.Equal(tc.expectedBalance), 
                "Expected balance %s, got %s", tc.expectedBalance, balance)
        })
    }
}
```

---

## Security & Compliance Testing

### **ABAC Authorization Testing**

#### **Policy Testing Framework**
```go
// @internal/core/finance/abac/policy_test.go

type ABACTestSuite struct {
    suite.Suite
    policyEngine abac.PolicyEngine
    testPolicies []abac.Policy
}

func (s *ABACTestSuite) SetupTest() {
    s.policyEngine = testutil.NewTestPolicyEngine()
    s.testPolicies = policies.FinancialPolicies
    
    // Load test policies
    for _, policy := range s.testPolicies {
        err := s.policyEngine.LoadPolicy(policy)
        s.Require().NoError(err)
    }
}

func (s *ABACTestSuite) TestTransactionCreationPolicies() {
    testCases := []struct {
        name           string
        subject        abac.Subject
        resource       abac.Resource
        action         string
        context        abac.Context
        expectAllowed  bool
        expectObligation bool
    }{
        {
            name: "accountant can create small transaction",
            subject: abac.Subject{
                Type: "user",
                ID:   "user-1",
                Attributes: map[string]interface{}{
                    "roles":      []string{"accountant"},
                    "department": "finance",
                },
            },
            resource: abac.Resource{
                Type: "transaction",
                Attributes: map[string]interface{}{
                    "amount": decimal.NewFromFloat(5000.00),
                    "type":   "manual",
                },
            },
            action: "create",
            context: abac.Context{
                TenantID:    tenant.ID("test-tenant"),
                RequestTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Business hours
                Environment: map[string]interface{}{
                    "location": "office",
                    "ip_address": "192.168.1.100",
                },
            },
            expectAllowed: true,
            expectObligation: false,
        },
        {
            name: "accountant cannot create large transaction",
            subject: abac.Subject{
                Type: "user",
                ID:   "user-1",
                Attributes: map[string]interface{}{
                    "roles":      []string{"accountant"},
                    "department": "finance",
                },
            },
            resource: abac.Resource{
                Type: "transaction",
                Attributes: map[string]interface{}{
                    "amount": decimal.NewFromFloat(50000.00), // Above limit
                    "type":   "manual",
                },
            },
            action: "create",
            context: abac.Context{
                TenantID:    tenant.ID("test-tenant"),
                RequestTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
            },
            expectAllowed: false,
        },
        {
            name: "finance manager can create large transaction with approval",
            subject: abac.Subject{
                Type: "user",
                ID:   "user-2",
                Attributes: map[string]interface{}{
                    "roles":      []string{"finance_manager"},
                    "department": "finance",
                },
            },
            resource: abac.Resource{
                Type: "transaction",
                Attributes: map[string]interface{}{
                    "amount": decimal.NewFromFloat(50000.00),
                    "type":   "manual",
                },
            },
            action: "create",
            context: abac.Context{
                TenantID:    tenant.ID("test-tenant"),
                RequestTime: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
            },
            expectAllowed:    true,
            expectObligation: true, // Should require dual approval
        },
    }
    
    for _, tc := range testCases {
        s.Run(tc.name, func() {
            request := abac.AuthorizationRequest{
                Subject:  tc.subject,
                Resource: tc.resource,
                Action:   tc.action,
                Context:  tc.context,
            }
            
            result, err := s.policyEngine.Evaluate(request)
            s.Require().NoError(err)
            
            s.Equal(tc.expectAllowed, result.Allowed, "Authorization result mismatch")
            
            if tc.expectObligation {
                s.NotEmpty(result.Obligations, "Expected obligations but got none")
            }
        })
    }
}

func (s *ABACTestSuite) TestSegregationOfDuties() {
    // Test that creator cannot approve their own transaction
    request := abac.AuthorizationRequest{
        Subject: abac.Subject{
            Type: "user",
            ID:   "user-1",
            Attributes: map[string]interface{}{
                "roles": []string{"finance_manager"},
            },
        },
        Resource: abac.Resource{
            Type: "transaction",
            ID:   "txn-1",
            Attributes: map[string]interface{}{
                "created_by": "user-1", // Same as subject
                "amount":     decimal.NewFromFloat(10000.00),
            },
        },
        Action: "approve",
        Context: abac.Context{
            TenantID: tenant.ID("test-tenant"),
        },
    }
    
    result, err := s.policyEngine.Evaluate(request)
    s.Require().NoError(err)
    
    s.False(result.Allowed, "Should not allow self-approval")
    s.Contains(result.Reason, "segregation of duties")
}
```

### **Audit Testing**

#### **Audit Trail Validation**
```go
// @internal/core/finance/audit/audit_test.go

func TestFinancialAuditTrail(t *testing.T) {
    mockStorage := &MockAuditStorage{}
    auditor := audit.NewFinancialAuditor(mockStorage)
    
    t.Run("account creation generates audit event", func(t *testing.T) {
        account := &domain.Account{
            ID:          domain.AccountID("acc-1"),
            TenantID:    tenant.ID("tenant-1"),
            Code:        "1000",
            Name:        "Cash",
            AccountType: domain.AccountTypeCash,
            RootType:    domain.RootTypeAsset,
        }
        
        err := auditor.LogAccountCreation(context.Background(), account, identity.UserID("user-1"))
        require.NoError(t, err)
        
        // Verify audit event was created
        events := mockStorage.GetEvents()
        require.Len(t, events, 1)
        
        event := events[0]
        assert.Equal(t, "finance.account.created", event.Type)
        assert.Equal(t, string(account.TenantID), event.TenantID)
        assert.Equal(t, string(account.ID), event.ResourceID)
        assert.Contains(t, event.ComplianceFrameworks, "SOX")
        
        // Verify sensitive data is encrypted
        assert.NotEmpty(t, event.EncryptedData)
        assert.Contains(t, event.EncryptedData, "account_name")
    })
    
    t.Run("high-risk transaction generates audit", func(t *testing.T) {
        transaction := &domain.Transaction{
            ID:          domain.TransactionID("txn-1"),
            TenantID:    tenant.ID("tenant-1"),
            TotalAmount: decimal.NewFromFloat(100000.00), // High amount
            Type:        domain.TransactionTypeManual,
            CreatedBy:   identity.UserID("user-1"),
        }
        
        err := auditor.LogTransactionCreation(context.Background(), transaction, transaction.CreatedBy)
        require.NoError(t, err)
        
        events := mockStorage.GetEvents()
        require.Len(t, events, 1)
        
        event := events[0]
        assert.Equal(t, audit.RiskHigh, event.RiskScore)
        assert.Contains(t, event.Metadata, "requires_approval")
        assert.True(t, event.Metadata["requires_approval"].(bool))
    })
}
```

---

## Performance Testing

### **Load Testing Framework**

#### **Transaction Processing Performance**
```go
// @test/performance/transaction_load_test.go

func TestTransactionProcessingLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }
    
    // Setup
    service := setupTransactionService(t)
    tenantID := createTestTenant(t)
    userID := createTestUser(t, tenantID)
    
    // Test parameters
    concurrentUsers := 50
    transactionsPerUser := 100
    maxDuration := 5 * time.Minute
    
    // Metrics collection
    var (
        successCount int64
        errorCount   int64
        totalLatency int64
        maxLatency   int64
        minLatency   int64 = math.MaxInt64
    )
    
    // Start load test
    startTime := time.Now()
    var wg sync.WaitGroup
    
    for i := 0; i < concurrentUsers; i++ {
        wg.Add(1)
        go func(userIndex int) {
            defer wg.Done()
            
            for j := 0; j < transactionsPerUser; j++ {
                if time.Since(startTime) > maxDuration {
                    return
                }
                
                // Create test transaction
                cmd := service.CreateTransactionCommand{
                    TenantID:    tenantID,
                    Type:        domain.TransactionTypeManual,
                    PostingDate: time.Now(),
                    Currency:    "USD",
                    ExchangeRate: decimal.NewFromFloat(1.0),
                    Entries: []service.CreateTransactionEntryCommand{
                        {
                            AccountID:    createRandomAccount(t, tenantID),
                            DebitAmount:  decimal.NewFromFloat(100.00),
                            CreditAmount: decimal.Zero,
                        },
                        {
                            AccountID:    createRandomAccount(t, tenantID),
                            DebitAmount:  decimal.Zero,
                            CreditAmount: decimal.NewFromFloat(100.00),
                        },
                    },
                    UserID: userID,
                }
                
                // Measure latency
                txnStart := time.Now()
                _, err := service.CreateTransaction(context.Background(), cmd)
                latency := time.Since(txnStart).Nanoseconds()
                
                // Update metrics
                if err != nil {
                    atomic.AddInt64(&errorCount, 1)
                } else {
                    atomic.AddInt64(&successCount, 1)
                }
                
                atomic.AddInt64(&totalLatency, latency)
                
                // Update min/max latency
                for {
                    current := atomic.LoadInt64(&maxLatency)
                    if latency <= current || atomic.CompareAndSwapInt64(&maxLatency, current, latency) {
                        break
                    }
                }
                
                for {
                    current := atomic.LoadInt64(&minLatency)
                    if latency >= current || atomic.CompareAndSwapInt64(&minLatency, current, latency) {
                        break
                    }
                }
            }
        }(i)
    }
    
    wg.Wait()
    
    // Calculate results
    totalTransactions := successCount + errorCount
    avgLatency := time.Duration(totalLatency / totalTransactions)
    errorRate := float64(errorCount) / float64(totalTransactions) * 100
    throughput := float64(totalTransactions) / time.Since(startTime).Seconds()
    
    // Assertions
    assert.Less(t, errorRate, 1.0, "Error rate should be less than 1%%")
    assert.Less(t, avgLatency, 50*time.Millisecond, "Average latency should be under 50ms")
    assert.Greater(t, throughput, 100.0, "Throughput should be over 100 TPS")
    
    t.Logf("Load test results:")
    t.Logf("  Total transactions: %d", totalTransactions)
    t.Logf("  Success rate: %.2f%%", 100-errorRate)
    t.Logf("  Average latency: %v", avgLatency)
    t.Logf("  Min latency: %v", time.Duration(minLatency))
    t.Logf("  Max latency: %v", time.Duration(maxLatency))
    t.Logf("  Throughput: %.2f TPS", throughput)
}
```

#### **Balance Calculation Performance**
```go
func BenchmarkAccountBalanceCalculation(b *testing.B) {
    service := setupTransactionService(b)
    tenantID := createTestTenant(b)
    accountID := createTestAccount(b, tenantID)
    
    // Create many transactions for the account
    createTransactionsForAccount(b, service, tenantID, accountID, 10000)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := service.GetAccountBalance(context.Background(), tenantID, accountID, time.Now())
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkTrialBalanceGeneration(b *testing.B) {
    service := setupTransactionService(b)
    tenantID := createTestTenant(b)
    
    // Create many accounts and transactions
    setupLargeChartOfAccounts(b, service, tenantID, 1000)
    createManyTransactions(b, service, tenantID, 50000)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := service.GetTrialBalance(context.Background(), tenantID, time.Now(), false)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

---

## Test Data Management

### **Test Data Builder Pattern**

#### **Financial Test Data Builders**
```go
// @test/testutil/financial_builders.go

type AccountBuilder struct {
    account *domain.Account
}

func NewAccountBuilder() *AccountBuilder {
    return &AccountBuilder{
        account: &domain.Account{
            TenantID:    tenant.ID("test-tenant"),
            Code:        domain.AccountCode("1000"),
            Name:        "Test Account",
            AccountType: domain.AccountTypeCash,
            RootType:    domain.RootTypeAsset,
            Currency:    domain.Currency("USD"),
            IsActive:    true,
            CreatedAt:   time.Now(),
            UpdatedAt:   time.Now(),
            Version:     1,
        },
    }
}

func (b *AccountBuilder) WithTenant(tenantID tenant.ID) *AccountBuilder {
    b.account.TenantID = tenantID
    return b
}

func (b *AccountBuilder) WithCode(code string) *AccountBuilder {
    b.account.Code = domain.AccountCode(code)
    return b
}

func (b *AccountBuilder) WithName(name string) *AccountBuilder {
    b.account.Name = name
    return b
}

func (b *AccountBuilder) WithType(accountType domain.AccountType, rootType domain.RootType) *AccountBuilder {
    b.account.AccountType = accountType
    b.account.RootType = rootType
    return b
}

func (b *AccountBuilder) AsGroup() *AccountBuilder {
    b.account.IsGroup = true
    return b
}

func (b *AccountBuilder) WithParent(parentID domain.AccountID) *AccountBuilder {
    b.account.ParentAccountID = &parentID
    return b
}

func (b *AccountBuilder) Build() *domain.Account {
    return b.account
}

type TransactionBuilder struct {
    transaction *domain.Transaction
}

func NewTransactionBuilder() *TransactionBuilder {
    return &TransactionBuilder{
        transaction: &domain.Transaction{
            TenantID:     tenant.ID("test-tenant"),
            Number:       domain.TransactionNumber("TEST-001"),
            Type:         domain.TransactionTypeManual,
            Status:       domain.TransactionStatusDraft,
            PostingDate:  time.Now(),
            Currency:     domain.Currency("USD"),
            ExchangeRate: decimal.NewFromFloat(1.0),
            TotalAmount:  decimal.Zero,
            Entries:      []domain.TransactionEntry{},
            CreatedBy:    identity.UserID("test-user"),
            CreatedAt:    time.Now(),
            UpdatedAt:    time.Now(),
            Version:      1,
        },
    }
}

func (b *TransactionBuilder) WithTenant(tenantID tenant.ID) *TransactionBuilder {
    b.transaction.TenantID = tenantID
    return b
}

func (b *TransactionBuilder) WithType(txnType domain.TransactionType) *TransactionBuilder {
    b.transaction.Type = txnType
    return b
}

func (b *TransactionBuilder) WithEntry(accountID domain.AccountID, debit, credit decimal.Decimal) *TransactionBuilder {
    entry := domain.TransactionEntry{
        AccountID:    accountID,
        DebitAmount:  debit,
        CreditAmount: credit,
        LineNumber:   len(b.transaction.Entries) + 1,
    }
    
    b.transaction.Entries = append(b.transaction.Entries, entry)
    
    // Recalculate total amount
    totalDebit := decimal.Zero
    for _, e := range b.transaction.Entries {
        totalDebit = totalDebit.Add(e.DebitAmount)
    }
    b.transaction.TotalAmount = totalDebit
    
    return b
}

func (b *TransactionBuilder) WithBalancedEntries(amount decimal.Decimal, debitAccount, creditAccount domain.AccountID) *TransactionBuilder {
    b.WithEntry(debitAccount, amount, decimal.Zero)
    b.WithEntry(creditAccount, decimal.Zero, amount)
    return b
}

func (b *TransactionBuilder) AsPosted(approverID identity.UserID) *TransactionBuilder {
    b.transaction.Status = domain.TransactionStatusPosted
    b.transaction.ApprovedBy = &approverID
    now := time.Now()
    b.transaction.ApprovedAt = &now
    return b
}

func (b *TransactionBuilder) Build() *domain.Transaction {
    return b.transaction
}

// Standard test data generators
func CreateStandardChartOfAccounts(t *testing.T, repo repository.AccountRepository, tenantID tenant.ID) map[string]*domain.Account {
    accounts := make(map[string]*domain.Account)
    
    // Assets
    cash, err := repo.Create(context.Background(), NewAccountBuilder().
        WithTenant(tenantID).
        WithCode("1000").
        WithName("Cash").
        WithType(domain.AccountTypeCash, domain.RootTypeAsset).
        Build())
    require.NoError(t, err)
    accounts["cash"] = cash
    
    // Liabilities
    payable, err := repo.Create(context.Background(), NewAccountBuilder().
        WithTenant(tenantID).
        WithCode("2000").
        WithName("Accounts Payable").
        WithType(domain.AccountTypePayable, domain.RootTypeLiability).
        Build())
    require.NoError(t, err)
    accounts["payable"] = payable
    
    // Income
    revenue, err := repo.Create(context.Background(), NewAccountBuilder().
        WithTenant(tenantID).
        WithCode("4000").
        WithName("Revenue").
        WithType(domain.AccountTypeIncome, domain.RootTypeIncome).
        Build())
    require.NoError(t, err)
    accounts["revenue"] = revenue
    
    return accounts
}
```

---

## Quality Gates & Metrics

### **Automated Quality Checks**

#### **CI/CD Pipeline Integration**
```yaml
# .github/workflows/financial-module-tests.yml
name: Financial Module Tests

on:
  push:
    paths:
      - 'internal/core/finance/**'
      - 'db/migration/*finance*'
      - 'db/queries/*finance*'
  pull_request:
    paths:
      - 'internal/core/finance/**'

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
          POSTGRES_USER: testuser
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run Financial Unit Tests
      run: |
        go test ./internal/core/finance/domain/... -v -coverprofile=domain.out
        go test ./internal/core/finance/service/... -v -coverprofile=service.out
        
    - name: Check Domain Coverage
      run: |
        go tool cover -func=domain.out | grep total | awk '{if($3+0 < 90) exit 1}'
        
    - name: Check Service Coverage  
      run: |
        go tool cover -func=service.out | grep total | awk '{if($3+0 < 85) exit 1}'
    
    - name: Run Integration Tests
      env:
        DATABASE_URL: postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable
      run: |
        make test-integration-finance
        
    - name: Run Performance Benchmarks
      run: |
        go test -bench=. ./internal/core/finance/... -benchmem -timeout=10m
        
    - name: Security Scan
      run: |
        gosec ./internal/core/finance/...
        
    - name: Compliance Check
      run: |
        # Check for proper audit logging
        grep -r "audit.LogEvent" internal/core/finance/service/ || exit 1
        # Check for ABAC integration
        grep -r "abac.Authorize" internal/core/finance/service/ || exit 1
```

### **Test Metrics Collection**

#### **Coverage and Quality Metrics**
```go
// @test/metrics/test_metrics.go

type TestMetrics struct {
    CoverageByPackage map[string]float64
    TestCounts        map[string]int
    BenchmarkResults  map[string]BenchmarkResult
    QualityScores     map[string]float64
}

type BenchmarkResult struct {
    Name           string
    NsPerOp        int64
    AllocsPerOp    int64
    BytesPerOp     int64
    MBPerSec       float64
}

func CollectFinancialModuleMetrics() (*TestMetrics, error) {
    metrics := &TestMetrics{
        CoverageByPackage: make(map[string]float64),
        TestCounts:        make(map[string]int),
        BenchmarkResults:  make(map[string]BenchmarkResult),
        QualityScores:     make(map[string]float64),
    }
    
    // Collect coverage metrics
    domainCoverage, err := getCoverageForPackage("internal/core/finance/domain")
    if err != nil {
        return nil, err
    }
    metrics.CoverageByPackage["domain"] = domainCoverage
    
    serviceCoverage, err := getCoverageForPackage("internal/core/finance/service")
    if err != nil {
        return nil, err
    }
    metrics.CoverageByPackage["service"] = serviceCoverage
    
    // Collect test counts
    metrics.TestCounts["unit"] = countTestsInPackage("internal/core/finance/domain")
    metrics.TestCounts["integration"] = countTestsInPackage("internal/core/finance/service")
    
    // Run and collect benchmark results
    benchmarks := []string{
        "BenchmarkAccountBalanceCalculation",
        "BenchmarkTrialBalanceGeneration",
        "BenchmarkTransactionValidation",
    }
    
    for _, benchmark := range benchmarks {
        result, err := runBenchmark(benchmark)
        if err != nil {
            return nil, err
        }
        metrics.BenchmarkResults[benchmark] = result
    }
    
    // Calculate quality scores
    metrics.QualityScores["overall"] = calculateOverallQualityScore(metrics)
    
    return metrics, nil
}

// Quality gates
func ValidateQualityGates(metrics *TestMetrics) []string {
    var violations []string
    
    // Coverage gates
    if metrics.CoverageByPackage["domain"] < 90.0 {
        violations = append(violations, 
            fmt.Sprintf("Domain coverage %.1f%% below 90%% threshold", 
                metrics.CoverageByPackage["domain"]))
    }
    
    if metrics.CoverageByPackage["service"] < 85.0 {
        violations = append(violations, 
            fmt.Sprintf("Service coverage %.1f%% below 85%% threshold", 
                metrics.CoverageByPackage["service"]))
    }
    
    // Performance gates
    balanceBenchmark := metrics.BenchmarkResults["BenchmarkAccountBalanceCalculation"]
    if balanceBenchmark.NsPerOp > 50_000_000 { // 50ms
        violations = append(violations, 
            fmt.Sprintf("Balance calculation too slow: %dms > 50ms", 
                balanceBenchmark.NsPerOp/1_000_000))
    }
    
    // Test count gates
    if metrics.TestCounts["unit"] < 50 {
        violations = append(violations, 
            fmt.Sprintf("Insufficient unit tests: %d < 50", metrics.TestCounts["unit"]))
    }
    
    return violations
}
```

This  testing strategy ensures that the AWO ERP Financial Module maintains the highest standards of quality, security, and performance while providing confidence in the system's correctness and reliability.