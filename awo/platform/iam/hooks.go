package iam

import (
	"context"
	"regexp"
	"strings"

	"awo.so/awo/def"
	"awo.so/awo/runtime"
)

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// UserNormalizeHook lowercases email before validation so downstream uniqueness
// checks are case-insensitive without a functional index.
type UserNormalizeHook struct{}

func (h *UserNormalizeHook) BeforeValidate(_ context.Context, rec *def.EntityRecord) error {
	email := strings.ToLower(strings.TrimSpace(rec.GetString("email")))
	rec.Set("email", email)
	return nil
}

// AccountLockHook auto-locks the account when failed_attempts reaches 5.
// Runs on Update — the auth layer increments failed_attempts on bad password.
type AccountLockHook struct{}

func (h *AccountLockHook) BeforeUpdate(_ context.Context, rec *def.EntityRecord, _ *def.EntityRecord) error {
	attempts := rec.GetInt("failed_attempts")
	if attempts >= 5 {
		rec.Set("status", "locked")
	}
	return nil
}

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
