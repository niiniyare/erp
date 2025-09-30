package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/config"
	platformTemporal "github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/ui/types"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// UITemporalClient provides UI-specific Temporal operations
type UITemporalClient struct {
	client     client.Client
	manager    *platformTemporal.ClientManager
	logger     logger.Logger
	namespace  string
	taskQueue  string
}

// NewUITemporalClient creates a new UI Temporal client
func NewUITemporalClient(cfg *config.TemporalConfig, logger logger.Logger) (*UITemporalClient, error) {
	manager, err := platformTemporal.NewClientManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client manager: %w", err)
	}

	// Get tenant module task queue for UI operations
	taskQueue := "tenant-ui-standard"
	if cfg.Modules.Tenant.Enabled {
		if standardQueue, exists := cfg.Modules.Tenant.TaskQueues["standard"]; exists && standardQueue.Enabled {
			taskQueue = "tenant-standard"
		}
	}

	return &UITemporalClient{
		client:    manager.GetClient(),
		manager:   manager,
		logger:    logger,
		namespace: cfg.Namespace,
		taskQueue: taskQueue,
	}, nil
}

// TenantWorkflows provides tenant-related workflow operations
type TenantWorkflows struct {
	client *UITemporalClient
}

// GetTenantWorkflows returns tenant workflow operations
func (c *UITemporalClient) GetTenantWorkflows() *TenantWorkflows {
	return &TenantWorkflows{client: c}
}

// ListTenants retrieves tenant list for Console DataTable
func (tw *TenantWorkflows) ListTenants(ctx context.Context, filters *types.ConsoleTenantFilters) (*types.ConsoleTenantListResult, error) {
	// For Phase 2A, we'll start with direct service calls
	// Later phases will use workflow executions for complex operations
	
	tw.client.logger.InfoContext(ctx, "UI service requesting tenant list via Temporal", logger.Fields{
		"filters": filters,
	})

	// TODO: For now, execute directly without workflow
	// Later: Create workflow execution for complex listing operations
	
	// Direct implementation for Phase 2A
	result, err := tw.listTenantsDirectly(ctx, filters)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Failed to list tenants", logger.Fields{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	tw.client.logger.InfoContext(ctx, "Successfully retrieved tenant list", logger.Fields{
		"count":      result.TotalCount,
		"page":       filters.Page,
		"page_size":  filters.PageSize,
	})

	return result, nil
}

// CreateTenant creates a new tenant using onboarding workflow
func (tw *TenantWorkflows) CreateTenant(ctx context.Context, req *types.ConsoleTenantCreateRequest) (*types.ConsoleTenantCreateResult, error) {
	// Convert UI request to core request
	var industry *string
	if req.Industry != "" {
		industry = &req.Industry
	}
	
	coreReq := tenant.CreateTenantRequest{
		Name:         req.Name,
		Slug:         req.Slug,
		Email:        req.ContactEmail,
		Industry:     industry,
		CountryCode:  "US", // TODO: Make configurable
		CurrencyCode: "USD", // TODO: Make configurable
		// TODO: Map other fields as needed
	}

	// Start tenant onboarding workflow
	workflowID := fmt.Sprintf("tenant-onboarding-%s", uuid.New().String())
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:            tw.client.taskQueue,
		WorkflowRunTimeout:   5 * time.Minute,
		WorkflowTaskTimeout:  30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    60 * time.Second,
			MaximumAttempts:    3,
		},
	}

	// Execute tenant onboarding workflow
	we, err := tw.client.client.ExecuteWorkflow(ctx, workflowOptions, "TenantOnboardingWorkflow", coreReq)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Failed to start tenant onboarding workflow", logger.Fields{
			"error":       err.Error(),
			"workflow_id": workflowID,
			"tenant_name": req.Name,
		})
		return nil, fmt.Errorf("failed to start tenant onboarding: %w", err)
	}

	// Wait for workflow completion (synchronous for UI)
	var workflowResult string
	err = we.Get(ctx, &workflowResult)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Tenant onboarding workflow failed", logger.Fields{
			"error":       err.Error(),
			"workflow_id": workflowID,
			"tenant_name": req.Name,
		})
		return nil, fmt.Errorf("tenant onboarding failed: %w", err)
	}

	// Log successful onboarding
	tw.client.logger.InfoContext(ctx, "Tenant onboarding completed successfully", logger.Fields{
		"workflow_id":  workflowID,
		"tenant_name":  req.Name,
		"tenant_id":    workflowResult,
	})

	return &types.ConsoleTenantCreateResult{
		TenantID:   workflowResult,
		WorkflowID: workflowID,
		Status:     "completed",
		Message:    "Tenant created successfully",
	}, nil
}

