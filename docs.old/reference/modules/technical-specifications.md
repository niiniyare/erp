# Technical Specifications

## ️ Overview

This document provides  technical specifications for the ERP system, including detailed architecture designs, database schemas, API specifications, integration patterns, and deployment configurations. It serves as the definitive technical reference for development, deployment, and maintenance teams.

##  Technology Stack

### Backend Technologies

```yaml
backend_stack:
  runtime_environment:
    language: "Go 1.21+"
    alternative: "Node.js 18+ (TypeScript)"
    container_runtime: "Docker 24+"
    orchestration: "Kubernetes 1.28+"
  
  web_framework:
    primary: "Gin (Go) / Fiber (Go)"
    alternative: "Express.js (Node.js)"
    features:
      - high_performance_routing
      - middleware_support
      - json_serialization
      - websocket_support
  
  databases:
    primary: "PostgreSQL 15+"
    cache: "Redis 7+"
    search: "Elasticsearch 8+"
    time_series: "InfluxDB 2.0+"
    document_store: "MongoDB 6+" # For audit logs and flexible schemas
  
  message_queues:
    primary: "Apache Kafka"
    alternative: "RabbitMQ"
    use_cases:
      - event_streaming
      - async_processing
      - inter_service_communication
  
  storage:
    object_storage: "MinIO / AWS S3"
    file_system: "Persistent Volumes"
    backup_storage: "AWS S3 / Azure Blob"
```

### Frontend Technologies

```yaml
frontend_stack:
  web_application:
    framework: "React 18+"
    language: "TypeScript 5+"
    bundler: "Vite 4+"
    routing: "React Router 6+"
    
  state_management:
    global_state: "Zustand / Redux Toolkit"
    server_state: "TanStack Query (React Query)"
    form_state: "React Hook Form"
    
  ui_framework:
    component_library: "Ant Design / Mantine"
    styling: "Tailwind CSS 3+"
    icons: "Lucide React"
    charts: "Recharts / Chart.js"
    
  mobile_applications:
    framework: "React Native 0.72+"
    alternative: "Flutter 3.10+"
    navigation: "React Navigation 6+"
    state_management: "Zustand"
```

### Infrastructure & DevOps

```yaml
infrastructure:
  cloud_platforms:
    primary: "Kubernetes (Cloud Agnostic)"
    supported_clouds:
      - "AWS EKS"
      - "Azure AKS" 
      - "Google GKE"
      - "On-premise"
  
  container_orchestration:
    platform: "Kubernetes 1.28+"
    service_mesh: "Istio 1.19+"
    ingress: "NGINX Ingress Controller"
    secrets_management: "Sealed Secrets / External Secrets"
  
  monitoring_observability:
    metrics: "Prometheus + Grafana"
    logging: "ELK Stack (Elasticsearch, Logstash, Kibana)"
    tracing: "Jaeger"
    apm: "New Relic / Datadog"
    uptime_monitoring: "UptimeRobot"
  
  ci_cd:
    version_control: "Git (GitHub/GitLab)"
    ci_cd_pipeline: "GitHub Actions / GitLab CI"
    artifact_registry: "Docker Hub / AWS ECR"
    deployment: "ArgoCD / Flux"
```

## ️ Database Architecture

### Multi-Tenant Database Design

#### Tenant Isolation Strategy

```sql
-- Schema-per-tenant approach for data isolation
CREATE SCHEMA IF NOT EXISTS tenant_${TENANT_ID};

-- Row-level security for shared tables
CREATE POLICY tenant_isolation ON shared_table_name
    FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Hybrid approach: Core tables shared, business tables isolated
CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    -- Shared across all tenants
);

-- Tenant-specific tables in dedicated schemas
CREATE TABLE tenant_${TENANT_ID}.sales_orders (
    id UUID PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL,
    -- Tenant-isolated data
);
```

#### Database Connection Management

