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
	baseUrl    string
	path       string
	method     string `default:"GET"`
	parameters map[string][]string
	headers    map[string]string
	body       string
	timeout    float64 `default:"60"`
	retryCount int64   `default:"0"`
}

func (hr *HttpRequest) BaseUrl(requestUrl string) *HttpRequest {
	hr.baseUrl = strings.TrimRight(requestUrl, "/")

	return hr
}

func (hr *HttpRequest) Path(requestUrl string) *HttpRequest {
	hr.path = strings.TrimLeft(requestUrl, "/")

	return hr
}

func (hr *HttpRequest) OverrideUrl(requestUrl string) *HttpRequest {
	parsedUrl, _ := url.Parse(hr.extractParametersFromUrl(requestUrl))

	hr.baseUrl = fmt.Sprintf(
		"%s://%s",
		parsedUrl.Scheme,
		parsedUrl.Host,
	)
	hr.path = parsedUrl.Path

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

func (hr *HttpRequest) RetryCount(retryCount int64) *HttpRequest {
	hr.retryCount = retryCount

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

func (hr *HttpRequest) Get() (response *HttpResponse, err error) {
	hr.method = GET

	return hr.execute()
}

func (hr *HttpRequest) Post() (response *HttpResponse, err error) {
	hr.method = POST

	return hr.execute()
}

func (hr *HttpRequest) Put() (response *HttpResponse, err error) {
	hr.method = PUT

	return hr.execute()
}

func (hr *HttpRequest) Patch() (response *HttpResponse, err error) {
	hr.method = PATCH

	return hr.execute()
}

func (hr *HttpRequest) Delete() (response *HttpResponse, err error) {
	hr.method = DELETE

	return hr.execute()
}

func (hr *HttpRequest) CustomMethod(method string) (response *HttpResponse, err error) {
	hr.method = method

	return hr.execute()
}

func (hr *HttpRequest) execute() (response *HttpResponse, err error) {
	hc := &http.Client{
		Timeout: time.Duration(hr.timeout) * time.Second,
	}

	req, err := http.NewRequest(
		hr.method,
		hr.generateUrl(),
		hr.parseBody(),
	)
	if err != nil {
		return
	}

	for k, v := range hr.headers {
		req.Header.Set(k, v)
	}

	// Tries the request atleast once
	var res *http.Response
	for i := 0; i < int(hr.retryCount+1); i++ {
		res, err = hc.Do(req)
		if err == nil {
			break
		}
	}

	// If no response was given the request has failed
	if res == nil {
		return
	}

	response = newHttpResponse(res)
	return
}

func (hr *HttpRequest) generateUrl() string {
	fullUrl := fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(hr.baseUrl, "/"),
		strings.TrimLeft(hr.path, "/"),
	)

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
		fullUrl,
		strings.Join(parameters, "&"),
	)
}

func (hr *HttpRequest) parseBody() io.Reader {
	if hr.method == GET {
		return nil
	}

	if hr.body != "" {
		return strings.NewReader(hr.body)
	}

	return nil
}

func (hr *HttpRequest) extractParametersFromUrl(requestUrl string) string {
	if !strings.Contains("?", requestUrl) {
		return strings.TrimRight(requestUrl, "/")
	}

	parts := strings.Split(requestUrl, "?")
	urlString := parts[0]
	queryString := parts[1]

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

	return strings.TrimRight(urlString, "/")
}
