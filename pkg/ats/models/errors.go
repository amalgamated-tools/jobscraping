package models

import "errors"

var (
	// ErrUnableToParseCompensation is returned when a compensation string cannot be parsed.
	ErrUnableToParseCompensation = errors.New("unable to parse compensation string")
)
