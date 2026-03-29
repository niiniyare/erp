# Airline Reservation System

## ✈️ Overview

The Airline Reservation System module transforms the core ERP into a  airline operations management platform. It provides flight scheduling, passenger booking, crew management, aircraft maintenance, revenue management, and operational control capabilities designed specifically for airlines and aviation companies.

##  Flight Operations Management

### Aircraft Fleet Management

```sql
-- Aircraft master data
CREATE TABLE aircraft (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Aircraft identification
    aircraft_registration VARCHAR(20) UNIQUE NOT NULL, -- Tail number (e.g., N12345)
    aircraft_type VARCHAR(50) NOT NULL, -- B737-800, A320-200, etc.
    manufacturer VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    series VARCHAR(50),
    
    -- Aircraft specifications
    maximum_seats INTEGER NOT NULL,
    maximum_cargo_weight DECIMAL(10,2), -- in kg
    maximum_fuel_capacity DECIMAL(10,2), -- in liters
    maximum_takeoff_weight DECIMAL(10,2), -- in kg
    cruise_speed INTEGER, -- in knots
    maximum_range INTEGER, -- in nautical miles
    
    -- Ownership and leasing
    ownership_type VARCHAR(20) DEFAULT 'owned', -- owned, leased, wet_lease, dry_lease
    owner_company VARCHAR(255),
    lease_start_date DATE,
    lease_end_date DATE,
    lease_cost_per_month DECIMAL(12,2),
    
    -- Aircraft status
    operational_status VARCHAR(20) DEFAULT 'active', -- active, maintenance, grounded, retired
    current_location VARCHAR(10), -- Airport code
    home_base_airport VARCHAR(10), -- Home airport code
    
    -- Maintenance tracking
    total_flight_hours DECIMAL(10,2) DEFAULT 0,
    total_flight_cycles INTEGER DEFAULT 0,
    last_major_maintenance DATE,
    next_major_maintenance_due DATE,
    maintenance_provider VARCHAR(255),
    
    -- Certification and compliance
    certificate_of_airworthiness_expiry DATE,
    insurance_expiry_date DATE,
    annual_inspection_due DATE,
    
    -- Financial tracking
    acquisition_cost DECIMAL(15,2),
    current_book_value DECIMAL(15,2),
    depreciation_method VARCHAR(30) DEFAULT 'straight_line',
    
    -- Configuration
    seat_configuration JSONB, -- Detailed seat map configuration
    galley_configuration JSONB,
    entertainment_system BOOLEAN DEFAULT false,
    wifi_available BOOLEAN DEFAULT false,
    
    -- Operational metrics
    daily_utilization_target DECIMAL(4,2), -- Target hours per day
    fuel_efficiency_rating VARCHAR(10), -- A, B, C, D rating
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_ownership_type CHECK (ownership_type IN ('owned', 'leased', 'wet_lease', 'dry_lease')),
    CONSTRAINT valid_operational_status CHECK (operational_status IN ('active', 'maintenance', 'grounded', 'retired'))
);

-- Airports and destinations
CREATE TABLE airports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Airport identification
    airport_code CHAR(3) UNIQUE NOT NULL, -- IATA code (e.g., LAX)
    icao_code CHAR(4) UNIQUE, -- ICAO code (e.g., KLAX)
    airport_name VARCHAR(255) NOT NULL,
    
    -- Location details
    city VARCHAR(100) NOT NULL,
    country_code CHAR(2) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    coordinates JSONB, -- {latitude, longitude}
    elevation_feet INTEGER,
    
    -- Airport capabilities
    runway_count INTEGER DEFAULT 1,
    max_aircraft_size VARCHAR(20), -- narrow_body, wide_body, jumbo
    hub_type VARCHAR(20) DEFAULT 'spoke', -- hub, focus_city, spoke, seasonal
    
    -- Operational details
    curfew_restrictions JSONB, -- Noise restrictions and operating hours
    slot_restrictions BOOLEAN DEFAULT false,
    customs_available BOOLEAN DEFAULT true,
    fuel_available BOOLEAN DEFAULT true,
    maintenance_services BOOLEAN DEFAULT false,
    
    -- Costs and fees
    landing_fee_structure JSONB,
    parking_fee_per_hour DECIMAL(8,2),
    fuel_price_per_liter DECIMAL(6,4),
    handling_fee DECIMAL(8,2),
    
    -- Status
    operational_status VARCHAR(20) DEFAULT 'active', -- active, closed, restricted
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_hub_type CHECK (hub_type IN ('hub', 'focus_city', 'spoke', 'seasonal')),
    CONSTRAINT valid_max_aircraft_size CHECK (max_aircraft_size IN ('narrow_body', 'wide_body', 'jumbo')),
    CONSTRAINT valid_operational_status CHECK (operational_status IN ('active', 'closed', 'restricted'))
);

-- Routes between airports
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Route identification
    route_code VARCHAR(20) UNIQUE NOT NULL,
    origin_airport_id UUID NOT NULL REFERENCES airports(id),
    destination_airport_id UUID NOT NULL REFERENCES airports(id),
    
    -- Route characteristics
    distance_nautical_miles INTEGER NOT NULL,
    flight_time_minutes INTEGER NOT NULL,
    route_type VARCHAR(20) DEFAULT 'domestic', -- domestic, international, regional
    
    -- Operational parameters
    preferred_aircraft_types JSONB, -- Array of suitable aircraft types
    minimum_aircraft_size VARCHAR(20),
    seasonal_route BOOLEAN DEFAULT false,
    seasonal_start_date DATE,
    seasonal_end_date DATE,
    
    -- Market analysis
    market_category VARCHAR(20) DEFAULT 'leisure', -- business, leisure, mixed
    competition_level VARCHAR(20) DEFAULT 'medium', -- low, medium, high
    demand_pattern VARCHAR(20) DEFAULT 'stable', -- growing, stable, declining
    
    -- Pricing and yield
    base_fare DECIMAL(8,2),
    fuel_surcharge DECIMAL(6,2),
    average_load_factor DECIMAL(5,2), -- Historical average
    
    -- Status
    route_status VARCHAR(20) DEFAULT 'active', -- active, suspended, cancelled
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_route_type CHECK (route_type IN ('domestic', 'international', 'regional')),
    CONSTRAINT valid_market_category CHECK (market_category IN ('business', 'leisure', 'mixed')),
    CONSTRAINT valid_competition_level CHECK (competition_level IN ('low', 'medium', 'high')),
    CONSTRAINT valid_demand_pattern CHECK (demand_pattern IN ('growing', 'stable', 'declining')),
    CONSTRAINT valid_route_status CHECK (route_status IN ('active', 'suspended', 'cancelled')),
    CONSTRAINT different_airports CHECK (origin_airport_id != destination_airport_id)
);
```

