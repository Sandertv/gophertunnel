package service

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/sandertv/gophertunnel/minecraft/service/internal"
)

// refreshingKeySet implements an OIDC KeySet backed by a cached JWKS fetch with a refresh interval.
// This is used for verifying OpenID ID tokens issued by the authorization service.
type refreshingKeySet struct {
	env *AuthorizationEnvironment

	jwksURL         string
	refreshInterval time.Duration
	sigAlgs         []jose.SignatureAlgorithm

	// ctx is used for configuration. Cancellation is ignored.
	ctx context.Context

	mu sync.RWMutex

	// inflight suppresses parallel key refresh and allows multiple goroutines to wait.
	inflight *refreshingKeySetInflight

	cachedKeys []jose.JSONWebKey
	lastFetch  time.Time
}

// refreshingKeySetInflight represents a single in-progress JWKS refresh shared by waiters.
type refreshingKeySetInflight struct {
	doneCh chan struct{}
	keys   []jose.JSONWebKey
	err    error
}

// newRefreshingKeySet constructs a refreshingKeySet for verifying JWT signatures against JWKS.
func newRefreshingKeySet(ctx context.Context, env *AuthorizationEnvironment, jwksURL string, refreshInterval time.Duration, supportedSigningAlgs []string) *refreshingKeySet {
	sigAlgs := make([]jose.SignatureAlgorithm, 0, len(supportedSigningAlgs))
	for _, alg := range supportedSigningAlgs {
		if alg == "" {
			continue
		}
		sigAlgs = append(sigAlgs, jose.SignatureAlgorithm(alg))
	}
	if len(sigAlgs) == 0 {
		// If discovery didn't specify algorithms, default to RS256, which is mandatory for OIDC.
		sigAlgs = []jose.SignatureAlgorithm{jose.RS256}
	}

	return &refreshingKeySet{
		env:             env,
		jwksURL:         jwksURL,
		refreshInterval: refreshInterval,
		sigAlgs:         sigAlgs,
		ctx:             context.WithoutCancel(ctx),
	}
}

// VerifySignature verifies a JWT signature using cached keys, optionally refreshing on mismatch.
func (r *refreshingKeySet) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	jws, err := jose.ParseSigned(jwt, r.sigAlgs)
	if err != nil {
		return nil, fmt.Errorf("minecraft/service: malformed jwt: %w", err)
	}

	// OIDC ID tokens are expected to have a single signature, use the first one to read kid.
	keyID := ""
	for _, sig := range jws.Signatures {
		keyID = sig.Header.KeyID
		break
	}

	keys, lastFetch := r.keysFromCache()
	if payload, ok := r.tryVerify(jws, keys, keyID); ok {
		return payload, nil
	}

	// Refresh keys only when one of the following conditions is met:
	// - We haven't fetched keys yet
	// - We don't recognise the kid and it's been long enough since last fetch.
	needsRefresh := len(keys) == 0
	if !needsRefresh && keyID != "" && !r.knowsKeyID(keys, keyID) && time.Since(lastFetch) >= r.refreshInterval {
		needsRefresh = true
	}
	if needsRefresh {
		keys, err := r.keysFromRemote(ctx)
		if err != nil {
			return nil, fmt.Errorf("minecraft/service: fetch jwks: %w", err)
		}
		if payload, ok := r.tryVerify(jws, keys, keyID); ok {
			return payload, nil
		}
	}

	return nil, errors.New("minecraft/service: failed to verify id token signature")
}

// keysFromCache returns the currently cached JWKS keys and the last refresh time.
func (r *refreshingKeySet) keysFromCache() (keys []jose.JSONWebKey, lastFetch time.Time) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cachedKeys, r.lastFetch
}

