#!/usr/bin/env bash
# Run this script once to complete the legacy documentation migration.
# Purpose: rename docs/ → docs_legacy_framework_a/ and add deprecation
# banners to every Markdown file inside it.
#
# Usage: bash LEGACY_MIGRATION.sh
# Then: rm LEGACY_MIGRATION.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
OLD_DIR="$REPO_ROOT/docs"
NEW_DIR="$REPO_ROOT/docs_legacy_framework_a"

# --- Step 1: rename ---
if [ -d "$NEW_DIR" ]; then
  echo "Already renamed: $NEW_DIR exists. Skipping mv."
elif [ -d "$OLD_DIR" ]; then
  mv "$OLD_DIR" "$NEW_DIR"
  echo "Renamed: docs/ → docs_legacy_framework_a/"
else
  echo "ERROR: neither docs/ nor docs_legacy_framework_a/ found. Nothing to do."
  exit 1
fi

# --- Step 2: prepend banner to every .md file ---
BANNER='> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

'

count=0
while IFS= read -r -d '' f; do
  # Skip if banner already present
  if head -1 "$f" | grep -q 'LEGACY DOCUMENTATION'; then
    echo "  already patched: $f"
    continue
  fi
  tmp=$(mktemp)
  printf '%s' "$BANNER" | cat - "$f" > "$tmp" && mv "$tmp" "$f"
  echo "  patched: $f"
  count=$((count + 1))
done < <(find "$NEW_DIR" -name "*.md" -print0)

echo ""
echo "Done. $count files patched."
echo "You may now: rm $REPO_ROOT/LEGACY_MIGRATION.sh"
