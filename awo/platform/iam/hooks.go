package iam

import (
	"context"
	"regexp"
	"strings"

	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// UserValidator enforces pre-create rules on iam_user records.
type UserValidator struct{}

func (v *UserValidator) BeforeCreate(_ context.Context, rec *def.EntityRecord) error {
	email := strings.ToLower(strings.TrimSpace(rec.GetString("email")))
	rec.Set("email", email)

	errs := make(map[string]string)
	if !emailRe.MatchString(email) {
		errs["email"] = "must be a valid email address"
	}
	// password_hash must be pre-hashed by the caller; we only check non-empty.
	if rec.GetString("password_hash") == "" {
		errs["password_hash"] = "required"
	}
	if len(errs) > 0 {
		return &runtime.ValidationError{Fields: errs}
	}
	return nil
}
