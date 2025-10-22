package ui

import (
	"context"
	"encoding/json"
	"fmt"
)

// ============================================================================
// ORGANISM COMPONENT EXAMPLES
// ============================================================================
// This file demonstrates practical usage of Modal, Table, and Tree organisms
// with the component registry system.

// CreateModalExample creates a modal component with typical configuration
func CreateModalExample(registry ComponentRegistry, title, content string) (Component, error) {
	config := map[string]any{
		"title":        title,
		"size":         ModalSizeMD,
		"closable":     true,
		"backdrop":     true,
		"keyboard":     true,
		"centered":     true,
		"close_button": true,
	}

	return registry.Create(context.Background(), ComponentModal, config)
}

// CreateTableExample creates a data table with columns and pagination
func CreateTableExample(registry ComponentRegistry) (Component, error) {
	columns := []TableColumn{
		{
			Key:        "id",
			Title:      "ID",
			DataType:   DataTypeNumber,
			Width:      "80px",
			Sortable:   true,
			Filterable: false,
		},
		{
			Key:        "name",
			Title:      "Name",
			DataType:   DataTypeText,
			Sortable:   true,
			Filterable: true,
			Searchable: true,
		},
		{
			Key:        "email",
			Title:      "Email",
			DataType:   DataTypeEmail,
			Sortable:   true,
			Filterable: true,
			Searchable: true,
		},
		{
			Key:        "status",
			Title:      "Status",
			DataType:   DataTypeBadge,
			Sortable:   true,
			Filterable: true,
		},
		{
			Key:      "actions",
			Title:    "Actions",
			DataType: DataTypeActions,
			Width:    "120px",
		},
	}

	config := map[string]any{
		"columns": columns,
		"pagination": map[string]any{
			"page_size":         10,
			"show_sizer":        true,
			"show_quick_jumper": true,
			"show_total":        true,
		},
		"sorting": map[string]any{
			"default_sort": []SortOrder{
				{Column: "name", Direction: SortAsc},
			},
			"multiple": true,
		},
		"filtering": map[string]any{
			"remote": false,
		},
		"search": map[string]any{
			"placeholder": "Search users...",
			"columns":     []string{"name", "email"},
			"debounce":    300,
		},
		"selection": map[string]any{
			"type": SelectionCheckbox,
		},
		"striped":     true,
		"bordered":    true,
		"hoverable":   true,
		"responsive":  true,
		"data_source": "/api/users",
	}

	return registry.Create(context.Background(), ComponentTable, config)
}

// CreateTreeExample creates a hierarchical tree view
func CreateTreeExample(registry ComponentRegistry) (Component, error) {
	treeData := []TreeNode{
		{
			Key:   "root",
			Title: "Root Folder",
			Icon:  "folder",
			Children: []TreeNode{
				{
					Key:   "documents",
					Title: "Documents",
					Icon:  "folder",
					Children: []TreeNode{
						{
							Key:    "doc1",
							Title:  "Report.pdf",
							Icon:   "file-text",
							IsLeaf: true,
						},
						{
							Key:    "doc2",
							Title:  "Presentation.pptx",
							Icon:   "file-text",
							IsLeaf: true,
						},
					},
				},
				{
					Key:   "images",
					Title: "Images",
					Icon:  "folder",
					Children: []TreeNode{
						{
							Key:    "img1",
							Title:  "photo1.jpg",
							Icon:   "image",
							IsLeaf: true,
						},
						{
							Key:    "img2",
							Title:  "photo2.png",
							Icon:   "image",
							IsLeaf: true,
						},
					},
				},
			},
		},
		{
			Key:   "trash",
			Title: "Trash",
			Icon:  "trash",
			Children: []TreeNode{
				{
					Key:    "deleted1",
					Title:  "old_file.txt",
					Icon:   "file",
					IsLeaf: true,
				},
			},
		},
	}

	config := map[string]any{
		"data":             treeData,
		"checkable":        true,
		"selectable":       true,
		"multiple":         true,
		"expandable":       true,
		"default_expanded": []string{"root", "documents"},
		"show_line":        true,
		"show_icon":        true,
		"draggable":        false,
		"virtual_scroll":   false,
		"height":           "400px",
	}

	return registry.Create(context.Background(), ComponentTree, config)
}

