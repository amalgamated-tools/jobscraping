package models

import (
	"strings"
)

type EmploymentType int64

const (
	FullTime EmploymentType = iota
	PartTime
	Contract
	Internship
	Temporary
	UnknownEmploymentType
)

func ParseEmploymentType(empType string) EmploymentType {
	switch strings.ToLower(empType) {
	case "full_time", "full time", "full-time", "fulltime":
		return FullTime
	case "part_time", "part time", "part-time", "parttime":
		return PartTime
	case "contract":
		return Contract
	case "internship", "intern":
		return Internship
	case "temporary", "temp":
		return Temporary
	default:
		return UnknownEmploymentType // Default to Unknown if unknown
	}
}

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

func (j *Job) ProcessCommitment(commitments []string) {
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
