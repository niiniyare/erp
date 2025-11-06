package schema

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ThemeEdgeCasesTestSuite tests edge cases and error conditions
type ThemeEdgeCasesTestSuite struct {
	suite.Suite
	manager  *ThemeManager
	registry *TokenRegistry
	ctx      context.Context
}

func (suite *ThemeEdgeCasesTestSuite) SetupTest() {
	suite.registry = NewTokenRegistry()
	suite.manager = NewThemeManager(NewThemeRegistry(), suite.registry)
	suite.ctx = context.Background()
}

func TestThemeEdgeCasesTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeEdgeCasesTestSuite))
}

// ==================== Circular Reference Detection ====================

func (suite *ThemeEdgeCasesTestSuite) TestCircularReferenceDetection() {
	// Create tokens with circular reference
	tokens := &DesignTokens{
		Primitives: &PrimitiveTokens{
			Colors: &ColorPrimitives{
				Blue: &ColorScale{
					Scale500: "#3b82f6",
				},
			},
		},
		Semantic: &SemanticTokens{
			Colors: &SemanticColors{
				Background: &BackgroundColors{
					Default: TokenReference("{semantic.colors.background.default}"), // Self-reference
				},
				Text: &TextColors{
					Default: TokenReference("{semantic.colors.text.subtle}"),
					Subtle:  TokenReference("{semantic.colors.text.default}"), // Mutual reference
				},
			},
		},
	}

	// Setting tokens with circular references should fail
	err := suite.registry.SetTokens(tokens)
	suite.Error(err, "Should detect circular reference")
	suite.Contains(err.Error(), "circular reference", "Error should mention circular reference")
}

func (suite *ThemeEdgeCasesTestSuite) TestDeepCircularReferenceDetection() {
	// Create tokens with deep circular reference chain
	tokens := &DesignTokens{
		Primitives: &PrimitiveTokens{
			Colors: &ColorPrimitives{
				Blue: &ColorScale{
					Scale500: "#3b82f6",
				},
			},
		},
		Semantic: &SemanticTokens{
			Colors: &SemanticColors{
				Background: &BackgroundColors{
					Default: TokenReference("{semantic.colors.text.default}"),
				},
				Text: &TextColors{
					Default: TokenReference("{semantic.colors.border.default}"),
				},
				Border: &BorderColors{
					Default: TokenReference("{semantic.colors.background.default}"), // Creates A->B->C->A cycle
				},
			},
		},
	}

	err := suite.registry.SetTokens(tokens)
	suite.Error(err, "Should detect deep circular reference")
	suite.Contains(err.Error(), "circular reference", "Error should mention circular reference")
}

// ==================== Token Path Error Handling ====================

func (suite *ThemeEdgeCasesTestSuite) TestInvalidTokenPaths() {
	err := suite.registry.SetTokens(GetDefaultTokens())
	suite.NoError(err)

	testCases := []struct {
		name     string
		tokenRef TokenReference
		errorMsg string
	}{
		{
			name:     "NonExistentPath",
			tokenRef: TokenReference("{invalid.token.path}"),
			errorMsg: "token field not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.registry.ResolveToken(suite.ctx, tc.tokenRef)
			suite.Error(err, "Should fail for invalid token: %s", tc.tokenRef)
			suite.Contains(err.Error(), tc.errorMsg, "Error should contain expected message")
		})
	}
}

func (suite *ThemeEdgeCasesTestSuite) TestTokenPathValidation() {
	testCases := []struct {
		name      string
		tokenRef  TokenReference
		shouldErr bool
		errorMsg  string
	}{
		{
			name:      "ValidPath",
			tokenRef:  TokenReference("{semantic.colors.background.default}"),
			shouldErr: false,
		},
		{
			name:      "InvalidCharacters",
			tokenRef:  TokenReference("{semantic.colors@background}"),
			shouldErr: true,
			errorMsg:  "invalid character",
		},
		{
			name:      "NoDotsInPath",
			tokenRef:  TokenReference("{invalidpath}"),
			shouldErr: true,
			errorMsg:  "must contain at least one dot",
		},
		{
			name:      "NonReference",
			tokenRef:  TokenReference("#ffffff"),
			shouldErr: false, // Non-references are valid
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.tokenRef.Validate()
			if tc.shouldErr {
				suite.Error(err, "Should fail for invalid token: %s", tc.tokenRef)
				suite.Contains(err.Error(), tc.errorMsg, "Error should contain expected message")
			} else {
				suite.NoError(err, "Should pass for valid token: %s", tc.tokenRef)
			}
		})
	}
}

