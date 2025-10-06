package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"
)

// TestGenerationSuite runs comprehensive tests for the UI generation system
func TestGenerationSuite(t *testing.T) {
	t.Run("TestAnalyzer", testAnalyzer)
	t.Run("TestReflector", testReflector)
	t.Run("TestPatternMatcher", testPatternMatcher)
	t.Run("TestSchemaGenerator", testSchemaGenerator)
	t.Run("TestTagSystem", testTagSystem)
	t.Run("TestEndToEndGeneration", testEndToEndGeneration)
}

// testAnalyzer tests the struct analyzer functionality
func testAnalyzer(t *testing.T) {
	analyzer := NewAnalyzer()

	// Test reflection-based analysis
	userType := reflect.TypeOf(TestUser{})
	structInfo := analyzer.AnalyzeStruct(userType)

	// Verify basic struct information
	if structInfo.Name != "TestUser" {
		t.Errorf("Expected struct name 'TestUser', got '%s'", structInfo.Name)
	}

	if len(structInfo.Fields) == 0 {
		t.Error("Expected fields to be parsed, got none")
	}

	// Verify specific field parsing
	var emailField *FieldInfo
	for _, field := range structInfo.Fields {
		if field.Name == "Email" {
			emailField = &field
			break
		}
	}

	if emailField == nil {
		t.Error("Email field not found")
	} else {
		if emailField.Component != "email" {
			t.Errorf("Expected email component, got '%s'", emailField.Component)
		}
		if !emailField.Required {
			t.Error("Email field should be required")
		}
		if !emailField.IsStringType {
			t.Error("Email field should be string type")
		}
	}

	// Verify business logic detection
	if !structInfo.IsCRUDEntity {
		t.Error("TestUser should be detected as CRUD entity")
	}

	if !structInfo.HasAuditTrail {
		t.Error("TestUser should have audit trail")
	}

	if !structInfo.HasSoftDelete {
		t.Error("TestUser should have soft delete")
	}

	if !structInfo.HasTenantScope {
		t.Error("TestUser should have tenant scope")
	}

	log.Printf("✓ Analyzer test passed - Found %d fields in TestUser", len(structInfo.Fields))
}

// testReflector tests the reflector functionality
func testReflector(t *testing.T) {
	reflector := NewReflector()

	// Test struct reflection
	user := TestUser{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Status:    "active",
		IsAdmin:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	structInfo, err := reflector.ReflectStruct(user)
	if err != nil {
		t.Fatalf("Failed to reflect struct: %v", err)
	}

	if structInfo.Name != "TestUser" {
		t.Errorf("Expected struct name 'TestUser', got '%s'", structInfo.Name)
	}

	// Test value extraction
	values, err := reflector.ReflectValue(user)
	if err != nil {
		t.Fatalf("Failed to reflect values: %v", err)
	}

	if values["Email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", values["Email"])
	}

	if values["IsAdmin"] != false {
		t.Errorf("Expected IsAdmin false, got '%v'", values["IsAdmin"])
	}

	log.Printf("✓ Reflector test passed - Extracted %d field values", len(values))
}

// testPatternMatcher tests the pattern matching functionality
func testPatternMatcher(t *testing.T) {
	matcher := NewPatternMatcher()
	analyzer := NewAnalyzer()

	// Test different entity types
	testCases := []struct {
		entityType      reflect.Type
		expectedPattern PatternType
		minScore        float64
	}{
		{reflect.TypeOf(TestUser{}), PatternUserProfile, 10.0},
		{reflect.TypeOf(TestProject{}), PatternTaskBoard, 10.0},
		{reflect.TypeOf(TestTask{}), PatternKanbanBoard, 10.0},
		{reflect.TypeOf(TestClient{}), PatternContactCard, 10.0},
		{reflect.TypeOf(TestInvoice{}), PatternInvoiceForm, 12.0},
		{reflect.TypeOf(TestEvent{}), PatternCalendarView, 10.0},
		{reflect.TypeOf(TestNotification{}), PatternNotifications, 8.0},
		{reflect.TypeOf(TestDepartment{}), PatternHierarchyTree, 10.0},
	}

	for _, testCase := range testCases {
		structInfo := analyzer.AnalyzeStruct(testCase.entityType)
		patterns := matcher.MatchPatterns(structInfo)

		if len(patterns) == 0 {
			t.Errorf("No patterns matched for %s", structInfo.Name)
			continue
		}

		// Find the expected pattern
		var foundPattern *PatternScore
		for _, pattern := range patterns {
			if pattern.Pattern == testCase.expectedPattern {
				foundPattern = &pattern
				break
			}
		}

		if foundPattern == nil {
			t.Errorf("Expected pattern %s not found for %s. Got: %v",
				testCase.expectedPattern, structInfo.Name, patterns[0].Pattern)
			continue
		}

		if foundPattern.Score < testCase.minScore {
			t.Errorf("Pattern %s score too low for %s: %.2f < %.2f",
				testCase.expectedPattern, structInfo.Name, foundPattern.Score, testCase.minScore)
		}

		log.Printf("✓ Pattern matcher test passed for %s - matched %s with score %.2f",
			structInfo.Name, foundPattern.Pattern, foundPattern.Score)
	}
}

