package middleware

import (
	"time"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Period context keys
const (
	AccountingPeriodKey   = "accounting_period"
	AccountingPeriodIDKey = "accounting_period_id"
)

// PeriodChecker is the minimal interface required by PeriodMiddleware to resolve periods.
// The concrete implementation is provided by the finance service layer.
type PeriodChecker interface {
	GetPeriodForDate(ctx *fiber.Ctx, tenantID uuid.UUID, date time.Time) (*domain.AccountingPeriod, error)
	GetCurrentPeriod(ctx *fiber.Ctx, tenantID uuid.UUID) (*domain.AccountingPeriod, error)
}

// PeriodMiddlewareConfig controls which request methods trigger period validation.
type PeriodMiddlewareConfig struct {
	// WriteMethods is the set of HTTP methods that require an open period.
	// Defaults to POST, PATCH, PUT, DELETE.
	WriteMethods map[string]bool
	// DateHeader is the name of the optional header a caller may supply to override
	// the posting date (e.g. "X-Posting-Date" in RFC3339 / "2006-01-02" format).
	DateHeader string
	// SkipPaths is a list of path prefixes that bypass period checking entirely
	// (e.g. "/finance/periods", "/finance/fiscal-years").
	SkipPaths []string
}

// defaultPeriodConfig returns sensible defaults.
func defaultPeriodConfig() PeriodMiddlewareConfig {
	return PeriodMiddlewareConfig{
		WriteMethods: map[string]bool{
			fiber.MethodPost:   true,
			fiber.MethodPatch:  true,
			fiber.MethodPut:    true,
			fiber.MethodDelete: true,
		},
		DateHeader: "X-Posting-Date",
		SkipPaths: []string{
			"/finance/periods",
			"/finance/fiscal-years",
		},
	}
}

// PeriodMiddleware returns a Fiber handler that:
//  1. Skips read-only requests (GET, HEAD, OPTIONS).
//  2. Resolves the accounting period from the posting date header (or today if absent).
//  3. Rejects the request with 422 if the resolved period is CLOSED or LOCKED.
//  4. Stores the period in Fiber Locals for downstream handlers.
func PeriodMiddleware(checker PeriodChecker, cfgOverrides ...PeriodMiddlewareConfig) fiber.Handler {
	cfg := defaultPeriodConfig()
	if len(cfgOverrides) > 0 {
		c := cfgOverrides[0]
		if len(c.WriteMethods) > 0 {
			cfg.WriteMethods = c.WriteMethods
		}
		if c.DateHeader != "" {
			cfg.DateHeader = c.DateHeader
		}
		if len(c.SkipPaths) > 0 {
			cfg.SkipPaths = c.SkipPaths
		}
	}

	return func(c *fiber.Ctx) error {
		// Only enforce on write methods
		if !cfg.WriteMethods[c.Method()] {
			return c.Next()
		}

		// Skip configured paths (e.g. the period-management endpoints themselves)
		path := c.Path()
		for _, skip := range cfg.SkipPaths {
			if len(path) >= len(skip) && path[:len(skip)] == skip {
				return c.Next()
			}
		}

		// Extract tenant ID from context (set by TenantMiddleware earlier in the chain)
		tenantID, ok := shared.GetTenantID(c.UserContext())
		if !ok {
			logger.WarnContext(c.UserContext(), "PeriodMiddleware: tenant_id missing from context")
			return sendErrorResponse(c, "tenant context is required", fiber.StatusBadRequest, "")
		}

		// Determine the posting date — caller may override via header
		postingDate := time.Now().UTC()
		if raw := c.Get(cfg.DateHeader); raw != "" {
			parsed, err := time.Parse("2006-01-02", raw)
			if err != nil {
				return sendErrorResponse(c,
					"invalid "+cfg.DateHeader+" header - expected YYYY-MM-DD format",
					fiber.StatusBadRequest, "")
			}
			postingDate = parsed.UTC()
		}

		// Resolve the accounting period for this date
		period, err := checker.GetPeriodForDate(c, tenantID, postingDate)
		if err != nil {
			logger.WarnContext(c.UserContext(), "PeriodMiddleware: period lookup failed",
				logger.Fields{
					"tenant_id":    tenantID.String(),
					"posting_date": postingDate.Format("2006-01-02"),
					"error":        err.Error(),
				})
			return sendErrorResponse(c,
				"no accounting period found for posting date "+postingDate.Format("2006-01-02"),
				fiber.StatusUnprocessableEntity, "")
		}

		// Reject if the period does not allow posting
		if !period.Status.AllowsPosting() {
			logger.WarnContext(c.UserContext(), "PeriodMiddleware: period is closed",
				logger.Fields{
					"tenant_id":     tenantID.String(),
					"period_id":     period.ID.String(),
					"period_name":   period.Name,
					"period_status": string(period.Status),
				})
			return sendErrorResponse(c,
				"accounting period '"+period.Name+"' is "+string(period.Status)+
					" - no new transactions may be posted",
				fiber.StatusUnprocessableEntity, "")
		}

		// Make period available to downstream handlers
		c.Locals(AccountingPeriodKey, period)
		c.Locals(AccountingPeriodIDKey, period.ID.String())

		return c.Next()
	}
}

// GetPeriodFromLocals retrieves the AccountingPeriod stored by PeriodMiddleware.
// Returns nil if the middleware was not applied or the request was read-only.
func GetPeriodFromLocals(c *fiber.Ctx) *domain.AccountingPeriod {
	if v := c.Locals(AccountingPeriodKey); v != nil {
		if p, ok := v.(*domain.AccountingPeriod); ok {
			return p
		}
	}
	return nil
}
