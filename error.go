package csvparser

import "fmt"

type parseError struct {
	Msg string
}

func (e parseError) Error() string {
	return fmt.Sprintf("csvparser: %s", e.Msg)
}

func newUnparsableHeaderErr(header string) parseError {
	return parseError{Msg: fmt.Sprintf("header \"%s\" doesn't have an associated parser", header)}
}

func newparseError(err error) parseError {
	return parseError{Msg: fmt.Sprintf("file couldn't be parsed: %s", err.Error())}
}
