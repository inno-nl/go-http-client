package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

func newHttpResponse(req *HttpRequest, res *http.Response) *HttpResponse {
	hr := &HttpResponse{}

	hr.StatusCode = int64(res.StatusCode)

	hr.Headers = make(map[string]string)
	for k, v := range res.Header {
		if len(v) == 0 {
			continue
		}

		hr.Headers[k] = strings.Join(v, ", ")
	}

	hr.Request = req
	hr.Response = res

	return hr
}

type HttpResponse struct {
	Request    *HttpRequest
	Response   *http.Response
	body       []byte
	StatusCode int64
	Headers    map[string]string
}

func (hr *HttpResponse) readBody() {
	if hr.body != nil {
		return
	}

	bytes, _ := io.ReadAll(hr.Response.Body)
	hr.body = bytes
}

func (hr *HttpResponse) Bytes() []byte {
	hr.readBody()

	return hr.body
}

func (hr *HttpResponse) String() string {
	hr.readBody()

	return string(hr.body)
}

func (hr *HttpResponse) Json(sliceOrMapOrStruct *any) error {
	hr.readBody()

	return json.Unmarshal(hr.body, sliceOrMapOrStruct)
}

func (hr *HttpResponse) Xml(sliceOrMapOrStruct *any) error {
	hr.readBody()

	return xml.Unmarshal(hr.body, sliceOrMapOrStruct)
}

func (hr *HttpResponse) Success() bool {
	return hr.StatusCode >= 200 && hr.StatusCode < 300
}

func (hr *HttpResponse) Retry() (*HttpResponse, error) {
	return hr.Request.Execute()
}
