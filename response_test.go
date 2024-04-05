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
	r.Request.URL.Scheme = "https"
	r.Request.URL.Host = "httpbin.org"
	r.Request.URL.RawQuery += "&test"
	r.Parameters.Add("reset", "added")
	r.Parameters.Set("reset", "updated")
	r.Request.Header.Set(customHeader, url)
	r.Post(struct{Greeting string}{"HI!"})

	r.Prepare()
	expect := "https://httpbin.org/anything?preset=&reset=updated"
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

func TestPath(t *testing.T) {
	url := "//localhost/basepath?init=first"
	c := NewURL(url)
	if c.Request.URL.String() != url {
		t.Fatalf("altered initial %s: %s", url, c.Request.URL)
	}
	if c.Parameters.Get("init") != "first" {
		t.Fatalf("unexpected parameters in initial %s: %v", url, c.Parameters)
	}

	add := "subpath/2?init=second#only+here" // override parameters
	r := c.NewPath(add)
	if r.Request.URL.String() != "//localhost/basepath/"+add {
		t.Fatalf("unexpected results of added %s: %s", add, r.Request.URL)
	}
	if r.Parameters.Get("init") != "second" {
		t.Fatalf("unexpected parameters in added %s: %v", add, r.Parameters)
	}

	add = "/newbase/3" // keep params not hash
	r = r.NewPath(add)
	if r.Request.URL.String() != "//localhost/newbase/3?init=second" {
		t.Fatalf("unexpected results of added %s: %s", add, r.Request.URL)
	}
	if r.Parameters.Get("init") != "second" {
		t.Fatalf("missing parameters in added %s: %v", add, r.Parameters)
	}

	add = "https://u:p@inno.nl:80?"
	r = r.NewPath(add)
	if r.Request.URL.String() != "https://u:p@inno.nl:80/newbase/3" {
		t.Fatalf("unexpected results of added %s: %s", add, r.Request.URL)
	}
	if len(r.Parameters) != 0 {
		t.Fatalf("retained parameters in added %s: %v", add, r.Parameters)
	}

	add = "//test@"
	r2 := r.NewPath(add)
	if v := r2.Request.URL.User.String(); v != "test" {
		t.Fatalf("unexpected user part of added %s: %s", add, v)
	}
	if v := r.Request.URL.User.String(); v != "u:p" {
		t.Fatalf("altered user part after added %s: %s", add, v)
	}

	if c.Request.URL.String() != url {
		t.Fatalf("initial %s altered along the way: %s", url, c.Request.URL)
	}
}

func TestAuthorize(t *testing.T) {
	url := "//test@0:80"
	r := NewURL(url)
	r.SetBasicAuth("us*r", "passw*rd")
	if r.Request.URL.String() != url {
		t.Fatalf("altered base request %s: %s", url, r.Request.URL)
	}
	if v := r.Request.Header.Get("Authorization"); v != "Basic dXMqcjpwYXNzdypyZA==" {
		t.Fatalf("unexpected basic auth header: %s", v)
	}

	r.SetBearerAuth("tok*n")
	if v := r.Request.Header.Get("Authorization"); v != "Bearer tok*n" {
		t.Fatalf("unexpected bearer auth header: %s", v)
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
		r.SetTimeout(i + 10) // distinct for each subtest
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

func TestResend(t *testing.T) {
	r := NewURL("https://httpbin.org/status/500")
	r.Tries = 4
	r.DoRetry = func (r *Request, e error) error {
		if r.StatusCode == 500 {
			// change to a client error for the next attempt
			r.Request.URL.Path = "/status/404"
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

func TestTimeout(t *testing.T) {
	url := "https://httpbin.org/delay/1"
	r := NewURL(url)
	r.SetTimeout(1) // insufficient for transfer overhead
	err := r.Send()
	if err == nil { // assume deadline exceeded
		t.Fatalf("downloaded %s despite timeout", url)
	}

	r.SetTimeout(5) // an additional 4s should be enough
	err = r.Send()
	if err != nil {
		t.Fatalf("download with increased timeout failed as well: %v", err)
	}
}
