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

sqlc: ## Generate SQLC store code
	@sqlc generate

mock: ## Generate mocks for interfaces
	@mockgen -package mockstore -destination internal/core/tenant/store/mock.go github.com/your/module/internal/core/tenant Store

fmt: ## Format Go code
	@go fmt ./...

lint: ## Lint Go code
	@golangci-lint run ./...

test: ## Run all tests
	@go test -v -cover ./...

test-core: ## Run only core layer tests
	@go test -v ./internal/core/...

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
	@go run cmd/server

.PHONY: help clean ci fmt lint test test-core test-repo \
	createdb dropdb migrateup migratedown migratedrop migrate-create \
	sqlc mock proto buf buf-lint evans dbdocs sql2dbml run check-tools
