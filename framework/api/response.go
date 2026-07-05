package api

// response.go — HTTP response helpers: serialisation, error mapping, and
// query-parameter parsing.
//
// Keeping these in a separate file from the handler methods makes it easier to
// audit what data leaves the system (recordToMap), how errors are translated to
// HTTP status codes (fiberErr), and what query parameters are accepted (listOptsFromQuery).

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/persistence"
	"awo.so/framework/validate"
)

// ──────────────────────────────────────────────────────────────────────────────
// Sentinel errors
// ──────────────────────────────────────────────────────────────────────────────

// errForbidden is a package-level sentinel used to signal a privacy or
// org-scope denial from within a transaction callback.  It is translated to
// HTTP 403 by [fiberErr].
//
// Using a sentinel (rather than fiber.ErrForbidden directly) keeps the
// persistence and privacy layers free of HTTP concerns.
var errForbidden = errors.New("forbidden")

// ──────────────────────────────────────────────────────────────────────────────
// Error translation
// ──────────────────────────────────────────────────────────────────────────────

// fiberErr translates domain errors returned from persistence, privacy, and
// hook layers into Fiber HTTP errors.
//
// Mapping:
//
//	persistence.ErrNotFound → 404 Not Found
//	persistence.ErrConflict → 409 Conflict  (carries the store's message)
//	errForbidden            → 403 Forbidden
//	everything else         → propagated as-is (Fiber renders unknown errors as 500)
func fiberErr(err error) error {
	switch {
	case errors.Is(err, persistence.ErrNotFound):
		return fiber.ErrNotFound

	case errors.Is(err, persistence.ErrConflict):
		// Include the store's conflict message (e.g. "duplicate invoice number")
		// so the client can surface it without an extra round-trip.
		return fiber.NewError(fiber.StatusConflict, err.Error())

	case errors.Is(err, errForbidden):
		return fiber.ErrForbidden

	default:
		// Fiber catches non-Fiber errors and renders them as 500.  Do not
		// attempt to unwrap or stringify here — that would leak internal detail.
		return err
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Validation response
// ──────────────────────────────────────────────────────────────────────────────

// validationResponse writes a 422 Unprocessable Entity response containing the
// structured validation errors.
//
// The response body is:
//
//	{ "status": 422, "errors": [ { "field": "...", "message": "..." }, ... ] }
//
// This is a dedicated helper (rather than inline JSON in each handler) so that
// the 422 shape is consistent across create and update.
func validationResponse(c *fiber.Ctx, verrs validate.ValidationErrors) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"status": fiber.StatusUnprocessableEntity,
		"errors": verrs,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Record serialisation
// ──────────────────────────────────────────────────────────────────────────────

// recordToMap serialises a [def.Record] to a plain map suitable for
// JSON encoding.
//
// Fields marked IsSensitive on the entity definition are omitted entirely —
// they are never included in HTTP responses regardless of the caller's
// permission level.  Sensitive fields (e.g. password hashes, raw encryption
// keys) must be accessed through dedicated, audited endpoints if needed at all.
//
// System fields included unconditionally:
//
//	id         — always present
//	created_at — always present
//	updated_at — always present
//
// Conditional system fields:
//
//	tenant_id  — included for non-global entities (IsGlobal() == false)
//	org_unit_id — included for unit-scoped entities whose record implements OrgUnitID()
func recordToMap(rec def.Record, def *def.EntityDefinition) map[string]any {
	// Pre-allocate with a reasonable capacity: 3 system fields + entity fields.
	m := make(map[string]any, len(def.Fields)+4)

	// System fields present on every non-global record.
	m["id"] = rec.ID()
	m["created_at"] = rec.Get("created_at")
	m["updated_at"] = rec.Get("updated_at")

	// tenant_id is meaningful only for tenant-scoped entities.
	// Global entities (e.g. currency codes, country list) are shared across
	// tenants and must not expose a tenant_id.
	if !def.IsGlobal() {
		m["tenant_id"] = rec.TenantID()
	}

	// org_unit_id is only relevant for unit-scoped entities.
	if def.IsUnitScoped() {
		if scoped, ok := rec.(interface{ OrgUnitID() uuid.UUID }); ok {
			m["org_unit_id"] = scoped.OrgUnitID()
		}
	}

	// Entity-declared fields — sensitive fields are skipped entirely.
	for _, f := range def.Fields {
		if !f.IsSensitive {
			m[f.Name] = rec.Get(f.Name)
		}
	}

	return m
}

// ──────────────────────────────────────────────────────────────────────────────
// Query parameter parsing
// ──────────────────────────────────────────────────────────────────────────────

// listOptsFromQuery builds a [persistence.ListOptions] from the query
// parameters of a list request.
//
// Accepted parameters:
//
//	q        string  — search / filter string forwarded verbatim to the store
//	order_by string  — column name; the store is responsible for allowlist validation
//	dir      string  — "asc" → ascending; anything else → descending (default)
//	limit    int     — page size; defaults to 50; the store may impose a maximum
//	offset   int     — zero-based record offset; defaults to 0
//
// OrgUnitIDs is left nil here; it is populated by the list handler after
// resolving the viewer's org-unit subtree.
func listOptsFromQuery(c *fiber.Ctx) persistence.ListOptions {
	opts := persistence.ListOptions{
		Search:    c.Query("q"),
		OrderBy:   c.Query("order_by"),
		Ascending: c.Query("dir") == "asc",
	}

	// Default limit to 50; a zero or negative value from the query string is
	// treated as "use default" rather than "return zero rows".
	if lim := c.QueryInt("limit", 50); lim > 0 {
		opts.Limit = lim
	}

	// Default offset to 0; negative offsets are clamped to 0.
	if off := c.QueryInt("offset", 0); off >= 0 {
		opts.Offset = off
	}

	return opts
}

// ──────────────────────────────────────────────────────────────────────────────
// Parameter parsing
// ──────────────────────────────────────────────────────────────────────────────

// parseUUID extracts and parses a UUID from a named Fiber route parameter.
// Returns HTTP 400 if the parameter is missing or not a valid UUID.
func parseUUID(c *fiber.Ctx, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, fiber.NewError(
			fiber.StatusBadRequest,
			"invalid UUID for parameter '"+param+"': "+c.Params(param),
		)
	}
	return id, nil
}
