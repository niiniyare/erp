# Component Development Guide: TypeScript Hooks + Alpine.js + Templ

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Complete guide for creating new components using TypeScript hooks, Alpine.js, and Templ templates
**SCOPE**: Component creation workflow, best practices, testing, and deployment
**TARGET AUDIENCE**: Developers building ERP UI components
**RELATED FILES**: `typescript-hooks-integration.md` (patterns), `getting-started.md` (setup), `components/` (examples)
<!-- LLM-CONTEXT-END -->

## Overview

This guide provides a complete workflow for developing new UI components in the ERP system using our integrated stack: TypeScript hooks for behavior, Alpine.js for reactivity, Templ for server-side rendering, and Flowbite + TailwindCSS for styling.

<!-- LLM-WORKFLOW-OVERVIEW-START -->
**DEVELOPMENT WORKFLOW:**
1. **Plan Component** - Define requirements and interface
2. **Create TypeScript Hook** - Implement business logic and state management
3. **Build Templ Component** - Create server-rendered template
4. **Integrate Styling** - Apply Flowbite/TailwindCSS design system
5. **Test Component** - Unit and integration testing
6. **Document Usage** - Create examples and API documentation
<!-- LLM-WORKFLOW-OVERVIEW-END -->

---

## Step 1: Component Planning

### Component Requirements Analysis

<!-- LLM-REQUIREMENTS-START -->
**COMPONENT PLANNING CHECKLIST:**

#### A. Functional Requirements
- [ ] **Purpose**: What problem does this component solve?
- [ ] **Input Data**: What data does it need to operate?
- [ ] **User Interactions**: How do users interact with it?
- [ ] **Output/Events**: What events or data does it emit?
- [ ] **State Management**: What state needs to be tracked?

#### B. Non-Functional Requirements
- [ ] **Performance**: Expected data size and interaction frequency
- [ ] **Accessibility**: Screen reader support, keyboard navigation
- [ ] **Responsive Design**: Mobile, tablet, desktop behavior
- [ ] **Browser Support**: Target browser compatibility
- [ ] **Integration**: How it connects with existing components

#### C. Technical Requirements
- [ ] **Server Dependencies**: Required backend APIs or data
- [ ] **Client Dependencies**: External libraries or resources
- [ ] **Security Considerations**: Data sanitization, XSS prevention
- [ ] **Error Handling**: Failure modes and recovery strategies
<!-- LLM-REQUIREMENTS-END -->

### Example: User Avatar Component Planning

<!-- LLM-EXAMPLE-PLANNING-START -->
**EXAMPLE: User Avatar Component**

```typescript
// Component Requirements Document
interface UserAvatarRequirements {
    functional: {
        purpose: "Display user avatar with status indicator and dropdown menu";
        inputs: {
            user: User;
            size: 'sm' | 'md' | 'lg';
            showStatus?: boolean;
            showDropdown?: boolean;
        };
        interactions: [
            "Click to open user menu",
            "Hover to show user info tooltip",
            "Keyboard navigation for accessibility"
        ];
        outputs: {
            onMenuSelect: (action: string) => void;
            onStatusChange: (status: UserStatus) => void;
        };
        state: [
            "isMenuOpen: boolean",
            "isLoading: boolean", 
            "userStatus: UserStatus"
        ];
    };
    nonFunctional: {
        performance: "Handle 100+ avatars on page";
        accessibility: "WCAG 2.1 AA compliance";
        responsive: "Adaptive sizing on mobile";
        browsers: ["Chrome 90+", "Firefox 88+", "Safari 14+"];
    };
    technical: {
        serverDeps: ["/api/users/:id", "/api/users/:id/status"];
        clientDeps: ["Alpine.js", "Flowbite tooltip"];
        security: ["XSS prevention", "CSRF token for status updates"];
        errorHandling: "Graceful fallback to initials if image fails";
    };
}
```
<!-- LLM-EXAMPLE-PLANNING-END -->

---

## Step 2: Creating TypeScript Hooks

### Hook Development Template

<!-- LLM-HOOK-TEMPLATE-START -->
**STANDARD HOOK TEMPLATE:**

