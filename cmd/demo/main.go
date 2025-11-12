package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/niiniyare/erp/views/pages"
)

func main() {
	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files
	e.Static("/static", "views/static")

	// Routes
	e.GET("/", handleHome)
	e.GET("/demo", handleDemo)
	e.GET("/simple", handleSimpleDemo)
	e.GET("/api/demo/htmx-content", handleHTMXDemo)
	e.POST("/api/contact", handleContact)
	e.GET("/api/search", handleSearch)

	// Start server
	log.Println("🚀 Starting demo server on :8080")
	log.Println("📱 Open http://localhost:8080/demo to view the component demo")
	log.Println("📱 Open http://localhost:8080/simple to view the simple demo")
	e.Logger.Fatal(e.Start(":8080"))
}

// handleHome serves the home page
func handleHome(c echo.Context) error {
	return c.HTML(http.StatusOK, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>ERP Demo</title>
			<script src="https://cdn.tailwindcss.com"></script>
			<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
			<script src="https://unpkg.com/htmx.org@1.9.10"></script>
		</head>
		<body class="bg-background text-foreground">
			<div class="min-h-screen flex items-center justify-center">
				<div class="text-center space-y-6">
					<h1 class="text-4xl font-bold">ERP Component Demo</h1>
					<p class="text-xl text-muted-foreground">Enterprise UI Component System</p>
					<div class="flex gap-4 justify-center">
						<a href="/demo" class="inline-flex items-center justify-center rounded-md bg-primary px-6 py-3 text-primary-foreground hover:bg-primary/90 transition-colors">
							Full Demo
						</a>
						<a href="/simple" class="inline-flex items-center justify-center rounded-md bg-green-600 px-6 py-3 text-white hover:bg-green-700 transition-colors">
							Simple Demo
						</a>
					</div>
				</div>
			</div>
		</body>
		</html>
	`)
}

// handleDemo serves the demo page
func handleDemo(c echo.Context) error {
	// Render the demo page
	component := pages.DemoPage()
	return component.Render(context.Background(), c.Response().Writer)
}

// handleSimpleDemo serves the simple demo page
func handleSimpleDemo(c echo.Context) error {
	// Render the simple demo page
	component := pages.SimpleDemoPage()
	return component.Render(context.Background(), c.Response().Writer)
}

// handleHTMXDemo serves dynamic HTMX content
func handleHTMXDemo(c echo.Context) error {
	time.Sleep(500 * time.Millisecond) // Simulate loading

	html := `
		<div class="space-y-4">
			<h5 class="font-medium text-green-600">✅ Content Loaded Successfully!</h5>
			<p class="text-sm">This content was loaded dynamically using HTMX. The request was made when you clicked the button.</p>
			<div class="bg-green-50 border border-green-200 rounded p-3">
				<p class="text-green-800 text-xs">
					<strong>Timestamp:</strong> ` + time.Now().Format("15:04:05") + `<br>
					<strong>Method:</strong> GET /api/demo/htmx-content<br>
					<strong>Status:</strong> 200 OK
				</p>
			</div>
			<button 
				class="text-xs text-blue-600 hover:text-blue-800"
				hx-get="/api/demo/htmx-content" 
				hx-target="#htmx-demo-content"
				hx-swap="innerHTML">
				🔄 Reload Content
			</button>
		</div>
	`
	return c.HTML(http.StatusOK, html)
}

// handleContact handles contact form submissions
func handleContact(c echo.Context) error {
	time.Sleep(300 * time.Millisecond) // Simulate processing

	name := c.FormValue("name")
	email := c.FormValue("email")
	message := c.FormValue("message")

	if name == "" || email == "" || message == "" {
		return c.HTML(http.StatusBadRequest, `
			<div class="bg-red-50 border border-red-200 rounded p-4 text-red-800">
				<strong>❌ Error:</strong> All fields are required.
			</div>
		`)
	}

	html := `
		<div class="bg-green-50 border border-green-200 rounded p-4 text-green-800">
			<h4 class="font-medium">✅ Message Sent Successfully!</h4>
			<p class="text-sm mt-2">
				Thank you <strong>` + name + `</strong>! We received your message and will respond to 
				<strong>` + email + `</strong> shortly.
			</p>
			<p class="text-xs mt-2 opacity-75">Message: "` + message[:min(len(message), 50)] + `..."</p>
		</div>
	`
	return c.HTML(http.StatusOK, html)
}

// handleSearch handles search requests
func handleSearch(c echo.Context) error {
	query := c.QueryParam("search")
	if query == "" {
		return c.HTML(http.StatusOK, `<p class="text-sm text-muted-foreground">Start typing to search...</p>`)
	}

	time.Sleep(200 * time.Millisecond) // Simulate search delay

	// Mock search results
	results := []string{
		"Button Component",
		"Input Component", 
		"Form Component",
		"Table Component",
		"Modal Component",
		"Navigation Component",
	}

	html := `<div class="space-y-2">`
	found := false
	for _, result := range results {
		if containsIgnoreCase(result, query) {
			html += `<div class="p-2 hover:bg-accent rounded text-sm cursor-pointer">🔍 ` + result + `</div>`
			found = true
		}
	}
	if !found {
		html += `<p class="text-sm text-muted-foreground">No results found for "` + query + `"</p>`
	}
	html += `</div>`

	return c.HTML(http.StatusOK, html)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		   strings.ToLower(s) == strings.ToLower(substr) ||
		   strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}