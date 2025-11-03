package schema

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/niiniyare/erp/internal/shared/format"
	"github.com/niiniyare/erp/pkg/condition"
)

// ValidatorFunc represents a custom validation function
type ValidatorFunc func(ctx context.Context, value any, params map[string]any) error

// AsyncValidatorFunc represents an async validation function with debouncing
type AsyncValidatorFunc func(ctx context.Context, value any, params map[string]any) error

// ValidationRegistry manages custom validators and provides extensible validation
type ValidationRegistry struct {
	validators      map[string]ValidatorFunc
	asyncValidators map[string]*AsyncValidator
	mu              sync.RWMutex
}

// AsyncValidator wraps async validation with caching and debouncing
type AsyncValidator struct {
	Name     string
	Validate AsyncValidatorFunc
	Debounce time.Duration
	Cache    bool
	CacheTTL time.Duration
	cache    map[string]*CacheEntry
	mu       sync.RWMutex
	cleanup  context.CancelFunc
}

// CacheEntry represents a cached validation result
type CacheEntry struct {
	Result    error
	ExpiresAt time.Time
}

// NewValidationRegistry creates a new validation registry with built-in validators
func NewValidationRegistry() *ValidationRegistry {
	registry := &ValidationRegistry{
		validators:      make(map[string]ValidatorFunc),
		asyncValidators: make(map[string]*AsyncValidator),
	}

	// Register built-in validators
	registry.registerBuiltInValidators()

	return registry
}

