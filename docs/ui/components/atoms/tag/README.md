# Tag Component

**FILE PURPOSE**: Categorization label and metadata implementation and specifications  
**SCOPE**: All tag variants, interactive features, and management patterns  
**TARGET AUDIENCE**: Developers implementing categorization, labeling, and tagging systems

## 📋 Component Overview

The Tag component provides visual labels for categorization, metadata, and content organization. It supports various styles, interactive features, and management capabilities while maintaining consistency with the overall design system.

### Schema Reference
- **Primary Schema**: `TagSchema.json`
- **Related Schemas**: `BadgeObject.json`, `Options.json`
- **Base Interface**: Display element with semantic categorization

## 🎨 Tag Types

### Basic Tag
**Purpose**: Simple categorization labels

```go
// Basic tag configuration
basicTag := TagProps{
    Text:    "JavaScript",
    Color:   "blue",
    Size:    "md",
    Variant: "solid",
}

// Generated Templ component
templ BasicTag(props TagProps) {
    <span class={ fmt.Sprintf("tag tag-%s tag-%s tag-%s", props.Size, props.Variant, props.Color) }
          role="listitem"
          aria-label={ fmt.Sprintf("Tag: %s", props.Text) }>
        
        if props.Icon != "" {
            <span class="tag-icon">
                @Icon(IconProps{Name: props.Icon, Size: getTagIconSize(props.Size)})
            </span>
        }
        
        <span class="tag-text">{ props.Text }</span>
        
        if props.Count > 0 {
            <span class="tag-count">{ formatCount(props.Count) }</span>
        }
        
        if props.Removable {
            <button 
                type="button"
                class="tag-remove"
                @click={ props.OnRemove }
                aria-label={ fmt.Sprintf("Remove %s tag", props.Text) }>
                @Icon(IconProps{Name: "x", Size: "xs"})
            </button>
        }
    </span>
}
```

### Interactive Tag
**Purpose**: Clickable tags with selection states

```go
interactiveTag := TagProps{
    Text:       "Frontend",
    Color:      "green",
    Clickable:  true,
    Selectable: true,
    OnClick:    "filterByTag",
}

templ InteractiveTag(props TagProps) {
    <span class={ fmt.Sprintf("tag interactive-tag tag-%s tag-%s", props.Size, props.Color) }
          x-data={ fmt.Sprintf(`{
              selected: %t,
              hover: false
          }`, props.Selected) }
          @mouseenter="hover = true"
          @mouseleave="hover = false"
          @click={ fmt.Sprintf("%s('%s')", props.OnClick, props.Text) }
          :class="{ 
              'tag-selected': selected,
              'tag-hover': hover
          }"
          role="button"
          tabindex="0"
          @keydown.enter={ props.OnClick }
          @keydown.space.prevent={ props.OnClick }>
        
        if props.Selectable {
            <span class="tag-checkbox" :class="{ 'checked': selected }">
                <span class="checkbox-icon" x-show="selected">
                    @Icon(IconProps{Name: "check", Size: "xs"})
                </span>
            </span>
        }
        
        if props.Icon != "" {
            <span class="tag-icon">
                @Icon(IconProps{Name: props.Icon, Size: getTagIconSize(props.Size)})
            </span>
        }
        
        <span class="tag-text">{ props.Text }</span>
        
        if props.Count > 0 {
            <span class="tag-count" x-text="selected ? `${props.Count} selected` : props.Count">
                { formatCount(props.Count) }
            </span>
        }
    </span>
}
```

### Tag Input
**Purpose**: Editable tag creation and management

