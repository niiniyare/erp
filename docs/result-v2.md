# Architectural Refinement and Production Readiness Enhancement

## Executive Summary

This document provides targeted architectural refinements for the multi-tenant organizational hierarchy system, focusing on leveraging existing Clean Architecture patterns and infrastructure. The recommendations prioritize production readiness while maintaining compatibility with Termux/Android constraints and established design patterns.

## 1. API Layer Completion and Consistency

### **Leveraging Service Facade Pattern**

The existing `service.go` facade in the organization module provides an excellent foundation. Here's how to complete the partially implemented handlers:

#### **Enhanced Service Facade Integration**

```go
// internal/core/entity/service.go - Enhanced facade methods
type Service interface {
    // Existing methods...
    
    // Enhanced list method with comprehensive filtering
    ListEntitiesWithFilters(ctx context.Context, req ListEntitiesRequest) (*PaginatedEntityResult, error)
    
    // Bulk hierarchy operations
    GetEntityHierarchyTree(ctx context.Context, rootID uuid.UUID, depth int) (*EntityHierarchyTree, error)
    
    // Enhanced update with validation hooks
    UpdateEntityWithValidation(ctx context.Context, id uuid.UUID, req UpdateEntityRequest) (*Entity, error)
    
    // Safe archive with dependency checks
    ArchiveEntitySafely(ctx context.Context, id uuid.UUID) (*ArchiveResult, error)
}

// Enhanced return types for better API mapping
type PaginatedEntityResult struct {
    Entities    []*Entity           `json:"entities"`
    Pagination  *PaginationMetadata `json:"pagination"`
    Filters     *AppliedFilters     `json:"filters"`
}

type EntityHierarchyTree struct {
    Root        *EntityWithHierarchy `json:"root"`
    Children    []*EntityWithHierarchy `json:"children"`
    Ancestors   []*EntityWithHierarchy `json:"ancestors"`
    TotalNodes  int                    `json:"total_nodes"`
    MaxDepth    int                    `json:"max_depth"`
}

type ArchiveResult struct {
    ArchivedEntityID uuid.UUID `json:"archived_entity_id"`
    AffectedChildren int       `json:"affected_children"`
    Warnings         []string  `json:"warnings"`
}
```

#### **Complete List Handler Implementation**

```go
// internal/api/handlers/entity.go - Complete List implementation
func (h *OrganizationGoaHandler) List(ctx context.Context, p *organization.ListPayload) (*organization.ListResult, error) {
    // Leverage service facade with proper validation
    req := h.buildListRequest(p)
    
    // Validate through service layer (leverages existing validation)
    if err := h.entityService.ValidateListRequest(ctx, req); err != nil {
        return nil, organization.MakeBadRequest(err)
    }
    
    // Use enhanced service method
    result, err := h.entityService.ListEntitiesWithFilters(ctx, req)
    if err != nil {
        return nil, h.mapServiceError(err)
    }
    
    // Transform using consistent mapping patterns
    return h.transformToGoaListResult(result), nil
}

func (h *OrganizationGoaHandler) buildListRequest(p *organization.ListPayload) entity.ListEntitiesRequest {
    req := entity.ListEntitiesRequest{
        Limit:  int(p.PageSize),
        Offset: int((p.Page - 1) * p.PageSize),
    }
    
    // Apply filters through type-safe mapping
    if p.EntityType != nil {
        entityType := entity.EntityType(*p.EntityType)
        req.Type = &entityType
    }
    
    if p.Status != nil {
        isActive := (*p.Status == organization.OrganizationStatusActive)
        req.IsActive = &isActive
    }
    
    if p.ParentID != nil {
        if parentUUID, err := uuid.Parse(*p.ParentID); err == nil {
            req.ParentID = &parentUUID
        }
    }
    
    return req
}
```

### **Consistent Error Mapping Pattern**

```go
// internal/api/handlers/common_errors.go - Centralized error mapping
type ErrorMapper struct {
    logger  logger.Logger
    metrics metrics.Provider
}

func NewErrorMapper(logger logger.Logger, metrics metrics.Provider) *ErrorMapper {
    return &ErrorMapper{logger: logger, metrics: metrics}
}

func (em *ErrorMapper) MapServiceError(ctx context.Context, err error, operation string) error {
    // Record metrics
    em.metrics.IncrementCounter("api_errors_total", metrics.Fields{
        "operation": operation,
        "error_type": em.categorizeError(err),
    })
    
    // Map to appropriate GOA errors
    switch {
    case errors.Is(err, sharedErrors.ErrEntityNotFound):
        return organization.MakeNotFound(err)
    case errors.Is(err, sharedErrors.ErrEntityNameExists):
        return organization.MakeConflict(errors.New("organization name already exists"))
    case errors.Is(err, sharedErrors.ErrInvalidEntityType):
        return organization.MakeBadRequest(errors.New("invalid entity type"))
    case errors.Is(err, sharedErrors.ErrCircularReference):
        return organization.MakeConflict(errors.New("operation would create circular reference"))
    case errors.Is(err, sharedErrors.ErrEntityHasChildren):
        return organization.MakeConflict(errors.New("cannot archive entity with active children"))
    case errors.Is(err, sharedErrors.ErrTenantIsolationViolation):
        return organization.MakeForbidden(errors.New("operation not permitted"))
    default:
        // Log unexpected errors
        em.logger.ErrorContext(ctx, "Unmapped service error", logger.Fields{
            "error": err.Error(),
            "operation": operation,
        })
        return organization.MakeInternalError(err)
    }
}

func (em *ErrorMapper) categorizeError(err error) string {
    switch {
    case errors.Is(err, sharedErrors.ErrEntityNotFound):
        return "not_found"
    case errors.Is(err, sharedErrors.ErrValidationError):
        return "validation"
    case errors.Is(err, sharedErrors.ErrBusinessRuleViolation):
        return "business_rule"
    default:
        return "internal"
    }
}
```

### **Validation Logic Placement Strategy**

#### **Service Layer Validation (Primary)**