// Register adds a custom validator
func (r *ValidationRegistry) Register(name string, validator ValidatorFunc) error {
	if name == "" {
		return NewValidationError("validator_name", "validator name is required")
	}
	if validator == nil {
		return NewValidationError("validator_func", "validator function is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.validators[name] = validator

	return nil
}

// RegisterAsync adds an async validator with cache cleanup
func (r *ValidationRegistry) RegisterAsync(validator *AsyncValidator) error {
	if validator == nil {
		return NewValidationError("async_validator", "async validator is required")
	}
	if validator.Name == "" {
		return NewValidationError("async_validator_name", "async validator name is required")
	}
	if validator.Validate == nil {
		return NewValidationError("async_validator_func", "async validator function is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if validator.cache == nil {
		validator.cache = make(map[string]*CacheEntry)
	}

	// Start cache cleanup goroutine if caching is enabled
	if validator.Cache && validator.CacheTTL > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		validator.cleanup = cancel
		validator.startCacheCleanup(ctx)
	}

	r.asyncValidators[validator.Name] = validator

	return nil
}

// startCacheCleanup starts a goroutine to clean up expired cache entries
func (av *AsyncValidator) startCacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				av.cleanupExpiredCache()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// cleanupExpiredCache removes expired entries from the cache
func (av *AsyncValidator) cleanupExpiredCache() {
	av.mu.Lock()
	defer av.mu.Unlock()

	now := time.Now()
	for key, entry := range av.cache {
		if now.After(entry.ExpiresAt) {
			delete(av.cache, key)
		}
	}
}

// Stop stops the cache cleanup goroutine
func (av *AsyncValidator) Stop() {
	if av.cleanup != nil {
		av.cleanup()
	}
}

// Validate executes a validator by name
func (r *ValidationRegistry) Validate(ctx context.Context, name string, value any, params map[string]any) error {
	r.mu.RLock()
	validator, exists := r.validators[name]
	r.mu.RUnlock()

	if !exists {
		return NewValidationError("validator_not_found", fmt.Sprintf("validator %s not found", name))
	}

	return validator(ctx, value, params)
}

// ValidateAsync executes an async validator with caching and debouncing
func (r *ValidationRegistry) ValidateAsync(ctx context.Context, name string, value any, params map[string]any) error {
	r.mu.RLock()
	validator, exists := r.asyncValidators[name]
	r.mu.RUnlock()

	if !exists {
		return NewValidationError("async_validator_not_found", fmt.Sprintf("async validator %s not found", name))
	}

	// Check cache if enabled
	if validator.Cache {
		cacheKey := fmt.Sprintf("%s:%v", name, value)
		validator.mu.RLock()
		entry, exists := validator.cache[cacheKey]
		validator.mu.RUnlock()

		if exists && time.Now().Before(entry.ExpiresAt) {
			return entry.Result
		}
	}

	// Apply debouncing if configured
	if validator.Debounce > 0 {
		select {
		case <-time.After(validator.Debounce):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Perform validation
	result := validator.Validate(ctx, value, params)

	// Cache result if enabled
	if validator.Cache {
		cacheKey := fmt.Sprintf("%s:%v", name, value)
		validator.mu.Lock()
		validator.cache[cacheKey] = &CacheEntry{
			Result:    result,
			ExpiresAt: time.Now().Add(validator.CacheTTL),
		}
		validator.mu.Unlock()
	}

	return result
}

// UnregisterAsync removes an async validator and stops its cleanup goroutine
func (r *ValidationRegistry) UnregisterAsync(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	validator, exists := r.asyncValidators[name]
	if !exists {
		return NewValidationError("async_validator_not_found", fmt.Sprintf("async validator %s not found", name))
	}

	// Stop cleanup goroutine
	validator.Stop()

	delete(r.asyncValidators, name)
	return nil
}

// ValidatorChain allows chaining multiple validators
type ValidatorChain struct {
	validators []ValidatorFunc
}

func NewValidatorChain() *ValidatorChain {
	return &ValidatorChain{
		validators: []ValidatorFunc{},
	}
}

func (vc *ValidatorChain) Add(validator ValidatorFunc) *ValidatorChain {
	vc.validators = append(vc.validators, validator)
	return vc
}

func (vc *ValidatorChain) Validate(ctx context.Context, value any, params map[string]any) error {
	for _, validator := range vc.validators {
		if err := validator(ctx, value, params); err != nil {
			return err
		}
	}
	return nil
}

// CrossFieldValidator validates multiple fields together using condition engine
type CrossFieldValidator struct {
	Name      string
	Fields    []string
	Condition *condition.ConditionGroup
	Validate  func(ctx context.Context, values map[string]any) error
}

// CrossFieldValidationRegistry manages cross-field validators
type CrossFieldValidationRegistry struct {
	validators map[string]*CrossFieldValidator
	evaluator  *condition.Evaluator
	mu         sync.RWMutex
}

// NewCrossFieldValidationRegistry creates a new cross-field validation registry
func NewCrossFieldValidationRegistry() *CrossFieldValidationRegistry {
	// Initialize condition evaluator with reasonable defaults
	config := &condition.Config{
		MaxLevel: 5,
		Types: map[string]condition.TypeConfig{
			"text": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual, condition.OpContains,
					condition.OpStartsWith, condition.OpEndsWith, condition.OpIsEmpty,
					condition.OpIsNotEmpty, condition.OpMatchRegexp,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
			"number": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual, condition.OpLess,
					condition.OpLessOrEqual, condition.OpGreater, condition.OpGreaterOrEqual,
					condition.OpBetween, condition.OpNotBetween,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
			"date": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual, condition.OpLess,
					condition.OpLessOrEqual, condition.OpGreater, condition.OpGreaterOrEqual,
					condition.OpBetween, condition.OpNotBetween,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
		},
	}

	evaluator := condition.NewEvaluator(config, condition.DefaultEvalOptions())

	registry := &CrossFieldValidationRegistry{
		validators: make(map[string]*CrossFieldValidator),
		evaluator:  evaluator,
	}

	// Register built-in cross-field validators
	registry.registerBuiltInValidators()

	return registry
}

// Register adds a cross-field validator
func (r *CrossFieldValidationRegistry) Register(validator *CrossFieldValidator) error {
	if validator == nil {
		return NewValidationError("cross_validator", "cross-field validator is required")
	}
	if validator.Name == "" {
		return NewValidationError("cross_validator_name", "cross-field validator name is required")
	}
	if len(validator.Fields) == 0 {
		return NewValidationError("cross_validator_fields", "cross-field validator must specify fields")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.validators[validator.Name] = validator

	return nil
}

// Validate executes a cross-field validator
func (r *CrossFieldValidationRegistry) Validate(ctx context.Context, name string, values map[string]any) error {
	r.mu.RLock()
	validator, exists := r.validators[name]
	r.mu.RUnlock()

	if !exists {
		return NewValidationError("cross_validator_not_found", fmt.Sprintf("cross-field validator %s not found", name))
	}

	// Check if all required fields are present
	fieldValues := make(map[string]any)
	for _, fieldName := range validator.Fields {
		if value, exists := values[fieldName]; exists {
			fieldValues[fieldName] = value
		}
	}

	// If condition is specified, evaluate it first
	if validator.Condition != nil {
		evalCtx := condition.NewEvalContext(fieldValues, condition.DefaultEvalOptions())
		shouldValidate, err := r.evaluator.Evaluate(ctx, validator.Condition, evalCtx)
		if err != nil {
			return WrapError(err, "condition_evaluation_failed", "failed to evaluate condition")
		}

		// Only validate if condition is true
		if !shouldValidate {
			return nil
		}
	}

	return validator.Validate(ctx, fieldValues)
}

// registerBuiltInValidators registers common validators using existing packages
func (r *ValidationRegistry) registerBuiltInValidators() {
	// Email validator
	r.Register("email", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "email validator requires string value")
		}

		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(str) {
			return NewValidationError("invalid_email", "invalid email format")
		}
		return nil
	})

	// Phone number validator
	r.Register("phone", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "phone validator requires string value")
		}

		// Remove common formatting characters and check if we have a valid phone number
		cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(str, "")

		// Must have at least 7 digits (minimum for a phone number) and at most 15 (E.164 standard)
		// Allow optional + at the beginning
		if len(cleanPhone) < 7 || len(cleanPhone) > 15 {
			return NewValidationError("invalid_phone", "invalid phone number format")
		}

		// Check if it starts with + followed by digits, or just digits
		phoneRegex := regexp.MustCompile(`^\+?\d{7,15}$`)
		if !phoneRegex.MatchString(cleanPhone) {
			return NewValidationError("invalid_phone", "invalid phone number format")
		}
		return nil
	})

	// URL validator
	r.Register("url", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "url validator requires string value")
		}

		urlRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
		if !urlRegex.MatchString(str) {
			return NewValidationError("invalid_url", "invalid URL format")
		}
		return nil
	})

	// Date format validator using format package
	r.Register("date_format", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "date_format validator requires string value")
		}

		layout, hasLayout := params["layout"].(string)
		if !hasLayout {
			layout = format.ISO8601Date
		}

		_, err := time.Parse(layout, str)
		if err != nil {
			return NewValidationError("invalid_date_format", fmt.Sprintf("invalid date format, expected: %s", layout))
		}
		return nil
	})

	// Time range validator using format package
	r.Register("time_range", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "time_range validator requires string value")
		}

		t, err := format.ParseDuration(str)
		if err != nil {
			return NewValidationError("invalid_duration", "invalid duration format")
		}

		if minDuration, hasMin := params["min"].(string); hasMin {
			min, err := format.ParseDuration(minDuration)
			if err != nil {
				return NewValidationError("invalid_min_duration", "invalid minimum duration")
			}
			if t < min {
				return NewValidationError("duration_too_short", fmt.Sprintf("duration must be at least %s", minDuration))
			}
		}

		if maxDuration, hasMax := params["max"].(string); hasMax {
			max, err := format.ParseDuration(maxDuration)
			if err != nil {
				return NewValidationError("invalid_max_duration", "invalid maximum duration")
			}
			if t > max {
				return NewValidationError("duration_too_long", fmt.Sprintf("duration must be at most %s", maxDuration))
			}
		}

		return nil
	})

	// Credit card Luhn algorithm validator
	r.Register("luhn", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "luhn validator requires string value")
		}

		// Remove spaces and dashes
		cleanStr := regexp.MustCompile(`[\s-]`).ReplaceAllString(str, "")

		if !isValidLuhn(cleanStr) {
			return NewValidationError("invalid_luhn", "invalid credit card number")
		}
		return nil
	})

	// Business day validator using format package
	r.Register("business_day", func(ctx context.Context, value any, params map[string]any) error {
		timeVal, ok := value.(time.Time)
		if !ok {
			// Try to parse from string
			str, isString := value.(string)
			if !isString {
				return NewValidationError("invalid_type", "business_day validator requires time.Time or string value")
			}

			var err error
			timeVal, err = time.Parse(format.ISO8601Date, str)
			if err != nil {
				return NewValidationError("invalid_date", "invalid date format for business day validation")
			}
		}

		if !format.IsBusinessDay(timeVal) {
			return NewValidationError("not_business_day", "date must be a business day (Monday-Friday)")
		}
		return nil
	})

	// Unique validator (async example - would need database access)
	r.RegisterAsync(&AsyncValidator{
		Name:     "unique",
		Debounce: 300 * time.Millisecond,
		Cache:    true,
		CacheTTL: 5 * time.Minute,
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			// Placeholder implementation - in real usage would check database
			// table := params["table"].(string)
			// field := params["field"].(string)
			// TODO: Check database for uniqueness
			return nil
		},
	})
}

