package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/awo/def"
	"awo.so/awo/registry"
)

// validateDefs returns (diagnostics, registryError).
// registryError is non-nil when registry.BuildFrom rejects the definitions
// before compiler.Validate even runs.
func validateDefs(defs []def.EntityDefinition) (Diagnostics, error) {
	reg, err := registry.BuildFrom(defs)
	if err != nil {
		return nil, err
	}
	return Validate(reg), nil
}

// mustValidate is for tests that expect the registry to accept the defs.
func mustValidate(t *testing.T, defs []def.EntityDefinition) Diagnostics {
	t.Helper()
	ds, err := validateDefs(defs)
	require.NoError(t, err, "registry.BuildFrom should accept these defs")
	return ds
}

// --- Tests for things caught by compiler.Validate ONLY (not by registry) ---

func TestValidate_Clean(t *testing.T) {
	ds := mustValidate(t, []def.EntityDefinition{simpleDef("mod_item", "mod")})
	assert.False(t, ds.HasErrors(), "expected no errors: %v", ds.Errors())
}

func TestValidate_RequiredAndDefault_Warning(t *testing.T) {
	d := &def.SystemDefinition{
		Name: "mod_item", Module: "mod", Label: "Item", LabelPlural: "Items",
		Fields: []def.FieldDef{
			{Name: "status", Type: def.FieldTypeData, Required: true, Default: func() any { return "draft" }},
		},
	}
	ds := mustValidate(t, []def.EntityDefinition{d})
	found := false
	for _, w := range ds.Warnings() {
		if w.FieldName == "status" {
			found = true
		}
	}
	assert.True(t, found, "expected Required+Default warning on field 'status'")
}

func TestValidate_ImmutableNoRequiredNoDefault_Warning(t *testing.T) {
	d := &def.SystemDefinition{
		Name: "mod_item", Module: "mod", Label: "Item", LabelPlural: "Items",
		Fields: []def.FieldDef{
			{Name: "code", Type: def.FieldTypeData, Immutable: true},
		},
	}
	ds := mustValidate(t, []def.EntityDefinition{d})
	found := false
	for _, w := range ds.Warnings() {
		if w.FieldName == "code" {
			found = true
		}
	}
	assert.True(t, found, "expected Immutable-without-Required-or-Default warning on field 'code'")
}

// FieldTypeInt with monetary suffix → compiler warning; registry only blocks FieldTypeFloat.
func TestValidate_MoneyInt_Warning(t *testing.T) {
	d := &def.SystemDefinition{
		Name: "mod_item", Module: "mod", Label: "Item", LabelPlural: "Items",
		Fields: []def.FieldDef{
			{Name: "unit_price", Type: def.FieldTypeInt},
		},
	}
	ds := mustValidate(t, []def.EntityDefinition{d})
	found := false
	for _, w := range ds.Warnings() {
		if w.FieldName == "unit_price" {
			found = true
		}
	}
	assert.True(t, found, "expected monetary-type warning on unit_price (FieldTypeInt)")
}

func TestValidate_Diagnostics_AsError_NilWhenClean(t *testing.T) {
	ds := mustValidate(t, []def.EntityDefinition{simpleDef("mod_x", "mod")})
	assert.Nil(t, ds.AsError(), "AsError should be nil when no errors")
}

// --- Tests for things caught by the registry (errors surface before compiler) ---

func TestValidate_NamingSeriesEmptySeries_CaughtByRegistry(t *testing.T) {
	d := &def.SystemDefinition{
		Name: "mod_item", Module: "mod", Label: "Item", LabelPlural: "Items",
		Fields: []def.FieldDef{
			{Name: "number", Type: def.FieldTypeNamingSeries, Series: ""},
		},
	}
	_, err := validateDefs([]def.EntityDefinition{d})
	require.Error(t, err, "registry should reject NamingSeries with empty Series")
	assert.Contains(t, err.Error(), "NamingSeries")
}

func TestValidate_NamingSeriesWithSeries_OK(t *testing.T) {
	d := &def.SystemDefinition{
		Name: "mod_item", Module: "mod", Label: "Item", LabelPlural: "Items",
		Fields: []def.FieldDef{
			{Name: "number", Type: def.FieldTypeNamingSeries, Series: "ITEM-{YYYY}-{SEQ:5}"},
		},
	}
	ds, err := validateDefs([]def.EntityDefinition{d})
	require.NoError(t, err)
	assert.False(t, ds.HasErrors(), "valid NamingSeries should produce no errors: %v", ds.Errors())
}

func TestValidate_MoneyFloat_CaughtByRegistry(t *testing.T) {
	d := &def.SystemDefinition{
		Name: "mod_item", Module: "mod", Label: "Item", LabelPlural: "Items",
		Fields: []def.FieldDef{
			{Name: "unit_price", Type: def.FieldTypeFloat},
		},
	}
	_, err := validateDefs([]def.EntityDefinition{d})
	require.Error(t, err, "registry should reject FieldTypeFloat with monetary field name")
}
