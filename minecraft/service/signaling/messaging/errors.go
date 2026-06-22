package messaging

import (
	"errors"
	"fmt"
)

// ErrCredentialsCall identifies failures while calling the signaling service for TURN credentials.
var ErrCredentialsCall = errors.New("signaling/messaging: credentials call failed")

// CredentialsError wraps a failed credentials JSON-RPC call.
type CredentialsError struct {
	// Method is the JSON-RPC method used for the failed credentials request.
	Method string
	// Err is the underlying call failure.
	Err error
}

func (e *CredentialsError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return fmt.Sprintf("call %q", e.Method)
	}
	return fmt.Sprintf("call %q: %v", e.Method, e.Err)
}

// Unwrap returns the underlying credentials call failure.
func (e *CredentialsError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is reports whether target is [ErrCredentialsCall].
func (e *CredentialsError) Is(target error) bool {
	return target == ErrCredentialsCall
}
