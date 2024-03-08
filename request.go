package httpclient

import (
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
	timeout := DEFAULT_TIMEOUT
	if r.timeout != nil {
		timeout = *r.timeout
	}

	// Client
	hc := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
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
	method := GET
	if r.method != nil {
		method = *r.method
	}

	// Request
	url := r.generateUrl()
	body := r.parseBody()
	req, err := http.NewRequest(method, url, body)
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
	retryCount := 0
	if r.retryCount != nil {
		retryCount = *r.retryCount
	}

	// Tries the request atleast once
	var res *http.Response
	for i := 0; i < retryCount+1; i++ {
		res, err = hc.Do(req)
		if err == nil {
			break
		}

		r.logError(req, i+1, err)
	}

	// If no response was given the request has failed
	if res == nil {
		return
	}

	response = newResponse(r, res)
	return
}

func (r *Request) logError(req *http.Request, attempt int, err error) {
	proxy := "<not set>"
	if r.proxyUrl != nil {
		proxy = *r.proxyUrl
	}

	url := r.generateUrl()
	method := *r.method

	headers := make([]string, 0)
	for k, v := range req.Header {
		headers = append(headers, fmt.Sprintf(" - %s : %s", k, strings.Join(v, ", ")))
	}

	body := "<not set>"
	if r.body != nil {
		body = *r.body
	}

	timeout := *r.timeout
	retryCount := *r.retryCount

	stringsToCount := []string{err.Error(), proxy, url, method}
	stringsToCount = append(stringsToCount, headers...)
	longestString := getLengthOfLongestString(stringsToCount)
	preCalculationCharactersLength := 14

	fmt.Println(strings.Repeat("-", longestString+preCalculationCharactersLength))
	fmt.Printf("Attempt %d/%d failed\n", attempt, retryCount+1)
	fmt.Printf("Error:      : %s\n", err.Error())
	fmt.Println("\nRequest:")
	fmt.Printf("Proxy       : %s\n", proxy)
	fmt.Printf("Url         : %s\n", url)
	fmt.Printf("Method      : %s\n", method)
	fmt.Println("Headers     :")
	for _, h := range headers {
		fmt.Println(h)
	}
	fmt.Printf("Body        : %s\n", body)
	fmt.Printf("Timeout     : %d\n", timeout)
	fmt.Printf("Retry count : %d\n", retryCount)
	fmt.Println(strings.Repeat("-", longestString+preCalculationCharactersLength))
}
