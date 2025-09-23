#!/bin/bash

# Script to generate Templ template files
# This script should be run whenever .templ files are modified

set -e

echo "Generating Templ templates..."

# Check if templ is installed
# if ! command -v templ &> /dev/null; then
#     echo "Installing templ..."
#     go install github.com/a-h/templ/cmd/templ@latest
# fi

# Generate templates
echo "Generating templates in internal/ui/templates..."
cd /data/data/com.termux/files/home/project/erp

# Generate all .templ files
templ generate

echo "Templ generation completed successfully!"

# Optional: Format generated Go files
echo "Formatting generated files..."
gofmt -w internal/ui/templates/

echo "All done!"

