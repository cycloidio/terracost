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

func TestSQSQueue_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("FIFO", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_sqs_queue.test",
			Type:         "aws_sqs_queue",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"fifo_queue": true,
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Requests .*Requests-FIFO.*",
				MonthlyQuantity: decimal.NewFromFloat(15000000),
				Unit:            "Requests",
				Details:         []string{"SQS queue", ".*Requests-FIFO.*"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AWSQueueService"),
					Family:   util.StringPtr("API Request"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Requests-FIFO.*")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Requests"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_sqs_queue")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("NOFIFO", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_sqs_queue.test",
			Type:         "aws_sqs_queue",
			Name:         "test",
			ProviderName: "aws",
			Values:       map[string]interface{}{},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Requests .*Requests-[^F].*",
				MonthlyQuantity: decimal.NewFromFloat(15000000),
				Unit:            "Requests",
				Details:         []string{"SQS queue", ".*Requests-[^F].*"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AWSQueueService"),
					Family:   util.StringPtr("API Request"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Requests-[^F].*")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Requests"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_sqs_queue")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})
}
