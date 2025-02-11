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

// S3BucketInventory represents an SQS queue definition that can be cost-estimated.
type S3BucketInventory struct {
	provider *Provider
	region   region.Code

	// Usage
	monthlyListedObjects decimal.Decimal
}

type s3BucketInventoryValues struct {
	Usage struct {
		MonthlyListedObjects float64 `mapstructure:"monthly_listed_objects"`
	} `mapstructure:"tc_usage"`
}

// decodeS3BucketInventoryValues decodes and returns s3BucketInventoryValues from a Terraform values map.
func decodeS3BucketInventoryValues(tfVals map[string]interface{}) (s3BucketInventoryValues, error) {
	var v s3BucketInventoryValues
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

// newS3BucketInventory creates a new S3BucketInventory from s3BucketInventoryValues.
func (p *Provider) newS3BucketInventory(_ map[string]terraform.Resource, vals s3BucketInventoryValues) *S3BucketInventory {
	v := &S3BucketInventory{
		provider: p,
		region:   p.region,

		// Usage
		monthlyListedObjects: decimal.NewFromFloat(vals.Usage.MonthlyListedObjects),
	}

	return v
}

// Components returns the price component queries that make up the S3BucketInventory.
func (v *S3BucketInventory) Components() []query.Component {
	components := []query.Component{v.s3BucketInventoryComponent()}
	return components
}

func (v *S3BucketInventory) s3BucketInventoryComponent() query.Component {
	return query.Component{
		Name:            "Objects listed",
		MonthlyQuantity: v.monthlyListedObjects,
		Details:         []string{"Listed"},
		Usage:           true,
		Unit:            "Objects",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonS3"),
			Family:   util.StringPtr("Fee"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(".*Inventory-ObjectsListed$")},
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
