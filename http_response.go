package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

func newHttpResponse(req *http.Request, res *http.Response) *HttpResponse {
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
	Request    *http.Request
	Response   *http.Response
	Body       []byte
	StatusCode int64
	Headers    map[string]string
}

func (hr *HttpResponse) readBody() {
	if hr.Body != nil {
		return
	}

	bytes, _ := io.ReadAll(hr.Response.Body)
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
