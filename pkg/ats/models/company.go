package models

type Company struct {
	Name        string  `json:"name"`
	HomepageURL *string `json:"homepage_url"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}
