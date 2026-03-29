#!/bin/bash

echo " Starting AWO ERP Documentation Server Test..."

# Start the server in the background
cd /data/data/com.termux/files/home/project/erp/docs
echo "Starting server..."
go run main.go --port 8082 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test endpoints
echo ""
echo "Testing endpoints:"

# Test root redirect
echo "1. Testing root endpoint..."
curl -s -o /dev/null -w "Root endpoint (/): HTTP %{http_code}\n" http://localhost:8082/

# Test MkDocs site
echo "2. Testing MkDocs site..."
curl -s -o /dev/null -w "Getting Started page: HTTP %{http_code}\n" http://localhost:8082/getting-started/01-developer-quick-start/

# Test schema documentation
echo "3. Testing Schema documentation..."
curl -s -o /dev/null -w "Schema index: HTTP %{http_code}\n" http://localhost:8082/schema/

# Test specific schema page
echo "4. Testing specific schema table..."
curl -s -o /dev/null -w "Finance accounts table: HTTP %{http_code}\n" http://localhost:8082/schema/tables/finance_accounts.html

# Test CSS and JS assets from MkDocs
echo "5. Testing MkDocs assets..."
curl -s -o /dev/null -w "MkDocs CSS: HTTP %{http_code}\n" http://localhost:8082/assets/stylesheets/main.7e37652d.min.css

# Test Schema assets
echo "6. Testing Schema assets..."
curl -s -o /dev/null -w "Schema CSS: HTTP %{http_code}\n" http://localhost:8082/schema/schemaSpy.css

echo ""
echo "✅ Test complete!"
echo " Documentation available at: http://localhost:8082"
echo "️ Database schema at: http://localhost:8082/schema/"

# Keep server running or stop it
echo ""
echo "Server running with PID: $SERVER_PID"
echo "Press Ctrl+C to stop the server or run: kill $SERVER_PID"

# Wait for user interrupt or stop after 30 seconds
read -t 30 -p "Press Enter to stop server (or wait 30s)..."
kill $SERVER_PID 2>/dev/null
echo ""
echo "Server stopped."