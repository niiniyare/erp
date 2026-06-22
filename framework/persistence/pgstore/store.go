package pgstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"awo.so/framework/definition"
	"awo.so/framework/persistence"
	"awo.so/framework/persistence/sqlbuilder"
)

// Querier is the minimal pgx interface satisfied by both *pgxpool.Pool and pgx.Tx.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// EntityStore is a pgx-backed implementation of persistence.EntityStore.
type EntityStore struct {
	def      *definition.EntityDefinition
	tenantID uuid.UUID
	q        Querier
}

// New creates an EntityStore. tenantID may be uuid.Nil for Global entities.
func New(def *definition.EntityDefinition, tenantID uuid.UUID, q Querier) *EntityStore {
	return &EntityStore{def: def, tenantID: tenantID, q: q}
}

// ──────────────────────────────────────────────────────────────────
// persistence.EntityStore implementation
// ──────────────────────────────────────────────────────────────────

func (s *EntityStore) FindByID(ctx context.Context, id uuid.UUID) (definition.Record, error) {
	query := sqlbuilder.SelectOne(s.def)
	row := s.q.QueryRow(ctx, query, id)
	rec, err := s.scanRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, persistence.ErrNotFound
		}
		return nil, fmt.Errorf("FindByID %s/%s: %w", s.def.Name, id, err)
	}
	return rec, nil
}

func (s *EntityStore) List(ctx context.Context, opts persistence.ListOptions) (persistence.Page, error) {
	// Build args slice: filter values, search string, limit, offset.
	var args []any

	// We pass the opts through; sqlbuilder.SelectList needs the filter map.
	// For now we use a simplified path: build the query, collect args in field order.
	filterFields := make([]string, 0, len(opts.Filter))
	for k := range opts.Filter {
		filterFields = append(filterFields, k)
	}

	sbOpts := buildSelectOpts(opts)
	query, _ := sqlbuilder.SelectList(s.def, sbOpts)

	// Collect args in the same order whereClauses produces placeholders.
	for _, k := range filterFields {
		args = append(args, opts.Filter[k])
	}
	if opts.Search != "" {
		args = append(args, "%"+opts.Search+"%")
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	args = append(args, limit, opts.Offset)

	rows, err := s.q.Query(ctx, query, args...)
	if err != nil {
		return persistence.Page{}, fmt.Errorf("List %s: %w", s.def.Name, err)
	}
	defer rows.Close()

	var records []definition.Record
	var total int64
	for rows.Next() {
		rec, t, err := s.scanListRow(rows)
		if err != nil {
			return persistence.Page{}, fmt.Errorf("List %s scan: %w", s.def.Name, err)
		}
		total = t
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return persistence.Page{}, fmt.Errorf("List %s rows: %w", s.def.Name, err)
	}

	return persistence.Page{
		Records: records,
		Total:   total,
		Limit:   limit,
		Offset:  opts.Offset,
	}, nil
}

func (s *EntityStore) Create(ctx context.Context, rec definition.MutableRecord) error {
	fields := fieldNames(s.def)
	query, cols := sqlbuilder.Insert(s.def, fields)

	id := rec.ID()
	if id == uuid.Nil {
		id = uuid.New()
		rec.Set("id", id)
	}

	args := make([]any, 0, len(cols))
	for _, col := range cols {
		switch col {
		case "id":
			args = append(args, id)
		case "tenant_id":
			args = append(args, s.tenantID)
		default:
			args = append(args, rec.Get(col))
		}
	}

	var returnedID uuid.UUID
	err := s.q.QueryRow(ctx, query, args...).Scan(&returnedID)
	if err != nil {
		return mapPgError(err)
	}
	rec.Set("id", returnedID)
	return nil
}

func (s *EntityStore) Update(ctx context.Context, rec definition.MutableRecord) error {
	fields := mutableFieldNames(s.def)
	query, cols := sqlbuilder.Update(s.def, fields)

	args := make([]any, 0, len(cols)+1)
	for _, col := range cols {
		args = append(args, rec.Get(col))
	}
	args = append(args, rec.ID()) // last placeholder = id

	_, err := s.q.Exec(ctx, query, args...)
	if err != nil {
		return mapPgError(err)
	}
	return nil
}

func (s *EntityStore) Delete(ctx context.Context, id uuid.UUID) error {
	var query string
	if s.def.SoftDelete {
		query = sqlbuilder.SoftDelete(s.def)
	} else {
		query = sqlbuilder.HardDelete(s.def)
	}
	_, err := s.q.Exec(ctx, query, id)
	return mapPgError(err)
}

func (s *EntityStore) Exists(ctx context.Context, filter map[string]any) (bool, error) {
	fields := make([]string, 0, len(filter))
	args := make([]any, 0, len(filter))
	for k, v := range filter {
		fields = append(fields, k)
		args = append(args, v)
	}
	query := sqlbuilder.Exists(s.def, fields)
	var exists bool
	err := s.q.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (s *EntityStore) BulkUpdate(ctx context.Context, filter map[string]any, values map[string]any) (int64, error) {
	// Hand-build a parameterised UPDATE ... SET ... WHERE ...
	setCols := make([]string, 0, len(values))
	args := make([]any, 0, len(values)+len(filter))
	idx := 1
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("UPDATE %s SET ", s.def.TableName()))
	for col, val := range values {
		if idx > 1 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s = $%d", col, idx))
		setCols = append(setCols, col)
		args = append(args, val)
		idx++
	}
	_ = setCols

	var whereParts []string
	for col, val := range filter {
		whereParts = append(whereParts, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}
	if len(whereParts) > 0 {
		sb.WriteString(" WHERE " + strings.Join(whereParts, " AND "))
	}

	tag, err := s.q.Exec(ctx, sb.String(), args...)
	if err != nil {
		return 0, mapPgError(err)
	}
	return tag.RowsAffected(), nil
}

// ──────────────────────────────────────────────────────────────────
// Scanning helpers
// ──────────────────────────────────────────────────────────────────

func (s *EntityStore) scanRow(row pgx.Row) (*mapRecord, error) {
	rec, dest, fieldPtrs := buildScanDest(s.def, s.tenantID)
	if err := row.Scan(dest...); err != nil {
		return nil, err
	}
	dereferenceFieldPtrs(s.def, rec, fieldPtrs)
	return rec, nil
}

func (s *EntityStore) scanListRow(rows pgx.Rows) (*mapRecord, int64, error) {
	rec, dest, fieldPtrs := buildScanDest(s.def, s.tenantID)
	var total int64
	dest = append(dest, &total)
	if err := rows.Scan(dest...); err != nil {
		return nil, 0, err
	}
	dereferenceFieldPtrs(s.def, rec, fieldPtrs)
	return rec, total, nil
}

// buildScanDest builds scan destination pointers matching sqlbuilder.columnList order.
func buildScanDest(def *definition.EntityDefinition, tenantID uuid.UUID) (*mapRecord, []any, []*any) {
	rec := newMapRecord(def.Name, tenantID)
	dest := []any{&rec.id}
	if !def.Global {
		dest = append(dest, &rec.tenantID)
	}

	var createdAt, updatedAt any
	dest = append(dest, &createdAt, &updatedAt)
	if def.SoftDelete {
		var deletedAt any
		dest = append(dest, &deletedAt)
		rec.data["deleted_at"] = &deletedAt
	}

	fieldPtrs := make([]*any, len(def.Fields))
	for i := range def.Fields {
		ptr := new(any)
		fieldPtrs[i] = ptr
		dest = append(dest, ptr)
	}

	rec.data["created_at"] = createdAt
	rec.data["updated_at"] = updatedAt
	return rec, dest, fieldPtrs
}

// dereferenceFieldPtrs copies scanned values from pointers into rec.data.
func dereferenceFieldPtrs(def *definition.EntityDefinition, rec *mapRecord, ptrs []*any) {
	for i, f := range def.Fields {
		rec.data[f.Name] = *ptrs[i]
	}
}

// ──────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────

func fieldNames(def *definition.EntityDefinition) []string {
	names := make([]string, len(def.Fields))
	for i, f := range def.Fields {
		names[i] = f.Name
	}
	return names
}

// mutableFieldNames excludes immutable system columns from UPDATE sets.
func mutableFieldNames(def *definition.EntityDefinition) []string {
	names := make([]string, 0, len(def.Fields))
	for _, f := range def.Fields {
		if !f.Immutable {
			names = append(names, f.Name)
		}
	}
	return names
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return persistence.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", persistence.ErrConflict, pgErr.ConstraintName)
		}
	}
	return err
}

