package tenantrepo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/tenant"
	"github.com/niiniyare/erp/internal/platform/database"
	"github.com/niiniyare/erp/internal/shared/errors"
)

type Repo struct {
	db database.Database
}

func New(db database.Database) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, t *tenant.Tenant) error {
	query := `
        INSERT INTO tenants (id, name, subdomain, plan_type, status, settings, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    `

	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.Name, t.Subdomain,
		t.PlanType, t.Status, t.Settings)

	return err
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	query := `
        SELECT id, name, subdomain, plan_type, status, settings, created_at, updated_at
        FROM tenants 
        WHERE id = $1
    `

	var t tenant.Tenant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Subdomain,
		&t.PlanType, &t.Status, &t.Settings,
		&t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return &t, nil
}

func (r *Repo) GetBySubdomain(ctx context.Context, subdomain string) (*tenant.Tenant, error) {
	query := `
        SELECT id, name, subdomain, plan_type, status, settings, created_at, updated_at
        FROM tenants 
        WHERE subdomain = $1
    `

	var t tenant.Tenant
	err := r.db.QueryRowContext(ctx, query, subdomain).Scan(
		&t.ID, &t.Name, &t.Subdomain,
		&t.PlanType, &t.Status, &t.Settings,
		&t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by subdomain: %w", err)
	}

	return &t, nil
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, updates tenant.UpdateTenantRequest) error {
	query := `
        UPDATE tenants 
        SET name = COALESCE($2, name),
            plan_type = COALESCE($3, plan_type),
            status = COALESCE($4, status),
            settings = COALESCE($5, settings),
            updated_at = NOW()
        WHERE id = $1
    `
	
	_, err := r.db.ExecContext(ctx, query, id, updates.Name, updates.PlanType, updates.Status, updates.Settings)
	return err
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenants SET status = 'suspended', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *Repo) List(ctx context.Context, offset, limit int) ([]*tenant.Tenant, error) {
	query := `
        SELECT id, name, subdomain, plan_type, status, settings, created_at, updated_at
        FROM tenants 
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tenants []*tenant.Tenant
	for rows.Next() {
		var t tenant.Tenant
		err := rows.Scan(
			&t.ID, &t.Name, &t.Subdomain,
			&t.PlanType, &t.Status, &t.Settings,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, &t)
	}
	
	return tenants, nil
}

func (r *Repo) Exists(ctx context.Context, subdomain string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE subdomain = $1)`
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, subdomain).Scan(&exists)
	if err != nil {
		return false, err
	}
	
	return exists, nil
}
