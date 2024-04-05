package httpclient

import (
	"fmt"
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

func New() (r *Request) {
	r = new(Request)
	r.Client = &http.Client{}

	// insufficient data for http.NewRequest()
	r.Request = &http.Request{
		Header: http.Header{"User-Agent": {DefaultAgent}},
	}
	r.Parameters = make(url.Values, 0)
	return
}

func NewURL(ref string) (r *Request) {
	r = New()
	_ = r.AddURL(ref) // invalid results reported by Client.Do()
	return
}

func (r *Request) NewURL(ref string) (d *Request) {
	d = r.Clone()
	d.Request.URL = nil
	err := d.AddURL(ref)
	if err != nil {
		d.Error = err
	}
	return
}

func (r *Request) NewPath(path string) (d *Request) {
	d = r.Clone()
	err := d.AddURL(path)
	if err != nil {
		d.Error = err // TODO join
	}
	return
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
	// Response will be reset by Send()
	return d
}

func (r *Request) Prepare() {
	if len(r.Parameters) > 0 {
		r.Request.URL.RawQuery = r.Parameters.Encode()
	}
}

func (r *Request) Send() error {
	r.Response = nil
	r.Attempt = 0
	return r.Resend()
}

func (r *Request) Resend() (err error) {
	err = r.Error
	if err != nil {
		return // TODO wrap error
	}

	r.Prepare()

	delay := time.Second
	for r.Attempt++; ; r.Attempt++ {
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
