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

//To check the available meterName and SkuName for the resource
// - curl -s "https://prices.azure.com/api/retail/prices?\$filter=productName eq 'IP Addresses'" | jq '.Items[] | {skuName, meterName}' | sort -u

// PublicIP is the entity that holds the logic to calculate price
// of the azurerm_public_ip
type PublicIP struct {
	provider *Provider

	location         string
	sku              string
	skuTier          string
	allocationMethod string

	// Usage
	monthlyHours decimal.Decimal
}

// publicIPValues is holds the terraform values that we need to estimate the price
type publicIPValues struct {

	//required params
	Location         string `mapstructure:"location"`
	AllocationMethod string `mapstructure:"allocation_method"` // Static or Dynamic

	//optional params
	Sku     string `mapstructure:"sku"`      // Basic or Standard(requires AllocationMethod=Static). Default=Standard
	SkuTier string `mapstructure:"sku_tier"` // Regional or Global(requires sku=Standard). Default=Regional

	// usage - with default values
	Usage struct {
		MonthlyHours int64 `mapstructure:"monthly_hours"`
	} `mapstructure:"tc_usage"`
}

// decodePublicIPValues decodes and returns computeInstanceValues from a Terraform values map.
func decodePublicIPValues(tfVals map[string]interface{}) (publicIPValues, error) {
	var v publicIPValues
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

// newPublicIP initializes a new PublicIP from the provider
func (p *Provider) newPublicIP(vals publicIPValues) *PublicIP {
	inst := &PublicIP{
		provider: p,

		location:         region.GetLocationName(vals.Location),
		allocationMethod: vals.AllocationMethod,
		sku:              vals.Sku,
		skuTier:          vals.SkuTier,
		// From Usage
		monthlyHours: decimal.NewFromInt(vals.Usage.MonthlyHours),
	}
	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *PublicIP) Components() []query.Component {

	components := []query.Component{}

	skuName := inst.sku

	// misconfiguration of user - return empty cost
	if skuName == "Standard" && inst.allocationMethod == "Dynamic" {
		return components
	}

	if inst.skuTier == "Global" {
		skuName = "Global"
	}

	meterName := skuName + " IPv4 " + inst.allocationMethod + " Public IP"

	components = []query.Component{
		inst.publicIPComponent(inst.provider.key, inst.location, skuName, meterName, inst.monthlyHours),
	}

	return components
}

func (inst *PublicIP) publicIPComponent(key, location, skuName, meterName string, monthlyHours decimal.Decimal) query.Component {
	return query.Component{
		Name:            "IP adress",
		MonthlyQuantity: monthlyHours,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Virtual Network"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "meterName", Value: util.StringPtr(meterName)},
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
