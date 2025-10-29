package ui

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// UIHandler handles web UI routes for the ERP system
type UIHandler struct {
	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.Service
}

// NewUIHandler creates a new UI handler
func NewUIHandler(
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.Service,
) *UIHandler {
	return &UIHandler{
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}
}

// ServeDemo renders a demo page showcasing the UI components
func (h *UIHandler) ServeDemo(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Awo ERP - Component Demo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body class="bg-gray-50 dark:bg-gray-900">
    <div class="min-h-screen flex">
        <!-- Sidebar -->
        <div class="hidden md:flex md:flex-shrink-0">
            <div class="flex flex-col w-64">
                <div class="flex flex-col h-0 flex-1 bg-gray-800">
                    <div class="flex items-center h-16 flex-shrink-0 px-4 bg-gray-900">
                        <span class="text-white text-xl font-semibold">Awo ERP</span>
                    </div>
                    <div class="flex-1 flex flex-col overflow-y-auto">
                        <nav class="flex-1 px-2 py-4 bg-gray-800 space-y-1">
                            <a href="/ui/demo" class="bg-gray-900 text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Dashboard
                            </a>
                            <a href="/ui/demo/components" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Components
                            </a>
                            <a href="/ui/demo/forms" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Forms
                            </a>
                        </nav>
                    </div>
                </div>
            </div>
        </div>

        <!-- Main content -->
        <div class="flex flex-col w-0 flex-1 overflow-hidden">
            <main class="flex-1 relative overflow-y-auto focus:outline-none">
                <div class="py-6">
                    <div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
                        <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">Welcome to Awo ERP</h1>
                        <p class="mt-2 text-gray-600 dark:text-gray-400">Enterprise Resource Planning System</p>
                    </div>
                    <div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 mt-8">
                        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            <!-- Dashboard Card -->
                            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                                <div class="p-6">
                                    <div class="flex items-center">
                                        <div class="flex-shrink-0">
                                            <svg class="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                            </svg>
                                        </div>
                                        <div class="ml-5 w-0 flex-1">
                                            <dl>
                                                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Dashboard</dt>
                                                <dd class="text-lg font-medium text-gray-900 dark:text-gray-100">Overview</dd>
                                            </dl>
                                        </div>
                                    </div>
                                    <div class="mt-4">
                                        <p class="text-sm text-gray-600 dark:text-gray-400">View key performance indicators and business insights.</p>
                                        <div class="mt-4">
                                            <a href="/ui/demo" class="text-blue-600 hover:text-blue-800 text-sm font-medium">View Dashboard →</a>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Components Card -->
                            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                                <div class="p-6">
                                    <div class="flex items-center">
                                        <div class="flex-shrink-0">
                                            <svg class="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                                            </svg>
                                        </div>
                                        <div class="ml-5 w-0 flex-1">
                                            <dl>
                                                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Components</dt>
                                                <dd class="text-lg font-medium text-gray-900 dark:text-gray-100">UI Library</dd>
                                            </dl>
                                        </div>
                                    </div>
                                    <div class="mt-4">
                                        <p class="text-sm text-gray-600 dark:text-gray-400">Explore our comprehensive UI component system.</p>
                                        <div class="mt-4">
                                            <a href="/ui/demo/components" class="text-blue-600 hover:text-blue-800 text-sm font-medium">View Components →</a>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <!-- Forms Card -->
                            <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
                                <div class="p-6">
                                    <div class="flex items-center">
                                        <div class="flex-shrink-0">
                                            <svg class="h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                            </svg>
                                        </div>
                                        <div class="ml-5 w-0 flex-1">
                                            <dl>
                                                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Forms</dt>
                                                <dd class="text-lg font-medium text-gray-900 dark:text-gray-100">Interactive</dd>
                                            </dl>
                                        </div>
                                    </div>
                                    <div class="mt-4">
                                        <p class="text-sm text-gray-600 dark:text-gray-400">Test form inputs, validation, and submission flows.</p>
                                        <div class="mt-4">
                                            <a href="/ui/demo/forms" class="text-blue-600 hover:text-blue-800 text-sm font-medium">View Forms →</a>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    </div>
    <script src="/static/js/main.js"></script>
</body>
</html>`

	return c.Type("html").Send([]byte(html))
}

// ServeLogin renders the login page
func (h *UIHandler) ServeLogin(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - Awo ERP</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body class="bg-gray-50 dark:bg-gray-900">
    <div class="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div class="max-w-md w-full space-y-8">
            <div>
                <h2 class="mt-6 text-center text-3xl font-bold text-gray-900 dark:text-gray-100">
                    Sign in to Awo ERP
                </h2>
                <p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
                    Enter your credentials to access your account
                </p>
            </div>
            <form class="mt-8 space-y-6" action="#" method="POST">
                <div class="rounded-md shadow-sm -space-y-px">
                    <div>
                        <label for="email-address" class="sr-only">Email address</label>
                        <input id="email-address" name="email" type="email" autocomplete="email" required 
                               class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100" 
                               placeholder="Email address">
                    </div>
                    <div>
                        <label for="password" class="sr-only">Password</label>
                        <input id="password" name="password" type="password" autocomplete="current-password" required 
                               class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100" 
                               placeholder="Password">
                    </div>
                </div>

                <div class="flex items-center justify-between">
                    <div class="flex items-center">
                        <input id="remember-me" name="remember-me" type="checkbox" 
                               class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                        <label for="remember-me" class="ml-2 block text-sm text-gray-900 dark:text-gray-100">
                            Remember me
                        </label>
                    </div>

                    <div class="text-sm">
                        <a href="#" class="font-medium text-blue-600 hover:text-blue-500">
                            Forgot your password?
                        </a>
                    </div>
                </div>

                <div>
                    <button type="submit" 
                            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        Sign in
                    </button>
                </div>
            </form>
        </div>
    </div>
    <script src="/static/js/main.js"></script>
</body>
</html>`

	return c.Type("html").Send([]byte(html))
}

