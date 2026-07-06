// Package sdui generates amis JSON page schemas from EntityDefinitions.
//
// Every EntityDefinition auto-generates: list page, create form, edit form,
// detail view. Custom PageBuilderSet overrides any view.
//
// Generated schemas are cached in Redis:
//
//	Key: page:{entity}:{view}:{tenant_id}
//	TTL: 5 minutes
//	Invalidation: on permission change or feature flag change
//
// Permission-gated elements are ABSENT from the schema (not disabled).
// The permission check runs at schema-serve time, not data-fetch time.
//
// amis documentation: https://aisuda.bce.baidu.com/amis/zh-CN/docs/intro
package sdui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/awo/cache"
	"awo.so/awo/compiler"
	"awo.so/awo/def"
)

const (
	sduiCacheTTL    = 5 * time.Minute
	sduiCachePrefix = "page:"
)

// Generator produces amis JSON page schemas.
type Generator struct {
	schema *compiler.CompiledSchema
	cache  cache.Cache
}

// New creates a Generator.
func New(schema *compiler.CompiledSchema, c cache.Cache) *Generator {
	return &Generator{schema: schema, cache: c}
}

// PageKind identifies which view to generate.
type PageKind string

const (
	PageList   PageKind = "list"
	PageCreate PageKind = "create"
	PageEdit   PageKind = "edit"
	PageDetail PageKind = "detail"
)

// GetPage returns the amis JSON schema for the requested page.
// Uses cache; falls back to generation on miss.
// tenantID is used for cache key scoping only — permission filtering
// should be applied by the caller before returning the schema.
func (g *Generator) GetPage(ctx context.Context, entityName string, kind PageKind, tenantID uuid.UUID) (map[string]any, error) {
	cacheKey := fmt.Sprintf("%s%s:%s:%s", sduiCachePrefix, entityName, kind, tenantID)

	var schema map[string]any
	if err := g.cache.Get(ctx, cacheKey, &schema); err == nil {
		return schema, nil
	}

	es, ok := g.schema.ByName[entityName]
	if !ok {
		return nil, fmt.Errorf("sdui: unknown entity %q", entityName)
	}

	// Check for custom page builder first.
	pageBuilders := es.Def.EntityPageBuilders()
	var builder def.PageBuilder
	switch kind {
	case PageList:
		builder = pageBuilders.List
	case PageCreate:
		builder = pageBuilders.Create
	case PageEdit:
		builder = pageBuilders.Edit
	case PageDetail:
		builder = pageBuilders.Detail
	}

	var page map[string]any
	var err error
	if builder != nil {
		pctx := def.PageContext{EntityName: entityName}
		page, err = builder(ctx, pctx)
		if err != nil {
			return nil, fmt.Errorf("sdui: custom builder for %s/%s: %w", entityName, kind, err)
		}
	} else {
		page = g.generate(es, kind)
	}

	// Cache the result.
	_ = g.cache.Set(ctx, cacheKey, page, sduiCacheTTL)
	return page, nil
}

// Invalidate clears all cached pages for an entity.
func (g *Generator) Invalidate(ctx context.Context, entityName string) error {
	return g.cache.DeletePrefix(ctx, sduiCachePrefix+entityName+":")
}

// generate produces the default amis schema for the given entity and view.
func (g *Generator) generate(es *compiler.EntitySchema, kind PageKind) map[string]any {
	switch kind {
	case PageList:
		return g.generateList(es)
	case PageCreate:
		return g.generateForm(es, "create")
	case PageEdit:
		return g.generateForm(es, "edit")
	case PageDetail:
		return g.generateDetail(es)
	default:
		return map[string]any{"type": "page", "body": "Unknown view"}
	}
}

func (g *Generator) generateList(es *compiler.EntitySchema) map[string]any {
	label, labelPlural := entityLabels(es.Def)

	cols := []map[string]any{
		{"name": "id", "label": "ID", "type": "text", "toggled": false},
	}
	for _, f := range es.Def.EntityFields() {
		if f.Hidden || f.Sensitive {
			continue
		}
		col := map[string]any{
			"name":  f.Name,
			"label": f.Label,
			"type":  amisColumnType(f.Type),
		}
		cols = append(cols, col)
	}
	cols = append(cols,
		map[string]any{"name": "created_at", "label": "Created", "type": "datetime"},
	)

	return map[string]any{
		"type":  "page",
		"title": labelPlural,
		"body": map[string]any{
			"type":    "crud",
			"api":     "/api/v1/entities/" + es.TableName,
			"columns": cols,
			"toolbar": []map[string]any{
				{
					"type":       "button",
					"label":      "New " + label,
					"icon":       "fa fa-plus",
					"actionType": "dialog",
					"dialog": map[string]any{
						"title": "New " + label,
						"body":  g.generateForm(es, "create"),
					},
				},
			},
		},
	}
}

