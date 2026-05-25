package amis

// ---------------------------------------------------------------------------
// Page
// ---------------------------------------------------------------------------

// PageBuilder builds an AMIS `page` component.
type PageBuilder struct{ base }

// Page creates a new page builder.
func Page(title string) *PageBuilder {
	b := &PageBuilder{base{s: M{"type": "page"}}}
	if title != "" {
		b.s["title"] = title
	}
	return b
}

// Body sets the page body. Accepts a single schema M or a slice A.
func (b *PageBuilder) Body(v any) *PageBuilder { b.set("body", v); return b }

// Aside sets the page aside panel.
func (b *PageBuilder) Aside(v any) *PageBuilder { b.set("aside", v); return b }

// Toolbar appends items to the page toolbar.
func (b *PageBuilder) Toolbar(items ...any) *PageBuilder {
	b.set("toolbar", items)
	return b
}

// Data injects initial data into the page scope (permissions, tenant context, etc.).
func (b *PageBuilder) Data(d M) *PageBuilder { b.set("data", d); return b }

// ClassName sets additional CSS classes on the page wrapper.
func (b *PageBuilder) ClassName(c string) *PageBuilder { b.set("className", c); return b }

// MarshalJSON implements json.Marshaler.
func (b *PageBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map (useful when embedding in larger schemas).
func (b *PageBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// ServiceBuilder builds an AMIS `service` component — loads data once and
// makes it available to all child components via the data chain.
type ServiceBuilder struct{ base }

// Service creates a new service builder. api is the GET endpoint, e.g.
// "get:/api/v1/payroll/dashboard-summary".
func Service(api string) *ServiceBuilder {
	b := &ServiceBuilder{base{s: M{"type": "service"}}}
	b.set("api", api)
	return b
}

// SchemaAPI tells the service to also load a remote schema (for dynamic pages).
func (b *ServiceBuilder) SchemaAPI(api string) *ServiceBuilder {
	b.set("schemaApi", api)
	return b
}

// Body sets the child schema rendered with the loaded data in scope.
func (b *ServiceBuilder) Body(items ...any) *ServiceBuilder {
	if len(items) == 1 {
		b.set("body", items[0])
	} else {
		b.set("body", items)
	}
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *ServiceBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *ServiceBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Grid
// ---------------------------------------------------------------------------

// GridBuilder builds an AMIS `grid` layout component.
type GridBuilder struct{ base }

// Grid creates a new grid layout. Each column is an AMIS column object.
// Columns should be wrapped with Col() for convenience.
func Grid(cols ...any) *GridBuilder {
	b := &GridBuilder{base{s: M{"type": "grid", "columns": cols}}}
	return b
}

// Gap sets the grid gap.
func (b *GridBuilder) Gap(px int) *GridBuilder { b.set("gap", px); return b }

// MarshalJSON implements json.Marshaler.
func (b *GridBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *GridBuilder) Build() Schema { return b.s }

// Col creates a grid column. Pass width (e.g. 4 out of 12) and body content.
func Col(width int, body any) M {
	return M{"md": width, "body": body}
}

// ---------------------------------------------------------------------------
// Panel
// ---------------------------------------------------------------------------

// PanelBuilder builds an AMIS `panel` component (a card with header/body/footer).
type PanelBuilder struct{ base }

// Panel creates a new panel with a title.
func Panel(title string) *PanelBuilder {
	b := &PanelBuilder{base{s: M{"type": "panel"}}}
	if title != "" {
		b.s["title"] = title
	}
	return b
}

// Body sets the panel body. Accepts any schema value.
func (b *PanelBuilder) Body(v any) *PanelBuilder { b.set("body", v); return b }

// Footer sets the panel footer.
func (b *PanelBuilder) Footer(v any) *PanelBuilder { b.set("footer", v); return b }

// ClassName sets additional CSS classes.
func (b *PanelBuilder) ClassName(c string) *PanelBuilder { b.set("className", c); return b }

// HeaderClassName sets CSS classes on the panel header.
func (b *PanelBuilder) HeaderClassName(c string) *PanelBuilder {
	b.set("headerClassName", c)
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *PanelBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *PanelBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Tabs
// ---------------------------------------------------------------------------

// TabsBuilder builds an AMIS `tabs` component.
type TabsBuilder struct{ base }

// Tabs creates a new tabs builder.
func Tabs() *TabsBuilder {
	return &TabsBuilder{base{s: M{"type": "tabs", "tabs": A{}}}}
}

// Tab adds a tab. title is the label; body is the tab content.
func (b *TabsBuilder) Tab(title string, body any) *TabsBuilder {
	tabs := b.s["tabs"].(A)
	b.s["tabs"] = append(tabs, M{"title": title, "body": body})
	return b
}

// Mode sets the tab display mode: "line" | "card" | "radio" | "vertical".
func (b *TabsBuilder) Mode(mode string) *TabsBuilder { b.set("tabsMode", mode); return b }

// MarshalJSON implements json.Marshaler.
func (b *TabsBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *TabsBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Stat (statistic card)
// ---------------------------------------------------------------------------

// StatBuilder builds an AMIS `statistic` component for KPI cards.
type StatBuilder struct{ base }

// Stat creates a new statistic component. source is the data expression e.g. "${total}".
func Stat(label, source string) *StatBuilder {
	b := &StatBuilder{base{s: M{
		"type":   "tpl",
		"tpl":    statTpl(label, source),
		"className": "stat-card",
	}}}
	return b
}

func statTpl(label, source string) string {
	return `<div class="stat-item"><div class="stat-value">` + source + `</div><div class="stat-label">` + label + `</div></div>`
}

// MarshalJSON implements json.Marshaler.
func (b *StatBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// ---------------------------------------------------------------------------
// Chart
// ---------------------------------------------------------------------------

// ChartBuilder builds an AMIS `chart` component (ECharts).
// Always sets backgroundColor: transparent per Decision 8.
type ChartBuilder struct{ base }

// Chart creates a new chart builder. Pass an empty string for static (config-only) charts.
// Always emits both:
//   - config.backgroundColor:"transparent" — ECharts canvas background
//   - style.background:"transparent"       — AMIS wrapper div (required by ValidateStage)
func Chart(api string) *ChartBuilder {
	s := M{
		"type":   "chart",
		"config": M{"backgroundColor": "transparent"},
		// style.background is checked by ruleValidateChartTransparentBg.
		// Without this the ValidateStage rejects every legacy chart page with HTTP 500.
		"style": M{"background": "transparent"},
	}
	if api != "" {
		s["api"] = api
	}
	return &ChartBuilder{base{s: s}}
}

// Config merges additional ECharts config. backgroundColor: transparent is
// always preserved.
func (b *ChartBuilder) Config(cfg M) *ChartBuilder {
	merged := M{"backgroundColor": "transparent"}
	for k, v := range cfg {
		merged[k] = v
	}
	b.s["config"] = merged
	return b
}

// Height sets the chart height in pixels.
func (b *ChartBuilder) Height(px int) *ChartBuilder { b.set("height", px); return b }

// MarshalJSON implements json.Marshaler.
func (b *ChartBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *ChartBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Tpl (inline template)
// ---------------------------------------------------------------------------

// Tpl creates an AMIS `tpl` component from a Go template string.
// Use CSS variables for any colours — no hardcoded hex values.
func Tpl(tpl string) M {
	return M{"type": "tpl", "tpl": tpl}
}

// ---------------------------------------------------------------------------
// Breadcrumb
// ---------------------------------------------------------------------------

// BreadcrumbItem is a single crumb. Set Href to make it a link.
type BreadcrumbItem struct {
	Label string
	Href  string
	Icon  string
}

// BC is a shorthand breadcrumb item constructor.
//   - BC("Home", "#dashboard")  → link crumb
//   - BC("Finance")             → plain crumb (last segment)
func BC(label string, href ...string) BreadcrumbItem {
	b := BreadcrumbItem{Label: label}
	if len(href) > 0 {
		b.Href = href[0]
	}
	return b
}

// Breadcrumb builds an AMIS `breadcrumb` component.
func Breadcrumb(items ...BreadcrumbItem) M {
	arr := make(A, len(items))
	for i, item := range items {
		m := M{"label": item.Label}
		if item.Href != "" {
			m["href"] = item.Href
		}
		if item.Icon != "" {
			m["icon"] = item.Icon
		}
		arr[i] = m
	}
	return M{"type": "breadcrumb", "items": arr, "className": "mb-3"}
}

// ---------------------------------------------------------------------------
// Divider (horizontal rule between sections)
// ---------------------------------------------------------------------------

// HDivider returns a horizontal divider schema element.
func HDivider() M {
	return M{"type": "divider", "className": "my-3"}
}

// ---------------------------------------------------------------------------
// Alert
// ---------------------------------------------------------------------------

// AlertBuilder builds an AMIS `alert` component.
type AlertBuilder struct{ base }

// Alert creates an inline alert. level: "info" | "success" | "warning" | "danger".
func Alert(level, body string) *AlertBuilder {
	b := &AlertBuilder{base{s: M{
		"type":  "alert",
		"level": level,
		"body":  body,
	}}}
	return b
}

// VisibleOn sets the visibility expression.
func (b *AlertBuilder) VisibleOn(expr string) *AlertBuilder {
	b.set("visibleOn", expr)
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *AlertBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *AlertBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Timeline
// ---------------------------------------------------------------------------

// TimelineBuilder builds an AMIS `timeline` component.
type TimelineBuilder struct{ base }

// Timeline creates a timeline from an API source.
func Timeline(api string) *TimelineBuilder {
	b := &TimelineBuilder{base{s: M{
		"type": "timeline",
		"api":  api,
	}}}
	return b
}

// Items sets static timeline items when no API is needed.
// Each item: M{"time": "...", "title": "...", "detail": "..."}
func (b *TimelineBuilder) Items(items ...M) *TimelineBuilder {
	arr := make(A, len(items))
	for i, item := range items {
		arr[i] = item
	}
	b.set("items", arr)
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *TimelineBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *TimelineBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// Descriptions (read-only key-value)
// ---------------------------------------------------------------------------

// DescriptionsBuilder builds an AMIS `descriptions` component.
type DescriptionsBuilder struct{ base }

// Descriptions creates a descriptions (key-value) display.
func Descriptions(title string) *DescriptionsBuilder {
	b := &DescriptionsBuilder{base{s: M{"type": "descriptions"}}}
	if title != "" {
		b.s["title"] = title
	}
	return b
}

// Item adds a field. label is the display name; name is the data key.
func (b *DescriptionsBuilder) Item(label, name string) *DescriptionsBuilder {
	items, _ := b.s["items"].(A)
	b.s["items"] = append(items, M{"label": label, "name": name})
	return b
}

// Columns sets how many columns to lay out (default 1).
func (b *DescriptionsBuilder) Columns(n int) *DescriptionsBuilder {
	b.set("columns", n)
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *DescriptionsBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *DescriptionsBuilder) Build() Schema { return b.s }
