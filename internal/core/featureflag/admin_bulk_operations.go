package featureflag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
)

// BulkEnableFlags enables multiple feature flags
func (s *adminServiceImpl) BulkEnableFlags(ctx context.Context, request *BulkEnableFlagsRequest) (*BulkOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.bulk_enable_flags")
	defer span.End()

	startTime := time.Now()

	// Validate request
	if len(request.FlagNames) > 100 {
		return nil, fmt.Errorf("bulk operation limited to 100 flags per request")
	}

	result := &BulkOperationResult{
		TotalRequested: len(request.FlagNames),
		Results:        make([]BulkOperationItemResult, 0, len(request.FlagNames)),
		ExecutedAt:     startTime,
	}

	// Process flags concurrently with limited parallelism
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent operations
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorCategories := make(map[string]int)

	for _, flagName := range request.FlagNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			itemResult := s.processBulkEnableFlag(ctx, name)

			mu.Lock()
			result.Results = append(result.Results, itemResult)
			if itemResult.Success {
				result.Successful++
			} else {
				result.Failed++
				// Categorize error
				if itemResult.Error != "" {
					if contains(itemResult.Error, "not found") {
						errorCategories["not_found"]++
					} else if contains(itemResult.Error, "permission") {
						errorCategories["permission"]++
					} else {
						errorCategories["other"]++
					}
				}
			}
			mu.Unlock()
		}(flagName)
	}

	wg.Wait()
	result.ExecutionTime = time.Since(startTime)

	// Calculate summary
	result.Summary = BulkOperationSummary{
		Operation:       "bulk_enable",
		SuccessRate:     float64(result.Successful) / float64(result.TotalRequested) * 100,
		AverageTime:     float64(result.ExecutionTime.Milliseconds()) / float64(result.TotalRequested),
		ErrorCategories: errorCategories,
	}

	// Audit the bulk operation
	s.auditBulkOperation(ctx, "bulk_enable_flags", request.Reason, result)

	s.logger.Info("Bulk enable flags completed", logger.Fields{
		"total":          result.TotalRequested,
		"successful":     result.Successful,
		"failed":         result.Failed,
		"execution_time": result.ExecutionTime,
		"reason":         request.Reason,
	})

	return result, nil
}

// BulkDisableFlags disables multiple feature flags
func (s *adminServiceImpl) BulkDisableFlags(ctx context.Context, request *BulkDisableFlagsRequest) (*BulkOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.bulk_disable_flags")
	defer span.End()

	startTime := time.Now()

	// Validate request
	if len(request.FlagNames) > 100 {
		return nil, fmt.Errorf("bulk operation limited to 100 flags per request")
	}

	result := &BulkOperationResult{
		TotalRequested: len(request.FlagNames),
		Results:        make([]BulkOperationItemResult, 0, len(request.FlagNames)),
		ExecutedAt:     startTime,
	}

	// Process flags concurrently
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorCategories := make(map[string]int)

	for _, flagName := range request.FlagNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			itemResult := s.processBulkDisableFlag(ctx, name)

			mu.Lock()
			result.Results = append(result.Results, itemResult)
			if itemResult.Success {
				result.Successful++
			} else {
				result.Failed++
				if itemResult.Error != "" {
					if contains(itemResult.Error, "not found") {
						errorCategories["not_found"]++
					} else if contains(itemResult.Error, "permission") {
						errorCategories["permission"]++
					} else {
						errorCategories["other"]++
					}
				}
			}
			mu.Unlock()
		}(flagName)
	}

	wg.Wait()
	result.ExecutionTime = time.Since(startTime)

	result.Summary = BulkOperationSummary{
		Operation:       "bulk_disable",
		SuccessRate:     float64(result.Successful) / float64(result.TotalRequested) * 100,
		AverageTime:     float64(result.ExecutionTime.Milliseconds()) / float64(result.TotalRequested),
		ErrorCategories: errorCategories,
	}

	s.auditBulkOperation(ctx, "bulk_disable_flags", request.Reason, result)

	return result, nil
}

