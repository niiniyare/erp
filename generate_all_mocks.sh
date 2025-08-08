#!/bin/bash

# This script generates mocks for all repository and service interfaces.

# Exit immediately if a command exits with a non-zero status.
set -e

# Base directory
BASE_DIR="internal/core"

# Find all interface files (repository.go, *.go, etc.) and generate mocks
find "$BASE_DIR" -type f \( -name "repository.go" -o -name "*_service.go" \) -print0 | while IFS= read -r -d $'\0' file; do
  # Get the directory and package name
  dir=$(dirname "$file")
  package_name=$(basename "$dir")
  destination_file="$dir/mock.go"

  echo "Generating mock for $package_name in $dir"

  # Generate the mock
  go run go.uber.org/mock/mockgen -source="$file" -destination="$destination_file" -package="$package_name"
done

echo "Mock generation complete."

# Generate mock for db/sqlc/store.go
DB_STORE_SOURCE="db/sqlc/store.go"
DB_STORE_DESTINATION="db/sqlc/mock_store.go"
DB_STORE_PACKAGE="db"

echo "Generating mock for $DB_STORE_SOURCE"
go run go.uber.org/mock/mockgen -source="$DB_STORE_SOURCE" -destination="$DB_STORE_DESTINATION" -package="$DB_STORE_PACKAGE"

# Generate mocks for shared infrastructure components
SHARED_DIR="internal/shared"

# Generate mock for logger
LOGGER_SOURCE="$SHARED_DIR/logger/logger.go"
LOGGER_DESTINATION="$SHARED_DIR/logger/mock_logger.go"
LOGGER_PACKAGE="logger"

if [ -f "$LOGGER_SOURCE" ]; then
  echo "Generating mock for logger"
  go run go.uber.org/mock/mockgen -source="$LOGGER_SOURCE" -destination="$LOGGER_DESTINATION" -package="$LOGGER_PACKAGE"
else
  echo "Logger source file not found: $LOGGER_SOURCE"
fi

# Generate mock for metrics
METRICS_SOURCE="$SHARED_DIR/metrics/metrics.go"
METRICS_DESTINATION="$SHARED_DIR/metrics/mock_metrics.go"
METRICS_PACKAGE="metrics"

if [ -f "$METRICS_SOURCE" ]; then
  echo "Generating mock for metrics"
  go run go.uber.org/mock/mockgen -source="$METRICS_SOURCE" -destination="$METRICS_DESTINATION" -package="$METRICS_PACKAGE"
else
  echo "Metrics source file not found: $METRICS_SOURCE"
fi

# Generate mock for tracing
TRACING_SOURCE="$SHARED_DIR/tracing/tracing.go"
TRACING_DESTINATION="$SHARED_DIR/tracing/mock_tracing.go"
TRACING_PACKAGE="tracing"

if [ -f "$TRACING_SOURCE" ]; then
  echo "Generating mock for tracing"
  go run go.uber.org/mock/mockgen -source="$TRACING_SOURCE" -destination="$TRACING_DESTINATION" -package="$TRACING_PACKAGE"
else
  echo "Tracing source file not found: $TRACING_SOURCE"
fi

echo "Generating cache mock"
go run go.uber.org/mock/mockgen -source="internal/platform/cache/interface.go" -destination="internal/platform/cache/mock.go" -package="cache"
echo "All mock generation complete."
