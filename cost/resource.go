package cost

import (
	"github.com/shopspring/decimal"
)

// Resource represents costs of a single cloud resource. Each Resource includes a Component map, keyed
// by the label.
type Resource struct {
	Components map[string]Component
	Skipped    bool
}

// Cost returns the sum of costs of every Component of this Resource.
func (re Resource) Cost() decimal.Decimal {
	var total decimal.Decimal
	for _, comp := range re.Components {
		total = total.Add(comp.Cost())
	}
	return total
}

// ResourceDiff is the difference in costs between prior and planned Resource. It contains a ComponentDiff
// map, keyed by the label.
type ResourceDiff struct {
	Address        string
	ComponentDiffs map[string]*ComponentDiff
}

// PriorCost returns the sum of costs of every Component's PriorCost.
func (rd ResourceDiff) PriorCost() decimal.Decimal {
	var total decimal.Decimal
	for _, cd := range rd.ComponentDiffs {
		total = total.Add(cd.PriorCost())
	}
	return total
}

// PlannedCost returns the sum of costs of every Component's PlannedCost.
func (rd ResourceDiff) PlannedCost() decimal.Decimal {
	var total decimal.Decimal
	for _, cd := range rd.ComponentDiffs {
		total = total.Add(cd.PlannedCost())
	}
	return total
}

// Errors returns a map of Component errors keyed by the Component label.
func (rd ResourceDiff) Errors() map[string]error {
	errs := make(map[string]error)
	for label, cd := range rd.ComponentDiffs {
		if cd.Prior != nil && cd.Prior.Error != nil {
			errs[label] = cd.Prior.Error
		} else if cd.Planned != nil && cd.Planned.Error != nil {
			errs[label] = cd.Planned.Error
		}
	}
	return errs
}

// Valid returns true if there are no errors in all of the ResourceDiff's components.
func (rd ResourceDiff) Valid() bool {
	for _, cd := range rd.ComponentDiffs {
		if (cd.Prior != nil && cd.Prior.Error != nil) || (cd.Planned != nil && cd.Planned.Error != nil) {
			return false
		}
	}
	return true
}