```go
// internal/core/entity/validation.go - Business validation in service layer
type EntityValidator struct {
    repository Repository
    rules      BusinessRules
}

func (v *EntityValidator) ValidateUpdateRequest(ctx context.Context, id uuid.UUID, req *UpdateEntityRequest) error {
    // Business logic validation belongs in service layer
    if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
        return sharedErrors.ErrInvalidEntityName
    }
    
    if req.ParentID != nil {
        if err := v.validateParentChange(ctx, id, *req.ParentID); err != nil {
            return err
        }
    }
    
    if req.Type != nil {
        if err := v.rules.ValidateEntityTypeChange(ctx, id, *req.Type); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *EntityValidator) validateParentChange(ctx context.Context, entityID, newParentID uuid.UUID) error {
    // Check circular reference
    isAncestor, err := v.repository.IsEntityAncestor(ctx, entityID, newParentID)
    if err != nil {
        return err
    }
    if isAncestor {
        return sharedErrors.ErrCircularReference
    }
    
    // Check depth limits
    return v.rules.ValidateHierarchyDepth(ctx, entityID, newParentID)
}
```

#### **API Layer Validation (Input Sanitization)**

```go
// internal/api/handlers/validation.go - Input validation in API layer
type PayloadValidator struct{}

func (pv *PayloadValidator) ValidateUpdatePayload(p *organization.UpdateOrganizationPayload) error {
    // Input format validation only
    if p.Name != nil && len(*p.Name) > 255 {
        return errors.New("name exceeds maximum length")
    }
    
    if p.Website != nil && *p.Website != "" {
        if !isValidURL(*p.Website) {
            return errors.New("invalid website URL format")
        }
    }
    
    if p.ParentID != nil {
        if _, err := uuid.Parse(*p.ParentID); err != nil {
            return errors.New("invalid parent_id format")
        }
    }
    
    return nil
}
```

## 2. Testing Strategy within Clean Architecture

### **Module-Based Testing Structure**

```go
// internal/core/entity/service_test.go - Service facade testing
type EntityServiceTestSuite struct {
    suite.Suite
    service        entity.Service
    mockRepo       *entity.MockRepository
    mockConfig     *configuration.MockService
    testTenantID   uuid.UUID
    testContext    context.Context
}

func (s *EntityServiceTestSuite) SetupTest() {
    s.mockRepo = entity.NewMockRepository(s.T())
    s.mockConfig = configuration.NewMockService(s.T())
    s.testTenantID = uuid.New()
    
    // Setup test context with tenant
    s.testContext = context.WithValue(context.Background(), "tenant_id", s.testTenantID)
    
    // Create service with mocked dependencies
    s.service = entity.NewService(s.mockRepo, s.mockConfig, createTestMetrics(), createTestTracing())
}

func (s *EntityServiceTestSuite) TestCreateEntity_Success() {
    // Arrange
    req := entity.CreateEntityRequest{
        Name: "Test Company",
        Code: "TEST_COMP",
        Type: entity.EntityTypeCompany,
    }
    
    expectedEntity := &entity.Entity{
        ID:   uuid.New(),
        Name: req.Name,
        Code: req.Code,
        Type: req.Type,
    }
    
    s.mockRepo.EXPECT().
        ValidateEntityName(s.testContext, req.Name, nil).
        Return(nil)
    
    s.mockRepo.EXPECT().
        ValidateEntityCode(s.testContext, req.Code, nil).
        Return(nil)
    
    s.mockRepo.EXPECT().
        Create(s.testContext, &req).
        Return(expectedEntity, nil)
    
    // Act
    result, err := s.service.CreateEntity(s.testContext, req)
    
    // Assert
    s.NoError(err)
    s.Equal(expectedEntity.Name, result.Name)
    s.Equal(expectedEntity.Code, result.Code)
    s.Equal(expectedEntity.Type, result.Type)
    
    s.mockRepo.AssertExpectations(s.T())
}

func TestEntityServiceTestSuite(t *testing.T) {
    suite.Run(t, new(EntityServiceTestSuite))
}
```

### **Cross-Module Integration Testing**

```go
// test/integration/hierarchy_operations_test.go - Integration testing across modules
type HierarchyIntegrationTestSuite struct {
    suite.Suite
    db              *sqlx.DB
    entityService   entity.Service
    configService   configuration.Service
    tenantService   tenant.Service
    testTenantID    uuid.UUID
    testContext     context.Context
}

func (s *HierarchyIntegrationTestSuite) SetupSuite() {
    // Setup test database
    s.db = setupTestDatabase(s.T())
    
    // Create real services (not mocked for integration tests)
    store := db.NewStore(s.db)
    
    s.entityService = entity.NewService(
        entity.NewRepository(store, createTestTracing(), createTestMetrics()),
        createTestTracing(),
        createTestMetrics(),
    )
    
    s.configService = configuration.NewService(
        configuration.NewRepository(store, createTestCache(), createTestMetrics()),
        createTestMetrics(),
    )
    
    s.tenantService = tenant.NewService(
        tenant.NewRepository(store, createTestMetrics()),
        createTestMetrics(),
    )
}

func (s *HierarchyIntegrationTestSuite) TestHierarchyMove_WithConfigurationInheritance() {
    // Create test tenant
    s.testTenantID = s.createTestTenant()
    s.testContext = s.setTenantContext(s.testTenantID)
    
    // Create hierarchy: Company -> Division -> Department
    company := s.createTestEntity("COMPANY", nil)
    division := s.createTestEntity("DIVISION", &company.ID)
    department := s.createTestEntity("DEPARTMENT", &division.ID)
    
    // Set configuration at company level
    s.setEntityConfiguration(company.ID, "finance", "default_currency", "USD")
    
    // Verify initial configuration inheritance
    config := s.getEffectiveConfiguration(department.ID, "finance", "default_currency")
    s.Equal("USD", config.Value)
    s.Equal("entity", string(config.Source))
    
    // Create new division
    newDivision := s.createTestEntity("DIVISION", &company.ID)
    
    // Set different configuration on new division
    s.setEntityConfiguration(newDivision.ID, "finance", "default_currency", "EUR")
    
    // Move department to new division
    err := s.entityService.UpdateEntity(s.testContext, department.ID, entity.UpdateEntityRequest{
        ParentID: &newDivision.ID,
    })
    s.NoError(err)
    
    // Verify configuration inheritance changed
    newConfig := s.getEffectiveConfiguration(department.ID, "finance", "default_currency")
    s.Equal("EUR", newConfig.Value)
    s.Equal("entity", string(newConfig.Source))
    
    // Verify hierarchy paths updated
    ancestors, err := s.entityService.GetEntityAncestors(s.testContext, department.ID)
    s.NoError(err)
    s.Len(ancestors, 2) // company, newDivision
    s.Equal(newDivision.ID, ancestors[0].ID)
    s.Equal(company.ID, ancestors[1].ID)
}
```

