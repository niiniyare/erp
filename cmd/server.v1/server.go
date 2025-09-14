package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/niiniyare/erp/internal/shared/logger"
)

// serve starts the HTTP server, listens for OS signals for graceful shutdown,
// and blocks until the server is closed.
func serve(app *application, handler http.Handler) error {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = app.config.Server.Port
	}
	if port == "" {
		return errors.New("server port must be configured")
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       app.config.Server.ReadTimeout,
		WriteTimeout:      app.config.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		app.logger.Info("HTTP server starting", logger.Fields{
			"address": srv.Addr,
		})
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		} else {
			serverErrors <- nil
		}
	}()

	// Give server a moment to fail on startup
	time.Sleep(50 * time.Millisecond)

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("server failed to start: %w", err)
		}
		return nil

	case sig := <-shutdownChannel:
		app.logger.Info("Caught signal, initiating graceful shutdown", logger.Fields{
			"signal": sig.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			app.logger.Error("Graceful shutdown failed, forcing close", logger.Fields{"error": err})
			if closeErr := srv.Close(); closeErr != nil {
				app.logger.Error("Failed to forcefully close server", logger.Fields{"error": closeErr})
			}
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		app.logger.Info("HTTP server stopped gracefully")
		return nil
	}
}
