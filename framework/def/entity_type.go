package def

// EntityType distinguishes how an entity's data is persisted.
//
// System entities own a dedicated PostgreSQL table generated from their
// EntityDefinition (column-per-field, full index support, typed SQL).
//
// Custom entities share the generic custom_entity_records table, storing
// field values in a JSONB data column. Use for tenant-defined runtime schema
// extensions that do not require schema migrations to add new fields.
type EntityType string

const (
	// EntityTypeSystem is the default. Entity owns a dedicated SQL table.
	// Table DDL is generated from EntityDefinition via the migration package.
	EntityTypeSystem EntityType = "system"

	// EntityTypeCustom stores all field values in the JSONB data column of the
	// custom_entity_records table. Fields may be added at runtime without
	// schema migrations. Querying is via GIN index on the data column.
	EntityTypeCustom EntityType = "custom"
)

// IsSystem reports whether t is EntityTypeSystem or unset (zero value defaults to system).
func (t EntityType) IsSystem() bool {
	return t == EntityTypeSystem || t == ""
}

// IsCustom reports whether t is EntityTypeCustom.
func (t EntityType) IsCustom() bool {
	return t == EntityTypeCustom
}
