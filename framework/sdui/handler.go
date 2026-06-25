// Package sdui exposes an HTTP endpoint that serves generated AMIS schemas
// for registered EntityDefinitions.
package sdui

import (
	"github.com/gofiber/fiber/v2"

	"awo.so/framework/definition"
	"awo.so/framework/sdui/amis"
)

// NavItem is one entry in the navigation tree returned by GET /sdui/nav.
type NavItem struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Module string `json:"module,omitempty"`
	Icon   string `json:"icon,omitempty"`
}

// NavGroup is a labelled section in the sidebar nav.
type NavGroup struct {
	Group string    `json:"group"`
	Items []NavItem `json:"items"`
}

// Register mounts the SDUI schema endpoint on router.
//
//	GET /sdui/nav            → nav tree for all registered entities
//	GET /sdui/:entity        → CRUDPage schema
//	GET /sdui/:entity/form   → FormPage schema
func Register(router fiber.Router, apiBase string) {
	g := router.Group("/sdui")
	g.Get("/nav", func(c *fiber.Ctx) error {
		return serveNav(c)
	})
	g.Get("/:entity", func(c *fiber.Ctx) error {
		return servePage(c, apiBase, false)
	})
	g.Get("/:entity/form", func(c *fiber.Ctx) error {
		return servePage(c, apiBase, true)
	})
}

// serveNav returns the sidebar navigation structure derived from all registered
// EntityDefinitions, grouped by EntityDefinition.Module.
func serveNav(c *fiber.Ctx) error {
	defs := definition.All()

	// Group by module, preserving first-seen order.
	order := []string{}
	groups := map[string][]NavItem{}

	for _, d := range defs {
		if d.IsGlobal() {
			continue // global entities are not user-navigable by default
		}
		mod := d.Module
		if mod == "" {
			mod = "Other"
		}
		if _, ok := groups[mod]; !ok {
			order = append(order, mod)
		}
		label := d.Label
		if label == "" {
			label = d.Name
		}
		groups[mod] = append(groups[mod], NavItem{
			ID:     d.Name,
			Label:  label,
			Module: mod,
		})
	}

	nav := make([]NavGroup, 0, len(order))
	for _, mod := range order {
		nav = append(nav, NavGroup{Group: mod, Items: groups[mod]})
	}

	return c.JSON(nav)
}

func servePage(c *fiber.Ctx, apiBase string, formOnly bool) error {
	name := c.Params("entity")
	def := definition.Lookup(name)
	if def == nil {
		return fiber.NewError(fiber.StatusNotFound, "unknown entity: "+name)
	}

	opts := amis.PageOpts{APIBase: apiBase}

	var schema map[string]any
	if formOnly {
		schema = amis.FormPage(def, opts)
	} else {
		schema = amis.CRUDPage(def, opts)
	}

	return c.JSON(schema)
}
