package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

var (
	schemaDir = flag.String("schema-dir", "docs/ui/Schema", "Path to UI schema directory")
	verbose   = flag.Bool("verbose", false, "Enable verbose output")
)

func main() {
	flag.Parse()

	if *verbose {
		fmt.Printf("Generating CSS DataTypes from schemas in: %s\n", *schemaDir)
	}

	// Create generator
	generator := css.NewDataTypeGenerator(*schemaDir)

	// Load all DataType schemas
	if err := generator.LoadAllDataTypes(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading DataType schemas: %v\n", err)
		os.Exit(1)
	}

	// Generate Go types
	if err := generator.GenerateAllTypes(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Go types: %v\n", err)
		os.Exit(1)
	}

	// Write all types to files
	if err := generator.WriteAllTypes(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing type files: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Println("✅ CSS DataType generation completed successfully!")
	}
}