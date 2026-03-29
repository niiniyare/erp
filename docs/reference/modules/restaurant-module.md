# Restaurant Management

## ️ Overview

The Restaurant Management module transforms the core ERP into a  restaurant operations platform. It provides menu management, kitchen operations, table service, inventory control, staff scheduling, and customer relationship management specifically designed for restaurants, cafes, and food service establishments.

##  Menu & Recipe Management

### Menu Engineering

```sql
-- Menu categories for organizing items
CREATE TABLE menu_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Category identification
    category_code VARCHAR(20) UNIQUE NOT NULL,
    category_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Category hierarchy
    parent_category_id UUID REFERENCES menu_categories(id),
    category_level INTEGER DEFAULT 1,
    sort_order INTEGER DEFAULT 0,
    
    -- Category properties
    is_active BOOLEAN DEFAULT true,
    available_for_online_ordering BOOLEAN DEFAULT true,
    requires_age_verification BOOLEAN DEFAULT false, -- For alcoholic beverages
    
    -- Display settings
    display_color VARCHAR(7), -- Hex color code
    category_image_url VARCHAR(500),
    
    -- Operational settings
    kitchen_display_name VARCHAR(100), -- Name shown in kitchen
    preparation_area VARCHAR(100), -- Kitchen section responsible
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Menu items
CREATE TABLE menu_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Item identification
    item_code VARCHAR(50) UNIQUE NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    
    -- Item classification
    menu_category_id UUID NOT NULL REFERENCES menu_categories(id),
    item_type VARCHAR(50) DEFAULT 'food', -- food, beverage, alcohol, dessert, appetizer
    cuisine_type VARCHAR(100), -- italian, chinese, mexican, etc.
    
    -- Pricing
    base_price DECIMAL(8,2) NOT NULL,
    cost_per_serving DECIMAL(8,2),
    profit_margin_percentage DECIMAL(5,2),
    
    -- Portion information
    serving_size VARCHAR(100),
    portion_weight_grams DECIMAL(8,2),
    calories_per_serving INTEGER,
    
    -- Preparation details
    preparation_time_minutes INTEGER,
    cooking_method VARCHAR(100), -- grilled, fried, baked, steamed, etc.
    skill_level_required VARCHAR(20) DEFAULT 'standard', -- basic, standard, advanced, expert
    
    -- Dietary information
    vegetarian BOOLEAN DEFAULT false,
    vegan BOOLEAN DEFAULT false,
    gluten_free BOOLEAN DEFAULT false,
    dairy_free BOOLEAN DEFAULT false,
    nut_free BOOLEAN DEFAULT false,
    halal BOOLEAN DEFAULT false,
    kosher BOOLEAN DEFAULT false,
    
    -- Allergen information
    contains_allergens JSONB, -- Array of allergen types
    allergen_warnings TEXT,
    
    -- Nutritional information
    nutritional_info JSONB, -- Detailed nutritional breakdown
    
    -- Availability and timing
    is_available BOOLEAN DEFAULT true,
    seasonal_item BOOLEAN DEFAULT false,
    available_days JSONB, -- Array of day numbers (1=Monday, 7=Sunday)
    available_time_start TIME,
    available_time_end TIME,
    
    -- Kitchen operations
    kitchen_station VARCHAR(100), -- grill, fryer, cold_prep, etc.
    special_instructions TEXT,
    holds_well BOOLEAN DEFAULT true, -- Can be kept warm after preparation
    
    -- Sales and popularity
    popularity_score DECIMAL(4,2) DEFAULT 0, -- 0-100 score
    featured_item BOOLEAN DEFAULT false,
    chef_recommendation BOOLEAN DEFAULT false,
    
    -- Digital assets
    primary_image_url VARCHAR(500),
    gallery_images JSONB, -- Array of image URLs
    
    -- Status
    item_status VARCHAR(20) DEFAULT 'active', -- active, inactive, discontinued, out_of_stock
    
    -- Recipe reference
    recipe_id UUID REFERENCES recipes(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_item_type CHECK (item_type IN ('food', 'beverage', 'alcohol', 'dessert', 'appetizer')),
    CONSTRAINT valid_skill_level CHECK (skill_level_required IN ('basic', 'standard', 'advanced', 'expert')),
    CONSTRAINT valid_item_status CHECK (item_status IN ('active', 'inactive', 'discontinued', 'out_of_stock'))
);

-- Recipes for menu items
CREATE TABLE recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Recipe identification
    recipe_name VARCHAR(255) NOT NULL,
    recipe_code VARCHAR(50) UNIQUE,
    recipe_version INTEGER DEFAULT 1,
    
    -- Recipe details
    description TEXT,
    yield_servings INTEGER NOT NULL DEFAULT 1,
    total_prep_time_minutes INTEGER,
    total_cook_time_minutes INTEGER,
    
    -- Recipe classification
    difficulty_level VARCHAR(20) DEFAULT 'medium', -- easy, medium, hard, expert
    recipe_type VARCHAR(50), -- main_dish, side_dish, sauce, marinade, etc.
    
    -- Cost calculation
    total_ingredient_cost DECIMAL(10,4),
    cost_per_serving DECIMAL(8,4),
    
    -- Instructions
    preparation_steps JSONB, -- Array of step objects with instructions
    cooking_instructions TEXT,
    plating_instructions TEXT,
    
    -- Quality standards
    quality_checkpoints JSONB, -- Array of quality check descriptions
    temperature_requirements JSONB, -- Cooking and serving temperatures
    
    -- Recipe status
    recipe_status VARCHAR(20) DEFAULT 'active', -- active, testing, archived, discontinued
    approved_by_chef BOOLEAN DEFAULT false,
    approved_date DATE,
    
    -- Recipe metadata
    created_by_chef_id UUID REFERENCES employees(id),
    last_modified_by UUID REFERENCES employees(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_difficulty_level CHECK (difficulty_level IN ('easy', 'medium', 'hard', 'expert')),
    CONSTRAINT valid_recipe_status CHECK (recipe_status IN ('active', 'testing', 'archived', 'discontinued'))
);

-- Recipe ingredients
CREATE TABLE recipe_ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    
    -- Ingredient reference
    ingredient_id UUID NOT NULL REFERENCES inventory_items(id),
    
    -- Quantity and measurement
    quantity DECIMAL(10,4) NOT NULL,
    unit_of_measure VARCHAR(50) NOT NULL,
    
    -- Ingredient preparation
    preparation_method VARCHAR(100), -- diced, sliced, minced, whole, etc.
    preparation_notes TEXT,
    
    -- Ingredient properties
    is_primary_ingredient BOOLEAN DEFAULT true, -- Primary vs garnish/seasoning
    is_optional BOOLEAN DEFAULT false,
    substitutable BOOLEAN DEFAULT false,
    substitute_ingredients JSONB, -- Array of alternative ingredient IDs
    
    -- Cost and nutrition contribution
    ingredient_cost DECIMAL(8,4),
    nutritional_contribution JSONB, -- Calories, protein, etc. from this ingredient
    
    -- Sequence and timing
    preparation_sequence INTEGER DEFAULT 1,
    add_at_step INTEGER, -- Which preparation step to add ingredient
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(recipe_id, ingredient_id)
);
```

