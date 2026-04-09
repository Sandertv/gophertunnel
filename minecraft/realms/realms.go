package realms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"golang.org/x/oauth2"
)

// Client is an instance of the realms api with a token.
type Client struct {
	tokenSrc   oauth2.TokenSource
	xblToken   *auth.XBLToken
	httpClient *http.Client
}

const (
	realmsBaseURL      = "https://bedrock.frontendlegacy.realms.minecraft-services.net"
	realmsRelyingParty = "https://pocket.realms.minecraft.net/"
)

var (
	ErrPlayerNotInRealm = errors.New("player not in realm")
	ErrRealmNotFound    = errors.New("realm not found")
)

// NewClient returns a new Client instance with the supplied token source for authentication.
// If httpClient is nil, http.DefaultClient will be used to request the realms api.
func NewClient(src oauth2.TokenSource, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		tokenSrc:   src,
		httpClient: httpClient,
	}
}

// Player is a player in a Realm.
type Player struct {
	UUID       string `json:"uuid"`
	Name       string `json:"Name"`
	Operator   bool   `json:"operator"`
	Accepted   bool   `json:"accepted"`
	Online     bool   `json:"online"`
	Permission string `json:"permission"`
}

// Realm is the realm structure returned from the api.
type Realm struct {
	// ID is the unique id for this realm.
	ID int `json:"id"`
	// RemoteSubscriptionID is The subscription ID of the realm.
	RemoteSubscriptionID string `json:"remoteSubscriptionID"`
	// Owner is always an empty string.
	Owner string `json:"owner"`
	// OwnerUUID is the XboxUserID (XUID) of the owner.
	OwnerUUID string `json:"ownerUUID"`
	// Name is the name of the Realm.
	Name string `json:"name"`
	// MOTD is always an empty string.
	MOTD string `json:"motd"`
	// DefaultPermission is the default permission level of the Realm world.
	// one of ["MEMBER", "OPERATOR"]
	DefaultPermission string `json:"defaultPermission"`
	// State is the current state of the realm
	// one of: ["OPEN", "CLOSED"]
	State string `json:"state"`
	// DaysLeft is the days remaining before renewal of the Realm as an integer.
	// (always 0 for Realms where the current user is not the owner)
	DaysLeft int `json:"daysLeft"`
	// Expired is whether the Realm has expired as a trial or not.
	Expired bool `json:"expired"`
	// ExpiredTrial is whether the Realm has expired as a trial or not.
	ExpiredTrial bool `json:"expiredTrial"`
	// GracePeriod is whether the Realm is in its grace period after expiry or not.
	GracePeriod bool `json:"gracePeriod"`
	// WorldType is the world type of the currently loaded world.
	WorldType string `json:"worldType"`
	// Players is a list of the players currently online in the realm
	// NOTE: this is only sent when directly requesting a realm.
	Players []Player `json:"players"`
	// MaxPlayers is how many player slots this realm has.
	MaxPlayers int `json:"maxPlayers"`
	// MinigameName is always null
	MinigameName string `json:"minigameName"`
	// MinigameID is always null
	MinigameID string `json:"minigameId"`
	// MinigameImage is always null
	MinigameImage string `json:"minigameImage"`
	// ActiveSlot is unused, always 1
	ActiveSlot int `json:"activeSlot"`
	// Slots is unused, always null
	Slots []struct{} `json:"slots"`
	// Member is Unknown, always false. (even when member or owner)
	Member bool `json:"member"`
	// ClubID is the ID of the associated Xbox Live club as an integer.
	ClubID int64 `json:"clubId"`
	// SubscriptionRefreshStatus is Unknown, always null.
	SubscriptionRefreshStatus struct{} `json:"subscriptionRefreshStatus"`

	// client is the instance of Client that this belongs to.
	client *Client
}

// RealmAddress contains the address returned by the Realms join endpoint along
// with the signalling protocol used for connecting to it.
type RealmAddress struct {
	NetworkProtocol string `json:"networkProtocol"`
	Address         string `json:"address"`

	SessionRegionData struct {
		RegionName     string `json:"regionName"`
		ServiceQuality int    `json:"serviceQuality"`
	} `json:"sessionRegionData"`

	Raw string `json:"-"`
}

