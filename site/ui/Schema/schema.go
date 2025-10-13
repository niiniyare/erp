// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    schema, err := UnmarshalSchema(bytes)
//    bytes, err = schema.Marshal()

package main

import "bytes"
import "errors"

import "encoding/json"

func UnmarshalSchema(data []byte) (Schema, error) {
	var r Schema
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Schema) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Schema struct {
	AllOf                []SchemaAllOf               `json:"allOf,omitempty"`
	Schema               string                      `json:"$schema"`
	ID                   *string                     `json:"$id,omitempty"`
	Type                 *TypeUnion                  `json:"type"`
	Properties           map[string]Property         `json:"properties,omitempty"`
	Required             []string                    `json:"required,omitempty"`
	AdditionalProperties *SchemaAdditionalProperties `json:"additionalProperties"`
	Description          *string                     `json:"description,omitempty"`
	AnyOf                []SchemaAnyOf               `json:"anyOf,omitempty"`
	Enum                 []string                    `json:"enum,omitempty"`
	Items                *SchemaItems                `json:"items,omitempty"`
	Ref                  *string                     `json:"$ref,omitempty"`
}

type PurpleAdditionalProperties struct {
	Type                 TypeElement      `json:"type"`
	Properties           PurpleProperties `json:"properties"`
	AdditionalProperties bool             `json:"additionalProperties"`
}

type PurpleProperties struct {
	Icon      ClassName `json:"icon"`
	Label     ClassName `json:"label"`
	Color     ClassName `json:"color"`
	ClassName ClassName `json:"className"`
}

type ClassName struct {
	Type *TypeElement `json:"type,omitempty"`
}

type SchemaAllOf struct {
	If                   *If               `json:"if,omitempty"`
	Then                 *Then             `json:"then,omitempty"`
	Ref                  *string           `json:"$ref,omitempty"`
	PatternProperties    map[string]Not    `json:"patternProperties,omitempty"`
	Type                 *TypeElement      `json:"type,omitempty"`
	AdditionalProperties *bool             `json:"additionalProperties,omitempty"`
	Properties           *FluffyProperties `json:"properties,omitempty"`
	Required             []string          `json:"required,omitempty"`
}

type If struct {
	Properties IfProperties `json:"properties"`
}

type IfProperties struct {
	Type             Align             `json:"type"`
	ActionType       *Align            `json:"actionType,omitempty"`
	ThumbMode        *Align            `json:"thumbMode,omitempty"`
	Mode             *Align            `json:"mode,omitempty"`
	LoadType         *Align            `json:"loadType,omitempty"`
	DialogType       *Align            `json:"dialogType,omitempty"`
	SubFormMode      *Align            `json:"subFormMode,omitempty"`
	ImageMode        *Align            `json:"imageMode,omitempty"`
	TabsMode         *Align            `json:"tabsMode,omitempty"`
	OptionType       *TooltipPlacement `json:"optionType,omitempty"`
	BuilderMode      *Align            `json:"builderMode,omitempty"`
	BorderMode       *Align            `json:"borderMode,omitempty"`
	DisplayMode      *Align            `json:"displayMode,omitempty"`
	ModalMode        *Align            `json:"modalMode,omitempty"`
	SelectMode       *Align            `json:"selectMode,omitempty"`
	LeftMode         *Align            `json:"leftMode,omitempty"`
	RightMode        *Align            `json:"rightMode,omitempty"`
	SearchResultMode *Align            `json:"searchResultMode,omitempty"`
}

