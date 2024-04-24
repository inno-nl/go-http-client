package httpclient

import (
	"testing"

	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"time"
)

const server = "httpbin.org"
const s = "https://" + server

func serverClient() (r *Request) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/status/500":
			w.WriteHeader(500)
		case "/forward":
			http.Redirect(w, r, "/missing", http.StatusFound)
		case "/missing", "/status/404":
			w.WriteHeader(404)
			w.Write([]byte(sampleHtml))
		case "/delay":
			amount, err := strconv.Atoi(r.URL.Query().Get("ms"))
			if err == nil {
				time.Sleep(time.Duration(amount) * time.Millisecond)
			} else {
				w.WriteHeader(400)
			}
		default:
			w.WriteHeader(501)
		}
		return
	})
	s := httptest.NewUnstartedServer(h)
	s.Start()
	// t.Cleanup(func() { s.Close() })
	r = NewURL(s.URL)
	r.Client = s.Client()
	return
}

func TestClientRedirect(t *testing.T) {
	r := serverClient()
	r = r.NewURL("forward")
	body, err := r.Text()
	if err == nil {
		t.Fatalf("unexpected success: %v", body)
	}
	var e *StatusError
	if !errors.As(err, &e) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if e.Code != 404 {
		t.Fatalf("unexpected download status: %v", err)
	}
	if v := r.Response.Request.URL; v.Path != "/missing" {
		t.Fatalf("unexpected final url: %v", v)
	}
	if body != sampleHtml {
		t.Fatalf("unexpected download body: %v", body)
	}
}

