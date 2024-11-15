package httpclient

import (
	"fmt"
	"reflect"
)

func AnyToString(a any) string {
	return fmt.Sprint(a)
}

func BoolToString(b bool) string {
	if b {
		return "true"
	}

	return "false"
}

func BoolToNumber(b bool) int {
	if b {
		return 1
	}

	return 0
}

func isSlice(a any) bool {
	varType := reflect.TypeOf(a).Kind().String()

	return varType == "slice"
}
