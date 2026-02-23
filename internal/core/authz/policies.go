package authz

import (
	"context"
	"fmt"
)

// AddPolicy inserts a p-rule. Effect must be "allow" or "deny".
func (s *service) AddPolicy(ctx context.Context, p Policy) error {
	if p.Effect != "allow" && p.Effect != "deny" {
		return fmt.Errorf("authz AddPolicy: effect must be \"allow\" or \"deny\", got %q", p.Effect)
	}
	if p.Subject == "" || p.Domain == "" || p.Object == "" || p.Action == "" {
		return ErrInvalidRequest
	}

	// Explicit duplicate check: Casbin's AddPolicy return value is unreliable
	// for nil-adapter (in-memory) enforcers — HasPolicy reads the in-memory
	// model directly and is consistent across all enforcer configurations.
	exists, err := s.enforcer.HasPolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	if err != nil {
		return fmt.Errorf("authz AddPolicy HasPolicy: %w", err)
	}
	if exists {
		return ErrPolicyConflict
	}

	_, err = s.enforcer.AddPolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	if err != nil {
		return fmt.Errorf("authz AddPolicy: %w", err)
	}
	return nil
}

// RemovePolicy deletes a p-rule.
func (s *service) RemovePolicy(ctx context.Context, p Policy) error {
	if p.Subject == "" || p.Domain == "" || p.Object == "" || p.Action == "" {
		return ErrInvalidRequest
	}

	_, err := s.enforcer.RemovePolicy(p.Subject, p.Domain, p.Object, p.Action, p.Effect)
	if err != nil {
		return fmt.Errorf("authz RemovePolicy: %w", err)
	}
	return nil
}

// GetPolicies returns all p-rules filtered to a specific domain (v1 field).
func (s *service) GetPolicies(ctx context.Context, domain string) ([]Policy, error) {
	// fieldIndex=1 matches v1 (the domain column in p-rules).
	raw, err := s.enforcer.GetFilteredPolicy(1, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered policy for domain %s: %w", domain, err)
	}

	out := make([]Policy, 0, len(raw))
	for _, r := range raw {
		// p = sub, dom, obj, act, eft
		if len(r) < 5 {
			continue
		}
		out = append(out, Policy{
			Subject: r[0],
			Domain:  r[1],
			Object:  r[2],
			Action:  r[3],
			Effect:  r[4],
		})
	}
	return out, nil
}
