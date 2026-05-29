//go:build !js && !wasm

package main

import (
	"bufio"
	"errors"
	"io"
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

func panicErr(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	return errors.New(valueText(r))
}

func valueText(v any) string {
	switch x := v.(type) {
	case nil:
		return "<nil>"
	case string:
		return x
	case error:
		return x.Error()
	case bool:
		return strconv.FormatBool(x)
	case int:
		return strconv.Itoa(x)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	default:
		return "unknown error"
	}
}

func writeText(w io.Writer, text string) {
	_, _ = io.WriteString(w, text)
}

func writeLine(w io.Writer, parts ...any) {
	for i, part := range parts {
		if i > 0 {
			writeText(w, " ")
		}
		writeText(w, valueText(part))
	}
	writeText(w, "\n")
}

func readWord(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}
