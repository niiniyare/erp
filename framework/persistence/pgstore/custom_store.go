package pgstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"awo.so/framework/def"
	filterPkg "awo.so/framework/filter"
	"awo.so/framework/persistence"
	"awo.so/framework/persistence/sqlbuilder"
)

// customEntityRecordsTable is the shared JSONB table used by all custom entities.
const customEntityRecordsTable = "custom_entity_records"

// CustomStore is a pgx-backed EntityStore for EntityTypeCustom entities.
// All field values are stored in the JSONB data column of custom_entity_records.
// The entity_type column acts as the table discriminator.
//
// Schema (see migration 000200_custom_entity_records.up.sql):
//
//	CREATE TABLE custom_entity_records (
//	    id          uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
//	    tenant_id   uuid NOT NULL,                       -- RLS enforcement
//	    entity_type text NOT NULL,                       -- EntityDefinition.Name
//	    data        jsonb NOT NULL DEFAULT '{}',
//	    created_at  timestamptz NOT NULL DEFAULT NOW(),
//	    updated_at  timestamptz NOT NULL DEFAULT NOW(),
//	    deleted_at  timestamptz                          -- soft-delete
//	);
//	CREATE INDEX ON custom_entity_records USING GIN (data);
//	CREATE INDEX ON custom_entity_records (tenant_id, entity_type);
type CustomStore struct {
	def      *def.EntityDefinition
	tenantID uuid.UUID
	q        Querier
}

// NewCustomStore creates a CustomStore for def scoped to tenantID.
func NewCustomStore(def *def.EntityDefinition, tenantID uuid.UUID, q Querier) *CustomStore {
	return &CustomStore{def: def, tenantID: tenantID, q: q}
}

// ── EntityStore implementation ─────────────────────────────────────────────────

func (s *CustomStore) FindByID(ctx context.Context, id uuid.UUID) (def.Record, error) {
	const q = `
		SELECT id, tenant_id, data, created_at, updated_at
		FROM custom_entity_records
		WHERE id = $1
		  AND entity_type = $2
		  AND deleted_at IS NULL`

	row := s.q.QueryRow(ctx, q, id, s.def.Name)
	return s.scanRow(row)
}

