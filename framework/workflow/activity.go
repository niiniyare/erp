package workflow

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"awo.so/framework/def"
	"awo.so/framework/persistence"
)

// StoreActivity wraps a persistence.TenantStore so Temporal activities can
// perform entity reads and writes without importing app-internal packages.
//
// Register with worker.RegisterActivity(storeActivity.FindByID) etc.
type StoreActivity struct {
	store persistence.TenantStore
}

// NewStoreActivity creates a StoreActivity backed by store.
func NewStoreActivity(store persistence.TenantStore) *StoreActivity {
	return &StoreActivity{store: store}
}

// FindByIDInput is the input for the FindByID activity.
type FindByIDInput struct {
	TenantID string `json:"tenant_id"`
	Entity   string `json:"entity"`
	ID       string `json:"id"`
}

// FindByIDOutput wraps the record fields as a plain map for Temporal serialisation.
type FindByIDOutput struct {
	Fields map[string]any `json:"fields"`
}

// FindByID is a Temporal activity that fetches an entity record by primary key.
func (a *StoreActivity) FindByID(ctx context.Context, input FindByIDInput) (*FindByIDOutput, error) {
	logger := activity.GetLogger(ctx)

	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("FindByID: invalid tenant_id: %w", err)
	}
	recID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, fmt.Errorf("FindByID: invalid id: %w", err)
	}

	es, err := a.store.ForEntity(ctx, tenantID, input.Entity)
	if err != nil {
		return nil, err
	}

	rec, err := es.FindByID(ctx, recID)
	if err != nil {
		return nil, err
	}

	logger.Info("FindByID", "entity", input.Entity, "id", input.ID)

	def := def.Lookup(input.Entity)
	fields := map[string]any{"id": rec.ID().String()}
	if def != nil {
		for _, f := range def.Fields {
			fields[f.Name] = rec.Get(f.Name)
		}
	}
	return &FindByIDOutput{Fields: fields}, nil
}

// UpdateInput is the input for the Update activity.
type UpdateInput struct {
	TenantID string         `json:"tenant_id"`
	Entity   string         `json:"entity"`
	ID       string         `json:"id"`
	Fields   map[string]any `json:"fields"`
}

// Update is a Temporal activity that applies field updates to an entity record.
// Does NOT run hooks — intended for system/workflow-initiated state transitions.
func (a *StoreActivity) Update(ctx context.Context, input UpdateInput) error {
	tenantID, err := uuid.Parse(input.TenantID)
	if err != nil {
		return fmt.Errorf("Update: invalid tenant_id: %w", err)
	}
	recID, err := uuid.Parse(input.ID)
	if err != nil {
		return fmt.Errorf("Update: invalid id: %w", err)
	}

	if err := a.store.WithTx(ctx, tenantID, func(tx persistence.TenantTx) error {
		es := tx.ForEntity(input.Entity)

		existing, err := es.FindByID(ctx, recID)
		if err != nil {
			return err
		}

		mut := &mutableWrapper{Record: existing, changes: input.Fields}
		return es.Update(ctx, mut)
	}); err != nil {
		return err
	}

	activity.GetLogger(ctx).Info("Update", "entity", input.Entity, "id", input.ID)
	return nil
}

// ──────────────────────────────────────────────────────────────────
// mutableWrapper adapts a read-only Record + change map → MutableRecord
// ──────────────────────────────────────────────────────────────────

type mutableWrapper struct {
	def.Record
	changes map[string]any
}

func (w *mutableWrapper) Get(field string) any {
	if v, ok := w.changes[field]; ok {
		return v
	}
	return w.Record.Get(field)
}

func (w *mutableWrapper) Set(field string, value any) {
	if w.changes == nil {
		w.changes = make(map[string]any)
	}
	w.changes[field] = value
}

var _ def.MutableRecord = (*mutableWrapper)(nil)
