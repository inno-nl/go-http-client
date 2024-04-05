package httpclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

func (r *Request) SetURL(ref string) {
	r.Request.URL, _ = url.Parse(ref) // final errors reported by Client.Do()
	if r.Request.URL != nil && r.Request.URL.RawQuery != "" {
		r.Parameters, _ = url.ParseQuery(r.Request.URL.RawQuery) // TODO error
	}
}

func (r *Request) SetTimeout(s int) {
	r.Client.Timeout = time.Duration(s) * time.Second
}

func (r *Request) SetProxyURL(ref string) {
	u, err := url.Parse(ref)
	if err != nil {
		r.Error = err
		return
	}
	r.Client.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
}

func (r *Request) SetBearerAuth(token string) {
	r.Request.Header.Set("Authorization", "Bearer "+token)
}

func (r *Request) SetBasicAuth(user, pass string) {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	r.Request.Header.Set("Authorization", "Basic "+token)
}

func (r *Request) Post(body any) {
	if r.Method == "" {
		r.Method = "POST"
	}

	var data []byte
	switch body.(type) {
	case nil:
	case string:
		data = []byte(body.(string)) // fallthrough
	case []byte:
		data = body.([]byte)
		if _, typeset := r.Request.Header["content-type"]; !typeset {
			r.Request.Header.Set("Content-Type", "text/plain")
		}
	default:
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			r.Error = err // TODO join
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
