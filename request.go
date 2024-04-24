package httpclient

import (
	"net/http"
	"time"
)

const DefaultAgent = "inno-go-http-client/2"

type Request struct {
	Error error // postponed until Send()

	*http.Client
	*http.Request
	*http.Response // on Send()

	// Override to check [Send] attempts for temporary exceptions.
	DoRetry func(*Request, error) error
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
	_ = d.AddURL(ref) // keep ignoring parse errors until Send()
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
	if err = r.Error; err != nil {
		return
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
			err = &StatusError{r.Response.StatusCode, r.Response.Status}
		}
		if err == nil {
			break
		}
		time.Sleep(delay)
		delay *= 2 // increase exponentially
	}
	return
}