### Menu Variants and Customization

```sql
-- Menu item variants (sizes, preparations, etc.)
CREATE TABLE menu_item_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    
    -- Variant identification
    variant_name VARCHAR(255) NOT NULL,
    variant_code VARCHAR(50),
    
    -- Variant type
    variant_type VARCHAR(50) NOT NULL, -- size, preparation, spice_level, temperature
    
    -- Pricing adjustment
    price_adjustment_type VARCHAR(20) DEFAULT 'fixed', -- fixed, percentage
    price_adjustment DECIMAL(8,2) DEFAULT 0,
    
    -- Recipe adjustment
    recipe_multiplier DECIMAL(6,4) DEFAULT 1.0, -- For size variations
    preparation_notes TEXT,
    
    -- Availability
    is_available BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    
    -- Display order
    sort_order INTEGER DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_variant_type CHECK (variant_type IN ('size', 'preparation', 'spice_level', 'temperature', 'style')),
    CONSTRAINT valid_price_adjustment_type CHECK (price_adjustment_type IN ('fixed', 'percentage'))
);

-- Menu item customizations (add-ons, modifications)
CREATE TABLE menu_item_customizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    
    -- Customization details
    customization_name VARCHAR(255) NOT NULL,
    customization_type VARCHAR(50) NOT NULL, -- add_on, substitution, removal, modification
    
    -- Ingredient reference (for add-ons/substitutions)
    ingredient_id UUID REFERENCES inventory_items(id),
    additional_quantity DECIMAL(8,4),
    
    -- Pricing
    additional_cost DECIMAL(6,2) DEFAULT 0,
    
    -- Availability and restrictions
    is_available BOOLEAN DEFAULT true,
    maximum_quantity INTEGER DEFAULT 1,
    requires_chef_approval BOOLEAN DEFAULT false,
    
    -- Impact on preparation
    affects_prep_time_minutes INTEGER DEFAULT 0,
    special_preparation_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_customization_type CHECK (customization_type IN ('add_on', 'substitution', 'removal', 'modification'))
);

-- Menu availability schedules
CREATE TABLE menu_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Schedule identification
    schedule_name VARCHAR(255) NOT NULL,
    schedule_type VARCHAR(50) NOT NULL, -- breakfast, lunch, dinner, happy_hour, late_night
    
    -- Time periods
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    days_of_week JSONB NOT NULL, -- Array of day numbers
    
    -- Date restrictions
    effective_start_date DATE,
    effective_end_date DATE,
    exclude_holidays BOOLEAN DEFAULT false,
    
    -- Menu items included
    included_categories JSONB, -- Array of category IDs
    excluded_items JSONB, -- Array of specific item IDs to exclude
    
    -- Pricing adjustments
    schedule_price_multiplier DECIMAL(4,3) DEFAULT 1.0, -- Happy hour discounts, etc.
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_schedule_type CHECK (schedule_type IN ('breakfast', 'lunch', 'dinner', 'happy_hour', 'late_night', 'all_day'))
);
```

