package httpclient

type Client struct {
	httpBase
}

func (hc *Client) NewRequest() *Request {
	r := &Request{}

	r.baseUrl = hc.baseUrl
	r.path = hc.path
	r.method = hc.method
	r.parameters = hc.parameters
	r.headers = hc.headers
	r.contentType = hc.contentType
	r.body = hc.body
	r.timeout = hc.timeout
	r.retryCount = hc.retryCount

	return r
}