### Flight Scheduling

```sql
-- Flight schedules (recurring flight patterns)
CREATE TABLE flight_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Schedule identification
    flight_number VARCHAR(10) NOT NULL, -- e.g., AA1234
    route_id UUID NOT NULL REFERENCES routes(id),
    
    -- Schedule pattern
    effective_start_date DATE NOT NULL,
    effective_end_date DATE,
    days_of_week JSONB NOT NULL, -- Array of day numbers (1=Monday, 7=Sunday)
    
    -- Timing
    departure_time_local TIME NOT NULL,
    arrival_time_local TIME NOT NULL,
    flight_duration_minutes INTEGER NOT NULL,
    
    -- Aircraft assignment
    aircraft_type VARCHAR(50), -- Preferred aircraft type
    aircraft_id UUID REFERENCES aircraft(id), -- Specific aircraft assignment
    
    -- Schedule metadata
    schedule_type VARCHAR(20) DEFAULT 'regular', -- regular, charter, seasonal, cargo
    service_class VARCHAR(20) DEFAULT 'full_service', -- full_service, low_cost, premium
    
    -- Operational parameters
    boarding_gate_time_minutes INTEGER DEFAULT 30,
    minimum_connection_time_minutes INTEGER DEFAULT 45,
    block_time_minutes INTEGER, -- Total scheduled block time
    
    -- Commercial parameters
    booking_class_configuration JSONB, -- Available booking classes and allocations
    meal_service BOOLEAN DEFAULT true,
    entertainment_service BOOLEAN DEFAULT false,
    
    -- Status
    schedule_status VARCHAR(20) DEFAULT 'active', -- active, suspended, cancelled
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_schedule_type CHECK (schedule_type IN ('regular', 'charter', 'seasonal', 'cargo')),
    CONSTRAINT valid_service_class CHECK (service_class IN ('full_service', 'low_cost', 'premium')),
    CONSTRAINT valid_schedule_status CHECK (schedule_status IN ('active', 'suspended', 'cancelled'))
);

-- Individual flights (instances of scheduled flights)
CREATE TABLE flights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Flight identification
    flight_number VARCHAR(10) NOT NULL,
    flight_date DATE NOT NULL,
    flight_schedule_id UUID REFERENCES flight_schedules(id),
    
    -- Route information
    origin_airport_id UUID NOT NULL REFERENCES airports(id),
    destination_airport_id UUID NOT NULL REFERENCES airports(id),
    
    -- Aircraft assignment
    aircraft_id UUID NOT NULL REFERENCES aircraft(id),
    
    -- Scheduled times (local timezone)
    scheduled_departure_local TIMESTAMPTZ NOT NULL,
    scheduled_arrival_local TIMESTAMPTZ NOT NULL,
    
    -- Actual times
    actual_departure_local TIMESTAMPTZ,
    actual_arrival_local TIMESTAMPTZ,
    
    -- Flight status
    flight_status VARCHAR(20) DEFAULT 'scheduled', -- scheduled, boarding, departed, arrived, cancelled, delayed
    delay_minutes INTEGER DEFAULT 0,
    delay_reason VARCHAR(100),
    cancellation_reason VARCHAR(255),
    
    -- Gate and terminal information
    departure_gate VARCHAR(10),
    arrival_gate VARCHAR(10),
    departure_terminal VARCHAR(10),
    arrival_terminal VARCHAR(10),
    
    -- Capacity and bookings
    total_seats INTEGER NOT NULL,
    available_seats INTEGER,
    booked_passengers INTEGER DEFAULT 0,
    checked_in_passengers INTEGER DEFAULT 0,
    
    -- Seat class breakdown
    first_class_seats INTEGER DEFAULT 0,
    business_class_seats INTEGER DEFAULT 0,
    premium_economy_seats INTEGER DEFAULT 0,
    economy_class_seats INTEGER DEFAULT 0,
    
    -- Commercial information
    base_fare DECIMAL(10,2),
    fuel_surcharge DECIMAL(8,2),
    taxes_and_fees DECIMAL(8,2),
    
    -- Operational metrics
    block_time_minutes INTEGER,
    air_time_minutes INTEGER,
    fuel_consumption_liters DECIMAL(10,2),
    
    -- Weather and conditions
    weather_conditions JSONB,
    
    -- Crew assignment
    captain_id UUID REFERENCES employees(id),
    first_officer_id UUID REFERENCES employees(id),
    cabin_crew_ids JSONB, -- Array of crew member IDs
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_flight_status CHECK (flight_status IN ('scheduled', 'boarding', 'departed', 'arrived', 'cancelled', 'delayed')),
    UNIQUE(flight_number, flight_date)
);

-- Flight segments for multi-leg flights
CREATE TABLE flight_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    
    -- Segment identification
    segment_number INTEGER NOT NULL,
    origin_airport_id UUID NOT NULL REFERENCES airports(id),
    destination_airport_id UUID NOT NULL REFERENCES airports(id),
    
    -- Segment timing
    scheduled_departure TIMESTAMPTZ NOT NULL,
    scheduled_arrival TIMESTAMPTZ NOT NULL,
    actual_departure TIMESTAMPTZ,
    actual_arrival TIMESTAMPTZ,
    
    -- Segment distance and duration
    distance_nautical_miles INTEGER,
    scheduled_duration_minutes INTEGER,
    actual_duration_minutes INTEGER,
    
    -- Ground time
    minimum_ground_time_minutes INTEGER DEFAULT 30,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(flight_id, segment_number)
);
```

