# Documentation Guide: Adding New Documentation

## Overview

This guide walks through the process of adding new documentation to the AWO ERP MkDocs system. The documentation follows a structured approach with clear organization, search optimization, and mobile-friendly design.

## Prerequisites

- Basic understanding of Markdown
- Knowledge of MkDocs structure
- Access to the project repository
- Text editor or IDE

## Project Structure

```
docs/
├── getting-started/          # Quick start guides
├── architecture/             # System design documentation
├── contributing/             # Developer guides (this file)
├── reference/
│   ├── api/                  # API documentation
│   │   ├── auth/            # Authentication API
│   │   ├── users/           # Users API
│   │   ├── organizations/   # Organizations API
│   │   ├── abac/           # ABAC Authorization API
│   │   └── tutorials/      # API tutorials
│   └── modules/             # Module-specific docs
├── javascripts/             #  search & help system
└── mkdocs.yml              # Main configuration file
```

## Step-by-Step Guide

### Step 1: Determine Documentation Category

Choose the appropriate category for your new documentation:

1. **Getting Started** - Tutorials, quick starts, installation guides
2. **Architecture** - System design, technical specifications
3. **Contributing** - Developer guides, best practices, workflows
4. **API Reference** - API endpoints, examples, integration guides
5. **Module Reference** - Feature-specific documentation

### Step 2: Create the Documentation File

#### 2.1 Choose the Right Location

Create your file in the appropriate directory:

```bash
# For getting started guides
touch docs/getting-started/my-new-guide.md

# For API documentation
mkdir -p docs/reference/api/my-new-api/
touch docs/reference/api/my-new-api/index.md

# For module documentation  
mkdir -p docs/reference/modules/my-module/
touch docs/reference/modules/my-module/overview.md

# For contributing guides
touch docs/contributing/my-guide.md
```

#### 2.2 Use the Standard Template

Start with this template structure:

```markdown
# Page Title

## Overview

Brief description of what this documentation covers and who it's for.

## Prerequisites

- Requirement 1
- Requirement 2
- Requirement 3

## Key Concepts/Features

- **Feature 1**: Description
- **Feature 2**: Description
- **Feature 3**: Description

## Quick Start

### Basic Example
```bash
# Code example with clear comments
command --option value
```

## Detailed Sections

### Section 1: Implementation
Detailed explanation with code examples.

### Section 2: Configuration
Configuration options and examples.

### Section 3: Best Practices
Recommended approaches and common patterns.

## Error Handling

### Common Issues
- **Issue 1**: Description and solution
- **Issue 2**: Description and solution

## Advanced Topics

### Advanced Feature 1
In-depth explanation for advanced users.

## Related Resources

- Link to related internal guides
- Link to external resources

---

**Next**: Link to the next logical topic | **Up**: Link to the parent topic
```

### Step 3: Add Navigation to mkdocs.yml

Edit the `mkdocs.yml` file to include your new documentation:

```yaml
nav:
  - 'Getting Started': 'getting-started/01-developer-quick-start.md'
  - 'Your New Section':
    - 'Overview': 'your-section/index.md'
    - 'Sub-topic 1': 'your-section/sub-topic-1.md'
    - 'Sub-topic 2': 'your-section/sub-topic-2.md'
```

**Navigation Best Practices:**
- Use clear, descriptive titles
- Group related topics logically
- Limit nesting to 3-4 levels maximum
- Use emoji sparingly (🚀 for interactive elements only)

### Step 4: Follow Content Guidelines

#### 4.1 Writing Style
- **Be Clear and Concise**: Use simple, direct language
- **Use Active Voice**: "Create the file" not "The file should be created"
- **Include Examples**: Provide practical, working examples
- **Add Context**: Explain why, not just how

#### 4.2 Code Examples
```bash
# Good: Clear, commented example
# Create a new user with proper role assignment
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new.user@company.com",
    "roles": ["user", "developer"]
  }'
```

#### 4.3 Link Structure
```markdown
# Internal links (relative)
[API Guide](../api/index.md)
[Configuration](../../modules/config.md)

# External links (absolute)
[Goa Framework](https://goa.design/)

# API endpoint references
Check the [Users API](../api/users/) for details.
```

### Step 5: Add Search Optimization

#### 5.1 Include Keywords
Make your documentation discoverable by including relevant keywords:

- In headings: Use terms users will search for
- In first paragraph: Include main concepts
- In code comments: Add searchable terms

#### 5.2 Add Metadata (Optional)
For complex guides, add metadata at the top:

```markdown
---
title: "Custom Page Title"
description: "SEO-friendly description"
tags: ["api", "authentication", "jwt"]
---

# Your Documentation Title
```

### Step 6: Test Your Documentation

#### 6.1 Local Testing
```bash
# Build and serve locally
mkdocs serve

# Check for broken links and warnings
mkdocs build --strict
```

#### 6.2 Validation Checklist
- [ ] All links work correctly
- [ ] Code examples are tested
- [ ] Images load properly (if any)
- [ ] Navigation flows logically
- [ ] Mobile-friendly formatting

### Step 7: API Documentation (Special Case)

For API documentation, follow this structure:

#### 7.1 API Guide Template
```markdown
# API Name

## Overview
Brief description and key features.

## Endpoints Overview
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/resource` | List resources |
| POST | `/api/v1/resource` | Create resource |