##  Restaurant Operations

### Table Management

```sql
-- Restaurant areas and sections
CREATE TABLE restaurant_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Area identification
    area_name VARCHAR(255) NOT NULL,
    area_code VARCHAR(20) UNIQUE NOT NULL,
    area_type VARCHAR(50) DEFAULT 'dining', -- dining, bar, patio, private, takeout
    
    -- Area properties
    total_tables INTEGER DEFAULT 0,
    total_capacity INTEGER DEFAULT 0,
    smoking_allowed BOOLEAN DEFAULT false,
    
    -- Service characteristics
    service_style VARCHAR(50) DEFAULT 'table_service', -- table_service, counter_service, buffet, self_service
    requires_reservation BOOLEAN DEFAULT false,
    accepts_walk_ins BOOLEAN DEFAULT true,
    
    -- Operational hours
    operating_hours JSONB, -- Daily operating schedule
    
    -- Staff assignment
    manager_id UUID REFERENCES employees(id),
    host_station_required BOOLEAN DEFAULT true,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_area_type CHECK (area_type IN ('dining', 'bar', 'patio', 'private', 'takeout')),
    CONSTRAINT valid_service_style CHECK (service_style IN ('table_service', 'counter_service', 'buffet', 'self_service'))
);

-- Restaurant tables
CREATE TABLE restaurant_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Table identification
    table_number VARCHAR(20) NOT NULL,
    table_name VARCHAR(100),
    restaurant_area_id UUID NOT NULL REFERENCES restaurant_areas(id),
    
    -- Table properties
    base_capacity INTEGER NOT NULL DEFAULT 2,
    maximum_capacity INTEGER NOT NULL DEFAULT 4,
    minimum_capacity INTEGER DEFAULT 1,
    table_shape VARCHAR(20) DEFAULT 'square', -- square, round, rectangular, booth
    
    -- Table features
    has_booth BOOLEAN DEFAULT false,
    window_table BOOLEAN DEFAULT false,
    private_table BOOLEAN DEFAULT false,
    high_top_table BOOLEAN DEFAULT false,
    wheelchair_accessible BOOLEAN DEFAULT true,
    
    -- Service requirements
    requires_server BOOLEAN DEFAULT true,
    vip_table BOOLEAN DEFAULT false,
    smoking_table BOOLEAN DEFAULT false,
    
    -- Table position
    position_x DECIMAL(6,2), -- X coordinate for floor plan
    position_y DECIMAL(6,2), -- Y coordinate for floor plan
    
    -- Status
    table_status VARCHAR(20) DEFAULT 'available', -- available, occupied, reserved, cleaning, out_of_order
    is_active BOOLEAN DEFAULT true,
    
    -- Current service
    current_order_id UUID REFERENCES orders(id),
    current_server_id UUID REFERENCES employees(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_table_shape CHECK (table_shape IN ('square', 'round', 'rectangular', 'booth')),
    CONSTRAINT valid_table_status CHECK (table_status IN ('available', 'occupied', 'reserved', 'cleaning', 'out_of_order')),
    UNIQUE(tenant_id, table_number)
);

-- Table reservations
CREATE TABLE table_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Reservation identification
    reservation_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Customer information
    customer_id UUID REFERENCES customers(id),
    customer_name VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(20),
    customer_email VARCHAR(255),
    
    -- Reservation details
    reservation_date DATE NOT NULL,
    reservation_time TIME NOT NULL,
    party_size INTEGER NOT NULL,
    
    -- Table assignment
    preferred_table_id UUID REFERENCES restaurant_tables(id),
    assigned_table_id UUID REFERENCES restaurant_tables(id),
    restaurant_area_id UUID REFERENCES restaurant_areas(id),
    
    -- Special requests
    special_requests TEXT,
    dietary_restrictions JSONB,
    occasion VARCHAR(100), -- birthday, anniversary, business_meeting, etc.
    
    -- Reservation status
    reservation_status VARCHAR(20) DEFAULT 'confirmed', -- confirmed, pending, seated, no_show, cancelled
    
    -- Timing
    estimated_duration_minutes INTEGER DEFAULT 90,
    arrival_time TIMESTAMPTZ,
    seated_time TIMESTAMPTZ,
    departure_time TIMESTAMPTZ,
    
    -- Service assignment
    server_id UUID REFERENCES employees(id),
    host_id UUID REFERENCES employees(id),
    
    -- Notes
    reservation_notes TEXT,
    internal_notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_reservation_status CHECK (reservation_status IN ('confirmed', 'pending', 'seated', 'no_show', 'cancelled'))
);
```

