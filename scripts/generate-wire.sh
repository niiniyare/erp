#!/bin/bash

# Wire generation script for Awo ERP
# This script generates Wire dependency injection code

set -e

echo " Generating Wire dependency injection code..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

# Install Wire if not present
if ! command -v wire &>/dev/null; then
  echo " Installing Wire..."
  go install github.com/google/wire/cmd/wire@latest
fi

# Generate Wire code for the main server
echo "⚡ Generating Wire code for cmd/server..."
cd cmd/server
wire
cd ../..

# Generate Wire code for any other injectors (if they exist)
# echo "⚡ Generating Wire code for tests..."
# cd test
# wire generate || echo "⚠️  No Wire files found in test directory"
# cd ..

# Verify that generated files are valid Go
echo "✅ Verifying generated code..."
go build -o /dev/null ./cmd/server/

echo " Wire generation completed successfully!"
echo ""
echo "Generated files:"
find . -name "wire_gen.go" -type f

echo ""
echo "Next steps:"
echo "1. Review the generated wire_gen.go files"
echo "2. Test the application: go run ./cmd/server/"
echo "3. Run tests: go test ./..."