```go
// Database connection pool management
type DatabaseManager struct {
    pools map[string]*sql.DB // tenant_id -> connection pool
    mutex sync.RWMutex
}

func (dm *DatabaseManager) GetTenantDB(tenantID string) (*sql.DB, error) {
    dm.mutex.RLock()
    if pool, exists := dm.pools[tenantID]; exists {
        dm.mutex.RUnlock()
        return pool, nil
    }
    dm.mutex.RUnlock()
    
    dm.mutex.Lock()
    defer dm.mutex.Unlock()
    
    // Double-check pattern
    if pool, exists := dm.pools[tenantID]; exists {
        return pool, nil
    }
    
    // Create new connection pool for tenant
    config := dm.getTenantDBConfig(tenantID)
    pool, err := sql.Open("postgres", config.ConnectionString)
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    pool.SetMaxOpenConns(config.MaxOpenConns)
    pool.SetMaxIdleConns(config.MaxIdleConns)
    pool.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    dm.pools[tenantID] = pool
    return pool, nil
}
```

### Database Performance Optimization

#### Indexing Strategy

```sql
-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY idx_sales_orders_tenant_date_status 
ON sales_orders(tenant_id, order_date DESC, status) 
WHERE status IN ('pending', 'confirmed', 'shipped');

-- Partial indexes for filtered queries
CREATE INDEX CONCURRENTLY idx_active_employees 
ON employees(tenant_id, department_id, hire_date) 
WHERE status = 'active';

-- Covering indexes for frequently accessed columns
CREATE INDEX CONCURRENTLY idx_inventory_items_covering 
ON inventory_items(tenant_id, item_code) 
INCLUDE (item_name, category_id, current_stock);

-- Function-based indexes for computed values
CREATE INDEX CONCURRENTLY idx_orders_total_amount_range
ON sales_orders(tenant_id, (CASE 
    WHEN total_amount < 100 THEN 'small'
    WHEN total_amount < 1000 THEN 'medium'
    ELSE 'large'
END));
```

#### Partitioning Strategy

```sql
-- Time-based partitioning for large transactional tables
CREATE TABLE audit_logs (
    id UUID DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB,
    PRIMARY KEY (id, event_timestamp)
) PARTITION BY RANGE (event_timestamp);

-- Create monthly partitions
CREATE TABLE audit_logs_y2025m01 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit_logs_y2025m02 PARTITION OF audit_logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Automated partition management
CREATE OR REPLACE FUNCTION create_monthly_partition(table_name TEXT, start_date DATE)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    end_date DATE;
BEGIN
    partition_name := table_name || '_y' || EXTRACT(year FROM start_date) || 'm' || 
                     LPAD(EXTRACT(month FROM start_date)::TEXT, 2, '0');
    end_date := start_date + INTERVAL '1 month';
    
    EXECUTE format('CREATE TABLE %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
                   partition_name, table_name, start_date, end_date);
                   
    EXECUTE format('CREATE INDEX ON %I (tenant_id, event_timestamp DESC)', partition_name);
END;
$$ LANGUAGE plpgsql;
```

### Database Backup and Recovery

```yaml
backup_strategy:
  continuous_backup:
    tool: "WAL-E / pgBackRest"
    frequency: "continuous_wal_streaming"
    retention: "30_days"
    encryption: "AES-256"
    
  point_in_time_recovery:
    granularity: "second_level"
    retention: "7_days"
    automated_testing: "daily"
    
  cross_region_replication:
    read_replicas: 2
    geographic_distribution: true
    failover_automation: true
    
  backup_verification:
    automated_restore_testing: "weekly"
    data_integrity_checks: "daily"
    recovery_time_objective: "15_minutes"
    recovery_point_objective: "1_minute"
```

##  API Architecture

### RESTful API Design

#### API Structure and Conventions

