package iam

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// errAuthzService is a minimal AuthzService stub whose Enforce always returns
// a non-nil error. All other methods are no-ops.
// Used for AZ-MID-030: Enforce error must produce 500, not 403.
type errAuthzService struct{}

func (errAuthzService) Enforce(_ context.Context, _ Request) (bool, error) {
	return false, fmt.Errorf("simulated enforcer failure")
}

func (errAuthzService) EnforceBatch(_ context.Context, _ []Request) ([]bool, error) {
	return nil, nil
}

func (errAuthzService) AssignRole(_ context.Context, _, _, _, _ string, _ ...AssignOpt) error {
	return nil
}

func (errAuthzService) RevokeRole(_ context.Context, _, _, _ string) error {
	return nil
}

func (errAuthzService) GetRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (errAuthzService) GetImplicitRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (errAuthzService) HasRole(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}

func (errAuthzService) GetAssignments(_ context.Context, _, _ string) ([]RoleAssignment, error) {
	return nil, nil
}

func (errAuthzService) AddPolicy(_ context.Context, _ Policy) error {
	return nil
}

func (errAuthzService) RemovePolicy(_ context.Context, _ Policy) error {
	return nil
}

func (errAuthzService) GetPolicies(_ context.Context, _ string) ([]Policy, error) {
	return nil, nil
}

func (errAuthzService) InvalidateCache(_ context.Context) error {
	return nil
}

func (errAuthzService) BootstrapTenantAdmin(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
