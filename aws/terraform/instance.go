package terraform

import (
	"github.com/shopspring/decimal"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/util"
)

// Instance represents an EC2 instance.
type Instance struct {
	providerName string
	region       string
	instanceType string
	tenancy      string

	operatingSystem string
	capacityStatus  string
	preInstalledSW  string

	rootVolume *Volume
}

// NewInstance creates a new Instance from Terraform values.
func (p *Provider) NewInstance(values map[string]interface{}) *Instance {
	instType, _ := values["instance_type"].(string)
	tenancyVal, _ := values["tenancy"].(string)
	zone, _ := values["availability_zone"].(string)

	inst := &Instance{
		providerName: p.name,
		instanceType: instType,
		region:       zoneToRegion(zone),

		// Note: every Instance is estimated as a Linux without pre-installed S/W
		operatingSystem: "Linux",
		preInstalledSW:  "NA",
		capacityStatus:  "Used",
	}

	// Terraform uses "default"/"dedicated", AWS expects "Shared"/"Dedicated"
	switch tenancyVal {
	case "dedicated":
		inst.tenancy = "Dedicated"
	default:
		inst.tenancy = "Shared"
	}

	// Use provider region if AZ not specified on the resource
	if inst.region == "" {
		inst.region = p.region
	}

	var volParams map[string]interface{}
	if rbss, ok := values["root_block_device"].([]interface{}); ok && len(rbss) > 0 {
		if rbsVals, ok := rbss[0].(map[string]interface{}); ok {
			// root_block_device attribute uses `volume_type` and `volume_size`, whereas the aws_ebs_volume
			// (and subsequently the NewVolume func) expects `type` and `size` instead
			volParams = map[string]interface{}{
				"availability_zone": zone,
				"type":              rbsVals["volume_type"],
				"size":              rbsVals["volume_size"],
			}
		}
	}
	if volParams == nil {
		// Default EBS Volume values
		volParams = map[string]interface{}{"availability_zone": zone}
	}
	inst.rootVolume = p.NewVolume(volParams)

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *Instance) Components() []query.Component {
	components := []query.Component{inst.computeComponent()}

	if inst.rootVolume != nil {
		volumeComponents := inst.rootVolume.Components()
		for i, c := range volumeComponents {
			volumeComponents[i].Name = "Root volume: " + c.Name
		}
		components = append(components, volumeComponents...)
	}

	return components
}

func (inst *Instance) computeComponent() query.Component {
	return query.Component{
		Name:           "Compute",
		Details:        []string{"Linux", "on-demand", inst.instanceType},
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(inst.providerName),
			Service:  util.StringPtr("AmazonEC2"),
			Family:   util.StringPtr("Compute Instance"),
			Location: util.StringPtr(inst.region),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "capacitystatus", Value: util.StringPtr(inst.capacityStatus)},
				{Key: "instanceType", Value: util.StringPtr(inst.instanceType)},
				{Key: "tenancy", Value: util.StringPtr(inst.tenancy)},
				{Key: "operatingSystem", Value: util.StringPtr(inst.operatingSystem)},
				{Key: "preInstalledSw", Value: util.StringPtr(inst.preInstalledSW)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Hrs"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "purchaseOption", Value: util.StringPtr("on_demand")},
			},
		},
	}
}

func zoneToRegion(zone string) string {
	if len(zone) < 1 {
		return ""
	}
	return zone[:len(zone)-1]
}
