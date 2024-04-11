package httpclient

import (
	"testing"

	"bytes"
	"net/http"
	"net/http/httptest"
)

const sampleText = "Eĥoŝanĝº ĉiĵaŭde" // valid unicode
const sampleData = "Eĥoŝanĝ\272 ĉiĵaŭde" // text with utf8 error
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
	r = New()
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

	expect := "unsuccessful response code 404 Not Found"
	if err == nil || r.StatusCode != 404 {
		t.Fatalf("error status not reported: %v", r.Status)
	}
	if err.Error() != expect {
		t.Fatalf("unexpected download error: %v", err)
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

func TestRequestJsonError(t *testing.T) {
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
