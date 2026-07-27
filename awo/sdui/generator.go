// Package sdui is the SDUI generator — the orchestrator that converts an
// EntitySchema and ViewerContext into a WidgetTree (*widget.Node) for a
// given page view, then renders it to amis JSON.
//
// Architecture (ADR-006):
//
//	EntityDefinition
//	    ↓  (compiler)
//	EntitySchema
//	    ↓  (sdui.Generator.GetPage)
//	*widget.Node           ← WidgetTree IR (awo/sdui/widget)
//	    ↓  (amis.Renderer.Render)
//	map[string]any         ← amis JSON
//	    ↓  (json.Marshal + HTTP response)
//	browser
//
// The generator reads EntitySchema (compiler output), never the original
// EntityDefinition. Permission-gated elements are absent from the tree —
// not present with Hidden: true — per WIDGET_TREE_SPEC.md §7.
//
// If the entity declares a PageBuilder for the requested PageKind and the
// builder returns a non-nil map, that map is used directly (bypasses the
// WidgetTree path). Returning nil falls back to auto-generation.
//
// Package dependencies (ADR-006):
//
//	sdui       → sdui/widget, sdui/amis, compiler, def, auth, cache
//	sdui/amis  → sdui/widget  (no compiler, def, or auth)
//	sdui/widget → (none)
//
// Cache key format (AMIS_RENDERER_SPEC.md §8):
//
//	page:{entity_qualified_name}:{view}:{viewer_roles_hash}:{tenant_id}
//	TTL: 5 minutes
package sdui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"awo.so/awo/auth"
	"awo.so/awo/cache"
	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/sdui/amis"
	"awo.so/awo/sdui/widget"
)

const (
	sduiCacheTTL    = 5 * time.Minute
	sduiCachePrefix = "page:"
)

// Generator produces amis JSON page schemas for entity views.
//
// Internally the generator builds a WidgetTree (*widget.Node) and renders it
// through the amis renderer. The rendered schema is cached in Redis per the
// cache key format specified in AMIS_RENDERER_SPEC.md §8.
//
// All methods are goroutine-safe; the Generator holds no mutable state.
type Generator struct {
	schema    *compiler.CompiledSchema
	evaluator auth.PolicyEvaluator // may be nil (all actions permitted)
	cache     cache.Cache
	renderer  amis.Renderer
}

// New returns a Generator backed by schema, evaluator, and c.
//
//   - evaluator may be nil: all actions permitted (tests / admin tools).
//   - c may be nil: caching disabled (schemas generated on every request).
func New(schema *compiler.CompiledSchema, evaluator auth.PolicyEvaluator, c cache.Cache) *Generator {
	if schema == nil {
		panic("sdui.New: schema must not be nil")
	}
	return &Generator{
		schema:    schema,
		evaluator: evaluator,
		cache:     c,
		renderer:  amis.New(),
	}
}

// GetPage returns the amis JSON schema for the given entity view.
//
// entityName is the qualified entity name (e.g. "finance_invoice").
// view is one of the def.PageKind constants.
// viewer is the authenticated principal.
//
// Cache lookup is attempted first (key: AMIS_RENDERER_SPEC.md §8 format).
// On miss the WidgetTree is generated, rendered to amis JSON, and cached.
//
// If the entity's PageBuilder for view returns a non-nil map, that map is
// returned directly without going through the WidgetTree path.
func (g *Generator) GetPage(ctx context.Context, entityName string, view def.PageKind, viewer auth.ViewerContext) (map[string]any, error) {
	es, ok := g.schema.ByName[entityName]
	if !ok {
		return nil, fmt.Errorf("sdui: entity %q not found in compiled schema", entityName)
	}

	// Check for custom PageBuilder — bypasses WidgetTree entirely.
	if builder := pageBuilderFor(es, view); builder != nil {
		pctx := def.PageContext{
			EntityName: entityName,
			Kind:       view,
		}
		if viewer != nil {
			pctx.Actor = viewer.Actor()
		}
		result, err := builder(ctx, pctx)
		if err != nil {
			return nil, fmt.Errorf("sdui: custom builder for %s/%s: %w", entityName, view, err)
		}
		if result != nil {
			return result, nil
		}
		// nil → fall through to auto-generation.
	}

	// Cache key: page:{entity}:{view}:{roles_hash}:{tenant_id}
	cacheKey := g.cacheKey(entityName, view, viewer)
	if g.cache != nil {
		var cached map[string]any
		if err := g.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached, nil
		}
	}

	// Build WidgetTree.
	gen := &pageGen{
		schema:    g.schema,
		evaluator: g.evaluator,
	}
	root, err := gen.GetPage(ctx, entityName, view, viewer)
	if err != nil {
		return nil, err
	}

	// Render WidgetTree → amis JSON.
	schema, err := g.renderer.Render(root)
	if err != nil {
		return nil, fmt.Errorf("sdui: render %s/%s: %w", entityName, view, err)
	}

	// Cache the rendered schema.
	if g.cache != nil {
		_ = g.cache.Set(ctx, cacheKey, schema, sduiCacheTTL)
	}

	return schema, nil
}

