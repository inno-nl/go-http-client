package httpclient

import (
	"net/http"
	"net/url"
	"time"
)

type Request struct {
	httpBase
}

func (hr *Request) Send() (response *Response, err error) {
	// Timeout
	timeout := DEFAULT_TIMEOUT
	if hr.timeout != nil {
		timeout = *hr.timeout
	}

	hc := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Proxy url
	if hr.proxyUrl != nil {
		proxyUrl, err := url.Parse(*hr.proxyUrl)
		if err != nil {
			return nil, err
		}
		hc.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
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

	response = newResponse(hr, res)
	return
}
