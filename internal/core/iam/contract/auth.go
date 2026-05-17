package contract

import "context"

// LoginResult carries the session token and resolved context returned by a
// successful Login call.
type LoginResult struct {
	// Token is the raw bearer token. The caller is responsible for setting
	// it in an HttpOnly cookie or Authorization header.
	Token string

	// Session is the resolved identity and runtime context for the new session.
	Session SessionContext
}

// AuthService is the narrow IAM authentication contract for non-IAM modules.
//
// It exposes only the auth entrypoints that external modules need — Login,
// Logout, and session validation. It deliberately hides:
//   - MFA completion flow (handled by the auth handler, not modules)
//   - SSO login path (handled by the SSO handler)
//   - Casbin enforcement (handled by IAM middleware)
//
// Obtain an implementation via [NewServiceAdapter].
type AuthService interface {
	// Login authenticates the user identified by email + password.
	// Returns (LoginResult, nil) on success.
	// Returns (zero, ErrMFARequired) when MFA is configured — the caller
	// must redirect to the MFA completion endpoint; the returned LoginResult
	// carries an empty Session but a non-empty Token (the pending MFA token).
	// Returns (zero, err) for invalid credentials or other errors.
	Login(ctx context.Context, email, password string) (LoginResult, error)

	// Logout invalidates the session identified by token.
	// Returns nil on success or when the token is already invalid.
	Logout(ctx context.Context, token string) error

	// ValidateSession checks whether token identifies a live, non-expired
	// session and returns its context.
	// Returns (zero, false) for missing, expired, or invalidated tokens.
	ValidateSession(ctx context.Context, token string) (SessionContext, bool)
}
