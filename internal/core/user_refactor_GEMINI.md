# Gemini Code-Assist: Refactoring `internal/core/user`

This document provides the context and step-by-step plan to complete the refactoring of the `internal/core/user` "god package" into smaller, domain-focused packages.

## 1. Objective

The primary goal is to finalize the decomposition of the `user` package as outlined in `internal/core/user/TODO_v2.md`. The file structure has been created and files have been moved, but the code is in a non-compilable state. The task is to refactor the code within the new packages to make them self-contained, update all dependencies, and ensure the entire application compiles and passes all tests.

## 2. Current State

- **Files Moved**: All files from `internal/core/user` have been moved to their respective new domain packages (`identity`, `access/*`, `audit`, etc.) using `git mv`.
- **Code Not Refactored**: The code within the moved files has not been updated. Package declarations are still `package user`, imports are broken, and interfaces/models have not been split.
- **Application is Broken**: The project will not compile in its current state.

## 3. Refactoring Workflow

Follow this workflow systematically. The process is divided into three main phases: refactoring each domain, integrating them, and final cleanup.

### Phase 1: Domain-by-Domain Refactoring

For each new domain package, perform the following steps. Process the domains in the specified order to manage dependencies effectively.

**Processing Order:**
1.  `internal/core/audit`
2.  `internal/core/notification`
3.  `internal/core/identity`
4.  `internal/core/access/permission`
5.  `internal/core/access/approval`
6.  `internal/core/access/execution`
7.  `internal/core/access/conditional`
8.  `internal/core/access/request`
9.  `internal/core/analytics`

**Per-Domain Checklist:**
1.  **Navigate to the directory**: `cd internal/core/{domain}`
2.  **Update Package Declaration**: In all `.go` files, change `package user` to the correct package name (e.g., `package audit`, `package identity`).
3.  **Isolate Domain Models**: Identify structs and interfaces specific to this domain in the old `model.go` (which is now likely copied in multiple places or needs to be created) and consolidate them into a `model.go` file within the domain. Remove unrelated models.
4.  **Fix Imports**: Update all import paths within the package to point to the new locations of dependencies.
5.  **Refactor Interfaces**: Create focused `Service` and `Repository` interfaces for the domain.
6.  **Update Constructors**: Create or update `NewService` and `NewRepository` functions, ensuring they accept the correct dependencies via their interfaces.
7.  **Compile & Test**:
    - Run `go build .` within the package directory to ensure it compiles independently.
    - Move relevant tests from the old `user` package tests and update them. Run `go test .` to verify correctness.

### Phase 2: Cross-Package Integration

Once all individual domains are refactored and compiling, integrate them into the larger application.

1.  **Update Upstream Imports**: Search the entire codebase (especially `internal/api/handlers`, `cmd/server`, `cmd/worker`) for imports pointing to the old `internal/core/user` and update them to the new domain packages.
2.  **Refactor Dependency Injection**: Modify `cmd/server/main.go` (and other application entrypoints) to initialize and inject all the new services correctly. The dependency graph from `TODO.md` should be followed.
3.  **Check for Circular Dependencies**: Run a full build (`go build ./...`) to detect any circular dependencies introduced during the refactoring. Resolve them by adjusting service interfaces or dependencies.

### Phase 3: Final Verification & Cleanup

1.  **Run Full Verification Suite**:
    - `go mod tidy`
    - `go build ./...`
    - `go test ./...`
2.  **API Smoke Test**: Manually or with scripts, test the primary API endpoints to ensure the refactored services are wired correctly and the application is functional.
3.  **Update Documentation**: Search the `/docs` directory for any references to the old `internal/core/user` package and update them.
4.  **Delete Old User Package**: Once confident that the refactoring is complete and stable, delete the now-empty `internal/core/user` directory.
5.  **Commit**: Create a comprehensive commit message summarizing the refactoring.

## 4. Guiding Principles

- **Clean Architecture**: Strictly adhere to the dependency rule: `Handlers -> Services -> Repositories`. Foundational services (`identity`) should not depend on higher-level services (`access`).
- **Single Responsibility**: Each package should have a single, well-defined responsibility.
- **Dependency Inversion**: Depend on interfaces, not concrete implementations, especially for cross-domain communication.
