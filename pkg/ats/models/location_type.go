package models

import (
	"strings"
)

type LocationType int64

const (
	RemoteLocation LocationType = iota
	OnsiteLocation
	HybridLocation
	UnknownLocation
)

func (j *Job) ProcessLocationType(locations []string) {
	if len(locations) == 0 {
		// observability.GetGlobalLogger().Debug("No locations provided, setting job location type to Unknown")
		j.LocationType = UnknownLocation
		return
	}

	updatedLocations := make([]string, 0, len(locations))
	for _, location := range locations {
		stringValue := strings.ToLower(strings.TrimSpace(location))
		if stringValue == "" {
			// observability.GetGlobalLogger().With(
			// 	zap.String("location", location),
			// ).Debug("Empty location string, skipping")
			continue
		}

		updatedLocations = append(updatedLocations, stringValue)
	}

	if j.LocationType != UnknownLocation {
		// observability.GetGlobalLogger().With(
		// 	zap.String("location", j.LocationType.String()),
		// ).Debug("Location type already set, ignoring new value")
		// we already have a location type set, so we don't override it
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
			if strings.Contains(strings.ToLower(location), "remote") {
				j.LocationType = RemoteLocation
				j.IsRemote = true

				return
			} else if strings.Contains(strings.ToLower(location), "onsite") {
				j.LocationType = OnsiteLocation
				return
			} else if strings.Contains(strings.ToLower(location), "hybrid") {
				j.LocationType = HybridLocation
				return
			} else {
				// observability.GetGlobalLogger().With(
				// 	zap.String("location", location),
				// ).Warn("Unknown job location type")
				j.AddMetadata("alternate_locations", location)
				j.LocationType = UnknownLocation
			}
		}
	}
}

func ParseLocationType(value string) LocationType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "remote", "telecommute":
		return RemoteLocation
	case "onsite":
		return OnsiteLocation
	case "hybrid":
		return HybridLocation
	default:
		return UnknownLocation // Default to Unknown if unknown
	}
}

func (e LocationType) String() string {
	return [...]string{
		"Remote",
		"Onsite",
		"Hybrid",
		"Unknown",
	}[e]
}
