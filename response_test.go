package httpclient

import (
	"testing"

	"fmt"
	"strings"
)

func TestInvalid(t *testing.T) {
	url := "invalid:blopplop"
	_, err := New(url).Bytes()
	expect := fmt.Sprintf(`Get "%s": unsupported protocol scheme "invalid"`, url)
	if err.Error() != expect {
		t.Fatalf("missing protocol error: %v", err)
	}
}

func TestText(t *testing.T) {
	url := "http://sheet.shiar.nl/sample.txt"
	body, err := New(url).String()
	if err != nil {
		t.Fatalf("could not download %s: %v", url, err)
	}
	if !strings.HasPrefix(body, "Unicode sample") {
		t.Fatalf("error in downloaded %s:\n%s", url, body[:140])
	}
}

func TestError(t *testing.T) {
	url := "https://httpbin.org/status/404"
	_, err := New(url).Bytes()
	if err == nil || err.Error() != "unsuccessful response code 404 Not Found" {
		t.Fatalf("unexpected error from %s: %v", url, err)
	}
}

type HttpbinEcho struct {
	Origin  string
	Url     string
	Data    string
	Json    map[string]any
	Headers map[string]string
}

func TestParameters(t *testing.T) {
	const customHeader = "X-Hello"
	url := "invalid:///anything?preset&reset=initial"
	r := New(url)
	r.URL.Scheme = "https"
	r.URL.Host = "httpbin.org"
	r.URL.RawQuery += "&test"
	r.Parameters.Add("reset", "added")
	r.Parameters.Set("reset", "updated")
	r.Request.Header.Set(customHeader, url)
	r.PostJson(struct{Greeting string}{"HI!"})

	var res HttpbinEcho
	err := r.Json(&res)
	if err != nil {
		t.Fatalf("could not download %s: %v", url, err)
	}
	u := r.URL.String()
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

func TestPost(t *testing.T) {
	url := "https://httpbin.org/post"
	input := "hi"
	r := New(url)
	r.Post(input)
	var res HttpbinEcho
	err := r.Json(&res)
	if err != nil {
		t.Fatalf("could not post %s: %v", url, err)
	}
	if res.Data != input {
		t.Fatalf("unexpected post results: %v", res)
	}
}

func TestRetry(t *testing.T) {
	url := "https://httpbin.org/status/500"
	r := New(url)
	r.Tries = 2
	err := r.Send()
	if err != nil {
		t.Fatalf("error downloading %s: %v", url, err)
	}
	if r.StatusCode != 500 {
		t.Fatalf("downloaded %s with incorrect status: %s", url, r.Status)
	}
	if r.Attempt != r.Tries {
		t.Fatalf("tried %d downloads of %s", r.Attempt, url)
	}
}

func TestTimeout(t *testing.T) {
	url := "https://httpbin.org/delay/2"
	r := New(url)
	r.Timeout(1)
	err := r.Send()
	if err == nil { // assume deadline exceeded
		t.Fatalf("downloaded %s despite timeout", url)
	}

	r.Timeout(3)
	err = r.Send()
	if err != nil {
		t.Fatalf("download with increased timeout failed as well: %v", err)
	}
}