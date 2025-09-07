#!/bin/bash

# AWO ERP Documentation Server Startup Script
# This script starts the documentation server that serves both:
# - MkDocs generated HTML documentation from ../site/
# - SchemaSpy generated database schema documentation from ./schema/

set -e

# Configuration
DEFAULT_PORT="8081"
PORT=${DOC_PORT:-${1:-$DEFAULT_PORT}}

echo "🏢 AWO ERP Documentation Server"
echo "================================"
echo ""
echo "📚 Serving MkDocs documentation from: ../site/"
echo "🗄️ Serving Schema documentation from: ./schema/"
echo "🌐 Server will start on: http://localhost:$PORT"
echo ""

# Check if required directories exist
if [ ! -d "../site" ]; then
    echo "❌ Error: ../site/ directory not found!"
    echo "   Please run 'mkdocs build' from the project root to generate the site."
    exit 1
fi

if [ ! -d "./schema" ]; then
    echo "⚠️  Warning: ./schema/ directory not found!"
    echo "   Schema documentation will not be available."
    echo "   Run SchemaSpy to generate schema documentation:"
    echo "   java -jar schemaspy.jar [options] -o docs/schema"
    echo ""
fi

# Check if site has index.html
if [ ! -f "../site/index.html" ]; then
    echo "ℹ️  Note: No index.html found in ../site/, using fallback redirect."
fi

echo "🚀 Starting server..."
echo ""
echo "Available endpoints:"
echo "  📖 Main Documentation: http://localhost:$PORT/"
echo "  🗄️ Database Schema:     http://localhost:$PORT/schema/"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start the Go server
cd "$(dirname "$0")"
go run main.go --port "$PORT"