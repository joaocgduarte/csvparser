package csvparser

import "fmt"

type ParseError struct {
	Msg string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("csvparser: %s", e.Msg)
}

func newUnparsableHeaderErr(header string) ParseError {
	return ParseError{Msg: fmt.Sprintf("header \"%s\" doesn't have an associated parser", header)}
}

func newParseError(err error) ParseError {
	return ParseError{Msg: fmt.Sprintf("file couldn't be parsed: %s", err.Error())}
}
