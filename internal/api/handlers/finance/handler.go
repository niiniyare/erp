// Package finance provides HTTP handlers for all finance module endpoints.
//
// File layout:
//
//	handler.go        — FinanceHandler struct, constructor, Wire provider set
//	base.go           — Shared response helpers, error handling, pagination
//	helpers.go        — Decimal conversion, UUID parsing, custom validators
//	accounts.go       — Chart of accounts CRUD
//	transactions.go   — Transaction CRUD + submit/approve/reject workflow
//	fiscal_periods.go — Fiscal years and accounting periods
//	exchange_rates.go — Exchange rates and currencies
//	budgets.go        — Budget lifecycle (draft → submit → approve → close)
//	cost_centers.go   — Cost center hierarchy
//	tax.go            — Tax authorities and tax codes
//	reconciliation.go — Bank statement import + line matching
package finance

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"

	financeService "awo.so/internal/core/finance/service"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// FinanceHandler is the single handler struct for all finance HTTP endpoints.
//
// It holds no state beyond its dependencies — all per-request state lives on
// the fiber.Ctx or in local variables. Every method follows the same four-step
// pattern documented in base.go.
//
// Dependencies are injected by Wire; see ProviderSet below.
type FinanceHandler struct {
	// services groups every finance sub-service so handlers never import
	// individual service packages directly.
	services *financeService.Services

	// logger is the zerolog-backed structured logger. Use logger.Fields for
	// key/value pairs; avoid fmt.Sprintf inside log calls.
	logger logger.Logger

	// metrics provides Prometheus counters/histograms. Increment after every
	// successful response and on every error path.
	metrics metrics.MetricsProvider

	// tracer wraps OpenTelemetry span creation. Every handler opens a span
	// as its first action and defers span.End().
	tracer tracing.Service

	// validator is a shared go-playground/validator instance with AWO custom
	// rules registered at construction time (see helpers.go).
	validator *validator.Validate
}

// NewFinanceHandler constructs a FinanceHandler with all required dependencies.
//
// Called by Wire — do not call directly in application code.
// Custom validators are registered here once; they are safe for concurrent use.
func NewFinanceHandler(
	services *financeService.Services,
	log logger.Logger,
	m metrics.MetricsProvider,
	tracer tracing.Service,
) *FinanceHandler {
	v := validator.New()
	registerCustomValidators(v) // see helpers.go

	return &FinanceHandler{
		services:  services,
		logger:    log,
		metrics:   m,
		tracer:    tracer,
		validator: v,
	}
}

// ProviderSet is the Wire provider set for the finance handler.
// Register this in the top-level wire.go alongside other handler providers.
var ProviderSet = wire.NewSet(NewFinanceHandler)
