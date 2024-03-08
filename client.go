package httpclient

type Client struct {
	httpBase
}

func (hc *Client) NewRequest() *Request {
	r := &Request{hc.httpBase}

	return r
}
