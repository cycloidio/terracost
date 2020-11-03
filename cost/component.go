package cost

import (
	"github.com/shopspring/decimal"
)

// Component describes the pricing of a single resource cost component. This includes Rate and Quantity
// and allows for final cost computation.
type Component struct {
	Quantity decimal.Decimal
	Unit     string
	Rate     decimal.Decimal
	Details  []string
}

// Cost returns the cost of this component (Rate multiplied by Quantity).
func (c Component) Cost() decimal.Decimal {
	return c.Rate.Mul(c.Quantity)
}

// ComponentDiff is a difference between the Prior and Planned Component.
type ComponentDiff struct {
	Prior, Planned *Component
}

// PriorCost returns the full cost of the Prior Component or decimal.Zero if it doesn't exist.
func (cd ComponentDiff) PriorCost() decimal.Decimal {
	if cd.Prior == nil {
		return decimal.Zero
	}
	return cd.Prior.Cost()
}

// PlannedCost returns the full cost of the Planned Component or decimal.Zero if it doesn't exist.
func (cd ComponentDiff) PlannedCost() decimal.Decimal {
	if cd.Planned == nil {
		return decimal.Zero
	}
	return cd.Planned.Cost()
}
