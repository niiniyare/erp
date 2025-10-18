package css

// Enums for common CSS values
type Display string

const (
	DisplayBlock       Display = "block"
	DisplayInline      Display = "inline"
	DisplayInlineBlock Display = "inline-block"
	DisplayFlex        Display = "flex"
	DisplayGrid        Display = "grid"
	DisplayNone        Display = "none"
)

type Position string

const (
	PositionStatic   Position = "static"
	PositionRelative Position = "relative"
	PositionAbsolute Position = "absolute"
	PositionFixed    Position = "fixed"
	PositionSticky   Position = "sticky"
)

type BoxSizing string

const (
	BoxSizingBorderBox  BoxSizing = "border-box"
	BoxSizingContentBox BoxSizing = "content-box"
)

type Overflow string

const (
	OverflowVisible Overflow = "visible"
	OverflowHidden  Overflow = "hidden"
	OverflowScroll  Overflow = "scroll"
	OverflowAuto    Overflow = "auto"
)

type FlexDirection string

const (
	FlexDirectionRow           FlexDirection = "row"
	FlexDirectionRowReverse    FlexDirection = "row-reverse"
	FlexDirectionColumn        FlexDirection = "column"
	FlexDirectionColumnReverse FlexDirection = "column-reverse"
)

type JustifyContent string

const (
	JustifyContentFlexStart    JustifyContent = "flex-start"
	JustifyContentFlexEnd      JustifyContent = "flex-end"
	JustifyContentCenter       JustifyContent = "center"
	JustifyContentSpaceBetween JustifyContent = "space-between"
	JustifyContentSpaceAround  JustifyContent = "space-around"
)

type AlignItems string

const (
	AlignItemsFlexStart AlignItems = "flex-start"
	AlignItemsFlexEnd   AlignItems = "flex-end"
	AlignItemsCenter    AlignItems = "center"
	AlignItemsStretch   AlignItems = "stretch"
	AlignItemsBaseline  AlignItems = "baseline"
)

type TextAlign string

const (
	TextAlignLeft    TextAlign = "left"
	TextAlignRight   TextAlign = "right"
	TextAlignCenter  TextAlign = "center"
	TextAlignJustify TextAlign = "justify"
)

type FontWeight string

const (
	FontWeightNormal FontWeight = "normal"
	FontWeightBold   FontWeight = "bold"
	FontWeight100    FontWeight = "100"
	FontWeight200    FontWeight = "200"
	FontWeight300    FontWeight = "300"
	FontWeight400    FontWeight = "400"
	FontWeight500    FontWeight = "500"
	FontWeight600    FontWeight = "600"
	FontWeight700    FontWeight = "700"
	FontWeight800    FontWeight = "800"
	FontWeight900    FontWeight = "900"
)

type Visibility string

const (
	VisibilityVisible  Visibility = "visible"
	VisibilityHidden   Visibility = "hidden"
	VisibilityCollapse Visibility = "collapse"
)

type BorderStyle string

const (
	BorderStyleNone   BorderStyle = "none"
	BorderStyleHidden BorderStyle = "hidden"
	BorderStyleDotted BorderStyle = "dotted"
	BorderStyleDashed BorderStyle = "dashed"
	BorderStyleSolid  BorderStyle = "solid"
	BorderStyleDouble BorderStyle = "double"
	BorderStyleGroove BorderStyle = "groove"
	BorderStyleRidge  BorderStyle = "ridge"
	BorderStyleInset  BorderStyle = "inset"
	BorderStyleOutset BorderStyle = "outset"
)

type Cursor string

const (
	CursorAuto       Cursor = "auto"
	CursorDefault    Cursor = "default"
	CursorPointer    Cursor = "pointer"
	CursorText       Cursor = "text"
	CursorMove       Cursor = "move"
	CursorNotAllowed Cursor = "not-allowed"
	CursorHelp       Cursor = "help"
)