```typescript
/**
 * UserAvatar AlpineJS Hook (TypeScript)
 * 
 * Provides user avatar functionality with status management and dropdown menu
 * 
 * @version 1.0.0
 */

/// <reference path="./types.d.ts" />

// Configuration Interface
interface UserAvatarConfig {
    user: User;
    size?: 'sm' | 'md' | 'lg';
    showStatus?: boolean;
    showDropdown?: boolean;
    onMenuSelect?: (action: string) => void;
    onStatusChange?: (status: UserStatus) => void;
}

// State Interface
interface UserAvatarStore {
    // Data properties
    user: User;
    size: string;
    showStatus: boolean;
    showDropdown: boolean;
    
    // State properties
    isMenuOpen: boolean;
    isLoading: boolean;
    hasError: boolean;
    errorMessage: string | null;
    
    // Computed properties
    readonly avatarUrl: string;
    readonly initials: string;
    readonly statusColor: string;
    readonly sizeClasses: string;
    
    // Lifecycle methods
    init(): void;
    destroy(): void;
    
    // Interaction methods
    toggleMenu(): void;
    closeMenu(): void;
    handleMenuSelect(action: string): void;
    updateStatus(status: UserStatus): Promise<void>;
    
    // Utility methods
    getAvatarUrl(): string;
    getInitials(): string;
    handleImageError(): void;
}

// User and UserStatus type definitions
interface User {
    id: string;
    name: string;
    email: string;
    avatarUrl?: string;
    status: UserStatus;
}

interface UserStatus {
    online: boolean;
    lastSeen?: Date;
    customStatus?: string;
}

// Hook Implementation
function useUserAvatar(config: UserAvatarConfig): UserAvatarStore {
    return {
        // Initialize data from config
        user: config.user,
        size: config.size || 'md',
        showStatus: config.showStatus ?? true,
        showDropdown: config.showDropdown ?? true,
        
        // Initialize state
        isMenuOpen: false,
        isLoading: false,
        hasError: false,
        errorMessage: null,
        
        // Computed properties
        get avatarUrl(): string {
            return this.getAvatarUrl();
        },
        
        get initials(): string {
            return this.getInitials();
        },
        
        get statusColor(): string {
            if (!this.showStatus) return '';
            return this.user.status.online ? 'bg-green-400' : 'bg-gray-400';
        },
        
        get sizeClasses(): string {
            const sizeMap = {
                sm: 'w-8 h-8',
                md: 'w-10 h-10',
                lg: 'w-12 h-12'
            };
            return sizeMap[this.size as keyof typeof sizeMap] || sizeMap.md;
        },
        
        // Lifecycle methods
        init(): void {
            // Setup click outside listener for menu
            if (this.showDropdown) {
                this.setupClickOutside();
            }
            
            // Setup keyboard listeners
            this.setupKeyboardListeners();
            
            // Initialize tooltips if needed
            this.initializeTooltips();
        },
        
        destroy(): void {
            // Cleanup event listeners
            this.cleanupEventListeners();
        },
        
        // Interaction methods
        toggleMenu(): void {
            if (!this.showDropdown) return;
            
            this.isMenuOpen = !this.isMenuOpen;
            
            if (this.isMenuOpen) {
                (this as any).$dispatch('avatar-menu-opened', { user: this.user });
            } else {
                (this as any).$dispatch('avatar-menu-closed', { user: this.user });
            }
        },
        
        closeMenu(): void {
            if (this.isMenuOpen) {
                this.isMenuOpen = false;
                (this as any).$dispatch('avatar-menu-closed', { user: this.user });
            }
        },
        
        handleMenuSelect(action: string): void {
            this.closeMenu();
            
            if (config.onMenuSelect) {
                config.onMenuSelect(action);
            }
            
            (this as any).$dispatch('avatar-menu-select', { 
                action, 
                user: this.user 
            });
        },
        
        async updateStatus(status: UserStatus): Promise<void> {
            this.isLoading = true;
            this.hasError = false;
            
            try {
                // Make API call to update status
                const response = await fetch(`/api/users/${this.user.id}/status`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': this.getCSRFToken()
                    },
                    body: JSON.stringify(status)
                });
                
                if (!response.ok) {
                    throw new Error(`Failed to update status: ${response.statusText}`);
                }
                
                // Update local state
                this.user.status = status;
                
                if (config.onStatusChange) {
                    config.onStatusChange(status);
                }
                
                (this as any).$dispatch('avatar-status-updated', { 
                    user: this.user, 
                    status 
                });
                
            } catch (error) {
                this.hasError = true;
                this.errorMessage = error instanceof Error ? error.message : 'Unknown error';
                
                (this as any).$dispatch('avatar-error', { 
                    user: this.user, 
                    error: this.errorMessage 
                });
            } finally {
                this.isLoading = false;
            }
        },
        
        // Utility methods
        getAvatarUrl(): string {
            if (this.user.avatarUrl) {
                return this.user.avatarUrl;
            }
            
            // Generate avatar from initials or use default
            return this.generateInitialsAvatar();
        },
        
        getInitials(): string {
            const names = this.user.name.split(' ');
            if (names.length >= 2) {
                return (names[0][0] + names[names.length - 1][0]).toUpperCase();
            }
            return this.user.name.substring(0, 2).toUpperCase();
        },
        
        handleImageError(): void {
            // Fallback to initials when image fails to load
            this.user.avatarUrl = undefined;
        },
        
        // Private helper methods
        setupClickOutside(): void {
            const handleClickOutside = (event: Event) => {
                const target = event.target as Element;
                const avatarElement = (this as any).$el;
                
                if (avatarElement && !avatarElement.contains(target)) {
                    this.closeMenu();
                }
            };
            
            document.addEventListener('click', handleClickOutside);
            
            // Store reference for cleanup
            (this as any)._clickOutsideHandler = handleClickOutside;
        },
        
        setupKeyboardListeners(): void {
            const handleKeydown = (event: KeyboardEvent) => {
                if (event.key === 'Escape' && this.isMenuOpen) {
                    this.closeMenu();
                }
            };
            
            document.addEventListener('keydown', handleKeydown);
            (this as any)._keydownHandler = handleKeydown;
        },
        
        cleanupEventListeners(): void {
            if ((this as any)._clickOutsideHandler) {
                document.removeEventListener('click', (this as any)._clickOutsideHandler);
            }
            
            if ((this as any)._keydownHandler) {
                document.removeEventListener('keydown', (this as any)._keydownHandler);
            }
        },
        
        initializeTooltips(): void {
            // Initialize Flowbite tooltips if present
            const tooltipElement = (this as any).$el.querySelector('[data-tooltip-target]');
            if (tooltipElement && (window as any).Flowbite) {
                new (window as any).Flowbite.Tooltip(tooltipElement);
            }
        },
        
        generateInitialsAvatar(): string {
            // Create data URL for initials avatar
            const canvas = document.createElement('canvas');
            const context = canvas.getContext('2d');
            const size = 100;
            
            canvas.width = size;
            canvas.height = size;
            
            if (context) {
                // Background color based on user ID
                const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'];
                const colorIndex = this.user.id.charCodeAt(0) % colors.length;
                
                context.fillStyle = colors[colorIndex];
                context.fillRect(0, 0, size, size);
                
                // Text
                context.fillStyle = '#FFFFFF';
                context.font = 'bold 36px sans-serif';
                context.textAlign = 'center';
                context.textBaseline = 'middle';
                context.fillText(this.getInitials(), size / 2, size / 2);
            }
            
            return canvas.toDataURL();
        },
        
        getCSRFToken(): string {
            const token = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
            return token || '';
        }
    };
}

// Export to global scope for Templ components
(window as any).useUserAvatar = useUserAvatar;

// Export for TypeScript modules
export { useUserAvatar, UserAvatarConfig, UserAvatarStore, User, UserStatus };
```
<!-- LLM-HOOK-TEMPLATE-END -->

