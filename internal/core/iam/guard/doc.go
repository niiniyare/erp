// Package guard provides compile-time and test-time enforcement of the IAM
// contract boundary.
//
// Rules it enforces:
//   - No non-IAM package may import "awo.so/internal/core/iam" directly
//     (only "awo.so/internal/core/iam/contract" is permitted).
//   - No non-IAM Go file may contain a field of type iam.Service or
//     iam.SessionService.
//   - No source file outside internal/core/iam/ may call .Enforce( on a
//     live code path (comment-only occurrences are separately flagged).
//   - shared/context.go must not define Permissions / Features / Modules
//     maps on any struct (shadow authorization system).
//
// Tests in this package run as part of the standard `go test ./...` suite.
// They perform static AST/string analysis of the source tree — no DB or
// external service required.
package guard
