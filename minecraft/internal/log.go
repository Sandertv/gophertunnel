package internal

import (
	"context"
	"log/slog"
)

// DiscardHandler implements a slog.Handler that is always disabled. Each of its
// methods return immediately without any code running.
type DiscardHandler struct{}

func (d DiscardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (d DiscardHandler) Handle(context.Context, slog.Record) error { return nil }
func (d DiscardHandler) WithAttrs([]slog.Attr) slog.Handler        { return d }
func (d DiscardHandler) WithGroup(string) slog.Handler             { return d }
