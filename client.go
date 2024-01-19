package httpclient

type BaseClient struct {
	baseUrl string
}

func (bc *BaseClient) SetBaseUrl(baseUrl string) {
	bc.baseUrl = baseUrl
}

func (bc *BaseClient) Request() *HttpRequest {
	return NewRequest().BaseUrl(bc.baseUrl)
}
