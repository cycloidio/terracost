package terraform

import (
	"fmt"
	"strings"

	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// virtualNetworkGatewayConnectionValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type virtualNetworkGatewayConnectionValues struct {
	VirtualNetworkGatewayID string `mapstructure:"virtual_network_gateway_id"`
	Location                string `mapstructure:"location"`
	Type                    string `mapstructure:"type"`
}

// decodeVirtualNetworkGatewayConnectionValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeVirtualNetworkGatewayConnectionValues(tfVals map[string]interface{}) (virtualNetworkGatewayConnectionValues, error) {
	var v virtualNetworkGatewayConnectionValues
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

// newVirtualNetworkGatewayConnection initializes a new VirtualNetworkGatewayConnection from the provider
func (p *Provider) newVirtualNetworkGatewayConnection(rss map[string]terraform.Resource, vals virtualNetworkGatewayConnectionValues) *VirtualNetworkGateway {
	inst := &VirtualNetworkGateway{
		provider: p,
		location: region.GetLocationName(vals.Location),
		sku:      "Basic",
		gwType:   vals.Type,
	}

	vng, err := decodeVirtualNetworkGatewayValues(rss[vals.VirtualNetworkGatewayID].Values)
	if err != nil {
		return inst
	}

	if strings.ToLower(vng.SKU) != "" {
		inst.sku = vng.SKU
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *VirtualNetworkGateway) connectionComponent() []query.Component {

	if strings.ToLower(inst.gwType) != "ipsec" {
		return []query.Component{}
	}

	components := []query.Component{
		inst.virtualNetworkGatewayConnectionS2SComponent(inst.provider.key, inst.location, inst.sku, inst.gwType),
	}

	return components
}

func (inst *VirtualNetworkGateway) virtualNetworkGatewayConnectionS2SComponent(key, location, sku string, gwType string) query.Component {
	return query.Component{
		Name:           fmt.Sprintf("VPN gateway Connection S2S (%s-%s)", sku, gwType),
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("VPN Gateway"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "skuName", Value: util.StringPtr(sku)},
				{Key: "meterName", Value: util.StringPtr("S2S Connection")},
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
