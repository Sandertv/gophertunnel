package internal

// Result wraps the basic structure of response body sent by most franchise-related services.
// Make sure to specify the T generic type to whatever you want in the Data.
type Result[T any] struct {
	// Data is the payload of the result, if successful.
	Data T `json:"result"`

	// I'm not sure if there was an error result, but I've seen errors in HTTP status code with
	// an empty response body rather than a body describing the error so this isn't the case.
}
