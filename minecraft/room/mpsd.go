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

// XBLAnnouncer announces a Status through the Multiplayer Session Directory (MPSD) of Xbox Live.
type XBLAnnouncer struct {
	// TokenSource provides the [xsapi.Token] required to publish a session when the Session is nil.
	TokenSource xsapi.TokenSource

	// SessionReference specifies the internal ID of the session being published when the Session is nil.
	SessionReference mpsd.SessionReference

	// PublishConfig specifies custom configuration for publishing a session when the Session is nil.
	PublishConfig mpsd.PublishConfig

	// Session is the session where the Status will be committed. If nil, a [mpsd.Session] will be published
	// using the PublishConfig.
	Session *mpsd.Session

	// custom properties are encoded from Status for comparison in announcements.
	custom []byte

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
	if bytes.Compare(custom, a.custom) == 0 {
		return nil
	} else {
		a.custom = custom
	}

	if a.Session == nil {
		if a.PublishConfig.Description == nil {
			a.PublishConfig.Description = a.description(status)
		}
		a.PublishConfig.Description.Properties.Custom = custom

		if a.SessionReference.ServiceConfigID == uuid.Nil {
			a.SessionReference.ServiceConfigID = uuid.MustParse("4fc10100-5f7a-4470-899b-280835760c07")
		}
		if a.SessionReference.TemplateName == "" {
			a.SessionReference.TemplateName = "MinecraftLobby"
		}
		if a.SessionReference.Name == "" {
			a.SessionReference.Name = strings.ToUpper(uuid.NewString())
		}

		a.Session, err = a.PublishConfig.PublishContext(ctx, a.TokenSource, a.SessionReference)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		return nil
	}
	_, err = a.Session.Commit(ctx, a.description(status))
	return err
}

// description returns a [mpsd.SessionDescription] to be committed or published on the Session.
// It uses custom properties encoded from the Status in [XBLAnnouncer.Announce].
func (a *XBLAnnouncer) description(status Status) *mpsd.SessionDescription {
	read, join := a.restrictions(status.BroadcastSetting)
	return &mpsd.SessionDescription{
		Properties: &mpsd.SessionProperties{
			System: &mpsd.SessionPropertiesSystem{
				ReadRestriction: read,
				JoinRestriction: join,
			},
			Custom: a.custom,
		},
	}
}

// restrictions determines the read and join restrictions for the session based on [Status.BroadcastSetting].
func (a *XBLAnnouncer) restrictions(setting int32) (read, join string) {
	switch setting {
	case BroadcastSettingFriendsOfFriends, BroadcastSettingFriendsOnly:
		return mpsd.SessionRestrictionFollowed, mpsd.SessionRestrictionFollowed
	case BroadcastSettingInviteOnly:
		return mpsd.SessionRestrictionLocal, mpsd.SessionRestrictionFollowed
	default:
		return mpsd.SessionRestrictionFollowed, mpsd.SessionRestrictionFollowed
	}
}

func (a *XBLAnnouncer) Close() (err error) {
	a.Lock()
	defer a.Unlock()

	if a.Session != nil {
		return a.Session.Close()
	}
	return nil
}
