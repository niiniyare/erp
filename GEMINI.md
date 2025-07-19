# Gemini Code-Assist Configuration for Awo ERP

This file provides context and guidelines for interacting with the Awo ERP codebase.

## About the Project

Awo is a multi-tenant ERP system built on a modern, modular architecture. It is designed to be flexible and scalable, supporting various organizational structures and business workflows.

**Key Features:**
- Multi-tenancy with isolated data and configurations.
- Modular design where features can be enabled/disabled per tenant.
- Workflow automation using Temporal.io.
- Goa for API design (REST and gRPC).
- PostgreSQL with RLS for data storage.
- `sqlc` for generating Go typesafe query code.

## Core Principles

1.  **Multi-Tenant First:** All features must be designed with multi-tenancy in mind. Every database query and API endpoint must be scoped to the current tenant context, passed via `context.Context`.
2.  **Clean Architecture:** Adhere to the established data flow (Handler → Service → Repository).
3.  **Modularity:** Keep features self-contained and loosely coupled.
4.  **Database is the Source of Truth:** Business logic should be enforced at the database level where possible (e.g., constraints, RLS policies). `sqlc` is used to generate a type-safe data access layer from raw SQL.
5.  **API-Driven:** All functionality should be exposed via the Goa-defined APIs.

## Data Flow Architecture

The application follows a strict Clean Architecture pattern. Understand this flow before implementing features.

`Client → Handler → Service → Repository → Database`

1.  **Handler (`/internal/api/handlers`)**:
    *   Receives HTTP requests.
    *   Validates request payloads (e.g., JSON).
    *   Calls the appropriate `Service` method, passing the `context.Context`.
    *   Formats the response (success or error) and sends it to the client.

2.  **Service (`/internal/core/*`)**:
    *   Contains the core business logic, validation, and rules.
    *   Orchestrates data access by calling `Repository` methods.
    *   Handles caching logic (check, set, invalidate).
    *   Should be framework-agnostic.

3.  **Repository (`/internal/repository`)**:
    *   Acts as an abstraction layer over the database.
    *   Converts domain models (used by the Service) to `sqlc`-generated models (used by the Store).
    *   Calls the `sqlc`-generated `Store` to execute database queries.
    *   Handles database transaction management (`WithTx`).
    *   Extracts `tenant_id` from the context to enforce data isolation.

4.  **Store (`/db/sqlc`)**:
    *   This is the `sqlc`-generated Data Access Layer (DAL).
    *   **Do not edit manually.**
    *   Provides type-safe methods to execute the SQL queries defined in `/db/queries`.

## Development Workflow

When asked to add a feature, follow this specific workflow:

1.  **API Design (`/internal/api/design`):** Define or modify API endpoints, payloads, and services using the Goa DSL.
2.  **Generate API Code:** Run `goa gen` to generate server and client code into the `/gen` directory.
3.  **Database Migration (`/db/migration`):** If the schema changes, create a new `*.up.sql` and `*.down.sql` migration file.
4.  **SQL Queries (`/db/queries`):** Write or update the necessary SQL queries for `sqlc`. Name them clearly (e.g., `GetUserByID`, `CreateTenant`).
5.  **Generate DAL:** Run `sqlc generate` to update the `Store` in `/db/sqlc`.
6.  **Implement Repository Logic (`/internal/repository`):** Create or update repository methods that call the newly generated `Store` functions. Handle the conversion between domain models and `sqlc` models here.
7.  **Implement Service Logic (`/internal/core/*`):** Implement the core business logic in the service layer, calling the new repository methods.
8.  **Implement Handler Logic (`/internal/api/handlers`):** Wire the service methods into the `goa`-generated handler stubs.
9.  **Testing:** Add unit and integration tests for the new functionality.

Do not write code directly in the `gen` or `db/sqlc` directories, as they are auto-generated.
