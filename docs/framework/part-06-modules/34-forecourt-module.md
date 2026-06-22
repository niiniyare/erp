---
title: "Chapter 34: Forecourt Module"
part: "Part VI — Built-In Modules"
chapter: 34
section: "34-forecourt-module"
related:
  - "[Chapter 33: Finance Module](33-finance-module.md)"
  - "[Chapter 30: Signals, Queries](../part-05-workflow/30-signals-queries.md)"
---

# Chapter 34: Forecourt Module

The Forecourt module manages petroleum retail station operations: fuel dispensing, meter readings, wetstock reconciliation, shift management, and fleet card transactions. It is Kenya-specific, addressing requirements from EPRA (Energy and Petroleum Regulatory Authority) and NEMA (National Environment Management Authority).

---

## 34.1. Entity Hierarchy

### 34.1.1. Site → Tank → Pump → Nozzle

```
Site (forecourt station)
├── Tank 1 (Petrol - 20,000L capacity)
│   ├── Pump 1
│   │   ├── Nozzle 1A (Petrol)
│   │   └── Nozzle 1B (Petrol)
│   └── Pump 2
│       └── Nozzle 2A (Petrol)
├── Tank 2 (Diesel - 20,000L capacity)
│   └── Pump 3
│       └── Nozzle 3A (Diesel)
└── Tank 3 (Kerosene - 10,000L capacity)
    └── Pump 4
        └── Nozzle 4A (Kerosene)
```

Each nozzle is a separate dispensing point. Meter readings are taken per nozzle. Shifts are managed per site.

### 34.1.2. Product Grades

```go
type FuelGrade string

const (
    FuelPetrol   FuelGrade = "petrol"    // 93-octane (common in Kenya)
    FuelDiesel   FuelGrade = "diesel"    // standard diesel
    FuelKerosene FuelGrade = "kerosene"  // illuminating kerosene
    FuelLPG      FuelGrade = "lpg"       // liquefied petroleum gas
    FuelPremium  FuelGrade = "premium"   // 95-octane premium petrol (some stations)
)
```

Fuel prices are set per grade per site. Price changes require updating the `FuelPrice` entity, which triggers a re-calculation of the current shift's projected revenue.

### 34.1.3. PTS-2 Pump Controller

The PTS-2 (Pump Telemetry System) is an electronic pump controller used by most Kenyan forecourt operators. It provides:
- Electronic volume meters (totaliser readings in litres)
- Electronic cash registers (cumulative cash collected)
- Real-time pump status (idle, authorised, dispensing, stopped)

```go
type PTS2Reading struct {
    PumpID           uuid.UUID
    NozzleID         uuid.UUID
    ElectronicVolume decimal.Decimal  // cumulative litres dispensed (totaliser)
    ElectronicCash   decimal.Decimal  // cumulative cash (totaliser)
    ReadingTime      time.Time
    Source           string           // "electronic" | "manual"
}
```

---

## 34.2. Meter Readings

### 34.2.1. Electronic Volume Reading — From PTS-2

The electronic volume meter reads cumulative litres dispensed since the pump was installed. It never resets (unlike cash meters which may reset). Volume readings are taken at shift open and shift close.

### 34.2.2. Electronic Cash Reading

Cumulative cash received at the pump. Subtracted at shift open to get shift cash.

### 34.2.3. Manual Mechanical Reading

A mechanical odometer-style display on the physical pump. Used as backup when electronic readings are unavailable. Cashiers enter these during shift close.

### 34.2.4. Meter Cross-Validation — Detecting Fraud and Equipment Errors

For each nozzle, cross-validate electronic vs manual readings:

```go
func validateMeterReadings(electronic, manual MeterReading) error {
    volumeDiff := electronic.Volume.Sub(manual.Volume).Abs()
    maxAllowedDiff := decimal.NewFromFloat(0.5) // 0.5 litres tolerance

    if volumeDiff.GreaterThan(maxAllowedDiff) {
        return errs.NewBusinessError("METER_MISMATCH",
            "nozzle %s: electronic reading %s L differs from manual reading %s L by %s L (exceeds %s L tolerance)",
            reading.NozzleCode,
            electronic.Volume.StringFixed(2),
            manual.Volume.StringFixed(2),
            volumeDiff.StringFixed(2),
            maxAllowedDiff.StringFixed(2))
    }
    return nil
}
```

### 34.2.5. Cumulative vs Incremental Reading Logic

Meters record cumulative totals. Shift volume is computed as:
```
shift_volume = close_reading - open_reading
```

If a meter was reset or replaced during the shift, the system detects a negative incremental reading and flags it for manual review.

---

## 34.3. Dip Readings and Wetstock

### 34.3.1. Manual Dip Reading Entity

