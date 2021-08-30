package cost

import (
	"context"
	"fmt"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
)

// Backend represents a storage method used to query pricing data. It must include concrete implementations
// of all repositories.
type Backend interface {
	Product() product.Repository
	Price() price.Repository
}

// State represents a collection of all the Resource costs (either prior or planned.) It is not tied to any specific
// cloud provider or IaC tool. Instead, it is a representation of a snapshot of cloud resources at a given point
// in time, with their associated costs.
type State struct {
	Resources map[string]Resource
}

// Errors that might be returned from NewState if either a product or a price are not found.
var (
	ErrProductNotFound = fmt.Errorf("product not found")
	ErrPriceNotFound   = fmt.Errorf("price not found")
)

// NewState returns a new State from a query.Resource slice by using the Backend to fetch the pricing data.
func NewState(ctx context.Context, backend Backend, queries []query.Resource) (*State, error) {
	state := &State{Resources: make(map[string]Resource)}

	if len(queries) == 0 {
		return nil, terraform.ErrNoQueries
	}
	for _, res := range queries {
		// Mark the Resource as skipped if there are no valid Components.
		state.ensureResource(res.Address, res.Provider, res.Type, len(res.Components) == 0)

		for _, comp := range res.Components {
			prods, err := backend.Product().Filter(ctx, comp.ProductFilter)
			if err != nil {
				state.addComponent(res.Address, comp.Name, Component{Error: err})
				continue
			}
			if len(prods) < 1 {
				state.addComponent(res.Address, comp.Name, Component{Error: ErrProductNotFound})
				continue
			}
			prices, err := backend.Price().Filter(ctx, prods[0].ID, comp.PriceFilter)
			if err != nil {
				state.addComponent(res.Address, comp.Name, Component{Error: err})
				continue
			}
			if len(prices) < 1 {
				state.addComponent(res.Address, comp.Name, Component{Error: ErrPriceNotFound})
				continue
			}

			quantity := comp.MonthlyQuantity
			rate := NewMonthly(prices[0].Value, prices[0].Currency)

			if quantity.IsZero() {
				quantity = comp.HourlyQuantity
				rate = NewHourly(prices[0].Value, prices[0].Currency)
			}

			component := Component{
				Quantity: quantity,
				Unit:     comp.Unit,
				Rate:     rate,
				Details:  comp.Details,
			}

			state.addComponent(res.Address, comp.Name, component)
		}
	}

	return state, nil
}

// Cost returns the sum of the costs of every Resource included in this State.
// Error is returned if there is a mismatch in resource currencies.
func (s *State) Cost() (Cost, error) {
	var total Cost
	for name, re := range s.Resources {
		rCost, err := re.Cost()
		if err != nil {
			return Zero, fmt.Errorf("failed to get cost of resource %s: %w", name, err)
		}
		total, err = total.Add(rCost)
		if err != nil {
			return Zero, fmt.Errorf("failed to add cost of resource %s: %w", name, err)
		}
	}

	return total, nil
}

// ensureResource creates Resource at the given address if it doesn't already exist.
func (s *State) ensureResource(address, provider, typ string, skipped bool) {
	if _, ok := s.Resources[address]; !ok {
		res := Resource{
			Provider: provider,
			Type:     typ,
			Skipped:  skipped,
		}

		if !skipped {
			res.Components = make(map[string]Component)
		}

		s.Resources[address] = res
	}
}

// addComponent adds the Component with given label to the Resource at given address.
func (s *State) addComponent(resAddress, compLabel string, component Component) {
	s.Resources[resAddress].Components[compLabel] = component
}