// Layout Properties
type LayoutProperties struct {
	Display  Display  `json:"display,omitempty"`
	Position Position `json:"position,omitempty"`
	Top      string   `json:"top,omitempty"`
	Right    string   `json:"right,omitempty"`
	Bottom   string   `json:"bottom,omitempty"`
	Left     string   `json:"left,omitempty"`
	ZIndex   string   `json:"z-index,omitempty"`
	Float    string   `json:"float,omitempty"`
	Clear    string   `json:"clear,omitempty"`
}

// Box Model Properties
type BoxModelProperties struct {
	Width     string    `json:"width,omitempty"`
	Height    string    `json:"height,omitempty"`
	MinWidth  string    `json:"min-width,omitempty"`
	MinHeight string    `json:"min-height,omitempty"`
	MaxWidth  string    `json:"max-width,omitempty"`
	MaxHeight string    `json:"max-height,omitempty"`
	BoxSizing BoxSizing `json:"box-sizing,omitempty"`
}

// Margin Properties
type MarginProperties struct {
	Margin      string `json:"margin,omitempty"`
	Top         string `json:"margin-top,omitempty"`
	Right       string `json:"margin-right,omitempty"`
	Bottom      string `json:"margin-bottom,omitempty"`
	Left        string `json:"margin-left,omitempty"`
	Block       string `json:"margin-block,omitempty"`
	BlockStart  string `json:"margin-block-start,omitempty"`
	BlockEnd    string `json:"margin-block-end,omitempty"`
	Inline      string `json:"margin-inline,omitempty"`
	InlineStart string `json:"margin-inline-start,omitempty"`
	InlineEnd   string `json:"margin-inline-end,omitempty"`
}

// Padding Properties
type PaddingProperties struct {
	Padding     string `json:"padding,omitempty"`
	Top         string `json:"padding-top,omitempty"`
	Right       string `json:"padding-right,omitempty"`
	Bottom      string `json:"padding-bottom,omitempty"`
	Left        string `json:"padding-left,omitempty"`
	Block       string `json:"padding-block,omitempty"`
	BlockStart  string `json:"padding-block-start,omitempty"`
	BlockEnd    string `json:"padding-block-end,omitempty"`
	Inline      string `json:"padding-inline,omitempty"`
	InlineStart string `json:"padding-inline-start,omitempty"`
	InlineEnd   string `json:"padding-inline-end,omitempty"`
}

// Border Properties
type BorderProperties struct {
	Border string      `json:"border,omitempty"`
	Width  string      `json:"border-width,omitempty"`
	Style  BorderStyle `json:"border-style,omitempty"`
	Color  string      `json:"border-color,omitempty"`
	Radius string      `json:"border-radius,omitempty"`

	Top            string      `json:"border-top,omitempty"`
	TopWidth       string      `json:"border-top-width,omitempty"`
	TopStyle       BorderStyle `json:"border-top-style,omitempty"`
	TopColor       string      `json:"border-top-color,omitempty"`
	TopLeftRadius  string      `json:"border-top-left-radius,omitempty"`
	TopRightRadius string      `json:"border-top-right-radius,omitempty"`

	Right      string      `json:"border-right,omitempty"`
	RightWidth string      `json:"border-right-width,omitempty"`
	RightStyle BorderStyle `json:"border-right-style,omitempty"`
	RightColor string      `json:"border-right-color,omitempty"`

	Bottom            string      `json:"border-bottom,omitempty"`
	BottomWidth       string      `json:"border-bottom-width,omitempty"`
	BottomStyle       BorderStyle `json:"border-bottom-style,omitempty"`
	BottomColor       string      `json:"border-bottom-color,omitempty"`
	BottomLeftRadius  string      `json:"border-bottom-left-radius,omitempty"`
	BottomRightRadius string      `json:"border-bottom-right-radius,omitempty"`

	Left      string      `json:"border-left,omitempty"`
	LeftWidth string      `json:"border-left-width,omitempty"`
	LeftStyle BorderStyle `json:"border-left-style,omitempty"`
	LeftColor string      `json:"border-left-color,omitempty"`

	Block       string `json:"border-block,omitempty"`
	BlockStart  string `json:"border-block-start,omitempty"`
	BlockEnd    string `json:"border-block-end,omitempty"`
	Inline      string `json:"border-inline,omitempty"`
	InlineStart string `json:"border-inline-start,omitempty"`
	InlineEnd   string `json:"border-inline-end,omitempty"`
}

