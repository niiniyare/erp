package infrastructure

import (
	"context"

	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/platform/temporal"
	"github.com/niiniyare/erp/internal/shared/logger"
)

type Temporal struct {
	Platform *temporal.Platform
}

func InitializeTemporal(cfg *config.Config, logger logger.Logger) (*Temporal, error) {
	// Initialize Temporal platform
	temporalPlatform, err := temporal.NewPlatform(&cfg.Temporal, logger)
	if err != nil {
		return nil, err
	}

	return &Temporal{
		Platform: temporalPlatform,
	}, nil
}

func (t *Temporal) Shutdown(ctx context.Context) error {
	if t.Platform != nil {
		return t.Platform.Stop(ctx)
	}
	return nil
}

// GetPlatform returns the Temporal platform
func (t *Temporal) GetPlatform() *temporal.Platform {
	return t.Platform
}
