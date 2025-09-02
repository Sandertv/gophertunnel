package auth

import "github.com/df-mc/go-playfab"

type XBOXPlayfabLoginConfig struct {
	playfab.LoginConfig
	XBOXToken string `json:"XboxToken,omitempty"`
}

func (l XBOXPlayfabLoginConfig) Login(x *XBLToken) (*playfab.Identity, error) {
	l.XBOXToken = "XBL3.0 x=" + x.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash + ";" + x.AuthorizationToken.Token
	return l.LoginConfig.Login("/Client/LoginWithXbox", l)
}
