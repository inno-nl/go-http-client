package httpclient

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
)

type base struct {
	logErrors   bool
	proxyUrl    *string
	baseUrl     *string
	path        *string
	method      *string
	parameters  map[string][]string
	headers     map[string]string
	contentType *string
	body        *string
	timeout     *int
	retryCount  *int
}

func (b *base) LogErrors(shouldLog bool) {
	b.logErrors = shouldLog
}

func (b *base) ProxyUrl(proxyUrl string) {
	b.proxyUrl = &proxyUrl
}

func (b *base) BaseUrl(requestUrl string) {
	baseUrl := strings.TrimRight(requestUrl, "/")

	b.baseUrl = &baseUrl
}

func (b *base) Path(requestUrl string) {
	path := strings.TrimLeft(requestUrl, "/")

	b.path = &path
}

func (b *base) FullUrl(requestUrl string) {
	parsedUrl, _ := url.Parse(b.extractParametersFromUrl(requestUrl))

	baseUrl := fmt.Sprintf(
		"%s://%s",
		parsedUrl.Scheme,
		parsedUrl.Host,
	)
	path := parsedUrl.Path

	b.baseUrl = &baseUrl
	b.path = &path
}

func (b *base) Method(method string) {
	b.method = &method
}

func (b *base) Parameter(key string, value string) {
	b.initParameters()

	_, exists := b.parameters[key]
	if !exists {
		b.parameters[key] = make([]string, 0)
	}

	b.parameters[key] = append(
		b.parameters[key],
		value,
	)
}

func (b *base) Parameters(parameters map[string]any) {
	b.initParameters()

	for k, v := range parameters {
		vType := fmt.Sprint(reflect.TypeOf(v).Kind())

		if vType == "string" {
			b.Parameter(k, v.(string))
			continue
		}

		if vType == "slice" {
			slice := v.([]string)

			for _, sv := range slice {
				b.Parameter(k, sv)
			}
		}
	}
}

func (b *base) Header(key string, value string) {
	b.initHeaders()

	b.headers[key] = value
}

func (b *base) Headers(headers map[string]string) {
	b.initHeaders()

	for k, v := range headers {
		b.Header(k, v)
	}
}

func (b *base) ContentType(contentType string) {
	b.contentType = &contentType
}

func (b *base) Body(body string) {
	b.body = &body

	if b.contentType == nil {
		b.ContentType("text/plain")
	}

	if b.method == nil {
		method := POST
		b.method = &method
	}
}

func (b *base) Json(body any) {
	bytes, _ := json.Marshal(body)
	bytesString := string(bytes)
	b.body = &bytesString

	if b.contentType == nil {
		b.ContentType("application/json")
	}

	if b.method == nil {
		method := POST
		b.method = &method
	}
}

func (b *base) Timeout(timeout int) {
	b.timeout = &timeout
}

func (b *base) RetryCount(retryCount int) {
	b.retryCount = &retryCount
}

func (b *base) BasicAuth(user string, pass string) {
	b.initHeaders()

	b.headers[AUTHORIZATION_HEADER] = fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass))),
	)
}

func (b *base) BearerAuth(token string) {
	b.initHeaders()

	b.headers[AUTHORIZATION_HEADER] = fmt.Sprintf(
		"Bearer %s",
		token,
	)
}

func (b *base) generateUrl() string {
	fullUrl := fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(*b.baseUrl, "/"),
		strings.TrimLeft(*b.path, "/"),
	)

	if len(b.parameters) == 0 {
		return fullUrl
	}

	parameters := make([]string, 0)

	for key, values := range b.parameters {
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

func (b *base) parseBody() io.Reader {
	if b.method != nil && *b.method == GET {
		return nil
	}

	if b.body != nil {
		return strings.NewReader(*b.body)
	}

	return nil
}

func (b *base) extractParametersFromUrl(requestUrl string) string {
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
			b.Parameter(key, "")
			continue
		}

		queryParamSplit := strings.Split(q, "=")
		key, _ := url.QueryUnescape(queryParamSplit[0])
		value, _ := url.QueryUnescape(queryParamSplit[1])

		b.Parameter(key, value)
	}

	return strings.TrimRight(urlString, "/")
}

func (b *base) initParameters() {
	if b.parameters == nil {
		b.parameters = make(map[string][]string)
	}
}

func (b *base) initHeaders() {
	if b.headers == nil {
		b.headers = make(map[string]string)
	}
}
