#!/usr/bin/env bash

DIR="db/migration"

echo "Migration | File | Line | Type | Object"
echo "---------------------------------------------------------------"

for file in "$DIR"/*.sql; do
  awk -v fname="$(basename "$file")" '
    BEGIN {
      IGNORECASE = 1
    }

    /^[[:space:]]*CREATE[[:space:]]+(OR[[:space:]]+REPLACE[[:space:]]+)?(TABLE|VIEW|FUNCTION|TRIGGER|INDEX|SEQUENCE)[[:space:]]+/ {
      type = toupper($3)
      if ($3 == "OR") {
        type = toupper($5)
        name = $6
      } else {
        name = $4
      }

      # clean name (remove parentheses or schema extras)
      gsub(/\(.*/, "", name)

      # migration number = first part of filename
      split(fname, parts, "_")
      mig = parts[1]

      printf "%s | %s | %d | %s | %s\n", mig, fname, NR, type, name
      exit
    }
  ' "$file"
done
