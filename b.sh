for f in docs/ui/Schema/definitions/*.json; do
  new=$(basename "$f")
  new=$(echo "$new" | sed -E 's/<[^>]*>//g; s/(StringOrNumber|String|Number)//g; s/\.\././g')
  if [ "$new" != "$(basename "$f")" ]; then
    mv "$f" "docs/ui/Schema/definitions/$new"
    echo "Renamed: $(basename "$f") → $new"
  fi
done
