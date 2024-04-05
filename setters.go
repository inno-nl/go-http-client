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

func (r *Request) AddURL(ref string) error {
	u, err := url.Parse(ref)
	if err != nil {
		return err
	}
	if v := u.RawQuery; v != "" {
		r.Parameters, err = url.ParseQuery(v)
		if err != nil {
			return err
		}
	} else if u.ForceQuery {
		r.Parameters = make(url.Values, 0)
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

	if v := u.Path; v != "" {
		if v[0] != '/' {
			// append relative path to existing base
			v = r.Request.URL.Path + "/" + v
		}
		r.Request.URL.Path = v
	}
	r.Request.URL.Fragment = u.Fragment // assume related

	return nil
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
