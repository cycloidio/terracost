package terraform

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

// CloudwatchLogGroup represents an SQS queue definition that can be cost-estimated.
type CloudwatchLogGroup struct {
	provider *Provider
	region   region.Code

	// Usage
	monthlyDataIngestedGB        decimal.Decimal
	storageGB                    decimal.Decimal
	monthlyDataScannedInsightsGB decimal.Decimal
}

type cloudwatchLogGroupValues struct {
	Usage struct {
		MonthlyDataIngestedGB        float64 `mapstructure:"monthly_data_ingested_gb"`
		StorageGB                    float64 `mapstructure:"storage_gb"`
		MonthlyDataScannedInsightsGB float64 `mapstructure:"monthly_data_scanned_insights_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeCloudwatchLogGroupValues decodes and returns cloudwatchLogGroupValues from a Terraform values map.
func decodeCloudwatchLogGroupValues(tfVals map[string]interface{}) (cloudwatchLogGroupValues, error) {
	var v cloudwatchLogGroupValues
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

// newCloudwatchLogGroup creates a new CloudwatchLogGroup from cloudwatchLogGroupValues.
func (p *Provider) newCloudwatchLogGroup(rss map[string]terraform.Resource, vals cloudwatchLogGroupValues) *CloudwatchLogGroup {
	// The 'rss' variable contains information from linked resources.
	// Currently, it is not utilized in this resource.
	_ = rss

	v := &CloudwatchLogGroup{
		provider: p,
		region:   p.region,

		// From Usage
		monthlyDataIngestedGB:        decimal.NewFromFloat(vals.Usage.MonthlyDataIngestedGB),
		storageGB:                    decimal.NewFromFloat(vals.Usage.StorageGB),
		monthlyDataScannedInsightsGB: decimal.NewFromFloat(vals.Usage.MonthlyDataScannedInsightsGB),
	}

	return v
}

// Components returns the price component queries that make up the CloudwatchLogGroup.
func (v *CloudwatchLogGroup) Components() []query.Component {
	components := []query.Component{v.cloudwatchLogGroupComponent()}
	components = append(components, v.cloudwatchLogGroupArchivalStorageComponent())
	components = append(components, v.cloudwatchLogGroupInsightsScannedComponent())
	return components
}

func (v *CloudwatchLogGroup) cloudwatchLogGroupComponent() query.Component {
	return query.Component{
		Name:            "Data ingested",
		MonthlyQuantity: v.monthlyDataIngestedGB,
		Details:         []string{"Data ingested"},
		Usage:           true,
		Unit:            "GB",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonCloudWatch"),
			Family:   util.StringPtr("Data Payload"),
			Location: util.StringPtr(v.region.String()),
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
	}
}

func (v *CloudwatchLogGroup) cloudwatchLogGroupArchivalStorageComponent() query.Component {
	return query.Component{
		Name:            "Archival Storage",
		MonthlyQuantity: v.storageGB,
		Details:         []string{"Archival Storage"},
		Usage:           true,
		Unit:            "GB-Mo",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonCloudWatch"),
			Family:   util.StringPtr("Storage Snapshot"),
			Location: util.StringPtr(v.region.String()),
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
	}
}

func (v *CloudwatchLogGroup) cloudwatchLogGroupInsightsScannedComponent() query.Component {
	return query.Component{
		Name:            "Insights queries data scanned",
		MonthlyQuantity: v.monthlyDataScannedInsightsGB,
		Details:         []string{"Insights queries", "data scanned", "Storage"},
		Usage:           true,
		Unit:            "GB",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonCloudWatch"),
			Family:   util.StringPtr("Data Payload"),
			Location: util.StringPtr(v.region.String()),
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
	}
}
