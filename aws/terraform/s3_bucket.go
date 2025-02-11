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

// S3Bucket represents an SQS queue definition that can be cost-estimated.
type S3Bucket struct {
	provider *Provider
	region   region.Code

	// Usage
	monthlyOutboundDataGB decimal.Decimal
	storageGB             decimal.Decimal
}

type s3BucketValues struct {

	// Usage
	Usage struct {
		MonthlyOutboundDataGB float64 `mapstructure:"monthly_outbound_data_gb"`
		StorageGB             float64 `mapstructure:"storage_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeS3BucketValues decodes and returns S3BucketValues from a Terraform values map.
func decodeS3BucketValues(tfVals map[string]interface{}) (s3BucketValues, error) {
	var v s3BucketValues
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

// newS3Bucket creates a new S3Bucket from s3BucketValues.
func (p *Provider) newS3Bucket(_ map[string]terraform.Resource, vals s3BucketValues) *S3Bucket {
	v := &S3Bucket{
		provider: p,
		region:   p.region,

		// From Usage
		monthlyOutboundDataGB: decimal.NewFromFloat(vals.Usage.MonthlyOutboundDataGB),
		storageGB:             decimal.NewFromFloat(vals.Usage.StorageGB),
	}

	return v
}

// Components returns the price component queries that make up the S3Bucket.
func (v *S3Bucket) Components() []query.Component {
	components := []query.Component{}

	if v.storageGB.GreaterThan(decimal.NewFromInt(512000)) {
		extraSize := v.storageGB.Sub(decimal.NewFromInt(512000))
		components = append(components, v.S3BucketComponent("0", decimal.NewFromInt(51200)))
		components = append(components, v.S3BucketComponent("51200", decimal.NewFromInt(460800)))
		components = append(components, v.S3BucketComponent("512000", extraSize))

	} else if v.storageGB.GreaterThan(decimal.NewFromInt(51200)) {
		extraSize := v.storageGB.Sub(decimal.NewFromInt(51200))
		components = append(components, v.S3BucketComponent("0", decimal.NewFromInt(51200)))
		components = append(components, v.S3BucketComponent("51200", extraSize))
	} else {
		components = append(components, v.S3BucketComponent("0", v.storageGB))
	}

	if v.monthlyOutboundDataGB.GreaterThan(decimal.NewFromInt(153600)) {
		extraOut := v.monthlyOutboundDataGB.Sub(decimal.NewFromInt(153600))
		components = append(components, v.S3BucketOutboundDataTransferComponent("0", decimal.NewFromInt(10240)))
		components = append(components, v.S3BucketOutboundDataTransferComponent("10240", decimal.NewFromInt(40960)))
		components = append(components, v.S3BucketOutboundDataTransferComponent("51200", decimal.NewFromInt(102400)))
		components = append(components, v.S3BucketOutboundDataTransferComponent("153600", extraOut))
	} else if v.monthlyOutboundDataGB.GreaterThan(decimal.NewFromInt(51200)) {
		extraOut := v.monthlyOutboundDataGB.Sub(decimal.NewFromInt(51200))
		components = append(components, v.S3BucketOutboundDataTransferComponent("0", decimal.NewFromInt(51200)))
		components = append(components, v.S3BucketOutboundDataTransferComponent("10240", decimal.NewFromInt(40960)))
		components = append(components, v.S3BucketOutboundDataTransferComponent("51200", extraOut))
	} else if v.monthlyOutboundDataGB.GreaterThan(decimal.NewFromInt(10240)) {
		extraOut := v.monthlyOutboundDataGB.Sub(decimal.NewFromInt(10240))
		components = append(components, v.S3BucketOutboundDataTransferComponent("0", decimal.NewFromInt(51200)))
		components = append(components, v.S3BucketOutboundDataTransferComponent("10240", extraOut))
	} else {
		components = append(components, v.S3BucketOutboundDataTransferComponent("0", v.monthlyOutboundDataGB))
	}

	return components
}

func (v *S3Bucket) S3BucketComponent(startingRange string, storage decimal.Decimal) query.Component {
	return query.Component{
		Name:            fmt.Sprintf("Storage %s", startingRange),
		MonthlyQuantity: storage,
		Details:         []string{"Standard"},
		Usage:           true,
		Unit:            "GB-Mo",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonS3"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(".*TimedStorage-ByteHrs$")},
				{Key: "VolumeType", Value: util.StringPtr("Standard")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("GB-Mo"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr(startingRange)},
			},
		},
	}
}

func (v *S3Bucket) S3BucketOutboundDataTransferComponent(startingRange string, outboundGB decimal.Decimal) query.Component {
	shortRegion := region.GetRegionToShortName(v.region.String())
	usageType := "DataTransfer-Out-Bytes"
	// us-east-1 is a special case where no shortRegion should be used
	if shortRegion != "" && shortRegion != "us-east-1" {
		usageType = fmt.Sprintf("%s-DataTransfer-Out-Bytes", shortRegion)
	}

	return query.Component{
		Name:            fmt.Sprintf("Outbound Data Transfer %s", startingRange),
		MonthlyQuantity: outboundGB,
		Details:         []string{"Outbound"},
		Usage:           true,
		Unit:            "GB",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AWSDataTransfer"),
			Family:   util.StringPtr("Data Transfer"),
			Location: util.StringPtr(""),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", Value: util.StringPtr(usageType)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("GB"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr(startingRange)},
			},
		},
	}
}
