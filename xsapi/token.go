package xsapi

import (
	"net/http"
)

type Token interface {
	SetAuthHeader(req *http.Request)
}

type TokenSource interface {
	Token() (Token, error)
}

type DisplayClaimer interface {
	DisplayClaims() DisplayClaims
}

type DisplayClaims struct {
	GamerTag string `json:"gtg"`
	XUID     string `json:"xid"`
	UserHash string `json:"uhs"`
}
