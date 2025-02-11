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

func TestKMSKey_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("SYMMETRIC_DEFAULT", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_kms_key.test",
			Type:         "aws_kms_key",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"customer_master_key_spec": "SYMMETRIC_DEFAULT",
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Customer master key",
				MonthlyQuantity: decimal.NewFromFloat(1),
				Unit:            "Keys",
				Details:         []string{"master key"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("awskms"),
					Family:   util.StringPtr("Encryption Key"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*KMS-Keys")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Keys"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Requests",
				MonthlyQuantity: decimal.NewFromFloat(1),
				Unit:            "Requests",
				Details:         []string{"Request"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("awskms"),
					Family:   util.StringPtr("API Request"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*KMS-Requests$")},
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
			{
				Name:            "ECC GenerateDataKeyPair requests",
				MonthlyQuantity: decimal.NewFromFloat(1),
				Unit:            "Requests",
				Details:         []string{"Request"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("awskms"),
					Family:   util.StringPtr(""),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*KMS-Requests-GenerateDatakeyPair-ECC$")},
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
			{
				Name:            "RSA GenerateDataKeyPair requests",
				MonthlyQuantity: decimal.NewFromFloat(1),
				Unit:            "Requests",
				Details:         []string{"Request"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("awskms"),
					Family:   util.StringPtr(""),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*KMS-Requests-GenerateDatakeyPair-RSA$")},
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

		us := usage.Default.GetUsage("aws_cloudwatch_metric_alarm")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("RSA_4096", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_kms_key.test",
			Type:         "aws_kms_key",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"customer_master_key_spec": "RSA_4096",
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Customer master key",
				MonthlyQuantity: decimal.NewFromFloat(1),
				Unit:            "Keys",
				Details:         []string{"master key"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("awskms"),
					Family:   util.StringPtr("Encryption Key"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*KMS-Keys")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Keys"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Requests (asymmetric)",
				MonthlyQuantity: decimal.NewFromFloat(1),
				Unit:            "Requests",
				Details:         []string{"Request"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("awskms"),
					Family:   util.StringPtr(""),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*KMS-Requests-Asymmetric$")},
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

		us := usage.Default.GetUsage("aws_cloudwatch_metric_alarm")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

}
