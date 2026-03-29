#!/bin/bash

# Documentation Accuracy Validation Script
# Verifies all claims in documentation against actual project implementation

echo " Starting documentation accuracy validation..."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

# Function to check file existence
check_file() {
    local file="$1"
    local description="$2"
    
    if [[ -f "$file" ]]; then
        echo -e "${GREEN}✅ $description${NC}: $file"
        return 0
    else
        echo -e "${RED}❌ $description${NC}: $file (MISSING)"
        ((ERRORS++))
        return 1
    fi
}

# Function to check directory existence
check_directory() {
    local dir="$1"
    local description="$2"
    
    if [[ -d "$dir" ]]; then
        local count=$(find "$dir" -type f | wc -l)
        echo -e "${GREEN}✅ $description${NC}: $dir ($count files)"
        return 0
    else
        echo -e "${RED}❌ $description${NC}: $dir (MISSING)"
        ((ERRORS++))
        return 1
    fi
}

echo " Validating critical file paths mentioned in documentation..."

# Check critical Go files
check_file "web/engine/schema_factory.go" "Schema Factory"
check_file "web/engine/schema_templ.go" "Schema Templ Renderer"
check_file "web/engine/enhanced_registry.go" "Enhanced Registry"
check_file "web/engine/css_integration.go" "CSS Integration"
check_file "pkg/schema/ui/css/properties.go" "CSS Properties"

# Check component directories
check_directory "web/components" "Components Directory"
check_directory "web/components/atoms" "Atoms Components"
check_directory "web/components/molecules" "Molecules Components"
check_directory "web/components/organisms" "Organisms Components"

# Check schema directories
check_directory "docs/ui/Schema/definitions" "Schema Definitions"
check_directory "docs/ui/Schema/definitions/core" "Core Schemas"
check_directory "docs/ui/Schema/definitions/components" "Component Schemas"

echo ""
echo " Validating schema counts..."

# Get actual schema counts
if [[ -d "docs/ui/Schema/definitions" ]]; then
    TOTAL_SCHEMAS=$(find docs/ui/Schema/definitions -name "*.json" | wc -l)
    CORE_SCHEMAS=$(find docs/ui/Schema/definitions/core -name "*.json" 2>/dev/null | wc -l)
    COMPONENT_SCHEMAS=$(find docs/ui/Schema/definitions/components -name "*.json" 2>/dev/null | wc -l)
    INTERACTION_SCHEMAS=$(find docs/ui/Schema/definitions/interactions -name "*.json" 2>/dev/null | wc -l)
    UTILITY_SCHEMAS=$(find docs/ui/Schema/definitions/utility -name "*.json" 2>/dev/null | wc -l)
    
    echo -e "${GREEN} Actual Schema Counts:${NC}"
    echo "   • Total: $TOTAL_SCHEMAS schemas"
    echo "   • Core: $CORE_SCHEMAS schemas"
    echo "   • Components: $COMPONENT_SCHEMAS schemas"
    echo "   • Interactions: $INTERACTION_SCHEMAS schemas"
    echo "   • Utility: $UTILITY_SCHEMAS schemas"
    
    # Check if previously documented count (913) matches reality
    if [[ $TOTAL_SCHEMAS -eq 913 ]]; then
        echo -e "${GREEN}✅ Schema count matches documentation${NC}"
    else
        echo -e "${YELLOW}⚠️  Schema count discrepancy${NC}: Documentation claims 913+, actual is $TOTAL_SCHEMAS"
        ((WARNINGS++))
    fi
else
    echo -e "${RED}❌ Schema definitions directory not found${NC}"
    ((ERRORS++))
fi

echo ""
echo " Validating specific schema files mentioned in documentation..."

# Check for schemas specifically mentioned in documentation
DOCUMENTED_SCHEMAS=(
    "ButtonSchema.json"
    "CheckboxControlSchema.json"
    "ButtonGroupSchema.json"
    "InputControlSchema.json"
    "FormSchema.json"
)

for schema in "${DOCUMENTED_SCHEMAS[@]}"; do
    # Search for the schema in any subdirectory
    SCHEMA_PATH=$(find docs/ui/Schema/definitions -name "$schema" 2>/dev/null)
    
    if [[ -n "$SCHEMA_PATH" ]]; then
        echo -e "${GREEN}✅ Schema found${NC}: $schema at $SCHEMA_PATH"
    else
        echo -e "${RED}❌ Schema missing${NC}: $schema (referenced in documentation)"
        ((ERRORS++))
    fi
done

echo ""
echo " Validating Go implementations..."

# Check for Go structs/interfaces mentioned in documentation
GO_TYPES=(
    "SchemaFactory"
    "SchemaTemplRenderer" 
    "EnhancedComponentRegistry"
    "TemplComponent"
    "ComponentRenderer"
)

for type_name in "${GO_TYPES[@]}"; do
    # Search for type definition in Go files
    if grep -r "type $type_name" . --include="*.go" >/dev/null 2>&1; then
        FOUND_FILE=$(grep -r "type $type_name" . --include="*.go" | head -1 | cut -d: -f1)
        echo -e "${GREEN}✅ Go type found${NC}: $type_name in $FOUND_FILE"
    else
        echo -e "${RED}❌ Go type missing${NC}: $type_name (referenced in documentation)"
        ((ERRORS++))
    fi
done

echo ""
echo " Summary Report:"
echo "==================="

if [[ $ERRORS -eq 0 ]]; then
    echo -e "${GREEN}✅ Documentation is accurate!${NC}"
    echo "   • All file paths verified"
    echo "   • All Go types found"
    echo "   • Schema organization confirmed"
else
    echo -e "${RED}❌ Found $ERRORS accuracy issues${NC}"
    echo "   • Please fix missing files/types before publishing documentation"
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}⚠️  Found $WARNINGS warnings${NC}"
    echo "   • Consider updating documentation for accuracy"
fi

echo ""
echo " Next Steps:"
if [[ $ERRORS -gt 0 ]]; then
    echo "1.  Fix missing files/implementations"
    echo "2.  Update documentation to match reality"
    echo "3.  Re-run this validation script"
    echo "4. ✅ Proceed with documentation updates"
else
    echo "1. ✅ Documentation is ready for updates"
    echo "2.  Use actual counts from this report"
    echo "3.  Run this script after any changes"
fi

# Exit with error code if issues found
exit $ERRORS