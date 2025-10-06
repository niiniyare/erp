package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../static/"))))
	r.PathPrefix("/styles/").Handler(http.StripPrefix("/styles/", http.FileServer(http.Dir("../styles/"))))

	// Basic routes for testing components
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/components", componentsHandler).Methods("GET")
	r.HandleFunc("/builder", builderHandler).Methods("GET")
	r.HandleFunc("/demo/atoms", atomsHandler).Methods("GET")
	r.HandleFunc("/demo/molecules", moleculesHandler).Methods("GET")
	r.HandleFunc("/demo/organisms", organismsHandler).Methods("GET")
	r.HandleFunc("/demo/schemas", schemasHandler).Methods("GET")
	r.HandleFunc("/demo/interactive", interactiveHandler).Methods("GET")
	r.HandleFunc("/demo/forms", formsHandler).Methods("GET")

	// Debug console route
	r.HandleFunc("/debug", debugHandler).Methods("GET")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	}).Methods("GET")

	port := ":8090"
	fmt.Printf("🚀 Test server starting on http://localhost%s\n", port)
	fmt.Printf("📝 Available routes:\n")
	fmt.Printf("   http://localhost%s/ - Home page\n", port)
	fmt.Printf("   http://localhost%s/components - Component showcase\n", port)
	fmt.Printf("   http://localhost%s/builder - Visual schema builder\n", port)
	fmt.Printf("   http://localhost%s/demo/atoms - Atomic components\n", port)
	fmt.Printf("   http://localhost%s/demo/molecules - Molecular components\n", port)
	fmt.Printf("   http://localhost%s/demo/organisms - Organism components\n", port)
	fmt.Printf("   http://localhost%s/demo/schemas - Schema examples\n", port)

	log.Fatal(http.ListenAndServe(port, r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ERP UI System Test</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="/styles/themes.css" rel="stylesheet">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold text-center mb-8">ERP UI System Test Suite</h1>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div class="bg-white rounded-lg shadow p-6">
                <h2 class="text-xl font-semibold mb-4">🎨 Component Testing</h2>
                <div class="space-y-2">
                    <a href="/demo/atoms" class="block px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Atomic Components</a>
                    <a href="/demo/molecules" class="block px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600">Molecular Components</a>
                    <a href="/demo/organisms" class="block px-4 py-2 bg-purple-500 text-white rounded hover:bg-purple-600">Organism Components</a>
                </div>
            </div>

            <div class="bg-white rounded-lg shadow p-6">
                <h2 class="text-xl font-semibold mb-4">🏗️ Schema System</h2>
                <div class="space-y-2">
                    <a href="/demo/schemas" class="block px-4 py-2 bg-indigo-500 text-white rounded hover:bg-indigo-600">Schema Examples</a>
                    <a href="/builder" class="block px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">Visual Builder</a>
                </div>
            </div>

            <div class="bg-white rounded-lg shadow p-6">
                <h2 class="text-xl font-semibold mb-4">⚡ Interactive Testing</h2>
                <div class="space-y-2">
                    <button 
                        hx-get="/health" 
                        hx-target="#htmx-test"
                        class="block w-full px-4 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600">
                        Test HTMX
                    </button>
                    <div id="htmx-test" class="mt-2 p-2 bg-gray-100 rounded"></div>
                </div>
                
                <div x-data="{ message: 'Alpine.js works!' }" class="mt-4">
                    <button 
                        @click="message = 'Alpine.js is working perfectly!'"
                        class="block w-full px-4 py-2 bg-teal-500 text-white rounded hover:bg-teal-600">
                        Test Alpine.js
                    </button>
                    <div x-text="message" class="mt-2 p-2 bg-gray-100 rounded"></div>
                </div>
            </div>
        </div>

        <div class="mt-8 text-center">
            <p class="text-gray-600">🔥 All systems loaded and ready for testing!</p>
        </div>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func componentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Component Showcase</h1><p>Coming soon...</p>")
}

func builderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Visual Schema Builder</h1><p>Coming soon...</p>")
}

func atomsHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Atomic Components Test</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100 p-8">
    <div class="container mx-auto">
        <h1 class="text-3xl font-bold mb-8">Atomic Components</h1>
        
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Buttons</h2>
                <div class="space-y-2">
                    <button class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Primary</button>
                    <button class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">Secondary</button>
                    <button class="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">Danger</button>
                </div>
            </div>

            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Inputs</h2>
                <div class="space-y-2">
                    <input type="text" placeholder="Text input" class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <input type="email" placeholder="Email input" class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <textarea placeholder="Textarea" class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500" rows="3"></textarea>
                </div>
            </div>

            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Form Controls</h2>
                <div class="space-y-2">
                    <label class="flex items-center">
                        <input type="checkbox" class="mr-2">
                        <span>Checkbox</span>
                    </label>
                    <label class="flex items-center">
                        <input type="radio" name="radio" class="mr-2">
                        <span>Radio Option 1</span>
                    </label>
                    <label class="flex items-center">
                        <input type="radio" name="radio" class="mr-2">
                        <span>Radio Option 2</span>
                    </label>
                </div>
            </div>
        </div>

        <div class="mt-8">
            <a href="/" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">← Back to Home</a>
        </div>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func moleculesHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Molecular Components Test</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100 p-8">
    <div class="container mx-auto">
        <h1 class="text-3xl font-bold mb-8">Molecular Components</h1>
        
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Cards</h2>
                <div class="border rounded-lg p-4">
                    <h3 class="font-semibold">Sample Card</h3>
                    <p class="text-gray-600 mt-2">This is a sample card component with some content.</p>
                    <button class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Action</button>
                </div>
            </div>

            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Alerts</h2>
                <div class="space-y-2">
                    <div class="bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded">
                        Info: This is an info message
                    </div>
                    <div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
                        Success: Operation completed successfully
                    </div>
                    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                        Error: Something went wrong
                    </div>
                </div>
            </div>

            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Search</h2>
                <div class="relative">
                    <input type="text" placeholder="Search..." class="w-full pl-10 pr-4 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <div class="absolute left-3 top-2.5">
                        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                        </svg>
                    </div>
                </div>
            </div>

            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">Dropdown</h2>
                <div x-data="{ open: false }" class="relative">
                    <button @click="open = !open" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">
                        Dropdown ▼
                    </button>
                    <div x-show="open" @click.away="open = false" x-transition class="absolute top-full left-0 mt-2 w-48 bg-white border rounded shadow-lg">
                        <a href="#" class="block px-4 py-2 hover:bg-gray-100">Option 1</a>
                        <a href="#" class="block px-4 py-2 hover:bg-gray-100">Option 2</a>
                        <a href="#" class="block px-4 py-2 hover:bg-gray-100">Option 3</a>
                    </div>
                </div>
            </div>
        </div>

        <div class="mt-8">
            <a href="/" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">← Back to Home</a>
        </div>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func organismsHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Organism Components Test</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-blue-600 text-white p-4">
            <div class="container mx-auto flex justify-between items-center">
                <h1 class="text-xl font-bold">ERP System</h1>
                <div class="space-x-4">
                    <a href="#" class="hover:text-blue-200">Dashboard</a>
                    <a href="#" class="hover:text-blue-200">Users</a>
                    <a href="#" class="hover:text-blue-200">Settings</a>
                </div>
            </div>
        </nav>

        <div class="container mx-auto p-8">
            <h1 class="text-3xl font-bold mb-8">Organism Components</h1>
            
            <div class="bg-white rounded-lg shadow overflow-hidden">
                <div class="px-6 py-4 border-b">
                    <h2 class="text-xl font-semibold">Data Table</h2>
                </div>
                <div class="overflow-x-auto">
                    <table class="w-full">
                        <thead class="bg-gray-50">
                            <tr>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody class="bg-white divide-y divide-gray-200">
                            <tr>
                                <td class="px-6 py-4 whitespace-nowrap">John Doe</td>
                                <td class="px-6 py-4 whitespace-nowrap">john@example.com</td>
                                <td class="px-6 py-4 whitespace-nowrap">Admin</td>
                                <td class="px-6 py-4 whitespace-nowrap">
                                    <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">Active</span>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap space-x-2">
                                    <button class="px-3 py-1 bg-blue-500 text-white text-sm rounded hover:bg-blue-600">Edit</button>
                                    <button class="px-3 py-1 bg-red-500 text-white text-sm rounded hover:bg-red-600">Delete</button>
                                </td>
                            </tr>
                            <tr>
                                <td class="px-6 py-4 whitespace-nowrap">Jane Smith</td>
                                <td class="px-6 py-4 whitespace-nowrap">jane@example.com</td>
                                <td class="px-6 py-4 whitespace-nowrap">User</td>
                                <td class="px-6 py-4 whitespace-nowrap">
                                    <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">Pending</span>
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap space-x-2">
                                    <button class="px-3 py-1 bg-blue-500 text-white text-sm rounded hover:bg-blue-600">Edit</button>
                                    <button class="px-3 py-1 bg-red-500 text-white text-sm rounded hover:bg-red-600">Delete</button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>

            <div class="mt-8 bg-white rounded-lg shadow">
                <div class="px-6 py-4 border-b">
                    <h2 class="text-xl font-semibold">Sidebar Navigation</h2>
                </div>
                <div class="flex">
                    <div class="w-64 bg-gray-50 p-4">
                        <nav class="space-y-2">
                            <a href="#" class="flex items-center px-4 py-2 text-gray-700 bg-blue-100 rounded">
                                <span>Dashboard</span>
                            </a>
                            <a href="#" class="flex items-center px-4 py-2 text-gray-700 hover:bg-gray-100 rounded">
                                <span>Users</span>
                            </a>
                            <a href="#" class="flex items-center px-4 py-2 text-gray-700 hover:bg-gray-100 rounded">
                                <span>Reports</span>
                            </a>
                            <a href="#" class="flex items-center px-4 py-2 text-gray-700 hover:bg-gray-100 rounded">
                                <span>Settings</span>
                            </a>
                        </nav>
                    </div>
                    <div class="flex-1 p-6">
                        <p class="text-gray-600">Main content area with sidebar navigation</p>
                    </div>
                </div>
            </div>

            <div class="mt-8">
                <a href="/" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">← Back to Home</a>
            </div>
        </div>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func schemasHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Schema Examples</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100 p-8">
    <div class="container mx-auto">
        <h1 class="text-3xl font-bold mb-8">Schema System Examples</h1>
        
        <div class="bg-white rounded-lg shadow p-6">
            <h2 class="text-xl font-semibold mb-4">JSON-Driven UI Rendering</h2>
            <p class="text-gray-600 mb-4">This demonstrates how our schema system can render UI from JSON definitions.</p>
            
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div>
                    <h3 class="font-semibold mb-2">Schema Definition</h3>
                    <pre class="bg-gray-100 p-4 rounded text-sm overflow-x-auto"><code>{
  "type": "form",
  "title": "User Registration",
  "fields": [
    {
      "name": "firstName",
      "type": "text",
      "label": "First Name",
      "required": true
    },
    {
      "name": "lastName",
      "type": "text",
      "label": "Last Name",
      "required": true
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "required": true
    },
    {
      "name": "role",
      "type": "select",
      "label": "Role",
      "options": [
        {"value": "admin", "label": "Administrator"},
        {"value": "user", "label": "User"}
      ]
    }
  ],
  "actions": [
    {
      "type": "submit",
      "label": "Create User",
      "style": "primary"
    }
  ]
}</code></pre>
                </div>
                
                <div>
                    <h3 class="font-semibold mb-2">Rendered Output</h3>
                    <div class="border rounded-lg p-4">
                        <h4 class="text-lg font-semibold mb-4">User Registration</h4>
                        <form class="space-y-4">
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">First Name *</label>
                                <input type="text" class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">Last Name *</label>
                                <input type="text" class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">Email Address *</label>
                                <input type="email" class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                            </div>
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
                                <select class="w-full px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500">
                                    <option value="admin">Administrator</option>
                                    <option value="user">User</option>
                                </select>
                            </div>
                            <button type="submit" class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Create User</button>
                        </form>
                    </div>
                </div>
            </div>
        </div>

        <div class="mt-8 bg-white rounded-lg shadow p-6">
            <h2 class="text-xl font-semibold mb-4">Dashboard Schema Example</h2>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div class="bg-blue-50 p-4 rounded-lg">
                    <h3 class="font-semibold text-blue-800">Total Users</h3>
                    <p class="text-2xl font-bold text-blue-900">1,234</p>
                    <p class="text-sm text-blue-600">+12% from last month</p>
                </div>
                <div class="bg-green-50 p-4 rounded-lg">
                    <h3 class="font-semibold text-green-800">Revenue</h3>
                    <p class="text-2xl font-bold text-green-900">$45,678</p>
                    <p class="text-sm text-green-600">+8% from last month</p>
                </div>
                <div class="bg-purple-50 p-4 rounded-lg">
                    <h3 class="font-semibold text-purple-800">Orders</h3>
                    <p class="text-2xl font-bold text-purple-900">567</p>
                    <p class="text-sm text-purple-600">+15% from last month</p>
                </div>
            </div>
        </div>

        <div class="mt-8">
            <a href="/" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">← Back to Home</a>
        </div>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