// registerBuiltInValidators for cross-field validation using condition engine
func (r *CrossFieldValidationRegistry) registerBuiltInValidators() {
	// Date range validator (start_date < end_date)
	r.Register(&CrossFieldValidator{
		Name:   "date_range",
		Fields: []string{"start_date", "end_date"},
		Validate: func(ctx context.Context, values map[string]any) error {
			startDate, hasStart := values["start_date"]
			endDate, hasEnd := values["end_date"]

			if !hasStart || !hasEnd {
				return nil // Skip if either date is missing
			}

			// Convert to time.Time
			var start, end time.Time
			var err error

			switch v := startDate.(type) {
			case time.Time:
				start = v
			case string:
				start, err = time.Parse(format.ISO8601Date, v)
				if err != nil {
					return NewValidationError("invalid_start_date", "invalid start date format")
				}
			default:
				return NewValidationError("invalid_start_date_type", "start_date must be time.Time or string")
			}

			switch v := endDate.(type) {
			case time.Time:
				end = v
			case string:
				end, err = time.Parse(format.ISO8601Date, v)
				if err != nil {
					return NewValidationError("invalid_end_date", "invalid end date format")
				}
			default:
				return NewValidationError("invalid_end_date_type", "end_date must be time.Time or string")
			}

			if start.After(end) {
				return NewValidationError("invalid_date_range", "start date must be before end date")
			}

			return nil
		},
	})

	// Password confirmation validator
	r.Register(&CrossFieldValidator{
		Name:   "password_confirmation",
		Fields: []string{"password", "password_confirmation"},
		Validate: func(ctx context.Context, values map[string]any) error {
			password, hasPassword := values["password"]
			confirmation, hasConfirmation := values["password_confirmation"]

			if !hasPassword || !hasConfirmation {
				return nil // Skip if either field is missing
			}

			if password != confirmation {
				return NewValidationError("password_mismatch", "password confirmation does not match")
			}

			return nil
		},
	})

	// Business hours validator using format package
	r.Register(&CrossFieldValidator{
		Name:   "business_hours",
		Fields: []string{"start_time", "end_time", "date"},
		Validate: func(ctx context.Context, values map[string]any) error {
			date, hasDate := values["date"]
			if hasDate {
				var dateTime time.Time
				var err error

				switch v := date.(type) {
				case time.Time:
					dateTime = v
				case string:
					dateTime, err = time.Parse(format.ISO8601Date, v)
					if err != nil {
						return NewValidationError("invalid_date", "invalid date format")
					}
				}

				// Check if it's a business day
				if !format.IsBusinessDay(dateTime) {
					return NewValidationError("not_business_day", "date must be a business day")
				}
			}

			startTime, hasStart := values["start_time"]
			endTime, hasEnd := values["end_time"]

			if hasStart && hasEnd {
				// Validate business hours (e.g., 9 AM to 5 PM)
				startStr := fmt.Sprintf("%v", startTime)
				endStr := fmt.Sprintf("%v", endTime)

				if startStr >= "09:00" && endStr <= "17:00" && startStr < endStr {
					return nil
				}

				return NewValidationError("invalid_business_hours", "times must be within business hours (9:00 AM - 5:00 PM)")
			}

			return nil
		},
	})
}