type Align struct {
	Type        TypeElement `json:"type"`
	Const       *string     `json:"const,omitempty"`
	Description *string     `json:"description,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

type TooltipPlacement struct {
	Type TypeElement `json:"type"`
	Enum []string    `json:"enum"`
}

type Not struct {
}

type FluffyProperties struct {
	Label              *CanAccessSuperData `json:"label,omitempty"`
	LabelClassName     *LabelClassName     `json:"labelClassName,omitempty"`
	Name               *CanAccessSuperData `json:"name,omitempty"`
	PopOver            *ColumnClassName    `json:"popOver,omitempty"`
	QuickEdit          *ColumnClassName    `json:"quickEdit,omitempty"`
	Copyable           *ColumnClassName    `json:"copyable,omitempty"`
	Unique             *CanAccessSuperData `json:"unique,omitempty"`
	ColumnClassName    *ColumnClassName    `json:"columnClassName,omitempty"`
	Testid             *ClassName          `json:"testid,omitempty"`
	X                  *CanAccessSuperData `json:"x,omitempty"`
	Y                  *CanAccessSuperData `json:"y,omitempty"`
	W                  *CanAccessSuperData `json:"w,omitempty"`
	H                  *CanAccessSuperData `json:"h,omitempty"`
	Width              *Height             `json:"width,omitempty"`
	Height             *Height             `json:"height,omitempty"`
	Align              *Align              `json:"align,omitempty"`
	Valign             *Align              `json:"valign,omitempty"`
	GridClassName      *CanAccessSuperData `json:"gridClassName,omitempty"`
	ColumnRatio        *ColumnRatio        `json:"columnRatio,omitempty"`
	RowClassName       *ClassName          `json:"rowClassName,omitempty"`
	CellClassName      *ClassName          `json:"cellClassName,omitempty"`
	InnerClassName     *ColumnClassName    `json:"innerClassName,omitempty"`
	Fixed              *Align              `json:"fixed,omitempty"`
	QuickEditOnUpdate  *ColumnClassName    `json:"quickEditOnUpdate,omitempty"`
	Sortable           *CanAccessSuperData `json:"sortable,omitempty"`
	Searchable         *Searchable         `json:"searchable,omitempty"`
	Toggled            *CanAccessSuperData `json:"toggled,omitempty"`
	VAlign             *Align              `json:"vAlign,omitempty"`
	HeaderAlign        *Align              `json:"headerAlign,omitempty"`
	ClassName          *CanAccessSuperData `json:"className,omitempty"`
	ClassNameExpr      *CanAccessSuperData `json:"classNameExpr,omitempty"`
	Filterable         *Filterable         `json:"filterable,omitempty"`
	Breakpoint         *Align              `json:"breakpoint,omitempty"`
	Remark             *ColumnClassName    `json:"remark,omitempty"`
	Value              *DataClass          `json:"value,omitempty"`
	CanAccessSuperData *CanAccessSuperData `json:"canAccessSuperData,omitempty"`
	LazyRenderAfter    *LazyRenderAfter    `json:"lazyRenderAfter,omitempty"`
	InnerStyle         *CanAccessSuperData `json:"innerStyle,omitempty"`
}

type CanAccessSuperData struct {
	Type        TypeElement `json:"type"`
	Description *string     `json:"description,omitempty"`
}

type ColumnClassName struct {
	Ref         ColumnClassNameRef `json:"$ref"`
	Description *string            `json:"description,omitempty"`
}

type ColumnRatio struct {
	AnyOf       []ColumnRatioAnyOf `json:"anyOf"`
	Description string             `json:"description"`
}

type ColumnRatioAnyOf struct {
	Type  TypeElement `json:"type"`
	Const *string     `json:"const,omitempty"`
}

type Filterable struct {
	AnyOf       []FilterableAnyOf `json:"anyOf"`
	Description string            `json:"description"`
}

type FilterableAnyOf struct {
	Type                 TypeElement          `json:"type"`
	Properties           *TentacledProperties `json:"properties,omitempty"`
	AdditionalProperties *bool                `json:"additionalProperties,omitempty"`
}

type TentacledProperties struct {
	Source  ClassName `json:"source"`
	Options Options   `json:"options"`
}

type Options struct {
	Type  TypeElement `json:"type"`
	Items Not         `json:"items"`
}

type Height struct {
	Type        []TypeElement `json:"type"`
	Description string        `json:"description"`
}

type LabelClassName struct {
	Ref         *ColumnClassNameRef `json:"$ref,omitempty"`
	Description *string             `json:"description,omitempty"`
	Type        *TypeElement        `json:"type,omitempty"`
	Items       *Then               `json:"items,omitempty"`
}

type Then struct {
	Ref string `json:"$ref"`
}

type LazyRenderAfter struct {
	Type        TypeElement `json:"type"`
	Description string      `json:"description"`
	Default     int64       `json:"default"`
}

type Searchable struct {
	AnyOf       []Desc `json:"anyOf"`
	Description string `json:"description"`
}

type Desc struct {
	Type *TypeElement `json:"type,omitempty"`
	Ref  *DescRef     `json:"$ref,omitempty"`
}

type DataClass struct {
	Description string `json:"description"`
}

type SchemaAnyOf struct {
	Ref                  *string                     `json:"$ref,omitempty"`
	Type                 *TypeElement                `json:"type,omitempty"`
	AdditionalProperties *IndigoAdditionalProperties `json:"additionalProperties"`
	Items                *PurpleItems                `json:"items,omitempty"`
	Properties           *IndecentProperties         `json:"properties,omitempty"`
	Required             []string                    `json:"required,omitempty"`
	Const                *ConstElement               `json:"const"`
	Not                  *Not                        `json:"not,omitempty"`
	PatternProperties    map[string]Not              `json:"patternProperties,omitempty"`
	AllOf                []PurpleAllOf               `json:"allOf,omitempty"`
	MinItems             *int64                      `json:"minItems,omitempty"`
	MaxItems             *int64                      `json:"maxItems,omitempty"`
}

type FluffyAdditionalProperties struct {
	AnyOf []PurpleAnyOf `json:"anyOf"`
}

type PurpleAnyOf struct {
	Type *TypeElement        `json:"type,omitempty"`
	Not  *Not                `json:"not,omitempty"`
	Ref  *ColumnClassNameRef `json:"$ref,omitempty"`
}

type PurpleAllOf struct {
	Ref                  *DescRef                 `json:"$ref,omitempty"`
	PatternProperties    *PurplePatternProperties `json:"patternProperties,omitempty"`
	Type                 *TypeElement             `json:"type,omitempty"`
	AdditionalProperties *bool                    `json:"additionalProperties,omitempty"`
	Properties           *StickyProperties        `json:"properties,omitempty"`
}

type PurplePatternProperties struct {
	SaveImmediatelyResetOnFailedReloadModeIcon Not `json:"^(saveImmediately|resetOnFailed|reload|mode|icon)$"`
}

type StickyProperties struct {
	SaveImmediately CanAccessSuperData `json:"saveImmediately"`
	ResetOnFailed   CanAccessSuperData `json:"resetOnFailed"`
	Reload          CanAccessSuperData `json:"reload"`
	Mode            Align              `json:"mode"`
	Icon            CanAccessSuperData `json:"icon"`
}

type PurpleItems struct {
	Type                 *TypeElement      `json:"type,omitempty"`
	Properties           *IndigoProperties `json:"properties,omitempty"`
	Required             []string          `json:"required,omitempty"`
	AdditionalProperties *bool             `json:"additionalProperties,omitempty"`
	AnyOf                []Then            `json:"anyOf,omitempty"`
}

type IndigoProperties struct {
	Value ClassName `json:"value"`
	Color ClassName `json:"color"`
}

type IndecentProperties struct {
	Type                 *Align              `json:"type,omitempty"`
	Label                *CanAccessSuperData `json:"label,omitempty"`
	Name                 *ClassName          `json:"name,omitempty"`
	Children             *Children           `json:"children,omitempty"`
	Prototype            *Not                `json:"prototype,omitempty"`
	Length               *ClassName          `json:"length,omitempty"`
	Arguments            *Not                `json:"arguments,omitempty"`
	Caller               *Then               `json:"caller,omitempty"`
	ID                   *CanAccessSuperData `json:"$$id,omitempty"`
	ClassName            *ColumnClassName    `json:"className,omitempty"`
	Ref                  *CanAccessSuperData `json:"$ref,omitempty"`
	Disabled             *CanAccessSuperData `json:"disabled,omitempty"`
	DisabledOn           *ColumnClassName    `json:"disabledOn,omitempty"`
	Hidden               *CanAccessSuperData `json:"hidden,omitempty"`
	HiddenOn             *ColumnClassName    `json:"hiddenOn,omitempty"`
	Visible              *CanAccessSuperData `json:"visible,omitempty"`
	VisibleOn            *ColumnClassName    `json:"visibleOn,omitempty"`
	PropertiesID         *CanAccessSuperData `json:"id,omitempty"`
	OnEvent              *OnEvent            `json:"onEvent,omitempty"`
	Static               *CanAccessSuperData `json:"static,omitempty"`
	StaticOn             *ColumnClassName    `json:"staticOn,omitempty"`
	StaticPlaceholder    *CanAccessSuperData `json:"staticPlaceholder,omitempty"`
	StaticClassName      *ColumnClassName    `json:"staticClassName,omitempty"`
	StaticLabelClassName *ColumnClassName    `json:"staticLabelClassName,omitempty"`
	StaticInputClassName *ColumnClassName    `json:"staticInputClassName,omitempty"`
	StaticSchema         *Not                `json:"staticSchema,omitempty"`
	Style                *CanAccessSuperData `json:"style,omitempty"`
	EditorSetting        *EditorSetting      `json:"editorSetting,omitempty"`
	UseMobileUI          *CanAccessSuperData `json:"useMobileUI,omitempty"`
	TestIDBuilder        *Then               `json:"testIdBuilder,omitempty"`
	Testid               *ClassName          `json:"testid,omitempty"`
	Block                *CanAccessSuperData `json:"block,omitempty"`
	DisabledTip          *CanAccessSuperData `json:"disabledTip,omitempty"`
	Icon                 *LabelClassName     `json:"icon,omitempty"`
	IconClassName        *ColumnClassName    `json:"iconClassName,omitempty"`
	RightIcon            *ColumnClassName    `json:"rightIcon,omitempty"`
	RightIconClassName   *ColumnClassName    `json:"rightIconClassName,omitempty"`
	LoadingClassName     *ColumnClassName    `json:"loadingClassName,omitempty"`
	Level                *Align              `json:"level,omitempty"`
	Primary              *ClassName          `json:"primary,omitempty"`
	Size                 *Align              `json:"size,omitempty"`
	Tooltip              *Then               `json:"tooltip,omitempty"`
	TooltipPlacement     *TooltipPlacement   `json:"tooltipPlacement,omitempty"`
	ConfirmText          *CanAccessSuperData `json:"confirmText,omitempty"`
	Required             *RequiredClass      `json:"required,omitempty"`
	ActiveLevel          *CanAccessSuperData `json:"activeLevel,omitempty"`
	ActiveClassName      *CanAccessSuperData `json:"activeClassName,omitempty"`
	Close                *Height             `json:"close,omitempty"`
	RequireSelected      *CanAccessSuperData `json:"requireSelected,omitempty"`
	MergeData            *CanAccessSuperData `json:"mergeData,omitempty"`
	Target               *LabelClassName     `json:"target,omitempty"`
	CountDown            *CanAccessSuperData `json:"countDown,omitempty"`
	CountDownTpl         *CanAccessSuperData `json:"countDownTpl,omitempty"`
	Badge                *ColumnClassName    `json:"badge,omitempty"`
	HotKey               *CanAccessSuperData `json:"hotKey,omitempty"`
	LoadingOn            *CanAccessSuperData `json:"loadingOn,omitempty"`
	OnClick              *OnClick            `json:"onClick,omitempty"`
	Body                 *LabelClassName     `json:"body,omitempty"`
	ActionType           *Align              `json:"actionType,omitempty"`
	API                  *ColumnClassName    `json:"api,omitempty"`
	Feedback             *Then               `json:"feedback,omitempty"`
	Reload               *LabelClassName     `json:"reload,omitempty"`
	Redirect             *ClassName          `json:"redirect,omitempty"`
	IgnoreConfirm        *ClassName          `json:"ignoreConfirm,omitempty"`
	IsolateScope         *CanAccessSuperData `json:"isolateScope,omitempty"`
	Blank                *CanAccessSuperData `json:"blank,omitempty"`
	URL                  *CanAccessSuperData `json:"url,omitempty"`
	Link                 *CanAccessSuperData `json:"link,omitempty"`
	Dialog               *ColumnClassName    `json:"dialog,omitempty"`
	NextCondition        *ColumnClassName    `json:"nextCondition,omitempty"`
	Data                 *DataClass          `json:"data,omitempty"`
	Drawer               *ColumnClassName    `json:"drawer,omitempty"`
	Toast                *ColumnClassName    `json:"toast,omitempty"`
	Copy                 *ColumnClassName    `json:"copy,omitempty"`
	To                   *CanAccessSuperData `json:"to,omitempty"`
	Cc                   *CanAccessSuperData `json:"cc,omitempty"`
	Bcc                  *CanAccessSuperData `json:"bcc,omitempty"`
	Subject              *CanAccessSuperData `json:"subject,omitempty"`
	DownloadFileName     *ClassName          `json:"downloadFileName,omitempty"`
	Value                *PurpleValue        `json:"value,omitempty"`
	ValueTypes           *Types              `json:"valueTypes,omitempty"`
	Operators            *Operators          `json:"operators,omitempty"`
	Funcs                *Funcs              `json:"funcs,omitempty"`
	DefaultValue         *Not                `json:"defaultValue,omitempty"`
	Placeholder          *ClassName          `json:"placeholder,omitempty"`
	MinLength            *ClassName          `json:"minLength,omitempty"`
	MaxLength            *ClassName          `json:"maxLength,omitempty"`
	Maximum              *ClassName          `json:"maximum,omitempty"`
	Minimum              *ClassName          `json:"minimum,omitempty"`
	Step                 *ClassName          `json:"step,omitempty"`
	Precision            *ClassName          `json:"precision,omitempty"`
	Format               *ClassName          `json:"format,omitempty"`
	InputFormat          *ClassName          `json:"inputFormat,omitempty"`
	MinDate              *Not                `json:"minDate,omitempty"`
	MaxDate              *Not                `json:"maxDate,omitempty"`
	MinTime              *Not                `json:"minTime,omitempty"`
	MaxTime              *Not                `json:"maxTime,omitempty"`
	TimeFormat           *ClassName          `json:"timeFormat,omitempty"`
	Multiple             *ClassName          `json:"multiple,omitempty"`
	Options              *Options            `json:"options,omitempty"`
	Source               *Source             `json:"source,omitempty"`
	Searchable           *ClassName          `json:"searchable,omitempty"`
	MaxTagCount          *ClassName          `json:"maxTagCount,omitempty"`
	OverflowTagPopover   *Not                `json:"overflowTagPopover,omitempty"`
	AutoComplete         *Searchable         `json:"autoComplete,omitempty"`
	Color                *ClassName          `json:"color,omitempty"`
	Title                *CanAccessSuperData `json:"title,omitempty"`
	SaveImmediately      *CanAccessSuperData `json:"saveImmediately,omitempty"`
	ResetOnFailed        *CanAccessSuperData `json:"resetOnFailed,omitempty"`
	Mode                 *Align              `json:"mode,omitempty"`
	TooltipClassName     *Then               `json:"tooltipClassName,omitempty"`
	Trigger              *Trigger            `json:"trigger,omitempty"`
	Content              *LabelClassName     `json:"content,omitempty"`
	Placement            *Align              `json:"placement,omitempty"`
	RootClose            *CanAccessSuperData `json:"rootClose,omitempty"`
	Shape                *Align              `json:"shape,omitempty"`
}

type Children struct {
	Type  TypeElement `json:"type"`
	Items Then        `json:"items"`
}

type EditorSetting struct {
	Type        TypeElement              `json:"type"`
	Properties  EditorSettingProperties  `json:"properties"`
	Description EditorSettingDescription `json:"description"`
}

type EditorSettingProperties struct {
	Behavior    CanAccessSuperData `json:"behavior"`
	DisplayName CanAccessSuperData `json:"displayName"`
	Mock        DataClass          `json:"mock"`
}

type Funcs struct {
	Type  TypeElement `json:"type"`
	Items *ClassName  `json:"items,omitempty"`
}

type OnClick struct {
	AnyOf       []ClassName        `json:"anyOf"`
	Description OnClickDescription `json:"description"`
}

type OnEvent struct {
	Type                 TypeElement                 `json:"type"`
	AdditionalProperties OnEventAdditionalProperties `json:"additionalProperties"`
	Description          OnEventDescription          `json:"description"`
}

type OnEventAdditionalProperties struct {
	Type                 TypeElement                    `json:"type"`
	Properties           HilariousProperties            `json:"properties"`
	Required             []AdditionalPropertiesRequired `json:"required"`
	AdditionalProperties bool                           `json:"additionalProperties"`
}

type HilariousProperties struct {
	Weight   ClassName `json:"weight"`
	Actions  Children  `json:"actions"`
	Debounce Then      `json:"debounce"`
	Track    Then      `json:"track"`
}

type Operators struct {
	Type  TypeElement    `json:"type"`
	Items OperatorsItems `json:"items"`
}

type OperatorsItems struct {
	AnyOf []FluffyAnyOf `json:"anyOf"`
}

type FluffyAnyOf struct {
	Type                 TypeElement          `json:"type"`
	Properties           *AmbitiousProperties `json:"properties,omitempty"`
	Required             []AnyOfRequired      `json:"required,omitempty"`
	AdditionalProperties *bool                `json:"additionalProperties,omitempty"`
}

type AmbitiousProperties struct {
	Lable  ClassName `json:"lable"`
	Value  ClassName `json:"value"`
	Values Options   `json:"values"`
}

type RequiredClass struct {
	Type        TypeElement `json:"type"`
	Items       ClassName   `json:"items"`
	Description string      `json:"description"`
	Default     []int64     `json:"default,omitempty"`
}

type Source struct {
	AnyOf []Desc `json:"anyOf"`
}

type Trigger struct {
	Type        TypeElement      `json:"type"`
	Items       TooltipPlacement `json:"items"`
	Description string           `json:"description"`
}

type PurpleValue struct {
	AnyOf []ClassName  `json:"anyOf,omitempty"`
	Type  *TypeElement `json:"type,omitempty"`
}

type Types struct {
	Type  TypeElement      `json:"type"`
	Items TooltipPlacement `json:"items"`
}

type SchemaItems struct {
	Ref   *string          `json:"$ref,omitempty"`
	AnyOf []TentacledAnyOf `json:"anyOf,omitempty"`
}

type TentacledAnyOf struct {
	Ref                  *string            `json:"$ref,omitempty"`
	Type                 *TypeElement       `json:"type,omitempty"`
	Properties           *CunningProperties `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	AdditionalProperties *bool              `json:"additionalProperties,omitempty"`
}

