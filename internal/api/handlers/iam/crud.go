package iam

// crud.go — Policy and role assignment handler methods.
//
// Routes (registered in routes.go):
//
//	GET    /api/v1/iam/policies              → ListPolicies
//	POST   /api/v1/iam/policies              → AddPolicy
//	DELETE /api/v1/iam/policies              → RemovePolicy
//	GET    /api/v1/iam/assignments           → ListAssignments
//	POST   /api/v1/iam/roles/assign          → AssignRole
//	POST   /api/v1/iam/roles/revoke          → RevokeRole
//	GET    /api/v1/iam/roles                 → GetRoles

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"

	"awo.so/internal/core/iam/contract"
	sharedErrors "awo.so/internal/shared/errors"
)

// ============================================================================
// Policy management
// ============================================================================

// ListPolicies returns all Casbin p-rules for a domain.
//
// GET /api/v1/iam/policies
// Query: domain (defaults to current tenant)
func (h *IAMHandler) ListPolicies(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.ListPolicies")
	defer span.End()
	c.SetUserContext(ctx)

	domain := tenantDomain(c)
	if domain == "" {
		return h.fail(c, sharedErrors.NewBusinessError("MISSING_DOMAIN", "domain is required (X-Tenant-ID or ?domain=)").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation))
	}

	policies, err := h.service.GetPolicies(ctx, domain)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(
		attribute.String("iam.domain", domain),
		attribute.Int("iam.policy_count", len(policies)),
	)
	return h.ok200(c, policies)
}

// addPolicyRequest is the request body for AddPolicy.
type addPolicyRequest struct {
	Subject string `json:"subject" validate:"required"`
	Domain  string `json:"domain"` // optional: defaults to tenant domain
	Object  string `json:"object" validate:"required"`
	Action  string `json:"action" validate:"required"`
	Effect  string `json:"effect" validate:"required,oneof=allow deny"`
}

// AddPolicy adds a Casbin p-rule.
//
// POST /api/v1/iam/policies
func (h *IAMHandler) AddPolicy(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.AddPolicy")
	defer span.End()
	c.SetUserContext(ctx)

	var req addPolicyRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	domain := req.Domain
	if domain == "" {
		domain = tenantDomain(c)
	}
	if domain == "" {
		return h.fail(c, sharedErrors.NewBusinessError("MISSING_DOMAIN", "domain is required").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation))
	}

	p := contract.Policy{
		Subject: req.Subject,
		Domain:  domain,
		Object:  req.Object,
		Action:  req.Action,
		Effect:  req.Effect,
	}

	if err := h.service.AddPolicy(ctx, p); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(
		attribute.String("iam.domain", domain),
		attribute.String("iam.subject", p.Subject),
	)
	return h.ok201(c, p)
}

// RemovePolicy removes a Casbin p-rule.
//
// DELETE /api/v1/iam/policies (body with Policy fields)
func (h *IAMHandler) RemovePolicy(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.RemovePolicy")
	defer span.End()
	c.SetUserContext(ctx)

	var req addPolicyRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	domain := req.Domain
	if domain == "" {
		domain = tenantDomain(c)
	}

	p := contract.Policy{
		Subject: req.Subject,
		Domain:  domain,
		Object:  req.Object,
		Action:  req.Action,
		Effect:  req.Effect,
	}

	if err := h.service.RemovePolicy(ctx, p); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok204(c)
}

// ============================================================================
// Role assignment management
// ============================================================================

// ListAssignments returns role assignments for a subject in a domain.
//
// GET /api/v1/iam/assignments?subject=<sub>&domain=<dom>
func (h *IAMHandler) ListAssignments(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.ListAssignments")
	defer span.End()
	c.SetUserContext(ctx)

	subject := c.Query("subject")
	if subject == "" {
		return h.fail(c, sharedErrors.NewBusinessError("MISSING_SUBJECT", "subject query param is required").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation))
	}

	domain := c.Query("domain")
	if domain == "" {
		domain = tenantDomain(c)
	}

	assignments, err := h.service.GetAssignments(ctx, subject, domain)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(
		attribute.String("iam.subject", subject),
		attribute.String("iam.domain", domain),
	)
	return h.ok200(c, assignments)
}

// assignRoleRequest is the request body for AssignRole.
type assignRoleRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Subject  string `json:"subject" validate:"required"`
	Role     string `json:"role" validate:"required"`
	Domain   string `json:"domain"` // optional: defaults to tenant domain
}

// AssignRole assigns a role to a subject.
//
// POST /api/v1/iam/roles/assign
func (h *IAMHandler) AssignRole(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.AssignRole")
	defer span.End()
	c.SetUserContext(ctx)

	var req assignRoleRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	domain := req.Domain
	if domain == "" {
		domain = tenantDomain(c)
	}

	if err := h.service.AssignRole(ctx, req.TenantID, req.Subject, req.Role, domain); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(
		attribute.String("iam.subject", req.Subject),
		attribute.String("iam.role", req.Role),
		attribute.String("iam.domain", domain),
	)
	return h.ok201(c, fiber.Map{
		"subject": req.Subject,
		"role":    req.Role,
		"domain":  domain,
	})
}

// revokeRoleRequest is the request body for RevokeRole.
type revokeRoleRequest struct {
	Subject string `json:"subject" validate:"required"`
	Role    string `json:"role" validate:"required"`
	Domain  string `json:"domain"` // optional: defaults to tenant domain
}

// RevokeRole revokes a role from a subject.
//
// POST /api/v1/iam/roles/revoke
func (h *IAMHandler) RevokeRole(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.RevokeRole")
	defer span.End()
	c.SetUserContext(ctx)

	var req revokeRoleRequest
	if err := h.bind(c, &req); err != nil {
		return h.fail(c, err)
	}

	domain := req.Domain
	if domain == "" {
		domain = tenantDomain(c)
	}

	if err := h.service.RevokeRole(ctx, req.Subject, req.Role, domain); err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	return h.ok204(c)
}

// GetRoles returns the roles for a subject in a domain.
//
// GET /api/v1/iam/roles?subject=<sub>&domain=<dom>
func (h *IAMHandler) GetRoles(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.UserContext(), "iam.GetRoles")
	defer span.End()
	c.SetUserContext(ctx)

	subject := c.Query("subject")
	if subject == "" {
		return h.fail(c, sharedErrors.NewBusinessError("MISSING_SUBJECT", "subject query param is required").
			WithHTTPStatus(fiber.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation))
	}

	domain := c.Query("domain")
	if domain == "" {
		domain = tenantDomain(c)
	}

	roles, err := h.service.GetRoles(ctx, subject, domain)
	if err != nil {
		span.RecordError(err)
		return h.fail(c, err)
	}

	span.SetAttributes(
		attribute.String("iam.subject", subject),
		attribute.String("iam.domain", domain),
	)
	return h.ok200(c, fiber.Map{
		"subject": subject,
		"domain":  domain,
		"roles":   roles,
	})
}