func (s *CustomStore) List(ctx context.Context, opts persistence.ListOptions) (persistence.Page, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	args := []any{s.def.Name}
	idx := 2

	var filters []string
	if opts.Search != "" {
		// Full-text search across all text values in the JSONB document.
		filters = append(filters, fmt.Sprintf("data::text ILIKE $%d", idx))
		args = append(args, "%"+escapeLikeCustom(opts.Search)+"%")
		idx++
	}

	where := "entity_type = $1 AND deleted_at IS NULL"
	if len(filters) > 0 {
		where += " AND " + strings.Join(filters, " AND ")
	}

	orderDir := "DESC"
	if opts.Ascending {
		orderDir = "ASC"
	}
	orderCol := "created_at"
	if opts.OrderBy != "" {
		// Only allow ordering by top-level JSON keys or system columns.
		switch opts.OrderBy {
		case "created_at", "updated_at", "id":
			orderCol = opts.OrderBy
		default:
			// JSONB key path ordering: ORDER BY data->>'field_name'
			orderCol = fmt.Sprintf("data->>'%s'", strings.ReplaceAll(opts.OrderBy, "'", "''"))
		}
	}

	countQ := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s",
		customEntityRecordsTable, where,
	)
	var total int64
	if err := s.q.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return persistence.Page{}, fmt.Errorf("CustomStore.List count %s: %w", s.def.Name, err)
	}

	listQ := fmt.Sprintf(
		`SELECT id, tenant_id, data, created_at, updated_at
		 FROM %s
		 WHERE %s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		customEntityRecordsTable, where, orderCol, orderDir, idx, idx+1,
	)
	args = append(args, limit, opts.Offset)

	rows, err := s.q.Query(ctx, listQ, args...)
	if err != nil {
		return persistence.Page{}, fmt.Errorf("CustomStore.List %s: %w", s.def.Name, err)
	}
	defer rows.Close()

	var records []def.Record
	for rows.Next() {
		rec, err := s.scanRows(rows)
		if err != nil {
			return persistence.Page{}, fmt.Errorf("CustomStore.List scan %s: %w", s.def.Name, err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return persistence.Page{}, fmt.Errorf("CustomStore.List rows %s: %w", s.def.Name, err)
	}

	return persistence.Page{
		Records: records,
		Total:   total,
		Limit:   limit,
		Offset:  opts.Offset,
	}, nil
}

func (s *CustomStore) Create(ctx context.Context, rec def.MutableRecord) error {
	if rec.ID() == uuid.Nil {
		rec.Set("_id", uuid.New())
	}

	data, err := s.marshalData(rec)
	if err != nil {
		return fmt.Errorf("CustomStore.Create marshal %s: %w", s.def.Name, err)
	}

	id := rec.ID()
	if id == uuid.Nil {
		id = uuid.New()
	}
	now := time.Now().UTC()

	const q = `
		INSERT INTO custom_entity_records
		    (id, tenant_id, entity_type, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (id) DO NOTHING`

	tag, err := s.q.Exec(ctx, q, id, s.tenantID, s.def.Name, data, now)
	if err != nil {
		return fmt.Errorf("CustomStore.Create %s: %w", s.def.Name, mapPgError(err))
	}
	if tag.RowsAffected() == 0 {
		return persistence.ErrConflict
	}

	// Stamp generated ID and timestamps back onto the mutable record.
	rec.Set("id", id)
	rec.Set("created_at", now)
	rec.Set("updated_at", now)
	return nil
}

func (s *CustomStore) Update(ctx context.Context, rec def.MutableRecord) error {
	data, err := s.marshalData(rec)
	if err != nil {
		return fmt.Errorf("CustomStore.Update marshal %s: %w", s.def.Name, err)
	}

	now := time.Now().UTC()
	const q = `
		UPDATE custom_entity_records
		SET    data = $1, updated_at = $2
		WHERE  id = $3
		  AND  entity_type = $4
		  AND  deleted_at IS NULL`

	tag, err := s.q.Exec(ctx, q, data, now, rec.ID(), s.def.Name)
	if err != nil {
		return fmt.Errorf("CustomStore.Update %s: %w", s.def.Name, mapPgError(err))
	}
	if tag.RowsAffected() == 0 {
		return persistence.ErrNotFound
	}
	rec.Set("updated_at", now)
	return nil
}

func (s *CustomStore) Delete(ctx context.Context, id uuid.UUID) error {
	// Custom entities always use soft-delete via deleted_at.
	const q = `
		UPDATE custom_entity_records
		SET    deleted_at = NOW()
		WHERE  id = $1
		  AND  entity_type = $2
		  AND  deleted_at IS NULL`

	tag, err := s.q.Exec(ctx, q, id, s.def.Name)
	if err != nil {
		return fmt.Errorf("CustomStore.Delete %s: %w", s.def.Name, mapPgError(err))
	}
	if tag.RowsAffected() == 0 {
		return persistence.ErrNotFound
	}
	return nil
}

func (s *CustomStore) Exists(ctx context.Context, filter map[string]any) (bool, error) {
	where, args := buildJSONBFilter(filter, s.def.Name)
	q := fmt.Sprintf(
		"SELECT EXISTS(SELECT 1 FROM %s WHERE %s AND deleted_at IS NULL)",
		customEntityRecordsTable, where,
	)
	var exists bool
	if err := s.q.QueryRow(ctx, q, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("CustomStore.Exists %s: %w", s.def.Name, err)
	}
	return exists, nil
}

func (s *CustomStore) Count(ctx context.Context, filter map[string]any) (int64, error) {
	where, args := buildJSONBFilter(filter, s.def.Name)
	q := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s AND deleted_at IS NULL",
		customEntityRecordsTable, where,
	)
	var count int64
	if err := s.q.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("CustomStore.Count %s: %w", s.def.Name, err)
	}
	return count, nil
}

func (s *CustomStore) BulkCreate(ctx context.Context, recs []def.MutableRecord) error {
	for _, rec := range recs {
		if err := s.Create(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

func (s *CustomStore) BulkUpdate(ctx context.Context, pred *filterPkg.Filter, values map[string]any) (int64, error) {
	// Merge values into the JSONB data column using the || (concat) operator.
	patch, err := json.Marshal(values)
	if err != nil {
		return 0, fmt.Errorf("CustomStore.BulkUpdate marshal patch: %w", err)
	}

	// Anchor on entity_type and soft-delete. $1 = entity name.
	baseWhere := "entity_type = $1 AND deleted_at IS NULL"
	args := []any{s.def.Name}

	// Translate the caller's filter predicate. Callers should use
	// filter.JSONPath("data", "field_name", filter.Eq("", value)) for JSONB
	// fields and filter.Eq("id", id) for system columns.
	if !pred.IsNone() {
		if clause, predArgs, _ := sqlbuilder.ToSQL(pred, len(args)+1); clause != "" {
			baseWhere += " AND " + clause
			args = append(args, predArgs...)
		}
	}

	patchIdx := len(args) + 1
	args = append(args, string(patch))

	q := fmt.Sprintf(
		`UPDATE %s SET data = data || $%d::jsonb, updated_at = NOW() WHERE %s`,
		customEntityRecordsTable, patchIdx, baseWhere,
	)
	tag, err := s.q.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("CustomStore.BulkUpdate %s: %w", s.def.Name, mapPgError(err))
	}
	return tag.RowsAffected(), nil
}

// ── Internal helpers ───────────────────────────────────────────────────────────

// marshalData serialises non-sensitive fields from rec into a JSON byte slice.
func (s *CustomStore) marshalData(rec def.Record) ([]byte, error) {
	m := make(map[string]any, len(s.def.Fields))
	for _, f := range s.def.Fields {
		if f.IsSensitive {
			continue
		}
		m[f.Name] = rec.Get(f.Name)
	}
	return json.Marshal(m)
}

// scanRow scans one row returned by QueryRow into a customRecord.
func (s *CustomStore) scanRow(row pgx.Row) (def.Record, error) {
	var (
		id        uuid.UUID
		tenantID  uuid.UUID
		rawData   []byte
		createdAt time.Time
		updatedAt time.Time
	)
	if err := row.Scan(&id, &tenantID, &rawData, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNotFound
		}
		return nil, err
	}
	return unmarshalCustomRecord(s.def.Name, id, tenantID, rawData, createdAt, updatedAt)
}

// scanRows scans one row returned by Query into a customRecord.
func (s *CustomStore) scanRows(rows pgx.Rows) (def.Record, error) {
	var (
		id        uuid.UUID
		tenantID  uuid.UUID
		rawData   []byte
		createdAt time.Time
		updatedAt time.Time
	)
	if err := rows.Scan(&id, &tenantID, &rawData, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	return unmarshalCustomRecord(s.def.Name, id, tenantID, rawData, createdAt, updatedAt)
}

// unmarshalCustomRecord decodes JSONB data into a customRecord.
func unmarshalCustomRecord(
	entityName string,
	id, tenantID uuid.UUID,
	rawData []byte,
	createdAt, updatedAt time.Time,
) (*customRecord, error) {
	var data map[string]any
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf("unmarshal custom record %s/%s: %w", entityName, id, err)
	}
	if data == nil {
		data = make(map[string]any)
	}
	data["created_at"] = createdAt
	data["updated_at"] = updatedAt
	return &customRecord{
		id:         id,
		tenantID:   tenantID,
		entityName: entityName,
		data:       data,
	}, nil
}

// buildJSONBFilter constructs a WHERE clause for JSONB equality filters.
// Simple string/numeric values use data->>'key' = $N.
// entity_type is always included as the first predicate.
func buildJSONBFilter(filter map[string]any, entityType string) (string, []any) {
	args := []any{entityType}
	parts := []string{"entity_type = $1"}
	idx := 2
	for k, v := range filter {
		switch k {
		case "id", "tenant_id", "created_at", "updated_at":
			// System columns — use directly.
			parts = append(parts, fmt.Sprintf("%s = $%d", k, idx))
		default:
			// JSONB data field.
			parts = append(parts, fmt.Sprintf("data->>'%s' = $%d",
				strings.ReplaceAll(k, "'", "''"), idx))
		}
		args = append(args, v)
		idx++
	}
	return strings.Join(parts, " AND "), args
}

func escapeLikeCustom(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "%", `\%`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
}

// ── customRecord ───────────────────────────────────────────────────────────────

// customRecord implements def.Record and def.MutableRecord for JSONB-backed entities.
type customRecord struct {
	id         uuid.UUID
	tenantID   uuid.UUID
	entityName string
	data       map[string]any
}

func (r *customRecord) ID() uuid.UUID       { return r.id }
func (r *customRecord) TenantID() uuid.UUID { return r.tenantID }
func (r *customRecord) EntityName() string  { return r.entityName }
func (r *customRecord) Get(field string) any {
	switch field {
	case "id":
		return r.id
	case "tenant_id":
		return r.tenantID
	}
	return r.data[field]
}

func (r *customRecord) Set(field string, val any) {
	switch field {
	case "_id":
		if id, ok := val.(uuid.UUID); ok {
			r.id = id
		}
		return
	case "id":
		if id, ok := val.(uuid.UUID); ok {
			r.id = id
		}
	case "tenant_id":
		if id, ok := val.(uuid.UUID); ok {
			r.tenantID = id
		}
	default:
		r.data[field] = val
	}
}

// compile-time checks
var (
	_ def.Record              = (*customRecord)(nil)
	_ def.MutableRecord       = (*customRecord)(nil)
	_ persistence.EntityStore = (*CustomStore)(nil)
)
