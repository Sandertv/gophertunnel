package authclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestSendRequestWithRetriesRetriesTransientStatus(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := SendRequestWithRetries(context.Background(), server.Client(), req, RetryOptions{
		Attempts: 2,
		MinDelay: time.Millisecond,
		MaxDelay: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts mismatch: got %d want 2", got)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status mismatch: got %d", resp.StatusCode)
	}
}

func TestSendRequestWithRetriesReturnsFinalRetryableResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := SendRequestWithRetries(context.Background(), server.Client(), req, RetryOptions{Attempts: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if string(body) != "rate limited" {
		t.Fatalf("body mismatch: got %q", body)
	}
}

func TestSendRequestWithRetriesBindsAttemptToRetryContext(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := SendRequestWithRetries(ctx, server.Client(), req, RetryOptions{Attempts: 1})
	if resp != nil {
		_ = resp.Body.Close()
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if got := attempts.Load(); got != 0 {
		t.Fatalf("attempts mismatch: got %d want 0", got)
	}
}
