package terraform

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
)

func TestNatGateway_Components(t *testing.T) {
	p, err := NewProvider("aws", "us-east-1")
	require.NoError(t, err)

	t.Run("NAT", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_nat_gateway.test",
			Type:         "aws_nat_gateway",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"allocation_id": "id",
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:           "NAT gateway",
				Details:        []string{"NatGateway"},
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("NAT Gateway"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*NatGateway-Hours")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},

			{
				Name:            "NAT Data processed",
				Details:         []string{"NatGateway Data processed"},
				Usage:           true,
				MonthlyQuantity: decimal.NewFromFloat(10),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("NAT Gateway"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*NatGateway-Bytes")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_nat_gateway")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		assert.Equal(t, expected, actual)
	})
}
