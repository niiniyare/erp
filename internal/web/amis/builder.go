// Package amis provides a fluent Go builder for AMIS JSON schemas.
// Every builder implements json.Marshaler so it works directly with
// fiber's c.JSON() — no .Build() call needed.
package amis

import "encoding/json"

// M is a schema map. Alias for clarity.
type M = map[string]any

// A is a schema array. Alias for clarity.
type A = []any

// Schema is the root type returned by every page function.
type Schema = M

// Ctx carries request-scoped data into schema functions.
// Schema functions are pure: given the same Ctx they produce the same JSON.
type Ctx struct {
	Flags map[string]bool
	User  CtxUser
	// Can checks a permission for the current user.
	// Returns false if access service is not provided.
	Can func(action, resource string) bool
}

// CtxUser holds the minimal user info schema functions need.
type CtxUser struct {
	ID         string
	TenantID   string
	TenantName string
	Role       string
}

// SchemaFn is the signature every page builder function must satisfy.
type SchemaFn func(ctx Ctx) Schema

// base is the shared builder core.
type base struct {
	s M
}

func (b *base) set(key string, val any) {
	if val != nil {
		b.s[key] = val
	}
}

func (b *base) setIf(key string, val any, cond bool) {
	if cond {
		b.s[key] = val
	}
}

func (b *base) marshalJSON() ([]byte, error) {
	return json.Marshal(b.s)
}
