package token

import (
	"time"

	"github.com/google/uuid"
)

// Payload contains the payload data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// AuthorizationPayloadKey is the key for storing the authorization payload in the context.
const AuthorizationPayloadKey = "authorization_payload"
