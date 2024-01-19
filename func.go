package httpclient

func Get() *HttpRequest {
	return NewRequest().Method(GET)
}

func Post() *HttpRequest {
	return NewRequest().Method(POST)
}

func Put() *HttpRequest {
	return NewRequest().Method(PUT)
}

func Patch() *HttpRequest {
	return NewRequest().Method(PATCH)
}

func Delete() *HttpRequest {
	return NewRequest().Method(DELETE)
}

func NewRequest() *HttpRequest {
	return &HttpRequest{}
}
