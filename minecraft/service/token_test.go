package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/google/uuid"
)

func TestDecodeClaimsAllowsSmallIssuedAtSkew(t *testing.T) {
	t.Parallel()

	token := &Token{AuthorizationHeader: testAuthorizationHeader(t, time.Now().Add(2*time.Minute))}
	if err := decodeClaims(token); err != nil {
		t.Fatalf("decodeClaims: %v", err)
	}
	if token.Claims.PlayerMessagingID == uuid.Nil {
		t.Fatal("PlayerMessagingID was not decoded")
	}
}

func TestDecodeClaimsRejectsLargeIssuedAtSkew(t *testing.T) {
	t.Parallel()

	token := &Token{AuthorizationHeader: testAuthorizationHeader(t, time.Now().Add(10*time.Minute))}
	err := decodeClaims(token)
	if !errors.Is(err, jwt.ErrIssuedInTheFuture) {
		t.Fatalf("decodeClaims error = %v, want ErrIssuedInTheFuture", err)
	}
}

func testAuthorizationHeader(t *testing.T, issuedAt time.Time) string {
	t.Helper()

	payload, err := json.Marshal(struct {
		PlayerMessagingID uuid.UUID `json:"pmid"`
		IssuedAt          int64     `json:"iat"`
		Expiry            int64     `json:"exp"`
	}{
		PlayerMessagingID: uuid.New(),
		IssuedAt:          issuedAt.Unix(),
		Expiry:            time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return "MCToken header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}