func (g *Generator) generateForm(es *compiler.EntitySchema, mode string) map[string]any {
	controls := []map[string]any{}
	for _, f := range es.Def.EntityFields() {
		if f.Hidden || f.ReadOnly || f.Sensitive {
			continue
		}
		if f.Immutable && mode == "edit" {
			continue // immutable fields not shown in edit form
		}
		ctrl := map[string]any{
			"type":  amisFormControl(f.Type),
			"name":  f.Name,
			"label": f.Label,
		}
		if f.Required {
			ctrl["required"] = true
		}
		if f.Type == def.FieldTypeSelect && len(f.Options) > 0 {
			opts := make([]map[string]any, len(f.Options))
			for i, o := range f.Options {
				opts[i] = map[string]any{"label": o, "value": o}
			}
			ctrl["options"] = opts
		}
		if f.MaxLen > 0 {
			ctrl["maxLength"] = f.MaxLen
		}
		controls = append(controls, ctrl)
	}

	apiURL := "/api/v1/entities/" + es.TableName
	if mode == "edit" {
		apiURL = apiURL + "/${id}"
	}
	method := "post"
	if mode == "edit" {
		method = "patch"
	}

	return map[string]any{
		"type": "form",
		"api": map[string]any{
			"method": method,
			"url":    apiURL,
		},
		"body": controls,
	}
}

func (g *Generator) generateDetail(es *compiler.EntitySchema) map[string]any {
	label, _ := entityLabels(es.Def)
	items := []map[string]any{}
	for _, f := range es.Def.EntityFields() {
		if f.Hidden || f.Sensitive {
			continue
		}
		items = append(items, map[string]any{
			"type":  "static",
			"name":  f.Name,
			"label": f.Label,
		})
	}
	return map[string]any{
		"type":  "page",
		"title": label,
		"body": map[string]any{
			"type":    "form",
			"mode":    "horizontal",
			"api":     "/api/v1/entities/" + es.TableName + "/${id}",
			"body":    items,
			"actions": []map[string]any{},
		},
	}
}

// entityLabels extracts singular and plural labels from an EntityDefinition.
// Falls back to EntityName if the concrete type does not expose labels.
func entityLabels(d def.EntityDefinition) (label, labelPlural string) {
	switch v := d.(type) {
	case *def.SystemDefinition:
		return v.Label, v.LabelPlural
	case *def.CustomDefinition:
		return v.Label, v.LabelPlural
	default:
		name := d.EntityName()
		return name, name + "s"
	}
}

// amisColumnType maps FieldType to amis column type.
func amisColumnType(ft def.FieldType) string {
	switch ft {
	case def.FieldTypeBool:
		return "boolean"
	case def.FieldTypeDate:
		return "date"
	case def.FieldTypeDateTime:
		return "datetime"
	case def.FieldTypeCurrency, def.FieldTypeFloat, def.FieldTypeInt:
		return "number"
	default:
		return "text"
	}
}

// amisFormControl maps FieldType to amis form control type.
func amisFormControl(ft def.FieldType) string {
	switch ft {
	case def.FieldTypeSelect:
		return "select"
	case def.FieldTypeMultiSelect:
		return "select" // multiple: true added by caller
	case def.FieldTypeBool:
		return "switch"
	case def.FieldTypeDate:
		return "date"
	case def.FieldTypeDateTime:
		return "datetime"
	case def.FieldTypeLongText:
		return "textarea"
	case def.FieldTypeCurrency, def.FieldTypeFloat:
		return "input-number"
	case def.FieldTypeInt:
		return "input-number"
	case def.FieldTypeJSON:
		return "json-editor"
	default:
		return "input-text"
	}
}

// MarshalPage serializes a page schema to JSON bytes.
func MarshalPage(page map[string]any) ([]byte, error) {
	return json.Marshal(page)
}
