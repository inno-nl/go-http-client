/*
Replacement HTTP client, building on core net/http objects
but with a greatly simplified interface.

[New] or [NewURL] prepare an outgoing request:

	c := httpclient.NewURL("http://localhost")
	r := c.NewURL("endpoint")

Then any of [Send], [Preview], [Bytes], [Text], [Json] or [Xml]
to send and process depending on the wanted format:

	body, err := r.Text()
*/
package httpclient

import (
	"net/http"
	"time"
)

const DefaultAgent = "inno-go-http-client/2"

// Combined [Client], [Request] and [Response],
// containing everything to prepare and download a HTTP request.
// Should be setup by either [New] or [NewURL]
// and then altered and executed by its methods.
type Request struct {
	Error error // Setup exceptions postponed until [Send].

	// Shared [http.Client] with common transportation details
	// like cookies and timeouts.
	*http.Client
	// Current [http.Request] including URL and Headers.
	// Typically modified from a base Request.
	*http.Request
	// Received [http.Response] populated by [Send].
	*http.Response

	// Override to check [Send] attempts for temporary exceptions.
	DoRetry func(*Request, error) error
	Attempt int // Do() counter in [Send]
	Tries   int // retry Do() if more than 1
}

// Initialise a new [Request] with a default user agent.
// Can be further prepared by client options and common setup,
// before being cloned by [NewURL] for specific downloads.
func New() (r *Request) {
	r = new(Request)
	r.Client = &http.Client{}

	// insufficient data for http.NewRequest()
	r.Request = &http.Request{
		Header: http.Header{"User-Agent": {DefaultAgent}},
	}
	return
}

// Create a new [Request] initialised with an URL.
// This can either be a complete link ready to be downloaded,
// or a root path to be extended by endpoints and parameters.
func NewURL(ref string) (r *Request) {
	r = New()
	_ = r.AddURL(ref) // invalid results reported by Client.Do()
	return
}

// Clone a Request with its URL replaced or appended
// by [AddURL]ing the given URI fragment.
func (r *Request) NewURL(ref string) (d *Request) {
	d = r.Clone()
	_ = d.AddURL(ref) // keep ignoring parse errors until Send()
	return
}

// Make a deep copy of all contained objects,
// so the original will not be affected by further use.
func (r *Request) Clone() *Request {
	d := new(Request)
	*d = *r
	if d.Client != nil {
		d.Client = new(http.Client)
		*d.Client = *r.Client
	}
	if d.Request != nil {
		d.Request = r.Request.Clone(r.Context())
	}
	// Response will be reset by Send()
	return d
}

// Send the prepared HTTP request, possibly retrying on server errors.
// Saves a [Response] of query results, but does not download contents yet,
// expecting manual intervention such as checking [StatusCode].
func (r *Request) Send() error {
	r.Response = nil
	r.Attempt = 0
	return r.Resend()
}

func (r *Request) Resend() (err error) {
	if err = r.Error; err != nil {
		return
	}

	delay := time.Second
	for r.Attempt++; ; r.Attempt++ {
		r.Response, err = r.Client.Do(r.Request)
		if r.Attempt >= r.Tries {
			break
		}
		if r.DoRetry != nil {
			err = r.DoRetry(r, err)
		} else if err == nil && r.StatusCode >= 500 {
			err = &StatusError{r.Response.StatusCode, r.Response.Status}
		}
		if err == nil {
			break
		}
		time.Sleep(delay)
		delay *= 2 // increase exponentially
	}
	return
}
