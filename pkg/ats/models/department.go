package models

import (
	"log/slog"
	"strings"
)

// Department represents various departments within a company.
type Department int64

const (
	// AI represents the Artificial Intelligence department.
	AI Department = iota
	// CustomerSuccessSupport represents the Customer Success and Support department.
	CustomerSuccessSupport
	// Data represents the Data department.
	Data
	// Design represents the Design department.
	Design
	// Marketing represents the Marketing department.
	Marketing
	// ProductManagement represents the Product Management department.
	ProductManagement
	// Sales represents the Sales department.
	Sales
	// Security represents the Security department.
	Security
	// SoftwareEngineering represents the Software Engineering department.
	SoftwareEngineering
	// Unsure represents an unknown or unspecified department.
	Unsure
)

// DepartmentNames returns a slice of all department names.
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

// String returns the string representation of the Department.
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

// ParseDepartment converts a string representation of a department to its corresponding Department constant.
func ParseDepartment(dept string) Department { //nolint:cyclop
	switch strings.ToLower(strings.TrimSpace(dept)) {
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
		slog.Warn("Unknown department encountered", slog.String("department", dept))
		return Unsure // Default to Unsure if unknown
	}
}
