package internal

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"sort"
	"strings"
)

func Assert(cond bool, msg string, i ...interface{}) {
	if cond {
		return
	}
	panic("Assertion failed: " + fmt.Sprintf(msg, i...))
}

func FileUriPath(u *url.URL) string {
	path := u.Path
	if u.Host == "." && u.Path[0] == os.PathSeparator {
		path = path[1:]
	}
	return path
}

func CoerceToLikelyGoType(v interface{}) interface{} {
	if f, ok := v.(float64); ok {
		isIntegral := f == float64(int64(f))
		if isIntegral {
			return int(f)
		}
	}
	return v
}

func MinInt(i1, i2 int) int {
	if i1 <= i2 {
		return i1
	}
	return i2
}

func MinIntPtr(i1, i2 *int) *int {
	if i1 == nil {
		return i2
	} else if i2 == nil {
		return i1
	}
	x := MinInt(*i1, *i2)
	return &x
}

func MaxInt(i1, i2 int) int {
	if i1 >= i2 {
		return i1
	}
	return i2
}

func MaxIntPtr(i1, i2 *int) *int {
	if i1 == nil {
		return i2
	} else if i2 == nil {
		return i1
	}
	x := MaxInt(*i1, *i2)
	return &x
}

func MinFloat64(i1, i2 float64) float64 {
	return math.Min(i1, i2)
}

func MinFloat64Ptr(i1, i2 *float64) *float64 {
	if i1 == nil {
		return i2
	} else if i2 == nil {
		return i1
	}
	x := MinFloat64(*i1, *i2)
	return &x
}

func MaxFloat64(i1, i2 float64) float64 {
	return math.Max(i1, i2)
}

func MaxFloat64Ptr(i1, i2 *float64) *float64 {
	if i1 == nil {
		return i2
	} else if i2 == nil {
		return i1
	}
	x := MaxFloat64(*i1, *i2)
	return &x
}

func UniqueSortedStrings(x []string) []string {
	accum := make(map[string]struct{}, len(x))
	for _, i := range x {
		if _, ok := accum[i]; !ok {
			accum[i] = struct{}{}
		}
	}
	uniq := make([]string, 0, len(accum))
	for i := range accum {
		uniq = append(uniq, i)
	}
	sort.Strings(uniq)
	return uniq
}

func CompactStrings(x ...*string) []string {
	r := make([]string, 0, len(x))
	for _, s := range x {
		if s != nil {
			r = append(r, *s)
		}
	}
	return r
}

func SliceIToStr(x interface{}) []string {
	tslice, ok := x.([]string)
	if ok {
		return tslice
	}
	islice := x.([]interface{})
	r := make([]string, len(islice))
	for i, o := range islice {
		r[i] = o.(string)
	}
	return r
}

func SliceIToInt(x interface{}) []int {
	tslice, ok := x.([]int)
	if ok {
		return tslice
	}
	islice := x.([]interface{})
	r := make([]int, len(islice))
	for i, o := range islice {
		r[i] = o.(int)
	}
	return r
}

func IntPtrToFloatPtr(i *int) *float64 {
	if i == nil {
		return nil
	}
	f := float64(*i)
	return &f
}

func LowercaseKeys(m map[string]interface{}) map[string]interface{} {
	r := make(map[string]interface{}, len(m))
	for k, v := range m {
		r[strings.ToLower(k)] = v
	}
	return r
}

func UrlValuesToMap(values url.Values) map[string]interface{} {
	r := make(map[string]interface{}, len(values))
	for k, v := range values {
		if len(v) == 1 {
			r[k] = v[0]
		} else {
			varr := make([]interface{}, len(v))
			for i, a := range v {
				varr[i] = a
			}
			r[k] = varr
		}
	}
	return r
}

func ToPlainMap(s map[string]interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var r map[string]interface{}
	return r, json.Unmarshal(b, &r)
}

func LastString(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}
