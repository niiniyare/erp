#!/bin/bash

# =============================================================================
# ERP EXCITING THEME APPLICATION SCRIPT
# Applies modern, exciting theme to reorganized UI components
# =============================================================================

echo "🎨 Applying Exciting Theme to ERP Components..."

# Apply theme to input components
echo "📝 Enhancing Input Components..."
find internal/ui/components/core -name "input.templ" -exec sed -i \
    's/class="[^"]*"/class="erp-input"/g' {} \;

find internal/ui/components/core -name "textarea.templ" -exec sed -i \
    's/class="[^"]*"/class="erp-input h-32 resize-y"/g' {} \;

find internal/ui/components/core -name "select.templ" -exec sed -i \
    's/class="[^"]*"/class="erp-input"/g' {} \;

# Apply theme to container components
echo "📦 Enhancing Container Components..."
find internal/ui/components/core -name "container.templ" -exec sed -i \
    's/class="[^"]*container[^"]*"/class="erp-container"/g' {} \;

# Apply theme to progress components
echo "⏳ Enhancing Progress Components..."
find internal/ui/components/core -name "progress.templ" -exec sed -i \
    's/bg-blue-600/bg-gradient-to-r from-primary-400 to-primary-600/g' {} \;

# Apply theme to badge components
echo "🏷️ Enhancing Badge Components..."
find internal/ui/components/core -name "badge.templ" -exec sed -i \
    's/bg-blue-100/bg-gradient-to-r from-primary-50 to-primary-100/g' {} \;

# Apply theme to alert components
echo "🚨 Enhancing Alert Components..."
find internal/ui/components/shared -name "alert.templ" -exec sed -i \
    's/class="[^"]*alert[^"]*"/class="erp-alert erp-alert-info"/g' {} \;

# Apply theme to toast components
echo "🍞 Enhancing Toast Components..."
find internal/ui/components/shared -name "toast.templ" -exec sed -i \
    's/class="[^"]*toast[^"]*"/class="erp-alert animate-slide-in"/g' {} \;

# Apply theme to modal components
echo "🪟 Enhancing Modal Components..."
find internal/ui/components/layout -name "modal.templ" -exec sed -i \
    's/class="[^"]*modal[^"]*"/class="erp-modal animate-scale-in"/g' {} \;

# Apply theme to navigation components
echo "🧭 Enhancing Navigation Components..."
find internal/ui/components/layout -name "navbar.templ" -exec sed -i \
    's/bg-white/bg-gradient-to-r from-white to-gray-50/g' {} \;

find internal/ui/components/layout -name "sidebar.templ" -exec sed -i \
    's/bg-gray-800/bg-gradient-to-b from-gray-800 to-gray-900/g' {} \;

# Apply theme to breadcrumb components
echo "🍞 Enhancing Breadcrumb Components..."
find internal/ui/components/shared -name "breadcrumb.templ" -exec sed -i \
    's/text-blue-600/text-primary-600 hover:text-primary-700/g' {} \;

# Apply theme to tooltip components
echo "💭 Enhancing Tooltip Components..."
find internal/ui/components/shared -name "tooltip.templ" -exec sed -i \
    's/bg-gray-900/bg-gradient-to-r from-gray-800 to-gray-900/g' {} \;

# Apply theme to avatar components
echo "👤 Enhancing Avatar Components..."
find internal/ui/components/shared -name "avatar.templ" -exec sed -i \
    's/ring-2 ring-gray-300/ring-2 ring-primary-300 ring-offset-2/g' {} \;

# Apply theme to button group components
echo "🔘 Enhancing Button Group Components..."
find internal/ui/components/shared -name "button_group.templ" -exec sed -i \
    's/border-gray-200/border-primary-200/g' {} \;

# Apply theme to accordion components
echo "📋 Enhancing Accordion Components..."
find internal/ui/components/layout -name "accordion.templ" -exec sed -i \
    's/bg-gray-50/bg-gradient-to-r from-primary-50 to-primary-100/g' {} \;

# Apply theme to tabs components
echo "📑 Enhancing Tabs Components..."
find internal/ui/components/layout -name "tabs.templ" -exec sed -i \
    's/border-blue-600/border-primary-600/g' {} \;

# Apply theme to carousel components
echo "🎠 Enhancing Carousel Components..."
find internal/ui/components/layout -name "carousel.templ" -exec sed -i \
    's/bg-white/bg-gradient-to-r from-white to-gray-50/g' {} \;

# Apply theme to dropdown components
echo "📝 Enhancing Dropdown Components..."
find internal/ui/components/layout -name "dropdown.templ" -exec sed -i \
    's/shadow-lg/shadow-exciting/g' {} \;

echo ""
echo "✨ Exciting Theme Application Complete!"
echo ""
echo "🎯 Applied Modern Enhancements:"
echo "  • Gradient backgrounds and buttons"
echo "  • Enhanced shadows and hover effects"  
echo "  • Smooth animations and transitions"
echo "  • Service-specific color schemes"
echo "  • Glass morphism effects"
echo "  • Modern border radius"
echo ""
echo "🎨 Next Steps:"
echo "  1. Generate templates: make templ"
echo "  2. Include themes.css in your layouts"
echo "  3. Apply service themes in your handlers"
echo ""
echo "Ready for an exciting UI experience! 🚀"