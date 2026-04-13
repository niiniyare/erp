package amis

// ---------------------------------------------------------------------------
// App shell (AMIS `app` component)
// ---------------------------------------------------------------------------

// AppBuilder builds the top-level AMIS `app` component, which provides
// the full shell: sidebar, routing, breadcrumbs, mobile drawer.
//
// Use this for the /schema/app endpoint. The frontend bootstraps with:
//
//	amis.embed('#root', { type: 'service', schemaApi: '/schema/app' }, ...)
type AppBuilder struct{ base }

// App creates a new app shell builder.
func App(brandName string) *AppBuilder {
	b := &AppBuilder{base{s: M{
		"type":      "app",
		"brandName": brandName,
		"pages":     A{},
	}}}
	return b
}

// Logo sets the logo URL.
func (b *AppBuilder) Logo(url string) *AppBuilder { b.set("logo", url); return b }

// Header sets header components (e.g. tenant name tpl, theme toggle button).
func (b *AppBuilder) Header(items ...any) *AppBuilder {
	b.set("header", items)
	return b
}

// Page adds a top-level nav page.
func (b *AppBuilder) Page(p NavPage) *AppBuilder {
	pages := b.s["pages"].(A)
	b.s["pages"] = append(pages, p.build())
	return b
}

// Pages adds multiple top-level nav pages.
func (b *AppBuilder) Pages(pages ...NavPage) *AppBuilder {
	for _, p := range pages {
		b.Page(p)
	}
	return b
}

// MarshalJSON implements json.Marshaler.
func (b *AppBuilder) MarshalJSON() ([]byte, error) { return b.marshalJSON() }

// Build returns the raw schema map.
func (b *AppBuilder) Build() Schema { return b.s }

// ---------------------------------------------------------------------------
// NavPage — a single entry in the app sidebar
// ---------------------------------------------------------------------------

// NavPage represents a single navigation entry.
// Use NavGroup() for entries with children, NavLink() for leaf pages.
type NavPage struct {
	label     string
	icon      string
	url       string
	schemaAPI string
	children  []NavPage
}

// NavLink creates a leaf navigation page.
// schemaAPI is the endpoint Go returns the page schema from, e.g. "/schema/finance/invoices".
func NavLink(label, icon, url, schemaAPI string) NavPage {
	return NavPage{label: label, icon: icon, url: url, schemaAPI: schemaAPI}
}

// NavGroup creates a navigation group with child items.
func NavGroup(label, icon string, children ...NavPage) NavPage {
	return NavPage{label: label, icon: icon, children: children}
}

func (p NavPage) build() M {
	m := M{"label": p.label}
	if p.icon != "" {
		m["icon"] = p.icon
	}
	if len(p.children) > 0 {
		children := make(A, len(p.children))
		for i, c := range p.children {
			children[i] = c.build()
		}
		m["children"] = children
	} else {
		m["url"] = p.url
		if p.schemaAPI != "" {
			m["schema"] = M{
				"type":      "service",
				"schemaApi": p.schemaAPI,
			}
		}
	}
	return m
}