---

## Step 3: Creating Templ Components

### Component Template Structure

<!-- LLM-TEMPL-TEMPLATE-START -->
**TEMPL COMPONENT TEMPLATE:**

```go
// web/components/user/avatar.templ
package user

import (
    "encoding/json"
    "context"
)

// User represents a user in the system
type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    AvatarURL *string   `json:"avatarUrl,omitempty"`
    Status    UserStatus `json:"status"`
}

// UserStatus represents the user's online status
type UserStatus struct {
    Online       bool   `json:"online"`
    LastSeen     *int64 `json:"lastSeen,omitempty"`
    CustomStatus string `json:"customStatus,omitempty"`
}

// UserAvatarProps defines the properties for the UserAvatar component
type UserAvatarProps struct {
    User         User   `json:"user" validate:"required"`
    Size         string `json:"size" validate:"oneof=sm md lg"`
    ShowStatus   bool   `json:"showStatus"`
    ShowDropdown bool   `json:"showDropdown"`
    Class        string `json:"class"`
    ID           string `json:"id"`
}

// UserAvatar renders a user avatar with status and dropdown menu
templ UserAvatar(props UserAvatarProps) {
    <div 
        if props.ID != "" {
            id={ props.ID }
        }
        class={ "relative inline-block " + props.Class }
        x-data={ getUserAvatarData(props) }
        x-init="init()"
        @click.away="closeMenu()"
    >
        <!-- Avatar Image/Initials -->
        <div class={ "relative " + getSizeClasses(props.Size) }>
            <!-- Image with fallback to initials -->
            <div 
                class="w-full h-full rounded-full bg-gray-300 flex items-center justify-center overflow-hidden cursor-pointer"
                @click="toggleMenu()"
                :class="{ 'ring-2 ring-blue-500': isMenuOpen }"
            >
                if props.User.AvatarURL != nil && *props.User.AvatarURL != "" {
                    <img 
                        :src="avatarUrl"
                        :alt="user.name + ' avatar'"
                        class="w-full h-full object-cover"
                        @error="handleImageError()"
                        loading="lazy"
                    />
                } else {
                    <span 
                        class="text-white font-medium"
                        :class="getSizeTextClasses()"
                        x-text="initials"
                    ></span>
                }
            </div>
            
            <!-- Status Indicator -->
            if props.ShowStatus {
                <div 
                    class={ "absolute bottom-0 right-0 block rounded-full ring-2 ring-white " + getStatusSizeClasses(props.Size) }
                    :class="statusColor"
                    data-tooltip-target={ "status-tooltip-" + props.User.ID }
                    x-show="showStatus"
                ></div>
                
                <!-- Status Tooltip -->
                <div 
                    id={ "status-tooltip-" + props.User.ID }
                    role="tooltip" 
                    class="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip"
                >
                    <span x-text="user.status.online ? 'Online' : 'Offline'"></span>
                    <div class="tooltip-arrow" data-popper-arrow></div>
                </div>
            }
            
            <!-- Loading Indicator -->
            <div 
                x-show="isLoading"
                class="absolute inset-0 bg-black bg-opacity-50 rounded-full flex items-center justify-center"
                x-transition
            >
                <svg class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            </div>
        </div>
        
        <!-- Dropdown Menu -->
        if props.ShowDropdown {
            <div 
                x-show="isMenuOpen && showDropdown"
                x-transition:enter="transition ease-out duration-100"
                x-transition:enter-start="transform opacity-0 scale-95"
                x-transition:enter-end="transform opacity-100 scale-100"
                x-transition:leave="transition ease-in duration-75"
                x-transition:leave-start="transform opacity-100 scale-100"
                x-transition:leave-end="transform opacity-0 scale-95"
                class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none z-50"
                role="menu"
                aria-orientation="vertical"
                style="display: none;"
            >
                <div class="py-1" role="none">
                    <!-- User Info Header -->
                    <div class="px-4 py-2 border-b border-gray-100">
                        <p class="text-sm font-medium text-gray-900" x-text="user.name"></p>
                        <p class="text-sm text-gray-500" x-text="user.email"></p>
                    </div>
                    
                    <!-- Menu Items -->
                    <a 
                        href="#"
                        @click.prevent="handleMenuSelect('profile')"
                        class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
                        role="menuitem"
                    >
                        <i class="fas fa-user mr-3"></i>
                        View Profile
                    </a>
                    
                    <a 
                        href="#"
                        @click.prevent="handleMenuSelect('settings')"
                        class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
                        role="menuitem"
                    >
                        <i class="fas fa-cog mr-3"></i>
                        Settings
                    </a>
                    
                    <!-- Status Toggle -->
                    <div class="border-t border-gray-100">
                        <button 
                            @click="updateStatus({ online: !user.status.online })"
                            class="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
                            :disabled="isLoading"
                        >
                            <i :class="user.status.online ? 'fas fa-moon' : 'fas fa-sun'" class="mr-3"></i>
                            <span x-text="user.status.online ? 'Go Offline' : 'Go Online'"></span>
                        </button>
                    </div>
                    
                    <div class="border-t border-gray-100">
                        <a 
                            href="#"
                            @click.prevent="handleMenuSelect('logout')"
                            class="block px-4 py-2 text-sm text-red-700 hover:bg-red-50 hover:text-red-900"
                            role="menuitem"
                        >
                            <i class="fas fa-sign-out-alt mr-3"></i>
                            Sign Out
                        </a>
                    </div>
                </div>
            </div>
        }
        
        <!-- Error Toast -->
        <div 
            x-show="hasError"
            x-transition
            class="absolute top-full left-0 mt-2 p-3 bg-red-100 border border-red-400 text-red-700 rounded-md text-sm"
            style="display: none;"
        >
            <div class="flex items-center">
                <i class="fas fa-exclamation-triangle mr-2"></i>
                <span x-text="errorMessage"></span>
                <button @click="hasError = false; errorMessage = null;" class="ml-2 text-red-500 hover:text-red-700">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        </div>
    </div>
}

// Helper functions for styling
func getSizeClasses(size string) string {
    switch size {
    case "sm":
        return "w-8 h-8"
    case "lg":
        return "w-12 h-12"
    default: // md
        return "w-10 h-10"
    }
}

func getStatusSizeClasses(size string) string {
    switch size {
    case "sm":
        return "w-2 h-2"
    case "lg":
        return "w-4 h-4"
    default: // md
        return "w-3 h-3"
    }
}

func getSizeTextClasses(size string) string {
    switch size {
    case "sm":
        return "text-xs"
    case "lg":
        return "text-lg"
    default: // md
        return "text-sm"
    }
}

// Helper function to convert props to Alpine.js data
func getUserAvatarData(props UserAvatarProps) string {
    config := map[string]interface{}{
        "user":         props.User,
        "size":         props.Size,
        "showStatus":   props.ShowStatus,
        "showDropdown": props.ShowDropdown,
    }
    
    data, _ := json.Marshal(config)
    return "useUserAvatar(" + string(data) + ")"
}
```
<!-- LLM-TEMPL-TEMPLATE-END -->