##  Passenger Reservation System

### Booking and Reservation Management

```sql
-- Passenger reservations (PNR - Passenger Name Record)
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Reservation identification
    pnr VARCHAR(6) UNIQUE NOT NULL, -- 6-character alphanumeric PNR
    confirmation_number VARCHAR(20) UNIQUE,
    
    -- Booking details
    booking_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    booking_channel VARCHAR(30) DEFAULT 'online', -- online, mobile, call_center, travel_agent, airport
    booking_agent_id UUID REFERENCES employees(id),
    travel_agent_id UUID REFERENCES travel_agents(id),
    
    -- Customer information
    lead_passenger_id UUID NOT NULL REFERENCES customers(id),
    total_passengers INTEGER NOT NULL DEFAULT 1,
    
    -- Trip information
    trip_type VARCHAR(20) DEFAULT 'round_trip', -- one_way, round_trip, multi_city, open_jaw
    booking_class VARCHAR(5) DEFAULT 'Y', -- Booking class code (Y, M, H, etc.)
    fare_basis VARCHAR(20),
    
    -- Reservation status
    reservation_status VARCHAR(20) DEFAULT 'confirmed', -- confirmed, cancelled, no_show, refunded
    
    -- Payment information
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency_code CHAR(3) DEFAULT 'USD',
    payment_status VARCHAR(20) DEFAULT 'pending', -- pending, paid, partial, failed, refunded
    
    -- Special services
    special_meal_requests JSONB,
    special_assistance_requests JSONB,
    frequent_flyer_number VARCHAR(20),
    
    -- Contact information
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    emergency_contact JSONB,
    
    -- Ticketing
    ticket_issued BOOLEAN DEFAULT false,
    ticket_number VARCHAR(20),
    ticket_issue_date TIMESTAMPTZ,
    
    -- Time limits
    ticketing_time_limit TIMESTAMPTZ,
    payment_time_limit TIMESTAMPTZ,
    
    -- Metadata
    group_booking BOOLEAN DEFAULT false,
    corporate_booking BOOLEAN DEFAULT false,
    corporate_account_id UUID REFERENCES customers(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_trip_type CHECK (trip_type IN ('one_way', 'round_trip', 'multi_city', 'open_jaw')),
    CONSTRAINT valid_reservation_status CHECK (reservation_status IN ('confirmed', 'cancelled', 'no_show', 'refunded')),
    CONSTRAINT valid_payment_status CHECK (payment_status IN ('pending', 'paid', 'partial', 'failed', 'refunded')),
    CONSTRAINT valid_booking_channel CHECK (booking_channel IN ('online', 'mobile', 'call_center', 'travel_agent', 'airport'))
);

-- Flight segments within a reservation
CREATE TABLE reservation_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    
    -- Segment identification
    segment_number INTEGER NOT NULL,
    flight_id UUID NOT NULL REFERENCES flights(id),
    
    -- Passenger details for this segment
    passenger_id UUID NOT NULL REFERENCES customers(id),
    
    -- Seat assignment
    seat_number VARCHAR(5),
    seat_class VARCHAR(20) DEFAULT 'economy', -- first, business, premium_economy, economy
    upgrade_eligible BOOLEAN DEFAULT false,
    
    -- Booking details
    booking_class VARCHAR(5) NOT NULL,
    fare_basis VARCHAR(20),
    fare_amount DECIMAL(10,2) NOT NULL,
    taxes_amount DECIMAL(8,2) DEFAULT 0,
    
    -- Segment status
    segment_status VARCHAR(20) DEFAULT 'confirmed', -- confirmed, waitlisted, cancelled, no_show
    
    -- Check-in status
    checked_in BOOLEAN DEFAULT false,
    check_in_time TIMESTAMPTZ,
    boarding_pass_issued BOOLEAN DEFAULT false,
    
    -- Baggage allowance
    baggage_allowance_kg INTEGER DEFAULT 23,
    extra_baggage_kg INTEGER DEFAULT 0,
    
    -- Special services for this segment
    meal_preference VARCHAR(50),
    special_assistance JSONB,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_seat_class CHECK (seat_class IN ('first', 'business', 'premium_economy', 'economy')),
    CONSTRAINT valid_segment_status CHECK (segment_status IN ('confirmed', 'waitlisted', 'cancelled', 'no_show')),
    UNIQUE(reservation_id, segment_number)
);

-- Passenger information (extends customers table for airline-specific needs)
CREATE TABLE passengers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    
    -- Travel document information
    passport_number VARCHAR(20),
    passport_country_code CHAR(2),
    passport_expiry_date DATE,
    visa_required BOOLEAN DEFAULT false,
    visa_number VARCHAR(20),
    
    -- Passenger preferences
    seat_preference VARCHAR(20), -- window, aisle, middle, any
    meal_preference VARCHAR(50), -- vegetarian, kosher, halal, etc.
    frequent_flyer_programs JSONB, -- Array of {airline, number, status}
    
    -- Special requirements
    wheelchair_required BOOLEAN DEFAULT false,
    assistance_required VARCHAR(100),
    unaccompanied_minor BOOLEAN DEFAULT false,
    age_category VARCHAR(20) DEFAULT 'adult', -- infant, child, adult, senior
    
    -- Travel history
    total_flights INTEGER DEFAULT 0,
    miles_flown INTEGER DEFAULT 0,
    last_flight_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_age_category CHECK (age_category IN ('infant', 'child', 'adult', 'senior'))
);
```

