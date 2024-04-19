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

func (r *Request) SetHeader(key string, value any) {
	if value == nil {
		r.Request.Header.Del(key)
		return
	}
	s := fmt.Sprintf("%v", value) // stringify any
	r.Request.Header.Set(key, s)
}

func (r *Request) SetQuery(replacement url.Values) {
	r.Request.URL.RawQuery = replacement.Encode()
}

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

func (r *Request) SetTimeout(s int) {
	r.Client.Timeout = time.Duration(s) * time.Second
}

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