---

## Step 4: Styling Integration

### Flowbite + TailwindCSS Integration

<!-- LLM-STYLING-INTEGRATION-START -->
**STYLING BEST PRACTICES:**

#### A. Using Flowbite Design System
```html
<!-- Use Flowbite component classes for consistency -->
<div class="relative inline-flex items-center justify-center w-10 h-10 overflow-hidden bg-gray-100 rounded-full dark:bg-gray-600">
    <span class="font-medium text-gray-600 dark:text-gray-300">JL</span>
</div>

<!-- Flowbite tooltip integration -->
<div data-tooltip-target="tooltip-default" data-tooltip="Tooltip content">
    <!-- Content -->
</div>
<div id="tooltip-default" role="tooltip" class="absolute z-10 invisible inline-block px-3 py-2 text-sm font-medium text-white transition-opacity duration-300 bg-gray-900 rounded-lg shadow-sm opacity-0 tooltip dark:bg-gray-700">
    Tooltip content
    <div class="tooltip-arrow" data-popper-arrow></div>
</div>
```

#### B. Custom Theme Integration
```css
/* web/styles/components.css - Custom component styles */
.erp-avatar {
    @apply relative inline-block;
}

.erp-avatar-sm {
    @apply w-8 h-8;
}

.erp-avatar-md {
    @apply w-10 h-10;
}

.erp-avatar-lg {
    @apply w-12 h-12;
}

.erp-avatar-status {
    @apply absolute bottom-0 right-0 block rounded-full ring-2 ring-white;
}

.erp-avatar-status-online {
    @apply bg-green-400;
}

.erp-avatar-status-offline {
    @apply bg-gray-400;
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
    .erp-avatar-status {
        @apply ring-gray-800;
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .erp-avatar {
        @apply ring-2 ring-gray-900;
    }
}
```

