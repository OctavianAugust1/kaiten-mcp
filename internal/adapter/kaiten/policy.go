package kaiten

import (
	"errors"
	"net/http"
)

// ErrNonGETRequest is returned before the underlying network transport sees a
// mutating or otherwise unapproved HTTP method.
var ErrNonGETRequest = errors.New("kaiten transport permits GET requests only")

// ReadOnlyTransport is a defense-in-depth policy around the HTTP transport.
type ReadOnlyTransport struct {
	next http.RoundTripper
}

// NewReadOnlyTransport wraps next with a strict GET-only method check.
func NewReadOnlyTransport(next http.RoundTripper) *ReadOnlyTransport {
	if next == nil {
		next = http.DefaultTransport
	}
	return &ReadOnlyTransport{next: next}
}

// RoundTrip rejects non-GET requests before delegating to the network.
func (transport *ReadOnlyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request == nil || request.Method != http.MethodGet {
		return nil, ErrNonGETRequest
	}
	return transport.next.RoundTrip(request)
}
