package entity

import (
	"net/http"
	"time"
)

type Token struct {
	Entity     Key       `json:"Entity,omitempty"`
	Token      string    `json:"EntityToken,omitempty"`
	Expiration time.Time `json:"TokenExpiration,omitempty"`
}

func (tok *Token) Expired() bool                   { return time.Now().After(tok.Expiration) }
func (tok *Token) SetAuthHeader(req *http.Request) { req.Header.Set("X-EntityToken", tok.Token) }

type Key struct {
	ID   string `json:"Id,omitempty"`
	Type Type   `json:"Type,omitempty"`
}

type Type string

const (
	TypeNamespace           Type = "namespace"
	TypeTitle               Type = "title"
	TypeMasterPlayerAccount Type = "master_player_account"
	TypeTitlePlayerAccount  Type = "title_player_account"
	TypeCharacter           Type = "character"
	TypeGroup               Type = "group"
)
