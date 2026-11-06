package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/internal"
)

// Sign requests the Chat service to sign the chat message.
//
// The resulting SignResult can be sent to a server by using the
// [SignResult.Protocol] method.
// The server can then distribute the signed message to other players
// in the same session. Clients receiving the chat payload can verify
// the signature using the [Environment.Verify] method.
func (e *Environment) Sign(ctx context.Context, src service.TokenSource, request SignRequest) (*SignResult, error) {
	if request.RoomID == "" {
		return nil, errors.New("service/chat: SignRequest.RoomID cannot be empty")
	}
	if request.MessageID == "" {
		return nil, errors.New("service/chat: SignRequest.MessageID cannot be empty")
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return nil, fmt.Errorf("service/chat: encode request body: %w", err)
	}

	token, err := src.ServiceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("service/chat: request service token: %w", err)
	}
	requestURL := e.ServiceURI.JoinPath("/api/v1.0/chat/sign").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, buf)
	if err != nil {
		return nil, fmt.Errorf("service/chat: make request: %w", err)
	}
	req.Header.Set("User-Agent", internal.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	token.SetAuthHeader(req)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result internal.Result[*SignResult]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("service/chat: decode response body: %w", err)
		}
		if result.Data == nil || !result.Data.valid() {
			return nil, errors.New("service/chat: invalid sign chat response")
		}
		return result.Data, nil
	default:
		return nil, internal.Err(resp)
	}
}

// SignRequest represents a request to sign a chat message.
type SignRequest struct {
	// RoomID identifies the world or server the message is intended for.
	// It is used by the clients to verify whether the chat message sent by
	// the server was really sent in the same session.
	// It is typically derived from [packet.StartGame.LevelID].
	RoomID string `json:"roomId"`
	// XUID identifies the sender of the message. It can be left empty, in that
	// case, the Chat service automatically fills the XUID in the result.
	XUID string `json:"xuid,omitempty"`
	// MessageID is a client-generated UUID for the message.
	MessageID string `json:"messageId"`
	// Message is the original chat message to be signed. It may be redacted
	// by the Chat service.
	Message string `json:"message"`
}

// SignResult represents the result returned by the Chat service after signing a chat message.
type SignResult struct {
	// RoomID identifies the world or server the signature was minted for.
	RoomID string `json:"roomId"`
	// XUID identifies the sender of the message.
	XUID string `json:"xuid"`
	// MessageID is the client-generated UUID specified in [SignRequest.MessageID].
	MessageID string `json:"messageId"`
	// ScanText is the message text scanned by the Chat service.
	ScanText string `json:"scanText"`
	// RedactedText is the redacted form of the message. It is always non-empty
	// string even if nothing needed redacting.
	RedactedText string `json:"redactedText"`
	// Signature is the signature produced by the Chat service.
	Signature string `json:"signature"`
	// KeyID identifies the key used to produce the signature.
	KeyID string `json:"kid"`
	// Filtered indicates whether the message was filtered during the safety scan.
	// It is not included in the signature.
	Filtered bool `json:"filtered"`
	// Timestamp is the timestamp associated with the result.
	// It is currently 0 for unknown reason and is not included in the signature.
	Timestamp int `json:"timestamp"`
	// RoutingResult is unknown. It is typically 'LocalBroadcast'.
	RoutingResult string `json:"routingResult"`
}

// valid reports whether all fields required to convert the result
// to a [protocol.SignedChat] are non-empty.
func (r SignResult) valid() bool {
	return !slices.ContainsFunc([]string{
		r.KeyID,
		r.XUID,
		r.RoomID,
		r.MessageID,
		r.ScanText,
		r.RedactedText,
		r.Signature,
	}, func(s string) bool {
		return s == ""
	})
}

// Protocol converts the result to a [protocol.SignedChat] that can be sent over a connection.
func (r SignResult) Protocol() protocol.SignedChat {
	return protocol.SignedChat{
		KeyID:        r.KeyID,
		SenderID:     r.XUID,
		RoomID:       r.RoomID,
		MessageID:    r.MessageID,
		ScanText:     r.ScanText,
		RedactedText: r.RedactedText,
		Signature:    r.Signature,
	}
}
