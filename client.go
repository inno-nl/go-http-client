package httpclient

type Client struct {
	base
}

func (c *Client) NewRequest() *Request {
	return &Request{c.base}
}