```go
tagInput := TagInputProps{
    Placeholder: "Add tags...",
    MaxTags:     10,
    Suggestions: []string{"React", "Vue", "Angular", "JavaScript", "TypeScript"},
    Separator:   ",",
}

templ TagInput(props TagInputProps) {
    <div class="tag-input-container" 
         x-data={ fmt.Sprintf(`{
             tags: %s,
             inputValue: '',
             suggestions: %s,
             showSuggestions: false,
             filteredSuggestions: [],
             maxTags: %d,
             get canAddMore() {
                 return this.maxTags === 0 || this.tags.length < this.maxTags;
             },
             addTag(tagText) {
                 const trimmedTag = tagText.trim();
                 if (trimmedTag && !this.tags.includes(trimmedTag) && this.canAddMore) {
                     this.tags.push(trimmedTag);
                     this.inputValue = '';
                     this.showSuggestions = false;
                     $dispatch('tags-changed', { tags: this.tags });
                 }
             },
             removeTag(index) {
                 this.tags.splice(index, 1);
                 $dispatch('tags-changed', { tags: this.tags });
             },
             filterSuggestions() {
                 if (this.inputValue.length === 0) {
                     this.filteredSuggestions = [];
                     this.showSuggestions = false;
                     return;
                 }
                 this.filteredSuggestions = this.suggestions.filter(suggestion => 
                     suggestion.toLowerCase().includes(this.inputValue.toLowerCase()) &&
                     !this.tags.includes(suggestion)
                 );
                 this.showSuggestions = this.filteredSuggestions.length > 0;
             },
             handleKeydown(event) {
                 if (event.key === 'Enter' || event.key === '%s') {
                     event.preventDefault();
                     this.addTag(this.inputValue);
                 } else if (event.key === 'Backspace' && this.inputValue === '' && this.tags.length > 0) {
                     this.removeTag(this.tags.length - 1);
                 }
             }
         }`, toJSON(props.InitialTags), toJSON(props.Suggestions), props.MaxTags, props.Separator) }>
        
        <div class="tag-input-wrapper">
            <div class="existing-tags">
                <template x-for="(tag, index) in tags" :key="index">
                    <span class="tag tag-md tag-secondary tag-removable">
                        <span class="tag-text" x-text="tag"></span>
                        <button 
                            type="button"
                            class="tag-remove"
                            @click="removeTag(index)"
                            :aria-label="`Remove ${tag} tag`">
                            @Icon(IconProps{Name: "x", Size: "xs"})
                        </button>
                    </span>
                </template>
            </div>
            
            <input 
                type="text"
                class="tag-input"
                x-model="inputValue"
                @input="filterSuggestions()"
                @keydown="handleKeydown($event)"
                @focus="filterSuggestions()"
                @blur="showSuggestions = false"
                :placeholder="canAddMore ? '{ props.Placeholder }' : 'Maximum tags reached'"
                :disabled="!canAddMore"
                autocomplete="off"
            />
        </div>
        
        <div x-show="showSuggestions" 
             x-transition
             class="tag-suggestions">
            <template x-for="suggestion in filteredSuggestions" :key="suggestion">
                <button 
                    type="button"
                    class="suggestion-item"
                    @mousedown.prevent="addTag(suggestion)"
                    x-text="suggestion">
                </button>
            </template>
        </div>
        
        if props.MaxTags > 0 {
            <div class="tag-counter">
                <span x-text="`${tags.length}/${maxTags} tags`"></span>
            </div>
        }
    </div>
}
```

### Tag Group
**Purpose**: Organized collection of related tags

```go
tagGroup := TagGroupProps{
    Title:      "Skills",
    Tags:       []string{"React", "Node.js", "PostgreSQL", "Docker"},
    Color:      "blue",
    Removable:  true,
    Collapsible: true,
}

templ TagGroup(props TagGroupProps) {
    <div class="tag-group" 
         x-data={ fmt.Sprintf(`{
             collapsed: %t,
             tags: %s,
             removeTag(tagToRemove) {
                 this.tags = this.tags.filter(tag => tag !== tagToRemove);
                 $dispatch('tag-removed', { tag: tagToRemove, group: '%s' });
             }
         }`, props.Collapsed, toJSON(props.Tags), props.Title) }>
        
        <div class="group-header">
            <h4 class="group-title">{ props.Title }</h4>
            
            if props.Count > 0 {
                <span class="group-count">({ formatCount(props.Count) })</span>
            }
            
            if props.Collapsible {
                <button 
                    type="button"
                    class="collapse-toggle"
                    @click="collapsed = !collapsed"
                    :aria-label="collapsed ? 'Expand group' : 'Collapse group'">
                    <span :class="{ 'rotate-180': !collapsed }" class="transition-transform">
                        @Icon(IconProps{Name: "chevron-down", Size: "sm"})
                    </span>
                </button>
            }
        </div>
        
        <div x-show="!collapsed" 
             x-transition
             class="group-content">
            
            <div class="tag-list">
                <template x-for="tag in tags" :key="tag">
                    <span class={ fmt.Sprintf("tag tag-%s tag-%s", props.Size, props.Color) }>
                        <span class="tag-text" x-text="tag"></span>
                        if props.Removable {
                            <button 
                                type="button"
                                class="tag-remove"
                                @click="removeTag(tag)"
                                :aria-label="`Remove ${tag} tag`">
                                @Icon(IconProps{Name: "x", Size: "xs"})
                            </button>
                        }
                    </span>
                </template>
            </div>
            
            if props.AddMore {
                <button 
                    type="button"
                    class="add-tag-button"
                    @click="$dispatch('add-tag-request', { group: '{ props.Title }' })">
                    @Icon(IconProps{Name: "plus", Size: "sm"})
                    <span>Add Tag</span>
                </button>
            }
        </div>
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type TagProps struct {
    // Identity
    ID     string `json:"id"`
    TestID string `json:"testid"`
    
    // Content
    Text  string `json:"text"`   // Tag text
    Icon  string `json:"icon"`   // Optional icon
    Count int    `json:"count"`  // Count/number
    
    // Appearance
    Size     TagSize     `json:"size"`     // xs, sm, md, lg
    Variant  TagVariant  `json:"variant"`  // solid, outline, soft
    Color    TagColor    `json:"color"`    // primary, secondary, success, etc.
    Shape    TagShape    `json:"shape"`    // rounded, pill, square
    
    // States
    Selected   bool `json:"selected"`   // Selection state
    Disabled   bool `json:"disabled"`   // Disabled state
    Loading    bool `json:"loading"`    // Loading state
    
    // Interaction
    Clickable   bool   `json:"clickable"`   // Make tag clickable
    Selectable  bool   `json:"selectable"`  // Allow selection
    Removable   bool   `json:"removable"`   // Show remove button
    OnClick     string `json:"onClick"`     // Click handler
    OnRemove    string `json:"onRemove"`    // Remove handler
    OnSelect    string `json:"onSelect"`    // Select handler
    
    // Styling
    Class string            `json:"className"`
    Style map[string]string `json:"style"`
    
    // Accessibility
    AriaLabel string `json:"ariaLabel"`
    Role      string `json:"role"`
}
```

