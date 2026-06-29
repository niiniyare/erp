// Package migrateCmd implements the `awo migrate` subcommand.
package migrateCmd

import (
	"context"
	"fmt"
	"os"

	"awo.so/framework/migrate"
)

// Run executes the migrate subcommand with args.
func Run(ctx context.Context, args []string) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	sub := "up"
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "up":
		fmt.Println("awo migrate: applying pending migrations…")
		if err := migrate.Run(ctx, dsn, migrate.Up); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
		fmt.Println("awo migrate: up — done")
		return nil

	case "down":
		fmt.Println("awo migrate: reverting last migration…")
		if err := migrate.Run(ctx, dsn, migrate.Down); err != nil {
			return fmt.Errorf("migrate down: %w", err)
		}
		fmt.Println("awo migrate: down — done")
		return nil

	case "status", "version":
		v, dirty, err := migrate.Version(ctx, dsn)
		if err != nil {
			return fmt.Errorf("migrate version: %w", err)
		}
		if v == 0 {
			fmt.Println("awo migrate: no migrations applied")
			return nil
		}
		dirtyStr := ""
		if dirty {
			dirtyStr = " (dirty)"
		}
		fmt.Printf("awo migrate: current version %d%s\n", v, dirtyStr)
		return nil

	default:
		return fmt.Errorf("unknown migrate subcommand %q — use up|down|status", sub)
	}
}
