package httpclient

import (
	"encoding/base64"
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
	parameters map[string]string
	headers    map[string]string
	body       string
	timeout    int64 `default:"60"`
}

func (hr *HttpRequest) parseUrl() string {
	if !strings.Contains("?", hr.url) {
		return hr.url
	}

	baseUrl := strings.Split(hr.url, "?")[0]

	parameters := make([]string, 0)

	for k, v := range hr.parameters {
		parameters = append(parameters, fmt.Sprintf(
			"%s=%s",
			url.QueryEscape(k),
			url.QueryEscape(v),
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
		Timeout: time.Duration(hr.timeout) * time.Second,
	}

	req, err := http.NewRequest(
		hr.method,
		hr.parseUrl(),
		hr.parseBody(),
	)
	if err != nil {
		return
	}

	for k, v := range hr.headers {
		req.Header.Set(k, v)
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
	hr.parameters[key] = value

	return hr
}

func (hr *HttpRequest) Header(key string, value string) *HttpRequest {
	hr.headers[key] = value

	return hr
}

func (hr *HttpRequest) Body(body string) *HttpRequest {
	hr.body = body
	hr.Header("Content-type", "text/plain")

	return hr
}

func (hr *HttpRequest) Json(body any) *HttpRequest {
	bytes, _ := json.Marshal(body)
	hr.body = string(bytes)
	hr.Header("Content-type", "application/json")

	return hr
}

func (hr *HttpRequest) Timeout(timeout int64) *HttpRequest {
	hr.timeout = timeout

	return hr
}

func (hr *HttpRequest) BasicAuth(user string, pass string) *HttpRequest {
	hr.headers[AUTHORIZATION_HEADER] = fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass))),
	)

	return hr
}

func (hr *HttpRequest) BearerAuth(token string) *HttpRequest {
	hr.headers[AUTHORIZATION_HEADER] = fmt.Sprintf(
		"Basic %s",
		token,
	)

	return hr
}
