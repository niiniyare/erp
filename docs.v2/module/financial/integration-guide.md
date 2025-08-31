# AWO ERP Financial Module - Integration Guide

**Version**: 1.0  
**Date**: January 2025  
**Status**: Technical Integration Manual  

---

## Table of Contents

1. [Integration Architecture](#integration-architecture)
2. [ERP Module Integration](#erp-module-integration)
3. [External System Integration](#external-system-integration)
4. [Event-Driven Integration](#event-driven-integration)
5. [Data Synchronization](#data-synchronization)
6. [API Integration Patterns](#api-integration-patterns)
7. [Workflow Integration](#workflow-integration)
8. [Real-time Integration](#real-time-integration)
9. [Integration Testing](#integration-testing)
10. [Troubleshooting Integration Issues](#troubleshooting-integration-issues)

---

## Integration Architecture

### **Multi-Layer Integration Strategy**

```
┌─────────────────────────────────────────────────────────────────┐
│                    External Systems Layer                      │
│  Banks, Payment Gateways, Tax Services, Accounting Software   │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Integration Gateway                         │
│  API Management, Protocol Translation, Security, Rate Limiting │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Event Bus Layer                          │
│   Redis Streams, Event Routing, Dead Letter Queue            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Financial Module Core                       │
│     Account Management, Transactions, Reporting               │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Internal ERP Modules                         │
│  Sales, Purchasing, Inventory, HR, CRM, Project Management    │
└─────────────────────────────────────────────────────────────────┘
```

### **Integration Patterns**
```yaml
Pattern Types:
  Event-Driven: Asynchronous communication via event streams
  Request-Response: Synchronous API calls for immediate data
  Batch Processing: Scheduled data transfers and reconciliation
  Real-time Streaming: Continuous data flow for real-time updates
  
Communication Protocols:
  Internal: gRPC, Redis Streams, Message Queues
  External: REST APIs, SOAP, EDI, SFTP, WebHooks
  
Data Formats:
  JSON: Primary format for API communication
  XML: Legacy system integration
  CSV: Batch file transfers
  EDI: B2B financial transactions
```

---

## ERP Module Integration

### **Sales Module Integration**

#### **Sales Order to Invoice Flow**
```go
// Event-driven integration for sales invoice creation
package integration

import (
    "context"
    "encoding/json"
    
    "github.com/niiniyare/erp/internal/core/finance/domain"
    "github.com/niiniyare/erp/internal/core/sales/events"
)

type SalesIntegrationHandler struct {
    financialService FinancialService
    eventBus        EventBus
}

// Handle sales order fulfillment
func (h *SalesIntegrationHandler) HandleSalesOrderFulfilled(ctx context.Context, event events.SalesOrderFulfilled) error {
    // Create sales invoice from fulfilled sales order
    invoice := &domain.SalesInvoice{
        TenantID:         event.TenantID,
        CustomerID:       event.CustomerID,
        SalesOrderID:     event.SalesOrderID,
        InvoiceDate:      event.FulfillmentDate,
        DueDate:          calculateDueDate(event.FulfillmentDate, event.PaymentTerms),
        Currency:         event.Currency,
        ExchangeRate:     event.ExchangeRate,
        LineItems:        convertOrderItems(event.LineItems),
        TaxCalculations:  calculateTaxes(event.LineItems, event.TaxJurisdiction),
        TotalAmount:      event.TotalAmount,
        Status:          domain.InvoiceStatusDraft,
    }
    
    // Create invoice through financial service
    createdInvoice, err := h.financialService.CreateSalesInvoice(ctx, invoice)
    if err != nil {
        return err
    }
    
    // Publish invoice created event
    return h.eventBus.Publish(ctx, events.SalesInvoiceCreated{
        TenantID:      event.TenantID,
        InvoiceID:     createdInvoice.ID,
        SalesOrderID:  event.SalesOrderID,
        CustomerID:    event.CustomerID,
        Amount:        createdInvoice.TotalAmount,
        CreatedAt:     createdInvoice.CreatedAt,
    })
}

// Handle payment received from sales
func (h *SalesIntegrationHandler) HandlePaymentReceived(ctx context.Context, event events.PaymentReceived) error {
    // Apply payment to outstanding invoices
    payment := &domain.CustomerPayment{
        TenantID:       event.TenantID,
        CustomerID:     event.CustomerID,
        PaymentDate:    event.PaymentDate,
        Amount:         event.Amount,
        Currency:       event.Currency,
        PaymentMethod:  event.PaymentMethod,
        Reference:      event.Reference,
        BankAccountID:  event.BankAccountID,
    }
    
    return h.financialService.ProcessCustomerPayment(ctx, payment)
}
```

#### **Revenue Recognition Integration**
```go
// Revenue recognition based on sales fulfillment
func (h *SalesIntegrationHandler) HandleRevenueRecognition(ctx context.Context, event events.RevenueRecognitionRequired) error {
    // Create journal entry for revenue recognition
    transaction := &domain.Transaction{
        TenantID:        event.TenantID,
        TransactionType: domain.TransactionTypeRevenueRecognition,
        Date:           event.RecognitionDate,
        Reference:      fmt.Sprintf("REV-REC-%s", event.SalesOrderID),
        Description:    fmt.Sprintf("Revenue recognition for SO %s", event.SalesOrderNumber),
        Currency:       event.Currency,
        Entries: []domain.TransactionEntry{
            {
                AccountID:    event.RevenueAccountID,
                CreditAmount: event.RevenueAmount,
                Description:  "Revenue recognition",
                CostCenter:   event.CostCenter,
                Department:   event.Department,
            },
            {
                AccountID:   event.DeferredRevenueAccountID,
                DebitAmount: event.RevenueAmount,
                Description: "Deferred revenue reduction",
            },
        },
    }
    
    return h.financialService.CreateTransaction(ctx, transaction)
}
```

### **Purchasing Module Integration**

#### **Purchase Order to Invoice Flow**
```go
// Purchase invoice creation and three-way matching
func (h *PurchasingIntegrationHandler) HandleGoodsReceived(ctx context.Context, event events.GoodsReceived) error {
    // Trigger three-way matching process
    matchingRequest := &domain.ThreeWayMatchingRequest{
        TenantID:        event.TenantID,
        PurchaseOrderID: event.PurchaseOrderID,
        ReceiptID:       event.ReceiptID,
        VendorID:        event.VendorID,
        ReceivedItems:   event.ReceivedItems,
        ReceivedDate:    event.ReceivedDate,
    }
    
    result, err := h.financialService.ProcessThreeWayMatching(ctx, matchingRequest)
    if err != nil {
        return err
    }
    
    // If matching is successful, create accrual entry
    if result.Status == domain.MatchingStatusMatched {
        return h.createAccrualEntry(ctx, event)
    }
    
    // If exceptions exist, route for approval
    if result.Status == domain.MatchingStatusException {
        return h.routeForApproval(ctx, event, result.Exceptions)
    }
    
    return nil
}

func (h *PurchasingIntegrationHandler) createAccrualEntry(ctx context.Context, event events.GoodsReceived) error {
    transaction := &domain.Transaction{
        TenantID:        event.TenantID,
        TransactionType: domain.TransactionTypeAccrual,
        Date:           event.ReceivedDate,
        Reference:      fmt.Sprintf("ACCRUAL-%s", event.PurchaseOrderID),
        Description:    fmt.Sprintf("Goods received accrual for PO %s", event.PurchaseOrderNumber),
        Currency:       event.Currency,
        Entries:        h.buildAccrualEntries(event),
    }
    
    return h.financialService.CreateTransaction(ctx, transaction)
}
```

### **Inventory Module Integration**

#### **Inventory Valuation Updates**
```go
// Real-time inventory valuation updates
func (h *InventoryIntegrationHandler) HandleInventoryMovement(ctx context.Context, event events.InventoryMovement) error {
    // Calculate inventory valuation impact
    valuationUpdate := &domain.InventoryValuationUpdate{
        TenantID:     event.TenantID,
        ProductID:    event.ProductID,
        MovementType: event.MovementType,
        Quantity:     event.Quantity,
        UnitCost:     event.UnitCost,
        MovementDate: event.MovementDate,
        LocationID:   event.LocationID,
    }
    
    // Update inventory accounts based on movement type
    switch event.MovementType {
    case "receipt":
        return h.processInventoryReceipt(ctx, valuationUpdate)
    case "issue":
        return h.processInventoryIssue(ctx, valuationUpdate)
    case "adjustment":
        return h.processInventoryAdjustment(ctx, valuationUpdate)
    case "transfer":
        return h.processInventoryTransfer(ctx, valuationUpdate)
    }
    
    return nil
}

func (h *InventoryIntegrationHandler) processInventoryReceipt(ctx context.Context, update *domain.InventoryValuationUpdate) error {
    totalValue := update.Quantity * update.UnitCost
    
    transaction := &domain.Transaction{
        TenantID:        update.TenantID,
        TransactionType: domain.TransactionTypeInventoryReceipt,
        Date:           update.MovementDate,
        Reference:      fmt.Sprintf("INV-RCPT-%s", update.ProductID),
        Description:    "Inventory receipt valuation",
        Currency:       "USD", // Or get from product/location
        Entries: []domain.TransactionEntry{
            {
                AccountID:   h.getInventoryAccount(update.ProductID),
                DebitAmount: totalValue,
                Description: fmt.Sprintf("Inventory receipt - %s", update.ProductID),
                CostCenter:  h.getCostCenter(update.LocationID),
            },
            {
                AccountID:    h.getInventoryReceiptAccount(),
                CreditAmount: totalValue,
                Description:  "Inventory receipt offset",
            },
        },
    }
    
    return h.financialService.CreateTransaction(ctx, transaction)
}
```

---

## External System Integration

### **Banking Integration**

#### **Bank Statement Import**
```go
// Automated bank statement processing
package external

import (
    "encoding/csv"
    "time"
)

type BankStatementImporter struct {
    bankService    BankService
    reconciler     ReconciliationService
    notifier       NotificationService
}

// Import bank statement from SFTP
func (b *BankStatementImporter) ImportDailyStatements(ctx context.Context) error {
    // Connect to bank SFTP server
    sftpClient, err := b.connectToBank()
    if err != nil {
        return err
    }
    defer sftpClient.Close()
    
    // Download new statement files
    files, err := sftpClient.ListFiles("/statements/")
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if b.isNewStatement(file.Name) {
            err := b.processStatementFile(ctx, file)
            if err != nil {
                b.notifier.SendAlert(ctx, "Statement Import Failed", err.Error())
                continue
            }
        }
    }
    
    return nil
}

func (b *BankStatementImporter) processStatementFile(ctx context.Context, file StatementFile) error {
    // Parse statement file
    transactions, err := b.parseStatementFile(file)
    if err != nil {
        return err
    }
    
    // Import transactions
    for _, txn := range transactions {
        bankTransaction := &domain.BankTransaction{
            BankAccountID:   txn.AccountNumber,
            TransactionDate: txn.Date,
            Amount:          txn.Amount,
            Description:     txn.Description,
            Reference:       txn.Reference,
            Type:           txn.Type,
        }
        
        // Create bank transaction record
        _, err := b.bankService.CreateBankTransaction(ctx, bankTransaction)
        if err != nil {
            return err
        }
    }
    
    // Trigger automatic reconciliation
    return b.reconciler.StartAutoReconciliation(ctx, file.AccountNumber, file.StatementDate)
}
```

#### **Payment Gateway Integration**
```go
// Payment processing integration
type PaymentGatewayIntegrator struct {
    gateway        PaymentGateway
    financialSvc   FinancialService
    cryptoService  CryptoService
}

// Process payment through gateway
func (p *PaymentGatewayIntegrator) ProcessPayment(ctx context.Context, payment *domain.PaymentRequest) (*domain.PaymentResult, error) {
    // Encrypt sensitive payment data
    encryptedData, err := p.cryptoService.Encrypt(payment.PaymentDetails)
    if err != nil {
        return nil, err
    }
    
    // Call payment gateway
    gatewayRequest := &PaymentGatewayRequest{
        MerchantID:     p.gateway.MerchantID,
        Amount:         payment.Amount,
        Currency:       payment.Currency,
        PaymentMethod:  payment.PaymentMethod,
        EncryptedData:  encryptedData,
        CallbackURL:    p.gateway.CallbackURL,
        Reference:      payment.Reference,
    }
    
    response, err := p.gateway.ProcessPayment(ctx, gatewayRequest)
    if err != nil {
        return nil, err
    }
    
    // Record payment transaction
    if response.Status == "approved" {
        return p.recordSuccessfulPayment(ctx, payment, response)
    }
    
    return p.recordFailedPayment(ctx, payment, response)
}

// Handle payment gateway webhook
func (p *PaymentGatewayIntegrator) HandleWebhook(ctx context.Context, webhook *PaymentWebhook) error {
    // Verify webhook signature
    if !p.verifyWebhookSignature(webhook) {
        return errors.New("invalid webhook signature")
    }
    
    switch webhook.EventType {
    case "payment.completed":
        return p.handlePaymentCompleted(ctx, webhook)
    case "payment.failed":
        return p.handlePaymentFailed(ctx, webhook)
    case "refund.processed":
        return p.handleRefundProcessed(ctx, webhook)
    }
    
    return nil
}
```

### **Tax Service Integration**

#### **Real-time Tax Calculation**
```go
// Tax calculation service integration
type TaxServiceIntegrator struct {
    taxService     ExternalTaxService
    configService  ConfigService
}

func (t *TaxServiceIntegrator) CalculateTax(ctx context.Context, request *domain.TaxCalculationRequest) (*domain.TaxCalculationResult, error) {
    // Get tax configuration for tenant/jurisdiction
    config, err := t.configService.GetTaxConfig(ctx, request.TenantID, request.Jurisdiction)
    if err != nil {
        return nil, err
    }
    
    // Prepare tax service request
    taxRequest := &ExternalTaxRequest{
        APIKey:           config.APIKey,
        TransactionDate:  request.TransactionDate,
        CustomerAddress:  request.CustomerAddress,
        VendorAddress:    request.VendorAddress,
        LineItems:        t.convertLineItems(request.LineItems),
        TransactionType:  request.TransactionType,
        CurrencyCode:     request.Currency,
    }
    
    // Call external tax service
    response, err := t.taxService.CalculateTax(ctx, taxRequest)
    if err != nil {
        // Fallback to internal tax calculation
        return t.calculateTaxFallback(ctx, request)
    }
    
    // Convert response to domain model
    return &domain.TaxCalculationResult{
        TotalTax:     response.TotalTax,
        TaxBreakdown: t.convertTaxBreakdown(response.TaxDetails),
        Jurisdiction: response.Jurisdiction,
        TaxRate:      response.EffectiveRate,
    }, nil
}
```

---

## Event-Driven Integration

### **Event Bus Configuration**
```go
// Redis Streams-based event bus
package events

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/go-redis/redis/v8"
)

type EventBus struct {
    client   *redis.Client
    handlers map[string][]EventHandler
}

type Event interface {
    EventType() string
    TenantID() string
    Timestamp() time.Time
}

type EventHandler interface {
    Handle(ctx context.Context, event Event) error
}

// Publish event to stream
func (e *EventBus) Publish(ctx context.Context, event Event) error {
    streamKey := fmt.Sprintf("events:%s", event.EventType())
    
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    return e.client.XAdd(ctx, &redis.XAddArgs{
        Stream: streamKey,
        Values: map[string]interface{}{
            "tenant_id": event.TenantID(),
            "event_type": event.EventType(),
            "data": string(data),
            "timestamp": event.Timestamp().Unix(),
        },
    }).Err()
}

// Subscribe to event stream
func (e *EventBus) Subscribe(ctx context.Context, eventType string, consumerGroup string) error {
    streamKey := fmt.Sprintf("events:%s", eventType)
    
    // Create consumer group if not exists
    e.client.XGroupCreateMkStream(ctx, streamKey, consumerGroup, "0")
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Read from stream
            streams, err := e.client.XReadGroup(ctx, &redis.XReadGroupArgs{
                Group:    consumerGroup,
                Consumer: "financial-service",
                Streams:  []string{streamKey, ">"},
                Count:    10,
                Block:    time.Second,
            }).Result()
            
            if err != nil {
                continue
            }
            
            // Process messages
            for _, stream := range streams {
                for _, message := range stream.Messages {
                    err := e.processMessage(ctx, eventType, message)
                    if err != nil {
                        // Handle error, possibly move to dead letter queue
                        continue
                    }
                    
                    // Acknowledge message
                    e.client.XAck(ctx, streamKey, consumerGroup, message.ID)
                }
            }
        }
    }
}
```

### **Financial Events**
```go
// Financial domain events
package financial_events

// Account events
type AccountCreated struct {
    TenantID    string    `json:"tenant_id"`
    AccountID   string    `json:"account_id"`
    AccountCode string    `json:"account_code"`
    AccountName string    `json:"account_name"`
    AccountType string    `json:"account_type"`
    CreatedAt   time.Time `json:"created_at"`
}

func (e AccountCreated) EventType() string { return "financial.account.created" }
func (e AccountCreated) TenantID() string  { return e.TenantID }
func (e AccountCreated) Timestamp() time.Time { return e.CreatedAt }

// Transaction events
type TransactionPosted struct {
    TenantID        string              `json:"tenant_id"`
    TransactionID   string              `json:"transaction_id"`
    TransactionType string              `json:"transaction_type"`
    Amount          decimal.Decimal     `json:"amount"`
    Currency        string              `json:"currency"`
    Entries         []TransactionEntry  `json:"entries"`
    PostedAt        time.Time           `json:"posted_at"`
}

func (e TransactionPosted) EventType() string { return "financial.transaction.posted" }
func (e TransactionPosted) TenantID() string  { return e.TenantID }
func (e TransactionPosted) Timestamp() time.Time { return e.PostedAt }

// Invoice events
type InvoiceCreated struct {
    TenantID    string          `json:"tenant_id"`
    InvoiceID   string          `json:"invoice_id"`
    CustomerID  string          `json:"customer_id"`
    Amount      decimal.Decimal `json:"amount"`
    Currency    string          `json:"currency"`
    DueDate     time.Time       `json:"due_date"`
    CreatedAt   time.Time       `json:"created_at"`
}

func (e InvoiceCreated) EventType() string { return "financial.invoice.created" }
func (e InvoiceCreated) TenantID() string  { return e.TenantID }
func (e InvoiceCreated) Timestamp() time.Time { return e.CreatedAt }
```

---

## Data Synchronization

### **Master Data Synchronization**
```go
// Customer master data synchronization
type CustomerSyncService struct {
    crmService      CRMService
    financialSvc    FinancialService
    syncRepository  SyncRepository
}

func (c *CustomerSyncService) SyncCustomerData(ctx context.Context) error {
    // Get last sync timestamp
    lastSync, err := c.syncRepository.GetLastSyncTime(ctx, "customer_sync")
    if err != nil {
        return err
    }
    
    // Fetch updated customers from CRM
    customers, err := c.crmService.GetUpdatedCustomers(ctx, lastSync)
    if err != nil {
        return err
    }
    
    // Process each customer
    for _, customer := range customers {
        err := c.processCustomerUpdate(ctx, customer)
        if err != nil {
            // Log error but continue processing
            log.Printf("Failed to sync customer %s: %v", customer.ID, err)
            continue
        }
    }
    
    // Update sync timestamp
    return c.syncRepository.UpdateSyncTime(ctx, "customer_sync", time.Now())
}

func (c *CustomerSyncService) processCustomerUpdate(ctx context.Context, customer *CRMCustomer) error {
    // Check if customer exists in financial system
    existing, err := c.financialSvc.GetCustomer(ctx, customer.ID)
    if err != nil && !errors.Is(err, ErrCustomerNotFound) {
        return err
    }
    
    financialCustomer := &domain.Customer{
        ID:              customer.ID,
        TenantID:        customer.TenantID,
        CustomerCode:    customer.Code,
        Name:            customer.Name,
        Email:           customer.Email,
        Phone:           customer.Phone,
        BillingAddress:  convertAddress(customer.BillingAddress),
        ShippingAddress: convertAddress(customer.ShippingAddress),
        PaymentTerms:    convertPaymentTerms(customer.PaymentTerms),
        CreditLimit:     customer.CreditLimit,
        Currency:        customer.Currency,
        TaxID:           customer.TaxID,
        IsActive:        customer.IsActive,
    }
    
    if existing == nil {
        // Create new customer
        return c.financialSvc.CreateCustomer(ctx, financialCustomer)
    } else {
        // Update existing customer
        return c.financialSvc.UpdateCustomer(ctx, financialCustomer)
    }
}
```

### **Chart of Accounts Synchronization**
```go
// Multi-tenant chart of accounts management
type ChartOfAccountsSync struct {
    templateService TemplateService
    accountService  AccountService
}

func (c *ChartOfAccountsSync) SyncTenantAccounts(ctx context.Context, tenantID string) error {
    // Get tenant's chart of accounts template
    template, err := c.templateService.GetChartTemplate(ctx, tenantID)
    if err != nil {
        return err
    }
    
    // Get existing accounts
    existingAccounts, err := c.accountService.GetAccounts(ctx, tenantID)
    if err != nil {
        return err
    }
    
    existingMap := make(map[string]*domain.Account)
    for _, account := range existingAccounts {
        existingMap[account.Code] = account
    }
    
    // Sync template accounts
    for _, templateAccount := range template.Accounts {
        if existing, exists := existingMap[templateAccount.Code]; exists {
            // Update if template is newer
            if templateAccount.Version > existing.TemplateVersion {
                err := c.updateAccountFromTemplate(ctx, existing, templateAccount)
                if err != nil {
                    return err
                }
            }
        } else {
            // Create new account from template
            err := c.createAccountFromTemplate(ctx, tenantID, templateAccount)
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

---

## API Integration Patterns

### **Retry and Circuit Breaker Pattern**
```go
// Resilient external API integration
package integration

import (
    "context"
    "time"
    
    "github.com/sony/gobreaker"
)

type ResilientAPIClient struct {
    client         HTTPClient
    circuitBreaker *gobreaker.CircuitBreaker
    retryConfig    RetryConfig
}

type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func NewResilientAPIClient(client HTTPClient) *ResilientAPIClient {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "external-api",
        MaxRequests: 3,
        Interval:    60 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures > 5
        },
    })
    
    return &ResilientAPIClient{
        client:         client,
        circuitBreaker: cb,
        retryConfig: RetryConfig{
            MaxAttempts: 3,
            BaseDelay:   100 * time.Millisecond,
            MaxDelay:    5 * time.Second,
            Multiplier:  2.0,
        },
    }
}

func (r *ResilientAPIClient) CallAPI(ctx context.Context, request APIRequest) (APIResponse, error) {
    operation := func() (interface{}, error) {
        return r.callWithRetry(ctx, request)
    }
    
    result, err := r.circuitBreaker.Execute(operation)
    if err != nil {
        return APIResponse{}, err
    }
    
    return result.(APIResponse), nil
}

func (r *ResilientAPIClient) callWithRetry(ctx context.Context, request APIRequest) (APIResponse, error) {
    var lastErr error
    delay := r.retryConfig.BaseDelay
    
    for attempt := 0; attempt < r.retryConfig.MaxAttempts; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return APIResponse{}, ctx.Err()
            case <-time.After(delay):
                // Continue with retry
            }
            
            // Exponential backoff with jitter
            delay = time.Duration(float64(delay) * r.retryConfig.Multiplier)
            if delay > r.retryConfig.MaxDelay {
                delay = r.retryConfig.MaxDelay
            }
        }
        
        response, err := r.client.Call(ctx, request)
        if err == nil {
            return response, nil
        }
        
        // Check if error is retryable
        if !isRetryableError(err) {
            return APIResponse{}, err
        }
        
        lastErr = err
    }
    
    return APIResponse{}, lastErr
}
```

### **Webhook Management**
```go
// Webhook receiver and processor
type WebhookManager struct {
    processors map[string]WebhookProcessor
    validator  WebhookValidator
    queue      WebhookQueue
}

type WebhookData struct {
    Source    string                 `json:"source"`
    EventType string                 `json:"event_type"`
    Data      map[string]interface{} `json:"data"`
    Signature string                 `json:"signature"`
    Timestamp time.Time              `json:"timestamp"`
}

// Receive webhook from external system
func (w *WebhookManager) ReceiveWebhook(ctx context.Context, webhook *WebhookData) error {
    // Validate webhook signature
    if !w.validator.ValidateSignature(webhook) {
        return errors.New("invalid webhook signature")
    }
    
    // Check timestamp to prevent replay attacks
    if time.Since(webhook.Timestamp) > 5*time.Minute {
        return errors.New("webhook timestamp too old")
    }
    
    // Queue webhook for async processing
    return w.queue.Enqueue(ctx, webhook)
}

// Process webhooks asynchronously
func (w *WebhookManager) ProcessWebhooks(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            webhook, err := w.queue.Dequeue(ctx)
            if err != nil {
                time.Sleep(time.Second)
                continue
            }
            
            err = w.processWebhook(ctx, webhook)
            if err != nil {
                // Handle failed webhook processing
                w.handleWebhookError(ctx, webhook, err)
            }
        }
    }
}

func (w *WebhookManager) processWebhook(ctx context.Context, webhook *WebhookData) error {
    processor, exists := w.processors[webhook.Source]
    if !exists {
        return fmt.Errorf("no processor for webhook source: %s", webhook.Source)
    }
    
    return processor.Process(ctx, webhook)
}
```

---

## Workflow Integration

### **Temporal Workflow Integration**
```go
// Temporal workflow for financial processes
package workflows

import (
    "time"
    
    "go.temporal.io/sdk/workflow"
)

// Invoice approval workflow
func InvoiceApprovalWorkflow(ctx workflow.Context, request InvoiceApprovalRequest) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting invoice approval workflow", "invoice_id", request.InvoiceID)
    
    // Step 1: Validate invoice
    var validationResult ValidationResult
    err := workflow.ExecuteActivity(ctx, 
        workflow.ActivityOptions{
            StartToCloseTimeout: 30 * time.Second,
        }, 
        ValidateInvoiceActivity, request.InvoiceID).Get(ctx, &validationResult)
    
    if err != nil {
        return err
    }
    
    if !validationResult.IsValid {
        return workflow.NewApplicationError("invoice validation failed", "VALIDATION_ERROR")
    }
    
    // Step 2: Route for approval based on amount
    var approvalResult ApprovalResult
    if request.Amount.GreaterThan(decimal.NewFromFloat(10000)) {
        // High value - requires multiple approvals
        err = workflow.ExecuteChildWorkflow(ctx,
            workflow.ChildWorkflowOptions{
                WorkflowID: fmt.Sprintf("multi-approval-%s", request.InvoiceID),
            },
            MultiLevelApprovalWorkflow, request).Get(ctx, &approvalResult)
    } else {
        // Standard approval
        err = workflow.ExecuteActivity(ctx,
            workflow.ActivityOptions{
                StartToCloseTimeout: 5 * time.Minute,
                HeartbeatTimeout:    30 * time.Second,
            },
            SingleApprovalActivity, request).Get(ctx, &approvalResult)
    }
    
    if err != nil {
        return err
    }
    
    // Step 3: Process approval result
    if approvalResult.Approved {
        // Create payment if approved
        var paymentResult PaymentResult
        err = workflow.ExecuteActivity(ctx,
            workflow.ActivityOptions{
                StartToCloseTimeout: 2 * time.Minute,
                RetryPolicy: &temporal.RetryPolicy{
                    MaximumAttempts: 3,
                },
            },
            CreatePaymentActivity, request.InvoiceID).Get(ctx, &paymentResult)
        
        if err != nil {
            return err
        }
        
        logger.Info("Invoice approved and payment created", 
            "invoice_id", request.InvoiceID,
            "payment_id", paymentResult.PaymentID)
    } else {
        // Handle rejection
        err = workflow.ExecuteActivity(ctx,
            workflow.ActivityOptions{
                StartToCloseTimeout: 30 * time.Second,
            },
            RejectInvoiceActivity, 
            RejectInvoiceRequest{
                InvoiceID: request.InvoiceID,
                Reason:    approvalResult.RejectionReason,
            }).Get(ctx, nil)
        
        if err != nil {
            return err
        }
        
        logger.Info("Invoice rejected", 
            "invoice_id", request.InvoiceID,
            "reason", approvalResult.RejectionReason)
    }
    
    return nil
}

// Activities
func ValidateInvoiceActivity(ctx context.Context, invoiceID string) (ValidationResult, error) {
    // Implement invoice validation logic
    return ValidationResult{IsValid: true}, nil
}

func SingleApprovalActivity(ctx context.Context, request InvoiceApprovalRequest) (ApprovalResult, error) {
    // Implement single approval logic
    return ApprovalResult{Approved: true}, nil
}

func CreatePaymentActivity(ctx context.Context, invoiceID string) (PaymentResult, error) {
    // Implement payment creation logic
    return PaymentResult{PaymentID: "pay-123"}, nil
}
```

---

## Real-time Integration

### **WebSocket Integration for Real-time Updates**
```go
// Real-time financial data streaming
package realtime

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/gorilla/websocket"
)

