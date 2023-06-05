package terraform_test

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

func TestInstance_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
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
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Linux", "on-demand", "m5.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "CapacityStatus", Value: util.StringPtr("Used")},
						{Key: "InstanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "Tenancy", Value: util.StringPtr("Shared")},
						{Key: "OperatingSystem", Value: util.StringPtr("Linux")},
						{Key: "PreInstalledSW", Value: util.StringPtr("NA")},
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
				Name:            "Root volume: Storage",
				MonthlyQuantity: decimal.NewFromFloat(8),
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
			Address:      "aws_instance.test",
			Type:         "aws_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"availability_zone": "eu-west-1a",
				"instance_type":     "m5.xlarge",
				"tenancy":           "dedicated",
				"ebs_optimized":     true,
				"monitoring":        true,
				"credit_specification": []interface{}{
					map[string]interface{}{
						"cpu_credits": "unlimited",
					},
				},
				"root_block_device": []interface{}{
					map[string]interface{}{
						"volume_type": "st1",
						"volume_size": float64(42),
					},
				},
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Linux", "on-demand", "m5.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "CapacityStatus", Value: util.StringPtr("Used")},
						{Key: "InstanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "Tenancy", Value: util.StringPtr("Dedicated")},
						{Key: "OperatingSystem", Value: util.StringPtr("Linux")},
						{Key: "PreInstalledSW", Value: util.StringPtr("NA")},
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
				Name:            "Root volume: Storage",
				MonthlyQuantity: decimal.NewFromFloat(42),
				Unit:            "GB",
				Details:         []string{"st1"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "VolumeAPIName", Value: util.StringPtr("st1")},
					},
				},
			},
			{
				Name:           "CPUCreditCost",
				Details:        []string{"Linux", "on-demand", "m5.xlarge"},
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("CPU Credits"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "OperatingSystem", Value: util.StringPtr("Linux")},
						{Key: "UsageType", Value: util.StringPtr(fmt.Sprintf("%s-CPUCredits:%s", "EU", "m5"))},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("vCPU-Hours"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
			{
				Name:            "EC2 detailed monitoring",
				Details:         []string{"on-demand", "monitoring"},
				MonthlyQuantity: decimal.NewFromInt(int64(7)),
				ProductFilter: &product.Filter{
					Provider:         util.StringPtr("aws"),
					Service:          util.StringPtr("AmazonCloudWatch"),
					Family:           util.StringPtr("Metric"),
					Location:         util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Metrics"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:           "EBS-optimized usage",
				Details:        []string{"EBS", "Optimizes", "m5.xlarge"},
				HourlyQuantity: decimal.NewFromInt(1),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("m5.xlarge")},
						{Key: "UsageType", Value: util.StringPtr(fmt.Sprintf("%s-EBSOptimized:%s", "EU", "m5.xlarge"))},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
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
