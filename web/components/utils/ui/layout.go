package ui

import (
	"context"
	"encoding/json"
	"fmt"
)

// Container component configuration
type ContainerConfig struct {
	MaxWidth     string `json:"max_width,omitempty"`
	Padding      string `json:"padding,omitempty"`
	Margin       string `json:"margin,omitempty"`
	Background   string `json:"background,omitempty"`
	Border       string `json:"border,omitempty"`
	BorderRadius string `json:"border_radius,omitempty"`
	Shadow       string `json:"shadow,omitempty"`
	Fluid        bool   `json:"fluid,omitempty"`
}

// Card component configuration
type CardConfig struct {
	Header    *CardSection `json:"header,omitempty"`
	Body      *CardSection `json:"body,omitempty"`
	Footer    *CardSection `json:"footer,omitempty"`
	Elevation int          `json:"elevation,omitempty"`
	Hoverable bool         `json:"hoverable,omitempty"`
	Clickable bool         `json:"clickable,omitempty"`
	Loading   bool         `json:"loading,omitempty"`
	Image     *CardImage   `json:"image,omitempty"`
}

// CardSection represents a section within a card
type CardSection struct {
	Content      string `json:"content,omitempty"`
	Class        string `json:"class,omitempty"`
	Padding      string `json:"padding,omitempty"`
	Background   string `json:"background,omitempty"`
	BorderBottom bool   `json:"border_bottom,omitempty"`
}

// CardImage represents an image in a card
type CardImage struct {
	Src       string   `json:"src" validate:"required"`
	Alt       string   `json:"alt,omitempty"`
	Position  Position `json:"position,omitempty"`
	Height    string   `json:"height,omitempty"`
	ObjectFit string   `json:"object_fit,omitempty"`
}

// Panel component configuration
type PanelConfig struct {
	Title         string      `json:"title,omitempty"`
	Subtitle      string      `json:"subtitle,omitempty"`
	Icon          string      `json:"icon,omitempty"`
	Collapsible   bool        `json:"collapsible,omitempty"`
	Collapsed     bool        `json:"collapsed,omitempty"`
	Closable      bool        `json:"closable,omitempty"`
	HeaderActions []Component `json:"header_actions,omitempty"`
	FooterActions []Component `json:"footer_actions,omitempty"`
}

// Tabs component configuration
type TabsConfig struct {
	Tabs         []Tab    `json:"tabs" validate:"required"`
	ActiveTab    string   `json:"active_tab,omitempty"`
	Position     Position `json:"position,omitempty"`
	Type         TabType  `json:"type,omitempty"`
	Size         Size     `json:"size,omitempty"`
	Closable     bool     `json:"closable,omitempty"`
	AddButton    bool     `json:"add_button,omitempty"`
	ScrollButton bool     `json:"scroll_button,omitempty"`
	LazyLoad     bool     `json:"lazy_load,omitempty"`
}

// Tab represents a single tab
type Tab struct {
	ID       string    `json:"id" validate:"required"`
	Label    string    `json:"label" validate:"required"`
	Icon     string    `json:"icon,omitempty"`
	Content  Component `json:"content,omitempty"` // Component is defined in types.go
	Disabled bool      `json:"disabled,omitempty"`
	Closable bool      `json:"closable,omitempty"`
	Badge    string    `json:"badge,omitempty"`
}

// TabType represents different tab styles
type TabType string

const (
	TabTypeDefault    TabType = "default"
	TabTypeCard       TabType = "card"
	TabTypeBorderless TabType = "borderless"
	TabTypePills      TabType = "pills"
)

// Modal component configuration
type ModalConfig struct {
	Title       string     `json:"title,omitempty"`
	Size        ModalSize  `json:"size,omitempty"`
	Closable    bool       `json:"closable,omitempty"`
	Backdrop    bool       `json:"backdrop,omitempty"`
	Keyboard    bool       `json:"keyboard,omitempty"`
	Focus       bool       `json:"focus,omitempty"`
	Centered    bool       `json:"centered,omitempty"`
	Scrollable  bool       `json:"scrollable,omitempty"`
	Fullscreen  bool       `json:"fullscreen,omitempty"`
	Animation   string     `json:"animation,omitempty"`
	ZIndex      int        `json:"z_index,omitempty"`
	CloseButton bool       `json:"close_button,omitempty"`
	Header      *Component `json:"header,omitempty"`
	Footer      *Component `json:"footer,omitempty"`
}

// ModalSize represents modal size options
type ModalSize string

const (
	ModalSizeSM   ModalSize = "sm"
	ModalSizeMD   ModalSize = "md"
	ModalSizeLG   ModalSize = "lg"
	ModalSizeXL   ModalSize = "xl"
	ModalSizeFull ModalSize = "full"
)

