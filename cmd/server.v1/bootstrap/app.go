package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/niiniyare/erp/cmd/server/services"
	"github.com/niiniyare/erp/internal/platform/config"
	"github.com/niiniyare/erp/internal/shared/logger"
)

type Application struct {
	Core   *services.Core
	Config *config.Config
}

func NewApplication() (*Application, error) {
	// Load configuration
	cfg, _ := config.LoadWithViper()
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Log successful configuration load
	logger.Info("Configuration loaded successfully", logger.Fields{
		"app_name": cfg.App.Name,
		"version":  cfg.App.Version,
		"stage":    cfg.App.Stage,
		"port":     cfg.Server.Port,
	})

	// Create application core
	app := services.NewCore(cfg)

	return &Application{
		Core:   app,
		Config: cfg,
	}, nil
}

func (a *Application) Start(ctx context.Context) error {
	return a.Core.Start(ctx)
}

func (a *Application) CreateHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + a.Config.Server.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 60,
		ReadTimeout:       a.Config.Server.ReadTimeout,
		WriteTimeout:      a.Config.Server.WriteTimeout,
	}
}

func (a *Application) Stop(ctx context.Context) error {
	return a.Core.Stop(ctx)
}
