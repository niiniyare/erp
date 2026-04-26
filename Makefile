# =============================================================================
# AWO ERP — Makefile
# Ref: https://tech.davis-hansson.com/p/make/
# =============================================================================

SHELL        := bash
.ONESHELL:
.SHELLFLAGS  := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS    += --warn-undefined-variables --no-builtin-rules --no-print-directory
.DEFAULT_GOAL := help

# =============================================================================
#  Configuration
# =============================================================================

# Database
DB_NAME  ?= ledger
DB_USER  ?= admin
DB_PSSWD ?= admin
DB_HOST  ?= localhost
DB_PORT  ?= 5432
TEST_DATABASE_URL="postgres://user:pass@localhost:5432/erp_test?sslmode=di
DB_URL   ?= postgresql://$(DB_USER):$(DB_PSSWD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
# Paths
MIGRATION_PATH := db/migration
SQLC_OUT       := db/sqlc
DOCS_PATH      := docs
COVERAGE_FILE  := coverage.out
COVERAGE_HTML  := coverage.html

# Ports
DOC_PORT  ?= 8081
GRPC_PORT ?= 9090

# Required tools (checked by check-tools)
REQUIRED_TOOLS := sqlc mockgen migrate psql golangci-lint mkdocs sleek awoctl

# Test configuration
TEST_TIMEOUT             := 10m
TEST_TIMEOUT_FAST        := 5m
TEST_TIMEOUT_E2E         := 15m
TEST_TIMEOUT_INTEGRATION := 8m
TEST_PARALLEL_JOBS       := 4
DOCKER_COMPOSE_TEST      := docker-compose.test.yml

# Test directories
UNIT_TEST_DIRS        := ./internal/core/... ./internal/shared/... ./internal/adapters/...
INTEGRATION_TEST_DIRS := ./test/integration/...
E2E_TEST_DIRS         := ./test/e2e/...

# Server log / PID
LOG_FILE := server.log
PID_FILE := server.pid

