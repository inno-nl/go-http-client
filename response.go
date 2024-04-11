package httpclient

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func (r *Request) Success() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

func (r *Request) Receive() (err error) {
	if r.Response == nil {
		err = r.Send()
		if err != nil {
			return
		}
		if r.Response == nil {
			panic("missing response")
		}
	}
	if !r.Success() {
		err = fmt.Errorf("unsuccessful response code %s", r.Status)
	}
	return
}

func (r *Request) Bytes() (out []byte, err error) {
	err = r.Receive()
	if r.Response == nil {
		return
	}
	defer r.Response.Body.Close()
	out, ioerr := io.ReadAll(r.Response.Body)
	if err == nil {
		err = ioerr
	}
	return
}

// Error given by Text() if the body does not seem to be proper Unicode.
// If a different encoding is expected, use Bytes() instead to get raw data
// without this sanity check.
var ErrTextInvalid = fmt.Errorf("response body contains invalid UTF-8")

func (r *Request) Text() (string, error) {
	body, err := r.Bytes()
	if err == nil && !utf8.Valid(body) {
		err = ErrTextInvalid
	}
	return string(body), err
}

// Specific error message given if Json() encounters an xml body (for example
// an unexpected html error page), instead of a generic json.Unmarshal error:
// `invalid character '<' looking for beginning of value`
//
// In such cases, use Preview() for further debugging:
//
//	err := r.Json(&res)
//	if err == httpclient.ErrJsonLikeXml {
//		fmt.Printf("failed to get %s: %v (%s)", r.URL, err, r.Preview())
//	}
var ErrJsonLikeXml = fmt.Errorf("initial '<' indicates xml not json")

func (r *Request) Json(serial any) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}
	if len(body) > 0 && body[0] == '<' {
		buf := bytes.NewReader(body)
		r.Response.Body = io.NopCloser(buf) // copy for rereading
		return ErrJsonLikeXml
	}
	return json.Unmarshal(body, serial)
}

func (r *Request) Xml(serial any) error {
	if err := r.Receive(); err != nil {
		return err
	}
	d := xml.NewDecoder(r.Response.Body)
	d.CharsetReader = func(xmlenc string, in io.Reader) (out io.Reader, err error) {
		// support for some common non-utf8 encoding declarations
		switch strings.ToLower(xmlenc) {
		case "us-ascii":
			out = in // subset of utf-8
		case "iso-8859-1", "windows-1252":
			out = charmap.Windows1252.NewDecoder().Reader(in)
		default:
			err = fmt.Errorf("unrecognised by httpclient.Xml")
		}
		return
	}
	return d.Decode(&serial)
}

// Abbreviate the first line of a response text,
// usually after failure to unmarshal Json()
// to give an indication of any retrieved (error) message instead.
func (r *Request) Preview() (body string) {
	body, _ = r.Text()
	if body == "" {
		return
	}
	cut := 160 // maximum length
	if eol := strings.IndexByte(body, '\n'); eol > 0 && eol < cut {
		// reduce length to first line ending
		cut = eol + 3
	}
	if cut < len(body) {
		// abbreviate exceeding line to truncated start and marker
		body = body[:cut-3] + "..."
	}
	return
}
