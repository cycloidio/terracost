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

// NatGateway is the entity that holds the logic to calculate price
type NatGateway struct {
	provider *Provider

	location string
	skuName  string

	// Usage
	monthlyDataProcessedGB decimal.Decimal
}

// natGatewayValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type natGatewayValues struct {
	Location string `mapstructure:"location"`
	SkuName  string `mapstructure:"sku_name"`

	Usage struct {
		MonthlyDataProcessedGB float64 `mapstructure:"monthly_data_processed_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeNatGatewayValues decodes and returns Values from a Terraform values map.
func decodeNatGatewayValues(tfVals map[string]interface{}) (natGatewayValues, error) {
	var v natGatewayValues
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

// newNatGateway initializes a new NatGateway from the provider
func (p *Provider) newNatGateway(vals natGatewayValues) *NatGateway {
	inst := &NatGateway{
		provider: p,

		location: region.GetLocationName(vals.Location),
		skuName:  "Standard",
		// From Usage
		monthlyDataProcessedGB: decimal.NewFromFloat(vals.Usage.MonthlyDataProcessedGB),
	}

	if vals.SkuName != "" {
		inst.skuName = vals.SkuName
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *NatGateway) Components() []query.Component {
	components := []query.Component{
		inst.natGatewayComponent(inst.provider.key, inst.location, inst.skuName),
		inst.natGatewayDataProcessedComponent(inst.provider.key, inst.location, inst.skuName, inst.monthlyDataProcessedGB),
	}

	return components
}

func (inst *NatGateway) natGatewayComponent(key string, location string, skuName string) query.Component {
	return query.Component{
		Name:           "NAT Gateway",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("NAT Gateway"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr("Global"),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "meterName", Value: util.StringPtr(fmt.Sprintf("%s Gateway", skuName))},
				{Key: "skuName", Value: util.StringPtr(skuName)},
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

func (inst *NatGateway) natGatewayDataProcessedComponent(key string, location string, skuName string, monthlyDataProcessedGB decimal.Decimal) query.Component {
	return query.Component{
		Name:            "NAT Gateway Data Processed",
		MonthlyQuantity: monthlyDataProcessedGB,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("NAT Gateway"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr("Global"),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "meterName", Value: util.StringPtr(fmt.Sprintf("%s Data Processed", skuName))},
				{Key: "skuName", Value: util.StringPtr(skuName)},
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
