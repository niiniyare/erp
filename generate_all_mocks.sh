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
