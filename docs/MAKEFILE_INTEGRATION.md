# 📋 Makefile Documentation Commands

The Makefile has been updated with comprehensive documentation commands that integrate with the new documentation server implementation.

## 🚀 Available Commands

### Primary Commands

| Command | Description | Usage |
|---------|-------------|--------|
| `make docs` | Start documentation server on port 8081 | `make docs` |
| `make docs-port` | Start server on custom port | `make docs-port DOC_PORT=9000` |
| `make docs-build` | Build MkDocs documentation | `make docs-build` |

### Development Commands

| Command | Description | Usage |
|---------|-------------|--------|
| `make docs-dev` | Build docs + start server | `make docs-dev` |
| `make docs-test` | Test documentation server | `make docs-test` |
| `make docs-mkdocs-safe` | Test MkDocs safety features | `make docs-mkdocs-safe` |

### Utility Commands

| Command | Description | Usage |
|---------|-------------|--------|
| `make docs-schema` | Show SchemaSpy command template | `make docs-schema` |

## 🎯 Quick Start Examples

### Basic Usage
```bash
# Start documentation server on default port (8081)
make docs

# Start on custom port
make docs-port DOC_PORT=3000
DOC_PORT=3000 make docs-port  # Alternative syntax
```

### Development Workflow
```bash
# Build MkDocs and start server
make docs-dev

# Just build MkDocs documentation
make docs-build

# Test the server functionality
make docs-test
```

## 🏗️ What Each Command Does

### `make docs`
- 🏢 Starts AWO ERP Documentation Server
- 📚 Serves MkDocs documentation at `/`
- 🗄️ Serves Schema documentation at `/schema/`
- 🌐 Default port: 8081
- 🔗 Uses `docs/start-docs.sh`

**Output:**
```
🏢 Starting AWO ERP Documentation Server...
📚 MkDocs documentation: http://localhost:8081/
🗄️ Schema documentation: http://localhost:8081/schema/
```

### `make docs-build`
- 📖 Runs `mkdocs build`
- ✅ Generates static HTML in `site/` directory
- 🛡️ Safe: Won't overwrite `docs/index.html`

### `make docs-test`
- 🧪 Runs comprehensive server tests
- ✅ Tests all endpoints (MkDocs, Schema, Assets)
- 📊 Reports HTTP status codes
- 🚀 Uses port 8082 for testing

### `make docs-dev`
- 👨‍💻 Development workflow automation
- 1️⃣ First: Builds MkDocs documentation
- 2️⃣ Then: Starts documentation server
- 🔄 Perfect for iterative development

### `make docs-mkdocs-safe`
- 🛡️ Tests MkDocs integration safety
- ✅ Verifies `docs/index.html` protection
- 🔍 Confirms server serves custom landing page
- 📝 Shows that `mkdocs build` won't break custom navigation

## 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DOC_PORT` | `8081` | Port for documentation server |
| `DB_NAME` | `ledger` | Database name for schema generation |
| `DB_USER` | `admin` | Database user for schema generation |
| `DB_PSSWD` | `admin` | Database password for schema generation |

## 🔗 Integration Points

### With MkDocs
- ✅ `make docs-build` → `mkdocs build`
- ✅ Safe integration (no overwriting custom files)
- ✅ Serves generated content from `site/`

### With SchemaSpy
- 📝 `make docs-schema` shows generation command
- 🎯 Outputs to `docs/schema/`
- 🔄 Manual generation (requires database connection)

### With Go Server
- 🚀 Uses `docs/start-docs.sh` for production
- 🧪 Uses `docs/test-server.sh` for testing
- ⚙️ Direct Go execution with proper error handling

## 🧪 Testing Commands

```bash
# Test server functionality
make docs-test

# Test MkDocs safety
make docs-mkdocs-safe

# Manual verification
make docs &        # Start server in background
curl http://localhost:8081/                    # Test root
curl http://localhost:8081/schema/             # Test schema
curl http://localhost:8081/getting-started/   # Test MkDocs
```

## 📝 Notes

- **Port Conflicts**: Commands use different ports for testing
- **Background Jobs**: Use `Ctrl+C` to stop servers
- **Prerequisites**: Ensure `docs/start-docs.sh` is executable
- **Database**: Schema generation requires running PostgreSQL
- **MkDocs**: Requires `mkdocs` and `mkdocs-material` installed

## 🚨 Troubleshooting

### "Permission denied" errors
```bash
chmod +x docs/start-docs.sh
chmod +x docs/test-server.sh
```

### "Port already in use"
```bash
make docs-port DOC_PORT=8082
```

### "Command not found: mkdocs"
```bash
pip install mkdocs mkdocs-material
```

### Schema documentation missing
```bash
# Generate with SchemaSpy (requires database running)
java -jar schemaspy.jar -t pgsql -host localhost -port 5432 \
  -db ledger -u admin -p admin -o docs/schema
```