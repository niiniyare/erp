# Retail Management

## ️ Overview

The Retail Management module transforms the core ERP into a  omnichannel retail platform. It provides multi-channel sales management, customer analytics, inventory synchronization, merchandising, and e-commerce integration designed for retailers operating across physical stores, online platforms, and mobile channels.

##  Multi-Channel Retail Operations

### Store Management

```sql
-- Retail store locations
CREATE TABLE retail_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Store identification
    store_code VARCHAR(20) UNIQUE NOT NULL,
    store_name VARCHAR(255) NOT NULL,
    store_format VARCHAR(50) DEFAULT 'standard', -- flagship, standard, outlet, popup, kiosk
    
    -- Location details
    address JSONB NOT NULL,
    gps_coordinates JSONB, -- {latitude, longitude}
    timezone VARCHAR(50) NOT NULL,
    
    -- Store characteristics
    store_size_sqft DECIMAL(10,2),
    selling_area_sqft DECIMAL(10,2),
    storage_area_sqft DECIMAL(10,2),
    
    -- Operational details
    opening_date DATE,
    store_manager_id UUID REFERENCES employees(id),
    phone_number VARCHAR(20),
    email VARCHAR(255),
    
    -- Operating hours
    operating_hours JSONB, -- Daily schedule with exceptions
    seasonal_hours JSONB, -- Holiday and seasonal variations
    
    -- Store capabilities
    has_fitting_rooms BOOLEAN DEFAULT false,
    has_cafe BOOLEAN DEFAULT false,
    has_pharmacy BOOLEAN DEFAULT false,
    has_click_and_collect BOOLEAN DEFAULT true,
    has_returns_desk BOOLEAN DEFAULT true,
    
    -- POS and technology
    pos_system_type VARCHAR(50),
    number_of_registers INTEGER DEFAULT 1,
    self_checkout_available BOOLEAN DEFAULT false,
    mobile_pos_enabled BOOLEAN DEFAULT false,
    
    -- Financial targets
    annual_sales_target DECIMAL(15,2),
    monthly_sales_target DECIMAL(12,2),
    
    -- Store status
    store_status VARCHAR(20) DEFAULT 'active', -- active, renovation, closed, seasonal
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_store_format CHECK (store_format IN ('flagship', 'standard', 'outlet', 'popup', 'kiosk')),
    CONSTRAINT valid_store_status CHECK (store_status IN ('active', 'renovation', 'closed', 'seasonal'))
);

-- Sales channels
CREATE TABLE sales_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Channel identification
    channel_code VARCHAR(20) UNIQUE NOT NULL,
    channel_name VARCHAR(255) NOT NULL,
    channel_type VARCHAR(30) NOT NULL, -- physical_store, ecommerce, mobile_app, marketplace, social_commerce
    
    -- Channel configuration
    channel_url VARCHAR(500), -- For online channels
    api_configuration JSONB, -- Integration settings
    
    -- Channel properties
    supports_inventory_sync BOOLEAN DEFAULT true,
    supports_order_sync BOOLEAN DEFAULT true,
    supports_customer_sync BOOLEAN DEFAULT true,
    supports_pricing_sync BOOLEAN DEFAULT true,
    
    -- Commission and fees
    commission_rate DECIMAL(5,4) DEFAULT 0, -- For marketplaces
    transaction_fee_rate DECIMAL(5,4) DEFAULT 0,
    monthly_fee DECIMAL(8,2) DEFAULT 0,
    
    -- Geographic scope
    supported_countries JSONB, -- Array of country codes
    supported_currencies JSONB, -- Array of currency codes
    
    -- Channel status
    is_active BOOLEAN DEFAULT true,
    integration_status VARCHAR(20) DEFAULT 'pending', -- pending, connected, error, disabled
    
    -- Associated store (for physical channels)
    retail_store_id UUID REFERENCES retail_stores(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_channel_type CHECK (channel_type IN ('physical_store', 'ecommerce', 'mobile_app', 'marketplace', 'social_commerce')),
    CONSTRAINT valid_integration_status CHECK (integration_status IN ('pending', 'connected', 'error', 'disabled'))
);

-- Product catalog for retail
CREATE TABLE retail_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Product identification
    sku VARCHAR(100) UNIQUE NOT NULL,
    upc VARCHAR(20), -- Universal Product Code
    ean VARCHAR(20), -- European Article Number
    product_name VARCHAR(255) NOT NULL,
    brand VARCHAR(100),
    
    -- Product classification
    category_id UUID REFERENCES item_categories(id),
    product_type VARCHAR(50) DEFAULT 'physical', -- physical, digital, service, subscription
    
    -- Product details
    description TEXT,
    short_description VARCHAR(500),
    features JSONB, -- Array of product features
    specifications JSONB, -- Technical specifications
    
    -- Dimensions and weight
    length_cm DECIMAL(8,2),
    width_cm DECIMAL(8,2),
    height_cm DECIMAL(8,2),
    weight_kg DECIMAL(8,4),
    
    -- Variants and options
    has_variants BOOLEAN DEFAULT false,
    variant_attributes JSONB, -- Array of variant types (size, color, etc.)
    
    -- Pricing
    cost_price DECIMAL(10,4),
    wholesale_price DECIMAL(10,2),
    retail_price DECIMAL(10,2),
    msrp DECIMAL(10,2), -- Manufacturer's Suggested Retail Price
    
    -- Inventory management
    track_inventory BOOLEAN DEFAULT true,
    safety_stock_level INTEGER DEFAULT 0,
    reorder_point INTEGER DEFAULT 0,
    max_stock_level INTEGER,
    
    -- Product lifecycle
    launch_date DATE,
    discontinue_date DATE,
    lifecycle_stage VARCHAR(20) DEFAULT 'active', -- new, active, mature, declining, discontinued
    
    -- SEO and marketing
    seo_title VARCHAR(200),
    seo_description VARCHAR(500),
    meta_keywords JSONB,
    
    -- Digital assets
    primary_image_url VARCHAR(500),
    image_gallery JSONB, -- Array of image URLs
    video_urls JSONB, -- Array of video URLs
    
    -- Compliance and certifications
    age_restriction INTEGER DEFAULT 0, -- Minimum age required
    requires_id_verification BOOLEAN DEFAULT false,
    restricted_shipping BOOLEAN DEFAULT false,
    
    -- Vendor information
    vendor_id UUID REFERENCES vendors(id),
    vendor_sku VARCHAR(100),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_product_type CHECK (product_type IN ('physical', 'digital', 'service', 'subscription')),
    CONSTRAINT valid_lifecycle_stage CHECK (lifecycle_stage IN ('new', 'active', 'mature', 'declining', 'discontinued'))
);

-- Product variants (size, color, style combinations)
CREATE TABLE product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_product_id UUID NOT NULL REFERENCES retail_products(id) ON DELETE CASCADE,
    
    -- Variant identification
    variant_sku VARCHAR(100) UNIQUE NOT NULL,
    variant_upc VARCHAR(20),
    variant_ean VARCHAR(20),
    
    -- Variant attributes
    variant_name VARCHAR(255),
    attribute_values JSONB, -- {color: "Red", size: "Large", style: "Modern"}
    
    -- Variant-specific pricing
    cost_price DECIMAL(10,4),
    retail_price DECIMAL(10,2),
    price_adjustment DECIMAL(8,2) DEFAULT 0, -- Premium or discount vs base product
    
    -- Variant-specific properties
    weight_kg DECIMAL(8,4),
    dimensions JSONB,
    
    -- Inventory tracking
    track_inventory_separately BOOLEAN DEFAULT true,
    
    -- Images and media
    variant_images JSONB, -- Array of variant-specific image URLs
    
    -- Availability
    is_active BOOLEAN DEFAULT true,
    available_for_sale BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Omnichannel Inventory Management

```sql
-- Channel-specific inventory levels
CREATE TABLE channel_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Product and channel
    product_id UUID NOT NULL REFERENCES retail_products(id),
    variant_id UUID REFERENCES product_variants(id),
    sales_channel_id UUID NOT NULL REFERENCES sales_channels(id),
    warehouse_id UUID REFERENCES warehouses(id),
    
    -- Inventory levels
    available_quantity INTEGER DEFAULT 0,
    allocated_quantity INTEGER DEFAULT 0, -- Reserved for orders
    on_order_quantity INTEGER DEFAULT 0,
    in_transit_quantity INTEGER DEFAULT 0,
    
    -- Channel-specific settings
    minimum_stock_level INTEGER DEFAULT 0,
    maximum_stock_level INTEGER,
    auto_replenish BOOLEAN DEFAULT true,
    
    -- Listing status on channel
    is_listed BOOLEAN DEFAULT true,
    listing_status VARCHAR(20) DEFAULT 'active', -- active, inactive, out_of_stock, discontinued
    listing_title VARCHAR(255),
    listing_description TEXT,
    
    -- Channel-specific pricing
    channel_price DECIMAL(10,2),
    promotional_price DECIMAL(10,2),
    price_override BOOLEAN DEFAULT false,
    
    -- Synchronization
    last_sync_at TIMESTAMPTZ,
    sync_status VARCHAR(20) DEFAULT 'pending', -- pending, synced, error
    sync_error_message TEXT,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_listing_status CHECK (listing_status IN ('active', 'inactive', 'out_of_stock', 'discontinued')),
    CONSTRAINT valid_sync_status CHECK (sync_status IN ('pending', 'synced', 'error')),
    UNIQUE(product_id, sales_channel_id, COALESCE(variant_id, '00000000-0000-0000-0000-000000000000'::UUID))
);

