package gatherings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/df-mc/go-playfab/v2"
	"github.com/df-mc/go-playfab/v2/catalog"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"github.com/sandertv/gophertunnel/minecraft/service/internal"
	"golang.org/x/text/language"
)

// Environment represents an environment for gatherings service.
type Environment struct {
	// ServiceURI is the base endpoint URL for gatherings service.
	ServiceURI *url.URL `json:"serviceUri"`

	// HTTPClient is the HTTP client used by Environment and Client
	// to make requests to the endpoints provided by gatherings service.
	// If nil, [http.DefaultClient] will be used instead.
	HTTPClient *http.Client `json:"-"`
}

// DefaultEnvironment is the default [Environment] used when callers do not
// provide one explicitly. It may become outdated if Mojang changes the service endpoint.
var DefaultEnvironment = &Environment{
	ServiceURI: &url.URL{
		Scheme: "https",
		Host:   "gatherings-secondary.franchise.minecraft-services.net",
	},
}

// NewClient returns a new [Client] using [DefaultEnvironment].
func NewClient(src service.TokenSource) *Client {
	return DefaultEnvironment.New(src)
}

// New returns a new Client using the provided [service.TokenSource] for authorization.
func (e *Environment) New(src service.TokenSource) *Client {
	return &Client{
		src:    src,
		client: e.httpClient(),
		env:    e,
	}
}

// httpClient returns the HTTP client used for requests made by [Environment].
func (e *Environment) httpClient() *http.Client {
	if e.HTTPClient != nil {
		return e.HTTPClient
	}
	return http.DefaultClient
}

// Client provides access to the Minecraft Gatherings service.
// It can search for online experiences and featured servers offered by
// Mojang-partnered creators.
type Client struct {
	src    service.TokenSource
	client *http.Client
	env    *Environment
}

// ParseExperience parses the display properties of item into an [Experience].
// It accepts [catalog.Item] values returned by other APIs, such as
// [catalog.Client.SearchItems]. The returned [Experience] is ready to use with Client.
func (c *Client) ParseExperience(item catalog.Item) (*Experience, error) {
	var exp Experience
	if err := json.Unmarshal(item.DisplayProperties, &exp); err != nil {
		return nil, fmt.Errorf("decode display properties for %s: %w", item.ID, err)
	}
	exp.Item, exp.client = item, c
	return &exp, nil
}

// ParseFeaturedServer parses the display properties of item into a
// [FeaturedServer]. It accepts [catalog.Item] values returned by other APIs,
// such as [catalog.Client.SearchItems]. The returned [FeaturedServer] is ready
// to use.
func (c *Client) ParseFeaturedServer(item catalog.Item) (*FeaturedServer, error) {
	var server FeaturedServer
	if err := json.Unmarshal(item.DisplayProperties, &server); err != nil {
		return nil, fmt.Errorf("decode display properties for %s: %w", item.ID, err)
	}
	server.Item = item
	return &server, nil
}

// SearchItems performs search for catalog items that describe experiences or
// featured servers.
//
// This method is similar to [catalog.Client.SearchItems], but the gatherings
// service may return items that are not visible to regular player entities.
// The endpoint appears to return the same result regardless of the provided
// filter, so this method is only useful for listing gatherings experiences
// and featured servers.
func (c *Client) SearchItems(ctx context.Context, filter catalog.SearchFilter, opts ...playfab.RequestOption) (*catalog.SearchResult, error) {
	buf := &bytes.Buffer{}
	defer buf.Reset()
	if err := json.NewEncoder(buf).Encode(filter); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}

	requestURL := c.env.ServiceURI.JoinPath("/api/v2.0/discovery/blob/client").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, buf)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "libhttpclient/1.0.0.0")

	opts = append(opts, playfab.AcceptLanguage([]language.Tag{language.AmericanEnglish}))
	for _, opt := range opts {
		if err := opt(req); err != nil {
			return nil, fmt.Errorf("apply request option: %w", err)
		}
	}

	token, err := c.src.ServiceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request service token: %w", err)
	}
	token.SetAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		// PlayFab API Result
		var result struct {
			Data *catalog.SearchResult `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}
		if result.Data == nil {
			return nil, errors.New("service/gatherings: invalid search result")
		}
		return result.Data, nil
	default:
		return nil, internal.Err(resp)
	}
}

// Experiences returns the currently listed online experiences.
func (c *Client) Experiences(ctx context.Context) ([]*Experience, error) {
	result, err := c.SearchItems(ctx, catalog.SearchFilter{
		// The gatherings service appears to ignore this [catalog.SearchFilter]
		// and return the same result for any filter.
		Filter:  "(ContentType eq '3PP_V2.0') and (DisplayProperties/experienceId ne '')",
		OrderBy: "startDate desc",
		Select:  "images",
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("service/gatherings: empty search result")
	}
	experiences := make([]*Experience, 0, len(result.Items))
	for _, item := range result.Items {
		var exp Experience
		if err := json.Unmarshal(item.DisplayProperties, &exp); err != nil {
			return nil, fmt.Errorf("service/gatherings: decode experience (%s): %w", item.ID, err)
		}
		if !exp.Valid() {
			continue
		}
		exp.Item, exp.client = item, c
		experiences = append(experiences, &exp)
	}
	return experiences, nil
}

// FeaturedServers returns the currently listed featured servers.
func (c *Client) FeaturedServers(ctx context.Context) ([]*FeaturedServer, error) {
	result, err := c.SearchItems(ctx, catalog.SearchFilter{
		// The gatherings service appears to ignore this [catalog.SearchFilter]
		// and return the same result for any filter.
		Filter:  "(ContentType eq '3PP_V2.0') and (DisplayProperties/experienceId eq '')",
		OrderBy: "startDate desc",
		Select:  "images", // I don't think this is needed
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("service/gatherings: empty search result")
	}
	servers := make([]*FeaturedServer, 0, len(result.Items))
	for _, item := range result.Items {
		server, err := c.ParseFeaturedServer(item)
		if err != nil {
			return nil, err
		}
		if !server.Valid() {
			continue
		}
		servers = append(servers, server)
	}
	return servers, nil
}

// JoinExperience attempts to join the experience using the ID.
// The resulting [Address] may point to the server nearest to the caller.
func (c *Client) JoinExperience(ctx context.Context, id uuid.UUID) (*Address, error) {
	if id == uuid.Nil {
		return nil, errors.New("service/gatherings: experience ID is nil")
	}

	buf := &bytes.Buffer{}
	defer buf.Reset()
	if err := json.NewEncoder(buf).Encode(map[string]any{
		"experienceId": id,
	}); err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	requestURL := c.env.ServiceURI.JoinPath("/api/v2.0/join/experience").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, buf)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "libhttpclient/1.0.0.0")

	token, err := c.src.ServiceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request service token: %w", err)
	}
	token.SetAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		var result internal.Result[*Address]
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}
		if result.Data == nil {
			return nil, errors.New("service/gatherings: invalid join result")
		}
		return result.Data, nil
	default:
		return nil, internal.Err(resp)
	}
}
