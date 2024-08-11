package catalog

import (
	"fmt"
	"github.com/sandertv/gophertunnel/playfab/entity"
	"github.com/sandertv/gophertunnel/playfab/internal"
	"github.com/sandertv/gophertunnel/playfab/title"
	"golang.org/x/text/language"
	"net/http"
	"slices"
)

type Filter struct {
	// Count is the number of returned items included in the SearchResult. The maximum value is 50 and is stored
	// to 10 service-side by default.
	Count int `json:"Count,omitempty"`
	// ContinuationToken is the opaque token used for continuing the query of Search, if any are available. It is
	// normally filled from SearchResult.ContinuationToken.
	ContinuationToken string `json:"ContinuationToken,omitempty"`
	// CustomTags is the optional properties associated with the request.
	CustomTags map[string]any `json:"CustomTags,omitempty"`
	// Entity is the nullable entity.Key to perform any actions.
	Entity *entity.Key `json:"Entity,omitempty"`
	// Filter is an OData query for filtering the SearchResult.
	Filter string `json:"Filter,omitempty"`
	// OrderBy is an OData sort query for sorting the index of SearchResult. Defaulted to relevance.
	OrderBy string `json:"OrderBy,omitempty"`
	// Term is the string terms to be searched.
	Term string `json:"Search,omitempty"`
	// Select is an OData selection query for filtering the fields of returned items included in the SearchResult.
	Select string `json:"Select,omitempty"`
	// Store ...
	Store *StoreReference `json:"Store,omitempty"`

	// Language is used as the `Accept-Language` header of the request and is generally used to display
	// a localized dictionaries catalog items. It must be one of the supported Languages, otherwise it
	// will be ignored by the request hook.
	Language language.Tag `json:"-"`
}

// Search will perform the search query in the catalog using the title. An authorization entity is optionally required in the service-side.
func (f Filter) Search(t title.Title, tok *entity.Token) (*SearchResult, error) {
	if f.Count > 50 {
		return nil, fmt.Errorf("playfab/catalog: Filter: count must be <= 50, got %d", f.Count)
	}
	if f.Entity == nil && tok != nil {
		f.Entity = &tok.Entity
	}

	return internal.Post[*SearchResult](t, "/Catalog/SearchItems", f, func(req *http.Request) {
		if tok != nil {
			tok.SetAuthHeader(req)
		}
		if f.Language != language.Und && slices.ContainsFunc(Languages, func(cmp language.Tag) bool { return f.Language == cmp }) {
			req.Header.Set("Accept-Language", f.Language.String())
		}
	})
}

type SearchResult struct {
	ContinuationToken string `json:"ContinuationToken,omitempty"`
	Items             []Item `json:"Items,omitempty"`
}