-- Inventory allocation rules
CREATE TABLE inventory_allocation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Rule identification
    rule_name VARCHAR(255) NOT NULL,
    rule_priority INTEGER DEFAULT 1, -- Higher number = higher priority
    
    -- Rule conditions
    product_categories JSONB, -- Array of category IDs
    specific_products JSONB, -- Array of product IDs
    sales_channels JSONB, -- Array of channel IDs
    
    -- Allocation strategy
    allocation_method VARCHAR(30) NOT NULL, -- percentage, priority, demand_based, profitability
    allocation_percentage DECIMAL(5,2), -- For percentage method
    
    -- Rule parameters
    parameters JSONB, -- Method-specific parameters
    
    -- Rule timing
    effective_start_date DATE,
    effective_end_date DATE,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_allocation_method CHECK (allocation_method IN ('percentage', 'priority', 'demand_based', 'profitability'))
);

-- Cross-channel order management
CREATE TABLE omnichannel_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Order identification
    order_number VARCHAR(50) UNIQUE NOT NULL,
    external_order_id VARCHAR(100), -- Order ID from external channel
    
    -- Channel and customer
    sales_channel_id UUID NOT NULL REFERENCES sales_channels(id),
    customer_id UUID REFERENCES customers(id),
    
    -- Order type and fulfillment
    order_type VARCHAR(30) DEFAULT 'standard', -- standard, pickup, delivery, ship_to_store, return
    fulfillment_method VARCHAR(30) DEFAULT 'ship_to_customer', -- ship_to_customer, store_pickup, curbside_pickup, same_day_delivery
    
    -- Addresses
    billing_address JSONB,
    shipping_address JSONB,
    pickup_store_id UUID REFERENCES retail_stores(id),
    
    -- Order amounts
    subtotal DECIMAL(12,2) DEFAULT 0,
    shipping_amount DECIMAL(8,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0,
    
    -- Order status
    order_status VARCHAR(30) DEFAULT 'pending', -- pending, confirmed, processing, shipped, delivered, cancelled, returned
    payment_status VARCHAR(20) DEFAULT 'pending', -- pending, authorized, captured, failed, refunded
    fulfillment_status VARCHAR(30) DEFAULT 'pending', -- pending, allocated, picked, packed, shipped, delivered
    
    -- Timing
    order_date TIMESTAMPTZ DEFAULT NOW(),
    required_ship_date DATE,
    estimated_delivery_date DATE,
    actual_delivery_date DATE,
    
    -- Customer service
    customer_notes TEXT,
    internal_notes TEXT,
    gift_message TEXT,
    
    -- Promotions and discounts
    applied_promotions JSONB, -- Array of promotion codes applied
    loyalty_points_used INTEGER DEFAULT 0,
    loyalty_points_earned INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_order_type CHECK (order_type IN ('standard', 'pickup', 'delivery', 'ship_to_store', 'return')),
    CONSTRAINT valid_fulfillment_method CHECK (fulfillment_method IN ('ship_to_customer', 'store_pickup', 'curbside_pickup', 'same_day_delivery')),
    CONSTRAINT valid_order_status CHECK (order_status IN ('pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled', 'returned')),
    CONSTRAINT valid_payment_status CHECK (payment_status IN ('pending', 'authorized', 'captured', 'failed', 'refunded')),
    CONSTRAINT valid_fulfillment_status CHECK (fulfillment_status IN ('pending', 'allocated', 'picked', 'packed', 'shipped', 'delivered'))
);

