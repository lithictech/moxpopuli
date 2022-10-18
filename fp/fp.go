package fp

import (
	"math/rand"
)

// SampleOut returns a new slice, removing random elements from in until
// it is at most length n.
func SampleOut(in []interface{}, n int) []interface{} {
	result := make([]interface{}, len(in))
	copy(result, in)
	for len(result) > n {
		idx := rand.Intn(len(result))
		result = append(result[:idx], result[idx+1:]...)
	}
	return result
}

func Sample[T interface{}](s []T) T {
	ls := len(s)
	if ls == 0 {
		var t T
		return t
	} else if ls == 1 {
		return s[0]
	}
	return s[rand.Intn(ls-1)]
}

func Values[K comparable, V interface{}](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
