package httpclient

import (
	"testing"

	"fmt"
	"strings"
)

func TestInvalid(t *testing.T) {
	url := "invalid:blopplop"
	_, err := NewURL(url).Bytes()
	expect := fmt.Sprintf(`Get "%s": unsupported protocol scheme "invalid"`, url)
	if err.Error() != expect {
		t.Fatalf("missing protocol error: %v", err)
	}
}

func TestText(t *testing.T) {
	url := "http://sheet.shiar.nl/sample.txt"
	body, err := NewURL(url).String()
	if err != nil {
		t.Fatalf("could not download %s: %v", url, err)
	}
	if !strings.HasPrefix(body, "Unicode sample") {
		t.Fatalf("error in downloaded %s:\n%s", url, body[:140])
	}
}

func TestError(t *testing.T) {
	url := "https://httpbin.org/status/404"
	_, err := NewURL(url).Bytes()
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
	Args    map[string]any
}

func TestParameters(t *testing.T) {
	const customHeader = "X-Hello"
	url := "invalid:///anything?preset&reset=initial"
	r := NewURL(url)
	r.URL.Scheme = "https"
	r.URL.Host = "httpbin.org"
	r.URL.RawQuery += "&test"
	r.Parameters.Add("reset", "added")
	r.Parameters.Set("reset", "updated")
	r.Request.Header.Set(customHeader, url)
	r.Post(struct{Greeting string}{"HI!"})

	r.Prepare()
	expect := "https://httpbin.org/anything?preset=&reset=updated"
	if u := r.URL.String(); u != expect {
		t.Fatalf("prepared url turned out incorrectly: %s", u)
	}

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

func TestReuse(t *testing.T) {
	rtypes := []string{"image/jpeg", "text/plain"}
	url := "https://httpbin.org/anything"
	c := NewURL(url)

	for i, rtype := range rtypes {
		r := c.Clone()
		r.Timeout(i + 10) // distinct for each subtest
		r.Request.Header.Set("X-Accept", rtype)
		r.Parameters.Add("type", rtype)
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

func TestRetry(t *testing.T) {
	url := "https://httpbin.org/status/500"
	r := NewURL(url)
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
	url := "https://httpbin.org/delay/1"
	r := NewURL(url)
	r.Timeout(1) // insufficient for transfer overhead
	err := r.Send()
	if err == nil { // assume deadline exceeded
		t.Fatalf("downloaded %s despite timeout", url)
	}

	r.Timeout(5) // an additional 4s should be enough
	err = r.Send()
	if err != nil {
		t.Fatalf("download with increased timeout failed as well: %v", err)
	}
}
