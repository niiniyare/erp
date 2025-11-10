package validate

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/niiniyare/erp/internal/shared/format"
)

// registerBuiltInValidators registers common validators using existing packages
func (r *ValidationRegistry) registerBuiltInValidators() {
	// Email validator
	r.Register("email", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("email", "invalid_type", "email validator requires string value")
		}

		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(str) {
			return NewValidationError("email", "invalid_email", "invalid email format")
		}
		return nil
	})

	// Phone number validator
	r.Register("phone", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("phone", "invalid_type", "phone validator requires string value")
		}

		// Remove common formatting characters and check if we have a valid phone number
		cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(str, "")

		// Must have at least 7 digits (minimum for a phone number) and at most 15 (E.164 standard)
		// Allow optional + at the beginning
		if len(cleanPhone) < 7 || len(cleanPhone) > 15 {
			return NewValidationError("phone", "invalid_phone", "invalid phone number format")
		}

		// Check if it starts with + followed by digits, or just digits
		phoneRegex := regexp.MustCompile(`^\+?\d{7,15}$`)
		if !phoneRegex.MatchString(cleanPhone) {
			return NewValidationError("phone", "invalid_phone", "invalid phone number format")
		}
		return nil
	})

	// URL validator
	r.Register("url", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("url", "invalid_type", "url validator requires string value")
		}

		urlRegex := regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
		if !urlRegex.MatchString(str) {
			return NewValidationError("url", "invalid_url", "invalid URL format")
		}
		return nil
	})

	// Date format validator using format package
	r.Register("date_format", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("date_format", "invalid_type", "date_format validator requires string value")
		}

		layout, hasLayout := params["layout"].(string)
		if !hasLayout {
			layout = format.ISO8601Date
		}

		_, err := time.Parse(layout, str)
		if err != nil {
			return NewValidationError("date_format", fmt.Sprintf("invalid date format, expected: %s", layout), "invalid_date_format")
		}
		return nil
	})

	// Time range validator using format package
	r.Register("time_range", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("time_range", "time_range validator requires string value", "invalid_type")
		}

		t, err := format.ParseDuration(str)
		if err != nil {
			return NewValidationError("time_range", "invalid duration format", "invalid_duration")
		}

		if minDuration, hasMin := params["min"].(string); hasMin {
			min, err := format.ParseDuration(minDuration)
			if err != nil {
				return NewValidationError("time_range", "invalid minimum duration", "invalid_min_duration")
			}
			if t < min {
				return NewValidationError("time_range", fmt.Sprintf("duration must be at least %s", minDuration), "duration_too_short")
			}
		}

		if maxDuration, hasMax := params["max"].(string); hasMax {
			max, err := format.ParseDuration(maxDuration)
			if err != nil {
				return NewValidationError("time_range", "invalid maximum duration", "invalid_max_duration")
			}
			if t > max {
				return NewValidationError("time_range", fmt.Sprintf("duration must be at most %s", maxDuration), "duration_too_long")
			}
		}

		return nil
	})

	// Credit card Luhn algorithm validator
	r.Register("luhn", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("luhn", "luhn validator requires string value", "invalid_type")
		}

		// Remove spaces and dashes
		cleanStr := regexp.MustCompile(`[\s-]`).ReplaceAllString(str, "")

		if !isValidLuhn(cleanStr) {
			return NewValidationError("luhn", "invalid credit card number", "invalid_luhn")
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
				return NewValidationError("business_day", "business_day validator requires time.Time or string value", "invalid_type")
			}

			var err error
			timeVal, err = time.Parse(format.ISO8601Date, str)
			if err != nil {
				return NewValidationError("business_day", "invalid date format for business day validation", "invalid_date")
			}
		}

		if !format.IsBusinessDay(timeVal) {
			return NewValidationError("business_day", "date must be a business day (Monday-Friday)", "not_business_day")
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
			// Database uniqueness check would be implemented here with actual DB connection
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
					return NewValidationError("start_date", "invalid start date format", "invalid_start_date")
				}
			default:
				return NewValidationError("start_date", "start_date must be time.Time or string", "invalid_start_date_type")
			}

			switch v := endDate.(type) {
			case time.Time:
				end = v
			case string:
				end, err = time.Parse(format.ISO8601Date, v)
				if err != nil {
					return NewValidationError("end_date", "invalid end date format", "invalid_end_date")
				}
			default:
				return NewValidationError("end_date", "end_date must be time.Time or string", "invalid_end_date_type")
			}

			if start.After(end) {
				return NewValidationError("date_range", "start date must be before end date", "invalid_date_range")
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
				return NewValidationError("password_confirmation", "password confirmation does not match", "password_mismatch")
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
						return NewValidationError("date", "invalid date format", "invalid_date")
					}
				}

				// Check if it's a business day
				if !format.IsBusinessDay(dateTime) {
					return NewValidationError("date", "date must be a business day", "not_business_day")
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

				return NewValidationError("business_hours", "times must be within business hours (9:00 AM - 5:00 PM)", "invalid_business_hours")
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