### **RLS Policy Testing Strategy**

```go
// test/integration/rls_test.go - Direct RLS testing
type RLSTestSuite struct {
    suite.Suite
    db       *sqlx.DB
    tenant1  uuid.UUID
    tenant2  uuid.UUID
}

func (s *RLSTestSuite) SetupSuite() {
    s.db = setupTestDatabase(s.T())
    s.tenant1 = uuid.New()
    s.tenant2 = uuid.New()
    
    // Create test tenants
    s.createTenant(s.tenant1, "Tenant 1")
    s.createTenant(s.tenant2, "Tenant 2")
}

func (s *RLSTestSuite) TestEntityTenantIsolation() {
    // Create entity in tenant1
    entity1ID := uuid.New()
    
    // Set tenant context for tenant1
    _, err := s.db.Exec("SELECT validate_and_set_tenant_context($1)", s.tenant1)
    s.NoError(err)
    
    // Insert entity
    _, err = s.db.Exec(`
        INSERT INTO entities (uuid, tenant_id, name, type, is_active, accrual_method, fy_start_month)
        VALUES ($1, current_tenant_id(), 'Test Entity', 'COMPANY', true, true, 1)
    `, entity1ID)
    s.NoError(err)
    
    // Switch to tenant2 context
    _, err = s.db.Exec("SELECT validate_and_set_tenant_context($1)", s.tenant2)
    s.NoError(err)
    
    // Try to access entity from tenant1 (should fail due to RLS)
    var count int
    err = s.db.QueryRow("SELECT COUNT(*) FROM entities WHERE uuid = $1", entity1ID).Scan(&count)
    s.NoError(err)
    s.Equal(0, count, "Entity from tenant1 should not be visible in tenant2 context")
    
    // Try to insert entity with tenant1's ID (should fail due to RLS CHECK)
    entity2ID := uuid.New()
    _, err = s.db.Exec(`
        INSERT INTO entities (uuid, tenant_id, name, type, is_active, accrual_method, fy_start_month)
        VALUES ($1, $2, 'Cross Tenant Entity', 'COMPANY', true, true, 1)
    `, entity2ID, s.tenant1)
    s.Error(err, "Should not be able to insert entity with different tenant_id")
}

func (s *RLSTestSuite) TestHierarchyPathIsolation() {
    // Test hierarchy_paths RLS policies
    path1ID := uuid.New()
    path2ID := uuid.New()
    
    // Set tenant1 context and create hierarchy path
    _, err := s.db.Exec("SELECT validate_and_set_tenant_context($1)", s.tenant1)
    s.NoError(err)
    
    _, err = s.db.Exec(`
        INSERT INTO hierarchy_paths (tenant_id, ancestor_id, descendant_id, depth)
        VALUES (current_tenant_id(), $1, $2, 1)
    `, path1ID, path2ID)
    s.NoError(err)
    
    // Switch to tenant2 and verify isolation
    _, err = s.db.Exec("SELECT validate_and_set_tenant_context($1)", s.tenant2)
    s.NoError(err)
    
    var count int
    err = s.db.QueryRow(`
        SELECT COUNT(*) FROM hierarchy_paths 
        WHERE ancestor_id = $1 AND descendant_id = $2
    `, path1ID, path2ID).Scan(&count)
    s.NoError(err)
    s.Equal(0, count, "Hierarchy path should not be visible across tenants")
}
```

## 3. Bulk Operations and Temporal Workflows

### **Temporal Workflow Integration with Service Facades**

