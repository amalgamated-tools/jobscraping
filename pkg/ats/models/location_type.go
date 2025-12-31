package models

import (
	"strings"
)

// LocationType represents the type of job location.
type LocationType int64

const (
	// RemoteLocation represents a remote job location.
	RemoteLocation LocationType = iota
	// OnsiteLocation represents an onsite job location.
	OnsiteLocation
	// HybridLocation represents a hybrid job location.
	HybridLocation
	// UnknownLocationType represents an unknown or unspecified job location.
	UnknownLocationType
)

// ProcessLocationType processes a list of location strings to determine and set the LocationType of the Job.
func (j *Job) ProcessLocationType(locations []string) { //nolint:cyclop
	if len(locations) == 0 {
		// observability.GetGlobalLogger().Debug("No locations provided, setting job location type to Unknown")
		j.LocationType = UnknownLocationType
		return
	}

	updatedLocations := make([]string, 0, len(locations))
	for _, location := range locations {
		stringValue := strings.ToLower(strings.TrimSpace(location))
		if stringValue == "" {
			continue
		}

		updatedLocations = append(updatedLocations, stringValue)
	}

	if j.LocationType != UnknownLocationType {
		for _, location := range updatedLocations {
			j.AddMetadata("alternate_locations", location)
		}

		return
	}

	for _, location := range updatedLocations {
		switch location {
		case "remote", "telecommute":
			j.LocationType = RemoteLocation
			j.IsRemote = true

			return
		case "onsite":
			j.LocationType = OnsiteLocation
			return
		case "hybrid":
			j.LocationType = HybridLocation
			return
		default:
			switch {
			case strings.Contains(strings.ToLower(location), "remote"):
				j.LocationType = RemoteLocation
				j.IsRemote = true

				return
			case strings.Contains(strings.ToLower(location), "anywhere"):
				j.LocationType = RemoteLocation
				j.IsRemote = true

				return
			case strings.Contains(strings.ToLower(location), "onsite"):
				j.LocationType = OnsiteLocation
				return
			case strings.Contains(strings.ToLower(location), "hybrid"):
				j.LocationType = HybridLocation
				return
			default:
				j.AddMetadata("alternate_locations", location)
				j.LocationType = UnknownLocationType
			}
		}
	}
}

// ParseLocationType converts a string representation of a location type to its corresponding LocationType constant.
func ParseLocationType(value string) LocationType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "remote", "telecommute":
		return RemoteLocation
	case "onsite", "in_office":
		return OnsiteLocation
	case "hybrid":
		return HybridLocation
	default:
		return UnknownLocationType // Default to Unknown if unknown
	}
}

// String returns the string representation of the LocationType.
func (e LocationType) String() string {
	return [...]string{
		"Remote",
		"Onsite",
		"Hybrid",
		"Unknown",
	}[e]
}
