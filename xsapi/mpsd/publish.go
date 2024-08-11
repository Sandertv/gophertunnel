package mpsd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/xsapi"
	"github.com/sandertv/gophertunnel/xsapi/rta"
	"log/slog"
	"net/http"
	"strings"
)

type PublishConfig struct {
	RTADialer *rta.Dialer
	RTAConn   *rta.Conn

	Description *SessionDescription

	Client *http.Client
	Logger *slog.Logger
}

func (conf PublishConfig) PublishContext(ctx context.Context, src xsapi.TokenSource, ref SessionReference) (s *Session, err error) {
	if conf.Logger == nil {
		conf.Logger = slog.Default()
	}
	if conf.Client == nil {
		conf.Client = &http.Client{}
	}
	var hasTransport bool
	if conf.Client.Transport != nil {
		_, hasTransport = conf.Client.Transport.(*xsapi.Transport)
	}
	if !hasTransport {
		conf.Client.Transport = &xsapi.Transport{
			Source: src,
			Base:   conf.Client.Transport,
		}
	}

	if conf.RTAConn == nil {
		if conf.RTADialer == nil {
			conf.RTADialer = &rta.Dialer{}
		}
		conf.RTAConn, err = conf.RTADialer.DialContext(ctx, src)
		if err != nil {
			return nil, fmt.Errorf("dial rta: %w", err)
		}
	}

	tok, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("obtain token: %w", err)
	}

	sub, err := conf.RTAConn.Subscribe(ctx, resourceURI)
	if err != nil {
		return nil, fmt.Errorf("subscribe with rta: %w", err)
	}
	var custom subscription
	if err := json.Unmarshal(sub.Custom, &custom); err != nil {
		return nil, fmt.Errorf("decode subscription custom: %w", err)
	}

	if conf.Description == nil {
		conf.Description = &SessionDescription{}
	}
	if conf.Description.Members == nil {
		conf.Description.Members = make(map[string]*MemberDescription, 1)
	}

	me, ok := conf.Description.Members["me"]
	if !ok {
		me = &MemberDescription{}
	}
	if me.Constants == nil {
		me.Constants = &MemberConstants{}
	}
	if me.Constants.System == nil {
		me.Constants.System = &MemberConstantsSystem{}
	}
	me.Constants.System.Initialize = true
	if claimer, ok := tok.(xsapi.DisplayClaimer); ok {
		me.Constants.System.XUID = claimer.DisplayClaims().XUID
	}
	if me.Properties == nil {
		me.Properties = &MemberProperties{}
	}
	if me.Properties.System == nil {
		me.Properties.System = &MemberPropertiesSystem{}
	}
	me.Properties.System.Active = true
	me.Properties.System.Connection = custom.ConnectionID
	if me.Properties.System.Subscription == nil {
		me.Properties.System.Subscription = &MemberPropertiesSystemSubscription{}
	}
	if me.Properties.System.Subscription.ID == "" {
		me.Properties.System.Subscription.ID = strings.ToUpper(uuid.NewString())
	}
	me.Properties.System.Subscription.ChangeTypes = []string{
		ChangeTypeEverything,
	}
	conf.Description.Members["me"] = me

	if _, err := conf.commit(ctx, ref, conf.Description); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	if err := conf.commitActivity(ctx, ref); err != nil {
		return nil, fmt.Errorf("commit activity: %w", err)
	}

	return &Session{
		ref:  ref,
		conf: conf,
		rta:  conf.RTAConn,
		log:  conf.Logger,
		sub:  sub,
	}, nil
}

const resourceURI = "https://sessiondirectory.xboxlive.com/connections/"

type subscription struct {
	ConnectionID uuid.UUID `json:"ConnectionId,omitempty"`
}
