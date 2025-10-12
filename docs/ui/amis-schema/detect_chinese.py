import os, re, json, sys

# --- Settings ---
dir_path = sys.argv[1] if len(sys.argv) > 1 else "./output"
pattern_chinese = re.compile(r"[\u4E00-\u9FFF]")  # Any Chinese character
# Match all "key": "value" pairs containing Chinese
pattern_kv = re.compile(r'"([^"]+)"\s*:\s*"([^"]*[\u4E00-\u9FFF][^"]*)"')

results = []

for root, _, files in os.walk(dir_path):
    for file in files:
        if not file.endswith(".json"):
            continue
        path = os.path.join(root, file)
        try:
            with open(path, "r", encoding="utf-8") as f:
                for i, line in enumerate(f, start=1):
                    # Skip lines without any Chinese
                    if not pattern_chinese.search(line):
                        continue

                    # Find *all* key/value pairs with Chinese
                    matches = list(pattern_kv.finditer(line))
                    if matches:
                        for m in matches:
                            tag, value = m.groups()
                            results.append({
                                "file": path,
                                "line": i,
                                # "text": {
                                "tag": tag,
                                "value": value,
                                "translation": ""

                                # },
                                # "translation": ""
                            })
                    else:
                        # fallback if Chinese exists but not in "key": "value" form
                        results.append({
                            "file": path,
                            "line": i,
                            "text": {
                                "tag": None,
                                "value": line.strip()
                            },
                            "translation": ""
                        })

        except Exception as e:
            print(f"⚠️ Skipped {path}: {e}", file=sys.stderr)

# Output nicely formatted JSON
json.dump(results, sys.stdout, ensure_ascii=False, indent=2)


# import os, re, json, sys
#
# dir_path = sys.argv[1] if len(sys.argv) > 1 else "."
# pattern = re.compile(r"[\u4E00-\u9FFF]")  # Chinese Unicode range
#
# results = []
# for root, _, files in os.walk(dir_path):
#     for file in files:
#         if not file.endswith(".json"):
#             continue
#         path = os.path.join(root, file)
#         try:
#             with open(path, "r", encoding="utf-8") as f:
#                 for i, line in enumerate(f, start=1):
#                     if pattern.search(line):
#                         results.append({
#                             "file": path,
#                             "line": i,
#                             "text": line.strip(),
#                             "translation": ""
#                         })
#         except Exception as e:
#             print(f"Skipped {path}: {e}", file=sys.stderr)
#
# json.dump(results, sys.stdout, ensure_ascii=False, indent=2)
