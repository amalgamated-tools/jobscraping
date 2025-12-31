package models

import "github.com/buger/jsonparser"

// Location represents a geographical location with city, state, postal code, and country.
type Location struct {
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"addressCountry,omitempty"`
}

// String returns a human-readable representation of the Location.
func (l Location) String() string {
	result := ""
	if l.City != "" {
		result += l.City
	}

	if l.State != "" {
		if result != "" {
			result += ", "
		}

		result += l.State
	}

	if l.PostalCode != "" {
		if result != "" {
			result += " "
		}

		result += l.PostalCode
	}

	if l.Country != "" {
		if result != "" {
			result += ", "
		}

		result += l.Country
	}

	return result
}

// ParseLocation parses a Location from JSON data.
func ParseLocation(data []byte) Location {
	location := Location{}

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "city":
			location.City = string(value)
		case "state", "region":
			location.State = string(value)
		case "postalCode":
			location.PostalCode = string(value)
		case "addressCountry", "country":
			location.Country = string(value)
		}

		return nil
	})
	if err != nil {
		return Location{}
	}

	return location
}
