# AWO ERP Documentation Structure Guide

## Overview

This guide defines the standardized documentation structure for the AWO ERP system. It provides templates and guidelines for creating consistent,  documentation across all modules and features.

## Standard Documentation Structure

Each module should follow this standardized structure:

```
docs/reference/modules/{module-name}/
├── README.md                    # Module overview and quick start  
├── TASK.md                     # Development progress tracking
├── PRD.md                      # Product requirements document
├── testing.md                  # Test requirements and progress
├── architecture-guide.md       # Technical architecture
├── api-reference.md           # API endpoints and schemas
├── integration-guide.md        # Integration with other modules
├── deployment-guide.md         # Deployment and configuration
├── security-compliance-guide.md # Security considerations
└── examples/                   # Code examples and samples
    ├── basic-usage.md
    ├── advanced-scenarios.md
    └── integration-examples.md
```

## Template Usage Guide

### 1. Module README Template

**Location**: `docs/templates/module-README-template.md`

**Usage**:
1. Copy template to your module directory as `README.md`
2. Replace all `{Module Name}` placeholders
3. Update architecture diagrams with actual relationships
4. Fill in real API endpoints and status
5. Update integration points with actual dependencies
6. Set current development status and metrics

**Key Sections to Customize**:
- **Architecture Overview**: Include actual domain entities and services
- **API Endpoints**: List real endpoints with current implementation status
- **Integration Points**: Actual module dependencies and external services
- **Development Status**: Current progress percentages and completed features

### 2. TASK.md Template

**Location**: `docs/templates/TASK-template.md`

**Usage**:
1. Copy template as `TASK.md` in module directory
2. Replace `{Module Name}` with actual module name
3. Update progress percentages based on current completion
4. Customize phases based on your module's implementation plan
5. Fill in actual completion dates and developer assignments
6. Update technical achievements and next priorities

**Key Customizations**:
- **Progress Overview**: Real completion percentages per phase
- **Implementation Plan**: Module-specific phases and milestones  
- **Code Metrics**: Actual test coverage and quality scores
- **Recent Completions**: Real completion dates and achievements

### 3. testing.md Template

**Location**: `docs/templates/testing-template.md`

**Usage**:
1. Copy template as `testing.md` in module directory
2. Replace `{Module Name}` and `{MOD}` placeholders
3. Update test IDs to match your module prefix
4. Customize test cases for your module's business logic
5. Update test status checkboxes as tests are implemented
6. Add module-specific business rules and security requirements

**Test ID Convention**:
- Domain tests: `{MOD}-DOMAIN-001`, `{MOD}-DOMAIN-002`, etc.
- Service tests: `{MOD}-SERVICE-001`, `{MOD}-SERVICE-002`, etc.
- Repository tests: `{MOD}-REPO-001`, `{MOD}-REPO-002`, etc.

### 4. PRD.md Template

**Location**: `docs/templates/PRD-template.md`

**Usage**:
1. Copy template as `PRD.md` in module directory
2. Replace `{Module Name}` with actual module name
3. Fill in real business requirements and user stories
4. Update technical architecture with actual service interfaces
5. Add real risk assessments and mitigation strategies
6. Set actual timeline and success metrics

**Critical Sections**:
- **Problem Statement**: Real business problem being solved
- **User Stories**: Actual user needs and acceptance criteria
- **Technical Architecture**: Real service interfaces and data models
- **Implementation Timeline**: Actual project phases and deliverables

### 5. API Reference Template

**Location**: `docs/templates/api-specification-template.md`

**Usage**:
1. Copy template as `api-reference.md` in module directory
2. Replace all `{module}` and `{Module Name}` placeholders
3. Update base URL with actual API paths
4. Add real endpoint documentation with actual request/response examples
5. Update data models with actual entity schemas
6. Add real error codes and business rules

**Integration**:
- Generate OpenAPI specs from Goa designs
- Keep examples in sync with actual API behavior
- Update error codes based on actual service implementation

## Documentation Templates

### Available Templates

| Template | Purpose | Location |
|----------|---------|----------|
| Module README | Main module documentation | `docs/templates/module-README-template.md` |
| TASK.md | Development progress tracking | `docs/templates/TASK-template.md` |
| Testing | Test strategy and cases | `docs/templates/testing-template.md` |  
| PRD | Product requirements | `docs/templates/PRD-template.md` |
| API Reference | API documentation | `docs/templates/api-specification-template.md` |

### Using Templates

1. **Copy Template**: Copy the appropriate template to your module directory
2. **Replace Placeholders**: Search and replace all placeholder values
3. **Customize Content**: Update with module-specific information
4. **Update Status**: Set current implementation status and progress
5. **Review & Commit**: Have template reviewed before committing

### Placeholder Convention

All templates use consistent placeholder naming:
- `{Module Name}`: Full module name (e.g., "Financial", "User Management")
- `{module}`: Lowercase module identifier (e.g., "finance", "user") 
- `{MOD}`: Uppercase module abbreviation (e.g., "FIN", "USR")
- `{Current Date}`: Today's date in YYYY-MM-DD format
- `{Date}`: Actual date to be filled in

