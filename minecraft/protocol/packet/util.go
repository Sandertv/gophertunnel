package packet

import (
	"errors"
	"strings"
)

// ChainErr chains together a variadic amount of errors into a single error and returns it. If all errors
// passed are nil, the error returned will also be nil.
func ChainErr(err ...error) error {
	var msg string
	for _, e := range err {
		if e == nil {
			continue
		}
		msg += e.Error() + "\n"
	}
	if msg == "" {
		return nil
	}
	return errors.New(strings.TrimRight(msg, "\n"))
}