// Drawer component configuration
type DrawerConfig struct {
	Title     string     `json:"title,omitempty"`
	Position  Position   `json:"position,omitempty"` // Position is defined in types.go
	Width     string     `json:"width,omitempty"`
	Height    string     `json:"height,omitempty"`
	Closable  bool       `json:"closable,omitempty"`
	Backdrop  bool       `json:"backdrop,omitempty"`
	Keyboard  bool       `json:"keyboard,omitempty"`
	Push      bool       `json:"push,omitempty"`
	Resizable bool       `json:"resizable,omitempty"`
	MinWidth  string     `json:"min_width,omitempty"`
	MaxWidth  string     `json:"max_width,omitempty"`
	Header    *Component `json:"header,omitempty"`
	Footer    *Component `json:"footer,omitempty"`
}

// Grid system configuration
type GridConfig struct {
	Columns     int         `json:"columns,omitempty"`
	Rows        int         `json:"rows,omitempty"`
	Gap         string      `json:"gap,omitempty"`
	ColumnGap   string      `json:"column_gap,omitempty"`
	RowGap      string      `json:"row_gap,omitempty"`
	Template    string      `json:"template,omitempty"`
	Areas       []string    `json:"areas,omitempty"`
	AutoFit     bool        `json:"auto_fit,omitempty"`
	AutoFill    bool        `json:"auto_fill,omitempty"`
	MinItemSize string      `json:"min_item_size,omitempty"`
	MaxItemSize string      `json:"max_item_size,omitempty"`
	Responsive  *Responsive `json:"responsive,omitempty"`
}

// Responsive configuration for different screen sizes
type Responsive struct {
	XS *BreakpointConfig `json:"xs,omitempty"`
	SM *BreakpointConfig `json:"sm,omitempty"`
	MD *BreakpointConfig `json:"md,omitempty"`
	LG *BreakpointConfig `json:"lg,omitempty"`
	XL *BreakpointConfig `json:"xl,omitempty"`
}

// BreakpointConfig defines layout at specific breakpoints
type BreakpointConfig struct {
	Columns int    `json:"columns,omitempty"`
	Gap     string `json:"gap,omitempty"`
	Hidden  bool   `json:"hidden,omitempty"`
}

// Flex container configuration
type FlexConfig struct {
	Direction    FlexDirection    `json:"direction,omitempty"`
	Wrap         FlexWrap         `json:"wrap,omitempty"`
	Justify      FlexJustify      `json:"justify,omitempty"`
	Align        FlexAlign        `json:"align,omitempty"`
	AlignContent FlexAlignContent `json:"align_content,omitempty"`
	Gap          string           `json:"gap,omitempty"`
	Inline       bool             `json:"inline,omitempty"`
}

// Flex direction options
type FlexDirection string

const (
	FlexDirectionRow           FlexDirection = "row"
	FlexDirectionRowReverse    FlexDirection = "row-reverse"
	FlexDirectionColumn        FlexDirection = "column"
	FlexDirectionColumnReverse FlexDirection = "column-reverse"
)

// Flex wrap options
type FlexWrap string

const (
	FlexWrapNoWrap      FlexWrap = "nowrap"
	FlexWrapWrap        FlexWrap = "wrap"
	FlexWrapWrapReverse FlexWrap = "wrap-reverse"
)

// Flex justify options
type FlexJustify string

const (
	FlexJustifyStart        FlexJustify = "flex-start"
	FlexJustifyEnd          FlexJustify = "flex-end"
	FlexJustifyCenter       FlexJustify = "center"
	FlexJustifySpaceBetween FlexJustify = "space-between"
	FlexJustifySpaceAround  FlexJustify = "space-around"
	FlexJustifySpaceEvenly  FlexJustify = "space-evenly"
)

// Flex align options
type FlexAlign string

const (
	FlexAlignStart    FlexAlign = "flex-start"
	FlexAlignEnd      FlexAlign = "flex-end"
	FlexAlignCenter   FlexAlign = "center"
	FlexAlignBaseline FlexAlign = "baseline"
	FlexAlignStretch  FlexAlign = "stretch"
)

// Flex align content options
type FlexAlignContent string

const (
	FlexAlignContentStart        FlexAlignContent = "flex-start"
	FlexAlignContentEnd          FlexAlignContent = "flex-end"
	FlexAlignContentCenter       FlexAlignContent = "center"
	FlexAlignContentSpaceBetween FlexAlignContent = "space-between"
	FlexAlignContentSpaceAround  FlexAlignContent = "space-around"
	FlexAlignContentStretch      FlexAlignContent = "stretch"
)

// Layout component factories
type (
	ContainerFactory struct{}
	CardFactory      struct{}
	PanelFactory     struct{}
	TabsFactory      struct{}
	ModalFactory     struct{}
	DrawerFactory    struct{}
)

// Container factory implementation
func (f *ContainerFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var containerConfig ContainerConfig
	if err := mapToStruct(config, &containerConfig); err != nil {
		return Component{}, fmt.Errorf("invalid container config: %w", err)
	}

	component := NewComponent(ComponentContainer, generateID())
	return component.WithConfig(containerConfig).Build(), nil
}

func (f *ContainerFactory) Validate(ctx context.Context, component Component) error {
	var containerConfig ContainerConfig
	if err := json.Unmarshal(component.Config, &containerConfig); err != nil {
		return fmt.Errorf("invalid container config: %w", err)
	}
	return nil
}

