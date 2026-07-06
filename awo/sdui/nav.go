package sdui

import (
	"sort"
	"strings"

	"awo.so/awo/compiler"
)

// BuildNav generates sidebar navigation from a compiled schema, grouping
// entities by module. Each module becomes a nav group; each entity within
// becomes a leaf item linking to its list page.
func BuildNav(schema *compiler.CompiledSchema) []NavItem {
	// Group entities by module.
	groups := make(map[string][]NavItem)
	var moduleOrder []string

	for _, es := range schema.Entities {
		moduleName := es.Def.EntityModule()
		name := es.Def.EntityName()
		label := es.Def.EntityLabel()

		item := NavItem{
			Label: label,
			To:    "/entities/" + name,
		}

		if _, seen := groups[moduleName]; !seen {
			moduleOrder = append(moduleOrder, moduleName)
		}
		groups[moduleName] = append(groups[moduleName], item)
	}

	sort.Strings(moduleOrder)

	nav := make([]NavItem, 0, len(moduleOrder))
	for _, mod := range moduleOrder {
		children := groups[mod]
		sort.Slice(children, func(i, j int) bool {
			return children[i].Label < children[j].Label
		})
		nav = append(nav, NavItem{
			Label:    strings.ToUpper(mod[:1]) + mod[1:],
			Children: children,
		})
	}
	return nav
}
