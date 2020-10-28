package terraform_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/cost-estimation/aws/terraform"
	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/terraform"
	"github.com/cycloidio/cost-estimation/util"
)

func TestInstance_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-3")
	require.NoError(t, err)

	t.Run("DefaultValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_instance.test",
			Type:         "aws_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"instance_type": "m5.xlarge",
			},
		}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Linux", "on-demand", "m5.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "capacitystatus", Value: util.StringPtr("Used")},
						{Key: "instanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "tenancy", Value: util.StringPtr("Shared")},
						{Key: "operatingSystem", Value: util.StringPtr("Linux")},
						{Key: "preInstalledSw", Value: util.StringPtr("NA")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "purchaseOption", Value: util.StringPtr("on_demand")},
					},
				},
			},
			{
				Name:            "Root volume: Storage",
				MonthlyQuantity: decimal.NewFromInt(8),
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
			Address:      "aws_instance.test",
			Type:         "aws_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"availability_zone": "us-east-1a",
				"instance_type":     "m5.xlarge",
				"tenancy":           "dedicated",
				"root_block_device": []interface{}{
					map[string]interface{}{
						"volume_type": "st1",
						"volume_size": float64(42),
					},
				},
			},
		}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Linux", "on-demand", "m5.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "capacitystatus", Value: util.StringPtr("Used")},
						{Key: "instanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "tenancy", Value: util.StringPtr("Dedicated")},
						{Key: "operatingSystem", Value: util.StringPtr("Linux")},
						{Key: "preInstalledSw", Value: util.StringPtr("NA")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "purchaseOption", Value: util.StringPtr("on_demand")},
					},
				},
			},
			{
				Name:            "Root volume: Storage",
				MonthlyQuantity: decimal.NewFromInt(42),
				Unit:            "GB",
				Details:         []string{"st1"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "volumeApiName", Value: util.StringPtr("st1")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})
}