// testSchemaGenerator tests the schema generation functionality
func testSchemaGenerator(t *testing.T) {
	config := GeneratorConfig{
		DefaultLayout:     "app",
		EnablePermissions: true,
		EnableAuditTrail:  true,
		APIPrefix:         "/api",
		TablePageSize:     20,
		EnableFiltering:   true,
		EnableSorting:     true,
		EnableBulkActions: true,
	}

	generator := NewSchemaGenerator(config)
	analyzer := NewAnalyzer()

	// Test schema generation for TestUser
	userType := reflect.TypeOf(TestUser{})
	structInfo := analyzer.AnalyzeStruct(userType)

	uiSchema, err := generator.GenerateUISchema(structInfo)
	if err != nil {
		t.Fatalf("Failed to generate UI schema: %v", err)
	}

	// Verify schema structure
	if uiSchema.Entity != "TestUser" {
		t.Errorf("Expected entity 'TestUser', got '%s'", uiSchema.Entity)
	}

	if uiSchema.ListSchema == nil {
		t.Error("List schema should be generated")
	}

	if uiSchema.CreateSchema == nil {
		t.Error("Create schema should be generated")
	}

	if uiSchema.EditSchema == nil {
		t.Error("Edit schema should be generated")
	}

	if uiSchema.DetailSchema == nil {
		t.Error("Detail schema should be generated")
	}

	if uiSchema.FilterSchema == nil {
		t.Error("Filter schema should be generated")
	}

	// Verify metadata
	if !uiSchema.Metadata.IsCRUD {
		t.Error("Metadata should indicate CRUD entity")
	}

	if !uiSchema.Metadata.HasAuditTrail {
		t.Error("Metadata should indicate audit trail")
	}

	if uiSchema.Metadata.PrimaryKey != "ID" {
		t.Errorf("Expected primary key 'ID', got '%s'", uiSchema.Metadata.PrimaryKey)
	}

	log.Printf("✓ Schema generator test passed - Generated complete UI schema for TestUser")
}

// testTagSystem tests the tag parsing and customization system
func testTagSystem(t *testing.T) {
	config := TagSystemConfig{
		DefaultComponent: "text",
		StrictValidation: true,
		AllowUnknownTags: false,
	}

	tagSystem := NewTagSystem(config)

	// Test tag parsing on TestUser fields
	userType := reflect.TypeOf(TestUser{})
	emailField, found := userType.FieldByName("Email")
	if !found {
		t.Fatal("Email field not found")
	}

	tagData, err := tagSystem.ParseFieldTags(emailField)
	if err != nil {
		t.Fatalf("Failed to parse field tags: %v", err)
	}

	// Verify UI tag parsing
	uiTags, exists := tagData.Tags["ui"]
	if !exists {
		t.Error("UI tags should be parsed")
	} else {
		if uiTags["component"] != "email" {
			t.Errorf("Expected component 'email', got '%v'", uiTags["component"])
		}

		if uiTags["required"] != true {
			t.Errorf("Expected required true, got '%v'", uiTags["required"])
		}

		if uiTags["label"] != "Email Address" {
			t.Errorf("Expected label 'Email Address', got '%v'", uiTags["label"])
		}
	}

	// Verify validation tag parsing
	validateTags, exists := tagData.Tags["validate"]
	if !exists {
		t.Error("Validate tags should be parsed")
	} else {
		rules, ok := validateTags["rules"].([]map[string]interface{})
		if !ok || len(rules) == 0 {
			t.Error("Validation rules should be parsed")
		}
	}

	// Verify database tag parsing
	dbTags, exists := tagData.Tags["db"]
	if !exists {
		t.Error("Database tags should be parsed")
	} else {
		if dbTags["column"] != "email" {
			t.Errorf("Expected db column 'email', got '%v'", dbTags["column"])
		}

		if dbTags["unique"] != true {
			t.Error("Email field should be marked as unique")
		}
	}

	log.Printf("✓ Tag system test passed - Parsed %d tag types with %d issues",
		len(tagData.Tags), len(tagData.Issues))
}

