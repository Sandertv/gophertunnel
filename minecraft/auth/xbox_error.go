package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// XboxError represents a structured error from Xbox Live authentication.
type XboxError struct {
	// URL is the endpoint that was called when the error occurred
	URL string
	// Method is the HTTP method used (usually POST)
	Method string
	// StatusCode is the HTTP status code returned (if any)
	StatusCode int
	// Status is the HTTP status text
	Status string
	// XboxErrorCode is the custom error code from the x-err header
	XboxErrorCode string
	// ResponseBody contains the raw response body for debugging
	ResponseBody string
	// Underlying is the underlying error that caused this (network errors, etc.)
	Underlying error
}

// AccountCreationRequiredError is returned when the Microsoft account has no Xbox profile.
// SignupURL opens Microsoft's hosted flow that creates and links the Xbox profile to the
// already-authenticated Microsoft account.
type AccountCreationRequiredError struct {
	*XboxError
	SignupURL *url.URL
}

// Error implements the error interface.
func (e *AccountCreationRequiredError) Error() string {
	if e == nil || e.XboxError == nil {
		return "xbox account creation required"
	}
	return e.XboxError.Error()
}

// Unwrap returns the underlying Xbox error for errors.As compatibility.
func (e *AccountCreationRequiredError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.XboxError
}

// Error implements the error interface.
func (e *XboxError) Error() string {
	var parts []string

	if e.Method != "" && e.URL != "" {
		parts = append(parts, fmt.Sprintf("%s %s", e.Method, e.URL))
	}

	if e.XboxErrorCode != "" {
		// Use the parsed error message for Xbox-specific errors
		parts = append(parts, parseXboxErrorCode(e.XboxErrorCode))
	} else if e.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status))
	}

	if e.Underlying != nil {
		parts = append(parts, e.Underlying.Error())
	}

	return strings.Join(parts, ": ")
}

// Unwrap returns the underlying error for error unwrapping.
func (e *XboxError) Unwrap() error {
	return e.Underlying
}

// IsNetworkError returns true if this error was caused by a network issue.
func (e *XboxError) IsNetworkError() bool {
	return e.Underlying != nil && e.StatusCode == 0
}

// IsXboxSpecificError returns true if this is a known Xbox Live error code.
func (e *XboxError) IsXboxSpecificError() bool {
	return e.XboxErrorCode != ""
}

// GetParsedXboxError returns the human-readable Xbox error message if available.
func (e *XboxError) GetParsedXboxError() string {
	if e.XboxErrorCode == "" {
		return ""
	}
	return parseXboxErrorCode(e.XboxErrorCode)
}

// newXboxError creates a new XboxError for network-related failures.
func newXboxNetworkError(method, url string, err error, responseBody []byte) *XboxError {
	return &XboxError{
		Method:       method,
		URL:          url,
		Underlying:   err,
		ResponseBody: string(responseBody),
	}
}

// newXboxHTTPError creates a new XboxError for HTTP response errors.
func newXboxHTTPError(method, url string, resp *http.Response, responseBody []byte) *XboxError {
	xboxErr := &XboxError{
		Method:       method,
		URL:          url,
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		ResponseBody: string(responseBody),
	}

	// Check for Xbox-specific error code in headers
	if errorCode := resp.Header.Get("x-err"); errorCode != "" {
		xboxErr.XboxErrorCode = errorCode
	}

	return xboxErr
}

func newAccountCreationRequiredError(xboxErr *XboxError, header http.Header, responseBody []byte, device *deviceToken) (*AccountCreationRequiredError, error) {
	if xboxErr == nil || xboxErr.XboxErrorCode != "2148916233" {
		return nil, fmt.Errorf("xbox error is not account creation required")
	}
	sessionID := header.Get("X-SessionId")
	if sessionID == "" {
		return nil, fmt.Errorf("header X-SessionId is absent from authorization response")
	}
	if device == nil || device.proofKey == nil {
		return nil, fmt.Errorf("device token or proof key is absent")
	}
	deviceID := device.DisplayClaims.DeviceInfo.DeviceID
	if deviceID == "" {
		return nil, fmt.Errorf("device ID is absent from device token")
	}

	var r struct {
		WebPage string `json:"WebPage"`
	}
	if err := json.Unmarshal(responseBody, &r); err != nil {
		return nil, fmt.Errorf("decode authorization response: %w", err)
	}
	if r.WebPage == "" {
		return nil, fmt.Errorf("account creation page URL is absent from authorization response")
	}

	signingURL := (&url.URL{
		Scheme: "https",
		Host:   "sisu.xboxlive.com",
		Path:   "/proxy",
	}).String()
	signingRequest, err := http.NewRequest(http.MethodPost, signingURL+"?sessionid="+url.QueryEscape(sessionID), nil)
	if err != nil {
		return nil, fmt.Errorf("make request for computing signature: %w", err)
	}
	if err := sign(signingRequest, nil, device.proofKey); err != nil {
		return nil, fmt.Errorf("sign account creation request: %w", err)
	}
	signature := signingRequest.Header.Get("Signature")
	if signature == "" {
		return nil, fmt.Errorf("signature is absent from account creation request")
	}

	signupURL, err := url.Parse(r.WebPage)
	if err != nil {
		return nil, fmt.Errorf("parse account creation page URL: %w", err)
	}
	query := signupURL.Query()
	query.Set("sig", signature)
	query.Set("did", "0x"+deviceID)
	query.Set("redirect", "https://sisu.xboxlive.com/sisu_desktop.srf")
	query.Set("sid", sessionID)
	signupURL.RawQuery = query.Encode()

	return &AccountCreationRequiredError{
		XboxError: xboxErr,
		SignupURL: signupURL,
	}, nil
}

// parseXboxError returns the message associated with an Xbox Live error code.
func parseXboxErrorCode(code string) string {
	switch code {
	case "2148916227":
		return "Your account was banned by Xbox for violating one or more Community Standards for Xbox and is unable to be used."
	case "2148916229":
		return "Your account is currently restricted and your guardian has not given you permission to play online. Login to https://account.microsoft.com/family/ and have your guardian change your permissions."
	case "2148916233":
		return "Your account currently does not have an Xbox profile. Please create one at https://signup.live.com/signup"
	case "2148916234":
		return "Your account has not accepted Xbox's Terms of Service. Please login and accept them."
	case "2148916235":
		return "Your account resides in a region that Xbox has not authorized use from. Xbox has blocked your attempt at logging in."
	case "2148916236":
		return "Your account requires proof of age. Please login to https://login.live.com/login.srf and provide proof of age."
	case "2148916237":
		return "Your account has reached its limit for playtime. Your account has been blocked from logging in."
	case "2148916238":
		return "The account date of birth is under 18 years and cannot proceed unless the account is added to a family by an adult."
	default:
		return fmt.Sprintf("unknown error code: %v", code)
	}
}
