package s3

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

// Transport makes an ordinary https request, but it always adds
// a Date header (if missing) and an S3 signature.
// It treats all requests as https.
type Transport struct {
	// The Keys used to sign requests.
	// This must be set before use; see KeysFromEnvironment.
	Keys *Keys

	// The Service used to sign requests.
	// If nil, uses DefaultService.
	Service *Service

	// The underlying RoundTripper used to execute the request.
	// If nil, uses http.DefaultTransport.
	Transport http.RoundTripper
}

// On startup, DefaultTransport is registered under the URL scheme "s3"
// in http.DefaultTransport.
var DefaultTransport = &Transport{}

func init() {
	t, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		t.RegisterProtocol("s3", DefaultTransport)
	}
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.Keys == nil {
		return nil, errors.New("s3: uninitialized keys")
	}
	r1 := new(http.Request)
	*r1 = *r // includes shallow copies of maps, but okay
	r1.URL = new(url.URL)
	*r1.URL = *r.URL

	r1.URL.Scheme = "https"
	r1.Header = make(http.Header)
	copyHeader(r1.Header, r.Header)
	if r1.Header.Get("Date") == "" {
		r1.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}
	service := t.Service
	if service == nil {
		service = DefaultService
	}
	service.Sign(r1, *t.Keys)
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(r1)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
