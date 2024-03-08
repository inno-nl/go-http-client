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

type httpBase struct {
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

func (hb *httpBase) BaseUrl(requestUrl string) {
	baseUrl := strings.TrimRight(requestUrl, "/")

	hb.baseUrl = &baseUrl

}

func (hb *httpBase) Path(requestUrl string) {
	path := strings.TrimLeft(requestUrl, "/")

	hb.path = &path

}

func (hb *httpBase) Method(method string) {
	hb.method = &method

}

func (hb *httpBase) FullUrl(requestUrl string) {
	parsedUrl, _ := url.Parse(hb.extractParametersFromUrl(requestUrl))

	baseUrl := fmt.Sprintf(
		"%s://%s",
		parsedUrl.Scheme,
		parsedUrl.Host,
	)
	path := parsedUrl.Path

	hb.baseUrl = &baseUrl
	hb.path = &path

}

func (hb *httpBase) Parameter(key string, value string) {
	if hb.parameters == nil {
		hb.parameters = make(map[string][]string)
	}

	_, exists := hb.parameters[key]
	if !exists {
		hb.parameters[key] = make([]string, 0)
	}

	hb.parameters[key] = append(
		hb.parameters[key],
		value,
	)
}

func (hb *httpBase) Parameters(parameters map[string]any) {
	if hb.parameters == nil {
		hb.parameters = make(map[string][]string)
	}

	for k, v := range parameters {
		vType := fmt.Sprint(reflect.TypeOf(v).Kind())

		if vType == "string" {
			hb.Parameter(k, v.(string))
			continue
		}

		if vType == "slice" {
			slice := v.([]string)

			for _, sv := range slice {
				hb.Parameter(k, sv)
			}
		}
	}

}

func (hb *httpBase) Header(key string, value string) {
	if hb.headers == nil {
		hb.headers = make(map[string]string)
	}

	hb.headers[key] = value

}

func (hb *httpBase) Headers(headers map[string]string) {
	if hb.headers == nil {
		hb.headers = make(map[string]string)
	}

	for k, v := range headers {
		hb.Header(k, v)
	}

}

func (hb *httpBase) ContentType(contentType string) {
	hb.contentType = &contentType

}

func (hb *httpBase) Body(body string) {
	hb.body = &body

	if hb.contentType == nil {
		hb.ContentType("text/plain")
	}

	if hb.method == nil {
		method := POST
		hb.method = &method
	}

}

func (hb *httpBase) Json(body any) {
	bytes, _ := json.Marshal(body)
	bytesString := string(bytes)
	hb.body = &bytesString

	if hb.contentType == nil {
		hb.ContentType("application/json")
	}

	if hb.method == nil {
		method := POST
		hb.method = &method
	}
}

func (hb *httpBase) Timeout(timeout int) {
	hb.timeout = &timeout

}

func (hb *httpBase) RetryCount(retryCount int) {
	hb.retryCount = &retryCount

}

func (hb *httpBase) BasicAuth(user string, pass string) {
	hb.headers[AUTHORIZATION_HEADER] = fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass))),
	)

}

func (hb *httpBase) BearerAuth(token string) {
	hb.headers[AUTHORIZATION_HEADER] = fmt.Sprintf(
		"Bearer %s",
		token,
	)

}

func (hb *httpBase) generateUrl() string {
	fullUrl := fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(*hb.baseUrl, "/"),
		strings.TrimLeft(*hb.path, "/"),
	)

	if len(hb.parameters) == 0 {
		return fullUrl
	}

	parameters := make([]string, 0)

	for key, values := range hb.parameters {
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

func (hb *httpBase) parseBody() io.Reader {
	if *hb.method == GET {
		return nil
	}

	if hb.body != nil {
		return strings.NewReader(*hb.body)
	}

	return nil
}

func (hb *httpBase) extractParametersFromUrl(requestUrl string) string {
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
			hb.Parameter(key, "")
			continue
		}

		queryParamSplit := strings.Split(q, "=")
		key, _ := url.QueryUnescape(queryParamSplit[0])
		value, _ := url.QueryUnescape(queryParamSplit[1])

		hb.Parameter(key, value)
	}

	return strings.TrimRight(urlString, "/")
}
