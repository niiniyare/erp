package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strconv"
)

// ConfigValue is an immutable value object handling different data types
type ConfigValue struct {
	Raw      any
	DataType DataType
}

// NewConfigValue creates a new configuration value with validation
func NewConfigValue(value any, dataType DataType) (ConfigValue, error) {
	if !dataType.Valid() {
		return ConfigValue{}, ErrInvalidDataType
	}

	if !isValidForType(value, dataType) {
		return ConfigValue{}, ErrInvalidValueForType
	}

	// Normalize value based on type
	normalized, err := normalizeValue(value, dataType)
	if err != nil {
		return ConfigValue{}, err
	}

	return ConfigValue{
		Raw:      normalized,
		DataType: dataType,
	}, nil
}

// NewStringValue creates a string configuration value
func NewStringValue(value string) ConfigValue {
	return ConfigValue{
		Raw:      value,
		DataType: DataTypeString,
	}
}

// NewIntegerValue creates an integer configuration value
func NewIntegerValue(value int) ConfigValue {
	return ConfigValue{
		Raw:      value,
		DataType: DataTypeInteger,
	}
}

// NewBooleanValue creates a boolean configuration value
func NewBooleanValue(value bool) ConfigValue {
	return ConfigValue{
		Raw:      value,
		DataType: DataTypeBoolean,
	}
}

// NewDecimalValue creates a decimal configuration value
func NewDecimalValue(value float64) ConfigValue {
	return ConfigValue{
		Raw:      value,
		DataType: DataTypeDecimal,
	}
}

// NewJSONValue creates a JSON configuration value
func NewJSONValue(value map[string]any) ConfigValue {
	return ConfigValue{
		Raw:      value,
		DataType: DataTypeJSON,
	}
}

// FromString creates a ConfigValue from a string representation based on data type
func FromString(strValue string, dataType DataType) (ConfigValue, error) {
	if !dataType.Valid() {
		return ConfigValue{}, ErrInvalidDataType
	}

	var value any
	var err error

	switch dataType {
	case DataTypeString:
		value = strValue

	case DataTypeInteger:
		value, err = strconv.Atoi(strValue)
		if err != nil {
			return ConfigValue{}, fmt.Errorf("invalid integer value '%s': %w", strValue, err)
		}

	case DataTypeBoolean:
		value, err = strconv.ParseBool(strValue)
		if err != nil {
			return ConfigValue{}, fmt.Errorf("invalid boolean value '%s': %w", strValue, err)
		}

	case DataTypeDecimal:
		value, err = strconv.ParseFloat(strValue, 64)
		if err != nil {
			return ConfigValue{}, fmt.Errorf("invalid decimal value '%s': %w", strValue, err)
		}

	case DataTypeJSON:
		var jsonValue map[string]any
		err = json.Unmarshal([]byte(strValue), &jsonValue)
		if err != nil {
			return ConfigValue{}, fmt.Errorf("invalid JSON value '%s': %w", strValue, err)
		}
		value = jsonValue

	default:
		return ConfigValue{}, ErrInvalidDataType
	}

	return ConfigValue{
		Raw:      value,
		DataType: dataType,
	}, nil
}

// IsValidForType checks if the value is compatible with the specified data type
func (cv ConfigValue) IsValidForType(dataType DataType) bool {
	return isValidForType(cv.Raw, dataType)
}

// AsString returns the value as a string
func (cv ConfigValue) AsString() (string, error) {
	if cv.DataType != DataTypeString {
		return "", ErrWrongDataType
	}
	str, ok := cv.Raw.(string)
	if !ok {
		return "", ErrValueNotString
	}
	return str, nil
}

// AsInt returns the value as an integer
func (cv ConfigValue) AsInt() (int, error) {
	if cv.DataType != DataTypeInteger {
		return 0, ErrWrongDataType
	}
	intVal, ok := cv.Raw.(int)
	if !ok {
		return 0, ErrValueNotInteger
	}
	return intVal, nil
}

// AsBool returns the value as a boolean
func (cv ConfigValue) AsBool() (bool, error) {
	if cv.DataType != DataTypeBoolean {
		return false, ErrWrongDataType
	}
	boolVal, ok := cv.Raw.(bool)
	if !ok {
		return false, ErrValueNotBoolean
	}
	return boolVal, nil
}

// AsDecimal returns the value as a float64
func (cv ConfigValue) AsDecimal() (float64, error) {
	if cv.DataType != DataTypeDecimal {
		return 0, ErrWrongDataType
	}
	floatVal, ok := cv.Raw.(float64)
	if !ok {
		return 0, ErrValueNotDecimal
	}
	return floatVal, nil
}

// AsJSON returns the value as a JSON object
func (cv ConfigValue) AsJSON() (map[string]any, error) {
	if cv.DataType != DataTypeJSON {
		return nil, ErrWrongDataType
	}
	jsonVal, ok := cv.Raw.(map[string]any)
	if !ok {
		return nil, ErrValueNotJSON
	}
	return jsonVal, nil
}

