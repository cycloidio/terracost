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

// S3BucketAnalyticsConfiguration represents an SQS queue definition that can be cost-estimated.
type S3BucketAnalyticsConfiguration struct {
	provider *Provider
	region   region.Code

	// Usage
	monthlyMonitoredObjects decimal.Decimal
}

type s3BucketAnalyticsConfigurationValues struct {
	Usage struct {
		MonthlyMonitoredObjects float64 `mapstructure:"monthly_monitored_objects"`
	} `mapstructure:"tc_usage"`
}

// decodeS3BucketAnalyticsConfigurationValues decodes and returns s3BucketAnalyticsConfigurationValues from a Terraform values map.
func decodeS3BucketAnalyticsConfigurationValues(tfVals map[string]interface{}) (s3BucketAnalyticsConfigurationValues, error) {
	var v s3BucketAnalyticsConfigurationValues
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

// newS3BucketAnalyticsConfiguration creates a new S3BucketAnalyticsConfiguration from s3BucketAnalyticsConfigurationValues.
func (p *Provider) newS3BucketAnalyticsConfiguration(rss map[string]terraform.Resource, vals s3BucketAnalyticsConfigurationValues) *S3BucketAnalyticsConfiguration {
	// The 'rss' variable contains information from linked resources.
	// Currently, it is not utilized in this resource.
	_ = rss

	v := &S3BucketAnalyticsConfiguration{
		provider: p,
		region:   p.region,

		// Usage
		monthlyMonitoredObjects: decimal.NewFromFloat(vals.Usage.MonthlyMonitoredObjects),
	}

	return v
}

// Components returns the price component queries that make up the S3BucketAnalyticsConfiguration.
func (v *S3BucketAnalyticsConfiguration) Components() []query.Component {
	components := []query.Component{v.s3BucketAnalyticsConfigurationComponent()}
	return components
}

func (v *S3BucketAnalyticsConfiguration) s3BucketAnalyticsConfigurationComponent() query.Component {
	return query.Component{
		Name:            "Objects monitored",
		MonthlyQuantity: v.monthlyMonitoredObjects,
		Details:         []string{"Monitored"},
		Usage:           true,
		Unit:            "Objects",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonS3"),
			Family:   util.StringPtr("Fee"),
			Location: util.StringPtr(v.region.String()),
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
	}
}
