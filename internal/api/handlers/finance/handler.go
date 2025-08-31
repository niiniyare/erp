package finance

import (
	"errors"

	goaFinance "github.com/niiniyare/erp/internal/api/gen/finance"
	"github.com/niiniyare/erp/internal/core/finance/service"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// FinanceHandler implements the Goa finance service interface
type FinanceHandler struct {
	financeServices *service.Services
	tracing         tracing.TracingService
	metrics         metrics.MetricsProvider
}

// NewFinanceHandler creates a new finance handler that implements the Goa service interface
func NewFinanceHandler(
	financeServices *service.Services,
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
) goaFinance.Service {
	return &FinanceHandler{
		financeServices: financeServices,
		tracing:         tracing,
		metrics:         metrics,
	}
}

// handleError converts domain errors to appropriate Goa errors
func (h *FinanceHandler) handleError(err error) error {
	var businessErr *sharedErrors.BusinessError
	if errors.As(err, &businessErr) {
		switch businessErr.HTTPStatus {
		case 400:
			return goaFinance.MakeBadRequest(err)
		case 404:
			return goaFinance.MakeNotFound(err)
		case 409:
			return goaFinance.MakeConflict(err)
		case 422:
			return goaFinance.MakeUnprocessableEntity(err)
		default:
			return goaFinance.MakeBadRequest(err)
		}
	}
	return goaFinance.MakeBadRequest(err)
}