// ToString returns the string representation of the value
func (cv ConfigValue) ToString() string {
	switch cv.DataType {
	case DataTypeString:
		if str, ok := cv.Raw.(string); ok {
			return str
		}

	case DataTypeInteger:
		if intVal, ok := cv.Raw.(int); ok {
			return strconv.Itoa(intVal)
		}

	case DataTypeBoolean:
		if boolVal, ok := cv.Raw.(bool); ok {
			return strconv.FormatBool(boolVal)
		}

	case DataTypeDecimal:
		if floatVal, ok := cv.Raw.(float64); ok {
			return strconv.FormatFloat(floatVal, 'f', -1, 64)
		}

	case DataTypeJSON:
		if jsonVal, ok := cv.Raw.(map[string]any); ok {
			jsonBytes, err := json.Marshal(jsonVal)
			if err == nil {
				return string(jsonBytes)
			}
		}
	}

	return fmt.Sprintf("%v", cv.Raw)
}

// ToJSON returns the JSON representation of the value
func (cv ConfigValue) ToJSON() ([]byte, error) {
	return json.Marshal(cv.Raw)
}

// Equals compares two configuration values for equality
func (cv ConfigValue) Equals(other ConfigValue) bool {
	if cv.DataType != other.DataType {
		return false
	}

	return reflect.DeepEqual(cv.Raw, other.Raw)
}

// IsEmpty returns true if the value is considered empty
func (cv ConfigValue) IsEmpty() bool {
	switch cv.DataType {
	case DataTypeString:
		if str, ok := cv.Raw.(string); ok {
			return str == ""
		}

	case DataTypeInteger:
		if intVal, ok := cv.Raw.(int); ok {
			return intVal == 0
		}

	case DataTypeBoolean:
		if boolVal, ok := cv.Raw.(bool); ok {
			return !boolVal
		}

	case DataTypeDecimal:
		if floatVal, ok := cv.Raw.(float64); ok {
			return floatVal == 0.0
		}

	case DataTypeJSON:
		if jsonVal, ok := cv.Raw.(map[string]any); ok {
			return len(jsonVal) == 0
		}
	}

	return cv.Raw == nil
}

// Clone creates a deep copy of the configuration value
func (cv ConfigValue) Clone() ConfigValue {
	var clonedRaw any

	switch cv.DataType {
	case DataTypeJSON:
		if jsonVal, ok := cv.Raw.(map[string]any); ok {
			cloned := make(map[string]any)
			maps.Copy(cloned, jsonVal)
			clonedRaw = cloned
		} else {
			clonedRaw = cv.Raw
		}
	default:
		clonedRaw = cv.Raw
	}

	return ConfigValue{
		Raw:      clonedRaw,
		DataType: cv.DataType,
	}
}

// Validate performs validation on the configuration value
func (cv ConfigValue) Validate() error {
	if !cv.DataType.Valid() {
		return ErrInvalidDataType
	}

	if !cv.IsValidForType(cv.DataType) {
		return ErrInvalidValueForType
	}

	return nil
}

// String returns a string representation of the configuration value
func (cv ConfigValue) String() string {
	return fmt.Sprintf("ConfigValue{Type: %s, Value: %v}", cv.DataType, cv.Raw)
}

// Helper functions

// isValidForType checks if a value is compatible with the specified data type
func isValidForType(value any, dataType DataType) bool {
	if value == nil {
		return true // Nil values are valid for all types
	}

	switch dataType {
	case DataTypeString:
		_, ok := value.(string)
		return ok

	case DataTypeInteger:
		_, ok := value.(int)
		return ok

	case DataTypeBoolean:
		_, ok := value.(bool)
		return ok

	case DataTypeDecimal:
		_, ok := value.(float64)
		return ok

	case DataTypeJSON:
		_, ok := value.(map[string]any)
		return ok

	default:
		return false
	}
}

// normalizeValue normalizes a value based on its data type
func normalizeValue(value any, dataType DataType) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch dataType {
	case DataTypeString:
		if str, ok := value.(string); ok {
			return str, nil
		}
		return fmt.Sprintf("%v", value), nil

	case DataTypeInteger:
		switch v := value.(type) {
		case int:
			return v, nil
		case int64:
			return int(v), nil
		case float64:
			return int(v), nil
		case string:
			return strconv.Atoi(v)
		default:
			return 0, ErrInvalidValueForType
		}

	case DataTypeBoolean:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		case int:
			return v != 0, nil
		default:
			return false, ErrInvalidValueForType
		}

	case DataTypeDecimal:
		switch v := value.(type) {
		case float64:
			return v, nil
		case float32:
			return float64(v), nil
		case int:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return 0.0, ErrInvalidValueForType
		}

	case DataTypeJSON:
		switch v := value.(type) {
		case map[string]any:
			return v, nil
		case string:
			var jsonVal map[string]any
			err := json.Unmarshal([]byte(v), &jsonVal)
			return jsonVal, err
		default:
			return nil, ErrInvalidValueForType
		}

	default:
		return nil, ErrInvalidDataType
	}
}
