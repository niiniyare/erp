package iam_test

import (
	"context"
	"testing"

	"awo.so/awo/def"
	"awo.so/awo/platform/iam"
	"awo.so/awo/runtime"
)

func TestUserValidator_ValidEmail(t *testing.T) {
	v := &iam.UserValidator{}
	rec := &def.EntityRecord{
		Data: map[string]any{
			"email":         "alice@example.com",
			"password_hash": "$2a$12$somehash",
		},
	}
	if err := v.BeforeCreate(context.Background(), rec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Email should be lowercased.
	if got := rec.GetString("email"); got != "alice@example.com" {
		t.Errorf("email not normalised: %q", got)
	}
}

func TestUserValidator_InvalidEmail(t *testing.T) {
	v := &iam.UserValidator{}
	rec := &def.EntityRecord{
		Data: map[string]any{
			"email":         "not-an-email",
			"password_hash": "$2a$12$somehash",
		},
	}
	err := v.BeforeCreate(context.Background(), rec)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !runtime.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestUserValidator_MissingPasswordHash(t *testing.T) {
	v := &iam.UserValidator{}
	rec := &def.EntityRecord{
		Data: map[string]any{
			"email":         "bob@example.com",
			"password_hash": "",
		},
	}
	err := v.BeforeCreate(context.Background(), rec)
	if err == nil {
		t.Fatal("expected validation error for empty password_hash, got nil")
	}
	var ve *runtime.ValidationError
	if !runtime.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
	_ = ve
}