### Check-in and Boarding

```sql
-- Check-in process
CREATE TABLE check_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_segment_id UUID NOT NULL REFERENCES reservation_segments(id) ON DELETE CASCADE,
    
    -- Check-in details
    check_in_method VARCHAR(20) NOT NULL, -- online, mobile, kiosk, counter
    check_in_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_in_location VARCHAR(100), -- Gate, counter, or device identifier
    
    -- Seat assignment
    seat_number VARCHAR(5),
    seat_assignment_method VARCHAR(20), -- auto, selected, upgraded, changed
    upgrade_applied BOOLEAN DEFAULT false,
    upgrade_type VARCHAR(50),
    
    -- Boarding information
    boarding_group VARCHAR(5),
    boarding_sequence INTEGER,
    priority_boarding BOOLEAN DEFAULT false,
    
    -- Baggage check
    checked_bags_count INTEGER DEFAULT 0,
    carry_on_bags_count INTEGER DEFAULT 1,
    baggage_tags JSONB, -- Array of baggage tag numbers
    
    -- Document verification
    documents_verified BOOLEAN DEFAULT false,
    document_verification_time TIMESTAMPTZ,
    verified_by_agent_id UUID REFERENCES employees(id),
    
    -- Special services
    special_meal_confirmed BOOLEAN DEFAULT false,
    wheelchair_service_requested BOOLEAN DEFAULT false,
    
    -- Boarding pass
    boarding_pass_number VARCHAR(20),
    boarding_pass_barcode VARCHAR(100),
    mobile_boarding_pass BOOLEAN DEFAULT false,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_check_in_method CHECK (check_in_method IN ('online', 'mobile', 'kiosk', 'counter')),
    CONSTRAINT valid_seat_assignment_method CHECK (seat_assignment_method IN ('auto', 'selected', 'upgraded', 'changed'))
);

-- Boarding process
CREATE TABLE boarding_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_segment_id UUID NOT NULL REFERENCES reservation_segments(id) ON DELETE CASCADE,
    
    -- Boarding details
    boarding_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    boarding_gate VARCHAR(10) NOT NULL,
    boarding_agent_id UUID REFERENCES employees(id),
    
    -- Boarding verification
    boarding_pass_scanned BOOLEAN DEFAULT true,
    document_checked BOOLEAN DEFAULT false,
    
    -- Boarding status
    boarding_status VARCHAR(20) DEFAULT 'boarded', -- boarded, denied, no_show
    denial_reason VARCHAR(255),
    
    -- Seat confirmation
    final_seat_number VARCHAR(5),
    seat_change_reason VARCHAR(100),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_boarding_status CHECK (boarding_status IN ('boarded', 'denied', 'no_show'))
);

-- Baggage tracking
CREATE TABLE baggage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Baggage identification
    baggage_tag_number VARCHAR(20) UNIQUE NOT NULL,
    reservation_segment_id UUID NOT NULL REFERENCES reservation_segments(id),
    
    -- Baggage details
    baggage_type VARCHAR(20) DEFAULT 'checked', -- checked, carry_on, special
    weight_kg DECIMAL(6,2) NOT NULL,
    dimensions JSONB, -- {length, width, height}
    baggage_description TEXT,
    
    -- Special handling
    fragile BOOLEAN DEFAULT false,
    valuable BOOLEAN DEFAULT false,
    priority BOOLEAN DEFAULT false,
    special_handling_code VARCHAR(10),
    
    -- Routing information
    origin_airport_id UUID NOT NULL REFERENCES airports(id),
    destination_airport_id UUID NOT NULL REFERENCES airports(id),
    routing_flights JSONB, -- Array of flight IDs for baggage routing
    
    -- Status tracking
    baggage_status VARCHAR(20) DEFAULT 'checked_in', -- checked_in, loaded, in_transit, arrived, delivered, delayed, lost
    current_location VARCHAR(100),
    
    -- Timeline
    checked_in_time TIMESTAMPTZ DEFAULT NOW(),
    loaded_time TIMESTAMPTZ,
    arrival_time TIMESTAMPTZ,
    delivered_time TIMESTAMPTZ,
    
    -- Claims and issues
    delayed BOOLEAN DEFAULT false,
    delay_reason VARCHAR(255),
    claim_filed BOOLEAN DEFAULT false,
    claim_number VARCHAR(20),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_baggage_type CHECK (baggage_type IN ('checked', 'carry_on', 'special')),
    CONSTRAINT valid_baggage_status CHECK (baggage_status IN ('checked_in', 'loaded', 'in_transit', 'arrived', 'delivered', 'delayed', 'lost'))
);
```

