package httpclient

import (
	"testing"

	"encoding/json"
	"reflect"
)

func TestAssArray(t *testing.T) {
	cases := map[string]AssArray{
		`"2, 3, 4"`:                  nil,

		`[]`:                         {},
		`[2, 3, 4]`:                  {2., 3., 4.},
		`["un", "du", "tri"]`:        {"un", "du", "tri"},
		`["0":1]`:                    nil,

		`{"0":2, "1":3, "2":4}`:      {2., 3., 4.},
		`{"0":"un", "1":2, "2":[3]}`: {"un", 2., []any{3.}}, // mixed values
		`{"0":1, "mixed":2}`:         nil,                   // mixed keys
	}

	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			var out AssArray
			err := json.Unmarshal([]byte(in), &out)
			if want == nil {
				if err == nil {
					t.Fatalf("unmarshal did not fail: %v", out)
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if !reflect.DeepEqual(out, want) {
				t.Fatalf("unexpected output: %v instead of %v", out, want)
			}
		})
	}
}
