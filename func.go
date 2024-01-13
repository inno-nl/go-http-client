package httpclient

func Get() *HttpRequest {
	return &HttpRequest{
		method: GET,
	}
}

func Post() *HttpRequest {
	return &HttpRequest{
		method: POST,
	}
}

func Put() *HttpRequest {
	return &HttpRequest{
		method: PUT,
	}
}

func Patch() *HttpRequest {
	return &HttpRequest{
		method: PATCH,
	}
}

func Delete() *HttpRequest {
	return &HttpRequest{
		method: DELETE,
	}
}

func NewRequest() *HttpRequest {
	return &HttpRequest{}
}
