package httpclient

import (
	"io"
	"net/http"
	"strings"
)

type Request struct {
	*http.Client
	*http.Request  // TODO Prepare()
	*http.Response // on Send()
}

func New(ref string) *Request {
	r, _ := http.NewRequest("", ref, nil) // TODO manual http.Request{}
	// TODO retain error? same as Do()
	return &Request{
		Client:  &http.Client{},
		Request: r,
		// TODO agent
	}
}

// TODO Proxy

func (r *Request) Post(body string) {
	rc := strings.NewReader(body)
	r.Request.Body = io.NopCloser(rc)
	r.Request.ContentLength = int64(rc.Len())
	if r.Method == "" { // TODO lost by http.NewRequest()
		r.Method = "POST"
	}
	if _, typeset := r.Request.Header["content-type"]; !typeset {
		r.Request.Header.Set("Content-Type", "text/plain")
	}
}

// TODO PostJson(any)

func (r *Request) Send() (err error) {
	// TODO retry
	r.Response, err = r.Client.Do(r.Request)
	return
}