// CreateFormModalExample creates a modal containing a form
func CreateFormModalExample(registry ComponentRegistry) (Component, error) {
	// Create the modal first
	modal, err := CreateModalExample(registry, "User Form", "")
	if err != nil {
		return Component{}, fmt.Errorf("failed to create modal: %w", err)
	}

	// Create form components to go inside the modal
	nameInput, err := registry.Create(context.Background(), ComponentInput, map[string]any{
		"label":       "Full Name",
		"placeholder": "Enter your full name",
		"required":    true,
		"input_type":  InputText,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create name input: %w", err)
	}

	emailInput, err := registry.Create(context.Background(), ComponentInput, map[string]any{
		"label":       "Email Address",
		"placeholder": "Enter your email",
		"required":    true,
		"input_type":  InputEmail,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create email input: %w", err)
	}

	submitButton, err := registry.Create(context.Background(), ComponentButton, map[string]any{
		"text":    "Submit",
		"variant": VariantPrimary,
		"type":    ButtonSubmit,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create submit button: %w", err)
	}

	cancelButton, err := registry.Create(context.Background(), ComponentButton, map[string]any{
		"text":    "Cancel",
		"variant": VariantSecondary,
		"type":    ButtonButton,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create cancel button: %w", err)
	}

	// Create a form to contain the inputs
	form, err := registry.Create(context.Background(), ComponentForm, map[string]any{
		"action": "/api/users",
		"method": "POST",
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create form: %w", err)
	}

	// Add form fields as children
	form.Children = []Component{nameInput, emailInput, submitButton, cancelButton}

	// Add form as child of modal
	modal.Children = []Component{form}

	return modal, nil
}

// CreateDataTableWithActionsExample creates a table with row actions
func CreateDataTableWithActionsExample(registry ComponentRegistry) (Component, error) {
	table, err := CreateTableExample(registry)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create base table: %w", err)
	}

	// Add row actions to the existing table config
	var tableConfig TableConfig
	if err := json.Unmarshal(table.Config, &tableConfig); err != nil {
		return Component{}, fmt.Errorf("failed to parse table config: %w", err)
	}

	// Define row actions
	tableConfig.RowActions = []Action{
		{
			Key:     "edit",
			Label:   "Edit",
			Icon:    "edit",
			Type:    ActionButton,
			Variant: VariantPrimary,
			OnClick: "editUser",
		},
		{
			Key:     "delete",
			Label:   "Delete",
			Icon:    "trash",
			Type:    ActionButton,
			Variant: VariantDanger,
			OnClick: "deleteUser",
			Confirm: &ConfirmDialog{
				Title:       "Confirm Deletion",
				Description: "Are you sure you want to delete this user?",
				OkText:      "Delete",
				CancelText:  "Cancel",
			},
		},
	}

	// Define table-level actions
	tableConfig.Actions = []Action{
		{
			Key:     "add",
			Label:   "Add User",
			Icon:    "plus",
			Type:    ActionButton,
			Variant: VariantPrimary,
			OnClick: "addUser",
		},
		{
			Key:     "export",
			Label:   "Export",
			Icon:    "download",
			Type:    ActionDropdown,
			Variant: VariantSecondary,
		},
	}

	// Update the table configuration
	if configBytes, err := json.Marshal(tableConfig); err == nil {
		table.Config = configBytes
	}

	return table, nil
}

// CreateFileManagerExample creates a complete file manager using tree and table
func CreateFileManagerExample(registry ComponentRegistry) (Component, error) {
	// Create container to hold both tree and table
	container, err := registry.Create(context.Background(), ComponentContainer, map[string]any{
		"fluid": true,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create container: %w", err)
	}

	// Create file tree (left panel)
	tree, err := CreateTreeExample(registry)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create tree: %w", err)
	}

	// Create file table (right panel)
	fileColumns := []TableColumn{
		{
			Key:      "name",
			Title:    "Name",
			DataType: DataTypeText,
			Sortable: true,
		},
		{
			Key:      "size",
			Title:    "Size",
			DataType: DataTypeText,
			Sortable: true,
		},
		{
			Key:      "modified",
			Title:    "Modified",
			DataType: DataTypeDateTime,
			Sortable: true,
		},
		{
			Key:      "actions",
			Title:    "Actions",
			DataType: DataTypeActions,
		},
	}

	fileTable, err := registry.Create(context.Background(), ComponentTable, map[string]any{
		"columns":     fileColumns,
		"data_source": "/api/files",
		"pagination": map[string]any{
			"page_size": 20,
		},
		"striped":   true,
		"hoverable": true,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create file table: %w", err)
	}

	// Add both components as children of the container
	container.Children = []Component{tree, fileTable}

	return container, nil
}

// ValidateOrganismExamples validates all the example components
func ValidateOrganismExamples(registry ComponentRegistry) error {
	examples := []func(ComponentRegistry) (Component, error){
		func(r ComponentRegistry) (Component, error) {
			return CreateModalExample(r, "Test Modal", "Test content")
		},
		CreateTableExample,
		CreateTreeExample,
		CreateFormModalExample,
		CreateDataTableWithActionsExample,
		CreateFileManagerExample,
	}

	for i, createExample := range examples {
		component, err := createExample(registry)
		if err != nil {
			return fmt.Errorf("example %d failed creation: %w", i+1, err)
		}

		if err := registry.Validate(context.Background(), component); err != nil {
			return fmt.Errorf("example %d failed validation: %w", i+1, err)
		}
	}

	return nil
}

// GetOrganismUsageStatistics returns usage statistics for organism components
func GetOrganismUsageStatistics(registry ComponentRegistry) map[ComponentType]int {
	stats := make(map[ComponentType]int)
	
	// Track which organism components are registered
	organismTypes := []ComponentType{
		ComponentModal,
		ComponentTable,
		ComponentTree,
		ComponentTabs,
		ComponentCard,
		ComponentDrawer,
	}

	for _, componentType := range organismTypes {
		if registry.IsRegistered(componentType) {
			stats[componentType] = 1
		} else {
			stats[componentType] = 0
		}
	}

	return stats
}

// DemonstrateOrganismComposition shows how to compose organisms with other components
func DemonstrateOrganismComposition(registry ComponentRegistry) (Component, error) {
	// Create a dashboard layout with multiple organisms
	dashboard, err := registry.Create(context.Background(), ComponentContainer, map[string]any{
		"fluid":   true,
		"padding": "24px",
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create dashboard container: %w", err)
	}

	// Create a card containing a table
	statsCard, err := registry.Create(context.Background(), ComponentCard, map[string]any{
		"header": map[string]any{
			"content": "User Statistics",
		},
		"elevation": 2,
		"hoverable": true,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create stats card: %w", err)
	}

	// Add table to the card
	userTable, err := CreateTableExample(registry)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create user table: %w", err)
	}
	statsCard.Children = []Component{userTable}

	// Create a card containing a tree
	navCard, err := registry.Create(context.Background(), ComponentCard, map[string]any{
		"header": map[string]any{
			"content": "Navigation",
		},
		"elevation": 2,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create nav card: %w", err)
	}

	// Add tree to the card
	navTree, err := CreateTreeExample(registry)
	if err != nil {
		return Component{}, fmt.Errorf("failed to create nav tree: %w", err)
	}
	navCard.Children = []Component{navTree}

	// Create tabs to organize the content
	tabs, err := registry.Create(context.Background(), ComponentTabs, map[string]any{
		"tabs": []Tab{
			{
				ID:    "users",
				Label: "Users",
				Icon:  "users",
				Content: statsCard,
			},
			{
				ID:    "files",
				Label: "Files",
				Icon:  "folder",
				Content: navCard,
			},
		},
		"position": PositionTop,
		"type":     TabTypeCard,
	})
	if err != nil {
		return Component{}, fmt.Errorf("failed to create tabs: %w", err)
	}

	// Add tabs to dashboard
	dashboard.Children = []Component{tabs}

	return dashboard, nil
}