// Background Properties
type BackgroundProperties struct {
	Background string `json:"background,omitempty"`
	Color      string `json:"background-color,omitempty"`
	Image      string `json:"background-image,omitempty"`
	Repeat     string `json:"background-repeat,omitempty"`
	Position   string `json:"background-position,omitempty"`
	PositionX  string `json:"background-position-x,omitempty"`
	PositionY  string `json:"background-position-y,omitempty"`
	Size       string `json:"background-size,omitempty"`
	Origin     string `json:"background-origin,omitempty"`
	Clip       string `json:"background-clip,omitempty"`
	Attachment string `json:"background-attachment,omitempty"`
	BlendMode  string `json:"background-blend-mode,omitempty"`
}

// Typography Properties
type TypographyProperties struct {
	FontFamily          string     `json:"font-family,omitempty"`
	FontSize            string     `json:"font-size,omitempty"`
	FontWeight          FontWeight `json:"font-weight,omitempty"`
	FontStyle           string     `json:"font-style,omitempty"`
	FontVariant         string     `json:"font-variant,omitempty"`
	FontStretch         string     `json:"font-stretch,omitempty"`
	LineHeight          string     `json:"line-height,omitempty"`
	LetterSpacing       string     `json:"letter-spacing,omitempty"`
	WordSpacing         string     `json:"word-spacing,omitempty"`
	TextAlign           TextAlign  `json:"text-align,omitempty"`
	Decoration          string     `json:"text-decoration,omitempty"`
	DecorationColor     string     `json:"text-decoration-color,omitempty"`
	DecorationLine      string     `json:"text-decoration-line,omitempty"`
	DecorationStyle     string     `json:"text-decoration-style,omitempty"`
	DecorationThickness string     `json:"text-decoration-thickness,omitempty"`
	TextTransform       string     `json:"text-transform,omitempty"`
	TextIndent          string     `json:"text-indent,omitempty"`
	TextShadow          string     `json:"text-shadow,omitempty"`
	Color               string     `json:"color,omitempty"`
	WhiteSpace          string     `json:"white-space,omitempty"`
	WordWrap            string     `json:"word-wrap,omitempty"`
	WordBreak           string     `json:"word-break,omitempty"`
	OverflowWrap        string     `json:"overflow-wrap,omitempty"`
}

// Flexbox Properties
type FlexboxProperties struct {
	Direction      FlexDirection  `json:"flex-direction,omitempty"`
	Wrap           string         `json:"flex-wrap,omitempty"`
	Flow           string         `json:"flex-flow,omitempty"`
	JustifyContent JustifyContent `json:"justify-content,omitempty"`
	AlignItems     AlignItems     `json:"align-items,omitempty"`
	AlignContent   string         `json:"align-content,omitempty"`
	Flex           string         `json:"flex,omitempty"`
	Grow           string         `json:"flex-grow,omitempty"`
	Shrink         string         `json:"flex-shrink,omitempty"`
	Basis          string         `json:"flex-basis,omitempty"`
	AlignSelf      string         `json:"align-self,omitempty"`
	Order          string         `json:"order,omitempty"`
}

