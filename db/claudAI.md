# ERP Hierarchy Design Guide

## Hierarchy Levels Explained

### 1. **COMPANY/SUBSIDIARY**
- **When to use**: Multi-company scenarios, holding companies, or legal entities
- **Examples**: 
  - ABC Holdings → ABC Manufacturing, ABC Services
  - Main Company → Regional Subsidiaries
- **Financial Impact**: Separate P&L, balance sheets, tax filings

### 2. **REGION/GEOGRAPHIC AREA**
- **When to use**: Geographically distributed operations
- **Examples**: 
  - North America, Europe, Asia-Pacific
  - State-level divisions (California, Texas, New York)
- **Purpose**: Regional reporting, compliance, management structure

### 3. **BRANCH/LOCATION**
- **When to use**: Physical locations, offices, stores, warehouses
- **Examples**: 
  - Store locations, office branches, manufacturing plants
  - Distribution centers, service centers
- **Purpose**: Location-based inventory, local P&L, address-specific operations

### 4. **DEPARTMENT/DIVISION**
- **When to use**: Functional or business unit separation
- **Examples**: 
  - Sales, Marketing, IT, HR, Finance
  - Product divisions (Electronics, Clothing, Home Goods)
- **Purpose**: Budget allocation, functional reporting, team management

### 5. **COST_CENTER**
- **When to use**: Detailed financial tracking and budgeting
- **Examples**: 
  - IT Support, Marketing Campaigns, R&D Projects
  - Maintenance, Quality Control, Training
- **Purpose**: Expense allocation, budget control, performance metrics

### 6. **PROJECT**
- **When to use**: Temporary initiatives with specific goals and timelines
- **Examples**: 
  - Product launches, system implementations, construction projects
  - Marketing campaigns, research initiatives
- **Purpose**: Project accounting, resource allocation, milestone tracking

## Real-World Hierarchy Examples

### Example 1: Manufacturing Company
```
TechCorp (Tenant)
├── TechCorp USA (Company)
│   ├── West Coast (Region)
│   │   ├── Los Angeles Plant (Branch)
│   │   │   ├── Production Department
│   │   │   │   ├── Assembly Line 1 (Cost Center)
│   │   │   │   └── Quality Control (Cost Center)
│   │   │   └── Maintenance Department
│   │   └── San Francisco Office (Branch)
│   │       ├── Sales Department
│   │       └── R&D Department
│   └── East Coast (Region)
│       └── New York Office (Branch)
└── TechCorp Europe (Company)
    └── London Office (Branch)
```

### Example 2: Retail Chain
```
RetailMax (Tenant)
├── RetailMax Corp (Company)
│   ├── Northern Region
│   │   ├── Store #001 - Downtown (Branch)
│   │   │   ├── Electronics Department
│   │   │   ├── Clothing Department
│   │   │   └── Customer Service (Cost Center)
│   │   └── Store #002 - Mall (Branch)
│   ├── Southern Region
│   │   └── Store #003 - Suburb (Branch)
│   └── Corporate Office (Branch)
│       ├── IT Department
│       ├── Marketing Department
│       └── Finance Department
```

### Example 3: Service Company
```
ConsultCorp (Tenant)
├── ConsultCorp Inc (Company)
│   ├── North America (Region)
│   │   ├── New York Office (Branch)
│   │   │   ├── Technology Practice (Department)
│   │   │   │   ├── Cloud Migration Project
│   │   │   │   └── Digital Transformation Project
│   │   │   └── Strategy Practice (Department)
│   │   └── Toronto Office (Branch)
│   └── Europe (Region)
│       └── London Office (Branch)
```

## Decision Matrix: Which Levels Do You Need?

| Business Type | Company | Region | Branch | Department | Cost Center | Project |
|---------------|---------|--------|--------|------------|-------------|---------|
| Single Location SMB | Optional | No | No | Yes | Optional | Yes |
| Multi-Location SMB | Optional | Optional | Yes | Yes | Optional | Yes |
| Regional Business | Yes | Yes | Yes | Yes | Yes | Yes |
| Multi-National | Yes | Yes | Yes | Yes | Yes | Yes |
| Franchise | Yes | Yes | Yes | Optional | Optional | Yes |
| E-commerce Only | Optional | Optional | Optional | Yes | Optional | Yes |

## Flexibility Rules

### Skip Levels When Not Needed
- Small businesses might go: Company → Department → Cost Center
- Simple structures: Company → Branch → Department
- Project-focused: Company → Department → Project

### Multiple Valid Paths
- Department can be under Company OR Branch OR Region
- Cost Centers can exist at any level
- Projects can be attached to any entity level

## Budget and Financial Reporting

### Budget Hierarchy
```sql
-- Budgets can exist at any level
INSERT INTO budgets (tenant_id, entity_id, name, budget_type, fiscal_year, total_amount)
VALUES 
  (1, 5, 'IT Department Annual Budget', 'DEPARTMENT', 2024, 500000),
  (1, 8, 'Store #001 Operating Budget', 'OPERATIONAL', 2024, 1200000),
  (1, 12, 'Cloud Migration Project Budget', 'PROJECT', 2024, 150000);
```

### Reporting Rollups
- Cost centers roll up to departments
- Departments roll up to branches/regions
- Everything rolls up to company level
- Projects can span multiple entities but report to their assigned entity

## Implementation Tips

### Start Simple, Grow Complex
1. Begin with: Company → Department → (Projects as needed)
2. Add regions when you expand geographically
3. Add branches when you have multiple locations
4. Add cost centers when you need detailed financial tracking

### Entity Codes for Integration
```sql
-- Use meaningful codes for external system integration
INSERT INTO entities (tenant_id, name, code, type)
VALUES 
  (1, 'Sales Department', 'SALES-001', 'DEPARTMENT'),
  (1, 'New York Store', 'NYC-001', 'BRANCH'),
  (1, 'West Coast Region', 'WC-REG', 'REGION');
```

### Metadata for Industry-Specific Needs
```sql
-- Store industry-specific data in metadata
UPDATE entities 
SET metadata = jsonb_build_object(
  'store_type', 'flagship',
  'square_footage', 15000,
  'lease_expiry', '2026-12-31',
  'manager_contact', '+1-555-0123'
)
WHERE code = 'NYC-001';
```
