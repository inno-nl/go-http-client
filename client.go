package httpclient

type Client struct {
	base
}

func (c *Client) NewRequest() *Request {
	r := &Request{c.base}

	return r
}
