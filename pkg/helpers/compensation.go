// Package helpers provides utility functions for parsing and handling common data structures.
package helpers

import (
	"regexp"
	"strconv"
	"strings"
)

// Compensation represents parsed compensation details from a job listing.
type Compensation struct {
	Currency     string
	MinSalary    float64
	MaxSalary    float64
	OffersEquity bool
	Parsed       bool
}

var compRegex = regexp.MustCompile(`(?i)([A-Z]*\$|€|£)?\s*\$?([\d.,]+)(K)?(?:\s*[–-]\s*([A-Z]*\$|€|£)?\s*\$?([\d.,]+)(K)?)?`)

// ParseCompensation parses a compensation string and extracts currency, minimum and maximum salary, and equity information.
func ParseCompensation(compensationString string) Compensation {
	match := compRegex.FindStringSubmatch(compensationString)

	result := Compensation{
		Currency:     "",
		MinSalary:    0,
		MaxSalary:    0,
		OffersEquity: strings.Contains(strings.ToLower(compensationString), "equity"),
		Parsed:       false,
	}

	if match == nil {
		return result
	}

	result.Parsed = true

	parseAmount := func(val string, hasK string) float64 {
		if val == "" {
			return 0
		}

		val = strings.ReplaceAll(val, ",", "")

		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}

		if strings.EqualFold(hasK, "K") {
			num *= 1000
		}

		return num
	}

	currency := match[1]
	if currency == "" {
		currency = match[4]
	}

	minSalary := parseAmount(match[2], match[3])

	maxSalary := parseAmount(match[5], match[6])
	if maxSalary == 0 {
		maxSalary = minSalary
	}

	result.Currency = strings.TrimSpace(currency)
	result.MinSalary = minSalary
	result.MaxSalary = maxSalary

	return result
}
