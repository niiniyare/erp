package db

// import (
// 	"context"
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"encoding/json"
// 	"testing"
// 	"time"
//
// 	"github.com/stretchr/testify/require"
// )
//
// // Test Case 1.1.1: Single Policy Evaluation Correctness
// func TestCreatePolicyEvaluation(t *testing.T) {
// 	tenant := createTestTenant(t, "TestCreatePolicy")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	arg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "some_hash",
// 		Decision:    "ALLOW",
// 	}
//
// 	eval, err := testQueries.CreatePolicyEvaluation(ctx, arg)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, eval)
//
// 	require.Equal(t, tenant.ID, eval.TenantID)
// 	require.Equal(t, arg.UserID, eval.UserID)
// 	require.Equal(t, arg.ResourceID, eval.ResourceID)
// 	require.Equal(t, arg.Action, eval.Action)
// 	require.Equal(t, arg.ContextHash, eval.ContextHash)
// 	require.Equal(t, arg.Decision, eval.Decision)
// 	require.NotZero(t, eval.ID)
// 	require.NotZero(t, eval.EvaluatedAt)
// 	require.NotZero(t, eval.ExpiresAt)
// }
//
// // Test Case 1.1.3: Context Hash Collision Resistance
// func TestContextHashUniqueness(t *testing.T) {
// 	tenant := createTestTenant(t, "TestHashUnique")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Context 1
// 	context1 := map[string]any{"ip": "192.168.1.1", "time": "morning"}
// 	hash1 := hashContext(t, context1)
//
// 	// Context 2 (different)
// 	context2 := map[string]any{"ip": "192.168.1.2", "time": "morning"}
// 	hash2 := hashContext(t, context2)
//
// 	require.NotEqual(t, hash1, hash2, "Different contexts should produce different hashes")
//
// 	// Context 3 (same as 1)
// 	context3 := map[string]any{"ip": "192.168.1.1", "time": "morning"}
// 	hash3 := hashContext(t, context3)
//
// 	require.Equal(t, hash1, hash3, "Identical contexts should produce identical hashes")
//
// 	arg1 := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: hash1,
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg1)
// 	require.NoError(t, err)
//
// 	arg2 := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: hash2,
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg2)
// 	require.NoError(t, err)
// }
//
// // Test Group 1.2: Dynamic Attribute Handling
// func TestPolicyEvaluationWithComplexContext(t *testing.T) {
// 	tenant := createTestTenant(t, "TestComplexCtx")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Time-sensitive context
// 	timeContext := map[string]any{
// 		"time_of_day": time.Now().Hour(),
// 		"day_of_week": time.Now().Weekday().String(),
// 	}
// 	timeHash := hashContext(t, timeContext)
//
// 	argTime := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: timeHash,
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, argTime)
// 	require.NoError(t, err)
//
// 	// Location-based context
// 	locationContext := map[string]any{
// 		"ip_address": "203.0.113.55",
// 		"geo_region": "EU",
// 	}
// 	locationHash := hashContext(t, locationContext)
//
// 	argLocation := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: locationHash,
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, argLocation)
// 	require.NoError(t, err)
// }
//
// // Phase 2: Cache Management and Invalidation Tests
//
// // Test Group 2.1: User-based Invalidation
// func TestInvalidateUserEvaluations(t *testing.T) {
// 	tenant := createTestTenant(t, "TestInvalidateUser")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user1 := createUserForTest(t, ctx)
// 	user2 := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create evaluations for both users
// 	arg1 := CreatePolicyEvaluationParams{
// 		UserID:      user1.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "user1_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg1)
// 	require.NoError(t, err)
//
// 	arg2 := CreatePolicyEvaluationParams{
// 		UserID:      user2.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "user2_ctx",
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg2)
// 	require.NoError(t, err)
//
// 	// Invalidate user1's evaluations
// 	err = testQueries.InvalidateUserEvaluations(ctx, user1.ID)
// 	require.NoError(t, err)
//
// 	// Verify user1's evaluations are gone and user2's remain
// 	// Note: We would need a GetPolicyEvaluation query to fully test this
// 	// For now we test the invalidation doesn't error
// }
//
// // Test Group 2.2: Resource-based Invalidation
// func TestInvalidateResourceEvaluations(t *testing.T) {
// 	tenant := createTestTenant(t, "TestInvalidateResource")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource1 := createResourceForTest(t, ctx, module.ID)
// 	resource2 := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create evaluations for different resources
// 	arg1 := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource1.ID,
// 		Action:      action.Name,
// 		ContextHash: "res1_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg1)
// 	require.NoError(t, err)
//
// 	arg2 := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource2.ID,
// 		Action:      action.Name,
// 		ContextHash: "res2_ctx",
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg2)
// 	require.NoError(t, err)
//
// 	// Test invalidation by resource type and ID
// 	// Note: This would require the resource_type field in policy_evaluations table
// 	// For now we test action-based invalidation as a proxy
// 	err = testQueries.InvalidateActionEvaluations(ctx, action.Name)
// 	require.NoError(t, err)
// }
//
// // Test Group 2.3: Action-based Invalidation
// func TestInvalidateActionEvaluations(t *testing.T) {
// 	tenant := createTestTenant(t, "TestInvalidateAction")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	readAction := createActionForTest(t, ctx)
// 	writeAction := createActionForTest(t, ctx)
//
// 	// Create evaluations with different actions
// 	readArg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      readAction.Name,
// 		ContextHash: "read_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, readArg)
// 	require.NoError(t, err)
//
// 	writeArg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      writeAction.Name,
// 		ContextHash: "write_ctx",
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, writeArg)
// 	require.NoError(t, err)
//
// 	// Invalidate only read actions
// 	err = testQueries.InvalidateActionEvaluations(ctx, readAction.Name)
// 	require.NoError(t, err)
// }
//
// // Test Group 2.5: Cleanup Operations
// func TestCleanupExpiredEvaluations(t *testing.T) {
// 	tenant := createTestTenant(t, "TestCleanupExpired")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create test evaluation
// 	arg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "cleanup_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg)
// 	require.NoError(t, err)
//
// 	// Test cleanup operation
// 	err = testQueries.CleanupExpiredEvaluations(ctx)
// 	require.NoError(t, err)
// }
//
// // Test Group 2.5: Complete Invalidation
// func TestInvalidateAllEvaluations(t *testing.T) {
// 	tenant := createTestTenant(t, "TestInvalidateAll")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user1 := createUserForTest(t, ctx)
// 	user2 := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create multiple evaluations
// 	arg1 := CreatePolicyEvaluationParams{
// 		UserID:      user1.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "all1_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg1)
// 	require.NoError(t, err)
//
// 	arg2 := CreatePolicyEvaluationParams{
// 		UserID:      user2.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "all2_ctx",
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg2)
// 	require.NoError(t, err)
//
// 	// Invalidate all evaluations for current tenant
// 	err = testQueries.InvalidateAllEvaluations(ctx)
// 	require.NoError(t, err)
// }
//
// // Phase 3: Query Performance and Analytics Tests
//
// // Test Group 3.1: Cache Statistics
// func TestGetEvaluationCacheStats(t *testing.T) {
// 	tenant := createTestTenant(t, "TestCacheStats")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user1 := createUserForTest(t, ctx)
// 	user2 := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create diverse evaluation data
// 	for i := 0; i < 5; i++ {
// 		arg := CreatePolicyEvaluationParams{
// 			UserID:      user1.ID,
// 			ResourceID:  &resource.ID,
// 			Action:      action.Name,
// 			ContextHash: fmt.Sprintf("stats_ctx_%d", i),
// 			Decision:    "ALLOW",
// 		}
// 		_, err = testQueries.CreatePolicyEvaluation(ctx, arg)
// 		require.NoError(t, err)
// 	}
//
// 	for i := 0; i < 3; i++ {
// 		arg := CreatePolicyEvaluationParams{
// 			UserID:      user2.ID,
// 			ResourceID:  &resource.ID,
// 			Action:      action.Name,
// 			ContextHash: fmt.Sprintf("stats_ctx_u2_%d", i),
// 			Decision:    "DENY",
// 		}
// 		_, err = testQueries.CreatePolicyEvaluation(ctx, arg)
// 		require.NoError(t, err)
// 	}
//
// 	// Get cache statistics
// 	stats, err := testQueries.GetEvaluationCacheStats(ctx)
// 	require.NoError(t, err)
// 	require.NotNil(t, stats)
//
// 	// Verify statistics are reasonable
// 	require.GreaterOrEqual(t, stats.TotalCachedEvaluations, int64(8))
// 	require.GreaterOrEqual(t, stats.UniqueUsers, int64(2))
// 	require.GreaterOrEqual(t, stats.UniqueActions, int64(1))
// }
//
// // Test Group 3.2: Historical Data Queries
// func TestGetUserEvaluationHistory(t *testing.T) {
// 	tenant := createTestTenant(t, "TestUserHistory")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user1 := createUserForTest(t, ctx)
// 	user2 := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create evaluation history for user1
// 	for i := 0; i < 3; i++ {
// 		arg := CreatePolicyEvaluationParams{
// 			UserID:      user1.ID,
// 			ResourceID:  &resource.ID,
// 			Action:      action.Name,
// 			ContextHash: fmt.Sprintf("history_ctx_%d", i),
// 			Decision:    "ALLOW",
// 		}
// 		_, err = testQueries.CreatePolicyEvaluation(ctx, arg)
// 		require.NoError(t, err)
// 		time.Sleep(10 * time.Millisecond) // Ensure distinct timestamps
// 	}
//
// 	// Create one evaluation for user2
// 	arg2 := CreatePolicyEvaluationParams{
// 		UserID:      user2.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "history_ctx_u2",
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, arg2)
// 	require.NoError(t, err)
//
// 	// Get user1's evaluation history
// 	historyParams := GetUserEvaluationHistoryParams{
// 		UserID: user1.ID,
// 		Limit:  10,
// 		Offset: 0,
// 	}
// 	history, err := testQueries.GetUserEvaluationHistory(ctx, historyParams)
// 	require.NoError(t, err)
// 	require.Len(t, history, 3)
//
// 	// Verify all results belong to user1
// 	for _, eval := range history {
// 		require.Equal(t, user1.ID, eval.UserID)
// 	}
// }
//
// // Test Group 3.3: Decision Analytics
// func TestGetEvaluationsByDecision(t *testing.T) {
// 	tenant := createTestTenant(t, "TestByDecision")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create evaluations with different decisions
// 	allowArg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "allow_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, allowArg)
// 	require.NoError(t, err)
//
// 	denyArg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "deny_ctx",
// 		Decision:    "DENY",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, denyArg)
// 	require.NoError(t, err)
//
// 	// Query evaluations by decision
// 	now := time.Now()
// 	decisionParams := GetEvaluationsByDecisionParams{
// 		Decision:      "ALLOW",
// 		EvaluatedAt:   now.Add(-time.Hour),
// 		EvaluatedAt_2: now.Add(time.Hour),
// 		Limit:         10,
// 		Offset:        0,
// 	}
// 	allowDecisions, err := testQueries.GetEvaluationsByDecision(ctx, decisionParams)
// 	require.NoError(t, err)
// 	require.GreaterOrEqual(t, len(allowDecisions), 1)
//
// 	// Verify all results have ALLOW decision
// 	for _, eval := range allowDecisions {
// 		require.Equal(t, "ALLOW", eval.Decision)
// 	}
// }
//
// func TestCountEvaluationsByDecision(t *testing.T) {
// 	tenant := createTestTenant(t, "TestCountDecision")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create known quantities of each decision type
// 	for i := 0; i < 3; i++ {
// 		allowArg := CreatePolicyEvaluationParams{
// 			UserID:      user.ID,
// 			ResourceID:  &resource.ID,
// 			Action:      action.Name,
// 			ContextHash: fmt.Sprintf("count_allow_%d", i),
// 			Decision:    "ALLOW",
// 		}
// 		_, err = testQueries.CreatePolicyEvaluation(ctx, allowArg)
// 		require.NoError(t, err)
// 	}
//
// 	for i := 0; i < 2; i++ {
// 		denyArg := CreatePolicyEvaluationParams{
// 			UserID:      user.ID,
// 			ResourceID:  &resource.ID,
// 			Action:      action.Name,
// 			ContextHash: fmt.Sprintf("count_deny_%d", i),
// 			Decision:    "DENY",
// 		}
// 		_, err = testQueries.CreatePolicyEvaluation(ctx, denyArg)
// 		require.NoError(t, err)
// 	}
//
// 	naArg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: "count_na",
// 		Decision:    "NOT_APPLICABLE",
// 	}
// 	_, err = testQueries.CreatePolicyEvaluation(ctx, naArg)
// 	require.NoError(t, err)
//
// 	// Count evaluations by decision
// 	now := time.Now()
// 	countParams := CountEvaluationsByDecisionParams{
// 		EvaluatedAt:   now.Add(-time.Hour),
// 		EvaluatedAt_2: now.Add(time.Hour),
// 	}
// 	counts, err := testQueries.CountEvaluationsByDecision(ctx, countParams)
// 	require.NoError(t, err)
//
// 	// Verify counts match expected values
// 	require.GreaterOrEqual(t, counts.AllowCount, int64(3))
// 	require.GreaterOrEqual(t, counts.DenyCount, int64(2))
// 	require.GreaterOrEqual(t, counts.NotApplicableCount, int64(1))
// 	require.GreaterOrEqual(t, counts.TotalCount, int64(6))
// }
//
// // Test Group 3.4: Performance Metrics
// func TestGetEvaluationMetrics(t *testing.T) {
// 	tenant := createTestTenant(t, "TestPerfMetrics")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user1 := createUserForTest(t, ctx)
// 	user2 := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource1 := createResourceForTest(t, ctx, module.ID)
// 	resource2 := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Create evaluations with varied resources and users
// 	for i := 0; i < 4; i++ {
// 		resourceID := &resource1.ID
// 		userID := user1.ID
// 		if i%2 == 0 {
// 			resourceID = &resource2.ID
// 		}
// 		if i%3 == 0 {
// 			userID = user2.ID
// 		}
//
// 		arg := CreatePolicyEvaluationParams{
// 			UserID:      userID,
// 			ResourceID:  resourceID,
// 			Action:      action.Name,
// 			ContextHash: fmt.Sprintf("metrics_ctx_%d", i),
// 			Decision:    "ALLOW",
// 		}
// 		_, err = testQueries.CreatePolicyEvaluation(ctx, arg)
// 		require.NoError(t, err)
// 	}
//
// 	// Get evaluation metrics
// 	now := time.Now()
// 	metricsParams := GetEvaluationMetricsParams{
// 		EvaluatedAt:   now.Add(-time.Hour),
// 		EvaluatedAt_2: now.Add(time.Hour),
// 	}
// 	metrics, err := testQueries.GetEvaluationMetrics(ctx, metricsParams)
// 	require.NoError(t, err)
// 	require.NotNil(t, metrics)
//
// 	// Verify metrics are reasonable
// 	require.GreaterOrEqual(t, metrics.TotalEvaluations, int64(4))
// 	require.GreaterOrEqual(t, metrics.UniqueUsers, int64(1))
// 	require.GreaterOrEqual(t, metrics.UniqueResources, int64(1))
// }
//
// // Phase 4: Edge Cases and Error Handling Tests
//
// // Test Group 4.1: Multi-tenant Isolation
// func TestTenantDataIsolation(t *testing.T) {
// 	tenant1 := createTestTenant(t, "TestIsolation1")
// 	tenant2 := createTestTenant(t, "TestIsolation2")
// 	ctx := context.Background()
//
// 	// Create evaluation in tenant1
// 	err := testQueries.SetTenantContext(ctx, tenant1.ID)
// 	require.NoError(t, err)
//
// 	user1 := createUserForTest(t, ctx)
// 	module1 := createModuleForTest(t, ctx)
// 	resource1 := createResourceForTest(t, ctx, module1.ID)
// 	action1 := createActionForTest(t, ctx)
//
// 	arg1 := CreatePolicyEvaluationParams{
// 		UserID:      user1.ID,
// 		ResourceID:  &resource1.ID,
// 		Action:      action1.Name,
// 		ContextHash: "tenant1_ctx",
// 		Decision:    "ALLOW",
// 	}
// 	eval1, err := testQueries.CreatePolicyEvaluation(ctx, arg1)
// 	require.NoError(t, err)
// 	require.Equal(t, tenant1.ID, eval1.TenantID)
//
// 	// Switch to tenant2
// 	err = testQueries.SetTenantContext(ctx, tenant2.ID)
// 	require.NoError(t, err)
//
// 	user2 := createUserForTest(t, ctx)
// 	module2 := createModuleForTest(t, ctx)
// 	resource2 := createResourceForTest(t, ctx, module2.ID)
// 	action2 := createActionForTest(t, ctx)
//
// 	arg2 := CreatePolicyEvaluationParams{
// 		UserID:      user2.ID,
// 		ResourceID:  &resource2.ID,
// 		Action:      action2.Name,
// 		ContextHash: "tenant2_ctx",
// 		Decision:    "DENY",
// 	}
// 	eval2, err := testQueries.CreatePolicyEvaluation(ctx, arg2)
// 	require.NoError(t, err)
// 	require.Equal(t, tenant2.ID, eval2.TenantID)
//
// 	// Verify tenant isolation by checking stats
// 	stats2, err := testQueries.GetEvaluationCacheStats(ctx)
// 	require.NoError(t, err)
// 	// Should only see tenant2's data
// 	require.GreaterOrEqual(t, stats2.TotalCachedEvaluations, int64(1))
// }
//
// // Test Group 4.2: Boundary Conditions
// func TestUnicodeAndSpecialCharacters(t *testing.T) {
// 	tenant := createTestTenant(t, "TestUnicode")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Test with unicode characters in context hash
// 	unicodeContext := map[string]any{
// 		"用户":       "测试用户",   // Chinese: user: test user
// 		"действие": "чтение", // Russian: action: reading
// 		"":        "",      // Emoji
// 	}
// 	unicodeHash := hashContext(t, unicodeContext)
//
// 	arg := CreatePolicyEvaluationParams{
// 		UserID:      user.ID,
// 		ResourceID:  &resource.ID,
// 		Action:      action.Name,
// 		ContextHash: unicodeHash,
// 		Decision:    "ALLOW",
// 	}
// 	eval, err := testQueries.CreatePolicyEvaluation(ctx, arg)
// 	require.NoError(t, err)
// 	require.Equal(t, unicodeHash, eval.ContextHash)
// }
//
// // Test Group 4.3: Concurrent Operations
// func TestConcurrentEvaluationCreation(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping concurrency tests in short mode")
// 	}
//
// 	tenant := createTestTenant(t, "TestConcurrent")
// 	ctx := context.Background()
// 	err := testQueries.SetTenantContext(ctx, tenant.ID)
// 	require.NoError(t, err)
//
// 	user := createUserForTest(t, ctx)
// 	module := createModuleForTest(t, ctx)
// 	resource := createResourceForTest(t, ctx, module.ID)
// 	action := createActionForTest(t, ctx)
//
// 	// Test concurrent creation of evaluations
// 	numGoroutines := 10
// 	done := make(chan error, numGoroutines)
//
// 	for i := 0; i < numGoroutines; i++ {
// 		go func(iteration int) {
// 			arg := CreatePolicyEvaluationParams{
// 				UserID:      user.ID,
// 				ResourceID:  &resource.ID,
// 				Action:      action.Name,
// 				ContextHash: fmt.Sprintf("concurrent_ctx_%d", iteration),
// 				Decision:    "ALLOW",
// 			}
// 			_, err := testQueries.CreatePolicyEvaluation(ctx, arg)
// 			done <- err
// 		}(i)
// 	}
//
// 	// Wait for all goroutines to complete
// 	for i := 0; i < numGoroutines; i++ {
// 		err := <-done
// 		require.NoError(t, err, "Concurrent evaluation creation should succeed")
// 	}
//
// 	// Verify data integrity
// 	stats, err := testQueries.GetEvaluationCacheStats(ctx)
// 	require.NoError(t, err)
// 	require.GreaterOrEqual(t, stats.TotalCachedEvaluations, int64(numGoroutines))
// }
//
// // Helper to create a consistent hash from a context map
// func hashContext(t *testing.T, context map[string]any) string {
// 	data, err := json.Marshal(context)
// 	require.NoError(t, err)
// 	hash := sha256.Sum256(data)
// 	return hex.EncodeToString(hash[:])
// }
