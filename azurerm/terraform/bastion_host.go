package terraform

import (
	"fmt"

	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// BastionHost is the entity that holds the logic to calculate price
type BastionHost struct {
	provider *Provider
	location string

	sku string

	// Usage
	monthlyOutboundDataGB decimal.Decimal
}

// bastionHostValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type bastionHostValues struct {
	Location string `mapstructure:"location"`

	SKU string `mapstructure:"sku"`

	Usage struct {
		MonthlyOutboundDataGB float64 `mapstructure:"monthly_outbound_data_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeBastionHostValues decodes and returns Values from a Terraform values map.
func decodeBastionHostValues(tfVals map[string]interface{}) (bastionHostValues, error) {
	var v bastionHostValues
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

// newBastionHost initializes a new BastionHost from the provider
func (p *Provider) newBastionHost(vals bastionHostValues) *BastionHost {
	inst := &BastionHost{
		provider: p,
		location: region.GetLocationName(vals.Location),
		sku:      "Basic",

		// From Usage
		monthlyOutboundDataGB: decimal.NewFromFloat(vals.Usage.MonthlyOutboundDataGB),
	}

	if vals.SKU != "" {
		inst.sku = vals.SKU
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *BastionHost) Components() []query.Component {
	components := []query.Component{
		inst.bastionHostComponent(inst.provider.key, inst.location, inst.sku, inst.monthlyOutboundDataGB),
		inst.bastionHostOutboundDataTransferComponent(inst.provider.key, inst.location, inst.sku, inst.monthlyOutboundDataGB),
	}
	return components
}

func (inst *BastionHost) bastionHostComponent(key string, location string, sku string, monthlyOutboundDataGB decimal.Decimal) query.Component {
	return query.Component{
		Name:           "Bastion host",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Azure Bastion"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "skuName", Value: util.StringPtr(sku)},
				{Key: "meterName", Value: util.StringPtr(fmt.Sprintf("%s Gateway", sku))},
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

func (inst *BastionHost) bastionHostOutboundDataTransferComponent(key string, location string, sku string, monthlyOutboundDataGB decimal.Decimal) query.Component {
	return query.Component{
		Name:            fmt.Sprintf("Bastion Outbound Data Transfer %s", sku),
		MonthlyQuantity: monthlyOutboundDataGB,
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Azure Bastion"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "skuName", Value: util.StringPtr(sku)},
				{Key: "meterName", Value: util.StringPtr(fmt.Sprintf("%s Data Transfer Out", sku))},
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
