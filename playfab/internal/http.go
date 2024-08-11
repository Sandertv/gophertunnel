package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sandertv/gophertunnel/playfab/title"
	"io"
	"net/http"
)

func Post[T any](t title.Title, route string, r any, hooks ...func(req *http.Request)) (zero T, err error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(r); err != nil {
		return zero, fmt.Errorf("encode: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, t.URL(route), buf)
	if err != nil {
		return zero, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, hook := range hooks {
		hook(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("POST %s: %w", route, err)
	}
	switch {
	case StatusRange(resp.StatusCode, http.StatusOK):
		var body Result[T]
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return zero, fmt.Errorf("decode: %w", err)
		}
		return body.Data, nil
	default:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return zero, fmt.Errorf("POST %s: %s", route, resp.Status)
		}
		var body Error
		if err := json.Unmarshal(b, &body); err != nil {
			return zero, fmt.Errorf("POST %s: %s: %s (%w)", route, resp.Status, b, err)
		}
		return zero, body
	}
}

func StatusRange(code, region int) bool {
	if region%100 != 0 {
		panic(fmt.Sprintf("playfab/internal: StatusRange: invalid http status region: %d", region))
	}
	return code >= region && code < region+100
}
