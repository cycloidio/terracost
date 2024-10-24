package terraform

import (
	"fmt"

	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// DNSZone is the entity that holds the logic to calculate price
type DNSZone struct {
	provider *Provider
	location string

	zoneType string
}

// privateDNSZoneValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type privateDNSZoneValues struct {
	Location string `mapstructure:"location"`

	ResourceGroupName string `mapstructure:"resource_group_name"`
}

// decodePrivateDNSZoneValues decodes and returns Values from a Terraform values map.
func decodePrivateDNSZoneValues(tfVals map[string]interface{}) (privateDNSZoneValues, error) {
	var v privateDNSZoneValues
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

// newPrivateDNSZone initializes a new PrivateDNSZone from the provider
func (p *Provider) newPrivateDNSZone(rss map[string]terraform.Resource, vals privateDNSZoneValues) *DNSZone {
	inst := &DNSZone{
		provider: p,
		location: "Zone 1",
		zoneType: "Private",
	}

	rg, err := decodeResourceGroupValues(rss[vals.ResourceGroupName].Values)
	if err != nil {
		return inst
	}

	// Get the location from RG
	if rg.Location != "" {
		inst.location = region.GetRegionToVNETZone(region.GetLocationName(rg.Location))
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *DNSZone) Components() []query.Component {
	components := []query.Component{
		inst.dnsZoneComponent(inst.provider.key, inst.location, inst.zoneType),
	}

	return components
}

func (inst *DNSZone) dnsZoneComponent(key string, location string, zoneType string) query.Component {
	return query.Component{
		Name:            fmt.Sprintf("Hosted zone %s", zoneType),
		MonthlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Azure DNS"),
			Family:   util.StringPtr("Networking"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "meterName", Value: util.StringPtr(fmt.Sprintf("%s Zone", zoneType))},
				// Use price for 25 or less zones
				{Key: "tierMinimumUnits", Value: util.StringPtr(fmt.Sprintf("%f", float64(0)))},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