## Quick Start
### Authentication
```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'
```

### Basic Operations
```bash
# Create resource
curl -X POST "endpoint" -H "Auth header" -d 'data'

# Get resource
curl -X GET "endpoint" -H "Auth header"
```

## Detailed Examples
### Python Client
```python
# Production-ready example
class APIClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
    
    def create_resource(self, data):
        # Implementation
        pass
```

## Error Handling
### Common Errors
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Description"
  }
}
```

## Best Practices
- Security considerations
- Performance tips
- Integration patterns
```

#### 7.2 Add to Search Enhancement
Update `docs/javascripts/search-enhancements.js` to include your API in search:

```javascript
const apiEndpoints = [
    // Add your new endpoints here
    { 
        path: '/api/v1/your-endpoint', 
        method: 'POST', 
        category: 'Your Category', 
        description: 'Your endpoint description' 
    },
];
```

### Step 8: Advanced Features

#### 8.1 Interactive Elements
For interactive examples, you can include:

```markdown
<!-- Try it out button -->
<div style="margin: 1rem 0;">
  <a href="../swagger-ui/" class="md-button md-button--primary">
    🚀 Try in API Explorer
  </a>
</div>

<!-- Code with copy button (automatic) -->
```bash
# This code block will have a copy button automatically
make run
```
```

#### 8.2 Contextual Help Integration
Your documentation will automatically benefit from the contextual help system if you include relevant keywords:

- **API Documentation**: Include "api", "endpoint", "authentication"
- **User Guides**: Include "user", "profile", "management"
- **Development**: Include "development", "contributing", "architecture"

### Step 9: Review and Quality Assurance

#### 9.1 Content Review
- [ ] Clear purpose and audience defined
- [ ] Logical information flow
- [ ] Consistent terminology
- [ ] Working code examples
- [ ] Proper grammar and spelling

#### 9.2 Technical Review
- [ ] All links functional
- [ ] Code examples tested
- [ ] API endpoints verified
- [ ] Cross-references accurate

#### 9.3 User Experience Review
- [ ] Mobile-friendly formatting
- [ ] Logical navigation
- [ ] Search-friendly content
- [ ] Appropriate detail level

## Common Patterns and Examples

### Pattern 1: Step-by-Step Tutorial
```markdown
## Installation Guide

### Step 1: Prerequisites
Ensure you have the required software installed.

### Step 2: Download
```bash
git clone https://github.com/repo/project.git
```

### Step 3: Configure
Edit the configuration file:
```yaml
# config.yml
setting: value
```

### Step 4: Run
```bash
make start
```
```

### Pattern 2: Reference Documentation
```markdown
## Configuration Reference

### Database Settings
| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `host` | string | `localhost` | Database host |
| `port` | int | `5432` | Database port |

### Example Configuration
```yaml
database:
  host: "localhost"
  port: 5432
  name: "erp_db"
```
```

### Pattern 3: API Endpoint Documentation
```markdown
## Create User

### Endpoint
```
POST /api/v1/users
```

### Request Body
```json
{
  "username": "string",
  "email": "string", 
  "roles": ["string"]
}
```

### Response
```json
{
  "data": {
    "id": "uuid",
    "username": "string",
    "created_at": "timestamp"
  }
}
```

### Example
```bash
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"username": "john.doe", "email": "john@example.com"}'
```
```

## Troubleshooting

### Common Issues

#### 1. Navigation Not Showing
**Problem**: New page doesn't appear in navigation
**Solution**: Add the page to `nav` section in `mkdocs.yml`

#### 2. Broken Internal Links
**Problem**: Links show 404 errors
**Solution**: Use relative paths and verify file locations
```markdown
# Correct
[API Guide](../api/index.md)

# Incorrect  
[API Guide](/api/index.md)
```

#### 3. Search Not Finding Content
**Problem**: New documentation doesn't appear in search
**Solution**: Rebuild the site and ensure keywords in content
```bash
mkdocs build --clean
```

#### 4. Code Blocks Not Formatting
**Problem**: Code doesn't have syntax highlighting
**Solution**: Specify language after opening backticks
```markdown
# Good
```python
def example():
    pass
```

# Not highlighted
```
def example():
    pass
```
```

## Best Practices Summary

### Content
- ✅ Start with user needs and goals
- ✅ Provide working examples
- ✅ Include error handling
- ✅ Link to related resources
- ✅ Keep content updated

### Structure  
- ✅ Use consistent headings hierarchy
- ✅ Group related topics logically
- ✅ Provide clear navigation paths
- ✅ Include "Next Steps" sections

### Technical
- ✅ Test all code examples
- ✅ Verify all links work
- ✅ Optimize for search
- ✅ Ensure mobile compatibility
- ✅ Use semantic markup

### Maintenance
- ✅ Review quarterly for accuracy
- ✅ Update examples with API changes
- ✅ Monitor for broken links
- ✅ Gather user feedback

## Getting Help

If you need assistance with documentation:

1. **Technical Issues**: Check the [troubleshooting guide](01-best-practices.md)
2. **Content Questions**: Review existing similar documentation
3. **Style Guidelines**: Follow patterns from established docs
4. **API Documentation**: Refer to existing API guides in `reference/api/`

---

**Related**: [Best Practices](01-best-practices.md) | **Up**: [Contributing](../contributing/)