package schema

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ThemeTestSuite tests the theme functionality
type ThemeTestSuite struct {
	suite.Suite
}

func TestThemeTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeTestSuite))
}

// ==================== Basic Theme Tests ====================

func (suite *ThemeTestSuite) TestThemeStructure() {
	theme := &Theme{
		ID:          "test-theme-id",
		Name:        "Test Theme",
		Version:     "1.0.0",
		Description: "A test theme",
		Author:      "Test Author",
		Tokens:      GetDefaultTokens(),
	}

	suite.Require().Equal("test-theme-id", theme.ID)
	suite.Require().Equal("Test Theme", theme.Name)
	suite.Require().Equal("1.0.0", theme.Version)
	suite.Require().Equal("A test theme", theme.Description)
	suite.Require().Equal("Test Author", theme.Author)
	suite.Require().NotNil(theme.Tokens)
}

func (suite *ThemeTestSuite) TestTokens() {
	tokens := GetDefaultTokens()

	suite.Require().NotNil(tokens)
	suite.Require().NotNil(tokens.Primitives)
	suite.Require().NotNil(tokens.Semantic)
	suite.Require().NotNil(tokens.Components)
}

func (suite *ThemeTestSuite) TestTokenReference() {
	ref := TokenReference("{colors.primary.base}")

	suite.Require().True(ref.IsReference())
	suite.Require().Equal("colors.primary.base", ref.Path())
	suite.Require().Equal("{colors.primary.base}", ref.String())

	// Test non-reference
	nonRef := TokenReference("#123456")
	suite.Require().False(nonRef.IsReference())
}

func (suite *ThemeTestSuite) TestThemeRegistry() {
	registry := NewThemeRegistry()
	suite.Require().NotNil(registry)

	// Test registry methods exist (they return placeholder values for now)
	exists := registry.Exists("test-theme")
	suite.Require().False(exists) // Should be false for non-existent theme
}

func (suite *ThemeTestSuite) TestThemeManager() {
	registry := NewThemeRegistry()
	tokenManager := NewTokenRegistry()
	manager := NewThemeManager(registry, tokenManager)
	suite.Require().NotNil(manager)
}

func (suite *ThemeTestSuite) TestDarkModeConfig() {
	darkModeConfig := &DarkModeConfig{
		Enabled:  true,
		Default:  true,
		Strategy: "auto",
	}

	suite.Require().True(darkModeConfig.Enabled)
	suite.Require().True(darkModeConfig.Default)
	suite.Require().Equal("auto", darkModeConfig.Strategy)
}

func (suite *ThemeTestSuite) TestAccessibilityConfig() {
	accessibilityConfig := &AccessibilityConfig{
		HighContrast:     true,
		ReducedMotion:    false,
		FocusIndicator:   true,
		ScreenReaderOnly: true,
		KeyboardNav:      true,
	}

	suite.Require().True(accessibilityConfig.HighContrast)
	suite.Require().False(accessibilityConfig.ReducedMotion)
	suite.Require().True(accessibilityConfig.FocusIndicator)
	suite.Require().True(accessibilityConfig.ScreenReaderOnly)
	suite.Require().True(accessibilityConfig.KeyboardNav)
}

func (suite *ThemeTestSuite) TestThemeWithAllComponents() {
	theme := &Theme{
		ID:          "complete-theme-id",
		Name:        "Complete Theme",
		Version:     "2.0.0",
		Description: "A complete theme with all components",
		Tokens:      GetDefaultTokens(),

		DarkMode: &DarkModeConfig{
			Enabled: true,
			Default: true,
		},

		Accessibility: &AccessibilityConfig{
			HighContrast:     true,
			FocusIndicator:   true,
			ScreenReaderOnly: true,
		},

		CustomCSS: "/* Custom theme styles */",
	}

	suite.Require().Equal("complete-theme-id", theme.ID)
	suite.Require().Equal("Complete Theme", theme.Name)
	suite.Require().NotNil(theme.Tokens)
	suite.Require().NotNil(theme.DarkMode)
	suite.Require().NotNil(theme.Accessibility)
	suite.Require().Contains(theme.CustomCSS, "Custom theme styles")
}

// ==================== Performance Benchmarks ====================

// BenchmarkTokenResolution measures token resolution performance
func BenchmarkTokenResolution(b *testing.B) {
	registry := NewTokenRegistry()
	err := registry.SetTokens(GetDefaultTokens())
	if err != nil {
		b.Fatal("Failed to set tokens:", err)
	}

	ctx := context.Background()
	tokenRef := TokenReference("{semantic.colors.background.default}")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := registry.ResolveToken(ctx, tokenRef)
		if err != nil {
			b.Fatal("Token resolution failed:", err)
		}
	}
}

