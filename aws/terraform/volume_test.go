package terraform_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

func TestVolume_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("DefaultValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_ebs_volume.test",
			Type:         "aws_ebs_volume",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"size": float64(42),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Storage",
				MonthlyQuantity: decimal.NewFromFloat(42),
				Unit:            "GB",
				Details:         []string{"gp3"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "VolumeAPIName", Value: util.StringPtr("gp3")},
					},
				},
			},
		}

		actual := p.ResourceComponents(rss, tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("WithAllValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_ebs_volume.test",
			Type:         "aws_ebs_volume",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"availability_zone": "eu-west-1a",
				"type":              "io2",
				"size":              float64(42),
				"iops":              float64(123),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Storage",
				MonthlyQuantity: decimal.NewFromFloat(42),
				Unit:            "GB",
				Details:         []string{"io2"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "VolumeAPIName", Value: util.StringPtr("io2")},
					},
				},
			},
			{
				Name:            "Provisioned IOPS",
				MonthlyQuantity: decimal.NewFromFloat(123),
				Unit:            "IOPS",
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "VolumeAPIName", Value: util.StringPtr("io2")},
						{Key: "UsageType", ValueRegex: util.StringPtr("^EBS:VolumeP-IOPS")},
					},
				},
			},
		}

		actual := p.ResourceComponents(rss, tfres)
		assert.Equal(t, expected, actual)
	})
}
