package handlers

import (
	"context"

	"github.com/niiniyare/erp/gen/openapi"
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

	// TODO: Return actual OpenAPI specification
	// This should be generated from the GOA design files
	result := &openapi.SpecResult{
		Spec: map[string]interface{}{
			"openapi": "3.0.0",
			"info": map[string]interface{}{
				"title":   "AWO ERP System API",
				"version": "1.0.0",
			},
			"paths": map[string]interface{}{},
		},
	}

	return result, nil
}

// UI redirects to the Swagger UI
func (h *OpenapiHandler) UI(ctx context.Context) (*openapi.UIResult, error) {
	logger.InfoContext(ctx, "OpenAPI UI requested", logger.Fields{
		"service": "openapi",
		"method":  "ui",
	})

	result := &openapi.UIResult{
		Location: "/swagger-ui/",
	}

	return result, nil
}
