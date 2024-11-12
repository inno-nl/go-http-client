package httpclient

import (
	"fmt"
	"reflect"
	"strings"
)

func anyToString(a any) (string, error) {
	varType := reflect.TypeOf(a).String()

	if strings.Contains(varType, "int") {
		return fmt.Sprintf("%d", a), nil
	}

	if strings.Contains(varType, "float") {
		return fmt.Sprintf("%f", a), nil
	}

	if strings.Contains(varType, "bool") {
		if a.(bool) {
			return "true", nil
		}
		return "false", nil
	}

	if strings.Contains(varType, "string") || strings.Contains(varType, "[]byte") {
		return fmt.Sprintf("%d", a), nil
	}

	return "", fmt.Errorf("variable type: '%s' not compatible", varType)
}