// Grid Properties
type GridProperties struct {
	Template        string `json:"grid-template,omitempty"`
	TemplateColumns string `json:"grid-template-columns,omitempty"`
	TemplateRows    string `json:"grid-template-rows,omitempty"`
	TemplateAreas   string `json:"grid-template-areas,omitempty"`
	AutoColumns     string `json:"grid-auto-columns,omitempty"`
	AutoRows        string `json:"grid-auto-rows,omitempty"`
	AutoFlow        string `json:"grid-auto-flow,omitempty"`
	Grid            string `json:"grid,omitempty"`
	Area            string `json:"grid-area,omitempty"`
	Column          string `json:"grid-column,omitempty"`
	ColumnStart     string `json:"grid-column-start,omitempty"`
	ColumnEnd       string `json:"grid-column-end,omitempty"`
	Row             string `json:"grid-row,omitempty"`
	RowStart        string `json:"grid-row-start,omitempty"`
	RowEnd          string `json:"grid-row-end,omitempty"`
	Gap             string `json:"gap,omitempty"`
	GridGap         string `json:"grid-gap,omitempty"`
	ColumnGap       string `json:"column-gap,omitempty"`
	GridColumnGap   string `json:"grid-column-gap,omitempty"`
	RowGap          string `json:"row-gap,omitempty"`
	GridRowGap      string `json:"grid-row-gap,omitempty"`
}

// Visual Effects Properties
type VisualEffectsProperties struct {
	Opacity            string     `json:"opacity,omitempty"`
	Visibility         Visibility `json:"visibility,omitempty"`
	BoxShadow          string     `json:"box-shadow,omitempty"`
	Filter             string     `json:"filter,omitempty"`
	BackdropFilter     string     `json:"backdrop-filter,omitempty"`
	MixBlendMode       string     `json:"mix-blend-mode,omitempty"`
	Transform          string     `json:"transform,omitempty"`
	TransformOrigin    string     `json:"transform-origin,omitempty"`
	TransformStyle     string     `json:"transform-style,omitempty"`
	Perspective        string     `json:"perspective,omitempty"`
	PerspectiveOrigin  string     `json:"perspective-origin,omitempty"`
	BackfaceVisibility string     `json:"backface-visibility,omitempty"`
}

// Animation Properties
type AnimationProperties struct {
	Animation      string `json:"animation,omitempty"`
	Name           string `json:"animation-name,omitempty"`
	Duration       string `json:"animation-duration,omitempty"`
	TimingFunction string `json:"animation-timing-function,omitempty"`
	Delay          string `json:"animation-delay,omitempty"`
	IterationCount string `json:"animation-iteration-count,omitempty"`
	Direction      string `json:"animation-direction,omitempty"`
	FillMode       string `json:"animation-fill-mode,omitempty"`
	PlayState      string `json:"animation-play-state,omitempty"`
}

// Transition Properties
type TransitionProperties struct {
	Transition     string `json:"transition,omitempty"`
	Property       string `json:"transition-property,omitempty"`
	Duration       string `json:"transition-duration,omitempty"`
	TimingFunction string `json:"transition-timing-function,omitempty"`
	Delay          string `json:"transition-delay,omitempty"`
}

// Overflow Properties
type OverflowProperties struct {
	Overflow   Overflow `json:"overflow,omitempty"`
	X          Overflow `json:"overflow-x,omitempty"`
	Y          Overflow `json:"overflow-y,omitempty"`
	ClipMargin string   `json:"overflow-clip-margin,omitempty"`
}

// Table Properties
type TableProperties struct {
	Layout         string `json:"table-layout,omitempty"`
	BorderCollapse string `json:"border-collapse,omitempty"`
	BorderSpacing  string `json:"border-spacing,omitempty"`
	CaptionSide    string `json:"caption-side,omitempty"`
	EmptyCells     string `json:"empty-cells,omitempty"`
}

// List Properties
type ListProperties struct {
	ListStyle string `json:"list-style,omitempty"`
	Type      string `json:"list-style-type,omitempty"`
	Position  string `json:"list-style-position,omitempty"`
	Image     string `json:"list-style-image,omitempty"`
}

