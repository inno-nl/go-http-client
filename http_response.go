package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

type HttpResponse struct {
	readerIsRead bool
	readCloser   io.ReadCloser
	body         string

	StatusCode int64
	Headers    map[string]string
}

func (hr *HttpResponse) readBodyAsBytes(saveAsString bool) []byte {
	if hr.readerIsRead {
		return []byte(hr.body)
	}

	bytes := []byte{}
	_, _ = hr.readCloser.Read(bytes)

	if saveAsString && !hr.readerIsRead {
		hr.body = string(bytes)
	}

	hr.readerIsRead = true

	return bytes
}

func (hr *HttpResponse) String(saveAsString bool) string {
	_ = hr.readBodyAsBytes(saveAsString)

	return hr.body
}

func (hr *HttpResponse) Json(sliceOrMapOrStruct *any, saveAsString bool) error {
	bytes := hr.readBodyAsBytes(saveAsString)

	return json.Unmarshal(bytes, sliceOrMapOrStruct)
}

func (hr *HttpResponse) Xml(sliceOrMapOrStruct *any, saveAsString bool) error {
	bytes := hr.readBodyAsBytes(saveAsString)

	return xml.Unmarshal(bytes, sliceOrMapOrStruct)
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

	hr.readCloser = resp.Body

	return hr
}
