package mpsd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
)

func (s *Session) Invite(xuid string, titleID int) (*InviteHandle, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&inviteHandle{
		Type:             "invite",
		SessionReference: s.ref,
		Version:          1,
		InvitedXUID:      xuid,
		InviteAttributes: map[string]any{
			"titleId": strconv.Itoa(titleID),
		},
	}); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, handlesURL.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xbl-Contract-Version", strconv.Itoa(contractVersion))

	resp, err := s.conf.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusCreated:
		// It seems the C++ implementation only decodes "id" field from the response.
		var handle *InviteHandle
		if err := json.NewDecoder(resp.Body).Decode(&handle); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}
		return handle, nil
	default:
		return nil, fmt.Errorf("%s %s: %s", req.Method, req.URL, resp.Status)
	}
}

type inviteHandle struct {
	Type             string           `json:"type,omitempty"`    // Always "invite".
	Version          int              `json:"version,omitempty"` // Always 1.
	InviteAttributes map[string]any   `json:"inviteAttributes,omitempty"`
	InvitedXUID      string           `json:"invitedXuid,omitempty"`
	SessionReference SessionReference `json:"sessionRef,omitempty"`
}

type InviteHandle struct {
	inviteHandle
	Expiration     time.Time       `json:"expiration,omitempty"`
	ID             uuid.UUID       `json:"id,omitempty"`
	InviteProtocol string          `json:"inviteProtocol,omitempty"`
	SenderXUID     string          `json:"senderXuid,omitempty"`
	GameTypes      json.RawMessage `json:"gameTypes,omitempty"`
}
