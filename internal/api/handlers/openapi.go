package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/niiniyare/erp/internal/api/gen/openapi"
	"github.com/niiniyare/erp/internal/api/swagger"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// OpenapiHandler implements the GOA openapi service following the data flow pattern
type OpenapiHandler struct{}

// NewOpenapiHandler creates a new GOA openapi handler following Clean Architecture pattern
func NewOpenapiHandler() openapi.Service {
	return &OpenapiHandler{}
}

// Spec returns the OpenAPI specification in JSON format
func (h *OpenapiHandler) Spec(ctx context.Context) (*openapi.SpecResult, error) {
	logger.InfoContext(ctx, "OpenAPI spec requested", logger.Fields{
		"service": "openapi",
		"method":  "spec",
	})

	// Read the generated OpenAPI specification file
	specPath := filepath.Join("gen", "http", "openapi3.json")
	specData, err := os.ReadFile(specPath)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to read OpenAPI spec file", logger.Fields{
			"error": err.Error(),
			"path":  specPath,
		})
		return nil, err
	}

	// Parse the OpenAPI specification
	var spec map[string]any
	if err := json.Unmarshal(specData, &spec); err != nil {
		logger.ErrorContext(ctx, "Failed to parse OpenAPI spec", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	result := &openapi.SpecResult{
		Spec: spec,
	}

	logger.InfoContext(ctx, "OpenAPI spec served successfully", logger.Fields{
		"endpoints_count": len(spec["paths"].(map[string]any)),
	})

	return result, nil
}

// UI redirects to the local Swagger UI
func (h *OpenapiHandler) UI(ctx context.Context) (*openapi.UIResult, error) {
	logger.InfoContext(ctx, "OpenAPI UI requested", logger.Fields{
		"service": "openapi",
		"method":  "ui",
	})

	// Redirect to the locally served Swagger UI
	result := &openapi.UIResult{
		Location: "/swagger-ui/",
	}

	logger.InfoContext(ctx, "Swagger UI redirect provided", logger.Fields{
		"redirect_url": result.Location,
	})

	return result, nil
}

// ServeSwaggerUI creates an HTTP handler for serving the embedded Swagger UI files
func ServeSwaggerUI() http.Handler {
	return swagger.ServeHTTP()
}
