// Package models contains data models used in the ATS system.
package models

import "net/url"

// Company represents a company with its basic information.
type Company struct {
	Name        string  `json:"name"`
	Homepage    url.URL `json:"homepage"`
	Description *string `json:"description"`
	Logo        url.URL `json:"logo"`
}

// NewCompany creates a new Company instance.
func NewCompany() *Company {
	return &Company{}
}