```go
// internal/workflows/organization/bulk_import.go - Temporal workflow for bulk operations
type BulkImportWorkflow struct {
    entityService entity.Service
    auditService  audit.Service
}

// Workflow definition
func (w *BulkImportWorkflow) BulkImportOrganizations(ctx workflow.Context, req BulkImportRequest) (*BulkImportResult, error) {
    logger := workflow.GetLogger(ctx)
    
    // Activity options for reliable execution
    activityOptions := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOptions)
    
    result := &BulkImportResult{
        OperationID:    req.OperationID,
        TotalItems:     len(req.Organizations),
        ProcessedItems: 0,
        SuccessItems:   0,
        FailedItems:    0,
    }
    
    logger.Info("Starting bulk import workflow", "operation_id", req.OperationID, "total_items", len(req.Organizations))
    
    // Phase 1: Validation
    var validationResult ValidationResult
    err := workflow.ExecuteActivity(ctx, w.ValidateImportRequest, req).Get(ctx, &validationResult)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    if !validationResult.Valid {
        result.Errors = validationResult.Errors
        return result, nil
    }
    
    // Phase 2: Build dependency graph
    var dependencyGraph DependencyGraph
    err = workflow.ExecuteActivity(ctx, w.BuildDependencyGraph, req.Organizations).Get(ctx, &dependencyGraph)
    if err != nil {
        return nil, fmt.Errorf("dependency analysis failed: %w", err)
    }
    
    // Phase 3: Process in batches
    batches := dependencyGraph.GetProcessingBatches(req.BatchSize)
    
    for i, batch := range batches {
        logger.Info("Processing batch", "batch_number", i+1, "batch_size", len(batch))
        
        var batchResult BatchResult
        err = workflow.ExecuteActivity(ctx, w.ProcessBatch, batch, req, result.EntityMappings).Get(ctx, &batchResult)
        if err != nil {
            logger.Error("Batch processing failed", "batch_number", i+1, "error", err)
            return nil, fmt.Errorf("batch %d failed: %w", i+1, err)
        }
        
        // Update progress
        result.ProcessedItems += batchResult.ProcessedCount
        result.SuccessItems += batchResult.SuccessCount
        result.FailedItems += batchResult.FailureCount
        result.Errors = append(result.Errors, batchResult.Errors...)
        
        // Merge entity mappings
        for tempID, realID := range batchResult.EntityMappings {
            result.EntityMappings[tempID] = realID
        }
        
        // Send progress update
        progressUpdate := ProgressUpdate{
            OperationID:    req.OperationID,
            ProcessedItems: result.ProcessedItems,
            TotalItems:     result.TotalItems,
            SuccessRate:    float64(result.SuccessItems) / float64(result.ProcessedItems),
        }
        
        err = workflow.ExecuteActivity(ctx, w.UpdateProgress, progressUpdate).Get(ctx, nil)
        if err != nil {
            logger.Warn("Failed to update progress", "error", err)
        }
        
        // Fail fast on critical errors
        if batchResult.CriticalFailure {
            return nil, fmt.Errorf("critical failure in batch %d", i+1)
        }
    }
    
    // Phase 4: Finalization
    err = workflow.ExecuteActivity(ctx, w.FinalizeImport, req, result).Get(ctx, nil)
    if err != nil {
        logger.Error("Import finalization failed", "error", err)
        return nil, fmt.Errorf("finalization failed: %w", err)
    }
    
    logger.Info("Bulk import completed", "operation_id", req.OperationID, 
        "success_items", result.SuccessItems, "failed_items", result.FailedItems)
    
    return result, nil
}

// Activity implementations using service facades
func (w *BulkImportWorkflow) ProcessBatch(ctx context.Context, batch []*OrganizationImport, 
    req BulkImportRequest, existingMappings map[string]uuid.UUID) (*BatchResult, error) {
    
    result := &BatchResult{
        ProcessedCount:   0,
        SuccessCount:     0,
        FailureCount:     0,
        EntityMappings:   make(map[string]uuid.UUID),
        Errors:          []ImportError{},
        CriticalFailure: false,
    }
    
    for _, org := range batch {
        // Use service facade for entity creation
        createReq := entity.CreateEntityRequest{
            Name:          org.Name,
            Code:          org.Code,
            Type:          org.EntityType,
            IsActive:      true,
            IsHidden:      false,
            AccrualMethod: org.AccrualMethod,
            FYStartMonth:  org.FiscalYearStart,
            Settings:      org.Settings,
            Metadata:      org.Metadata,
        }
        
        // Resolve parent from existing mappings
        if org.ParentTempID != nil {
            if parentID, exists := existingMappings[*org.ParentTempID]; exists {
                createReq.ParentID = &parentID
            } else {
                result.Errors = append(result.Errors, ImportError{
                    TempID:  org.TempID,
                    Type:    ErrorTypeParentNotFound,
                    Message: fmt.Sprintf("parent with temp_id %s not found", *org.ParentTempID),
                })
                result.FailureCount++
                result.ProcessedCount++
                continue
            }
        }
        
        // Create entity through service facade
        createdEntity, err := w.entityService.CreateEntity(ctx, createReq)
        if err != nil {
            result.Errors = append(result.Errors, ImportError{
                TempID:  org.TempID,
                Type:    ErrorTypeCreationFailed,
                Message: err.Error(),
            })
            result.FailureCount++
        } else {
            result.EntityMappings[org.TempID] = createdEntity.ID
            result.SuccessCount++
        }
        
        result.ProcessedCount++
    }
    
    return result, nil
}
```

### **Temporal Worker Configuration**

```go
// cmd/worker/main.go - Temporal worker setup
func main() {
    // Initialize services using existing service facades
    store := db.NewStore(mustConnectDB())
    
    entityService := entity.NewService(
        entity.NewRepository(store, tracing, metrics),
        tracing,
        metrics,
    )
    
    configService := configuration.NewService(
        configuration.NewRepository(store, cache, metrics),
        metrics,
    )
    
    auditService := audit.NewService(
        audit.NewRepository(store, metrics),
        metrics,
    )
    
    // Create Temporal client
    client, err := temporal.NewClient(temporal.ClientOptions{
        HostPort:  config.TemporalHostPort,
        Namespace: config.TemporalNamespace,
    })
    if err != nil {
        log.Fatalf("Failed to create Temporal client: %v", err)
    }
    defer client.Close()
    
    // Create worker
    worker := worker.New(client, config.TaskQueue, worker.Options{})
    
    // Register workflows and activities
    bulkImportWorkflow := &organization.BulkImportWorkflow{
        entityService: entityService,
        auditService:  auditService,
    }
    
    worker.RegisterWorkflow(bulkImportWorkflow.BulkImportOrganizations)
    worker.RegisterActivity(bulkImportWorkflow.ValidateImportRequest)
    worker.RegisterActivity(bulkImportWorkflow.BuildDependencyGraph)
    worker.RegisterActivity(bulkImportWorkflow.ProcessBatch)
    worker.RegisterActivity(bulkImportWorkflow.UpdateProgress)
    worker.RegisterActivity(bulkImportWorkflow.FinalizeImport)
    
    // Start worker
    err = worker.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalf("Worker failed: %v", err)
    }
}
```

## 4. Advanced Hierarchy Features Without ltree

### **Closure Table Optimization for Sub-tree Operations**

