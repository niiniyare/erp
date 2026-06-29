package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/contracts/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/tracing"
)

// postgres is the PostgreSQL-backed implementation of contracts Repository.
// All data access is wrapped in WithTenantFromCtx to enforce RLS tenant isolation.
type postgres struct {
	store  db.Store
	tracer tracing.Service
	log    logger.Logger
}

// NewPostgres returns a new contract repository backed by PostgreSQL.
func NewPostgres(store db.Store, tracer tracing.Service, log logger.Logger) Repository {
	return &postgres{
		store:  store,
		tracer: tracer,
		log:    log,
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

// Create inserts a new contract row. The tenant_id is injected by the DB
// function current_tenant_id() which is set by WithTenantFromCtx.
func (r *postgres) Create(ctx context.Context, c *domain.Contract) error {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.Create")
	defer span.End()

	// Verify tenant context is present before hitting the DB.
	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	r.log.DebugContext(ctx, "creating contract", logger.Fields{
		"number":    c.Number,
		"entity_id": c.EntityID,
	})

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		metadata, err := marshalMetadata(c.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}

		params := db.CreateContractParams{
			EntityID:          c.EntityID,
			Number:            c.Number,
			Title:             c.Title,
			Status:            string(c.Status),
			ContractType:      string(c.ContractType),
			CounterpartyName:  c.CounterpartyName,
			CounterpartyEmail: c.CounterpartyEmail,
			StartDate:         c.StartDate,
			EndDate:           ptrToTimeOrZero(c.EndDate), // NOT NULL in DB; zero if unset
			Value:             decimalToPgNumeric(c.Value),
			CurrencyCode:      c.CurrencyCode,
			Description:       c.Description,
			Terms:             c.Terms,
			Metadata:          metadata,
			CreatedBy:         c.CreatedBy,
		}

		row, err := s.CreateContract(ctx, params)
		if err != nil {
			r.log.ErrorContext(ctx, "failed to create contract", logger.Fields{
				"number": c.Number,
				"error":  err.Error(),
			})
			return parseContractDBError(err, "Create")
		}

		// Populate generated fields back onto the domain object.
		c.ID = row.ID
		c.TenantID = row.TenantID
		c.Version = row.Version
		c.CreatedAt = row.CreatedAt
		c.UpdatedAt = row.UpdatedAt
		return nil
	})
}

// ── GetByID ───────────────────────────────────────────────────────────────────

// GetByID fetches a non-deleted contract by its UUID within the current tenant.
func (r *postgres) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contract, error) {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.GetByID")
	defer span.End()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.Contract

	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		row, err := s.GetContractByID(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrContractNotFound
			}
			r.log.ErrorContext(ctx, "failed to get contract by ID", logger.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return parseContractDBError(err, "GetByID")
		}
		result = mapRowToDomain(row)
		return nil
	})

	return result, err
}

// ── GetByNumber ───────────────────────────────────────────────────────────────

// GetByNumber fetches a contract by its human-readable number (e.g. "CTR-2024-001").
func (r *postgres) GetByNumber(ctx context.Context, number string) (*domain.Contract, error) {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.GetByNumber")
	defer span.End()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	var result *domain.Contract

	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		row, err := s.GetContractByNumber(ctx, number)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrContractNotFound
			}
			r.log.ErrorContext(ctx, "failed to get contract by number", logger.Fields{
				"number": number,
				"error":  err.Error(),
			})
			return parseContractDBError(err, "GetByNumber")
		}
		result = mapRowToDomain(row)
		return nil
	})

	return result, err
}

// ── Update ────────────────────────────────────────────────────────────────────

