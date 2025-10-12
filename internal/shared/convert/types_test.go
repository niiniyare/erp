package convert_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/niiniyare/erp/internal/shared/convert"
)

func TestInt64ToUint(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		u, err := convert.Int64ToUint(123)
		require.NoError(t, err)
		require.Equal(t, uint(123), u)
	})
	t.Run("underflow (negative)", func(t *testing.T) {
		_, err := convert.Int64ToUint(-1)
		require.ErrorIs(t, err, convert.ErrUnderflow)
	})
	t.Run("overflow only on 32-bit uint", func(t *testing.T) {
		// Overflow exists if max(uint) == 2^32-1
		if ^uint(0) == uint(math.MaxUint32) {
			// Pick a value > MaxUint32 but within int64
			v := int64(math.MaxUint32) + 1
			_, err := convert.Int64ToUint(v)
			require.ErrorIs(t, err, convert.ErrOverflow)
		}
	})
}

func TestUintToInt64(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		i, err := convert.UintToInt64(123)
		require.NoError(t, err)
		require.Equal(t, int64(123), i)
	})
	t.Run("overflow if uint > MaxInt64 (64-bit+)", func(t *testing.T) {
		if ^uint(0) > uint(math.MaxInt64) {
			_, err := convert.UintToInt64(uint(math.MaxInt64) + 1)
			require.ErrorIs(t, err, convert.ErrOverflow)
		}
	})
}

func TestIntToUint32(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		u, err := convert.IntToUint32(123)
		require.NoError(t, err)
		require.Equal(t, uint32(123), u)
	})
	t.Run("underflow (negative)", func(t *testing.T) {
		_, err := convert.IntToUint32(-1)
		require.ErrorIs(t, err, convert.ErrUnderflow)
	})
	t.Run("overflow only when int can exceed MaxUint32 (64-bit)", func(t *testing.T) {
		if uint64(math.MaxInt) > uint64(math.MaxUint32) {
			val := int(math.MaxUint32) + 1
			_, err := convert.IntToUint32(val)
			require.ErrorIs(t, err, convert.ErrOverflow)
		}
	})
}

func TestUint32ToInt(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		i, err := convert.Uint32ToInt(123)
		require.NoError(t, err)
		require.Equal(t, int(123), i)
	})
	t.Run("overflow only when MaxUint32 > MaxInt (32-bit)", func(t *testing.T) {
		if uint64(math.MaxUint32) > uint64(math.MaxInt) {
			_, err := convert.Uint32ToInt(math.MaxUint32)
			require.ErrorIs(t, err, convert.ErrOverflow)
		}
	})
}

func TestInt64ToInt32(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		i, err := convert.Int64ToInt32(123)
		require.NoError(t, err)
		require.Equal(t, int32(123), i)
	})
	t.Run("overflow", func(t *testing.T) {
		_, err := convert.Int64ToInt32(math.MaxInt64)
		require.ErrorIs(t, err, convert.ErrOverflow)
	})
	t.Run("underflow", func(t *testing.T) {
		_, err := convert.Int64ToInt32(math.MinInt64)
		require.ErrorIs(t, err, convert.ErrUnderflow)
	})
}

func TestIntToInt32(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		i, err := convert.IntToInt32(50)
		require.NoError(t, err)
		require.Equal(t, int32(50), i)
	})
	t.Run("overflow", func(t *testing.T) {
		if math.MaxInt > math.MaxInt32 {
			_, err := convert.IntToInt32(math.MaxInt32 + 1)
			require.ErrorIs(t, err, convert.ErrOverflow)
		}
	})
	t.Run("underflow", func(t *testing.T) {
		if math.MinInt < math.MinInt32 {
			_, err := convert.IntToInt32(math.MinInt32 - 1)
			require.ErrorIs(t, err, convert.ErrUnderflow)
		}
	})
}

func TestUint64ToInt64(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		i, err := convert.Uint64ToInt64(100)
		require.NoError(t, err)
		require.Equal(t, int64(100), i)
	})
	t.Run("overflow", func(t *testing.T) {
		_, err := convert.Uint64ToInt64(uint64(math.MaxInt64) + 1)
		require.ErrorIs(t, err, convert.ErrOverflow)
	})
}

func TestInt64ToUint64(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		u, err := convert.Int64ToUint64(123)
		require.NoError(t, err)
		require.Equal(t, uint64(123), u)
	})
	t.Run("underflow", func(t *testing.T) {
		_, err := convert.Int64ToUint64(-10)
		require.ErrorIs(t, err, convert.ErrUnderflow)
	})
}

// package convert_test
//
// import (
// 	"math"
// 	"testing"
//
// 	"github.com/stretchr/testify/require"
//
// 	"github.com/niiniyare/erp/internal/shared/convert"
// )
//
// func TestInt64ToUint(t *testing.T) {
// 	t.Run("valid conversion", func(t *testing.T) {
// 		u, err := convert.Int64ToUint(123)
// 		require.NoError(t, err)
// 		require.Equal(t, uint(123), u)
// 	})
//
// 	t.Run("negative value", func(t *testing.T) {
// 		_, err := convert.Int64ToUint(-1)
// 		require.ErrorIs(t, err, convert.ErrUnderflow)
// 	})
// }
//
// func TestUintToInt64(t *testing.T) {
// 	t.Run("valid conversion", func(t *testing.T) {
// 		i, err := convert.UintToInt64(123)
// 		require.NoError(t, err)
// 		require.Equal(t, int64(123), i)
// 	})
//
// 	t.Run("overflow value", func(t *testing.T) {
// 		// On 64-bit machines, math.MaxUint > math.MaxInt64
// 		if ^uint(0) > math.MaxInt64 {
// 			_, err := convert.UintToInt64(uint(math.MaxInt64) + 1)
// 			require.ErrorIs(t, err, convert.ErrOverflow)
// 		}
// 	})
// }
//
// func TestIntToUint32(t *testing.T) {
// 	t.Run("valid conversion", func(t *testing.T) {
// 		u, err := convert.IntToUint32(123)
// 		require.NoError(t, err)
// 		require.Equal(t, uint32(123), u)
// 	})
//
// 	t.Run("negative value", func(t *testing.T) {
// 		_, err := convert.IntToUint32(-1)
// 		require.ErrorIs(t, err, convert.ErrUnderflow)
// 	})
//
// 	t.Run("overflow value", func(t *testing.T) {
// 		_, err := convert.IntToUint32(math.MaxInt32 + 1)
// 		require.ErrorIs(t, err, convert.ErrOverflow)
// 	})
// }
//
// func TestUint32ToInt(t *testing.T) {
// 	t.Run("valid conversion", func(t *testing.T) {
// 		i, err := convert.Uint32ToInt(123)
// 		require.NoError(t, err)
// 		require.Equal(t, int(123), i)
// 	})
//
// 	t.Run("overflow value", func(t *testing.T) {
// 		if uint64(math.MaxUint32) > uint64(math.MaxInt) {
// 			_, err := convert.Uint32ToInt(math.MaxUint32)
// 			require.ErrorIs(t, err, convert.ErrOverflow)
// 		}
// 	})
// }
