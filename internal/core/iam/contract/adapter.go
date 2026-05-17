package contract

import (
	"context"

	"awo.so/internal/core/iam"
)

// SessionServiceAdapter wraps [iam.SessionService] to satisfy [AuthService].
//
// This is the production implementation — create it once at wire time and
// inject it into any module that needs to trigger auth flows.
//
// The adapter does NOT duplicate session logic; it delegates entirely to the
// core SessionService and converts the returned domain types into contract DTOs.
type SessionServiceAdapter struct {
	sessions iam.SessionService
}

// NewServiceAdapter constructs an [AuthService] backed by the core SessionService.
func NewServiceAdapter(sessions iam.SessionService) AuthService {
	return &SessionServiceAdapter{sessions: sessions}
}

// Login delegates to [iam.SessionService.Login] and wraps the result in
// contract DTOs. When MFA is required, the error is propagated unchanged so
// that callers can detect [iam.ErrMFARequired] and redirect appropriately.
func (a *SessionServiceAdapter) Login(ctx context.Context, email, password string) (LoginResult, error) {
	resolved, token, err := a.sessions.Login(ctx, email, password)
	if err != nil {
		// Propagate as-is — callers check for specific sentinel errors
		// (e.g. ErrMFARequired) via errors.Is.
		return LoginResult{Token: token}, err
	}
	return LoginResult{
		Token:   token,
		Session: newSessionContext(resolved),
	}, nil
}

// Logout delegates to [iam.SessionService.Logout].
func (a *SessionServiceAdapter) Logout(ctx context.Context, token string) error {
	return a.sessions.Logout(ctx, token)
}

// ValidateSession delegates to [iam.SessionService.ValidateSession].
// Returns (zero, false) when validation fails for any reason.
func (a *SessionServiceAdapter) ValidateSession(ctx context.Context, token string) (SessionContext, bool) {
	resolved, err := a.sessions.ValidateSession(ctx, token)
	if err != nil || resolved == nil {
		return SessionContext{}, false
	}
	return newSessionContext(resolved), true
}
