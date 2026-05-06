package test

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/df-mc/go-playfab"
	"github.com/df-mc/go-xsapi/v2"
	"github.com/df-mc/go-xsapi/v2/xal/sisu"
	"github.com/df-mc/go-xsapi/v2/xal/xasd"
	"github.com/df-mc/go-xsapi/v2/xal/xsts"
	"github.com/go-jose/go-jose/v4"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"golang.org/x/oauth2"
)

func PleaseRemoveThisBeforePRingSebWontAllowThis(t testing.TB) *User {
	cacheDir := filepath.Join(os.TempDir(), "gophertunnel-integration-testing-credentials")
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		t.Fatalf("error making credentials directory at temp: %s", err)
	}

	cache, _, err := readCache[UserCache](filepath.Join(cacheDir, "user.json"))
	if err != nil {
		t.Fatalf("error reading cache: %s", err)
	}

	u, err := newUser(t.Context(), t, cache)
	if err != nil {
		t.Fatalf("error creating user: %s", err)
	}
	t.Cleanup(func() {
		if err := writeJSON(filepath.Join(cacheDir, "user.json"), u.Cache()); err != nil {
			t.Fatalf("cleanup: error writing user.json: %s", err)
		}
		if err := u.Close(); err != nil {
			t.Fatalf("error closing user: %s", err)
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), time.Second*15)
	defer cancel()
	if err := u.login(ctx); err != nil {
		var acct *sisu.AccountCreationRequiredError
		if errors.As(err, &acct) {
			t.Fatalf("create an Xbox Live account at %s", acct.SignupURL)
		}
		t.Fatalf("error logging in to network services: %s", err)
	}
	if !slices.Contains(u.XSAPI().UserInfo().Privileges, xsts.PrivilegeMultiplayer) {
		t.Logf("user doesn't have privilege to do multiplayer")
	}
	t.Logf("logged in as %s (%s)", u.XSAPI().UserInfo().GamerTag, u.XSAPI().UserInfo().XUID)
	return u
}

func newUser(ctx context.Context, t testing.TB, cache UserCache) (*User, error) {
	if cache.MSAToken == nil {
		d, err := auth.AndroidConfig.DeviceAuth(ctx)
		if err != nil {
			return nil, fmt.Errorf("request device auth code flow: %w", err)
		}
		t.Logf("sign in to your Microsoft Account at %s using the code %s", d.VerificationURI, d.UserCode)
		pollCtx, cancel := context.WithTimeout(ctx, time.Minute*5)
		defer cancel()
		token, err := auth.AndroidConfig.DeviceAccessToken(pollCtx, d)
		if err != nil {
			return nil, fmt.Errorf("poll device access token: %w", err)
		}
		cache.MSAToken = token
	}

	u := &User{
		t: t,

		msa: auth.AndroidConfig.TokenSource(context.Background(), cache.MSAToken),
	}
	u.session = auth.AndroidConfig.New(u.msa, &sisu.SessionConfig{
		Snapshot:          cache.Snapshot,
		DeviceTokenSource: xasd.ReuseTokenSource(auth.AndroidConfig.Config.Config, cache.Device.Token, cache.Device.ProofKey),
	})
	return u, nil
}

type User struct {
	t testing.TB

	// msa supplies [oauth2.Token] for user's Microsoft Account.
	msa oauth2.TokenSource
	// session is the SISU session used to sign in to Xbox Live.
	session *sisu.Session

	xsapi   *xsapi.Client
	playfab *playfab.Client
	// minecraft supplies service tokens for authorizing with Minecraft network services.
	minecraft service.TokenSource

	discovery *service.Discovery
	authEnv   *service.AuthorizationEnvironment
}

