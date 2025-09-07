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
	var port = flag.String("port", "", "Port to serve on")
	flag.Parse()

	// Get port from environment variable, flag, or default
	portNum := "8081"
	if envPort := os.Getenv("DOC_PORT"); envPort != "" {
		portNum = envPort
	}
	if *port != "" {
		portNum = *port
	}

	// Get the current directory (site directory)
	siteDir := "./site"

	// Handle all requests with custom handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFileWithIndex(w, r, siteDir)
	})

	// Start server on specified port
	portAddr := ":" + portNum
	url := fmt.Sprintf("http://localhost:%s", portNum)
	fmt.Printf("Starting server on %s\n", url)
	absPath, _ := filepath.Abs(siteDir)
	fmt.Printf("Serving files from: %s\n", absPath)

	// Open URL in default browser
	go func() {
		// time.Sleep(1 * time.Second) // Wait for server to start
		openBrowser(url)
	}()

	log.Fatal(http.ListenAndServe(portAddr, nil))
}

// serveFileWithIndex serves files and directories, automatically serving index.html for directories
func serveFileWithIndex(w http.ResponseWriter, r *http.Request, root string) {
	path := filepath.Join(root, r.URL.Path)

	// Clean the path
	path = filepath.Clean(path)

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// If it's a directory, try to serve index.html
	if info.IsDir() {
		indexPath := filepath.Join(path, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		// If no index.html, return 404 to avoid directory listing
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, path)
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