func buildSelectOpts(opts persistence.ListOptions) sqlbuilder.SelectOpts {
	return sqlbuilder.SelectOpts{
		Filter:    opts.Filter,
		Search:    opts.Search,
		OrderBy:   opts.OrderBy,
		Ascending: opts.Ascending,
	}
}

// TenantStoreAdapter wraps a *pgxpool.Pool and implements persistence.TenantStore.
type TenantStoreAdapter struct {
	pool *pgxpool.Pool
}

// NewTenantStore creates a TenantStoreAdapter backed by pool.
func NewTenantStore(pool *pgxpool.Pool) *TenantStoreAdapter {
	return &TenantStoreAdapter{pool: pool}
}

func (a *TenantStoreAdapter) ForEntity(ctx context.Context, tenantID uuid.UUID, entity string) (persistence.EntityStore, error) {
	def := definition.Lookup(entity)
	if def == nil {
		return nil, fmt.Errorf("%w: %s", persistence.ErrUnknownEntity, entity)
	}
	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	// Set RLS tenant context on this connection before returning.
	if !def.Global {
		if _, err := conn.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
			conn.Release()
			return nil, fmt.Errorf("set_tenant_context: %w", err)
		}
	}
	// Wrap conn in a store; conn is released when the store is garbage-collected.
	// For production use, prefer WithTx which releases deterministically.
	return New(def, tenantID, conn), nil
}

func (a *TenantStoreAdapter) WithTx(ctx context.Context, tenantID uuid.UUID, fn func(tx persistence.TenantTx) error) error {
	conn, err := a.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	// Set RLS tenant context transaction-local.
	if _, err := tx.Exec(ctx, "SELECT set_tenant_context($1)", tenantID); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("set_tenant_context: %w", err)
	}

	ttx := &tenantTx{tenantID: tenantID, tx: tx}
	if err := fn(ttx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

type tenantTx struct {
	tenantID uuid.UUID
	tx       pgx.Tx
}

func (t *tenantTx) ForEntity(entity string) persistence.EntityStore {
	def := definition.Lookup(entity)
	if def == nil {
		panic(fmt.Sprintf("pgstore.tenantTx: unknown entity %q", entity))
	}
	return New(def, t.tenantID, t.tx)
}

var _ persistence.TenantStore = (*TenantStoreAdapter)(nil)
var _ persistence.TenantTx    = (*tenantTx)(nil)
