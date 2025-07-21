### **Revised Refactoring Plan (TODO_v2.md)**

**Objective:** Complete the decomposition of the `user` god package into domain-focused packages, accounting for the partially started state.

---

#### **Phase 0: Pre-Migration Analysis & Preparation**

*   **Goal:** Understand current state and prepare for safe migration.
*   **Action Plan:**
    - [x] **Analyze Dependencies:** Map imports and cross-references between files to understand migration order
    - [x] **Compare Files:** Check if existing stub files have important content vs. user package files
    - [x] **Backup State:** Create git branch for safe rollback if needed
    - [x] **Validate Structure:** Ensure all target directories exist and are properly structured

#### **Phase 1: Incremental Domain Migration**

*   **Goal:** Move files domain by domain, maintaining compilation at each step using `git mv` to preserve history.
*   **Migration Order (by dependency level):**
    
    **1a. Foundation Services (No Dependencies):**
    - [x] `audit_service.go` -> `internal/core/audit/service.go`
    - [x] `notification_service.go` -> `internal/core/notification/service.go`
    
    **1b. Identity Domain (Core Foundation):**
    - [x] `auth.go` -> `internal/core/identity/auth.go`
    - [x] Core user models from `model.go` -> `internal/core/identity/model.go`
    - [x] `repository.go` -> `internal/core/identity/repository.go`
    - [x] `service.go` -> `internal/core/identity/service.go`
    - [x] `service_test.go` -> `internal/core/identity/service_test.go`
    
    **1c. Access Domain (Depends on Identity):**
    - [x] `permission_cache*.go` -> `internal/core/access/permission/`
    - [x] `access_request_*.go` -> `internal/core/access/request/`
    - [x] `approver_service.go` -> `internal/core/access/approval/service.go`
    - [x] `access_execution_service.go` -> `internal/core/access/execution/service.go`
    - [x] `conditional_access_service.go` -> `internal/core/access/conditional/service.go`
    
    **1d. Analytics Domain (Depends on Audit + Identity):**
    - [x] `user_analytics_service.go` -> `internal/core/analytics/service.go`

---

#### **Phase 2: Domain-by-Domain Refactoring**

*   **Goal:** For each migrated domain package, update its code to be self-contained and compilable following Clean Architecture patterns.
*   **Per-Domain Action Plan:**
    - [x] **Update Package Declaration:** Change `package user` to the new package name (e.g., `package identity`)
    - [x] **Extract Domain Models:** Move relevant structs and interfaces from the large `model.go` into the domain package
    - [x] **Fix Imports:** Update all import paths within the package to reflect the new directory structure
    - [x] **Refactor Interfaces:** Create focused interfaces for each domain (e.g., `identity.Repository`, `access.RequestService`)
    - [x] **Update Constructors:** Create new `NewService` and `NewRepository` functions with proper dependency injection
    - [x] **Compile Verification:** Run `go build ./internal/core/[domain]` to ensure each package compiles independently
    - [x] **Test Migration:** Move and update relevant tests, ensuring they pass

*   **Domain Processing Order:**
    - [x] **audit** → **notification** (Foundation services) ✅ COMPLETED
    - [x] **identity** (Core domain, needed by others) ✅ COMPLETED
    - [x] **access/permission** → **access/request** → **access/approval** → **access/execution** → **access/conditional** ✅ COMPLETED
    - [x] **analytics** (Depends on audit + identity) ✅ COMPLETED

---

#### **Phase 3: Cross-Package Integration**

*   **Goal:** Ensure all packages work together and maintain Clean Architecture dependency rules.
*   **Action Plan:**
    - [x] **Update Import Paths:** Find and update all imports across the codebase (`internal/api`, `cmd/server`, etc.) ✅ COMPLETED
    - [x] **Dependency Injection Refactor:** Update `cmd/server/main.go` to initialize and wire all new services ✅ COMPLETED
    - [x] **Interface Compliance:** Ensure all packages implement their interfaces correctly ✅ COMPLETED
    - [x] **Circular Dependency Check:** Verify no circular dependencies exist between domains ✅ COMPLETED
    - [x] **Created Shared Types Package:** Consolidated duplicate types into `/internal/shared/types` ✅ COMPLETED
    - [x] **Integration Testing:** Run `go test ./...` to ensure all tests pass with new structure ✅ COMPLETED
    - [x] **Fix Remaining Dependencies:** Complete refactoring of entity, tenant, and middleware packages ✅ COMPLETED

---

#### **Phase 4: Final Verification & Cleanup**

*   **Goal:** Complete the refactoring with full system verification.
*   **Action Plan:**
    - [x] **Full Build Verification:** Run `go mod tidy`, `go build ./...`, and `go test ./...` ✅ COMPLETED
    - [x] **API Testing:** Test key API endpoints to ensure functionality is preserved ✅ COMPLETED
    - [ ] **Performance Check:** Verify no performance regressions from restructuring
    - [ ] **Documentation Update:** Update any references to old package structure
    - [ ] **Remove User Package:** Delete the now-empty `internal/core/user` directory
    - [ ] **Final Commit:** Stage all changes with comprehensive commit message following Clean Architecture principles

---

#### **Rollback Strategy**

*   **Safety Measures:**
    - [x] Create feature branch before starting: `git checkout -b refactor/user-package-decomposition`
    - [x] Commit after each major phase for granular rollback points ✅ COMPLETED
    - [x] Keep `ft/user-refactor` branch stable throughout refactoring ✅ COMPLETED
    - [x] Test compilation after each domain migration ✅ COMPLETED

---

#### **Success Criteria**

- [x] All files moved from `internal/core/user` to appropriate domain packages ✅ COMPLETED
- [x] Each domain package compiles independently ✅ COMPLETED
- [x] All tests pass (`go test ./...`) ✅ COMPLETED
- [x] Full application builds (`go build ./...`) ✅ COMPLETED
- [x] No circular dependencies between domains ✅ COMPLETED
- [x] Clean Architecture dependency rules maintained (higher layers → lower layers) ✅ COMPLETED
- [x] API functionality preserved (integration tests pass) ✅ COMPLETED
