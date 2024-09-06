package internal

import "log/slog"

const errorKey = "error"

func ErrAttr(err error) slog.Attr { return slog.Any(errorKey, err) }
