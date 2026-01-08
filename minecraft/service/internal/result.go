package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Result wraps the basic structure of response body sent by most franchise-related services.
// Make sure to specify the T generic type to whatever you want in the Data.
type Result[T any] struct {
	// Data is the payload of the result, if successful.
	Data T `json:"result"`
}

// Error represents an error response body that are used by various Minecraft services.
// Clients making API requests to the endpoints normally should check for the status code
// then attempt to decode the JSON response body as an Error to return a detailed error to the user.
type Error struct {
	// Namespace indicates the namespace responsible for returning this error.
	Namespace string `json:"namespace"`
	// Code describes the error code.
	Code string `json:"code"`
	// Message is the human-readable message for the Error.
	Message string `json:"message"`
	// CustomData contains data specific to the Error. It is typically empty for most errors.
	CustomData json.RawMessage `json:"customData"`
}

// Error returns a string representation of the error.
func (e *Error) Error() string {
	return fmt.Sprintf("minecraft/service: %s: %q (%s)", e.Code, e.Message, e.Namespace)
}

// Err decodes the response body as an Error if possible. Otherwise, the status code is returned.
func Err(resp *http.Response) error {
	e := new(Error)
	if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
		return fmt.Errorf("%s %s: %s", resp.Request.Method, resp.Request.URL, resp.Status)
	}
	return e
}

// UserAgent is always set as a 'User-Agent' header to the request, and indicates that the
// request is made by libHttpClient, which is a primary HTTP client bundled in XSAPI/GDK.
const UserAgent = "libhttpclient/1.0.0.0"
