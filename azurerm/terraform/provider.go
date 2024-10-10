package terraform

import (
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
)

// Provider is an implementation of the terraform.Provider, used to extract component queries from
// terraform resources.
type Provider struct {
	key string
}

// NewProvider initializes a new Google provider with key and region
func NewProvider(key string) (*Provider, error) {
	return &Provider{
		key: key,
	}, nil
}

// Name returns the Provider's common name.
func (p *Provider) Name() string { return p.key }

// ResourceComponents returns Component queries for a given terraform.Resource.
func (p *Provider) ResourceComponents(rss map[string]terraform.Resource, tfRes terraform.Resource) []query.Component {
	switch tfRes.Type {
	case "azurerm_linux_virtual_machine":
		vals, err := decodeLinuxVirtualMachineValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newLinuxVirtualMachine(vals).Components()
	case "azurerm_virtual_machine":
		vals, err := decodeVirtualMachineValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newVirtualMachine(vals).Components()
	case "azurerm_virtual_network_gateway":
		vals, err := decodeVirtualNetworkGatewayValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newVirtualNetworkGateway(vals).Components()
	default:
		return nil
	}
}
