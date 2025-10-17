# Installation Guide

**FILE PURPOSE**: Environment setup for ERP UI development  
**ESTIMATED TIME**: 15 minutes  
**OUTCOME**: Working development environment with schema-driven components

## 🎯 Quick Setup

### Prerequisites
```bash
# Verify Go version (1.21+ required)
go version

# Install Templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Verify installation
templ --help
```

### Project Setup
```bash
# Navigate to ERP project
cd /path/to/erp

# Install dependencies
go mod tidy

# Generate Templ components
make templ

# Generate CSS properties
make sqlc
```

### Verify Installation
```bash
# Check web components exist
ls web/components/atoms/

# Verify schema system
find docs/ui/Schema/definitions -name "*.json" | wc -l
# Expected: 937+ schemas

# Test component generation
go run cmd/test-schema/main.go
```

## 🔧 Development Environment

### Required Tools
- **Go 1.21+**: Core language
- **Templ CLI**: Template compilation
- **Make**: Build automation

### Optional Tools
- **Air**: Hot reload during development
- **Node.js**: For npm-based tooling

### IDE Setup
```bash
# VS Code extensions (recommended)
code --install-extension bradlc.vscode-tailwindcss
code --install-extension golang.go
```

## 🚀 Next Steps

✅ **Environment Ready** → Continue to [First Component](first-component.md)

**Quick Start Path**:
1. ✅ Installation (you are here)
2. 🎯 [Create First Component](first-component.md)
3. 📚 [Basic Examples](basic-examples.md)

**Troubleshooting**: See [Development Guide](../development/creating-components.md) for detailed setup issues.