func interactiveHandler(w http.ResponseWriter, r *http.Request) {
	html := `<\!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Interactive Demo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100 p-8">
    <div class="container mx-auto">
        <h1 class="text-3xl font-bold mb-8">⚡ Interactive Demo</h1>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div class="bg-white p-6 rounded-lg shadow">
                <h2 class="text-xl font-semibold mb-4">HTMX Testing</h2>
                <button hx-get="/health" hx-target="#htmx-response" class="w-full px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Test HTMX</button>
                <div id="htmx-response" class="p-4 bg-gray-100 rounded mt-4">Response will appear here</div>
            </div>
            <div class="bg-white p-6 rounded-lg shadow" x-data="{ count: 0 }">
                <h2 class="text-xl font-semibold mb-4">Alpine.js Testing</h2>
                <p class="text-lg">Count: <span x-text="count"></span></p>
                <button @click="count++" class="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600">+</button>
                <button @click="count--" class="ml-2 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">-</button>
            </div>
        </div>
        <div class="mt-8 text-center">
            <a href="/" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">← Back to Home</a>
        </div>
    </div>
    <script>console.log('🚀 Interactive demo loaded');</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func formsHandler(w http.ResponseWriter, r *http.Request) {
	html := `<\!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Forms Demo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js" defer></script>
</head>
<body class="bg-gray-100 p-8">
    <div class="container mx-auto">
        <h1 class="text-3xl font-bold mb-8">📝 Forms Demo</h1>
        <div class="bg-white p-6 rounded-lg shadow" x-data="{ formData: {} }">
            <form @submit.prevent="console.log('Form submitted:', formData)" class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
                    <input x-model="formData.name" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
                    <input x-model="formData.email" type="email" class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
                </div>
                <button type="submit" class="w-full px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">Submit</button>
            </form>
            <div x-show="Object.keys(formData).length > 0" class="mt-4 p-3 bg-gray-100 rounded">
                <pre x-text="JSON.stringify(formData, null, 2)"></pre>
            </div>
        </div>
        <div class="mt-8 text-center">
            <a href="/" class="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600">← Back to Home</a>
        </div>
    </div>
    <script>console.log('📝 Forms demo loaded');</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "debug.html")
}
EOF < /dev/null