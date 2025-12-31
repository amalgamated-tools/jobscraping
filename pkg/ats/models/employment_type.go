package models

import (
	"strings"
)

// EmploymentType represents various types of employment commitments.
type EmploymentType int64

const (
	// FullTime represents full-time employment.
	FullTime EmploymentType = iota
	// PartTime represents part-time employment.
	PartTime
	// Contract represents contract-based employment.
	Contract
	// Internship represents internship positions.
	Internship
	// Temporary represents temporary employment.
	Temporary
	// UnknownEmploymentType represents an unknown or unspecified employment type.
	UnknownEmploymentType
)

// ParseEmploymentType converts a string representation of an employment type to its corresponding EmploymentType constant.
func ParseEmploymentType(empType string) EmploymentType {
	switch strings.ToLower(empType) {
	case "full_time", "full time", "full-time", "fulltime", "hourly_ft", "salaried_ft":
		return FullTime
	case "part_time", "part time", "part-time", "parttime", "hourly_pt":
		return PartTime
	case "contract", "contractor":
		return Contract
	case "internship", "intern":
		return Internship
	case "temporary", "temp":
		return Temporary
	default:
		return UnknownEmploymentType // Default to Unknown if unknown
	}
}

// String returns the string representation of the EmploymentType.
func (e EmploymentType) String() string {
	return [...]string{
		"Full Time",
		"Part Time",
		"Contract",
		"Internship",
		"Temporary",
		"Unknown",
	}[e]
}

// ProcessCommitment processes a list of commitment strings to determine and set the EmploymentType of the Job.
func (j *Job) ProcessCommitment(commitments []string) { //nolint:cyclop
	if len(commitments) == 0 {
		j.EmploymentType = FullTime
		return
	}

	updatedCommitments := make([]string, 0, len(commitments))
	for _, commitment := range commitments {
		stringValue := strings.ToLower(strings.TrimSpace(commitment))
		if stringValue == "" {
			// observability.GetGlobalLogger().With(
			// 	zap.String("commitment", commitment),
			// ).Debug("Empty commitment string, skipping")
			continue
		}

		updatedCommitments = append(updatedCommitments, stringValue)
	}

	if j.EmploymentType != UnknownEmploymentType {
		// observability.GetGlobalLogger().With(
		// 	zap.String("employment_type", j.EmploymentType.String()),
		// ).Debug("Employment type already set, ignoring new value")
		for _, c := range updatedCommitments {
			j.AddMetadata("alternate_commitments", c)
		}

		return
	}

	for _, c := range updatedCommitments {
		if strings.Contains(c, "part-time") || strings.Contains(c, "part time") {
			j.EmploymentType = PartTime
			return
		}

		if strings.Contains(c, "contract") || strings.Contains(c, "contractor") || strings.Contains(c, "term") {
			j.EmploymentType = Contract
			return
		}
	}

	j.EmploymentType = FullTime
}
