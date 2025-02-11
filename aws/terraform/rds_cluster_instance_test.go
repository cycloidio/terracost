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

func TestRDSClusterInstance_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("RDSClusterInstanceMysql", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_rds_cluster_instance.test",
			Type:         "aws_rds_cluster_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"engine":                                "aurora-mysql",
				"instance_class":                        "db.r4.large",
				"performance_insights_retention_period": 31,
				"performance_insights_enabled":          true,
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:           "Database instance",
				HourlyQuantity: decimal.NewFromFloat(1),
				Unit:           "Hrs",
				Details:        []string{"instance"},
				Usage:          false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Instance"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("db.r4.large")},
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora MySQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*InstanceUsage:.*")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Performance Insights Long Term Retention (serverless)",
				MonthlyQuantity: decimal.NewFromFloat(0.5),
				Unit:            "ACU-Months",
				Details:         []string{"Insights"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Performance Insights"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora MySQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*PI_LTR_FMR:Serverless$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("ACU-Months"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Performance Insights API",
				MonthlyQuantity: decimal.NewFromFloat(500000),
				Unit:            "API Calls",
				Details:         []string{"Requests"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Performance Insights"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*PI_API$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("API Calls"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_rds_cluster_instance")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})
}
