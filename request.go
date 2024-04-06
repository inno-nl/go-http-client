package httpclient

import (
	"fmt"
	"net/http"
	"time"
)

const DefaultAgent = "inno-go-http-client/2"

type Request struct {
	Error      error      // postponed

	*http.Client
	*http.Request
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
	return
}

func NewURL(ref string) (r *Request) {
	r = New()
	_ = r.AddURL(ref) // invalid results reported by Client.Do()
	return
}

func (r *Request) NewURL(ref string) (d *Request) {
	d = r.Clone()
	err := d.AddURL(ref)
	if err != nil {
		d.Error = err
	}
	return
}

func (r *Request) Clone() *Request {
	d := new(Request)
	*d = *r
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
