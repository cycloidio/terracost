package terraform

import (
	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/terraform"
	"github.com/mitchellh/mapstructure"
)

// dnsZoneValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type dnsZoneValues struct {
	Location          string `mapstructure:"location"`
	ResourceGroupName string `mapstructure:"resource_group_name"`
}

// decodeDNSZoneValues decodes and returns Values from a Terraform values map.
func decodeDNSZoneValues(tfVals map[string]interface{}) (dnsZoneValues, error) {
	var v dnsZoneValues
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

// newDNSZone initializes a new DNSZone from the provider
func (p *Provider) newDNSZone(rss map[string]terraform.Resource, vals dnsZoneValues) *DNSZone {
	inst := &DNSZone{
		provider: p,
		location: "Zone 1",
		zoneType: "Public",
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
