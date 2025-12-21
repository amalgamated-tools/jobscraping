package helpers

import (
	"regexp"
	"strconv"
	"strings"
)

type Compensation struct {
	Currency     string
	MinSalary    *int
	MaxSalary    *int
	OffersEquity bool
	Parsed       bool
}

var compRegex = regexp.MustCompile(`(?i)([A-Z]*\$|€|£)?\s*\$?([\d.,]+)(K)?(?:\s*[–-]\s*([A-Z]*\$|€|£)?\s*\$?([\d.,]+)(K)?)?`)

func ParseCompensation(s string) Compensation {
	match := compRegex.FindStringSubmatch(s)

	result := Compensation{
		Currency:     "",
		MinSalary:    nil,
		MaxSalary:    nil,
		OffersEquity: strings.Contains(strings.ToLower(s), "equity"),
		Parsed:       false,
	}

	if match == nil {
		return result
	}
	result.Parsed = true

	parseAmount := func(val string, hasK string) *int {
		if val == "" {
			return nil
		}
		val = strings.ReplaceAll(val, ",", "")
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil
		}
		if strings.EqualFold(hasK, "K") {
			num *= 1000
		}
		n := int(num)
		return &n
	}

	currency := match[1]
	if currency == "" {
		currency = match[4]
	}

	minSalary := parseAmount(match[2], match[3])
	maxSalary := parseAmount(match[5], match[6])
	if maxSalary == nil {
		maxSalary = minSalary
	}

	result.Currency = strings.TrimSpace(currency)
	result.MinSalary = minSalary
	result.MaxSalary = maxSalary
	return result
}