// BulkDeleteFlags deletes multiple feature flags (destructive operation)
func (s *adminServiceImpl) BulkDeleteFlags(ctx context.Context, request *BulkDeleteFlagsRequest) (*BulkOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.bulk_delete_flags")
	defer span.End()

	startTime := time.Now()

	// Stricter validation for destructive operations
	if len(request.FlagIDs) > 50 {
		return nil, fmt.Errorf("bulk delete limited to 50 flags per request for safety")
	}

	result := &BulkOperationResult{
		TotalRequested: len(request.FlagIDs),
		Results:        make([]BulkOperationItemResult, 0, len(request.FlagIDs)),
		ExecutedAt:     startTime,
	}

	// Process deletions with more careful error handling
	semaphore := make(chan struct{}, 5) // Lower concurrency for destructive ops
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorCategories := make(map[string]int)

	for _, flagID := range request.FlagIDs {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			itemResult := s.processBulkDeleteFlag(ctx, id)

			mu.Lock()
			result.Results = append(result.Results, itemResult)
			if itemResult.Success {
				result.Successful++
			} else {
				result.Failed++
				if itemResult.Error != "" {
					if contains(itemResult.Error, "not found") {
						errorCategories["not_found"]++
					} else if contains(itemResult.Error, "permission") {
						errorCategories["permission"]++
					} else if contains(itemResult.Error, "dependency") {
						errorCategories["dependency"]++
					} else {
						errorCategories["other"]++
					}
				}
			}
			mu.Unlock()
		}(flagID)
	}

	wg.Wait()
	result.ExecutionTime = time.Since(startTime)

	result.Summary = BulkOperationSummary{
		Operation:       "bulk_delete",
		SuccessRate:     float64(result.Successful) / float64(result.TotalRequested) * 100,
		AverageTime:     float64(result.ExecutionTime.Milliseconds()) / float64(result.TotalRequested),
		ErrorCategories: errorCategories,
	}

	// High-severity audit for bulk deletions
	s.auditBulkOperation(ctx, "bulk_delete_flags", request.Reason, result)

	return result, nil
}

// BulkUpdateRollout updates rollout percentages for multiple flags
func (s *adminServiceImpl) BulkUpdateRollout(ctx context.Context, request *BulkUpdateRolloutRequest) (*BulkOperationResult, error) {
	ctx, span := s.tracing.StartSpan(ctx, "admin.bulk_update_rollout")
	defer span.End()

	startTime := time.Now()

	if len(request.Updates) > 100 {
		return nil, fmt.Errorf("bulk rollout update limited to 100 flags per request")
	}

	result := &BulkOperationResult{
		TotalRequested: len(request.Updates),
		Results:        make([]BulkOperationItemResult, 0, len(request.Updates)),
		ExecutedAt:     startTime,
	}

	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorCategories := make(map[string]int)

	for _, update := range request.Updates {
		wg.Add(1)
		go func(upd RolloutUpdate) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			itemResult := s.processBulkRolloutUpdate(ctx, upd)

			mu.Lock()
			result.Results = append(result.Results, itemResult)
			if itemResult.Success {
				result.Successful++
			} else {
				result.Failed++
				if itemResult.Error != "" {
					if contains(itemResult.Error, "not found") {
						errorCategories["not_found"]++
					} else if contains(itemResult.Error, "validation") {
						errorCategories["validation"]++
					} else {
						errorCategories["other"]++
					}
				}
			}
			mu.Unlock()
		}(update)
	}

	wg.Wait()
	result.ExecutionTime = time.Since(startTime)

	result.Summary = BulkOperationSummary{
		Operation:       "bulk_update_rollout",
		SuccessRate:     float64(result.Successful) / float64(result.TotalRequested) * 100,
		AverageTime:     float64(result.ExecutionTime.Milliseconds()) / float64(result.TotalRequested),
		ErrorCategories: errorCategories,
	}

	s.auditBulkOperation(ctx, "bulk_update_rollout", request.Reason, result)

	return result, nil
}

// Helper methods for processing individual items in bulk operations

func (s *adminServiceImpl) processBulkEnableFlag(ctx context.Context, flagName string) BulkOperationItemResult {
	// Get the flag first
	flag, err := s.baseService.GetFeatureFlag(ctx, flagName)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: flagName,
			Success:    false,
			Error:      err.Error(),
		}
	}

	// Update to enable (set default value to true for boolean flags)
	defaultValue := true
	updateRequest := &UpdateFeatureFlagRequest{
		DefaultValue: &defaultValue,
	}

	_, err = s.baseService.UpdateFeatureFlag(ctx, flag.ID, updateRequest)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: flagName,
			Success:    false,
			Error:      err.Error(),
		}
	}

	return BulkOperationItemResult{
		Identifier: flagName,
		Success:    true,
	}
}

