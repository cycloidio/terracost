package terraform_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
)

func TestEFSFileSystem_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("DefaultValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_efs_file_system.test",
			Type:         "aws_efs_file_system",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Storage .*-TimedStorage-ByteHrs",
				MonthlyQuantity: decimal.NewFromFloat(180),
				Unit:            "GB",
				Details:         []string{"EFS storage", ".*-TimedStorage-ByteHrs"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEFS"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-TimedStorage-ByteHrs")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_efs_file_system")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("WithAllValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_efs_file_system.test",
			Type:         "aws_efs_file_system",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"availability_zone": "eu-west-1a",
				"lifecycle_policy": []interface{}{map[string]interface{}{
					"transition_to_ia": "AFTER_30_DAYS",
				}},
				"throughput_mode":                 "provisioned",
				"provisioned_throughput_in_mibps": float64(20),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Storage .*-TimedStorage-ByteHrs",
				MonthlyQuantity: decimal.NewFromFloat(180),
				Unit:            "GB",
				Details:         []string{"EFS storage", ".*-TimedStorage-ByteHrs"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEFS"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-TimedStorage-ByteHrs")},
					},
				},
			},
			{
				Name:            "Provisioned throughput",
				MonthlyQuantity: decimal.NewFromFloat(11),
				Unit:            "MBps",
				Details:         []string{"Througput"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEFS"),
					Family:   util.StringPtr("Provisioned Throughput"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr("ProvisionedTP-MiBpsHrs")},
					},
				},
			},
			{
				Name:            "Storage .*-IATimedStorage-ByteHrs",
				MonthlyQuantity: decimal.NewFromFloat(10),
				Unit:            "GB", Details: []string{"EFS storage", ".*-IATimedStorage-ByteHrs"},
				Usage: true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEFS"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-IATimedStorage-ByteHrs")},
					},
				},
			},
			{
				Name:            "Requests Read",
				MonthlyQuantity: decimal.NewFromFloat(20),
				Unit:            "GB",
				Details:         []string{"Requests", "Infrequent Access", "Read"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEFS"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "AccessType", Value: util.StringPtr("Read")},
						{Key: "StorageClass", Value: util.StringPtr("Infrequent Access")},
					},
				},
			},
			{
				Name:            "Requests Write",
				MonthlyQuantity: decimal.NewFromFloat(30),
				Unit:            "GB",
				Details:         []string{"Requests", "Infrequent Access", "Write"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonEFS"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "AccessType", Value: util.StringPtr("Write")},
						{Key: "StorageClass", Value: util.StringPtr("Infrequent Access")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_efs_file_system")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})
}
