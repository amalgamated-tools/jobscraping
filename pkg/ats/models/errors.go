package models

import "errors"

var (
	// ErrUnableToParseCompensation is returned when a compensation string cannot be parsed.
	ErrUnableToParseCompensation = errors.New("received non-OK status code")
)
