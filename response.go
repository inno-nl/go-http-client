package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

func newResponse(req *Request, res *http.Response) *Response {
	r := &Response{}

	r.StatusCode = res.StatusCode

	r.Headers = make(map[string]string)
	for k, v := range res.Header {
		if len(v) == 0 {
			continue
		}

		r.Headers[k] = strings.Join(v, ", ")
	}

	r.Request = req
	r.Response = res

	return r
}

type Response struct {
	Request    *Request
	Response   *http.Response
	body       []byte
	StatusCode int
	Headers    map[string]string
}

func (r *Response) readBody() {
	if r.body != nil {
		return
	}

	bytes, _ := io.ReadAll(r.Response.Body)
	r.body = bytes
}

func (r *Response) Bytes() []byte {
	r.readBody()

	return r.body
}

func (r *Response) String() string {
	r.readBody()

	return string(r.body)
}

func (r *Response) Json(serializable any) error {
	r.readBody()

	return json.Unmarshal(r.body, serializable)
}

func (r *Response) Xml(serializable any) error {
	r.readBody()

	return xml.Unmarshal(r.body, serializable)
}

func (r *Response) Success() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

func (r *Response) Retry() (*Response, error) {
	return r.Request.Send()
}
