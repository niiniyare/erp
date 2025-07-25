#!/bin/bash

cd db/migration || exit 1
echo "[INFO] cd to db/migration"

# Check if dry run
dry_run=false
if [[ "$1" == "--dry-run" ]]; then
  dry_run=true
  echo "[INFO] Running in dry-run mode — no files will be renamed"
fi

# Sorted list of all migration SQL files
files=($(ls -1 | grep -E '^[0-9]{6}_.+\.(up|down)\.sql$' | sort))

echo "[INFO] Found ${#files[@]} files"

start_fixing=false
seen_000021=0
counter=22

for ((i = 0; i < ${#files[@]}; i += 2)); do
  f1="${files[i]}"
  f2="${files[i + 1]}"

  num=$(echo "$f1" | cut -d'_' -f1)
  suffix=$(echo "$f1" | cut -d'_' -f2- | sed 's/\.down\.sql$//')

  echo "[DEBUG] Processing pair: $f1 & $f2 (num=$num, suffix=$suffix)"

  if [[ "$num" == "000021" ]]; then
    ((seen_000021++))
    echo "[DEBUG] Seen 000021 $seen_000021 times"
    if [[ $seen_000021 -eq 1 ]]; then
      echo "[INFO] First 000021 found, skipping"
      continue
    fi
    echo "[INFO] Duplicate 000021 found, start renaming from here"
    start_fixing=true
  fi

  if [[ "$start_fixing" == true ]]; then
    new_num=$(printf "%06d" "$counter")
    new_down="${new_num}_${suffix}.down.sql"
    new_up="${new_num}_${suffix}.up.sql"

    echo "[RENAME] $f1 -> $new_down"
    echo "[RENAME] $f2 -> $new_up"

    if [[ "$dry_run" == false ]]; then
      mv "$f1" "$new_down"
      mv "$f2" "$new_up"
    fi

    ((counter++))
  fi
done