type CunningProperties struct {
	Label    ClassName `json:"label"`
	Children Children  `json:"children"`
}

type Property struct {
	Type                 *TypeUnion                    `json:"type"`
	Description          *string                       `json:"description,omitempty"`
	Ref                  *string                       `json:"$ref,omitempty"`
	Default              *DefaultUnion                 `json:"default"`
	AdditionalProperties *PropertyAdditionalProperties `json:"additionalProperties"`
	Properties           *PropertyProperties           `json:"properties,omitempty"`
	Enum                 []EnumUnion                   `json:"enum,omitempty"`
	Const                *EnumUnion                    `json:"const"`
	AnyOf                []PropertyAnyOf               `json:"anyOf,omitempty"`
	Items                *PropertyItems                `json:"items,omitempty"`
	Required             []string                      `json:"required,omitempty"`
	MinItems             *int64                        `json:"minItems,omitempty"`
	MaxItems             *int64                        `json:"maxItems,omitempty"`
}

type TentacledAdditionalProperties struct {
	Type                 *TypeUnion                     `json:"type"`
	Properties           *FriskyProperties              `json:"properties,omitempty"`
	Required             []AdditionalPropertiesRequired `json:"required,omitempty"`
	AdditionalProperties *bool                          `json:"additionalProperties,omitempty"`
	Ref                  *string                        `json:"$ref,omitempty"`
	AnyOf                []StickyAnyOf                  `json:"anyOf,omitempty"`
}

type StickyAnyOf struct {
	Type                 *TypeElement                  `json:"type,omitempty"`
	AdditionalProperties *IndecentAdditionalProperties `json:"additionalProperties"`
	Ref                  *DescRef                      `json:"$ref,omitempty"`
	Properties           *MagentaProperties            `json:"properties,omitempty"`
}

type StickyAdditionalProperties struct {
	Type                 TypeElement `json:"type"`
	AdditionalProperties *ClassName  `json:"additionalProperties,omitempty"`
}

type MagentaProperties struct {
	Style Then      `json:"style"`
	Label ClassName `json:"label"`
}

type FriskyProperties struct {
	Weight   *ClassName        `json:"weight,omitempty"`
	Actions  *Children         `json:"actions,omitempty"`
	Debounce *Then             `json:"debounce,omitempty"`
	Track    *Then             `json:"track,omitempty"`
	Title    *ClassName        `json:"title,omitempty"`
	Type     *TooltipPlacement `json:"type,omitempty"`
}

type PropertyAnyOf struct {
	Ref                  *string                        `json:"$ref,omitempty"`
	Type                 *TypeElement                   `json:"type,omitempty"`
	Items                *FluffyItems                   `json:"items,omitempty"`
	Const                *EnumUnion                     `json:"const"`
	Properties           *BraggadociousProperties       `json:"properties,omitempty"`
	AdditionalProperties *HilariousAdditionalProperties `json:"additionalProperties"`
	Required             []string                       `json:"required,omitempty"`
}