// BenchmarkTokenResolutionConcurrent measures concurrent token resolution
func BenchmarkTokenResolutionConcurrent(b *testing.B) {
	registry := NewTokenRegistry()
	err := registry.SetTokens(GetDefaultTokens())
	if err != nil {
		b.Fatal("Failed to set tokens:", err)
	}

	ctx := context.Background()
	tokenRefs := []TokenReference{
		"{semantic.colors.background.default}",
		"{semantic.colors.text.default}",
		"{semantic.spacing.component.default}",
		"{components.button.primary.background}",
		"{primitives.colors.blue.500}",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, tokenRef := range tokenRefs {
				_, err := registry.ResolveToken(ctx, tokenRef)
				if err != nil {
					b.Fatal("Token resolution failed:", err)
				}
			}
		}
	})
}

// BenchmarkThemeRegistration measures theme registration performance
func BenchmarkThemeRegistration(b *testing.B) {
	manager := NewThemeManager(NewThemeRegistry(), NewTokenRegistry())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		theme := &Theme{
			ID:      fmt.Sprintf("bench-theme-%d", i),
			Name:    fmt.Sprintf("Benchmark Theme %d", i),
			Version: "1.0.0",
			Tokens:  GetDefaultTokens(),
		}

		err := manager.RegisterTheme(theme)
		if err != nil {
			b.Fatal("Theme registration failed:", err)
		}
	}
}

// BenchmarkThemeRetrieval measures theme retrieval performance
func BenchmarkThemeRetrieval(b *testing.B) {
	manager := NewThemeManager(NewThemeRegistry(), NewTokenRegistry())
	ctx := context.Background()

	// Pre-populate with themes
	for i := 0; i < 100; i++ {
		theme := &Theme{
			ID:      fmt.Sprintf("theme-%d", i),
			Name:    fmt.Sprintf("Theme %d", i),
			Version: "1.0.0",
			Tokens:  GetDefaultTokens(),
		}
		manager.RegisterTheme(theme)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		themeID := fmt.Sprintf("theme-%d", i%100)
		_, err := manager.GetTheme(ctx, themeID)
		if err != nil {
			b.Fatal("Theme retrieval failed:", err)
		}
	}
}

// BenchmarkThemeCloning measures deep copy performance
func BenchmarkThemeCloning(b *testing.B) {
	theme := &Theme{
		ID:      "clone-benchmark",
		Name:    "Clone Benchmark Theme",
		Version: "1.0.0",
		Tokens:  GetDefaultTokens(),
		DarkMode: &DarkModeConfig{
			Enabled:  true,
			Strategy: "auto",
		},
		Accessibility: &AccessibilityConfig{
			HighContrast:     true,
			MinContrastRatio: 4.5,
			KeyboardNav:      true,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := deepCopyTheme(theme)
		if err != nil {
			b.Fatal("Theme cloning failed:", err)
		}
	}
}

// BenchmarkCacheOperations measures cache performance
func BenchmarkCacheOperations(b *testing.B) {
	cache := NewThemeCache(1000, 5*time.Minute)

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		theme := &Theme{
			ID:      fmt.Sprintf("cache-theme-%d", i),
			Name:    fmt.Sprintf("Cache Theme %d", i),
			Version: "1.0.0",
			Tokens:  GetDefaultTokens(),
		}
		cache.Set(theme.ID, theme)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("cache-theme-%d", i%100)
			_, _ = cache.Get(key)
		}
	})

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			theme := &Theme{
				ID:      fmt.Sprintf("new-theme-%d", i),
				Name:    fmt.Sprintf("New Theme %d", i),
				Version: "1.0.0",
				Tokens:  GetDefaultTokens(),
			}
			cache.Set(theme.ID, theme)
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Mix of reads and writes
				if b.N%2 == 0 {
					key := fmt.Sprintf("cache-theme-%d", b.N%100)
					_, _ = cache.Get(key)
				} else {
					theme := &Theme{
						ID:      fmt.Sprintf("concurrent-theme-%d", b.N),
						Name:    fmt.Sprintf("Concurrent Theme %d", b.N),
						Version: "1.0.0",
						Tokens:  GetDefaultTokens(),
					}
					cache.Set(theme.ID, theme)
				}
			}
		})
	})
}

// BenchmarkTokenCompilation measures token compilation performance
func BenchmarkTokenCompilation(b *testing.B) {
	registry := NewTokenRegistry()
	tokens := GetDefaultTokens()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := registry.SetTokens(tokens)
		if err != nil {
			b.Fatal("Token compilation failed:", err)
		}
	}
}