#### C. Responsive Design Patterns
```go
// Responsive avatar component
templ ResponsiveUserAvatar(props UserAvatarProps) {
    <div class="flex items-center space-x-3">
        <!-- Mobile: Small avatar -->
        <div class="sm:hidden">
            @UserAvatar(UserAvatarProps{
                User: props.User,
                Size: "sm",
                ShowStatus: true,
                ShowDropdown: false,
            })
        </div>
        
        <!-- Tablet and up: Full avatar with dropdown -->
        <div class="hidden sm:block">
            @UserAvatar(props)
        </div>
        
        <!-- Mobile: User name next to avatar -->
        <div class="sm:hidden">
            <p class="text-sm font-medium text-gray-900">{ props.User.Name }</p>
            <p class="text-xs text-gray-500">{ props.User.Email }</p>
        </div>
    </div>
}
```
<!-- LLM-STYLING-INTEGRATION-END -->

---

## Step 5: Testing Components

### Unit Testing Strategy

<!-- LLM-TESTING-STRATEGY-START -->
**COMPREHENSIVE TESTING APPROACH:**

#### A. TypeScript Hook Testing
```typescript
// test/hooks/userAvatar.test.ts
import { useUserAvatar, User, UserStatus } from '../../../web/hooks/userAvatar';

describe('useUserAvatar Hook', () => {
    let mockUser: User;
    let hookInstance: ReturnType<typeof useUserAvatar>;
    
    beforeEach(() => {
        mockUser = {
            id: 'user-123',
            name: 'John Doe',
            email: 'john@example.com',
            avatarUrl: 'https://example.com/avatar.jpg',
            status: { online: true }
        };
        
        hookInstance = useUserAvatar({ user: mockUser });
        
        // Mock Alpine.js methods
        (hookInstance as any).$dispatch = jest.fn();
        (hookInstance as any).$el = document.createElement('div');
    });
    
    afterEach(() => {
        hookInstance.destroy();
    });
    
    describe('Initialization', () => {
        test('should initialize with correct default values', () => {
            expect(hookInstance.user).toEqual(mockUser);
            expect(hookInstance.size).toBe('md');
            expect(hookInstance.showStatus).toBe(true);
            expect(hookInstance.isMenuOpen).toBe(false);
        });
        
        test('should compute initials correctly', () => {
            expect(hookInstance.initials).toBe('JD');
        });
        
        test('should compute status color correctly', () => {
            expect(hookInstance.statusColor).toBe('bg-green-400');
            
            hookInstance.user.status.online = false;
            expect(hookInstance.statusColor).toBe('bg-gray-400');
        });
    });
    
    describe('Menu Interactions', () => {
        test('should toggle menu open/close', () => {
            expect(hookInstance.isMenuOpen).toBe(false);
            
            hookInstance.toggleMenu();
            expect(hookInstance.isMenuOpen).toBe(true);
            expect((hookInstance as any).$dispatch).toHaveBeenCalledWith(
                'avatar-menu-opened', 
                { user: mockUser }
            );
            
            hookInstance.toggleMenu();
            expect(hookInstance.isMenuOpen).toBe(false);
            expect((hookInstance as any).$dispatch).toHaveBeenCalledWith(
                'avatar-menu-closed', 
                { user: mockUser }
            );
        });
        
        test('should handle menu selection', () => {
            const onMenuSelect = jest.fn();
            const hookWithCallback = useUserAvatar({ 
                user: mockUser, 
                onMenuSelect 
            });
            
            (hookWithCallback as any).$dispatch = jest.fn();
            
            hookWithCallback.handleMenuSelect('profile');
            
            expect(onMenuSelect).toHaveBeenCalledWith('profile');
            expect((hookWithCallback as any).$dispatch).toHaveBeenCalledWith(
                'avatar-menu-select',
                { action: 'profile', user: mockUser }
            );
            expect(hookWithCallback.isMenuOpen).toBe(false);
        });
    });
    
    describe('Status Updates', () => {
        beforeEach(() => {
            // Mock fetch
            global.fetch = jest.fn();
            
            // Mock CSRF token
            document.head.innerHTML = '<meta name="csrf-token" content="test-token">';
        });
        
        test('should update status successfully', async () => {
            const newStatus: UserStatus = { online: false };
            
            (global.fetch as jest.Mock).mockResolvedValueOnce({
                ok: true,
                json: () => Promise.resolve({ success: true })
            });
            
            await hookInstance.updateStatus(newStatus);
            
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/users/user-123/status',
                expect.objectContaining({
                    method: 'PUT',
                    headers: expect.objectContaining({
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': 'test-token'
                    }),
                    body: JSON.stringify(newStatus)
                })
            );
            
            expect(hookInstance.user.status).toEqual(newStatus);
            expect(hookInstance.hasError).toBe(false);
        });
        
        test('should handle status update errors', async () => {
            (global.fetch as jest.Mock).mockResolvedValueOnce({
                ok: false,
                statusText: 'Forbidden'
            });
            
            await hookInstance.updateStatus({ online: false });
            
            expect(hookInstance.hasError).toBe(true);
            expect(hookInstance.errorMessage).toContain('Forbidden');
            expect((hookInstance as any).$dispatch).toHaveBeenCalledWith(
                'avatar-error',
                expect.objectContaining({
                    user: mockUser,
                    error: expect.stringContaining('Forbidden')
                })
            );
        });
    });
    
    describe('Image Handling', () => {
        test('should handle image error gracefully', () => {
            hookInstance.handleImageError();
            expect(hookInstance.user.avatarUrl).toBeUndefined();
        });
        
        test('should generate initials avatar', () => {
            const avatarUrl = hookInstance.generateInitialsAvatar();
            expect(avatarUrl).toMatch(/^data:image\/png;base64,/);
        });
    });
    
    describe('Accessibility', () => {
        test('should setup keyboard listeners', () => {
            const addEventListenerSpy = jest.spyOn(document, 'addEventListener');
            
            hookInstance.init();
            
            expect(addEventListenerSpy).toHaveBeenCalledWith(
                'keydown',
                expect.any(Function)
            );
        });
        
        test('should close menu on Escape key', () => {
            hookInstance.isMenuOpen = true;
            
            const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
            document.dispatchEvent(escapeEvent);
            
            expect(hookInstance.isMenuOpen).toBe(false);
        });
    });
});
```

