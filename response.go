package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

func newResponse(req *Request, res *http.Response) *Response {
	hr := &Response{}

	hr.StatusCode = res.StatusCode

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

type Response struct {
	Request    *Request
	Response   *http.Response
	body       []byte
	StatusCode int
	Headers    map[string]string
}

func (hr *Response) readBody() {
	if hr.body != nil {
		return
	}

	bytes, _ := io.ReadAll(hr.Response.Body)
	hr.body = bytes
}

func (hr *Response) Bytes() []byte {
	hr.readBody()

	return hr.body
}

func (hr *Response) String() string {
	hr.readBody()

	return string(hr.body)
}

func (hr *Response) Json(serializable any) error {
	hr.readBody()

	return json.Unmarshal(hr.body, serializable)
}

func (hr *Response) Xml(serializable any) error {
	hr.readBody()

	return xml.Unmarshal(hr.body, serializable)
}

func (hr *Response) Success() bool {
	return hr.StatusCode >= 200 && hr.StatusCode < 300
}

func (hr *Response) Retry() (*Response, error) {
	return hr.Request.Send()
}
