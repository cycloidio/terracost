package cost

import (
	"github.com/shopspring/decimal"
)

// HoursPerMonth is an approximate number of hours in a month.
// It is calculated as 365 days in a year x 24 hours in a day / 12 months in year.
var HoursPerMonth = decimal.NewFromInt(730)

// Cost represents a monthly or hourly cost of a cloud resource or its component.
type Cost struct {
	// Decimal is price per month.
	decimal.Decimal
}

// Zero is Cost with zero value.
var Zero = Cost{}

// NewMonthly returns a new Cost from price per month.
func NewMonthly(monthly decimal.Decimal) Cost {
	return Cost{Decimal: monthly}
}

// NewHourly returns a new Cost from price per hour.
func NewHourly(hourly decimal.Decimal) Cost {
	return Cost{Decimal: hourly.Mul(HoursPerMonth)}
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
func (c Cost) Add(c2 Cost) Cost {
	return Cost{Decimal: c.Decimal.Add(c2.Monthly())}
}

// MulDecimal multiplies the Cost by the given decimal.Decimal.
func (c Cost) MulDecimal(d decimal.Decimal) Cost {
	return Cost{Decimal: c.Decimal.Mul(d)}
}
