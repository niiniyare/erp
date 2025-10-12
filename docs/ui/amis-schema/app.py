import json
import os
from pathlib import Path

def chunk_json_schema(schema_path: str, output_dir: str = "chunked_schemas"):
    """
    Split a large JSON Schema file into smaller sub-schemas based on top-level definitions.
    Works with $defs, definitions, or components/schemas (OpenAPI).
    If no definitions are found, falls back to top-level properties.
    """

    schema_path = Path(schema_path)
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    created_files = []

    # Load schema
    print(f" Loading schema: {schema_path}")
    if not schema_path.exists():
        raise FileNotFoundError(f"Schema file not found: {schema_path}")

    with open(schema_path, "r", encoding="utf-8") as f:
        try:
            schema = json.load(f)
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in {schema_path}: {e}")

    print("Top-level keys:", list(schema.keys()))

    # Detect definition source
    if "$defs" in schema and isinstance(schema["$defs"], dict):
        definitions = schema.pop("$defs")
        defs_source = "$defs"
    elif "definitions" in schema and isinstance(schema["definitions"], dict):
        definitions = schema.pop("definitions")
        defs_source = "definitions"
    elif (
        "components" in schema
        and isinstance(schema["components"], dict)
        and "schemas" in schema["components"]
        and isinstance(schema["components"]["schemas"], dict)
    ):
        definitions = schema["components"].pop("schemas")
        if not schema["components"]:
            schema.pop("components", None)
        defs_source = "components/schemas"
    elif "properties" in schema and isinstance(schema["properties"], dict):
        print("No $defs/definitions/components found. Falling back to 'properties'.")
        definitions = schema.pop("properties")
        defs_source = "properties (fallback)"
    else:
        print(" No definitions or properties found — nothing to chunk.")
        return []

    print(f"Using definitions source: {defs_source} ({len(definitions)} items)")

    # Write each sub-schema file
    for name, sub_schema in definitions.items():
        if not isinstance(sub_schema, dict):
            print(f"⏭️ Skipping {name}: not a valid object.")
            continue

        sub_file = output_dir / f"{name}.json"

        # Ensure $schema and $id exist
        if "$schema" not in sub_schema and "$schema" in schema:
            sub_schema["$schema"] = schema["$schema"]
        if "$id" not in sub_schema:
            sub_schema["$id"] = f"./{name}.json"

        with open(sub_file, "w", encoding="utf-8") as f:
            json.dump(sub_schema, f, indent=2, ensure_ascii=False)

        created_files.append(str(sub_file))
        print(f"✔️ Wrote: {sub_file}")

    # Create main schema referencing sub-schemas
    main_schema = schema.copy()
    main_schema.setdefault("properties", {})
    for name in definitions.keys():
        main_schema["properties"][name] = {"$ref": f"./{name}.json"}

    main_file = output_dir / f"{schema_path.stem}_main.json"
    with open(main_file, "w", encoding="utf-8") as f:
        json.dump(main_schema, f, indent=2, ensure_ascii=False)

    created_files.append(str(main_file))
    print(f"\nMain schema saved: {main_file}")
    print(f"Each sub-schema now referenced via $ref.\n")

    return created_files


# Example usage:
chunk_json_schema("schema.json", output_dir="output")