#### B. Templ Component Testing
```go
// test/components/user/avatar_test.go
package user_test

import (
    "context"
    "strings"
    "testing"
    
    "your-project/web/components/user"
)

func TestUserAvatarComponent(t *testing.T) {
    testUser := user.User{
        ID:        "test-user",
        Name:      "Test User", 
        Email:     "test@example.com",
        AvatarURL: stringPtr("https://example.com/avatar.jpg"),
        Status:    user.UserStatus{Online: true},
    }
    
    tests := []struct {
        name     string
        props    user.UserAvatarProps
        contains []string
        notContains []string
    }{
        {
            name: "renders avatar with all features",
            props: user.UserAvatarProps{
                User:         testUser,
                Size:         "md",
                ShowStatus:   true,
                ShowDropdown: true,
                ID:           "test-avatar",
            },
            contains: []string{
                `id="test-avatar"`,
                `x-data="useUserAvatar(`,
                `x-init="init()"`,
                `@click="toggleMenu()"`,
                `src="https://example.com/avatar.jpg"`,
                `alt="Test User avatar"`,
                `class="w-10 h-10"`, // md size
                `role="menu"`,
            },
        },
        {
            name: "renders without dropdown when disabled",
            props: user.UserAvatarProps{
                User:         testUser,
                Size:         "sm",
                ShowStatus:   true,
                ShowDropdown: false,
            },
            contains: []string{
                `class="w-8 h-8"`, // sm size
                `x-show="showStatus"`,
            },
            notContains: []string{
                `role="menu"`,
                `x-show="isMenuOpen && showDropdown"`,
            },
        },
        {
            name: "renders without status when disabled",
            props: user.UserAvatarProps{
                User:         testUser,
                Size:         "lg",
                ShowStatus:   false,
                ShowDropdown: true,
            },
            contains: []string{
                `class="w-12 h-12"`, // lg size
            },
            notContains: []string{
                `x-show="showStatus"`,
                `data-tooltip-target="status-tooltip-`,
            },
        },
        {
            name: "renders initials when no avatar URL",
            props: user.UserAvatarProps{
                User: user.User{
                    ID:     "no-avatar-user",
                    Name:   "No Avatar User",
                    Email:  "noavatar@example.com", 
                    Status: user.UserStatus{Online: false},
                },
                Size: "md",
            },
            contains: []string{
                `x-text="initials"`,
            },
            notContains: []string{
                `src=`,
                `@error="handleImageError()"`,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf strings.Builder
            err := user.UserAvatar(tt.props).Render(context.Background(), &buf)
            
            if err != nil {
                t.Fatalf("Failed to render UserAvatar: %v", err)
            }
            
            html := buf.String()
            
            // Check required content
            for _, expected := range tt.contains {
                if !strings.Contains(html, expected) {
                    t.Errorf("Expected HTML to contain %q, but it didn't.\nHTML: %s", expected, html)
                }
            }
            
            // Check excluded content
            for _, notExpected := range tt.notContains {
                if strings.Contains(html, notExpected) {
                    t.Errorf("Expected HTML NOT to contain %q, but it did.\nHTML: %s", notExpected, html)
                }
            }
        })
    }
}

func TestHelperFunctions(t *testing.T) {
    tests := []struct {
        name     string
        function func(string) string
        input    string
        expected string
    }{
        {"getSizeClasses sm", user.GetSizeClasses, "sm", "w-8 h-8"},
        {"getSizeClasses md", user.GetSizeClasses, "md", "w-10 h-10"},
        {"getSizeClasses lg", user.GetSizeClasses, "lg", "w-12 h-12"},
        {"getSizeClasses default", user.GetSizeClasses, "invalid", "w-10 h-10"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.function(tt.input)
            if result != tt.expected {
                t.Errorf("Expected %q, got %q", tt.expected, result)
            }
        })
    }
}

