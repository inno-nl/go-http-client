package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

type HttpResponse struct {
	body []byte

	StatusCode int64
	Headers    map[string]string
}

func (hr *HttpResponse) String() string {
	return string(hr.body)
}

func (hr *HttpResponse) Json(sliceOrMapOrStruct *any) error {
	return json.Unmarshal(hr.body, sliceOrMapOrStruct)
}

func (hr *HttpResponse) Xml(sliceOrMapOrStruct *any, saveAsString bool) error {
	return xml.Unmarshal(hr.body, sliceOrMapOrStruct)
}

func newHttpResponse(resp *http.Response) *HttpResponse {
	hr := &HttpResponse{}

	hr.StatusCode = int64(resp.StatusCode)

	for hK, hV := range resp.Header {
		if len(hV) == 0 {
			continue
		}

		hr.Headers[hK] = strings.Join(hV, ", ")
	}

	bytes := []byte{}
	_, _ = resp.Body.Read(bytes)
	hr.body = bytes

	return hr
}
