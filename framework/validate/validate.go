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
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

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
	// Convert value to float64 for comparison.
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