### Input Properties
```go
type TagInputProps struct {
    // Identity
    Name string `json:"name"`
    ID   string `json:"id"`
    
    // Content
    InitialTags []string `json:"initialTags"`
    Placeholder string   `json:"placeholder"`
    
    // Behavior
    MaxTags     int      `json:"maxTags"`     // Maximum allowed tags
    MinTags     int      `json:"minTags"`     // Minimum required tags
    Separator   string   `json:"separator"`   // Key to separate tags (comma, enter)
    Duplicates  bool     `json:"duplicates"`  // Allow duplicate tags
    
    // Suggestions
    Suggestions     []string `json:"suggestions"`
    SuggestionsAPI  string   `json:"suggestionsAPI"`  // API endpoint for suggestions
    MinSearchLength int      `json:"minSearchLength"` // Minimum characters for suggestions
    
    // Validation
    Pattern     string   `json:"pattern"`     // Regex pattern for validation
    MaxLength   int      `json:"maxLength"`   // Maximum tag length
    Forbidden   []string `json:"forbidden"`   // Forbidden tag values
    
    // Styling
    Size    TagSize  `json:"size"`
    Color   TagColor `json:"color"`
    Variant TagVariant `json:"variant"`
    
    // Events
    OnAdd    string `json:"onAdd"`
    OnRemove string `json:"onRemove"`
    OnChange string `json:"onChange"`
    
    // Accessibility
    AriaLabel       string `json:"ariaLabel"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
}
```

### Group Properties
```go
type TagGroupProps struct {
    // Identity
    ID    string `json:"id"`
    Title string `json:"title"`
    
    // Content
    Tags        []string `json:"tags"`
    Description string   `json:"description"`
    Count       int      `json:"count"`
    
    // Features
    Collapsible bool `json:"collapsible"` // Allow collapse/expand
    Collapsed   bool `json:"collapsed"`   // Initial collapsed state
    Removable   bool `json:"removable"`   // Allow removing tags
    AddMore     bool `json:"addMore"`     // Show add button
    Sortable    bool `json:"sortable"`    // Allow drag-and-drop sorting
    
    // Layout
    Layout  TagLayout `json:"layout"`  // horizontal, vertical, wrap, grid
    Columns int       `json:"columns"` // Grid layout columns
    
    // Appearance
    Size    TagSize    `json:"size"`
    Color   TagColor   `json:"color"`
    Variant TagVariant `json:"variant"`
    
    // Events
    OnTagAdd    string `json:"onTagAdd"`
    OnTagRemove string `json:"onTagRemove"`
    OnCollapse  string `json:"onCollapse"`
}
```

### Size Variants
```go
type TagSize string

const (
    TagXS TagSize = "xs"    // 16px height, tiny text
    TagSM TagSize = "sm"    // 20px height, small text
    TagMD TagSize = "md"    // 24px height, normal text (default)
    TagLG TagSize = "lg"    // 28px height, larger text
)
```

### Visual Variants
```go
type TagVariant string

const (
    TagSolid   TagVariant = "solid"   // Filled background
    TagOutline TagVariant = "outline" // Border only
    TagSoft    TagVariant = "soft"    // Light background
    TagGhost   TagVariant = "ghost"   // Minimal styling
)
```

### Color Options
```go
type TagColor string

const (
    TagPrimary   TagColor = "primary"   // Brand color
    TagSecondary TagColor = "secondary" // Neutral gray
    TagSuccess   TagColor = "success"   // Green
    TagWarning   TagColor = "warning"   // Yellow
    TagDanger    TagColor = "danger"    // Red
    TagInfo      TagColor = "info"      // Blue
    TagPurple    TagColor = "purple"    // Purple
    TagPink      TagColor = "pink"      // Pink
    TagOrange    TagColor = "orange"    // Orange
    TagTeal      TagColor = "teal"      // Teal
)
```

### Shape Options
```go
type TagShape string

