#!/bin/bash

# Multi-Tenant ERP Test Database Setup Script
# This script sets up a proper test database for running database tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DB_NAME="${TEST_DB_NAME:-ledger_test}"
TEST_DB_USER="${TEST_DB_USER:-admin}"
TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-admin}"
TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
TEST_DB_PORT="${TEST_DB_PORT:-5432}"
TEST_DATABASE_URL="postgresql://${TEST_DB_USER}:${TEST_DB_PASSWORD}@${TEST_DB_HOST}:${TEST_DB_PORT}/${TEST_DB_NAME}?sslmode=disable"

MIGRATION_PATH="./db/migration/"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if PostgreSQL is running
check_postgres() {
    log_info "Checking PostgreSQL connection..."
    if ! pg_isready -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" >/dev/null 2>&1; then
        log_error "PostgreSQL is not running or not accessible at ${TEST_DB_HOST}:${TEST_DB_PORT}"
        log_info "Please ensure PostgreSQL is running and accessible."
        log_info "You can start PostgreSQL with: brew services start postgresql@14 (macOS) or systemctl start postgresql (Linux)"
        exit 1
    fi
    log_success "PostgreSQL is running"
}

# Check if required tools are available
check_tools() {
    log_info "Checking required tools..."
    local missing_tools=()
    
    for tool in createdb dropdb migrate sqlc psql; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools:"
        log_info "  - PostgreSQL client tools (createdb, dropdb, psql)"
        log_info "  - golang-migrate: https://github.com/golang-migrate/migrate"
        log_info "  - sqlc: https://sqlc.dev/"
        exit 1
    fi
    log_success "All required tools are available"
}

# Create test database user if it doesn't exist
create_test_user() {
    log_info "Creating test database user if needed..."
    
    if psql -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U postgres -lqt | cut -d \| -f 1 | grep -qw "${TEST_DB_USER}"; then
        log_info "User ${TEST_DB_USER} already exists"
    else
        log_info "Creating user ${TEST_DB_USER}..."
        psql -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U postgres -c "CREATE USER ${TEST_DB_USER} WITH CREATEDB PASSWORD '${TEST_DB_PASSWORD}';" || {
            log_warning "Could not create user as postgres superuser, trying with current user..."
            createuser -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" --createdb "${TEST_DB_USER}" || {
                log_warning "Could not create user, assuming it exists or will be created manually"
            }
        }
    fi
}

# Drop existing test database if it exists
drop_test_db() {
    log_info "Dropping existing test database if it exists..."
    if psql -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U "${TEST_DB_USER}" -lqt | cut -d \| -f 1 | grep -qw "${TEST_DB_NAME}"; then
        log_info "Dropping existing database ${TEST_DB_NAME}..."
        dropdb -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U "${TEST_DB_USER}" "${TEST_DB_NAME}" || {
            log_warning "Could not drop database, it may not exist"
        }
    fi
}

# Create test database
create_test_db() {
    log_info "Creating test database ${TEST_DB_NAME}..."
    createdb -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U "${TEST_DB_USER}" -O "${TEST_DB_USER}" "${TEST_DB_NAME}" || {
        log_error "Failed to create test database"
        exit 1
    }
    log_success "Test database ${TEST_DB_NAME} created"
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."
    if [ ! -d "${MIGRATION_PATH}" ]; then
        log_error "Migration path ${MIGRATION_PATH} does not exist"
        exit 1
    fi
    
    migrate -path "${MIGRATION_PATH}" -database "${TEST_DATABASE_URL}" -verbose up || {
        log_error "Failed to run migrations"
        exit 1
    }
    log_success "Database migrations completed"
}

# Generate SQLC code
generate_sqlc() {
    log_info "Generating SQLC code..."
    sqlc generate || {
        log_error "Failed to generate SQLC code"
        exit 1
    }
    log_success "SQLC code generated"
}

# Verify database setup
verify_setup() {
    log_info "Verifying database setup..."
    
    # Check if we can connect
    psql -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U "${TEST_DB_USER}" -d "${TEST_DB_NAME}" -c "SELECT 1;" >/dev/null || {
        log_error "Cannot connect to test database"
        exit 1
    }
    
    # Check if key tables exist
    local tables=("tenants" "tenant_configurations" "users")
    for table in "${tables[@]}"; do
        if ! psql -h "${TEST_DB_HOST}" -p "${TEST_DB_PORT}" -U "${TEST_DB_USER}" -d "${TEST_DB_NAME}" -c "\dt ${table}" | grep -q "${table}"; then
            log_error "Required table ${table} does not exist"
            exit 1
        fi
    done
    
    log_success "Database setup verified"
}

# Create environment file for tests
create_env_file() {
    log_info "Creating test environment configuration..."
    
    cat > .env.test << EOF
# Test Database Configuration
TEST_DATABASE_URL=${TEST_DATABASE_URL}
DB_HOST=${TEST_DB_HOST}
DB_PORT=${TEST_DB_PORT}
DB_USER=${TEST_DB_USER}
DB_PASSWORD=${TEST_DB_PASSWORD}
DB_NAME=${TEST_DB_NAME}
EOF

    log_success "Test environment file .env.test created"
}

# Main setup function
main() {
    echo
    log_info "Setting up test database for Multi-Tenant ERP system"
    echo
    
    check_tools
    check_postgres
    create_test_user
    drop_test_db
    create_test_db
    run_migrations
    generate_sqlc
    verify_setup
    create_env_file
    
    echo
    log_success "🎉 Test database setup completed successfully!"
    echo
    log_info "Database URL: ${TEST_DATABASE_URL}"
    log_info "Environment file: .env.test"
    echo
    log_info "You can now run database tests with:"
    log_info "  export TEST_DATABASE_URL='${TEST_DATABASE_URL}'"
    log_info "  go test -tags=database ./internal/core/tenant/... -v"
    echo
    log_info "Or source the environment file:"
    log_info "  source .env.test && go test -tags=database ./internal/core/tenant/... -v"
    echo
}

# Handle script arguments
case "${1:-setup}" in
    "setup")
        main
        ;;
    "drop")
        log_info "Dropping test database..."
        drop_test_db
        log_success "Test database dropped"
        ;;
    "recreate")
        log_info "Recreating test database..."
        drop_test_db
        create_test_db
        run_migrations
        verify_setup
        log_success "Test database recreated"
        ;;
    "verify")
        log_info "Verifying test database..."
        verify_setup
        log_success "Test database verification completed"
        ;;
    *)
        echo "Usage: $0 {setup|drop|recreate|verify}"
        echo
        echo "Commands:"
        echo "  setup     - Full test database setup (default)"
        echo "  drop      - Drop test database"
        echo "  recreate  - Drop and recreate test database with migrations"
        echo "  verify    - Verify test database is properly configured"
        exit 1
        ;;
esac