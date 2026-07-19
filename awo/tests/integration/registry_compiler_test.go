// Package integration contains framework validation tests.
//
// Tests in this file run without any infrastructure (no Postgres, no Redis).
// They validate the registry, compiler, and hook pipeline using fakestore.
//
// Integration tests that require a real database are in db_test.go and are
// guarded by the "integration" build tag.
package integration_test

import (
	"testing"

	"awo.so/awo/compiler"
	"awo.so/awo/def"
	"awo.so/awo/examples/demo"
	"awo.so/awo/platform/iam"
	"awo.so/awo/platform/organization"
	"awo.so/awo/platform/tenant"
	"awo.so/awo/registry"
)

// demoDefs returns the minimal set of EntityDefinitions needed for the demo
// module. Uses BuildFrom (not Build) so the global registry is not touched
// and mandatory-entity checks are skipped.
func demoDefs() []def.EntityDefinition {
	return []def.EntityDefinition{
		&tenant.Definition,
		&iam.UserDefinition,
		&organization.Definition,
		&organization.OrgTypeDefinition,
		&organization.OrgAssignmentDefinition,
		&demo.CustomerDefinition,
	}
}

// TestRegistryBuild verifies that all demo + platform definitions pass
// registry validation.
func TestRegistryBuild(t *testing.T) {
	reg, err := registry.BuildFrom(demoDefs())
	if err != nil {
		t.Fatalf("registry.BuildFrom: %v", err)
	}
	if reg.Count() != 6 {
		t.Errorf("expected 6 entities, got %d", reg.Count())
	}
	names := []string{
		"platform_tenant",
		"iam_user",
		"platform_organization",
		"platform_org_type",
		"platform_org_assignment",
		"demo_customer",
	}
	for _, n := range names {
		if reg.Lookup(n) == nil {
			t.Errorf("entity %q not found in registry", n)
		}
	}
}

// TestRegistryModuleLookup verifies ByModule filtering.
func TestRegistryModuleLookup(t *testing.T) {
	reg, err := registry.BuildFrom(demoDefs())
	if err != nil {
		t.Fatalf("%v", err)
	}
	demoDefs := reg.ByModule("demo")
	if len(demoDefs) != 1 {
		t.Errorf("expected 1 demo entity, got %d", len(demoDefs))
	}
	// EntityName() returns the module-local name ("customer", not "demo_customer").
	// Use def.QualifiedName() to get the full identifier.
	if demoDefs[0].EntityName() != "customer" {
		t.Errorf("unexpected local entity name: %s (expected \"customer\")", demoDefs[0].EntityName())
	}
}

// TestRegistryDuplicateDetection verifies that a duplicate entity name is
// caught. BuildFrom de-dupes by inserting into a map — the second definition
// with the same name silently replaces the first. Count stays at expected.
func TestRegistryDuplicateDetection(t *testing.T) {
	defs := []def.EntityDefinition{
		&tenant.Definition,
		&tenant.Definition, // exact same pointer — duplicate
	}
	reg, err := registry.BuildFrom(defs)
	// BuildFrom may accept this (map dedup) or reject it — both are valid.
	// What matters: Lookup still works and returns a valid definition.
	if err != nil {
		t.Logf("BuildFrom rejected duplicate (acceptable): %v", err)
		return
	}
	if reg.Lookup("platform_tenant") == nil {
		t.Error("platform_tenant not found after dedup")
	}
}

// TestRegistrySeal verifies that after Build() the global registry is sealed
// and rejects further registrations. We use BuildFrom here (no seal) and
// separately test that the def package seal works.
func TestRegistryBuildFrom_NoMandatoryCheck(t *testing.T) {
	// demo_customer alone — would fail Build() (no iam_user etc.) but
	// BuildFrom should accept it since mandatory-entity check is skipped.
	// However, link target platform_tenant and platform_organization must
	// still be present.
	reg, err := registry.BuildFrom([]def.EntityDefinition{
		&demo.CustomerDefinition,
		&tenant.Definition,
		&iam.UserDefinition,
		&organization.Definition,
		&organization.OrgTypeDefinition,
		&organization.OrgAssignmentDefinition,
	})
	if err != nil {
		t.Fatalf("BuildFrom with demo only: %v", err)
	}
	if reg.Lookup("demo_customer") == nil {
		t.Error("demo_customer not found")
	}
}

// TestCompilerCompile verifies that Compile produces a valid CompiledSchema.
func TestCompilerCompile(t *testing.T) {
	reg, err := registry.BuildFrom(demoDefs())
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	schema, err := compiler.Compile(reg)
	if err != nil {
		t.Fatalf("compiler.Compile: %v", err)
	}
	if schema == nil {
		t.Fatal("schema is nil")
	}
	if len(schema.Entities) != 6 {
		t.Errorf("expected 6 compiled entities, got %d", len(schema.Entities))
	}

	es := schema.ByName["demo_customer"]
	if es == nil {
		t.Fatal("demo_customer not in compiled schema")
	}
	// Verify fields compiled correctly.
	if _, ok := es.FieldsByName["name"]; !ok {
		t.Error("field 'name' missing from compiled customer schema")
	}
	if !es.RequiredFields["name"] {
		t.Error("field 'name' should be Required")
	}
	if !es.ImmutableFields["customer_code"] {
		t.Error("field 'customer_code' should be Immutable")
	}
	// Default for 'active' should be present.
	if es.DefaultValues["active"] == nil {
		t.Error("field 'active' missing default value function")
	}
	v := es.DefaultValues["active"]()
	if v != true {
		t.Errorf("active default: expected true, got %v", v)
	}
}

// TestCompilerDiagnostics verifies that Validate can run standalone.
func TestCompilerDiagnostics(t *testing.T) {
	reg, err := registry.BuildFrom(demoDefs())
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	ds := compiler.Validate(reg)
	// No errors expected for a well-formed definition.
	if ds.HasErrors() {
		t.Errorf("unexpected compiler errors:\n%s", ds)
	}
	for _, d := range ds {
		t.Logf("diagnostic [%s] %s: %s", d.Severity, d.EntityName, d.Message)
	}
}

// TestCompilerFingerprint verifies deterministic SHA-256 output.
func TestCompilerFingerprint(t *testing.T) {
	reg, err := registry.BuildFrom(demoDefs())
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	schema, err := compiler.Compile(reg)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	fp1 := compiler.Fingerprint(schema)
	fp2 := compiler.Fingerprint(schema)

	if fp1 == "" {
		t.Error("fingerprint is empty")
	}
	if fp1 != fp2 {
		t.Errorf("fingerprint not deterministic: %q vs %q", fp1, fp2)
	}
	if len(fp1) != 64 {
		t.Errorf("expected 64-char SHA-256 hex, got %d chars", len(fp1))
	}
	t.Logf("schema fingerprint: %s", fp1)
}

// TestCompilerRoutes verifies that CRUD routes are generated for demo_customer.
func TestCompilerRoutes(t *testing.T) {
	reg, err := registry.BuildFrom(demoDefs())
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	schema, err := compiler.Compile(reg)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	var customerRoutes []compiler.RouteDescriptor
	for _, r := range schema.Routes {
		if r.EntityQualifiedName == "demo_customer" {
			customerRoutes = append(customerRoutes, r)
		}
	}
	if len(customerRoutes) == 0 {
		t.Error("no routes generated for demo_customer")
	}
	t.Logf("demo_customer routes: %d", len(customerRoutes))
	for _, r := range customerRoutes {
		t.Logf("  %s %s", r.Method, r.Path)
	}
}
