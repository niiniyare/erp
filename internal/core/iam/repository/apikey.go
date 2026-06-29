package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
)

// Port (interface)

// APIKeyRepository is the persistence port for API key management.
//
// Create / Revoke / List use WithTenantFromCtx — the caller must ensure
// tenant ID is in ctx (cache.TenantIDKey).
//
// GetByHash intentionally bypasses tenant context: the key_hash is globally
// unique and the tenant is unknown until after the lookup.  This method must
// be backed by a DB role with BYPASSRLS (admin_role) at the infrastructure
// level — the same pattern used by session token validation.
type APIKeyRepository interface {
	// Create inserts a new API key for the current tenant.
	// keyHash is the pre-computed SHA-256 hex of the raw bearer token.
	Create(ctx context.Context, req *domain.CreateAPIKeyRequest, keyHash string) (*domain.APIKey, error)

	// GetByHash fetches an active (non-revoked, non-expired) key by its SHA-256
	// hash.  Returns nil, nil when not found.  Runs without tenant context.
	GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)

	// Revoke marks the key as revoked for the current tenant.
	Revoke(ctx context.Context, keyID uuid.UUID) error

	// List returns all API keys for the current tenant, newest first.
	List(ctx context.Context) ([]*domain.APIKey, error)
}

// Adapter (implementation)

type apiKeyRepo struct {
	store db.Store
}

// NewAPIKeyRepository constructs an APIKeyRepository backed by PostgreSQL.
func NewAPIKeyRepository(store db.Store) APIKeyRepository {
	return &apiKeyRepo{store: store}
}

func (r *apiKeyRepo) Create(ctx context.Context, req *domain.CreateAPIKeyRequest, keyHash string) (*domain.APIKey, error) {
	var expiresAt sql.NullTime
	if req.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *req.ExpiresAt, Valid: true}
	}

	var result *domain.APIKey
	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		row, err := s.CreateAPIKey(ctx, db.CreateAPIKeyParams{
			Name:      req.Name,
			KeyHash:   keyHash,
			Scopes:    req.Scopes,
			ExpiresAt: expiresAt,
			CreatedBy: req.CreatedBy,
		})
		if err != nil {
			return fmt.Errorf("api key repo: create: %w", err)
		}
		result = rowToAPIKey(row)
		return nil
	})
	return result, err
}

// GetByHash fetches an active key by hash — no tenant context (see interface comment).
func (r *apiKeyRepo) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	row, err := r.store.GetAPIKeyByHash(ctx, keyHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // not found / expired / revoked
	}
	if err != nil {
		return nil, fmt.Errorf("api key repo: get by hash: %w", err)
	}
	return rowHashToAPIKey(row), nil
}

func (r *apiKeyRepo) Revoke(ctx context.Context, keyID uuid.UUID) error {
	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		if err := s.RevokeAPIKey(ctx, keyID); err != nil {
			return fmt.Errorf("api key repo: revoke: %w", err)
		}
		return nil
	})
}

func (r *apiKeyRepo) List(ctx context.Context) ([]*domain.APIKey, error) {
	var out []*domain.APIKey
	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		rows, err := s.ListAPIKeys(ctx)
		if err != nil {
			return fmt.Errorf("api key repo: list: %w", err)
		}
		out = make([]*domain.APIKey, 0, len(rows))
		for _, row := range rows {
			out = append(out, rowListToAPIKey(row))
		}
		return nil
	})
	return out, err
}

// Mapping helpers

func rowToAPIKey(row *db.ApiKey) *domain.APIKey {
	k := &domain.APIKey{
		ID:        row.ID,
		TenantID:  row.TenantID,
		Name:      row.Name,
		Scopes:    row.Scopes,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		k.ExpiresAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		k.RevokedAt = &t
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		k.LastUsedAt = &t
	}
	return k
}

func rowHashToAPIKey(row *db.GetAPIKeyByHashRow) *domain.APIKey {
	k := &domain.APIKey{
		ID:        row.ID,
		TenantID:  row.TenantID,
		Name:      row.Name,
		Scopes:    row.Scopes,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		k.ExpiresAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		k.RevokedAt = &t
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		k.LastUsedAt = &t
	}
	return k
}

func rowListToAPIKey(row *db.ListAPIKeysRow) *domain.APIKey {
	k := &domain.APIKey{
		ID:        row.ID,
		TenantID:  row.TenantID,
		Name:      row.Name,
		Scopes:    row.Scopes,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		k.ExpiresAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		k.RevokedAt = &t
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		k.LastUsedAt = &t
	}
	return k
}
