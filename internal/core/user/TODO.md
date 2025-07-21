# Refactoring Plan: Decomposing the `user` God Package

This document outlines the plan to refactor the oversized `user` package into smaller, domain-focused packages based on Clean Architecture principles. The goal is to improve modularity, reduce coupling, and enhance the overall maintainability of the codebase.

## 1. New Domain-Driven Structure

The current `user` package will be decomposed into the following new packages under `internal/core/`:

- **`internal/core/identity`**: Core Identity Management.
  - **Responsibility**: Manages the core concepts of a user's identity.
  - **Key Components**: `User`, `Person`, `Employee` models, `AuthService` for authentication, and a repository for basic identity CRUD.

- **`internal/core/access`**: A new parent package for all Authorization and Access Control.
  - **`access/request`**: The complete access request workflow.
  - **`access/permission`**: Core permission/role models and the evaluation engine.
  - **`access/approval`**: Logic for determining approvers for access requests.
  - **`access/execution`**: Logic for granting/revoking access after approval.
  - **`access/conditional`**: Logic for conditional access policies (e.g., time, location).

- **`internal/core/audit`**: Centralized Auditing Service.
  - **Responsibility**: Provides a single, reliable service for logging all security, access, and workflow events.

- **`internal/core/notification`**: Centralized Notification Service.
  - **Responsibility**: Manages sending notifications via various channels (email, Slack, etc.) for system events.

- **`internal/core/analytics`**: User Behavior Analytics.
  - **Responsibility**: Analyzes user activity to detect anomalies, assess risk, and provide insights.

## 2. Domain Interaction Model

The new services will interact in a clear, dependency-managed way. The core principle is that foundational services (like `identity`) should not depend on higher-level business logic services (like `access/request`).

```
[ High-Level Services ]
       |
       v
[ Core Business Logic ]
       |
       v
[ Foundational Services ]
       |
       v
[ Core Identity & Platform ]
```

**Dependency Flow:**

- **`access/request` (Access Request Service)**
  - `->` **`identity.Service`**: To get details of the requester, target user, and approver.
  - `->` **`access/approval.Service`**: To determine the list of eligible approvers for a request.
  - `->` **`access/execution.Service`**: To trigger the grant/revoke of access once a request is approved.
  - `->` **`notification.Service`**: To send notifications to requesters and approvers.
  - `->` **`audit.Service`**: To log all workflow events (creation, approval, rejection).

- **`access/execution` (Access Execution Service)**
  - `->` **`identity.Service`**: To perform the final assignment of roles/permissions to a user.
  - `->` **`audit.Service`**: To log the successful grant or revocation of privileges.

- **`access/approval` (Approver Service)**
  - `->` **`identity.Service`**: To fetch user hierarchy, roles, and other attributes needed to evaluate approval rules.
  - `->` **`access/permission.Service`**: To check if a potential approver has the necessary permissions to approve a request.

- **`analytics.Service` (User Analytics Service)**
  - `->` **`audit.Service`**: To get the raw stream of user activities and events for analysis.
  - `->` **`identity.Service`**: To enrich analytics data with user details.

- **`identity.Service` (Identity Service)**
  - **(No dependencies)** on other new business-level domains. This is a foundational service.

- **`audit.Service` & `notification.Service`**
  - **(Few to no dependencies)**. These are cross-cutting concerns designed to be used by other services.

## 3. Refactoring Checklist

This checklist will be followed to execute the refactoring systematically.

- [ ] **1. Create New Directories:**
  - `internal/core/identity`
  - `internal/core/access/request`
  - `internal/core/access/permission`
  - `internal/core/access/approval`
  - `internal/core/access/execution`
  - `internal/core/access/conditional`
  - `internal/core/audit`
  - `internal/core/notification`
  - `internal/core/analytics`

- [ ] **2. Move Files:** Move each file from `internal/core/user` to its new corresponding package directory.

- [ ] **3. Update Package Declarations:** Change the `package user` line in each moved file to its new package name (e.g., `package identity`, `package request`).

- [ ] **4. Split Models:** Move domain-specific models into their respective packages. For example, `AccessRequest` goes to `access/request`, while `User` and `Person` go to `identity`.

- [ ] **5. Split Interfaces:** Decompose the monolithic `Service` and `Repository` interfaces into smaller, more focused interfaces within each new package.

- [ ] **6. Update Imports:** Systematically update all import paths across the entire project to point to the new locations of types and functions. This will be the most extensive step.

- [ ] **7. Refactor Constructors:** Create a `NewService` and `NewRepository` for each new domain, ensuring dependencies are injected according to the interaction model described above.

- [ ] **8. Update Dependency Injection Root:** Modify the main application setup (e.g., in `cmd/server/main.go`) to initialize and inject the new services correctly.

- [ ] **9. Compile and Test:** Run `go mod tidy`, compile the entire project, and run all tests (`go test ./...`) to ensure the refactoring has not introduced regressions.

- [ ] **10. Final Cleanup:** Once all tests pass and the application runs correctly, delete the now-empty `internal/core/user` directory.
