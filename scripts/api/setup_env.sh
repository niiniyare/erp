#!/bin/bash

# Environment setup script for API tests
# Sources configuration from .env and config.yaml

# Get the project root directory (3 levels up from scripts/api/)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../" && pwd)"

# Source environment variables if .env exists
if [ -f "$PROJECT_ROOT/.env" ]; then
    echo " Loading environment from $PROJECT_ROOT/.env"
    # Export variables from .env, handling comments and empty lines
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
else
    echo "⚠️  No .env file found at $PROJECT_ROOT/.env"
fi

# Set default values from config.yaml if not already set
export SERVER_PORT="${SERVER_PORT:-8080}"
export GRPC_PORT="${GRPC_PORT:-9090}"
export SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_NAME="${DB_NAME:-ledger}"

echo "✅ Environment configured:"
echo "   SERVER_PORT: $SERVER_PORT"
echo "   GRPC_PORT: $GRPC_PORT"
echo "   SERVER_HOST: $SERVER_HOST"
echo "   DB_HOST: $DB_HOST"
echo "   DB_PORT: $DB_PORT"
echo "   DB_NAME: $DB_NAME"
echo ""