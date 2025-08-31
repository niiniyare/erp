# 2. Core Concepts & Architecture

To use the Feature Flag system effectively, it's important to understand its core components and how they fit together.

## Core Concepts

### 1. Feature Flag
A **Feature Flag** is the central entity in the system. It's a configuration record that represents a single, controllable feature. Each flag has several key properties:
-   **Name**: A unique, human-readable identifier (e.g., `enable-new-dashboard`).
-   **Type**: The kind of value the flag holds (`boolean`, `string`, `number`, `json`). Currently, the system is primarily focused on `boolean` flags.
-   **Default Value**: The "off" state of the flag. For a boolean flag, this is typically `false`.
-   **Enabled Status**: The "on" state of the flag. This is the value that will be served if a user falls within the rollout percentage.

### 2. Rollout Percentage
This is a simple but powerful mechanism for progressive delivery. It's a value from 0 to 100 that determines what percentage of users should have the feature enabled. The system uses a consistent hashing algorithm based on a user's ID to ensure that a specific user will consistently either see or not see the feature, preventing a flickering user experience.

### 3. Targeting
While the current implementation (`simple_service`) focuses on percentage-based rollouts, the database schema and design are built to support more advanced **Targeting**. This involves enabling a feature for a specific group of users based on their attributes. The `target_audience` JSONB column in the `feature_flags` table is designed to hold these rules, such as:
-   `"user_roles": ["admin", "manager"]`
-   `"subscription_plan": "enterprise"`
-   `"region": "EMEA"`

## System Architecture & Data Flow

The Feature Flag module follows the same Clean Architecture principles used throughout the Awo ERP system. This ensures a clear separation of concerns and a predictable data flow.

**Flow:** `API Handler` -> `Service` -> `Repository` -> `Database (SQLC)`

1.  **API Layer (`internal/api/`)**
    *   **Goa DSL (`internal/api/design/`)**: This is where the API contract is defined. The `admin.go` file specifies the administrative endpoints like `/bulk/enable` and `/health`.
    *   **Handlers (`internal/api/handlers/`)**: (Note: These are planned for a future phase). These will be the Go methods that receive HTTP requests, perform initial validation, and call the appropriate service method.

2.  **Service Layer (`internal/core/featureflag/`)**
    *   This is the heart of the module, containing all the business logic.
    *   **`service_simple.go`**: The primary service implementation. It's responsible for validating tenant context and orchestrating the creation, update, and evaluation of flags.
    *   **`service_cached.go`**: A decorator that wraps the `simpleServiceImpl`. It introduces a Redis caching layer to dramatically speed up flag evaluations and reduce database load. It intercepts calls, checks the cache, and if necessary, calls the underlying `simpleServiceImpl` before caching the result.
    *   **`admin_service.go`**: Defines the interface for all administrative functions. The implementation files (`admin_bulk_operations.go`, `admin_system_management.go`, etc.) contain the logic for the admin-facing features.

3.  **Repository Layer (`internal/core/featureflag/repository_simple.go`)**
    *   This layer is the bridge between the service's business logic and the database.
    *   It implements the `SimpleRepository` interface, providing methods like `CreateFeatureFlag` and `GetFeatureFlagByName`.
    *   Its only job is to execute database queries using the `sqlc`-generated `Store`. It knows nothing about business rules; it just fetches and stores data.

4.  **Database Layer (`db/`)**
    *   **Schema (`db/base.sql` & Migrations)**: The `feature_flags` table is defined here. It's the source of truth for all flag configurations. Key columns include `name`, `default_value`, `rollout_percentage`, and the `target_audience` and `metadata` JSONB fields for future enhancements. Crucially, it is tenant-isolated via a `tenant_id` foreign key and a Row-Level Security (RLS) policy.
    *   **Queries (`db/queries/feature_flags_simple.sql`)**: This file contains the raw SQL queries that `sqlc` uses. Writing queries here allows us to keep SQL separate from our Go code.
    *   **`sqlc` (`db/sqlc/`)**: This is the auto-generated Go code that provides a type-safe interface for executing the queries defined in the `queries` directory. The application code **always** calls these generated methods, never raw SQL.
