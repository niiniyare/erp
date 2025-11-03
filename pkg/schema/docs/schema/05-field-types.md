# Field Types Reference

Complete reference for all supported field types in Awo ERP's schema-driven UI system.

## Table of Contents
- [Basic Text Fields](#basic-text-fields)
- [Date & Time Fields](#date--time-fields)
- [Selection Fields](#selection-fields)
- [File & Media Fields](#file--media-fields)
- [Text Content Fields](#text-content-fields)
- [Specialized Fields](#specialized-fields)
- [Relational Fields](#relational-fields)
- [Layout & Container Fields](#layout--container-fields)
- [Display-Only Fields](#display-only-fields)
- [Database Type Mapping](#database-type-mapping)
- [Validation Rules by Type](#validation-rules-by-type)

---

## Basic Text Fields

### `text`
Single-line text input for general text data.

**HTML:** `<input type="text">`  
**Database:** `VARCHAR(255)` or `TEXT`

```json
{
  "name": "username",
  "type": "text",
  "label": "Username",
  "required": true,
  "placeholder": "Enter username",
  "config": {
    "maxLength": 50,
    "autocomplete": "username",
    "prefix": "@",
    "suffix": null
  },
  "validation": {
    "minLength": 3,
    "maxLength": 50,
    "pattern": "^[a-zA-Z0-9_-]+$"
  },
  "default": null
}
```

**Config Options:**
- `placeholder` (string): Placeholder text
- `autocomplete` (string): HTML autocomplete attribute
- `prefix` (string): Text/icon before input
- `suffix` (string): Text/icon after input
- `maxLength` (int): Character limit

**Validation:**
- `minLength`, `maxLength`: Length constraints
- `pattern`: Regex pattern
- `custom`: Custom validation function

---

### `email`
Email input with built-in validation.

**HTML:** `<input type="email">`  
**Database:** `VARCHAR(255)`

```json
{
  "name": "email",
  "type": "email",
  "label": "Email Address",
  "required": true,
  "placeholder": "user@example.com",
  "config": {
    "autocomplete": "email",
    "multiple": false
  },
  "validation": {
    "format": "rfc5322"
  }
}
```

**Config Options:**
- `multiple` (bool): Allow multiple comma-separated emails
- `domains` (array): Whitelist/blacklist domains

**Built-in Validation:** RFC 5322 email format

---

### `password`
Masked password input.

**HTML:** `<input type="password">`  
**Database:** `VARCHAR(255)` (store hashed)

```json
{
  "name": "password",
  "type": "password",
  "label": "Password",
  "required": true,
  "config": {
    "showStrength": true,
    "showToggle": true,
    "autocomplete": "new-password"
  },
  "validation": {
    "minLength": 8,
    "pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).+$"
  }
}
```

**Config Options:**
- `showStrength` (bool): Display password strength indicator
- `showToggle` (bool): Show/hide password toggle
- `confirmField` (string): Name of confirmation field

---

### `number`
Numeric input with optional step controls.

**HTML:** `<input type="number">`  
**Database:** `INTEGER`, `BIGINT`, `NUMERIC`

```json
{
  "name": "quantity",
  "type": "number",
  "label": "Quantity",
  "required": true,
  "config": {
    "step": 1,
    "showStepper": true,
    "precision": 0
  },
  "validation": {
    "min": 0,
    "max": 9999,
    "step": 1
  },
  "default": 1
}
```

**Config Options:**
- `step` (number): Increment/decrement value
- `showStepper` (bool): Show +/- buttons
- `precision` (int): Decimal places (0 for integers)
- `thousandsSeparator` (bool): Format with commas

**Validation:**
- `min`, `max`: Value range
- `step`: Value must be multiple of step

---

### `phone`
Phone number input with formatting.

**HTML:** `<input type="tel">`  
**Database:** `VARCHAR(50)`

```json
{
  "name": "phone",
  "type": "phone",
  "label": "Phone Number",
  "config": {
    "format": "international",
    "defaultCountry": "US",
    "allowExtension": true
  },
  "validation": {
    "pattern": "^\\+?[1-9]\\d{1,14}$"
  }
}
```

**Config Options:**
- `format` (string): `"international"`, `"national"`, `"custom"`
- `defaultCountry` (string): ISO country code
- `allowExtension` (bool): Allow phone extensions

---

### `url`
URL input with validation.

**HTML:** `<input type="url">`  
**Database:** `VARCHAR(2048)` or `TEXT`

```json
{
  "name": "website",
  "type": "url",
  "label": "Website",
  "placeholder": "https://example.com",
  "validation": {
    "protocols": ["http", "https"],
    "requireProtocol": true
  }
}
```

**Config Options:**
- `protocols` (array): Allowed protocols
- `requireProtocol` (bool): Must include http(s)://

---

### `hidden`
Hidden field for storing values not displayed to users.

**HTML:** `<input type="hidden">`  
**Database:** Any type

```json
{
  "name": "user_id",
  "type": "hidden",
  "default": "${session.user_id}"
}
```

**Use Cases:**
- System-generated values
- Session data
- Workflow state

---

### `color`
Color picker input.

**HTML:** `<input type="color">`  
**Database:** `VARCHAR(7)` (hex) or `VARCHAR(20)` (rgba)

```json
{
  "name": "brand_color",
  "type": "color",
  "label": "Brand Color",
  "config": {
    "format": "hex",
    "alpha": false,
    "presets": ["#FF0000", "#00FF00", "#0000FF"]
  },
  "default": "#000000"
}
```

**Config Options:**
- `format` (string): `"hex"`, `"rgb"`, `"rgba"`, `"hsl"`
- `alpha` (bool): Support transparency
- `presets` (array): Quick-select colors

---

## Date & Time Fields

### `date`
Date picker input.

**HTML:** `<input type="date">`  
**Database:** `DATE`

```json
{
  "name": "birth_date",
  "type": "date",
  "label": "Date of Birth",
  "required": true,
  "config": {
    "format": "YYYY-MM-DD",
    "displayFormat": "MM/DD/YYYY",
    "minDate": "1900-01-01",
    "maxDate": "today",
    "firstDayOfWeek": 0
  },
  "validation": {
    "min": "1900-01-01",
    "max": "${today}"
  }
}
```

**Config Options:**
- `format` (string): Internal date format
- `displayFormat` (string): User-facing format
- `minDate`, `maxDate` (string): Date boundaries
- `firstDayOfWeek` (int): 0=Sunday, 1=Monday
- `shortcuts` (bool): Show "Today", "Yesterday" buttons

**Validation:**
- `min`, `max`: Date range
- `businessDays` (bool): Only allow weekdays

---

### `time`
Time picker input.

**HTML:** `<input type="time">`  
**Database:** `TIME`

```json
{
  "name": "meeting_time",
  "type": "time",
  "label": "Meeting Time",
  "config": {
    "format": "HH:mm",
    "use24Hour": true,
    "minuteStep": 15
  },
  "validation": {
    "min": "08:00",
    "max": "18:00"
  }
}
```

**Config Options:**
- `format` (string): Time format
- `use24Hour` (bool): 24-hour vs 12-hour format
- `minuteStep` (int): Minute increment (e.g., 15, 30)

---

### `datetime`
Combined date and time picker.

**HTML:** `<input type="datetime-local">`  
**Database:** `TIMESTAMP`

```json
{
  "name": "scheduled_at",
  "type": "datetime",
  "label": "Schedule Date & Time",
  "config": {
    "format": "YYYY-MM-DD HH:mm:ss",
    "displayFormat": "MM/DD/YYYY hh:mm A",
    "timezone": "UTC"
  },
  "default": "${now}"
}
```

**Config Options:**
- `timezone` (string): Timezone for display/storage
- `utc` (bool): Store in UTC

---

### `daterange`
Date range picker (start and end dates).

**HTML:** Two date inputs  
**Database:** Two `DATE` columns or `DATERANGE`

```json
{
  "name": "project_duration",
  "type": "daterange",
  "label": "Project Duration",
  "config": {
    "startField": "start_date",
    "endField": "end_date",
    "minDays": 1,
    "maxDays": 365,
    "singleCalendar": false
  },
  "validation": {
    "required": true,
    "maxSpan": 365
  }
}
```

**Config Options:**
- `startField`, `endField` (string): Field names for range
- `minDays`, `maxDays` (int): Range constraints
- `singleCalendar` (bool): Show both dates in one calendar

---

### `month`
Month picker (year + month).

**HTML:** `<input type="month">`  
**Database:** `DATE` or `VARCHAR(7)` (YYYY-MM)

```json
{
  "name": "report_month",
  "type": "month",
  "label": "Report Month",
  "config": {
    "format": "YYYY-MM"
  }
}
```

---

### `year`
Year picker.

**HTML:** `<select>` or custom picker  
**Database:** `SMALLINT`

```json
{
  "name": "graduation_year",
  "type": "year",
  "label": "Graduation Year",
  "config": {
    "minYear": 1950,
    "maxYear": "${currentYear + 10}"
  }
}
```

---

## Selection Fields

### `select`
Dropdown selection (single or multiple).

**HTML:** `<select>` or custom dropdown  
**Database:** `VARCHAR`, `INTEGER`, or `TEXT[]` (multiple)

```json
{
  "name": "country",
  "type": "select",
  "label": "Country",
  "required": true,
  "config": {
    "searchable": true,
    "clearable": true,
    "multiple": false,
    "placeholder": "Select a country",
    "maxItems": null
  },
  "options": [
    {"value": "us", "label": "United States", "icon": "🇺🇸"},
    {"value": "uk", "label": "United Kingdom", "icon": "🇬🇧"},
    {"value": "ke", "label": "Kenya", "icon": "🇰🇪"}
  ],
  "optionsSource": null
}
```

**Config Options:**
- `searchable` (bool): Enable search/filter
- `clearable` (bool): Show clear button
- `multiple` (bool): Allow multiple selections
- `maxItems` (int): Max selections (for multiple)
- `optionsSource` (string): API endpoint for dynamic options
- `createable` (bool): Allow creating new options

**Options Format:**
```json
{
  "value": "unique_value",
  "label": "Display Text",
  "icon": "icon-name",
  "disabled": false,
  "group": "category"
}
```

**Dynamic Options:**
```json
{
  "name": "category_id",
  "type": "select",
  "label": "Category",
  "optionsSource": "/api/categories",
  "config": {
    "dependsOn": ["parent_id"],
    "valueField": "id",
    "labelField": "name"
  }
}
```

---

### `radio`
Radio button group for single selection.

**HTML:** `<input type="radio">`  
**Database:** `VARCHAR`, `INTEGER`, `BOOLEAN`

```json
{
  "name": "payment_method",
  "type": "radio",
  "label": "Payment Method",
  "required": true,
  "config": {
    "inline": true,
    "showDescription": true
  },
  "options": [
    {
      "value": "cash",
      "label": "Cash",
      "description": "Pay with cash on delivery"
    },
    {
      "value": "card",
      "label": "Credit Card",
      "description": "Secure card payment"
    },
    {
      "value": "mpesa",
      "label": "M-Pesa",
      "description": "Mobile money payment"
    }
  ]
}
```

**Config Options:**
- `inline` (bool): Display horizontally
- `showDescription` (bool): Show option descriptions

---

### `checkbox`
Single checkbox for boolean values.

**HTML:** `<input type="checkbox">`  
**Database:** `BOOLEAN`

```json
{
  "name": "agree_terms",
  "type": "checkbox",
  "label": "I agree to the terms and conditions",
  "required": true,
  "config": {
    "checkedValue": true,
    "uncheckedValue": false
  },
  "default": false
}
```

**Config Options:**
- `checkedValue`: Value when checked
- `uncheckedValue`: Value when unchecked

---

### `checkboxes`
Multiple checkbox group for multi-selection.

**HTML:** Multiple `<input type="checkbox">`  
**Database:** `TEXT[]` or `JSONB`

```json
{
  "name": "permissions",
  "type": "checkboxes",
  "label": "User Permissions",
  "config": {
    "inline": false,
    "columns": 2
  },
  "options": [
    {"value": "read", "label": "Read"},
    {"value": "write", "label": "Write"},
    {"value": "delete", "label": "Delete"},
    {"value": "admin", "label": "Admin"}
  ],
  "validation": {
    "minChecked": 1,
    "maxChecked": 3
  }
}
```

**Config Options:**
- `inline` (bool): Display horizontally
- `columns` (int): Grid columns
- `selectAll` (bool): Show "Select All" option

---

### `switch`
Toggle switch for boolean values.

**HTML:** `<input type="checkbox">` (styled)  
**Database:** `BOOLEAN`

```json
{
  "name": "is_active",
  "type": "switch",
  "label": "Active",
  "config": {
    "onLabel": "Active",
    "offLabel": "Inactive",
    "onValue": true,
    "offValue": false
  },
  "default": true
}
```

**Config Options:**
- `onLabel`, `offLabel` (string): Toggle labels
- `onValue`, `offValue`: Values for each state

---

### `tree-select`
Hierarchical tree-based selector.

**HTML:** Custom tree component  
**Database:** `INTEGER` or `VARCHAR` (for selected node ID)

```json
{
  "name": "department_id",
  "type": "tree-select",
  "label": "Department",
  "config": {
    "checkable": false,
    "multiple": false,
    "showSearch": true,
    "treeDefaultExpandAll": false
  },
  "optionsSource": "/api/departments/tree",
  "treeData": [
    {
      "value": "1",
      "label": "Engineering",
      "children": [
        {"value": "1-1", "label": "Backend"},
        {"value": "1-2", "label": "Frontend"}
      ]
    },
    {
      "value": "2",
      "label": "Sales"
    }
  ]
}
```

**Config Options:**
- `checkable` (bool): Show checkboxes
- `multiple` (bool): Allow multiple selections
- `showSearch` (bool): Enable tree search
- `treeDefaultExpandAll` (bool): Expand all nodes by default

---

## File & Media Fields

### `file`
Generic file upload.

**HTML:** `<input type="file">`  
**Database:** `VARCHAR` (file path) or `TEXT` (metadata JSON)

```json
{
  "name": "document",
  "type": "file",
  "label": "Upload Document",
  "config": {
    "multiple": false,
    "maxSize": 10485760,
    "accept": [".pdf", ".doc", ".docx"],
    "uploadUrl": "/api/upload",
    "downloadUrl": "/api/files/{id}",
    "autoUpload": true
  },
  "validation": {
    "maxSize": 10485760,
    "accept": ["application/pdf", "application/msword"]
  }
}
```

**Config Options:**
- `multiple` (bool): Allow multiple files
- `maxSize` (int): Max file size in bytes
- `accept` (array): Allowed file types/extensions
- `uploadUrl` (string): Upload endpoint
- `autoUpload` (bool): Upload immediately on select
- `drag` (bool): Enable drag-and-drop

**Validation:**
- `maxSize`: File size limit
- `accept`: MIME types or extensions
- `maxFiles`: Maximum number of files

---

### `image`
Image upload with preview.

**HTML:** `<input type="file" accept="image/*">`  
**Database:** `VARCHAR` (image path)

```json
{
  "name": "avatar",
  "type": "image",
  "label": "Profile Picture",
  "config": {
    "multiple": false,
    "maxSize": 5242880,
    "accept": ["image/jpeg", "image/png", "image/webp"],
    "crop": true,
    "aspectRatio": 1,
    "resize": {
      "maxWidth": 800,
      "maxHeight": 800,
      "quality": 0.9
    },
    "preview": true
  },
  "validation": {
    "maxSize": 5242880,
    "dimensions": {
      "minWidth": 100,
      "minHeight": 100,
      "maxWidth": 2000,
      "maxHeight": 2000
    }
  }
}
```

**Config Options:**
- `crop` (bool): Enable image cropping
- `aspectRatio` (number): Crop aspect ratio (1 = square)
- `resize` (object): Auto-resize settings
- `preview` (bool): Show image preview

---

## Text Content Fields

### `textarea`
Multi-line text input.

**HTML:** `<textarea>`  
**Database:** `TEXT`

```json
{
  "name": "description",
  "type": "textarea",
  "label": "Description",
  "config": {
    "rows": 4,
    "maxLength": 1000,
    "showCount": true,
    "autoSize": false,
    "resize": "vertical"
  },
  "validation": {
    "maxLength": 1000
  }
}
```

**Config Options:**
- `rows` (int): Initial row count
- `maxLength` (int): Character limit
- `showCount` (bool): Display character counter
- `autoSize` (bool): Auto-grow height
- `resize` (string): `"none"`, `"vertical"`, `"horizontal"`, `"both"`

---

### `richtext`
WYSIWYG rich text editor.

**HTML:** `<div contenteditable>` with editor  
**Database:** `TEXT` or `JSONB`

```json
{
  "name": "content",
  "type": "richtext",
  "label": "Content",
  "config": {
    "toolbar": [
      "bold", "italic", "underline", "strike",
      "heading", "blockquote", "code",
      "bulletList", "orderedList",
      "link", "image", "table"
    ],
    "height": 400,
    "maxLength": 50000,
    "allowImages": true,
    "allowTables": true,
    "sanitize": true,
    "outputFormat": "html"
  },
  "validation": {
    "maxLength": 50000
  }
}
```

**Config Options:**
- `toolbar` (array): Available formatting tools
- `height` (int): Editor height in pixels
- `allowImages` (bool): Enable image uploads
- `allowTables` (bool): Enable table insertion
- `sanitize` (bool): XSS protection
- `outputFormat` (string): `"html"`, `"markdown"`, `"json"`

---

### `code`
Code editor with syntax highlighting.

**HTML:** Monaco Editor or similar  
**Database:** `TEXT`

```json
{
  "name": "query",
  "type": "code",
  "label": "SQL Query",
  "config": {
    "language": "sql",
    "theme": "vs-dark",
    "height": 300,
    "lineNumbers": true,
    "minimap": false,
    "wordWrap": true,
    "readOnly": false
  }
}
```

**Config Options:**
- `language` (string): `"sql"`, `"javascript"`, `"python"`, `"json"`, etc.
- `theme` (string): Editor theme
- `height` (int): Editor height
- `lineNumbers` (bool): Show line numbers
- `minimap` (bool): Show code minimap
- `readOnly` (bool): Make editor read-only

---

### `json`
JSON editor with validation.

**HTML:** Code editor with JSON validation  
**Database:** `JSONB` or `TEXT`

```json
{
  "name": "metadata",
  "type": "json",
  "label": "Metadata",
  "config": {
    "height": 200,
    "validate": true,
    "format": true,
    "schema": {
      "type": "object",
      "properties": {
        "tags": {"type": "array"},
        "priority": {"type": "number"}
      }
    }
  }
}
```

**Config Options:**
- `validate` (bool): Real-time JSON validation
- `format` (bool): Auto-format JSON
- `schema` (object): JSON Schema for validation

---

## Specialized Fields

### `currency`
Monetary value input with formatting.

**HTML:** `<input type="number">` with formatting  
**Database:** `NUMERIC(19,4)`

```json
{
  "name": "amount",
  "type": "currency",
  "label": "Amount",
  "required": true,
  "config": {
    "currency": "USD",
    "locale": "en-US",
    "decimals": 2,
    "min": 0,
    "showSymbol": true,
    "symbolPosition": "before"
  },
  "validation": {
    "min": 0,
    "max": 999999999.99,
    "step": 0.01
  },
  "default": 0
}
```

**Config Options:**
- `currency` (string): ISO currency code
- `locale` (string): Locale for formatting
- `decimals` (int): Decimal places (typically 2)
- `showSymbol` (bool): Display currency symbol
- `symbolPosition` (string): `"before"` or `"after"`

**Supported Currencies:**
```json
{
  "currencies": [
    {"code": "USD", "symbol": "$"},
    {"code": "EUR", "symbol": "€"},
    {"code": "KES", "symbol": "KSh"},
    {"code": "GBP", "symbol": "£"}
  ]
}
```

---

### `tags`
Tag input for comma-separated values.

**HTML:** Custom tag component  
**Database:** `TEXT[]` or `TEXT` (comma-separated)

```json
{
  "name": "skills",
  "type": "tags",
  "label": "Skills",
  "config": {
    "delimiter": ",",
    "maxTags": 10,
    "allowCustom": true,
    "suggestions": ["JavaScript", "Python", "Go", "SQL"],
    "suggestionsSource": "/api/skills/suggestions"
  },
  "validation": {
    "minTags": 1,
    "maxTags": 10,
    "tagPattern": "^[a-zA-Z0-9\\s-]+$"
  }
}
```

**Config Options:**
- `delimiter` (string): Tag separator
- `maxTags` (int): Maximum number of tags
- `allowCustom` (bool): Allow creating new tags
- `suggestions` (array): Pre-defined tag suggestions
- `suggestionsSource` (string): API for tag suggestions

---

### `rating`
Star rating or numeric scale.

**HTML:** Custom rating component  
**Database:** `SMALLINT` or `NUMERIC(3,2)`

```json
{
  "name": "product_rating",
  "type": "rating",
  "label": "Rating",
  "config": {
    "max": 5,
    "allowHalf": true,
    "allowClear": true,
    "character": "★",
    "showTooltips": true,
    "tooltips": ["Poor", "Fair", "Good", "Very Good", "Excellent"]
  },
  "validation": {
    "min": 1,
    "max": 5
  }
}
```

**Config Options:**
- `max` (int): Maximum rating value
- `allowHalf` (bool): Allow half-star ratings
- `allowClear` (bool): Allow clearing rating
- `character` (string): Rating character/icon
- `tooltips` (array): Text for each rating level

---

### `slider`
Numeric slider input.

**HTML:** `<input type="range">`  
**Database:** `INTEGER` or `NUMERIC`

```json
{
  "name": "volume",
  "type": "slider",
  "label": "Volume",
  "config": {
    "min": 0,
    "max": 100,
    "step": 1,
    "marks": {
      "0": "Min",
      "50": "Medium",
      "100": "Max"
    },
    "showValue": true,
    "range": false
  },
  "default": 50
}
```

**Config Options:**
- `min`, `max` (number): Value range
- `step` (number): Increment value
- `marks` (object): Labels at specific values
- `showValue` (bool): Display current value
- `range` (bool): Enable range selection (two handles)

---

### `icon-picker`
Icon selector from icon library.

**HTML:** Custom icon picker  
**Database:** `VARCHAR(50)`

```json
{
  "name": "menu_icon",
  "type": "icon-picker",
  "label": "Menu Icon",
  "config": {
    "library": "heroicons",
    "searchable": true,
    "categories": true,
    "preview": true
  }
}
```

**Config Options:**
- `library` (string): Icon library name
- `searchable` (bool): Enable icon search
- `categories` (bool): Show icon categories

---

## Relational Fields

### `relation`
Foreign key relationship selector.

**HTML:** Searchable select  
**Database:** `INTEGER` or `UUID` (foreign key)

```json
{
  "name": "customer_id",
  "type": "relation",
  "label": "Customer",
  "required": true,
  "config": {
    "targetSchema": "customers",
    "displayField": "name",
    "searchFields": ["name", "email", "phone"],
    "valueField": "id",
    "searchable": true,
    "createNew": true,
    "createUrl": "/customers/new",
    "sortBy": "name",
    "filters": {
      "status": "active"
    }
  },
  "validation": {
    "exists": true
  }
}
```

**Config Options:**
- `targetSchema` (string): Related table/schema
- `displayField` (string): Field to display in dropdown
- `searchFields` (array): Fields to search
- `valueField` (string): Field to use as value (typically `id`)
- `createNew` (bool): Allow creating new records inline
- `filters` (object): Additional filters for related records

---

### `autocomplete`
Autocomplete search input.

**HTML:** Input with dropdown suggestions  
**Database:** Depends on relation type

```json
{
  "name": "product_id",
  "type": "autocomplete",
  "label": "Product",
  "config": {
    "source": "/api/products/search",
    "minChars": 2,
    "debounce": 300,
    "maxResults": 10,
    "displayField": "name",
    "valueField": "id",
    "template": "${code} - ${name}"
  }
}
```

**Config Options:**
- `source` (string): API endpoint for search
- `minChars` (int): Minimum characters before search
- `debounce` (int): Delay in milliseconds
- `maxResults` (int): Maximum suggestions to show
- `template` (string): Display template with placeholders

---

## Layout & Container Fields

### `repeatable`
Repeatable field group for one-to-many relationships.

**HTML:** Dynamic list with add/remove buttons  
**Database:** Related table or `JSONB` array

```json
{
  "name": "line_items",
  "type": "repeatable",
  "label": "Invoice Items",
  "config": {
    "minItems": 1,
    "maxItems": 100,
    "addButtonText": "Add Item",
    "removeButtonText": "Remove",
    "sortable": true,
    "collapsible": false,
    "itemTitle": "Item ${index + 1}"
  },
  "fields": [
    {
      "name": "product_id",
      "type": "relation",
      "label": "Product",
      "config": {
        "targetSchema": "products"
      }
    },
    {
      "name": "quantity",
      "type": "number",
      "label": "Qty",
      "default": 1
    },
    {
      "name": "price",
      "type": "currency",
      "label": "Price"
    },
    {
      "name": "total",
      "type": "currency",
      "label": "Total",
      "config": {
        "readOnly": true,
        "formula": "${quantity} * ${price}"
      }
    }
  ],
  "validation": {
    "minItems": 1,
    "maxItems": 100
  }
}
```

**Config Options:**
- `minItems`, `maxItems` (int): Item count constraints
- `addButtonText`, `removeButtonText` (string): Button labels
- `sortable` (bool): Allow drag-to-reorder
- `collapsible` (bool): Allow collapsing items
- `itemTitle` (string): Title template for each item

---

### `group`
Logical grouping of fields.

**HTML:** `<div>` container  
**Database:** N/A (layout only)

```json
{
  "name": "address_group",
  "type": "group",
  "label": "Address Information",
  "config": {
    "collapsible": true,
    "collapsed": false,
    "border": true
  },
  "fields": [
    {"name": "street", "type": "text", "label": "Street"},
    {"name": "city", "type": "text", "label": "City"},
    {"name": "postal_code", "type": "text", "label": "Postal Code"}
  ]
}
```

**Config Options:**
- `collapsible` (bool): Can collapse/expand
- `collapsed` (bool): Initial collapsed state
- `border` (bool): Show border around group

---

### `tabs`
Tabbed interface for organizing fields.

**HTML:** Tab component  
**Database:** N/A (layout only)

```json
{
  "name": "product_tabs",
  "type": "tabs",
  "config": {
    "position": "top",
    "animated": true
  },
  "tabs": [
    {
      "key": "general",
      "label": "General",
      "icon": "info",
      "fields": [
        {"name": "name", "type": "text", "label": "Product Name"},
        {"name": "sku", "type": "text", "label": "SKU"}
      ]
    },
    {
      "key": "pricing",
      "label": "Pricing",
      "icon": "currency",
      "fields": [
        {"name": "price", "type": "currency", "label": "Price"},
        {"name": "cost", "type": "currency", "label": "Cost"}
      ]
    }
  ]
}
```

**Config Options:**
- `position` (string): `"top"`, `"left"`, `"right"`, `"bottom"`
- `animated` (bool): Animate tab transitions

---

### `fieldset`
Fieldset with legend (HTML fieldset).

**HTML:** `<fieldset>`  
**Database:** N/A (layout only)

```json
{
  "name": "contact_fieldset",
  "type": "fieldset",
  "label": "Contact Information",
  "fields": [
    {"name": "email", "type": "email", "label": "Email"},
    {"name": "phone", "type": "phone", "label": "Phone"}
  ]
}
```

---

## Display-Only Fields

### `static`
Static text or HTML display (non-editable).

**HTML:** `<div>` with content  
**Database:** N/A

```json
{
  "name": "help_text",
  "type": "static",
  "content": "<p>Please fill in all required fields marked with *</p>",
  "config": {
    "className": "help-text",
    "allowHtml": true
  }
}
```

**Config Options:**
- `content` (string): Static content to display
- `allowHtml` (bool): Allow HTML in content
- `className` (string): CSS class

---

### `divider`
Visual separator line.

**HTML:** `<hr>` or styled `<div>`  
**Database:** N/A

```json
{
  "name": "section_divider",
  "type": "divider",
  "config": {
    "text": "Additional Information",
    "textPosition": "center"
  }
}
```

**Config Options:**
- `text` (string): Optional text label
- `textPosition` (string): `"left"`, `"center"`, `"right"`

---

### `formula`
Calculated/computed field (read-only).

**HTML:** Display field  
**Database:** Can be computed or stored

```json
{
  "name": "line_total",
  "type": "formula",
  "label": "Total",
  "config": {
    "formula": "${quantity} * ${unit_price}",
    "format": "currency",
    "precision": 2,
    "recalculate": "onChange"
  }
}
```

**Config Options:**
- `formula` (string): Calculation expression
- `format` (string): Output format
- `recalculate` (string): When to recalculate (`"onChange"`, `"onSubmit"`)

**Formula Syntax:**
- Variables: `${fieldName}`
- Operators: `+`, `-`, `*`, `/`, `%`
- Functions: `sum()`, `avg()`, `min()`, `max()`, `round()`

---

## Database Type Mapping

Map field types to PostgreSQL column types:

| Field Type | PostgreSQL Type | Example |
|------------|----------------|---------|
| `text` | `VARCHAR(n)` or `TEXT` | `VARCHAR(255)` |
| `email` | `VARCHAR(255)` | `VARCHAR(255)` |
| `password` | `VARCHAR(255)` | `VARCHAR(255)` (hashed) |
| `number` | `INTEGER`, `BIGINT`, `NUMERIC` | `INTEGER` |
| `currency` | `NUMERIC(19,4)` | `NUMERIC(19,4)` |
| `phone` | `VARCHAR(50)` | `VARCHAR(50)` |
| `url` | `VARCHAR(2048)` | `VARCHAR(2048)` |
| `date` | `DATE` | `DATE` |
| `time` | `TIME` | `TIME` |
| `datetime` | `TIMESTAMP` or `TIMESTAMPTZ` | `TIMESTAMPTZ` |
| `checkbox` | `BOOLEAN` | `BOOLEAN` |
| `switch` | `BOOLEAN` | `BOOLEAN` |
| `select` (single) | `VARCHAR` or `INTEGER` | `INTEGER` (FK) |
| `select` (multiple) | `TEXT[]` or `JSONB` | `TEXT[]` |
| `tags` | `TEXT[]` | `TEXT[]` |
| `textarea` | `TEXT` | `TEXT` |
| `richtext` | `TEXT` or `JSONB` | `TEXT` |
| `json` | `JSONB` | `JSONB` |
| `file` | `VARCHAR(255)` (path) | `VARCHAR(255)` |
| `image` | `VARCHAR(255)` (path) | `VARCHAR(255)` |
| `relation` | `INTEGER` or `UUID` (FK) | `INTEGER` |
| `rating` | `SMALLINT` or `NUMERIC(3,2)` | `NUMERIC(3,2)` |
| `color` | `VARCHAR(20)` | `VARCHAR(20)` |

**Recommended Indexes:**
- Foreign keys (`relation` fields): `CREATE INDEX idx_table_fk ON table(fk_column)`
- Search fields (`text`, `email`): `CREATE INDEX idx_table_search ON table USING gin(to_tsvector('english', column))`
- Date ranges: `CREATE INDEX idx_table_date ON table(date_column)`
- JSON fields: `CREATE INDEX idx_table_json ON table USING gin(json_column)`

---

## Validation Rules by Type

| Field Type | Supported Validations |
|------------|----------------------|
| **text** | `required`, `minLength`, `maxLength`, `pattern`, `custom` |
| **email** | `required`, `format`, `pattern`, `custom` |
| **password** | `required`, `minLength`, `pattern`, `custom` |
| **number** | `required`, `min`, `max`, `step`, `custom` |
| **currency** | `required`, `min`, `max`, `step`, `custom` |
| **phone** | `required`, `pattern`, `custom` |
| **url** | `required`, `format`, `pattern`, `custom` |
| **date** | `required`, `min`, `max`, `custom` |
| **datetime** | `required`, `min`, `max`, `custom` |
| **textarea** | `required`, `minLength`, `maxLength`, `pattern`, `custom` |
| **richtext** | `required`, `minLength`, `maxLength`, `custom` |
| **select** | `required`, `custom` |
| **checkbox** | `required`, `custom` |
| **checkboxes** | `required`, `minChecked`, `maxChecked`, `custom` |
| **file** | `required`, `maxSize`, `accept`, `custom` |
| **image** | `required`, `maxSize`, `accept`, `dimensions`, `custom` |
| **tags** | `required`, `minTags`, `maxTags`, `tagPattern`, `custom` |
| **rating** | `required`, `min`, `max`, `custom` |
| **relation** | `required`, `exists`, `custom` |
| **repeatable** | `minItems`, `maxItems`, field-specific validations |

**Common Validation Options:**

```json
{
  "validation": {
    "required": true,
    "custom": "${value.length >= 3}",
    "message": "Must be at least 3 characters"
  }
}
```

---

## Conditional Fields

Fields can be shown/hidden based on other field values:

```json
{
  "name": "discount_type",
  "type": "select",
  "options": [
    {"value": "percentage", "label": "Percentage"},
    {"value": "fixed", "label": "Fixed Amount"}
  ]
},
{
  "name": "discount_percentage",
  "type": "number",
  "label": "Discount %",
  "showIf": {
    "field": "discount_type",
    "equals": "percentage"
  },
  "validation": {
    "min": 0,
    "max": 100
  }
},
{
  "name": "discount_amount",
  "type": "currency",
  "label": "Discount Amount",
  "showIf": {
    "field": "discount_type",
    "equals": "fixed"
  }
}
```

**Conditional Operators:**
- `equals`: Exact match
- `notEquals`: Not equal
- `in`: Value in array
- `notIn`: Value not in array
- `gt`, `gte`: Greater than (or equal)
- `lt`, `lte`: Less than (or equal)
- `contains`: String contains
- `isEmpty`: Field is empty
- `isNotEmpty`: Field has value

**Complex Conditions:**
```json
{
  "showIf": {
    "all": [
      {"field": "status", "equals": "active"},
      {"field": "type", "in": ["premium", "enterprise"]}
    ]
  }
}
```

---

## Default Values

All fields support default values:

**Static Defaults:**
```json
{
  "name": "status",
  "type": "select",
  "default": "draft"
}
```

**Dynamic Defaults:**
```json
{
  "name": "created_at",
  "type": "datetime",
  "default": "${now}"
}

{
  "name": "created_by",
  "type": "hidden",
  "default": "${session.user_id}"
}
```

**Available Variables:**
- `${now}`: Current timestamp
- `${today}`: Today's date
- `${session.user_id}`: Current user ID
- `${session.tenant_id}`: Current tenant ID
- `${uuid}`: Generate UUID

---

## Best Practices

1. **Field Naming**: Use `snake_case` for field names to match database columns
2. **Labels**: Use clear, concise labels with proper capitalization
3. **Placeholders**: Provide helpful placeholder text with examples
4. **Validation**: Always validate on both client and server side
5. **Required Fields**: Mark required fields clearly in the UI
6. **Help Text**: Add descriptions for complex fields
7. **Defaults**: Set sensible defaults where appropriate
8. **Performance**: Use `optionsSource` for large option lists (>100 items)
9. **Accessibility**: Include proper labels and ARIA attributes
10. **Security**: Sanitize rich text and validate file uploads

---

[← Back to Schema Structure](04-schema-structure.md) | [Next: Validation →](06-validation.md)
