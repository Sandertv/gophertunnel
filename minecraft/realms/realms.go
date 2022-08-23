package realms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// RealmsApi is an instance of the realms api with a token
type RealmsApi struct {
	token_src  oauth2.TokenSource
	xbox_token *auth.XBLToken
}

// NewRealmsApi returns a new RealmsApi instance with the supplied token source for authentication
func NewRealmsApi(src oauth2.TokenSource) *RealmsApi {
	return &RealmsApi{
		token_src: src,
	}
}

// RealmPlayer in a realm returned from the api
type RealmPlayer struct {
	UUID       string `json:"uuid"`
	Name       string `json:"Name"`
	Operator   bool   `json:"operator"`
	Accepted   bool   `json:"accepted"`
	Online     bool   `json:"online"`
	Permission string `json:"permission"`
}

// Realm returned from the api
type Realm struct {
	// the unique id for this realm.
	Id int `json:"id"`
	// The subscription ID of the realm.
	RemoteSubscriptionId string `json:"remoteSubscriptionId"`
	// This is always an empty string.
	Owner string `json:"owner"`
	// The XboxUserID (XUID) of the owner.
	OwnerUUID string `json:"ownerUUID"`
	// The name of the Realm.
	Name string `json:"name"`
	// This is always an empty string.
	Motd string `json:"motd"`
	// The default permission level of the Realm world.
	// one of ["MEMBER", "OPERATOR"]
	DefaultPermission string `json:"defaultPermission"`
	// The current state of the realm
	// one of: ["OPEN", "CLOSED"]
	State string `json:"state"`
	//  The days remaining before renewal of the Realm as an integer.
	// (always 0 for Realms where the current user is not the owner)
	DaysLeft int `json:"daysLeft"`
	// whether the Realm has expired as a trial or not.
	Expired bool `json:"expired"`
	// whether the Realm has expired as a trial or not.
	ExpiredTrial bool `json:"expiredTrial"`
	// whether the Realm is in its grace period after expiry or not.
	GracePeriod bool `json:"gracePeriod"`
	// The world type of the currently loaded world
	// one of: ["NORMAL", "FLAT"?]
	WorldType string `json:"worldType"`
	// Players is a list of the players currently online in the realm
	// NOTE: this is only sent when directly requesting a realm
	Players []RealmPlayer `json:"players"`
	// MaxPlayers how many player slots this realm has
	MaxPlayers int `json:"maxPlayers"`
	// always null
	MinigameName string `json:"minigameName"`
	// always null
	MinigameId string `json:"minigameId"`
	// always null
	MinigameImage string `json:"minigameImage"`
	// unused, always 1
	ActiveSlot int `json:"activeSlot"`
	// unused, always null
	Slots []struct{} `json:"slots"`
	// Unknown, always false. (even when member or owner)
	Member bool `json:"member"`
	// The ID of the associated Xbox Live club as an integer.
	ClubId int `json:"clubId"`
	// Unknown, always null
	SubscriptionRefreshStatus struct{} `json:"subscriptionRefreshStatus"`

	// instance of RealmsApi that this belongs to
	_realmsApi *RealmsApi
}

// GetAddress requests the address and port to connect to this realm from the api
// if wait is true it will poll the api until the server is started
// else it will just return the http error
func (r *Realm) GetAddress(ctx context.Context, wait bool) (address string, err error) {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	i := 0
	for range ticker.C {
		i++
		body, status, err := r._realmsApi.getRequest(ctx, fmt.Sprintf("/worlds/%d/join", r.Id))
		if err != nil {
			if status == 503 && wait {
				fmt.Printf("Waiting for the realm to start... %d\033[K\r", i)
				continue
			}
			return "", err
		}
		println()

		var data struct {
			Address       string `json:"address"`
			PendingUpdate bool   `json:"pendingUpdate"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return "", err
		}
		return data.Address, nil
	}
	panic("unreachable")
}

// GetPlayers gets all the players currently on this realm
func (r *Realm) GetPlayers(ctx context.Context) (players []RealmPlayer, err error) {
	body, _, err := r._realmsApi.getRequest(ctx, fmt.Sprintf("/worlds/%d", r.Id))
	if err != nil {
		return nil, err
	}

	var response Realm
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Players, nil
}

// getXboxToken returns the xbox token used for the api
func (r *RealmsApi) getXboxToken(ctx context.Context) (*auth.XBLToken, error) {
	if r.xbox_token != nil {
		return r.xbox_token, nil
	}

	t, err := r.token_src.Token()
	if err != nil {
		return nil, err
	}

	r.xbox_token, err = auth.RequestXBLToken(ctx, t, "https://pocket.realms.minecraft.net/")
	return r.xbox_token, err
}

// getRequest sends a get request to path with the right headers for the api set
func (r *RealmsApi) getRequest(ctx context.Context, path string) (body []byte, status int, err error) {
	if string(path[0]) != "/" {
		path = "/" + path
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://pocket.realms.minecraft.net%s", path), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "MCPE/UWP")
	req.Header.Set("Client-Version", "1.10.1")
	xbl, err := r.getXboxToken(ctx)
	if err != nil {
		return nil, 0, err
	}
	xbl.SetAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
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

// GetRealms gets a list of all realms the token has access to
func (r *RealmsApi) GetRealms(ctx context.Context) ([]Realm, error) {
	body, _, err := r.getRequest(ctx, "/worlds")
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
		realms[i]._realmsApi = r
	}

	return realms, nil
}
