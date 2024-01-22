package httpclient

type BaseClient struct {
	baseUrl string
}

func (bc *BaseClient) BaseUrl(requestUrl string) {
	bc.baseUrl = requestUrl
}

func (bc *BaseClient) NewRequest() *HttpRequest {
	return NewRequest().BaseUrl(bc.baseUrl)
}
