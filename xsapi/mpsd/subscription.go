package mpsd

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/xsapi/internal"
	"strings"
)

const resourceURI = "https://sessiondirectory.xboxlive.com/connections/"

type subscription struct {
	ConnectionID uuid.UUID `json:"ConnectionId,omitempty"`
}

type subscriptionHandler struct {
	*Session
}

func (h *subscriptionHandler) HandleEvent(data json.RawMessage) {
	var event subscriptionEvent
	if err := json.Unmarshal(data, &event); err != nil {
		h.conf.Logger.Error("error decoding subscription event", internal.ErrAttr(err))
	}
	for _, tap := range event.ShoulderTaps {
		ref, err := h.parseReference(tap.Resource)
		if err != nil {
			h.conf.Logger.Error("handle subscription event: error parsing shoulder tap", internal.ErrAttr(err))
			continue
		}
		h.handler().HandleSessionChange(ref, tap.Branch, tap.ChangeNumber)
	}
}

func (h *subscriptionHandler) parseReference(s string) (ref SessionReference, err error) {
	segments := strings.Split(s, "~")
	if len(segments) != 3 {
		return ref, fmt.Errorf("unexpected segmentations: %s", s)
	}
	ref.ServiceConfigID, err = uuid.Parse(segments[0])
	if err != nil {
		return ref, fmt.Errorf("parse service config ID: %w", err)
	}
	ref.TemplateName = segments[1]
	ref.Name = segments[2]
	return ref, nil
}

type subscriptionEvent struct {
	ShoulderTaps []shoulderTap `json:"shoulderTaps"`
}

type shoulderTap struct {
	Resource     string    `json:"resource"`
	ChangeNumber uint64    `json:"changeNumber"`
	Branch       uuid.UUID `json:"branch"`
}
