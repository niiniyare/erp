package definition

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
}

// Edge starts a fluent EdgeDef declaration.
func Edge(name string) *EdgeDef {
	return &EdgeDef{Name: name}
}

func (e *EdgeDef) To(entity string) *EdgeDef          { e.TargetEntity = entity; return e }
func (e *EdgeDef) One() *EdgeDef                       { e.Cardinality = CardinalityOne; return e }
func (e *EdgeDef) Many() *EdgeDef                      { e.Cardinality = CardinalityMany; return e }
func (e *EdgeDef) ManyToMany() *EdgeDef                { e.Cardinality = CardinalityManyToMany; return e }
func (e *EdgeDef) WithForeignKey(col string) *EdgeDef  { e.ForeignKey = col; return e }
func (e *EdgeDef) WithJoinTable(t string) *EdgeDef     { e.JoinTable = t; return e }
func (e *EdgeDef) CascadeDelete() *EdgeDef             { e.Cascade = true; return e }
func (e *EdgeDef) Eager() *EdgeDef                     { e.EagerLoad = true; return e }
func (e *EdgeDef) WithLabel(l string) *EdgeDef         { e.Label = l; return e }
