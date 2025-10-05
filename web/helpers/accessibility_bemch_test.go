package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkUniqueIDGeneration benchmarks ID generation methods
func BenchmarkUniqueIDGeneration(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	b.Run("Atomic", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = helper.GenerateUniqueID("test")
			}
		})
	})

	b.Run("Secure", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = helper.GenerateSecureUniqueID("test")
			}
		})
	})
}

// BenchmarkAriaLabelGeneration benchmarks ARIA label generation
func BenchmarkAriaLabelGeneration(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	tests := []struct {
		name    string
		context map[string]string
	}{
		{
			name:    "NoContext",
			context: nil,
		},
		{
			name: "SingleContext",
			context: map[string]string{
				"status": "enabled",
			},
		},
		{
			name: "MultipleContext",
			context: map[string]string{
				"status": "active",
				"count":  "5",
				"level":  "2",
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = helper.GenerateAriaLabel("Submit Button", tt.context)
			}
		})
	}
}

// BenchmarkAttributeGeneration benchmarks attribute generation
func BenchmarkAttributeGeneration(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	b.Run("FormFieldAttributes", func(b *testing.B) {
		config := NewFormFieldConfig().
			Required().
			WithError("error-1").
			AutoComplete("email").
			InputMode("email").
			Build()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.GenerateFormFieldAttributes(config)
		}
	})

	b.Run("ButtonAttributes", func(b *testing.B) {
		config := NewButtonConfig().
			Pressed(true).
			Controls("menu").
			HasPopup().
			Build()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.GenerateButtonAttributes(config)
		}
	})

	b.Run("TableAttributes", func(b *testing.B) {
		config := &TableConfig{
			Caption:  "Data Table",
			Sortable: true,
			RowCount: 100,
			ColCount: 5,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.GenerateTableAttributes(config)
		}
	})
}

// BenchmarkSanitizeID benchmarks ID sanitization
func BenchmarkSanitizeID(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Simple",
			input: "button",
		},
		{
			name:  "WithSpaces",
			input: "submit button form",
		},
		{
			name:  "WithSpecialChars",
			input: "my@special#button$element!",
		},
		{
			name:  "Long",
			input: "this is a very long identifier with many words and special characters @#$%",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = helper.SanitizeID(tt.input)
			}
		})
	}
}

// BenchmarkBuilderPattern benchmarks builder vs direct construction
func BenchmarkBuilderPattern(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	b.Run("FormFieldBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config := NewFormFieldConfig().
				Required().
				WithError("error-1").
				AutoComplete("email").
				Build()
			_ = helper.GenerateFormFieldAttributes(config)
		}
	})

	b.Run("FormFieldDirect", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config := &FormFieldConfig{
				Required:     true,
				HasError:     true,
				ErrorID:      "error-1",
				AutoComplete: "email",
			}
			_ = helper.GenerateFormFieldAttributes(config)
		}
	})

	b.Run("ButtonBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config := NewButtonConfig().
				Pressed(true).
				Controls("menu").
				Build()
			_ = helper.GenerateButtonAttributes(config)
		}
	})

	b.Run("ButtonDirect", func(b *testing.B) {
		pressed := true
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config := &ButtonConfig{
				Pressed:  &pressed,
				Controls: "menu",
			}
			_ = helper.GenerateButtonAttributes(config)
		}
	})
}

// BenchmarkValidation benchmarks validation functions
func BenchmarkValidation(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	b.Run("AriaLabelValidation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = helper.GenerateAriaLabelWithValidation("Valid Label", nil)
		}
	})

	b.Run("AttributeValidation", func(b *testing.B) {
		attrs := map[string]string{
			"aria-pressed": "true",
			"aria-invalid": "false",
			"aria-hidden":  "false",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.ValidateAccessibilityAttributes(attrs)
		}
	})
}

