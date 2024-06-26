package httpclient

import (
	"testing"

	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

const sampleText = "Eĥoŝanĝº ĉiĵaŭde" // valid unicode
const sampleData = "Eĥoŝanĝ\272 ĉiĵaŭde" // text with utf8 error
const sampleJson = ` {"data":"\u2714","origin": "...", "rows":[ ]}` // valid
const sampleJsoff = `{"data":"...…", "origin": "\x2714"}` // invalid
const sampleHtml = `<?xml version="1.0"?><html>
<h1>hell☺</h1><pre><span class=""><!-- HTM&#x4C; --></span>
` // greeting with some tags
const xmlDeclare = `<?xml version='1.0' encoding='us-ascii'?>`
const sampleXml  = xmlDeclare + `
<!-- copied from https://httpbin.org/xml -->
<slideshow author="Yours Truly" title="Sample Slide Show">
<slide type="all"><title>Wake up to WonderWidgets!</title></slide>
<slide/>
</slideshow>
`

func httpResult(status int, body string) (r *Request) {
	w := &httptest.ResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      bytes.NewBufferString(body),
		Code:      status,
	}
	w.HeaderMap.Add("Content-Length", strconv.Itoa(len(body)))
	r = New()
	r.Request = nil
	r.Response = w.Result()
	return
}

func TestRequestEmpty(t *testing.T) {
	r := httpResult(204, "")
	body, err := r.Text()
	if err != nil {
		t.Fatalf("could not simulate download: %v", err)
	}
	if r.StatusCode != 204 {
		t.Fatalf("unexpected response: %v", r.Status)
	}
	if body != "" {
		t.Fatalf("unexpected download: %v", body)
	}
}

func TestRequestStatus(t *testing.T) {
	r := httpResult(404, sampleHtml)
	body, err := r.Text()

	if err == nil || r.StatusCode != 404 {
		t.Fatalf("error status not reported: %v", r.Status)
	}
	var e *StatusError
	if !errors.As(err, &e) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if v := e.Code; v != 404 {
		t.Fatalf("unexpected status in error: %v", v)
	}
	if body != sampleHtml {
		t.Fatalf("unexpected download: %v", body)
	}
}

func TestRequestBytes(t *testing.T) {
	r := httpResult(200, sampleData)
	body, err := r.Bytes()

	if err != nil {
		t.Fatalf("unexpected download error: %v", err)
	}
	if string(body) != sampleData {
		t.Fatalf("missing invalid download: %v", body)
	}
}

func TestRequestUnicode(t *testing.T) {
	r := httpResult(200, sampleText)
	body, err := r.Text()

	if err != nil {
		t.Fatalf("could not simulate download: %v", err)
	}
	if body != sampleText {
		t.Fatalf("unexpected download: %v", body)
	}
}

func TestRequestInvalidUnicode(t *testing.T) {
	r := httpResult(200, sampleData)
	body, err := r.Text()

	if err == nil {
		t.Fatalf("unexpected download success: %v", body)
	}
	if err != ErrTextInvalid {
		t.Fatalf("unexpected download error: %v", err)
	}
	if body != sampleData {
		t.Fatalf("missing invalid download: %v", body)
	}
}

func TestRequestJson(t *testing.T) {
	r := httpResult(200, sampleJson)
	var res HttpbinEcho
	err := r.Json(&res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := res.Data; v != "✔" {
		t.Fatalf("unexpected payload: %v", v)
	}
}

func TestRequestJsonError(t *testing.T) {
	r := httpResult(200, sampleJsoff)
	var res HttpbinEcho
	err := r.Json(&res)
	// expect `invalid character 'x' in string escape code`
	// without any partial data
	if err == nil {
		t.Fatalf("uncaught syntax error: %v", res)
	}
	if v := res.Origin; v != "" {
		t.Fatalf("unexpected payload: %v", v)
	}

	var syntaxerr *json.SyntaxError
	if !errors.As(err, &syntaxerr) {
		t.Fatalf("unexpected error type: %T", err)
	}
	if v := syntaxerr.Offset; v != 31 {
		t.Fatalf("unexpected error offset: %v", v)
	}
}

func TestRequestJsonXml(t *testing.T) {
	r := httpResult(200, sampleXml)
	var res HttpbinEcho
	err := r.Json(&res)
	if err == nil {
		t.Fatalf("unexpected success: %v", res)
	}
	if err != ErrJsonLikeXml {
		t.Fatalf("unexpected error: %v", err)
	}
	if v := r.Preview(); v != xmlDeclare+"..." {
		t.Fatalf("unexpected preview: %v", v)
	}
}

func TestRequestJsonMojibake(t *testing.T) {
	s := strings.Replace(sampleJson, ".", "\205", 2) // invalid utf8
	s = strings.Replace(s, "rows", "url", 1) // array into Url string
	r := httpResult(200, s)
	var res HttpbinEcho
	err := r.Json(&res)
	// expect errors.joinError of `response body contains invalid UTF-8`
	// and `cannot unmarshal array into Go struct`... with partial data

	if err == nil {
		t.Fatalf("unexpected download success: %v", res)
	}
	if v := res.Data; v != "✔" {
		t.Fatalf("missing partial data: %v", v)
	}
	if v := res.Origin; v != "\uFFFD\uFFFD." {
		t.Fatalf("unexpected results: %v", v)
	}

	if !errors.Is(err, ErrTextInvalid) {
		t.Fatalf("missing unicode error: %v", err)
	}
	var syntaxerr *json.UnmarshalTypeError
	if !errors.As(err, &syntaxerr) {
		t.Fatalf("missing json type error: %T", err)
	}
	if v := syntaxerr.Field; v != "Url" {
		t.Fatalf("unexpected error field: %s", v)
	}
}

func TestRequestXml(t *testing.T) {
	r := httpResult(200, sampleXml)
	var res struct {
		Title string `xml:"title,attr"`
	}
	err := r.Xml(&res)
	if err != nil {
		t.Fatalf("unexpected download error: %v", err)
	}
	expect := `Sample Slide Show`
	if res.Title != expect {
		t.Fatalf("unexpected xml results for <slideshow title />: %v", res.Title)
	}
}

func TestRequestXmlEmpty(t *testing.T) {
	r := httpResult(200, "")
	var res struct {}
	err := r.Xml(&res)
	if err != ErrBodyEmpty {
		t.Fatalf("unexpected error: %v", err)
	}
}