// testEndToEndGeneration tests the complete generation workflow
func testEndToEndGeneration(t *testing.T) {
	// Test entities with different characteristics
	testEntities := []reflect.Type{
		reflect.TypeOf(TestUser{}),
		reflect.TypeOf(TestProject{}),
		reflect.TypeOf(TestTask{}),
		reflect.TypeOf(TestClient{}),
		reflect.TypeOf(TestInvoice{}),
		reflect.TypeOf(TestEvent{}),
		reflect.TypeOf(TestNotification{}),
		reflect.TypeOf(TestDepartment{}),
	}

	config := GeneratorConfig{
		DefaultLayout:     "app",
		EnablePermissions: true,
		EnableAuditTrail:  true,
		APIPrefix:         "/api",
		TablePageSize:     20,
		EnableFiltering:   true,
		EnableSorting:     true,
		EnableBulkActions: true,
	}

	analyzer := NewAnalyzer()
	generator := NewSchemaGenerator(config)
	matcher := NewPatternMatcher()

	for _, entityType := range testEntities {
		entityName := entityType.Name()

		// Step 1: Analyze the struct
		structInfo := analyzer.AnalyzeStruct(entityType)

		// Step 2: Match patterns
		patterns := matcher.MatchPatterns(structInfo)
		if len(patterns) == 0 {
			t.Errorf("No patterns matched for %s", entityName)
			continue
		}

		bestPattern := patterns[0]
		log.Printf("Best pattern for %s: %s (score: %.2f)",
			entityName, bestPattern.Pattern, bestPattern.Score)

		// Step 3: Generate UI schema
		uiSchema, err := generator.GenerateUISchema(structInfo)
		if err != nil {
			t.Errorf("Failed to generate schema for %s: %v", entityName, err)
			continue
		}

		// Step 4: Validate generated schema
		if err := validateGeneratedSchema(uiSchema); err != nil {
			t.Errorf("Generated schema validation failed for %s: %v", entityName, err)
			continue
		}

		// Step 5: Test JSON serialization
		jsonData, err := json.MarshalIndent(uiSchema, "", "  ")
		if err != nil {
			t.Errorf("Failed to serialize schema for %s: %v", entityName, err)
			continue
		}

		// Verify JSON is valid and contains expected elements
		if len(jsonData) == 0 {
			t.Errorf("Empty JSON generated for %s", entityName)
			continue
		}

		log.Printf("✓ End-to-end generation test passed for %s - generated %d bytes of JSON",
			entityName, len(jsonData))
	}
}

// validateGeneratedSchema performs validation on generated UI schemas
func validateGeneratedSchema(schema *UISchema) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if schema.Entity == "" {
		return fmt.Errorf("entity name is empty")
	}

	if schema.Title == "" {
		return fmt.Errorf("title is empty")
	}

	// Validate that at least one schema type was generated
	schemaCount := 0
	if schema.ListSchema != nil {
		schemaCount++
	}
	if schema.CreateSchema != nil {
		schemaCount++
	}
	if schema.EditSchema != nil {
		schemaCount++
	}
	if schema.DetailSchema != nil {
		schemaCount++
	}

	if schemaCount == 0 {
		return fmt.Errorf("no schemas were generated")
	}

	// Validate schema structure
	if schema.ListSchema != nil {
		if err := validatePageSchema(schema.ListSchema); err != nil {
			return fmt.Errorf("list schema validation failed: %w", err)
		}
	}

	if schema.CreateSchema != nil {
		if err := validatePageSchema(schema.CreateSchema); err != nil {
			return fmt.Errorf("create schema validation failed: %w", err)
		}
	}

	return nil
}

