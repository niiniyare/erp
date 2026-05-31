# Development Tools

## Overview

This directory contains tools, scripts, and utilities for developing, validating, and maintaining the ERP UI documentation and component system.

## Available Tools

### Schema Tools
- **[schema-tools/](./schema-tools/)** - Schema validation and generation tools
  - `check-doc-accuracy.sh` - Validates documentation accuracy against implementation
  - `generate-metadata.sh` - Generates schema metadata and documentation
  - `organize-schemas.sh` - Organizes and validates schema structure

### Project Tools
- **`restructure-progress.sh`** - Tracks documentation restructuring progress

## Tool Usage

### Schema Validation
```bash
# Validate all documentation accuracy
./tools/schema-tools/check-doc-accuracy.sh

# Generate schema metadata
./tools/schema-tools/generate-metadata.sh

# Organize schema files
./tools/schema-tools/organize-schemas.sh
```

### Development Workflow
```bash
# Check restructuring progress
./tools/restructure-progress.sh
```

## Adding New Tools

When creating new development tools:
1. Place scripts in appropriate subdirectory
2. Include clear usage documentation
3. Add error handling and validation
4. Update this README with tool description

## Related Documentation

- **[Development Guides](../development/)** - Development workflows and patterns
- **[Schema System](../schemas/)** - Schema validation and organization