package sidebar

import (
	"strconv"
	"strings"
	"github.com/niiniyare/erp/web/components/atoms"
)

// getSidebarClasses returns CSS classes for the main sidebar
func getSidebarClasses(props SidebarProps) string {
	classes := []string{
		"sidebar",
		"relative",
		"h-full",
		"bg-white",
		"border-r",
		"border-gray-200",
		"dark:bg-gray-800",
		"dark:border-gray-700",
		"transition-all",
		"duration-300",
		"ease-in-out",
	}
	
	// Width classes
	width := props.Width
	if width == "" {
		width = "w-64" // Default width
	}
	classes = append(classes, width)
	
	// Position classes
	if props.Position == "right" {
		classes = append(classes, "border-l", "border-r-0")
	}
	
	// Variant classes
	switch props.Variant {
	case "minimal":
		classes = append(classes, "shadow-none", "border-0")
	case "bordered":
		classes = append(classes, "border-2")
	default:
		classes = append(classes, "shadow-sm")
	}
	
	// Overlay classes for mobile
	if props.Overlay {
		classes = append(classes, 
			"fixed", 
			"inset-y-0", 
			"left-0", 
			"z-50", 
			"lg:relative", 
			"lg:translate-x-0",
		)
	}
	
	// Theme classes
	switch props.Theme {
	case "dark":
		classes = append(classes, "dark")
	}
	
	if props.Class != "" {
		classes = append(classes, props.Class)
	}
	
	return strings.Join(classes, " ")
}

// getSidebarItemClasses returns CSS classes for sidebar items
func getSidebarItemClasses(item SidebarItem, level int) string {
	baseClasses := []string{
		"group",
		"flex",
		"items-center",
		"px-2",
		"py-2",
		"text-sm",
		"font-medium",
		"rounded-lg",
		"transition-colors",
		"duration-150",
	}
	
	// Indentation for nested items
	if level > 0 {
		indentClass := "ml-" + strconv.Itoa(level*4)
		baseClasses = append(baseClasses, indentClass)
	}
	
	// State classes
	if item.Active {
		baseClasses = append(baseClasses,
			"bg-blue-100",
			"text-blue-700",
			"dark:bg-blue-900",
			"dark:text-blue-200",
		)
	} else if item.Highlight {
		baseClasses = append(baseClasses,
			"bg-yellow-50",
			"text-yellow-800",
			"dark:bg-yellow-900/20",
			"dark:text-yellow-200",
		)
	} else {
		baseClasses = append(baseClasses,
			"text-gray-700",
			"hover:bg-gray-100",
			"hover:text-gray-900",
			"dark:text-gray-300",
			"dark:hover:bg-gray-700",
			"dark:hover:text-white",
		)
	}
	
	// Disabled state
	if item.Disabled {
		baseClasses = append(baseClasses,
			"opacity-50",
			"cursor-not-allowed",
			"pointer-events-none",
		)
	}
	
	if item.Class != "" {
		baseClasses = append(baseClasses, item.Class)
	}
	
	return strings.Join(baseClasses, " ")
}

// getSidebarHeaderClasses returns CSS classes for section headers
func getSidebarHeaderClasses(level int) string {
	classes := []string{
		"px-2",
		"py-3",
	}
	
	if level > 0 {
		classes = append(classes, "mt-4")
	}
	
	return strings.Join(classes, " ")
}

// getSidebarBadgeClasses returns CSS classes for badges
func getSidebarBadgeClasses(color string) string {
	baseClasses := "inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
	
	switch color {
	case "red":
		return baseClasses + " bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300"
	case "blue":
		return baseClasses + " bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300"
	case "green":
		return baseClasses + " bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300"
	case "yellow":
		return baseClasses + " bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300"
	case "purple":
		return baseClasses + " bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300"
	case "pink":
		return baseClasses + " bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-300"
	default:
		return baseClasses + " bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
	}
}

// getSidebarIconSize returns appropriate icon size based on item level
func getSidebarIconSize(level int) atoms.IconSize {
	if level > 0 {
		return atoms.IconSizeXS
	}
	return atoms.IconSizeSM
}

// getSidebarIconClasses returns CSS classes for icons
func getSidebarIconClasses(item SidebarItem) string {
	if item.Active {
		return "text-blue-600 dark:text-blue-400"
	} else if item.Highlight {
		return "text-yellow-600 dark:text-yellow-400"
	}
	return "text-gray-400 group-hover:text-gray-500 dark:text-gray-500 dark:group-hover:text-gray-400"
}

