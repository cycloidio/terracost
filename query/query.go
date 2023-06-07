package query

import (
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
)

// Resource represents a single cloud resource. It has a unique Address and a collection of multiple
// Component queries.
type Resource struct {
	// Address uniquely identifies this cloud Resource.
	Address string

	// Provider is the cloud provider that this Resource belongs to.
	Provider string

	// Type describes the type of the Resource.
	Type string

	// Components is a list of price components that make up this Resource. If it is empty, the resource
	// is considered to be skipped.
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
	Usage           bool
	ProductFilter   *product.Filter
	PriceFilter     *price.Filter
}