// BenchmarkWithLogger benchmarks performance with logger
func BenchmarkWithLogger(b *testing.B) {
	mockLogger := NewMockLogger()
	helperWithLogger := NewAccessibilityHelperWithLogger(nil, mockLogger)
	helperNoLogger := NewAccessibilityHelper(nil)

	b.Run("WithLogger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helperWithLogger.GenerateUniqueID("test")
		}
	})

	b.Run("NoLogger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helperNoLogger.GenerateUniqueID("test")
		}
	})
}

// BenchmarkContextOperations benchmarks context-related operations
func BenchmarkContextOperations(b *testing.B) {
	helper := NewAccessibilityHelper(nil)
	ctx := SetLanguageContext(context.Background(), "es")

	b.Run("GetLanguageAttribute", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.GetLanguageAttribute(ctx)
		}
	})

	b.Run("IsLanguageSupported", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.IsLanguageSupported("es")
		}
	})

	b.Run("SetLanguageContext", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SetLanguageContext(context.Background(), "es")
		}
	})
}

// BenchmarkMergeAttributes benchmarks attribute merging
func BenchmarkMergeAttributes(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	tests := []struct {
		name     string
		attrMaps []map[string]string
	}{
		{
			name: "TwoMaps",
			attrMaps: []map[string]string{
				{"id": "test", "class": "btn"},
				{"role": "button", "type": "submit"},
			},
		},
		{
			name: "FiveMaps",
			attrMaps: []map[string]string{
				{"id": "test"},
				{"class": "btn"},
				{"role": "button"},
				{"aria-label": "Submit"},
				{"data-action": "submit"},
			},
		},
		{
			name: "LargeMaps",
			attrMaps: []map[string]string{
				{
					"id":          "test",
					"class":       "btn btn-primary",
					"role":        "button",
					"aria-label":  "Submit",
					"data-action": "submit",
				},
				{
					"aria-pressed":  "false",
					"aria-expanded": "false",
					"aria-controls": "menu",
					"tabindex":      "0",
				},
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = helper.MergeAttributes(tt.attrMaps...)
			}
		})
	}
}

// BenchmarkConfigOperations benchmarks config operations
func BenchmarkConfigOperations(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	b.Run("GetConfig", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.GetConfig()
		}
	})

	b.Run("UpdateConfig", func(b *testing.B) {
		config := DefaultAccessibilityConfig()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			helper.UpdateConfig(config)
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent access
func BenchmarkConcurrentOperations(b *testing.B) {
	helper := NewAccessibilityHelper(nil)

	b.Run("ConcurrentIDGeneration", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = helper.GenerateUniqueID("test")
			}
		})
	})

	b.Run("ConcurrentAttributeGeneration", func(b *testing.B) {
		config := NewButtonConfig().Pressed(true).Build()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = helper.GenerateButtonAttributes(config)
			}
		})
	})

	b.Run("ConcurrentConfigRead", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = helper.GetConfig()
			}
		})
	})
}

// TestBenchmarkResults validates benchmark results meet performance requirements
func TestBenchmarkResults(t *testing.T) {
	helper := NewAccessibilityHelper(nil)

	t.Run("UniqueIDGeneration should be fast", func(t *testing.T) {
		iterations := 10000
		start := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < iterations; i++ {
				_ = helper.GenerateUniqueID("test")
			}
		})

		avgNsPerOp := start.NsPerOp()
		require.Less(t, avgNsPerOp, int64(1000),
			"ID generation should take less than 1μs per operation")
	})

	t.Run("AttributeGeneration should be fast", func(t *testing.T) {
		config := NewButtonConfig().Pressed(true).Build()
		iterations := 10000

		start := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < iterations; i++ {
				_ = helper.GenerateButtonAttributes(config)
			}
		})

		avgNsPerOp := start.NsPerOp()
		require.Less(t, avgNsPerOp, int64(500),
			"Attribute generation should take less than 500ns per operation")
	})
}
