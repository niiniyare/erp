package siteheader

import (
	"strings"
)

// getSiteHeaderClasses returns CSS classes for the main header
func getSiteHeaderClasses(props SiteHeaderProps) string {
	classes := []string{
		"bg-white",
		"border-gray-200",
		"dark:bg-gray-900",
		"dark:border-gray-700",
	}
	
	if props.Sticky {
		classes = append(classes, "sticky", "top-0", "z-50")
	}
	
	if props.Border {
		classes = append(classes, "border-b")
	}
	
	if props.Shadow {
		classes = append(classes, "shadow-sm")
	}
	
	if props.Transparent {
		classes = append(classes, "bg-transparent", "dark:bg-transparent")
	}
	
	if props.Class != "" {
		classes = append(classes, props.Class)
	}
	
	return strings.Join(classes, " ")
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
func getUserInitials(user HeaderUser) string {
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

// getUserStatusClasses returns CSS classes for user status indicator
func getUserStatusClasses(status string) string {
	baseClasses := "w-full h-full rounded-full"
	
	switch status {
	case "online":
		return baseClasses + " bg-green-500"
	case "away":
		return baseClasses + " bg-yellow-500"
	case "busy":
		return baseClasses + " bg-red-500"
	case "offline":
		return baseClasses + " bg-gray-400"
	default:
		return baseClasses + " bg-gray-400"
	}
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

// getUserDropdownClasses returns CSS classes for user dropdown
func getUserDropdownClasses(position string) string {
	baseClasses := "absolute z-10 mt-2 w-56 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none dark:bg-gray-700 dark:ring-gray-600"
	
	if position == "left" {
		return baseClasses + " left-0 origin-top-left"
	}
	
	return baseClasses + " right-0 origin-top-right"
}

// getUserMenuItemClasses returns CSS classes for user menu items
func getUserMenuItemClasses() string {
	return "group flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-600 transition-colors duration-150"
}

// getNavItemClasses returns CSS classes for navigation items
func getNavItemClasses(item NavItem) string {
	baseClasses := []string{
		"block",
		"py-2",
		"px-3",
		"text-gray-900",
		"rounded",
		"hover:bg-gray-100",
		"md:hover:bg-transparent",
		"md:border-0",
		"md:hover:text-blue-700",
		"md:p-0",
		"dark:text-white",
		"md:dark:hover:text-blue-500",
		"dark:hover:bg-gray-700",
		"dark:hover:text-white",
		"md:dark:hover:bg-transparent",
		"transition-colors",
		"duration-150",
	}
	
	if item.Active {
		baseClasses = append(baseClasses, 
			"text-blue-700", 
			"bg-blue-100", 
			"md:bg-transparent", 
			"md:text-blue-700", 
			"dark:text-blue-500",
		)
	}
	
	if item.Disabled {
		baseClasses = append(baseClasses, 
			"opacity-50", 
			"cursor-not-allowed", 
			"pointer-events-none",
		)
	}
	
	return strings.Join(baseClasses, " ")
}

// getDropdownItemClasses returns CSS classes for dropdown items
func getDropdownItemClasses(item NavItem) string {
	baseClasses := []string{
		"group",
		"flex",
		"items-center",
		"px-4",
		"py-2",
		"text-sm",
		"text-gray-700",
		"hover:bg-gray-100",
		"dark:text-gray-300",
		"dark:hover:bg-gray-600",
		"transition-colors",
		"duration-150",
	}
	
	if item.Active {
		baseClasses = append(baseClasses, "bg-gray-100", "dark:bg-gray-600")
	}
	
	if item.Disabled {
		baseClasses = append(baseClasses, 
			"opacity-50", 
			"cursor-not-allowed", 
			"pointer-events-none",
		)
	}
	
	return strings.Join(baseClasses, " ")
}

// getMobileNavItemClasses returns CSS classes for mobile navigation items
func getMobileNavItemClasses(item NavItem) string {
	baseClasses := []string{
		"block",
		"px-3",
		"py-2",
		"text-gray-700",
		"rounded-lg",
		"hover:bg-gray-100",
		"dark:text-gray-300",
		"dark:hover:bg-gray-600",
		"transition-colors",
		"duration-150",
	}
	
	if item.Active {
		baseClasses = append(baseClasses, 
			"text-blue-700", 
			"bg-blue-100", 
			"dark:text-blue-500", 
			"dark:bg-blue-900/20",
		)
	}
	
	if item.Disabled {
		baseClasses = append(baseClasses, 
			"opacity-50", 
			"cursor-not-allowed", 
			"pointer-events-none",
		)
	}
	
	return strings.Join(baseClasses, " ")
}

// getNavBadgeClasses returns CSS classes for navigation badges
func getNavBadgeClasses(color string) string {
	baseClasses := "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ml-2"
	
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
	default:
		return baseClasses + " bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
	}
}

// getSiteHeaderAlpineData returns the Alpine.js data for the header
func getSiteHeaderAlpineData() string {
	return `{
		mobileMenuOpen: false,
		
		// Mobile menu methods
		toggleMobileMenu() {
			this.mobileMenuOpen = !this.mobileMenuOpen;
		},
		
		closeMobileMenu() {
			this.mobileMenuOpen = false;
		},
		
		// Keyboard shortcuts
		init() {
			// Close mobile menu on escape
			this.$watch('mobileMenuOpen', (value) => {
				if (value) {
					document.body.style.overflow = 'hidden';
				} else {
					document.body.style.overflow = '';
				}
			});
			
			// Close mobile menu on window resize to desktop
			window.addEventListener('resize', () => {
				if (window.innerWidth >= 768) {
					this.mobileMenuOpen = false;
				}
			});
			
			// Handle keyboard shortcuts
			document.addEventListener('keydown', (e) => {
				// CMD/Ctrl + K for search (if search is enabled)
				if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
					e.preventDefault();
					const searchInput = document.querySelector('input[type="search"]');
					if (searchInput) {
						searchInput.focus();
					}
				}
				
				// Escape to close mobile menu
				if (e.key === 'Escape' && this.mobileMenuOpen) {
					this.mobileMenuOpen = false;
				}
			});
		}
	}`
}