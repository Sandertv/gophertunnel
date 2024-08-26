package franchise

import (
	"errors"
	"fmt"
	"net/http"
)

type Transport struct {
	IdentityProvider IdentityProvider
	Base             http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyClosed := false
	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				_ = req.Body.Close()
			}
		}()
	}

	if t.IdentityProvider == nil {
		return nil, errors.New("minecraft/franchise: Transport: IdentityProvider is nil")
	}
	config, err := t.IdentityProvider.TokenConfig()
	if err != nil {
		return nil, fmt.Errorf("request token config: %w", err)
	}
	token, err := config.Token()
	if err != nil {
		return nil, fmt.Errorf("request token: %w", err)
	}

	req2 := cloneRequest(req)
	token.SetAuthHeader(req2)

	reqBodyClosed = true
	return t.base().RoundTrip(req2)
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
