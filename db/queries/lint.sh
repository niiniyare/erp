#!/bin/bash

QUERY_DIR="./db/queries"
echo -e "\n Linting SQLC queries in \033[1m$QUERY_DIR\033[0m"

has_error=false

find "$QUERY_DIR" -name '*.sql' | while read -r file; do
  relpath="${file#$QUERY_DIR/}"

  # # 1. SELECT *
  # grep -n -iE '^\s*select\s+\*' "$file" | while IFS=: read -r line_num line; do
  #   echo -e "\n \033[1;31mSELECT *\033[0m found:"
  #   echo -e "    File: \033[1m$QUERY_DIR/$relpath\033[0m"
  #   echo -e "    Line $line_num: $line"
  #   has_error=true
  # done
  #
  # # 2. JOIN ... .* (ambiguous columns)
  # grep -n -iE 'join\s+[a-zA-Z0-9_]+\s+(as\s+)?[a-z]+\s+on\s+.*\.\*' "$file" | while IFS=: read -r line_num line; do
  #   echo -e "\n \033[1;31mJOIN with table.*\033[0m (ambiguous columns) found:"
  #   echo -e "    File: \033[1m$QUERY_DIR/$relpath\033[0m"
  #   echo -e "    Line $line_num: $line"
  #   has_error=true
  # done
  #
  # # 3. Missing -- name:
  # if ! grep -q '^-- name:' "$file"; then
  #   echo -e "\n \033[1;31mMissing '-- name:'\033[0m directive:"
  #   echo -e "    File: \033[1m$QUERY_DIR/$relpath\033[0m"
  #   has_error=true
  # fi

  # 4. Missing tenant_id = current_tenant_id()
  grep -n -i '\bwhere\b' "$file" | while IFS=: read -r line_num line; do
    if ! grep -iq 'tenant_id\s*=\s*current_tenant_id()' <<<"$line"; then
      echo -e "\n \033[1;33mPossibly missing tenant isolation\033[0m:"
      echo -e "    File: \033[1m$QUERY_DIR/$relpath\033[0m"
      echo -e "    Line $line_num: $line"
      has_error=true
    fi
  done
done

echo ""
if [ "$has_error" = false ]; then
  echo -e "✅ \033[1;32mNo SQLC issues found!\033[0m"
else
  echo -e " \033[1;31mSQLC issues detected — please review the output above.\033[0m"
fi
