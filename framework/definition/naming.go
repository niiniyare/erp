package definition

// NamingSeriesDef configures automatic document-number generation for an entity.
//
// The naming package increments a per-entity, per-tenant counter stored in
// naming_sequences and stamps the generated value onto Field during Create.
//
// Example configuration:
//
//	NamingSeries: &definition.NamingSeriesDef{
//	    Field:   "name",
//	    Prefix:  "INV-",
//	    Padding: 5,
//	}
//
// produces values like "INV-00001", "INV-00002", …
type NamingSeriesDef struct {
	// Field is the entity field that receives the generated value.
	// The field must exist in EntityDefinition.Fields.
	Field string

	// Prefix is prepended to the zero-padded sequence number.
	// May include date placeholders in future; currently treated as a
	// literal string.
	Prefix string

	// Padding is the minimum width of the numeric part, zero-padded.
	// Defaults to 5 when zero.
	Padding int
}
