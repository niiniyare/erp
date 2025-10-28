#!/bin/bash
set -e

# Ensure TEST_DATABASE_URL is set
if [ -z "${TEST_DATABASE_URL}" ]; then
  echo "Error: TEST_DATABASE_URL is not set."
  exit 1
fi

# Run database migrations
echo "Running database migrations..."
migrate -database "${TEST_DATABASE_URL}" -path db/migration up

# Run database tests
echo "Running database tests..."
go test -cover -v -tags=database ./internal/core/tenant