-- Order line items
CREATE TABLE omnichannel_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    omnichannel_order_id UUID NOT NULL REFERENCES omnichannel_orders(id) ON DELETE CASCADE,
    
    -- Product information
    product_id UUID NOT NULL REFERENCES retail_products(id),
    variant_id UUID REFERENCES product_variants(id),
    sku VARCHAR(100) NOT NULL,
    
    -- Quantity and pricing
    quantity_ordered INTEGER NOT NULL,
    quantity_allocated INTEGER DEFAULT 0,
    quantity_shipped INTEGER DEFAULT 0,
    quantity_cancelled INTEGER DEFAULT 0,
    quantity_returned INTEGER DEFAULT 0,
    
    unit_price DECIMAL(10,4) NOT NULL,
    unit_cost DECIMAL(10,4),
    discount_amount DECIMAL(8,2) DEFAULT 0,
    line_total DECIMAL(12,2) NOT NULL,
    
    -- Fulfillment details
    fulfillment_store_id UUID REFERENCES retail_stores(id),
    warehouse_id UUID REFERENCES warehouses(id),
    
    -- Item status
    item_status VARCHAR(30) DEFAULT 'pending', -- pending, allocated, picked, packed, shipped, delivered, cancelled, returned
    
    -- Personalization
    personalization_details JSONB, -- Custom engraving, gift wrap, etc.
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_item_status CHECK (item_status IN ('pending', 'allocated', 'picked', 'packed', 'shipped', 'delivered', 'cancelled', 'returned'))
);
```

##  Customer Experience & Analytics

### Customer Segmentation & Personalization

```typescript
interface RetailCustomerAnalytics {
  segmentCustomers(tenantId: string): Promise<CustomerSegmentation[]>;
  analyzeShoppingBehavior(customerId: string): Promise<ShoppingBehaviorProfile>;
  generatePersonalizedRecommendations(customerId: string): Promise<ProductRecommendation[]>;
  calculateCustomerLifetimeValue(customerId: string): Promise<CustomerValueMetrics>;
}