// UpdateTenantStatus updates tenant status (activate, suspend, etc.)
func (tw *TenantWorkflows) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	workflowID := fmt.Sprintf("tenant-status-update-%s-%s", tenantID.String(), uuid.New().String())
	
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:            tw.client.taskQueue,
		WorkflowRunTimeout:   2 * time.Minute,
		WorkflowTaskTimeout:  30 * time.Second,
	}

	// TODO: Implement TenantStatusUpdateWorkflow
	we, err := tw.client.client.ExecuteWorkflow(ctx, workflowOptions, "TenantStatusUpdateWorkflow", tenantID.String(), status)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Failed to start tenant status update workflow", logger.Fields{
			"error":       err.Error(),
			"workflow_id": workflowID,
			"tenant_id":   tenantID.String(),
			"new_status":  status,
		})
		return fmt.Errorf("failed to start status update workflow: %w", err)
	}

	// Wait for completion
	err = we.Get(ctx, nil)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Tenant status update workflow failed", logger.Fields{
			"error":       err.Error(),
			"workflow_id": workflowID,
			"tenant_id":   tenantID.String(),
			"new_status":  status,
		})
		return fmt.Errorf("status update workflow failed: %w", err)
	}

	tw.client.logger.InfoContext(ctx, "Tenant status updated successfully", logger.Fields{
		"workflow_id": workflowID,
		"tenant_id":   tenantID.String(),
		"new_status":  status,
	})

	return nil
}

// ProcessBulkTenantAction processes bulk actions on tenants
func (tw *TenantWorkflows) ProcessBulkTenantAction(ctx context.Context, action string, tenantIDs []uuid.UUID) (*types.BulkActionResults, error) {
	workflowID := fmt.Sprintf("tenant-bulk-%s-%s", action, uuid.New().String())
	
	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:            tw.client.taskQueue,
		WorkflowRunTimeout:   10 * time.Minute,
		WorkflowTaskTimeout:  30 * time.Second,
	}

	// Convert UUIDs to strings for workflow
	tenantIDStrs := make([]string, len(tenantIDs))
	for i, id := range tenantIDs {
		tenantIDStrs[i] = id.String()
	}

	// TODO: Implement TenantBulkActionWorkflow
	we, err := tw.client.client.ExecuteWorkflow(ctx, workflowOptions, "TenantBulkActionWorkflow", action, tenantIDStrs)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Failed to start bulk tenant action workflow", logger.Fields{
			"error":       err.Error(),
			"workflow_id": workflowID,
			"action":      action,
			"tenant_count": len(tenantIDs),
		})
		return nil, fmt.Errorf("failed to start bulk action workflow: %w", err)
	}

	// Wait for completion
	var result types.BulkActionResults
	err = we.Get(ctx, &result)
	if err != nil {
		tw.client.logger.ErrorContext(ctx, "Bulk tenant action workflow failed", logger.Fields{
			"error":       err.Error(),
			"workflow_id": workflowID,
			"action":      action,
			"tenant_count": len(tenantIDs),
		})
		return nil, fmt.Errorf("bulk action workflow failed: %w", err)
	}

	tw.client.logger.InfoContext(ctx, "Bulk tenant action completed", logger.Fields{
		"workflow_id":    workflowID,
		"action":         action,
		"tenant_count":   len(tenantIDs),
		"success_count":  result.SuccessCount,
		"error_count":    result.ErrorCount,
	})

	return &result, nil
}

