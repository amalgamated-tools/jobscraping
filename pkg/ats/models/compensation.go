package models

import (
	"log/slog"
	"regexp"
	"strconv"
	"strings"
)

// Compensation represents salary information including currency, minimum and maximum salary, equity offering, and parsing status.
type Compensation struct {
	Currency     string
	MinSalary    float64
	MaxSalary    float64
	OffersEquity bool
	Parsed       bool
}

var compRegex = regexp.MustCompile(`(?i)([A-Z]*\$|€|£)?\s*\$?([\d.,]+)(K)?(?:\s*[–-]\s*([A-Z]*\$|€|£)?\s*\$?([\d.,]+)(K)?)?`)

// ParseCompensation parses a Compensation from a string representation.
func ParseCompensation(s string) Compensation {
	match := compRegex.FindStringSubmatch(s)

	result := Compensation{
		Currency:     "",
		MinSalary:    0,
		MaxSalary:    0,
		OffersEquity: strings.Contains(strings.ToLower(s), "equity"),
		Parsed:       false,
	}

	if match == nil {
		slog.Debug("compensation string did not match regex", slog.String("input", s))
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
	result.MinSalary = float64(minSalary)
	result.MaxSalary = float64(maxSalary)

	return result
}