```yaml
api_conventions:
  base_url: "https://api.awo.com/v1"
  
  url_patterns:
    tenants: "/tenants/{tenant_id}"
    resources: "/tenants/{tenant_id}/{resource}"
    sub_resources: "/tenants/{tenant_id}/{resource}/{id}/{sub_resource}"
    
  http_methods:
    GET: "retrieve_resources"
    POST: "create_resources"
    PUT: "update_entire_resource"
    PATCH: "partial_update"
    DELETE: "remove_resource"
    
  status_codes:
    success: [200, 201, 202, 204]
    client_errors: [400, 401, 403, 404, 409, 422]
    server_errors: [500, 502, 503, 504]
    
  response_format:
    success: "data_envelope"
    error: "error_envelope_with_details"
    pagination: "cursor_based_pagination"
```

#### API Response Envelopes

```typescript
// Standard API Response Format
interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: APIError;
  meta?: ResponseMeta;
}

interface APIError {
  code: string;
  message: string;
  details?: Record<string, any>;
  trace_id?: string;
}

interface ResponseMeta {
  pagination?: PaginationMeta;
  rate_limit?: RateLimitMeta;
  cache?: CacheMeta;
  request_id: string;
  timestamp: string;
}

interface PaginationMeta {
  current_page: number;
  per_page: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
  next_cursor?: string;
  previous_cursor?: string;
}

// Example API Response
const salesOrderResponse: APIResponse<SalesOrder[]> = {
  success: true,
  data: [
    {
      id: "uuid",
      order_number: "SO-2025-001",
      customer_id: "uuid",
      total_amount: 1500.00,
      status: "confirmed"
    }
  ],
  meta: {
    pagination: {
      current_page: 1,
      per_page: 50,
      total_count: 250,
      total_pages: 5,
      has_next: true,
      has_previous: false
    },
    request_id: "req_abc123",
    timestamp: "2025-01-15T10:30:00Z"
  }
};
```

#### API Security Implementation

```go
// JWT Authentication Middleware
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := extractTokenFromHeader(c.GetHeader("Authorization"))
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        claims, err := validateJWT(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("permissions", claims.Permissions)
        
        c.Next()
    }
}

// Rate Limiting Middleware
func RateLimitMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        userID := c.GetString("user_id")
        
        // Create rate limit key (per user or per IP)
        key := fmt.Sprintf("rate_limit:%s:%s", clientIP, userID)
        
        if !limiter.Allow(key) {
            c.Header("X-RateLimit-Limit", "100")
            c.Header("X-RateLimit-Remaining", "0")
            c.Header("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Add(time.Hour).Unix())))
            
            c.JSON(429, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": 3600
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Permission-based Authorization
func RequirePermissions(permissions ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userPermissions := c.GetStringSlice("permissions")
        
        for _, required := range permissions {
            if !contains(userPermissions, required) {
                c.JSON(403, gin.H{
                    "error": "Insufficient permissions",
                    "required": required
                })
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}
```

### GraphQL API Design

#### Schema Definition

```graphql
# Core GraphQL Schema
scalar DateTime
scalar UUID
scalar JSON

type Query {
  # Tenant operations
  tenant(id: UUID!): Tenant
  
  # Sales operations
  salesOrders(
    filter: SalesOrderFilter
    pagination: PaginationInput
  ): SalesOrderConnection!
  
  salesOrder(id: UUID!): SalesOrder
  
  # Inventory operations
  inventoryItems(
    filter: InventoryItemFilter
    pagination: PaginationInput
  ): InventoryItemConnection!
  
  # Financial operations
  accounts(filter: AccountFilter): [Account!]!
  journalEntries(
    filter: JournalEntryFilter
    pagination: PaginationInput
  ): JournalEntryConnection!
}

type Mutation {
  # Sales operations
  createSalesOrder(input: CreateSalesOrderInput!): SalesOrderPayload!
  updateSalesOrder(id: UUID!, input: UpdateSalesOrderInput!): SalesOrderPayload!
  
  # Inventory operations
  createStockTransaction(input: CreateStockTransactionInput!): StockTransactionPayload!
  
  # Financial operations
  createJournalEntry(input: CreateJournalEntryInput!): JournalEntryPayload!
}

type Subscription {
  # Real-time updates
  salesOrderUpdates(tenantId: UUID!): SalesOrder!
  inventoryLevelChanges(itemId: UUID!): InventoryLevel!
  systemNotifications(userId: UUID!): Notification!
}

# Type definitions
type SalesOrder {
  id: UUID!
  orderNumber: String!
  customer: Customer!
  orderDate: DateTime!
  status: SalesOrderStatus!
  totalAmount: Float!
  lineItems: [SalesOrderLineItem!]!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type SalesOrderConnection {
  edges: [SalesOrderEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

enum SalesOrderStatus {
  DRAFT
  CONFIRMED
  SHIPPED
  DELIVERED
  CANCELLED
}

input SalesOrderFilter {
  status: SalesOrderStatus
  customerId: UUID
  dateRange: DateRangeInput
  amountRange: AmountRangeInput
}
```

