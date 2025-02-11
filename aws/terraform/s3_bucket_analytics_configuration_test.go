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

func TestS3BucketAnalyticsConfiguration_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("ObjectsMonitored", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_s3_bucket_analytics_configuration.test",
			Type:         "aws_s3_bucket_analytics_configuration",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Objects monitored",
				MonthlyQuantity: decimal.NewFromFloat(0),
				Unit:            "Objects",
				Details:         []string{"Monitored"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonS3"),
					Family:   util.StringPtr("Fee"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*StorageAnalytics-ObjCount$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Objects"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_cloudwatch_metric_alarm")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})
}