## Implementation Workflow

### For New Modules

1. **Create Module Directory**:
   ```bash
   mkdir -p docs/reference/modules/{module-name}
   ```

2. **Copy Core Templates**:
   ```bash
   cp docs/templates/module-README-template.md docs/reference/modules/{module-name}/README.md
   cp docs/templates/TASK-template.md docs/reference/modules/{module-name}/TASK.md
   cp docs/templates/testing-template.md docs/reference/modules/{module-name}/testing.md
   ```

3. **Customize Templates**: Replace placeholders and add module-specific content

4. **Add to Navigation**: Update `mkdocs.yml` to include new module

5. **Create Examples Directory**:
   ```bash
   mkdir docs/reference/modules/{module-name}/examples
   ```

### For Existing Modules

1. **Assess Current Documentation**: Identify gaps compared to standard structure

2. **Migrate Existing Content**: Preserve existing documentation while adopting templates

3. **Fill Documentation Gaps**: Create missing documentation using templates  

4. **Standardize Format**: Update existing docs to match template structure

5. **Update Navigation**: Ensure all docs are properly linked

## Quality Standards

### Documentation Completeness

Each module should have:
- [x] README.md with overview and quick start
- [x] TASK.md with current development progress  
- [x] testing.md with  test strategy
- [x] PRD.md with business requirements
- [x] api-reference.md with complete API documentation
- [x] Examples directory with code samples

### Content Quality

- **Accuracy**: Information matches actual implementation
- **Completeness**: All required sections are filled out
- **Clarity**: Writing is clear and easy to understand
- **Consistency**: Follows templates and style guidelines
- **Currency**: Documentation is kept up-to-date with code changes

### Review Process

1. **Self Review**: Author reviews documentation for completeness and accuracy
2. **Peer Review**: Another developer reviews for technical accuracy
3. **Product Review**: Product owner reviews PRD and user-facing content
4. **Final Approval**: Tech lead approves before merge

## Maintenance Guidelines

### Regular Updates

- **Weekly**: Update TASK.md progress tracking
- **Sprint End**: Update README.md development status
- **Release**: Update version numbers and changelog entries
- **Quarterly**: Review and update PRD business requirements

### Change Management

- **Code Changes**: Update affected documentation in same PR
- **API Changes**: Update api-reference.md immediately  
- **Business Changes**: Update PRD.md and cascade to other docs
- **Architecture Changes**: Update architecture diagrams and integration guides

### Documentation Debt

Track documentation debt in TASK.md:
- Missing documentation sections
- Outdated information
- Incomplete examples
- Broken links or references

## Integration with MkDocs

### Navigation Structure

Update `mkdocs.yml` for each new module:

```yaml
nav:
  - 'Module Name':
    - 'Overview': 'reference/modules/{module}/README.md'
    - 'Development Tasks': 'reference/modules/{module}/TASK.md'  
    - 'Product Requirements': 'reference/modules/{module}/PRD.md'
    - 'Testing Strategy': 'reference/modules/{module}/testing.md'
    - 'API Reference': 'reference/modules/{module}/api-reference.md'
    - 'Architecture': 'reference/modules/{module}/architecture-guide.md'
    - 'Integration': 'reference/modules/{module}/integration-guide.md'
    - 'Deployment': 'reference/modules/{module}/deployment-guide.md'
    - 'Security': 'reference/modules/{module}/security-compliance-guide.md'
```

### Cross-References

Use relative links between documentation files:
- `[API Reference](api-reference.md)` - within module
- `[User Module](../user/README.md)` - to other modules  
- `[Contributing](../../contributing/01-best-practices.md)` - to guides

## Best Practices

### Writing Guidelines

1. **Use Active Voice**: "The service processes requests" vs "Requests are processed"
2. **Be Specific**: Include concrete examples and data
3. **Structure Information**: Use headings, lists, and tables effectively
4. **Include Diagrams**: Visual representations improve understanding
5. **Provide Examples**: Code examples make documentation actionable

### Backend Focus

Since this is primarily a backend system:
- **Emphasize APIs**: Detailed API documentation is critical
- **Include Database Schema**: Document data models and relationships
- **Show Integration Patterns**: How services communicate
- **Cover Error Handling**: Comprehensive error documentation
- **Performance Metrics**: Include benchmarks and SLAs

### Maintenance Automation

- **Link Checking**: Automated broken link detection
- **API Sync**: Keep API docs in sync with OpenAPI specs
- **Template Updates**: Propagate template improvements to existing docs
- **Metric Tracking**: Automate documentation health metrics

---

**Next Steps**: 
1. Choose a module to start with documentation standardization
2. Copy and customize the appropriate templates  
3. Update MkDocs navigation to include the new structure
4. Review and iterate based on team feedback

This standardized approach ensures consistent,  documentation across the entire AWO ERP system while reducing the effort needed to create and maintain high-quality documentation.