func stringPtr(s string) *string {
    return &s
}
```

#### C. Integration Testing
```typescript
// test/integration/userAvatar.integration.test.ts
describe('UserAvatar Integration Tests', () => {
    let container: HTMLElement;
    
    beforeEach(() => {
        container = document.createElement('div');
        document.body.appendChild(container);
        
        // Mock global functions
        (window as any).useUserAvatar = useUserAvatar;
        (window as any).Alpine = {
            start: jest.fn(),
            data: jest.fn(),
        };
    });
    
    afterEach(() => {
        document.body.removeChild(container);
    });
    
    test('should integrate hook with rendered component', async () => {
        // Simulate server-rendered HTML
        container.innerHTML = `
            <div x-data="useUserAvatar({
                user: {
                    id: 'test-user',
                    name: 'Test User',
                    email: 'test@example.com',
                    status: { online: true }
                },
                size: 'md',
                showStatus: true,
                showDropdown: true
            })" x-init="init()">
                <div class="relative w-10 h-10">
                    <img x-bind:src="avatarUrl" x-bind:alt="user.name + ' avatar'" />
                    <div x-bind:class="statusColor" x-show="showStatus"></div>
                </div>
                <div x-show="isMenuOpen" style="display: none;">
                    <button @click="handleMenuSelect('profile')">Profile</button>
                </div>
            </div>
        `;
        
        // Initialize Alpine.js component manually
        const element = container.firstElementChild as HTMLElement;
        const hookInstance = useUserAvatar({
            user: {
                id: 'test-user',
                name: 'Test User',
                email: 'test@example.com',
                status: { online: true }
            }
        });
        
        // Simulate Alpine.js binding
        (element as any).hookInstance = hookInstance;
        
        // Test avatar image source
        const img = element.querySelector('img');
        expect(img).toBeTruthy();
        
        // Test status indicator
        const statusIndicator = element.querySelector('[x-bind\\:class="statusColor"]');
        expect(statusIndicator).toBeTruthy();
        
        // Test menu interaction
        hookInstance.toggleMenu();
        expect(hookInstance.isMenuOpen).toBe(true);
        
        // Test menu selection
        const profileButton = element.querySelector('button');
        const mockDispatch = jest.fn();
        (hookInstance as any).$dispatch = mockDispatch;
        
        hookInstance.handleMenuSelect('profile');
        expect(mockDispatch).toHaveBeenCalledWith('avatar-menu-select', {
            action: 'profile',
            user: expect.objectContaining({ id: 'test-user' })
        });
    });
});
```
<!-- LLM-TESTING-STRATEGY-END -->

---

## Step 6: Documentation and Examples

### Component Documentation Template

<!-- LLM-DOCUMENTATION-TEMPLATE-START -->
**COMPONENT DOCUMENTATION STRUCTURE:**

```markdown
# UserAvatar Component

## Overview
The UserAvatar component displays a user's profile picture with status indicator and dropdown menu. It provides a consistent interface for user identification and quick access to user actions across the ERP system.

## Features
- ✅ Responsive avatar display with fallback to initials
- ✅ Online/offline status indicator
- ✅ Dropdown menu with user actions
- ✅ Keyboard navigation support
- ✅ Loading states and error handling
- ✅ CSRF protection for status updates
- ✅ Accessibility compliant (WCAG 2.1 AA)

## Usage

### Basic Usage
```go
@user.UserAvatar(user.UserAvatarProps{
    User: currentUser,
    Size: "md",
    ShowStatus: true,
    ShowDropdown: true,
})
```

### Configuration Options
```go
type UserAvatarProps struct {
    User         User   `json:"user" validate:"required"`           // User data
    Size         string `json:"size" validate:"oneof=sm md lg"`     // Avatar size
    ShowStatus   bool   `json:"showStatus"`                        // Show status indicator
    ShowDropdown bool   `json:"showDropdown"`                      // Show dropdown menu
    Class        string `json:"class"`                             // Additional CSS classes
    ID           string `json:"id"`                                // Element ID
}
```

## Examples

### Small Avatar (Navigation Bar)
```go
@user.UserAvatar(user.UserAvatarProps{
    User: currentUser,
    Size: "sm",
    ShowStatus: false,
    ShowDropdown: false,
    Class: "mr-2",
})
```

### Large Avatar (Profile Page)
```go
@user.UserAvatar(user.UserAvatarProps{
    User: profileUser,
    Size: "lg",
    ShowStatus: true,
    ShowDropdown: true,
    ID: "profile-avatar",
})
```

### Custom Event Handling
```go
// In your Templ component
<div 
    @avatar-menu-select="handleAvatarMenuSelect($event.detail)"
    @avatar-status-updated="handleStatusUpdate($event.detail)"
>
    @user.UserAvatar(props)
</div>

<script>
function handleAvatarMenuSelect(detail) {
    switch(detail.action) {
        case 'profile':
            window.location.href = '/profile';
            break;
        case 'settings':
            window.location.href = '/settings';
            break;
        case 'logout':
            if(confirm('Are you sure you want to sign out?')) {
                window.location.href = '/logout';
            }
            break;
    }
}

function handleStatusUpdate(detail) {
    console.log('Status updated:', detail.status);
    // Show success notification
}
</script>
```