// keysFromRemote refreshes keys from the JWKS endpoint, and de-duplicates concurrent refreshes.
func (r *refreshingKeySet) keysFromRemote(ctx context.Context) ([]jose.JSONWebKey, error) {
	// Lock to inspect the inflight request field.
	r.mu.Lock()
	if r.inflight == nil {
		r.inflight = &refreshingKeySetInflight{doneCh: make(chan struct{})}
		inflight := r.inflight

		// This goroutine has exclusive ownership over the current inflight request.
		go func() {
			keys, err := r.updateKeys()

			inflight.keys = keys
			inflight.err = err
			close(inflight.doneCh)

			r.mu.Lock()
			defer r.mu.Unlock()

			if err == nil {
				r.cachedKeys = keys
				r.lastFetch = time.Now()
			}
			r.inflight = nil
		}()
	}
	inflight := r.inflight
	r.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-inflight.doneCh:
		return inflight.keys, inflight.err
	}
}

// updateKeys performs the HTTP request to fetch and decode the JWKS.
func (r *refreshingKeySet) updateKeys() ([]jose.JSONWebKey, error) {
	req, err := http.NewRequestWithContext(r.ctx, http.MethodGet, r.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", internal.UserAgent)

	resp, err := r.env.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, internal.Err(resp)
	}

	keyset, err := decodeKeySet(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode key set: %w", err)
	}
	return keyset.Keys, nil
}

// knowsKeyID returns whether the provided JWKS key list contains the given kid.
func (r *refreshingKeySet) knowsKeyID(keys []jose.JSONWebKey, keyID string) bool {
	for _, key := range keys {
		if key.KeyID == keyID {
			return true
		}
	}
	return false
}

// tryVerify attempts verification against keys, filtering by kid when present.
func (r *refreshingKeySet) tryVerify(jws *jose.JSONWebSignature, keys []jose.JSONWebKey, keyID string) ([]byte, bool) {
	for _, key := range keys {
		if keyID == "" || key.KeyID == keyID {
			if payload, err := jws.Verify(&key); err == nil {
				return payload, true
			}
		}
	}
	return nil, false
}

// decodeKeySet decodes the key set obtained from the authorization service
// with minor patches to support 'x5t' field being hex-encoded instead of
// [base64.RawURLEncoding].
func decodeKeySet(r io.Reader) (*jose.JSONWebKeySet, error) {
	var data struct {
		Keys []map[string]any `json:"keys"`
	}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}
	set := &jose.JSONWebKeySet{
		Keys: make([]jose.JSONWebKey, len(data.Keys)),
	}
	for i, key := range data.Keys {
		x5t, ok := key["x5t"].(string)
		if !ok {
			return nil, errors.New("no x5t found in jwk")
		}

		// Microsoft uses hex instead of base64 for the 'x5t' field, which violates the JOSE spec.
		// For a SHA-1 thumbprint (20 bytes), we expect either:
		// - Hex: 40 chars
		// - Base64URL (no padding): 27 chars
		var fingerprint []byte
		switch len(x5t) {
		case 40:
			var err error
			fingerprint, err = hex.DecodeString(x5t)
			if err != nil {
				return nil, fmt.Errorf("decode x5t hex: %w", err)
			}
			key["x5t"] = base64.RawURLEncoding.EncodeToString(fingerprint)
		case 27:
			var err error
			fingerprint, err = base64.RawURLEncoding.DecodeString(x5t)
			if err != nil {
				return nil, fmt.Errorf("decode x5t base64: %w", err)
			}
		default:
			return nil, fmt.Errorf("invalid x5t length: %d", len(x5t))
		}
		if n := len(fingerprint); n != 20 {
			return nil, fmt.Errorf("fingerprint is not 20 bytes long: %d", n)
		}

		// jose.JSONWebKey validates during JSON unmarshalling, so after patching x5t we re-encode the
		// JWK and let the custom UnmarshalJSON validate and decode it.
		b, err := json.Marshal(key)
		if err != nil {
			return nil, fmt.Errorf("encode reformatted jwk: %w", err)
		}
		var jwk jose.JSONWebKey
		if err := jwk.UnmarshalJSON(b); err != nil {
			return nil, fmt.Errorf("decode reformatted jwk: %w", err)
		}
		set.Keys[i] = jwk
	}
	return set, nil
}
