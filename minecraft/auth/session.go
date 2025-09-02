package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/df-mc/go-playfab"
	"github.com/df-mc/go-playfab/title"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/auth/franchise"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
)

type Session struct {
	env                  franchise.AuthorizationEnvironment
	obtainer             *XBLTokenObtainer
	legacyMultiplayerXBL *XBLToken
	playfabIdentity      *playfab.Identity
	mcToken              *franchise.Token
	src                  oauth2.TokenSource
	deviceType           Device
	conf                 *franchise.TokenConfig
}

// SessionFromTokenSource creates a session from an XBOX token source and returns it.
func SessionFromTokenSource(src oauth2.TokenSource, deviceType Device, ctx context.Context) (s *Session, err error) {
	s = &Session{src: src, deviceType: deviceType}
	if err := s.login(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) login(ctx context.Context) error {
	tok, err := s.src.Token()
	if err != nil {
		return fmt.Errorf("request token: %w", err)
	}
	err = s.initDiscovery()
	if err != nil {
		return fmt.Errorf("init discovery: %w", err)
	}

	region, _ := language.English.Region()
	s.conf = &franchise.TokenConfig{
		Device: &franchise.DeviceConfig{
			ApplicationType: franchise.ApplicationTypeMinecraftPE,
			Capabilities:    []string{franchise.CapabilityRayTracing},
			GameVersion:     protocol.CurrentVersion,
			ID:              uuid.New(),
			Memory:          strconv.FormatUint(rand.Uint64(), 10),
			Platform:        franchise.PlatformWindows10,
			PlayFabTitleID:  s.env.PlayFabTitleID,
			StorePlatform:   franchise.StorePlatformUWPStore,
			Type:            franchise.DeviceTypeWindows10,
		},
		User: &franchise.UserConfig{
			Language:     language.English,
			LanguageCode: language.AmericanEnglish,
			RegionCode:   region.String(),
			TokenType:    franchise.TokenTypePlayFab,
		},
		Environment: &s.env,
	}

	s.obtainer, err = NewXBLTokenObtainer(tok, s.deviceType, ctx)
	if err != nil {
		return fmt.Errorf("obtain device token: %w", err)
	}

	if err = s.loginWithPlayfab(ctx); err != nil {
		return err
	}

	return s.obtainMcToken(ctx)
}

func (s *Session) initDiscovery() error {
	discovery, err := franchise.Discover(protocol.CurrentVersion)
	if err != nil {
		return fmt.Errorf("discover: %w", err)
	}

	if err := discovery.Environment(&s.env, franchise.EnvironmentTypeProduction); err != nil {
		return fmt.Errorf("decode environment: %w", err)
	}

	return nil
}

func (s *Session) loginWithPlayfab(ctx context.Context) (err error) {
	playfabXBL, err := s.obtainer.RequestXBLToken(ctx, "http://playfab.xboxlive.com/")
	if err != nil {
		return fmt.Errorf("request playfab token: %w", err)
	}

	s.playfabIdentity, err = XBOXPlayfabLoginConfig{
		LoginConfig: playfab.LoginConfig{
			Title:         title.Title(s.env.PlayFabTitleID),
			CreateAccount: true,
		},
	}.Login(playfabXBL)
	if err != nil {
		return fmt.Errorf("error logging in to playfab: %w", err)
	}

	return nil
}

func (s *Session) obtainMcToken(ctx context.Context) (err error) {
	playfabIdentity, err := s.PlayfabIdentity(ctx)
	if err != nil {
		return err
	}
	s.conf.User.Token = playfabIdentity.SessionTicket
	s.mcToken, err = s.conf.Token()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	return nil
}

// Obtainer returns the Xbox token obtainer, which contains the device token
func (s *Session) Obtainer() *XBLTokenObtainer {
	return s.obtainer
}

// PlayfabIdentity returns the user's Playfab identity, which includes the session ticket.
func (s *Session) PlayfabIdentity(ctx context.Context) (*playfab.Identity, error) {
	if pastExpirationTime(s.playfabIdentity.EntityToken.Expiration) {
		if err := s.loginWithPlayfab(ctx); err != nil {
			return nil, err
		}
	}
	return s.playfabIdentity, nil
}

// MCToken returns the session token, or refreshes it if it has expired.
func (s *Session) MCToken(ctx context.Context) (*franchise.Token, error) {
	if pastExpirationTime(s.mcToken.ValidUntil) {
		if err := s.obtainMcToken(ctx); err != nil {
			return nil, err
		}
	}
	return s.mcToken, nil
}

// LegacyMultiplayerXBL requests an XBL token for the old multiplayer endpoint.
func (s *Session) LegacyMultiplayerXBL(ctx context.Context) (tok *XBLToken, err error) {
	if s.legacyMultiplayerXBL == nil || pastExpirationTime(s.legacyMultiplayerXBL.AuthorizationToken.NotAfter) {
		s.legacyMultiplayerXBL, err = s.obtainer.RequestXBLToken(ctx, "https://multiplayer.minecraft.net/")
		if err != nil {
			return nil, fmt.Errorf("request legacy multiplayer token: %w", err)
		}
	}
	return s.legacyMultiplayerXBL, nil
}

// MultiplayerToken requests a multiplayer token from Microsoft. The token can be reused, but is not
// reused by the vanilla client. Calling SetKey will clear the saved token.
func (s *Session) MultiplayerToken(ctx context.Context, key *ecdsa.PrivateKey) (tok *franchise.MultiplayerToken, err error) {
	mcToken, err := s.MCToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("obtain MCToken: %w", err)
	}
	return franchise.RequestMultiplayerToken(ctx, s.env, mcToken, key)
}

const expirationTimeDelta = time.Minute

func pastExpirationTime(expirationTime time.Time) bool {
	return time.Now().After(expirationTime.Add(-expirationTimeDelta))
}
