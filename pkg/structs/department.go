package structs

import "strings"

type Department int64

const (
	AI Department = iota
	CustomerSuccessSupport
	Data
	Design
	Marketing
	ProductManagement
	Sales
	Security
	SoftwareEngineering
	Unsure
)

func DepartmentNames() []string {
	return []string{
		"AI",
		"Customer Success and Support",
		"Data",
		"Design",
		"Marketing",
		"Product Management",
		"Sales",
		"Security",
		"Software Engineering",
	}
}

func (d Department) String() string {
	return [...]string{
		"Unsure",
		"AI",
		"Customer Success and Support",
		"Data",
		"Design",
		"Marketing",
		"Product Management",
		"Sales",
		"Security",
		"Software Engineering",
	}[d]
}

func ParseDepartment(dept string) Department {
	switch strings.ToLower(dept) {
	case "ai":
		return AI
	case "corporate it", "corporate":
		return SoftwareEngineering
	case "customer success", "customer support", "customer success & support":
		return CustomerSuccessSupport
	case "community":
		return CustomerSuccessSupport
	case "data", "data science", "data engineering":
		return Data
	case "design", "ux", "ui", "product design":
		return Design
	case "hardware", "hardware engineering":
		return SoftwareEngineering
	case "marketing", "growth":
		return Marketing
	case "product management", "product":
		return ProductManagement
	case "sales", "business development":
		return Sales
	case "security", "information security", "infosec":
		return Security
	case "software engineering", "engineering", "dev", "development":
		return SoftwareEngineering
	default:
		return Unsure // Default to Unsure if unknown
	}
}
