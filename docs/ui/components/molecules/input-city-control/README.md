# Input City Control Component

**FILE PURPOSE**: City selection control with autocomplete and geographic validation  
**SCOPE**: City picking, geographic filtering, location-based data entry, and regional selection  
**TARGET AUDIENCE**: Developers implementing location-based forms, address management, and geographic data entry

## 📋 Component Overview

Input City Control provides specialized city selection functionality with autocomplete, country/state filtering, and geographic validation. Essential for address forms and location-based data entry in ERP systems.

### Schema Reference
- **Primary Schema**: `InputCityControlSchema.json`
- **Related Schemas**: `BaseApiObject.json`, `Option.json`
- **Base Interface**: Form input control for city selection

## Basic Usage

```json
{
    "type": "input-city",
    "name": "city",
    "label": "City",
    "placeholder": "Enter city name..."
}
```

## Go Type Definition

```go
type InputCityControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Value              string              `json:"value"`
    Source             interface{}         `json:"source"`           // Cities data source
    SearchApi          interface{}         `json:"searchApi"`        // Search endpoint
    Country            string              `json:"country"`          // Filter by country
    State              string              `json:"state"`            // Filter by state/province
    Region             string              `json:"region"`           // Filter by region
    MinLength          int                 `json:"minLength"`        // Min search length
    SearchEnable       bool                `json:"searchEnable"`     // Enable search
    Clearable          bool                `json:"clearable"`
    Disabled           bool                `json:"disabled"`
    ReadOnly           bool                `json:"readOnly"`
    Required           bool                `json:"required"`
    ShowPopulation     bool                `json:"showPopulation"`   // Show population data
    ShowTimeZone       bool                `json:"showTimeZone"`     // Show timezone
    ShowCoordinates    bool                `json:"showCoordinates"`  // Show lat/lng
    AllowCustom        bool                `json:"allowCustom"`      // Allow custom entries
    ValidateExists     bool                `json:"validateExists"`   // Validate city exists
    AutoFill           interface{}         `json:"autoFill"`         // Auto-fill related fields
    ItemRender         interface{}         `json:"itemRender"`       // Custom item template
    ValueField         string              `json:"valueField"`       // Value property
    LabelField         string              `json:"labelField"`       // Display property
}
```

## Essential Variants

### Basic City Input
```json
{
    "type": "input-city",
    "name": "shipping_city",
    "label": "Shipping City",
    "placeholder": "Enter city name...",
    "searchEnable": true,
    "clearable": true,
    "minLength": 2
}
```

### Country-Filtered Cities
```json
{
    "type": "input-city",
    "name": "delivery_city",
    "label": "Delivery City",
    "placeholder": "Select delivery city...",
    "country": "${shipping_country}",
    "searchEnable": true,
    "validateExists": true,
    "clearable": true
}
```

### State/Province Filtered
```json
{
    "type": "input-city",
    "name": "branch_city",
    "label": "Branch City",
    "placeholder": "Select city for new branch...",
    "country": "${country}",
    "state": "${state}",
    "showPopulation": true,
    "showTimeZone": true,
    "searchEnable": true
}
```

### City with Auto-fill
```json
{
    "type": "input-city",
    "name": "customer_city",
    "label": "Customer City",
    "placeholder": "Enter customer city...",
    "searchEnable": true,
    "autoFill": {
        "api": "/api/cities/${value}/details",
        "fillMapping": {
            "timezone": "timezone",
            "postal_codes": "zip_codes",
            "area_code": "phone_area_code"
        }
    }
}
```

## Real-World Use Cases

### Shipping Address City
```json
{
    "type": "input-city",
    "name": "shipping_city",
    "label": "Shipping City",
    "placeholder": "Enter shipping city...",
    "country": "${shipping_country}",
    "state": "${shipping_state}",
    "searchApi": "/api/cities/search",
    "searchEnable": true,
    "minLength": 2,
    "clearable": true,
    "required": true,
    "validateExists": true,
    "showPopulation": false,
    "itemRender": {
        "type": "html",
        "html": "<div class='city-item'><strong>${name}</strong><br><small>${state}, ${country}</small></div>"
    },
    "autoFill": {
        "api": "/api/shipping/city-data/${value}",
        "fillMapping": {
            "shipping_zone": "delivery_zone",
            "shipping_cost": "base_shipping_cost",
            "delivery_days": "estimated_delivery_days",
            "timezone": "timezone",
            "postal_codes": "available_postal_codes"
        }
    },
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Shipping city is required"
    }
}
```

### Branch Location City
```json
{
    "type": "input-city",
    "name": "new_branch_city",
    "label": "New Branch City",
    "placeholder": "Select city for new branch...",
    "country": "${target_country}",
    "state": "${target_state}",
    "region": "${target_region}",
    "searchEnable": true,
    "showPopulation": true,
    "showTimeZone": true,
    "showCoordinates": true,
    "minLength": 2,
    "clearable": true,
    "required": true,
    "itemRender": {
        "type": "html",
        "html": "<div class='branch-city-item'><h4>${name}</h4><p>Population: ${population:number}</p><p>Timezone: ${timezone}</p><small>${state}, ${country}</small></div>"
    },
    "autoFill": {
        "api": "/api/market-analysis/city/${value}",
        "fillMapping": {
            "market_size": "population",
            "competition_level": "competitor_count",
            "economic_indicators": "gdp_per_capita",
            "business_climate": "business_rating",
            "cost_of_operations": "operating_cost_index",
            "talent_availability": "workforce_score"
        }
    }
}
```