func (u *User) login(ctx context.Context) (err error) {
	discovery, err := service.Default(ctx)
	if err != nil {
		return fmt.Errorf("discover minecraft network services: %w", err)
	}
	u.authEnv = new(service.AuthorizationEnvironment)
	if err := discovery.Environment(u.authEnv); err != nil {
		return fmt.Errorf("resolve environment for %q: %w", u.authEnv.ServiceName(), err)
	}

	u.xsapi, err = xsapi.NewClient(u.session)
	if err != nil {
		return fmt.Errorf("login to xbox live network services: %w", err)
	}
	u.playfab, err = playfab.LoginWithXbox(ctx, u.authEnv.PlayFabTitleID, u.xsapi, playfab.ClientConfig{CreateAccount: true})
	if err != nil {
		return fmt.Errorf("login to playfab account: %w", err)
	}
	u.minecraft = u.authEnv.TokenSource(u.playfab, service.TokenConfig{})
	return nil
}

func (u *User) MSA() oauth2.TokenSource {
	return u.msa
}

func (u *User) Session() *sisu.Session {
	return u.session
}

func (u *User) XSAPI() *xsapi.Client {
	if u.xsapi == nil {
		panic("not logged in to xbox live")
	}
	return u.xsapi
}

func (u *User) PlayFab() *playfab.Client {
	if u.playfab == nil {
		panic("not logged in to playfab")
	}
	return u.playfab
}

func (u *User) Minecraft() service.TokenSource {
	return u.minecraft
}

func (u *User) Cache() *UserCache {
	cache := &UserCache{}
	if u.msa != nil {
		token, err := u.msa.Token()
		if err != nil {
			u.t.Errorf("error retrieving MSA token: %s", err)
		} else {
			cache.MSAToken = token
		}
	}
	if u.session != nil {
		cache.Snapshot = u.session.Snapshot()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		token, err := u.session.DeviceToken(ctx)
		if err != nil {
			u.t.Errorf("error retrieving device token: %s", err)
		}
		cache.Device = DeviceCache{
			Token:    token,
			ProofKey: u.session.ProofKey(),
		}
	}
	return cache
}

func (u *User) Close() (err error) {
	if u.xsapi != nil {
		if err := u.xsapi.Close(); err != nil {
			err = errors.Join(err, fmt.Errorf("close XSAPI client: %w", err))
		}
	}
	if u.playfab != nil {
		// stop refreshing entity tokens
		if err := u.playfab.Close(); err != nil {
			err = errors.Join(err, fmt.Errorf("close playfab client: %w", err))
		}
	}
	return err
}

type UserCache struct {
	MSAToken *oauth2.Token  `json:"msa_token"`
	Snapshot *sisu.Snapshot `json:"session_snapshot"`
	Device   DeviceCache    `json:"device"`
}

type DeviceCache struct {
	ProofKey *ecdsa.PrivateKey `json:"proof_key"`
	Token    *xasd.Token       `json:"token"`
}

func (c *DeviceCache) MarshalJSON() ([]byte, error) {
	type Alias DeviceCache
	return json.Marshal(struct {
		*Alias
		ProofKey jose.JSONWebKey `json:"proof_key"`
	}{
		Alias: (*Alias)(c),
		ProofKey: jose.JSONWebKey{
			Algorithm: string(jose.ES256),
			Use:       "sig",
			Key:       c.ProofKey,
		},
	})
}

func (c *DeviceCache) UnmarshalJSON(b []byte) error {
	type Alias DeviceCache
	data := struct {
		*Alias
		ProofKey jose.JSONWebKey `json:"proof_key"`
	}{Alias: (*Alias)(c)}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	var ok bool
	c.ProofKey, ok = data.ProofKey.Key.(*ecdsa.PrivateKey)
	if !ok {
		return fmt.Errorf("invalid proof key type: %T, expected *ecdsa.PrivateKey", data.ProofKey.Key)
	}
	return nil
}

func readCache[T any](path string) (value T, ok bool, err error) {
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return value, false, nil
	} else if err != nil {
		return value, false, fmt.Errorf("stat: %w", err)
	} else if stat.IsDir() {
		return value, false, fmt.Errorf("%q is a directory", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return value, false, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&value); err != nil {
		return value, false, fmt.Errorf("decode contents of %q: %w", path, err)
	}
	return value, true, nil
}

func writeJSON(path string, v any) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(v); err != nil {
		return fmt.Errorf("encode contents of %q: %w", path, err)
	}
	return nil
}
