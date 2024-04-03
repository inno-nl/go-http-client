package httpclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const DefaultAgent = "inno-go-http-client/2"

type Request struct {
	// pre-Prepare() setup
	Parameters url.Values // additional fields added to URL.RawQuery
	Error      error      // postponed

	*http.Client
	*http.Request  // on Prepare()
	*http.Response // on Send()

	DoRetry func(*Request, error) (error)
	Attempt int // Do() counter in Send()
	Tries   int // retry Do() if more than 1
}

func New(ref string) *Request {
	// http.NewRequest() without method and body
	u, _ := url.Parse(ref) // final errors reported by Client.Do()
	r := &http.Request{
		URL:    u,
		Header: http.Header{"User-Agent": {DefaultAgent}},
	}
	params := make(url.Values, 0)
	if u != nil && u.RawQuery != "" {
		params, _ = url.ParseQuery(u.RawQuery) // TODO error
	}
	return &Request{
		Client:     &http.Client{},
		Request:    r,
		Parameters: params,
	}
}

func (r *Request) Clone() *Request {
	d := new(Request)
	*d = *r
	d.Parameters = make(url.Values, len(r.Parameters))
	for k, v := range r.Parameters {
		d.Parameters[k] = v
	}
	if d.Client != nil {
		d.Client = new(http.Client)
		*d.Client = *r.Client
	}
	if d.Request != nil {
		d.Request = r.Request.Clone(r.Context())
	}
	// Response intentionally kept
	return d
}

func (r *Request) Timeout(s int) {
	r.Client.Timeout = time.Duration(s) * time.Second
}

func (r *Request) ProxyUrl(ref string) {
	u, err := url.Parse(ref)
	if err != nil {
		r.Error = err
		return
	}
	r.Client.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
}

func (r *Request) BearerAuth(token string) {
	r.Request.Header.Set("Authorization", "Bearer "+token)
}

func (r *Request) BasicAuth(user, pass string) {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	r.Request.Header.Set("Authorization", "Basic "+token)
}

func (r *Request) Post(body string) {
	rc := bytes.NewBufferString(body)
	r.Request.Body = io.NopCloser(rc)
	r.Request.ContentLength = int64(rc.Len())
	if r.Method == "" {
		r.Method = "POST"
	}
	if _, typeset := r.Request.Header["content-type"]; !typeset {
		r.Request.Header.Set("Content-Type", "text/plain")
	}
}

func (r *Request) PostJson(body any) {
	data, err := json.Marshal(body)
	if err != nil {
		r.Error = err // TODO join
		return
	}
	rc := bytes.NewReader(data)
	r.Request.Body = io.NopCloser(rc)
	r.Request.ContentLength = int64(rc.Len())
	if r.Method == "" {
		r.Method = "POST"
	}
	if _, typeset := r.Request.Header["content-type"]; !typeset {
		r.Request.Header.Set("Content-Type", "application/json")
	}
}

func (r *Request) Send() (err error) {
	err = r.Error
	if err != nil {
		return // TODO wrap error
	}

	if len(r.Parameters) > 0 {
		r.URL.RawQuery = r.Parameters.Encode()
	}

	delay := time.Second
	for r.Attempt = 1; ; r.Attempt++ {
		r.Response, err = r.Client.Do(r.Request)
		if r.Attempt >= r.Tries {
			break
		}
		if r.DoRetry != nil {
			err = r.DoRetry(r, err)
		} else if err == nil && r.StatusCode >= 500 {
			err = fmt.Errorf("unsuccessful response code %s", r.Status)
		}
		if err == nil {
			break
		}
		time.Sleep(delay)
		delay *= 2 // increase exponentially
	}
	return
}
