package terraform

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

// CloudwatchMetricAlarm represents an SQS queue definition that can be cost-estimated.
type CloudwatchMetricAlarm struct {
	provider           *Provider
	region             region.Code
	comparisonOperator string
	metricsCount       decimal.Decimal
	period             decimal.Decimal
}

type cloudwatchMetricAlarmValues struct {
	ComparisonOperator string  `mapstructure:"comparison_operator"`
	Period             float64 `mapstructure:"period"`
	MetricQuery        []struct {
		Metric []struct {
			Period float64 `mapstructure:"period"`
		} `mapstructure:"metric"`
	} `mapstructure:"metric_query"`
}

// decodeCloudwatchMetricAlarmValues decodes and returns cloudwatchMetricAlarmValues from a Terraform values map.
func decodeCloudwatchMetricAlarmValues(tfVals map[string]interface{}) (cloudwatchMetricAlarmValues, error) {
	var v cloudwatchMetricAlarmValues
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &v,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return v, err
	}

	if err := decoder.Decode(tfVals); err != nil {
		return v, err
	}
	return v, nil
}

// newCloudwatchMetricAlarm creates a new CloudwatchMetricAlarm from cloudwatchMetricAlarmValues.
func (p *Provider) newCloudwatchMetricAlarm(_ map[string]terraform.Resource, vals cloudwatchMetricAlarmValues) *CloudwatchMetricAlarm {
	v := &CloudwatchMetricAlarm{
		provider:           p,
		region:             p.region,
		comparisonOperator: vals.ComparisonOperator,
		metricsCount:       decimal.NewFromFloat(1),
		period:             decimal.NewFromFloat(60),
	}

	if vals.Period > 0 {
		v.period = decimal.NewFromFloat(vals.Period)
	}

	if len(vals.MetricQuery) > 0 {
		metricCount := 0
		for _, metricQuery := range vals.MetricQuery {
			if len(metricQuery.Metric) > 0 {
				for _, metric := range metricQuery.Metric {
					if metric.Period > 0 {
						// if a period is defined, take the highest to estimate
						if v.period.Cmp(decimal.NewFromFloat(metric.Period)) < 0 {
							v.period = decimal.NewFromFloat(metric.Period)
						}
					}
					metricCount++
				}
			}
		}

		if metricCount > 0 {
			v.metricsCount = decimal.NewFromInt(int64(metricCount))
		}
	}

	return v
}

// Components returns the price component queries that make up the CloudwatchMetricAlarm.
func (v *CloudwatchMetricAlarm) Components() []query.Component {
	components := []query.Component{v.cloudwatchMetricAlarmComponent()}
	return components
}

func (v *CloudwatchMetricAlarm) cloudwatchMetricAlarmComponent() query.Component {
	quantity := v.metricsCount
	unit := "alarm metrics"
	anomalyDetection := ""
	alarmType := "Standard"
	alarmName := fmt.Sprintf("%s%s", "Standard resolution", anomalyDetection)

	switch v.comparisonOperator {
	case "LessThanLowerOrGreaterThanUpperThreshold", "LessThanLowerThreshold", "GreaterThanUpperThreshold":
		quantity = quantity.Mul(decimal.NewFromInt(3))
		unit = "Alarms"
		anomalyDetection = " anomaly detection"
	}

	if v.period.Div(decimal.NewFromInt(60)).LessThan(decimal.NewFromInt(1)) {
		alarmName = fmt.Sprintf("%s%s", "High resolution", anomalyDetection)
		alarmType = "High Resolution"
	}

	return query.Component{
		Name:            alarmName,
		MonthlyQuantity: quantity,
		Details:         []string{alarmName, alarmType},
		Usage:           true,
		Unit:            unit,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonCloudWatch"),
			Family:   util.StringPtr("Alarm"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(".*AlarmMonitorUsage")},
				{Key: "AlarmType", Value: util.StringPtr(alarmType)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Alarms"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}
