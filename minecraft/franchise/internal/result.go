package internal

// Result wraps the basic structure of response body sent by most franchise-related services.
// Make sure to specify the T generic type to whatever you want in the Data.
type Result[T any] struct {
	// Data is the payload of the result, if successful.
	Data T `json:"result"`
}
