package mpsd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (conf PublishConfig) commitActivity(ctx context.Context, ref SessionReference) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(map[string]any{
		"type":       "activity",
		"sessionRef": ref,
		"version":    1,
	}); err != nil {
		return fmt.Errorf("encode request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, handlesURL.String(), buf)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xbl-Contract-Version", strconv.Itoa(contractVersion))

	resp, err := conf.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("%s %s: %s", req.Method, req.URL, resp.Status)
	}
}

var handlesURL = &url.URL{
	Scheme: "https",
	Host:   "sessiondirectory.xboxlive.com",
	Path:   "/handles",
}
