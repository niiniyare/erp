// Package generateCmd implements the `awo generate` subcommand.
package generateCmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Run executes `awo generate <what> <name>`.
func Run(_ context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: awo generate entity|migration <name>")
	}
	what, name := args[0], args[1]
	switch what {
	case "entity":
		return generateEntity(name)
	case "migration":
		return generateMigration(name)
	default:
		return fmt.Errorf("unknown generate target %q — use entity|migration", what)
	}
}

// ── entity ────────────────────────────────────────────────────────────────────

const entityTmpl = `package {{.Pkg}}

import "awo.so/framework/definition"

var {{.TypeName}}Def = &definition.EntityDefinition{
	Name:   "{{.Name}}",
	Label:  "{{.Label}}",
	Module: "TODO",
	Fields: []*definition.FieldDef{
		// TODO: add fields
	},
}

func init() {
	definition.Register({{.TypeName}}Def)
}
`

func generateEntity(name string) error {
	pkg := strings.ToLower(name)
	typeName := toPascalCase(name)
	label := toLabel(name)

	tmpl, err := template.New("entity").Parse(entityTmpl)
	if err != nil {
		return err
	}

	filename := pkg + ".go"
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create %s: %w", filename, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, map[string]string{
		"Pkg":      pkg,
		"TypeName": typeName,
		"Name":     name,
		"Label":    label,
	}); err != nil {
		return err
	}

	fmt.Printf("generated entity: %s\n", filename)
	return nil
}

// ── migration ─────────────────────────────────────────────────────────────────

func generateMigration(name string) error {
	// Find highest existing migration number.
	seq, err := nextMigrationSeq()
	if err != nil {
		return err
	}

	slug := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	up := fmt.Sprintf("%06d_%s.up.sql", seq, slug)
	down := fmt.Sprintf("%06d_%s.down.sql", seq, slug)

	upContent := fmt.Sprintf("-- %s\n-- TODO: write your UP migration\n", up)
	downContent := fmt.Sprintf("-- %s\n-- TODO: write your DOWN migration\n", down)

	if err := os.WriteFile(up, []byte(upContent), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", up, err)
	}
	if err := os.WriteFile(down, []byte(downContent), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", down, err)
	}

	fmt.Printf("generated migration: %s / %s\n", up, down)
	return nil
}

func nextMigrationSeq() (int, error) {
	entries, err := filepath.Glob("*.up.sql")
	if err != nil {
		return 0, err
	}
	max := 0
	for _, e := range entries {
		base := filepath.Base(e)
		var n int
		fmt.Sscanf(base, "%d", &n)
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func toLabel(s string) string {
	return strings.ReplaceAll(toPascalCase(s), "_", " ")
}
