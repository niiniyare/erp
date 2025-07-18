#!/bin/bash

# Generic Database Setup Script for Multi-Tenant SaaS
# Executes seed files and test files in alphabetical order

set -euo pipefail # Exit on error, undefined vars, pipe failures

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Script directory
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load environment variables
if [[ -f "${SCRIPT_DIR}/.env" ]]; then
  set -a # Export all variables
  source "${SCRIPT_DIR}/.env"
  set +a
else
  echo -e "${RED}Error: .env file not found in ${SCRIPT_DIR}${NC}"
  exit 1
fi

# Validate required environment variables
readonly REQUIRED_VARS=("DB_NAME" "DB_USER" "DB_PSSWD")
for var in "${REQUIRED_VARS[@]}"; do
  if [[ -z "${!var:-}" ]]; then
    echo -e "${RED}Error: Required environment variable $var is not set${NC}"
    exit 1
  fi
done

# Database configuration
readonly DB_HOST="${DB_HOST:-localhost}"
readonly DB_PORT="${DB_PORT:-5432}"
readonly PGPASSWORD="${DB_PSSWD}"
export PGPASSWORD

# Directories
readonly DB_DIR="${SCRIPT_DIR}/db"
readonly SEED_DIR="${DB_DIR}/seed"
readonly TEST_DIR="${DB_DIR}/test"

# Logging
readonly LOG_FILE="${SCRIPT_DIR}/db_setup.log"

# Functions
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "${LOG_FILE}"
}

print_header() {
  echo -e "${BLUE}================================${NC}"
  echo -e "${BLUE}  Generic Database Setup Script${NC}"
  echo -e "${BLUE}================================${NC}"
  echo "Database: ${DB_NAME}"
  echo "User: ${DB_USER}"
  echo "Host: ${DB_HOST}:${DB_PORT}"
  echo "Seed Directory: ${SEED_DIR}"
  echo "Test Directory: ${TEST_DIR}"
  echo ""
}

check_dependencies() {
  local deps=("psql")
  for dep in "${deps[@]}"; do
    if ! command -v "$dep" &>/dev/null; then
      echo -e "${RED}Error: $dep is required but not installed${NC}"
      exit 1
    fi
  done
}

test_connection() {
  echo -e "${YELLOW}Testing database connection...${NC}"
  if psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -c '\q' &>/dev/null; then
    echo -e "${GREEN}✓ Database connection successful${NC}"
    log "Database connection test passed"
  else
    echo -e "${RED}✗ Database connection failed${NC}"
    echo "Please verify:"
    echo "  - Database server is running"
    echo "  - Database '${DB_NAME}' exists"
    echo "  - User '${DB_USER}' has proper permissions"
    echo "  - Password is correct"
    exit 1
  fi
}

create_directories() {
  echo -e "${YELLOW}Creating directory structure...${NC}"
  mkdir -p "${SEED_DIR}" "${TEST_DIR}"
  log "Created directories: ${SEED_DIR}, ${TEST_DIR}"
}

execute_sql_file() {
  local file="$1"
  local description="$2"

  if [[ ! -f "$file" ]]; then
    echo -e "${RED}✗ File not found: $file${NC}"
    return 1
  fi

  echo -e "${YELLOW}Executing: $description${NC}"
  log "Executing SQL file: $file"

  if psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -f "$file" >>"${LOG_FILE}" 2>&1; then
    echo -e "${GREEN}✓ $description completed${NC}"
    log "Successfully executed: $file"
    return 0
  else
    echo -e "${RED}✗ $description failed${NC}"
    log "Failed to execute: $file"
    return 1
  fi
}

run_sql_files() {
  local dir="$1"
  local context="$2"

  if [[ ! -d "$dir" ]]; then
    echo -e "${YELLOW}Directory not found: $dir - skipping $context${NC}"
    return 0
  fi

  echo -e "${YELLOW}Running $context files from $dir...${NC}"
  local files=()
  while IFS= read -r -d $'\0' file; do
    files+=("$file")
  done < <(find "$dir" -maxdepth 1 -name '*.sql' -print0 | sort -z)

  if [[ ${#files[@]} -eq 0 ]]; then
    echo -e "${YELLOW}No SQL files found in $dir${NC}"
    return 0
  fi

  for file in "${files[@]}"; do
    local description="${context}: $(basename "$file")"
    execute_sql_file "$file" "$description" || {
      echo -e "${RED}$context process failed at: $description${NC}"
      return 1
    }
  done

  echo -e "${GREEN}✓ All $context files executed successfully${NC}"
  log "$context process completed successfully"
  return 0
}

cleanup() {
  echo -e "${YELLOW}Cleaning up...${NC}"
  unset PGPASSWORD
  log "Cleanup completed"
}

show_summary() {
  echo -e "${BLUE}================================${NC}"
  echo -e "${BLUE}  Setup Summary${NC}"
  echo -e "${BLUE}================================${NC}"
  echo -e "${GREEN}✓ Database operations completed${NC}"
  echo ""
  echo "Seed files executed from: ${SEED_DIR}"
  echo "Test files executed from: ${TEST_DIR}"
  echo ""
  echo "Log file: ${LOG_FILE}"
  echo -e "${BLUE}================================${NC}"
}

main() {
  # Initialize log file
  echo "Database setup started at $(date)" >"${LOG_FILE}"

  print_header
  check_dependencies
  test_connection
  create_directories

  # Run seed files
  if ! run_sql_files "${SEED_DIR}" "Seed"; then
    echo -e "${RED}Seed process aborted${NC}"
    exit 1
  fi

  # Run test files
  if ! run_sql_files "${TEST_DIR}" "Test"; then
    echo -e "${YELLOW}Some tests failed - check logs for details${NC}"
  fi

  cleanup
  show_summary

  echo -e "${GREEN}Database setup completed!${NC}"
  log "Database setup process completed"
}

# Trap to cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