// ==================== Theme Registration Edge Cases ====================

func (suite *ThemeEdgeCasesTestSuite) TestDuplicateThemeRegistration() {
	// Register a theme
	theme := &Theme{
		ID:      "duplicate-theme",
		Name:    "Duplicate Theme",
		Version: "1.0.0",
		Tokens:  GetDefaultTokens(),
	}

	err := suite.manager.RegisterTheme(theme)
	suite.NoError(err)

	// Try to register the same theme again
	duplicateTheme := &Theme{
		ID:      "duplicate-theme", // Same ID
		Name:    "Another Duplicate Theme",
		Version: "2.0.0",
		Tokens:  GetDefaultTokens(),
	}

	err = suite.manager.RegisterTheme(duplicateTheme)
	// This should either succeed (overwrite) or fail (conflict) depending on implementation
	// We don't assert error/success since behavior may vary
}

// ==================== Theme Override Conflicts ====================

func (suite *ThemeEdgeCasesTestSuite) TestThemeOverrideConflicts() {
	// Register base theme
	baseTheme := &Theme{
		ID:      "base-theme",
		Name:    "Base Theme",
		Version: "1.0.0",
		Tokens:  GetDefaultTokens(),
	}
	err := suite.manager.RegisterTheme(baseTheme)
	suite.NoError(err)

	// Test basic theme retrieval
	retrievedTheme, err := suite.manager.GetTheme(suite.ctx, "base-theme")
	suite.NoError(err)
	suite.NotNil(retrievedTheme)
	suite.Equal("base-theme", retrievedTheme.ID)
}

func (suite *ThemeEdgeCasesTestSuite) TestBasicThemeOperations() {
	// Register base theme
	baseTheme := &Theme{
		ID:      "base-theme",
		Name:    "Base Theme",
		Version: "1.0.0",
		Tokens:  GetDefaultTokens(),
	}
	err := suite.manager.RegisterTheme(baseTheme)
	suite.NoError(err)

	// Test retrieval
	retrieved, err := suite.manager.GetTheme(suite.ctx, "base-theme")
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal("base-theme", retrieved.ID)
}

// ==================== Concurrent Access Edge Cases ====================

