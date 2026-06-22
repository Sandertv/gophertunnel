package messaging

import (
	"errors"
	"io"
	"testing"
)

func TestCredentialsError(t *testing.T) {
	err := &CredentialsError{Method: MethodSignalingCredentials, Err: io.EOF}

	if got, want := err.Error(), `call "Signaling_TurnAuth_v1_0": EOF`; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrCredentialsCall) {
		t.Fatal("expected errors.Is to match ErrCredentialsCall")
	}
	if !errors.Is(err, io.EOF) {
		t.Fatal("expected errors.Is to match wrapped EOF")
	}

	var credentialsErr *CredentialsError
	if !errors.As(err, &credentialsErr) {
		t.Fatal("expected errors.As to match CredentialsError")
	}
	if credentialsErr.Method != MethodSignalingCredentials {
		t.Fatalf("Method = %q, want %q", credentialsErr.Method, MethodSignalingCredentials)
	}
}
