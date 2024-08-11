package catalog

import (
	"encoding/json"
	"github.com/sandertv/gophertunnel/playfab/entity"
	"time"
)

type Item struct {
	AlternateIDs      []AlternateID              `json:"AlternateIds,omitempty"`
	ContentType       string                     `json:"ContentType,omitempty"`
	Contents          []Content                  `json:"Contents,omitempty"`
	CreationDate      time.Time                  `json:"CreationDate,omitempty"`
	CreatorEntity     entity.Key                 `json:"CreatorEntity,omitempty"`
	DeepLinks         []DeepLink                 `json:"DeepLinks,omitempty"`
	DefaultStackID    string                     `json:"DefaultStackId,omitempty"` // new?
	Description       Dictionary[string]         `json:"Description,omitempty"`
	DisplayProperties map[string]json.RawMessage `json:"DisplayProperties,omitempty"`
	DisplayVersion    string                     `json:"DisplayVersion,omitempty"`
	ETag              string                     `json:"ETag,omitempty"`
	EndDate           time.Time                  `json:"EndDate,omitempty"`
	ID                string                     `json:"Id,omitempty"`
	Images            []Image                    `json:"Images,omitempty"`
	Hidden            *bool                      `json:"IsHidden,omitempty"`
	ItemReferences    []ItemReference            `json:"ItemReferences,omitempty"`
	Keywords          Dictionary[*Keyword]       `json:"Keywords,omitempty"`
	LastModifiedDate  time.Time                  `json:"LastModifiedDate,omitempty"`
	Moderation        ModerationState            `json:"Moderation,omitempty"`
	Platforms         []string                   `json:"Platforms,omitempty"`
	PriceOptions      PriceOptions               `json:"PriceOptions,omitempty"`
	Rating            Rating                     `json:"Rating,omitempty"`
	StartDate         time.Time                  `json:"StartDate,omitempty"`
	StoreDetails      StoreDetails               `json:"StoreDetails,omitempty"`
	Tags              []string                   `json:"Tags,omitempty"`
	Title             Dictionary[string]         `json:"Title,omitempty"`
	Type              string                     `json:"Type,omitempty"`
}

type StoreReference struct {
	AlternateID AlternateID `json:"AlternateId,omitempty"`
	ID          string      `json:"Id,omitempty"`
}

type AlternateID struct {
	Type  string `json:"Type,omitempty"`
	Value string `json:"Value,omitempty"`
}

type Content struct {
	ID               string   `json:"Id,omitempty"`
	MaxClientVersion string   `json:"MaxClientVersion,omitempty"`
	MinClientVersion string   `json:"MinClientVersion,omitempty"`
	Tags             []string `json:"Tags,omitempty"`
	Type             string   `json:"Type,omitempty"`
	URL              string   `json:"Url,omitempty"`
}

type DeepLink struct {
	Platform string `json:"Platform,omitempty"`
	URL      string `json:"Url,omitempty"`
}

type Image struct {
	ID   string `json:"Id,omitempty"`
	Tag  string `json:"Tag,omitempty"`
	Type string `json:"Type,omitempty"`
	URL  string `json:"Url,omitempty"`
}

type ItemReference struct {
	Amount       int          `json:"Amount,omitempty"`
	ID           string       `json:"Id,omitempty"`
	PriceOptions PriceOptions `json:"PriceOptions,omitempty"`
}

type PriceOptions []Price

func (opts PriceOptions) MarshalJSON() ([]byte, error) {
	type raw struct {
		Prices []Price `json:"Prices,omitempty"`
	}
	return json.Marshal(raw{Prices: opts})
}

func (opts *PriceOptions) UnmarshalJSON(b []byte) error {
	var raw struct {
		Prices []Price `json:"Prices,omitempty"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*opts = raw.Prices
	return nil
}

type Price struct {
	Amounts               []PriceAmount `json:"Amounts,omitempty"`
	UnitDurationInSeconds int           `json:"UnitDurationInSeconds,omitempty"`
}

type PriceAmount struct {
	Amount int    `json:"Amount,omitempty"`
	ItemID string `json:"ItemId,omitempty"`
}

type Keyword []string

func (k *Keyword) MarshalJSON() ([]byte, error) {
	type raw struct {
		Values []string `json:"Values,omitempty"`
	}
	return json.Marshal(raw{Values: *k})
}

func (k *Keyword) UnmarshalJSON(b []byte) error {
	var raw struct {
		Values []string `json:"Values,omitempty"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*k = raw.Values
	return nil
}

type ModerationState struct {
	LastModifiedDate time.Time `json:"LastModifiedDate,omitempty"`
	Reason           string    `json:"Reason,omitempty"`
	Status           string    `json:"Status,omitempty"`
}

const (
	ModerationStatusApproved           string = "Approved"
	ModerationStatusAwaitingModeration string = "AwaitingModeration"
	ModerationStatusRejected           string = "Rejected"
	ModerationStatusUnknown            string = "Unknown"
)

type Rating struct {
	Average    float32 `json:"Average,omitempty"`
	Count1Star int     `json:"Count1Star,omitempty"`
	Count2Star int     `json:"Count2Star,omitempty"`
	Count3Star int     `json:"Count3Star,omitempty"`
	Count4Star int     `json:"Count4Star,omitempty"`
	Count5Star int     `json:"Count5Star,omitempty"`
	TotalCount int     `json:"TotalCount,omitempty"`
}

type StoreDetails struct {
	FilterOptions        FilterOptions        `json:"FilterOptions,omitempty"`
	PriceOptionsOverride PriceOptionsOverride `json:"PriceOptionsOverride,omitempty"`
}

type FilterOptions struct {
	Filter          string `json:"Filter,omitempty"`
	IncludeAllItems bool   `json:"IncludeAllItems,omitempty"`
}

type PriceOptionsOverride []PriceOverride

func (opts PriceOptionsOverride) MarshalJSON() ([]byte, error) {
	type raw struct {
		Prices []PriceOverride `json:"Prices,omitempty"`
	}
	return json.Marshal(raw{Prices: opts})
}

func (opts *PriceOptionsOverride) UnmarshalJSON(b []byte) error {
	var raw struct {
		Prices []PriceOverride `json:"Prices,omitempty"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*opts = raw.Prices
	return nil
}

type PriceOverride struct {
	Amounts []PriceAmountOverride `json:"Amounts,omitempty"`
}

type PriceAmountOverride struct {
	FixedValue int    `json:"FixedValue,omitempty"`
	ItemID     string `json:"ItemId,omitempty"`
	Multiplier int    `json:"Multiplier,omitempty"`
}

const (
	ItemTypeBundle      = "bundle"
	ItemTypeCatalogItem = "catalogItem"
	ItemTypeCurrency    = "currency"
	ItemTypeStore       = "store"
	ItemTypeUGC         = "ugc"
)
