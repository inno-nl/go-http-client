package httpclient

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type HttpRequest struct {
	baseUrl     string
	path        string
	method      string
	parameters  map[string][]string
	headers     map[string]string
	contentType string
	body        string
	timeout     float64 `default:"60"`
	retryCount  int64   `default:"0"`
}

func (hr *HttpRequest) BaseUrl(requestUrl string) *HttpRequest {
	hr.baseUrl = strings.TrimRight(requestUrl, "/")

	return hr
}

func (hr *HttpRequest) Path(requestUrl string) *HttpRequest {
	hr.path = strings.TrimLeft(requestUrl, "/")

	return hr
}

func (hr *HttpRequest) Method(method string) *HttpRequest {
	hr.method = method

	return hr
}

func (hr *HttpRequest) FullUrl(requestUrl string) *HttpRequest {
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
	if hr.parameters == nil {
		hr.parameters = make(map[string][]string)
	}

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

func (hr *HttpRequest) Parameters(parameters map[string]any) *HttpRequest {
	if hr.parameters == nil {
		hr.parameters = make(map[string][]string)
	}

	for k, v := range parameters {
		vType := fmt.Sprint(reflect.TypeOf(v).Kind())

		if vType == "string" {
			hr.Parameter(k, v.(string))
			continue
		}

		if vType == "slice" {
			slice := v.([]string)

			for _, sv := range slice {
				hr.Parameter(k, sv)
			}
		}
	}

	return hr
}

func (hr *HttpRequest) Header(key string, value string) *HttpRequest {
	if hr.headers == nil {
		hr.headers = make(map[string]string)
	}

	hr.headers[key] = value

	return hr
}

func (hr *HttpRequest) Headers(headers map[string]string) *HttpRequest {
	if hr.headers == nil {
		hr.headers = make(map[string]string)
	}

	for k, v := range headers {
		hr.Header(k, v)
	}

	return hr
}

func (hr *HttpRequest) ContentType(contentType string) *HttpRequest {
	hr.contentType = contentType

	return hr
}

func (hr *HttpRequest) Body(body string) *HttpRequest {
	hr.body = body

	if hr.contentType == "" {
		hr.ContentType("text/plain")
	}

	if hr.method == "" {
		hr.method = POST
	}

	return hr
}

func (hr *HttpRequest) Json(body any) *HttpRequest {
	bytes, _ := json.Marshal(body)
	hr.body = string(bytes)

	if hr.contentType == "" {
		hr.ContentType("application/json")
	}

	if hr.method == "" {
		hr.method = POST
	}

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

func (hr *HttpRequest) Execute() (response *HttpResponse, err error) {
	hc := &http.Client{
		Timeout: time.Duration(hr.timeout) * time.Second,
	}

	if hr.method == "" {
		hr.method = GET
	}

	req, err := http.NewRequest(
		hr.method,
		hr.generateUrl(),
		hr.parseBody(),
	)
	if err != nil {
		return
	}

	req.Header.Set(
		"user-agent",
		"inno-go-http-client",
	)

	if hr.contentType != "" {
		req.Header.Set("Content-type", hr.contentType)
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

	response = newHttpResponse(hr, res)
	return
}

func (hr *HttpRequest) generateUrl() string {
	fullUrl := fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(hr.baseUrl, "/"),
		strings.TrimLeft(hr.path, "/"),
	)

	if len(hr.parameters) == 0 {
		return fullUrl
	}

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