# Terminal colours
ESC    := $(shell printf '\033')
RED    := $(ESC)[0;31m
GREEN  := $(ESC)[0;32m
YELLOW := $(ESC)[0;33m
BLUE   := $(ESC)[0;34m
PURPLE := $(ESC)[0;35m
CYAN   := $(ESC)[0;36m
NC     := $(ESC)[0m

# =============================================================================
# 🆘 Help & Utilities
# =============================================================================

.PHONY: help
help: ##  Show this help message
	@echo "$(CYAN)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} \
		/^[a-zA-Z0-9_.-]+:.*?##/ { \
			gsub(/^[ \t]+/, "", $$2); \
			printf "  $(GREEN)%-26s$(NC) %s\n", $$1, $$2 \
		}' $(MAKEFILE_LIST)
	@echo ""

.PHONY: check-tools
check-tools: ##  Verify required tools are installed
	@echo "$(BLUE)Checking required tools...$(NC)"
	@$(foreach tool,$(REQUIRED_TOOLS), \
		command -v $(tool) >/dev/null 2>&1 || { \
			echo "$(RED)❌  Missing: $(tool)$(NC)"; exit 1; \
		};)
	@echo "$(GREEN)✅  All required tools are installed$(NC)"

.PHONY: status
status: ##  Show project configuration
	@echo "$(CYAN)Project Configuration:$(NC)"
	@echo "  Database      : $(DB_USER)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)"
	@echo "  Migrations    : $(MIGRATION_PATH)"
	@echo "  Docs port     : $(DOC_PORT)"
	@echo "  Test timeout  : $(TEST_TIMEOUT)"
	@echo "  Parallel jobs : $(TEST_PARALLEL_JOBS)"

# =============================================================================
# ️  Code Generation
# =============================================================================

.PHONY: generate
generate: sqlc mock wire ##  Run all code generators (SQLC → Mocks → Wire)

# --- SQLC -------------------------------------------------------------------

.PHONY: sqlc-install
sqlc-install: ##  Install the sqlc binary
	@echo "$(BLUE)Installing sqlc...$(NC)"
	CGO_CFLAGS="-D_GNU_SOURCE" go install -v github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "$(GREEN)✅  sqlc installed$(NC)"

.PHONY: sqlc
sqlc: ## ️  Generate SQLC store + db.Store mock
	@echo "$(BLUE)Generating SQLC code...$(NC)"
	sqlc generate
	@echo "$(GREEN)✅  SQLC generation complete$(NC)"
	@echo "$(BLUE)Generating db.Store mock...$(NC)"
	go generate ./db/sqlc/...
	@echo "$(GREEN)✅  Mock generation complete$(NC)"

.PHONY: sqlc-lint
sqlc-lint: ##  Lint SQL queries
	@./db/queries/lint.sh

# --- Mocks ------------------------------------------------------------------

.PHONY: mock
mock: ##  Regenerate all go:generate mocks
	@echo "$(BLUE)Generating mocks...$(NC)"
	go generate ./...
	@echo "$(GREEN)✅  Mocks generated$(NC)"

# --- Wire -------------------------------------------------------------------

.PHONY: wire-install
wire-install: ##  Install Google Wire
	@echo "$(BLUE)Installing Wire...$(NC)"
	go install github.com/google/wire/cmd/wire@latest
	@echo "$(GREEN)✅  Wire installed$(NC)"

.PHONY: wire
wire: ## ⚡ Generate Wire dependency-injection code
	@echo "$(BLUE)Generating Wire code...$(NC)"
	@./scripts/generate-wire.sh
	@echo "$(GREEN)✅  Wire generation complete$(NC)"

.PHONY: wire-check
wire-check: ##  Verify Wire files are up to date
	@echo "$(BLUE)Checking Wire files...$(NC)"
	cd cmd/server && wire check
	@echo "$(GREEN)✅  Wire files are up to date$(NC)"

# --- Misc --------------------------------------------------------------------

.PHONY: interface2any
interface2any: ##  Replace interface{} with any across all Go files
	@find . -type f -name '*.go' | xargs sed -i 's/interface{}/any/g'

# =============================================================================
# ️  Scaffolding  (awoctl)
# =============================================================================

.PHONY: scaffold-install
scaffold-install: ##  Build and install the awoctl scaffolding tool
	@echo "$(BLUE)Building awoctl...$(NC)"
	go build -o awoctl ./cmd/awoctl/
	@echo "$(GREEN)✅  awoctl built$(NC)"

.PHONY: scaffold-module
scaffold-module: ##  New module  — make scaffold-module name=<n> [docs=true] [tests=true]
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌  name is required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-module name=inventory [docs=true] [tests=true]$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating module: $(name)$(NC)"
	@cmd="./awoctl new module $(name) --verbose"; \
	[ "$(docs)"  = "true" ] && cmd="$$cmd --with-docs";  true; \
	[ "$(tests)" = "true" ] && cmd="$$cmd --with-tests"; true; \
	eval $$cmd
	@echo "$(GREEN)✅  Module '$(name)' generated$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Update domain models in internal/core/$(name)/domain/"
	@echo "  2. Run 'make sqlc' to regenerate DB layer"
	@echo "  3. Implement business logic"

.PHONY: scaffold-component
scaffold-component: ##  New component — make scaffold-component module=<m> name=<n> type=<t>
	@if [ -z "$(module)" ] || [ -z "$(name)" ] || [ -z "$(type)" ]; then \
		echo "$(RED)❌  module, name, and type are required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-component module=finance name=payment type=entity$(NC)"; \
		echo "$(YELLOW)Types: entity service repository handler dto middleware validator mapper$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating $(type): $(name) in module $(module)$(NC)"
	./awoctl new component $(module) $(name) $(type) --verbose
	@echo "$(GREEN)✅  Component '$(name)' generated$(NC)"

.PHONY: scaffold-feature
scaffold-feature: ## ⭐ New feature — make scaffold-feature path=<module/feature>
	@if [ -z "$(path)" ]; then \
		echo "$(RED)❌  path is required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-feature path=finance/budgets$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating feature: $(path)$(NC)"
	./awoctl new feature $(path) --verbose
	@echo "$(GREEN)✅  Feature '$(path)' generated$(NC)"

.PHONY: scaffold-docs
scaffold-docs: ##  Generate docs for existing module — make scaffold-docs name=<n>
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌  name is required$(NC)"; \
		echo "$(YELLOW)Usage: make scaffold-docs name=inventory$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Generating docs for module: $(name)$(NC)"
	./awoctl docs module $(name) --verbose
	@echo "$(GREEN)✅  Docs generated at: docs/reference/modules/$(name)/$(NC)"

.PHONY: scaffold-help
scaffold-help: ##  Show awoctl help
	@./awoctl --help

# =============================================================================
#  Tests
# =============================================================================

.PHONY: test
test: test-unit ##  Default: run unit tests

# --- Core suites ------------------------------------------------------------

.PHONY: test-unit
test-unit: ##  Unit tests with race detection
	@echo "$(BLUE)Running unit tests...$(NC)"
	go test -tags=unit -race -timeout=$(TEST_TIMEOUT) \
		-parallel=$(TEST_PARALLEL_JOBS) -v $(UNIT_TEST_DIRS)
	@echo "$(GREEN)✅  Unit tests passed$(NC)"

.PHONY: test-unit-fast
test-unit-fast: ## ⚡ Unit tests without race detection
	@echo "$(BLUE)Running unit tests (fast)...$(NC)"
	go test -tags=unit -timeout=$(TEST_TIMEOUT_FAST) \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	@echo "$(GREEN)✅  Fast unit tests passed$(NC)"

.PHONY: test-unit-short
test-unit-short: ##  Unit tests, skipping slow cases (-short)
	@echo "$(BLUE)Running unit tests (short)...$(NC)"
	go test -tags=unit -short -timeout=$(TEST_TIMEOUT_FAST) \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	@echo "$(GREEN)✅  Short unit tests passed$(NC)"

.PHONY: test-integration
test-integration: ##  Integration tests (spins up Docker dependencies)
	@echo "$(BLUE)Running integration tests...$(NC)"
	docker-compose -f $(DOCKER_COMPOSE_TEST) up -d postgres redis || true
	sleep 5
	go test -tags=integration -timeout=$(TEST_TIMEOUT_INTEGRATION) \
		-parallel=$(TEST_PARALLEL_JOBS) -v $(INTEGRATION_TEST_DIRS) || \
		{ docker-compose -f $(DOCKER_COMPOSE_TEST) down; exit 1; }
	docker-compose -f $(DOCKER_COMPOSE_TEST) down
	@echo "$(GREEN)✅  Integration tests passed$(NC)"

.PHONY: test-e2e
test-e2e: ##  End-to-end tests (full Docker environment)
	@echo "$(BLUE)Running E2E tests...$(NC)"
	docker-compose -f $(DOCKER_COMPOSE_TEST) up -d || true
	sleep 10
	go test -tags=e2e -timeout=$(TEST_TIMEOUT_E2E) -v $(E2E_TEST_DIRS) || \
		{ docker-compose -f $(DOCKER_COMPOSE_TEST) down; exit 1; }
	docker-compose -f $(DOCKER_COMPOSE_TEST) down
	@echo "$(GREEN)✅  E2E tests passed$(NC)"

# --- Coverage ---------------------------------------------------------------

.PHONY: test-coverage
test-coverage: ##  Full coverage report (HTML + func summary)
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -race -timeout=$(TEST_TIMEOUT) \
		-coverprofile=$(COVERAGE_FILE) -covermode=atomic \
		-parallel=$(TEST_PARALLEL_JOBS) ./...
	go tool cover -func=$(COVERAGE_FILE) | tail -1
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)✅  Coverage report: $(COVERAGE_HTML)$(NC)"

.PHONY: test-coverage-unit
test-coverage-unit: ##  Unit-test coverage only
	@echo "$(BLUE)Running unit tests with coverage...$(NC)"
	go test -tags=unit -race -timeout=$(TEST_TIMEOUT) \
		-coverprofile=$(COVERAGE_FILE) -covermode=atomic \
		-parallel=$(TEST_PARALLEL_JOBS) $(UNIT_TEST_DIRS)
	go tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo "$(GREEN)✅  Unit coverage complete$(NC)"

.PHONY: test-coverage-open
test-coverage-open: test-coverage ##  Generate and open HTML coverage report
	open $(COVERAGE_HTML) 2>/dev/null || xdg-open $(COVERAGE_HTML) 2>/dev/null || \
		echo "$(YELLOW)Open manually: $(COVERAGE_HTML)$(NC)"

# --- Benchmarks -------------------------------------------------------------

.PHONY: test-benchmark
test-benchmark: ##  Run benchmark tests
	@echo "$(BLUE)Running benchmarks...$(NC)"
	go test -bench=. -benchmem -timeout=$(TEST_TIMEOUT) -run=^$$ ./... | tee benchmark.out
	@echo "$(GREEN)✅  Benchmarks complete$(NC)"

.PHONY: test-benchmark-compare
test-benchmark-compare: ##  Benchmark and compare with previous run
	@echo "$(BLUE)Running benchmarks (count=5)...$(NC)"
	go test -bench=. -benchmem -count=5 -timeout=$(TEST_TIMEOUT) \
		-run=^$$ ./... | tee benchmark-new.out
	@if [ -f benchmark.out ]; then \
		echo "$(BLUE)Comparing with previous results...$(NC)"; \
		benchcmp benchmark.out benchmark-new.out || echo "$(YELLOW)benchcmp not available$(NC)"; \
	fi
	mv benchmark-new.out benchmark.out

# --- Domain tests -----------------------------------------------------------

.PHONY: test-identity
test-identity: ## 🆔 Identity domain tests
	go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/identity/...

.PHONY: test-access
test-access: ##  Access domain tests
	go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/access/...

.PHONY: test-audit
test-audit: ##  Audit domain tests
	go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/audit/...

.PHONY: test-notification
test-notification: ##  Notification domain tests
	go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/notification/...

.PHONY: test-analytics
test-analytics: ##  Analytics domain tests
	go test -race -timeout=$(TEST_TIMEOUT) -v ./internal/core/analytics/...

# --- Suites & utilities -----------------------------------------------------

.PHONY: test-all
test-all: ##  Full suite: unit + integration + e2e
	@echo "$(PURPLE)Running complete test suite...$(NC)"
	$(MAKE) test-unit
	$(MAKE) test-integration
	$(MAKE) test-e2e
	@echo "$(GREEN)✅  All tests completed$(NC)"

.PHONY: test-ci
test-ci: ##  CI pipeline: coverage + integration
	@echo "$(PURPLE)CI test pipeline...$(NC)"
	$(MAKE) test-coverage
	$(MAKE) test-integration
	@echo "$(GREEN)✅  CI pipeline complete$(NC)"

.PHONY: test-watch
test-watch: ##  Watch *.go files and re-run unit tests (requires entr)
	@echo "$(BLUE)Watching for changes (Ctrl-C to stop)...$(NC)"
	find . -name '*.go' | entr -c $(MAKE) test-unit-fast

.PHONY: test-failed
test-failed: ##  Re-run only tests that failed in the last run
	@echo "$(BLUE)Re-running failed tests...$(NC)"
	go test -json ./... | tee /tmp/test-output.json
	jq -r 'select(.Action=="fail" and .Test) | .Package + "/" + .Test' \
		/tmp/test-output.json | xargs -I{} go test -run {} -v

.PHONY: test-list
test-list: ##  List all tests
	@go test -list . ./... 2>/dev/null | grep -E '^(Test|Example|Benchmark)' | sort

.PHONY: test-clean
test-clean: ##  Remove coverage/benchmark artefacts and test containers
	@echo "$(YELLOW)Cleaning test artefacts...$(NC)"
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML) benchmark.out benchmark-new.out
	docker-compose -f $(DOCKER_COMPOSE_TEST) down -v --remove-orphans 2>/dev/null || true
	go clean -testcache -cache
	@echo "$(GREEN)✅  Test cleanup complete$(NC)"

.PHONY: test-deps
test-deps: ##  Install test tooling (mockgen, gotestsum)
	@echo "$(BLUE)Installing test dependencies...$(NC)"
	go install github.com/golang/mock/mockgen@latest
	go install gotest.tools/gotestsum@latest
	@command -v entr     >/dev/null || echo "$(YELLOW)Tip: install 'entr' for test-watch$(NC)"
	@command -v benchcmp >/dev/null || echo "$(YELLOW)Tip: install 'benchcmp' for benchmark-compare$(NC)"

# =============================================================================
#  Code Quality
# =============================================================================

.PHONY: fmt
fmt: ##  Format all Go source files
	@echo "$(BLUE)Formatting...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅  Done$(NC)"

.PHONY: vet
vet: ##  Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✅  Done$(NC)"

.PHONY: lint
lint: ##  Run golangci-lint
	@echo "$(BLUE)Linting...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)✅  Done$(NC)"

.PHONY: security
security: ##  Run gosec security scanner
	@echo "$(BLUE)Running security checks...$(NC)"
	command -v gosec >/dev/null 2>&1 || \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...
	@echo "$(GREEN)✅  Security check complete$(NC)"

.PHONY: quality
quality: fmt vet lint sqlc-lint ##  Run all quality checks (fmt + vet + lint + sqlc-lint)

# =============================================================================
# ️  Database
# =============================================================================

.PHONY: db-create
db-create: ## ️  Create database
	@echo "$(BLUE)Creating database $(DB_NAME)...$(NC)"
	createdb --username="$(DB_USER)" --owner="$(DB_USER)" $(DB_NAME)
	@echo "$(GREEN)✅  Database created$(NC)"

.PHONY: db-drop
db-drop: ##  Drop database
	@echo "$(RED)Dropping database $(DB_NAME)...$(NC)"
	dropdb $(DB_NAME)
	@echo "$(YELLOW)⚠️   Database dropped$(NC)"

.PHONY: db-reset
db-reset: db-drop db-create migrate-up ##  Drop → create → migrate

.PHONY: migrate-up
migrate-up: ## ⬆️  Apply all pending migrations
	@echo "$(BLUE)Running migrations (up)...$(NC)"
	migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose up
	@echo "$(GREEN)✅  Migrations applied$(NC)"

.PHONY: migrate-down
migrate-down: ## ⬇️  Roll back the last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose down
	@echo "$(YELLOW)⚠️   Migration rolled back$(NC)"

.PHONY: migrate-drop
migrate-drop: ##  Drop all schema objects
	@echo "$(RED)Dropping all schema objects...$(NC)"
	migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose drop -f
	@echo "$(YELLOW)⚠️   Schema dropped$(NC)"

.PHONY: migrate-create
migrate-create: ##  New migration — make migrate-create name=<migration_name>
	@if [ -z "$(name)" ]; then \
		echo "$(RED)❌  name is required$(NC)"; \
		echo "$(YELLOW)Usage: make migrate-create name=add_users_table$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(name)$(NC)"
	migrate create -ext sql -dir "$(MIGRATION_PATH)" -digits 2 -seq "$(name)" -verbose
	@echo "$(GREEN)✅  Migration files created$(NC)"

.PHONY: migrate-status
migrate-status: ##  Show current migration version
	@echo "$(BLUE)Migration status...$(NC)"
	migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" version

# =============================================================================
#  Documentation
# =============================================================================

.PHONY: docs
docs: ##  Serve documentation (MkDocs + schema)
	@echo "$(BLUE)Starting AWO ERP Documentation Server...$(NC)"
	@echo "  MkDocs  : http://localhost:$(DOC_PORT)/"
	@echo "  Schema  : http://localhost:$(DOC_PORT)/schema/"
	cd $(DOCS_PATH) && ./start-docs.sh $(DOC_PORT)

.PHONY: docs-build
docs-build: ## ️  Build MkDocs static site
	@echo "$(BLUE)Building documentation...$(NC)"
	mkdocs build
	@echo "$(GREEN)✅  Docs built in site/$(NC)"

.PHONY: docs-dev
docs-dev: docs-build docs ## ‍ Build then serve documentation

.PHONY: docs-schema
docs-schema: ## ️  Generate DB schema documentation (SchemaSpy)
	@echo "$(BLUE)Generating schema documentation...$(NC)"
	$(DOCS_PATH)/scripts/generate-schema-docs.sh
	@echo "$(GREEN)✅  Schema docs generated$(NC)"

.PHONY: dbdocs
dbdocs: ##  Publish DB docs from DBML
	@echo "$(BLUE)Building dbdocs...$(NC)"
	dbdocs build $(DOCS_PATH)/schema.dbml
	@echo "$(GREEN)✅  DB docs published$(NC)"

.PHONY: sql2dbml
sql2dbml: ##  Convert SQL migrations → DBML
	sql2dbml $(MIGRATION_PATH)/*.up.sql --postgres -o $(DOCS_PATH)/schema.dbml

# =============================================================================
#  Application
# =============================================================================

.PHONY: run
run: ##  Start the server (logs to stdout and $(LOG_FILE))
	@echo "$(BLUE)Starting server... (logs → $(LOG_FILE))$(NC)"
	go run ./cmd/server/ 2>&1 | tee $(LOG_FILE)

.PHONY: run-bg
run-bg: ##  Start the server in the background
	@echo "$(BLUE)Starting server in background...$(NC)"
	bash -c "go run ./cmd/server/ 2>&1 | tee -a $(LOG_FILE)" &
	echo $$! > $(PID_FILE)
	@echo "$(GREEN)Server started (PID: $$(cat $(PID_FILE))) — logs → $(LOG_FILE)$(NC)"

.PHONY: stop
stop: ##  Stop the background server
	@if [ -f $(PID_FILE) ]; then \
		echo "$(RED)Stopping server (PID: $$(cat $(PID_FILE)))...$(NC)"; \
		kill $$(cat $(PID_FILE)) 2>/dev/null || true; \
		rm -f $(PID_FILE); \
		echo "$(GREEN)Server stopped$(NC)"; \
	else \
		echo "$(YELLOW)No running server found$(NC)"; \
	fi

.PHONY: build
build: ##  Compile → bin/server
	@echo "$(BLUE)Building...$(NC)"
	go build -o bin/server ./cmd/server/
	@echo "$(GREEN)✅  Built: bin/server$(NC)"

.PHONY: temporal-server
temporal-server: ## ⏰ Start Temporal dev server
	@echo "$(BLUE)Starting Temporal server...$(NC)"
	temporal server start-dev

# =============================================================================
#  Docker
# =============================================================================

.PHONY: docker-build
docker-build: ##  Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t erp-app .
	@echo "$(GREEN)✅  Image built: erp-app$(NC)"

.PHONY: docker-run
docker-run: ##  Run Docker image
	docker run -p 8080:8080 erp-app

# =============================================================================
#  Dependency Management
# =============================================================================

.PHONY: deps-update
deps-update: ##  Update all Go dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✅  Dependencies updated$(NC)"

.PHONY: deps-check
deps-check: ##  Scan dependencies for known vulnerabilities (nancy)
	@echo "$(BLUE)Checking for vulnerabilities...$(NC)"
	go list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)✅  Dependency check complete$(NC)"

# =============================================================================
#  Code Statistics
# =============================================================================

.PHONY: stats
stats: ##  Show code statistics with per-language breakdown
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║$(NC) $(CYAN)           AWO ERP — CODE STATISTICS$(NC)                        $(BLUE)║$(NC)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@declare -A lines_map files_map; \
	total_lines=0; total_files=0; \
	for lang in go sql js ts tsx jsx css html json yaml yml md sh py; do \
		fc=$$(find . -name "*.$$lang" -not -path '*/vendor/*' \
			-not -path '*/node_modules/*' -type f 2>/dev/null | wc -l); \
		lc=$$(find . -name "*.$$lang" -not -path '*/vendor/*' \
			-not -path '*/node_modules/*' -type f \
			-exec cat {} + 2>/dev/null | wc -l); \
		files_map[$$lang]=$$fc; \
		lines_map[$$lang]=$${lc:-0}; \
		total_lines=$$((total_lines + $${lines_map[$$lang]})); \
		total_files=$$((total_files + fc)); \
	done; \
	printf "$(YELLOW)┌─────────────┬────────┬────────────┬──────────┐$(NC)\n"; \
	printf "$(YELLOW)│$(NC) $(PURPLE)%-11s$(NC) $(YELLOW)│$(NC) $(PURPLE)%6s$(NC) $(YELLOW)│$(NC) $(PURPLE)%10s$(NC) $(YELLOW)│$(NC) $(PURPLE)%8s$(NC) $(YELLOW)│$(NC)\n" \
		"Language" "Files" "Lines" "% Total"; \
	printf "$(YELLOW)├─────────────┼────────┼────────────┼──────────┤$(NC)\n"; \
	for lang in go sql js ts tsx jsx css html json yaml yml md sh py; do \
		fc=$${files_map[$$lang]}; lc=$${lines_map[$$lang]}; \
		[ $$fc -eq 0 ] && continue; \
		[ $$total_lines -gt 0 ] && \
			pct=$$(awk "BEGIN{printf \"%.1f\", $$lc*100/$$total_lines}") || pct="0.0"; \
		case "$$lang" in \
			go)            clr="$(GREEN)"  ;; \
			sql)           clr="$(CYAN)"   ;; \
			js|ts|tsx|jsx) clr="$(YELLOW)" ;; \
			css|html)      clr="$(PURPLE)" ;; \
			json|yaml|yml) clr="$(BLUE)"   ;; \
			md)            clr="$(CYAN)"   ;; \
			*)             clr="$(NC)"     ;; \
		esac; \
		printf "$(YELLOW)│$(NC) $${clr}%-11s$(NC) $(YELLOW)│$(NC) %6d $(YELLOW)│$(NC) %10d $(YELLOW)│$(NC) %7s%% $(YELLOW)│$(NC)\n" \
			"$$lang" "$$fc" "$$lc" "$$pct"; \
	done; \
	printf "$(YELLOW)├─────────────┼────────┼────────────┼──────────┤$(NC)\n"; \
	printf "$(YELLOW)│$(NC) $(RED)%-11s$(NC) $(YELLOW)│$(NC) %6d $(YELLOW)│$(NC) %10d $(YELLOW)│$(NC) %8s $(YELLOW)│$(NC)\n" \
		"TOTAL" "$$total_files" "$$total_lines" "100.0%"; \
	printf "$(YELLOW)└─────────────┴────────┴────────────┴──────────┘$(NC)\n"; \
	echo ""; \
	echo "$(GREEN) Directory breakdown:$(NC)"; \
	for dir in cmd internal pkg db docs test; do \
		[ -d "$$dir" ] || continue; \
		n=$$(find "$$dir" -type f | wc -l); \
		[ $$n -eq 0 ] && continue; \
		printf "  $(CYAN)%-12s$(NC) %d files\n" "$$dir/" "$$n"; \
	done; \
	echo ""; \
	echo "$(GREEN)✅  Analysis complete$(NC)"

# =============================================================================
#  Clean
# =============================================================================

.PHONY: clean
clean: ##  Remove build artefacts, logs, coverage files, and Go caches
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -f $(LOG_FILE) $(PID_FILE)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f benchmark.out benchmark-new.out
	rm -f docs/swagger/*.json
	go clean -testcache -cache -modcache
	@echo "$(GREEN)✅  Clean complete$(NC)"

# =============================================================================
#  Composite Workflows
# =============================================================================

.PHONY: dev-setup
dev-setup: check-tools scaffold-install db-create migrate-up generate test-deps ## ️  Bootstrap development environment

.PHONY: dev-test
dev-test: test-unit-fast build ## ‍ Quick development cycle (unit tests + build)

.PHONY: dev-reset
dev-reset: clean db-reset generate ##  Full reset (clean → db reset → codegen)

.PHONY: ci
ci: quality test-ci build ##  Standard CI pipeline

.PHONY: ci-full
ci-full: quality test-all build ##  Full CI pipeline (all test suites)

# =============================================================================
#  Short Aliases
# =============================================================================

.PHONY: dev setup reset gen coverage bench watch quick
dev:      dev-test       ## ‍ → dev-test
setup:    dev-setup      ## ️  → dev-setup
reset:    dev-reset      ##  → dev-reset
gen:      generate       ##  → generate
coverage: test-coverage  ##  → test-coverage
bench:    test-benchmark ##  → test-benchmark
watch:    test-watch     ##  → test-watch
quick:    test-unit-short ## ⚡ → test-unit-short

# Legacy names (backward compatibility)
.PHONY: createdb dropdb migrateup migratedown migratedrop
createdb:    db-create
dropdb:      db-drop
migrateup:   migrate-up
migratedown: migrate-down
migratedrop: migrate-drop
