package internal

import (
	"strconv"
	"strings"
)

type Result[T any] struct {
	StatusCode int    `json:"code,omitempty"`
	Data       T      `json:"data,omitempty"`
	Status     string `json:"status,omitempty"`
}

type Error struct {
	StatusCode int                 `json:"code,omitempty"`
	Type       string              `json:"error,omitempty"`
	Code       int                 `json:"errorCode,omitempty"`
	Details    map[string][]string `json:"errorDetails,omitempty"`
	Message    string              `json:"errorMessage,omitempty"`
	Status     string              `json:"status,omitempty"`
}

func (err Error) Error() string {
	b := &strings.Builder{}
	b.WriteString(errorHeader)
	b.WriteByte(errorSeparator)

	b.WriteByte(' ')
	b.WriteString(strconv.Itoa(err.Code))

	if err.Type != "" {
		b.WriteByte(' ')
		b.WriteByte(errorLeftBracket)
		b.WriteString(err.Type)
		b.WriteByte(errorRightBracket)
	}
	if err.Message != "" && err.Message != err.Type {
		// In some cases, message are the equal to the type, so we're trimming some unnecessary fields here...
		// tl;dl avoid returning `1041 (InvalidRequest): "InvalidRequest"`
		b.WriteByte(errorSeparator)
		b.WriteByte(' ')
		b.WriteString(strconv.Quote(err.Message))
	}
	if err.Details != nil {
		b.WriteByte(' ')
		b.WriteByte(errorLeftBracket)

		var index int
		for key, messages := range err.Details {
			b.WriteString(strconv.Quote(key))
			b.WriteByte(errorSeparator)
			b.WriteByte(' ')
			b.WriteByte(errorLeftSquareBracket)

			var elementIndex int
			for _, msg := range messages {
				b.WriteString(strconv.Quote(msg))
				if elementIndex++; elementIndex < len(messages) {
					b.WriteByte(errorBracketSeparator)
					b.WriteByte(' ')
				}
			}

			b.WriteByte(errorRightSquareBracket)

			if index > errorMaxDetails {
				b.WriteByte(errorBracketSeparator)
				b.WriteByte(' ')
				b.WriteString(errorDetailsSuffix)
				break
			}

			if index++; index < len(err.Details) {
				b.WriteByte(errorBracketSeparator)
				b.WriteByte(' ')
			}
		}
		b.WriteByte(errorRightBracket)
	}
	return b.String()
}

const (
	errorHeader           = "playfab"
	errorSeparator        = ':'
	errorBracketSeparator = ','

	errorLeftBracket  = '('
	errorRightBracket = ')'

	errorLeftSquareBracket  = '['
	errorRightSquareBracket = ']'

	errorDetailsSuffix = "..."

	errorMaxDetails = 5

	// playfab: 0001
	// playfab: 0001 (Foo)
	// playfab: 0001 (Foo): "..."
	//
	// playfab: 0001 ("fuga": ["hoge", "huga"])
	// playfab: 0001 (Foo) ("fuga": ["hoge", "huga"])
	// playfab: 0001 (Foo): "..." ("fuga": ["hoge", "huga"])
)
