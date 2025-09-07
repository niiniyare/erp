package temporal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/niiniyare/erp/internal/platform/config"
	loggerPkg "github.com/niiniyare/erp/internal/shared/logger"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

// ClientManager manages the Temporal client connection
type ClientManager struct {
	client client.Client
	config *config.TemporalConfig
	logger loggerPkg.Logger
}

// NewClientManager creates a new Temporal client manager
func NewClientManager(cfg *config.TemporalConfig, logger loggerPkg.Logger) (*ClientManager, error) {
	temporalClient, err := createTemporalClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	return &ClientManager{
		client: temporalClient,
		config: cfg,
		logger: logger,
	}, nil
}

// GetClient returns the Temporal client
func (cm *ClientManager) GetClient() client.Client {
	return cm.client
}

// Close closes the Temporal client connection
func (cm *ClientManager) Close() {
	if cm.client != nil {
		cm.client.Close()
		cm.logger.Info("Temporal client connection closed")
	}
}

// createTemporalClient creates a Temporal client with proper configuration
func createTemporalClient(cfg *config.TemporalConfig, logger loggerPkg.Logger) (client.Client, error) {
	clientOptions := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    &temporalLoggerAdapter{logger: logger},
		Identity:  cfg.Client.Identity,
	}

	// Configure TLS if enabled
	if cfg.TLS.Enabled {
		tlsConfig, err := createTLSConfig(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		clientOptions.ConnectionOptions = client.ConnectionOptions{
			TLS: tlsConfig,
		}
		logger.Info("TLS configuration applied to Temporal client", loggerPkg.Fields{
			"server_name":          cfg.TLS.ServerName,
			"insecure_skip_verify": cfg.TLS.InsecureSkipVerify,
		})
	}

	// Configure connection settings
	if cfg.Client.ConnectionTimeout > 0 {
		if clientOptions.ConnectionOptions.KeepAliveTime == 0 {
			clientOptions.ConnectionOptions.KeepAliveTime = cfg.Client.KeepAliveTime
		}
		if clientOptions.ConnectionOptions.KeepAliveTimeout == 0 {
			clientOptions.ConnectionOptions.KeepAliveTimeout = cfg.Client.KeepAliveTimeout
		}
		clientOptions.ConnectionOptions.DisableKeepAlivePermitWithoutStream = !cfg.Client.KeepAlivePermitWithoutStream
		logger.Debug("Connection keep-alive settings configured", loggerPkg.Fields{
			"keep_alive_time":    clientOptions.ConnectionOptions.KeepAliveTime,
			"keep_alive_timeout": clientOptions.ConnectionOptions.KeepAliveTimeout,
		})
	}

	// Configure context propagators if specified
	if len(cfg.Client.ContextPropagators) > 0 {
		// Create context propagators based on configuration
		var propagators []workflow.ContextPropagator
		for _, propName := range cfg.Client.ContextPropagators {
			switch propName {
			case "tracing":
				// Add OpenTelemetry tracing propagator if needed
				// This would typically be: propagators = append(propagators, temporal.NewTracingContextPropagator())
				logger.Debug("Tracing propagator configuration available")
			case "baggage":
				// Add baggage propagator if needed
				logger.Debug("Baggage propagator configuration available")
			default:
				logger.Warn("Unknown context propagator", loggerPkg.Fields{
					"propagator": propName,
				})
			}
		}
		if len(propagators) > 0 {
			clientOptions.ContextPropagators = propagators
		}
		logger.Info("Context propagators configured", loggerPkg.Fields{
			"propagators": cfg.Client.ContextPropagators,
			"count":       len(propagators),
		})
	}

	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		logger.Error("Failed to connect to Temporal server", loggerPkg.Fields{
			"error":     err.Error(),
			"host_port": cfg.HostPort,
			"namespace": cfg.Namespace,
		})
		return nil, fmt.Errorf("failed to dial temporal server: %w", err)
	}

	logger.Info("✅ Temporal Client Connected", loggerPkg.Fields{
		"host_port": cfg.HostPort,
		"namespace": cfg.Namespace,
		"identity":  cfg.Client.Identity,
		"status":    "connected",
	})

	return temporalClient, nil
}

// createTLSConfig creates TLS configuration from config
func createTLSConfig(tlsConfig config.TemporalTLSConfig) (*tls.Config, error) {
	config := &tls.Config{
		ServerName:         tlsConfig.ServerName,
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
	}

	if tlsConfig.CertPath != "" && tlsConfig.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.CertPath, tlsConfig.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Add CA certificate loading if needed
	if tlsConfig.CaPath != "" {
		caCert, err := os.ReadFile(tlsConfig.CaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		config.RootCAs = caCertPool
	}

	return config, nil
}

// temporalLoggerAdapter adapts our logger interface to Temporal's logger interface
type temporalLoggerAdapter struct {
	logger loggerPkg.Logger
}

func (l *temporalLoggerAdapter) Debug(msg string, keyvals ...any) {
	l.logger.Debug(msg, convertKeyVals(keyvals))
}

func (l *temporalLoggerAdapter) Info(msg string, keyvals ...any) {
	l.logger.Info(msg, convertKeyVals(keyvals))
}

func (l *temporalLoggerAdapter) Warn(msg string, keyvals ...any) {
	l.logger.Warn(msg, convertKeyVals(keyvals))
}

func (l *temporalLoggerAdapter) Error(msg string, keyvals ...any) {
	l.logger.Error(msg, convertKeyVals(keyvals))
}

func (l *temporalLoggerAdapter) With(keyvals ...any) log.Logger {
	// Return a new logger with additional context
	fields := convertKeyVals(keyvals)
	return &temporalLoggerAdapter{
		logger: l.logger.WithFields(fields),
	}
}

// convertKeyVals converts Temporal's key-value pairs to our logger format
func convertKeyVals(keyvals []any) loggerPkg.Fields {
	fields := make(loggerPkg.Fields)
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			if key, ok := keyvals[i].(string); ok {
				fields[key] = keyvals[i+1]
			}
		}
	}
	return fields
}

// CheckConnection verifies the Temporal connection is healthy
func (cm *ClientManager) CheckConnection(ctx context.Context) error {
	// Try to get workflow service info as a health check
	_, err := cm.client.WorkflowService().GetSystemInfo(ctx, &workflowservice.GetSystemInfoRequest{})
	if err != nil {
		cm.logger.Error("Temporal connection health check failed", loggerPkg.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("temporal health check failed: %w", err)
	}

	cm.logger.Debug("Temporal connection health check passed")
	return nil
}

// GetNamespace returns the configured namespace
func (cm *ClientManager) GetNamespace() string {
	return cm.config.Namespace
}

// GetHostPort returns the configured host:port
func (cm *ClientManager) GetHostPort() string {
	return cm.config.HostPort
}

