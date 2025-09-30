package bootstrap

import (
	"fmt"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type Dependencies struct {
	Store       db.Store
	RedisClient cache.Service
	Logger      logger.Logger
	Metrics     *metrics.MetricsService
	Tracing     tracing.TracingService
}

func InitializeDependencies(app *Application) (*Dependencies, error) {
	// Get application services
	appServices := app.Core.GetServices()

	return &Dependencies{
		Store:       appServices.Store,
		RedisClient: appServices.RedisClient,
		Logger:      appServices.Logger,
		Metrics:     appServices.Metrics,
		Tracing:     appServices.Tracing,
	}, nil
}

func (d *Dependencies) Validate() error {
	if d.Store == nil {
		return fmt.Errorf("database store is required")
	}
	if d.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	return nil
}
