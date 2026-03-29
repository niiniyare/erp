package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/iam/domain"
)

// Port (interface)

// SSORepository defines the persistence port for SSO provider configurations.
// All methods use current_tenant_id() via the DB session — the caller must
// ensure the tenant ID is in ctx (cache.TenantIDKey) before calling.
type SSORepository interface {
	// GetProvider returns the active SSO provider config for the current tenant.
	// Returns nil, nil when not found.
	GetProvider(ctx context.Context, provider domain.OAuthProvider) (*domain.SSOProvider, error)

	// UpsertProvider inserts or updates the SSO provider config for the current tenant.
	UpsertProvider(ctx context.Context, req *domain.CreateSSOProviderRequest, encryptedSecret string) (*domain.SSOProvider, error)

	// DeactivateProvider marks the provider as inactive for the current tenant.
	DeactivateProvider(ctx context.Context, provider domain.OAuthProvider) error

	// ListProviders returns all SSO providers configured for the current tenant.
	ListProviders(ctx context.Context) ([]*domain.SSOProvider, error)
}

// Adapter (implementation)

type ssoRepo struct {
	store db.Store
}

// NewSSORepository constructs an SSORepository backed by PostgreSQL.
func NewSSORepository(store db.Store) SSORepository {
	return &ssoRepo{store: store}
}

func (r *ssoRepo) GetProvider(ctx context.Context, provider domain.OAuthProvider) (*domain.SSOProvider, error) {
	var result *domain.SSOProvider
	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		row, err := s.GetSSOProvider(ctx, string(provider))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // not found
		}
		if err != nil {
			return fmt.Errorf("sso repo: get provider: %w", err)
		}
		p, err := rowToSSOProvider(row)
		if err != nil {
			return err
		}
		result = p
		return nil
	})
	return result, err
}

func (r *ssoRepo) UpsertProvider(ctx context.Context, req *domain.CreateSSOProviderRequest, encryptedSecret string) (*domain.SSOProvider, error) {
	extraJSON, err := json.Marshal(req.ExtraParams)
	if err != nil {
		return nil, fmt.Errorf("sso repo: marshal extra_params: %w", err)
	}

	var defaultEntityID *uuid.UUID
	if req.DefaultEntityID != uuid.Nil {
		defaultEntityID = &req.DefaultEntityID
	}

	var result *domain.SSOProvider
	err = r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		row, err := s.UpsertSSOProvider(ctx, db.UpsertSSOProviderParams{
			Provider:        string(req.Provider),
			ClientID:        req.ClientID,
			ClientSecretEnc: encryptedSecret,
			Scopes:          req.Scopes,
			RedirectUri:     req.RedirectURI,
			ExtraParams:     extraJSON,
			AutoProvision:   req.AutoProvision,
			DefaultEntityID: defaultEntityID,
		})
		if err != nil {
			return fmt.Errorf("sso repo: upsert provider: %w", err)
		}
		p, err := rowToSSOProvider(row)
		if err != nil {
			return err
		}
		result = p
		return nil
	})
	return result, err
}

func (r *ssoRepo) DeactivateProvider(ctx context.Context, provider domain.OAuthProvider) error {
	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		if err := s.DeactivateSSOProvider(ctx, string(provider)); err != nil {
			return fmt.Errorf("sso repo: deactivate provider: %w", err)
		}
		return nil
	})
}

func (r *ssoRepo) ListProviders(ctx context.Context) ([]*domain.SSOProvider, error) {
	var out []*domain.SSOProvider
	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		rows, err := s.ListSSOProviders(ctx)
		if err != nil {
			return fmt.Errorf("sso repo: list providers: %w", err)
		}
		out = make([]*domain.SSOProvider, 0, len(rows))
		for _, row := range rows {
			p, err := rowToSSOProvider(row)
			if err != nil {
				return err
			}
			out = append(out, p)
		}
		return nil
	})
	return out, err
}

// Mapping helpers

func rowToSSOProvider(row *db.SsoProvider) (*domain.SSOProvider, error) {
	extra := make(map[string]string)
	if row.ExtraParams != nil {
		if err := json.Unmarshal(row.ExtraParams, &extra); err != nil {
			return nil, fmt.Errorf("sso repo: unmarshal extra_params: %w", err)
		}
	}

	var defaultEntityID uuid.UUID
	if row.DefaultEntityID != nil {
		defaultEntityID = *row.DefaultEntityID
	}

	return &domain.SSOProvider{
		ID:              row.ID,
		TenantID:        row.TenantID,
		Provider:        domain.OAuthProvider(row.Provider),
		ClientID:        row.ClientID,
		ClientSecretEnc: row.ClientSecretEnc,
		Scopes:          row.Scopes,
		RedirectURI:     row.RedirectUri,
		ExtraParams:     extra,
		AutoProvision:   row.AutoProvision,
		DefaultEntityID: defaultEntityID,
		IsActive:        row.IsActive,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}
