// Package validate provides field-level validators and the validation pipeline.
//
// Built-in validators are constructor functions that return def.FieldValidator.
// They are designed to be chained on FieldDef declarations:
//
//	def.Field("email").OfType(def.FieldTypeData).
//	    MaxLen(254).
//	    Validate(validate.Email())
package validate

import (
	"fmt"
	"log/slog"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"

	"awo.so/framework/def"
)

// ── Built-in validators ────────────────────────────────────────────────────────

// Email validates RFC 5321 email format.
func Email() def.FieldValidator {
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		if _, err := mail.ParseAddress(s); err != nil {
			return &def.FieldError{Message: "must be a valid email address"}
		}
		return nil
	}
}

// Phone validates E.164 format (+[country][number], 8-15 digits total).
func Phone() def.FieldValidator {
	re := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		if !re.MatchString(s) {
			return &def.FieldError{Message: "must be in E.164 format (e.g. +254700000000)"}
		}
		return nil
	}
}

// URL validates an absolute HTTP or HTTPS URL.
func URL() def.FieldValidator {
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		u, err := url.ParseRequestURI(s)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return &def.FieldError{Message: "must be an absolute HTTP or HTTPS URL"}
		}
		return nil
	}
}

// KRAPin validates Kenyan KRA PIN format: one letter + 9 digits + one letter.
// Example valid value: A000000000X
func KRAPin() def.FieldValidator {
	re := regexp.MustCompile(`^[A-Z]\d{9}[A-Z]$`)
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		if !re.MatchString(strings.ToUpper(s)) {
			return &def.FieldError{Message: "must be a valid KRA PIN (e.g. A000000000X)"}
		}
		return nil
	}
}

// NHIF validates an NHIF member number (6–9 digits).
func NHIF() def.FieldValidator {
	re := regexp.MustCompile(`^\d{6,9}$`)
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		if !re.MatchString(s) {
			return &def.FieldError{Message: "must be a valid NHIF member number (6–9 digits)"}
		}
		return nil
	}
}

// Regex validates that the field value matches pattern. pattern is compiled
// once at call time and panics on invalid syntax.
func Regex(pattern string) def.FieldValidator {
	re := regexp.MustCompile(pattern) // panics on bad pattern at startup
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		if !re.MatchString(s) {
			return &def.FieldError{Message: fmt.Sprintf("must match pattern %s", pattern)}
		}
		return nil
	}
}

// MinLen validates minimum rune count for string fields.
func MinLen(n int) def.FieldValidator {
	return func(value any, _ def.Record) *def.FieldError {
		s, ok := toString(value)
		if !ok || s == "" {
			return nil
		}
		if utf8.RuneCountInString(s) < n {
			return &def.FieldError{Message: fmt.Sprintf("must be at least %d characters", n)}
		}
		return nil
	}
}

// Currency validates that the field value can be parsed as a precise decimal
// number. It rejects float64 values to prevent silent precision loss on
// financial amounts. Use this validator on FieldTypeCurrency fields.
//
// Accepted types: decimal.Decimal, string (parseable as decimal), int, int64.
// Rejected: float32, float64 (lossy — must be converted to string by caller).
func Currency() def.FieldValidator {
	return func(value any, _ def.Record) *def.FieldError {
		if value == nil {
			return nil
		}
		switch value.(type) {
		case decimal.Decimal:
			return nil // already precise
		case string:
			s, _ := value.(string)
			if s == "" {
				return nil
			}
			if _, err := decimal.NewFromString(s); err != nil {
				return &def.FieldError{Message: "must be a valid decimal number"}
			}
			return nil
		case int, int32, int64, uint, uint32, uint64:
			return nil // lossless integer → decimal
		case float32, float64:
			return &def.FieldError{
				Message: "currency value must not be a floating-point number (use string or decimal.Decimal)",
			}
		default:
			return &def.FieldError{Message: fmt.Sprintf("unsupported currency type %T", value)}
		}
	}
}

// CurrencyRange validates that a decimal.Decimal (or string-parseable decimal)
// field value falls within [min, max] inclusive. Either bound may be nil to
// skip that side of the range check.
//
// Returns a def.FieldValidator suitable for use in FieldDef.Validators.
func CurrencyRange(min, max *decimal.Decimal) def.FieldValidator {
	return func(value any, _ def.Record) *def.FieldError {
		if value == nil {
			return nil
		}
		d, err := toDecimal(value)
		if err != nil {
			return &def.FieldError{Message: "must be a valid decimal number"}
		}
		if min != nil && d.LessThan(*min) {
			return &def.FieldError{
				Message: fmt.Sprintf("must be at least %s", min.String()),
			}
		}
		if max != nil && d.GreaterThan(*max) {
			return &def.FieldError{
				Message: fmt.Sprintf("must be at most %s", max.String()),
			}
		}
		return nil
	}
}

// toDecimal converts v to decimal.Decimal without float precision loss.
// float64 inputs are accepted but logged as a warning — they originate from
// JSON unmarshalling and are an unavoidable ingestion artifact.
func toDecimal(v any) (decimal.Decimal, error) {
	switch t := v.(type) {
	case decimal.Decimal:
		return t, nil
	case string:
		return decimal.NewFromString(t)
	case int:
		return decimal.NewFromInt(int64(t)), nil
	case int32:
		return decimal.NewFromInt(int64(t)), nil
	case int64:
		return decimal.NewFromInt(t), nil
	case uint, uint32, uint64:
		return decimal.NewFromFloat(float64(t.(uint))), nil
	case float64:
		// Lossy but unavoidable when value originates from JSON decode.
		slog.Warn("currency: float64 value may lose precision", "value", t)
		return decimal.NewFromFloat(t), nil
	default:
		return decimal.Zero, fmt.Errorf("unsupported type %T for decimal conversion", v)
	}
}

