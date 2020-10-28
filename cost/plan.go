package cost

import (
	"github.com/shopspring/decimal"
)

// Plan is the cost difference between two State instances. It is not tied to any specific cloud provider or IaC tool.
// Instead, it is a representation of the differences between two snapshots of cloud resources, with their associated
// costs. The Plan instance can be used to calculate the total cost difference of a plan, as well as cost differences
// of each resource (and their components) separately.
type Plan struct {
	Prior, Planned *State
}

// NewPlan returns a new Plan from Prior and Planned State.
func NewPlan(prior, planned *State) *Plan {
	return &Plan{Prior: prior, Planned: planned}
}

// PriorCost returns the total cost of the Prior State or decimal.Zero if it isn't included in the plan.
func (p Plan) PriorCost() decimal.Decimal {
	if p.Prior == nil {
		return decimal.Zero
	}
	return p.Prior.Cost()
}

// PlannedCost returns the total cost of the Planned State or decimal.Zero if it isn't included in the plan.
func (p Plan) PlannedCost() decimal.Decimal {
	if p.Planned == nil {
		return decimal.Zero
	}
	return p.Planned.Cost()
}

// ResourceDifferences merges the Prior and Planned State and returns a slice of differences between resources.
// The order of the elements in the slice is undefined and unstable.
func (p Plan) ResourceDifferences() []ResourceDiff {
	rdmap := make(map[string]ResourceDiff)

	if p.Prior != nil {
		mergeResourceDiffsFromState(rdmap, p.Prior, false)
	}
	if p.Planned != nil {
		mergeResourceDiffsFromState(rdmap, p.Planned, true)
	}

	rds := make([]ResourceDiff, 0, len(rdmap))
	for _, rd := range rdmap {
		rds = append(rds, rd)
	}
	return rds
}

// mergeResourceDiffsFromState adds all the resources from the State to the provided ResourceDiff map. Each component
// of every resource is then placed into an appropriate ComponentDiff field based on the value of the `planned` argument.
func mergeResourceDiffsFromState(rdmap map[string]ResourceDiff, state *State, planned bool) {
	for address, res := range state.Resources {
		if _, ok := rdmap[address]; !ok {
			rdmap[address] = ResourceDiff{
				Address:        address,
				ComponentDiffs: make(map[string]*ComponentDiff),
			}
		}

		for label, comp := range res.Components {
			comp := comp

			cd, ok := rdmap[address].ComponentDiffs[label]
			if !ok {
				cd = &ComponentDiff{}
				rdmap[address].ComponentDiffs[label] = cd
			}

			if planned {
				cd.Planned = &comp
			} else {
				cd.Prior = &comp
			}
		}
	}
}
