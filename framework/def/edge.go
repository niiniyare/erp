package def

import (
	"fmt"
	"strings"
)

// Cardinality describes the multiplicity of a relationship between entities.
type Cardinality string

const (
	// CardinalityOne is a many-to-one or one-to-one relationship (FK on this side).
	CardinalityOne Cardinality = "one"

	// CardinalityMany is a one-to-many relationship (FK on the other side).
	CardinalityMany Cardinality = "many"

	// CardinalityManyToMany is a many-to-many relationship via a join table.
	CardinalityManyToMany Cardinality = "many_to_many"
)

// EdgeDef declares a relationship between two EntityDefinitions.
// Edges drive eager-load helpers, cascade-delete rules, and SDUI sub-form rendering.
type EdgeDef struct {
	// Name is the Go field name used in Record.Get and JSON wire format.
	Name string

	// TargetEntity is the registered entity name of the related object.
	TargetEntity string

	// Cardinality describes the multiplicity of the relationship.
	Cardinality Cardinality

	// ForeignKey is the FK column name. For CardinalityOne it lives on this
	// entity's table; for CardinalityMany it lives on the target's table.
	// Defaults to "{Name}_id" (CardinalityOne) or "{ThisEntity}_id" (CardinalityMany).
	ForeignKey string

	// JoinTable is required for CardinalityManyToMany.
	JoinTable string

	// Cascade indicates that deleting this record should cascade-delete edge records.
	Cascade bool

	// EagerLoad causes the edge to be fetched and embedded automatically in FindByID.
	EagerLoad bool

	// Label is the human-readable name shown in SDUI sub-forms.
	Label string

	// IsRequired indicates if this relationship is mandatory.
	IsRequired bool

	// RefField is the name of the edge on the target entity that this edge refers to (for inverse edges).
	RefField string

	// AnnotationsList holds custom schema annotations (e.g., cascade rules, DB indexes).
	AnnotationsList []any
}

// Package-level constructor helpers matching the documentation's simplified builder API.

// To creates an owner-side EdgeDef pointing to targetEntity (many-to-one).
func To(name string, targetEntity string) *EdgeDef {
	return Edge(name).To(targetEntity).One()
}

// From creates an inverse-side EdgeDef pointing from targetEntity (one-to-many).
func From(name string, targetEntity string) *EdgeDef {
	return Edge(name).To(targetEntity).Many()
}

// Edge starts a fluent EdgeDef declaration.
func Edge(name string) *EdgeDef {
	return &EdgeDef{Name: name}
}

func (e *EdgeDef) To(entity string) *EdgeDef          { e.TargetEntity = entity; return e }
func (e *EdgeDef) One() *EdgeDef                      { e.Cardinality = CardinalityOne; return e }
func (e *EdgeDef) Many() *EdgeDef                     { e.Cardinality = CardinalityMany; return e }
func (e *EdgeDef) ManyToMany() *EdgeDef               { e.Cardinality = CardinalityManyToMany; return e }
func (e *EdgeDef) WithForeignKey(col string) *EdgeDef { e.ForeignKey = col; return e }
func (e *EdgeDef) WithJoinTable(t string) *EdgeDef    { e.JoinTable = t; return e }
func (e *EdgeDef) CascadeDelete() *EdgeDef            { e.Cascade = true; return e }
func (e *EdgeDef) Eager() *EdgeDef                    { e.EagerLoad = true; return e }
func (e *EdgeDef) WithLabel(l string) *EdgeDef        { e.Label = l; return e }

// Documentation-aligned fluent builder API aliases and helpers

// Field specifies the explicit foreign key column name. Alias for WithForeignKey.
func (e *EdgeDef) Field(col string) *EdgeDef {
	return e.WithForeignKey(col)
}

// Required marks the relationship as mandatory.
func (e *EdgeDef) Required() *EdgeDef {
	e.IsRequired = true
	return e
}

// Ref specifies the owner-side edge name that this inverse edge references.
func (e *EdgeDef) Ref(ref string) *EdgeDef {
	e.RefField = ref
	return e
}

// Annotations stores custom metadata or db annotations on the edge.
// If any annotation contains "cascade", it automatically marks the edge as Cascade = true.
func (e *EdgeDef) Annotations(anns ...any) *EdgeDef {
	e.AnnotationsList = append(e.AnnotationsList, anns...)
	for _, ann := range anns {
		annStr := fmt.Sprintf("%v", ann)
		if strings.Contains(strings.ToLower(annStr), "cascade") {
			e.Cascade = true
		}
	}
	return e
}
