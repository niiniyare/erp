package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SchemaCompositionEngine manages the visual schema building process
type SchemaCompositionEngine struct {
	// Core state
	currentSchema *CompositionSchema
	history       *HistoryManager
	registry      *ComponentRegistry
	validator     *SchemaValidator

	// Real-time updates
	subscribers map[string]chan SchemaUpdate
	mu          sync.RWMutex

	// Configuration
	config EngineConfig
}

// CompositionSchema represents the current schema being built
type CompositionSchema struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Page structure
	Layout      string                 `json:"layout"`
	Title       string                 `json:"title"`
	Components  []ComponentInstance    `json:"components"`
	DataSources []DataSourceDefinition `json:"dataSources"`
	Actions     []ActionDefinition     `json:"actions"`

	// Visual builder metadata
	Canvas    CanvasMetadata `json:"canvas"`
	Variables map[string]any `json:"variables"`
	Settings  map[string]any `json:"settings"`
}

// ComponentInstance represents a component in the composition
type ComponentInstance struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	Props       map[string]any      `json:"props"`
	Layout      *LayoutRules        `json:"layout,omitempty"`
	Children    []ComponentInstance `json:"children,omitempty"`
	DataSource  string              `json:"dataSource,omitempty"`
	Permissions *PermissionRules    `json:"permissions,omitempty"`
	Conditions  []RenderCondition   `json:"conditions,omitempty"`

	// Editor-specific configuration
	Data   *DataBinding    `json:"data,omitempty"`
	Events *EventHandlers  `json:"events,omitempty"`
	Style  *ComponentStyle `json:"style,omitempty"`

	// Visual builder metadata
	Position Position `json:"position"`
	Size     Size     `json:"size"`
	ZIndex   int      `json:"zIndex"`
	Locked   bool     `json:"locked"`
	Hidden   bool     `json:"hidden"`
	Selected bool     `json:"selected"`
	ParentID string   `json:"parentId,omitempty"`
}

// CanvasMetadata stores canvas-specific settings
type CanvasMetadata struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	GridSize   int     `json:"gridSize"`
	ShowGrid   bool    `json:"showGrid"`
	SnapToGrid bool    `json:"snapToGrid"`
	Zoom       float64 `json:"zoom"`
	Background string  `json:"background"`
	ViewMode   string  `json:"viewMode"` // "design", "preview", "code"
}

