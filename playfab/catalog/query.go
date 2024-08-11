package catalog

import (
	"github.com/sandertv/gophertunnel/playfab/entity"
	"github.com/sandertv/gophertunnel/playfab/internal"
	"github.com/sandertv/gophertunnel/playfab/title"
)

type Query struct {
	AlternateID *AlternateID   `json:"AlternateId,omitempty"`
	CustomTags  map[string]any `json:"CustomTags,omitempty"`
	Entity      *entity.Key    `json:"Entity,omitempty"`
	ID          string         `json:"Id,omitempty"`
}

func (q Query) Item(t title.Title, tok *entity.Token) (zero Item, err error) {
	res, err := internal.Post[*queryResponse](t, "/Catalog/GetItem", q, tok.SetAuthHeader)
	if err != nil {
		return zero, err
	}
	return res.Item, nil
}

type queryResponse struct {
	Item Item `json:"Item,omitempty"`
}
