package internal

type Result[T any] struct {
	Data T `json:"result"`
}
