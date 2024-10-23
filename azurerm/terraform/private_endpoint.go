package terraform

import (
	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

//	This resource corresponds in the billing API as a single meterName within the productName Virtual Network Private Link
// To check it you can use: curl -s "https://prices.azure.com/api/retail/prices?\$filter=productName eq 'Virtual Network Private Link'" | jq '.Items[] | {skuName, meterName}' | sort -u

// PrivateEndpoint is the entity that holds the logic to calculate price
// of the azurerm_private_endpoint
type PrivateEndpoint struct {
	provider *Provider

	location string

	// Usage
	monthlyHours decimal.Decimal
}

// privateEndpointValues is holds the terraform values that we need to estimate the price
type privateEndpointValues struct {

	//required params
	Location string `mapstructure:"location"`

	// usage - with default values
	Usage struct {
		MonthlyHours int64 `mapstructure:"monthly_hours"`
	} `mapstructure:"tc_usage"`
}

// decodePrivateEndpointValues decodes and returns computeInstanceValues from a Terraform values map.
func decodePrivateEndpointValues(tfVals map[string]interface{}) (privateEndpointValues, error) {
	var v privateEndpointValues
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

// newPrivateEndpoint initializes a new PrivateEndpoint from the provider
func (p *Provider) newPrivateEndpoint(vals privateEndpointValues) *PrivateEndpoint {
	inst := &PrivateEndpoint{
		provider: p,

		location: region.GetLocationName(vals.Location),
		// From Usage
		monthlyHours: decimal.NewFromInt(vals.Usage.MonthlyHours),
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *PrivateEndpoint) Components() []query.Component {

	return []query.Component{
		inst.privateEndpointComponent(inst.provider.key, "Global", inst.monthlyHours),
	}
}

func (inst *PrivateEndpoint) privateEndpointComponent(key, location string, monthlyHours decimal.Decimal) query.Component {
	return query.Component{
		Name:            "Private Endpoint",
		MonthlyQuantity: monthlyHours,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Virtual Network"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr("Virtual Network Private Link")},
				{Key: "meterName", Value: util.StringPtr("Standard Private Endpoint")},
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
