package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

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
	if err != nil {
		return
	}
	defer r.Response.Body.Close()
	out, err = io.ReadAll(r.Response.Body)
	return
}

func (r *Request) String() (string, error) {
	body, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Specific error message given if Json() encounters an xml body (for example
// an unexpected html error page), instead of a generic json.Unmarshal error:
// `invalid character '<' looking for beginning of value`
var ErrJsonLikeXml = fmt.Errorf("initial '<' indicates xml not json")

func (r *Request) Json(serial any) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}
	if len(body) > 0 && body[0] == '<' {
		return ErrJsonLikeXml // TODO preview
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
