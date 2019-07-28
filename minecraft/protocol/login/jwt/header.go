package jwt

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

// allowedAlgorithm is a slice of allowed algorithms.
var allowedAlgorithms []string

// AllowAlg adds a variadic amount of algorithms which JWT headers may have. Any algorithm found in a JWT
// header will be assumed as an error and returned during verification.
func AllowAlg(alg ...string) {
	allowedAlgorithms = append(allowedAlgorithms, alg...)
}

// AllowedAlg checks if the algorithm passed has been allowed using a call to AllowAlg().
func AllowedAlg(algorithm string) bool {
	for _, alg := range allowedAlgorithms {
		if alg == algorithm {
			return true
		}
	}
	return false
}

// Header holds the header information of a JWT claim.
type Header struct {
	// Algorithm is the algorithm used for the signature in the signature section of the claim. Any algorithm
	// that isn't allowed using AllowAlg will result in an error instead.
	Algorithm string `json:"alg"`
	X5U       string `json:"x5u"`
}

// Header parses the JWT passed and returns the base64 decoded header section of the claim. The JSON data
// returned is not guaranteed to be valid JSON.
func HeaderFrom(jwt []byte) ([]byte, error) {
	fragments := bytes.Split(jwt, []byte{'.'})
	if len(fragments) != 3 {
		return nil, fmt.Errorf("expected claim to have 3 sections, but got %v", len(fragments))
	}
	// Some (faulty) JWT implementations use padded base64, whereas it should be raw. We trim this off.
	fragments[0] = bytes.TrimRight(fragments[0], "=")
	payload, err := base64.RawURLEncoding.DecodeString(string(fragments[0]))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding payload: %v (%v)", err, fragments[0])
	}
	return payload, nil
}
