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
	"github.com/cycloidio/terracost/util"
)

func TestEKSCluster_Components(t *testing.T) {
	p, err := NewProvider("aws", "us-east-1")
	require.NoError(t, err)

	t.Run("EKSCluster", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_eks_cluster.test",
			Type:         "aws_eks_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:           "EKS Cluster",
				Details:        []string{"EKSCluster:Compute"},
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEKS"),
					Family:   util.StringPtr("Compute"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", Value: util.StringPtr("USE1-AmazonEKS-Hours:perCluster")},
					},
				},
				PriceFilter: &price.Filter{
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
		}

		actual := p.ResourceComponents(rss, tfres)
		assert.Equal(t, expected, actual)
	})
}
