package audit

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"awo.so/awo/def"
)

func TestMarshalNullable_Nil(t *testing.T) {
	t.Parallel()

	b, err := marshalNullable(nil)
	if err != nil {
		t.Fatalf("marshalNullable(nil): unexpected error %v", err)
	}
	if b != nil {
		t.Errorf("marshalNullable(nil) = %v, want nil", b)
	}
}

func TestMarshalNullable_EmptyMap(t *testing.T) {
	t.Parallel()

	b, err := marshalNullable(map[string]any{})
	if err != nil {
		t.Fatalf("marshalNullable({}): unexpected error %v", err)
	}
	// Empty map → "{}" → should return nil (treated as NULL).
	if b != nil {
		t.Errorf("marshalNullable(empty map) = %q, want nil (NULL)", string(b))
	}
}

func TestMarshalNullable_EmptySlice(t *testing.T) {
	t.Parallel()

	b, err := marshalNullable([]string{})
	if err != nil {
		t.Fatalf("marshalNullable([]): unexpected error %v", err)
	}
	if b != nil {
		t.Errorf("marshalNullable(empty slice) = %q, want nil (NULL)", string(b))
	}
}

func TestMarshalNullable_NonEmpty(t *testing.T) {
	t.Parallel()

	b, err := marshalNullable(map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("marshalNullable: unexpected error %v", err)
	}
	if b == nil {
		t.Error("marshalNullable(non-empty map) must not return nil")
	}
	if string(b) != `{"key":"value"}` {
		t.Errorf("marshalNullable: got %q, want %q", string(b), `{"key":"value"}`)
	}
}

func TestMarshalNullable_NonEmptyStringSlice(t *testing.T) {
	t.Parallel()

	b, err := marshalNullable([]string{"status", "amount"})
	if err != nil {
		t.Fatalf("marshalNullable([]string): unexpected error %v", err)
	}
	if b == nil {
		t.Error("marshalNullable(non-empty slice) must not return nil")
	}
}

func TestNullableString(t *testing.T) {
	t.Parallel()

	if nullableString("") != nil {
		t.Error("nullableString(\"\") must return nil")
	}
	if v := nullableString("hello"); v != "hello" {
		t.Errorf("nullableString(\"hello\") = %v, want \"hello\"", v)
	}
}

func TestNullableUUID(t *testing.T) {
	t.Parallel()

	if nullableUUID(uuid.Nil) != nil {
		t.Error("nullableUUID(uuid.Nil) must return nil")
	}
	id := uuid.New()
	if v := nullableUUID(id); v != id {
		t.Errorf("nullableUUID(%v) = %v, want original UUID", id, v)
	}
}

// TestTransactionalWriter_Write_InvalidRecord verifies that Validate errors
// are returned before any SQL is attempted (no DB connection required).
func TestTransactionalWriter_Write_InvalidRecord(t *testing.T) {
	t.Parallel()

	w := NewTransactionalWriter(nil) // nil fallback: must not be reached
	rec := AuditRecord{}             // invalid: missing TenantID, EntityName, etc.

	err := w.Write(context.TODO(), rec)
	//nolint:staticcheck — intentionally nil ctx to prove we don't reach SQL
	if err == nil {
		t.Error("Write with invalid record: expected error from Validate()")
	}
}

// TestTransactionalWriter_Write_NilFallback verifies that Write returns a
// descriptive error when no TX is in context and no fallback querier is set.
// This guards against silent no-ops from misconfigured deployments.
func TestTransactionalWriter_Write_NilFallback(t *testing.T) {
	t.Parallel()

	w := NewTransactionalWriter(nil) // nil fallback — standalone writes must fail clearly
	rec := AuditRecord{
		TenantID:      uuid.New(),
		EntityName:    "finance_invoice",
		Operation:     OperationCreate,
		EventCategory: CategoryData,
		Actor:         &def.Actor{UserID: uuid.New()},
	}

	// context.Background() has no active transaction — fallback path is taken.
	err := w.Write(context.Background(), rec)
	if err == nil {
		t.Error("Write with nil fallback and no TX: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no database connection available") {
		t.Errorf("Write with nil fallback: expected descriptive error, got: %v", err)
	}
}