interface CustomerSegmentation {
  segment_id: string;
  segment_name: string;
  segment_description: string;
  
  criteria: {
    purchase_frequency: 'high' | 'medium' | 'low';
    average_order_value: 'high' | 'medium' | 'low';
    recency: 'recent' | 'moderate' | 'dormant';
    total_spend: 'high' | 'medium' | 'low';
  };
  
  customer_count: number;
  segment_percentage: number;
  
  characteristics: {
    average_age: number;
    preferred_categories: string[];
    preferred_channels: string[];
    shopping_patterns: ShoppingPattern[];
  };
  
  marketing_recommendations: {
    campaign_type: string;
    messaging_strategy: string;
    channel_preference: string[];
    discount_sensitivity: number;
  };
}

interface ShoppingBehaviorProfile {
  customer_id: string;
  analysis_period: DateRange;
  
  purchase_patterns: {
    total_orders: number;
    total_spend: number;
    average_order_value: number;
    purchase_frequency_days: number;
    seasonal_patterns: SeasonalPattern[];
  };
  
  product_preferences: {
    favorite_categories: CategoryPreference[];
    favorite_brands: BrandPreference[];
    size_preferences: SizePreference[];
    color_preferences: ColorPreference[];
  };
  
  channel_behavior: {
    preferred_channels: ChannelUsage[];
    cross_channel_journey: CustomerJourney[];
    device_preferences: DeviceUsage[];
  };
  
  price_sensitivity: {
    discount_responsiveness: number;
    price_tolerance: number;
    promotion_engagement: number;
  };
  
  loyalty_indicators: {
    repeat_purchase_rate: number;
    brand_loyalty_score: number;
    referral_likelihood: number;
    churn_risk_score: number;
  };
}

