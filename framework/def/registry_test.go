package def_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/framework/def"
)

func TestRegistry_RegisterAndLookup(t *testing.T) {
	r := def.NewRegistry()

	d := &def.EntityDefinition{
		Name:  "test_widget",
		Label: "Widget",
		Fields: []*def.FieldDef{
			def.String("name").RequiredField(),
		},
	}
	r.Register(d)

	got := r.Lookup("test_widget")
	require.NotNil(t, got)
	assert.Equal(t, "test_widget", got.Name)
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	r := def.NewRegistry()
	d := &def.EntityDefinition{Name: "dup_entity"}
	r.Register(d)

	assert.Panics(t, func() {
		r.Register(&def.EntityDefinition{Name: "dup_entity"})
	})
}

func TestRegistry_LookupMissing(t *testing.T) {
	r := def.NewRegistry()
	assert.Nil(t, r.Lookup("nonexistent"))
}

func TestRegistry_AllByType_System(t *testing.T) {
	r := def.NewRegistry()

	r.Register(&def.EntityDefinition{Name: "sys_a", Type: def.EntityTypeSystem})
	r.Register(&def.EntityDefinition{Name: "sys_b"}) // zero value = system
	r.Register(&def.EntityDefinition{Name: "cust_a", Type: def.EntityTypeCustom})

	system := r.AllByType(def.EntityTypeSystem)
	assert.Len(t, system, 2)

	names := make(map[string]bool, len(system))
	for _, d := range system {
		names[d.Name] = true
	}
	assert.True(t, names["sys_a"])
	assert.True(t, names["sys_b"])
	assert.False(t, names["cust_a"])
}

func TestRegistry_AllByType_Custom(t *testing.T) {
	r := def.NewRegistry()

	r.Register(&def.EntityDefinition{Name: "sys_x", Type: def.EntityTypeSystem})
	r.Register(&def.EntityDefinition{Name: "cust_x", Type: def.EntityTypeCustom})
	r.Register(&def.EntityDefinition{Name: "cust_y", Type: def.EntityTypeCustom})

	custom := r.AllByType(def.EntityTypeCustom)
	assert.Len(t, custom, 2)
}

func TestEntityType_Helpers(t *testing.T) {
	assert.True(t, def.EntityTypeSystem.IsSystem())
	assert.False(t, def.EntityTypeSystem.IsCustom())

	assert.True(t, def.EntityTypeCustom.IsCustom())
	assert.False(t, def.EntityTypeCustom.IsSystem())

	// Zero value defaults to system.
	var zero def.EntityType
	assert.True(t, zero.IsSystem())
	assert.False(t, zero.IsCustom())
}

func TestEntityDefinition_LabelPluralName(t *testing.T) {
	cases := []struct {
		def  def.EntityDefinition
		want string
	}{
		{def.EntityDefinition{Name: "invoice", Label: "Invoice", LabelPlural: "Invoices"}, "Invoices"},
		{def.EntityDefinition{Name: "invoice", Label: "Invoice"}, "Invoices"},
		{def.EntityDefinition{Name: "invoice"}, "invoices"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.def.LabelPluralName())
	}
}

func TestEntityDefinition_FieldSet(t *testing.T) {
	d := &def.EntityDefinition{
		Name: "item",
		Fields: []*def.FieldDef{
			def.String("code"),
			def.String("description"),
		},
	}
	fs := d.FieldSet()
	assert.Contains(t, fs, "code")
	assert.Contains(t, fs, "description")
	assert.NotContains(t, fs, "nonexistent")
}