type FluffyItems struct {
	Ref                  *ItemsRef              `json:"$ref,omitempty"`
	Type                 *TypeUnion             `json:"type"`
	Properties           *MischievousProperties `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties *bool                  `json:"additionalProperties,omitempty"`
}

type MischievousProperties struct {
	StartTime *ClassName `json:"startTime,omitempty"`
	EndTime   *ClassName `json:"endTime,omitempty"`
	Content   *Not       `json:"content,omitempty"`
	ClassName *ClassName `json:"className,omitempty"`
	Value     *Then      `json:"value,omitempty"`
	Color     *ClassName `json:"color,omitempty"`
}

type BraggadociousProperties struct {
	IsAlpha                *CanAccessSuperData       `json:"isAlpha,omitempty"`
	IsAlphanumeric         *CanAccessSuperData       `json:"isAlphanumeric,omitempty"`
	IsEmail                *CanAccessSuperData       `json:"isEmail,omitempty"`
	IsFloat                *CanAccessSuperData       `json:"isFloat,omitempty"`
	IsInt                  *CanAccessSuperData       `json:"isInt,omitempty"`
	IsJSON                 *CanAccessSuperData       `json:"isJson,omitempty"`
	IsLength               *CanAccessSuperData       `json:"isLength,omitempty"`
	IsNumeric              *CanAccessSuperData       `json:"isNumeric,omitempty"`
	IsRequired             *CanAccessSuperData       `json:"isRequired,omitempty"`
	IsURL                  *CanAccessSuperData       `json:"isUrl,omitempty"`
	MatchRegexp            *CanAccessSuperData       `json:"matchRegexp,omitempty"`
	MatchRegexp1           *CanAccessSuperData       `json:"matchRegexp1,omitempty"`
	MatchRegexp2           *CanAccessSuperData       `json:"matchRegexp2,omitempty"`
	MatchRegexp3           *CanAccessSuperData       `json:"matchRegexp3,omitempty"`
	MatchRegexp4           *CanAccessSuperData       `json:"matchRegexp4,omitempty"`
	MatchRegexp5           *CanAccessSuperData       `json:"matchRegexp5,omitempty"`
	MaxLength              *CanAccessSuperData       `json:"maxLength,omitempty"`
	Maximum                *CanAccessSuperData       `json:"maximum,omitempty"`
	MinLength              *CanAccessSuperData       `json:"minLength,omitempty"`
	Minimum                *CanAccessSuperData       `json:"minimum,omitempty"`
	IsDateTimeSame         *Is                       `json:"isDateTimeSame,omitempty"`
	IsDateTimeBefore       *Is                       `json:"isDateTimeBefore,omitempty"`
	IsDateTimeAfter        *Is                       `json:"isDateTimeAfter,omitempty"`
	IsDateTimeSameOrBefore *Is                       `json:"isDateTimeSameOrBefore,omitempty"`
	IsDateTimeSameOrAfter  *Is                       `json:"isDateTimeSameOrAfter,omitempty"`
	IsDateTimeBetween      *Is                       `json:"isDateTimeBetween,omitempty"`
	IsTimeSame             *Is                       `json:"isTimeSame,omitempty"`
	IsTimeBefore           *Is                       `json:"isTimeBefore,omitempty"`
	IsTimeAfter            *Is                       `json:"isTimeAfter,omitempty"`
	IsTimeSameOrBefore     *Is                       `json:"isTimeSameOrBefore,omitempty"`
	IsTimeSameOrAfter      *Is                       `json:"isTimeSameOrAfter,omitempty"`
	IsTimeBetween          *Is                       `json:"isTimeBetween,omitempty"`
	ShowSuggestion         *CanAccessSuperData       `json:"showSuggestion,omitempty"`
	DefaultSelection       *DataClass                `json:"defaultSelection,omitempty"`
	API                    *API                      `json:"api,omitempty"`
	Silent                 *PopOverContainerSelector `json:"silent,omitempty"`
	FillMappinng           *CanAccessSuperData       `json:"fillMappinng,omitempty"`
	Trigger                *Align                    `json:"trigger,omitempty"`
	Mode                   *Align                    `json:"mode,omitempty"`
	Position               *Align                    `json:"position,omitempty"`
	Size                   *Align                    `json:"size,omitempty"`
	Columns                *Columns                  `json:"columns,omitempty"`
	Filter                 *DataClass                `json:"filter,omitempty"`
	Type                   *Align                    `json:"type,omitempty"`
	Header                 *Header                   `json:"header,omitempty"`
	Body                   *LabelClassName           `json:"body,omitempty"`
	Media                  *Media                    `json:"media,omitempty"`
	Actions                *ActionsClass             `json:"actions,omitempty"`
	Toolbar                *LabelClassName           `json:"toolbar,omitempty"`
	Secondary              *ColumnClassName          `json:"secondary,omitempty"`
	UseCardLabel           *CanAccessSuperData       `json:"useCardLabel,omitempty"`
	Testid                 *ClassName                `json:"testid,omitempty"`
	ID                     *CanAccessSuperData       `json:"$$id,omitempty"`
	ClassName              *LabelClassName           `json:"className,omitempty"`
	Ref                    *CanAccessSuperData       `json:"$ref,omitempty"`
	Disabled               *CanAccessSuperData       `json:"disabled,omitempty"`
	DisabledOn             *ColumnClassName          `json:"disabledOn,omitempty"`
	Hidden                 *CanAccessSuperData       `json:"hidden,omitempty"`
	HiddenOn               *ColumnClassName          `json:"hiddenOn,omitempty"`
	Visible                *CanAccessSuperData       `json:"visible,omitempty"`
	VisibleOn              *ColumnClassName          `json:"visibleOn,omitempty"`
	PropertiesID           *CanAccessSuperData       `json:"id,omitempty"`
	OnEvent                *OnEvent                  `json:"onEvent,omitempty"`
	Static                 *CanAccessSuperData       `json:"static,omitempty"`
	StaticOn               *ColumnClassName          `json:"staticOn,omitempty"`
	StaticPlaceholder      *CanAccessSuperData       `json:"staticPlaceholder,omitempty"`
	StaticClassName        *ColumnClassName          `json:"staticClassName,omitempty"`
	StaticLabelClassName   *ColumnClassName          `json:"staticLabelClassName,omitempty"`
	StaticInputClassName   *ColumnClassName          `json:"staticInputClassName,omitempty"`
	StaticSchema           *Not                      `json:"staticSchema,omitempty"`
	Style                  *CanAccessSuperData       `json:"style,omitempty"`
	EditorSetting          *EditorSetting            `json:"editorSetting,omitempty"`
	UseMobileUI            *CanAccessSuperData       `json:"useMobileUI,omitempty"`
	TestIDBuilder          *Then                     `json:"testIdBuilder,omitempty"`
	Height                 *ClassName                `json:"height,omitempty"`
	MaxHeight              *ClassName                `json:"maxHeight,omitempty"`
	Enable                 *ClassName                `json:"enable,omitempty"`
	Types                  *Types                    `json:"types,omitempty"`
	Label                  *ClassName                `json:"label,omitempty"`
	ActiveLabel            *ClassName                `json:"activeLabel,omitempty"`
	Icon                   *ClassName                `json:"icon,omitempty"`
	ActiveIcon             *ClassName                `json:"activeIcon,omitempty"`
	Expand                 *TooltipPlacement         `json:"expand,omitempty"`
	Accordion              *CanAccessSuperData       `json:"accordion,omitempty"`
	Source                 *ClassName                `json:"source,omitempty"`
	Options                *Options                  `json:"options,omitempty"`
	Left                   *ClassName                `json:"left,omitempty"`
	Right                  *ClassName                `json:"right,omitempty"`
	Top                    *ClassName                `json:"top,omitempty"`
	Bottom                 *ClassName                `json:"bottom,omitempty"`
	AspectRatio            *CanAccessSuperData       `json:"aspectRatio,omitempty"`
	Guides                 *ClassName                `json:"guides,omitempty"`
	DragMode               *ClassName                `json:"dragMode,omitempty"`
	ViewMode               *ClassName                `json:"viewMode,omitempty"`
	Rotatable              *ClassName                `json:"rotatable,omitempty"`
	Scalable               *ClassName                `json:"scalable,omitempty"`
	Value                  *Then                     `json:"value,omitempty"`
	Color                  *ClassName                `json:"color,omitempty"`
	Block                  *CanAccessSuperData       `json:"block,omitempty"`
	DisabledTip            *CanAccessSuperData       `json:"disabledTip,omitempty"`
	IconClassName          *ColumnClassName          `json:"iconClassName,omitempty"`
	RightIcon              *ColumnClassName          `json:"rightIcon,omitempty"`
	RightIconClassName     *ColumnClassName          `json:"rightIconClassName,omitempty"`
	LoadingClassName       *ColumnClassName          `json:"loadingClassName,omitempty"`
	Level                  *Align                    `json:"level,omitempty"`
	Primary                *ClassName                `json:"primary,omitempty"`
	Tooltip                *Then                     `json:"tooltip,omitempty"`
	TooltipPlacement       *TooltipPlacement         `json:"tooltipPlacement,omitempty"`
	ConfirmText            *CanAccessSuperData       `json:"confirmText,omitempty"`
	Required               *RequiredClass            `json:"required,omitempty"`
	ActiveLevel            *CanAccessSuperData       `json:"activeLevel,omitempty"`
	ActiveClassName        *CanAccessSuperData       `json:"activeClassName,omitempty"`
	Close                  *Height                   `json:"close,omitempty"`
	RequireSelected        *CanAccessSuperData       `json:"requireSelected,omitempty"`
	MergeData              *CanAccessSuperData       `json:"mergeData,omitempty"`
	Target                 *LabelClassName           `json:"target,omitempty"`
	CountDown              *CanAccessSuperData       `json:"countDown,omitempty"`
	CountDownTpl           *CanAccessSuperData       `json:"countDownTpl,omitempty"`
	Badge                  *ColumnClassName          `json:"badge,omitempty"`
	HotKey                 *CanAccessSuperData       `json:"hotKey,omitempty"`
	LoadingOn              *CanAccessSuperData       `json:"loadingOn,omitempty"`
	OnClick                *OnClick                  `json:"onClick,omitempty"`
	ActionType             *Align                    `json:"actionType,omitempty"`
	Feedback               *Then                     `json:"feedback,omitempty"`
	Reload                 *Then                     `json:"reload,omitempty"`
	Redirect               *ClassName                `json:"redirect,omitempty"`
	IgnoreConfirm          *ClassName                `json:"ignoreConfirm,omitempty"`
	IsolateScope           *CanAccessSuperData       `json:"isolateScope,omitempty"`
	Blank                  *CanAccessSuperData       `json:"blank,omitempty"`
	URL                    *CanAccessSuperData       `json:"url,omitempty"`
	Link                   *CanAccessSuperData       `json:"link,omitempty"`
	Dialog                 *ColumnClassName          `json:"dialog,omitempty"`
	NextCondition          *ColumnClassName          `json:"nextCondition,omitempty"`
	Data                   *DataClass                `json:"data,omitempty"`
	Drawer                 *ColumnClassName          `json:"drawer,omitempty"`
	Toast                  *ColumnClassName          `json:"toast,omitempty"`
	Copy                   *ColumnClassName          `json:"copy,omitempty"`
	To                     *CanAccessSuperData       `json:"to,omitempty"`
	Cc                     *CanAccessSuperData       `json:"cc,omitempty"`
	Bcc                    *CanAccessSuperData       `json:"bcc,omitempty"`
	Subject                *CanAccessSuperData       `json:"subject,omitempty"`
	DownloadFileName       *ClassName                `json:"downloadFileName,omitempty"`
}

type API struct {
	AnyOf       []Desc              `json:"anyOf,omitempty"`
	Description APIDescription      `json:"description"`
	Ref         *ColumnClassNameRef `json:"$ref,omitempty"`
}

type ActionsClass struct {
	Type        TypeElement `json:"type"`
	Items       Then        `json:"items"`
	Description string      `json:"description"`
}

type Columns struct {
	Type        TypeElement        `json:"type"`
	Items       Not                `json:"items"`
	Description ColumnsDescription `json:"description"`
}

type Header struct {
	Type                 TypeElement      `json:"type"`
	Properties           HeaderProperties `json:"properties"`
	AdditionalProperties bool             `json:"additionalProperties"`
	Description          string           `json:"description"`
}

type HeaderProperties struct {
	ClassName              Then                 `json:"className"`
	Title                  ColumnClassName      `json:"title"`
	TitleClassName         Then                 `json:"titleClassName"`
	SubTitle               ColumnClassName      `json:"subTitle"`
	SubTitleClassName      Then                 `json:"subTitleClassName"`
	SubTitlePlaceholder    ClassName            `json:"subTitlePlaceholder"`
	Description            ColumnClassName      `json:"description"`
	DescriptionPlaceholder CanAccessSuperData   `json:"descriptionPlaceholder"`
	DescriptionClassName   ColumnClassName      `json:"descriptionClassName"`
	Desc                   Then                 `json:"desc"`
	DescPlaceholder        Then                 `json:"descPlaceholder"`
	DescClassName          Then                 `json:"descClassName"`
	Avatar                 ColumnClassName      `json:"avatar"`
	AvatarText             Then                 `json:"avatarText"`
	AvatarTextBackground   AvatarTextBackground `json:"avatarTextBackground"`
	AvatarTextClassName    Then                 `json:"avatarTextClassName"`
	AvatarClassName        ColumnClassName      `json:"avatarClassName"`
	ImageClassName         ColumnClassName      `json:"imageClassName"`
	Highlight              ColumnClassName      `json:"highlight"`
	HighlightClassName     Then                 `json:"highlightClassName"`
	Href                   ColumnClassName      `json:"href"`
	Blank                  CanAccessSuperData   `json:"blank"`
}

type AvatarTextBackground struct {
	Type  TypeElement               `json:"type"`
	Items AvatarTextBackgroundItems `json:"items"`
}

type AvatarTextBackgroundItems struct {
	Type                 TypeElement `json:"type"`
	Properties           Properties1 `json:"properties"`
	Required             []string    `json:"required"`
	AdditionalProperties ClassName   `json:"additionalProperties"`
}

type Properties1 struct {
	Length ClassName `json:"length"`
}

type Is struct {
	AnyOf       []Funcs `json:"anyOf"`
	Description string  `json:"description"`
}

type Media struct {
	Type                 TypeElement     `json:"type"`
	Properties           MediaProperties `json:"properties"`
	AdditionalProperties bool            `json:"additionalProperties"`
	Description          string          `json:"description"`
}

type MediaProperties struct {
	ClassName Then               `json:"className"`
	Type      Align              `json:"type"`
	URL       ColumnClassName    `json:"url"`
	Position  Align              `json:"position"`
	AutoPlay  CanAccessSuperData `json:"autoPlay"`
	IsLive    CanAccessSuperData `json:"isLive"`
	Poster    ColumnClassName    `json:"poster"`
}

type PopOverContainerSelector struct {
	Type        TypeElement `json:"type"`
	Description string      `json:"description"`
	Default     bool        `json:"default"`
}

type DefaultClass struct {
	Success  *string `json:"success,omitempty"`
	Pending  *string `json:"pending,omitempty"`
	Fail     *string `json:"fail,omitempty"`
	Queue    *string `json:"queue,omitempty"`
	Schedule *string `json:"schedule,omitempty"`
}

type PropertyItems struct {
	Type                 *TypeUnion    `json:"type"`
	Ref                  *string       `json:"$ref,omitempty"`
	AnyOf                []IndigoAnyOf `json:"anyOf,omitempty"`
	Properties           *Properties4  `json:"properties,omitempty"`
	AdditionalProperties *bool         `json:"additionalProperties,omitempty"`
	Required             []string      `json:"required,omitempty"`
	Enum                 []string      `json:"enum,omitempty"`
}

type IndigoAnyOf struct {
	Ref                  *string         `json:"$ref,omitempty"`
	Type                 *TypeElement    `json:"type,omitempty"`
	Properties           *Properties3    `json:"properties,omitempty"`
	Required             []AnyOfRequired `json:"required,omitempty"`
	AdditionalProperties *bool           `json:"additionalProperties,omitempty"`
	AllOf                []FluffyAllOf   `json:"allOf,omitempty"`
}

type FluffyAllOf struct {
	Ref                  *DescRef                 `json:"$ref,omitempty"`
	PatternProperties    *FluffyPatternProperties `json:"patternProperties,omitempty"`
	Type                 *TypeElement             `json:"type,omitempty"`
	AdditionalProperties *bool                    `json:"additionalProperties,omitempty"`
	Properties           *Properties2             `json:"properties,omitempty"`
}

type FluffyPatternProperties struct {
	Align Not `json:"^(align)$"`
}

type Properties2 struct {
	Align Align `json:"align"`
}

type Properties3 struct {
	Value                *ClassName          `json:"value,omitempty"`
	Label                *ClassName          `json:"label,omitempty"`
	Align                *Align              `json:"align,omitempty"`
	Testid               *ClassName          `json:"testid,omitempty"`
	ID                   *CanAccessSuperData `json:"$$id,omitempty"`
	ClassName            *ColumnClassName    `json:"className,omitempty"`
	Ref                  *CanAccessSuperData `json:"$ref,omitempty"`
	Disabled             *CanAccessSuperData `json:"disabled,omitempty"`
	DisabledOn           *ColumnClassName    `json:"disabledOn,omitempty"`
	Hidden               *CanAccessSuperData `json:"hidden,omitempty"`
	HiddenOn             *ColumnClassName    `json:"hiddenOn,omitempty"`
	Visible              *CanAccessSuperData `json:"visible,omitempty"`
	VisibleOn            *ColumnClassName    `json:"visibleOn,omitempty"`
	PropertiesID         *CanAccessSuperData `json:"id,omitempty"`
	OnEvent              *OnEvent            `json:"onEvent,omitempty"`
	Static               *CanAccessSuperData `json:"static,omitempty"`
	StaticOn             *ColumnClassName    `json:"staticOn,omitempty"`
	StaticPlaceholder    *CanAccessSuperData `json:"staticPlaceholder,omitempty"`
	StaticClassName      *ColumnClassName    `json:"staticClassName,omitempty"`
	StaticLabelClassName *ColumnClassName    `json:"staticLabelClassName,omitempty"`
	StaticInputClassName *ColumnClassName    `json:"staticInputClassName,omitempty"`
	StaticSchema         *Not                `json:"staticSchema,omitempty"`
	Style                *CanAccessSuperData `json:"style,omitempty"`
	EditorSetting        *EditorSetting      `json:"editorSetting,omitempty"`
	UseMobileUI          *CanAccessSuperData `json:"useMobileUI,omitempty"`
	TestIDBuilder        *Then               `json:"testIdBuilder,omitempty"`
	Type                 *Then               `json:"type,omitempty"`
	Lable                *ClassName          `json:"lable,omitempty"`
	Values               *Options            `json:"values,omitempty"`
}

type Properties4 struct {
	Key         *CanAccessSuperData `json:"key,omitempty"`
	Label       *CanAccessSuperData `json:"label,omitempty"`
	Remark      *CanAccessSuperData `json:"remark,omitempty"`
	Status      *Status             `json:"status,omitempty"`
	Title       *Then               `json:"title,omitempty"`
	Body        *Then               `json:"body,omitempty"`
	Level       *TooltipPlacement   `json:"level,omitempty"`
	ID          *ClassName          `json:"id,omitempty"`
	Position    *TooltipPlacement   `json:"position,omitempty"`
	CloseButton *ClassName          `json:"closeButton,omitempty"`
	ShowIcon    *ClassName          `json:"showIcon,omitempty"`
	Timeout     *ClassName          `json:"timeout,omitempty"`
	Rule        *ClassName          `json:"rule,omitempty"`
	Message     *ClassName          `json:"message,omitempty"`
	Name        *Name               `json:"name,omitempty"`
}

type Name struct {
	AnyOf []Funcs `json:"anyOf"`
}

type Status struct {
	Type        TypeElement `json:"type"`
	Enum        []int64     `json:"enum"`
	Description string      `json:"description"`
}

type PropertyProperties struct {
	Behavior                    *CanAccessSuperData       `json:"behavior,omitempty"`
	DisplayName                 *CanAccessSuperData       `json:"displayName,omitempty"`
	Mock                        *DataClass                `json:"mock,omitempty"`
	Count                       *ClassName                `json:"count,omitempty"`
	ValidateFailed              *CanAccessSuperData       `json:"validateFailed,omitempty"`
	MinLengthValidateFailed     *CanAccessSuperData       `json:"minLengthValidateFailed,omitempty"`
	MaxLengthValidateFailed     *CanAccessSuperData       `json:"maxLengthValidateFailed,omitempty"`
	Success                     *ClassName                `json:"success,omitempty"`
	Failed                      *ClassName                `json:"failed,omitempty"`
	IsAlpha                     *ClassName                `json:"isAlpha,omitempty"`
	IsAlphanumeric              *ClassName                `json:"isAlphanumeric,omitempty"`
	IsEmail                     *ClassName                `json:"isEmail,omitempty"`
	IsFloat                     *ClassName                `json:"isFloat,omitempty"`
	IsInt                       *ClassName                `json:"isInt,omitempty"`
	IsJSON                      *ClassName                `json:"isJson,omitempty"`
	IsLength                    *ClassName                `json:"isLength,omitempty"`
	IsNumeric                   *ClassName                `json:"isNumeric,omitempty"`
	IsRequired                  *ClassName                `json:"isRequired,omitempty"`
	IsURL                       *ClassName                `json:"isUrl,omitempty"`
	MatchRegexp                 *ClassName                `json:"matchRegexp,omitempty"`
	MatchRegexp2                *ClassName                `json:"matchRegexp2,omitempty"`
	MatchRegexp3                *ClassName                `json:"matchRegexp3,omitempty"`
	MatchRegexp4                *ClassName                `json:"matchRegexp4,omitempty"`
	MatchRegexp5                *ClassName                `json:"matchRegexp5,omitempty"`
	MaxLength                   *ClassName                `json:"maxLength,omitempty"`
	Maximum                     *ClassName                `json:"maximum,omitempty"`
	MinLength                   *ClassName                `json:"minLength,omitempty"`
	Minimum                     *ClassName                `json:"minimum,omitempty"`
	IsDateTimeSame              *ClassName                `json:"isDateTimeSame,omitempty"`
	IsDateTimeBefore            *ClassName                `json:"isDateTimeBefore,omitempty"`
	IsDateTimeAfter             *ClassName                `json:"isDateTimeAfter,omitempty"`
	IsDateTimeSameOrBefore      *ClassName                `json:"isDateTimeSameOrBefore,omitempty"`
	IsDateTimeSameOrAfter       *ClassName                `json:"isDateTimeSameOrAfter,omitempty"`
	IsDateTimeBetween           *ClassName                `json:"isDateTimeBetween,omitempty"`
	IsTimeSame                  *ClassName                `json:"isTimeSame,omitempty"`
	IsTimeBefore                *ClassName                `json:"isTimeBefore,omitempty"`
	IsTimeAfter                 *ClassName                `json:"isTimeAfter,omitempty"`
	IsTimeSameOrBefore          *ClassName                `json:"isTimeSameOrBefore,omitempty"`
	IsTimeSameOrAfter           *ClassName                `json:"isTimeSameOrAfter,omitempty"`
	IsTimeBetween               *ClassName                `json:"isTimeBetween,omitempty"`
	Top                         *ClassName                `json:"top,omitempty"`
	Left                        *ClassName                `json:"left,omitempty"`
	ArrayFormat                 *TooltipPlacement         `json:"arrayFormat,omitempty"`
	Indices                     *ClassName                `json:"indices,omitempty"`
	AllowDots                   *ClassName                `json:"allowDots,omitempty"`
	ClassName                   *LabelClassName           `json:"className,omitempty"`
	Title                       *LabelClassName           `json:"title,omitempty"`
	TitleClassName              *Then                     `json:"titleClassName,omitempty"`
	SubTitle                    *ColumnClassName          `json:"subTitle,omitempty"`
	SubTitleClassName           *Then                     `json:"subTitleClassName,omitempty"`
	SubTitlePlaceholder         *ClassName                `json:"subTitlePlaceholder,omitempty"`
	Description                 *LabelClassName           `json:"description,omitempty"`
	DescriptionPlaceholder      *CanAccessSuperData       `json:"descriptionPlaceholder,omitempty"`
	DescriptionClassName        *ColumnClassName          `json:"descriptionClassName,omitempty"`
	Desc                        *Desc                     `json:"desc,omitempty"`
	DescPlaceholder             *Then                     `json:"descPlaceholder,omitempty"`
	DescClassName               *Then                     `json:"descClassName,omitempty"`
	Avatar                      *ColumnClassName          `json:"avatar,omitempty"`
	AvatarText                  *Then                     `json:"avatarText,omitempty"`
	AvatarTextBackground        *AvatarTextBackground     `json:"avatarTextBackground,omitempty"`
	AvatarTextClassName         *Then                     `json:"avatarTextClassName,omitempty"`
	AvatarClassName             *ColumnClassName          `json:"avatarClassName,omitempty"`
	ImageClassName              *ColumnClassName          `json:"imageClassName,omitempty"`
	Highlight                   *ColumnClassName          `json:"highlight,omitempty"`
	HighlightClassName          *Then                     `json:"highlightClassName,omitempty"`
	Href                        *ColumnClassName          `json:"href,omitempty"`
	Blank                       *CanAccessSuperData       `json:"blank,omitempty"`
	Root                        *ClassName                `json:"root,omitempty"`
	Show                        *ClassName                `json:"show,omitempty"`
	Expand                      *Align                    `json:"expand,omitempty"`
	ExpandAll                   *CanAccessSuperData       `json:"expandAll,omitempty"`
	Accordion                   *CanAccessSuperData       `json:"accordion,omitempty"`
	Type                        *Align                    `json:"type,omitempty"`
	URL                         *ColumnClassName          `json:"url,omitempty"`
	Position                    *Align                    `json:"position,omitempty"`
	AutoPlay                    *CanAccessSuperData       `json:"autoPlay,omitempty"`
	IsLive                      *CanAccessSuperData       `json:"isLive,omitempty"`
	Poster                      *ColumnClassName          `json:"poster,omitempty"`
	Prev                        *Then                     `json:"prev,omitempty"`
	Next                        *Then                     `json:"next,omitempty"`
	Source                      *ClassName                `json:"source,omitempty"`
	Options                     *Options                  `json:"options,omitempty"`
	EvalMode                    *CanAccessSuperData       `json:"evalMode,omitempty"`
	MixedMode                   *CanAccessSuperData       `json:"mixedMode,omitempty"`
	Variables                   *LabelClassName           `json:"variables,omitempty"`
	VariableMode                *Align                    `json:"variableMode,omitempty"`
	Functions                   *LabelClassName           `json:"functions,omitempty"`
	Header                      *CanAccessSuperData       `json:"header,omitempty"`
	InputMode                   *Align                    `json:"inputMode,omitempty"`
	AllowInput                  *CanAccessSuperData       `json:"allowInput,omitempty"`
	Icon                        *ColumnClassName          `json:"icon,omitempty"`
	BtnLabel                    *CanAccessSuperData       `json:"btnLabel,omitempty"`
	Level                       *Align                    `json:"level,omitempty"`
	BtnSize                     *Align                    `json:"btnSize,omitempty"`
	BorderMode                  *Align                    `json:"borderMode,omitempty"`
	Placeholder                 *CanAccessSuperData       `json:"placeholder,omitempty"`
	VariableClassName           *CanAccessSuperData       `json:"variableClassName,omitempty"`
	FunctionClassName           *CanAccessSuperData       `json:"functionClassName,omitempty"`
	SelfVariableName            *CanAccessSuperData       `json:"selfVariableName,omitempty"`
	InputSettings               *ColumnClassName          `json:"inputSettings,omitempty"`
	Remark                      *ColumnClassName          `json:"remark,omitempty"`
	LabelRemark                 *ColumnClassName          `json:"labelRemark,omitempty"`
	Size                        *Align                    `json:"size,omitempty"`
	Label                       *EllipsisThreshold        `json:"label,omitempty"`
	LabelAlign                  *ColumnClassName          `json:"labelAlign,omitempty"`
	LabelWidth                  *Height                   `json:"labelWidth,omitempty"`
	LabelClassName              *CanAccessSuperData       `json:"labelClassName,omitempty"`
	Name                        *CanAccessSuperData       `json:"name,omitempty"`
	ExtraName                   *CanAccessSuperData       `json:"extraName,omitempty"`
	Hint                        *CanAccessSuperData       `json:"hint,omitempty"`
	SubmitOnChange              *CanAccessSuperData       `json:"submitOnChange,omitempty"`
	ReadOnly                    *CanAccessSuperData       `json:"readOnly,omitempty"`
	ReadOnlyOn                  *CanAccessSuperData       `json:"readOnlyOn,omitempty"`
	ValidateOnChange            *CanAccessSuperData       `json:"validateOnChange,omitempty"`
	Mode                        *Align                    `json:"mode,omitempty"`
	Horizontal                  *ColumnClassName          `json:"horizontal,omitempty"`
	Inline                      *CanAccessSuperData       `json:"inline,omitempty"`
	InputClassName              *ColumnClassName          `json:"inputClassName,omitempty"`
	Required                    *CanAccessSuperData       `json:"required,omitempty"`
	ValidationErrors            *ValidationErrors         `json:"validationErrors,omitempty"`
	Validations                 *Validations              `json:"validations,omitempty"`
	Value                       *DataClass                `json:"value,omitempty"`
	ClearValueOnHidden          *CanAccessSuperData       `json:"clearValueOnHidden,omitempty"`
	ValidateAPI                 *Searchable               `json:"validateApi,omitempty"`
	AutoFill                    *AutoFill                 `json:"autoFill,omitempty"`
	InitAutoFill                *InitAutoFill             `json:"initAutoFill,omitempty"`
	Row                         *ClassName                `json:"row,omitempty"`
	ID                          *CanAccessSuperData       `json:"$$id,omitempty"`
	Ref                         *CanAccessSuperData       `json:"$ref,omitempty"`
	Disabled                    *CanAccessSuperData       `json:"disabled,omitempty"`
	DisabledOn                  *ColumnClassName          `json:"disabledOn,omitempty"`
	Hidden                      *CanAccessSuperData       `json:"hidden,omitempty"`
	HiddenOn                    *ColumnClassName          `json:"hiddenOn,omitempty"`
	Visible                     *CanAccessSuperData       `json:"visible,omitempty"`
	VisibleOn                   *ColumnClassName          `json:"visibleOn,omitempty"`
	PropertiesID                *CanAccessSuperData       `json:"id,omitempty"`
	OnEvent                     *OnEvent                  `json:"onEvent,omitempty"`
	Static                      *CanAccessSuperData       `json:"static,omitempty"`
	StaticOn                    *ColumnClassName          `json:"staticOn,omitempty"`
	StaticPlaceholder           *CanAccessSuperData       `json:"staticPlaceholder,omitempty"`
	StaticClassName             *ColumnClassName          `json:"staticClassName,omitempty"`
	StaticLabelClassName        *ColumnClassName          `json:"staticLabelClassName,omitempty"`
	StaticInputClassName        *ColumnClassName          `json:"staticInputClassName,omitempty"`
	StaticSchema                *Not                      `json:"staticSchema,omitempty"`
	Style                       *CanAccessSuperData       `json:"style,omitempty"`
	EditorSetting               *EditorSetting            `json:"editorSetting,omitempty"`
	UseMobileUI                 *CanAccessSuperData       `json:"useMobileUI,omitempty"`
	TestIDBuilder               *Then                     `json:"testIdBuilder,omitempty"`
	X                           *ClassName                `json:"x,omitempty"`
	Y                           *ClassName                `json:"y,omitempty"`
	LowerCase                   *CanAccessSuperData       `json:"lowerCase,omitempty"`
	UpperCase                   *CanAccessSuperData       `json:"upperCase,omitempty"`
	Width                       *Width                    `json:"width,omitempty"`
	Align                       *Align                    `json:"align,omitempty"`
	FilterOption                *Align                    `json:"filterOption,omitempty"`
	Init                        *ClassName                `json:"init,omitempty"`
	Pending                     *ClassName                `json:"pending,omitempty"`
	Uploading                   *ClassName                `json:"uploading,omitempty"`
	Error                       *ClassName                `json:"error,omitempty"`
	Uploaded                    *ClassName                `json:"uploaded,omitempty"`
	Ready                       *ClassName                `json:"ready,omitempty"`
	LevelExpand                 *CanAccessSuperData       `json:"levelExpand,omitempty"`
	EnableClipboard             *CanAccessSuperData       `json:"enableClipboard,omitempty"`
	IconStyle                   *Align                    `json:"iconStyle,omitempty"`
	QuotesOnKeys                *CanAccessSuperData       `json:"quotesOnKeys,omitempty"`
	SortKeys                    *CanAccessSuperData       `json:"sortKeys,omitempty"`
	EllipsisThreshold           *EllipsisThreshold        `json:"ellipsisThreshold,omitempty"`
	MaxHeight                   *CanAccessSuperData       `json:"maxHeight,omitempty"`
	MaxWidth                    *CanAccessSuperData       `json:"maxWidth,omitempty"`
	AspectRatioLabel            *CanAccessSuperData       `json:"aspectRatioLabel,omitempty"`
	AspectRatio                 *CanAccessSuperData       `json:"aspectRatio,omitempty"`
	Height                      *CanAccessSuperData       `json:"height,omitempty"`
	MinHeight                   *CanAccessSuperData       `json:"minHeight,omitempty"`
	MinWidth                    *CanAccessSuperData       `json:"minWidth,omitempty"`
	ErrorMode                   *Align                    `json:"errorMode,omitempty"`
	Delimiter                   *CanAccessSuperData       `json:"delimiter,omitempty"`
	MatchFunc                   *OnClick                  `json:"matchFunc,omitempty"`
	Mini                        *CanAccessSuperData       `json:"mini,omitempty"`
	Enhance                     *CanAccessSuperData       `json:"enhance,omitempty"`
	Clearable                   *CanAccessSuperData       `json:"clearable,omitempty"`
	SearchImediately            *CanAccessSuperData       `json:"searchImediately,omitempty"`
	ValueField                  *CanAccessSuperData       `json:"valueField,omitempty"`
	Sticky                      *CanAccessSuperData       `json:"sticky,omitempty"`
	PullingText                 *ClassName                `json:"pullingText,omitempty"`
	LoosingText                 *ClassName                `json:"loosingText,omitempty"`
	MaxTagCount                 *CanAccessSuperData       `json:"maxTagCount,omitempty"`
	DisplayPosition             *Trigger                  `json:"displayPosition,omitempty"`
	OverflowTagPopover          *ColumnClassName          `json:"overflowTagPopover,omitempty"`
	OverflowTagPopoverInCRUD    *ColumnClassName          `json:"overflowTagPopoverInCRUD,omitempty"`
	Actions                     *ActionsClass             `json:"actions,omitempty"`
	Body                        *ColumnClassName          `json:"body,omitempty"`
	Tabs                        *Not                      `json:"tabs,omitempty"`
	FieldSet                    *Not                      `json:"fieldSet,omitempty"`
	Data                        *Not                      `json:"data,omitempty"`
	Debug                       *CanAccessSuperData       `json:"debug,omitempty"`
	DebugConfig                 *DebugConfig              `json:"debugConfig,omitempty"`
	InitAPI                     *Searchable               `json:"initApi,omitempty"`
	InitAsyncAPI                *Searchable               `json:"initAsyncApi,omitempty"`
	InitFinishedField           *CanAccessSuperData       `json:"initFinishedField,omitempty"`
	InitCheckInterval           *CanAccessSuperData       `json:"initCheckInterval,omitempty"`
	InitFetch                   *CanAccessSuperData       `json:"initFetch,omitempty"`
	InitFetchOn                 *CanAccessSuperData       `json:"initFetchOn,omitempty"`
	Interval                    *CanAccessSuperData       `json:"interval,omitempty"`
	SilentPolling               *CanAccessSuperData       `json:"silentPolling,omitempty"`
	StopAutoRefreshWhen         *CanAccessSuperData       `json:"stopAutoRefreshWhen,omitempty"`
	PersistData                 *CanAccessSuperData       `json:"persistData,omitempty"`
	PersistDataKeys             *RequiredClass            `json:"persistDataKeys,omitempty"`
	ClearPersistDataAfterSubmit *CanAccessSuperData       `json:"clearPersistDataAfterSubmit,omitempty"`
	API                         *Searchable               `json:"api,omitempty"`
	Feedback                    *DataClass                `json:"feedback,omitempty"`
	AsyncAPI                    *Searchable               `json:"asyncApi,omitempty"`
	CheckInterval               *CanAccessSuperData       `json:"checkInterval,omitempty"`
	FinishedField               *CanAccessSuperData       `json:"finishedField,omitempty"`
	ResetAfterSubmit            *CanAccessSuperData       `json:"resetAfterSubmit,omitempty"`
	ClearAfterSubmit            *CanAccessSuperData       `json:"clearAfterSubmit,omitempty"`
	ColumnCount                 *CanAccessSuperData       `json:"columnCount,omitempty"`
	AutoFocus                   *CanAccessSuperData       `json:"autoFocus,omitempty"`
	Messages                    *Messages                 `json:"messages,omitempty"`
	PanelClassName              *ColumnClassName          `json:"panelClassName,omitempty"`
	PrimaryField                *CanAccessSuperData       `json:"primaryField,omitempty"`
	Redirect                    *ClassName                `json:"redirect,omitempty"`
	Reload                      *ClassName                `json:"reload,omitempty"`
	SubmitOnInit                *CanAccessSuperData       `json:"submitOnInit,omitempty"`
	SubmitText                  *CanAccessSuperData       `json:"submitText,omitempty"`
	Target                      *CanAccessSuperData       `json:"target,omitempty"`
	WrapWithPanel               *CanAccessSuperData       `json:"wrapWithPanel,omitempty"`
	AffixFooter                 *CanAccessSuperData       `json:"affixFooter,omitempty"`
	PromptPageLeave             *CanAccessSuperData       `json:"promptPageLeave,omitempty"`
	PromptPageLeaveMessage      *CanAccessSuperData       `json:"promptPageLeaveMessage,omitempty"`
	Rules                       *Rules                    `json:"rules,omitempty"`
	PreventEnterSubmit          *CanAccessSuperData       `json:"preventEnterSubmit,omitempty"`
	Testid                      *ClassName                `json:"testid,omitempty"`
	Layout                      *Layout                   `json:"layout,omitempty"`
	MaxButtons                  *LazyRenderAfter          `json:"maxButtons,omitempty"`
	PerPageAvailable            *RequiredClass            `json:"perPageAvailable,omitempty"`
	PopOverContainerSelector    *PopOverContainerSelector `json:"popOverContainerSelector,omitempty"`
	Enable                      *ColumnClassName          `json:"enable,omitempty"`
	LoadDataOnce                *CanAccessSuperData       `json:"loadDataOnce,omitempty"`
	Prototype                   *Not                      `json:"prototype,omitempty"`
	Length                      *ClassName                `json:"length,omitempty"`
	Arguments                   *Not                      `json:"arguments,omitempty"`
	Caller                      *Then                     `json:"caller,omitempty"`
}

type AutoFill struct {
	AnyOf       []AutoFillAnyOf `json:"anyOf"`
	Description string          `json:"description"`
}

type AutoFillAnyOf struct {
	Type                 TypeElement                    `json:"type"`
	AdditionalProperties *AmbitiousAdditionalProperties `json:"additionalProperties"`
	Properties           *Properties5                   `json:"properties,omitempty"`
}

type Properties5 struct {
	ShowSuggestion   CanAccessSuperData       `json:"showSuggestion"`
	DefaultSelection DataClass                `json:"defaultSelection"`
	API              Searchable               `json:"api"`
	Silent           PopOverContainerSelector `json:"silent"`
	FillMappinng     CanAccessSuperData       `json:"fillMappinng"`
	Trigger          Align                    `json:"trigger"`
	Mode             Align                    `json:"mode"`
	Position         CanAccessSuperData       `json:"position"`
	Size             CanAccessSuperData       `json:"size"`
	Columns          Columns                  `json:"columns"`
	Filter           DataClass                `json:"filter"`
}

type DebugConfig struct {
	Type                 TypeElement           `json:"type"`
	Properties           DebugConfigProperties `json:"properties"`
	AdditionalProperties bool                  `json:"additionalProperties"`
	Description          string                `json:"description"`
}

type DebugConfigProperties struct {
	LevelExpand       CanAccessSuperData `json:"levelExpand"`
	EnableClipboard   CanAccessSuperData `json:"enableClipboard"`
	IconStyle         Align              `json:"iconStyle"`
	QuotesOnKeys      CanAccessSuperData `json:"quotesOnKeys"`
	SortKeys          CanAccessSuperData `json:"sortKeys"`
	EllipsisThreshold EllipsisThreshold  `json:"ellipsisThreshold"`
}

type EllipsisThreshold struct {
	AnyOf       []EllipsisThresholdAnyOf `json:"anyOf"`
	Description string                   `json:"description"`
}

type EllipsisThresholdAnyOf struct {
	Type  TypeElement `json:"type"`
	Const *bool       `json:"const,omitempty"`
}

type InitAutoFill struct {
	AnyOf []ColumnRatioAnyOf `json:"anyOf"`
}

type Layout struct {
	AnyOf       []Funcs `json:"anyOf"`
	Description string  `json:"description"`
	Default     string  `json:"default"`
}

type Messages struct {
	Type                 TypeElement        `json:"type"`
	Properties           MessagesProperties `json:"properties"`
	AdditionalProperties bool               `json:"additionalProperties"`
	Description          string             `json:"description"`
}

type MessagesProperties struct {
	ValidateFailed CanAccessSuperData `json:"validateFailed"`
}

type Rules struct {
	Type        TypeElement `json:"type"`
	Items       RulesItems  `json:"items"`
	Description string      `json:"description"`
}

type RulesItems struct {
	Type                 TypeElement `json:"type"`
	Properties           Properties6 `json:"properties"`
	Required             []string    `json:"required"`
	AdditionalProperties bool        `json:"additionalProperties"`
}

type Properties6 struct {
	Rule    ClassName `json:"rule"`
	Message ClassName `json:"message"`
	Name    Name      `json:"name"`
}

type ValidationErrors struct {
	Type        TypeElement          `json:"type"`
	Properties  map[string]ClassName `json:"properties"`
	Description string               `json:"description"`
}

type Validations struct {
	AnyOf []ValidationsAnyOf `json:"anyOf"`
}

type ValidationsAnyOf struct {
	Type       TypeElement  `json:"type"`
	Properties *Properties7 `json:"properties,omitempty"`
}

type Properties7 struct {
	IsAlpha                CanAccessSuperData `json:"isAlpha"`
	IsAlphanumeric         CanAccessSuperData `json:"isAlphanumeric"`
	IsEmail                CanAccessSuperData `json:"isEmail"`
	IsFloat                CanAccessSuperData `json:"isFloat"`
	IsInt                  CanAccessSuperData `json:"isInt"`
	IsJSON                 CanAccessSuperData `json:"isJson"`
	IsLength               CanAccessSuperData `json:"isLength"`
	IsNumeric              CanAccessSuperData `json:"isNumeric"`
	IsRequired             CanAccessSuperData `json:"isRequired"`
	IsURL                  CanAccessSuperData `json:"isUrl"`
	MatchRegexp            CanAccessSuperData `json:"matchRegexp"`
	MatchRegexp1           CanAccessSuperData `json:"matchRegexp1"`
	MatchRegexp2           CanAccessSuperData `json:"matchRegexp2"`
	MatchRegexp3           CanAccessSuperData `json:"matchRegexp3"`
	MatchRegexp4           CanAccessSuperData `json:"matchRegexp4"`
	MatchRegexp5           CanAccessSuperData `json:"matchRegexp5"`
	MaxLength              CanAccessSuperData `json:"maxLength"`
	Maximum                CanAccessSuperData `json:"maximum"`
	MinLength              CanAccessSuperData `json:"minLength"`
	Minimum                CanAccessSuperData `json:"minimum"`
	IsDateTimeSame         Is                 `json:"isDateTimeSame"`
	IsDateTimeBefore       Is                 `json:"isDateTimeBefore"`
	IsDateTimeAfter        Is                 `json:"isDateTimeAfter"`
	IsDateTimeSameOrBefore Is                 `json:"isDateTimeSameOrBefore"`
	IsDateTimeSameOrAfter  Is                 `json:"isDateTimeSameOrAfter"`
	IsDateTimeBetween      Is                 `json:"isDateTimeBetween"`
	IsTimeSame             Is                 `json:"isTimeSame"`
	IsTimeBefore           Is                 `json:"isTimeBefore"`
	IsTimeAfter            Is                 `json:"isTimeAfter"`
	IsTimeSameOrBefore     Is                 `json:"isTimeSameOrBefore"`
	IsTimeSameOrAfter      Is                 `json:"isTimeSameOrAfter"`
	IsTimeBetween          Is                 `json:"isTimeBetween"`
}

type Width struct {
	Type        *TypeUnion `json:"type"`
	Description string     `json:"description"`
}

type TypeElement string

const (
	Array   TypeElement = "array"
	Boolean TypeElement = "boolean"
	Null    TypeElement = "null"
	Number  TypeElement = "number"
	Object  TypeElement = "object"
	String  TypeElement = "string"
)

type ColumnClassNameRef string

const (
	DefinitionsBadgeObject                ColumnClassNameRef = "#/definitions/BadgeObject"
	DefinitionsClassName                  ColumnClassNameRef = "#/definitions/ClassName"
	DefinitionsDialogSchemaBase           ColumnClassNameRef = "#/definitions/DialogSchemaBase"
	DefinitionsDrawerSchemaBase           ColumnClassNameRef = "#/definitions/DrawerSchemaBase"
	DefinitionsFormHorizontal             ColumnClassNameRef = "#/definitions/FormHorizontal"
	DefinitionsFormulaPickerInputSettings ColumnClassNameRef = "#/definitions/FormulaPickerInputSettings"
	DefinitionsLabelAlign                 ColumnClassNameRef = "#/definitions/LabelAlign"
	DefinitionsSchemaAPI                  ColumnClassNameRef = "#/definitions/SchemaApi"
	DefinitionsSchemaClassName            ColumnClassNameRef = "#/definitions/SchemaClassName"
	DefinitionsSchemaCollection           ColumnClassNameRef = "#/definitions/SchemaCollection"
	DefinitionsSchemaCopyable             ColumnClassNameRef = "#/definitions/SchemaCopyable"
	DefinitionsSchemaExpression           ColumnClassNameRef = "#/definitions/SchemaExpression"
	DefinitionsSchemaIcon                 ColumnClassNameRef = "#/definitions/SchemaIcon"
	DefinitionsSchemaPopOver              ColumnClassNameRef = "#/definitions/SchemaPopOver"
	DefinitionsSchemaQuickEdit            ColumnClassNameRef = "#/definitions/SchemaQuickEdit"
	DefinitionsSchemaReload               ColumnClassNameRef = "#/definitions/SchemaReload"
	DefinitionsSchemaRemark               ColumnClassNameRef = "#/definitions/SchemaRemark"
	DefinitionsSchemaURLPath              ColumnClassNameRef = "#/definitions/SchemaUrlPath"
	DefinitionsToastSchemaBase            ColumnClassNameRef = "#/definitions/ToastSchemaBase"
	DefinitionsTooltipWrapperSchema       ColumnClassNameRef = "#/definitions/TooltipWrapperSchema"
	PurpleDefinitionsSchemaTpl            ColumnClassNameRef = "#/definitions/SchemaTpl"
)

type DescRef string

const (
	DefinitionsBaseAPIObject      DescRef = "#/definitions/BaseApiObject"
	DefinitionsStepStatus         DescRef = "#/definitions/StepStatus"
	FluffyDefinitionsSchemaTpl    DescRef = "#/definitions/SchemaTpl"
	PurpleDefinitionsSchemaObject DescRef = "#/definitions/SchemaObject"
)

type EditorSettingDescription string

const (
	EditorConfigurationCanBeIgnoredAtRuntime EditorSettingDescription = "Editor configuration, can be ignored at runtime"
)

type OnClickDescription string

const (
	CustomEventHandlerFunction OnClickDescription = "Custom event handler function"
	SearchMatchingFunction     OnClickDescription = "Search matching function"
)

type AdditionalPropertiesRequired string

const (
	Actions    AdditionalPropertiesRequired = "actions"
	PurpleType AdditionalPropertiesRequired = "type"
	Title      AdditionalPropertiesRequired = "title"
)

type OnEventDescription string

const (
	EventActionConfiguration OnEventDescription = "Event action configuration"
)

type AnyOfRequired string

const (
	FluffyType AnyOfRequired = "type"
	Lable      AnyOfRequired = "lable"
	Value      AnyOfRequired = "value"
)

type ItemsRef string

const (
	DefinitionsExpressionComplex  ItemsRef = "#/definitions/ExpressionComplex"
	DefinitionsIconItemSchema     ItemsRef = "#/definitions/IconItemSchema"
	DefinitionsOption             ItemsRef = "#/definitions/Option"
	DefinitionsShortCuts          ItemsRef = "#/definitions/ShortCuts"
	DefinitionsTrigger            ItemsRef = "#/definitions/Trigger"
	FluffyDefinitionsSchemaObject ItemsRef = "#/definitions/SchemaObject"
)

type APIDescription string

const (
	AutoFillAPI                 APIDescription = "Auto-fill API"
	AutofillAPI                 APIDescription = "autofill api"
	ConfigureAjaxSendingAddress APIDescription = "Configure ajax sending address"
)

type ColumnsDescription string

const (
	ReferToTheEntryDisplayedItems ColumnsDescription = "Refer to the entry displayed items"
)

type SchemaAdditionalProperties struct {
	Bool                       *bool
	PurpleAdditionalProperties *PurpleAdditionalProperties
}

func (x *SchemaAdditionalProperties) UnmarshalJSON(data []byte) error {
	x.PurpleAdditionalProperties = nil
	var c PurpleAdditionalProperties
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.PurpleAdditionalProperties = &c
	}
	return nil
}

func (x *SchemaAdditionalProperties) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, x.PurpleAdditionalProperties != nil, x.PurpleAdditionalProperties, false, nil, false, nil, false)
}

type IndigoAdditionalProperties struct {
	Bool                       *bool
	FluffyAdditionalProperties *FluffyAdditionalProperties
}

func (x *IndigoAdditionalProperties) UnmarshalJSON(data []byte) error {
	x.FluffyAdditionalProperties = nil
	var c FluffyAdditionalProperties
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffyAdditionalProperties = &c
	}
	return nil
}

func (x *IndigoAdditionalProperties) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, x.FluffyAdditionalProperties != nil, x.FluffyAdditionalProperties, false, nil, false, nil, false)
}

type ConstElement struct {
	Integer *int64
	String  *string
}

func (x *ConstElement) UnmarshalJSON(data []byte) error {
	object, err := unmarshalUnion(data, &x.Integer, nil, nil, &x.String, false, nil, false, nil, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *ConstElement) MarshalJSON() ([]byte, error) {
	return marshalUnion(x.Integer, nil, nil, x.String, false, nil, false, nil, false, nil, false, nil, false)
}

type PropertyAdditionalProperties struct {
	Bool                          *bool
	TentacledAdditionalProperties *TentacledAdditionalProperties
}

func (x *PropertyAdditionalProperties) UnmarshalJSON(data []byte) error {
	x.TentacledAdditionalProperties = nil
	var c TentacledAdditionalProperties
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.TentacledAdditionalProperties = &c
	}
	return nil
}

func (x *PropertyAdditionalProperties) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, x.TentacledAdditionalProperties != nil, x.TentacledAdditionalProperties, false, nil, false, nil, false)
}

type IndecentAdditionalProperties struct {
	Bool                       *bool
	StickyAdditionalProperties *StickyAdditionalProperties
}

func (x *IndecentAdditionalProperties) UnmarshalJSON(data []byte) error {
	x.StickyAdditionalProperties = nil
	var c StickyAdditionalProperties
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.StickyAdditionalProperties = &c
	}
	return nil
}

func (x *IndecentAdditionalProperties) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, x.StickyAdditionalProperties != nil, x.StickyAdditionalProperties, false, nil, false, nil, false)
}

type TypeUnion struct {
	Enum      *TypeElement
	EnumArray []TypeElement
}

func (x *TypeUnion) UnmarshalJSON(data []byte) error {
	x.EnumArray = nil
	x.Enum = nil
	object, err := unmarshalUnion(data, nil, nil, nil, nil, true, &x.EnumArray, false, nil, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *TypeUnion) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, x.EnumArray != nil, x.EnumArray, false, nil, false, nil, x.Enum != nil, x.Enum, false)
}

type HilariousAdditionalProperties struct {
	Bool *bool
	Desc *Desc
}

func (x *HilariousAdditionalProperties) UnmarshalJSON(data []byte) error {
	x.Desc = nil
	var c Desc
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.Desc = &c
	}
	return nil
}

func (x *HilariousAdditionalProperties) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, x.Desc != nil, x.Desc, false, nil, false, nil, false)
}

type EnumUnion struct {
	Bool   *bool
	String *string
}

func (x *EnumUnion) UnmarshalJSON(data []byte) error {
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, &x.String, false, nil, false, nil, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *EnumUnion) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, x.String, false, nil, false, nil, false, nil, false, nil, false)
}

type DefaultUnion struct {
	Bool         *bool
	DefaultClass *DefaultClass
	Integer      *int64
	String       *string
	UnionArray   []ConstElement
}

func (x *DefaultUnion) UnmarshalJSON(data []byte) error {
	x.UnionArray = nil
	x.DefaultClass = nil
	var c DefaultClass
	object, err := unmarshalUnion(data, &x.Integer, nil, &x.Bool, &x.String, true, &x.UnionArray, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.DefaultClass = &c
	}
	return nil
}

func (x *DefaultUnion) MarshalJSON() ([]byte, error) {
	return marshalUnion(x.Integer, nil, x.Bool, x.String, x.UnionArray != nil, x.UnionArray, x.DefaultClass != nil, x.DefaultClass, false, nil, false, nil, false)
}

type AmbitiousAdditionalProperties struct {
	Bool      *bool
	ClassName *ClassName
}

func (x *AmbitiousAdditionalProperties) UnmarshalJSON(data []byte) error {
	x.ClassName = nil
	var c ClassName
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.ClassName = &c
	}
	return nil
}

func (x *AmbitiousAdditionalProperties) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, x.ClassName != nil, x.ClassName, false, nil, false, nil, false)
}

func unmarshalUnion(data []byte, pi **int64, pf **float64, pb **bool, ps **string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) (bool, error) {
	if pi != nil {
			*pi = nil
	}
	if pf != nil {
			*pf = nil
	}
	if pb != nil {
			*pb = nil
	}
	if ps != nil {
			*ps = nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
			return false, err
	}

	switch v := tok.(type) {
	case json.Number:
			if pi != nil {
					i, err := v.Int64()
					if err == nil {
							*pi = &i
							return false, nil
					}
			}
			if pf != nil {
					f, err := v.Float64()
					if err == nil {
							*pf = &f
							return false, nil
					}
					return false, errors.New("Unparsable number")
			}
			return false, errors.New("Union does not contain number")
	case float64:
			return false, errors.New("Decoder should not return float64")
	case bool:
			if pb != nil {
					*pb = &v
					return false, nil
			}
			return false, errors.New("Union does not contain bool")
	case string:
			if haveEnum {
					return false, json.Unmarshal(data, pe)
			}
			if ps != nil {
					*ps = &v
					return false, nil
			}
			return false, errors.New("Union does not contain string")
	case nil:
			if nullable {
					return false, nil
			}
			return false, errors.New("Union does not contain null")
	case json.Delim:
			if v == '{' {
					if haveObject {
							return true, json.Unmarshal(data, pc)
					}
					if haveMap {
							return false, json.Unmarshal(data, pm)
					}
					return false, errors.New("Union does not contain object")
			}
			if v == '[' {
					if haveArray {
							return false, json.Unmarshal(data, pa)
					}
					return false, errors.New("Union does not contain array")
			}
			return false, errors.New("Cannot handle delimiter")
	}
	return false, errors.New("Cannot unmarshal union")
}

func marshalUnion(pi *int64, pf *float64, pb *bool, ps *string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) ([]byte, error) {
	if pi != nil {
			return json.Marshal(*pi)
	}
	if pf != nil {
			return json.Marshal(*pf)
	}
	if pb != nil {
			return json.Marshal(*pb)
	}
	if ps != nil {
			return json.Marshal(*ps)
	}
	if haveArray {
			return json.Marshal(pa)
	}
	if haveObject {
			return json.Marshal(pc)
	}
	if haveMap {
			return json.Marshal(pm)
	}
	if haveEnum {
			return json.Marshal(pe)
	}
	if nullable {
			return json.Marshal(nil)
	}
	return nil, errors.New("Union must not be null")
}
