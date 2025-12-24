package models

// EquityType represents whether equity is offered for a job position.
type EquityType int64

const (
	// EquityOffered indicates that equity is offered for the job position.
	EquityOffered EquityType = 1
	// EquityNotOffered indicates that equity is not offered for the job position.
	EquityNotOffered EquityType = 0
	// UnknownEquity represents an unknown or unspecified equity status.
	UnknownEquity EquityType = -1 // Default value if not specified
)
