package httpclient

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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

func (r *Request) Json(serial any) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}
	if len(body) > 0 && body[0] == '<' {
		return fmt.Errorf("initial '<' indicates xml not json") // TODO preview
	}
	return json.Unmarshal(body, serial)
}

func (r *Request) Xml(serial any) error {
	if err := r.Receive(); err != nil {
		return err
	}
	d := xml.NewDecoder(r.Response.Body)
	return d.Decode(&serial)
}