## Accessibility

### Keyboard Navigation
- **Tab**: Focus on avatar
- **Enter/Space**: Open dropdown menu
- **Escape**: Close dropdown menu
- **Arrow Keys**: Navigate menu items

### Screen Reader Support
- Proper ARIA labels and roles
- Status announcements for screen readers
- Keyboard-accessible menu navigation

## API Reference

### TypeScript Hook
```typescript
function useUserAvatar(config: UserAvatarConfig): UserAvatarStore
```

#### Configuration
```typescript
interface UserAvatarConfig {
    user: User;                              // Required user object
    size?: 'sm' | 'md' | 'lg';              // Avatar size (default: 'md')
    showStatus?: boolean;                    // Show status indicator (default: true)
    showDropdown?: boolean;                  // Show dropdown menu (default: true)
    onMenuSelect?: (action: string) => void; // Menu selection callback
    onStatusChange?: (status: UserStatus) => void; // Status change callback
}
```

#### Store Properties
```typescript
interface UserAvatarStore {
    // Data
    user: User;
    size: string;
    showStatus: boolean;
    showDropdown: boolean;
    
    // State
    isMenuOpen: boolean;
    isLoading: boolean;
    hasError: boolean;
    errorMessage: string | null;
    
    // Computed
    readonly avatarUrl: string;
    readonly initials: string;
    readonly statusColor: string;
    readonly sizeClasses: string;
    
    // Methods
    init(): void;
    destroy(): void;
    toggleMenu(): void;
    closeMenu(): void;
    handleMenuSelect(action: string): void;
    updateStatus(status: UserStatus): Promise<void>;
}
```

### Events

#### avatar-menu-opened
Fired when the dropdown menu is opened.
```typescript
detail: { user: User }
```

#### avatar-menu-closed
Fired when the dropdown menu is closed.
```typescript
detail: { user: User }
```

#### avatar-menu-select
Fired when a menu item is selected.
```typescript
detail: { action: string; user: User }
```

#### avatar-status-updated
Fired when user status is successfully updated.
```typescript
detail: { user: User; status: UserStatus }
```

#### avatar-error
Fired when an error occurs.
```typescript
detail: { user: User; error: string }
```

## Styling

### CSS Classes
```css
.erp-avatar               /* Base avatar container */
.erp-avatar-sm           /* Small size (32x32) */
.erp-avatar-md           /* Medium size (40x40) */
.erp-avatar-lg           /* Large size (48x48) */
.erp-avatar-status       /* Status indicator */
.erp-avatar-status-online  /* Online status color */
.erp-avatar-status-offline /* Offline status color */
```

### Customization
```go
@user.UserAvatar(user.UserAvatarProps{
    User: currentUser,
    Class: "ring-2 ring-blue-500 hover:ring-blue-600", // Custom styling
})
```

## Testing

### Unit Tests
```bash
# Run TypeScript hook tests
npm test hooks/userAvatar.test.ts

# Run Templ component tests  
go test ./test/components/user/...
```

### Integration Tests
```bash
# Run full integration test suite
npm test integration/userAvatar.integration.test.ts
```

## Browser Support
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Dependencies
- Alpine.js 3.x
- Flowbite 2.x
- TailwindCSS 3.x
```
<!-- LLM-DOCUMENTATION-TEMPLATE-END -->

---

## References

<!-- LLM-REFERENCES-START -->
**RELATED DOCUMENTATION:**
- [TypeScript Hooks Integration](../patterns/typescript-hooks-integration.md) - Hook development patterns
- [Getting Started](../getting-started.md) - Setup and configuration
- [Architecture](../fundamentals/architecture.md) - System design principles
- [Components](../components/elements.md) - Basic component library
- [Testing Strategies](./testing-guide.md) - Comprehensive testing approaches

**EXTERNAL REFERENCES:**
- [Alpine.js Documentation](https://alpinejs.dev/) - Reactive JavaScript framework
- [Templ Guide](https://templ.guide/) - Go templating language
- [Flowbite Components](https://flowbite.com/docs/components/) - UI component library
- [TailwindCSS](https://tailwindcss.com/docs) - Utility-first CSS framework
- [Web Accessibility Guidelines](https://www.w3.org/WAI/WCAG21/quickref/) - WCAG 2.1 reference

**TOOLS AND UTILITIES:**
- [Jest Testing Framework](https://jestjs.io/) - JavaScript testing
- [Go Testing Package](https://golang.org/pkg/testing/) - Go unit testing
- [TypeScript Compiler](https://www.typescriptlang.org/docs/) - Type checking and compilation
<!-- LLM-REFERENCES-END -->

<!-- LLM-METADATA-START -->
**METADATA FOR AI ASSISTANTS:**
- File Type: Development Guide
- Scope: Component creation workflow, TypeScript hooks, Templ templates, testing
- Target Audience: Frontend and full-stack developers
- Complexity: Intermediate to Advanced
- Focus: Complete component development lifecycle
- Dependencies: TypeScript, Alpine.js, Templ, Flowbite, TailwindCSS, Go
- Estimated Time: 2-4 hours for new component
- Last Updated: December 2024
<!-- LLM-METADATA-END -->