class RetailAnalyticsService implements RetailCustomerAnalytics {
  async segmentCustomers(tenantId: string): Promise<CustomerSegmentation[]> {
    const customers = await this.getCustomersWithPurchaseHistory(tenantId);
    const segments: CustomerSegmentation[] = [];
    
    // RFM Analysis (Recency, Frequency, Monetary)
    const rfmScores = this.calculateRFMScores(customers);
    
    // Define segment criteria
    const segmentDefinitions = [
      {
        name: 'Champions',
        description: 'High-value customers who buy frequently and recently',
        criteria: { recency: 'high', frequency: 'high', monetary: 'high' }
      },
      {
        name: 'Loyal Customers',
        description: 'Regular customers with good spending habits',
        criteria: { recency: 'medium', frequency: 'high', monetary: 'medium' }
      },
      {
        name: 'Potential Loyalists',
        description: 'Recent customers with good potential',
        criteria: { recency: 'high', frequency: 'medium', monetary: 'medium' }
      },
      {
        name: 'At Risk',
        description: 'Previously good customers who haven\'t purchased recently',
        criteria: { recency: 'low', frequency: 'high', monetary: 'high' }
      },
      {
        name: 'Cannot Lose Them',
        description: 'High-value customers who haven\'t purchased recently',
        criteria: { recency: 'low', frequency: 'high', monetary: 'high' }
      },
      {
        name: 'Price Sensitive',
        description: 'Customers who primarily buy during promotions',
        criteria: { promotion_response: 'high', average_discount: 'high' }
      }
    ];
    
    for (const segmentDef of segmentDefinitions) {
      const segmentCustomers = this.identifySegmentCustomers(customers, rfmScores, segmentDef.criteria);
      
      const segment: CustomerSegmentation = {
        segment_id: this.generateSegmentId(segmentDef.name),
        segment_name: segmentDef.name,
        segment_description: segmentDef.description,
        criteria: this.mapCriteriaToInterface(segmentDef.criteria),
        customer_count: segmentCustomers.length,
        segment_percentage: (segmentCustomers.length / customers.length) * 100,
        characteristics: await this.analyzeSegmentCharacteristics(segmentCustomers),
        marketing_recommendations: this.generateMarketingRecommendations(segmentDef, segmentCustomers)
      };
      
      segments.push(segment);
    }
    
    return segments;
  }
  
  async generatePersonalizedRecommendations(customerId: string): Promise<ProductRecommendation[]> {
    const customer = await this.getCustomerWithHistory(customerId);
    const purchaseHistory = await this.getPurchaseHistory(customerId);
    const browsingHistory = await this.getBrowsingHistory(customerId);
    
    const recommendations: ProductRecommendation[] = [];
    
    // Collaborative filtering recommendations
    const collaborativeRecs = await this.getCollaborativeFilteringRecommendations(customer, purchaseHistory);
    recommendations.push(...collaborativeRecs);
    
    // Content-based recommendations
    const contentBasedRecs = await this.getContentBasedRecommendations(purchaseHistory, browsingHistory);
    recommendations.push(...contentBasedRecs);
    
    // Cross-sell recommendations
    const crossSellRecs = await this.getCrossSellRecommendations(purchaseHistory);
    recommendations.push(...crossSellRecs);
    
    // Trending products in customer's preferred categories
    const trendingRecs = await this.getTrendingRecommendations(customer.preferredCategories);
    recommendations.push(...trendingRecs);
    
    // Score and rank all recommendations
    const scoredRecommendations = this.scoreRecommendations(recommendations, customer);
    
    // Return top 20 recommendations
    return scoredRecommendations
      .sort((a, b) => b.confidence_score - a.confidence_score)
      .slice(0, 20);
  }
  
  private calculateRFMScores(customers: CustomerWithHistory[]): Map<string, RFMScore> {
    const rfmScores = new Map<string, RFMScore>();
    const currentDate = new Date();
    
    // Calculate raw RFM values
    const rfmValues = customers.map(customer => {
      const lastPurchaseDate = new Date(customer.last_purchase_date);
      const recency = Math.floor((currentDate.getTime() - lastPurchaseDate.getTime()) / (1000 * 60 * 60 * 24));
      const frequency = customer.total_orders || 0;
      const monetary = customer.total_spent || 0;
      
      return {
        customer_id: customer.id,
        recency,
        frequency,
        monetary
      };
    });
    
    // Calculate quintiles for scoring
    const recencyQuintiles = this.calculateQuintiles(rfmValues.map(v => v.recency));
    const frequencyQuintiles = this.calculateQuintiles(rfmValues.map(v => v.frequency));
    const monetaryQuintiles = this.calculateQuintiles(rfmValues.map(v => v.monetary));
    
    // Assign scores (1-5, where 5 is best)
    for (const value of rfmValues) {
      const recencyScore = 6 - this.getQuintileScore(value.recency, recencyQuintiles); // Invert for recency
      const frequencyScore = this.getQuintileScore(value.frequency, frequencyQuintiles);
      const monetaryScore = this.getQuintileScore(value.monetary, monetaryQuintiles);
      
      rfmScores.set(value.customer_id, {
        recency_score: recencyScore,
        frequency_score: frequencyScore,
        monetary_score: monetaryScore,
        rfm_score: recencyScore + frequencyScore + monetaryScore
      });
    }
    
    return rfmScores;
  }
}
```

### Visual Merchandising & Store Layout

```sql
-- Store layout and zones
CREATE TABLE store_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    retail_store_id UUID NOT NULL REFERENCES retail_stores(id) ON DELETE CASCADE,
    
    -- Zone identification
    zone_name VARCHAR(255) NOT NULL,
    zone_code VARCHAR(20) NOT NULL,
    zone_type VARCHAR(50) NOT NULL, -- entrance, checkout, fitting_room, seasonal, promotional, regular
    
    -- Zone properties
    zone_area_sqft DECIMAL(8,2),
    foot_traffic_level VARCHAR(20) DEFAULT 'medium', -- high, medium, low
    visibility_level VARCHAR(20) DEFAULT 'medium', -- high, medium, low
    
    -- Zone layout
    position_coordinates JSONB, -- Store layout coordinates
    adjacent_zones JSONB, -- Array of adjacent zone IDs
    
    -- Merchandising rules
    category_assignment JSONB, -- Array of category IDs for this zone
    display_style VARCHAR(50), -- wall_display, island, endcap, checkout_display
    
    -- Performance metrics
    sales_per_sqft DECIMAL(10,2),
    conversion_rate DECIMAL(5,4),
    dwell_time_seconds INTEGER,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_zone_type CHECK (zone_type IN ('entrance', 'checkout', 'fitting_room', 'seasonal', 'promotional', 'regular')),
    CONSTRAINT valid_foot_traffic_level CHECK (foot_traffic_level IN ('high', 'medium', 'low')),
    CONSTRAINT valid_visibility_level CHECK (visibility_level IN ('high', 'medium', 'low')),
    UNIQUE(retail_store_id, zone_code)
);