const (
    ShapeRounded TagShape = "rounded" // Rounded corners (default)
    ShapePill    TagShape = "pill"    // Fully rounded
    ShapeSquare  TagShape = "square"  // Square corners
)
```

### Layout Types
```go
type TagLayout string

const (
    LayoutHorizontal TagLayout = "horizontal" // Side by side
    LayoutVertical   TagLayout = "vertical"   // Stacked
    LayoutWrap       TagLayout = "wrap"       // Wrap to new lines
    LayoutGrid       TagLayout = "grid"       // Grid layout
)
```

## 🎨 Styling Implementation

### Base Tag Styles
```css
.tag {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    padding: 2px 8px;
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    line-height: 1.2;
    border-radius: var(--radius-md);
    border: 1px solid transparent;
    white-space: nowrap;
    vertical-align: middle;
    transition: var(--transition-base);
    max-width: 200px;
    
    .tag-text {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    
    .tag-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    
    .tag-count {
        font-size: inherit;
        opacity: 0.8;
        font-weight: var(--font-weight-bold);
    }
    
    .tag-remove {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 16px;
        height: 16px;
        margin-left: var(--space-xs);
        background: none;
        border: none;
        color: currentColor;
        cursor: pointer;
        border-radius: var(--radius-xs);
        opacity: 0.7;
        transition: var(--transition-quick);
        
        &:hover {
            opacity: 1;
            background: rgba(255, 255, 255, 0.2);
        }
        
        &:focus {
            outline: none;
            box-shadow: 0 0 0 1px currentColor;
        }
    }
}
```

### Size Variants
```css
/* Extra Small */
.tag-xs {
    padding: 1px 4px;
    font-size: 10px;
    min-height: 16px;
    gap: 2px;
    
    .tag-remove {
        width: 12px;
        height: 12px;
        margin-left: 2px;
    }
}

/* Small */
.tag-sm {
    padding: 2px 6px;
    font-size: 11px;
    min-height: 20px;
    gap: var(--space-xs);
    
    .tag-remove {
        width: 14px;
        height: 14px;
        margin-left: var(--space-xs);
    }
}

/* Medium (Default) */
.tag-md {
    padding: 3px 8px;
    font-size: var(--font-size-xs);
    min-height: 24px;
    gap: var(--space-xs);
    
    .tag-remove {
        width: 16px;
        height: 16px;
        margin-left: var(--space-xs);
    }
}

/* Large */
.tag-lg {
    padding: 4px 12px;
    font-size: var(--font-size-sm);
    min-height: 28px;
    gap: var(--space-sm);
    
    .tag-remove {
        width: 18px;
        height: 18px;
        margin-left: var(--space-sm);
    }
}
```

### Color and Variant Styles
```css
/* Solid variant */
.tag-solid {
    &.tag-primary {
        background: var(--color-primary);
        color: white;
        border-color: var(--color-primary);
    }
    
    &.tag-secondary {
        background: var(--color-bg-secondary);
        color: var(--color-text-secondary);
        border-color: var(--color-border-medium);
    }
    
    &.tag-success {
        background: var(--color-success);
        color: white;
        border-color: var(--color-success);
    }
    
    &.tag-warning {
        background: var(--color-warning);
        color: var(--color-text-primary);
        border-color: var(--color-warning);
    }
    
    &.tag-danger {
        background: var(--color-danger);
        color: white;
        border-color: var(--color-danger);
    }
    
    &.tag-info {
        background: var(--color-info);
        color: white;
        border-color: var(--color-info);
    }
}

/* Outline variant */
.tag-outline {
    background: transparent;
    
    &.tag-primary {
        color: var(--color-primary);
        border-color: var(--color-primary);
    }
    
    &.tag-secondary {
        color: var(--color-text-secondary);
        border-color: var(--color-border-medium);
    }
    
    &.tag-success {
        color: var(--color-success);
        border-color: var(--color-success);
    }
    
    &.tag-warning {
        color: var(--color-warning-dark);
        border-color: var(--color-warning);
    }
    
    &.tag-danger {
        color: var(--color-danger);
        border-color: var(--color-danger);
    }
    
    &.tag-info {
        color: var(--color-info);
        border-color: var(--color-info);
    }
}

/* Soft variant */
.tag-soft {
    border: none;
    
    &.tag-primary {
        background: var(--color-primary-light);
        color: var(--color-primary-dark);
    }
    
    &.tag-secondary {
        background: var(--color-bg-secondary);
        color: var(--color-text-secondary);
    }
    
    &.tag-success {
        background: var(--color-success-light);
        color: var(--color-success-dark);
    }
    
    &.tag-warning {
        background: var(--color-warning-light);
        color: var(--color-warning-dark);
    }
    
    &.tag-danger {
        background: var(--color-danger-light);
        color: var(--color-danger-dark);
    }
    
    &.tag-info {
        background: var(--color-info-light);
        color: var(--color-info-dark);
    }
}

/* Ghost variant */
.tag-ghost {
    background: transparent;
    border: none;
    color: var(--color-text-secondary);
    
    &:hover {
        background: var(--color-bg-hover);
        color: var(--color-text-primary);
    }
}
```

### Interactive Tag Styles
```css
.interactive-tag {
    cursor: pointer;
    transition: var(--transition-base);
    
    &:hover {
        transform: translateY(-1px);
        box-shadow: var(--shadow-sm);
    }
    
    &:focus {
        outline: none;
        box-shadow: 0 0 0 2px var(--color-primary-light);
    }
    
    &:active {
        transform: translateY(0);
    }
    
    &.tag-selected {
        box-shadow: 0 0 0 2px var(--color-primary);
        transform: translateY(-1px);
    }
    
    .tag-checkbox {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 14px;
        height: 14px;
        border: 1px solid var(--color-border-medium);
        border-radius: var(--radius-xs);
        background: var(--color-bg-surface);
        margin-right: var(--space-xs);
        transition: var(--transition-quick);
        
        &.checked {
            background: var(--color-primary);
            border-color: var(--color-primary);
            color: white;
        }
        
        .checkbox-icon {
            font-size: 10px;
        }
    }
}
```

### Tag Input Styles
```css
.tag-input-container {
    position: relative;
    width: 100%;
}

.tag-input-wrapper {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-xs);
    min-height: 36px;
    padding: var(--space-xs) var(--space-sm);
    border: 1px solid var(--color-border-medium);
    border-radius: var(--radius-md);
    background: var(--color-bg-surface);
    transition: var(--transition-base);
    
    &:focus-within {
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    .existing-tags {
        display: flex;
        flex-wrap: wrap;
        gap: var(--space-xs);
    }
    
    .tag-input {
        flex: 1;
        min-width: 120px;
        background: none;
        border: none;
        outline: none;
        font-size: var(--font-size-sm);
        color: var(--color-text-primary);
        
        &::placeholder {
            color: var(--color-text-placeholder);
        }
        
        &:disabled {
            color: var(--color-text-disabled);
            cursor: not-allowed;
        }
    }
}

.tag-suggestions {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    z-index: 10;
    background: var(--color-bg-surface);
    border: 1px solid var(--color-border-medium);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    max-height: 200px;
    overflow-y: auto;
    margin-top: var(--space-xs);
    
    .suggestion-item {
        display: block;
        width: 100%;
        padding: var(--space-sm) var(--space-md);
        background: none;
        border: none;
        text-align: left;
        font-size: var(--font-size-sm);
        color: var(--color-text-primary);
        cursor: pointer;
        transition: var(--transition-quick);
        
        &:hover {
            background: var(--color-bg-hover);
        }
        
        &:focus {
            outline: none;
            background: var(--color-primary-light);
            color: var(--color-primary-dark);
        }
    }
}

.tag-counter {
    font-size: var(--font-size-xs);
    color: var(--color-text-secondary);
    margin-top: var(--space-xs);
}
```

### Tag Group Styles
```css
.tag-group {
    .group-header {
        display: flex;
        align-items: center;
        gap: var(--space-sm);
        margin-bottom: var(--space-md);
        
        .group-title {
            font-size: var(--font-size-sm);
            font-weight: var(--font-weight-semibold);
            color: var(--color-text-primary);
            margin: 0;
        }
        
        .group-count {
            font-size: var(--font-size-xs);
            color: var(--color-text-secondary);
        }
        
        .collapse-toggle {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 20px;
            height: 20px;
            background: none;
            border: none;
            color: var(--color-text-secondary);
            cursor: pointer;
            border-radius: var(--radius-xs);
            transition: var(--transition-quick);
            margin-left: auto;
            
            &:hover {
                background: var(--color-bg-hover);
                color: var(--color-text-primary);
            }
            
            &:focus {
                outline: none;
                box-shadow: 0 0 0 2px var(--color-primary-light);
            }
        }
    }
    
    .group-content {
        .tag-list {
            display: flex;
            flex-wrap: wrap;
            gap: var(--space-xs);
            margin-bottom: var(--space-md);
            
            &.layout-vertical {
                flex-direction: column;
                align-items: flex-start;
            }
            
            &.layout-grid {
                display: grid;
                grid-template-columns: repeat(var(--columns, 3), 1fr);
            }
        }
        
        .add-tag-button {
            display: inline-flex;
            align-items: center;
            gap: var(--space-xs);
            padding: 3px 8px;
            background: transparent;
            border: 1px dashed var(--color-border-medium);
            border-radius: var(--radius-md);
            color: var(--color-text-secondary);
            font-size: var(--font-size-xs);
            cursor: pointer;
            transition: var(--transition-base);
            
            &:hover {
                border-color: var(--color-primary);
                color: var(--color-primary);
                background: var(--color-primary-light);
            }
            
            &:focus {
                outline: none;
                box-shadow: 0 0 0 2px var(--color-primary-light);
            }
        }
    }
}
```

### Shape Variants
```css
/* Shape variants */
.tag-shape-rounded {
    border-radius: var(--radius-md);
}

.tag-shape-pill {
    border-radius: var(--radius-full);
}

.tag-shape-square {
    border-radius: 0;
}
```

## ⚙️ Advanced Features

### Tag Autocomplete
```go
templ TagAutocomplete(props TagInputProps) {
    <div class="tag-autocomplete" 
         x-data={ fmt.Sprintf(`{
            tags: %s,
            inputValue: '',
            suggestions: [],
            selectedIndex: -1,
            loading: false,
            async fetchSuggestions() {
                if (this.inputValue.length < %d) {
                    this.suggestions = [];
                    return;
                }
                
                this.loading = true;
                try {
                    const response = await fetch('%s?q=' + encodeURIComponent(this.inputValue));
                    this.suggestions = await response.json();
                } catch (error) {
                    console.error('Failed to fetch suggestions:', error);
                    this.suggestions = [];
                } finally {
                    this.loading = false;
                }
                this.selectedIndex = -1;
            },
            handleKeydown(event) {
                switch (event.key) {
                    case 'ArrowDown':
                        event.preventDefault();
                        this.selectedIndex = Math.min(this.selectedIndex + 1, this.suggestions.length - 1);
                        break;
                    case 'ArrowUp':
                        event.preventDefault();
                        this.selectedIndex = Math.max(this.selectedIndex - 1, -1);
                        break;
                    case 'Enter':
                        event.preventDefault();
                        if (this.selectedIndex >= 0) {
                            this.addTag(this.suggestions[this.selectedIndex].name);
                        } else if (this.inputValue.trim()) {
                            this.addTag(this.inputValue.trim());
                        }
                        break;
                    case 'Escape':
                        this.suggestions = [];
                        this.selectedIndex = -1;
                        break;
                }
            }
        }`, toJSON(props.InitialTags), props.MinSearchLength, props.SuggestionsAPI) }>
        
        <!-- Tag input implementation with autocomplete -->
    </div>
}
```

### Tag Validation
```go
templ ValidatedTagInput(props TagInputProps) {
    <div class="validated-tag-input" 
         x-data={ fmt.Sprintf(`{
            tags: %s,
            errors: {},
            validateTag(tagText) {
                const errors = [];
                
                // Length validation
                if (tagText.length > %d) {
                    errors.push('Tag is too long (max %d characters)');
                }
                
                // Pattern validation
                const pattern = new RegExp('%s');
                if (!pattern.test(tagText)) {
                    errors.push('Invalid tag format');
                }
                
                // Forbidden values
                const forbidden = %s;
                if (forbidden.includes(tagText.toLowerCase())) {
                    errors.push('This tag is not allowed');
                }
                
                // Duplicate check
                if (!%t && this.tags.includes(tagText)) {
                    errors.push('Tag already exists');
                }
                
                return errors;
            },
            addTag(tagText) {
                const trimmedTag = tagText.trim();
                const validationErrors = this.validateTag(trimmedTag);
                
                if (validationErrors.length > 0) {
                    this.errors[trimmedTag] = validationErrors;
                    return false;
                }
                
                this.tags.push(trimmedTag);
                delete this.errors[trimmedTag];
                return true;
            }
        }`, toJSON(props.InitialTags), props.MaxLength, props.MaxLength, props.Pattern, toJSON(props.Forbidden), props.Duplicates) }>
        
        <!-- Validated tag input implementation -->
    </div>
}
```

### Drag and Drop Tags
```go
templ SortableTagGroup(props TagGroupProps) {
    <div class="sortable-tag-group" 
         x-data={ fmt.Sprintf(`{
            tags: %s,
            draggedTag: null,
            draggedIndex: -1,
            startDrag(tag, index, event) {
                this.draggedTag = tag;
                this.draggedIndex = index;
                event.dataTransfer.effectAllowed = 'move';
                event.dataTransfer.setData('text/html', event.target.outerHTML);
            },
            allowDrop(event) {
                event.preventDefault();
                event.dataTransfer.dropEffect = 'move';
            },
            drop(targetIndex, event) {
                event.preventDefault();
                
                if (this.draggedIndex !== targetIndex) {
                    const draggedTag = this.tags[this.draggedIndex];
                    this.tags.splice(this.draggedIndex, 1);
                    this.tags.splice(targetIndex, 0, draggedTag);
                    
                    $dispatch('tags-reordered', { 
                        tags: this.tags,
                        from: this.draggedIndex,
                        to: targetIndex 
                    });
                }
                
                this.draggedTag = null;
                this.draggedIndex = -1;
            }
        }`, toJSON(props.Tags)) }>
        
        <div class="group-header">
            <h4 class="group-title">{ props.Title }</h4>
        </div>
        
        <div class="sortable-tag-list">
            <template x-for="(tag, index) in tags" :key="tag">
                <span class="tag tag-md tag-secondary sortable-tag"
                      draggable="true"
                      @dragstart="startDrag(tag, index, $event)"
                      @dragover="allowDrop($event)"
                      @drop="drop(index, $event)"
                      :class="{ 'dragging': draggedTag === tag }">
                    
                    <span class="drag-handle">
                        @Icon(IconProps{Name: "grip-vertical", Size: "xs"})
                    </span>
                    
                    <span class="tag-text" x-text="tag"></span>
                </span>
            </template>
        </div>
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific tag styling */
@media (max-width: 479px) {
    .tag {
        font-size: var(--font-size-xs);
        padding: 2px 6px;
        max-width: 150px;
        
        .tag-remove {
            width: 20px;
            height: 20px;
            margin-left: var(--space-xs);
        }
    }
    
    .tag-input-wrapper {
        min-height: 44px;
        padding: var(--space-sm);
        
        .tag-input {
            min-width: 100px;
            font-size: var(--font-size-base);
        }
    }
    
    .tag-group {
        .tag-list {
            gap: var(--space-xs);
            
            &.layout-grid {
                grid-template-columns: repeat(2, 1fr);
            }
        }
    }
    
    /* Stack suggestions full width */
    .tag-suggestions {
        left: var(--space-sm);
        right: var(--space-sm);
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .interactive-tag {
        /* Remove hover effects on touch devices */
        &:hover {
            transform: none;
            box-shadow: none;
        }
        
        /* Enhance tap feedback */
        &:active {
            transform: scale(0.95);
        }
    }
    
    .tag-remove {
        min-width: 32px;
        min-height: 32px;
        
        &:hover {
            background: none;
        }
    }
    
    .sortable-tag {
        /* Disable drag on touch devices */
        user-select: none;
        -webkit-user-select: none;
    }
}
```

## ♿ Accessibility Implementation

### ARIA Support
```go
func (tag TagProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Role based on usage
    if tag.Clickable {
        attrs["role"] = "button"
        attrs["tabindex"] = "0"
    } else if tag.Selectable {
        attrs["role"] = "option"
        attrs["aria-selected"] = fmt.Sprintf("%t", tag.Selected)
    } else {
        attrs["role"] = "listitem"
    }
    
    // Accessible name
    if tag.AriaLabel != "" {
        attrs["aria-label"] = tag.AriaLabel
    } else {
        attrs["aria-label"] = fmt.Sprintf("Tag: %s", tag.Text)
    }
    
    // Disabled state
    if tag.Disabled {
        attrs["aria-disabled"] = "true"
    }
    
    // Remove button accessibility
    if tag.Removable {
        attrs["aria-describedby"] = tag.ID + "-remove"
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleTag(props TagProps) {
    <span class={ props.GetClasses() }
          for attrName, attrValue := range props.GetAriaAttributes() {
              { attrName }={ attrValue }
          }>
        
        <span class="tag-text">{ props.Text }</span>
        
        if props.Count > 0 {
            <span class="tag-count" aria-label={ fmt.Sprintf("%d items", props.Count) }>
                { formatCount(props.Count) }
            </span>
        }
        
        if props.Removable {
            <button 
                type="button"
                class="tag-remove"
                id={ props.ID + "-remove" }
                aria-label={ fmt.Sprintf("Remove %s tag", props.Text) }
                @click={ props.OnRemove }>
                @Icon(IconProps{Name: "x", Size: "xs"})
            </button>
        }
        
        <!-- Screen reader only description -->
        if props.Selected {
            <span class="sr-only">Selected</span>
        }
    </span>
}
```

### Keyboard Navigation
```css
/* Focus management */
.tag {
    &[role="button"],
    &[role="option"] {
        cursor: pointer;
        
        &:focus {
            outline: none;
            box-shadow: 0 0 0 2px var(--color-primary-light);
            border-radius: var(--radius-md);
        }
        
        &:focus-visible {
            box-shadow: 0 0 0 2px var(--color-primary);
        }
    }
}

.tag-remove {
    &:focus {
        outline: none;
        box-shadow: 0 0 0 1px currentColor;
        border-radius: var(--radius-xs);
    }
}

.tag-input {
    &:focus {
        outline: none;
    }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
    .tag {
        border-width: 2px;
        border-style: solid;
        
        &.tag-outline {
            border-width: 2px;
        }
        
        &.tag-solid {
            border-color: currentColor;
        }
    }
    
    .tag-checkbox {
        border-width: 2px;
        border-color: currentColor;
    }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
    .interactive-tag {
        transition: none;
        
        &:hover {
            transform: none;
        }
    }
    
    .tag-suggestions {
        transition: none;
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestTagComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    TagProps
        expected []string
    }{
        {
            name: "basic tag",
            props: TagProps{
                Text:  "JavaScript",
                Color: "blue",
            },
            expected: []string{"tag", "tag-blue", "JavaScript"},
        },
        {
            name: "removable tag",
            props: TagProps{
                Text:      "React",
                Removable: true,
                OnRemove:  "handleRemove",
            },
            expected: []string{"tag-remove", "Remove React tag"},
        },
        {
            name: "selectable tag",
            props: TagProps{
                Text:       "Frontend",
                Selectable: true,
                Selected:   true,
            },
            expected: []string{"role=\"option\"", "aria-selected=\"true\""},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderTag(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Tag Accessibility', () => {
    test('has proper role for interactive tags', () => {
        const tag = render(<Tag text="Filter" clickable onClick={() => {}} />);
        const tagElement = screen.getByRole('button');
        expect(tagElement).toBeInTheDocument();
    });
    
    test('removable tags announce removal action', () => {
        const tag = render(<Tag text="JavaScript" removable onRemove={() => {}} />);
        const removeButton = screen.getByLabelText('Remove JavaScript tag');
        expect(removeButton).toBeInTheDocument();
    });
    
    test('selectable tags announce selection state', () => {
        const tag = render(<Tag text="Frontend" selectable selected />);
        const tagElement = screen.getByRole('option');
        expect(tagElement).toHaveAttribute('aria-selected', 'true');
    });
    
    test('supports keyboard interaction', () => {
        const handleClick = jest.fn();
        const tag = render(<Tag text="JavaScript" clickable onClick={handleClick} />);
        const tagElement = screen.getByRole('button');
        
        tagElement.focus();
        expect(tagElement).toHaveFocus();
        
        fireEvent.keyDown(tagElement, { key: 'Enter' });
        expect(handleClick).toHaveBeenCalled();
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Tag Visual Tests', () => {
    test('all tag variants and states', async ({ page }) => {
        await page.goto('/components/tag');
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`tag-${size}.png`);
        }
        
        // Test all variants
        const variants = ['solid', 'outline', 'soft', 'ghost'];
        for (const variant of variants) {
            await expect(page.locator(`[data-variant="${variant}"]`)).toHaveScreenshot(`tag-${variant}.png`);
        }
        
        // Test all colors
        const colors = ['primary', 'secondary', 'success', 'warning', 'danger'];
        for (const color of colors) {
            await expect(page.locator(`[data-color="${color}"]`)).toHaveScreenshot(`tag-${color}.png`);
        }
        
        // Test interactive states
        await expect(page.locator('[data-state="removable"]')).toHaveScreenshot('tag-removable.png');
        await expect(page.locator('[data-state="selectable"]')).toHaveScreenshot('tag-selectable.png');
    });
});
```

## 📚 Usage Examples

### Skill Tags
```go
templ SkillTags() {
    @TagGroup(TagGroupProps{
        Title:     "Technical Skills",
        Tags:      []string{"JavaScript", "React", "Node.js", "PostgreSQL"},
        Color:     "blue",
        Removable: true,
        AddMore:   true,
    })
}
```

### Filter Tags
```go
templ FilterTags() {
    <div class="filter-section">
        <h4>Active Filters</h4>
        @InteractiveTag(TagProps{
            Text:      "Category: Frontend",
            Color:     "primary",
            Clickable: true,
            Removable: true,
            OnClick:   "editFilter",
            OnRemove:  "removeFilter",
        })
        
        @InteractiveTag(TagProps{
            Text:      "Experience: 2+ years",
            Color:     "secondary",
            Clickable: true,
            Removable: true,
        })
    </div>
}
```

### Tag Input for Categories
```go
templ CategoryInput() {
    @TagInput(TagInputProps{
        Name:        "categories",
        Placeholder: "Add categories...",
        MaxTags:     5,
        Suggestions: []string{"Frontend", "Backend", "Mobile", "DevOps", "Design"},
        Color:       "primary",
        Size:        "md",
    })
}
```

### Status Tags
```go
templ ProjectStatus() {
    <div class="project-meta">
        @BasicTag(TagProps{
            Text:    "In Progress",
            Color:   "warning",
            Variant: "soft",
            Icon:    "clock",
        })
        
        @BasicTag(TagProps{
            Text:  "High Priority",
            Color: "danger",
            Icon:  "alert-triangle",
        })
        
        @BasicTag(TagProps{
            Text:  "Team Project",
            Color: "info",
            Icon:  "users",
            Count: 5,
        })
    </div>
}
```

## 🔗 Related Components

- **[Badge](../badge/)** - Status indicators and counts
- **[Button](../button/)** - Interactive action elements
- **[Chip](../../molecules/chip/)** - Enhanced tag functionality
- **[Filter Group](../../molecules/filter-group/)** - Tag-based filtering

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `TagSchema.json`  
**CSS Classes**: `.tag`, `.tag-{size}`, `.tag-{color}`, `.tag-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant