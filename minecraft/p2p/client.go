package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/df-mc/go-xsapi/v2"
	"github.com/df-mc/go-xsapi/v2/mpsd"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth"
)

// ClientConfig encapsulates configuration for creating a Client.
type ClientConfig struct {
	// Log is used to log errors encountered while decoding the custom
	// properties of a multiplayer session activity. If nil, [slog.Default]
	// will be used.
	Log *slog.Logger
}

// New returns a new Client using the underlying [xsapi.Client].
func (conf ClientConfig) New(client *xsapi.Client) *Client {
	if conf.Log == nil {
		conf.Log = slog.Default().With("src", "minecraft/p2p")
	}
	return &Client{
		client: client,
		conf:   conf,
	}
}

// NewClient returns a new Client using the underlying [xsapi.Client].
func NewClient(client *xsapi.Client) *Client {
	return ClientConfig{}.New(client)
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
		if activity.RelatedInfo == nil || activity.RelatedInfo.Closed {
			continue
		}
		var world World
		if err := json.Unmarshal(activity.CustomProperties, &world); err != nil {
			c.conf.Log.Error("error decoding world data",
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

// Join joins the multiplayer session on Xbox Live using the given handle ID
// and waits until the host publishes a usable connection and nonce for the caller.
func (c *Client) Join(ctx context.Context, handleID uuid.UUID) (_ *Session, err error) {
	s, err := c.client.MPSD().Join(ctx, handleID, mpsd.JoinConfig{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if err2 := s.Close(); err2 != nil {
				err = errors.Join(err, fmt.Errorf("cleanup session: %w", err2))
			}
		}
	}()
	session := &Session{
		client:  c,
		session: s,

		ready: make(chan struct{}),

		log: c.conf.Log.With("sessionRef", s.Reference()),
	}
	s.Handle(&handler{session})
	if err := session.updateWorldData(s.Properties().Custom); err != nil {
		return nil, err
	}
	if err := session.waitReady(ctx); err != nil {
		return nil, err
	}
	return session, nil
}

// Join joins the multiplayer session associated with the World on Xbox Live
// and waits until the host publishes a usable connection and nonce for the caller.
func (w *World) Join(ctx context.Context) (*Session, error) {
	if w.client == nil {
		return nil, errors.New("minecraft/p2p: client is not bound to world")
	}
	return w.client.Join(ctx, w.handleID)
}

// A Session represents a session for a peer-to-peer world.
// It provides useful methods for establishing a connection with the host.
type Session struct {
	client  *Client
	session *mpsd.Session

	world      World
	connection Connection
	worldMu    sync.RWMutex

	nonce string

	readyOnce sync.Once
	readyErr  error
	ready     chan struct{}

	log *slog.Logger
}

// World returns the World data decoded from the custom properties of the multiplayer session.
// The returned data is real-timely updated during the session.
// It might be changed when a new player joins the world.
func (s *Session) World() World {
	s.worldMu.RLock()
	defer s.worldMu.RUnlock()
	return s.world.clone()
}

// updateWorldData updates the world data used by the Session from the custom properties
// received from the Xbox Live MPSD service. An error may be returned if the provided custom
// data cannot be decoded into a World.
func (s *Session) updateWorldData(custom json.RawMessage) error {
	var world World
	if err := json.Unmarshal(custom, &world); err != nil {
		return fmt.Errorf("decode custom properties: %w", err)
	}
	connection, connectionErr := world.Connection()

	s.worldMu.Lock()
	defer s.worldMu.Unlock()

	s.world = world
	if connectionErr == nil {
		s.connection = connection
	} else if err := s.connection.Validate(); err != nil {
		return s.failReadyLocked(fmt.Errorf("select connection method: %w", connectionErr))
	}

	if s.nonce == "" {
		// If the host has not yet generated or published a nonce for the caller, check if
		// one has been added to the Nonces field since the last update.
		xuid := s.client.client.UserInfo().XUID
		if nonces := world.Nonces; nonces != nil {
			nonce, ok := nonces[xuid]
			if ok {
				if nonce == "" {
					return s.failReadyLocked(errors.New("host published empty nonce for caller"))
				}
				s.log.Debug("received nonce from host", slog.String("xuid", xuid), slog.String("nonce", nonce))
				s.nonce = nonce
			}
		}
	}
	if s.nonce != "" {
		s.readyOnce.Do(func() {
			close(s.ready)
		})
	}
	return nil
}

func (s *Session) failReadyLocked(err error) error {
	if s.readyErr == nil {
		s.readyErr = err
	}
	s.readyOnce.Do(func() {
		close(s.ready)
	})
	return err
}

// waitReady blocks until updateWorldData observes both a usable connection and
// the caller nonce, or until ctx is canceled. If both happen together, it
// returns the ready result so terminal host errors are not hidden by ctx.
func (s *Session) waitReady(ctx context.Context) error {
	select {
	case <-s.ready:
	case <-ctx.Done():
		select {
		case <-s.ready:
		default:
			return ctx.Err()
		}
	}
	return s.readyResult()
}

func (s *Session) readyResult() error {
	s.worldMu.RLock()
	defer s.worldMu.RUnlock()
	return s.readyErr
}

// Connection returns the supported connection selected from [World.SupportedConnections].
func (s *Session) Connection() Connection {
	s.worldMu.RLock()
	defer s.worldMu.RUnlock()

	return s.connection
}

// Nonce returns the nonce generated by the host that can be used as the
// [github.com/sandertv/gophertunnel/minecraft/protocol/login.ClientData.Nonce]
// field.
func (s *Session) Nonce() string {
	s.worldMu.RLock()
	defer s.worldMu.RUnlock()

	return s.nonce
}

// Close leaves the session with a 15 seconds timeout.
func (s *Session) Close() error {
	return s.session.Close()
}

// CloseContext leaves the session.
func (s *Session) CloseContext(ctx context.Context) error {
	return s.session.CloseContext(ctx)
}

// handler handles events that might occur in the Session.
type handler struct {
	*Session
}

// HandleSessionChange handles a session change notified by the RTA service.
func (h *handler) HandleSessionChange(session *mpsd.Session) {
	if err := h.updateWorldData(session.Properties().Custom); err != nil {
		h.log.Error("error updating world data", slog.Any("error", err))
	}
}