// Update applies partial changes to a contract using optimistic locking.
// The WHERE clause checks version = req.Version so concurrent edits fail fast.
func (r *postgres) Update(ctx context.Context, id uuid.UUID, req domain.UpdateContractRequest) (*domain.Contract, error) {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.Update")
	defer span.End()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	r.log.DebugContext(ctx, "updating contract", logger.Fields{
		"id":      id,
		"version": req.Version,
	})

	var result *domain.Contract

	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		// Lock the row first to detect optimistic-lock conflicts before UPDATE.
		existing, err := s.GetContractByIDForUpdate(ctx, id)
		if err != nil {
			if err == db.ErrNoRows {
				return domain.ErrContractNotFound
			}
			return parseContractDBError(err, "Update.Lock")
		}

		// Reject if the client is working from a stale version.
		if existing.Version != req.Version {
			r.log.WarnContext(ctx, "contract version conflict", logger.Fields{
				"id":       id,
				"expected": req.Version,
				"actual":   existing.Version,
			})
			return domain.ErrVersionConflict
		}

		// Build update params — only non-nil request fields are applied (COALESCE in SQL).
		// StartDate/EndDate are NOT NULL in DB so fall back to existing row values when
		// the caller did not provide new ones (COALESCE can't handle non-nullable params).
		startDate := existing.StartDate
		if req.StartDate != nil {
			startDate = *req.StartDate
		}
		endDate := existing.EndDate
		if req.EndDate != nil {
			endDate = *req.EndDate
		}

		params := db.UpdateContractParams{
			ID:                id,
			Version:           req.Version,
			Title:             req.Title,
			Status:            req.Status,
			ContractType:      req.ContractType,
			CounterpartyName:  req.CounterpartyName,
			CounterpartyEmail: req.CounterpartyEmail,
			StartDate:         startDate,
			EndDate:           endDate,
			CurrencyCode:      req.CurrencyCode,
			Description:       req.Description,
			Terms:             req.Terms,
			SignedBy:          req.SignedBy,
			SignedAt:          ptrToNullTime(req.SignedAt), // *time.Time → sql.NullTime
			Value:             existing.Value,              // default to existing pgtype.Numeric directly
		}

		// Override Value if caller explicitly supplied a new one.
		if req.Value != nil {
			params.Value = decimalToPgNumeric(*req.Value)
		}

		row, err := s.UpdateContract(ctx, params)
		if err != nil {
			if err == db.ErrNoRows {
				// No row updated means concurrent modification; treat as conflict.
				return domain.ErrVersionConflict
			}
			r.log.ErrorContext(ctx, "failed to update contract", logger.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return parseContractDBError(err, "Update")
		}

		result = mapRowToDomain(row)
		return nil
	})

	return result, err
}

// ── SoftDelete ────────────────────────────────────────────────────────────────

// SoftDelete marks a contract as deleted by setting deleted_at = NOW().
// The contract remains in the database for audit purposes.
func (r *postgres) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.SoftDelete")
	defer span.End()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	r.log.DebugContext(ctx, "soft-deleting contract", logger.Fields{"id": id})

	return r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		err := s.SoftDeleteContract(ctx, id)
		if err != nil {
			r.log.ErrorContext(ctx, "failed to soft-delete contract", logger.Fields{
				"id":    id,
				"error": err.Error(),
			})
			return parseContractDBError(err, "SoftDelete")
		}
		return nil
	})
}

// ── List ──────────────────────────────────────────────────────────────────────

// List returns a paginated, filtered slice of contracts and the total match count.
// Both the data query and the count query run inside the same tenant-scoped transaction.
func (r *postgres) List(ctx context.Context, filter domain.ContractFilter) ([]*domain.Contract, int64, error) {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.List")
	defer span.End()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return nil, 0, fmt.Errorf("tenant ID not found in context")
	}

	var (
		contracts []*domain.Contract
		total     int64
	)

	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		listParams := db.ListContractsParams{
			EntityID:     filter.EntityID,
			Status:       filter.Status,
			ContractType: filter.ContractType,
			Search:       filter.Search,
			LimitCount:   filter.Limit,
			OffsetCount:  filter.Offset,
		}

		rows, err := s.ListContracts(ctx, listParams)
		if err != nil {
			r.log.ErrorContext(ctx, "failed to list contracts", logger.Fields{"error": err.Error()})
			return parseContractDBError(err, "List")
		}

		// Count total matches (without pagination) for the caller's paging metadata.
		countParams := db.CountContractsParams{
			EntityID:     filter.EntityID,
			Status:       filter.Status,
			ContractType: filter.ContractType,
			Search:       filter.Search,
		}

		total, err = s.CountContracts(ctx, countParams)
		if err != nil {
			r.log.ErrorContext(ctx, "failed to count contracts", logger.Fields{"error": err.Error()})
			return parseContractDBError(err, "Count")
		}

		contracts = make([]*domain.Contract, 0, len(rows))
		for _, row := range rows {
			contracts = append(contracts, mapRowToDomain(row))
		}

		return nil
	})

	return contracts, total, err
}

