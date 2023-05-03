package terraform

import (
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// ComputeInstance is the entity that holds the logic to calculate price
// of the google_compute_instance
type ComputeInstance struct {
	provider     *Provider
	region       string
	instanceType string

	machineType string
}

// computeInstanceValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type computeInstanceValues struct {
	MachineType string `mapstructure:"machine_type"`
}

// decodeComputeInstanceValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeComputeInstanceValues(tfVals map[string]interface{}) (computeInstanceValues, error) {
	var v computeInstanceValues
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &v,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return v, err
	}

	if err := decoder.Decode(tfVals); err != nil {
		return v, err
	}
	return v, nil
}

// newComputeInstance initializes a new ComputeInstance from the provider
func (p *Provider) newComputeInstance(vals computeInstanceValues) *ComputeInstance {
	inst := &ComputeInstance{
		provider: p,
		region:   p.region,

		machineType: vals.MachineType,
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *ComputeInstance) Components() []query.Component {
	components := []query.Component{inst.computeComponent()}

	return components
}

// computeComponent returns the query needed to be able to calculate the price
func (inst *ComputeInstance) computeComponent() query.Component {
	return query.Component{
		Name:           "Compute",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(inst.provider.key),
			Service:  util.StringPtr("Compute Engine"),
			Family:   util.StringPtr("Compute"),
			Location: util.StringPtr(inst.region),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "machine_type", Value: util.StringPtr(inst.machineType)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("h"),
		},
	}
}
