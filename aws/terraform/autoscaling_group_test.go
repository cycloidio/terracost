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

func TestAutoscalingGroup_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("LaunchTemplate", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "module.test.aws_autoscaling_group.lt",
			Type:         "aws_autoscaling_group",
			Name:         "lt",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"desired_capacity":   2,
				"min_size":           1,
				"max_size":           5,
				"availability_zones": []string{"eu-west-1a", "eu-west-1b"},
				"launch_template":    []interface{}{map[string]interface{}{"id": "aws_launch_template.test"}},
			},
		}

		rss := map[string]terraform.Resource{
			"aws_launch_template.test": terraform.Resource{
				Address:      "aws_launch_template.test",
				Type:         "aws_launch_template",
				Name:         "test",
				ProviderName: "aws",
				Values: map[string]interface{}{
					"instance_type":        "m5.xlarge",
					"ebs_optimized":        true,
					"placement":            []interface{}{map[string]interface{}{"availability_zone": "eu-west-1a", "tenancy": "dedicated"}},
					"credit_specification": []interface{}{map[string]interface{}{"cpu_credits": "unlimited"}},
					"monitoring":           []interface{}{map[string]interface{}{"enabled": true}},
					"block_device_mappings": []interface{}{
						map[string]interface{}{
							"device_name": "/dev/sda1",
							"ebs": []interface{}{
								map[string]interface{}{"volume_size": float64(20)},
							},
						},
					},
				},
			},
		}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(2),
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
				MonthlyQuantity: decimal.NewFromFloat(20),
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
			{
				Name:           "CPUCreditCost",
				Details:        []string{"Linux", "on-demand", "m5.xlarge"},
				HourlyQuantity: decimal.NewFromInt(2),
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
				MonthlyQuantity: decimal.NewFromInt(int64(14)),
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
				HourlyQuantity: decimal.NewFromInt(2),
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

	t.Run("MixedInstancesPolicyLaunchTemplate", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_autoscaling_group.mixlt",
			Type:         "aws_autoscaling_group",
			Name:         "mixlt",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"desired_capacity":   3,
				"availability_zones": []string{"eu-west-1a"},
				"mixed_instances_policy": []interface{}{
					map[string]interface{}{
						"launch_template": []interface{}{
							map[string]interface{}{
								"launch_template_specification": []interface{}{
									map[string]interface{}{
										"launch_template_name": "aws_launch_template.testmix",
									},
								},
								"override": []interface{}{
									map[string]interface{}{
										"instance_type": "c5.large",
									},
								},
							},
						},
					},
				},
			},
		}

		rss := map[string]terraform.Resource{
			"aws_launch_template.testmix": terraform.Resource{
				Address:      "aws_launch_template.testmix",
				Type:         "aws_launch_template",
				Name:         "testmix",
				ProviderName: "aws",
				Values: map[string]interface{}{
					"instance_type": "m5.xlarge",
					"placement":     []interface{}{map[string]interface{}{"availability_zone": "eu-west-1a"}},
					"block_device_mappings": []interface{}{
						map[string]interface{}{
							"device_name": "/dev/sda1",
							"ebs": []interface{}{
								map[string]interface{}{"volume_size": float64(30)},
							},
						},
					},
				},
			},
		}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(3),
				Details:        []string{"Linux", "on-demand", "c5.large"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Compute Instance"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "CapacityStatus", Value: util.StringPtr("Used")},
						{Key: "InstanceType", Value: util.StringPtr("c5.large")},
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
				MonthlyQuantity: decimal.NewFromFloat(30),
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

	t.Run("LaunchConfiguration", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "module.test.aws_autoscaling_group.lt",
			Type:         "aws_autoscaling_group",
			Name:         "lt",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"desired_capacity":     2,
				"min_size":             1,
				"max_size":             5,
				"availability_zones":   []string{"eu-west-1a", "eu-west-1b"},
				"launch_configuration": "aws_launch_configuration.test",
			},
		}

		rss := map[string]terraform.Resource{
			"aws_launch_configuration.test": terraform.Resource{
				Address:      "aws_launch_configuration.test",
				Type:         "aws_launch_configuration",
				Name:         "test",
				ProviderName: "aws",
				Values: map[string]interface{}{
					"instance_type":     "m5.xlarge",
					"enable_monitoring": true,
					"placement_tenancy": "dedicated",

					"root_block_device": []interface{}{
						map[string]interface{}{
							"volume_type": "gp3",
							"volume_size": float64(42),
						},
					},
				},
			},
		}

		expected := []query.Component{
			{
				Name:           "Compute",
				HourlyQuantity: decimal.NewFromInt(2),
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
			{
				Name:            "EC2 detailed monitoring",
				Details:         []string{"on-demand", "monitoring"},
				MonthlyQuantity: decimal.NewFromInt(int64(14)),
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
		}

		actual := p.ResourceComponents(rss, tfres)
		assert.Equal(t, expected, actual)
	})

}
