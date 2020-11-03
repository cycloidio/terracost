package terraform

import (
	"github.com/cycloidio/cost-estimation/query"
	"testing"

	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/terraform"
	"github.com/cycloidio/cost-estimation/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestVolume_Components(t *testing.T) {
	p := NewProvider("aws", "eu-west-3")

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

		expected := []query.Component{
			{
				Name:            "Storage",
				MonthlyQuantity: decimal.NewFromInt(42),
				Unit:            "GB",
				Details:         []string{"gp2"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "volumeApiName", Value: util.StringPtr("gp2")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("WithAllValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_ebs_volume.test",
			Type:         "aws_ebs_volume",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"availability_zone": "us-east-1a",
				"type":              "io2",
				"size":              float64(42),
				"iops":              float64(123),
			},
		}

		expected := []query.Component{
			{
				Name:            "Storage",
				MonthlyQuantity: decimal.NewFromInt(42),
				Unit:            "GB",
				Details:         []string{"io2"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "volumeApiName", Value: util.StringPtr("io2")},
					},
				},
			},
			{
				Name:            "Provisioned IOPS",
				MonthlyQuantity: decimal.NewFromInt(123),
				Unit:            "IOPS",
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "volumeApiName", Value: util.StringPtr("io2")},
						{Key: "usagetype", ValueRegex: util.StringPtr("^EBS:VolumeP-IOPS")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})
}