// GetWidgetTree builds and returns the raw WidgetTree for the given entity view.
// Use this when the caller needs the IR rather than the rendered amis JSON
// (e.g. for testing or alternative renderers).
func (g *Generator) GetWidgetTree(ctx context.Context, entityName string, view def.PageKind, viewer auth.ViewerContext) (*widget.Node, error) {
	gen := &pageGen{schema: g.schema, evaluator: g.evaluator}
	return gen.GetPage(ctx, entityName, view, viewer)
}

// Invalidate clears all cached pages for an entity across all views, viewers,
// and tenants. Call after permission changes or feature flag changes.
func (g *Generator) Invalidate(ctx context.Context, entityName string) error {
	if g.cache == nil {
		return nil
	}
	return g.cache.DeletePrefix(ctx, sduiCachePrefix+entityName+":")
}

// MarshalPage serializes a page schema to JSON bytes.
func MarshalPage(page map[string]any) ([]byte, error) {
	return json.Marshal(page)
}

// cacheKey builds the canonical cache key for a page request.
// Format: page:{entity}:{view}:{roles_hash}:{tenant_id}
// (AMIS_RENDERER_SPEC.md §8)
func (g *Generator) cacheKey(entityName string, view def.PageKind, viewer auth.ViewerContext) string {
	var rolesHash string
	var tenantID uuid.UUID
	if viewer != nil {
		rolesHash = viewerRolesHash(viewer.Roles())
		tenantID = viewer.TenantID()
	}
	return fmt.Sprintf("%s%s:%s:%s:%s", sduiCachePrefix, entityName, view, rolesHash, tenantID)
}

