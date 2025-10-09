import os, re, json, sys

dir_path = sys.argv[1] if len(sys.argv) > 1 else "."
pattern = re.compile(r"[\u4E00-\u9FFF]")  # Chinese Unicode range

results = []
for root, _, files in os.walk(dir_path):
    for file in files:
        if not file.endswith(".json"):
            continue
        path = os.path.join(root, file)
        try:
            with open(path, "r", encoding="utf-8") as f:
                for i, line in enumerate(f, start=1):
                    if pattern.search(line):
                        results.append({
                            "file": path,
                            "line": i,
                            "text": line.strip(),
                            "translation": ""
                        })
        except Exception as e:
            print(f"⚠️ Skipped {path}: {e}", file=sys.stderr)

json.dump(results, sys.stdout, ensure_ascii=False, indent=2)
