package auth

import "github.com/df-mc/go-playfab"

type XboxPlayfabLoginConfig struct {
	playfab.LoginConfig
	XboxToken string `json:"XboxToken,omitempty"`
}

func (l XboxPlayfabLoginConfig) Login(x *XBLToken) (*playfab.Identity, error) {
	l.XboxToken = "XBL3.0 x=" + x.AuthorizationToken.DisplayClaims.UserInfo[0].UserHash + ";" + x.AuthorizationToken.Token
	return l.LoginConfig.Login("/Client/LoginWithXbox", l)
}
