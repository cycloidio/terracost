package terraform

import (
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// LinuxVirtualMachine is the entity that holds the logic to calculate price
// of the google_compute_instance
type LinuxVirtualMachine struct {
	provider *Provider

	location string
	size     string
}

// linuxVirtualMachineValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type linuxVirtualMachineValues struct {
	Size     string `mapstructure:"size"`
	Location string `mapstructure:"location"`
}

// decodeLinuxVirtualMachineValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeLinuxVirtualMachineValues(tfVals map[string]interface{}) (linuxVirtualMachineValues, error) {
	var v linuxVirtualMachineValues
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

// newLinuxVirtualMachine initializes a new LinuxVirtualMachine from the provider
func (p *Provider) newLinuxVirtualMachine(vals linuxVirtualMachineValues) *LinuxVirtualMachine {
	inst := &LinuxVirtualMachine{
		provider: p,

		location: getLocationName(vals.Location),
		size:     vals.Size,
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *LinuxVirtualMachine) Components() []query.Component {
	components := []query.Component{inst.linuxVirtualMachineComponent()}

	return components
}

// linuxVirtualMachineComponent returns the query needed to be able to calculate the price
func (inst *LinuxVirtualMachine) linuxVirtualMachineComponent() query.Component {
	return linuxVirtualMachineComponent(inst.provider.key, inst.location, inst.size)
}

// linuxVirtualMachineComponent is the abstraction of the same LinuxVirtualMachine.linuxVirtualMachineComponent
// so it can be reused
func linuxVirtualMachineComponent(key, location, size string) query.Component {
	return query.Component{
		Name:           "Compute",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Virtual Machines"),
			Family:   util.StringPtr("Compute"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "arm_sku_name", Value: util.StringPtr(size)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 Hour"),
		},
	}
}