```go
// internal/core/entity/hierarchy_operations.go - Advanced hierarchy without ltree
type HierarchyOperations struct {
    repository Repository
    validator  HierarchyValidator
    cache      cache.Provider
    metrics    metrics.Provider
}

func NewHierarchyOperations(repo Repository, validator HierarchyValidator, 
    cache cache.Provider, metrics metrics.Provider) *HierarchyOperations {
    return &HierarchyOperations{
        repository: repo,
        validator:  validator,
        cache:      cache,
        metrics:    metrics,
    }
}

// Efficient sub-tree replication using closure table
func (h *HierarchyOperations) ReplicateSubtree(ctx context.Context, req SubtreeReplicationRequest) (*SubtreeReplicationResult, error) {
    // Use transaction for consistency
    return h.repository.WithTransaction(ctx, func(txCtx context.Context) (*SubtreeReplicationResult, error) {
        // Get source subtree using closure table
        sourceNodes, err := h.getSubtreeNodes(txCtx, req.SourceEntityID, req.Options.MaxDepth)
        if err != nil {
            return nil, fmt.Errorf("failed to get source subtree: %w", err)
        }
        
        // Build replication plan
        plan := h.buildReplicationPlan(sourceNodes, req)
        
        // Execute replication in dependency order
        result := &SubtreeReplicationResult{
            OperationID:     uuid.New(),
            SourceEntityID:  req.SourceEntityID,
            TargetParentID:  req.TargetParentID,
            CreatedEntities: make(map[uuid.UUID]uuid.UUID),
        }
        
        for level := 0; level <= plan.MaxLevel; level++ {
            levelNodes := plan.GetNodesAtLevel(level)
            
            for _, node := range levelNodes {
                newEntity, err := h.replicateNode(txCtx, node, req, result)
                if err != nil {
                    return nil, fmt.Errorf("failed to replicate node %s: %w", node.ID, err)
                }
                
                result.CreatedEntities[node.ID] = newEntity.ID
            }
        }
        
        return result, nil
    })
}

// Get subtree nodes using optimized closure table queries
func (h *HierarchyOperations) getSubtreeNodes(ctx context.Context, rootID uuid.UUID, maxDepth int) ([]*EntityNode, error) {
    // Cache key for subtree
    cacheKey := fmt.Sprintf("subtree:%s:%d", rootID.String(), maxDepth)
    
    var nodes []*EntityNode
    if err := h.cache.Get(ctx, cacheKey, &nodes); err == nil {
        h.metrics.IncrementCounter("hierarchy_cache_hit", metrics.Fields{"operation": "subtree"})
        return nodes, nil
    }
    
    // Query using closure table for efficiency
    query := `
        WITH subtree AS (
            SELECT e.*, hp.depth
            FROM entities e
            JOIN hierarchy_paths hp ON e.uuid = hp.descendant_id
            WHERE hp.tenant_id = current_tenant_id()
              AND hp.ancestor_id = $1
              AND ($2 = 0 OR hp.depth <= $2)
              AND e.deleted_at IS NULL
        )
        SELECT uuid, name, type, parent_id, depth, settings, metadata
        FROM subtree
        ORDER BY depth, name
    `
    
    rows, err := h.repository.Query(ctx, query, rootID, maxDepth)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    nodes = make([]*EntityNode, 0)
    for rows.Next() {
        node := &EntityNode{}
        err := rows.Scan(&node.ID, &node.Name, &node.Type, &node.ParentID, 
            &node.Depth, &node.Settings, &node.Metadata)
        if err != nil {
            return nil, err
        }
        nodes = append(nodes, node)
    }
    
    // Cache result
    h.cache.Set(ctx, cacheKey, nodes, 5*time.Minute)
    h.metrics.IncrementCounter("hierarchy_cache_miss", metrics.Fields{"operation": "subtree"})
    
    return nodes, nil
}
```

### **Depth Validation Using Standard SQL**

```go
// internal/core/entity/depth_validator.go - Depth validation without recursive CTEs
type DepthValidator struct {
    repository Repository
    maxDepth   int
    cache      cache.Provider
}

func (v *DepthValidator) ValidateEntityMove(ctx context.Context, entityID, newParentID uuid.UUID) error {
    // Get depths using optimized queries
    parentDepth, err := v.getEntityDepthFast(ctx, newParentID)
    if err != nil {
        return fmt.Errorf("failed to get parent depth: %w", err)
    }
    
    subtreeDepth, err := v.getSubtreeMaxDepthFast(ctx, entityID)
    if err != nil {
        return fmt.Errorf("failed to get subtree depth: %w", err)
    }
    
    // Calculate resulting depth
    newEntityDepth := parentDepth + 1
    maxResultingDepth := newEntityDepth + subtreeDepth
    
    if maxResultingDepth > v.maxDepth {
        return &HierarchyDepthError{
            EntityID:          entityID,
            NewParentID:       newParentID,
            TargetDepth:       newEntityDepth,
            SubtreeDepth:      subtreeDepth,
            MaxResultingDepth: maxResultingDepth,
            BusinessLimit:     v.maxDepth,
        }
    }
    
    return nil
}

// Fast depth calculation using closure table
func (v *DepthValidator) getEntityDepthFast(ctx context.Context, entityID uuid.UUID) (int, error) {
    cacheKey := fmt.Sprintf("entity_depth:%s", entityID.String())
    
    var depth int
    if err := v.cache.Get(ctx, cacheKey, &depth); err == nil {
        return depth, nil
    }
    
    // Single query using closure table max depth
    query := `
        SELECT COALESCE(MAX(depth), 0) as depth
        FROM hierarchy_paths
        WHERE tenant_id = current_tenant_id()
          AND descendant_id = $1
    `
    
    err := v.repository.QueryRow(ctx, query, entityID).Scan(&depth)
    if err != nil {
        return 0, err
    }
    
    // Cache for 5 minutes
    v.cache.Set(ctx, cacheKey, depth, 5*time.Minute)
    
    return depth, nil
}

func (v *DepthValidator) getSubtreeMaxDepthFast(ctx context.Context, entityID uuid.UUID) (int, error) {
    cacheKey := fmt.Sprintf("subtree_depth:%s", entityID.String())
    
    var depth int
    if err := v.cache.Get(ctx, cacheKey, &depth); err == nil {
        return depth, nil
    }
    
    // Single query to get maximum subtree depth
    query := `
        SELECT COALESCE(MAX(hp.depth), 0) as max_depth
        FROM hierarchy_paths hp
        WHERE hp.tenant_id = current_tenant_id()
          AND hp.ancestor_id = $1
    `
    
    err := v.repository.QueryRow(ctx, query, entityID).Scan(&depth)
    if err != nil {
        return 0, err
    }
    
    // Cache for 5 minutes
    v.cache.Set(ctx, cacheKey, depth, 5*time.Minute)
    
    return depth, nil
}
```

