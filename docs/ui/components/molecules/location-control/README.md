# Location Control Component

**FILE PURPOSE**: Geographic location selection and coordinate input control  
**SCOPE**: Location picking, address input, coordinate selection, and geographic data management  
**TARGET AUDIENCE**: Developers implementing location-based features, address management, and geographic data entry

## 📋 Component Overview

Location Control provides comprehensive location selection functionality with address autocomplete, coordinate input, map integration, and geographic validation. Essential for location-based ERP features and spatial data management.

### Schema Reference
- **Primary Schema**: `LocationControlSchema.json`
- **Related Schemas**: `BaseApiObject.json`, `AddressSchema.json`
- **Base Interface**: Form input control for location/address selection

## Basic Usage

```json
{
    "type": "location-picker",
    "name": "facility_location",
    "label": "Facility Location",
    "placeholder": "Enter address or coordinates..."
}
```

## Go Type Definition

```go
type LocationControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Value              interface{}         `json:"value"`
    Vendor             string              `json:"vendor"`           // Map provider
    Ak                 string              `json:"ak"`               // API key
    MapType            string              `json:"mapType"`          // Map display type
    CoordinateType     string              `json:"coordinateType"`   // Coordinate system
    Format             string              `json:"format"`           // Output format
    OnlySelectCurrentLoc bool              `json:"onlySelectCurrentLoc"` // Current location only
    AutoLocation       bool                `json:"autoLocation"`     // Auto detect location
    RestrictLocation   interface{}         `json:"restrictLocation"` // Geographic bounds
    SearchRadius       int                 `json:"searchRadius"`     // Search radius (km)
    EnableAddressSearch bool               `json:"enableAddressSearch"` // Address autocomplete
    EnableCoordinates  bool                `json:"enableCoordinates"`   // Show coordinates
    EnableCurrentLocation bool             `json:"enableCurrentLocation"` // Current location button
    EnableMapPicker    bool                `json:"enableMapPicker"`     // Map selection
    DefaultZoom        int                 `json:"defaultZoom"`         // Default zoom level
    MinZoom            int                 `json:"minZoom"`             // Minimum zoom
    MaxZoom            int                 `json:"maxZoom"`             // Maximum zoom
    MarkerIcon         string              `json:"markerIcon"`          // Custom marker
    ShowAddress        bool                `json:"showAddress"`         // Display address
    ShowCoordinates    bool                `json:"showCoordinates"`     // Display coordinates
    AddressFormat      string              `json:"addressFormat"`       // Address display format
    ValidationRadius   int                 `json:"validationRadius"`    // Validation distance
}
```

## Essential Variants

### Basic Address Picker
```json
{
    "type": "location-picker",
    "name": "shipping_address",
    "label": "Shipping Address",
    "placeholder": "Enter shipping address...",
    "enableAddressSearch": true,
    "showAddress": true,
    "vendor": "google",
    "format": "address"
}
```

### Coordinate Input
```json
{
    "type": "location-picker",
    "name": "warehouse_coordinates",
    "label": "Warehouse Coordinates",
    "placeholder": "Enter coordinates or select on map...",
    "enableCoordinates": true,
    "enableMapPicker": true,
    "showCoordinates": true,
    "coordinateType": "wgs84",
    "format": "coordinates"
}
```

### Map-Based Location Picker
```json
{
    "type": "location-picker",
    "name": "site_location",
    "label": "Site Location",
    "placeholder": "Select location on map...",
    "enableMapPicker": true,
    "enableCurrentLocation": true,
    "defaultZoom": 10,
    "showAddress": true,
    "showCoordinates": true,
    "autoLocation": false
}
```

### Restricted Area Picker
```json
{
    "type": "location-picker",
    "name": "service_location",
    "label": "Service Location",
    "placeholder": "Select location within service area...",
    "enableAddressSearch": true,
    "enableMapPicker": true,
    "restrictLocation": {
        "bounds": {
            "north": 40.7829,
            "south": 40.7489,
            "east": -73.9441,
            "west": -73.9927
        }
    },
    "validationRadius": 50
}
```

## Real-World Use Cases

