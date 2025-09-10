package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "", "Port to serve on")
	flag.Parse()

	// Get port from environment variable, flag, or default
	portNum := "8081"
	if envPort := os.Getenv("DOC_PORT"); envPort != "" {
		portNum = envPort
	}
	if *port != "" {
		portNum = *port
	}

	// Get the parent directory to access both site/ and docs/schema/
	baseDir := ".."
	fmt.Printf("Serving MkDocs site from: ../site/\n")
	fmt.Printf("Serving Schema documentation from: schema/\n")

	// Handle schema documentation requests
	http.HandleFunc("/schema/", func(w http.ResponseWriter, r *http.Request) {
		// Remove /schema/ prefix and serve from docs/schema/
		path := r.URL.Path[8:] // Remove "/schema/" prefix
		if path == "" {
			path = "index.html"
		}
		schemaPath := filepath.Join("schema", path)
		serveSchemaFile(w, r, schemaPath)
	})

	// Handle all other requests from the site/ directory
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If root path, serve our custom index.html from docs/
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}

		sitePath := filepath.Join(baseDir, "site", r.URL.Path)
		serveStaticFile(w, r, sitePath)
	})

	// Start server on specified port
	portAddr := ":" + portNum
	url := fmt.Sprintf("http://localhost:%s", portNum)
	fmt.Printf("Starting server on %s\n", url)
	absPath, _ := filepath.Abs(baseDir)
	fmt.Printf("Base directory: %s\n", absPath)

	// Open URL in default browser
	go func() {
		// time.Sleep(1 * time.Second) // Wait for server to start
		openBrowser(url)
	}()

	log.Fatal(http.ListenAndServe(portAddr, nil))
}

// serveStaticFile serves files from the MkDocs site directory
func serveStaticFile(w http.ResponseWriter, r *http.Request, fullPath string) {
	// Clean the path
	fullPath = filepath.Clean(fullPath)

	// Check if path exists
	info, err := os.Stat(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// If it's a directory, try to serve index.html
	if info.IsDir() {
		indexPath := filepath.Join(fullPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		// If no index.html, return 404 to avoid directory listing
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// serveSchemaFile serves files from the schema directory
func serveSchemaFile(w http.ResponseWriter, r *http.Request, relativePath string) {
	// Clean the path
	relativePath = filepath.Clean(relativePath)

	// Check if path exists
	info, err := os.Stat(relativePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// If it's a directory, try to serve index.html
	if info.IsDir() {
		indexPath := filepath.Join(relativePath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		// If no index.html, return 404 to avoid directory listing
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, relativePath)
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) {
	var err error

	// Check if we're in Termux first
	if _, err := exec.LookPath("termux-open"); err == nil {
		err = exec.Command("termux-open", url).Start()
	} else {
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", url).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		case "darwin":
			err = exec.Command("open", url).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}
	}

	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Printf("Please open %s manually\n", url)
	} else {
		fmt.Printf("Opening %s in default browser...\n", url)
	}
}