### **Circular Reference Detection in Application Logic**

```go
// internal/core/entity/circular_reference_detector.go - Application-level detection
type CircularReferenceDetector struct {
    repository Repository
    cache      cache.Provider
}

func (crd *CircularReferenceDetector) WouldCreateCircularReference(ctx context.Context, entityID, newParentID uuid.UUID) (bool, error) {
    // Check if newParent is a descendant of entity
    return crd.isDescendant(ctx, entityID, newParentID)
}

func (crd *CircularReferenceDetector) isDescendant(ctx context.Context, ancestorID, potentialDescendantID uuid.UUID) (bool, error) {
    cacheKey := fmt.Sprintf("is_descendant:%s:%s", ancestorID.String(), potentialDescendantID.String())
    
    var isDescendant bool
    if err := crd.cache.Get(ctx, cacheKey, &isDescendant); err == nil {
        return isDescendant, nil
    }
    
    // Use closure table for efficient lookup
    query := `
        SELECT EXISTS(
            SELECT 1 FROM hierarchy_paths
            WHERE tenant_id = current_tenant_id()
              AND ancestor_id = $1
              AND descendant_id = $2
              AND depth > 0
        ) as is_descendant
    `
    
    err := crd.repository.QueryRow(ctx, query, ancestorID, potentialDescendantID).Scan(&isDescendant)
    if err != nil {
        return false, err
    }
    
    // Cache result for 1 minute (shorter cache for dynamic relationships)
    crd.cache.Set(ctx, cacheKey, isDescendant, time.Minute)
    
    return isDescendant, nil
}

// Batch circular reference detection for bulk operations
func (crd *CircularReferenceDetector) ValidateBulkMoves(ctx context.Context, moves []EntityMove) ([]ValidationError, error) {
    // Build query for all moves at once
    var entityIDs, parentIDs []uuid.UUID
    for _, move := range moves {
        entityIDs = append(entityIDs, move.EntityID)
        parentIDs = append(parentIDs, move.NewParentID)
    }
    
    // Single query to check all potential circular references
    query := `
        SELECT 
            moves.entity_id,
            moves.new_parent_id,
            CASE WHEN hp.ancestor_id IS NOT NULL THEN true ELSE false END as would_create_cycle
        FROM (
            SELECT unnest($1::uuid[]) as entity_id, unnest($2::uuid[]) as new_parent_id
        ) moves
        LEFT JOIN hierarchy_paths hp ON hp.ancestor_id = moves.entity_id 
            AND hp.descendant_id = moves.new_parent_id
            AND hp.tenant_id = current_tenant_id()
            AND hp.depth > 0
    `
    
    rows, err := crd.repository.Query(ctx, query, pq.Array(entityIDs), pq.Array(parentIDs))
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var errors []ValidationError
    for rows.Next() {
        var entityID, parentID uuid.UUID
        var wouldCreateCycle bool
        
        err := rows.Scan(&entityID, &parentID, &wouldCreateCycle)
        if err != nil {
            return nil, err
        }
        
        if wouldCreateCycle {
            errors = append(errors, ValidationError{
                EntityID: entityID,
                Type:     ErrorTypeCircularReference,
                Message:  fmt.Sprintf("Moving entity %s to parent %s would create circular reference", entityID, parentID),
            })
        }
    }
    
    return errors, nil
}
```

## 5. Configuration Management Optimization

### **Hierarchy-Aware Cache Warming**

```go
// internal/core/configuration/cache_warmer.go - Smart cache warming
type HierarchyCacheWarmer struct {
    configService   configuration.Service
    entityService   entity.Service
    cache          cache.Provider
    metrics        metrics.Provider
}

func (w *HierarchyCacheWarmer) WarmConfigurationCache(ctx context.Context, tenantID uuid.UUID) error {
    // Get hot configurations from metrics
    hotConfigs := w.getHotConfigurations(tenantID)
    
    // Get organizational structure for efficient warming
    rootEntities, err := w.entityService.GetEntityRoots(ctx)
    if err != nil {
        return err
    }
    
    // Warm configurations level by level (inheritance pattern)
    for _, root := range rootEntities {
        if err := w.warmEntityHierarchy(ctx, root.ID, hotConfigs, 0, 5); err != nil {
            w.metrics.IncrementCounter("cache_warming_errors", metrics.Fields{
                "entity_id": root.ID.String(),
                "error":     err.Error(),
            })
        }
    }
    
    return nil
}

func (w *HierarchyCacheWarmer) warmEntityHierarchy(ctx context.Context, entityID uuid.UUID, 
    configs []ConfigurationKey, currentDepth, maxDepth int) error {
    
    if currentDepth > maxDepth {
        return nil
    }
    
    // Warm configurations for this entity
    for _, config := range configs {
        go func(cfg ConfigurationKey) {
            _, err := w.configService.GetEffectiveConfiguration(ctx, &configuration.GetConfigurationRequest{
                EntityID:  &entityID,
                Module:    cfg.Module,
                ConfigKey: cfg.Key,
            })
            if err != nil {
                w.metrics.IncrementCounter("cache_warming_failures", metrics.Fields{
                    "entity_id": entityID.String(),
                    "module":    string(cfg.Module),
                    "key":       string(cfg.Key),
                })
            }
        }(config)
    }
    
    // Recursively warm children
    children, err := w.entityService.GetEntityChildren(ctx, entityID)
    if err != nil {
        return err
    }
    
    for _, child := range children {
        if err := w.warmEntityHierarchy(ctx, child.ID, configs, currentDepth+1, maxDepth); err != nil {
            return err
        }
    }
    
    return nil
}

func (w *HierarchyCacheWarmer) getHotConfigurations(tenantID uuid.UUID) []ConfigurationKey {
    // Get from metrics (most accessed configurations)
    // This would be implemented based on your metrics collection
    return []ConfigurationKey{
        {Module: "finance", Key: "default_currency"},
        {Module: "finance", Key: "fiscal_year_start"},
        {Module: "general", Key: "timezone"},
        {Module: "general", Key: "business_hours"},
    }
}
```