### Facility/Branch Location Management
```json
{
    "type": "location-picker",
    "name": "branch_location",
    "label": "Branch Location",
    "placeholder": "Enter branch address or select on map...",
    "enableAddressSearch": true,
    "enableMapPicker": true,
    "enableCurrentLocation": true,
    "showAddress": true,
    "showCoordinates": true,
    "vendor": "google",
    "ak": "${GOOGLE_MAPS_API_KEY}",
    "defaultZoom": 15,
    "minZoom": 5,
    "maxZoom": 20,
    "coordinateType": "wgs84",
    "format": "full",
    "addressFormat": "formatted",
    "markerIcon": "/images/branch-marker.png",
    "required": true,
    "autoFill": {
        "api": "/api/locations/geocode/${value}",
        "fillMapping": {
            "latitude": "lat",
            "longitude": "lng",
            "formatted_address": "address",
            "postal_code": "zip",
            "city": "city",
            "state": "state",
            "country": "country"
        }
    },
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Branch location is required"
    }
}
```

### Delivery Address Selection
```json
{
    "type": "location-picker",
    "name": "delivery_location",
    "label": "Delivery Address",
    "placeholder": "Enter delivery address...",
    "enableAddressSearch": true,
    "enableMapPicker": true,
    "showAddress": true,
    "vendor": "google",
    "searchRadius": 100,
    "restrictLocation": {
        "country": "US",
        "bounds": {
            "north": 49.3457868,
            "south": 24.7433195,
            "east": -66.9513812,
            "west": -124.7844079
        }
    },
    "validationRadius": 200,
    "autoFill": {
        "api": "/api/delivery/validate-address/${value}",
        "fillMapping": {
            "delivery_zone": "zone",
            "delivery_fee": "fee",
            "estimated_time": "eta",
            "service_available": "available"
        }
    },
    "hint": "Enter a valid delivery address within our service area"
}
```

### Asset Location Tracking
```json
{
    "type": "location-picker",
    "name": "asset_location",
    "label": "Asset Location",
    "placeholder": "Mark asset location on map...",
    "enableMapPicker": true,
    "enableCurrentLocation": true,
    "enableCoordinates": true,
    "showCoordinates": true,
    "showAddress": true,
    "defaultZoom": 18,
    "coordinateType": "wgs84",
    "format": "coordinates",
    "markerIcon": "/images/asset-marker.png",
    "autoLocation": true,
    "required": true,
    "autoFill": {
        "api": "/api/assets/location-metadata/${value}",
        "fillMapping": {
            "location_accuracy": "accuracy",
            "location_timestamp": "timestamp",
            "nearest_landmark": "landmark",
            "zone": "facility_zone"
        }
    }
}
```

### Customer Site Location
```json
{
    "type": "location-picker",
    "name": "customer_site",
    "label": "Customer Site",
    "placeholder": "Enter customer site address...",
    "enableAddressSearch": true,
    "enableMapPicker": true,
    "showAddress": true,
    "showCoordinates": true,
    "vendor": "google",
    "addressFormat": "detailed",
    "searchRadius": 50,
    "defaultZoom": 16,
    "autoFill": {
        "api": "/api/customers/site-details/${value}",
        "fillMapping": {
            "site_name": "name",
            "site_contact": "contact_person",
            "site_phone": "phone",
            "access_instructions": "instructions",
            "parking_available": "parking",
            "site_type": "type"
        }
    },
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Customer site location is required"
    }
}
```

### Warehouse/Distribution Center
```json
{
    "type": "location-picker",
    "name": "warehouse_location",
    "label": "Warehouse Location",
    "placeholder": "Select warehouse location...",
    "enableAddressSearch": true,
    "enableMapPicker": true,
    "enableCoordinates": true,
    "showAddress": true,
    "showCoordinates": true,
    "vendor": "google",
    "coordinateType": "wgs84",
    "format": "full",
    "defaultZoom": 15,
    "markerIcon": "/images/warehouse-marker.png",
    "restrictLocation": {
        "zoning": "industrial"
    },
    "autoFill": {
        "api": "/api/warehouses/location-analysis/${value}",
        "fillMapping": {
            "transportation_access": "transport_score",
            "nearest_highway": "highway_distance",
            "nearest_port": "port_distance",
            "nearest_airport": "airport_distance",
            "population_density": "density",
            "zoning_type": "zoning"
        }
    }
}
```

