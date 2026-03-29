#!/usr/bin/env bash
set -euo pipefail

echo " Searching for service interfaces with //go:generate mockgen..."

grep -rilE 'type\s+[A-Za-z0-9_]*Repository\s+interface' . \
  --exclude-dir=./internal/api/gen \
  --exclude=./db/sqlc/db.go \
  --exclude=./db/sqlc/querier.go \
  --include='*.go' |
  while read -r file; do
    if grep -q '//go:generate.*mockgen' "$file"; then
      echo -e "\n File: $file"
      echo "--------------------------------"
      grep '//go:generate.*mockgen' "$file"
      echo "--------------------------------"
      read -p "❓ Remove this mockgen line from $file? (y/N): " choice
      if [[ "$choice" =~ ^[Yy]$ ]]; then
        sed -i '/\/\/go:generate.*mockgen/d' "$file"
        echo "✅ Removed from $file"
      # else
      #   echo "⏩ Skipped $file"
      fi
    fi
  done
