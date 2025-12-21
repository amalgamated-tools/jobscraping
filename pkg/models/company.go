package models

type Company struct {
	ID          int     `db:"id"`
	Name        string  `db:"name"`
	HomepageURL *string `db:"homepage_url"`
	Description *string `db:"description"`
	LogoURL     *string `db:"logo_url"`
}
