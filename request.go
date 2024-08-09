package httpclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	base
}

func (r *Request) Send() (response *Response, err error) {
	// Timeout
	if r.timeout == nil {
		defaultTimeout := DEFAULT_TIMEOUT
		r.timeout = &defaultTimeout
	}

	// Client
	hc := &http.Client{
		Timeout: time.Duration(*r.timeout) * time.Second,
	}

	// Proxy url
	if r.proxyUrl != nil {
		proxyUrl, err := url.Parse(*r.proxyUrl)
		if err != nil {
			return nil, err
		}
		hc.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	// Method
	if r.method == nil {
		defaultMethod := GET
		r.method = &defaultMethod
	}

	// Request
	url := r.generateUrl()
	body := r.parseBody()
	req, err := http.NewRequest(*r.method, url, body)
	if err != nil {
		return
	}

	// User agent
	req.Header.Set(
		"user-agent",
		"inno-go-http-client",
	)

	// Override content type
	if r.contentType != nil {
		req.Header.Set("Content-type", *r.contentType)
	}

	// Headers
	if r.headers != nil {
		for k, v := range r.headers {
			req.Header.Set(k, v)
		}
	}

	// Retrycount
	if r.retryCount == nil {
		defaultRetryCount := 0
		r.retryCount = &defaultRetryCount
	}

	// Tries the request atleast once
	retryCount := *r.retryCount
	var res *http.Response
	for i := 0; i < retryCount+1; i++ {
		res, err = hc.Do(req)

		hasExponentialBackup := r.exponentialBackoff > 0
		tooManyRequests := res.StatusCode == 429

		if err != nil || (hasExponentialBackup && tooManyRequests) {
			if tooManyRequests {
				err = errors.New("too many requests")
			}

			r.logError(req, i+1, err)

			if hasExponentialBackup {
				time.Sleep(time.Duration(r.exponentialBackoff*(i+1)) * time.Second)
			}

			continue
		}

		break
	}

	// If no response was given the request has failed
	if res == nil {
		return
	}

	response = newResponse(r, res)
	return
}

func (r *Request) logError(req *http.Request, attempt int, err error) {
	if !r.logErrors || r.errorLogFunc == nil {
		return
	}

	proxyUrl := ""
	if r.proxyUrl != nil {
		proxyUrl = *r.proxyUrl
	}

	url := r.generateUrl()

	headers := make([]string, 0)
	for k, v := range req.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}

	body := ""
	if r.body != nil {
		body = *r.body
	}

	r.errorLogFunc(Error{
		Error:      err,
		Attempt:    attempt,
		ProxyUrl:   proxyUrl,
		Url:        url,
		Method:     *r.method,
		Headers:    headers,
		Body:       body,
		Timeout:    *r.timeout,
		RetryCount: *r.retryCount,
	})
}
