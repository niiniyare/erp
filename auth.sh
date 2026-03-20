#!/bin/bash
# Auth + protected route smoke test

BASE="http://localhost:8080"
COOKIE_JAR="/tmp/erp_test_cookies.txt"
EMAIL="admin@test.com"
PASSWORD="password123"

echo "=== 1. Login ==="
curl -s -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  -c "$COOKIE_JAR" |
  jq .

echo ""
echo "=== 2. Protected route WITH cookie (expect 200) ==="
curl -s -X GET "$BASE/api/v1/finance/accounts" \
  -b "$COOKIE_JAR" |
  jq .

echo ""
echo "=== 3. Protected route WITHOUT cookie (expect 401) ==="
curl -s -X GET "$BASE/api/v1/finance/accounts" |
  jq .

echo ""
echo "=== 4. Logout ==="
curl -s -X POST "$BASE/api/v1/auth/logout" \
  -b "$COOKIE_JAR" \
  -c "$COOKIE_JAR" \
  -o /dev/null -w "HTTP %{http_code}\n"

echo ""
echo "=== 5. Protected route AFTER logout (expect 401) ==="
curl -s -X GET "$BASE/api/v1/finance/accounts" \
  -b "$COOKIE_JAR" |
  jq .

echo ""
echo "=== 6. Locked account (expect 423) ==="
for i in 1 2 3 4 5 6; do
  echo -n "attempt $i: "
  curl -s -X POST "$BASE/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"locked@test.com","password":"wrongpass"}' |
    jq -r '.error // .message'
done

rm -f "$COOKIE_JAR"
