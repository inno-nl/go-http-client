package httpclient

import (
	"encoding/json"
	"slices"
)

// Assinine or "Associative" array as produced by PHP,
// with values either in a normal unassociative array,
// or a JSON object with index numbers as keys.
//
// The expected slice [1,2,3] could be encoded as either:
//
//	[1, 2, 3]
//	{"0":1, "1":2, "2":3}
//	{"1":2, "42":3, "0":1}
type AssArray []any

func (a *AssArray) UnmarshalJSON(v []byte) (err error) {
	var arr []any
	if err = json.Unmarshal(v, &arr); err == nil {
		*a = arr
		return
	}

	var obj map[int]any
	if err = json.Unmarshal(v, &obj); err == nil {
		keys := make([]int, len(obj))
		i := 0
		for ref, _ := range obj {
			keys[i] = ref
			i++
		}
		slices.Sort(keys)

		*a = make(AssArray, len(obj))
		for i, ref := range keys {
			(*a)[i] = obj[ref]
		}
	}
	return
}