### **Configuration Invalidation on Hierarchy Changes**

```go
// internal/core/configuration/hierarchy_listener.go - Cache invalidation
type HierarchyChangeListener struct {
    cache   cache.Provider
    metrics metrics.Provider
}

func (l *HierarchyChangeListener) OnEntityMoved(ctx context.Context, event EntityMovedEvent) error {
    // Invalidate configuration cache for moved entity and all descendants
    return l.invalidateEntityHierarchyCache(ctx, event.EntityID)
}

func (l *HierarchyChangeListener) OnEntityDeleted(ctx context.Context, event EntityDeletedEvent) error {
    // Invalidate cache for deleted entity
    return l.invalidateEntityCache(ctx, event.EntityID)
}

func (l *HierarchyChangeListener) invalidateEntityHierarchyCache(ctx context.Context, entityID uuid.UUID) error {
    // Pattern-based cache invalidation for entity and descendants
    patterns := []string{
        fmt.Sprintf("config:resolved:*:*:*:%s", entityID.String()),
        fmt.Sprintf("config:resolved:*:*:*:%s:*", entityID.String()),
    }
    
    for _, pattern := range patterns {
        count, err := l.cache.DeletePattern(ctx, pattern)
        if err != nil {
            return err
        }
        
        l.metrics.Gauge("cache_invalidation_count", float64(count), metrics.Fields{
            "pattern": pattern,
            "reason":  "hierarchy_change",
        })
    }
    
    return nil
}
```

## 6. Temporal for Workflow Orchestration

### **Complex Hierarchy Modification Workflows**

```go
// internal/workflows/organization/hierarchy_modification.go
type HierarchyModificationWorkflow struct {
    entityService entity.Service
    configService configuration.Service
    auditService  audit.Service
}

// Workflow for complex hierarchy restructuring
func (w *HierarchyModificationWorkflow) RestructureOrganization(ctx workflow.Context, req RestructureRequest) (*RestructureResult, error) {
    logger := workflow.GetLogger(ctx)
    
    activityOptions := workflow.ActivityOptions{
        StartToCloseTimeout: 15 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    time.Minute,
            MaximumAttempts:    3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOptions)
    
    result := &RestructureResult{
        OperationID: req.OperationID,
        StartTime:   workflow.Now(ctx),
    }
    
    // Phase 1: Validation and Planning
    logger.Info("Starting hierarchy restructure validation", "operation_id", req.OperationID)
    
    var validationResult ValidationResult
    err := workflow.ExecuteActivity(ctx, w.ValidateRestructurePlan, req).Get(ctx, &validationResult)
    if err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    if !validationResult.Valid {
        result.Status = "failed"
        result.Errors = validationResult.Errors
        return result, nil
    }
    
    // Phase 2: Create backup snapshot
    logger.Info("Creating backup snapshot")
    
    var backupID string
    err = workflow.ExecuteActivity(ctx, w.CreateHierarchySnapshot, req.TenantID).Get(ctx, &backupID)
    if err != nil {
        return nil, fmt.Errorf("backup creation failed: %w", err)
    }
    result.BackupID = backupID
    
    // Phase 3: Execute moves in dependency order
    logger.Info("Executing hierarchy moves", "move_count", len(req.Moves))
    
    for i, move := range req.Moves {
        logger.Info("Processing move", "move_number", i+1, "entity_id", move.EntityID)
        
        var moveResult MoveResult
        err = workflow.ExecuteActivity(ctx, w.ExecuteEntityMove, move).Get(ctx, &moveResult)
        if err != nil {
            // Rollback on failure
            logger.Error("Move failed, initiating rollback", "move_number", i+1, "error", err)
            
            rollbackErr := workflow.ExecuteActivity(ctx, w.RollbackToSnapshot, backupID).Get(ctx, nil)
            if rollbackErr != nil {
                logger.Error("Rollback failed", "error", rollbackErr)
                return nil, fmt.Errorf("move failed and rollback failed: %w", rollbackErr)
            }
            
            result.Status = "rolled_back"
            return result, fmt.Errorf("move %d failed: %w", i+1, err)
        }
        
        result.CompletedMoves = append(result.CompletedMoves, moveResult)
        
        // Send progress update
        progress := ProgressUpdate{
            OperationID:    req.OperationID,
            ProcessedItems: i + 1,
            TotalItems:     len(req.Moves),
            Progress:       float64(i+1) / float64(len(req.Moves)),
        }
        
        workflow.ExecuteActivity(ctx, w.UpdateProgress, progress)
    }
    
    // Phase 4: Configuration propagation
    logger.Info("Propagating configurations after restructure")
    
    err = workflow.ExecuteActivity(ctx, w.PropagateConfigurations, req.TenantID).Get(ctx, nil)
    if err != nil {
        logger.Warn("Configuration propagation failed", "error", err)
        // Don't fail the entire operation for config propagation failures
    }
    
    // Phase 5: Cleanup
    logger.Info("Cleaning up temporary resources")
    
    err = workflow.ExecuteActivity(ctx, w.CleanupSnapshot, backupID).Get(ctx, nil)
    if err != nil {
        logger.Warn("Snapshot cleanup failed", "error", err)
    }
    
    result.Status = "completed"
    result.EndTime = workflow.Now(ctx)
    
    logger.Info("Hierarchy restructure completed successfully", 
        "operation_id", req.OperationID, 
        "completed_moves", len(result.CompletedMoves))
    
    return result, nil
}

// Activity implementations
func (w *HierarchyModificationWorkflow) ExecuteEntityMove(ctx context.Context, move EntityMove) (*MoveResult, error) {
    // Use service facade for the actual move
    err := w.entityService.UpdateEntity(ctx, move.EntityID, entity.UpdateEntityRequest{
        ParentID: &move.NewParentID,
    })
    
    if err != nil {
        return nil, err
    }
    
    return &MoveResult{
        EntityID:      move.EntityID,
        OldParentID:   move.OldParentID,
        NewParentID:   move.NewParentID,
        CompletedAt:   time.Now(),
    }, nil
}

func (w *HierarchyModificationWorkflow) CreateHierarchySnapshot(ctx context.Context, tenantID uuid.UUID) (string, error) {
    // Create a snapshot of current hierarchy state for rollback
    snapshotID := uuid.New().String()
    
    // Implementation would save current state to a snapshot table
    // This is simplified for brevity
    
    return snapshotID, nil
}
```

