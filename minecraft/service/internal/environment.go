package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// ServiceEnvironment represents an environment for a service that hosts an endpoint
// on the [ServiceEnvironment.ServiceURI].
//
// Structs embedding this struct must be aware that, since it implements UnmarshalJSON,
// other fields outside of this struct may not be unmarshalled correctly. Therefore, it
// should be used for service environments where 'serviceUri' is the only key.
type ServiceEnvironment struct {
	// ServiceURI is the base URL for the service endpoint.
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