### Order Management & Kitchen Operations

```sql
-- Restaurant orders
CREATE TABLE restaurant_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Order identification
    order_number VARCHAR(50) UNIQUE NOT NULL,
    order_type VARCHAR(30) DEFAULT 'dine_in', -- dine_in, takeout, delivery, catering
    
    -- Table and customer information
    table_id UUID REFERENCES restaurant_tables(id),
    customer_id UUID REFERENCES customers(id),
    party_size INTEGER DEFAULT 1,
    
    -- Service staff
    server_id UUID NOT NULL REFERENCES employees(id),
    host_id UUID REFERENCES employees(id),
    
    -- Order timing
    order_started_at TIMESTAMPTZ DEFAULT NOW(),
    order_sent_to_kitchen_at TIMESTAMPTZ,
    order_ready_at TIMESTAMPTZ,
    order_served_at TIMESTAMPTZ,
    order_completed_at TIMESTAMPTZ,
    
    -- Order totals
    subtotal DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(8,2) DEFAULT 0,
    service_charge DECIMAL(8,2) DEFAULT 0,
    tip_amount DECIMAL(8,2) DEFAULT 0,
    discount_amount DECIMAL(8,2) DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0,
    
    -- Order status
    order_status VARCHAR(20) DEFAULT 'open', -- open, sent_to_kitchen, preparing, ready, served, paid, cancelled
    kitchen_status VARCHAR(20) DEFAULT 'pending', -- pending, received, preparing, ready, served
    
    -- Payment information
    payment_method VARCHAR(30),
    payment_status VARCHAR(20) DEFAULT 'pending', -- pending, paid, partial, failed
    payment_reference VARCHAR(100),
    
    -- Special instructions
    order_notes TEXT,
    kitchen_notes TEXT,
    allergy_notes TEXT,
    
    -- Delivery information (for delivery orders)
    delivery_address JSONB,
    delivery_instructions TEXT,
    estimated_delivery_time TIMESTAMPTZ,
    delivery_fee DECIMAL(6,2),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_order_type CHECK (order_type IN ('dine_in', 'takeout', 'delivery', 'catering')),
    CONSTRAINT valid_order_status CHECK (order_status IN ('open', 'sent_to_kitchen', 'preparing', 'ready', 'served', 'paid', 'cancelled')),
    CONSTRAINT valid_kitchen_status CHECK (kitchen_status IN ('pending', 'received', 'preparing', 'ready', 'served')),
    CONSTRAINT valid_payment_status CHECK (payment_status IN ('pending', 'paid', 'partial', 'failed'))
);

-- Order line items
CREATE TABLE restaurant_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_order_id UUID NOT NULL REFERENCES restaurant_orders(id) ON DELETE CASCADE,
    
    -- Item details
    menu_item_id UUID NOT NULL REFERENCES menu_items(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    
    -- Variant and customizations
    selected_variant_id UUID REFERENCES menu_item_variants(id),
    customizations JSONB, -- Array of customization IDs and specifications
    
    -- Pricing
    unit_price DECIMAL(8,2) NOT NULL,
    customization_charges DECIMAL(6,2) DEFAULT 0,
    line_total DECIMAL(8,2) NOT NULL,
    
    -- Kitchen information
    kitchen_station VARCHAR(100), -- Where this item should be prepared
    preparation_priority INTEGER DEFAULT 1, -- 1=highest priority
    cooking_instructions TEXT,
    
    -- Item status
    item_status VARCHAR(20) DEFAULT 'ordered', -- ordered, preparing, ready, served, cancelled
    
    -- Timing
    sent_to_kitchen_at TIMESTAMPTZ,
    preparation_started_at TIMESTAMPTZ,
    preparation_completed_at TIMESTAMPTZ,
    served_at TIMESTAMPTZ,
    
    -- Quality and satisfaction
    special_requests TEXT,
    customer_satisfaction_rating INTEGER, -- 1-5 scale
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_item_status CHECK (item_status IN ('ordered', 'preparing', 'ready', 'served', 'cancelled')),
    CONSTRAINT valid_satisfaction_rating CHECK (customer_satisfaction_rating IS NULL OR (customer_satisfaction_rating >= 1 AND customer_satisfaction_rating <= 5))
);

-- Kitchen display system tickets
CREATE TABLE kitchen_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_order_id UUID NOT NULL REFERENCES restaurant_orders(id) ON DELETE CASCADE,
    
    -- Ticket identification
    ticket_number INTEGER NOT NULL,
    kitchen_station VARCHAR(100) NOT NULL,
    
    -- Ticket timing
    received_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    estimated_ready_time TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    
    -- Ticket priority
    priority_level INTEGER DEFAULT 1, -- 1=highest, 5=lowest
    rush_order BOOLEAN DEFAULT false,
    
    -- Items on this ticket
    ticket_items JSONB, -- Array of order item IDs assigned to this station
    
    -- Ticket status
    ticket_status VARCHAR(20) DEFAULT 'pending', -- pending, in_progress, ready, completed, cancelled
    
    -- Kitchen staff
    assigned_chef_id UUID REFERENCES employees(id),
    expediter_id UUID REFERENCES employees(id),
    
    -- Quality control
    quality_checked BOOLEAN DEFAULT false,
    quality_notes TEXT,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_ticket_status CHECK (ticket_status IN ('pending', 'in_progress', 'ready', 'completed', 'cancelled'))
);
```

