# Your First Component

**FILE PURPOSE**: Create a working schema-driven component in under 10 minutes  
**PREREQUISITES**: [Installation](installation.md) completed  
**OUTCOME**: Functional component with schema validation

## 🎯 Goal: Create a Button Component

We'll create a simple button using our schema-driven architecture.

### Step 1: Find the Button Schema
```bash
# Locate button schemas
find docs/ui/Schema/definitions -name "*button*" -type f

# Expected output:
# docs/ui/Schema/definitions/components/atoms/ButtonGroupSchema.json
# docs/ui/Schema/definitions/components/atoms/ButtonToolbarSchema.json
```

### Step 2: Examine the Schema Structure
```bash
# View a component schema
cat docs/ui/Schema/definitions/components/atoms/ButtonGroupSchema.json | head -20
```

The schema defines properties like `text`, `variant`, `size`, `disabled`, etc.

### Step 3: Create Component Using Schema Factory
```go
// Create test file: cmd/test-button/main.go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/niiniyare/erp/web/engine"
)

func main() {
    // Initialize schema factory
    factory, err := engine.NewSchemaFactory("docs/ui/Schema")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create button props
    props := map[string]interface{}{
        "text":     "Click Me",
        "variant":  "primary",
        "size":     "md",
        "disabled": false,
    }
    
    // Generate component from schema
    component, err := factory.RenderToTempl(context.Background(), "ButtonGroupSchema", props)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ Button component created successfully!")
    fmt.Printf("Component type: %T\n", component)
}
```

### Step 4: Test Your Component
```bash
# Run the test
go run cmd/test-button/main.go

# Expected output:
# ✅ Button component created successfully!
# Component type: templ.Component
```

### Step 5: Add to Web Page
```go
// In your page template
package pages

import "github.com/niiniyare/erp/web/engine"

templ TestPage() {
    <!DOCTYPE html>
    <html>
    <head>
        <title>Test Page</title>
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        <script src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
        <link href="https://cdn.jsdelivr.net/npm/flowbite@2.0.0/dist/flowbite.min.css" rel="stylesheet">
    </head>
    <body>
        <div class="p-8">
            <h1 class="text-2xl font-bold mb-4">My First Schema Component</h1>
            
            // Schema-driven button will render here
            @renderSchemaButton()
        </div>
    </body>
    </html>
}

templ renderSchemaButton() {
    // This would be generated from your schema factory
    <button type="button" class="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5">
        Click Me
    </button>
}
```

## 🎉 Success!

You've successfully:
- ✅ Located a component schema
- ✅ Used the schema factory to generate a component
- ✅ Created a working button with proper styling

## 🚀 Next Steps

**Option A: Explore More Components**
```bash
# See all available component schemas
find docs/ui/Schema/definitions/components -name "*.json" | head -10
```

**Option B: Customize Your Component**
- Modify the `props` object to change appearance
- Try different `variant` values: `"secondary"`, `"success"`, `"danger"`
- Experiment with `size`: `"sm"`, `"md"`, `"lg"`, `"xl"`

**Option C: Add Interactivity**
- Add HTMX attributes for server interactions
- Include Alpine.js directives for client-side behavior

## 📚 Continue Learning

✅ **First Component Complete** → Continue to [Basic Examples](basic-examples.md)

**Quick Start Path**:
1. ✅ [Installation](installation.md)
2. ✅ First Component (you are here)
3. 🎯 [Basic Examples](basic-examples.md)

**Deep Dive**: [Schema System](../fundamentals/schema-system.md) | [Component Development](../development/creating-components.md)