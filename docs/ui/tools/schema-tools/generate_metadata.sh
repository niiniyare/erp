#!/bin/bash

# Schema Metadata Generation Script
# Generates basic metadata files for all schemas in the organized structure

echo "📋 Generating schema metadata files..."

# Function to determine category and subcategory from path
get_metadata() {
    local schema_path="$1"
    local schema_name="$2"
    
    # Extract category and subcategory from path
    if [[ "$schema_path" == *"core/layout"* ]]; then
        echo "core" "layout"
    elif [[ "$schema_path" == *"core/typography"* ]]; then
        echo "core" "typography"
    elif [[ "$schema_path" == *"core/color"* ]]; then
        echo "core" "color"
    elif [[ "$schema_path" == *"core/spacing"* ]]; then
        echo "core" "spacing"
    elif [[ "$schema_path" == *"core/datatypes"* ]]; then
        echo "core" "datatypes"
    elif [[ "$schema_path" == *"core/compatibility"* ]]; then
        echo "core" "compatibility"
    elif [[ "$schema_path" == *"components/atoms"* ]]; then
        echo "atoms" "forms"
    elif [[ "$schema_path" == *"components/molecules"* ]]; then
        echo "molecules" "forms"
    elif [[ "$schema_path" == *"components/organisms"* ]]; then
        echo "organisms" "forms"
    elif [[ "$schema_path" == *"components/templates"* ]]; then
        echo "templates" "layouts"
    elif [[ "$schema_path" == *"interactions/forms"* ]]; then
        echo "interactions" "forms"
    elif [[ "$schema_path" == *"interactions/navigation"* ]]; then
        echo "interactions" "navigation"
    elif [[ "$schema_path" == *"interactions/data"* ]]; then
        echo "interactions" "data"
    else
        echo "utility" "helpers"
    fi
}

# Function to generate human-readable title from schema name
generate_title() {
    local name="$1"
    # Remove Schema suffix and add spaces before capitals
    echo "$name" | sed 's/Schema$//' | sed 's/\([a-z]\)\([A-Z]\)/\1 \2/g'
}

# Function to generate description based on schema name and category
generate_description() {
    local name="$1"
    local category="$2"
    local subcategory="$3"
    
    case "$category" in
        "atoms")
            echo "A fundamental UI element for $subcategory functionality"
            ;;
        "molecules")
            echo "A composite UI component combining multiple atoms for $subcategory operations"
            ;;
        "organisms")
            echo "A complex UI component providing complete $subcategory functionality"
            ;;
        "templates")
            echo "A page-level template for $subcategory structure"
            ;;
        "interactions")
            echo "Interactive behavior schema for $subcategory operations"
            ;;
        "core")
            echo "Core CSS property definition for $subcategory styling"
            ;;
        "utility")
            echo "Utility schema for $subcategory support"
            ;;
        *)
            echo "Schema definition for $name"
            ;;
    esac
}

# Generate metadata for all JSON schema files
find . -name "*.json" -not -name "*.meta.json" -not -name "schema_metadata_template.json" | while read schema_file; do
    # Skip if metadata already exists
    meta_file="${schema_file%.json}.meta.json"
    if [[ -f "$meta_file" ]]; then
        echo "⏭️  Skipping $schema_file (metadata exists)"
        continue
    fi
    
    # Extract schema name and path info
    schema_name=$(basename "$schema_file" .json)
    schema_path=$(dirname "$schema_file")
    
    # Get category and subcategory
    read category subcategory <<< $(get_metadata "$schema_path" "$schema_name")
    
    # Generate human-readable title and description
    title=$(generate_title "$schema_name")
    description=$(generate_description "$schema_name" "$category" "$subcategory")
    
    # Create basic metadata file
    cat > "$meta_file" << EOF
{
  "\$schema": "../../schema_metadata_template.json",
  "meta": {
    "title": "$title",
    "description": "$description",
    "category": "$category",
    "subcategory": "$subcategory",
    "dependencies": [],
    "examples": [],
    "accessibility": {
      "wcagLevel": "AA",
      "keyboardSupport": true,
      "screenReaderSupport": true,
      "ariaRequirements": [],
      "notes": "Standard accessibility requirements apply"
    },
    "performance": {
      "renderCost": "low",
      "memoryCost": "low",
      "optimizations": [],
      "notes": "Standard performance characteristics"
    },
    "version": "1.0.0",
    "deprecated": false,
    "createdDate": "$(date -I)",
    "lastModified": "$(date -I)",
    "author": "ERP UI Team",
    "tags": [
      "$(echo $subcategory | tr '[:upper:]' '[:lower:]')",
      "$(echo $schema_name | tr '[:upper:]' '[:lower:]')"
    ],
    "framework": {
      "templ": {
        "componentFile": "",
        "propsInterface": ""
      },
      "htmx": {
        "supported": false,
        "attributes": []
      },
      "alpine": {
        "directives": []
      }
    },
    "testing": {
      "unitTests": "",
      "integrationTests": "",
      "visualTests": "",
      "testCoverage": 0
    }
  }
}
EOF

    echo "✅ Generated metadata for $schema_name"
done

echo "📊 Metadata generation complete!"
echo "Generated metadata files: $(find . -name "*.meta.json" | wc -l)"