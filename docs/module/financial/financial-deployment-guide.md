# AWO ERP Financial Module - Deployment & Operations Guide

**Version**: 1.0  
**Date**: January 2025  
**Status**: Technical Operations Manual  

---

## Table of Contents

1. [Deployment Architecture](#deployment-architecture)
2. [Infrastructure Requirements](#infrastructure-requirements)
3. [Database Setup & Configuration](#database-setup--configuration)
4. [Application Deployment](#application-deployment)
5. [Security Configuration](#security-configuration)
6. [Monitoring & Observability](#monitoring--observability)
7. [Backup & Recovery](#backup--recovery)
8. [Performance Optimization](#performance-optimization)
9. [Troubleshooting Guide](#troubleshooting-guide)
10. [Maintenance Procedures](#maintenance-procedures)

---

## Deployment Architecture

### **Multi-Environment Strategy**

```
┌─────────────────────────────────────────────────────────────────┐
│                         Production                              │
│  Load Balancer → API Gateway → Financial Services → Database   │
│     (HA)           (HA)          (Multi-AZ)        (HA+RO)     │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Staging                               │
│  Load Balancer → API Gateway → Financial Services → Database   │
│    (Single)        (Single)       (Single)         (Single)    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Development                             │
│      Local/Docker → Financial Services → Local Database        │
│                          (Single)           (Docker)           │
└─────────────────────────────────────────────────────────────────┘
```

### **Microservices Deployment Pattern**
```yaml
Financial Module Services:
  financial-api: 
    replicas: 3 (prod), 2 (staging), 1 (dev)
    ports: [8080, 9090] # HTTP, gRPC
    
  financial-worker:
    replicas: 2 (prod), 1 (staging), 1 (dev)
    purpose: Background processing, reconciliation
    
  financial-scheduler:
    replicas: 1 (prod), 1 (staging), 0 (dev)
    purpose: Periodic tasks, report generation
    
  financial-migration:
    replicas: 1 (on-demand)
    purpose: Database schema updates
```

---

## Infrastructure Requirements

### **Hardware Requirements**

#### **Production Environment**
```yaml
API Services (per instance):
  CPU: 4 vCPUs
  Memory: 8 GB RAM
  Storage: 100 GB SSD
  Network: 1 Gbps

Worker Services (per instance):
  CPU: 2 vCPUs
  Memory: 4 GB RAM
  Storage: 50 GB SSD
  Network: 1 Gbps

Database Server:
  CPU: 8 vCPUs
  Memory: 32 GB RAM
  Storage: 1 TB NVMe SSD (RAID 10)
  Network: 10 Gbps
  Backup Storage: 5 TB

Cache Server (Redis):
  CPU: 4 vCPUs
  Memory: 16 GB RAM
  Storage: 200 GB SSD
  Network: 1 Gbps
```

#### **Staging Environment**
```yaml
API Services:
  CPU: 2 vCPUs
  Memory: 4 GB RAM
  Storage: 50 GB SSD

Database Server:
  CPU: 4 vCPUs
  Memory: 16 GB RAM
  Storage: 500 GB SSD

Cache Server:
  CPU: 2 vCPUs
  Memory: 8 GB RAM
  Storage: 100 GB SSD
```

### **Cloud Infrastructure (AWS)**
```yaml
Production Setup:
  VPC: Multi-AZ with private subnets
  Load Balancer: Application Load Balancer (ALB)
  Compute: ECS Fargate or EKS
  Database: RDS PostgreSQL 15+ (Multi-AZ)
  Cache: ElastiCache Redis (Cluster Mode)
  Storage: EFS for shared storage
  Backup: RDS Automated Backups + S3
  Monitoring: CloudWatch + Prometheus
  
Instance Types:
  API: c5.xlarge (4 vCPU, 8 GB)
  Worker: c5.large (2 vCPU, 4 GB)
  Database: r5.2xlarge (8 vCPU, 64 GB)
  Cache: r5.xlarge (4 vCPU, 30.5 GB)
```

### **Networking Requirements**
```yaml
Load Balancer:
  - HTTPS Termination (TLS 1.3)
  - Health Check Endpoints
  - Rate Limiting
  - WAF Protection

Firewall Rules:
  Inbound:
    - 443 (HTTPS) from Internet
    - 80 (HTTP) redirect to HTTPS
    - 9090 (gRPC) from internal services
  
  Outbound:
    - 443 (HTTPS) to external APIs
    - 5432 (PostgreSQL) to database
    - 6379 (Redis) to cache
    - 53 (DNS) for name resolution

Security Groups:
  api-sg: 80,443 from ALB
  worker-sg: Internal communication only
  db-sg: 5432 from api-sg, worker-sg
  cache-sg: 6379 from api-sg, worker-sg
```

---

## Database Setup & Configuration

### **PostgreSQL Configuration**
```sql
-- postgresql.conf optimizations
shared_buffers = '8GB'                    # 25% of total RAM
effective_cache_size = '24GB'             # 75% of total RAM
work_mem = '256MB'                        # For complex queries
maintenance_work_mem = '2GB'              # For maintenance operations
wal_buffers = '64MB'                      # WAL buffer size
checkpoint_completion_target = 0.9        # Spread checkpoints
random_page_cost = 1.1                    # SSD optimization
effective_io_concurrency = 200            # SSD concurrency

# Row Level Security
row_security = on                         # Enable RLS
shared_preload_libraries = 'pg_stat_statements,auto_explain'

# Logging
log_min_duration_statement = 1000         # Log slow queries
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on
```

### **Database Schema Deployment**
```bash
#!/bin/bash
# deploy-database.sh

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-awo_erp}
DB_USER=${DB_USER:-postgres}
MIGRATION_PATH=${MIGRATION_PATH:-./db/migration}

echo "🔄 Starting database deployment..."

# 1. Validate database connection
echo "📡 Testing database connection..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT version();" || exit 1

# 2. Create database if not exists
echo "🏗️  Creating database if not exists..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "
  SELECT 'CREATE DATABASE $DB_NAME' 
  WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')
  \gexec"

# 3. Run migrations
echo "⬆️  Running database migrations..."
migrate -path $MIGRATION_PATH -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=require" up

# 4. Verify critical tables
echo "✅ Verifying database schema..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
  SELECT COUNT(*) as table_count 
  FROM information_schema.tables 
  WHERE table_schema = 'public' 
  AND table_name LIKE 'finance_%';"

# 5. Setup Row-Level Security policies
echo "🔒 Configuring Row-Level Security..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ./db/scripts/setup_rls.sql

echo "✅ Database deployment completed successfully!"
```

### **Migration Scripts Example**
```sql
-- 067_finance_enums.up.sql
BEGIN;

-- Account type enumeration
CREATE TYPE account_type_enum AS ENUM (
    'receivable',
    'payable', 
    'bank',
    'cash',
    'credit_card',
    'equity',
    'expense',
    'fixed_asset',
    'inventory',
    'liability',
    'revenue'
);

-- Root type enumeration  
CREATE TYPE root_type_enum AS ENUM (
    'asset',
    'liability', 
    'equity',
    'income',
    'expense'
);

-- Transaction type enumeration
CREATE TYPE transaction_type_enum AS ENUM (
    'manual',
    'sales_invoice',
    'purchase_invoice',
    'payment',
    'receipt',
    'journal_entry'
);

-- Transaction status enumeration
CREATE TYPE transaction_status_enum AS ENUM (
    'draft',
    'posted',
    'cancelled',
    'reversed'
);

COMMIT;
```

```sql
-- 068_finance_core_tables.up.sql
BEGIN;

-- Chart of accounts table
CREATE TABLE finance_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    account_type account_type_enum NOT NULL,
    root_type root_type_enum NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    parent_account_id UUID REFERENCES finance_accounts(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    
    -- Constraints
    CONSTRAINT finance_accounts_tenant_code_unique UNIQUE (tenant_id, code),
    CONSTRAINT finance_accounts_currency_valid CHECK (LENGTH(currency) = 3)
);

-- Row Level Security
ALTER TABLE finance_accounts ENABLE ROW LEVEL SECURITY;

CREATE POLICY finance_accounts_tenant_isolation ON finance_accounts
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Indexes for performance
CREATE INDEX idx_finance_accounts_tenant_id ON finance_accounts(tenant_id);
CREATE INDEX idx_finance_accounts_code ON finance_accounts(tenant_id, code);
CREATE INDEX idx_finance_accounts_type ON finance_accounts(tenant_id, account_type);
CREATE INDEX idx_finance_accounts_parent ON finance_accounts(parent_account_id);

COMMIT;
```

---

## Application Deployment

### **Docker Configuration**
```dockerfile
# Dockerfile.financial-api
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o financial-api cmd/financial-api/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/financial-api .
COPY --from=builder /app/configs ./configs

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080 9090
CMD ["./financial-api"]
```

### **Kubernetes Deployment**
```yaml
# k8s/financial-api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: financial-api
  labels:
    app: financial-api
    module: finance
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: financial-api
  template:
    metadata:
      labels:
        app: financial-api
        module: finance
    spec:
      containers:
      - name: financial-api
        image: awo-erp/financial-api:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: cache-secret
              key: url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /root/configs
      volumes:
      - name: config
        configMap:
          name: financial-config
```

### **Service Configuration**
```yaml
# k8s/financial-api-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: financial-api-service
  labels:
    app: financial-api
spec:
  selector:
    app: financial-api
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  type: ClusterIP

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: financial-api-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - api.awo-erp.com
    secretName: api-tls-secret
  rules:
  - host: api.awo-erp.com
    http:
      paths:
      - path: /v1/finance
        pathType: Prefix
        backend:
          service:
            name: financial-api-service
            port:
              number: 80
```

### **Configuration Management**
```yaml
# k8s/financial-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: financial-config
data:
  config.yaml: |
    server:
      port: 8080
      grpc_port: 9090
      read_timeout: 30s
      write_timeout: 30s
      idle_timeout: 120s
      
    database:
      max_open_conns: 25
      max_idle_conns: 5
      conn_max_lifetime: 300s
      
    cache:
      default_ttl: 300s
      max_memory_mb: 100
      
    logging:
      level: info
      format: json
      
    observability:
      metrics_enabled: true
      tracing_enabled: true
      jaeger_endpoint: "http://jaeger:14268/api/traces"
      
    features:
      multi_currency: true
      real_time_reconciliation: true
      automated_workflows: true
```

---

## Security Configuration

### **TLS/SSL Configuration**
```yaml
# TLS Configuration
TLS_VERSION: "1.3"
CIPHER_SUITES:
  - TLS_AES_256_GCM_SHA384
  - TLS_CHACHA20_POLY1305_SHA256
  - TLS_AES_128_GCM_SHA256

CERTIFICATE_MANAGEMENT:
  provider: "cert-manager"
  issuer: "letsencrypt-prod"
  auto_renewal: true
  renewal_threshold: "720h" # 30 days
```

### **Secrets Management**
```bash
#!/bin/bash
# setup-secrets.sh

# Database credentials
kubectl create secret generic database-secret \
  --from-literal=url="postgres://user:password@host:5432/dbname?sslmode=require"

# Cache credentials  
kubectl create secret generic cache-secret \
  --from-literal=url="redis://user:password@host:6379/0"

# JWT signing key
kubectl create secret generic jwt-secret \
  --from-literal=secret="$(openssl rand -base64 32)"

# API keys for external services
kubectl create secret generic external-apis \
  --from-literal=bank_api_key="bank-api-key" \
  --from-literal=payment_gateway_key="payment-gateway-key"
```

### **RBAC Configuration**
```yaml
# k8s/rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: financial-api-role
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: financial-api-binding
subjects:
- kind: ServiceAccount
  name: financial-api-sa
  namespace: default
roleRef:
  kind: Role
  name: financial-api-role
  apiGroup: rbac.authorization.k8s.io
```

---

## Monitoring & Observability

### **Prometheus Metrics Configuration**
```yaml
# prometheus-config.yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
- job_name: 'financial-api'
  static_configs:
  - targets: ['financial-api-service:8080']
  metrics_path: /metrics
  scrape_interval: 10s
  
- job_name: 'financial-worker'
  static_configs:
  - targets: ['financial-worker-service:8080']
  metrics_path: /metrics
  scrape_interval: 30s

rule_files:
- "financial_alerts.yml"

alerting:
  alertmanagers:
  - static_configs:
    - targets: ['alertmanager:9093']
```

### **Alert Rules**
```yaml
# financial_alerts.yml
groups:
- name: financial.rules
  rules:
  - alert: HighTransactionFailureRate
    expr: rate(financial_transactions_failed_total[5m]) > 0.01
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High transaction failure rate detected"
      description: "Transaction failure rate is {{ $value }} per second"
      
  - alert: DatabaseConnectionsHigh
    expr: finance_db_connections_active / finance_db_connections_max > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Database connection pool nearly exhausted"
      
  - alert: LongRunningTransaction
    expr: finance_transaction_processing_duration_seconds > 30
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "Transaction processing taking too long"
```

### **Grafana Dashboard Configuration**
```json
{
  "dashboard": {
    "title": "AWO ERP Financial Module",
    "panels": [
      {
        "title": "Transaction Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(financial_transactions_total[1m])",
            "legendFormat": "Transactions/sec"
          }
        ]
      },
      {
        "title": "Response Times",
        "type": "graph", 
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(financial_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(financial_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "Median"
          }
        ]
      },
      {
        "title": "Database Performance",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(financial_db_query_duration_seconds_sum[5m]) / rate(financial_db_query_duration_seconds_count[5m])",
            "legendFormat": "Avg Query Time"
          }
        ]
      }
    ]
  }
}
```

---

## Backup & Recovery

### **Database Backup Strategy**
```bash
#!/bin/bash
# backup-database.sh

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-awo_erp}
DB_USER=${DB_USER:-postgres}
BACKUP_DIR=${BACKUP_DIR:-/backups}
RETENTION_DAYS=${RETENTION_DAYS:-30}

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/finance_backup_$DATE.sql.gz"

echo "🔄 Starting database backup..."

# Create backup directory
mkdir -p $BACKUP_DIR

# Perform backup with compression
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
  --verbose \
  --format=custom \
  --compress=9 \
  --no-owner \
  --no-privileges \
  | gzip > $BACKUP_FILE

# Verify backup
if [ -f "$BACKUP_FILE" ] && [ -s "$BACKUP_FILE" ]; then
    echo "✅ Backup completed successfully: $BACKUP_FILE"
    
    # Upload to S3 (if configured)
    if [ -n "$AWS_S3_BUCKET" ]; then
        aws s3 cp $BACKUP_FILE s3://$AWS_S3_BUCKET/database-backups/
        echo "📤 Backup uploaded to S3"
    fi
else
    echo "❌ Backup failed"
    exit 1
fi

# Cleanup old backups
find $BACKUP_DIR -name "finance_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete
echo "🧹 Cleanup completed"
```

### **Disaster Recovery Procedure**
```bash
#!/bin/bash
# disaster-recovery.sh

set -e

BACKUP_FILE=${1:-latest}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-awo_erp_recovery}
DB_USER=${DB_USER:-postgres}

echo "🚨 Starting disaster recovery procedure..."

# 1. Create recovery database
echo "🏗️  Creating recovery database..."
createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME

# 2. Restore from backup
if [ "$BACKUP_FILE" = "latest" ]; then
    BACKUP_FILE=$(ls -t /backups/finance_backup_*.sql.gz | head -1)
fi

echo "📦 Restoring from backup: $BACKUP_FILE"
gunzip -c $BACKUP_FILE | pg_restore -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME --verbose

# 3. Verify data integrity
echo "🔍 Verifying data integrity..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
  SELECT 
    schemaname,
    tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes
  FROM pg_stat_user_tables 
  WHERE schemaname = 'public' 
  AND tablename LIKE 'finance_%'
  ORDER BY tablename;"

echo "✅ Disaster recovery completed"
echo "⚠️  Remember to update application configuration to point to recovery database"
```

---

## Performance Optimization

### **Database Query Optimization**
```sql
-- Common performance queries
-- 1. Identify slow queries
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE query LIKE '%finance_%'
ORDER BY total_time DESC
LIMIT 20;

-- 2. Index usage analysis
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

-- 3. Table size analysis  
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
FROM pg_tables 
WHERE schemaname = 'public' 
AND tablename LIKE 'finance_%'
ORDER BY size_bytes DESC;
```

### **Application-Level Caching**
```go
// Redis caching configuration
package cache

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type FinancialCache struct {
    client *redis.Client
}

func NewFinancialCache(redisURL string) *FinancialCache {
    opts, _ := redis.ParseURL(redisURL)
    client := redis.NewClient(opts)
    
    return &FinancialCache{client: client}
}

// Cache account balance with TTL
func (c *FinancialCache) SetAccountBalance(ctx context.Context, accountID string, balance AccountBalance) error {
    data, err := json.Marshal(balance)
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("account:balance:%s", accountID)
    return c.client.Set(ctx, key, data, 5*time.Minute).Err()
}

// Cache financial reports
func (c *FinancialCache) SetReport(ctx context.Context, reportKey string, report interface{}) error {
    data, err := json.Marshal(report)
    if err != nil {
        return err
    }
    
    key := fmt.Sprintf("report:%s", reportKey)
    return c.client.Set(ctx, key, data, 30*time.Minute).Err()
}
```

### **Connection Pooling Configuration**
```go
// Database connection pool optimization
func NewDatabasePool(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, err
    }
    
    // Connection pool settings
    db.SetMaxOpenConns(25)          // Maximum connections
    db.SetMaxIdleConns(5)           // Idle connections
    db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
    db.SetConnMaxIdleTime(2 * time.Minute) // Idle timeout
    
    return db, nil
}
```

---

## Troubleshooting Guide

### **Common Issues & Solutions**

#### **1. Database Connection Issues**
```bash
# Symptoms: Connection timeouts, pool exhaustion
# Diagnosis:
kubectl logs -f deployment/financial-api | grep "database"
psql -h $DB_HOST -U $DB_USER -c "SELECT count(*) FROM pg_stat_activity;"

# Solutions:
1. Check connection pool settings
2. Verify database server health
3. Review network connectivity
4. Scale database if needed
```

#### **2. High Memory Usage**
```bash
# Diagnosis:
kubectl top pods -l app=financial-api
kubectl describe pod <pod-name>

# Solutions:
1. Review query complexity
2. Optimize cache usage
3. Check for memory leaks
4. Increase pod memory limits
```

#### **3. Transaction Processing Failures**
```sql
-- Diagnosis queries
SELECT 
    transaction_id,
    status,
    error_message,
    created_at
FROM financial_transaction_logs 
WHERE status = 'failed' 
AND created_at > NOW() - INTERVAL '1 hour'
ORDER BY created_at DESC;

-- Check for deadlocks
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
```

### **Debug Mode Configuration**
```yaml
# Enable debug logging
apiVersion: v1
kind: ConfigMap
metadata:
  name: financial-debug-config
data:
  config.yaml: |
    logging:
      level: debug
      sql_debug: true
      include_stack_trace: true
      
    profiling:
      enabled: true
      port: 6060
      
    tracing:
      sample_rate: 1.0  # 100% sampling for debugging
```

---

## Maintenance Procedures

### **Regular Maintenance Tasks**

#### **Weekly Tasks**
```bash
#!/bin/bash
# weekly-maintenance.sh

echo "📅 Starting weekly maintenance..."

# 1. Database maintenance
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
  VACUUM ANALYZE;
  REINDEX DATABASE $DB_NAME;
"

# 2. Cache cleanup
redis-cli -h $REDIS_HOST FLUSHDB

# 3. Log rotation
find /var/log/financial -name "*.log" -mtime +7 -delete

# 4. Backup verification
./scripts/verify-backups.sh

echo "✅ Weekly maintenance completed"
```

#### **Monthly Tasks**
```bash
#!/bin/bash
# monthly-maintenance.sh

echo "📅 Starting monthly maintenance..."

# 1. Database statistics update
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
  VACUUM FULL;
  ANALYZE;
  UPDATE pg_class SET relpages = 0 WHERE relkind = 'r';
"

# 2. Certificate renewal check
./scripts/check-certificates.sh

# 3. Dependency updates check
./scripts/check-dependencies.sh

# 4. Performance review
./scripts/generate-performance-report.sh

echo "✅ Monthly maintenance completed"
```

### **Scaling Procedures**

#### **Horizontal Scaling**
```bash
#!/bin/bash
# scale-financial-api.sh

REPLICAS=${1:-3}

echo "🔄 Scaling financial API to $REPLICAS replicas..."

kubectl scale deployment financial-api --replicas=$REPLICAS

# Wait for rollout
kubectl rollout status deployment/financial-api --timeout=300s

# Verify scaling
kubectl get pods -l app=financial-api

echo "✅ Scaling completed"
```

#### **Database Scaling**
```sql
-- Read replica setup
CREATE SUBSCRIPTION financial_replica 
CONNECTION 'host=primary-db port=5432 user=replication dbname=awo_erp' 
PUBLICATION financial_publication;

-- Connection routing configuration
-- Read operations → Read replica
-- Write operations → Primary database
```

---

**Document Control**
- **Version**: 1.0
- **Last Updated**: January 2025
- **Next Review**: Quarterly
- **Approval Required**: DevOps Lead, Database Administrator, Security Team

**Related Documents**
- Financial Implementation Plan (@docs/module/financial/financial-implementation-plan.md)
- Security & Compliance Guide (@docs/module/financial/financial-security-compliance-guide.md)
- API Reference Guide (@docs/module/financial/financial-api-reference.md)