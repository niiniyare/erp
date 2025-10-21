# Documentation Validation Process

## 🎯 Purpose

Ensure all schema documentation accurately reflects the actual project implementation with automated validation checks.

## 📋 Pre-Publication Validation Checklist

### **1. File Path Verification**
```bash
# Verify every file path mentioned in documentation exists
find /data/data/com.termux/files/home/project/erp -name "*.go" | grep -E "(schema|component|css)" | head -10
find docs/ui/Schema/definitions -name "*.json" | wc -l  # Get exact count
```

### **2. Schema Count Accuracy**
```bash
# Always use actual counts, never estimates
TOTAL_SCHEMAS=$(find docs/ui/Schema/definitions -name "*.json" | wc -l)
echo "Actual schema count: $TOTAL_SCHEMAS"
```

### **3. Component File Verification**
```bash
# Check web/components structure before documenting
ls -la web/components/
ls -la web/engine/
```

### **4. Technical Claims Validation**
- [ ] Every Go struct mentioned exists in the codebase
- [ ] Every interface referenced has actual implementation
- [ ] Every method call shown is valid
- [ ] Every import path is correct

### **5. Schema File References**
- [ ] Every schema file mentioned by name exists
- [ ] Schema dependencies are real and accurate
- [ ] Example schema props match actual schema definitions

## 🔧 Validation Tools

### **1. File Existence Checker**
```bash
#!/bin/bash
# check_doc_accuracy.sh

echo "🔍 Validating documentation accuracy..."

# Check critical files mentioned in docs
CRITICAL_FILES=(
    "web/engine/schema_factory.go"
    "web/engine/schema_templ.go" 
    "web/engine/enhanced_registry.go"
    "web/engine/css_integration.go"
    "pkg/schema/ui/css/properties.go"
)

for file in "${CRITICAL_FILES[@]}"; do
    if [[ -f "$file" ]]; then
        echo "✅ $file exists"
    else
        echo "❌ $file MISSING - update docs"
    fi
done

# Validate schema counts
ACTUAL_COUNT=$(find docs/ui/Schema/definitions -name "*.json" | wc -l)
echo "📊 Actual schema count: $ACTUAL_COUNT"
```

### **2. Schema Reference Validator**
```go
// validate_schema_refs.go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

func main() {
    // Read documentation files
    docFiles, _ := filepath.Glob("docs/ui/**/*.md")
    
    // Extract schema references
    schemaPattern := regexp.MustCompile(`[A-Z][a-zA-Z]*Schema\.json`)
    
    for _, docFile := range docFiles {
        content, _ := os.ReadFile(docFile)
        matches := schemaPattern.FindAllString(string(content), -1)
        
        for _, match := range matches {
            schemaPath := filepath.Join("docs/ui/Schema/definitions", match)
            if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
                fmt.Printf("❌ %s references missing schema: %s\n", docFile, match)
            }
        }
    }
}
```

## 📝 Documentation Standards

### **1. Never Make Assumptions**
- ❌ "The button schema exists..."
- ✅ "After verifying ButtonSchema.json exists at [path]..."

### **2. Always Verify Before Writing**
```bash
# Before documenting any feature:
ls -la web/components/atoms/  # Check actual files
grep -r "ButtonProps" web/   # Verify type exists
```

### **3. Use Actual Examples**
- ❌ Generic examples that might not work
- ✅ Copy-paste examples from actual working code

### **4. Reference Real File Paths**
```markdown
❌ "The schema factory (located in the engine)"
✅ "The schema factory in `web/engine/schema_factory.go:89`"
```

## 🚀 Implementation Protocol

### **Before Any Documentation Update**

1. **Run Validation Script**
   ```bash
   ./check_doc_accuracy.sh
   ```

2. **Verify Schema References**
   ```bash
   go run validate_schema_refs.go
   ```

3. **Check Implementation**
   ```bash
   # Verify the actual Go code
   cd web/engine && ls -la
   grep -n "SchemaFactory" *.go
   ```

4. **Update Counts**
   ```bash
   # Get real numbers
   find docs/ui/Schema/definitions -name "*.json" | wc -l
   ```

### **Documentation Review Process**

1. **Self-Validation**: Author runs all validation scripts
2. **Peer Review**: Another developer verifies claims
3. **Code Cross-Reference**: Every technical claim verified against actual code
4. **Integration Test**: Documentation examples must actually work

## 🎯 Error Prevention

### **Common Accuracy Issues to Avoid**

1. **Schema File References**: Always verify existence before mentioning
2. **Component Counts**: Use actual counts, not estimates
3. **File Paths**: Copy-paste actual paths, don't type from memory
4. **Technical Claims**: Every interface/struct mentioned must exist
5. **Example Code**: All examples must be tested and work

### **Red Flags in Documentation**

- "Approximately X files..." (get exact count)
- "The schema probably..." (verify or don't document)
- "Similar to..." (show actual implementation)
- "Should be located at..." (verify location)

## 📊 Success Metrics

- **100% file path accuracy**: Every referenced file exists
- **100% technical claim validity**: Every interface/method exists
- **Real example success**: All code examples work when copy-pasted
- **Zero broken references**: No dead links or missing schemas

This validation process ensures documentation serves as a reliable source of truth rather than a source of confusion.