// Address requests the address and protocol used to connect to this realm.
// It will wait for the realm to start if it is currently offline.
func (r *Realm) Address(ctx context.Context) (RealmAddress, error) {
	if r.client == nil {
		return RealmAddress{}, fmt.Errorf("realm client is nil")
	}
	return r.client.RealmAddress(ctx, r.ID)
}

// RealmAddress requests the address and protocol used to connect to a realm
// from the api, and waits for the realm to start if it is currently offline.
func (r *Client) RealmAddress(ctx context.Context, realmID int) (RealmAddress, error) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return RealmAddress{}, ctx.Err()
		case <-ticker.C:
			body, status, err := r.request(ctx, fmt.Sprintf("/worlds/%d/join", realmID))
			if err != nil {
				switch status {
				case 503:
					continue
				case 404:
					return RealmAddress{}, ErrRealmNotFound
				case 403:
					return RealmAddress{}, ErrPlayerNotInRealm
				}
				return RealmAddress{}, err
			}

			var address RealmAddress
			if err := json.Unmarshal(body, &address); err != nil {
				return RealmAddress{}, err
			}
			return address, nil
		}
	}
}

// OnlinePlayers gets all the players currently on this realm,
// Returns a 403 error if the current user is not the owner of the Realm.
func (r *Realm) OnlinePlayers(ctx context.Context) (players []Player, err error) {
	body, status, err := r.client.request(ctx, fmt.Sprintf("/worlds/%d", r.ID))
	if err != nil {
		switch status {
		case 403:
			return nil, ErrPlayerNotInRealm
		case 404:
			return nil, ErrRealmNotFound
		}
		return nil, err
	}

	var response Realm
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Players, nil
}

// xboxToken returns the xbox token used for the api.
func (r *Client) xboxToken(ctx context.Context) (*auth.XBLToken, error) {
	if r.xblToken != nil && r.xblToken.Valid() {
		return r.xblToken, nil
	}
	if r.tokenSrc == nil {
		return nil, fmt.Errorf("token source is nil")
	}

	t, err := r.tokenSrc.Token()
	if err != nil {
		return nil, err
	}

	r.xblToken, err = auth.RequestXBLToken(ctx, t, realmsRelyingParty)
	return r.xblToken, err
}

// request sends an http get request to path with the right headers for the api set.
func (r *Client) request(ctx context.Context, path string) (body []byte, status int, err error) {
	if path == "" {
		return nil, 0, fmt.Errorf("path is empty")
	}
	if path[0] != '/' {
		path = "/" + path
	}
	req, err := http.NewRequestWithContext(ctx, "GET", realmsBaseURL+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "MCPE/UWP")
	req.Header.Set("Client-Version", protocol.CurrentVersion)
	xbl, err := r.xboxToken(ctx)
	if err != nil {
		return nil, 0, err
	}
	xbl.SetAuthHeader(req)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= 400 {
		return body, resp.StatusCode, fmt.Errorf("HTTP Error: %d", resp.StatusCode)
	}

	return body, resp.StatusCode, nil
}

// Realm gets a realm by its invite code.
func (c *Client) Realm(ctx context.Context, code string) (Realm, error) {
	body, _, err := c.request(ctx, fmt.Sprintf("/worlds/v1/link/%s", code))
	if err != nil {
		return Realm{}, err
	}

	var realm Realm
	if err := json.Unmarshal(body, &realm); err != nil {
		return Realm{}, err
	}
	realm.client = c

	return realm, nil
}

// Realms gets a list of all realms the token has access to.
func (c *Client) Realms(ctx context.Context) ([]Realm, error) {
	body, _, err := c.request(ctx, "/worlds")
	if err != nil {
		return nil, err
	}

	var response struct {
		Servers []Realm `json:"servers"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	realms := response.Servers
	for i := range realms {
		realms[i].client = c
	}

	return realms, nil
}
