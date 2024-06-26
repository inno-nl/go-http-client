package httpclient

import (
	"testing"

	"bytes"
	"fmt"
	"net/url"
	"strings"
)

func TestParsePath(t *testing.T) {
	url := "//localhost/basepath?init=first"
	c := NewURL(url)
	if c.Request.URL.String() != url {
		t.Fatalf("altered initial %s: %s", url, c.Request.URL)
	}
	if v := c.Request.URL.RawQuery; v != "init=first" {
		t.Fatalf("unexpected parameters in initial %s: %v", url, v)
	}

	add := "subpath/2?init=second#only+here" // override parameters
	r := c.NewURL(add)
	if r.Request.URL.String() != "//localhost/basepath/"+add {
		t.Fatalf("unexpected results of added %s: %s", add, r.Request.URL)
	}

	add = "/newbase/3" // keep params not hash
	r = r.NewURL(add)
	if r.Request.URL.String() != "//localhost/newbase/3?init=second" {
		t.Fatalf("unexpected results of added %s: %s", add, r.Request.URL)
	}

	add = "https://u:p@inno.nl:80?"
	r = r.NewURL(add)
	if r.Request.URL.String() != "https://u:p@inno.nl:80/newbase/3" {
		t.Fatalf("unexpected results of added %s: %s", add, r.Request.URL)
	}
	if v := r.Request.URL.RawQuery; v != "" {
		t.Fatalf("retained parameters in added %s: %v", add, v)
	}

	add = "//test@"
	r2 := r.NewURL(add)
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

func TestParseHeaders(t *testing.T) {
	url := "incomplete"
	r := NewURL(url)
	if v := r.Request.Header; v == nil || v.Get("User-Agent") != DefaultAgent {
		t.Fatalf("missing default headers after init: %s", v)
	}

	r.SetHeader("temp", "added")
	r.SetHeader("test", "first")
	r.SetHeader("x-empty", "")
	r.SetHeader("Invalid ", "no error?")
	r.Request.Header.Add("test", "second")
	r.SetHeader("TEMP", nil)
	r.SetHeader("X-String", "string!\n")
	r.SetHeader("X-Number", "NaN")
	r.SetHeader("x-number", 42)
	expect := []string{
		"Test: first",
		"Test: second",
		"User-Agent: " + DefaultAgent,
		"X-Empty: ",
		"X-Number: 42",
		"X-String: string!",
		"", // eol
	}
	w := bytes.NewBuffer([]byte{})
	if err := r.Request.Header.Write(w); err != nil {
		t.Fatalf("unexpected header error: %v", err)
	}
	if v := w.String(); v != strings.Join(expect, "\r\n") {
		t.Fatalf("unexpected header results:\n%v", v)
	}
}

func TestParseAuthorize(t *testing.T) {
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

func TestParseQuery(t *testing.T) {
	u := "invalid:///anything?preset&reset=initial"
	r := NewURL(u)
	r.AddURL("https:")
	r.SetQuery(url.Values{"reset": {"replaced", "&double"}})
	expect := "https:///anything?reset=replaced&reset=%26double"

	r.AddQuery("reset", "&again")
	r.AddQuery("empty", "")
	r.AddQuery("empty", nil)
	r.AddQuery("number", 42)
	r.AddQuery("error", fmt.Errorf("oopß"))
	expect += "&reset=%26again&empty=&empty&number=42&error=oop%C3%9F"
	if v := r.Request.URL.String(); v != expect {
		t.Fatalf("unexpected url results: %s", v)
	}
}
