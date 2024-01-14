package httpclient

import (
	"encoding/json"
	"net/http"
	"strings"
)

type HttpResponse struct {
	StatusCode int64
	Headers    map[string]string
	Payload    string
}

func (hr *HttpResponse) Unmarshal(sliceOrMapOrStruct *any) error {
	return json.Unmarshal([]byte(hr.Payload), sliceOrMapOrStruct)
}

func newHttpResponse(resp *http.Response) *HttpResponse {
	hr := &HttpResponse{}

	hr.StatusCode = int64(resp.StatusCode)

	for hK, hV := range resp.Header {
		if len(hV) == 0 {
			hr.Headers[hK] = hV[0]
		}

		hr.Headers[hK] = strings.Join(hV, ", ")
	}

	buffer := []byte{}
	_, _ = resp.Body.Read(buffer)
	hr.Payload = string(buffer)

	return hr
}