### Service Territory Definition
```json
{
    "type": "location-picker",
    "name": "service_territory_center",
    "label": "Service Territory Center",
    "placeholder": "Define center of service territory...",
    "enableMapPicker": true,
    "enableAddressSearch": true,
    "showAddress": true,
    "showCoordinates": true,
    "defaultZoom": 12,
    "format": "coordinates",
    "autoFill": {
        "api": "/api/territories/coverage-analysis/${value}",
        "fillMapping": {
            "coverage_radius": "radius",
            "customer_density": "density",
            "competitor_presence": "competition",
            "market_potential": "potential",
            "territory_code": "code"
        }
    }
}
```

### Incident/Emergency Location
```json
{
    "type": "location-picker",
    "name": "incident_location",
    "label": "Incident Location",
    "placeholder": "Mark exact incident location...",
    "enableMapPicker": true,
    "enableCurrentLocation": true,
    "enableCoordinates": true,
    "showCoordinates": true,
    "showAddress": true,
    "autoLocation": true,
    "defaultZoom": 18,
    "coordinateType": "wgs84",
    "format": "coordinates",
    "markerIcon": "/images/incident-marker.png",
    "required": true,
    "autoFill": {
        "api": "/api/incidents/location-data/${value}",
        "fillMapping": {
            "nearest_responder": "closest_unit",
            "response_time": "eta",
            "location_type": "area_type",
            "access_routes": "routes",
            "hazards": "potential_hazards"
        }
    }
}
```

### Construction Site Location
```json
{
    "type": "location-picker",
    "name": "construction_site",
    "label": "Construction Site",
    "placeholder": "Mark construction site location...",
    "enableMapPicker": true,
    "enableAddressSearch": true,
    "enableCoordinates": true,
    "showAddress": true,
    "showCoordinates": true,
    "vendor": "google",
    "defaultZoom": 16,
    "coordinateType": "wgs84",
    "format": "full",
    "markerIcon": "/images/construction-marker.png",
    "autoFill": {
        "api": "/api/construction/site-analysis/${value}",
        "fillMapping": {
            "soil_type": "soil_classification",
            "elevation": "altitude",
            "flood_zone": "flood_risk",
            "utility_access": "utilities",
            "permits_required": "permit_types",
            "environmental_concerns": "environmental_factors"
        }
    }
}
```

### Fleet Vehicle Location
```json
{
    "type": "location-picker",
    "name": "vehicle_location",
    "label": "Vehicle Location",
    "placeholder": "Current vehicle location...",
    "enableCurrentLocation": true,
    "enableCoordinates": true,
    "showCoordinates": true,
    "showAddress": true,
    "autoLocation": true,
    "onlySelectCurrentLoc": true,
    "coordinateType": "wgs84",
    "format": "coordinates",
    "defaultZoom": 16,
    "autoFill": {
        "api": "/api/fleet/location-update/${value}",
        "fillMapping": {
            "location_timestamp": "timestamp",
            "speed": "current_speed",
            "heading": "direction",
            "nearest_checkpoint": "checkpoint",
            "route_deviation": "off_route",
            "fuel_stations_nearby": "fuel_nearby"
        }
    }
}
```

### Vendor/Supplier Location
```json
{
    "type": "location-picker",
    "name": "supplier_location",
    "label": "Supplier Location",
    "placeholder": "Enter supplier facility address...",
    "enableAddressSearch": true,
    "enableMapPicker": true,
    "showAddress": true,
    "showCoordinates": true,
    "vendor": "google",
    "addressFormat": "business",
    "searchRadius": 100,
    "autoFill": {
        "api": "/api/suppliers/location-metrics/${value}",
        "fillMapping": {
            "shipping_distance": "distance_km",
            "shipping_cost": "transport_cost",
            "delivery_time": "lead_time",
            "customs_zone": "trade_zone",
            "time_zone": "timezone",
            "business_hours": "operating_hours"
        }
    }
}
```

This component provides essential location management functionality for ERP systems requiring geographic data handling, address management, and location-based business operations.