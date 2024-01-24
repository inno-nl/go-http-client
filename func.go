package httpclient

func NewRequest() *HttpRequest {
	return &HttpRequest{
		timeout:    60,
		retryCount: 0,
	}
}