func TestInvalid(t *testing.T) {
	u := "invalid:blopplop"
	r := NewURL(u)
	_, err := r.Bytes()
	var urlerr *url.Error
	if !errors.As(err, &urlerr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	expect := `unsupported protocol scheme "invalid"`
	if urlerr.Err.Error() != expect {
		t.Fatalf("missing protocol error: %v", urlerr)
	}
	if v := r.Response; v != nil {
		t.Fatalf("unexpected response: %v", v)
	}
}

type HttpbinEcho struct {
	Origin  string
	Url     string
	Data    string
	Json    map[string]any
	Headers map[string]string
	Args    map[string]any
}

func TestRemoteJson(t *testing.T) {
	const customHeader = "X-Hello"
	url := "invalid:///anything?preset&reset=initial"
	r := NewURL(url)
	r.Request.URL.Scheme = "https"
	r.Request.URL.Host = server
	r.AddURL("?reset=updated")
	r.AddURL("&reset=added")
	r.SetHeader(customHeader, url)
	r.Post(struct{Greeting string}{"HI!"})
	expect := s + "/anything?reset=updated&reset=added"
	if u := r.Request.URL.String(); u != expect {
		t.Fatalf("prepared url turned out incorrectly: %s", u)
	}

	var res HttpbinEcho
	err := r.Json(&res)
	if err != nil {
		t.Fatalf("could not download %s: %v", url, err)
	}
	u := r.Request.URL.String()
	if v := res.Url; v != u {
		t.Fatalf("sent url (%s) mismatch: %v", v, u)
	}
	if v := res.Headers["User-Agent"]; v != DefaultAgent {
		t.Fatalf("sent user agent mismatch: %v", v)
	}
	if v := res.Headers[customHeader]; v != url {
		t.Fatalf("missing custom header %s: %v", customHeader, v)
	}
	if v := res.Json["Greeting"]; v != "HI!" {
		t.Fatalf("sent json data mismatch: %v", res.Json)
	}
}

func TestRemotePost(t *testing.T) {
	url := s + "/post"
	r := NewURL(url)
	r.Post(nil)
	var res HttpbinEcho
	err := r.Json(&res)
	if err != nil {
		t.Fatalf("could not post %s: %v", url, err)
	}
	if res.Data != "" {
		t.Fatalf("unexpected post results: %v", res)
	}

	input := "hi"
	r.Response = nil // reset
	r.Post(input)
	err = r.Json(&res)
	if err != nil {
		t.Fatalf("could not post %s with json: %v", url, err)
	}
	if res.Data != input {
		t.Fatalf("mismatching post results: %v", res)
	}
}

func TestRemoteJsonError(t *testing.T) {
	u := s + "/xml"
	r := NewURL(u)
	var res HttpbinEcho
	err := r.Json(&res)
	if err == nil {
		t.Fatalf("unexpectedly parsed %s as json: %v", u, res)
	}
	if err != ErrJsonLikeXml {
		t.Fatalf("unexpected error parsing %s: %v", u, err)
	}
	if v := r.Preview(); v != xmlDeclare+"..." {
		t.Fatalf("unexpected preview of %s: %v", u, v)
	}
}

func TestRemoteReuse(t *testing.T) {
	rtypes := []string{"image/jpeg", "text/plain"}
	url := s + "/anything"
	c := NewURL(url)

	for i, rtype := range rtypes {
		r := c.Clone()
		r.SetTimeout(i + 10) // distinct for each subtest
		r.SetHeader("X-Accept", rtype)
		r.AddURL("?type=" + rtype)
		res := HttpbinEcho{}
		err := r.Json(&res)
		if err != nil {
			t.Fatalf("error downloading with %s: %v", rtype, err)
		}
		if v := res.Args["type"]; v != rtype {
			t.Fatalf("request for %s had tainted parameters: %v", rtype, v)
		}
		if v := res.Headers["X-Accept"]; v != rtype {
			t.Fatalf("request for %s instead gave %v", rtype, v)
		}
	}

	if v := c.Request.Header.Get("X-Accept"); v != "" {
		res := "no response though"
		if c.Response != nil {
			res = c.Status
		}
		t.Fatalf("client object modified by request %s: %s", v, res)
	}
	if v := c.Client.Timeout; v != 0 {
		t.Fatalf("client timeout modified by request: %s", v)
	}
}

func TestClientRetry(t *testing.T) {
	r := serverClient().NewURL("status/500")
	r.SetRetry(1)
	err := r.Send()
	if err != nil {
		t.Fatalf("error downloading %s: %v", r.URL, err)
	}
	if r.StatusCode != 500 {
		t.Fatalf("downloaded %s with incorrect status: %s", r.URL, r.Status)
	}
	if r.Attempt != r.Tries {
		t.Fatalf("tried %d downloads of %s", r.Attempt, r.URL)
	}
}

func TestClientResend(t *testing.T) {
	r := serverClient().NewURL("status/500")
	r.SetRetry(3)
	if v := r.Tries; v != 4 {
		t.Fatalf("misinterpreted retry count: %d", v)
	}
	r.DoRetry = func(r *Request, e error) error {
		if r.StatusCode == 500 {
			// change to a client error for the next attempt
			r.AddURL("/status/404")
			return fmt.Errorf("retry after initial code %s", r.Status)
		}
		return e
	}

	err := r.Send() // 500 then 404, no 3rd retry
	if err != nil {
		t.Fatalf("error downloading %s: %v", r.Request.URL, err)
	}
	if r.StatusCode != 404 {
		t.Fatalf("downloaded %s with unexpected status: %s", r.Request.URL, r.Status)
	}
	if r.Attempt != 2 {
		t.Fatalf("tried %d initial downloads of %s", r.Attempt, r.Request.URL)
	}

	err = r.Resend() // another 404, no retries
	if err != nil {
		t.Fatalf("error redownloading %s: %v", r.Request.URL, err)
	}
	if r.StatusCode != 404 {
		t.Fatalf("redownloaded %s with incorrect status: %s", r.Request.URL, r.Status)
	}
	if r.Attempt != 3 {
		t.Fatalf("tried %d total downloads of %s", r.Attempt, r.Request.URL)
	}
}

func TestClientTimeout(t *testing.T) {
	r := serverClient().NewURL("delay?ms=1001")
	r.SetTimeout(1) // insufficient for slightly longer response

	err := r.Send()
	if err == nil { // assume deadline exceeded
		t.Fatalf("downloaded %s despite timeout", r.URL)
	}
	var urlerr *url.Error
	if !errors.As(err, &urlerr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if v := fmt.Sprintf("%T", urlerr.Err); v != "*http.httpError" { // reflect.TypeOf
		t.Fatalf("unexpected wrapped error type: %s", v)
	}

	r.SetTimeout(5) // an additional 4s should be enough
	err = r.Send()
	if err != nil {
		t.Fatalf("download with increased timeout failed as well: %v", err)
	}
}
