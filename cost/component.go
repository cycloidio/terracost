package cost

import (
	"github.com/shopspring/decimal"
)

// Component describes the pricing of a single resource cost component. This includes Rate and Quantity
// and allows for final cost computation.
type Component struct {
	Quantity decimal.Decimal
	Unit     string
	Rate     Cost
	Details  []string
	Usage    bool

	Error error
}

// Cost returns the cost of this component (Rate multiplied by Quantity).
func (c Component) Cost() Cost {
	if c.Rate.IsZero() || c.Quantity.IsZero() {
		return Zero
	}
	return c.Rate.MulDecimal(c.Quantity)
}

// ComponentDiff is a difference between the Prior and Planned Component.
type ComponentDiff struct {
	Prior, Planned *Component
}

// PriorCost returns the full cost of the Prior Component or decimal.Zero if it doesn't exist.
func (cd ComponentDiff) PriorCost() Cost {
	if cd.Prior == nil {
		return Zero
	}
	return cd.Prior.Cost()
}

// PlannedCost returns the full cost of the Planned Component or decimal.Zero if it doesn't exist.
func (cd ComponentDiff) PlannedCost() Cost {
	if cd.Planned == nil {
		return Zero
	}
	return cd.Planned.Cost()
}

// Valid returns true if there are no errors in both the Planned and Prior components.
func (cd ComponentDiff) Valid() bool {
	return !((cd.Prior != nil && cd.Prior.Error != nil) || (cd.Planned != nil && cd.Planned.Error != nil))
}
