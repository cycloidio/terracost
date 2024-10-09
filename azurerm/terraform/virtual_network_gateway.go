package terraform

import (
	"fmt"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// VirtualNetworkGateway is the entity that holds the logic to calculate price
// of the google_compute_instance
type VirtualNetworkGateway struct {
	provider *Provider

	location  string
	sku       string
	meterName string
}

// virtualNetworkGatewayValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type virtualNetworkGatewayValues struct {
	SKU      string `mapstructure:"sku"`
	Location string `mapstructure:"location"`
}

// decodeVirtualNetworkGatewayValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeVirtualNetworkGatewayValues(tfVals map[string]interface{}) (virtualNetworkGatewayValues, error) {
	var v virtualNetworkGatewayValues
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

// newVirtualNetworkGateway initializes a new VirtualNetworkGateway from the provider
func (p *Provider) newVirtualNetworkGateway(vals virtualNetworkGatewayValues) *VirtualNetworkGateway {
	inst := &VirtualNetworkGateway{
		provider: p,

		location:  getLocationName(vals.Location),
		sku:       vals.SKU,
		meterName: vals.SKU,
	}

	if vals.SKU == "Basic" {
		inst.meterName = "Basic Gateway"
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *VirtualNetworkGateway) Components() []query.Component {
	components := []query.Component{inst.virtualNetworkGatewayComponent(inst.provider.key, inst.location, inst.sku, inst.meterName)}

	return components
}

func (inst *VirtualNetworkGateway) virtualNetworkGatewayComponent(key, location, sku string, meterName string) query.Component {
	return query.Component{
		Name:           fmt.Sprintf("VPN gateway (%s)", sku),
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("VPN Gateway"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "meter_name", Value: util.StringPtr(meterName)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 Hour"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
