package terraform_test

import (
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

	// resource "" "" {

	//   launch_template {
	//     id      = aws_launch_template.foobar.id
	//     version = "$Latest"
	//   }
	// }

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
				"launch_template":    []interface{}{map[string]interface{}{"id": []string{"aws_launch_template.test"}}},
			},
		}

		rss := map[string]terraform.Resource{
			"aws_launch_template.test": terraform.Resource{
				Address:      "aws_launch_template.test",
				Type:         "aws_launch_template",
				Name:         "test",
				ProviderName: "aws",
				Values: map[string]interface{}{
					"instance_type":        "t3.large",
					"ebs_optimized":        true,
					"placement":            []interface{}{map[string]interface{}{"availability_zone": "eu-west-1c", "tenancy": "dedicated"}},
					"credit_specification": []interface{}{map[string]interface{}{"cpu_credits": "unlimited"}},
					"monitoring":           []interface{}{map[string]interface{}{"enabled": true}},
					"block_device_mappings": []interface{}{
						map[string]interface{}{
							"device_name": "/dev/sda1",
							"ebs": []interface{}{
								map[string]interface{}{"volume_size": 20},
							},
						},
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
				MonthlyQuantity: decimal.NewFromInt(8),
				Unit:            "GB",
				Details:         []string{"gp2"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEC2"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "VolumeAPIName", Value: util.StringPtr("gp2")},
					},
				},
			},
		}

		actual := p.ResourceComponents(rss, tfres)
		assert.Equal(t, expected, actual)
	})

}
