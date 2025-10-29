package schema

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// NumberRenderer handles number input fields
type NumberRenderer struct {
	*BaseRenderer
}

func NewNumberRenderer(base *BaseRenderer) *NumberRenderer {
	return &NumberRenderer{BaseRenderer: base}
}

func (nr *NumberRenderer) Render(ctx context.Context, field *Field, value interface{}, errors []string) (string, error) {
	attrs := nr.buildInputAttributes(field, value)
	attrs["type"] = "number"

	// Add number-specific attributes
	if field.Config != nil {
		if step, exists := field.Config["step"]; exists {
			attrs["step"] = fmt.Sprintf("%v", step)
		}
	}

	inputHTML := fmt.Sprintf(`<input %s>`, nr.attributesToString(attrs))
	return nr.RenderContainer(field, inputHTML, errors), nil
}

func (nr *NumberRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldNumber
}

func (nr *NumberRenderer) GetRequiredAssets() []string {
	return []string{"css/input.css"}
}

// CurrencyRenderer handles currency input fields
type CurrencyRenderer struct {
	*BaseRenderer
}

func NewCurrencyRenderer(base *BaseRenderer) *CurrencyRenderer {
	return &CurrencyRenderer{BaseRenderer: base}
}

func (cr *CurrencyRenderer) Render(ctx context.Context, field *Field, value interface{}, errors []string) (string, error) {
	attrs := cr.buildInputAttributes(field, value)
	attrs["type"] = "text"
	attrs["inputmode"] = "decimal"

	// Add currency formatting
	currency := "USD"
	symbol := "$"

	if field.Config != nil {
		if cur, exists := field.Config["currency"]; exists {
			currency = fmt.Sprintf("%v", cur)
		}
		if sym, exists := field.Config["symbol"]; exists {
			symbol = fmt.Sprintf("%v", sym)
		}
	}

	attrs["data-currency"] = currency
	attrs["data-symbol"] = symbol
	attrs["data-mask"] = "currency"

	// Wrap input with currency symbol
	inputHTML := fmt.Sprintf(`
		<div class="currency-input">
			<span class="currency-symbol">%s</span>
			<input %s>
		</div>`, symbol, cr.attributesToString(attrs))

	return cr.RenderContainer(field, inputHTML, errors), nil
}

func (cr *CurrencyRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldCurrency
}

func (cr *CurrencyRenderer) GetRequiredAssets() []string {
	return []string{"css/currency.css", "js/currency-formatter.js"}
}

// SliderRenderer handles range/slider input fields
type SliderRenderer struct {
	*BaseRenderer
}

func NewSliderRenderer(base *BaseRenderer) *SliderRenderer {
	return &SliderRenderer{BaseRenderer: base}
}

func (sr *SliderRenderer) Render(ctx context.Context, field *Field, value interface{}, errors []string) (string, error) {
	attrs := sr.buildInputAttributes(field, value)
	attrs["type"] = "range"

	// Set default range if not specified
	if _, exists := attrs["min"]; !exists {
		attrs["min"] = "0"
	}
	if _, exists := attrs["max"]; !exists {
		attrs["max"] = "100"
	}

	// Add step
	if field.Config != nil {
		if step, exists := field.Config["step"]; exists {
			attrs["step"] = fmt.Sprintf("%v", step)
		} else {
			attrs["step"] = "1"
		}
	} else {
		attrs["step"] = "1"
	}

	// Show value indicator
	showValue := true
	if field.Config != nil {
		if show, exists := field.Config["showValue"]; exists {
			showValue = show.(bool)
		}
	}

	var inputHTML string
	if showValue {
		inputHTML = fmt.Sprintf(`
			<div class="slider-container">
				<input %s oninput="updateSliderValue('%s', this.value)">
				<span class="slider-value" id="%s-value">%s</span>
			</div>`, sr.attributesToString(attrs), field.Name, field.Name, attrs["value"])
	} else {
		inputHTML = fmt.Sprintf(`<input %s>`, sr.attributesToString(attrs))
	}

	return sr.RenderContainer(field, inputHTML, errors), nil
}

func (sr *SliderRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldSlider
}

func (sr *SliderRenderer) GetRequiredAssets() []string {
	return []string{"css/slider.css", "js/slider.js"}
}

// RatingRenderer handles star rating fields
type RatingRenderer struct {
	*BaseRenderer
}

func NewRatingRenderer(base *BaseRenderer) *RatingRenderer {
	return &RatingRenderer{BaseRenderer: base}
}

func (rr *RatingRenderer) Render(ctx context.Context, field *Field, value interface{}, errors []string) (string, error) {
	// Get configuration
	maxRating := 5
	allowHalf := false
	icon := "★"

	if field.Config != nil {
		if max, exists := field.Config["max"]; exists {
			if maxInt, err := strconv.Atoi(fmt.Sprintf("%v", max)); err == nil {
				maxRating = maxInt
			}
		}
		if half, exists := field.Config["allowHalf"]; exists {
			allowHalf = half.(bool)
		}
		if ic, exists := field.Config["icon"]; exists {
			icon = fmt.Sprintf("%v", ic)
		}
	}

	// Get current value
	currentValue := 0.0
	if value != nil {
		if val, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err == nil {
			currentValue = val
		}
	}

	// Build rating stars
	var stars []string
	for i := 1; i <= maxRating; i++ {
		starValue := float64(i)
		if allowHalf {
			// Add half star option
			halfValue := float64(i) - 0.5
			halfClass := "star-half"
			if currentValue >= halfValue {
				halfClass += " star-filled"
			}

			stars = append(stars, fmt.Sprintf(`
				<span class="%s" data-value="%.1f" onclick="setRating('%s', %.1f)">%s</span>`,
				halfClass, halfValue, field.Name, halfValue, icon))
		}

		starClass := "star"
		if currentValue >= starValue {
			starClass += " star-filled"
		}

		stars = append(stars, fmt.Sprintf(`
			<span class="%s" data-value="%.0f" onclick="setRating('%s', %.0f)">%s</span>`,
			starClass, starValue, field.Name, starValue, icon))
	}

	inputHTML := fmt.Sprintf(`
		<div class="rating-container" data-rating="%.1f">
			<input type="hidden" name="%s" id="%s" value="%.1f">
			<div class="rating-stars">%s</div>
		</div>`, currentValue, field.Name, field.Name, currentValue, strings.Join(stars, ""))

	return rr.RenderContainer(field, inputHTML, errors), nil
}

func (rr *RatingRenderer) SupportsFieldType(fieldType FieldType) bool {
	return fieldType == FieldRating
}

func (rr *RatingRenderer) GetRequiredAssets() []string {
	return []string{"css/rating.css", "js/rating.js"}
}