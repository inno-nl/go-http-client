package httpclient

type HttpClient struct {
	httpBase
}

func (hc *HttpClient) NewRequest() *HttpRequest {
	r := &HttpRequest{}

	r.baseUrl = hc.baseUrl
	r.path = hc.path
	r.method = hc.method
	r.parameters = hc.parameters
	r.parameters = hc.parameters
	r.headers = hc.headers
	r.contentType = hc.contentType
	r.body = hc.body
	r.timeout = hc.timeout
	r.retryCount = hc.retryCount

	return r
}
