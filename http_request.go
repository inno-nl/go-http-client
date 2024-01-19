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
	parameters map[string][]string
	headers    map[string]string
	body       string
	timeout    float64 `default:"60"`
}

func (hr *HttpRequest) parseUrl() string {
	parameters := make([]string, 0)

	for key, values := range hr.parameters {
		for _, v := range values {
			parameters = append(parameters, fmt.Sprintf(
				"%s=%s",
				url.QueryEscape(key),
				url.QueryEscape(v),
			))
		}
	}

	return fmt.Sprintf(
		"%s?%s",
		hr.url,
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

func (hr *HttpRequest) ExecuteRaw() (response *HttpResponse, err error) {
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

func (hr *HttpRequest) Url(requestUrl string) *HttpRequest {
	if !strings.Contains("?", requestUrl) {
		hr.url = requestUrl
		return hr
	}

	parts := strings.Split(hr.url, "?")
	baseUrl := parts[0]
	queryString := parts[1]

	hr.url = baseUrl

	queryStringSplit := strings.Split(queryString, "&")
	for _, q := range queryStringSplit {
		if !strings.Contains(q, "=") {
			key, _ := url.QueryUnescape(q)
			hr.Parameter(key, "")
			continue
		}

		queryParamSplit := strings.Split(q, "=")
		key, _ := url.QueryUnescape(queryParamSplit[0])
		value, _ := url.QueryUnescape(queryParamSplit[1])

		hr.Parameter(key, value)
	}

	return hr
}

func (hr *HttpRequest) Method(method string) *HttpRequest {
	hr.method = method

	return hr
}

func (hr *HttpRequest) Parameter(key string, value string) *HttpRequest {
	_, exists := hr.parameters[key]
	if !exists {
		hr.parameters[key] = make([]string, 0)
	}

	hr.parameters[key] = append(
		hr.parameters[key],
		value,
	)

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

func (hr *HttpRequest) Timeout(timeout float64) *HttpRequest {
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
		"Bearer %s",
		token,
	)

	return hr
}
