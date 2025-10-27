# See https://tech.davis-hansson.com/p/make/
# ============================================================================
# 🧱 Environment & Defaults
# ============================================================================

SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables --no-builtin-rules --no-print-directory
.DEFAULT_GOAL := help

# ============================================================================
# 📋 Variables & Configuration
# ============================================================================

# Database Configuration
DB_NAME ?= ledger
DB_USER ?= admin
DB_PSSWD ?= admin
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_URL ?= postgresql://$(DB_USER):$(DB_PSSWD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Paths
MIGRATION_PATH := db/migration
SQLC_OUT := db/sqlc
DOCS_PATH := docs
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Ports
DOC_PORT ?= 8081
GRPC_PORT ?= 9090

# Tools
REQUIRED_TOOLS := sqlc mockgen migrate psql golangci-lint mkdocs java sleek awoctl

# Test Configuration
TEST_TIMEOUT := 10m
TEST_TIMEOUT_FAST := 5m
TEST_TIMEOUT_E2E := 15m
TEST_TIMEOUT_INTEGRATION := 8m
DOCKER_COMPOSE_TEST := docker-compose.test.yml
TEST_PARALLEL_JOBS := 4

# Test Directories
UNIT_TEST_DIRS := ./internal/core/... ./internal/shared/... ./internal/adapters/...
INTEGRATION_TEST_DIRS := ./test/integration/...
E2E_TEST_DIRS := ./test/e2e/...


# ====== SERVER LOG ======
LOG_FILE=server.log
PID_FILE=server.pid