#### GraphQL Resolver Implementation

```go
// GraphQL Resolver Structure
type Resolver struct {
    DB       *database.DB
    Services *services.Services
}

type QueryResolver struct{ *Resolver }
type MutationResolver struct{ *Resolver }
type SubscriptionResolver struct{ *Resolver }

// Sales Order Resolver
func (r *QueryResolver) SalesOrders(ctx context.Context, filter *model.SalesOrderFilter, pagination *model.PaginationInput) (*model.SalesOrderConnection, error) {
    tenantID := auth.GetTenantIDFromContext(ctx)
    
    // Build query with filters
    query := r.DB.WithTenant(tenantID).Model(&models.SalesOrder{})
    
    if filter != nil {
        if filter.Status != nil {
            query = query.Where("status = ?", *filter.Status)
        }
        if filter.CustomerID != nil {
            query = query.Where("customer_id = ?", *filter.CustomerID)
        }
        if filter.DateRange != nil {
            query = query.Where("order_date BETWEEN ? AND ?", filter.DateRange.Start, filter.DateRange.End)
        }
    }
    
    // Apply pagination
    var totalCount int64
    query.Count(&totalCount)
    
    offset := 0
    limit := 50
    if pagination != nil {
        if pagination.Offset != nil {
            offset = *pagination.Offset
        }
        if pagination.Limit != nil {
            limit = *pagination.Limit
        }
    }
    
    var salesOrders []models.SalesOrder
    err := query.Offset(offset).Limit(limit).Find(&salesOrders).Error
    if err != nil {
        return nil, err
    }
    
    // Convert to GraphQL types
    edges := make([]*model.SalesOrderEdge, len(salesOrders))
    for i, order := range salesOrders {
        edges[i] = &model.SalesOrderEdge{
            Node:   convertSalesOrderToGraphQL(order),
            Cursor: encodeCursor(order.ID),
        }
    }
    
    return &model.SalesOrderConnection{
        Edges: edges,
        PageInfo: &model.PageInfo{
            HasNextPage:     offset+limit < int(totalCount),
            HasPreviousPage: offset > 0,
            TotalCount:      int(totalCount),
        },
        TotalCount: int(totalCount),
    }, nil
}

// Real-time subscriptions using WebSocket
func (r *SubscriptionResolver) SalesOrderUpdates(ctx context.Context, tenantID string) (<-chan *model.SalesOrder, error) {
    ch := make(chan *model.SalesOrder)
    
    // Subscribe to Redis pub/sub for sales order updates
    subscriber := r.Services.Redis.Subscribe(ctx, fmt.Sprintf("sales_orders:%s", tenantID))
    
    go func() {
        defer close(ch)
        defer subscriber.Close()
        
        for {
            select {
            case msg := <-subscriber.Channel():
                var order model.SalesOrder
                if err := json.Unmarshal([]byte(msg.Payload), &order); err == nil {
                    ch <- &order
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return ch, nil
}
```

##  Event-Driven Architecture

### Event Sourcing Implementation

