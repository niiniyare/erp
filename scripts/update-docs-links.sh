#!/bin/bash

# update-docs-links.sh (Version 2.1 - Production Ready)
#
# A robust script to safely find and rewrite relative Markdown links.
#
# USAGE:
#   ./scripts/update-docs-links.sh         (Runs all stages: backup, manifest, rewrite)
#   ./scripts/update-docs-links.sh --dry-run (Previews changes without writing files)

set -eo pipefail

# --- Configuration ---
DOCS_DIR="docs"
BACKUP_DIR="docs_backup_$(date +%Y%m%d_%H%M%S)"
MANIFEST="docs_path_manifest.csv"
DRY_RUN=0

if [[ "$1" == "--dry-run" ]]; then
  DRY_RUN=1
  echo "[MODE] Dry run enabled. No files will be modified."
fi

# --- Helper Functions ---
log() { echo "[INFO] $1"; }
warn() { echo "[WARN] $1"; }
error() { echo "[ERROR] $1" >&2; exit 1; }

# Function to escape strings for use in sed/perl regex
escape_regex() {
  echo "$1" | sed -e 's/[]\/$.*^|[]/\\&/g'
}

# --- Dependency Check ---
REALPATH_CMD="realpath"
if ! command -v realpath &> /dev/null; then
    if command -v grealpath &> /dev/null; then
        REALPATH_CMD="grealpath"
    else
        error "'realpath' command not found. On macOS, run 'brew install coreutils'. On Debian/Ubuntu, run 'sudo apt-get install realpath'."
    fi
fi

# --- Stage 1: Backup ---
log "Starting documentation restructuring..."
if [[ $DRY_RUN -eq 0 ]]; then
  log "Creating backup of '${DOCS_DIR}' directory to '${BACKUP_DIR}'..."
  [ -d "$DOCS_DIR" ] && cp -r "$DOCS_DIR" "$BACKUP_DIR" || error "Directory '${DOCS_DIR}' not found."
  log "Backup complete."
else
  log "Skipping backup in dry-run mode."
fi

# --- Stage 2: Create Manifest ---
log "Creating file manifest from current structure..."
[ -f "$MANIFEST" ] && rm "$MANIFEST"
echo "old_path,new_path" > "$MANIFEST"

find "$DOCS_DIR" -name "*.md" | while read -r current_filepath; do
  current_path="${current_filepath#$DOCS_DIR/}"
  old_path=""

  case "$current_path" in
    "getting-started/01-developer-quick-start.md") old_path="api/QUICK_START.md" ;;
    "core-concepts/03-abac-and-permissions.md") old_path="module/user/abac/policy.md" ;;
    "how-to-guides/02-migration-guide.md") old_path="MIGRATION_VALIDATION_GUIDE.md" ;;
    "reference/api/financials.md") old_path="module/financial/api-reference.md" ;;
    "architecture/abac-design-and-roadmap.md") old_path="ABAC_IMPLEMENTATION_PLAN.md" ;;
    "contributing/01-best-practices.md") old_path="dev/best-practices.md" ;;
    "architecture/system-overview.md") old_path="module/system-architecture.md" ;;
    *) old_path="$current_path" ;;
  esac

  if [ -n "$old_path" ]; then
    echo "$old_path,$current_path" >> "$MANIFEST"
  fi
done

log "Manifest created at '${MANIFEST}'."

# --- Stage 3: Rewrite Links ---
log "Starting link rewrite process..."
AWK_SCRIPT=$(mktemp)
cat > "$AWK_SCRIPT" << 'AWK_EOF'
BEGIN { in_code_block = 0 }
/^```/ { in_code_block = !in_code_block }
{
    if (in_code_block) {
        print
    } else {
        cmd = "perl -pe '" ENVIRON["SUBSTITUTION"] "'"
        printf "%s", $0 | cmd
        close(cmd)
    }
}
AWK_EOF

tail -n +2 "$MANIFEST" | while IFS=, read -r old_path new_path; do
  if [[ "$old_path" == "$new_path" ]]; then continue; fi
  log "Processing links pointing to '${old_path}' -> should now point to '${new_path}'"

  # Use process substitution for a more robust loop
  while IFS= read -r file_to_update; do
    source_dir=$(dirname "$file_to_update")
    new_relative_link=$($REALPATH_CMD --relative-to="$source_dir" "$DOCS_DIR/$new_path")
    
    escaped_old_path=$(escape_regex "$old_path")
    escaped_new_link=$(escape_regex "$new_relative_link")

    export SUBSTITUTION="s|(\(\s*)${escaped_old_path}(\s*#.*?)?(\s*\))|\1${escaped_new_link}\2\3|g"

    if [[ $DRY_RUN -eq 1 ]]; then
      changed_content=$(awk -f "$AWK_SCRIPT" "$file_to_update")
      original_content=$(cat "$file_to_update")
      if [[ "$changed_content" != "$original_content" ]]; then
        warn "  [DRY RUN] Would update link in: $file_to_update"
      fi
    else
      temp_file=$(mktemp)
      awk -f "$AWK_SCRIPT" "$file_to_update" > "$temp_file"
      mv "$temp_file" "$file_to_update"
      log "  Updated link in: $file_to_update"
    fi
  done < <(grep -rl --include='*.md' "($old_path" "$DOCS_DIR" || true)

done

rm "$AWK_SCRIPT"
echo
log "Script finished."
if [[ $DRY_RUN -eq 0 ]]; then
  log "Backup is in '${BACKUP_DIR}'. Review changes and run validation."
else
  log "Dry run complete. No files were changed."
fi