func (f *ContainerFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentContainer,
		Title:       "Container",
		Description: "A layout container for organizing content",
		Properties: map[string]Property{
			"max_width": {
				Type:        "string",
				Description: "Maximum width of the container",
				Default:     "100%",
			},
			"fluid": {
				Type:        "boolean",
				Description: "Whether container should be fluid width",
				Default:     false,
			},
			"padding": {
				Type:        "string",
				Description: "Container padding",
			},
		},
		Examples: []map[string]any{
			{
				"max_width": "1200px",
				"padding":   "20px",
				"fluid":     false,
			},
		},
	}
}

// Card factory implementation
func (f *CardFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var cardConfig CardConfig
	if err := mapToStruct(config, &cardConfig); err != nil {
		return Component{}, fmt.Errorf("invalid card config: %w", err)
	}

	component := NewComponent(ComponentCard, generateID())
	return component.WithConfig(cardConfig).Build(), nil
}

func (f *CardFactory) Validate(ctx context.Context, component Component) error {
	var cardConfig CardConfig
	if err := json.Unmarshal(component.Config, &cardConfig); err != nil {
		return fmt.Errorf("invalid card config: %w", err)
	}
	return nil
}

func (f *CardFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentCard,
		Title:       "Card",
		Description: "A flexible content container",
		Properties: map[string]Property{
			"elevation": {
				Type:        "integer",
				Description: "Card shadow elevation level",
				Min:         new(float64),
				Max:         func() *float64 { v := 24.0; return &v }(),
				Default:     1,
			},
			"hoverable": {
				Type:        "boolean",
				Description: "Whether card has hover effect",
				Default:     false,
			},
			"clickable": {
				Type:        "boolean",
				Description: "Whether card is clickable",
				Default:     false,
			},
		},
		Examples: []map[string]any{
			{
				"header": map[string]any{
					"content": "Card Title",
				},
				"elevation": 2,
				"hoverable": true,
			},
		},
	}
}

// Tabs factory implementation
func (f *TabsFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var tabsConfig TabsConfig
	if err := mapToStruct(config, &tabsConfig); err != nil {
		return Component{}, fmt.Errorf("invalid tabs config: %w", err)
	}

	component := NewComponent(ComponentTabs, generateID())
	return component.WithConfig(tabsConfig).Build(), nil
}

func (f *TabsFactory) Validate(ctx context.Context, component Component) error {
	var tabsConfig TabsConfig
	if err := json.Unmarshal(component.Config, &tabsConfig); err != nil {
		return fmt.Errorf("invalid tabs config: %w", err)
	}

	if len(tabsConfig.Tabs) == 0 {
		return fmt.Errorf("tabs component must have at least one tab")
	}

	return nil
}

func (f *TabsFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentTabs,
		Title:       "Tabs",
		Description: "A tabbed interface component",
		Properties: map[string]Property{
			"tabs": {
				Type:        "array",
				Description: "Array of tab objects",
			},
			"position": {
				Type:        "string",
				Description: "Position of tab headers",
				Enum:        []string{"top", "bottom", "left", "right"},
				Default:     "top",
			},
			"type": {
				Type:        "string",
				Description: "Tab style type",
				Enum:        []string{"default", "card", "borderless", "pills"},
				Default:     "default",
			},
		},
		Required: []string{"tabs"},
		Examples: []map[string]any{
			{
				"tabs": []Tab{
					{ID: "tab1", Label: "First Tab"},
					{ID: "tab2", Label: "Second Tab"},
				},
				"type": "card",
			},
		},
	}
}

// Panel factory implementation
func (f *PanelFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var panelConfig PanelConfig
	if err := mapToStruct(config, &panelConfig); err != nil {
		return Component{}, fmt.Errorf("invalid panel config: %w", err)
	}
	component := NewComponent(ComponentPanel, generateID())
	return component.WithConfig(panelConfig).Build(), nil
}

func (f *PanelFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *PanelFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentPanel,
		Title:       "Panel",
		Description: "A collapsible panel component",
	}
}

// Modal factory implementation
func (f *ModalFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var modalConfig ModalConfig
	if err := mapToStruct(config, &modalConfig); err != nil {
		return Component{}, fmt.Errorf("invalid modal config: %w", err)
	}
	component := NewComponent(ComponentModal, generateID())
	return component.WithConfig(modalConfig).Build(), nil
}

func (f *ModalFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *ModalFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentModal,
		Title:       "Modal",
		Description: "A modal dialog component",
	}
}

// Drawer factory implementation
func (f *DrawerFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var drawerConfig DrawerConfig
	if err := mapToStruct(config, &drawerConfig); err != nil {
		return Component{}, fmt.Errorf("invalid drawer config: %w", err)
	}
	component := NewComponent(ComponentDrawer, generateID())
	return component.WithConfig(drawerConfig).Build(), nil
}

func (f *DrawerFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *DrawerFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentDrawer,
		Title:       "Drawer",
		Description: "A slide-out drawer component",
	}
}
