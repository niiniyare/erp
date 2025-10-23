#!/usr/bin/env bash

interface_any() {
  local dry_run=false
  local verbose=false
  local target="."
  local exts=("md" "go" "templ")

  # Parse CLI args safely
  while [[ $# -gt 0 ]]; do
    case "$1" in
    -d | --dry-run)
      dry_run=true
      shift
      ;;
    -v | --verbose)
      verbose=true
      shift
      ;;
    -p | --path)
      target="$2"
      if [[ -z "$target" ]]; then
        echo "Error: --path requires an argument" >&2
        return 1
      fi
      shift 2
      ;;
    -e | --ext)
      exts=()
      shift
      while [[ $# -gt 0 && "$1" != -* ]]; do
        exts+=("${1#.}") # strip dot if user typed .go
        shift
      done
      ;;
    -h | --help)
      cat <<'EOF'
Usage: interface_any [options]

Options:
  -d, --dry-run         Show which files would be modified (no changes made)
  -v, --verbose         Print each matched file as processed
  -p, --path <path>     Limit search to a directory or a single file
  -e, --ext <exts...>   Restrict to specific file extensions (default: md go templ)
  -h, --help            Show this help message

Examples:
  interface_any -d                   # dry run on all default extensions
  interface_any -v                   # verbose actual run
  interface_any -p ./src -e go templ # only .go and .templ in ./src
  interface_any -p main.go           # single file
EOF
      return 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      return 1
      ;;
    esac
  done

  # If path is a single file
  if [[ -f "$target" ]]; then
    if grep -q 'interface{}' "$target"; then
      if $dry_run; then
        echo "[DRY-RUN] Would modify: $target"
      else
        $verbose && echo "Modifying: $target"
        # Use a temporary file for sed to avoid any flag issues
        local temp_file
        temp_file=$(mktemp)
        sed 's/interface{}/any/g' "$target" >"$temp_file"
        mv "$temp_file" "$target"
      fi
    fi
    return 0
  fi

  # Build name expression for find
  local name_expr=()
  for ext in "${exts[@]}"; do
    name_expr+=(-name "*.${ext}")
    if [[ ${#name_expr[@]} -lt $((${#exts[@]} * 2 - 1)) ]]; then
      name_expr+=(-o)
    fi
  done

  # Execute find + grep + sed logic
  if $dry_run; then
    $verbose && echo "[INFO] Dry run mode enabled"
    find "$target" -type f \( "${name_expr[@]}" \) \
      ! -name "*_templ.go" ! -name "*.sqlc.go" \
      -not -path "*/.git/*" -not -path "*/vendor/*" -not -path "*/node_modules/*" \
      -exec grep -l 'interface{}' {} \; | while read -r file; do
      echo "[DRY-RUN] Would modify: $file"
    done
  else
    find "$target" -type f \( "${name_expr[@]}" \) \
      ! -name "*_templ.go" ! -name "*.sqlc.go" \
      -not -path "*/.git/*" -not -path "*/vendor/*" -not -path "*/node_modules/*" \
      -exec grep -l 'interface{}' {} \; | while read -r file; do
      if $verbose; then
        echo "Modifying: $file"
      fi
      # Use temporary file approach for all files
      local temp_file
      temp_file=$(mktemp)
      sed 's/interface{}/any/g' "$file" >"$temp_file"
      mv "$temp_file" "$file"
    done
  fi
}