#### Event Store Design

```go
// Event Store Interface
type EventStore interface {
    SaveEvents(aggregateID string, events []Event, expectedVersion int) error
    GetEvents(aggregateID string, fromVersion int) ([]Event, error)
    GetSnapshot(aggregateID string) (*Snapshot, error)
    SaveSnapshot(snapshot Snapshot) error
}

// Event Structure
type Event struct {
    ID            string                 `json:"id"`
    AggregateID   string                 `json:"aggregate_id"`
    AggregateType string                 `json:"aggregate_type"`
    EventType     string                 `json:"event_type"`
    EventVersion  int                    `json:"event_version"`
    Data          map[string]interface{} `json:"data"`
    Metadata      map[string]interface{} `json:"metadata"`
    Timestamp     time.Time              `json:"timestamp"`
    UserID        string                 `json:"user_id"`
    TenantID      string                 `json:"tenant_id"`
}

// Aggregate Root
type AggregateRoot struct {
    ID            string
    Version       int
    Changes       []Event
    TenantID      string
}

func (ar *AggregateRoot) ApplyEvent(event Event) {
    ar.Changes = append(ar.Changes, event)
    ar.Version++
}

func (ar *AggregateRoot) GetUncommittedEvents() []Event {
    return ar.Changes
}

func (ar *AggregateRoot) ClearChanges() {
    ar.Changes = []Event{}
}

// Sales Order Aggregate Example
type SalesOrderAggregate struct {
    AggregateRoot
    OrderNumber   string
    CustomerID    string
    Status        string
    TotalAmount   decimal.Decimal
    LineItems     []SalesOrderLineItem
}

func (so *SalesOrderAggregate) CreateOrder(orderNumber, customerID string, lineItems []SalesOrderLineItem) error {
    if so.ID != "" {
        return errors.New("sales order already exists")
    }
    
    // Business logic validation
    if len(lineItems) == 0 {
        return errors.New("sales order must have at least one line item")
    }
    
    // Calculate total
    totalAmount := decimal.Zero
    for _, item := range lineItems {
        totalAmount = totalAmount.Add(item.LineTotal)
    }
    
    // Raise domain event
    event := Event{
        ID:            uuid.New().String(),
        AggregateID:   uuid.New().String(),
        AggregateType: "SalesOrder",
        EventType:     "SalesOrderCreated",
        EventVersion:  1,
        Data: map[string]interface{}{
            "order_number":  orderNumber,
            "customer_id":   customerID,
            "line_items":    lineItems,
            "total_amount":  totalAmount,
            "status":        "draft",
        },
        Timestamp: time.Now(),
        TenantID:  so.TenantID,
    }
    
    so.ApplyEvent(event)
    so.apply(event) // Apply to current state
    
    return nil
}

func (so *SalesOrderAggregate) apply(event Event) {
    switch event.EventType {
    case "SalesOrderCreated":
        so.ID = event.AggregateID
        so.OrderNumber = event.Data["order_number"].(string)
        so.CustomerID = event.Data["customer_id"].(string)
        so.Status = event.Data["status"].(string)
        so.TotalAmount = event.Data["total_amount"].(decimal.Decimal)
        // ... apply other fields
    case "SalesOrderStatusChanged":
        so.Status = event.Data["new_status"].(string)
    // ... handle other events
    }
}
```

### Message Queue Integration

#### Kafka Integration

