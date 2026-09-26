package realms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	compatiblePath             = "/mco/client/compatible"
	compatibleResponse         = "COMPATIBLE"
	unknownClientVersionReason = "unknown_client_version"
	maxVersionFallback         = 24
	versionSearchCooldown      = 5 * time.Minute
)

// SetClientVersion sets the game version sent to Realms. An empty version uses
// protocol.CurrentVersion. Invalid versions are ignored.
func (c *Client) SetClientVersion(version string) {
	version = strings.TrimSpace(version)
	c.versionMu.Lock()
	defer c.versionMu.Unlock()
	if version != "" {
		if _, _, _, ok := parseVersion(version); !ok {
			return
		}
	}
	if version == c.preferredVersion {
		return
	}
	c.preferredVersion, c.acceptedVersion = version, ""
	c.searchFailedAt = time.Time{}
}

// clientVersion returns the version to send as Client-Version.
func (c *Client) clientVersion() string {
	c.versionMu.Lock()
	defer c.versionMu.Unlock()
	if c.acceptedVersion != "" {
		return c.acceptedVersion
	}
	if c.preferredVersion != "" {
		return c.preferredVersion
	}
	return protocol.CurrentVersion
}

// negotiateClientVersion finds and caches an accepted version after a rejection.
func (c *Client) negotiateClientVersion(ctx context.Context, failed string) (string, bool) {
	c.negotiateMu.Lock()
	defer c.negotiateMu.Unlock()
	// Another caller may have negotiated a version while this request was in flight.
	if current := c.clientVersion(); current != failed {
		return current, true
	}

	c.versionMu.Lock()
	preferred, failedAt := c.preferredVersion, c.searchFailedAt
	c.versionMu.Unlock()
	if preferred == "" {
		preferred = protocol.CurrentVersion
	}
	if !failedAt.IsZero() && time.Since(failedAt) < versionSearchCooldown {
		return failed, false
	}

	for _, candidate := range versionCandidates(preferred) {
		if candidate == failed {
			continue
		}
		accepted, err := c.versionCompatible(ctx, candidate)
		if err != nil {
			// Stop searching if the compatibility endpoint fails.
			c.recordSearchFailure()
			return failed, false
		}
		if !accepted {
			continue
		}
		c.versionMu.Lock()
		c.acceptedVersion, c.searchFailedAt = candidate, time.Time{}
		c.versionMu.Unlock()
		return candidate, true
	}
	c.recordSearchFailure()
	return failed, false
}

// recordSearchFailure holds off further searches until the cooldown passes.
func (c *Client) recordSearchFailure() {
	c.versionMu.Lock()
	defer c.versionMu.Unlock()
	c.searchFailedAt = time.Now()
}

// versionCompatible reports whether Realms still accepts a game version.
func (c *Client) versionCompatible(ctx context.Context, version string) (bool, error) {
	body, status, err := c.sendWithMethod(ctx, http.MethodGet, realmsBaseURL, compatiblePath, nil, version)
	if err != nil {
		return false, err
	}
	if status != http.StatusOK {
		return false, fmt.Errorf("realms compatibility check: unexpected status %d", status)
	}
	return strings.EqualFold(strings.TrimSpace(string(body)), compatibleResponse), nil
}

// versionCandidates lists versions to try, starting at version and walking its
// patch component down towards zero.
func versionCandidates(version string) []string {
	major, minor, patch, ok := parseVersion(version)
	if !ok {
		return []string{version}
	}
	candidates := make([]string, 0, maxVersionFallback+1)
	for i := 0; i <= maxVersionFallback && patch-i >= 0; i++ {
		candidates = append(candidates, fmt.Sprintf("%d.%d.%d", major, minor, patch-i))
	}
	return candidates
}

// parseVersion splits the leading MAJOR.MINOR.PATCH of a version, ignoring any
// build revision after it.
func parseVersion(version string) (major, minor, patch int, ok bool) {
	fields := strings.Split(strings.TrimSpace(version), ".")
	if len(fields) < 3 {
		return 0, 0, 0, false
	}
	var parsed [3]int
	for i := range parsed {
		n, err := strconv.Atoi(fields[i])
		if err != nil || n < 0 {
			return 0, 0, 0, false
		}
		parsed[i] = n
	}
	return parsed[0], parsed[1], parsed[2], true
}

// unknownClientVersion reports whether a response rejects the Client-Version
// header rather than the request itself.
func unknownClientVersion(status int, body []byte) bool {
	if status != http.StatusBadRequest {
		return false
	}
	var payload struct {
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	return payload.Reason == unknownClientVersionReason
}
