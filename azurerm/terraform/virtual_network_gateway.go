package terraform

import (
	"fmt"
	"strings"

	"github.com/cycloidio/terracost/azurerm/region"
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
	meterName string
	sku       string
	gwType    string
	// Usage
	monthlyDataTransferGB decimal.Decimal
}

// virtualNetworkGatewayValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type virtualNetworkGatewayValues struct {
	SKU      string `mapstructure:"sku"`
	Location string `mapstructure:"location"`
	Type     string `mapstructure:"type"`

	Usage struct {
		MonthlyDataTransferGB float64 `mapstructure:"monthly_data_transfer_gb"`
	} `mapstructure:"tc_usage"`
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

		location:  region.GetLocationName(vals.Location),
		meterName: vals.SKU,
		sku:       vals.SKU,
		gwType:    vals.Type,
		// From Usage
		monthlyDataTransferGB: decimal.NewFromFloat(vals.Usage.MonthlyDataTransferGB),
	}

	if strings.ToLower(vals.SKU) == "basic" {
		inst.meterName = "Basic Gateway"
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *VirtualNetworkGateway) Components() []query.Component {
	components := []query.Component{
		inst.virtualNetworkGatewayComponent(inst.provider.key, inst.location, inst.sku, inst.meterName),
		inst.virtualNetworkGatewayP2SComponent(inst.provider.key, inst.location, inst.sku),
		inst.virtualNetworkGatewayDataTransfersComponent(inst.provider.key, inst.location),
	}

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
				{Key: "meterName", Value: util.StringPtr(meterName)},
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

func (inst *VirtualNetworkGateway) virtualNetworkGatewayP2SComponent(key, location, sku string) query.Component {
	return query.Component{
		Name:           "VPN gateway P2S tunnels (over 128)",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("VPN Gateway"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "skuName", Value: util.StringPtr(sku)},
				{Key: "meterName", Value: util.StringPtr("P2S Connection")},
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

func (inst *VirtualNetworkGateway) virtualNetworkGatewayDataTransfersComponent(key string, location string) query.Component {
	return query.Component{
		Name:            "VPN gateway data tranfer",
		MonthlyQuantity: inst.monthlyDataTransferGB,
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("VPN Gateway"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(region.GetRegionToVNETZone(location)),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr("VPN Gateway Bandwidth")},
				{Key: "meterName", Value: util.StringPtr("Standard Inter-Virtual Network Data Transfer Out")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 GB"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
