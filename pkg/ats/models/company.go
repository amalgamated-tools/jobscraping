// Package models contains data models used in the ATS system.
package models

// Company represents a company with its basic information.
type Company struct {
	Name        string  `json:"name"`
	HomepageURL *string `json:"homepage_url"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}
