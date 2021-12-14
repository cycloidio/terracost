package terraform

import (
	"github.com/cycloidio/terracost/query"
	"github.com/mitchellh/mapstructure"
)

// VirtualMachine is the entity that holds the logic to calculate price
// of the google_compute_instance
type VirtualMachine struct {
	provider *Provider

	location string
	vmSize   string
}

// virtualMachineValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type virtualMachineValues struct {
	VMSize   string `mapstructure:"vm_size"`
	Location string `mapstructure:"location"`
}

// decodeVirtualMachineValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeVirtualMachineValues(tfVals map[string]interface{}) (virtualMachineValues, error) {
	var v virtualMachineValues
	if err := mapstructure.Decode(tfVals, &v); err != nil {
		return v, err
	}
	return v, nil
}

// newVirtualMachine initializes a new VirtualMachine from the provider
func (p *Provider) newVirtualMachine(vals virtualMachineValues) *VirtualMachine {
	inst := &VirtualMachine{
		provider: p,

		location: getLocationName(vals.Location),
		vmSize:   vals.VMSize,
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *VirtualMachine) Components() []query.Component {
	components := []query.Component{linuxVirtualMachineComponent(inst.provider.key, inst.location, inst.vmSize)}

	return components
}
