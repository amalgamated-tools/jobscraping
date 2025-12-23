package models

type EquityType int64

const (
	EquityOffered    EquityType = 1
	EquityNotOffered EquityType = 0
	UnknownEquity    EquityType = -1 // Default value if not specified
)