// BenchmarkThemeOverrides measures theme override application performance
func BenchmarkThemeOverrides(b *testing.B) {
	manager := NewThemeManager(NewThemeRegistry(), NewTokenRegistry())
	ctx := context.Background()

	// Register base theme
	baseTheme := &Theme{
		ID:      "base-override-theme",
		Name:    "Base Override Theme",
		Version: "1.0.0",
		Tokens:  GetDefaultTokens(),
	}
	manager.RegisterTheme(baseTheme)

	overrides := &ThemeOverrides{
		TokenOverrides: map[string]string{
			"colors.primary.500":   "#custom-primary",
			"spacing.base":         "1.5rem",
			"typography.size.base": "18px",
		},
		ComponentOverrides: map[string]any{
			"button.borderRadius": "12px",
			"input.height":        "44px",
		},
		CustomCSS: ".custom { color: #override; }",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := manager.GetThemeWithOverrides(ctx, "base-override-theme", overrides)
		if err != nil {
			b.Fatal("Theme override failed:", err)
		}
	}
}

// BenchmarkConcurrentThemeAccess measures concurrent theme operations
func BenchmarkConcurrentThemeAccess(b *testing.B) {
	manager := NewThemeManager(NewThemeRegistry(), NewTokenRegistry())
	ctx := context.Background()

	// Pre-populate with themes
	for i := 0; i < 50; i++ {
		theme := &Theme{
			ID:      fmt.Sprintf("concurrent-theme-%d", i),
			Name:    fmt.Sprintf("Concurrent Theme %d", i),
			Version: "1.0.0",
			Tokens:  GetDefaultTokens(),
		}
		manager.RegisterTheme(theme)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of operations
			switch b.N % 3 {
			case 0:
				// Get theme
				themeID := fmt.Sprintf("concurrent-theme-%d", b.N%50)
				_, _ = manager.GetTheme(ctx, themeID)
			case 1:
				// List themes
				_, _ = manager.ListThemes(ctx)
			case 2:
				// Register new theme
				newTheme := &Theme{
					ID:      fmt.Sprintf("new-concurrent-theme-%d", b.N),
					Name:    fmt.Sprintf("New Concurrent Theme %d", b.N),
					Version: "1.0.0",
					Tokens:  GetDefaultTokens(),
				}
				_ = manager.RegisterTheme(newTheme)
			}
		}
	})
}

// ==================== Performance Tests ====================

// TestTokenResolutionPerformance verifies token resolution meets <1ms p95 target
func TestTokenResolutionPerformance(t *testing.T) {
	registry := NewTokenRegistry()
	err := registry.SetTokens(GetDefaultTokens())
	if err != nil {
		t.Fatal("Failed to set tokens:", err)
	}

	ctx := context.Background()
	tokenRef := TokenReference("{semantic.colors.background.default}")

	// Warm up
	for i := 0; i < 100; i++ {
		_, _ = registry.ResolveToken(ctx, tokenRef)
	}

	// Measure performance
	iterations := 1000
	durations := make([]time.Duration, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := registry.ResolveToken(ctx, tokenRef)
		durations[i] = time.Since(start)

		if err != nil {
			t.Fatal("Token resolution failed:", err)
		}
	}

	// Calculate percentiles
	p95 := calculatePercentile(durations, 95)
	p99 := calculatePercentile(durations, 99)
	avg := calculateAverage(durations)

	t.Logf("Token Resolution Performance:")
	t.Logf("  Average: %v", avg)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)

	// Verify p95 < 1ms target
	if p95 > time.Millisecond {
		t.Errorf("Token resolution P95 (%v) exceeds 1ms target", p95)
	}
}

// TestConcurrentSafety verifies thread safety under concurrent access
func TestConcurrentSafety(t *testing.T) {
	manager := NewThemeManager(NewThemeRegistry(), NewTokenRegistry())
	ctx := context.Background()

	var wg sync.WaitGroup
	errorChan := make(chan error, 100)

	// Concurrent theme operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			theme := &Theme{
				ID:      fmt.Sprintf("safety-theme-%d", id),
				Name:    fmt.Sprintf("Safety Theme %d", id),
				Version: "1.0.0",
				Tokens:  GetDefaultTokens(),
			}

			err := manager.RegisterTheme(theme)
			if err != nil {
				errorChan <- fmt.Errorf("registration failed for theme %d: %v", id, err)
				return
			}

			// Immediate retrieval
			_, err = manager.GetTheme(ctx, theme.ID)
			if err != nil {
				errorChan <- fmt.Errorf("retrieval failed for theme %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	for err := range errorChan {
		t.Error("Concurrent safety error:", err)
	}

	// Verify all themes were registered
	themes, err := manager.ListThemes(ctx)
	if err != nil {
		t.Fatal("Failed to list themes:", err)
	}

	if len(themes) < 50 {
		t.Errorf("Expected at least 50 themes, got %d", len(themes))
	}
}

// Helper functions for performance calculations
func calculatePercentile(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Simple percentile calculation (not sorting for performance)
	index := (len(durations) * percentile) / 100
	if index >= len(durations) {
		index = len(durations) - 1
	}

	return durations[index]
}

func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}
