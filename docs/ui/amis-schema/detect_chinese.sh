#!/usr/bin/env bash
# Detect Chinese text in JSON files and output JSON-formatted results.
# Usage: ./detect_chinese_json.sh [directory]

DIR=${1:-.}
FIRST=1

echo "["

# Find all .json files (recursively)
find "$DIR" -type f -name "*.json" | while read -r file; do
  # Find Chinese characters (Unicode range \u4E00–\u9FFF)
  grep -nP "[\x{4E00}-\x{9FFF}]" "$file" | while IFS=: read -r line_num line_text; do
    # Escape double quotes in the text
    esc_text=$(echo "$line_text" | sed 's/"/\\"/g')
    # Add comma between JSON objects (not before the first one)
    if [ $FIRST -eq 0 ]; then
      echo ","
    fi
    FIRST=0
    # Print JSON object
    echo "  {"
    echo "    \"file\": \"${file}\","
    echo "    \"line\": ${line_num},"
    echo "    \"text\": \"${esc_text}\","
    echo "    \"translation\": \"\""
    echo -n "  }"
  done
done

echo
echo "]"
## Scan all JSON files and show lines containing Chinese characters
#
#DIR=${1:-.}
#
#echo " Scanning directory: $DIR"
#echo
#
#for file in "$DIR"/*.json; do
#  matches=$(grep -nP "[\x{4E00}-\x{9FFF}]" "$file")
#  if [ -n "$matches" ]; then
#    echo "------------------------------------"
#    echo "------------------------------------"
#
#    echo "⚠️  Chinese characters found in: $file"
#    echo "$matches"
#    echo "------------------------------------"
#  fi
#done
