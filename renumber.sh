#!/bin/bash

cd db/migration || exit 1

# Sort and group files
files=($(ls -1 | grep -E '^[0-9]{6}_.+\.(up|down)\.sql$' | sort))

start_fixing=false
counter=22

for ((i = 0; i < ${#files[@]}; i += 2)); do
  f1="${files[i]}"
  f2="${files[i + 1]}"

  # Extract number and suffix
  num=$(echo "$f1" | cut -d'_' -f1)
  suffix=$(echo "$f1" | cut -d'_' -f2- | sed 's/\.down\.sql$//')

  # Check if this is the second 000021
  if [[ "$num" == "000021" && "$start_fixing" == false ]]; then
    # First 000021 — skip
    start_fixing=true
    continue
  fi

  if [[ "$start_fixing" == true ]]; then
    new_num=$(printf "%06d" "$counter")
    new_down="${new_num}_${suffix}.down.sql"
    new_up="${new_num}_${suffix}.up.sql"

    echo "Renaming: $f1 -> $new_down"
    echo "Renaming: $f2 -> $new_up"

    mv "$f1" "$new_down"
    mv "$f2" "$new_up"

    ((counter++))
  fi
done
