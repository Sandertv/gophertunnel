package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// liveTokenURL is the URL that an access token may be retrieved from using several methods.
const liveTokenURL = `https://login.live.com/oauth20_token.srf`

// requestURL is the first URL that a GET request is made to in order to authenticate to Live.
const requestURL = `https://login.live.com/oauth20_authorize.srf?client_id=00000000441cc96b&redirect_uri=https://login.live.com/oauth20_desktop.srf&response_type=token&display=touch&scope=service::user.auth.xboxlive.com::MBI_SSL&locale=en`

// The flowToken seems like to have a fixed length
const flowTokenLen = 216

// RequestLiveToken does a login request for Microsoft Live using the login and password passed. If
// successful, a token containing the access token, refresh token, expiry and user ID is returned.
func RequestLiveToken(login, password string) (*TokenPair, error) {
	// We first create a new http client and send a request to the first URL.
	c := &http.Client{}
	resp, err := c.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("GET requestURL: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading GET requestURL body: %v", err)
	}
	_ = resp.Body.Close()
	// We're looking for a JavaScript object (not JSON) that holds data on the next URL that we need to
	// send a request to in order to be continue the authentication.
	sFTTagIndex := bytes.Index(body, []byte("sFTTag:"))
	if sFTTagIndex == -1 {
		return nil, fmt.Errorf("sFTTag not found in response body")
	}
	valueIndex := bytes.Index(body[sFTTagIndex:], []byte("value=\""))
	if valueIndex == -1 {
		return nil, fmt.Errorf("sFTTag value not found in response body")
	}
	valueIndex += sFTTagIndex + 7
	flowToken := string(body[valueIndex : valueIndex+flowTokenLen])

	credentialTypeURLPrefixIndex := bytes.Index(body, []byte("https://login.live.com/GetCredentialType.srf"))
	if credentialTypeURLPrefixIndex == -1 {
		return nil, fmt.Errorf("credentialTypeURL prefix not found in response body")
	}
	endIndex := bytes.Index(body[credentialTypeURLPrefixIndex:], []byte("',"))
	if endIndex == -1 {
		return nil, fmt.Errorf("credentialTypeURL end not found")
	}
	endIndex += credentialTypeURLPrefixIndex
	credentialTypeURL := string(body[credentialTypeURLPrefixIndex:endIndex])

	urlPostIndex := bytes.Index(body, []byte("urlPost:'"))
	if urlPostIndex == -1 {
		return nil, fmt.Errorf("urlPost not found in response body")
	}
	urlPostIndex += 9
	endIndex = bytes.Index(body[urlPostIndex:], []byte("',"))
	if endIndex == -1 {
		return nil, fmt.Errorf("urlPost end not found")
	}
	endIndex += urlPostIndex
	urlPost := string(body[urlPostIndex:endIndex])

	var uaid string
	for _, cookie := range resp.Cookies() {
		// We need the uaid for the next request we send, so we look through the cookies of the last response
		// and search the cookie that holds the uaid.
		if cookie.Name == "uaid" {
			uaid = cookie.Value
		}
	}

	// The next POST request is the request that issues the username.
	jsonData, _ := json.Marshal(map[string]interface{}{
		"username":             login,
		"uaid":                 uaid,
		"isOtherIdpSupported":  "false",
		"checkPhones":          false,
		"isRemoteNGCSupported": true,
		"isCookieBannerShown":  false,
		"isFidoSupported":      false,
		"flowToken":            flowToken,
	})
	request, _ := http.NewRequest("POST", credentialTypeURL, bytes.NewReader(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	transferCookies(request, resp)
	resp, err = c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", credentialTypeURL, err)
	}
	_ = resp.Body.Close()

	// Don't allow redirecting on the next POST request. We need the Location data from that request, which
	// we need.
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	postData := url.Values{
		"login":        []string{login},
		"passwd":       []string{password},
		"PPFT":         []string{flowToken},
		"PPSX":         []string{"P"},
		"SI":           []string{"Sign in"},
		"type":         []string{"11"},
		"NewUser":      []string{"1"},
		"LoginOptions": []string{"1"},
	}
	request, _ = http.NewRequest("POST", urlPost, strings.NewReader(postData.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	transferCookies(request, resp)

	resp, err = c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", urlPost, err)
	}
	_ = resp.Body.Close()
	location, err := resp.Location()
	if err != nil {
		return nil, fmt.Errorf("final response had no location: %v", err)
	} else if location.String() == "" {
		// The Location was empty, meaning authentication failed.
		return nil, fmt.Errorf("incorrect login or password passed")
	}

	location.RawQuery = location.Fragment
	// Finally we use the Location data to find the access and refresh tokens.
	values := location.Query()

	t, _ := strconv.Atoi(values.Get("expires_in"))
	c.CloseIdleConnections()
	return NewTokenPair(values.Get("access_token"), values.Get("refresh_token"), time.Duration(t)), nil
}

// transferCookies transfers the cookies from the previous http response to a request and sets the referer
// field in the header of the request to the URL of the previous response.
func transferCookies(request *http.Request, previous *http.Response) {
	request.Header.Set("Referer", previous.Request.URL.String())
	for _, cookie := range previous.Cookies() {
		request.AddCookie(cookie)
	}
}
