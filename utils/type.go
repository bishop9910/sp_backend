package utils

import "reflect"

func IsType[T any](v any) bool {
	var zero T
	return reflect.TypeOf(v) == reflect.TypeOf(zero)
}

func AnyToStringSafe(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func AnyToIntSafe(v any) int {
	i, ok := v.(int)
	if !ok {
		return 0
	}
	return i
}