func (s *adminServiceImpl) processBulkDisableFlag(ctx context.Context, flagName string) BulkOperationItemResult {
	flag, err := s.baseService.GetFeatureFlag(ctx, flagName)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: flagName,
			Success:    false,
			Error:      err.Error(),
		}
	}

	defaultValue := false
	updateRequest := &UpdateFeatureFlagRequest{
		DefaultValue: &defaultValue,
	}

	_, err = s.baseService.UpdateFeatureFlag(ctx, flag.ID, updateRequest)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: flagName,
			Success:    false,
			Error:      err.Error(),
		}
	}

	return BulkOperationItemResult{
		Identifier: flagName,
		Success:    true,
	}
}

func (s *adminServiceImpl) processBulkDeleteFlag(ctx context.Context, flagID uuid.UUID) BulkOperationItemResult {
	err := s.baseService.DeleteFeatureFlag(ctx, flagID)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: flagID.String(),
			Success:    false,
			Error:      err.Error(),
		}
	}

	return BulkOperationItemResult{
		Identifier: flagID.String(),
		Success:    true,
	}
}

func (s *adminServiceImpl) processBulkRolloutUpdate(ctx context.Context, update RolloutUpdate) BulkOperationItemResult {
	// Validate percentage
	if update.Percentage < 0 || update.Percentage > 100 {
		return BulkOperationItemResult{
			Identifier: update.FlagName,
			Success:    false,
			Error:      fmt.Sprintf("invalid percentage: %d (must be 0-100)", update.Percentage),
		}
	}

	flag, err := s.baseService.GetFeatureFlag(ctx, update.FlagName)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: update.FlagName,
			Success:    false,
			Error:      err.Error(),
		}
	}

	updateRequest := &UpdateFeatureFlagRequest{
		RolloutPercentage: &update.Percentage,
	}

	_, err = s.baseService.UpdateFeatureFlag(ctx, flag.ID, updateRequest)
	if err != nil {
		return BulkOperationItemResult{
			Identifier: update.FlagName,
			Success:    false,
			Error:      err.Error(),
		}
	}

	return BulkOperationItemResult{
		Identifier: update.FlagName,
		Success:    true,
	}
}

// auditBulkOperation logs bulk operations for compliance and monitoring
func (s *adminServiceImpl) auditBulkOperation(ctx context.Context, operation, reason string, result *BulkOperationResult) {
	contextData, _ := json.Marshal(map[string]any{
		"operation":         operation,
		"reason":            reason,
		"total_requested":   result.TotalRequested,
		"successful":        result.Successful,
		"failed":            result.Failed,
		"execution_time_ms": result.ExecutionTime.Milliseconds(),
		"success_rate":      result.Summary.SuccessRate,
		"error_categories":  result.Summary.ErrorCategories,
		"executed_at":       result.ExecutedAt,
	})

	severity := AuditSeverityWarn
	if operation == "bulk_delete_flags" {
		severity = AuditSeverityHigh // Destructive operations are high severity
	}
	if result.Summary.SuccessRate < 50 {
		severity = AuditSeverityHigh // Low success rate is concerning
	}

	decision := fmt.Sprintf("EXECUTED_%s", operation)
	reasonMsg := fmt.Sprintf("Admin bulk operation: %s (%d/%d successful)", operation, result.Successful, result.TotalRequested)
	s.auditService.CreateAuditEvent(ctx, audit.CreateAuditEventRequest{
		EventType:     "admin_bulk_operation",
		EventCategory: AuditCategoryAdmin,
		Severity:      severity,
		Decision:      &decision,
		Reason:        &reasonMsg,
		Context:       contextData,
	})

	// Also record metrics
	if s.metrics != nil {
		labels := metrics.Fields{
			"operation": operation,
		}

		s.metrics.IncrementCounter("admin_bulk_operations_total", labels)
		s.metrics.ObserveHistogram("admin_bulk_operation_duration", float64(result.ExecutionTime.Milliseconds()), labels)
		s.metrics.SetGauge("admin_bulk_operation_success_rate", result.Summary.SuccessRate, labels)
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}
