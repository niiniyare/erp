// Package token defines JWT/session token payload types for the Awo Framework.
package token

import (
	"time"

	"github.com/google/uuid"
)

// Payload contains the claims carried in an auth token.
type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// AuthorizationPayloadKey is the context/locals key for the token payload.
const AuthorizationPayloadKey = "authorization_payload"

// IsExpired reports whether the token has passed its expiry time.
func (p *Payload) IsExpired() bool {
	return time.Now().After(p.ExpiredAt)
}