// isValidLuhn implements the Luhn algorithm for credit card validation
func isValidLuhn(number string) bool {
	if len(number) == 0 {
		return false
	}

	var sum int
	var alternate bool

	// Loop through digits from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + (digit / 10)
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// Enhanced field validation that uses the registry
func (f *Field) ValidateWithRegistry(ctx context.Context, value any, registry *ValidationRegistry) error {
	// First run standard validation (excluding custom validation)
	if err := f.validateValueExcludingCustom(ctx, value); err != nil {
		return err
	}

	// Then run custom validators if configured
	if f.Validation != nil && f.Validation.Custom != "" {
		params := make(map[string]any)

		// Add validation parameters
		if f.Validation.Min != nil {
			params["min"] = *f.Validation.Min
		}
		if f.Validation.Max != nil {
			params["max"] = *f.Validation.Max
		}
		if f.Validation.MinLength != nil {
			params["minLength"] = *f.Validation.MinLength
		}
		if f.Validation.MaxLength != nil {
			params["maxLength"] = *f.Validation.MaxLength
		}
		if f.Validation.Pattern != "" {
			params["pattern"] = f.Validation.Pattern
		}
		if f.Validation.Format != "" {
			params["format"] = f.Validation.Format
		}

		return registry.Validate(ctx, f.Validation.Custom, value, params)
	}

	return nil
}

// ValidateSchemaWithRegistries validates a schema using both validation registries
func ValidateSchemaWithRegistries(ctx context.Context, schema *Schema, data map[string]any,
	fieldRegistry *ValidationRegistry, crossFieldRegistry *CrossFieldValidationRegistry,
) error {
	collector := NewErrorCollector()

	// Validate individual fields
	for _, field := range schema.Fields {
		if value, exists := data[field.Name]; exists || field.Required {
			if err := field.ValidateWithRegistry(ctx, value, fieldRegistry); err != nil {
				if schemaErr, ok := err.(SchemaError); ok {
					collector.AddFieldError(field.Name, schemaErr)
				} else {
					collector.AddValidationError(field.Name, "validation_failed", err.Error())
				}
			}
		}
	}

	// Validate cross-field rules if configured
	if schema.Validation != nil {
		// This would integrate with a cross-field validation system
		// For now, we'll use built-in validators
		for _, validator := range []string{"date_range", "password_confirmation", "business_hours"} {
			if err := crossFieldRegistry.Validate(ctx, validator, data); err != nil {
				if schemaErr, ok := err.(SchemaError); ok {
					collector.AddError(schemaErr)
				} else {
					collector.AddValidationError("", "cross_validation_failed", err.Error())
				}
			}
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}