```go
// Kafka Producer Configuration
type KafkaProducer struct {
    producer sarama.SyncProducer
    config   *KafkaConfig
}

type KafkaConfig struct {
    Brokers               []string
    RetryMax              int
    RequiredAcks          sarama.RequiredAcks
    CompressionType       sarama.CompressionCodec
    FlushFrequency        time.Duration
    BatchSize             int
}

func NewKafkaProducer(config *KafkaConfig) (*KafkaProducer, error) {
    saramaConfig := sarama.NewConfig()
    saramaConfig.Producer.RequiredAcks = config.RequiredAcks
    saramaConfig.Producer.Compression = config.CompressionType
    saramaConfig.Producer.Flush.Frequency = config.FlushFrequency
    saramaConfig.Producer.Flush.Messages = config.BatchSize
    saramaConfig.Producer.Return.Successes = true
    saramaConfig.Producer.Return.Errors = true
    saramaConfig.Producer.Retry.Max = config.RetryMax
    
    producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
    if err != nil {
        return nil, err
    }
    
    return &KafkaProducer{
        producer: producer,
        config:   config,
    }, nil
}

func (kp *KafkaProducer) PublishEvent(topic string, event Event) error {
    eventBytes, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    message := &sarama.ProducerMessage{
        Topic:     topic,
        Key:       sarama.StringEncoder(event.AggregateID),
        Value:     sarama.ByteEncoder(eventBytes),
        Headers: []sarama.RecordHeader{
            {
                Key:   []byte("tenant_id"),
                Value: []byte(event.TenantID),
            },
            {
                Key:   []byte("event_type"),
                Value: []byte(event.EventType),
            },
        },
        Timestamp: event.Timestamp,
    }
    
    partition, offset, err := kp.producer.SendMessage(message)
    if err != nil {
        return err
    }
    
    log.Printf("Event published to topic %s, partition %d, offset %d", topic, partition, offset)
    return nil
}

// Kafka Consumer for Event Processing
type EventConsumer struct {
    consumer sarama.ConsumerGroup
    handlers map[string]EventHandler
}

type EventHandler interface {
    Handle(ctx context.Context, event Event) error
}

func (ec *EventConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (ec *EventConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (ec *EventConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for {
        select {
        case message := <-claim.Messages():
            var event Event
            if err := json.Unmarshal(message.Value, &event); err != nil {
                log.Printf("Error unmarshaling event: %v", err)
                continue
            }
            
            handler, exists := ec.handlers[event.EventType]
            if !exists {
                log.Printf("No handler found for event type: %s", event.EventType)
                session.MarkMessage(message, "")
                continue
            }
            
            ctx := context.WithValue(context.Background(), "tenant_id", event.TenantID)
            if err := handler.Handle(ctx, event); err != nil {
                log.Printf("Error handling event: %v", err)
                // Implement retry logic or dead letter queue
                continue
            }
            
            session.MarkMessage(message, "")
            
        case <-session.Context().Done():
            return nil
        }
    }
}
```

##  Security Architecture

### Authentication & Authorization

#### OAuth 2.0 / OIDC Implementation

```yaml
auth_configuration:
  oauth2_settings:
    authorization_server: "https://auth.awo.com"
    client_id: "erp_client_web"
    client_secret: "${OAUTH_CLIENT_SECRET}"
    scopes: ["openid", "profile", "email", "erp:read", "erp:write"]
    
  jwt_configuration:
    issuer: "https://auth.awo.com"
    audience: "erp-api"
    algorithm: "RS256"
    public_key_url: "https://auth.awo.com/.well-known/jwks.json"
    token_expiry: "1h"
    refresh_token_expiry: "7d"
    
  session_management:
    session_store: "redis"
    session_timeout: "8h"
    max_concurrent_sessions: 5
    session_binding: "ip_and_user_agent"
```

#### Role-Based Access Control (RBAC)

```go
// RBAC Implementation
type Permission struct {
    ID       string `json:"id"`
    Resource string `json:"resource"` // e.g., "sales_orders"
    Action   string `json:"action"`   // e.g., "read", "write", "delete"
    Scope    string `json:"scope"`    // e.g., "tenant", "organization", "own"
}

type Role struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions"`
    TenantID    string       `json:"tenant_id"`
}

type RBACManager struct {
    cache      cache.Cache
    repository RoleRepository
}

