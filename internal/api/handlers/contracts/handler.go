package contracts

import (
	"github.com/go-playground/validator/v10"

	"awo.so/internal/core/contracts"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
)

// Handler handles all HTTP requests for the contracts module.
type Handler struct {
	svc       contracts.Service
	log       logger.Logger
	metrics   metrics.MetricsProvider
	tracer    tracing.Service
	validator *validator.Validate
}

// New creates a contracts Handler.
func New(
	svc contracts.Service,
	log logger.Logger,
	m metrics.MetricsProvider,
	tracer tracing.Service,
) *Handler {
	return &Handler{
		svc:       svc,
		log:       log,
		metrics:   m,
		tracer:    tracer,
		validator: validator.New(),
	}
}
