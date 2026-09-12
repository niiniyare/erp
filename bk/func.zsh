function gitall() { 
	git add -A 
	if [ "$1" != "" ] 
	
	then git commit -m "$1" 

	else git commit -m update # default commit message is `update` 
	fi # closing statement of if-else block git push origin HEAD
}

function set_editor() {
  local editors=("nvim" "lvim" "micro")
  
  # Use argument as editor if provided and available
  if [[ -n "$1" ]] && command -v "$1" >/dev/null 2>&1; then
    export EDITOR="$1"
    return
  fi

  # Use pre-set VISUAL or EDITOR if available
  if [[ -n "$VISUAL" ]]; then
    export EDITOR="$VISUAL"
    return
  elif [[ -n "$EDITOR" ]]; then
    export EDITOR="$EDITOR"
    return
  fi

  # Loop through editors and set the first one found
  for editor in "${editors[@]}"; do
    if command -v "$editor" >/dev/null 2>&1; then
      export EDITOR="$editor"
      return
    fi
  done

  # If no editors are found, fallback to a default editor
  export EDITOR="nano"
}

# in your .bashrc/.zshrc/*rc
alias bathelp='bat --plain --language=help'
function help() {
    "$@" --help 2>&1 | bathelp
}

function pg_ctl_cmd() {
  pg_isready && echo "PostgreSQL is already running" || pg_ctl -D ~/pg -l ~/pg/logfile "$1"
}


function interface_any() {
  local dry_run=false
  local verbose=false
  local target="."
  local exts=("md" "go" "templ")

  # Parse CLI args safely
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -d|--dry-run) dry_run=true; shift ;;
      -v|--verbose) verbose=true; shift ;;
      -p|--path)
        target="$2"
        if [[ -z "$target" ]]; then
          echo "Error: --path requires an argument" >&2
          return 1
        fi
        shift 2 ;;
      -e|--ext)
        exts=()
        shift
        while [[ $# -gt 0 && "$1" != -* ]]; do
          exts+=("${1#.}") # strip dot if user typed .go
          shift
        done
        ;;
      -h|--help)
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
        return 0 ;;
      *)
        echo "Unknown option: $1" >&2
        return 1 ;;
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
        sed 's/interface{}/any/g' "$target" > "$temp_file"
        mv "$temp_file" "$target"
      fi
    fi
    return 0
  fi

  # Build name expression for find
  local name_expr=()
  for ext in "${exts[@]}"; do
    name_expr+=( -name "*.${ext}" )
    if [[ ${#name_expr[@]} -lt $(( ${#exts[@]} * 2 - 1 )) ]]; then
      name_expr+=( -o )
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
        sed 's/interface{}/any/g' "$file" > "$temp_file"
        mv "$temp_file" "$file"
      done
  fi
}

function gocover() {
    local pkg=${1:-./...}
    local output=${2:-coverage.html}
    
    go test -cover -v "$pkg" -coverprofile=coverage.out
    if [[ $? -eq 0 ]]; then
        go tool cover -html=coverage.out -o "$output"
        echo "Coverage report generated: $output"
        
        # Auto-open after 3 seconds unless user presses a key
        echo -n "Opening coverage report in 3 seconds... (Press any key to cancel)"
        if read -r -t 3 -n 1; then
            echo -e "\nCoverage report saved to: $output"
        else
            echo -e "\nOpening coverage report..."
            if command -v termux-open >/dev/null 2>&1; then
                termux-open "$output"
            elif command -v open >/dev/null 2>&1; then
                open "$output"
            elif command -v xdg-open >/dev/null 2>&1; then
                xdg-open "$output"
            else
                echo "Please open '$output' manually."
            fi
        fi
    else
        echo "Tests failed, skipping coverage report generation"
        return 1
    fi
}


emoji_rm (){
 rg -l "[\x{1F300}-\x{1FAFF}]" | fzf -m | while read -r file; do
  sd "[\x{1F300}-\x{1FAFF}]" "" "$file"
done 
}