// ── NumberExists ──────────────────────────────────────────────────────────────

// NumberExists returns true if a non-deleted contract with the given number
// already exists for the current tenant. Used to enforce uniqueness before insert.
func (r *postgres) NumberExists(ctx context.Context, number string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx, "contracts.repo.NumberExists")
	defer span.End()

	if _, ok := shared.GetTenantID(ctx); !ok {
		return false, fmt.Errorf("tenant ID not found in context")
	}

	var exists bool

	err := r.store.WithTenantFromCtx(ctx, func(ctx context.Context, s db.Store) error {
		var err error
		exists, err = s.CheckContractNumberExists(ctx, number)
		if err != nil {
			r.log.ErrorContext(ctx, "failed to check contract number existence", logger.Fields{
				"number": number,
				"error":  err.Error(),
			})
			return parseContractDBError(err, "NumberExists")
		}
		return nil
	})

	return exists, err
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// mapRowToDomain converts a SQLC-generated Contract row to the domain entity.
func mapRowToDomain(row *db.Contract) *domain.Contract {
	c := &domain.Contract{
		ID:                row.ID,
		TenantID:          row.TenantID,
		EntityID:          row.EntityID,
		Number:            row.Number,
		Title:             row.Title,
		Status:            domain.ContractStatus(row.Status),
		ContractType:      domain.ContractType(row.ContractType),
		CounterpartyName:  row.CounterpartyName,
		CounterpartyEmail: row.CounterpartyEmail,
		StartDate:         row.StartDate,
		EndDate:           timeToPtr(row.EndDate), // DB is NOT NULL; treat zero as nil in domain
		Value:             pgNumericToDecimal(row.Value),
		CurrencyCode:      row.CurrencyCode,
		Description:       row.Description,
		Terms:             row.Terms,
		SignedBy:          row.SignedBy,
		Version:           row.Version,
		CreatedBy:         row.CreatedBy,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}

	// Map nullable timestamptz to *time.Time for signed_at.
	if row.SignedAt.Valid {
		t := row.SignedAt.Time
		c.SignedAt = &t
	}

	// Map nullable timestamptz to *time.Time for deleted_at.
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		c.DeletedAt = &t
	}

	// Unmarshal JSONB metadata; fall back to empty map on error.
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &c.Metadata)
	}
	if c.Metadata == nil {
		c.Metadata = make(map[string]any)
	}

	return c
}

// decimalToPgNumeric converts a shopspring Decimal to pgtype.Numeric for storage.
// The Exp field preserves fractional precision (e.g. 12.34 → Int=1234, Exp=-2).
func decimalToPgNumeric(d decimal.Decimal) pgtype.Numeric {
	if d.IsZero() {
		return pgtype.Numeric{Int: new(big.Int), Exp: 0, Valid: true}
	}
	return pgtype.Numeric{
		Int:   d.BigInt(),
		Exp:   int32(d.Exponent()),
		Valid: true,
	}
}

// pgNumericToDecimal converts a pgtype.Numeric from the DB back to a shopspring Decimal.
func pgNumericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid || n.Int == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(n.Int, int32(n.Exp))
}

// marshalMetadata serialises the contract's metadata map to JSONB bytes.
// An empty/nil map is written as '{}' to satisfy the NOT NULL constraint.
func marshalMetadata(m map[string]any) ([]byte, error) {
	if len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// nullTimeToPtr converts sql.NullTime → *time.Time for use in domain structs.
func nullTimeToPtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

// ptrToTimeOrZero dereferences a *time.Time; returns zero time.Time when nil.
// Used when the DB column is NOT NULL and we must always provide a value.
func ptrToTimeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// timeToPtr converts a time.Time to *time.Time, returning nil for the zero value.
// Bridges NOT NULL DB columns to optional domain fields.
func timeToPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// ptrToNullTime converts *time.Time → sql.NullTime for nullable DB columns.
func ptrToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
