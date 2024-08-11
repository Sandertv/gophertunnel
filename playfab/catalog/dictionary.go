package catalog

import (
	"encoding/json"
	"errors"
	"golang.org/x/text/language"
	"sort"
	"strings"
)

type Dictionary[T comparable] struct{ self map[string]T }

func (dict *Dictionary[T]) Message(tag language.Tag) (zero T) {
	msg, ok := dict.Lookup(tag.String())
	if !ok || msg == zero {
		return dict.Neutral()
	}
	return msg
}

func (dict *Dictionary[T]) Lookup(key string) (zero T, ok bool) {
	for compare, msg := range dict.self {
		if strings.EqualFold(compare, key) {
			return msg, true
		}
	}
	return zero, false
}

const neutralKey = "NEUTRAL"

func (dict *Dictionary[T]) Neutral() (zero T) {
	for key, msg := range dict.self {
		if strings.EqualFold(key, neutralKey) && msg != zero {
			return msg
		}
	}
	return
}

func (dict *Dictionary[T]) Map() map[string]T { return dict.self }

func (dict *Dictionary[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(dict.self)
}

func (dict *Dictionary[T]) UnmarshalJSON(b []byte) error {
	if dict == nil {
		return errors.New("playfab/catalog: cannot unmarshal a nil *Dictionary")
	}
	return json.Unmarshal(b, &dict.self)
}

var Languages []language.Tag

var unsortedLanguages = []string{
	"en-US",
	"ja-JP",
	"ko-KR",
	"ru-RU",
	"en-GB",
}

var languages = []string{
	"hu-HU",
	"pl-PL",
	"fr-CA",
	"nl-NL",
	"tr-TR",
	"uk-UA",
	"zh-CN",
	"es-MX",
	"id-ConnectionID",
	"sk-SK",
	"pt-BR",
	"sv-SE",
	"de-DE",
	"fi-FI",
	"fr-FR",
	"nb-NO",
	"bg-BG",
	"cs-CZ",
	"pt-PT",
	"da-DK",
	"it-IT",
	"el-GR",
	"es-ES",
	"zh-TW",
}

func init() {
	sort.Strings(languages)

	for _, key := range append(unsortedLanguages, languages...) {
		Languages = append(Languages, language.MustParse(key))
	}
}
