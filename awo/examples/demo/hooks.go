package demo

import (
	"context"
	"strings"

	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

// CustomerValidator runs before every demo_customer Create.
// Validates that name is non-empty and email (if supplied) contains "@".
type CustomerValidator struct{}

func (v *CustomerValidator) BeforeCreate(_ context.Context, r *def.EntityRecord) error {
	ve := &runtime.ValidationError{Fields: map[string]string{}}

	name, _ := r.Data["name"].(string)
	if strings.TrimSpace(name) == "" {
		ve.Fields["name"] = "required"
	}

	code, _ := r.Data["customer_code"].(string)
	if strings.TrimSpace(code) == "" {
		ve.Fields["customer_code"] = "required"
	}

	email, _ := r.Data["email"].(string)
	if email != "" && !strings.Contains(email, "@") {
		ve.Fields["email"] = "must be a valid email address"
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