// Position represents component position on canvas
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size represents component dimensions
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// SchemaUpdate represents a real-time update
type SchemaUpdate struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"userId,omitempty"`
	Data      any       `json:"data"`
}

// EngineConfig configures the composition engine
type EngineConfig struct {
	MaxHistorySize      int           `json:"maxHistorySize"`
	AutoSaveInterval    time.Duration `json:"autoSaveInterval"`
	EnableCollaboration bool          `json:"enableCollaboration"`
	ValidationLevel     string        `json:"validationLevel"` // "strict", "warning", "off"
}

// NewSchemaCompositionEngine creates a new schema composition engine
func NewSchemaCompositionEngine(config EngineConfig) *SchemaCompositionEngine {
	if config.MaxHistorySize == 0 {
		config.MaxHistorySize = 50
	}
	if config.AutoSaveInterval == 0 {
		config.AutoSaveInterval = 30 * time.Second
	}

	engine := &SchemaCompositionEngine{
		history:     NewHistoryManager(config.MaxHistorySize),
		registry:    NewComponentRegistry(),
		validator:   NewSchemaValidator(),
		subscribers: make(map[string]chan SchemaUpdate),
		config:      config,
	}

	engine.initializeDefaultSchema()
	return engine
}

// initializeDefaultSchema creates a new empty schema
func (e *SchemaCompositionEngine) initializeDefaultSchema() {
	e.currentSchema = &CompositionSchema{
		ID:          uuid.New().String(),
		Name:        "Untitled Schema",
		Description: "New schema composition",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Layout:      "app",
		Title:       "New Page",
		Components:  []ComponentInstance{},
		DataSources: []DataSourceDefinition{},
		Actions:     []ActionDefinition{},
		Canvas: CanvasMetadata{
			Width:      1200,
			Height:     800,
			GridSize:   8,
			ShowGrid:   true,
			SnapToGrid: true,
			Zoom:       1.0,
			Background: "#ffffff",
			ViewMode:   "design",
		},
		Variables: make(map[string]any),
		Settings:  make(map[string]any),
	}
}

// AddComponent adds a new component to the schema
func (e *SchemaCompositionEngine) AddComponent(componentType string, position Position, props map[string]any) (*ComponentInstance, error) {
	// Validate component type
	if !e.registry.IsValidComponent(componentType) {
		return nil, fmt.Errorf("invalid component type: %s", componentType)
	}

	// Create component instance
	component := ComponentInstance{
		ID:       uuid.New().String(),
		Type:     componentType,
		Props:    props,
		Position: position,
		Size: Size{
			Width:  200,
			Height: 100,
		},
		ZIndex:   e.getNextZIndex(),
		Selected: true,
	}

	// Apply default props from registry
	defaultProps := e.registry.GetDefaultProps(componentType)
	for key, value := range defaultProps {
		if _, exists := component.Props[key]; !exists {
			component.Props[key] = value
		}
	}

	// Add to schema
	e.mu.Lock()
	e.currentSchema.Components = append(e.currentSchema.Components, component)
	e.currentSchema.UpdatedAt = time.Now()
	e.mu.Unlock()

	// Create history entry
	e.history.Push(&HistoryEntry{
		Type:        "add_component",
		Description: fmt.Sprintf("Added %s component", componentType),
		Data:        component,
		Timestamp:   time.Now(),
	})

	// Notify subscribers
	e.notifySubscribers(SchemaUpdate{
		Type:      "component_added",
		Timestamp: time.Now(),
		Data:      component,
	})

	return &component, nil
}

// UpdateComponent updates an existing component
func (e *SchemaCompositionEngine) UpdateComponent(componentID string, updates map[string]any) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Find component
	component := e.findComponent(componentID)
	if component == nil {
		return fmt.Errorf("component not found: %s", componentID)
	}

	// Store old state for history
	oldState := *component

	// Apply updates
	for field, value := range updates {
		switch field {
		case "props":
			if props, ok := value.(map[string]any); ok {
				if component.Props == nil {
					component.Props = make(map[string]any)
				}
				for key, val := range props {
					component.Props[key] = val
				}
			}
		case "position":
			if pos, ok := value.(Position); ok {
				component.Position = pos
			}
		case "size":
			if size, ok := value.(Size); ok {
				component.Size = size
			}
		case "zIndex":
			if zIndex, ok := value.(int); ok {
				component.ZIndex = zIndex
			}
		case "locked":
			if locked, ok := value.(bool); ok {
				component.Locked = locked
			}
		case "hidden":
			if hidden, ok := value.(bool); ok {
				component.Hidden = hidden
			}
		}
	}

	e.currentSchema.UpdatedAt = time.Now()

	// Create history entry
	e.history.Push(&HistoryEntry{
		Type:        "update_component",
		Description: fmt.Sprintf("Updated component %s", component.Type),
		Data: map[string]any{
			"before": oldState,
			"after":  *component,
		},
		Timestamp: time.Now(),
	})

	// Notify subscribers
	e.notifySubscribers(SchemaUpdate{
		Type:      "component_updated",
		Timestamp: time.Now(),
		Data:      *component,
	})

	return nil
}

// DeleteComponent removes a component from the schema
func (e *SchemaCompositionEngine) DeleteComponent(componentID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Find and remove component
	for i, component := range e.currentSchema.Components {
		if component.ID == componentID {
			// Store for history
			deletedComponent := component

			// Remove from slice
			e.currentSchema.Components = append(
				e.currentSchema.Components[:i],
				e.currentSchema.Components[i+1:]...,
			)

			e.currentSchema.UpdatedAt = time.Now()

			// Create history entry
			e.history.Push(&HistoryEntry{
				Type:        "delete_component",
				Description: fmt.Sprintf("Deleted %s component", component.Type),
				Data:        deletedComponent,
				Timestamp:   time.Now(),
			})

			// Notify subscribers
			e.notifySubscribers(SchemaUpdate{
				Type:      "component_deleted",
				Timestamp: time.Now(),
				Data:      deletedComponent,
			})

			return nil
		}
	}

	return fmt.Errorf("component not found: %s", componentID)
}

// MoveComponent changes component position
func (e *SchemaCompositionEngine) MoveComponent(componentID string, newPosition Position) error {
	// Snap to grid if enabled
	if e.currentSchema.Canvas.SnapToGrid {
		gridSize := float64(e.currentSchema.Canvas.GridSize)
		newPosition.X = float64(int(newPosition.X/gridSize)) * gridSize
		newPosition.Y = float64(int(newPosition.Y/gridSize)) * gridSize
	}

	return e.UpdateComponent(componentID, map[string]any{
		"position": newPosition,
	})
}

// ResizeComponent changes component size
func (e *SchemaCompositionEngine) ResizeComponent(componentID string, newSize Size) error {
	// Snap to grid if enabled
	if e.currentSchema.Canvas.SnapToGrid {
		gridSize := float64(e.currentSchema.Canvas.GridSize)
		newSize.Width = float64(int(newSize.Width/gridSize)) * gridSize
		newSize.Height = float64(int(newSize.Height/gridSize)) * gridSize
	}

	return e.UpdateComponent(componentID, map[string]any{
		"size": newSize,
	})
}

// SelectComponents sets selection state for components
func (e *SchemaCompositionEngine) SelectComponents(componentIDs []string, clearOthers bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear other selections if requested
	if clearOthers {
		for i := range e.currentSchema.Components {
			e.currentSchema.Components[i].Selected = false
		}
	}

	// Set selection for specified components
	selectedComponents := []ComponentInstance{}
	for _, componentID := range componentIDs {
		if component := e.findComponent(componentID); component != nil {
			component.Selected = true
			selectedComponents = append(selectedComponents, *component)
		}
	}

	// Notify subscribers
	e.notifySubscribers(SchemaUpdate{
		Type:      "selection_changed",
		Timestamp: time.Now(),
		Data:      selectedComponents,
	})

	return nil
}

// Undo reverts the last action
func (e *SchemaCompositionEngine) Undo() error {
	entry := e.history.Undo()
	if entry == nil {
		return fmt.Errorf("nothing to undo")
	}

	// Apply undo operation based on entry type
	switch entry.Type {
	case "add_component":
		if component, ok := entry.Data.(ComponentInstance); ok {
			e.DeleteComponent(component.ID)
		}
	case "delete_component":
		if component, ok := entry.Data.(ComponentInstance); ok {
			e.currentSchema.Components = append(e.currentSchema.Components, component)
		}
	case "update_component":
		if data, ok := entry.Data.(map[string]any); ok {
			if before, ok := data["before"].(ComponentInstance); ok {
				// Restore previous state
				if component := e.findComponent(before.ID); component != nil {
					*component = before
				}
			}
		}
	}

	// Notify subscribers
	e.notifySubscribers(SchemaUpdate{
		Type:      "undo",
		Timestamp: time.Now(),
		Data:      entry,
	})

	return nil
}

// Redo reapplies the last undone action
func (e *SchemaCompositionEngine) Redo() error {
	entry := e.history.Redo()
	if entry == nil {
		return fmt.Errorf("nothing to redo")
	}

	// Apply redo operation (reverse of undo)
	switch entry.Type {
	case "add_component":
		if component, ok := entry.Data.(ComponentInstance); ok {
			e.currentSchema.Components = append(e.currentSchema.Components, component)
		}
	case "delete_component":
		if component, ok := entry.Data.(ComponentInstance); ok {
			e.DeleteComponent(component.ID)
		}
	case "update_component":
		if data, ok := entry.Data.(map[string]any); ok {
			if after, ok := data["after"].(ComponentInstance); ok {
				// Apply updated state
				if component := e.findComponent(after.ID); component != nil {
					*component = after
				}
			}
		}
	}

	// Notify subscribers
	e.notifySubscribers(SchemaUpdate{
		Type:      "redo",
		Timestamp: time.Now(),
		Data:      entry,
	})

	return nil
}

// GetCurrentSchema returns the current schema
func (e *SchemaCompositionEngine) GetCurrentSchema() *CompositionSchema {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent external modification
	schemaCopy := *e.currentSchema
	return &schemaCopy
}

// LoadSchema loads a schema into the engine
func (e *SchemaCompositionEngine) LoadSchema(schema *CompositionSchema) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate schema
	if err := e.validator.ValidateSchema(schema); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Store current schema for history
	if e.currentSchema != nil {
		e.history.Push(&HistoryEntry{
			Type:        "load_schema",
			Description: "Loaded new schema",
			Data:        *e.currentSchema,
			Timestamp:   time.Now(),
		})
	}

	e.currentSchema = schema
	e.currentSchema.UpdatedAt = time.Now()

	// Notify subscribers
	e.notifySubscribers(SchemaUpdate{
		Type:      "schema_loaded",
		Timestamp: time.Now(),
		Data:      *schema,
	})

	return nil
}

// ExportSchema exports the current schema to JSON
func (e *SchemaCompositionEngine) ExportSchema() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return json.MarshalIndent(e.currentSchema, "", "  ")
}

// Subscribe to real-time updates
func (e *SchemaCompositionEngine) Subscribe(ctx context.Context, subscriberID string) <-chan SchemaUpdate {
	e.mu.Lock()
	updateChan := make(chan SchemaUpdate, 100)
	e.subscribers[subscriberID] = updateChan
	e.mu.Unlock()

	// Remove subscriber when context is cancelled
	go func() {
		<-ctx.Done()
		e.mu.Lock()
		delete(e.subscribers, subscriberID)
		close(updateChan)
		e.mu.Unlock()
	}()

	return updateChan
}

// Helper methods

func (e *SchemaCompositionEngine) findComponent(componentID string) *ComponentInstance {
	for i := range e.currentSchema.Components {
		if e.currentSchema.Components[i].ID == componentID {
			return &e.currentSchema.Components[i]
		}
	}
	return nil
}

func (e *SchemaCompositionEngine) getNextZIndex() int {
	maxZ := 0
	for _, component := range e.currentSchema.Components {
		if component.ZIndex > maxZ {
			maxZ = component.ZIndex
		}
	}
	return maxZ + 1
}

func (e *SchemaCompositionEngine) notifySubscribers(update SchemaUpdate) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, ch := range e.subscribers {
		select {
		case ch <- update:
		default:
			// Channel is full, skip this update
		}
	}
}

// Additional data structures from existing schema system

type DataSourceDefinition struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Endpoint string            `json:"endpoint,omitempty"`
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

type ActionDefinition struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Label       string            `json:"label,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Target      string            `json:"target,omitempty"`
	Method      string            `json:"method,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
}

type LayoutRules struct {
	Container  string            `json:"container,omitempty"`
	Grid       *GridLayout       `json:"grid,omitempty"`
	Classes    []string          `json:"classes,omitempty"`
	Responsive map[string]string `json:"responsive,omitempty"`
}

type GridLayout struct {
	Columns    string `json:"columns"`
	Gap        string `json:"gap,omitempty"`
	ColumnSpan int    `json:"columnSpan,omitempty"`
}

type PermissionRules struct {
	RequiredRoles       []string          `json:"requiredRoles,omitempty"`
	RequiredPermissions []string          `json:"requiredPermissions,omitempty"`
	Conditions          []PermissionCheck `json:"conditions,omitempty"`
}

type PermissionCheck struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Source   string `json:"source,omitempty"`
}

type RenderCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Source   string `json:"source,omitempty"`
}

// DataBinding represents component data configuration
type DataBinding struct {
	Source     map[string]any    `json:"source,omitempty"`
	Bindings   []PropertyBinding `json:"bindings,omitempty"`
	Filters    []DataFilter      `json:"filters,omitempty"`
	Sorting    []DataSort        `json:"sorting,omitempty"`
	Pagination *PaginationConfig `json:"pagination,omitempty"`
	Cache      *CacheConfig      `json:"cache,omitempty"`
	Transform  map[string]any    `json:"transform,omitempty"`
}

type PropertyBinding struct {
	Property  string `json:"property"`
	Path      string `json:"path"`
	Transform string `json:"transform,omitempty"`
	Default   any    `json:"default,omitempty"`
}

type DataFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Logic    string `json:"logic,omitempty"`
}

type DataSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
	Priority  int    `json:"priority,omitempty"`
}

type PaginationConfig struct {
	Enabled  bool   `json:"enabled"`
	PageSize int    `json:"pageSize"`
	Strategy string `json:"strategy"`
	Total    string `json:"total,omitempty"`
}

type CacheConfig struct {
	Enabled  bool   `json:"enabled"`
	TTL      int    `json:"ttl"`
	Strategy string `json:"strategy"`
	Key      string `json:"key,omitempty"`
}

// EventHandlers represents component event configuration
type EventHandlers struct {
	Handlers   []EventHandler   `json:"handlers,omitempty"`
	Actions    []EventAction    `json:"actions,omitempty"`
	Conditions []EventCondition `json:"conditions,omitempty"`
	Validators []EventValidator `json:"validators,omitempty"`
}

type EventHandler struct {
	ID         string         `json:"id"`
	Event      string         `json:"event"`
	Target     string         `json:"target"`
	Actions    []string       `json:"actions"`
	Conditions []string       `json:"conditions,omitempty"`
	Debounce   int            `json:"debounce,omitempty"`
	Throttle   int            `json:"throttle,omitempty"`
	Once       bool           `json:"once,omitempty"`
	Passive    bool           `json:"passive,omitempty"`
	Capture    bool           `json:"capture,omitempty"`
	Enabled    bool           `json:"enabled"`
	Options    map[string]any `json:"options,omitempty"`
}

type EventAction struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Target      string         `json:"target,omitempty"`
	Config      map[string]any `json:"config"`
	Async       bool           `json:"async,omitempty"`
	RetryCount  int            `json:"retryCount,omitempty"`
	Timeout     int            `json:"timeout,omitempty"`
	OnSuccess   []string       `json:"onSuccess,omitempty"`
	OnError     []string       `json:"onError,omitempty"`
	OnComplete  []string       `json:"onComplete,omitempty"`
	Enabled     bool           `json:"enabled"`
	Description string         `json:"description,omitempty"`
}

type EventCondition struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Expression  string         `json:"expression,omitempty"`
	Config      map[string]any `json:"config"`
	Invert      bool           `json:"invert,omitempty"`
	Description string         `json:"description,omitempty"`
}

type EventValidator struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Field    string         `json:"field"`
	Config   map[string]any `json:"config"`
	Message  string         `json:"message"`
	Severity string         `json:"severity"`
}

// ComponentStyle represents component styling configuration
type ComponentStyle struct {
	Classes      []string                   `json:"classes,omitempty"`
	InlineStyles map[string]string          `json:"inlineStyles,omitempty"`
	Tokens       map[string]any             `json:"tokens,omitempty"`
	Responsive   map[string]ResponsiveStyle `json:"responsive,omitempty"`
	States       map[string]StateStyle      `json:"states,omitempty"`
	Animation    *AnimationConfig           `json:"animation,omitempty"`
	Theme        string                     `json:"theme,omitempty"`
}

type ResponsiveStyle struct {
	Classes []string          `json:"classes"`
	Styles  map[string]string `json:"styles"`
}

type StateStyle struct {
	Classes []string          `json:"classes"`
	Styles  map[string]string `json:"styles"`
}

type AnimationConfig struct {
	Type       string         `json:"type"`
	Duration   string         `json:"duration"`
	Timing     string         `json:"timing"`
	Delay      string         `json:"delay,omitempty"`
	Iterations string         `json:"iterations,omitempty"`
	Direction  string         `json:"direction,omitempty"`
	Config     map[string]any `json:"config"`
}
