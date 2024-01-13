package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpRequest struct {
	url        string
	method     string
	parameters []Parameter
	headers    []Header
	body       string
	timeout    int64 `default:"60"`
}

func (hr *HttpRequest) parseUrl() string {
	if !strings.Contains("?", hr.url) {
		return hr.url
	}

	baseUrl := strings.Split(hr.url, "?")[0]

	parameters := make([]string, 0)

	for _, p := range hr.parameters {
		parameters = append(parameters, fmt.Sprintf(
			"%s=%s",
			url.QueryEscape(p.Key),
			url.QueryEscape(p.Value),
		))
	}

	return fmt.Sprintf(
		"%s?%s",
		baseUrl,
		strings.Join(parameters, "&"),
	)
}

func (hr *HttpRequest) parseBody() io.Reader {
	if hr.body != "" {
		return strings.NewReader(hr.body)
	}

	return nil
}

func (hr *HttpRequest) Execute() (response *HttpResponse, err error) {
	hc := &http.Client{
		Timeout: time.Duration(hr.timeout),
	}

	req, err := http.NewRequest(
		hr.method,
		hr.parseUrl(),
		hr.parseBody(),
	)
	if err != nil {
		return
	}

	for _, h := range hr.headers {
		req.Header.Set(
			h.Key,
			h.Value,
		)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return
	}

	response = newHttpResponse(resp)
	return
}

func (hr *HttpRequest) Url(url string) *HttpRequest {
	hr.url = url

	return hr
}

func (hr *HttpRequest) Method(method string) *HttpRequest {
	hr.method = method

	return hr
}

func (hr *HttpRequest) Parameter(key string, value string) *HttpRequest {
	hr.parameters = append(hr.parameters, Parameter{
		Key:   key,
		Value: value,
	})

	return hr
}

func (hr *HttpRequest) Header(key string, value string) *HttpRequest {
	hr.headers = append(hr.headers, Header{
		Key:   key,
		Value: value,
	})

	return hr
}

func (hr *HttpRequest) Body(body string) *HttpRequest {
	hr.body = body

	return hr
}

func (hr *HttpRequest) Json(body any) *HttpRequest {
	hr.body = hr.encodeToJson(body)

	return hr
}

func (hr *HttpRequest) Timeout(timeout int64) *HttpRequest {
	hr.timeout = timeout

	return hr
}

func (hr *HttpRequest) encodeToJson(data any) string {
	bytes, _ := json.Marshal(data)

	return string(bytes)
}