### **Configuration Propagation Workflows**

```go
// internal/workflows/configuration/propagation.go
type ConfigurationPropagationWorkflow struct {
    configService configuration.Service
    entityService entity.Service
}

func (w *ConfigurationPropagationWorkflow) PropagateConfigurationChange(ctx workflow.Context, req ConfigPropagationRequest) error {
    logger := workflow.GetLogger(ctx)
    
    activityOptions := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval:    time.Second,
            BackoffCoefficient: 1.5,
            MaximumInterval:    30 * time.Second,
            MaximumAttempts:    5,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, activityOptions)
    
    logger.Info("Starting configuration propagation", 
        "entity_id", req.EntityID, 
        "module", req.Module, 
        "key", req.ConfigKey)
    
    // Get affected entities (descendants)
    var affectedEntities []uuid.UUID
    err := workflow.ExecuteActivity(ctx, w.GetAffectedEntities, req.EntityID).Get(ctx, &affectedEntities)
    if err != nil {
        return fmt.Errorf("failed to get affected entities: %w", err)
    }
    
    logger.Info("Found affected entities", "count", len(affectedEntities))
    
    // Process in batches to avoid overwhelming the system
    batchSize := 50
    for i := 0; i < len(affectedEntities); i += batchSize {
        end := i + batchSize
        if end > len(affectedEntities) {
            end = len(affectedEntities)
        }
        
        batch := affectedEntities[i:end]
        
        err = workflow.ExecuteActivity(ctx, w.InvalidateConfigurationCache, batch, req.Module, req.ConfigKey).Get(ctx, nil)
        if err != nil {
            logger.Warn("Cache invalidation failed for batch", "batch_start", i, "error", err)
            // Continue processing other batches
        }
        
        // Small delay between batches to prevent overwhelming the cache
        workflow.Sleep(ctx, 100*time.Millisecond)
    }
    
    logger.Info("Configuration propagation completed")
    
    return nil
}

func (w *ConfigurationPropagationWorkflow) GetAffectedEntities(ctx context.Context, entityID uuid.UUID) ([]uuid.UUID, error) {
    // Get all descendants of the entity
    descendants, err := w.entityService.GetEntityDescendants(ctx, entityID)
    if err != nil {
        return nil, err
    }
    
    entityIDs := make([]uuid.UUID, len(descendants))
    for i, entity := range descendants {
        entityIDs[i] = entity.ID
    }
    
    return entityIDs, nil
}

func (w *ConfigurationPropagationWorkflow) InvalidateConfigurationCache(ctx context.Context, 
    entityIDs []uuid.UUID, module domain.ModuleName, key domain.ConfigKey) error {
    
    // Use configuration service to invalidate cache for affected entities
    for _, entityID := range entityIDs {
        err := w.configService.InvalidateEntityConfiguration(ctx, entityID, module, key)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

---

## Implementation Roadmap and Prioritization

### **Phase 1: Foundation Enhancement (2-3 weeks)**
**Priority: HIGH - Production Critical**

1. **Complete API Handler Implementation**
   - Implement List, Update, Hierarchy, Archive handlers using service facade pattern
   - Establish consistent error mapping across all endpoints
   - Add comprehensive input validation at API layer

2. **Enhanced Service Layer Testing**
   - Implement service facade test suites using testify/suite
   - Add cross-module integration tests
   - Create RLS policy direct testing framework

### **Phase 2: Bulk Operations and Temporal (3-4 weeks)**
**Priority: HIGH - Business Value**

1. **Temporal Workflow Integration**
   - Implement bulk import workflows using existing service facades
   - Add hierarchy modification workflows with rollback capabilities
   - Create configuration propagation workflows

2. **Enhanced Bulk Operations**
   - Extend existing bulk operations pattern to organizational entities
   - Add dependency graph processing for correct creation order
   - Implement progress tracking and result handling

### **Phase 3: Advanced Hierarchy Features (2-3 weeks)**
**Priority: MEDIUM - Feature Enhancement**

1. **Sub-tree Replication**
   - Implement efficient replication using closure table
   - Add name transformation and configuration policy handling
   - Create replication validation and conflict resolution

2. **Depth Management**
   - Add proactive depth validation using standard SQL
   - Implement circular reference detection in application logic
   - Create monitoring and alerting for hierarchy health

### **Phase 4: Performance Optimization (1-2 weeks)**
**Priority: MEDIUM - Performance**

1. **Configuration Cache Enhancement**
   - Implement hierarchy-aware cache warming
   - Add intelligent cache invalidation on hierarchy changes
   - Create cache performance monitoring

2. **Query Optimization**
   - Optimize closure table queries for large hierarchies
   - Add query result caching for frequently accessed data
   - Implement batch operations for hierarchy modifications

### **Timeline Summary**
- **Total Duration**: 8-12 weeks
- **Critical Path**: API completion → Temporal integration → Advanced features
- **Risk Mitigation**: Phased approach allows for early delivery of core functionality

### **Success Metrics**
- **API Completeness**: 100% handler implementation with consistent error handling
- **Test Coverage**: >90% for service layer, >80% for integration tests
- **Performance**: <100ms response time for hierarchy operations, >95% cache hit rate
- **Reliability**: Zero data loss during bulk operations, automatic rollback on failures

This architectural refinement maintains compatibility with existing patterns while adding production-ready features that leverage the established Clean Architecture and service facade patterns.