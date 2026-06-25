// Package convert provides safe numeric type conversion helpers.
package convert

import (
	"errors"
	"math"
)

var (
	ErrOverflow  = errors.New("value overflows destination type")
	ErrUnderflow = errors.New("value underflows destination type")
)

func Int64ToUint(v int64) (uint, error) {
	if v < 0 {
		return 0, ErrUnderflow
	}
	if uint64(v) > uint64(^uint(0)) {
		return 0, ErrOverflow
	}
	return uint(v), nil
}

func UintToInt64(v uint) (int64, error) {
	if uint64(v) > uint64(math.MaxInt64) {
		return 0, ErrOverflow
	}
	return int64(v), nil
}

func IntToUint32(v int) (uint32, error) {
	if v < 0 {
		return 0, ErrUnderflow
	}
	if uint64(v) > uint64(math.MaxUint32) {
		return 0, ErrOverflow
	}
	return uint32(v), nil
}

func Uint32ToInt(v uint32) (int, error) {
	if uint64(v) > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}
	return int(v), nil
}

func Int64ToInt32(v int64) (int32, error) {
	if v > math.MaxInt32 {
		return 0, ErrOverflow
	}
	if v < math.MinInt32 {
		return 0, ErrUnderflow
	}
	return int32(v), nil
}

func IntToInt32(v int) (int32, error) {
	if v > math.MaxInt32 {
		return 0, ErrOverflow
	}
	if v < math.MinInt32 {
		return 0, ErrUnderflow
	}
	return int32(v), nil
}

func IntToInt16(v int) (int16, error) {
	if v > math.MaxInt16 {
		return 0, ErrOverflow
	}
	if v < math.MinInt16 {
		return 0, ErrUnderflow
	}
	return int16(v), nil
}

func Uint64ToInt64(v uint64) (int64, error) {
	if v > uint64(math.MaxInt64) {
		return 0, ErrOverflow
	}
	return int64(v), nil
}

func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, ErrUnderflow
	}
	return uint64(v), nil
}
