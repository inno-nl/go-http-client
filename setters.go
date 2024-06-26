package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Alter the request URL by replacing or appending some or all parts
// of an RFC 3986 URI.
//
// Most parts will be replaced if present after [url.Parse].
//
//	r.AddURL("http://localhost") // initial location
//	r.AddURL("https:")           // upgrade protocol
//	r.AddURL(":8080")            // add explicit port number
//	r.AddURL("#hello there")     // set fragment component
//
// A path without leading / is joined to an(y) existing base path
// after an implied / if necessary:
//
//	r.AddURL("/api/v1")   // replacement root
//	r.AddURL("hello")     // appended endpoint
//	r.URL.Path += "hello" // concatenated without /
//
// A query can be indicated either by standard ? to replace the RawQuery part,
// or & to keep any existing value:
//
//	r.AddURL("&extra") // set in addition to earlier parameters
//	r.AddURL("?")      // delete everything
func (r *Request) AddURL(ref string) error {
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}
	if r.Request.URL == nil {
		r.Request.URL = u
		return nil
	}

	if v := u.Scheme; v != "" {
		r.Request.URL.Scheme = v
	}
	if v := u.User; v != nil {
		r.Request.URL.User = v
	}
	if v := u.Host; v != "" {
		r.Request.URL.Host = v
	}
	if v := u.RawQuery; v != "" || u.ForceQuery {
		r.Request.URL.RawQuery = v
	}
	if queryCut := strings.IndexByte(u.Path, '&'); queryCut >= 0 {
		// split parameters from path with unencoded ampersand
		if r.Request.URL.RawQuery != "" {
			r.Request.URL.RawQuery += "&"
		}
		r.Request.URL.RawQuery += u.Path[queryCut+1:] // after
		u.Path = u.Path[:queryCut]                    // before
	}

	if v := u.Path; v != "" {
		if v[0] != '/' {
			// append relative path to existing base
			v = strings.TrimRight(r.Request.URL.Path, "/") + "/" + v
		}
		r.Request.URL.Path = v
	}
	r.Request.URL.Fragment = u.Fragment // assume related

	return nil
}

// Replaces a request header value, equivalent to [Request.Header.Set]
// but also stringifies values and deletes if nil.
//
//	r.SetHeader("user-agent", nil) // delete default
//	r.SetHeader("x-Error", errors.New("message"))
//	message := r.Request.Header.Get("X-error") // inherited interface
func (r *Request) SetHeader(key string, value any) {
	if value == nil {
		r.Request.Header.Del(key)
		return
	}
	s := fmt.Sprintf("%v", value) // stringify any
	r.Request.Header.Set(key, s)
}

// Completely replace any query parameters by an [url.Values] map.
//
//	params = url.Values{
//	    "config": {"default"},
//	}
//	params.Set("config", "override")
//	params.Add("limit", 10)
//	r.SetQuery(params)
//	r.AddQuery("debug", nil)
func (r *Request) SetQuery(replacement url.Values) {
	r.Request.URL.RawQuery = replacement.Encode()
}

// Append a &key=value parameter to RawQuery.
// Like AddURL("&...") but with %-escaping both key and value,
// and values stringified or omitted if nil.
//
//	r.AddQuery("limit", 42)  // AddURL("&limit=42")
//	r.AddQuery("debug", nil) // AddURL("&debug")
func (r *Request) AddQuery(k string, v any) {
	q := &r.Request.URL.RawQuery
	if *q != "" {
		*q += "&"
	}
	*q += url.QueryEscape(k)
	if v != nil {
		s := fmt.Sprintf("%v", v) // stringify any
		*q += "=" + url.QueryEscape(s)
	}
}

// Override the number of [Tries] so an additional number of [Send] attempts
// are made on receiving server errors.
//
// This feature is disabled if kept or reset to 0.
// A value of 1 will retry once after waiting for a second.
// Higher values will keep trying, each time doubling the delay in between.
// [Response] will be the first success or last error,
// with [Attempt] set to the final number of tries.
func (r *Request) SetRetry(num int) {
	r.Tries = num + 1
}

// Shorthand to set the client timeout duration to a number of seconds.
func (r *Request) SetTimeout(s float64) {
	r.Client.Timeout = time.Duration(s * float64(time.Second))
}

// Configure a proxy URL as [Client.Transport.Proxy].
func (r *Request) SetProxyURL(ref string) error {
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}
	r.Client.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	return nil
}

// Set the request's Authorization header to the given Bearer token,
// like [SetBasicAuth] but simply relaying a single value.
func (r *Request) SetBearerAuth(token string) {
	r.Request.Header.Set("Authorization", "Bearer "+token)
}

// Provide a request body to be sent along with expected headers.
// The Method will be changed to POST if not yet explicitly set.
// Data can be given as []byte to be sent literally, a string
// which also applies a Content-Type of text/plain unless already defined,
// or a struct automatically marshalled as JSON and sent as application/json.
func (r *Request) Post(body any) {
	if r.Method == "" {
		r.Method = "POST"
	}

	var data []byte
	switch body.(type) {
	case nil:
	case []byte:
		data = body.([]byte)
	case string:
		data = []byte(body.(string))
		if _, typeset := r.Request.Header["content-type"]; !typeset {
			r.Request.Header.Set("Content-Type", "text/plain")
		}
	default:
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			r.Error = fmt.Errorf("Post data invalid: %w", err) // wrap
			return
		}
		if _, typeset := r.Request.Header["content-type"]; !typeset {
			r.Request.Header.Set("Content-Type", "application/json")
		}
	}
	rc := bytes.NewReader(data)
	r.Request.Body = io.NopCloser(rc)
	r.Request.ContentLength = int64(rc.Len())
	// TODO r.Request.GetBody
}
