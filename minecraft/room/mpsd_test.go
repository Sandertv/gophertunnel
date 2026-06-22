package room

import (
	"testing"

	"github.com/df-mc/go-xsapi/v2/mpsd"
	"github.com/sandertv/gophertunnel/minecraft/p2p"
)

func TestXBLAnnouncerPublishConfigUsesBroadcastRestrictions(t *testing.T) {
	t.Parallel()

	_, read, join := (&XBLAnnouncer{}).publishConfig(Status{BroadcastSetting: p2p.BroadcastSettingInviteOnly}, nil)
	if read != mpsd.SessionRestrictionFollowed {
		t.Fatalf("read restriction mismatch: got %q", read)
	}
	if join != mpsd.SessionRestrictionLocal {
		t.Fatalf("join restriction mismatch: got %q", join)
	}
}