##  Revenue Management & Pricing

### Dynamic Pricing Engine

```typescript
interface RevenueManagementEngine {
  calculateOptimalPrices(flightId: string): Promise<PricingRecommendation>;
  updateInventoryAllocation(flightId: string): Promise<InventoryAllocation>;
  analyzeBookingPatterns(routeId: string, period: DateRange): Promise<BookingAnalysis>;
  forecastDemand(flightId: string): Promise<DemandForecast>;
}

interface PricingRecommendation {
  flight_id: string;
  analysis_date: Date;
  
  fare_classes: FareClassRecommendation[];
  
  demand_forecast: {
    total_demand: number;
    price_elasticity: number;
    confidence_level: number;
  };
  
  competitive_analysis: {
    competitor_prices: CompetitorPrice[];
    market_position: 'premium' | 'competitive' | 'value';
  };
  
  optimization_metrics: {
    expected_revenue: number;
    expected_load_factor: number;
    revenue_per_available_seat: number;
  };
}

interface FareClassRecommendation {
  booking_class: string;
  current_price: number;
  recommended_price: number;
  price_change_percentage: number;
  seat_allocation: number;
  expected_bookings: number;
  revenue_contribution: number;
}

class AirlineRevenueService implements RevenueManagementEngine {
  async calculateOptimalPrices(flightId: string): Promise<PricingRecommendation> {
    const flight = await this.getFlight(flightId);
    const route = await this.getRoute(flight.origin_airport_id, flight.destination_airport_id);
    const historicalData = await this.getHistoricalBookingData(route.id, flight.flight_date);
    
    // Analyze current booking status
    const currentBookings = await this.getCurrentBookings(flightId);
    const daysToFlight = this.calculateDaysToFlight(flight.flight_date);
    
    // Demand forecasting
    const demandForecast = await this.forecastDemand(flightId);
    
    // Competitive analysis
    const competitorPrices = await this.getCompetitorPrices(route.id, flight.flight_date);
    
    // Price optimization using revenue management algorithms
    const fareClassRecommendations = await this.optimizeFareClasses(
      flight,
      demandForecast,
      competitorPrices,
      currentBookings,
      daysToFlight
    );
    
    // Calculate expected metrics
    const optimizationMetrics = this.calculateOptimizationMetrics(
      fareClassRecommendations,
      flight.total_seats
    );
    
    return {
      flight_id: flightId,
      analysis_date: new Date(),
      fare_classes: fareClassRecommendations,
      demand_forecast: {
        total_demand: demandForecast.total_demand,
        price_elasticity: demandForecast.price_elasticity,
        confidence_level: demandForecast.confidence_level
      },
      competitive_analysis: {
        competitor_prices: competitorPrices,
        market_position: this.determineMarketPosition(fareClassRecommendations, competitorPrices)
      },
      optimization_metrics: optimizationMetrics
    };
  }
  
  private async optimizeFareClasses(
    flight: Flight,
    demandForecast: DemandForecast,
    competitorPrices: CompetitorPrice[],
    currentBookings: BookingSummary,
    daysToFlight: number
  ): Promise<FareClassRecommendation[]> {
    
    const fareClasses = await this.getFareClasses(flight.route_id);
    const recommendations: FareClassRecommendation[] = [];
    
    for (const fareClass of fareClasses) {
      // Get current price and booking velocity
      const currentPrice = await this.getCurrentPrice(flight.id, fareClass.booking_class);
      const bookingVelocity = this.calculateBookingVelocity(
        currentBookings,
        fareClass.booking_class,
        daysToFlight
      );
      
      // Apply pricing strategy based on days to flight
      let pricingStrategy: 'aggressive' | 'moderate' | 'conservative';
      if (daysToFlight > 60) {
        pricingStrategy = 'aggressive'; // Lower prices to stimulate early bookings
      } else if (daysToFlight > 14) {
        pricingStrategy = 'moderate'; // Balanced approach
      } else {
        pricingStrategy = 'conservative'; // Higher prices for last-minute bookings
      }
      
      // Calculate optimal price using demand-based pricing model
      const optimalPrice = this.calculateOptimalPrice({
        basePrice: fareClass.base_price,
        currentPrice,
        demandForecast,
        competitorPrices,
        bookingVelocity,
        daysToFlight,
        pricingStrategy,
        fareClassElasticity: fareClass.price_elasticity
      });
      
      // Calculate seat allocation using bid-price control
      const seatAllocation = this.calculateOptimalSeatAllocation(
        fareClass,
        demandForecast,
        flight.available_seats,
        optimalPrice
      );
      
      // Forecast bookings and revenue
      const expectedBookings = Math.min(
        seatAllocation,
        demandForecast.demand_by_class[fareClass.booking_class] || 0
      );
      
      recommendations.push({
        booking_class: fareClass.booking_class,
        current_price: currentPrice,
        recommended_price: optimalPrice,
        price_change_percentage: ((optimalPrice - currentPrice) / currentPrice) * 100,
        seat_allocation: seatAllocation,
        expected_bookings: expectedBookings,
        revenue_contribution: expectedBookings * optimalPrice
      });
    }
    
    return recommendations;
  }
  
  private calculateOptimalPrice(params: {
    basePrice: number;
    currentPrice: number;
    demandForecast: DemandForecast;
    competitorPrices: CompetitorPrice[];
    bookingVelocity: number;
    daysToFlight: number;
    pricingStrategy: 'aggressive' | 'moderate' | 'conservative';
    fareClassElasticity: number;
  }): number {
    
    let optimalPrice = params.basePrice;
    
    // Demand adjustment
    const demandMultiplier = this.calculateDemandMultiplier(
      params.demandForecast.total_demand,
      params.daysToFlight
    );
    optimalPrice *= demandMultiplier;
    
    // Competitive adjustment
    const avgCompetitorPrice = params.competitorPrices.reduce(
      (sum, comp) => sum + comp.price, 0
    ) / params.competitorPrices.length;
    
    const competitiveAdjustment = this.calculateCompetitiveAdjustment(
      optimalPrice,
      avgCompetitorPrice,
      params.pricingStrategy
    );
    optimalPrice *= competitiveAdjustment;
    
    // Booking velocity adjustment
    const velocityAdjustment = this.calculateVelocityAdjustment(
      params.bookingVelocity,
      params.daysToFlight
    );
    optimalPrice *= velocityAdjustment;
    
    // Apply price elasticity
    if (params.fareClassElasticity) {
      const elasticityAdjustment = Math.pow(
        (params.demandForecast.total_demand / 100), 
        -1 / params.fareClassElasticity
      );
      optimalPrice *= elasticityAdjustment;
    }
    
    // Apply strategy constraints
    switch (params.pricingStrategy) {
      case 'aggressive':
        optimalPrice = Math.max(optimalPrice, params.basePrice * 0.7); // No less than 70% of base
        break;
      case 'moderate':
        optimalPrice = Math.max(optimalPrice, params.basePrice * 0.85); // No less than 85% of base
        break;
      case 'conservative':
        optimalPrice = Math.min(optimalPrice, params.basePrice * 3.0); // No more than 300% of base
        break;
    }
    
    return Math.round(optimalPrice * 100) / 100; // Round to nearest cent
  }
}
```

