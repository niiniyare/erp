// Package bootstrap handles one-time idempotent startup tasks.
// Currently: seeding the initial platform administrator account.
package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"awo.so/internal/platform/config"
)

const bcryptCost = 12

// AdminConfig holds credentials for the initial platform admin.
type AdminConfig struct {
	Email       string
	Password    string
	DisplayName string
}

// AdminConfigFromEnv reads bootstrap config from environment variables.
//
//	PLATFORM_ADMIN_EMAIL     — required; skips bootstrap when absent
//	PLATFORM_ADMIN_PASSWORD  — required; skips bootstrap when absent
//	PLATFORM_ADMIN_NAME      — optional display name (default: "Platform Administrator")
//
// Returns nil when either required var is unset — bootstrap is skipped silently.
func AdminConfigFromEnv() *AdminConfig {
	email := os.Getenv("PLATFORM_ADMIN_EMAIL")
	password := os.Getenv("PLATFORM_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return nil
	}
	name := os.Getenv("PLATFORM_ADMIN_NAME")
	if name == "" {
		name = "Platform Administrator"
	}
	return &AdminConfig{Email: email, Password: password, DisplayName: name}
}

// Run bootstraps the platform admin using a short-lived connection derived
// from cfg. Safe to call on every startup — inserts only when no SYSADMIN
// user exists yet.
//
// Returns nil (no-op) when PLATFORM_ADMIN_EMAIL or PLATFORM_ADMIN_PASSWORD
// are not set.
func Run(ctx context.Context, cfg *config.Config) error {
	adminCfg := AdminConfigFromEnv()
	if adminCfg == nil {
		return nil
	}

	pool, err := pgxpool.New(ctx, cfg.Database.GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("bootstrap: connect db: %w", err)
	}
	defer pool.Close()

	return EnsurePlatformAdmin(ctx, pool, adminCfg)
}

// EnsurePlatformAdmin inserts a SYSADMIN user when none exists.
// Idempotent — the WHERE NOT EXISTS guard makes repeated calls safe.
//
// Platform users are not scoped to any tenant: tenant_id and entity_id are
// both set to the nil UUID (00000000-…). The awo.tenant_id GUC is set to
// the nil UUID inside the transaction so that any RLS policies that match
// on current_tenant_id() do not reject the insert.
func EnsurePlatformAdmin(ctx context.Context, pool *pgxpool.Pool, cfg *AdminConfig) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcryptCost)
	if err != nil {
		return fmt.Errorf("bootstrap: hash password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Set tenant GUC to nil UUID so RLS policies (if present) accept the row.
	// Ignore the error — the GUC may not exist when RLS is not configured.
	_, _ = tx.Exec(ctx, `SET LOCAL awo.tenant_id = '00000000-0000-0000-0000-000000000000'`)

	const q = `
		INSERT INTO users (
			id,          tenant_id,   entity_id,
			email,       username,    display_name,
			password_hash,           user_type,
			account_status,          is_active,
			created_at,              updated_at
		)
		SELECT
			gen_random_uuid(),
			'00000000-0000-0000-0000-000000000000'::uuid,
			'00000000-0000-0000-0000-000000000000'::uuid,
			$1, $1, $2,
			$3, 'SYSADMIN',
			'active', true,
			now(), now()
		WHERE NOT EXISTS (
			SELECT 1 FROM users
			WHERE user_type IN ('SYSADMIN', 'PLATFORM')
			  AND deleted_at IS NULL
		)`

	tag, err := tx.Exec(ctx, q, cfg.Email, cfg.DisplayName, string(hash))
	if err != nil {
		return fmt.Errorf("bootstrap: insert platform admin: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("bootstrap: commit: %w", err)
	}

	if tag.RowsAffected() > 0 {
		fmt.Printf("[bootstrap] platform admin created: %s\n", cfg.Email)
	}
	return nil
}