##  Restaurant Analytics & Performance

### Sales Analytics

```typescript
interface RestaurantAnalytics {
  generateSalesReport(locationId: string, period: DateRange): Promise<SalesReport>;
  analyzeMenuPerformance(period: DateRange): Promise<MenuPerformanceReport>;
  calculateTableTurnover(areaId: string, period: DateRange): Promise<TableTurnoverReport>;
  analyzeCustomerSatisfaction(period: DateRange): Promise<CustomerSatisfactionReport>;
}

interface SalesReport {
  location_id: string;
  reporting_period: DateRange;
  
  revenue_summary: {
    total_revenue: number;
    food_revenue: number;
    beverage_revenue: number;
    alcohol_revenue: number;
    average_check_size: number;
    revenue_per_seat: number;
  };
  
  transaction_summary: {
    total_transactions: number;
    dine_in_transactions: number;
    takeout_transactions: number;
    delivery_transactions: number;
    average_party_size: number;
  };
  
  hourly_breakdown: HourlySales[];
  daily_breakdown: DailySales[];
  
  payment_methods: {
    cash_percentage: number;
    card_percentage: number;
    digital_percentage: number;
    average_tip_percentage: number;
  };
  
  operational_metrics: {
    table_turnover_rate: number;
    average_dining_duration: number;
    kitchen_ticket_time_average: number;
    customer_wait_time_average: number;
  };
}

interface MenuPerformanceReport {
  reporting_period: DateRange;
  
  top_performers: {
    highest_revenue_items: MenuItemPerformance[];
    highest_quantity_items: MenuItemPerformance[];
    highest_margin_items: MenuItemPerformance[];
    trending_items: MenuItemPerformance[];
  };
  
  category_analysis: {
    category_id: string;
    category_name: string;
    total_revenue: number;
    total_quantity_sold: number;
    average_price: number;
    profit_margin: number;
    contribution_percentage: number;
  }[];
  
  underperformers: {
    low_sales_items: MenuItemPerformance[];
    low_margin_items: MenuItemPerformance[];
    slow_moving_items: MenuItemPerformance[];
  };
  
  pricing_analysis: {
    price_elasticity_items: MenuItemElasticity[];
    optimal_pricing_suggestions: PricingSuggestion[];
  };
}

class RestaurantAnalyticsService implements RestaurantAnalytics {
  async generateSalesReport(locationId: string, period: DateRange): Promise<SalesReport> {
    const orders = await this.getOrdersForPeriod(locationId, period);
    const orderItems = await this.getOrderItemsForPeriod(locationId, period);
    
    // Calculate revenue summary
    const revenueSummary = this.calculateRevenueSummary(orders, orderItems);
    
    // Calculate transaction summary
    const transactionSummary = this.calculateTransactionSummary(orders);
    
    // Generate hourly and daily breakdowns
    const hourlyBreakdown = this.generateHourlyBreakdown(orders);
    const dailyBreakdown = this.generateDailyBreakdown(orders);
    
    // Analyze payment methods
    const paymentMethods = this.analyzePaymentMethods(orders);
    
    // Calculate operational metrics
    const operationalMetrics = await this.calculateOperationalMetrics(locationId, period);
    
    return {
      location_id: locationId,
      reporting_period: period,
      revenue_summary: revenueSummary,
      transaction_summary: transactionSummary,
      hourly_breakdown: hourlyBreakdown,
      daily_breakdown: dailyBreakdown,
      payment_methods: paymentMethods,
      operational_metrics: operationalMetrics
    };
  }
  
  async analyzeMenuPerformance(period: DateRange): Promise<MenuPerformanceReport> {
    const orderItems = await this.getOrderItemsWithMenuDetails(period);
    const menuItems = await this.getMenuItemsWithCosts();
    
    // Calculate performance metrics for each menu item
    const itemPerformance = this.calculateMenuItemPerformance(orderItems, menuItems);
    
    // Identify top performers
    const topPerformers = {
      highest_revenue_items: itemPerformance
        .sort((a, b) => b.total_revenue - a.total_revenue)
        .slice(0, 10),
      highest_quantity_items: itemPerformance
        .sort((a, b) => b.quantity_sold - a.quantity_sold)
        .slice(0, 10),
      highest_margin_items: itemPerformance
        .sort((a, b) => b.profit_margin_percentage - a.profit_margin_percentage)
        .slice(0, 10),
      trending_items: await this.identifyTrendingItems(itemPerformance, period)
    };
    
    // Analyze by category
    const categoryAnalysis = this.analyzeByCat|egory(itemPerformance);
    
    // Identify underperformers
    const underperformers = {
      low_sales_items: itemPerformance
        .filter(item => item.quantity_sold < this.calculateAverageQuantity(itemPerformance) * 0.3)
        .sort((a, b) => a.quantity_sold - b.quantity_sold)
        .slice(0, 10),
      low_margin_items: itemPerformance
        .filter(item => item.profit_margin_percentage < 60) // Industry standard
        .sort((a, b) => a.profit_margin_percentage - b.profit_margin_percentage)
        .slice(0, 10),
      slow_moving_items: await this.identifySlowMovingItems(itemPerformance, period)
    };
    
    // Pricing analysis
    const pricingAnalysis = await this.analyzePricing(itemPerformance, period);
    
    return {
      reporting_period: period,
      top_performers: topPerformers,
      category_analysis: categoryAnalysis,
      underperformers: underperformers,
      pricing_analysis: pricingAnalysis
    };
  }
  
  private calculateMenuItemPerformance(
    orderItems: OrderItemWithDetails[], 
    menuItems: MenuItemWithCosts[]
  ): MenuItemPerformance[] {
    
    const performanceMap = new Map<string, MenuItemPerformance>();
    
    for (const orderItem of orderItems) {
      const menuItem = menuItems.find(m => m.id === orderItem.menu_item_id);
      if (!menuItem) continue;
      
      const existing = performanceMap.get(orderItem.menu_item_id) || {
        menu_item_id: orderItem.menu_item_id,
        item_name: menuItem.item_name,
        category: menuItem.category_name,
        quantity_sold: 0,
        total_revenue: 0,
        total_cost: 0,
        profit_margin: 0,
        profit_margin_percentage: 0,
        average_selling_price: 0,
        velocity: 0
      };
      
      existing.quantity_sold += orderItem.quantity;
      existing.total_revenue += orderItem.line_total;
      existing.total_cost += (menuItem.cost_per_serving * orderItem.quantity);
      
      performanceMap.set(orderItem.menu_item_id, existing);
    }
    
    // Calculate derived metrics
    for (const [itemId, performance] of performanceMap.entries()) {
      performance.profit_margin = performance.total_revenue - performance.total_cost;
      performance.profit_margin_percentage = performance.total_revenue > 0 
        ? (performance.profit_margin / performance.total_revenue) * 100 
        : 0;
      performance.average_selling_price = performance.quantity_sold > 0 
        ? performance.total_revenue / performance.quantity_sold 
        : 0;
      performance.velocity = performance.quantity_sold; // Could be more sophisticated
    }
    
    return Array.from(performanceMap.values());
  }
}
```

