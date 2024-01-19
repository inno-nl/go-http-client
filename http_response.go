package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

type HttpResponse struct {
	Response   *http.Response
	Body       []byte
	StatusCode int64
	Headers    map[string]string
}

func (hr *HttpResponse) readBody() {
	if hr.Body != nil {
		return
	}

	bytes := []byte{}
	_, _ = hr.Response.Body.Read(bytes)
	hr.Body = bytes
}

func (hr *HttpResponse) String() string {
	hr.readBody()

	return string(hr.Body)
}

func (hr *HttpResponse) Json(sliceOrMapOrStruct *any) error {
	hr.readBody()

	return json.Unmarshal(hr.Body, sliceOrMapOrStruct)
}

func (hr *HttpResponse) Xml(sliceOrMapOrStruct *any) error {
	hr.readBody()

	return xml.Unmarshal(hr.Body, sliceOrMapOrStruct)
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

	hr.Response = resp

	return hr
}