// ── Pipeline ───────────────────────────────────────────────────────────────────

// ValidationErrors is a collection of field-level validation failures.
// Implements error; serialises to the AMIS-compatible JSON error envelope.
type ValidationErrors []*FieldErr

// FieldErr is one field-level validation failure with the field name populated.
type FieldErr struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Field + ": " + e.Message
	}
	return strings.Join(msgs, "; ")
}

// Run executes the full synchronous validation pipeline against rec.
//
// Pipeline stages (in order):
//  1. Required — non-empty check for IsRequired fields
//  2. MaxLength — rune-count check for Data / SmallText fields
//  3. MinVal / MaxVal — numeric range check
//  4. Options — value-in-set check for Select fields
//  5. Custom FieldValidators registered on the FieldDef
//
// All field errors are collected before returning so the user receives the
// complete error list in one response (no short-circuit within a field).
// A field that fails Required is not further validated (skip subsequent stages).
//
// Returns nil when all validations pass.
func Run(def *def.EntityDefinition, rec def.MutableRecord) ValidationErrors {
	var errs ValidationErrors

	for _, f := range def.Fields {
		value := rec.Get(f.Name)
		fieldErrs := validateField(f, value, rec)
		errs = append(errs, fieldErrs...)
	}

	// Entity-level cross-field validators only run when all field validators pass.
	// This prevents confusing errors (e.g. date range check when both dates are empty).
	if len(errs) == 0 {
		for _, ev := range def.EntityValidators {
			for _, fe := range ev(rec) {
				if fe != nil {
					errs = append(errs, &FieldErr{Field: fe.Field, Message: fe.Message})
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

// validateField runs all stages for one field and returns any errors.
func validateField(f *def.FieldDef, value any, rec def.Record) []*FieldErr {
	var errs []*FieldErr

	empty := isEmpty(value)

	// Stage 1: Required.
	if f.IsRequired && empty {
		label := f.Label
		if label == "" {
			label = f.Name
		}
		errs = append(errs, &FieldErr{Field: f.Name, Message: label + " is required"})
		return errs // skip further stages — value is absent
	}

	if empty {
		return nil // optional field with no value — nothing to validate
	}

	// Stage 2: MaxLength.
	if f.MaxLength > 0 {
		if s, ok := toString(value); ok {
			if utf8.RuneCountInString(s) > f.MaxLength {
				errs = append(errs, &FieldErr{
					Field:   f.Name,
					Message: fmt.Sprintf("must be at most %d characters", f.MaxLength),
				})
			}
		}
	}

	// Stage 3: MinVal / MaxVal.
	if f.MinVal != nil || f.MaxVal != nil {
		if fe := validateNumericRange(f, value); fe != nil {
			errs = append(errs, &FieldErr{Field: f.Name, Message: fe.Message})
		}
	}

	// Stage 4: Options (Select / MultiSelect).
	if f.Type == def.FieldTypeSelect && len(f.Options) > 0 {
		if s, ok := toString(value); ok {
			if !contains(f.Options, s) {
				errs = append(errs, &FieldErr{
					Field:   f.Name,
					Message: fmt.Sprintf("must be one of: %s", strings.Join(f.Options, ", ")),
				})
			}
		}
	}

	// Stage 5: Custom validators.
	for _, v := range f.Validators {
		if fe := v(value, rec); fe != nil {
			errs = append(errs, &FieldErr{Field: f.Name, Message: fe.Message})
		}
	}

	return errs
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toString(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func contains(opts []string, v string) bool {
	for _, o := range opts {
		if o == v {
			return true
		}
	}
	return false
}

func validateNumericRange(f *def.FieldDef, value any) *def.FieldError {
	// For decimal.Decimal values (Currency fields), use exact decimal arithmetic
	// to avoid the precision loss that float64 comparison would introduce on
	// amounts like 1234567890.1234.
	if d, err := toDecimal(value); err == nil {
		if f.MinVal != nil && d.LessThan(*f.MinVal) {
			return &def.FieldError{
				Message: fmt.Sprintf("must be at least %s", f.MinVal.String()),
			}
		}
		if f.MaxVal != nil && d.GreaterThan(*f.MaxVal) {
			return &def.FieldError{
				Message: fmt.Sprintf("must be at most %s", f.MaxVal.String()),
			}
		}
		return nil
	}

	// Fallback for plain integer/float fields that don't carry decimal.Decimal.
	var n float64
	switch v := value.(type) {
	case int:
		n = float64(v)
	case int32:
		n = float64(v)
	case int64:
		n = float64(v)
	case float32:
		n = float64(v)
	case float64:
		n = v
	default:
		return nil // non-numeric — skip range check
	}

	if f.MinVal != nil {
		min, _ := f.MinVal.Float64()
		if n < min {
			return &def.FieldError{
				Message: fmt.Sprintf("must be at least %s", f.MinVal.String()),
			}
		}
	}
	if f.MaxVal != nil {
		max, _ := f.MaxVal.Float64()
		if n > max {
			return &def.FieldError{
				Message: fmt.Sprintf("must be at most %s", f.MaxVal.String()),
			}
		}
	}
	return nil
}
