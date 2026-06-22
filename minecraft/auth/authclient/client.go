package authclient

import (
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

type RetryOptions struct {
	Attempts int           // how many times to send the request (default: 3)
	Factor   float64       // factor to multiply the delay by on each attempt (default: 2.0)
	MinDelay time.Duration // minimum delay (default: 500ms)
	MaxDelay time.Duration // Maximum delay (default: 8s)
}

// SendRequestWithRetries sends a request and retries on 429, 5xx and network errors.
func SendRequestWithRetries(ctx context.Context, c *http.Client, request *http.Request, r ...RetryOptions) (*http.Response, error) {
	var opts RetryOptions
	if len(r) > 0 {
		opts = r[0]
	}
	if opts.Attempts <= 0 {
		opts.Attempts = 3
	}
	if opts.Factor <= 0 {
		opts.Factor = 2.0
	}
	if opts.MinDelay <= 0 {
		opts.MinDelay = 500 * time.Millisecond
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 8 * time.Second
	}
	if opts.MaxDelay < opts.MinDelay {
		opts.MaxDelay = opts.MinDelay
	}

	var resp *http.Response
	var err error
	var retryAfterDelay time.Duration

	for i := range opts.Attempts {
		if i > 0 {
			delay := min(opts.MinDelay*time.Duration(math.Pow(opts.Factor, float64(i))), opts.MaxDelay)

			// Use any retry-after delay from previous response
			if retryAfterDelay > 0 {
				delay = max(delay, retryAfterDelay)
				retryAfterDelay = 0 // Reset for next iteration
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Clone the request for each attempt to avoid issues with consumed request bodies
		req := request.Clone(ctx)
		if request.Body != nil && request.GetBody != nil {
			req.Body, err = request.GetBody()
			if err != nil {
				return nil, err
			}
		}

		resp, err = c.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return resp, err
			}
			// Some proxies close the connection without returning a response, which often surfaces as EOF.
			// Treat that as retryable.
			if errors.Is(err, io.EOF) {
				continue
			}
			var netErr net.Error
			if errors.As(err, &netErr) {
				continue
			}
			// Not a network error, so don't retry
			return resp, err
		}

		// Retry on 429, 408, and server errors.
		//
		// Some proxies/edges return non-standard status codes (e.g. 999) for upstream/internal failures.
		// Treat those as retryable too.
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusRequestTimeout ||
			resp.StatusCode == 999 ||
			(resp.StatusCode >= 500 && resp.StatusCode < 600) {
			// Read Retry-After header before closing body
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
						retryAfterDelay = time.Duration(seconds) * time.Second
					}
				}
			}
			if i+1 < opts.Attempts {
				resp.Body.Close()
				continue
			}
			return resp, nil
		}

		// Success or a non-5xx error code.
		return resp, nil
	}

	// No more attempts, return last resp and error.
	return resp, err
}
