package middleware

// This file is replaced by jwt_auth.go and authorization.go
// which provide comprehensive authentication and authorization
// middleware specifically designed for GOA framework integration

// The new middleware structure:
// - jwt_auth.go: Handles JWT, Basic Auth, and API Key authentication using IAM service
// - authorization.go: Handles permission and role-based authorization using IAM service
// - rate_limit.go: Provides comprehensive rate limiting
// - security_headers.go: Adds security headers
// - validation.go: Input validation and sanitization
// - tenant.go: Multi-tenant context handling (existing)

// Usage example:
// 
// // In your GOA service initialization:
// jwtMiddleware := middleware.NewJWTAuthMiddleware(iamService, logger, metrics, tracer)
// authzMiddleware := middleware.NewAuthorizationMiddleware(iamService, config, logger, metrics, tracer)
// 
// // For GOA security middleware:
// service.Use(jwtMiddleware.JWTAuth)
// service.Use(authzMiddleware.RequirePermission("resource", "action"))
// 
// // For HTTP middleware:
// handler.Use(jwtMiddleware.HTTPJWTMiddleware())
// handler.Use(authzMiddleware.HTTPMiddleware())