// viewerRolesHash returns the first 16 hex chars of SHA-256(sorted_roles).
// Per AMIS_RENDERER_SPEC.md §8: ensures per-role-set cache isolation.
func viewerRolesHash(roles []string) string {
	sorted := append([]string(nil), roles...)
	sort.Strings(sorted)
	h := sha256.New()
	for _, r := range sorted {
		h.Write([]byte(r))
		h.Write([]byte{0}) // separator
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// NavEntry is a single navigation item in the sidebar menu.
type NavEntry struct {
	// Module is the owning module (e.g. "finance").
	Module string `json:"module"`
	// Label is the human-readable plural display name (e.g. "Invoices").
	Label string `json:"label"`
	// Entity is the qualified entity name (e.g. "finance_invoice").
	Entity string `json:"entity"`
	// ListURL is the SDUI page URL for the list view.
	ListURL string `json:"listUrl"`
}

// NavModule groups NavEntry values by module.
type NavModule struct {
	Module  string     `json:"module"`
	Label   string     `json:"label"`
	Entries []NavEntry `json:"entries"`
}

// Nav returns the navigation schema for the sidebar menu.
// Entities are grouped by module in the order they appear in CompiledSchema.
// Only entities with a declared Read permission are included.
func (g *Generator) Nav() []NavModule {
	seen := make(map[string]int) // module → index in result
	var result []NavModule

	for _, es := range g.schema.Entities {
		if len(es.Permissions.Read) == 0 {
			continue // inaccessible entity — omit from nav
		}
		idx, ok := seen[es.Module]
		if !ok {
			idx = len(result)
			seen[es.Module] = idx
			// Capitalise first letter for the module display label.
			moduleLabel := es.Module
			if len(moduleLabel) > 0 {
				moduleLabel = strings.ToUpper(moduleLabel[:1]) + moduleLabel[1:]
			}
			result = append(result, NavModule{
				Module: es.Module,
				Label:  moduleLabel,
			})
		}
		result[idx].Entries = append(result[idx].Entries, NavEntry{
			Module:  es.Module,
			Label:   es.LabelPlural,
			Entity:  es.QualifiedName,
			ListURL: "/ui/" + es.Module + "/" + es.APIResource,
		})
	}
	return result
}

// pageBuilderFor returns the PageBuilder for the given view kind, or nil.
func pageBuilderFor(es *compiler.EntitySchema, view def.PageKind) def.PageBuilder {
	switch view {
	case def.PageKindList:
		return es.PageBuilders.List
	case def.PageKindCreate:
		return es.PageBuilders.Create
	case def.PageKindEdit:
		return es.PageBuilders.Edit
	case def.PageKindDetail:
		return es.PageBuilders.Detail
	}
	return nil
}

// ── pageGen: WidgetTree construction ─────────────────────────────────────────

// pageGen builds WidgetTree nodes from EntitySchema + ViewerContext.
// It is the inner layer of the Generator — separated for testability.
type pageGen struct {
	schema    *compiler.CompiledSchema
	evaluator auth.PolicyEvaluator
}

// GetPage generates the WidgetTree root node for the given entity view.
func (g *pageGen) GetPage(ctx context.Context, entityName string, view def.PageKind, viewer auth.ViewerContext) (*widget.Node, error) {
	es, ok := g.schema.ByName[entityName]
	if !ok {
		return nil, fmt.Errorf("sdui: entity %q not found in compiled schema", entityName)
	}
	switch view {
	case def.PageKindList:
		return g.buildList(ctx, es, viewer)
	case def.PageKindCreate:
		return g.buildForm(ctx, es, viewer, false)
	case def.PageKindEdit:
		return g.buildForm(ctx, es, viewer, true)
	case def.PageKindDetail:
		return g.buildDetail(ctx, es, viewer)
	default:
		return nil, fmt.Errorf("sdui: unknown PageKind %q", view)
	}
}

// ── List view ─────────────────────────────────────────────────────────────────

func (g *pageGen) buildList(ctx context.Context, es *compiler.EntitySchema, viewer auth.ViewerContext) (*widget.Node, error) {
	var cols []*widget.Node
	for _, f := range es.Fields {
		if skipListColumn(f) {
			continue
		}
		cols = append(cols, fieldToNode(f, es, false))
	}

	var toolbarActions []*widget.ActionNode
	if canPerform(ctx, g.evaluator, viewer, es.QualifiedName, "create") {
		toolbarActions = append(toolbarActions, &widget.ActionNode{
			Label:      "New " + es.Label,
			ActionType: "link",
			Level:      "primary",
			Href:       "/ui/" + es.Module + "/" + es.APIResource + "/create",
		})
	}

	// Row actions: absent (not hidden) when permission denied.
	var rowActions []any
	if canPerform(ctx, g.evaluator, viewer, es.QualifiedName, "read") {
		rowActions = append(rowActions, map[string]any{
			"type": "button", "label": "View", "actionType": "link", "level": "default",
			"link": "/ui/" + es.Module + "/" + es.APIResource + "/${id}",
		})
	}
	if canPerform(ctx, g.evaluator, viewer, es.QualifiedName, "update") {
		rowActions = append(rowActions, map[string]any{
			"type": "button", "label": "Edit", "actionType": "link", "level": "default",
			"link": "/ui/" + es.Module + "/" + es.APIResource + "/${id}/edit",
		})
	}
	if canPerform(ctx, g.evaluator, viewer, es.QualifiedName, "delete") {
		rowActions = append(rowActions, map[string]any{
			"type": "button", "label": "Delete", "actionType": "ajax", "level": "danger",
			"api":         "DELETE:" + es.RoutePrefix + "/${id}",
			"confirmText": "Delete this " + es.Label + "? This action cannot be undone.",
		})
	}
	for _, a := range es.Actions {
		if a.Hidden || !canPerform(ctx, g.evaluator, viewer, es.QualifiedName, a.Name) {
			continue
		}
		btn := map[string]any{
			"type": "button", "label": a.Label,
			"actionType": "ajax", "level": "default",
			"api": string(a.Method) + ":" + es.RoutePrefix + "/${id}/" + a.Name,
		}
		if a.ConfirmMessage != "" {
			btn["confirmText"] = a.ConfirmMessage
		}
		rowActions = append(rowActions, btn)
	}

	listNode := &widget.Node{
		Kind:  widget.NodeList,
		Label: es.LabelPlural,
		DataSource: &widget.DataSource{
			URL:    es.RoutePrefix + "?q=${keywords}&page=${page}&perPage=${perPage}",
			Method: "GET",
		},
		Children: cols,
		Actions:  toolbarActions,
	}
	if len(rowActions) > 0 {
		if listNode.Props == nil {
			listNode.Props = make(map[string]any)
		}
		listNode.Props["rowActions"] = rowActions
	}

	return &widget.Node{
		Kind:     widget.NodePage,
		Label:    es.LabelPlural,
		Children: []*widget.Node{listNode},
		Actions:  toolbarActions,
	}, nil
}

// skipListColumn returns true for fields excluded from auto-generated list columns.
// (SDUI_FIELD_WIDGET_MAP.md — List View Column Exclusions)
func skipListColumn(f def.FieldDef) bool {
	switch f.Name {
	case "id", "tenant_id", "custom_fields":
		return true
	}
	if f.Type == def.FieldTypeLongText || f.Type == def.FieldTypeJSON {
		return true
	}
	return f.Sensitive || f.Hidden
}

// ── Form views (Create / Edit) ────────────────────────────────────────────────

func (g *pageGen) buildForm(_ context.Context, es *compiler.EntitySchema, _ auth.ViewerContext, isEdit bool) (*widget.Node, error) {
	var children []*widget.Node

	// Use declared layout if present; fall back to flat field list.
	if layoutNodes := buildLayoutNodes(es, isEdit); layoutNodes != nil {
		children = layoutNodes
	} else {
		for _, f := range es.Fields {
			if f.Hidden || f.Name == "tenant_id" {
				continue
			}
			children = append(children, fieldToNode(f, es, isEdit))
		}
	}

	for _, edge := range es.Edges {
		if edge.Hidden || edge.Type != def.EdgeOneToMany {
			continue
		}
		// Generate column nodes for the inline child table by looking up
		// the target EntitySchema. Fall back to an empty table on miss.
		var cols []*widget.Node
		if target, ok := g.schema.ByName[edge.Target]; ok {
			for _, tf := range target.Fields {
				if skipListColumn(tf) {
					continue
				}
				cols = append(cols, fieldToNode(tf, target, false))
			}
		}
		children = append(children, &widget.Node{
			Kind:  widget.NodeTable,
			Label: edge.Label,
			DataSource: &widget.DataSource{
				URL:    "/api/v1/" + edge.Target + "?parent_id=${id}",
				Method: "GET",
			},
			Children: cols,
		})
	}

	var method, apiURL, readURL, label string
	if isEdit {
		method = "PATCH"
		apiURL = es.RoutePrefix + "/${id}"
		readURL = es.RoutePrefix + "/${id}" // initApi: pre-populate form with existing record
		label = "Edit " + es.Label
	} else {
		method = "POST"
		apiURL = es.RoutePrefix
		// readURL is empty for create — no existing record to load.
		label = "Create " + es.Label
	}

	return &widget.Node{
		Kind:  widget.NodePage,
		Label: label,
		Children: []*widget.Node{{
			Kind:  widget.NodeForm,
			Label: label,
			DataSource: &widget.DataSource{
				URL:     apiURL,
				Method:  method,
				ReadURL: readURL,
			},
			Children: children,
		}},
	}, nil
}

// ── Detail view ───────────────────────────────────────────────────────────────

func (g *pageGen) buildDetail(ctx context.Context, es *compiler.EntitySchema, viewer auth.ViewerContext) (*widget.Node, error) {
	var children []*widget.Node

	// Layout for detail view: same as form but all fields forced read-only.
	if layoutNodes := buildLayoutNodes(es, false); layoutNodes != nil {
		// Walk every leaf field node in the layout tree and set ReadOnly.
		setAllReadOnly(layoutNodes)
		children = layoutNodes
	} else {
		for _, f := range es.Fields {
			if f.Hidden || f.Name == "tenant_id" {
				continue
			}
			n := fieldToNode(f, es, false)
			n.ReadOnly = true
			children = append(children, n)
		}
	}

	var actions []*widget.ActionNode
	if canPerform(ctx, g.evaluator, viewer, es.QualifiedName, "update") {
		actions = append(actions, &widget.ActionNode{
			Label: "Edit", ActionType: "link", Level: "primary",
			Href: "/ui/" + es.Module + "/" + es.APIResource + "/${id}/edit",
		})
	}
	for _, a := range es.Actions {
		if a.Hidden || !canPerform(ctx, g.evaluator, viewer, es.QualifiedName, a.Name) {
			continue
		}
		actions = append(actions, &widget.ActionNode{
			Label: a.Label, ActionType: "ajax", Level: "default",
			API:         string(a.Method) + ":" + es.RoutePrefix + "/${id}/" + a.Name,
			ConfirmText: a.ConfirmMessage,
		})
	}

	detailURL := es.RoutePrefix + "/${id}"
	return &widget.Node{
		Kind:  widget.NodePage,
		Label: es.Label,
		Children: []*widget.Node{{
			Kind:  widget.NodeForm,
			Label: es.Label,
			// Detail view: no submit (no api/URL), only initApi to load the record.
			// Use ReadURL so the renderer sets initApi; URL empty → no api config.
			DataSource: &widget.DataSource{ReadURL: detailURL},
			Children:   children,
			Actions:    actions,
		}},
		Actions: actions,
	}, nil
}

// ── Field → Node translation ──────────────────────────────────────────────────

// buildLayoutNodes converts es.Layout into WidgetTree section/tab nodes.
// Returns nil when the layout is empty (signals caller to use flat field list).
func buildLayoutNodes(es *compiler.EntitySchema, isEdit bool) []*widget.Node {
	layout := es.Layout
	if len(layout.Tabs) == 0 && len(layout.Sections) == 0 {
		return nil
	}
	if len(layout.Tabs) > 0 {
		tabsNode := &widget.Node{Kind: widget.NodeTabs}
		for _, tab := range layout.Tabs {
			tabNode := &widget.Node{
				Kind:     widget.NodeSection,
				Label:    tab.Label,
				Children: buildSectionNodes(es, tab.Sections, isEdit),
			}
			tabsNode.Children = append(tabsNode.Children, tabNode)
		}
		return []*widget.Node{tabsNode}
	}
	return buildSectionNodes(es, layout.Sections, isEdit)
}

// buildSectionNodes converts a slice of SectionDef into NodeSection nodes.
// Each section's columns are flattened into field children; column span hints
// are applied as columnClassName Props on individual field nodes.
func buildSectionNodes(es *compiler.EntitySchema, sections []def.SectionDef, isEdit bool) []*widget.Node {
	var nodes []*widget.Node
	for _, sec := range sections {
		secNode := &widget.Node{
			Kind:        widget.NodeSection,
			Label:       sec.Label,
			Collapsible: sec.Collapsible,
			Collapsed:   sec.Collapsed,
		}

		if len(sec.Columns) == 0 {
			// No columns declared — section has no fields (unusual but valid;
			// caller may use Props to inject content via PageBuilder).
		} else if len(sec.Columns) == 1 {
			// Single column: render fields in a vertical flow with no span class.
			for _, fname := range sec.Columns[0].Fields {
				f, ok := es.FieldsByName[fname]
				if !ok || f.Hidden || f.Name == "tenant_id" {
					continue
				}
				secNode.Children = append(secNode.Children, fieldToNode(f, es, isEdit))
			}
		} else {
			// Multiple columns: compute span class for each column's fields.
			// Equal span = 12 / numColumns when ColumnDef.Span is 0.
			autoSpan := 12 / len(sec.Columns)
			for _, col := range sec.Columns {
				span := col.Span
				if span <= 0 {
					span = autoSpan
				}
				colClass := fmt.Sprintf("col-md-%d", span)
				for _, fname := range col.Fields {
					f, ok := es.FieldsByName[fname]
					if !ok || f.Hidden || f.Name == "tenant_id" {
						continue
					}
					n := fieldToNode(f, es, isEdit)
					if n.Props == nil {
						n.Props = make(map[string]any)
					}
					n.Props["columnClassName"] = colClass
					secNode.Children = append(secNode.Children, n)
				}
			}
		}
		nodes = append(nodes, secNode)
	}
	return nodes
}

// setAllReadOnly recursively marks every leaf field node in a WidgetTree
// branch as read-only. Used by buildDetail to force the layout into a
// read-only state without duplicating the layout construction logic.
func setAllReadOnly(nodes []*widget.Node) {
	for _, n := range nodes {
		if len(n.Children) > 0 {
			setAllReadOnly(n.Children)
		} else {
			// Leaf node: field widget. Force read-only; clear expression
			// so DisabledOn doesn't override (detail is always read-only).
			n.ReadOnly = true
			n.DisabledOn = ""
		}
	}
}

func fieldToNode(f def.FieldDef, es *compiler.EntitySchema, isEdit bool) *widget.Node {
	n := &widget.Node{
		Name:        f.Name,
		Label:       fieldLabel(f),
		Description: f.Description,
		Required:    f.Required,
		ReadOnly:    f.ReadOnly,
		VisibleOn:   f.VisibleOn,
		HiddenOn:    f.HiddenOn,
		DisabledOn:  f.DisabledOn,
		RequiredOn:  f.RequiredOn,
	}
	if f.Immutable && isEdit {
		n.ReadOnly = true
	}

	switch f.Type {
	case def.FieldTypeData:
		n.Kind = widget.NodeText
		if f.MaxLen > 0 {
			n.Props = map[string]any{"maxLength": f.MaxLen}
		}
	case def.FieldTypeSmallText:
		n.Kind = widget.NodeTextArea
		n.Props = map[string]any{"rows": 2}
	case def.FieldTypeLongText:
		n.Kind = widget.NodeTextArea
		n.Props = map[string]any{"rows": 4}
	case def.FieldTypeInt:
		n.Kind = widget.NodeNumber
		n.Props = numberProps(f, 0)
	case def.FieldTypeFloat:
		n.Kind = widget.NodeNumber
		n.Props = numberProps(f, 6)
	case def.FieldTypeCurrency:
		n.Kind = widget.NodeNumber
		n.Props = numberProps(f, 4)
	case def.FieldTypeBool:
		n.Kind = widget.NodeSwitch
	case def.FieldTypeDate:
		n.Kind = widget.NodeDate
	case def.FieldTypeDateTime:
		n.Kind = widget.NodeDateTime
	case def.FieldTypeTime:
		n.Kind = widget.NodeDateTime
		n.Props = map[string]any{"format": "HH:mm"}
	case def.FieldTypeSelect:
		n.Kind = widget.NodeSelect
		if len(f.Options) > 0 {
			n.Props = map[string]any{"options": optionsList(f.Options)}
		}
	case def.FieldTypeMultiSelect:
		n.Kind = widget.NodeSelect
		n.Props = map[string]any{"multiple": true, "options": optionsList(f.Options)}
	case def.FieldTypeNamingSeries:
		n.Kind = widget.NodeText
		if isEdit {
			n.ReadOnly = true
		}
	case def.FieldTypeJSON:
		n.Kind = widget.NodeEditor
	case def.FieldTypeLink:
		n.Kind = widget.NodeSelect
		n.DataSource = linkDataSource(f, es)
		n.Props = map[string]any{"searchable": true}
	case def.FieldTypeLinkList:
		n.Kind = widget.NodeSelect
		n.DataSource = linkDataSource(f, es)
		n.Props = map[string]any{"searchable": true, "multiple": true}
	case def.FieldTypeDynamicLink:
		n.Kind = widget.NodeField
	default:
		n.Kind = widget.NodeText
	}
	return n
}

func optionsList(options []string) []map[string]any {
	out := make([]map[string]any, 0, len(options))
	for _, o := range options {
		out = append(out, map[string]any{"label": o, "value": o})
	}
	return out
}

func linkDataSource(f def.FieldDef, es *compiler.EntitySchema) *widget.DataSource {
	if lookup, ok := es.FieldLookups[f.Name]; ok {
		return &widget.DataSource{
			URL:        lookup.SearchURL,
			Method:     "GET",
			LabelField: lookup.LabelField,
			ValueField: lookup.ValueField,
		}
	}
	return &widget.DataSource{
		URL: "/api/v1/" + f.LinkTarget + "?q=${keywords}", Method: "GET",
		LabelField: "name", ValueField: "id",
	}
}

func numberProps(f def.FieldDef, precision int) map[string]any {
	props := map[string]any{"precision": precision}
	if f.Min != nil {
		props["min"] = *f.Min
	}
	if f.Max != nil {
		props["max"] = *f.Max
	}
	return props
}

// canPerform checks whether viewer may perform action on entity.
// Fails open on evaluator error — the API layer enforces the actual permission.
// A superfluous button that gets a 403 is acceptable; a missing button for an
// admin panel is a UX regression.
func canPerform(ctx context.Context, evaluator auth.PolicyEvaluator, viewer auth.ViewerContext, entity, action string) bool {
	if evaluator == nil || viewer == nil || viewer.IsPlatformAdmin() {
		return true
	}
	ok, _ := evaluator.CanPerform(ctx, viewer, entity, action)
	return ok
}