// ServeComponents renders a comprehensive components showcase page
func (h *UIHandler) ServeComponents(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UI Components - Awo ERP</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/static/css/main.css">
    <style>
        .component-section {
            border: 1px solid #e5e7eb;
            border-radius: 12px;
            padding: 24px;
            margin-bottom: 24px;
            background: white;
            box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
            transition: all 0.2s ease-in-out;
        }
        
        .component-section:hover {
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        }
        
        html[data-theme="dark"] .component-section,
        .dark .component-section {
            border-color: #374151;
            background: #1f2937;
            box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.3);
        }
        
        .component-demo {
            padding: 20px;
            border: 2px dashed #d1d5db;
            border-radius: 8px;
            margin: 16px 0;
            background: #f9fafb;
            position: relative;
        }
        
        html[data-theme="dark"] .component-demo,
        .dark .component-demo {
            border-color: #4b5563;
            background: #111827;
        }

        /* Theme switching styles */
        html {
            transition: background-color 0.3s ease, color 0.3s ease;
        }
        
        html[data-theme="dark"] {
            background-color: #111827;
            color: #f9fafb;
        }
        
        html[data-theme="dark"] body {
            background-color: #111827;
            color: #f9fafb;
        }
    </style>
</head>
<body class="bg-gray-50 dark:bg-gray-900">
    <div class="min-h-screen flex">
        <!-- Sidebar -->
        <div class="hidden md:flex md:flex-shrink-0">
            <div class="flex flex-col w-64">
                <div class="flex flex-col h-0 flex-1 bg-gray-800">
                    <div class="flex items-center h-16 flex-shrink-0 px-4 bg-gray-900">
                        <span class="text-white text-xl font-semibold">Awo ERP</span>
                    </div>
                    <div class="flex-1 flex flex-col overflow-y-auto">
                        <nav class="flex-1 px-2 py-4 bg-gray-800 space-y-1">
                            <a href="/ui/demo" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Dashboard
                            </a>
                            <a href="/ui/demo/components" class="bg-gray-900 text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Components
                            </a>
                            <a href="/ui/login" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Login
                            </a>
                        </nav>
                    </div>
                </div>
            </div>
        </div>

        <!-- Main content -->
        <div class="flex flex-col w-0 flex-1 overflow-hidden">
            <main class="flex-1 relative overflow-y-auto focus:outline-none">
                <div class="py-6">
                    <div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
                        <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">UI Components Showcase</h1>
                        <p class="mt-2 text-gray-600 dark:text-gray-400">Comprehensive collection of UI components for Awo ERP</p>
                    </div>
                    
                    <div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 mt-8">
                        
                        <!-- Buttons Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Buttons</h2>
                            <div class="component-demo">
                                <div class="flex flex-wrap gap-4">
                                    <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-blue-600 text-white shadow-sm hover:bg-blue-700 h-9 px-4 py-2">Default</button>
                                    <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-red-600 text-white shadow-sm hover:bg-red-700 h-9 px-4 py-2">Destructive</button>
                                    <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 border bg-white shadow-sm hover:bg-gray-50 h-9 px-4 py-2">Outline</button>
                                    <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-gray-100 text-gray-900 shadow-sm hover:bg-gray-200 h-9 px-4 py-2">Secondary</button>
                                    <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 hover:bg-gray-100 h-9 px-4 py-2">Ghost</button>
                                    <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 text-blue-600 underline-offset-4 hover:underline h-9 px-4 py-2">Link</button>
                                </div>
                                <div class="mt-4">
                                    <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Sizes:</h4>
                                    <div class="flex items-center gap-4">
                                        <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-blue-600 text-white shadow-sm hover:bg-blue-700 h-8 px-3">Small</button>
                                        <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-blue-600 text-white shadow-sm hover:bg-blue-700 h-9 px-4 py-2">Default</button>
                                        <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-blue-600 text-white shadow-sm hover:bg-blue-700 h-10 px-6">Large</button>
                                        <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-blue-600 text-white shadow-sm hover:bg-blue-700 size-9">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                                            </svg>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Form Inputs Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Form Inputs</h2>
                            <div class="component-demo">
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Text Input</label>
                                        <input type="text" placeholder="Enter text..." class="flex h-9 w-full rounded-md border border-gray-300 bg-white px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100">
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Email Input</label>
                                        <input type="email" placeholder="Enter email..." class="flex h-9 w-full rounded-md border border-gray-300 bg-white px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100">
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Password Input</label>
                                        <input type="password" placeholder="Enter password..." class="flex h-9 w-full rounded-md border border-gray-300 bg-white px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100">
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Select</label>
                                        <select class="flex h-9 w-full rounded-md border border-gray-300 bg-white px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100">
                                            <option>Option 1</option>
                                            <option>Option 2</option>
                                            <option>Option 3</option>
                                        </select>
                                    </div>
                                </div>
                                <div class="mt-6">
                                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Textarea</label>
                                    <textarea placeholder="Enter description..." rows="4" class="flex min-h-[60px] w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100"></textarea>
                                </div>
                            </div>
                        </div>

                        <!-- Checkboxes and Radio Buttons -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Checkboxes & Radio Buttons</h2>
                            <div class="component-demo">
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div>
                                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Checkboxes</h4>
                                        <div class="space-y-2">
                                            <div class="flex items-center space-x-2">
                                                <input type="checkbox" id="check1" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                                <label for="check1" class="text-sm text-gray-700 dark:text-gray-300">Option 1</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="checkbox" id="check2" checked class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                                <label for="check2" class="text-sm text-gray-700 dark:text-gray-300">Option 2 (checked)</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="checkbox" id="check3" disabled class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                                <label for="check3" class="text-sm text-gray-700 dark:text-gray-300 opacity-50">Option 3 (disabled)</label>
                                            </div>
                                        </div>
                                    </div>
                                    <div>
                                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Radio Buttons</h4>
                                        <div class="space-y-2">
                                            <div class="flex items-center space-x-2">
                                                <input type="radio" id="radio1" name="radio-group" class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300">
                                                <label for="radio1" class="text-sm text-gray-700 dark:text-gray-300">Option A</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="radio" id="radio2" name="radio-group" checked class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300">
                                                <label for="radio2" class="text-sm text-gray-700 dark:text-gray-300">Option B (selected)</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="radio" id="radio3" name="radio-group" disabled class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300">
                                                <label for="radio3" class="text-sm text-gray-700 dark:text-gray-300 opacity-50">Option C (disabled)</label>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Cards Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Cards</h2>
                            <div class="component-demo">
                                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                    <div class="rounded-lg border bg-white text-gray-950 shadow-sm dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50">
                                        <div class="flex flex-col space-y-1.5 p-6">
                                            <h3 class="text-2xl font-semibold leading-none tracking-tight">Card Title</h3>
                                            <p class="text-sm text-gray-500 dark:text-gray-400">Card description goes here</p>
                                        </div>
                                        <div class="p-6 pt-0">
                                            <p class="text-sm">This is the card content area. You can put any content here.</p>
                                        </div>
                                        <div class="flex items-center p-6 pt-0">
                                            <button class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 bg-blue-600 text-white shadow-sm hover:bg-blue-700 h-9 px-4 py-2">Action</button>
                                        </div>
                                    </div>

                                    <div class="rounded-lg border bg-white text-gray-950 shadow-sm dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50">
                                        <div class="flex flex-col space-y-1.5 p-6">
                                            <h3 class="text-2xl font-semibold leading-none tracking-tight">Statistics</h3>
                                            <p class="text-sm text-gray-500 dark:text-gray-400">Key performance indicators</p>
                                        </div>
                                        <div class="p-6 pt-0">
                                            <div class="text-3xl font-bold text-green-600">$24,500</div>
                                            <p class="text-xs text-gray-500 dark:text-gray-400">+12% from last month</p>
                                        </div>
                                    </div>

                                    <div class="rounded-lg border bg-white text-gray-950 shadow-sm dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50">
                                        <div class="aspect-video overflow-hidden rounded-t-lg bg-gray-100 dark:bg-gray-800"></div>
                                        <div class="flex flex-col space-y-1.5 p-6">
                                            <h3 class="text-2xl font-semibold leading-none tracking-tight">Image Card</h3>
                                            <p class="text-sm text-gray-500 dark:text-gray-400">Card with image placeholder</p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Alerts Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Alerts</h2>
                            <div class="component-demo">
                                <div class="space-y-4">
                                    <div class="relative w-full rounded-lg border border-gray-200 px-4 py-3 text-sm dark:border-gray-800 bg-blue-50 text-blue-900 border-blue-200 dark:bg-blue-950 dark:text-blue-100 dark:border-blue-800">
                                        <div class="flex items-center gap-2">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                            </svg>
                                            <strong>Info:</strong> This is an informational alert message.
                                        </div>
                                    </div>

                                    <div class="relative w-full rounded-lg border border-gray-200 px-4 py-3 text-sm dark:border-gray-800 bg-green-50 text-green-900 border-green-200 dark:bg-green-950 dark:text-green-100 dark:border-green-800">
                                        <div class="flex items-center gap-2">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                                            </svg>
                                            <strong>Success:</strong> Operation completed successfully!
                                        </div>
                                    </div>

                                    <div class="relative w-full rounded-lg border border-gray-200 px-4 py-3 text-sm dark:border-gray-800 bg-yellow-50 text-yellow-900 border-yellow-200 dark:bg-yellow-950 dark:text-yellow-100 dark:border-yellow-800">
                                        <div class="flex items-center gap-2">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.664-.833-2.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
                                            </svg>
                                            <strong>Warning:</strong> Please review this action carefully.
                                        </div>
                                    </div>

                                    <div class="relative w-full rounded-lg border border-gray-200 px-4 py-3 text-sm dark:border-gray-800 bg-red-50 text-red-900 border-red-200 dark:bg-red-950 dark:text-red-100 dark:border-red-800">
                                        <div class="flex items-center gap-2">
                                            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                                            </svg>
                                            <strong>Error:</strong> Something went wrong. Please try again.
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Badges Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Badges</h2>
                            <div class="component-demo">
                                <div class="flex flex-wrap gap-4">
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">Default</span>
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">Primary</span>
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">Success</span>
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">Warning</span>
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">Danger</span>
                                </div>
                                <div class="mt-4">
                                    <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Badge Sizes:</h4>
                                    <div class="flex items-center gap-4">
                                        <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-800">Small</span>
                                        <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-blue-100 text-blue-800">Default</span>
                                        <span class="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium bg-blue-100 text-blue-800">Large</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Progress Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Progress Bars</h2>
                            <div class="component-demo">
                                <div class="space-y-6">
                                    <div>
                                        <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400 mb-1">
                                            <span>Progress</span>
                                            <span>25%</span>
                                        </div>
                                        <div class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                                            <div class="bg-blue-600 h-2.5 rounded-full" style="width: 25%"></div>
                                        </div>
                                    </div>

                                    <div>
                                        <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400 mb-1">
                                            <span>Loading</span>
                                            <span>60%</span>
                                        </div>
                                        <div class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                                            <div class="bg-green-600 h-2.5 rounded-full" style="width: 60%"></div>
                                        </div>
                                    </div>

                                    <div>
                                        <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400 mb-1">
                                            <span>Complete</span>
                                            <span>100%</span>
                                        </div>
                                        <div class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                                            <div class="bg-green-600 h-2.5 rounded-full" style="width: 100%"></div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Tables Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Tables</h2>
                            <div class="component-demo">
                                <div class="overflow-x-auto">
                                    <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                                        <thead class="bg-gray-50 dark:bg-gray-800">
                                            <tr>
                                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Email</th>
                                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Role</th>
                                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                                            <tr>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">John Doe</td>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">john@example.com</td>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">Admin</td>
                                                <td class="px-6 py-4 whitespace-nowrap">
                                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-green-100 text-green-800">Active</span>
                                                </td>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                    <button class="text-blue-600 hover:text-blue-900 mr-2">Edit</button>
                                                    <button class="text-red-600 hover:text-red-900">Delete</button>
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">Jane Smith</td>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">jane@example.com</td>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">User</td>
                                                <td class="px-6 py-4 whitespace-nowrap">
                                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-yellow-100 text-yellow-800">Pending</span>
                                                </td>
                                                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                    <button class="text-blue-600 hover:text-blue-900 mr-2">Edit</button>
                                                    <button class="text-red-600 hover:text-red-900">Delete</button>
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>

                        <!-- Navigation Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Navigation</h2>
                            <div class="component-demo">
                                <div class="space-y-6">
                                    <div>
                                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Breadcrumb</h4>
                                        <nav class="flex" aria-label="Breadcrumb">
                                            <ol class="inline-flex items-center space-x-1 md:space-x-3">
                                                <li class="inline-flex items-center">
                                                    <a href="#" class="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-400 dark:hover:text-white">
                                                        Home
                                                    </a>
                                                </li>
                                                <li>
                                                    <div class="flex items-center">
                                                        <svg class="w-6 h-6 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                                            <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"></path>
                                                        </svg>
                                                        <a href="#" class="ml-1 text-sm font-medium text-gray-700 hover:text-blue-600 md:ml-2 dark:text-gray-400 dark:hover:text-white">Dashboard</a>
                                                    </div>
                                                </li>
                                                <li aria-current="page">
                                                    <div class="flex items-center">
                                                        <svg class="w-6 h-6 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                                                            <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"></path>
                                                        </svg>
                                                        <span class="ml-1 text-sm font-medium text-gray-500 md:ml-2 dark:text-gray-400">Components</span>
                                                    </div>
                                                </li>
                                            </ol>
                                        </nav>
                                    </div>

                                    <div>
                                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Tabs</h4>
                                        <div class="border-b border-gray-200 dark:border-gray-700">
                                            <nav class="-mb-px flex space-x-8" aria-label="Tabs">
                                                <button class="py-2 px-1 border-b-2 border-blue-500 font-medium text-sm text-blue-600 whitespace-nowrap">Current</button>
                                                <button class="py-2 px-1 border-b-2 border-transparent font-medium text-sm text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap dark:text-gray-400 dark:hover:text-gray-300">Tab 1</button>
                                                <button class="py-2 px-1 border-b-2 border-transparent font-medium text-sm text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap dark:text-gray-400 dark:hover:text-gray-300">Tab 2</button>
                                                <button class="py-2 px-1 border-b-2 border-transparent font-medium text-sm text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap dark:text-gray-400 dark:hover:text-gray-300">Tab 3</button>
                                            </nav>
                                        </div>
                                    </div>

                                    <div>
                                        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Pagination</h4>
                                        <nav class="flex items-center justify-between">
                                            <div class="flex-1 flex justify-between sm:hidden">
                                                <button class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300">Previous</button>
                                                <button class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300">Next</button>
                                            </div>
                                            <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                                                <div>
                                                    <p class="text-sm text-gray-700 dark:text-gray-300">
                                                        Showing <span class="font-medium">1</span> to <span class="font-medium">10</span> of <span class="font-medium">97</span> results
                                                    </p>
                                                </div>
                                                <div>
                                                    <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
                                                        <button class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300">Previous</button>
                                                        <button class="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300">1</button>
                                                        <button class="bg-blue-50 border-blue-500 text-blue-600 relative inline-flex items-center px-4 py-2 border text-sm font-medium">2</button>
                                                        <button class="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300">3</button>
                                                        <button class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300">Next</button>
                                                    </nav>
                                                </div>
                                            </div>
                                        </nav>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Loading States Section -->
                        <div class="component-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Loading States</h2>
                            <div class="component-demo">
                                <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                                    <div class="text-center">
                                        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
                                        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">Spinner</p>
                                    </div>
                                    
                                    <div class="text-center">
                                        <div class="flex space-x-1 justify-center">
                                            <div class="h-2 w-2 bg-blue-600 rounded-full animate-bounce"></div>
                                            <div class="h-2 w-2 bg-blue-600 rounded-full animate-bounce" style="animation-delay: 0.1s"></div>
                                            <div class="h-2 w-2 bg-blue-600 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
                                        </div>
                                        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">Dots</p>
                                    </div>
                                    
                                    <div class="text-center">
                                        <div class="h-4 bg-gray-200 rounded-full dark:bg-gray-700 overflow-hidden">
                                            <div class="h-full bg-blue-600 rounded-full animate-pulse" style="width: 70%"></div>
                                        </div>
                                        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">Progress</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                    </div>
                </div>
            </main>
        </div>
    </div>
    <script src="/static/js/main.js"></script>
</body>
</html>`

	return c.Type("html").Send([]byte(html))
}
