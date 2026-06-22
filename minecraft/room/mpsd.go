package room

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/df-mc/go-xsapi/v2/mpsd"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/p2p"
)

// XBLAnnouncer announces a Status through the Multiplayer Session Directory (MPSD) of Xbox Live.
type XBLAnnouncer struct {
	// Client publishes and updates MPSD sessions.
	Client *mpsd.Client

	// SessionReference specifies the internal ID of the session being published when the Session is nil.
	SessionReference mpsd.SessionReference

	// PublishConfig specifies custom configuration for publishing a session when the Session is nil.
	PublishConfig mpsd.PublishConfig

	// Session is the session where the Status will be committed. If nil, a [mpsd.Session] will be published
	// using the PublishConfig.
	Session *mpsd.Session

	// custom properties are encoded from Status for comparison in announcements.
	custom []byte
	// readRestriction and joinRestriction track the effective MPSD restrictions used by Session.
	readRestriction string
	joinRestriction string

	// Mutex ensures atomic read/write access to the fields.
	sync.Mutex
}

// Announce commits or publishes a [mpsd.Session] with the given Status. The status will be encoded as custom properties
// of the session description. The [context.Context] may be used to control the deadline or cancellation of announcement.
//
// If the Status has not changed since the last announcement, the method will return immediately.
func (a *XBLAnnouncer) Announce(ctx context.Context, status Status) error {
	a.Lock()
	defer a.Unlock()

	custom, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	config, read, join := a.publishConfig(status, custom)
	if a.Session != nil && a.readRestriction == "" && a.joinRestriction == "" {
		if properties := a.Session.Properties(); properties.System != nil {
			a.readRestriction = properties.System.ReadRestriction
			a.joinRestriction = properties.System.JoinRestriction
		}
	}
	if bytes.Equal(custom, a.custom) && read == a.readRestriction && join == a.joinRestriction {
		return nil
	}

	restrictionsChanged := read != a.readRestriction || join != a.joinRestriction
	if a.Session != nil && restrictionsChanged {
		if a.Client == nil {
			return errors.New("room: XBLAnnouncer.Client is nil and MPSD restrictions changed")
		}
		if err := a.Session.CloseContext(ctx); err != nil {
			return fmt.Errorf("close stale session: %w", err)
		}
		a.Session = nil
	}

	if a.Session == nil {
		if a.Client == nil {
			return errors.New("room: XBLAnnouncer.Client is nil")
		}
		if a.SessionReference.ServiceConfigID == uuid.Nil {
			a.SessionReference.ServiceConfigID = auth.ServiceConfigID
		}
		if a.SessionReference.TemplateName == "" {
			a.SessionReference.TemplateName = "MinecraftLobby"
		}
		if a.SessionReference.Name == "" {
			a.SessionReference.Name = strings.ToUpper(uuid.NewString())
		}

		a.Session, err = a.Client.Publish(ctx, a.SessionReference, config)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		a.custom = custom
		a.readRestriction = read
		a.joinRestriction = join
		return nil
	}
	if err := a.Session.SetCustomProperties(ctx, custom); err != nil {
		return fmt.Errorf("set custom properties: %w", err)
	}
	a.custom = custom
	a.readRestriction = read
	a.joinRestriction = join
	return nil
}

// publishConfig returns the effective [mpsd.PublishConfig] for status.
func (a *XBLAnnouncer) publishConfig(status Status, custom []byte) (mpsd.PublishConfig, string, string) {
	setting := status.BroadcastSetting
	if !setting.Valid() {
		setting = p2p.BroadcastSettingFriendsOfFriends
	}
	read, join := setting.ReadRestriction(), setting.JoinRestriction()
	config := a.PublishConfig
	config.CustomProperties = custom
	if config.ReadRestriction == "" {
		config.ReadRestriction = read
	} else {
		read = config.ReadRestriction
	}
	if config.JoinRestriction == "" {
		config.JoinRestriction = join
	} else {
		join = config.JoinRestriction
	}
	return config, read, join
}

func (a *XBLAnnouncer) Close() (err error) {
	a.Lock()
	defer a.Unlock()

	if a.Session != nil {
		return a.Session.Close()
	}
	return nil
}
