package query

import (
	"github.com/shopspring/decimal"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
)

// Resource represents a single cloud resource. It has a unique Address and a collection of multiple
// Component queries.
type Resource struct {
	// Address uniquely identifies this cloud Resource.
	Address string

	// Components is a list of price components that make up this Resource.
	Components []Component
}

// Component represents a price component of a cloud Resource. It is used to fetch the price for a single
// component of a resource. For example, a compute instance might be have different pricing for the number
// of CPU's, amount of RAM, etc. - each of these would be a Component.
type Component struct {
	Name            string
	HourlyQuantity  decimal.Decimal
	MonthlyQuantity decimal.Decimal
	Unit            string
	Details         []string
	ProductFilter   *product.Filter
	PriceFilter     *price.Filter
}
