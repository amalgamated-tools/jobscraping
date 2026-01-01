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
	// UnknownDepartment represents an unknown or unspecified department.
	UnknownDepartment
)

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
		slog.Debug("Unknown department encountered", slog.String("department", dept))
		return UnknownDepartment // Default to Unsure if unknown
	}
}

func (d Department) String() string {
	switch d {
	case AI:
		return "ai"
	case CustomerSuccessSupport:
		return "customer_success_support"
	case Data:
		return "data"
	case Design:
		return "design"
	case Marketing:
		return "marketing"
	case ProductManagement:
		return "product_management"
	case Sales:
		return "sales"
	case Security:
		return "security"
	case SoftwareEngineering:
		return "software_engineering"
	case UnknownDepartment:
		return "unknown"
	default:
		return "unknown"
	}
}
