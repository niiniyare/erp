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

DB_NAME ?= ledger
DB_USER ?= admin
DB_PSSWD ?= admin
DB_URL ?= postgresql://$(DB_USER):$(DB_PSSWD)@localhost:5432/$(DB_NAME)?sslmode=disable

MIGRATION_PATH := db/migration/
SQLC_OUT := db/sqlc
PROTO := api/proto/v1
PB := api/pb/v1

REQUIRED_TOOLS := sqlc protoc mockgen migrate dbdocs psql buf

# ============================================================================
# 🧪 General
# ============================================================================

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36mtarget\033[0m\n"} \
	/^[a-zA-Z0-9_.-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

check-tools: ## Check required tools
	@$(foreach tool,$(REQUIRED_TOOLS),\
		command -v $(tool) >/dev/null 2>&1 || { echo >&2 "Missing: $(tool)"; exit 1; };)

clean: ## Clean generated files
	@rm -f $(PB)/*.go doc/swagger/*.json

ci: check-tools fmt lint test proto sqlc ## Run all essential checks

# ============================================================================
# 🧬 API Layer (Handlers, Middleware, Routes)
# ============================================================================

proto: check-tools ## Generate gRPC and gateway files
	@mkdir -p $(PB)
	@protoc --proto_path=$(PROTO) \
		--go_out=$(PB) --go_opt=paths=source_relative \
		--go-grpc_out=$(PB) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PB) --grpc-gateway_opt=paths=source_relative \
		$(PROTO)/*.proto

buf: ## Use Buf to generate proto files
	@buf generate --output $(PB)

buf-lint: ## Lint proto files with Buf
	@buf lint | jq

evans: ## Start Evans gRPC REPL
	@evans --host localhost --port 9090 -r repl

# ============================================================================
# 🧠 Core Business Layer (Domain, Services, Interfaces)
# ============================================================================

sqlc-lint: ## detection of basd queries
	@./db/queries/lint.sh
sqlc: ## Generate SQLC store code
	@sqlc generate

mock: ## Generate mocks for interfaces
	@./generate_all_mocks.sh
fmt: ## Format Go code
	@go fmt ./...

goa: ## generate 7oa 
	@goa gen github.com/niiniyare/erp/internal/api/design -o internal/api

lint: ## Lint Go code
	@golangci-lint run ./...

gen: sqlc goa ##  generate All 


test: ## Run all tests
	@go test -v -cover ./...

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@go test -tags=unit -race -timeout=10m -v ./internal/core/... ./internal/shared/... ./internal/adapters/...

test-unit-fast: ## Run unit tests without race detection (faster)
	@echo "⚡ Running unit tests (fast mode)..."
	@go test -tags=unit -timeout=5m ./internal/core/... ./internal/shared/... ./internal/adapters/...

test-integration: ## Run integration tests
	@echo "🔗 Running integration tests..."
	@echo "🐳 Starting dependencies..."
	@docker-compose -f docker-compose.test.yml up -d postgres redis || true
	@sleep 5
	@go test -tags=integration -timeout=5m -v ./... || (docker-compose -f docker-compose.test.yml down; exit 1)
	@echo "🧹 Cleaning up dependencies..."
	@docker-compose -f docker-compose.test.yml down

test-e2e: ## Run end-to-end tests
	@echo "🌐 Running E2E tests..."
	@echo "🐳 Starting full environment..."
	@docker-compose -f docker-compose.test.yml up -d || true
	@sleep 10
	@go test -tags=e2e -timeout=15m -v ./test/e2e/... || (docker-compose -f docker-compose.test.yml down; exit 1)
	@echo "🧹 Cleaning up environment..."
	@docker-compose -f docker-compose.test.yml down

test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@go test -race -timeout=10m -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-benchmark: ## Run benchmark tests
	@echo "🏃 Running benchmark tests..."
	@go test -bench=. -benchmem -timeout=10m ./...

test-core: ## Run only core layer tests
	@go test -v ./internal/core/...

# Domain-specific tests
test-identity: ## Run identity domain tests
	@echo "🆔 Testing identity domain..."
	@go test -race -timeout=10m -v ./internal/core/identity/...

test-access: ## Run access domain tests  
	@echo "🔐 Testing access domain..."
	@go test -race -timeout=10m -v ./internal/core/access/...

test-audit: ## Run audit domain tests
	@echo "📋 Testing audit domain..."
	@go test -race -timeout=10m -v ./internal/core/audit/...

test-notification: ## Run notification domain tests
	@echo "📢 Testing notification domain..."
	@go test -race -timeout=10m -v ./internal/core/notification/...

test-analytics: ## Run analytics domain tests
	@echo "📈 Testing analytics domain..."
	@go test -race -timeout=10m -v ./internal/core/analytics/...

test-all: ## Run complete test suite (unit + integration + e2e)
	@echo "🎯 Running complete test suite..."
	@$(MAKE) test-unit
	@$(MAKE) test-integration  
	@$(MAKE) test-e2e
	@echo "✅ All tests completed successfully!"

# Development workflow tests
dev-test: ## Quick development test cycle
	@echo "👨‍💻 Running development test cycle..."
	@$(MAKE) test-unit-fast
	@go build ./...

dev-test-coverage: ## Development test with coverage
	@echo "👨‍💻 Running development test with coverage..."
	@go test -tags=unit -race -timeout=5m -coverprofile=coverage.out ./internal/core/... ./internal/shared/... ./internal/adapters/...
	@go tool cover -func=coverage.out

# CI/CD pipeline commands  
ci-test: ## CI test pipeline
	@echo "🚀 Running CI test pipeline..."
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@go test -race -timeout=10m -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

ci-full: ## Full CI pipeline
	@echo "🚀 Running full CI pipeline..."
	@go build ./...
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@$(MAKE) test-e2e
	@go test -race -timeout=10m -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

# Test cleanup
test-clean: ## Clean test artifacts
	@echo "🧹 Cleaning test artifacts..."
	@rm -f coverage.out coverage.html benchmark.out
	@docker-compose -f docker-compose.test.yml down -v --remove-orphans || true

test-clean-cache: ## Clean Go test cache
	@echo "🧹 Cleaning Go test cache..."
	@go clean -testcache
	@go clean -cache

# ============================================================================
# 🗃️ Repository Layer (SQLC Store, Repos, Converters)
# ============================================================================

test-repo: ## Run tests related to repositories
	@go test -v ./internal/core/tenant/store/...

# ============================================================================
# 🏗️ Infrastructure Layer (Database, Redis, Config)
# ============================================================================

createdb: ## Create the database
	@createdb --username="$(DB_USER)" --owner="$(DB_USER)" $(DB_NAME)

dropdb: ## Drop the database
	@dropdb $(DB_NAME)

migrateup: ## Run all up migrations
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose up

migratedown: ## Roll back the last migration
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose down

migratedrop: ## Drop all schema objects
	@migrate -path "$(MIGRATION_PATH)" -database "$(DB_URL)" -verbose drop

migrate-create: ## Create new migration file: make migrate-create name=init_users
	@name=$(name); \
	if [ -z "$$name" ]; then echo "Missing name param. Use: make migrate-create name=xyz"; exit 1; fi; \
	migrate create -ext sql -dir "$(MIGRATION_PATH)" -digits 2 -seq "$$name" -verbose

dbdocs: ## Generate DB docs from DBML
	@dbdocs build docs/schema.dbml

sql2dbml: ## Convert SQL migration to DBML
	@sql2dbml $(MIGRATION_PATH)/*.up.sql --postgres -o docs/schema.dbml

# ============================================================================
# 🚀 App Entry (main.go / wiring)
# ============================================================================
run: ## Run the app server
	@go run ./cmd/server/*.go

.PHONY: help clean ci fmt lint test test-core test-repo \
	createdb dropdb migrateup migratedown migratedrop migrate-create \
	sqlc mock proto buf buf-lint evans dbdocs sql2dbml run check-tools
