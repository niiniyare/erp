#!/usr/bin/env bash
set -euo pipefail

SRC_DIR="db/migration"
OUT_FILE="db/base.up.sql"

rm -f "$OUT_FILE"
touch "$OUT_FILE"

for file in $(find "$SRC_DIR" -type f -name "*.up.sql" | sort); do
  echo "-- BEGIN FILE: $file" >>"$OUT_FILE"
  cat "$file" >>"$OUT_FILE"
  echo -e "\n-- END FILE: $file\n" >>"$OUT_FILE"
done

echo "✅ Built $OUT_FILE with all migrations"