-- Planograms for product placement
CREATE TABLE planograms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Planogram identification
    planogram_name VARCHAR(255) NOT NULL,
    planogram_code VARCHAR(50) UNIQUE NOT NULL,
    version INTEGER DEFAULT 1,
    
    -- Scope and application
    store_id UUID REFERENCES retail_stores(id), -- Specific store or template for all stores
    zone_id UUID REFERENCES store_zones(id),
    fixture_type VARCHAR(50), -- gondola, wall_bay, endcap, island, counter
    
    -- Planogram dimensions
    width_cm DECIMAL(8,2),
    height_cm DECIMAL(8,2),
    depth_cm DECIMAL(8,2),
    shelf_count INTEGER,
    
    -- Effective dates
    effective_start_date DATE NOT NULL,
    effective_end_date DATE,
    
    -- Planogram objectives
    objective VARCHAR(50) DEFAULT 'sales_optimization', -- sales_optimization, inventory_turnover, margin_optimization, new_product_launch
    target_metrics JSONB, -- Target sales, margins, etc.
    
    -- Approval and status
    status VARCHAR(20) DEFAULT 'draft', -- draft, approved, active, archived
    approved_by UUID REFERENCES employees(id),
    approved_at TIMESTAMPTZ,
    
    -- Implementation tracking
    implemented_stores INTEGER DEFAULT 0,
    total_target_stores INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_objective CHECK (objective IN ('sales_optimization', 'inventory_turnover', 'margin_optimization', 'new_product_launch')),
    CONSTRAINT valid_status CHECK (status IN ('draft', 'approved', 'active', 'archived'))
);

-- Planogram product placements
CREATE TABLE planogram_placements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    planogram_id UUID NOT NULL REFERENCES planograms(id) ON DELETE CASCADE,
    
    -- Product placement
    product_id UUID NOT NULL REFERENCES retail_products(id),
    variant_id UUID REFERENCES product_variants(id),
    
    -- Position details
    shelf_number INTEGER NOT NULL,
    position_x DECIMAL(6,2) NOT NULL, -- X coordinate on shelf
    position_y DECIMAL(6,2), -- Y coordinate (for vertical placement)
    width_cm DECIMAL(6,2) NOT NULL,
    depth_cm DECIMAL(6,2),
    height_cm DECIMAL(6,2),
    
    -- Inventory levels
    minimum_stock INTEGER DEFAULT 1,
    maximum_stock INTEGER NOT NULL,
    restock_trigger INTEGER,
    
    -- Placement strategy
    placement_reason VARCHAR(100), -- high_margin, fast_moving, new_launch, promotional
    facing_count INTEGER DEFAULT 1, -- Number of product facings
    
    -- Performance tracking
    planned_weekly_sales INTEGER,
    actual_weekly_sales INTEGER DEFAULT 0,
    inventory_turns_per_week DECIMAL(4,2),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(planogram_id, shelf_number, position_x)
);

