// Package design serves as a centralized repository for common and domain-specific types that can be reused across the entire API ecosystem.
// Key responsibilities include:
//
//   - Common Validation Patterns: Standard regex patterns for UUID, email, phone,
//     currency codes, and other frequently validated data formats
//
//   - Domain Entity Definitions: Core business objects and data structures that
//     represent the application's domain model
//
//   - API Contract Types: Request/response structures, pagination models, error formats,
//     and other API-specific types
//
//   - Cross-Cutting Concerns: Audit fields, localization types, security headers,
//     and other shared functionality
//
// • gRPC Integration: gRPC-specific metadata headers and protocol support types
//
// This package ensures consistency across all API endpoints by providing a single
// source of truth for type definitions, validation rules, and data structures.
package design
