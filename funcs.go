package leopards

import (
	"encoding/binary"
	"reflect"
)

// Bit2Uint SQL bit to int
func Bit2Uint(v []byte) uint64 {
	x := make([]byte, 8)
	j := 0
	for i := 8 - len(v); i < 8; i++ {
		x[i] = v[j]
		j++
	}
	return binary.BigEndian.Uint64(x)
}

// Key fetch value with type from map data
func Key[K comparable](data any, key any) K {
	v := reflect.ValueOf(data)

	var t K

	if v.Kind() == reflect.Map {
		x := v.MapIndex(reflect.ValueOf(key)).Interface()
		if vt, ok := x.(K); ok {
			return vt
		}
	}

	return t
}

// Pick fetch value with type from slice
func Pick[K comparable](data any, index int) K {
	v := reflect.ValueOf(data)

	var t K

	if v.Kind() == reflect.Slice {
		if index < v.Len()-1 {
			vt := v.Index(index).Interface()
			if vv, ok := vt.(K); ok {
				return vv
			}
		}
	}

	return t
}
