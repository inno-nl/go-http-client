package httpclient

import (
	"encoding/json"
)

// Assinine or "Associative" array as produced by PHP,
// with values either in a normal unassociative array,
// or a JSON object with index numbers as keys.
//
// The expected slice [1,2,3] could be encoded as either:
//
//	[1, 2, 3]
//	{"0":1, "1":2, "2":3}
type AssArray []any

func (a *AssArray) UnmarshalJSON(v []byte) (err error) {
	var arr []any
	if err = json.Unmarshal(v, &arr); err == nil {
		*a = arr
		return
	}

	var obj map[int]any
	if err = json.Unmarshal(v, &obj); err == nil {
		*a = make(AssArray, len(obj))
		for i, v := range obj {
			(*a)[i] = v
		}
	}
	return
}
