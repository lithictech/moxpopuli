package jsontype

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type JsonType string

//goland:noinspection GoSnakeCaseUsage
const (
	T_INTEGER JsonType = "integer"
	T_STRING  JsonType = "string"
	T_BOOLEAN JsonType = "boolean"
	T_NUMBER  JsonType = "number"
	T_OBJECT  JsonType = "object"
	T_ARRAY   JsonType = "array"
	T_NOTYPE  JsonType = ""
)

// Sniff returns the JsonType for the given value.
// Value can be a Go primitive or a special type like json.Number.
// Nil returns T_NOTYPE.
func Sniff(value interface{}) JsonType {
	if value == nil {
		return T_NOTYPE
	}
	if j, ok := value.(json.Number); ok {
		value = unwrapJson(j)
	}
	v := reflect.TypeOf(value)
	switch v.Kind() {
	case reflect.String:
		return T_STRING
	case reflect.Bool:
		return T_BOOLEAN
	case reflect.Int:
		return T_INTEGER
	case reflect.Float64:
		return T_NUMBER
	case reflect.Map:
		return T_OBJECT
	case reflect.Slice:
		return T_ARRAY
	default:
		panic(fmt.Sprintf("could not sniff type from %v", value))
	}
}

func unwrapJson(n json.Number) interface{} {
	i, err := n.Int64()
	if err == nil {
		return int(i)
	}
	f, err := n.Float64()
	if err != nil {
		panic("number is not a float or int? " + n.String())
	}
	return f
}
