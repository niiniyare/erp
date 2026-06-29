// Command awo is the Awo Framework CLI — analogous to Frappe's `bench`.
//
// Usage:
//
//	awo migrate [up|down|status|version]
//	awo generate entity <name>
//	awo generate migration <name>
package main

import (
	"context"
	"fmt"
	"os"

	"awo.so/framework/cmd/awo/internal/migrateCmd"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "awo:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	ctx := context.Background()

	switch args[0] {
	case "migrate":
		return migrateCmd.Run(ctx, args[1:])
	case "version", "--version", "-v":
		fmt.Println("awo v0.1.0 (framework)")
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q — run 'awo help'", args[0])
	}
}

func printUsage() {
	fmt.Print(`awo — Awo Framework CLI

Usage:
  awo migrate [up|down|status|version]   Database migration management
  awo version                            Print CLI version
  awo help                               Show this help

Environment:
  DATABASE_URL   PostgreSQL connection string (required for migrate)
`)
}
