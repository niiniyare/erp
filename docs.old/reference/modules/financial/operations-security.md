# Financial Module - Operations & Security Guide

> **Comprehensive guide covering security architecture, compliance requirements, deployment strategies, testing frameworks, and operational procedures for the AWO ERP Financial Module.**

## Table of Contents

1. [Security Architecture](#security-architecture)
2. [Compliance Framework](#compliance-framework)
3. [Deployment & Infrastructure](#deployment--infrastructure)
4. [Testing Strategy](#testing-strategy)
5. [Monitoring & Observability](#monitoring--observability)
6. [Backup & Recovery](#backup--recovery)
7. [Performance Optimization](#performance-optimization)
8. [Maintenance Procedures](#maintenance-procedures)

---

## Security Architecture

### Defense-in-Depth Strategy

The AWO ERP Financial Module implements military-grade security through multiple layers:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Application Security Layer                   │
│  • Input validation and sanitization                           │
│  • Business logic authorization                                │
│  • Rate limiting and DDoS protection                          │
│  • Session management and timeout                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  ABAC Authorization Layer                       │
│  • Attribute-based access control                             │
│  • Context-aware decision making                              │
│  • Policy evaluation engine                                   │
│  • Real-time risk assessment                                  │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Data Access Security Layer                   │
│  • Row-level security (RLS)                                   │
│  • Column-level encryption                                    │
│  • Database audit logging                                     │
│  • Tenant isolation enforcement                               │
└─────────────────────────────────────────────────────────────────┘
```

### Multi-Tenant Security

**Row Level Security (RLS) Implementation:**
```sql
-- Tenant isolation function
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID AS $$
BEGIN
    RETURN COALESCE(
        current_setting('app.current_tenant_id', true)::UUID,
        NULL
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Automatic tenant isolation
CREATE POLICY tenant_isolation ON finance_accounts
FOR ALL TO application_role USING (
    current_tenant_id() IS NOT NULL 
    AND tenant_id = current_tenant_id()
    AND deleted_at IS NULL
);

-- Admin bypass for system operations
CREATE POLICY admin_full_access ON finance_accounts
FOR ALL TO admin_role USING (true);
```

**Application-Level Security:**
```go
// Middleware for tenant context injection
func TenantMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-ID")
        if tenantID == "" {
            return fiber.NewError(fiber.StatusUnauthorized, "Tenant ID required")
        }
        
        // Validate tenant access
        if !hasValidTenantAccess(c.Context(), tenantID, c.Get("Authorization")) {
            return fiber.NewError(fiber.StatusForbidden, "Invalid tenant access")
        }
        
        // Inject tenant context into database session
        if err := setTenantContext(c.Context(), tenantID); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "Failed to set tenant context")
        }
        
        return c.Next()
    }
}

// ABAC policy enforcement
type ABACPolicy struct {
    Subject   map[string]interface{} `json:"subject"`
    Resource  map[string]interface{} `json:"resource"`
    Action    string                 `json:"action"`
    Context   map[string]interface{} `json:"context"`
}

func (p *ABACPolicy) Evaluate() bool {
    // Policy evaluation logic
    if p.Action == "transactions:create" {
        maxAmount := p.Subject["max_transaction_amount"].(float64)
        requestedAmount := p.Resource["amount"].(float64)
        
        if requestedAmount > maxAmount {
            timeOfDay := p.Context["time_of_day"].(string)
            return timeOfDay == "business_hours" && 
                   p.Subject["override_permissions"].(bool)
        }
        return true
    }
    
    return false
}
```

### Data Protection & Encryption

**Encryption at Rest:**
```sql
-- Transparent Data Encryption (TDE) for sensitive fields
CREATE TABLE sensitive_financial_data (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    account_number BYTEA,  -- Encrypted field
    routing_number BYTEA,  -- Encrypted field
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Application-layer encryption functions
CREATE OR REPLACE FUNCTION encrypt_sensitive(plain_text TEXT)
RETURNS BYTEA AS $$
BEGIN
    RETURN pgp_sym_encrypt(plain_text, current_setting('app.encryption_key'));
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE OR REPLACE FUNCTION decrypt_sensitive(encrypted_data BYTEA)
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted_data, current_setting('app.encryption_key'));
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

**Encryption in Transit:**
```nginx
# TLS 1.3 configuration
server {
    listen 443 ssl http2;
    ssl_protocols TLSv1.3;
    ssl_ciphers ECDHE+AESGCM:ECDHE+CHACHA20:DHE+AESGCM:DHE+CHACHA20:!aNULL:!MD5:!DSS;
    ssl_prefer_server_ciphers off;
    
    # HSTS header
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
    
    # Certificate pinning
    add_header Public-Key-Pins 'pin-sha256="base64+primary=="; pin-sha256="base64+backup=="; max-age=5184000; includeSubDomains';
}
```

### Audit & Logging

**Comprehensive Audit Trail:**
```sql
-- Audit log table structure
CREATE TABLE finance_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    table_name VARCHAR(50) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values JSONB,
    new_values JSONB,
    changed_by UUID NOT NULL REFERENCES users(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    session_id UUID,
    correlation_id UUID
);

-- Audit trigger function
CREATE OR REPLACE FUNCTION audit_financial_changes()
RETURNS TRIGGER AS $$
DECLARE
    correlation_id UUID := current_setting('app.correlation_id', true)::UUID;
BEGIN
    INSERT INTO finance_audit_log (
        tenant_id, table_name, record_id, action,
        old_values, new_values, changed_by, ip_address, 
        user_agent, correlation_id
    ) VALUES (
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        CASE WHEN TG_OP = 'DELETE' THEN row_to_json(OLD) ELSE NULL END,
        CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN row_to_json(NEW) ELSE NULL END,
        current_setting('app.current_user_id')::UUID,
        current_setting('app.client_ip')::INET,
        current_setting('app.user_agent'),
        correlation_id
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

**Security Event Monitoring:**
```go
type SecurityEvent struct {
    EventType   string                 `json:"event_type"`
    Severity    string                 `json:"severity"`
    UserID      uuid.UUID              `json:"user_id"`
    TenantID    uuid.UUID              `json:"tenant_id"`
    IPAddress   string                 `json:"ip_address"`
    Details     map[string]interface{} `json:"details"`
    Timestamp   time.Time              `json:"timestamp"`
}

// Security event logger
func LogSecurityEvent(ctx context.Context, event SecurityEvent) {
    // Log to security SIEM
    siem.LogEvent(event)
    
    // Real-time alerting for critical events
    if event.Severity == "CRITICAL" {
        alertManager.TriggerAlert("security_breach", event)
    }
    
    // Compliance audit log
    complianceLogger.Log(event)
}

// Example usage
LogSecurityEvent(ctx, SecurityEvent{
    EventType: "UNAUTHORIZED_ACCESS_ATTEMPT",
    Severity:  "HIGH",
    UserID:    userID,
    TenantID:  tenantID,
    IPAddress: clientIP,
    Details: map[string]interface{}{
        "attempted_resource": "/api/v1/finance/accounts",
        "reason": "insufficient_permissions",
        "action_taken": "access_denied",
    },
})
```

---

## Compliance Framework

### Regulatory Compliance

**SOX (Sarbanes-Oxley) Compliance:**
```yaml
SOX_Requirements:
  Section_302:
    - CEO/CFO certification of financial reports
    - Internal control effectiveness attestation
    - Quarterly compliance certifications
  
  Section_404:
    - Management assessment of internal controls
    - External auditor attestation
    - Control deficiency remediation
  
  Section_409:
    - Real-time disclosure of material changes
    - Rapid reporting of financial events
    - Insider trading prevention
```

**GAAP (Generally Accepted Accounting Principles):**
```sql
-- Double-entry bookkeeping enforcement
ALTER TABLE finance_transactions ADD CONSTRAINT balanced_transaction
CHECK (
    CASE WHEN transaction_status IN ('POSTED', 'APPROVED')
    THEN ABS(total_debit_amount - total_credit_amount) < 0.01
    ELSE TRUE END
);

-- Revenue recognition compliance
CREATE TABLE revenue_recognition_schedule (
    id UUID PRIMARY KEY,
    transaction_id UUID REFERENCES finance_transactions(id),
    recognition_date DATE NOT NULL,
    recognition_amount DECIMAL(15,2) NOT NULL,
    recognition_percentage DECIMAL(5,2),
    performance_obligation_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Data Privacy & GDPR

**Data Classification:**
```go
type DataClassification string

const (
    PublicData       DataClassification = "PUBLIC"
    InternalData     DataClassification = "INTERNAL"
    ConfidentialData DataClassification = "CONFIDENTIAL"
    RestrictedData   DataClassification = "RESTRICTED"
    PersonalData     DataClassification = "PERSONAL" // GDPR protected
)

type FinancialData struct {
    ID           uuid.UUID          `json:"id"`
    Classification DataClassification `json:"classification"`
    RetentionPolicy string           `json:"retention_policy"`
    EncryptionRequired bool         `json:"encryption_required"`
    AccessLog    []AccessLogEntry    `json:"access_log"`
}

// GDPR compliance functions
func (fd *FinancialData) IsPersonalData() bool {
    return fd.Classification == PersonalData
}

func (fd *FinancialData) CanBeDeleted() bool {
    return time.Since(fd.CreatedAt) > fd.GetRetentionPeriod()
}

func (fd *FinancialData) AnonymizeData() error {
    if !fd.IsPersonalData() {
        return nil
    }
    
    // Implement data anonymization
    return fd.applyAnonymization()
}
```

**Right to be Forgotten Implementation:**
```sql
-- GDPR data deletion procedure
CREATE OR REPLACE FUNCTION gdpr_delete_personal_data(
    p_data_subject_id UUID,
    p_retention_override BOOLEAN DEFAULT FALSE
)
RETURNS JSON AS $$
DECLARE
    deletion_report JSON;
BEGIN
    -- Check retention policies
    IF NOT p_retention_override THEN
        -- Verify legal basis for deletion
        PERFORM check_retention_requirements(p_data_subject_id);
    END IF;
    
    -- Anonymize financial records
    UPDATE finance_accounts 
    SET account_name = 'ANONYMIZED_' || id::text,
        created_by = NULL,
        updated_by = NULL
    WHERE created_by = p_data_subject_id 
       OR updated_by = p_data_subject_id;
    
    -- Create deletion audit record
    INSERT INTO gdpr_deletion_log (
        data_subject_id, deletion_date, deletion_reason,
        records_affected, performed_by
    ) VALUES (
        p_data_subject_id, NOW(), 'RIGHT_TO_BE_FORGOTTEN',
        json_build_object('accounts', ROW_COUNT),
        current_setting('app.current_user_id')::UUID
    );
    
    RETURN json_build_object(
        'status', 'completed',
        'deletion_date', NOW(),
        'records_affected', ROW_COUNT
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

### Audit Requirements

**Internal Control Testing:**
```go
type ControlTest struct {
    ControlID       string    `json:"control_id"`
    TestType        string    `json:"test_type"` // WALKTHROUGH, DESIGN, OPERATING
    TestProcedure   string    `json:"test_procedure"`
    SampleSize      int       `json:"sample_size"`
    TestResults     string    `json:"test_results"`
    Deficiencies    []string  `json:"deficiencies"`
    TestDate        time.Time `json:"test_date"`
    TestedBy        uuid.UUID `json:"tested_by"`
}

// Automated control testing
func RunControlTests(ctx context.Context, period string) (*ControlTestReport, error) {
    tests := []ControlTest{
        {
            ControlID: "FIN-001",
            TestType: "OPERATING",
            TestProcedure: "Verify transaction approval workflow",
            SampleSize: 25,
        },
        {
            ControlID: "FIN-002", 
            TestType: "OPERATING",
            TestProcedure: "Test segregation of duties in transaction posting",
            SampleSize: 50,
        },
    }
    
    var results []ControlTestResult
    for _, test := range tests {
        result := executeControlTest(ctx, test)
        results = append(results, result)
    }
    
    return generateControlReport(results), nil
}
```

---

## Deployment & Infrastructure

### Multi-Environment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Production                              │
│  • High availability (3 AZs)                                   │
│  • Auto-scaling (2-20 instances)                              │
│  • Read replicas (3x)                                         │
│  • 99.99% uptime SLA                                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Staging                                │
│  • Production mirror                                           │
│  • Performance testing                                         │
│  • User acceptance testing                                     │
│  • Blue-green deployment                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Development                              │
│  • Feature development                                         │
│  • Integration testing                                         │
│  • Code review validation                                      │
│  • Automated testing                                          │
└─────────────────────────────────────────────────────────────────┘
```

### Infrastructure Requirements

**Database Infrastructure:**
```yaml
PostgreSQL_Configuration:
  Version: "15.x"
  Memory: "32GB minimum (64GB recommended)"
  Storage: "SSD NVMe (minimum 10K IOPS)"
  CPU: "8 cores minimum (16 cores recommended)"
  
  High_Availability:
    Primary: "Multi-AZ deployment"
    Replicas: "3 read replicas"
    Backup: "Point-in-time recovery (PITR)"
    Failover: "Automatic failover < 60 seconds"
  
  Performance_Tuning:
    shared_buffers: "8GB"
    work_mem: "256MB" 
    maintenance_work_mem: "2GB"
    effective_cache_size: "24GB"
    max_connections: "200"
    
Redis_Configuration:
  Version: "7.x"
  Memory: "8GB minimum"
  Persistence: "AOF + RDB snapshots"
  Clustering: "Redis Cluster (3 masters, 3 replicas)"
```

**Application Infrastructure:**
```yaml
Kubernetes_Deployment:
  Namespace: "finance-module"
  
  API_Service:
    Replicas: 3
    Resources:
      CPU: "500m-2000m"
      Memory: "1Gi-4Gi"
    Health_Checks:
      Liveness: "/health/live"
      Readiness: "/health/ready"
      
  Background_Workers:
    Replicas: 2
    Resources:
      CPU: "200m-1000m" 
      Memory: "512Mi-2Gi"
      
  Monitoring:
    Prometheus: "Metrics collection"
    Grafana: "Dashboard visualization"
    Jaeger: "Distributed tracing"
    ELK_Stack: "Log aggregation"
```

### Database Setup & Migration

**Migration Management:**
```sql
-- Migration version tracking
CREATE TABLE schema_migrations (
    version VARCHAR(14) PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW(),
    applied_by TEXT DEFAULT current_user
);

-- Migration procedure
CREATE OR REPLACE FUNCTION apply_migration(
    migration_version VARCHAR(14),
    migration_sql TEXT
)
RETURNS BOOLEAN AS $$
DECLARE
    migration_exists BOOLEAN;
BEGIN
    -- Check if migration already applied
    SELECT EXISTS(
        SELECT 1 FROM schema_migrations 
        WHERE version = migration_version
    ) INTO migration_exists;
    
    IF migration_exists THEN
        RAISE NOTICE 'Migration % already applied', migration_version;
        RETURN FALSE;
    END IF;
    
    -- Execute migration in transaction
    BEGIN
        EXECUTE migration_sql;
        
        INSERT INTO schema_migrations (version) 
        VALUES (migration_version);
        
        RAISE NOTICE 'Migration % applied successfully', migration_version;
        RETURN TRUE;
        
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'Migration % failed: %', migration_version, SQLERRM;
    END;
END;
$$ LANGUAGE plpgsql;
```

**Zero-Downtime Deployment:**
```bash
#!/bin/bash
# Blue-Green deployment script

ENVIRONMENT=$1
NEW_VERSION=$2

echo "Starting blue-green deployment for $ENVIRONMENT"

# 1. Deploy new version to green environment
kubectl apply -f k8s/green-deployment.yaml
kubectl set image deployment/finance-api-green finance-api=finance:$NEW_VERSION

# 2. Wait for green deployment to be ready
kubectl rollout status deployment/finance-api-green --timeout=600s

# 3. Run health checks
./scripts/health-check.sh green-service

if [ $? -eq 0 ]; then
    echo "Health checks passed. Switching traffic to green."
    
    # 4. Switch traffic to green
    kubectl patch service finance-api-service -p '{"spec":{"selector":{"version":"green"}}}'
    
    # 5. Wait and verify
    sleep 30
    ./scripts/health-check.sh finance-api-service
    
    if [ $? -eq 0 ]; then
        echo "Deployment successful. Cleaning up blue environment."
        kubectl delete deployment finance-api-blue
    else
        echo "Health check failed. Rolling back to blue."
        kubectl patch service finance-api-service -p '{"spec":{"selector":{"version":"blue"}}}'
        exit 1
    fi
else
    echo "Health check failed. Deployment aborted."
    kubectl delete deployment finance-api-green
    exit 1
fi
```

---

## Testing Strategy

### Test Pyramid Implementation

```
                    ┌─────────────────────┐
                    │   E2E Tests (5%)    │ ← Manual & Automated
                    └─────────────────────┘
                ┌─────────────────────────────┐
                │  Integration Tests (15%)    │ ← API & Database
                └─────────────────────────────┘
        ┌─────────────────────────────────────────┐
        │      Unit Tests (80%)                   │ ← Business Logic
        └─────────────────────────────────────────┘
```

### Unit Testing Framework

**Domain Logic Testing:**
```go
// Test double-entry bookkeeping validation
func TestTransactionBalanceValidation(t *testing.T) {
    tests := []struct {
        name        string
        transaction *domain.Transaction
        expectError bool
    }{
        {
            name: "valid_balanced_transaction",
            transaction: &domain.Transaction{
                Entries: []domain.TransactionEntry{
                    {AccountID: cashAccountID, DebitAmount: decimal.NewFromFloat(1000)},
                    {AccountID: revenueAccountID, CreditAmount: decimal.NewFromFloat(1000)},
                },
            },
            expectError: false,
        },
        {
            name: "unbalanced_transaction_should_fail",
            transaction: &domain.Transaction{
                Entries: []domain.TransactionEntry{
                    {AccountID: cashAccountID, DebitAmount: decimal.NewFromFloat(1000)},
                    {AccountID: revenueAccountID, CreditAmount: decimal.NewFromFloat(500)},
                },
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.transaction.Validate()
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "debits must equal credits")
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Test account hierarchy validation
func TestAccountHierarchyValidation(t *testing.T) {
    account := &domain.Account{
        ID:              uuid.New(),
        ParentAccountID: &uuid.UUID{}, // Same as ID - circular reference
    }
    account.ParentAccountID = &account.ID
    
    err := account.ValidateHierarchy()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular reference")
}
```

**Repository Testing with Test Containers:**
```go
func TestAccountRepository(t *testing.T) {
    // Start test database container
    ctx := context.Background()
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "testpass",
                "POSTGRES_DB":       "testdb",
            },
            WaitingFor: wait.ForLog("database system is ready to accept connections"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)
    
    // Get connection details and run tests
    host, _ := postgres.Host(ctx)
    port, _ := postgres.MappedPort(ctx, "5432")
    
    db := setupTestDB(host, port.Port())
    repo := repository.NewAccountRepository(db)
    
    t.Run("create_account", func(t *testing.T) {
        account := &domain.Account{
            AccountCode: "1001",
            AccountName: "Test Cash Account",
            RootType:    domain.RootTypeAsset,
        }
        
        created, err := repo.Create(ctx, account)
        assert.NoError(t, err)
        assert.Equal(t, account.AccountCode, created.AccountCode)
        assert.NotZero(t, created.ID)
    })
}
```

### Integration Testing

**API Integration Tests:**
```go
func TestTransactionAPI(t *testing.T) {
    // Setup test environment
    app := setupTestApp()
    
    t.Run("create_transaction_workflow", func(t *testing.T) {
        // 1. Create accounts
        cashAccount := createTestAccount(t, app, "1001", "Cash")
        revenueAccount := createTestAccount(t, app, "4001", "Revenue")
        
        // 2. Create transaction
        transactionReq := map[string]interface{}{
            "transaction_number": "TEST-001",
            "description":        "Test transaction",
            "entries": []map[string]interface{}{
                {
                    "account_id":    cashAccount.ID,
                    "debit_amount":  "1000.00",
                    "description":   "Cash receipt",
                },
                {
                    "account_id":     revenueAccount.ID,
                    "credit_amount":  "1000.00", 
                    "description":    "Sales revenue",
                },
            },
        }
        
        resp := performRequest(app, "POST", "/api/v1/finance/transactions", transactionReq)
        assert.Equal(t, 201, resp.Code)
        
        var transaction map[string]interface{}
        json.Unmarshal(resp.Body.Bytes(), &transaction)
        
        // 3. Submit for approval
        submitResp := performRequest(app, "PUT", 
            fmt.Sprintf("/api/v1/finance/transactions/%s/submit", transaction["id"]), nil)
        assert.Equal(t, 200, submitResp.Code)
        
        // 4. Approve transaction
        approveResp := performRequest(app, "PUT",
            fmt.Sprintf("/api/v1/finance/transactions/%s/approve", transaction["id"]),
            map[string]string{"approval_notes": "Test approval"})
        assert.Equal(t, 200, approveResp.Code)
        
        // 5. Post transaction
        postResp := performRequest(app, "PUT",
            fmt.Sprintf("/api/v1/finance/transactions/%s/post", transaction["id"]), nil)
        assert.Equal(t, 200, postResp.Code)
        
        // 6. Verify account balances updated
        verifyAccountBalance(t, app, cashAccount.ID, "1000.00")
        verifyAccountBalance(t, app, revenueAccount.ID, "-1000.00") // Credit balance
    })
}
```

### Security & Compliance Testing

**Security Test Suite:**
```go
func TestSecurityControls(t *testing.T) {
    app := setupTestApp()
    
    t.Run("unauthorized_access_prevention", func(t *testing.T) {
        // Test without authentication
        resp := performRequest(app, "GET", "/api/v1/finance/accounts", nil)
        assert.Equal(t, 401, resp.Code)
        
        // Test with invalid token
        resp = performRequestWithAuth(app, "GET", "/api/v1/finance/accounts", nil, "invalid-token")
        assert.Equal(t, 401, resp.Code)
    })
    
    t.Run("tenant_isolation", func(t *testing.T) {
        // Create account in tenant A
        accountA := createTestAccountForTenant(t, app, "tenantA", "1001", "Cash A")
        
        // Try to access with tenant B token
        resp := performRequestWithTenant(app, "GET", 
            fmt.Sprintf("/api/v1/finance/accounts/%s", accountA.ID), nil, "tenantB")
        assert.Equal(t, 404, resp.Code) // Should not find due to RLS
    })
    
    t.Run("sql_injection_prevention", func(t *testing.T) {
        // Attempt SQL injection in search parameter
        maliciousPayload := "'; DROP TABLE finance_accounts; --"
        resp := performRequest(app, "GET", 
            fmt.Sprintf("/api/v1/finance/accounts?search=%s", url.QueryEscape(maliciousPayload)), nil)
        
        // Should return normal response, not error
        assert.Equal(t, 200, resp.Code)
        
        // Verify table still exists
        verifyTableExists(t, app, "finance_accounts")
    })
}
```

**Performance Testing:**
```go
func TestPerformanceRequirements(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance tests in short mode")
    }
    
    app := setupTestApp()
    
    t.Run("transaction_creation_performance", func(t *testing.T) {
        const numTransactions = 1000
        const maxLatency = 100 * time.Millisecond
        
        var latencies []time.Duration
        
        for i := 0; i < numTransactions; i++ {
            start := time.Now()
            
            createTestTransaction(t, app, fmt.Sprintf("PERF-%d", i))
            
            latency := time.Since(start)
            latencies = append(latencies, latency)
        }
        
        // Calculate statistics
        avgLatency := calculateAverage(latencies)
        p95Latency := calculatePercentile(latencies, 95)
        
        t.Logf("Average latency: %v", avgLatency)
        t.Logf("95th percentile latency: %v", p95Latency)
        
        assert.Less(t, avgLatency, maxLatency, "Average latency exceeds requirement")
        assert.Less(t, p95Latency, maxLatency*2, "95th percentile latency exceeds requirement")
    })
}
```

### Test Data Management

**Test Data Factory:**
```go
type TestDataFactory struct {
    db       *sql.DB
    tenantID uuid.UUID
}

func (f *TestDataFactory) CreateAccount(opts ...AccountOption) *domain.Account {
    account := &domain.Account{
        TenantID:    f.tenantID,
        AccountCode: fmt.Sprintf("TEST-%d", rand.Int()),
        AccountName: "Test Account",
        RootType:    domain.RootTypeAsset,
        IsActive:    true,
    }
    
    // Apply options
    for _, opt := range opts {
        opt(account)
    }
    
    // Save to database
    err := f.saveAccount(account)
    if err != nil {
        panic(err)
    }
    
    return account
}

type AccountOption func(*domain.Account)

func WithAccountCode(code string) AccountOption {
    return func(a *domain.Account) {
        a.AccountCode = code
    }
}

func WithRootType(rootType domain.RootType) AccountOption {
    return func(a *domain.Account) {
        a.RootType = rootType
    }
}

// Usage
account := factory.CreateAccount(
    WithAccountCode("1001"),
    WithRootType(domain.RootTypeAsset),
)
```

---

## Monitoring & Observability

### Metrics Collection

**Application Metrics:**
```go
// Prometheus metrics
var (
    transactionsCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "finance_transactions_created_total",
            Help: "Total number of transactions created",
        },
        []string{"tenant_id", "transaction_type", "status"},
    )
    
    transactionProcessingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "finance_transaction_processing_duration_seconds", 
            Help:    "Transaction processing duration",
            Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
        },
        []string{"operation", "tenant_id"},
    )
    
    accountBalanceGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "finance_account_balance",
            Help: "Current account balance",
        },
        []string{"tenant_id", "account_code", "account_type"},
    )
)

// Metrics middleware
func MetricsMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        
        err := c.Next()
        
        duration := time.Since(start).Seconds()
        method := c.Method()
        path := c.Path()
        status := c.Response().StatusCode()
        tenantID := c.Get("X-Tenant-ID")
        
        httpRequestDuration.WithLabelValues(method, path, strconv.Itoa(status), tenantID).Observe(duration)
        
        return err
    }
}
```

**Business Metrics Dashboard:**
```yaml
# Grafana dashboard configuration
Finance_Module_Dashboard:
  Panels:
    - Title: "Transaction Volume"
      Query: "rate(finance_transactions_created_total[5m])"
      Type: "Graph"
      
    - Title: "Account Balance Distribution"  
      Query: "finance_account_balance"
      Type: "Stat"
      
    - Title: "Processing Latency"
      Query: "histogram_quantile(0.95, rate(finance_transaction_processing_duration_seconds_bucket[5m]))"
      Type: "Graph"
      
    - Title: "Error Rate"
      Query: "rate(finance_errors_total[5m])"
      Type: "Singlestat"
      
    - Title: "Active Users by Tenant"
      Query: "count by (tenant_id) (finance_active_sessions)"
      Type: "Table"
```

### Distributed Tracing

**OpenTelemetry Implementation:**
```go
// Tracer setup
var tracer trace.Tracer

func init() {
    tp := trace.NewTracerProvider(
        trace.WithBatcher(jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://jaeger:14268/api/traces")))),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("finance-module"),
            semconv.ServiceVersionKey.String("4.0.0"),
        )),
    )
    
    otel.SetTracerProvider(tp)
    tracer = tp.Tracer("finance-module")
}

// Service tracing
func (s *TransactionService) Create(ctx context.Context, req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
    ctx, span := tracer.Start(ctx, "transaction.create")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("tenant_id", req.TenantID.String()),
        attribute.String("transaction_type", string(req.TransactionType)),
        attribute.Int("entry_count", len(req.Entries)),
    )
    
    // Validate request
    ctx, validateSpan := tracer.Start(ctx, "transaction.validate")
    if err := s.validateRequest(ctx, req); err != nil {
        validateSpan.RecordError(err)
        validateSpan.SetStatus(codes.Error, err.Error())
        validateSpan.End()
        return nil, err
    }
    validateSpan.End()
    
    // Create transaction
    ctx, createSpan := tracer.Start(ctx, "transaction.create.database")
    transaction, err := s.repo.Create(ctx, req)
    createSpan.End()
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    span.SetAttributes(
        attribute.String("transaction_id", transaction.ID.String()),
        attribute.String("transaction_number", transaction.TransactionNumber),
    )
    
    return transaction, nil
}
```

### Log Aggregation

**Structured Logging:**
```go
// Logger configuration
logger := logrus.New()
logger.SetFormatter(&logrus.JSONFormatter{})
logger.SetLevel(logrus.InfoLevel)

type FinanceLogger struct {
    logger *logrus.Logger
}

func (l *FinanceLogger) LogTransactionEvent(ctx context.Context, event string, transaction *domain.Transaction) {
    entry := l.logger.WithContext(ctx).WithFields(logrus.Fields{
        "event":              event,
        "transaction_id":     transaction.ID,
        "transaction_number": transaction.TransactionNumber,
        "tenant_id":          transaction.TenantID,
        "amount":             transaction.TotalDebitAmount,
        "status":             transaction.TransactionStatus,
    })
    
    // Add trace context
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        entry = entry.WithFields(logrus.Fields{
            "trace_id": span.SpanContext().TraceID().String(),
            "span_id":  span.SpanContext().SpanID().String(),
        })
    }
    
    entry.Info("Transaction event processed")
}
```

**ELK Stack Configuration:**
```yaml
# Logstash configuration
input:
  beats:
    port: 5044
    
filter:
  if [fields][service] == "finance-module" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
    
    if [level] == "ERROR" {
      mutate {
        add_tag => [ "alert" ]
      }
    }
  }

output:
  elasticsearch:
    hosts => ["elasticsearch:9200"]
    index => "finance-logs-%{+YYYY.MM.dd}"
```

---

## Backup & Recovery

### Backup Strategy

**Automated Backup Schedule:**
```yaml
Backup_Configuration:
  Full_Backup:
    Frequency: "Daily at 2:00 AM UTC"
    Retention: "30 days"
    Storage: "AWS S3 with encryption"
    
  Incremental_Backup:
    Frequency: "Every 4 hours"
    Retention: "7 days"
    Storage: "Local SSD + S3 sync"
    
  Transaction_Log_Backup:
    Frequency: "Every 15 minutes"
    Retention: "7 days" 
    Storage: "High-speed SSD"
    
  Point_In_Time_Recovery:
    Enabled: true
    Retention: "7 days"
    Granularity: "15 minutes"
```

**Backup Verification:**
```bash
#!/bin/bash
# Automated backup verification script

BACKUP_DATE=$(date +%Y%m%d)
BACKUP_FILE="finance-backup-${BACKUP_DATE}.sql.gz"

echo "Verifying backup: $BACKUP_FILE"

# 1. Check file integrity
if ! gzip -t "$BACKUP_FILE"; then
    echo "ERROR: Backup file corruption detected"
    exit 1
fi

# 2. Test restore in isolated environment
docker run --rm -d --name backup-test postgres:15
sleep 10

# Restore backup to test database
gunzip -c "$BACKUP_FILE" | docker exec -i backup-test psql -U postgres

# 3. Verify critical tables
CRITICAL_TABLES=("finance_accounts" "finance_transactions" "finance_transaction_entries")

for table in "${CRITICAL_TABLES[@]}"; do
    count=$(docker exec backup-test psql -U postgres -t -c "SELECT COUNT(*) FROM $table;")
    if [ "$count" -lt 1 ]; then
        echo "ERROR: Table $table appears empty in backup"
        exit 1
    fi
    echo "✓ Table $table: $count records"
done

# 4. Cleanup
docker stop backup-test

echo "Backup verification completed successfully"
```

### Disaster Recovery

**RTO/RPO Targets:**
```yaml
Recovery_Objectives:
  RTO: "< 4 hours"    # Recovery Time Objective
  RPO: "< 15 minutes" # Recovery Point Objective
  
DR_Procedures:
  Primary_Failure:
    - Automatic failover to standby database
    - DNS cutover to backup region
    - Application restart with new DB endpoint
    
  Region_Failure:
    - Manual failover to disaster recovery region
    - Restore from latest backup
    - Data validation and integrity checks
    
  Data_Corruption:
    - Point-in-time recovery to last known good state
    - Transaction log replay
    - Business validation of restored data
```

**DR Testing Schedule:**
```yaml
DR_Testing:
  Automated_Tests:
    Frequency: "Weekly"
    Scope: "Backup restore simulation"
    Duration: "30 minutes"
    
  Tabletop_Exercises:
    Frequency: "Monthly"
    Participants: ["DevOps", "Finance", "Security"]
    Scenarios: ["Database failure", "Region outage", "Security incident"]
    
  Full_DR_Test:
    Frequency: "Quarterly"
    Scope: "Complete system failover"
    Duration: "4 hours"
    Validation: "End-to-end business process testing"
```

---

## Performance Optimization

### Database Performance

**Query Optimization:**
```sql
-- Slow query identification
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    rows
FROM pg_stat_statements 
WHERE query LIKE '%finance_%'
ORDER BY total_time DESC
LIMIT 10;

-- Index usage analysis
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
  AND tablename LIKE 'finance_%'
ORDER BY idx_scan DESC;

-- Table statistics
SELECT 
    schemaname,
    tablename,
    n_tup_ins,
    n_tup_upd,
    n_tup_del,
    n_dead_tup,
    last_autovacuum
FROM pg_stat_user_tables
WHERE tablename LIKE 'finance_%';
```

**Performance Monitoring:**
```go
// Database connection pool monitoring
func MonitorConnectionPool(db *sql.DB) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := db.Stats()
        
        // Export metrics to Prometheus
        dbConnectionsInUse.Set(float64(stats.InUse))
        dbConnectionsIdle.Set(float64(stats.Idle))
        dbConnectionsWaitCount.Set(float64(stats.WaitCount))
        dbConnectionsWaitDuration.Set(float64(stats.WaitDuration.Seconds()))
        
        // Alert if connection pool is stressed
        if float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.8 {
            alertManager.TriggerAlert("high_db_connection_usage", map[string]interface{}{
                "in_use":         stats.InUse,
                "max_open":       stats.MaxOpenConnections,
                "usage_percent":  (float64(stats.InUse) / float64(stats.MaxOpenConnections)) * 100,
            })
        }
    }
}
```

### Application Performance

**Caching Strategy:**
```go
type CacheStrategy struct {
    redis  *redis.Client
    local  *cache.Cache
    ttl    time.Duration
}

func (c *CacheStrategy) GetAccountBalance(ctx context.Context, accountID uuid.UUID, asOfDate time.Time) (*decimal.Decimal, error) {
    cacheKey := fmt.Sprintf("account_balance:%s:%s", accountID.String(), asOfDate.Format("2006-01-02"))
    
    // L1 Cache: Check local cache first
    if value, found := c.local.Get(cacheKey); found {
        balance := value.(*decimal.Decimal)
        return balance, nil
    }
    
    // L2 Cache: Check Redis cache
    cached, err := c.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var balance decimal.Decimal
        if err := json.Unmarshal([]byte(cached), &balance); err == nil {
            // Store in L1 cache for faster subsequent access
            c.local.Set(cacheKey, &balance, c.ttl)
            return &balance, nil
        }
    }
    
    // Cache miss: Query database
    balance, err := c.queryAccountBalance(ctx, accountID, asOfDate)
    if err != nil {
        return nil, err
    }
    
    // Store in both cache levels
    balanceJSON, _ := json.Marshal(balance)
    c.redis.Set(ctx, cacheKey, balanceJSON, c.ttl)
    c.local.Set(cacheKey, balance, c.ttl)
    
    return balance, nil
}

// Cache invalidation on account updates
func (c *CacheStrategy) InvalidateAccountCache(accountID uuid.UUID) {
    pattern := fmt.Sprintf("account_balance:%s:*", accountID.String())
    
    // Clear Redis cache
    keys, _ := c.redis.Keys(context.Background(), pattern).Result()
    if len(keys) > 0 {
        c.redis.Del(context.Background(), keys...)
    }
    
    // Clear local cache
    c.local.DeleteExpired()
}
```

---

## Maintenance Procedures

### Routine Maintenance

**Weekly Maintenance Tasks:**
```bash
#!/bin/bash
# Weekly maintenance script

echo "Starting weekly maintenance $(date)"

# 1. Database maintenance
echo "Running database maintenance..."
psql $DATABASE_URL -c "VACUUM ANALYZE;"
psql $DATABASE_URL -c "REINDEX DATABASE finance_db;"

# 2. Log rotation and cleanup
echo "Cleaning up logs..."
find /var/log/finance-module -name "*.log" -mtime +7 -delete
logrotate /etc/logrotate.d/finance-module

# 3. Cache warming
echo "Warming application cache..."
curl -s "http://localhost:8080/admin/cache/warm" > /dev/null

# 4. Performance report generation
echo "Generating performance reports..."
python3 /scripts/generate-performance-report.py

# 5. Backup verification
echo "Verifying recent backups..."
./verify-backup.sh

echo "Weekly maintenance completed $(date)"
```

**Monthly Maintenance Tasks:**
```yaml
Monthly_Procedures:
  Database_Optimization:
    - Full database statistics update
    - Index fragmentation analysis
    - Table partition maintenance
    - Storage usage review
    
  Security_Review:
    - Access control audit
    - Certificate renewal check
    - Security patch application
    - Vulnerability scanning
    
  Performance_Analysis:
    - Query performance review
    - Capacity planning updates
    - SLA compliance reporting
    - Resource utilization analysis
```

### Incident Response

**Incident Classification:**
```yaml
Severity_Levels:
  P0_Critical:
    Description: "Complete system outage affecting all users"
    Response_Time: "15 minutes"
    Resolution_Target: "4 hours"
    Escalation: "Immediate C-level notification"
    
  P1_High:
    Description: "Major functionality impacted, affecting multiple tenants"
    Response_Time: "30 minutes" 
    Resolution_Target: "8 hours"
    Escalation: "VP Engineering notification"
    
  P2_Medium:
    Description: "Limited functionality impacted, workaround available"
    Response_Time: "2 hours"
    Resolution_Target: "24 hours"
    Escalation: "Team lead notification"
    
  P3_Low:
    Description: "Minor issues, cosmetic problems"
    Response_Time: "8 hours"
    Resolution_Target: "72 hours"
    Escalation: "Standard team notification"
```

**Incident Response Playbook:**
```bash
#!/bin/bash
# Incident response automation

INCIDENT_ID=$1
SEVERITY=$2

echo "Incident $INCIDENT_ID (Severity: $SEVERITY) detected at $(date)"

# 1. Alert team based on severity
case $SEVERITY in
    "P0"|"P1")
        # Page on-call engineer
        curl -X POST "https://api.pagerduty.com/incidents" \
             -H "Authorization: Token $PAGERDUTY_TOKEN" \
             -d "{\"incident\":{\"type\":\"incident\",\"title\":\"Finance Module Incident $INCIDENT_ID\",\"service\":{\"id\":\"$SERVICE_ID\"},\"urgency\":\"high\"}}"
        ;;
    "P2"|"P3")
        # Slack notification
        curl -X POST -H 'Content-type: application/json' \
             --data "{\"text\":\"Finance Module Incident $INCIDENT_ID (Severity: $SEVERITY)\"}" \
             $SLACK_WEBHOOK
        ;;
esac

# 2. Gather diagnostic information
kubectl logs deployment/finance-api --tail=1000 > /tmp/incident-${INCIDENT_ID}-logs.txt
kubectl describe pods -l app=finance-api > /tmp/incident-${INCIDENT_ID}-pods.txt

# 3. Run automated recovery procedures
if [ "$SEVERITY" = "P0" ]; then
    echo "Running automated recovery procedures..."
    kubectl rollout restart deployment/finance-api
    ./scripts/cache-clear.sh
    ./scripts/health-check.sh
fi

# 4. Generate incident report template
cat > /tmp/incident-${INCIDENT_ID}-report.md << EOF
# Incident Report: $INCIDENT_ID

**Severity:** $SEVERITY
**Detection Time:** $(date)
**Status:** Investigating

## Timeline
- $(date): Incident detected

## Impact
- TBD

## Root Cause
- Under investigation

## Resolution
- TBD

## Post-Incident Actions
- [ ] Review monitoring alerts
- [ ] Update runbooks if needed
- [ ] Implement preventive measures
EOF

echo "Incident response initiated. Report template created at /tmp/incident-${INCIDENT_ID}-report.md"
```

---

**This comprehensive operations and security guide provides the foundation for maintaining a secure, compliant, and highly available financial module in production environments.**