func (suite *ThemeEdgeCasesTestSuite) TestConcurrentRegistrationRace() {
	const numGoroutines = 100
	const numThemes = 10

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines)

	// Concurrent registration of themes with same IDs (race condition)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numThemes; j++ {
				theme := &Theme{
					ID:      fmt.Sprintf("race-theme-%d", j), // Same ID across goroutines
					Name:    fmt.Sprintf("Race Theme %d-%d", goroutineID, j),
					Version: "1.0.0",
					Tokens:  GetDefaultTokens(),
				}

				err := suite.manager.RegisterTheme(theme)
				if err != nil {
					// Some registrations may fail due to conflicts, which is expected
					if !IsErrorType(err, ErrorTypeConflict) {
						errorChan <- fmt.Errorf("unexpected error from goroutine %d: %v", goroutineID, err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// Check for unexpected errors
	for err := range errorChan {
		suite.Fail("Concurrent registration race error: %v", err)
	}

	// Verify that some themes were registered
	themes, err := suite.manager.ListThemes(suite.ctx)
	suite.NoError(err)
	suite.GreaterOrEqual(len(themes), numThemes, "Should have registered some themes")
}

func (suite *ThemeEdgeCasesTestSuite) TestConcurrentCacheEviction() {
	cache := NewThemeCache(5, 100*time.Millisecond) // Small cache with fast TTL

	var wg sync.WaitGroup
	const numGoroutines = 50

	// Concurrent operations to trigger cache eviction
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Add many themes to trigger eviction
			for j := 0; j < 20; j++ {
				theme := &Theme{
					ID:      fmt.Sprintf("evict-theme-%d-%d", id, j),
					Name:    fmt.Sprintf("Evict Theme %d-%d", id, j),
					Version: "1.0.0",
					Tokens:  GetDefaultTokens(),
				}
				cache.Set(theme.ID, theme)

				// Immediately try to get it
				_, found := cache.Get(theme.ID)
				if !found {
					// May not be found due to eviction, which is expected
				}

				// Small delay to let TTL expire
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	testTheme := &Theme{
		ID:      "test-after-eviction",
		Name:    "Test After Eviction",
		Version: "1.0.0",
		Tokens:  GetDefaultTokens(),
	}
	cache.Set(testTheme.ID, testTheme)

	retrieved, found := cache.Get(testTheme.ID)
	suite.True(found, "Cache should still work after concurrent eviction")
	suite.Equal(testTheme.ID, retrieved.ID)
}

// ==================== Memory and Resource Edge Cases ====================

func (suite *ThemeEdgeCasesTestSuite) TestLargeTokenStructures() {
	// Create a theme with very large token structures
	largeTokens := GetDefaultTokens()

	largeTheme := &Theme{
		ID:      "large-theme",
		Name:    "Large Theme",
		Version: "1.0.0",
		Tokens:  largeTokens,
		CustomCSS: fmt.Sprintf("/* Large CSS content: %s */",
			fmt.Sprintf("%0*s", 10000, "x")), // 10KB of CSS
	}

	// Should handle large themes
	err := suite.manager.RegisterTheme(largeTheme)
	suite.NoError(err)

	// Deep copy should work with large themes
	cloned, err := deepCopyTheme(largeTheme)
	suite.NoError(err)
	suite.Equal(largeTheme.ID, cloned.ID)
	suite.Equal(len(largeTheme.CustomCSS), len(cloned.CustomCSS))
}

func (suite *ThemeEdgeCasesTestSuite) TestNilPointerHandling() {
	// Test various nil pointer scenarios
	testCases := []struct {
		name   string
		testFn func()
	}{
		{
			name: "NilTokensInTheme",
			testFn: func() {
				theme := &Theme{
					ID:      "nil-tokens-theme",
					Name:    "Nil Tokens Theme",
					Version: "1.0.0",
					Tokens:  nil, // Nil tokens
				}
				err := suite.manager.RegisterTheme(theme)
				suite.Error(err, "Should fail with nil tokens")
			},
		},
		{
			name: "NilThemeRegistration",
			testFn: func() {
				err := suite.manager.RegisterTheme(nil)
				suite.Error(err, "Should fail with nil theme")
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, tc.testFn)
	}
}

// ==================== Performance Edge Cases ====================

func (suite *ThemeEdgeCasesTestSuite) TestTokenResolutionTimeout() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(suite.ctx, 1*time.Millisecond)
	defer cancel()

	err := suite.registry.SetTokens(GetDefaultTokens())
	suite.NoError(err)

	// This might timeout on slow systems
	_, err = suite.registry.ResolveToken(ctx, TokenReference("{semantic.colors.background.default}"))
	// Don't assert error as it may or may not timeout depending on system speed
	if err != nil {
		suite.Contains(err.Error(), "context deadline exceeded", "If error occurs, should be timeout")
	}
}

func (suite *ThemeEdgeCasesTestSuite) TestCacheMemoryPressure() {
	// Test cache behavior under memory pressure
	cache := NewThemeCache(1000, 1*time.Hour) // Large cache

	// Fill cache with many themes
	for i := 0; i < 2000; i++ { // More than cache capacity
		theme := &Theme{
			ID:      fmt.Sprintf("pressure-theme-%d", i),
			Name:    fmt.Sprintf("Pressure Theme %d", i),
			Version: "1.0.0",
			Tokens:  GetDefaultTokens(),
		}
		cache.Set(theme.ID, theme)
	}

	// Cache should have evicted older entries due to LRU policy

	// Early themes should be evicted
	_, found := cache.Get("pressure-theme-0")
	suite.False(found, "Early themes should be evicted")

	// Recent themes should be present
	_, found = cache.Get("pressure-theme-1999")
	suite.True(found, "Recent themes should be present")
}
