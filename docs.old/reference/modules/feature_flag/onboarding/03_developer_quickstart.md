# 3. Developer Quickstart: Using Feature Flags

This guide provides a hands-on walkthrough for developers to integrate and use the feature flag system in their application code.

## Objective

Your goal as a developer is typically to wrap a new piece of functionality in a conditional block that is controlled by a feature flag.

```go
// Psuedo-code example
if featureflags.IsEnabled("new-cool-feature") {
    // Show the new, cool feature
} else {
    // Show the old, stable feature
}
```

## Step 1: Creating a Feature Flag

Before you can use a flag in the code, it must be created in the system. While this can be done via the API, you will often do it through a service layer for testing or programmatic setup.

Here is how to use the `SimpleService` to create a basic boolean flag.

```go
// Assume 'flagService' is an initialized instance of featureflag.SimpleService
// and 'ctx' is a context.Context with the tenant_id already set.

import (
    "context"
    "log"
    "github.com/google/uuid"
    "awo.so/internal/core/featureflag"
)

func createMyNewFeatureFlag(ctx context.Context, flagService featureflag.SimpleService) {
    // Start with a 10% rollout to a small group of users
    rollout := int32(10)

    req := &featureflag.CreateFeatureFlagRequest{
        Name:              "enable-new-dashboard",
        Description:       "Activates the new V2 dashboard for users.",
        FlagType:          featureflag.FlagTypeBoolean,
        DefaultValue:      false, // The feature is OFF by default
        RolloutPercentage: &rollout,
        Metadata: map[string]interface{}{
            "owner":       "frontend-team",
            "jira_ticket": "DASH-123",
        },
    }

    newFlag, err := flagService.CreateFeatureFlag(ctx, req)
    if err != nil {
        // Handle error, perhaps the flag already exists
        log.Printf("Failed to create feature flag: %v", err)
        return
    }

    log.Printf("Successfully created flag '%s' (ID: %s)", newFlag.Name, newFlag.ID)
}
```

## Step 2: Evaluating a Feature Flag

Once a flag exists, you can evaluate it to decide which code path to execute. The evaluation takes into account the flag's default value, its enabled status, and the rollout percentage.

The `EvaluateFlag` method is the primary way to do this.

```go
// Assume 'flagService' is an initialized instance of featureflag.SimpleService
// and 'ctx' has the tenant_id.

func checkDashboardFeature(ctx context.Context, flagService featureflag.SimpleService, userID uuid.UUID) {
    // The EvaluationContext provides the necessary details for the evaluation engine
    // to make a decision. The UserID is critical for consistent percentage rollouts.
    evalCtx := &featureflag.EvaluationContext{
        TenantID: getTenantFromContext(ctx), // Helper to extract tenant ID
        UserID:   &userID,
        Attributes: map[string]string{
            "subscription_plan": "enterprise", // For future targeting rules
        },
    }

    result, err := flagService.EvaluateFlag(ctx, "enable-new-dashboard", evalCtx)
    if err != nil {
        log.Printf("Error evaluating flag, falling back to default behavior: %v", err)
        // Always handle errors gracefully and default to the "off" state
        renderOldDashboard()
        return
    }

    // The 'Enabled' field tells you if the feature is active for the user.
    if result.Enabled {
        log.Printf("User %s gets the new dashboard! Reason: %s", userID, result.Reason)
        renderNewDashboard()
    } else {
        log.Printf("User %s gets the old dashboard. Reason: %s", userID, result.Reason)
        renderOldDashboard()
    }
}
```

### The `EvaluationResult`

The `EvaluateFlag` method returns a rich `EvaluationResult` object, not just a boolean. This is important for observability and debugging.

-   `Value`: The raw value of the flag (`true`/`false`).
-   `Enabled`: The final decision. For a boolean flag, this is the same as `Value`.
-   `Reason`: Explains *why* this result was returned (e.g., `percentage_rollout`, `default_value`). This is invaluable for debugging.
-   `Metadata.CacheHit`: Tells you if the result came from the high-speed cache or the database, which is useful for performance analysis.

## Step 3: Bulk Evaluation

To avoid making multiple network calls when checking several flags on the same page, use the `EvaluateFlags` method.

```go
func checkMultipleFeatures(ctx context.Context, flagService featureflag.SimpleService, evalCtx *featureflag.EvaluationContext) {
    flagNames := []string{"enable-new-dashboard", "show-beta-feedback-button"}

    response, err := flagService.EvaluateFlags(ctx, flagNames, evalCtx)
    if err != nil {
        log.Printf("Error during bulk evaluation: %v", err)
        // Fallback to all features being off
        return
    }

    // Check the new dashboard
    if result, ok := response.Results["enable-new-dashboard"]; ok && result.Enabled {
        renderNewDashboard()
    } else {
        renderOldDashboard()
    }

    // Check the feedback button
    if result, ok := response.Results["show-beta-feedback-button"]; ok && result.Enabled {
        showFeedbackButton()
    }
}
```