A dipstick measurement of the physical fuel level in each tank, taken at specified times (typically morning before trade and evening after trade close):

```go
type DipReading struct {
    TankID      uuid.UUID
    DipDate     time.Time
    DipCM       decimal.Decimal  // dipstick reading in centimetres
    VolumeL     decimal.Decimal  // computed from tank calibration chart
    Temperature decimal.Decimal  // optional — affects volume calculation
    TakenBy     uuid.UUID
    Verified    bool
    VerifiedBy  *uuid.UUID
}
```

### 34.3.2. Computed Theoretical Stock

```
theoretical_stock = opening_dip + deliveries_in_period - meter_sales_in_period
```

### 34.3.3. Variance Calculation

```
variance = actual_dip - theoretical_stock
variance_pct = variance / theoretical_stock * 100
```

Positive variance: more fuel in tank than expected (possible meter under-reading).
Negative variance: less fuel than expected (possible theft, evaporation, meter over-reading).

### 34.3.4. NEMA Environmental Compliance Thresholds

The National Environment Management Authority (NEMA) sets maximum acceptable variance limits for underground storage tanks to detect leaks:

```go
const (
    NEMAMonthlyVolumeThreshold   = 0.5  // 0.5% of throughput
    NEMADailyVolumeThreshold     = 0.3  // 0.3% of opening dip
)

func checkNEMACompliance(variance, throughput decimal.Decimal) bool {
    variancePct := variance.Abs().Div(throughput).Mul(decimal.NewFromInt(100))
    return variancePct.LessThan(decimal.NewFromFloat(NEMAMonthlyVolumeThreshold))
}
```

When the NEMA threshold is breached, the system generates a compliance alert and notifies the site manager.

---

## 34.4. Shift Management

### 34.4.1. Shift Open

```go
func openShift(ctx context.Context, input ShiftOpenInput) (*Shift, error) {
    // Record opening meter readings for each nozzle
    for _, reading := range input.OpeningReadings {
        meterReadingRepo.Create(ctx, MeterReadingCreate{
            NozzleID:    reading.NozzleID,
            Type:        "shift_open",
            Electronic:  reading.ElectronicVolume,
            Mechanical:  reading.MechanicalVolume,
            ReadingTime: time.Now(),
        })
    }

    return shiftRepo.Create(ctx, ShiftCreate{
        SiteID:      input.SiteID,
        CashierID:   input.CashierID,
        OpenedAt:    time.Now(),
        Status:      "open",
        OpeningCash: input.OpeningFloat,
    })
}
```

### 34.4.2. Shift Close — Cashier Submits

```
POST /api/v1/shifts/{id}/close
{
  "closing_readings": [...],
  "cash_declared": 45000.00,
  "discrepancy_explanation": "KES 500 shortage, coins counted wrong"
}

→ Shift status: "pending_approval"
→ ShiftCloseApprovalWorkflow starts
```

### 34.4.3. Forecourt Reconciliation Engine

After approval, the reconciliation engine computes variances per nozzle and per grade:

```
Nozzle 1A (Petrol):
  Volume sold (meter): 2,340.5 L
  Cash expected (price × volume): KES 274,838.50
  Cash declared: KES 274,450.00
  Cash variance: KES -388.50 (SHORT)

  Volume sold (meter): 2,340.5 L
  Volume expected (dip): 2,337.2 L
  Volume variance: 3.3 L (OVER — likely meter calibration drift)
```

---

## 34.5. Fleet Card Management

### 34.5.1. Fleet Customer Entity

Fleet customers are companies with credit accounts at the station. They receive a fleet card instead of paying cash.

```go
type FleetCustomer struct {
    CustomerID    uuid.UUID   // linked to Customer entity
    AccountNumber string
    CreditLimit   decimal.Decimal
    CurrentBalance decimal.Decimal
    Cards          []FleetCard
}
```

### 34.5.2. Card Authorisation Request

When a fleet vehicle presents a card at the pump, the PTS-2 queries the Awo API for authorisation:

```
POST /api/v1/fleet/authorise
{
  "card_number": "FLEET-001-004",
  "pump_id": "uuid",
  "requested_amount": 5000.00
}

→ Check: card is active, account has sufficient credit
→ Create AuthorisationRequest record
→ Return: { "authorised": true, "transaction_id": "uuid", "limit": 5000.00 }
```

### 34.5.3. Card Transaction Entity

After dispensing, the PTS-2 posts the actual transaction:

```
POST /api/v1/fleet/transactions
{
  "authorisation_id": "uuid",
  "volume_dispensed": 42.3,
  "amount": 4960.20,
  "odometer": 124567
}

→ Deduct from fleet account balance
→ Post GL entry (Debit: Accounts Receivable, Credit: Sales Revenue)
→ Generate transaction receipt
```
