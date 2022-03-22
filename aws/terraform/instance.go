package terraform

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
)

// Instance represents an EC2 instance definition that can be cost-estimated.
type Instance struct {
	provider     *Provider
	region       region.Code
	instanceType string

	// tenancy describes the tenancy of an instance.
	// Valid values include: Shared, Dedicated, Host.
	// Note: only "Shared" and "Dedicated" are supported at the moment.
	tenancy string

	// operatingSystem denotes the OS that the instance is using that may affect pricing.
	// Valid values include: Linux, RHEL, SUSE, Windows.
	// Note: only "Linux" is supported at the moment.
	operatingSystem string

	// capacityStatus describes the status of capacity reservations.
	// Valid values include: Used, UnusedCapacityReservation, AllocatedCapacityReservation.
	// Note: only "Used" is supported at the moment.
	capacityStatus string

	// preInstalledSW denotes any pre-installed software that may affect pricing.
	// Valid values include: NA, SQL Std, SQL Web, SQL Ent.
	// Note: only "NA" (no pre-installed software) is supported at the moment.
	preInstalledSW string

	rootVolume *Volume
}

// instanceValues represents the structure of Terraform values for aws_instance resource.
type instanceValues struct {
	InstanceType     string `mapstructure:"instance_type"`
	Tenancy          string `mapstructure:"tenancy"`
	AvailabilityZone string `mapstructure:"availability_zone"`

	RootBlockDevice []struct {
		VolumeType string  `mapstructure:"volume_type"`
		VolumeSize float64 `mapstructure:"volume_size"`
		IOPS       float64 `mapstructure:"iops"`
	} `mapstructure:"root_block_device"`
}

// decodeInstanceValues decodes and returns instanceValues from a Terraform values map.
func decodeInstanceValues(tfVals map[string]interface{}) (instanceValues, error) {
	var v instanceValues
	if err := mapstructure.Decode(tfVals, &v); err != nil {
		return v, err
	}
	return v, nil
}

// newInstance creates a new Instance from instanceValues.
func (p *Provider) newInstance(vals instanceValues) *Instance {
	inst := &Instance{
		provider: p,
		region:   p.region,
		tenancy:  "Shared",

		// Note: every Instance is estimated as a Linux without pre-installed S/W
		operatingSystem: "Linux",
		capacityStatus:  "Used",
		preInstalledSW:  "NA",

		instanceType: vals.InstanceType,
	}

	if reg := region.NewFromZone(vals.AvailabilityZone); reg.Valid() {
		inst.region = reg
	}

	if vals.Tenancy == "dedicated" {
		inst.tenancy = "Dedicated"
	}

	volVals := volumeValues{AvailabilityZone: vals.AvailabilityZone}
	if len(vals.RootBlockDevice) > 0 {
		rbd := vals.RootBlockDevice[0]
		volVals.Type = rbd.VolumeType
		volVals.Size = rbd.VolumeSize
		volVals.IOPS = rbd.IOPS
	}
	inst.rootVolume = p.newVolume(volVals)

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *Instance) Components() []query.Component {
	components := []query.Component{inst.computeComponent()}

	if inst.rootVolume != nil {
		for _, comp := range inst.rootVolume.Components() {
			comp.Name = "Root volume: " + comp.Name
			components = append(components, comp)
		}
	}

	return components
}

func (inst *Instance) computeComponent() query.Component {
	return query.Component{
		Name:           "Compute",
		Details:        []string{"Linux", "on-demand", inst.instanceType},
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(inst.provider.key),
			Service:  util.StringPtr("AmazonEC2"),
			Family:   util.StringPtr("Compute Instance"),
			Location: util.StringPtr(inst.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "CapacityStatus", Value: util.StringPtr(inst.capacityStatus)},
				{Key: "InstanceType", Value: util.StringPtr(inst.instanceType)},
				{Key: "Tenancy", Value: util.StringPtr(inst.tenancy)},
				{Key: "OperatingSystem", Value: util.StringPtr(inst.operatingSystem)},
				{Key: "PreInstalledSW", Value: util.StringPtr(inst.preInstalledSW)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Hrs"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
			},
		},
	}
}