type FinancialDataStreamer struct {
    connections map[string]*websocket.Conn
    eventBus    EventBus
    upgrader    websocket.Upgrader
}

type StreamMessage struct {
    Type      string      `json:"type"`
    TenantID  string      `json:"tenant_id"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
}

// WebSocket endpoint for real-time updates
func (f *FinancialDataStreamer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade connection
    conn, err := f.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // Authenticate and get tenant ID
    tenantID, err := f.authenticateConnection(r)
    if err != nil {
        conn.WriteMessage(websocket.CloseMessage, []byte("Authentication failed"))
        return
    }
    
    // Store connection
    connectionID := fmt.Sprintf("%s-%d", tenantID, time.Now().Unix())
    f.connections[connectionID] = conn
    defer delete(f.connections, connectionID)
    
    // Listen for events
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go f.streamEvents(ctx, tenantID, conn)
    
    // Keep connection alive
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            break
        }
    }
}

func (f *FinancialDataStreamer) streamEvents(ctx context.Context, tenantID string, conn *websocket.Conn) {
    // Subscribe to tenant-specific events
    eventChan := f.eventBus.Subscribe(ctx, tenantID, []string{
        "financial.transaction.posted",
        "financial.invoice.created",
        "financial.payment.received",
        "financial.balance.updated",
    })
    
    for {
        select {
        case <-ctx.Done():
            return
        case event := <-eventChan:
            message := StreamMessage{
                Type:      event.EventType(),
                TenantID:  tenantID,
                Data:      event,
                Timestamp: time.Now(),
            }
            
            data, err := json.Marshal(message)
            if err != nil {
                continue
            }
            
            err = conn.WriteMessage(websocket.TextMessage, data)
            if err != nil {
                return
            }
        }
    }
}

// Broadcast balance updates to connected clients
func (f *FinancialDataStreamer) BroadcastBalanceUpdate(ctx context.Context, tenantID string, balanceUpdate BalanceUpdate) {
    message := StreamMessage{
        Type:      "balance_update",
        TenantID:  tenantID,
        Data:      balanceUpdate,
        Timestamp: time.Now(),
    }
    
    data, _ := json.Marshal(message)
    
    for connectionID, conn := range f.connections {
        if strings.HasPrefix(connectionID, tenantID) {
            conn.WriteMessage(websocket.TextMessage, data)
        }
    }
}
```

---

## Integration Testing

### **Integration Test Framework**
```go
// Integration testing for financial module
package integration_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type FinancialIntegrationTestSuite struct {
    suite.Suite
    testDB          *sql.DB
    testRedis       *redis.Client
    financialSvc    FinancialService
    salesSvc        SalesService
    eventBus        EventBus
    testTenantID    string
}

func (s *FinancialIntegrationTestSuite) SetupSuite() {
    // Setup test database
    s.testDB = setupTestDatabase()
    s.testRedis = setupTestRedis()
    
    // Initialize services
    s.financialSvc = NewFinancialService(s.testDB, s.testRedis)
    s.salesSvc = NewSalesService(s.testDB, s.testRedis)
    s.eventBus = NewEventBus(s.testRedis)
    
    // Create test tenant
    s.testTenantID = "test-tenant-123"
    s.setupTestTenant()
}

func (s *FinancialIntegrationTestSuite) TestSalesOrderToInvoiceFlow() {
    ctx := context.Background()
    
    // Create test customer
    customer := &domain.Customer{
        TenantID:     s.testTenantID,
        CustomerCode: "CUST-001",
        Name:         "Test Customer",
        Currency:     "USD",
    }
    
    createdCustomer, err := s.financialSvc.CreateCustomer(ctx, customer)
    s.Require().NoError(err)
    
    // Create sales order
    salesOrder := &domain.SalesOrder{
        TenantID:   s.testTenantID,
        CustomerID: createdCustomer.ID,
        OrderDate:  time.Now(),
        LineItems: []domain.SalesOrderLineItem{
            {
                ProductID:   "prod-001",
                Quantity:    decimal.NewFromFloat(10),
                UnitPrice:   decimal.NewFromFloat(100),
                LineAmount:  decimal.NewFromFloat(1000),
            },
        },
        TotalAmount: decimal.NewFromFloat(1000),
        Status:      "confirmed",
    }
    
    createdOrder, err := s.salesSvc.CreateSalesOrder(ctx, salesOrder)
    s.Require().NoError(err)
    
    // Fulfill sales order (this should trigger invoice creation)
    err = s.salesSvc.FulfillSalesOrder(ctx, createdOrder.ID)
    s.Require().NoError(err)
    
    // Wait for event processing
    time.Sleep(100 * time.Millisecond)
    
    // Verify invoice was created
    invoices, err := s.financialSvc.GetInvoicesBySalesOrder(ctx, createdOrder.ID)
    s.Require().NoError(err)
    s.Assert().Len(invoices, 1)
    
    invoice := invoices[0]
    s.Assert().Equal(createdCustomer.ID, invoice.CustomerID)
    s.Assert().Equal(createdOrder.TotalAmount, invoice.TotalAmount)
    s.Assert().Equal("draft", invoice.Status)
    
    // Verify accounting entries were created
    transactions, err := s.financialSvc.GetTransactionsByInvoice(ctx, invoice.ID)
    s.Require().NoError(err)
    s.Assert().Len(transactions, 1)
    
    transaction := transactions[0]
    s.Assert().Equal(2, len(transaction.Entries)) // AR debit + Revenue credit
    
    // Verify AR debit entry
    arEntry := findEntryByAccountType(transaction.Entries, "receivable")
    s.Assert().NotNil(arEntry)
    s.Assert().Equal(invoice.TotalAmount, arEntry.DebitAmount)
    
    // Verify Revenue credit entry
    revenueEntry := findEntryByAccountType(transaction.Entries, "revenue")
    s.Assert().NotNil(revenueEntry)
    s.Assert().Equal(invoice.TotalAmount, revenueEntry.CreditAmount)
}

func (s *FinancialIntegrationTestSuite) TestPaymentProcessingFlow() {
    ctx := context.Background()
    
    // Create invoice first
    invoice := s.createTestInvoice(ctx)
    
    // Process payment
    payment := &domain.CustomerPayment{
        TenantID:       s.testTenantID,
        CustomerID:     invoice.CustomerID,
        PaymentDate:    time.Now(),
        Amount:         invoice.TotalAmount,
        Currency:       invoice.Currency,
        PaymentMethod:  "bank_transfer",
        Reference:      "TEST-PAY-001",
        InvoiceIDs:     []string{invoice.ID},
    }
    
    processedPayment, err := s.financialSvc.ProcessCustomerPayment(ctx, payment)
    s.Require().NoError(err)
    
    // Verify payment allocation
    s.Assert().Equal("completed", processedPayment.Status)
    
    // Verify invoice status updated
    updatedInvoice, err := s.financialSvc.GetInvoice(ctx, invoice.ID)
    s.Require().NoError(err)
    s.Assert().Equal("paid", updatedInvoice.Status)
    s.Assert().Equal(decimal.Zero, updatedInvoice.OutstandingAmount)
    
    // Verify accounting entries
    transactions, err := s.financialSvc.GetTransactionsByPayment(ctx, processedPayment.ID)
    s.Require().NoError(err)
    s.Assert().Len(transactions, 1)
    
    transaction := transactions[0]
    s.Assert().Equal(2, len(transaction.Entries)) // Cash debit + AR credit
}

func TestFinancialIntegrationSuite(t *testing.T) {
    suite.Run(t, new(FinancialIntegrationTestSuite))
}
```

---

## Troubleshooting Integration Issues

### **Common Integration Problems**

#### **Event Processing Delays**
```bash
# Check Redis Streams lag
redis-cli XINFO STREAM events:financial.transaction.posted

# Monitor consumer group processing
redis-cli XINFO GROUPS events:financial.transaction.posted

# Check for failed messages
redis-cli XPENDING events:financial.transaction.posted financial-service

# Replay failed messages
redis-cli XCLAIM events:financial.transaction.posted financial-service consumer1 60000 <message-id>
```

#### **Database Deadlocks**
```sql
-- Monitor deadlock information
SELECT 
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Check for long-running transactions
SELECT 
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
AND state = 'active';
```

#### **API Integration Failures**
```go
// Health check for external APIs
func (h *HealthChecker) CheckExternalAPIs(ctx context.Context) map[string]HealthStatus {
    results := make(map[string]HealthStatus)
    
    // Check payment gateway
    results["payment_gateway"] = h.checkPaymentGateway(ctx)
    
    // Check tax service
    results["tax_service"] = h.checkTaxService(ctx)
    
    // Check bank APIs
    results["bank_api"] = h.checkBankAPI(ctx)
    
    return results
}

func (h *HealthChecker) checkPaymentGateway(ctx context.Context) HealthStatus {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    _, err := h.paymentGateway.HealthCheck(ctx)
    if err != nil {
        return HealthStatus{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    }
    
    return HealthStatus{
        Status:  "healthy",
        Message: "Payment gateway responding normally",
    }
}
```

---

**Document Control**
- **Version**: 1.0
- **Last Updated**: January 2025
- **Next Review**: Monthly during development
- **Approval Required**: Integration Architect, Technical Lead

**Related Documents**
- Financial Implementation Plan (@docs/module/financial/financial-implementation-plan.md)
- API Reference Guide (@docs/module/financial/financial-api-reference.md)
- Deployment Guide (@docs/module/financial/financial-deployment-guide.md)