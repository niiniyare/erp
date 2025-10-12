#!/usr/bin/env bash

current_file=""
grep -rn --include="*.json" --exclude="*.bak" "[一-鿿]" ./output |
  while IFS=: read -r file line content; do
    if [ "$file" != "$current_file" ]; then
      printf "\n%s\n" "═▶ $file ═══════════════════════════════════"
      current_file="$file"
    fi
    printf "  Line %-4s: %s\n" "$line" "$content"
  done
