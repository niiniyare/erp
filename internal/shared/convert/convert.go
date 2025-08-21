package convert

import (
	"errors"
	"math"
)

// Canonical errors for all conversions.
var (
	ErrOverflow  = errors.New("value overflows destination type")
	ErrUnderflow = errors.New("value underflows destination type")
)

// ---------- INT64 <-> UINT ----------

// Int64ToUint converts int64 → uint safely (checks uint max on this arch).
func Int64ToUint(v int64) (uint, error) {
	if v < 0 {
		return 0, ErrUnderflow
	}
	// Max uint on this arch.
	maxUint := ^uint(0)
	if uint64(v) > uint64(maxUint) {
		return 0, ErrOverflow
	}
	return uint(v), nil
}

// UintToInt64 converts uint → int64 safely.
func UintToInt64(v uint) (int64, error) {
	if uint64(v) > uint64(math.MaxInt64) {
		return 0, ErrOverflow
	}
	return int64(v), nil
}

// ---------- INT <-> UINT32 ----------

// IntToUint32 converts int → uint32 safely.
func IntToUint32(v int) (uint32, error) {
	if v < 0 {
		return 0, ErrUnderflow
	}
	if uint64(v) > uint64(math.MaxUint32) {
		return 0, ErrOverflow
	}
	return uint32(v), nil
}

// Uint32ToInt converts uint32 → int safely.
func Uint32ToInt(v uint32) (int, error) {
	// Compare without converting to int first.
	if uint64(v) > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}
	return int(v), nil
}

// ---------- INT64 <-> INT32 ----------

// Int64ToInt32 converts int64 → int32 safely.
func Int64ToInt32(v int64) (int32, error) {
	if v > math.MaxInt32 {
		return 0, ErrOverflow
	}
	if v < math.MinInt32 {
		return 0, ErrUnderflow
	}
	return int32(v), nil
}

// IntToInt32 converts int → int32 safely.
func IntToInt32(v int) (int32, error) {
	if v > math.MaxInt32 {
		return 0, ErrOverflow
	}
	if v < math.MinInt32 {
		return 0, ErrUnderflow
	}
	return int32(v), nil
}

// ---------- UINT64 <-> INT64 ----------

// Uint64ToInt64 converts uint64 → int64 safely.
func Uint64ToInt64(v uint64) (int64, error) {
	if v > uint64(math.MaxInt64) {
		return 0, ErrOverflow
	}
	return int64(v), nil
}

// Int64ToUint64 converts int64 → uint64 safely.
func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, ErrUnderflow
	}
	return uint64(v), nil
}

// package convert
//
// import (
// 	"errors"
// 	"math"
// )
//
// // Common errors
// var (
// 	ErrOverflow  = errors.New("value overflows destination type")
// 	ErrUnderflow = errors.New("value underflows destination type")
// 	ErrNegative  = errors.New("value can not be Negative")
// )
// // ---------- INT <-> UINT32 ----------
//
// func IntToUint32(v int) (uint32, error) {
// 	if v < 0 {
// 		return 0, ErrUnderflow
// 	}
// 	if v > math.MaxUint32 {
// 		return 0, ErrOverflow
// 	}
// 	return uint32(v), nil
// }
//
// // ---------- INT <-> UINT ----------
//
// // Int64ToUint converts int64 → uint safely
// func Int64ToUint(v int64) (uint, error) {
// 	if v < 0 {
// 		return 0, ErrUnderflow
// 	}
// 	if uint64(v) > uint64(^uint(0)) {
// 		return 0, ErrOverflow
// 	}
// 	return uint(v), nil
// }
//
// // UintToInt64 converts uint → int64 safely
// func UintToInt64(v uint) (int64, error) {
// 	if v > math.MaxInt64 {
// 		return 0, ErrOverflow
// 	}
// 	return int64(v), nil
// }
//
// // Uint32ToInt converts uint32 → int safely
// func Uint32ToInt(v uint32) (int, error) {
// 	if int(v) > math.MaxInt {
// 		return 0, ErrOverflow
// 	}
// 	return int(v), nil
// }
//
// // ---------- INT <-> INT32 ----------
//
// // Int64ToInt32 converts int64 → int32 safely
// func Int64ToInt32(v int64) (int32, error) {
// 	if v > math.MaxInt32 {
// 		return 0, ErrOverflow
// 	}
// 	if v < math.MinInt32 {
// 		return 0, ErrUnderflow
// 	}
// 	return int32(v), nil
// }
//
// // IntToInt32 converts int → int32 safely
// func IntToInt32(v int) (int32, error) {
// 	if v > math.MaxInt32 {
// 		return 0, ErrOverflow
// 	}
// 	if v < math.MinInt32 {
// 		return 0, ErrUnderflow
// 	}
// 	return int32(v), nil
// }
//
// // ---------- UINT64 <-> INT64 ----------
//
// // Uint64ToInt64 converts uint64 → int64 safely
// func Uint64ToInt64(v uint64) (int64, error) {
// 	if v > math.MaxInt64 {
// 		return 0, ErrOverflow
// 	}
// 	return int64(v), nil
// }
//
// // Int64ToUint64 converts int64 → uint64 safely
// func Int64ToUint64(v int64) (uint64, error) {
// 	if v < 0 {
// 		return 0, ErrUnderflow
// 	}
// 	return uint64(v), nil
// }