-- Promotional displays and campaigns
CREATE TABLE promotional_displays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Display identification
    display_name VARCHAR(255) NOT NULL,
    campaign_id UUID REFERENCES marketing_campaigns(id),
    
    -- Display location
    store_id UUID REFERENCES retail_stores(id),
    zone_id UUID REFERENCES store_zones(id),
    display_type VARCHAR(50) NOT NULL, -- endcap, island, window, entrance, checkout
    
    -- Display period
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    -- Display products
    featured_products JSONB, -- Array of product IDs
    promotional_pricing JSONB, -- Special pricing for display
    
    -- Display materials
    signage_required BOOLEAN DEFAULT true,
    signage_content TEXT,
    display_fixtures JSONB, -- Required fixtures and props
    
    -- Performance tracking
    target_sales DECIMAL(12,2),
    actual_sales DECIMAL(12,2) DEFAULT 0,
    foot_traffic_increase DECIMAL(5,2), -- Percentage increase
    
    -- Setup and maintenance
    setup_date DATE,
    setup_staff_id UUID REFERENCES employees(id),
    maintenance_schedule JSONB, -- Restocking and refresh schedule
    
    -- Status
    display_status VARCHAR(20) DEFAULT 'planned', -- planned, active, completed, cancelled
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_display_type CHECK (display_type IN ('endcap', 'island', 'window', 'entrance', 'checkout')),
    CONSTRAINT valid_display_status CHECK (display_status IN ('planned', 'active', 'completed', 'cancelled'))
);
```

##  E-commerce Integration

### Multi-Platform Synchronization

```typescript
interface EcommerceIntegration {
  syncProductCatalog(channelId: string): Promise<SyncResult>;
  syncInventoryLevels(channelId: string): Promise<SyncResult>;
  syncOrdersFromChannel(channelId: string): Promise<OrderSyncResult>;
  syncCustomerData(channelId: string): Promise<CustomerSyncResult>;
  managePricingRules(channelId: string, rules: PricingRule[]): Promise<void>;
}

interface SyncResult {
  channel_id: string;
  sync_type: string;
  sync_timestamp: Date;
  
  summary: {
    total_records: number;
    successful_syncs: number;
    failed_syncs: number;
    skipped_records: number;
  };
  
  details: {
    created_records: number;
    updated_records: number;
    deleted_records: number;
  };
  
  errors: SyncError[];
  warnings: SyncWarning[];
}

interface PricingRule {
  rule_id: string;
  rule_name: string;
  rule_type: 'markup' | 'discount' | 'fixed' | 'competitive';
  
  conditions: {
    product_categories?: string[];
    brands?: string[];
    price_range?: { min: number; max: number; };
    inventory_level?: { min: number; max: number; };
  };
  
  pricing_action: {
    action_type: 'percentage' | 'fixed_amount' | 'formula';
    value: number;
    formula?: string;
  };
  
  channel_specific: {
    channel_id: string;
    minimum_price?: number;
    maximum_price?: number;
    round_to?: number; // Round to nearest value
  };
  
  schedule: {
    effective_start: Date;
    effective_end?: Date;
    days_of_week?: number[];
    time_ranges?: TimeRange[];
  };
}

class EcommerceService implements EcommerceIntegration {
  async syncProductCatalog(channelId: string): Promise<SyncResult> {
    const channel = await this.getSalesChannel(channelId);
    const products = await this.getProductsForChannel(channelId);
    
    const syncResult: SyncResult = {
      channel_id: channelId,
      sync_type: 'product_catalog',
      sync_timestamp: new Date(),
      summary: { total_records: 0, successful_syncs: 0, failed_syncs: 0, skipped_records: 0 },
      details: { created_records: 0, updated_records: 0, deleted_records: 0 },
      errors: [],
      warnings: []
    };
    
    syncResult.summary.total_records = products.length;
    
    for (const product of products) {
      try {
        // Apply channel-specific transformations
        const channelProduct = await this.transformProductForChannel(product, channel);
        
        // Check if product exists on channel
        const existingProduct = await this.getProductFromChannel(channel, product.sku);
        
        if (existingProduct) {
          // Update existing product
          const updateResult = await this.updateProductOnChannel(channel, channelProduct);
          if (updateResult.success) {
            syncResult.details.updated_records++;
            syncResult.summary.successful_syncs++;
          } else {
            syncResult.summary.failed_syncs++;
            syncResult.errors.push({
              record_id: product.id,
              error_type: 'update_failed',
              error_message: updateResult.error_message
            });
          }
        } else {
          // Create new product
          const createResult = await this.createProductOnChannel(channel, channelProduct);
          if (createResult.success) {
            syncResult.details.created_records++;
            syncResult.summary.successful_syncs++;
            
            // Store channel-specific product ID for future syncs
            await this.storeChannelProductMapping(product.id, channelId, createResult.channel_product_id);
          } else {
            syncResult.summary.failed_syncs++;
            syncResult.errors.push({
              record_id: product.id,
              error_type: 'create_failed',
              error_message: createResult.error_message
            });
          }
        }
      } catch (error) {
        syncResult.summary.failed_syncs++;
        syncResult.errors.push({
          record_id: product.id,
          error_type: 'sync_exception',
          error_message: error.message
        });
      }
    }
    
    // Log sync result
    await this.logSyncResult(syncResult);
    
    return syncResult;
  }
  
