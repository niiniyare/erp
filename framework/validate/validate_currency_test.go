package validate_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"awo.so/framework/validate"
)

func TestCurrency_AcceptsDecimal(t *testing.T) {
	v := validate.Currency()
	d := decimal.NewFromFloat(1234567890.1234)
	if fe := v(d, nil); fe != nil {
		t.Errorf("want nil for decimal.Decimal, got: %v", fe)
	}
}

func TestCurrency_AcceptsString(t *testing.T) {
	v := validate.Currency()
	if fe := v("9999999999.9999", nil); fe != nil {
		t.Errorf("want nil for valid string, got: %v", fe)
	}
}

func TestCurrency_RejectsFloat64(t *testing.T) {
	v := validate.Currency()
	if fe := v(float64(123.45), nil); fe == nil {
		t.Error("want error for float64, got nil")
	}
}

func TestCurrency_RejectsInvalidString(t *testing.T) {
	v := validate.Currency()
	if fe := v("not-a-number", nil); fe == nil {
		t.Error("want error for non-numeric string, got nil")
	}
}

func TestCurrency_NilIsNoOp(t *testing.T) {
	v := validate.Currency()
	if fe := v(nil, nil); fe != nil {
		t.Errorf("want nil for nil value, got: %v", fe)
	}
}

func TestCurrencyRange_WithinBounds(t *testing.T) {
	min := decimal.NewFromFloat(0)
	max := decimal.NewFromFloat(1_000_000)
	v := validate.CurrencyRange(&min, &max)

	cases := []any{
		decimal.NewFromFloat(500_000),
		"999999.9999",
		int64(0),
	}
	for _, val := range cases {
		if fe := v(val, nil); fe != nil {
			t.Errorf("want nil for in-range value %v, got: %v", val, fe)
		}
	}
}

func TestCurrencyRange_BelowMin(t *testing.T) {
	min := decimal.NewFromFloat(100)
	v := validate.CurrencyRange(&min, nil)
	if fe := v(decimal.NewFromFloat(99.99), nil); fe == nil {
		t.Error("want error for value below min, got nil")
	}
}

func TestCurrencyRange_AboveMax(t *testing.T) {
	max := decimal.NewFromFloat(1000)
	v := validate.CurrencyRange(nil, &max)
	if fe := v(decimal.NewFromFloat(1000.01), nil); fe == nil {
		t.Error("want error for value above max, got nil")
	}
}

func TestCurrencyRange_NoPrecisionLoss(t *testing.T) {
	// Verifies exact decimal comparison — not float64 approximation.
	// float64(0.1) + float64(0.2) != 0.3 in IEEE 754.
	min := decimal.RequireFromString("0.3")
	v := validate.CurrencyRange(&min, nil)

	// decimal.NewFromString("0.1") + decimal.NewFromString("0.2") == 0.3 exactly.
	val := decimal.RequireFromString("0.1").Add(decimal.RequireFromString("0.2"))
	if fe := v(val, nil); fe != nil {
		t.Errorf("want no error for exact 0.3, got: %v (precision loss via float would fail)", fe)
	}
}