// getSidebarUserClasses returns CSS classes for user info
func getSidebarUserClasses(compact bool) string {
	baseClasses := []string{
		"flex",
		"items-center",
		"p-2",
		"rounded-lg",
		"hover:bg-gray-100",
		"dark:hover:bg-gray-700",
		"transition-colors",
		"duration-150",
		"cursor-pointer",
	}
	
	if compact {
		baseClasses = append(baseClasses, "justify-center")
	}
	
	return strings.Join(baseClasses, " ")
}

// getSidebarFooterItemClasses returns CSS classes for footer items
func getSidebarFooterItemClasses(item SidebarItem) string {
	baseClasses := []string{
		"group",
		"flex",
		"items-center",
		"px-2",
		"py-1.5",
		"text-sm",
		"rounded-lg",
		"hover:bg-gray-100",
		"dark:hover:bg-gray-700",
		"transition-colors",
		"duration-150",
	}
	
	if item.Active {
		baseClasses = append(baseClasses,
			"bg-gray-100",
			"dark:bg-gray-700",
		)
	}
	
	return strings.Join(baseClasses, " ")
}

// getBrandHref returns the brand link href with fallback
func getBrandHref(href string) string {
	if href == "" {
		return "/"
	}
	return href
}

// getSearchPlaceholder returns the search placeholder with fallback
func getSearchPlaceholder(placeholder string) string {
	if placeholder == "" {
		return "Search..."
	}
	return placeholder
}

// getUserInitials generates user initials from name
func getUserInitials(user SidebarUser) string {
	if user.Initials != "" {
		return user.Initials
	}
	
	if user.Name == "" {
		return "U"
	}
	
	parts := strings.Fields(user.Name)
	if len(parts) == 0 {
		return "U"
	}
	
	if len(parts) == 1 {
		return strings.ToUpper(string(parts[0][0]))
	}
	
	return strings.ToUpper(string(parts[0][0]) + string(parts[len(parts)-1][0]))
}

// getUserStatusColor returns color classes for status indicators
func getUserStatusColor(status string) string {
	switch status {
	case "online":
		return "bg-green-500"
	case "away":
		return "bg-yellow-500"
	case "busy":
		return "bg-red-500"
	case "offline":
		return "bg-gray-400"
	default:
		return "bg-gray-400"
	}
}

// getSidebarAlpineData returns the Alpine.js data for the sidebar
func getSidebarAlpineData(props SidebarProps) string {
	return `{
		collapsed: ` + strconv.FormatBool(props.Collapsed) + `,
		mobileOpen: false,
		searchTerm: '',
		
		// Toggle methods
		toggleCollapsed() {
			this.collapsed = !this.collapsed;
			
			// Dispatch event for layout adjustments
			this.$dispatch('sidebar-toggle', { collapsed: this.collapsed });
			
			// Store state in localStorage
			localStorage.setItem('sidebar-collapsed', this.collapsed);
		},
		
		toggleMobile() {
			this.mobileOpen = !this.mobileOpen;
			
			// Prevent body scroll when mobile menu is open
			if (this.mobileOpen) {
				document.body.style.overflow = 'hidden';
			} else {
				document.body.style.overflow = '';
			}
		},
		
		closeMobile() {
			this.mobileOpen = false;
			document.body.style.overflow = '';
		},
		
		// Search functionality
		filterItems() {
			// TODO: Implement search filtering
			// FIXME: Add search logic to filter sidebar items
			if (this.searchTerm.length > 0) {
				// Filter logic would go here
				console.log('Searching for:', this.searchTerm);
			}
		},
		
		// Initialization
		init() {
			// Restore collapsed state from localStorage
			const stored = localStorage.getItem('sidebar-collapsed');
			if (stored !== null) {
				this.collapsed = stored === 'true';
			}
			
			// Handle window resize
			window.addEventListener('resize', () => {
				if (window.innerWidth >= 1024) {
					this.mobileOpen = false;
					document.body.style.overflow = '';
				}
			});
			
			// Handle escape key for mobile menu
			document.addEventListener('keydown', (e) => {
				if (e.key === 'Escape' && this.mobileOpen) {
					this.closeMobile();
				}
			});
			
			// Auto-collapse on small screens
			if (window.innerWidth < 1024) {
				this.collapsed = true;
			}
		}
	}`
}