package terraform_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
)

func TestCloudwatchLogGroup_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("LogGroup", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_cloudwatch_log_group.test",
			Type:         "aws_cloudwatch_log_group",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Data ingested",
				MonthlyQuantity: decimal.NewFromFloat(10),
				Unit:            "GB",
				Details:         []string{"Data ingested"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonCloudWatch"),
					Family:   util.StringPtr("Data Payload"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-DataProcessing-Bytes")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Archival Storage",
				MonthlyQuantity: decimal.NewFromFloat(200),
				Unit:            "GB-Mo",
				Details:         []string{"Archival Storage"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonCloudWatch"),
					Family:   util.StringPtr("Storage Snapshot"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-TimedStorage-ByteHrs")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB-Mo"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Insights queries data scanned",
				MonthlyQuantity: decimal.NewFromFloat(20),
				Unit:            "GB",
				Details:         []string{"Insights queries", "data scanned", "Storage"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonCloudWatch"),
					Family:   util.StringPtr("Data Payload"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-DataScanned-Bytes")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_cloudwatch_log_group")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})
}
