package sdui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"awo.so/framework/def"
	"awo.so/framework/sdui/amis"
)

// ViewType identifies which page variant to build.
type ViewType string

const (
	ViewCRUD       ViewType = "crud"
	ViewFormCreate ViewType = "form_create"
	ViewFormEdit   ViewType = "form_edit"
	ViewDetail     ViewType = "detail"
)

// Builder generates AMIS page schemas for one EntityDefinition.
// It applies permission-based field filtering, honours PageBuilderSet overrides,
// and reads/writes the schema cache transparently.
type Builder struct {
	entDef       *def.EntityDefinition
	apiBase      string
	cache        Cache
	extraFields  []*def.FieldDef
}

// NewBuilder creates a Builder for entDef.
func NewBuilder(entDef *def.EntityDefinition, apiBase string, opts ...BuilderOption) *Builder {
	b := &Builder{entDef: entDef, apiBase: apiBase}
	for _, o := range opts {
		o(b)
	}
	return b
}

// BuilderOption configures a Builder.
type BuilderOption func(*Builder)

// WithBuilderCache injects a schema cache.
func WithBuilderCache(c Cache) BuilderOption {
	return func(b *Builder) { b.cache = c }
}

// WithExtraFields injects tenant-specific custom fields to append.
func WithExtraFields(fields []*def.FieldDef) BuilderOption {
	return func(b *Builder) { b.extraFields = fields }
}

// Build generates or retrieves from cache the AMIS schema for viewType.
//
// Cache key: page:{entityHash}:{viewType}:{tenantID}
// Cache is bypassed for PageBuilderSet overrides (may carry per-request state).
//
// viewer may be nil for unauthenticated routes — sensitive fields are still
// excluded, and the schema is marked read-only.
func (b *Builder) Build(ctx context.Context, viewType ViewType, viewer def.ViewerContext, tenantID uuid.UUID) (map[string]any, error) {
	// Custom page builder override — bypasses cache.
	if schema := b.customSchema(viewType, viewer); schema != nil {
		return schema, nil
	}

	entityHash := entityNameHash(b.entDef.Name)
	cacheKey := buildKey(entityHash, string(viewType), tenantID.String())

	// Cache read (viewer-agnostic — filtering happens after cache retrieval).
	if b.cache != nil {
		if raw, err := CacheGet(ctx, b.cache, cacheKey); err == nil && raw != nil {
			var schema map[string]any
			if err := json.Unmarshal(raw, &schema); err == nil {
				return applyViewerFilter(schema, b.entDef, viewer, viewType), nil
			}
			// Corrupt cache entry — fall through to regenerate.
		}
	}

	// Generate base schema (no viewer filtering — cache is viewer-agnostic).
	schema := b.generateBase(viewType)

	// Cache write (best-effort).
	if b.cache != nil {
		if raw, err := json.Marshal(schema); err == nil {
			_ = CacheSet(ctx, b.cache, cacheKey, raw)
		}
	}

	// Apply viewer-specific filtering after caching.
	return applyViewerFilter(schema, b.entDef, viewer, viewType), nil
}

// Invalidate removes all cached schemas for this entity and tenant.
func (b *Builder) Invalidate(ctx context.Context, tenantID uuid.UUID) error {
	if b.cache == nil {
		return nil
	}
	return InvalidateEntitySchemas(ctx, b.cache, entityNameHash(b.entDef.Name), tenantID.String())
}

// ── Internal ──────────────────────────────────────────────────────────────────

// customSchema returns the schema from PageBuilderSet if a matching override exists.
// Returns nil when no override is configured for viewType.
func (b *Builder) customSchema(viewType ViewType, _ def.ViewerContext) map[string]any {
	pb := b.entDef.PageBuilders
	if pb == nil {
		return nil
	}
	switch viewType {
	case ViewCRUD:
		if pb.CRUDPage != nil {
			return pb.CRUDPage(b.entDef, b.apiBase)
		}
	case ViewFormCreate, ViewFormEdit:
		if pb.FormPage != nil {
			return pb.FormPage(b.entDef, b.apiBase)
		}
	case ViewDetail:
		if pb.DetailPage != nil {
			return pb.DetailPage(b.entDef, b.apiBase)
		}
	}
	return nil
}

// generateBase builds the viewer-agnostic base schema for caching.
func (b *Builder) generateBase(viewType ViewType) map[string]any {
	opts := amis.PageOpts{
		APIBase:     b.apiBase,
		ExtraFields: b.extraFields,
	}
	switch viewType {
	case ViewCRUD:
		return amis.CRUDPage(b.entDef, opts)
	case ViewFormCreate:
		return amis.FormPage(b.entDef, opts)
	case ViewFormEdit:
		// Edit form is identical to create form structurally;
		// the AMIS api.sendOn and initApi fields differ (handled by client URL params).
		editOpts := opts
		return amis.FormPage(b.entDef, editOpts)
	case ViewDetail:
		if b.entDef.PageBuilders != nil && b.entDef.PageBuilders.DetailPage != nil {
			return b.entDef.PageBuilders.DetailPage(b.entDef, b.apiBase)
		}
		// Fall back to read-only form when no detail builder is registered.
		roOpts := opts
		roOpts.ReadOnly = true
		return amis.FormPage(b.entDef, roOpts)
	default:
		return amis.CRUDPage(b.entDef, opts)
	}
}

// applyViewerFilter removes sensitive fields and marks read-only when the viewer
// lacks write permission. Operates on a shallow copy of the base schema.
//
// This runs AFTER cache retrieval so the cached schema always contains all fields;
// the filtering result is never stored in cache (it varies per viewer).
func applyViewerFilter(schema map[string]any, entDef *def.EntityDefinition, viewer def.ViewerContext, viewType ViewType) map[string]any {
	excluded := FilteredOpts(entDef, viewer)
	if len(excluded) == 0 {
		return schema
	}

	// Walk the schema and remove excluded field names from columns/form controls.
	// AMIS stores field refs as "name" keys inside body.columns and body.controls arrays.
	filtered := shallowCopySchema(schema)
	if body, ok := filtered["body"].(map[string]any); ok {
		filtered["body"] = filterBody(body, excluded)
	}
	return filtered
}

// filterBody removes excluded fields from AMIS body maps (crud/form bodies).
func filterBody(body map[string]any, excluded map[string]bool) map[string]any {
	out := make(map[string]any, len(body))
	for k, v := range body {
		switch k {
		case "columns", "controls", "body":
			out[k] = filterList(v, excluded)
		default:
			out[k] = v
		}
	}
	return out
}

// filterList removes items from an AMIS columns/controls array whose "name" key
// is in excluded. Non-map items are passed through unchanged.
func filterList(v any, excluded map[string]bool) any {
	items, ok := v.([]any)
	if !ok {
		return v
	}
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			filtered = append(filtered, item)
			continue
		}
		name, _ := m["name"].(string)
		if name == "" || !excluded[name] {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// shallowCopySchema returns a one-level-deep copy of the schema map so we
// don't mutate the cached value when applying per-viewer filtering.
func shallowCopySchema(schema map[string]any) map[string]any {
	out := make(map[string]any, len(schema))
	for k, v := range schema {
		out[k] = v
	}
	return out
}

// entityNameHash returns the first 8 bytes of sha256(name) as a 16-char hex string.
func entityNameHash(name string) string {
	h := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%x", h[:8])
}
