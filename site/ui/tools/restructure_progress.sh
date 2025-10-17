#!/bin/bash

# UI Documentation Restructuring Progress Tracker
# Validates completion of tasks in RoadMap.md

echo "📊 UI Documentation Restructuring Progress"
echo "==========================================="

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TOTAL_TASKS=0
COMPLETED_TASKS=0

# Function to check task completion
check_task() {
    local task_name="$1"
    local verification_command="$2"
    local description="$3"
    
    ((TOTAL_TASKS++))
    
    echo -e "\n${BLUE}🔍 Checking: $task_name${NC}"
    echo "   $description"
    
    if eval "$verification_command" >/dev/null 2>&1; then
        echo -e "   ${GREEN}✅ COMPLETED${NC}"
        ((COMPLETED_TASKS++))
        return 0
    else
        echo -e "   ${RED}❌ NOT COMPLETED${NC}"
        return 1
    fi
}

# Function to check file count
check_file_count() {
    local path="$1"
    local expected_min="$2"
    local expected_max="$3"
    local description="$4"
    
    if [[ -d "$path" ]]; then
        local count=$(find "$path" -type f | wc -l)
        echo -e "   ${BLUE}📁 $description: $count files${NC}"
        if [[ $count -ge $expected_min && $count -le $expected_max ]]; then
            return 0
        fi
    fi
    return 1
}

echo -e "${BLUE}📋 Phase 1: Foundation Setup${NC}"

check_task "New Directory Structure" \
    "[[ -d 'quick-start' && -d 'fundamentals' && -d 'development' && -d 'components' && -d 'schemas' && -d 'integration' && -d 'reference' && -d 'tools' ]]" \
    "All new directories created according to plan"

check_task "Content Inventory Complete" \
    "[[ -f 'CONTENT_INVENTORY.md' && -f 'RoadMap.md' ]]" \
    "Roadmap and content inventory files exist"

check_task "Documentation Accuracy Validation" \
    "[[ -f 'Schema/check_doc_accuracy.sh' && -x 'Schema/check_doc_accuracy.sh' ]]" \
    "Accuracy validation script exists and is executable"

echo -e "\n${BLUE}📋 Phase 2: Core Documentation${NC}"

check_task "Enhanced README.md" \
    "[[ -f 'README.md' && $(wc -c < README.md) -gt 2000 ]]" \
    "README.md exists and is substantial (>2KB)"

check_task "Quick Start Section" \
    "[[ -f 'quick-start/installation.md' && -f 'quick-start/first-component.md' && -f 'quick-start/basic-examples.md' ]]" \
    "All three quick-start files exist"

check_task "Fundamentals Consolidation" \
    "[[ -f 'fundamentals/architecture.md' && -f 'fundamentals/schema-system.md' ]]" \
    "Core fundamentals files exist"

echo -e "\n${BLUE}📋 Phase 3: Flowbite Cleanup${NC}"

# Check if Flowbite has been consolidated
FLOWBITE_FILES=$(find flowbite/ -name "*.md" 2>/dev/null | wc -l)
if [[ $FLOWBITE_FILES -lt 20 ]]; then
    check_task "Flowbite Demo Consolidation" \
        "true" \
        "Flowbite files reduced from 89 to manageable number"
else
    check_task "Flowbite Demo Consolidation" \
        "false" \
        "Flowbite still has $FLOWBITE_FILES files (target: <20)"
fi

check_task "Flowbite Integration Guide" \
    "[[ -f 'integration/flowbite.md' ]]" \
    "Consolidated Flowbite integration guide exists"

echo -e "\n${BLUE}📋 Phase 4: Developer Experience${NC}"

check_task "Development Guides" \
    "[[ -f 'development/creating-components.md' && -f 'development/validation-patterns.md' ]]" \
    "Core development guides exist"

check_task "Component Documentation" \
    "[[ -f 'components/README.md' && -d 'components/atoms' ]]" \
    "Component documentation structure exists"

check_task "Schema System Documentation" \
    "[[ -f 'schemas/README.md' && -d 'schemas/core' ]]" \
    "Schema system documentation organized"

echo -e "\n${BLUE}📋 Phase 5: Integration Guides${NC}"

check_task "Technology Integration" \
    "[[ -f 'integration/htmx.md' || -f 'integration/alpine-js.md' ]]" \
    "At least one technology integration guide exists"

check_task "Reference Documentation" \
    "[[ -d 'reference/api' || -f 'reference/component-registry.md' ]]" \
    "Reference documentation structure exists"

echo -e "\n${BLUE}📋 Phase 6: Quality Assurance${NC}"

# Check for broken links (basic check)
check_task "Internal Link Validation" \
    "! grep -r '\](' . --include='*.md' | grep -E '\]\([^h]' | grep -v '(' | head -1 >/dev/null" \
    "No obvious broken internal links found"

check_task "Schema System Preserved" \
    "check_file_count 'Schema/definitions' 900 950 'Schema definitions'" \
    "All 937+ schemas preserved during restructuring"

echo -e "\n${BLUE}📊 Overall Progress Summary${NC}"
echo "================================"

COMPLETION_PERCENTAGE=$((COMPLETED_TASKS * 100 / TOTAL_TASKS))

echo -e "Tasks Completed: ${GREEN}$COMPLETED_TASKS${NC} / $TOTAL_TASKS"
echo -e "Progress: ${GREEN}$COMPLETION_PERCENTAGE%${NC}"

if [[ $COMPLETION_PERCENTAGE -eq 100 ]]; then
    echo -e "\n${GREEN}🎉 RESTRUCTURING COMPLETE!${NC}"
    echo "All tasks have been completed successfully."
elif [[ $COMPLETION_PERCENTAGE -ge 75 ]]; then
    echo -e "\n${YELLOW}🚀 NEARLY COMPLETE${NC}"
    echo "Most tasks completed, focus on remaining items."
elif [[ $COMPLETION_PERCENTAGE -ge 50 ]]; then
    echo -e "\n${BLUE}📈 GOOD PROGRESS${NC}"
    echo "Halfway through the restructuring process."
else
    echo -e "\n${RED}🚧 EARLY STAGE${NC}"
    echo "Just getting started with the restructuring."
fi

echo -e "\n${BLUE}📝 Next Priority Actions:${NC}"

if [[ ! -d "quick-start" ]]; then
    echo "1. 🏗️  Create new directory structure"
fi

if [[ ! -f "quick-start/installation.md" ]]; then
    echo "2. 📚 Create quick-start content"
fi

if [[ $FLOWBITE_FILES -gt 20 ]]; then
    echo "3. 🧹 Consolidate Flowbite documentation ($FLOWBITE_FILES files)"
fi

if [[ ! -f "integration/htmx.md" ]]; then
    echo "4. 🔗 Create integration guides"
fi

echo -e "\n${BLUE}🔧 To continue restructuring:${NC}"
echo "1. Follow the tasks in RoadMap.md"
echo "2. Use CONTENT_INVENTORY.md for content mapping"
echo "3. Run this script to track progress"
echo "4. Validate with: docs/ui/Schema/check_doc_accuracy.sh"

exit $((TOTAL_TASKS - COMPLETED_TASKS))