#!/bin/bash

# Alpine.js JavaScript Bundle Optimization Script
# Target: <50KB total bundle size

set -e

PROJECT_ROOT="/data/data/com.termux/files/home/project/erp/web"
OUTPUT_DIR="$PROJECT_ROOT/static/js/dist"
TEMP_DIR="$PROJECT_ROOT/build/temp"

echo "🚀 Starting JavaScript bundle optimization..."

# Create output directories
mkdir -p "$OUTPUT_DIR"
mkdir -p "$TEMP_DIR"

# Remove any existing bundles
rm -f "$OUTPUT_DIR"/*.js

echo "📦 Creating optimized bundle..."

# Create main bundle file with core Alpine.js stores
cat > "$TEMP_DIR/bundle.js" << 'EOF'
/**
 * ERP System - Optimized JavaScript Bundle
 * Includes: Alpine.js stores + essential utilities
 * Target: <50KB total size
 */

EOF

# Combine and minify Alpine.js stores (highest priority)
echo "// === ALPINE.JS STORES ===" >> "$TEMP_DIR/bundle.js"

# Add core app store (minified)
echo "// App Store" >> "$TEMP_DIR/bundle.js"
# Remove comments and minify while preserving functionality
sed -e '/^[[:space:]]*\/\*/,/\*\//d' \
    -e '/^[[:space:]]*\/\//d' \
    -e '/^[[:space:]]*$/d' \
    "$PROJECT_ROOT/static/js/stores/app.js" >> "$TEMP_DIR/bundle.js"

echo "" >> "$TEMP_DIR/bundle.js"

# Add user store (minified)
echo "// User Store" >> "$TEMP_DIR/bundle.js"
sed -e '/^[[:space:]]*\/\*/,/\*\//d' \
    -e '/^[[:space:]]*\/\//d' \
    -e '/^[[:space:]]*$/d' \
    "$PROJECT_ROOT/static/js/stores/user.js" >> "$TEMP_DIR/bundle.js"

echo "" >> "$TEMP_DIR/bundle.js"

# Add notifications store (minified)
echo "// Notifications Store" >> "$TEMP_DIR/bundle.js"
sed -e '/^[[:space:]]*\/\*/,/\*\//d' \
    -e '/^[[:space:]]*\/\//d' \
    -e '/^[[:space:]]*$/d' \
    "$PROJECT_ROOT/static/js/stores/notifications.js" >> "$TEMP_DIR/bundle.js"

echo "" >> "$TEMP_DIR/bundle.js"

# Add essential form utilities (minified)
echo "// === ESSENTIAL UTILITIES ===" >> "$TEMP_DIR/bundle.js"
echo "// Form Utilities" >> "$TEMP_DIR/bundle.js"
sed -e '/^[[:space:]]*\/\*/,/\*\//d' \
    -e '/^[[:space:]]*\/\//d' \
    -e '/^[[:space:]]*$/d' \
    "$PROJECT_ROOT/styles/forms.js" >> "$TEMP_DIR/bundle.js"

echo "" >> "$TEMP_DIR/bundle.js"

# Add datatable utilities (minified)
echo "// DataTable Utilities" >> "$TEMP_DIR/bundle.js"
sed -e '/^[[:space:]]*\/\*/,/\*\//d' \
    -e '/^[[:space:]]*\/\//d' \
    -e '/^[[:space:]]*$/d' \
    "$PROJECT_ROOT/styles/datatable.js" >> "$TEMP_DIR/bundle.js"

# Further minify the bundle (basic minification)
echo "🔧 Applying advanced minification..."

# Remove extra whitespace and optimize
sed -i \
    -e 's/[[:space:]]\+/ /g' \
    -e 's/^[[:space:]]*//' \
    -e 's/[[:space:]]*$//' \
    -e '/^$/d' \
    "$TEMP_DIR/bundle.js"

# Copy to final output
cp "$TEMP_DIR/bundle.js" "$OUTPUT_DIR/app-bundle.js"

# Create core bundle (minimal version for critical paths)
echo "📦 Creating core bundle..."

cat > "$OUTPUT_DIR/core-bundle.js" << 'EOF'
/**
 * ERP System - Core Bundle (Critical Path)
 * Essential functionality only
 */
EOF

# Add only app store for core bundle
sed -e '/^[[:space:]]*\/\*/,/\*\//d' \
    -e '/^[[:space:]]*\/\//d' \
    -e '/^[[:space:]]*$/d' \
    -e 's/[[:space:]]\+/ /g' \
    "$PROJECT_ROOT/static/js/stores/app.js" >> "$OUTPUT_DIR/core-bundle.js"

# Check final sizes
echo "📊 Bundle size analysis:"
echo "===================="

FULL_SIZE=$(wc -c < "$OUTPUT_DIR/app-bundle.js")
CORE_SIZE=$(wc -c < "$OUTPUT_DIR/core-bundle.js")

echo "Full bundle: $(echo "scale=1; $FULL_SIZE / 1024" | bc)KB ($FULL_SIZE bytes)"
echo "Core bundle: $(echo "scale=1; $CORE_SIZE / 1024" | bc)KB ($CORE_SIZE bytes)"

# Target check
TARGET_SIZE=51200  # 50KB
if [ $FULL_SIZE -lt $TARGET_SIZE ]; then
    echo "✅ SUCCESS: Bundle is under 50KB target!"
else
    echo "⚠️  WARNING: Bundle is over 50KB target ($(echo "scale=1; $FULL_SIZE / 1024" | bc)KB)"
fi

# Cleanup
rm -rf "$TEMP_DIR"

echo "🎉 Bundle optimization complete!"
echo "Output files:"
echo "  - $OUTPUT_DIR/app-bundle.js (Full bundle)"
echo "  - $OUTPUT_DIR/core-bundle.js (Core bundle)"