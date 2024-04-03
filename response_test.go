package httpclient

import (
	"testing"

	"bytes"
	"fmt"
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
	body, err := New(url).Bytes()
	if err != nil {
		t.Fatalf("could not download %s: %v", url, err)
	}
	if !bytes.HasPrefix(body, []byte("Unicode sample")) {
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

func TestPost(t *testing.T) {
	url := "https://httpbin.org/post"
	input := "hi"
	r := New(url)
	r.Post(input)
	r.Method = "POST" // TODO workaround
	res := struct{Data string}{}
	err := r.Json(&res)
	if err != nil {
		t.Fatalf("could not post %s: %v", url, err)
	}
	if res.Data != input {
		t.Fatalf("unexpected post results: %v", res)
	}
}