### Customer Experience Management

```sql
-- Customer feedback and reviews
CREATE TABLE customer_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Feedback identification
    restaurant_order_id UUID REFERENCES restaurant_orders(id),
    customer_id UUID REFERENCES customers(id),
    
    -- Feedback details
    overall_rating INTEGER NOT NULL, -- 1-5 scale
    food_quality_rating INTEGER,
    service_rating INTEGER,
    atmosphere_rating INTEGER,
    value_rating INTEGER,
    
    -- Detailed feedback
    feedback_text TEXT,
    favorite_items JSONB, -- Array of menu item IDs
    improvement_suggestions TEXT,
    
    -- Feedback source
    feedback_source VARCHAR(30) DEFAULT 'in_person', -- in_person, online, phone, email, social_media
    feedback_channel VARCHAR(100), -- Specific platform if online
    
    -- Staff mentioned
    server_mentioned_id UUID REFERENCES employees(id),
    chef_mentioned_id UUID REFERENCES employees(id),
    
    -- Response and follow-up
    response_required BOOLEAN DEFAULT false,
    management_response TEXT,
    response_date TIMESTAMPTZ,
    follow_up_required BOOLEAN DEFAULT false,
    follow_up_completed BOOLEAN DEFAULT false,
    
    -- Public review
    public_review BOOLEAN DEFAULT false,
    review_platform VARCHAR(50), -- google, yelp, tripadvisor, etc.
    review_url VARCHAR(500),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_overall_rating CHECK (overall_rating >= 1 AND overall_rating <= 5),
    CONSTRAINT valid_feedback_source CHECK (feedback_source IN ('in_person', 'online', 'phone', 'email', 'social_media'))
);

-- Loyalty program for restaurants
CREATE TABLE restaurant_loyalty_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Member identification
    customer_id UUID NOT NULL REFERENCES customers(id),
    membership_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Membership details
    membership_tier VARCHAR(20) DEFAULT 'bronze', -- bronze, silver, gold, platinum
    join_date DATE NOT NULL DEFAULT CURRENT_DATE,
    tier_achieved_date DATE,
    
    -- Points and rewards
    total_points_earned INTEGER DEFAULT 0,
    current_points_balance INTEGER DEFAULT 0,
    total_points_redeemed INTEGER DEFAULT 0,
    lifetime_spend DECIMAL(12,2) DEFAULT 0,
    
    -- Visit tracking
    total_visits INTEGER DEFAULT 0,
    last_visit_date DATE,
    average_spend_per_visit DECIMAL(8,2) DEFAULT 0,
    
    -- Preferences
    favorite_cuisine_types JSONB,
    dietary_restrictions JSONB,
    preferred_dining_times JSONB,
    communication_preferences JSONB,
    
    -- Status
    membership_status VARCHAR(20) DEFAULT 'active', -- active, inactive, suspended
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_membership_tier CHECK (membership_tier IN ('bronze', 'silver', 'gold', 'platinum')),
    CONSTRAINT valid_membership_status CHECK (membership_status IN ('active', 'inactive', 'suspended')),
    UNIQUE(tenant_id, customer_id)
);

-- Loyalty points transactions
CREATE TABLE loyalty_point_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    loyalty_member_id UUID NOT NULL REFERENCES restaurant_loyalty_members(id),
    
    -- Transaction details
    transaction_type VARCHAR(20) NOT NULL, -- earned, redeemed, expired, bonus, adjustment
    points_amount INTEGER NOT NULL,
    
    -- Source information
    restaurant_order_id UUID REFERENCES restaurant_orders(id),
    promotion_id UUID REFERENCES promotions(id),
    
    -- Transaction description
    description TEXT,
    expiry_date DATE, -- For earned points
    
    -- Processed by
    processed_by UUID REFERENCES employees(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_transaction_type CHECK (transaction_type IN ('earned', 'redeemed', 'expired', 'bonus', 'adjustment'))
);
```

This  restaurant management system provides sophisticated menu engineering, kitchen operations, table service management, and customer analytics specifically designed for restaurant operations while integrating seamlessly with the core ERP inventory, financial, and HR modules.
