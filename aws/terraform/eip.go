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

// ElasticIP represents an ElasticIP instance definition that can be cost-estimated.
type ElasticIP struct {
	providerKey           string
	region                region.Code
	customerOwnedIpv4Pool string
	instance              string
	networkInterface      string
}

type elasticIPValues struct {
	CustomerOwnedIpv4Pool string `mapstructure:"customer_owned_ipv4_pool"`
	Instance              string `mapstructure:"instance"`
	NetworkInterface      string `mapstructure:"network_interface"`
}

func decodeElasticIPValues(tfVals map[string]interface{}) (elasticIPValues, error) {
	var v elasticIPValues
	if err := mapstructure.Decode(tfVals, &v); err != nil {
		return v, err
	}
	return v, nil
}

// NewInstance creates a new Instance from Terraform values.
func (p *Provider) newElasticIP(vals elasticIPValues) *ElasticIP {

	inst := &ElasticIP{
		providerKey:           p.key,
		region:                p.region,
		customerOwnedIpv4Pool: vals.CustomerOwnedIpv4Pool,
		instance:              vals.Instance,
		networkInterface:      vals.NetworkInterface,
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *ElasticIP) Components() []query.Component {
	// An Elastic IP address doesn’t incur charges as long as all the following conditions are true:
	// * The Elastic IP address is associated with an EC2 instance.
	// * The instance associated with the Elastic IP address is running.
	// * The instance has only one Elastic IP address attached to it.
	// * The Elastic IP address is associated with an attached network interface
	if len(inst.customerOwnedIpv4Pool) > 0 || len(inst.instance) > 0 || len(inst.networkInterface) > 0 {
		return []query.Component{}
	}

	components := []query.Component{inst.elasticIPInstanceComponent()}

	return components
}

func (inst *ElasticIP) elasticIPInstanceComponent() query.Component {

	attrFilters := []*product.AttributeFilter{
		{Key: "Group", Value: util.StringPtr("ElasticIP:IdleAddress")},
	}

	return query.Component{
		Name:           "Elastic IP",
		Details:        []string{"ElasticIP:IdleAddress"},
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider:         util.StringPtr(inst.providerKey),
			Service:          util.StringPtr("AmazonEC2"),
			Family:           util.StringPtr("IP Address"),
			Location:         util.StringPtr(inst.region.String()),
			AttributeFilters: attrFilters,
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Hrs"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("1")},
			},
		},
	}
}
