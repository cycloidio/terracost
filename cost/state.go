package cost

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
)

// Backend represents a storage method used to query pricing data. It must include concrete implementations
// of all repositories.
type Backend interface {
	Product() product.Repository
	Price() price.Repository
}

// State represents a collection of all the Resource costs (either prior or planned.)
type State struct {
	Resources map[string]Resource
}

// NewState returns a new State from a query.Resource slice by using the Backend to fetch the pricing data.
func NewState(ctx context.Context, backend Backend, queries []query.Resource) (*State, error) {
	state := &State{Resources: make(map[string]Resource)}

	for _, res := range queries {
		for _, comp := range res.Components {
			prods, err := backend.Product().Filter(ctx, comp.ProductFilter)
			if err != nil {
				return nil, err
			}
			if len(prods) < 1 {
				return nil, fmt.Errorf("product not found")
			}
			prices, err := backend.Price().Filter(ctx, prods[0].ID, comp.PriceFilter)
			if err != nil {
				return nil, err
			}
			if len(prices) < 1 {
				return nil, fmt.Errorf("price not found")
			}

			quantity := comp.MonthlyQuantity
			rate := prices[0].Value
			if quantity.Equals(decimal.Zero) {
				quantity = comp.HourlyQuantity
				rate = rate.Mul(decimal.NewFromInt(730))
			}

			component := Component{
				Quantity: quantity,
				Unit:     prices[0].Unit,
				Rate:     rate,
				Details:  comp.Details,
			}

			state.addComponent(res.Address, comp.Name, component)
		}
	}

	return state, nil
}

// Cost returns the sum of the costs of every Resource included in this State.
func (s *State) Cost() decimal.Decimal {
	total := decimal.Zero
	for _, re := range s.Resources {
		total = total.Add(re.Cost())
	}
	return total
}

func (s *State) addComponent(resAddress, compLabel string, component Component) {
	if _, ok := s.Resources[resAddress]; !ok {
		s.Resources[resAddress] = Resource{Components: make(map[string]Component)}
	}

	s.Resources[resAddress].Components[compLabel] = component
}