# Colors for output
ESC := $(shell printf '\033')
RED := $(ESC)[0;31m
GREEN := $(ESC)[0;32m
YELLOW := $(ESC)[0;33m
BLUE := $(ESC)[0;34m
PURPLE := $(ESC)[0;35m
CYAN := $(ESC)[0;36m
NC := $(ESC)[0m

# ============================================================================
# 🆘 Help & Utilities
# ============================================================================

.PHONY: help
help: ## 📚 Show this help message
	@echo "$(CYAN)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} \
	/^[a-zA-Z0-9_.-]+:.*?##/ { \
		gsub(/^[ \t]+/, "", $$2); \
		printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)
	@echo ""

.PHONY: check-tools
check-tools: ## 🔧 Check if required tools are installed
	@echo "$(BLUE)Checking required tools...$(NC)"
	@$(foreach tool,$(REQUIRED_TOOLS),\
		command -v $(tool) >/dev/null 2>&1 || { \
			echo "$(RED)❌ Missing: $(tool)$(NC)"; exit 1; \
		};)
	@echo "$(GREEN)✅ All required tools are installed$(NC)"


.PHONY: status
status: ## 📊 Show project status and configuration
	@echo "$(CYAN)Project Configuration:$(NC)"
	@echo "  Database: $(DB_USER)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	@echo "  Migration Path: $(MIGRATION_PATH)"
	@echo "  Documentation Port: $(DOC_PORT)"
	@echo "  Test Timeout: $(TEST_TIMEOUT)"
	@echo "  Parallel Jobs: $(TEST_PARALLEL_JOBS)"

# ============================================================================
# 🏗️ Build & Generation
# ============================================================================

.PHONY: generate
generate: sqlc mock wire ## 🔄 Generate all code (SQLC, Mocks, Wire)

# ============================================================================
# 🏗️ Scaffolding & Code Generation
# ============================================================================

.PHONY: scaffold-install
scaffold-install: ## 🔧 Build and install awoctl scaffolding tool
	@echo "$(BLUE)Building awoctl scaffolding tool...$(NC)"
	@go build -o awoctl ./cmd/awoctl/
	@echo "$(GREEN)✅ awoctl built successfully$(NC)"

.PHONY: scaffold-module
scaffold-module: ## 📦 Generate new module (usage: make scaffold-module name=moduleName [docs=true] [tests=true])
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌ Error: module name required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-module name=inventory$(NC)"; \
		echo "$(YELLOW)Options: docs=true tests=true$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating module: $(name)$(NC)"
	@cmd="./awoctl new module $(name) --verbose"; \
	if [ "$(docs)" = "true" ]; then cmd="$$cmd --with-docs"; fi; \
	if [ "$(tests)" = "true" ]; then cmd="$$cmd --with-tests"; fi; \
	eval $$cmd
	@echo "$(GREEN)✅ Module '$(name)' generated successfully$(NC)"
	@if [ "$(docs)" = "true" ]; then \
		echo "$(GREEN)📚 Documentation generated at: docs/reference/modules/$(name)/$(NC)"; \
	fi
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Update domain models in internal/core/$(name)/domain/"
	@echo "  2. Run 'make sqlc goa' to generate code"
	@echo "  3. Implement business logic"

.PHONY: scaffold-component
scaffold-component: ## 🧩 Generate component (usage: make scaffold-component module=finance name=payment type=entity)
	@if [ -z "$(module)" ] || [ -z "$(name)" ] || [ -z "$(type)" ]; then \
		echo "$(RED)❌ Error: module, name, and type are required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-component module=finance name=payment type=entity$(NC)"; \
		echo "$(YELLOW)Types: entity, service, repository, handler, dto, middleware, validator, mapper$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating $(type) component: $(name) in module $(module)$(NC)"
	@./awoctl new component $(module) $(name) $(type) --verbose
	@echo "$(GREEN)✅ Component '$(name)' generated successfully$(NC)"

.PHONY: scaffold-feature
scaffold-feature: ## ⭐ Generate feature (usage: make scaffold-feature path=finance/budgets)
	@if [ -z "$(path)" ]; then \
		echo "$(RED)❌ Error: feature path required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-feature path=finance/budgets$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating feature: $(path)$(NC)"
	@./awoctl new feature $(path) --verbose
	@echo "$(GREEN)✅ Feature '$(path)' generated successfully$(NC)"

.PHONY: scaffold-docs
scaffold-docs: ## 📖 Generate documentation for existing module (usage: make scaffold-docs name=moduleName)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌ Error: module name required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-docs name=inventory$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating documentation for module: $(name)$(NC)"
	@./awoctl docs module $(name) --verbose
	@echo "$(GREEN)✅ Documentation for '$(name)' generated successfully$(NC)"
	@echo "$(GREEN)📚 Documentation available at: docs/reference/modules/$(name)/$(NC)"

.PHONY: scaffold-help
scaffold-help: ## 📚 Show scaffolding tool help
	@./awoctl --help

.PHONY: sqlc-install
sqlc-install: ## Install sqlc excutable of if it's not in the system
	@echo "$(BLUE)Installing SQLC ...$(NC)"
	@CGO_CFLAGS="-D_GNU_SOURCE" go install -v github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: sqlc
sqlc: ## 🗄️ Generate SQLC store code
	@echo "$(BLUE)Generating SQLC code...$(NC)"
	@sqlc generate &&
	@echo "$(GREEN)✅ SQLC generation complete$(NC)"
	@echo "$(BLUE) Mock Generation for db.Store interface...$(NC)"
	@ go generate ./db/sqlc/...
	@echo "$(GREEN)✅ Mock generation complete$(NC)"

.PHONY: sqlc-lint
sqlc-lint: ## 🔍 Lint SQL queries
	@./db/queries/lint.sh


.PHONY: templ
templ: 
	@echo "$(BLUE)Generating Templ code...$(NC)"
	@templ generate ./web/...
	@echo "$(GREEN)Templ code generation complete$(NC)"
	# @echo "$(BLUE)compilig TypeScript...$(NC)"
	# @tsc --noEmit
	# @echo "$(GREEN)TypeScript compilation successful!$(NC)"



.PHONY: templ-fmt 
templ-fmt:
	@echo "$(BLUE)Formatting Templ code...$(NC)"
	@find . -type f -name "*.templ" -exec templ fmt {} +
.PHONY: mock
mock: ## 🎭 Generate mocks for interfaces
	@echo "$(BLUE)Generating mocks...$(NC)"
	@go generate ./...
	@echo "$(GREEN)✅ Mocks generated$(NC)"

.PHONY: wire-install
wire-install: ## 📦 Install Google Wire
	@echo "$(BLUE)Installing Google Wire...$(NC)"
	@go install github.com/google/wire/cmd/wire@latest
	@echo "$(GREEN)✅ Wire installed$(NC)"

.PHONY: wire
wire: ## ⚡ Generate Wire dependency injection code
	@echo "$(BLUE)Generating Wire code...$(NC)"
	@./scripts/generate-wire.sh
	@echo "$(GREEN)✅ Wire generation complete$(NC)"

.PHONY: wire-check
wire-check: ## 🔍 Check if Wire files are up to date
	@echo "$(BLUE)Checking Wire files...$(NC)"
	@cd cmd/server && wire check
	@echo "$(GREEN)✅ Wire files are up to date$(NC)"

.PHONY: goa
goa: ## 🎯 Generate Goa code
	@echo "$(BLUE)Generating Goa code...$(NC)"
	@goa gen github.com/niiniyare/erp/internal/api/design -o internal/api
	@rm -rf ./internal/api/swagger/openapi
	@cp -f ./internal/api/gen/http/openapi3.json ./internal/api/swagger
	@echo "$(GREEN)✅ Goa generation complete$(NC)"

.PHONY: interface2any
interface2any: ## 🔄 Convert interfaces to any
	@find . -type f -name '*.go' | xargs sed -i 's/interface{}/any/g'
	# @./scripts/interface_any.sh

# ============================================================================
# 🧪 Testing Framework
# ============================================================================

.PHONY: test
test: test-unit ## 🧪 Run default tests (unit tests)

# ============================================================================
# 🎯 Core Test Suites
# ============================================================================

.PHONY: test-unit
test-unit: ## 🎯 Run unit tests with race detection
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test -tags=unit -race -timeout=$(TEST_TIMEOUT) -v \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	@echo "$(GREEN)✅ Unit tests passed$(NC)"

.PHONY: test-unit-fast
test-unit-fast: ## ⚡ Run unit tests without race detection (faster)
	@echo "$(BLUE)Running unit tests (fast mode)...$(NC)"
	@go test -tags=unit -timeout=$(TEST_TIMEOUT_FAST) \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	@echo "$(GREEN)✅ Fast unit tests passed$(NC)"

.PHONY: test-unit-short
test-unit-short: ## 🏃 Run unit tests with short flag (skip slow tests)
	@echo "$(BLUE)Running unit tests (short mode)...$(NC)"
	@go test -tags=unit -short -timeout=$(TEST_TIMEOUT_FAST) \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	@echo "$(GREEN)✅ Short unit tests passed$(NC)"

.PHONY: test-integration
test-integration: ## 🔗 Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	@echo "🐳 Starting dependencies..."
	@docker-compose -f $(DOCKER_COMPOSE_TEST) up -d postgres redis || true
	@sleep 5
	@go test -tags=integration -timeout=$(TEST_TIMEOUT_INTEGRATION) \
		-parallel=$(TEST_PARALLEL_JOBS) -v $(INTEGRATION_TEST_DIRS) || \
		(docker-compose -f $(DOCKER_COMPOSE_TEST) down; exit 1)
	@echo "🧹 Cleaning up dependencies..."
	@docker-compose -f $(DOCKER_COMPOSE_TEST) down
	@echo "$(GREEN)✅ Integration tests passed$(NC)"

.PHONY: test-e2e
test-e2e: ## 🌐 Run end-to-end tests
	@echo "$(BLUE)Running E2E tests...$(NC)"
	@echo "🐳 Starting full environment..."
	@docker-compose -f $(DOCKER_COMPOSE_TEST) up -d || true
	@sleep 10
	@go test -tags=e2e -timeout=$(TEST_TIMEOUT_E2E) -v $(E2E_TEST_DIRS) || \
		(docker-compose -f $(DOCKER_COMPOSE_TEST) down; exit 1)
	@echo "🧹 Cleaning up environment..."
	@docker-compose -f $(DOCKER_COMPOSE_TEST) down
	@echo "$(GREEN)✅ E2E tests passed$(NC)"

# ============================================================================
# 📊 Test Analysis & Reporting
# ============================================================================

.PHONY: test-coverage
test-coverage: ## 📊 Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -race -timeout=$(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) \
		-covermode=atomic -parallel=$(TEST_PARALLEL_JOBS) ./...
	@go tool cover -func=$(COVERAGE_FILE) | tail -1
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)✅ Coverage report generated: $(COVERAGE_HTML)$(NC)"

.PHONY: test-coverage-unit
test-coverage-unit: ## 📊 Run unit tests with coverage
	@echo "$(BLUE)Running unit tests with coverage...$(NC)"
	@go test -tags=unit -race -timeout=$(TEST_TIMEOUT) \
		-coverprofile=$(COVERAGE_FILE) -covermode=atomic \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	@go tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo "$(GREEN)✅ Unit test coverage complete$(NC)"

.PHONY: test-coverage-html
test-coverage-html: test-coverage ## 🌐 Generate and open HTML coverage report
	@echo "$(BLUE)Opening coverage report in browser...$(NC)"
	@open $(COVERAGE_HTML) || xdg-open $(COVERAGE_HTML) || echo "Please open $(COVERAGE_HTML) manually"

.PHONY: test-benchmark
test-benchmark: ## 🏃 Run benchmark tests
	@echo "$(BLUE)Running benchmark tests...$(NC)"
	@go test -bench=. -benchmem -timeout=$(TEST_TIMEOUT) \
		-run=^$$ ./... | tee benchmark.out
	@echo "$(GREEN)✅ Benchmarks complete$(NC)"

.PHONY: test-benchmark-compare
test-benchmark-compare: ## 🏃 Run benchmarks and compare with previous results
	@echo "$(BLUE)Running benchmarks for comparison...$(NC)"
	@go test -bench=. -benchmem -count=5 -timeout=$(TEST_TIMEOUT) \
		-run=^$$ ./... | tee benchmark-new.out
	@if [ -f benchmark.out ]; then \
		echo "$(BLUE)Comparing with previous results...$(NC)"; \
		benchcmp benchmark.out benchmark-new.out || echo "benchcmp not available"; \
	fi
	@mv benchmark-new.out benchmark.out

# ============================================================================
# 🎯 Domain-Specific Tests
# ============================================================================

.PHONY: test-identity
test-identity: ## 🆔 Run identity domain tests
	@echo "$(BLUE)Testing identity domain...$(NC)"
	@go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/identity/...

.PHONY: test-access
test-access: ## 🔐 Run access domain tests
	@echo "$(BLUE)Testing access domain...$(NC)"
	@go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/access/...

.PHONY: test-audit
test-audit: ## 📋 Run audit domain tests
	@echo "$(BLUE)Testing audit domain...$(NC)"
	@go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/audit/...

.PHONY: test-notification
test-notification: ## 📢 Run notification domain tests
	@echo "$(BLUE)Testing notification domain...$(NC)"
	@go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/notification/...

.PHONY: test-analytics
test-analytics: ## 📈 Run analytics domain tests
	@echo "$(BLUE)Testing analytics domain...$(NC)"
	@go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/analytics/...

# ============================================================================
# 🚀 Test Suites & Workflows
# ============================================================================

.PHONY: test-all
test-all: ## 🎯 Run complete test suite (unit + integration + e2e)
	@echo "$(PURPLE)Running complete test suite...$(NC)"
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@$(MAKE) test-e2e
	@echo "$(GREEN)✅ All tests completed successfully!$(NC)"

.PHONY: test-quick
test-quick: ## ⚡ Quick test run (unit tests in short mode)
	@echo "$(BLUE)Running quick tests...$(NC)"
	@$(MAKE) test-unit-short

.PHONY: test-ci
test-ci: ## 🚀 CI test pipeline (coverage + all tests)
	@echo "$(PURPLE)Running CI test pipeline...$(NC)"
	@$(MAKE) test-coverage
	@$(MAKE) test-integration
	@echo "$(GREEN)✅ CI test pipeline completed!$(NC)"

.PHONY: test-watch
test-watch: ## 👀 Watch for changes and run tests (requires entr)
	@echo "$(BLUE)Watching for changes (press Ctrl+C to stop)...$(NC)"
	@find . -name "*.go" | entr -c make test-unit-fast

.PHONY: test-failed
test-failed: ## 🔄 Re-run only failed tests from last run
	@echo "$(BLUE)Re-running failed tests...$(NC)"
	@go test -json ./... | tee /tmp/test-output.json
	@cat /tmp/test-output.json | jq -r 'select(.Action=="fail" and .Test) | .Package + "/" + .Test' | \
		xargs -I {} go test -run {} -v

# ============================================================================
# 🧹 Test Maintenance & Utilities
# ============================================================================

.PHONY: test-clean
test-clean: ## 🧹 Clean test artifacts and docker containers
	@echo "$(YELLOW)Cleaning test artifacts...$(NC)"
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) benchmark.out benchmark-new.out
	@docker-compose -f $(DOCKER_COMPOSE_TEST) down -v --remove-orphans || true
	@go clean -testcache -cache
	@echo "$(GREEN)✅ Test cleanup complete$(NC)"

.PHONY: test-deps
test-deps: ## 📦 Install test dependencies
	@echo "$(BLUE)Installing test dependencies...$(NC)"
	@go install github.com/golang/mock/mockgen@latest
	@go install gotest.tools/gotestsum@latest
	@command -v entr >/dev/null || echo "$(YELLOW)Consider installing 'entr' for test-watch$(NC)"
	@command -v benchcmp >/dev/null || echo "$(YELLOW)Consider installing 'benchcmp' for benchmark comparison$(NC)"

.PHONY: test-list
test-list: ## 📋 List all available tests
	@echo "$(BLUE)Available tests:$(NC)"
	@go test -list . ./... 2>/dev/null | grep -E '^Test|^Example|^Benchmark' | sort

.PHONY: test-verbose
test-verbose: ## 🔊 Run tests with verbose output and detailed timing
	@echo "$(BLUE)Running tests with verbose output...$(NC)"
	@go test -v -timeout=$(TEST_TIMEOUT) -parallel=$(TEST_PARALLEL_JOBS) ./... | \
		grep -E "(PASS|FAIL|RUN)"

# ============================================================================
# 🎨 Code Quality
# ============================================================================

.PHONY: fmt
fmt: ## 🎨 Format Go code
	@echo "$(BLUE)Formatting Go code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

.PHONY: lint
lint: ## 📝 Lint Go code
	@echo "$(BLUE)Linting Go code...$(NC)"
	@golangci-lint run ./...
	@echo "$(GREEN)✅ Linting complete$(NC)"

.PHONY: vet
vet: ## 🔍 Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ Vet check complete$(NC)"

.PHONY: security
security: ## 🔒 Run security checks (gosec)
	@echo "$(BLUE)Running security checks...$(NC)"
	@command -v gosec >/dev/null 2>&1 || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec ./...
	@echo "$(GREEN)✅ Security check complete$(NC)"

.PHONY: quality
quality: fmt vet lint sqlc-lint templ-fmt vet  ## 🏆 Run all code quality checks

# ============================================================================
# 🗄️ Database Operations
# ============================================================================

.PHONY: db-create
db-create: ## 🏗️ Create the database
	@echo "$(BLUE)Creating database $(DB_NAME)...$(NC)"
	@createdb --username="$(DB_USER)" --owner="$(DB_USER)" $(DB_NAME)
	@echo "$(GREEN)✅ Database created$(NC)"

.PHONY: db-drop
db-drop: ## 💥 Drop the database
	@echo "$(RED)Dropping database $(DB_NAME)...$(NC)"
	@dropdb $(DB_NAME)
	@echo "$(YELLOW)⚠️ Database dropped$(NC)"

.PHONY: db-reset
db-reset: db-drop db-create migrate-up ## 🔄 Reset database (drop, create, migrate)

.PHONY: migrate-up
migrate-up: ## ⬆️ Run all up migrations
	@echo "$(BLUE)Running up migrations...$(NC)"
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose up
	@echo "$(GREEN)✅ Migrations applied$(NC)"

.PHONY: migrate-down
migrate-down: ## ⬇️ Roll back the last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose down
	@echo "$(YELLOW)⚠️ Migration rolled back$(NC)"

.PHONY: migrate-drop
migrate-drop: ## 💥 Drop all schema objects
	@echo "$(RED)Dropping all schema objects...$(NC)"
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose drop -f
	@echo "$(YELLOW)⚠️ Schema dropped$(NC)"

.PHONY: migrate-create
migrate-create: ## 📝 Create new migration file: make migrate-create name=init_users
	@name=$(name); \
	if [ -z "$$name" ]; then \
		echo "$(RED)Missing name param. Use: make migrate-create name=xyz$(NC)"; exit 1; \
	fi; \
	echo "$(BLUE)Creating migration: $$name$(NC)"; \
	migrate create -ext sql -dir "$(MIGRATION_PATH)" -digits 2 -seq "$$name" -verbose
	@echo "$(GREEN)✅ Migration created$(NC)"

.PHONY: migrate-status
migrate-status: ## 📊 Show migration status
	@echo "$(BLUE)Checking migration status...$(NC)"
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" version

# ============================================================================
# 📚 Documentation
# ============================================================================

.PHONY: docs
docs: ## 📖 Serve documentation on default port
	@echo "$(BLUE)🏢 Starting AWO ERP Documentation Server...$(NC)"
	@echo "📚 MkDocs documentation: http://localhost:$(DOC_PORT)/"
	@echo "🗄️ Schema documentation: http://localhost:$(DOC_PORT)/schema/"
	@cd $(DOCS_PATH) && ./start-docs.sh $(DOC_PORT)

.PHONY: docs-build
docs-build: ## 🏗️ Build MkDocs documentation
	@echo "$(BLUE)📖 Building MkDocs documentation...$(NC)"
	@mkdocs build
	@echo "$(GREEN)✅ Documentation built in site/ directory$(NC)"

.PHONY: docs-dev
docs-dev: docs-build docs ## 👨‍💻 Development docs workflow (build + serve)

.PHONY: docs-schema
docs-schema: ## 🗄️ Generate database schema documentation
	@echo "$(BLUE)🗄️ Generating database schema documentation...$(NC)"
	@echo "Note: Requires SchemaSpy JAR and database connection"
	@$(DOCS_PATH)/scripts/generate-schema-docs.sh
	@echo "$(GREEN)✅ Schema documentation generated$(NC)"

.PHONY: docs-test
docs-test: ## 🧪 Test documentation server
	@echo "$(BLUE)🧪 Testing documentation server...$(NC)"
	@cd $(DOCS_PATH) && ./test-server.sh

.PHONY: dbdocs
dbdocs: ## 📊 Generate DB docs from DBML
	@echo "$(BLUE)Generating database documentation...$(NC)"
	@dbdocs build $(DOCS_PATH)/schema.dbml
	@echo "$(GREEN)✅ Database documentation generated$(NC)"

.PHONY: sql2dbml
sql2dbml: ## 🔄 Convert SQL migration to DBML
	@sql2dbml $(MIGRATION_PATH)/*.up.sql --postgres -o $(DOCS_PATH)/schema.dbml

# ============================================================================
# 🚀 Application & Services
# ============================================================================

# .PHONY: run
# run: ## 🚀 Run the application server
# 	@echo "$(BLUE)Starting application server...$(NC)"
# 	@go run ./cmd/server/

.PHONY: run
run: ## Run the application server (logs printed and saved to server.log)
	@echo "$(BLUE)Starting application server...$(NC)"
	@echo "$(YELLOW)Logs will be written to $(LOG_FILE)$(NC)"
	@go run ./cmd/server/ 2>&1 | tee $(LOG_FILE)


.PHONY: run-bg
run-bg: ##  Run the server in background (Termux-friendly)
	@echo "$(BLUE)Starting server in background...$(NC)"
	@bash -c "go run ./cmd/server/ 2>&1 | tee -a $(LOG_FILE)" &
	@echo $$! > $(PID_FILE)
	@echo "$(GREEN)Server started in background (PID: $$(cat $(PID_FILE)))$(NC)"
	@echo "$(YELLOW)Logs are being written to $(LOG_FILE)$(NC)"


.PHONY: stop
stop:  ## 🛑 Stop the background server
	@if [ -f $(PID_FILE) ]; then \
		echo "$(RED)Stopping server (PID: $$(cat $(PID_FILE)))...$(NC)"; \
		kill $$(cat $(PID_FILE)) 2>/dev/null || true; \
		rm -f $(PID_FILE); \
		echo "$(GREEN)Server stopped.$(NC)"; \
	else \
		echo "$(YELLOW)No running server found.$(NC)"; \
	fi

.PHONY: clean
clean: ## 🧹 Clean up log and pid files 
	@echo "$(RED)Cleaning up logs and pid file...$(NC)"
	@rm -f $(LOG_FILE) $(PID_FILE)
	@echo "$(YELLOW)Cleaning generated files...$(NC)"
	@rm -f doc/swagger/*.json
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) benchmark.out
	@go clean -testcache -cache -modcache
	@echo "$(GREEN)✅ Cleanup complete$(NC)"


.PHONY: build
build: ## 🔨 Build the application
	@echo "$(BLUE)Building application...$(NC)"
	@go build -o bin/server ./cmd/server/
	@echo "$(GREEN)✅ Application built: bin/server$(NC)"

.PHONY: temporal-server
temporal-server: ## ⏰ Start Temporal server
	@echo "$(BLUE)Starting Temporal server...$(NC)"
	@temporal server start-dev

# ============================================================================
# 🔄 Development Workflows
# ============================================================================

.PHONY: dev-setup
dev-setup: check-tools scaffold-install db-create migrate-up generate test-deps ## 🛠️ Setup development environment

.PHONY: dev-test
dev-test: test-unit-fast build ## 👨‍💻 Quick development test cycle

.PHONY: dev-reset
dev-reset: clean db-reset generate ## 🔄 Reset development environment

.PHONY: ci
ci: quality test-ci build templ-fmt  ## 🚀 Run CI pipeline

.PHONY: ci-full
ci-full: quality test-all build ## 🚀 Run full CI pipeline

# ============================================================================
# 🔧 Maintenance
# ============================================================================

.PHONY: deps-update
deps-update: ## 📦 Update Go dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

.PHONY: deps-check
deps-check: ## 🔍 Check for dependency vulnerabilities
	@echo "$(BLUE)Checking dependencies for vulnerabilities...$(NC)"
	@go list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)✅ Dependency check complete$(NC)"

.PHONY: docker-build
docker-build: ## 🐳 Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t erp-app .
	@echo "$(GREEN)✅ Docker image built$(NC)"

.PHONY: docker-run
docker-run: ## 🐳 Run application in Docker
	@echo "$(BLUE)Running application in Docker...$(NC)"
	@docker run -p 8080:8080 erp-app

# ============================================================================
# 🎯 Aliases & Shortcuts
# ============================================================================

.PHONY: dev
dev: dev-test ## 👨‍💻 Alias for dev-test

.PHONY: setup
setup: dev-setup ## 🛠️ Alias for dev-setup

.PHONY: reset
reset: dev-reset ## 🔄 Alias for dev-reset

.PHONY: gen
gen: generate ## 🔄 Alias for generate

.PHONY: coverage
coverage: test-coverage ## 📊 Alias for test-coverage

.PHONY: bench
bench: test-benchmark ## 🏃 Alias for test-benchmark

.PHONY: watch
watch: test-watch ## 👀 Alias for test-watch

.PHONY: quick
quick: test-quick ## ⚡ Alias for test-quick

# Legacy aliases for backward compatibility
.PHONY: createdb dropdb migrateup migratedown migratedrop
createdb: db-create
dropdb: db-drop  
migrateup: migrate-up
migratedown: migrate-down
migratedrop: migrate-drop
