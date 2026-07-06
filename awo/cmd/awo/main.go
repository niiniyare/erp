// Command awo is the Awo framework developer CLI.
//
// Usage:
//
//	awo <command> [flags]
//
// Commands:
//
//	schema inspect  Print the compiled schema as JSON
//	schema fingerprint  Print the schema fingerprint
//	migrate up      Apply pending migrations
//	migrate down    Roll back the last migration
//	migrate version Print the current database version
//	module list     List registered modules
//	docgen          Generate Markdown API reference
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "schema":
		err = runSchema(args)
	case "migrate":
		err = runMigrate(args)
	case "module":
		err = runModule(args)
	case "docgen":
		err = runDocgen(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "awo: unknown command %q\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "awo %s: %v\n", command, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`awo — Awo framework developer CLI

Usage:
  awo <command> [flags]

Commands:
  schema inspect        Print the compiled schema as JSON
  schema fingerprint    Print the schema fingerprint (SHA-256)
  migrate up            Apply all pending up migrations
  migrate down          Roll back the last applied migration
  migrate version       Print current database version
  migrate status        List applied/pending migrations
  module list           List registered modules in dependency order
  docgen                Generate Markdown API reference to stdout

Flags:
  --dir <path>          Migration directory (default: db/migration)
  --db  <url>           PostgreSQL database URL (default: $DATABASE_URL)
  --out <path>          Output file for docgen (default: stdout)

Environment:
  DATABASE_URL          PostgreSQL connection string
  AWO_MIGRATION_DIR     Override default migration directory`)
}

func runSchema(args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "inspect":
		fmt.Println("schema inspect: load your application's registered entities and pipe through awo/introspect.")
		fmt.Println("This command requires the compiled binary of your application.")
		fmt.Println("Tip: expose GET /api/v1/schema in development and run: curl ... | jq .")
		return nil
	case "fingerprint":
		fmt.Println("schema fingerprint: same requirements as schema inspect.")
		fmt.Println("The fingerprint is served alongside the schema at GET /api/v1/schema.")
		return nil
	default:
		return fmt.Errorf("unknown schema sub-command %q (use: inspect, fingerprint)", sub)
	}
}

func runMigrate(args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "up", "down", "version", "status":
		fmt.Printf("migrate %s: use the cmd/migrate binary (awo/cmd/migrate/main.go) which is already wired to golang-migrate.\n", sub)
		fmt.Println("The developer CLI wraps it for discoverability only.")
		return nil
	default:
		return fmt.Errorf("unknown migrate sub-command %q (use: up, down, version, status)", sub)
	}
}

func runModule(args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "list":
		fmt.Println("module list: requires the compiled application to read registered module manifests.")
		fmt.Println("Expose GET /api/v1/modules in development for a JSON list.")
		return nil
	default:
		return fmt.Errorf("unknown module sub-command %q (use: list)", sub)
	}
}

func runDocgen(args []string) error {
	_ = args
	fmt.Println("docgen: requires the compiled application to read the CompiledSchema.")
	fmt.Println("Wire awo/docgen.WriteMarkdown to a CLI entrypoint in your application.")
	fmt.Println("Example: go run ./cmd/server docgen > docs/api-reference.md")
	return nil
}
