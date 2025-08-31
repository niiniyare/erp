#!/bin/bash

# doc-quality-check.sh - Documentation Quality Assurance Framework
# Integrates with CI/CD pipeline for automated validation

set -e

DOCS_DIR="docs"
REPORT_FILE="doc-quality-report.json"
ERROR_COUNT=0

echo "🔍 Running Documentation Quality Checks..."

# 1. MkDocs Build Validation
echo "📋 Testing MkDocs build..."
if ! mkdocs build --strict 2>&1 | tee mkdocs-build.log; then
    ERROR_COUNT=$((ERROR_COUNT + 1))
    echo "❌ MkDocs build failed"
else
    echo "✅ MkDocs build successful"
fi

# 2. Link Validation
echo "🔗 Checking internal links..."
if command -v markdown-link-check &> /dev/null; then
    find "$DOCS_DIR" -name "*.md" -exec markdown-link-check {} \; > link-check.log 2>&1
    if grep -q "ERROR" link-check.log; then
        ERROR_COUNT=$((ERROR_COUNT + 1))
        echo "❌ Broken links found"
    else
        echo "✅ All links valid"
    fi
else
    echo "⚠️ markdown-link-check not installed"
fi

# 3. Navigation Completeness
echo "📚 Validating navigation coverage..."
TOTAL_MD_FILES=$(find "$DOCS_DIR" -name "*.md" | wc -l)
INCLUDED_FILES=$(grep -o '\.md' mkdocs.yml | wc -l)
COVERAGE=$((INCLUDED_FILES * 100 / TOTAL_MD_FILES))

if [ "$COVERAGE" -lt 90 ]; then
    ERROR_COUNT=$((ERROR_COUNT + 1))
    echo "❌ Navigation coverage: ${COVERAGE}% (target: 90%+)"
else
    echo "✅ Navigation coverage: ${COVERAGE}%"
fi

# 4. Generate Quality Report
cat > "$REPORT_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "build_status": "$([ $ERROR_COUNT -eq 0 ] && echo "PASS" || echo "FAIL")",
  "error_count": $ERROR_COUNT,
  "navigation_coverage": $COVERAGE,
  "total_files": $TOTAL_MD_FILES,
  "warnings": $(grep -c "WARNING" mkdocs-build.log || echo 0),
  "checks": {
    "mkdocs_build": "$([ -s mkdocs-build.log ] && echo "COMPLETED" || echo "FAILED")",
    "link_validation": "$([ -f link-check.log ] && echo "COMPLETED" || echo "SKIPPED")",
    "navigation_check": "COMPLETED"
  }
}
EOF

echo "📊 Quality report generated: $REPORT_FILE"

# Exit with appropriate code for CI/CD
if [ $ERROR_COUNT -eq 0 ]; then
    echo "🎉 All quality checks passed!"
    exit 0
else
    echo "💥 $ERROR_COUNT quality issues detected"
    exit 1
fi