### Fare Classes and Inventory Control

```sql
-- Fare classes and pricing rules
CREATE TABLE fare_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Fare class identification
    booking_class CHAR(1) NOT NULL, -- Y, M, H, Q, etc.
    fare_basis VARCHAR(20) NOT NULL, -- YCA21, MCA14, etc.
    cabin_class VARCHAR(20) NOT NULL, -- economy, premium_economy, business, first
    
    -- Fare rules
    advance_purchase_days INTEGER DEFAULT 0,
    minimum_stay_days INTEGER DEFAULT 0,
    maximum_stay_days INTEGER,
    saturday_night_stay_required BOOLEAN DEFAULT false,
    
    -- Cancellation and changes
    refundable BOOLEAN DEFAULT false,
    changeable BOOLEAN DEFAULT true,
    change_fee DECIMAL(8,2) DEFAULT 0,
    cancellation_fee DECIMAL(8,2) DEFAULT 0,
    
    -- Booking restrictions
    blackout_dates JSONB,
    seasonal_restrictions JSONB,
    route_restrictions JSONB,
    
    -- Pricing
    base_price DECIMAL(10,2) NOT NULL,
    price_elasticity DECIMAL(4,3), -- Price elasticity coefficient
    
    -- Revenue management
    booking_priority INTEGER DEFAULT 1, -- Higher number = higher priority
    revenue_contribution_score DECIMAL(6,3),
    
    -- Seat allocation
    initial_allocation_percentage DECIMAL(5,2) DEFAULT 10,
    maximum_allocation_percentage DECIMAL(5,2) DEFAULT 100,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    effective_date DATE NOT NULL,
    expiry_date DATE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_cabin_class CHECK (cabin_class IN ('economy', 'premium_economy', 'business', 'first')),
    UNIQUE(booking_class, fare_basis)
);

-- Flight inventory allocation
CREATE TABLE flight_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    
    -- Inventory by cabin class
    first_class_total INTEGER DEFAULT 0,
    first_class_available INTEGER DEFAULT 0,
    first_class_oversold INTEGER DEFAULT 0,
    
    business_class_total INTEGER DEFAULT 0,
    business_class_available INTEGER DEFAULT 0,
    business_class_oversold INTEGER DEFAULT 0,
    
    premium_economy_total INTEGER DEFAULT 0,
    premium_economy_available INTEGER DEFAULT 0,
    premium_economy_oversold INTEGER DEFAULT 0,
    
    economy_class_total INTEGER DEFAULT 0,
    economy_class_available INTEGER DEFAULT 0,
    economy_class_oversold INTEGER DEFAULT 0,
    
    -- Inventory controls
    oversale_protection BOOLEAN DEFAULT true,
    maximum_oversale_percentage DECIMAL(5,2) DEFAULT 5.0,
    
    -- Revenue management
    last_revenue_update TIMESTAMPTZ,
    next_review_time TIMESTAMPTZ,
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(flight_id)
);

-- Fare class inventory allocation per flight
CREATE TABLE flight_fare_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    fare_class_id UUID NOT NULL REFERENCES fare_classes(id),
    
    -- Inventory allocation
    seats_allocated INTEGER NOT NULL DEFAULT 0,
    seats_sold INTEGER DEFAULT 0,
    seats_available INTEGER DEFAULT 0,
    
    -- Dynamic pricing
    current_price DECIMAL(10,2) NOT NULL,
    price_last_updated TIMESTAMPTZ DEFAULT NOW(),
    price_update_reason VARCHAR(100),
    
    -- Revenue management metrics
    bid_price DECIMAL(10,2), -- Minimum price to accept booking
    opportunity_cost DECIMAL(10,2), -- Cost of displacing higher-value passenger
    
    -- Booking controls
    nested_inventory BOOLEAN DEFAULT true,
    authorization_level INTEGER DEFAULT 1, -- Level required to override controls
    
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(flight_id, fare_class_id)
);
```

This  airline reservation system provides sophisticated flight operations, passenger management, and revenue optimization capabilities specifically designed for airline operations while integrating seamlessly with the core ERP financial and operational modules.