// validatePageSchema validates the structure of a page schema
func validatePageSchema(schema *PageSchema) error {
	if schema == nil {
		return fmt.Errorf("page schema is nil")
	}

	if schema.ID == "" {
		return fmt.Errorf("page schema ID is empty")
	}

	if schema.Title == "" {
		return fmt.Errorf("page schema title is empty")
	}

	if schema.Layout == "" {
		return fmt.Errorf("page schema layout is empty")
	}

	// Validate that components exist
	if len(schema.Components) == 0 {
		return fmt.Errorf("page schema has no components")
	}

	// Validate component structure
	for i, component := range schema.Components {
		if component.ID == "" {
			return fmt.Errorf("component %d has empty ID", i)
		}

		if component.Type == "" {
			return fmt.Errorf("component %d has empty type", i)
		}
	}

	return nil
}

// RunGenerationDemo runs a demonstration of the generation system
func RunGenerationDemo() {
	fmt.Println("=== UI Generation System Demo ===")

	// Initialize components
	analyzer := NewAnalyzer()
	matcher := NewPatternMatcher()

	config := GeneratorConfig{
		DefaultLayout:     "app",
		EnablePermissions: true,
		EnableAuditTrail:  true,
		APIPrefix:         "/api",
		TablePageSize:     20,
		EnableFiltering:   true,
		EnableSorting:     true,
		EnableBulkActions: true,
	}
	generator := NewSchemaGenerator(config)

	// Demo entities
	entities := []struct {
		Type reflect.Type
		Name string
	}{
		{reflect.TypeOf(TestUser{}), "User Management"},
		{reflect.TypeOf(TestProject{}), "Project Management"},
		{reflect.TypeOf(TestInvoice{}), "Invoice Management"},
		{reflect.TypeOf(TestEvent{}), "Event Calendar"},
	}

	for _, entity := range entities {
		fmt.Printf("🔍 Analyzing %s...\n", entity.Name)

		// Analyze structure
		structInfo := analyzer.AnalyzeStruct(entity.Type)
		fmt.Printf("   Fields: %d\n", len(structInfo.Fields))
		fmt.Printf("   CRUD Entity: %t\n", structInfo.IsCRUDEntity)
		fmt.Printf("   Has Workflow: %t\n", structInfo.HasWorkflow)
		fmt.Printf("   Has Relationships: %d\n", len(structInfo.Fields))

		// Match patterns
		patterns := matcher.MatchPatterns(structInfo)
		if len(patterns) > 0 {
			fmt.Printf("   Best Pattern: %s (score: %.1f)\n", patterns[0].Pattern, patterns[0].Score)
		}

		// Generate schema
		schema, err := generator.GenerateUISchema(structInfo)
		if err != nil {
			fmt.Printf("   ❌ Generation failed: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Generated schemas: ")
		schemaTypes := []string{}
		if schema.ListSchema != nil {
			schemaTypes = append(schemaTypes, "List")
		}
		if schema.CreateSchema != nil {
			schemaTypes = append(schemaTypes, "Create")
		}
		if schema.EditSchema != nil {
			schemaTypes = append(schemaTypes, "Edit")
		}
		if schema.DetailSchema != nil {
			schemaTypes = append(schemaTypes, "Detail")
		}
		if schema.FilterSchema != nil {
			schemaTypes = append(schemaTypes, "Filter")
		}

		fmt.Printf("%s\n", schemaTypes)
		fmt.Println()
	}

	fmt.Println("✅ Demo completed successfully!")
}

// BenchmarkGeneration benchmarks the generation performance
func BenchmarkGeneration(b *testing.B) {
	analyzer := NewAnalyzer()
	config := GeneratorConfig{
		DefaultLayout:     "app",
		EnablePermissions: true,
		EnableAuditTrail:  true,
		APIPrefix:         "/api",
		TablePageSize:     20,
	}
	generator := NewSchemaGenerator(config)

	userType := reflect.TypeOf(TestUser{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		structInfo := analyzer.AnalyzeStruct(userType)
		_, err := generator.GenerateUISchema(structInfo)
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}
