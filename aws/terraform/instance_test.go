package terraform

import (
	"testing"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/terraform"
	"github.com/cycloidio/cost-estimation/util"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestInstance_Components(t *testing.T) {
	p := NewProvider("aws", "eu-west-3")

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
				Name:     "EC2 instance hours",
				Quantity: decimal.NewFromInt(730),
				Unit:     "Hrs",
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "instanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "tenancy", Value: util.StringPtr("Shared")},
						{Key: "operatingSystem", Value: util.StringPtr("Linux")},
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
				Name:     "Root volume: Storage",
				Quantity: decimal.NewFromInt(8),
				Unit:     "GB-Mo",
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
				Name:     "EC2 instance hours",
				Quantity: decimal.NewFromInt(730),
				Unit:     "Hrs",
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "instanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "tenancy", Value: util.StringPtr("Dedicated")},
						{Key: "operatingSystem", Value: util.StringPtr("Linux")},
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
				Name:     "Root volume: Storage",
				Quantity: decimal.NewFromInt(42),
				Unit:     "GB-Mo",
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