  async managePricingRules(channelId: string, rules: PricingRule[]): Promise<void> {
    const channel = await this.getSalesChannel(channelId);
    const products = await this.getProductsForChannel(channelId);
    
    for (const product of products) {
      // Find applicable pricing rules
      const applicableRules = rules.filter(rule => 
        this.isRuleApplicableToProduct(rule, product) &&
        this.isRuleActiveForChannel(rule, channelId)
      );
      
      if (applicableRules.length === 0) continue;
      
      // Apply rules in priority order
      const sortedRules = applicableRules.sort((a, b) => a.priority - b.priority);
      let finalPrice = product.retail_price;
      
      for (const rule of sortedRules) {
        finalPrice = this.applyPricingRule(finalPrice, rule, product);
      }
      
      // Apply channel-specific constraints
      const channelPrice = this.applyChannelConstraints(finalPrice, rule.channel_specific);
      
      // Update channel inventory with new price
      await this.updateChannelPrice(channelId, product.id, channelPrice);
      
      // Sync price to external channel if needed
      if (channel.supports_pricing_sync) {
        await this.syncPriceToChannel(channel, product, channelPrice);
      }
    }
  }
  
  private async transformProductForChannel(product: RetailProduct, channel: SalesChannel): Promise<ChannelProduct> {
    const transformation = await this.getChannelTransformationRules(channel.id);
    
    let channelProduct: ChannelProduct = {
      sku: product.sku,
      title: product.product_name,
      description: product.description,
      price: product.retail_price,
      inventory_quantity: await this.getAvailableInventory(product.id, channel.id),
      images: product.image_gallery || [],
      category: await this.mapCategoryForChannel(product.category_id, channel.id),
      attributes: this.extractProductAttributes(product),
      seo: {
        title: product.seo_title || product.product_name,
        description: product.seo_description || product.short_description,
        keywords: product.meta_keywords || []
      }
    };
    
    // Apply channel-specific transformations
    if (transformation) {
      channelProduct = this.applyTransformationRules(channelProduct, transformation);
    }
    
    // Apply channel-specific pricing rules
    const pricingRules = await this.getPricingRulesForChannel(channel.id);
    channelProduct.price = this.calculateChannelPrice(product, pricingRules);
    
    return channelProduct;
  }
  
  private applyPricingRule(currentPrice: number, rule: PricingRule, product: RetailProduct): number {
    let newPrice = currentPrice;
    
    switch (rule.pricing_action.action_type) {
      case 'percentage':
        if (rule.rule_type === 'markup') {
          newPrice = currentPrice * (1 + rule.pricing_action.value / 100);
        } else if (rule.rule_type === 'discount') {
          newPrice = currentPrice * (1 - rule.pricing_action.value / 100);
        }
        break;
        
      case 'fixed_amount':
        if (rule.rule_type === 'markup') {
          newPrice = currentPrice + rule.pricing_action.value;
        } else if (rule.rule_type === 'discount') {
          newPrice = currentPrice - rule.pricing_action.value;
        }
        break;
        
      case 'fixed':
        newPrice = rule.pricing_action.value;
        break;
        
      case 'formula':
        newPrice = this.evaluatePricingFormula(rule.pricing_action.formula!, {
          current_price: currentPrice,
          cost_price: product.cost_price,
          retail_price: product.retail_price,
          inventory_level: this.getInventoryLevel(product.id)
        });
        break;
    }
    
    return Math.max(0, newPrice); // Ensure price is not negative
  }
}
```

This  retail management system provides sophisticated omnichannel operations, customer analytics, visual merchandising, and e-commerce integration capabilities specifically designed for modern retail operations while integrating seamlessly with the core ERP inventory, financial, and customer management modules.