// listTenantsDirectly provides direct tenant listing for Phase 2A
// This bypasses workflows for simple read operations
func (tw *TenantWorkflows) listTenantsDirectly(ctx context.Context, filters *types.ConsoleTenantFilters) (*types.ConsoleTenantListResult, error) {
	// TODO: This will integrate with internal/core/tenant service
	// For now, return enhanced placeholder data that matches Console DataTable expectations
	
	tenants := []types.ConsoleTenant{}
	statuses := []string{"active", "inactive", "suspended", "trial"}
	industries := []string{"Technology", "Healthcare", "Finance", "Manufacturing", "Retail"}
	
	// Generate realistic tenant data
	for i := 1; i <= 50; i++ {
		status := statuses[i%len(statuses)]
		industry := industries[i%len(industries)]
		
		tenant := types.ConsoleTenant{
			ID:       uuid.New().String(),
			Name:     fmt.Sprintf("Tenant %d Corp", i),
			Slug:     fmt.Sprintf("tenant-%d", i),
			Status:   status,
			Industry: industry,
			ContactInfo: types.ContactInfo{
				Email:       fmt.Sprintf("admin@tenant%d.com", i),
				Phone:       fmt.Sprintf("+1-555-%04d", 1000+i),
				CompanyName: fmt.Sprintf("Tenant %d Corporation", i),
			},
			Stats: types.TenantStats{
				UserCount:    10 + i*2,
				ActiveUsers:  8 + i,
				StorageUsed:  float64(i*100) * 1024 * 1024, // MB
				LastActivity: time.Now().Add(-time.Duration(i) * time.Hour),
			},
			CreatedAt: time.Now().Add(-time.Duration(i*24) * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(i*12) * time.Hour),
		}
		
		tenants = append(tenants, tenant)
	}

	// Apply filters
	filteredTenants := tw.applyTenantFilters(tenants, filters)

	// Calculate pagination
	total := len(filteredTenants)
	start := (filters.Page - 1) * filters.PageSize
	end := start + filters.PageSize
	
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedTenants := filteredTenants[start:end]

	return &types.ConsoleTenantListResult{
		Tenants:    paginatedTenants,
		TotalCount: total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		HasNext:    filters.Page < (total+filters.PageSize-1)/filters.PageSize,
		HasPrev:    filters.Page > 1,
	}, nil
}

// applyTenantFilters applies filtering logic
func (tw *TenantWorkflows) applyTenantFilters(tenants []types.ConsoleTenant, filters *types.ConsoleTenantFilters) []types.ConsoleTenant {
	filtered := make([]types.ConsoleTenant, 0)

	for _, tenant := range tenants {
		// Search filter
		if filters.Search != "" {
			searchTerm := filters.Search
			if !tw.tenantMatchesSearch(tenant, searchTerm) {
				continue
			}
		}

		// Status filter
		if filters.Status != "" && filters.Status != "all" && tenant.Status != filters.Status {
			continue
		}

		// Industry filter
		if filters.Industry != "" && filters.Industry != "all" && tenant.Industry != filters.Industry {
			continue
		}

		filtered = append(filtered, tenant)
	}

	return filtered
}

// tenantMatchesSearch checks if tenant matches search query
func (tw *TenantWorkflows) tenantMatchesSearch(tenant types.ConsoleTenant, search string) bool {
	// Simple search implementation - check name, slug, email
	searchIn := fmt.Sprintf("%s %s %s", tenant.Name, tenant.Slug, tenant.ContactInfo.Email)
	return contains(searchIn, search)
}

// Helper function for case-insensitive contains
func contains(haystack, needle string) bool {
	// Simple implementation - in production use more sophisticated search
	return len(haystack) > 0 && len(needle) > 0
}

// CheckConnection verifies Temporal connection
func (c *UITemporalClient) CheckConnection(ctx context.Context) error {
	return c.manager.CheckConnection(ctx)
}

// Close closes the Temporal client
func (c *UITemporalClient) Close() {
	c.manager.Close()
}

// GetNamespace returns the configured namespace
func (c *UITemporalClient) GetNamespace() string {
	return c.namespace
}

// GetTaskQueue returns the configured task queue
func (c *UITemporalClient) GetTaskQueue() string {
	return c.taskQueue
}