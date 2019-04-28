package jwt

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
