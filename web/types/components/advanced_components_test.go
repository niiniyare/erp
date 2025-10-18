// Test file for advanced component schemas
package components

import (
	"encoding/json"
	"testing"
)

// TestArrayControlSchema tests the ArrayControlSchema creation and validation
func TestArrayControlSchema(t *testing.T) {
	// Test factory creation
	arrayControl := NewArrayControl("addresses", []string{"address_form"})
	
	if arrayControl.Type != "input-array" {
		t.Errorf("Expected type 'input-array', got %s", arrayControl.Type)
	}
	
	if arrayControl.Name != "addresses" {
		t.Errorf("Expected name 'addresses', got %s", arrayControl.Name)
	}
	
	// Test validation
	err := arrayControl.Validate()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
	
	// Test JSON serialization
	jsonStr, err := arrayControl.ToJSON()
	if err != nil {
		t.Errorf("JSON serialization failed: %v", err)
	}
	
	// Test JSON deserialization
	_, err = ArrayControlFromJSON(jsonStr)
	if err != nil {
		t.Errorf("JSON deserialization failed: %v", err)
	}
}

// TestFileControlSchema tests the FileControlSchema creation and validation
func TestFileControlSchema(t *testing.T) {
	// Test basic file control
	fileControl := NewFileControl("avatar")
	
	if fileControl.Type != "input-file" {
		t.Errorf("Expected type 'input-file', got %s", fileControl.Type)
	}
	
	if fileControl.Name != "avatar" {
		t.Errorf("Expected name 'avatar', got %s", fileControl.Name)
	}
	
	// Test validation
	err := fileControl.Validate()
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
	
	// Test image file control with cropping
	imageControl := NewImageFileControl("profile_pic", 1.0) // Square aspect ratio
	
	if imageControl.Crop == nil {
		t.Error("Expected crop config for image control")
	}
	
	if imageControl.Crop.AspectRatio != 1.0 {
		t.Errorf("Expected aspect ratio 1.0, got %f", imageControl.Crop.AspectRatio)
	}
	
	// Test validation for image control
	err = imageControl.Validate()
	if err != nil {
		t.Errorf("Image control validation failed: %v", err)
	}
}

// TestEditorControlSchema tests the EditorControlSchema creation and validation
func TestEditorControlSchema(t *testing.T) {
	// Test code editor
	codeEditor := NewCodeEditor("source_code", LanguageJavaScript)
	
	if codeEditor.Type != "editor" {
		t.Errorf("Expected type 'editor', got %s", codeEditor.Type)
	}
	
	if codeEditor.Language != LanguageJavaScript {
		t.Errorf("Expected language 'javascript', got %s", codeEditor.Language)
	}
	
	// Test validation
	err := codeEditor.Validate()
	if err != nil {
		t.Errorf("Code editor validation failed: %v", err)
	}
	
	// Test rich text editor
	richEditor := NewRichTextEditor("content")
	
	if richEditor.EditorType != EditorTypeTinyMCE {
		t.Errorf("Expected editor type 'tinymce', got %s", richEditor.EditorType)
	}
	
	// Test validation
	err = richEditor.Validate()
	if err != nil {
		t.Errorf("Rich text editor validation failed: %v", err)
	}
	
	// Test JSON editor
	jsonEditor := NewJSONEditor("config")
	
	if jsonEditor.Language != LanguageJSON {
		t.Errorf("Expected language 'json', got %s", jsonEditor.Language)
	}
	
	// Test validation
	err = jsonEditor.Validate()
	if err != nil {
		t.Errorf("JSON editor validation failed: %v", err)
	}
}

// TestWizardSchema tests the WizardSchema creation and validation
func TestWizardSchema(t *testing.T) {
	// Create test steps
	step1 := NewWizardStep("Personal Info", []any{map[string]string{"type": "input-text", "name": "name"}})
	step2 := NewWizardStep("Contact Info", []any{map[string]string{"type": "input-email", "name": "email"}})
	step3 := NewWizardStep("Review", []any{map[string]string{"type": "static", "name": "review"}})
	
	// Test wizard creation
	wizard := NewThreeStepWizard(step1, step2, step3)
	
	if wizard.Type != "wizard" {
		t.Errorf("Expected type 'wizard', got %s", wizard.Type)
	}
	
	if len(wizard.Steps) != 3 {
		t.Errorf("Expected 3 steps, got %d", len(wizard.Steps))
	}
	
	// Test validation
	err := wizard.Validate()
	if err != nil {
		t.Errorf("Wizard validation failed: %v", err)
	}
	
	// Test vertical wizard
	verticalWizard := NewVerticalWizard([]WizardStepSchema{step1, step2})
	
	if verticalWizard.Mode != WizardModeVertical {
		t.Errorf("Expected vertical mode, got %s", verticalWizard.Mode)
	}
}

