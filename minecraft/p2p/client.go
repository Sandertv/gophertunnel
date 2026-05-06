package p2p

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/df-mc/go-xsapi/v2"
	"github.com/sandertv/gophertunnel/minecraft/auth"
)

// ClientConfig encapsulates configuration for creating a Client.
type ClientConfig struct {
	// ErrorLog is used to log errors encountered while decoding the custom
	// properties of a multiplayer session activity. If nil, [slog.Default]
	// will be used.
	ErrorLog *slog.Logger
}

// New returns a new Client using the underlying [xsapi.Client].
func (conf ClientConfig) New(client *xsapi.Client) *Client {
	if conf.ErrorLog == nil {
		conf.ErrorLog = slog.Default().With("src", "minecraft/p2p")
	}
	return &Client{
		client: client,
		conf:   conf,
	}
}

// NewClient returns a new Client using the underlying [xsapi.Client].
func NewClient(client *xsapi.Client) *Client {
	var c ClientConfig
	return c.New(client)
}

// Client implements an API client for searching peer-to-peer worlds hosted by players.
type Client struct {
	client *xsapi.Client
	conf   ClientConfig
}

// Worlds returns a list of worlds available to join.
func (c *Client) Worlds(ctx context.Context) ([]World, error) {
	activities, err := c.client.MPSD().Activities(ctx, auth.ServiceConfigID)
	if err != nil {
		return nil, err
	}
	worlds := make([]World, 0, len(activities))
	for _, activity := range activities {
		var world World
		if err := json.Unmarshal(activity.CustomProperties, &world); err != nil {
			c.conf.ErrorLog.Error("error decoding world data",
				slog.Any("error", err),
				slog.String("customProperties", string(activity.CustomProperties)),
			)
			continue
		}
		world.client, world.handleID = c, activity.ID
		worlds = append(worlds, world)
	}
	return worlds, nil
}
