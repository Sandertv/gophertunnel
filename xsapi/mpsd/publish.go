package mpsd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/xsapi"
	"github.com/sandertv/gophertunnel/xsapi/internal"
	"github.com/sandertv/gophertunnel/xsapi/rta"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type PublishConfig struct {
	RTADialer *rta.Dialer
	RTAConn   *rta.Conn

	Description *SessionDescription

	Client *http.Client
	Logger *slog.Logger
}

func (conf PublishConfig) publish(ctx context.Context, src xsapi.TokenSource, u *url.URL, ref SessionReference) (*Session, error) {
	if conf.Logger == nil {
		conf.Logger = slog.Default()
	}
	if conf.Client == nil {
		conf.Client = &http.Client{}
	}
	internal.SetTransport(conf.Client, src)

	if conf.RTAConn == nil {
		if conf.RTADialer == nil {
			conf.RTADialer = &rta.Dialer{}
		}
		var err error
		conf.RTAConn, err = conf.RTADialer.DialContext(ctx, src)
		if err != nil {
			return nil, fmt.Errorf("prepare subscription: dial: %w", err)
		}
	}

	tok, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("obtain token: %w", err)
	}

	sub, err := conf.RTAConn.Subscribe(ctx, resourceURI)
	if err != nil {
		return nil, fmt.Errorf("prepare subscription: subscribe: %w", err)
	}
	var custom subscription
	if err := json.Unmarshal(sub.Custom, &custom); err != nil {
		return nil, fmt.Errorf("prepare subscription: decode: %w", err)
	}

	if conf.Description == nil {
		conf.Description = &SessionDescription{}
	}
	if conf.Description.Members == nil {
		conf.Description.Members = make(map[string]*MemberDescription, 1)
	}

	if ref.Name == "" {
		ref.Name = strings.ToUpper(uuid.NewString())
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

	if _, err := conf.commit(ctx, u, conf.Description); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	if err := conf.commitActivity(ctx, ref); err != nil {
		return nil, fmt.Errorf("commit activity: %w", err)
	}

	s := &Session{
		ref:  ref,
		conf: conf,
		rta:  conf.RTAConn,
		sub:  sub,
	}
	s.Handle(nil)
	sub.Handle(&subscriptionHandler{s})
	return s, nil
}

func (conf PublishConfig) PublishContext(ctx context.Context, src xsapi.TokenSource, ref SessionReference) (s *Session, err error) {
	return conf.publish(ctx, src, ref.URL(), ref)
}
