package design

import (
	"errors"
	"strconv"
)

type wrappedError struct {
	prefix string
	err    error
}

func (e wrappedError) Error() string {
	if e.err == nil {
		return e.prefix
	}
	return e.prefix + ": " + e.err.Error()
}

func (e wrappedError) Unwrap() error {
	return e.err
}

func wrapError(prefix string, err error) error {
	if err == nil {
		return errors.New(prefix)
	}
	return wrappedError{prefix: prefix, err: err}
}

func invalidColorLiteral(value string) error {
	return errors.New("invalid color literal " + strconv.Quote(value))
}

func lowerHex4(value uint16) string {
	hex := strconv.FormatUint(uint64(value), 16)
	if len(hex) >= 4 {
		return hex
	}
	return "0000"[:4-len(hex)] + hex
}