func (rbac *RBACManager) CheckPermission(userID, tenantID, resource, action string) (bool, error) {
    // Get user roles from cache or database
    cacheKey := fmt.Sprintf("user_roles:%s:%s", tenantID, userID)
    roles, err := rbac.getUserRoles(cacheKey, userID, tenantID)
    if err != nil {
        return false, err
    }
    
    // Check if any role has the required permission
    for _, role := range roles {
        for _, permission := range role.Permissions {
            if rbac.matchesPermission(permission, resource, action) {
                return true, nil
            }
        }
    }
    
    return false, nil
}

func (rbac *RBACManager) matchesPermission(permission Permission, resource, action string) bool {
    // Exact match
    if permission.Resource == resource && permission.Action == action {
        return true
    }
    
    // Wildcard matches
    if permission.Resource == "*" || permission.Action == "*" {
        return true
    }
    
    // Hierarchical resource matching (e.g., "sales:*" matches "sales:orders")
    if strings.HasSuffix(permission.Resource, ":*") {
        prefix := strings.TrimSuffix(permission.Resource, ":*")
        return strings.HasPrefix(resource, prefix+":")
    }
    
    return false
}

// Middleware for API endpoint protection
func RequirePermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        tenantID := c.GetString("tenant_id")
        
        rbacManager := c.MustGet("rbac").(*RBACManager)
        
        hasPermission, err := rbacManager.CheckPermission(userID, tenantID, resource, action)
        if err != nil {
            c.JSON(500, gin.H{"error": "Authorization check failed"})
            c.Abort()
            return
        }
        
        if !hasPermission {
            c.JSON(403, gin.H{
                "error": "Forbidden",
                "message": fmt.Sprintf("Missing permission: %s:%s", resource, action),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Data Encryption

#### Encryption at Rest

```go
// Database Field Encryption
type EncryptedField struct {
    Value     []byte `gorm:"column:encrypted_value"`
    KeyID     string `gorm:"column:key_id"`
    Algorithm string `gorm:"column:algorithm"`
}

type EncryptionService struct {
    keys map[string][]byte // Key ID -> Encryption Key
    aead cipher.AEAD
}

func NewEncryptionService() (*EncryptionService, error) {
    // Load encryption keys from secure key management service
    keys := make(map[string][]byte)
    
    // Example: Load from environment or external key service
    masterKey := os.Getenv("MASTER_ENCRYPTION_KEY")
    if masterKey == "" {
        return nil, errors.New("master encryption key not found")
    }
    
    key, err := hex.DecodeString(masterKey)
    if err != nil {
        return nil, err
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    
    aead, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    keys["default"] = key
    
    return &EncryptionService{
        keys: keys,
        aead: aead,
    }, nil
}

func (es *EncryptionService) Encrypt(plaintext string, keyID string) (*EncryptedField, error) {
    if keyID == "" {
        keyID = "default"
    }
    
    nonce := make([]byte, es.aead.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := es.aead.Seal(nil, nonce, []byte(plaintext), nil)
    
    // Prepend nonce to ciphertext
    encrypted := append(nonce, ciphertext...)
    
    return &EncryptedField{
        Value:     encrypted,
        KeyID:     keyID,
        Algorithm: "AES-256-GCM",
    }, nil
}

func (es *EncryptionService) Decrypt(field *EncryptedField) (string, error) {
    if len(field.Value) < es.aead.NonceSize() {
        return "", errors.New("invalid encrypted data")
    }
    
    nonce := field.Value[:es.aead.NonceSize()]
    ciphertext := field.Value[es.aead.NonceSize():]
    
    plaintext, err := es.aead.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}

// GORM Hook for automatic encryption
func (ef *EncryptedField) BeforeCreate(tx *gorm.DB) error {
    // Encryption happens in application layer before saving
    return nil
}

func (ef *EncryptedField) AfterFind(tx *gorm.DB) error {
    // Decryption happens after loading from database
    return nil
}
```

This  technical specification provides the foundation for building a robust, scalable, and secure ERP system that can handle enterprise-level requirements while maintaining high performance and reliability.
