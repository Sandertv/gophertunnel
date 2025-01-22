package auth

import (
	"fmt"
	"net/http"
)

// XBLToken holds info on the authorization token used for authenticating with XBOX Live.
type XBLToken struct {
	AuthorizationToken string
}

// SetAuthHeader returns a string that may be used for the 'Authorization' header used for Minecraft
// related endpoints that need an XBOX Live authenticated caller.
func (t XBLToken) SetAuthHeader(r *http.Request) {
	r.Header.Set("Authorization", fmt.Sprintf("XBL3.0 x=%v", t.AuthorizationToken))
}

// NewXBLToken creates a new XBLToken with the given authorization token.
func NewXBLToken(token string) *XBLToken {
	return &XBLToken{
		AuthorizationToken: token,
	}
}
