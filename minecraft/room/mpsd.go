package room

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/df-mc/go-xsapi"
	"github.com/df-mc/go-xsapi/mpsd"
	"github.com/google/uuid"
	"strings"
	"sync"
)

var serviceConfigID = uuid.MustParse("4fc10100-5f7a-4470-899b-280835760c07")

func NewSessionAnnouncer(s *mpsd.Session) *SessionAnnouncer {
	return &SessionAnnouncer{
		s: s,
	}
}

type SessionPublishConfig struct {
	PublishConfig mpsd.PublishConfig
	Reference     mpsd.SessionReference
}

func (conf SessionPublishConfig) New(src xsapi.TokenSource) *SessionAnnouncer {
	return &SessionAnnouncer{
		p:   conf,
		src: src,
	}
}

func (conf SessionPublishConfig) publish(ctx context.Context, src xsapi.TokenSource) (*mpsd.Session, error) {
	if conf.Reference.ServiceConfigID == uuid.Nil {
		conf.Reference.ServiceConfigID = serviceConfigID
	}
	if conf.Reference.TemplateName == "" {
		conf.Reference.TemplateName = "MinecraftLobby"
	}
	if conf.Reference.Name == "" {
		conf.Reference.Name = strings.ToUpper(uuid.NewString())
	}

	s, err := conf.PublishConfig.PublishContext(ctx, src, conf.Reference)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type SessionAnnouncer struct {
	p SessionPublishConfig

	src xsapi.TokenSource

	s           *mpsd.Session
	description *mpsd.SessionDescription
	mu          sync.Mutex
}

func (a *SessionAnnouncer) Announce(ctx context.Context, status Status) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	custom, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("encode status: %w", err)
	}
	a.updateDescription(status)
	if bytes.Compare(a.description.Properties.Custom, custom) == 0 {
		return nil // Avoid committing same properties
	}
	a.description.Properties.Custom = custom

	if a.s == nil {
		a.p.PublishConfig.Description = a.description
		s, err := a.p.publish(ctx, a.src)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		a.s = s
		return nil
	}

	commit, err := a.s.Commit(ctx, a.description)
	if err == nil {
		a.description = commit.SessionDescription
	}
	return err
}

func (a *SessionAnnouncer) Close() error {
	return a.s.Close()
}

func (a *SessionAnnouncer) updateDescription(status Status) {
	if a.description == nil {
		a.description = &mpsd.SessionDescription{}
	}
	if a.description.Properties == nil {
		a.description.Properties = &mpsd.SessionProperties{}
	}
	if a.description.Properties.System == nil {
		a.description.Properties.System = &mpsd.SessionPropertiesSystem{}
	}

	switch status.BroadcastSetting {
	case BroadcastSettingFriendsOfFriends, BroadcastSettingFriendsOnly:
		a.description.Properties.System.JoinRestriction = mpsd.SessionRestrictionFollowed
		a.description.Properties.System.ReadRestriction = mpsd.SessionRestrictionFollowed
	case BroadcastSettingInviteOnly:
		a.description.Properties.System.JoinRestriction = mpsd.SessionRestrictionLocal
		a.description.Properties.System.ReadRestriction = mpsd.SessionRestrictionLocal
	}
}