### Customer Location City
```json
{
    "type": "input-city",
    "name": "customer_city",
    "label": "Customer City",
    "placeholder": "Enter customer city...",
    "country": "${customer_country}",
    "searchEnable": true,
    "minLength": 2,
    "clearable": true,
    "allowCustom": true,
    "validateExists": false,
    "autoFill": {
        "api": "/api/customers/city-insights/${value}",
        "fillMapping": {
            "sales_territory": "territory",
            "local_currency": "currency",
            "tax_jurisdiction": "tax_zone",
            "business_hours": "standard_hours",
            "language": "primary_language"
        }
    },
    "hint": "Enter the city where your customer is located"
}
```

### Vendor/Supplier City
```json
{
    "type": "input-city",
    "name": "vendor_city",
    "label": "Vendor City",
    "placeholder": "Enter vendor city...",
    "country": "${vendor_country}",
    "state": "${vendor_state}",
    "searchEnable": true,
    "showTimeZone": true,
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/vendors/city-logistics/${value}",
        "fillMapping": {
            "shipping_hub": "nearest_hub",
            "customs_office": "customs_location",
            "transport_costs": "logistics_cost_index",
            "lead_times": "average_lead_time",
            "trade_routes": "available_routes"
        }
    }
}
```

### Employee Residence City
```json
{
    "type": "input-city",
    "name": "employee_city",
    "label": "Residence City",
    "placeholder": "Enter residence city...",
    "country": "${residence_country}",
    "state": "${residence_state}",
    "searchEnable": true,
    "showTimeZone": true,
    "clearable": true,
    "autoFill": {
        "api": "/api/hr/city-benefits/${value}",
        "fillMapping": {
            "cost_of_living": "living_cost_index",
            "tax_implications": "local_tax_rate",
            "commute_options": "transport_options",
            "remote_work_policy": "remote_eligibility"
        }
    }
}
```

### Warehouse/Distribution City
```json
{
    "type": "input-city",
    "name": "warehouse_city",
    "label": "Warehouse City",
    "placeholder": "Select warehouse city...",
    "country": "${facility_country}",
    "region": "${target_region}",
    "searchEnable": true,
    "showPopulation": true,
    "clearable": true,
    "required": true,
    "itemRender": {
        "type": "html",
        "html": "<div class='warehouse-city'><strong>${name}</strong><br><span class='population'>Pop: ${population:number}</span><br><small>${state}, ${country}</small></div>"
    },
    "autoFill": {
        "api": "/api/logistics/city-analysis/${value}",
        "fillMapping": {
            "transportation_score": "transport_rating",
            "warehouse_availability": "facility_availability",
            "labor_costs": "warehouse_labor_cost",
            "infrastructure_rating": "infrastructure_score",
            "proximity_to_customers": "customer_density"
        }
    }
}
```

### Sales Territory City
```json
{
    "type": "input-city",
    "name": "territory_city",
    "label": "Territory City",
    "placeholder": "Select territory city...",
    "country": "${sales_country}",
    "state": "${sales_state}",
    "searchEnable": true,
    "showPopulation": true,
    "clearable": true,
    "autoFill": {
        "api": "/api/sales/territory-data/${value}",
        "fillMapping": {
            "market_potential": "market_score",
            "competitor_presence": "competition_level",
            "sales_rep_assignment": "assigned_rep",
            "quota_allocation": "sales_quota",
            "historical_performance": "past_performance"
        }
    }
}
```

### Service Area City
```json
{
    "type": "input-city",
    "name": "service_city",
    "label": "Service City",
    "placeholder": "Enter service city...",
    "country": "${service_country}",
    "state": "${service_state}",
    "searchEnable": true,
    "showTimeZone": true,
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/service/city-coverage/${value}",
        "fillMapping": {
            "service_available": "coverage_available",
            "response_time": "avg_response_time",
            "service_center": "nearest_center",
            "technician_count": "available_technicians",
            "service_hours": "operating_hours"
        }
    }
}
```

### Event/Conference City
```json
{
    "type": "input-city",
    "name": "event_city",
    "label": "Event City",
    "placeholder": "Select event city...",
    "country": "${event_country}",
    "searchEnable": true,
    "showPopulation": true,
    "showTimeZone": true,
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/events/city-venues/${value}",
        "fillMapping": {
            "venue_options": "available_venues",
            "hotel_capacity": "accommodation_options",
            "airport_access": "airport_connections",
            "transport_options": "local_transport",
            "event_costs": "estimated_costs"
        }
    }
}
```

### Manufacturing Site City
```json
{
    "type": "input-city",
    "name": "manufacturing_city",
    "label": "Manufacturing City",
    "placeholder": "Select manufacturing city...",
    "country": "${manufacturing_country}",
    "region": "${industrial_region}",
    "searchEnable": true,
    "showPopulation": true,
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/manufacturing/city-profile/${value}",
        "fillMapping": {
            "industrial_zones": "available_zones",
            "utility_costs": "energy_costs",
            "labor_availability": "skilled_workforce",
            "raw_material_access": "supplier_proximity",
            "environmental_regulations": "compliance_requirements",
            "tax_incentives": "manufacturing_incentives"
        }
    }
}
```

This component provides essential city selection functionality for ERP systems requiring location-based data entry, geographic filtering, and address management with intelligent autocomplete and validation.