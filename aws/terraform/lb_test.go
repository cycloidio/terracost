package terraform_test

import (
	"testing"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
)

func TestLB_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-3")
	require.NoError(t, err)

	t.Run("DefaultValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_lb.test",
			Type:         "aws_lb",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		expected := []query.Component{
			{
				Name:           "Application Load Balancer",
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AWSELB"),
					Family:   util.StringPtr("Load Balancer-Application"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr("LoadBalancerUsage")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("NetworkLoadBalancer", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_lb.test",
			Type:         "aws_lb",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"load_balancer_type": "network",
			},
		}
		expected := []query.Component{
			{
				Name:           "Network Load Balancer",
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AWSELB"),
					Family:   util.StringPtr("Load Balancer-Network"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr("LoadBalancerUsage")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("GatewayLoadBalancer", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_lb.test",
			Type:         "aws_lb",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"load_balancer_type": "gateway",
			},
		}
		expected := []query.Component{
			{
				Name:           "Gateway Load Balancer",
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AWSELB"),
					Family:   util.StringPtr("Load Balancer-Gateway"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr("LoadBalancerUsage")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("ClassicLoadBalancer", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elb.test",
			Type:         "aws_elb",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		expected := []query.Component{
			{
				Name:           "Classic Load Balancer",
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AWSELB"),
					Family:   util.StringPtr("Load Balancer"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr("LoadBalancerUsage")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})
}
