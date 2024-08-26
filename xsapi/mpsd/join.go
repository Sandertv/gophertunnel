package mpsd

import (
	"context"
	"github.com/sandertv/gophertunnel/xsapi"
)

type JoinConfig struct {
	PublishConfig
}

func (conf JoinConfig) JoinContext(ctx context.Context, src xsapi.TokenSource, handle ActivityHandle) (*Session, error) {
	return conf.publish(ctx, src, handle.URL().JoinPath("session"), handle.SessionReference)
}
