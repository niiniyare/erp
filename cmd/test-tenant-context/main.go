//go:build database
// +build database

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/niiniyare/erp/internal/core/tenant"
)

func main() {
	fmt.Println("Tenant Context Database Testing Tool")
	fmt.Println("====================================")

	// Check if database URL is set
	if os.Getenv("TEST_DATABASE_URL") == "" {
		fmt.Println("❌ TEST_DATABASE_URL environment variable not set")
		fmt.Println("Set it to something like: postgres://user:password@localhost:5432/dbname?sslmode=disable")
		os.Exit(1)
	}

	// Create database test runner
	runner, err := tenant.NewDatabaseTestRunner()
	if err != nil {
		log.Fatalf("❌ Failed to create database test runner: %v", err)
	}
	defer runner.Close()

	fmt.Println("✅ Successfully connected to database")

	// Run tenant context operations test
	fmt.Println("\n" + strings.Repeat("=", 60))
	if err := runner.TestTenantContextOperations(); err != nil {
		log.Fatalf("❌ Tenant context operations test failed: %v", err)
	}

	// Run session persistence test
	fmt.Println("\n" + strings.Repeat("=", 60))
	if err := runner.TestSessionPersistence(); err != nil {
		log.Fatalf("❌ Session persistence test failed: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 All database tests completed successfully!")
	fmt.Println("The tenant context system is working correctly with PostgreSQL.")
}
