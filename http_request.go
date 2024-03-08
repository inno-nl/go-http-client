package httpclient

import (
	"net/http"
	"time"
)

type HttpRequest struct {
	httpBase
}

func (hr *HttpRequest) Send() (response *HttpResponse, err error) {
	// Timeout
	timeout := DEFAULT_TIMEOUT
	if hr.timeout != nil {
		timeout = *hr.timeout
	}

	hc := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Method
	method := GET
	if hr.method != nil {
		method = *hr.method
	}

	req, err := http.NewRequest(
		method,
		hr.generateUrl(),
		hr.parseBody(),
	)
	if err != nil {
		return
	}

	// User agent
	req.Header.Set(
		"user-agent",
		"inno-go-http-client",
	)

	// Override content type
	if hr.contentType != nil {
		req.Header.Set("Content-type", *hr.contentType)
	}

	// Headers
	if hr.headers != nil {
		for k, v := range hr.headers {
			req.Header.Set(k, v)
		}
	}

	// Retrycount
	retryCount := 0
	if hr.retryCount != nil {
		retryCount = *hr.retryCount
	}

	// Tries the request atleast once
	var res *http.Response
	for i := 0; i < retryCount+1; i++ {
		res, err = hc.Do(req)
		if err == nil {
			break
		}
	}

	// If no response was given the request has failed
	if res == nil {
		return
	}

	response = newHttpResponse(hr, res)
	return
}