// Interaction Properties
type InteractionProperties struct {
	Cursor        Cursor `json:"cursor,omitempty"`
	PointerEvents string `json:"pointer-events,omitempty"`
	UserSelect    string `json:"user-select,omitempty"`
	TouchAction   string `json:"touch-action,omitempty"`
}

// Scroll Properties
type ScrollProperties struct {
	Behavior      string `json:"scroll-behavior,omitempty"`
	Margin        string `json:"scroll-margin,omitempty"`
	MarginTop     string `json:"scroll-margin-top,omitempty"`
	MarginRight   string `json:"scroll-margin-right,omitempty"`
	MarginBottom  string `json:"scroll-margin-bottom,omitempty"`
	MarginLeft    string `json:"scroll-margin-left,omitempty"`
	Padding       string `json:"scroll-padding,omitempty"`
	PaddingTop    string `json:"scroll-padding-top,omitempty"`
	PaddingRight  string `json:"scroll-padding-right,omitempty"`
	PaddingBottom string `json:"scroll-padding-bottom,omitempty"`
	PaddingLeft   string `json:"scroll-padding-left,omitempty"`
	SnapType      string `json:"scroll-snap-type,omitempty"`
	SnapAlign     string `json:"scroll-snap-align,omitempty"`
}

// Modern Layout Properties
type ModernLayoutProperties struct {
	InlineSize    string `json:"inline-size,omitempty"`
	BlockSize     string `json:"block-size,omitempty"`
	MinInlineSize string `json:"min-inline-size,omitempty"`
	MinBlockSize  string `json:"min-block-size,omitempty"`
	MaxInlineSize string `json:"max-inline-size,omitempty"`
	MaxBlockSize  string `json:"max-block-size,omitempty"`
}

// Logical Properties
type LogicalProperties struct {
	Inset       string `json:"inset,omitempty"`
	Block       string `json:"inset-block,omitempty"`
	BlockStart  string `json:"inset-block-start,omitempty"`
	BlockEnd    string `json:"inset-block-end,omitempty"`
	Inline      string `json:"inset-inline,omitempty"`
	InlineStart string `json:"inset-inline-start,omitempty"`
	InlineEnd   string `json:"inset-inline-end,omitempty"`
}

// Advanced Typography
type AdvancedTypoProperties struct {
	OpticalSizing     string `json:"font-optical-sizing,omitempty"`
	VariationSettings string `json:"font-variation-settings,omitempty"`
	FeatureSettings   string `json:"font-feature-settings,omitempty"`
	UnderlineOffset   string `json:"text-underline-offset,omitempty"`
	UnderlinePosition string `json:"text-underline-position,omitempty"`
}

// Mask Properties
type MaskProperties struct {
	Mask      string `json:"mask,omitempty"`
	Image     string `json:"mask-image,omitempty"`
	Mode      string `json:"mask-mode,omitempty"`
	Repeat    string `json:"mask-repeat,omitempty"`
	Position  string `json:"mask-position,omitempty"`
	Clip      string `json:"mask-clip,omitempty"`
	Origin    string `json:"mask-origin,omitempty"`
	Size      string `json:"mask-size,omitempty"`
	Composite string `json:"mask-composite,omitempty"`
}

// Clip Path Properties
type ClipPathProperties struct {
	ClipPath string `json:"clip-path,omitempty"`
}

// Shape Properties
type ShapeProperties struct {
	Outside        string `json:"shape-outside,omitempty"`
	Margin         string `json:"shape-margin,omitempty"`
	ImageThreshold string `json:"shape-image-threshold,omitempty"`
}

// Container Properties
type ContainerProperties struct {
	Container string `json:"container,omitempty"`
	Name      string `json:"container-name,omitempty"`
	Type      string `json:"container-type,omitempty"`
}