// TestComboControlSchema tests the ComboControlSchema creation and validation
func TestComboControlSchema(t *testing.T) {
	// Create test sub-controls
	subControls := []ComboSubControl{
		NewComboSubControl("input-text", "first_name", "First Name"),
		NewComboSubControl("input-text", "last_name", "Last Name"),
	}
	
	// Test combo creation
	combo := NewComboControl("person", subControls)
	
	if combo.Type != "combo" {
		t.Errorf("Expected type 'combo', got %s", combo.Type)
	}
	
	if combo.Name != "person" {
		t.Errorf("Expected name 'person', got %s", combo.Name)
	}
	
	if len(combo.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(combo.Items))
	}
	
	// Test validation
	err := combo.Validate()
	if err != nil {
		t.Errorf("Combo validation failed: %v", err)
	}
	
	// Test multiple combo
	multipleCombo := NewMultipleComboControl("contacts", subControls)
	
	if !multipleCombo.Multiple {
		t.Error("Expected multiple to be true")
	}
	
	// Test conditional combo
	conditions := []ComboCondition{
		NewComboCondition("data.type === 'person'", "Person", []any{
			map[string]string{"type": "input-text", "name": "name"},
		}),
		NewComboCondition("data.type === 'company'", "Company", []any{
			map[string]string{"type": "input-text", "name": "company_name"},
		}),
	}
	
	conditionalCombo := NewConditionalComboControl("entity", conditions)
	
	if !conditionalCombo.TypeSwitchable {
		t.Error("Expected typeSwitchable to be true")
	}
	
	if len(conditionalCombo.Conditions) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(conditionalCombo.Conditions))
	}
}

// TestComponentEnums tests that all enums have valid values
func TestComponentEnums(t *testing.T) {
	// Test ArrayControl enums
	validSizes := []ArrayControlSize{ArraySizeXS, ArraySizeSM, ArraySizeMD, ArraySizeLG, ArraySizeFull}
	for _, size := range validSizes {
		if string(size) == "" && size != ArraySizeXS { // Only XS can be empty since it's the first
			t.Errorf("Empty size enum: %s", size)
		}
	}
	
	// Test FileControl enums
	validCaptureModes := []FileCaptureMode{FileCaptureUser, FileCaptureEnvironment, FileCaptureCamera, FileCaptureCamcorder, FileCaptureFile}
	for _, mode := range validCaptureModes {
		if string(mode) == "" {
			t.Errorf("Empty capture mode enum: %s", mode)
		}
	}
	
	// Test Editor enums
	validLanguages := []EditorLanguage{LanguageJavaScript, LanguageJSON, LanguageHTML, LanguageCSS, LanguageMarkdown}
	for _, lang := range validLanguages {
		if string(lang) == "" {
			t.Errorf("Empty language enum: %s", lang)
		}
	}
	
	// Test Wizard enums
	if WizardModeHorizontal == "" || WizardModeVertical == "" {
		t.Error("Empty wizard mode enum")
	}
}

// TestComponentJSONSerialization tests that all components can serialize to/from JSON properly
func TestComponentJSONSerialization(t *testing.T) {
	// Test ArrayControl JSON
	arrayControl := NewArrayControl("test", []string{"test"})
	data, err := json.Marshal(arrayControl)
	if err != nil {
		t.Errorf("ArrayControl JSON marshal failed: %v", err)
	}
	
	var unmarshaledArray ArrayControlSchema
	err = json.Unmarshal(data, &unmarshaledArray)
	if err != nil {
		t.Errorf("ArrayControl JSON unmarshal failed: %v", err)
	}
	
	// Test FileControl JSON
	fileControl := NewFileControl("test")
	data, err = json.Marshal(fileControl)
	if err != nil {
		t.Errorf("FileControl JSON marshal failed: %v", err)
	}
	
	var unmarshaledFile FileControlSchema
	err = json.Unmarshal(data, &unmarshaledFile)
	if err != nil {
		t.Errorf("FileControl JSON unmarshal failed: %v", err)
	}
	
	// Test EditorControl JSON
	editorControl := NewCodeEditor("test", LanguageJavaScript)
	data, err = json.Marshal(editorControl)
	if err != nil {
		t.Errorf("EditorControl JSON marshal failed: %v", err)
	}
	
	var unmarshaledEditor EditorControlSchema
	err = json.Unmarshal(data, &unmarshaledEditor)
	if err != nil {
		t.Errorf("EditorControl JSON unmarshal failed: %v", err)
	}
}