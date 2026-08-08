package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// Err decodes the response body as an Error if possible. Otherwise, the status code is returned
// along with a truncated body for debugging.
func Err(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	e := new(Error)
	if err := json.Unmarshal(body, e); err != nil {
		if len(body) > 0 {
			return fmt.Errorf("%s %s: %s: %q", resp.Request.Method, resp.Request.URL, resp.Status, body)
		}
		return fmt.Errorf("%s %s: %s", resp.Request.Method, resp.Request.URL, resp.Status)
	}
	return e
}

// UserAgent is always set as a 'User-Agent' header to the request, and indicates that the
// request is made by libHttpClient, which is a primary HTTP client bundled in XSAPI/GDK.
const UserAgent = "libhttpclient/1.0.0.0"

// ServiceEnvironment represents an environment for a service that hosts an endpoint
// on the [ServiceEnvironment.ServiceURI].
type ServiceEnvironment struct {
	ServiceURI *url.URL `json:"serviceUri"`
}

// UnmarshalJSON implements [json.Unmarshaler.UnmarshalJSON].
// Since [url.URL] does not implement [json.Unmarshaler] or [encoding.TextUnmarshaler],
// we decode URL fields as strings first, then manually parse them into URLs.
// See: https://github.com/golang/go/issues/52638
func (e *ServiceEnvironment) UnmarshalJSON(b []byte) error {
	type Alias ServiceEnvironment
	data := struct {
		*Alias
		ServiceURI string `json:"serviceUri"`
		Issuer     string `json:"issuer"`
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if data.ServiceURI == "" {
		return errors.New("service/internal: ServiceEnvironment.ServiceURI is empty")
	}
	var err error
	e.ServiceURI, err = url.Parse(data.ServiceURI)
	if err != nil {
		return fmt.Errorf("parse ServiceURI: %w", err)
	}
	return nil
}
