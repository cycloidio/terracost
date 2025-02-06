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

func TestCloudwatchMetricAlarm_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("AlarmStandard", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_cloudwatch_metric_alarm.test",
			Type:         "aws_cloudwatch_metric_alarm",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"comparison_operator": "GreaterThanOrEqualToThreshold",
				"period":              60,
				"metric_query": []interface{}{
					map[string]interface{}{
						"metric": map[string]interface{}{
							"period": 180,
						},
					},
					map[string]interface{}{
						"metric": map[string]interface{}{
							"period": 60,
						},
					},
				},
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Standard resolution",
				MonthlyQuantity: decimal.NewFromFloat(2),
				Unit:            "alarm metrics",
				Details:         []string{"Standard resolution", "Standard"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonCloudWatch"),
					Family:   util.StringPtr("Alarm"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*AlarmMonitorUsage")},
						{Key: "AlarmType", Value: util.StringPtr("Standard")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Alarms"),
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

	t.Run("AlarmHigh", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_cloudwatch_metric_alarm.test",
			Type:         "aws_cloudwatch_metric_alarm",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"comparison_operator": "LessThanLowerOrGreaterThanUpperThreshold",
				"period":              30,
				"metric_query": []interface{}{
					map[string]interface{}{
						"metric": map[string]interface{}{},
					},
				},
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "High resolution anomaly detection",
				MonthlyQuantity: decimal.NewFromFloat(3),
				Unit:            "Alarms",
				Details:         []string{"High resolution anomaly detection", "High Resolution"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonCloudWatch"),
					Family:   util.StringPtr("Alarm"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*AlarmMonitorUsage")},
						{Key: "AlarmType", Value: util.StringPtr("High Resolution")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Alarms"),
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
