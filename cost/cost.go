package cost

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// HoursPerMonth is an approximate number of hours in a month.
// It is calculated as 365 days in a year x 24 hours in a day / 12 months in year.
var HoursPerMonth = decimal.NewFromInt(730)

// Cost represents a monthly or hourly cost of a cloud resource or its component.
type Cost struct {
	// Decimal is price per month.
	decimal.Decimal
	// Currency of the cost.
	Currency string
}

// Zero is Cost with zero value.
var Zero = Cost{}

// NewMonthly returns a new Cost from price per month with currency.
func NewMonthly(monthly decimal.Decimal, currency string) Cost {
	return Cost{Decimal: monthly, Currency: currency}
}

// NewHourly returns a new Cost from price per hour with currency.
func NewHourly(hourly decimal.Decimal, currency string) Cost {
	return Cost{Decimal: hourly.Mul(HoursPerMonth), Currency: currency}
}

// Monthly returns the cost per month.
func (c Cost) Monthly() decimal.Decimal {
	return c.Decimal
}

// Hourly returns the cost per hour.
func (c Cost) Hourly() decimal.Decimal {
	return c.DivRound(HoursPerMonth, 6)
}

// Add adds the values of two Cost structs.
// If the currency of both costs doesn't match, error is returned.
func (c Cost) Add(c2 Cost) (Cost, error) {
	// if cost addition iz Zero, ignore it
	if c2 == Zero {
		return c, nil
	}

	// If there is no currency, use the currency of the addition
	if c.Currency == "" {
		c.Currency = c2.Currency
	}

	if c.Currency != c2.Currency {
		return Zero, fmt.Errorf("currency mismatch: expected %s, got %s", c.Currency, c2.Currency)
	}

	return Cost{Decimal: c.Decimal.Add(c2.Monthly()), Currency: c.Currency}, nil
}

// MulDecimal multiplies the Cost by the given decimal.Decimal.
func (c Cost) MulDecimal(d decimal.Decimal) Cost {
	return Cost{Decimal: c.Decimal.Mul(d), Currency: c.Currency}
}
