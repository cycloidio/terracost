package terraform

import (
	"github.com/shopspring/decimal"

	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/util"
)

// Volume represents an EBS volume.
type Volume struct {
	providerName string
	region       string
	volumeType   string
	size         decimal.Decimal
	iops         decimal.Decimal
}

// NewVolume creates a new Volume from Terraform values.
func (p *Provider) NewVolume(values map[string]interface{}) *Volume {
	zone, _ := values["availability_zone"].(string)
	volType, _ := values["type"].(string)
	volSize, _ := values["size"].(float64)
	volIops, _ := values["iops"].(float64)

	v := &Volume{
		providerName: p.name,
		region:       zoneToRegion(zone),
		volumeType:   volType,
		size:         decimal.NewFromFloat(volSize),
		iops:         decimal.NewFromFloat(volIops),
	}

	// Use provider region if AZ not specified on the resource
	if v.region == "" {
		v.region = p.region
	}
	// "gp2" is the default volume type
	if v.volumeType == "" {
		v.volumeType = "gp2"
	}
	// 8 GiB is the default volume size if none is provided
	if v.size.LessThanOrEqual(decimal.Zero) {
		v.size = decimal.NewFromInt(8)
	}

	return v
}

// Components returns the price component queries that make up the Volume.
func (v *Volume) Components() []query.Component {
	comps := []query.Component{v.storageComponent()}

	if v.volumeType == "io1" || v.volumeType == "io2" {
		comps = append(comps, v.iopsComponent())
	}

	return comps
}

func (v *Volume) storageComponent() query.Component {
	return query.Component{
		Name:     "Storage",
		Quantity: v.size,
		Unit:     "GB-Mo",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.providerName),
			Service:  util.StringPtr("AmazonEC2"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(v.region),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "volumeApiName", Value: util.StringPtr(v.volumeType)},
			},
		},
	}
}

func (v *Volume) iopsComponent() query.Component {
	return query.Component{
		Name:     "Provisioned IOPS",
		Quantity: v.iops,
		Unit:     "IOPS-Mo",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.providerName),
			Service:  util.StringPtr("AmazonEC2"),
			Family:   util.StringPtr("System Operation"),
			Location: util.StringPtr(v.region),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "volumeApiName", Value: util.StringPtr(v.volumeType)},
				{Key: "usagetype", ValueRegex: util.StringPtr("^EBS:VolumeP-IOPS")},
			},
		},
	}
}
