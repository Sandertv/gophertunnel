package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
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
	index := bytes.Index(body, []byte("var ServerData = "))
	offset := len("var ServerData = ")
	obj := parseJSObject(body[index+offset:])

	// m is a container of the flowtoken. It is in an XML attribute, so we first unmarshal it in this
	// container struct.
	m := struct {
		Value string `xml:"value,attr"`
	}{}
	if err := xml.Unmarshal([]byte(obj["sFTTag"]), &m); err != nil {
		return nil, fmt.Errorf("error decoding flowtoken XML container: %v", err)
	}

	flowToken := m.Value

	var credentialTypeURL, uaid string
	for _, value := range obj {
		// The field that holds the GetCredentialType URL differs each time, so we need to loop through all
		// fields and find the one that holds it.
		if strings.HasPrefix(value, "https://login.live.com/GetCredentialType.srf") {
			credentialTypeURL = value
		}
	}
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
	request, _ = http.NewRequest("POST", obj["urlPost"], strings.NewReader(postData.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	transferCookies(request, resp)

	resp, err = c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("POST %v: %v", obj["urlPost"], err)
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

// parseJSObject is used to parse a javascript object from the data slice passed. It is used to parse data
// found in the Server data object found in the first login request payload.
func parseJSObject(data []byte) map[string]string {
	buf := bytes.NewBuffer(data[1:])
	reader := bufio.NewReader(buf)
	m := make(map[string]string)
	for {
		if bytes.Index(buf.Bytes(), []byte{'}'}) < bytes.Index(buf.Bytes(), []byte{':'}) {
			break
		}
		name, err := reader.ReadString(':')
		if err != nil {
			break
		}
		name = name[:len(name)-1]

		value, err := reader.ReadString(',')
		if err != nil {
			break
		}
		// Remove the comma at the end of the value and trim the quotes away on the outsides of the value so
		// that we obtain a clean string.
		m[name] = strings.Trim(value[:len(value)-1], `'`)
	}
	return m
}
