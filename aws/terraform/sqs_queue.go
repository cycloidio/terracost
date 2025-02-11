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

// SQSQueue represents an SQS queue definition that can be cost-estimated.
type SQSQueue struct {
	provider *Provider
	region   region.Code

	fifoQueue bool

	// Usage
	monthlyRequests decimal.Decimal
	requestSizeKB   decimal.Decimal
}

type sqsQueueValues struct {
	FifoQueue bool `mapstructure:"fifo_queue"`

	Usage struct {
		MonthlyRequests float64 `mapstructure:"monthly_requests"`
		RequestSizeKB   float64 `mapstructure:"request_size_kb"`
	} `mapstructure:"tc_usage"`
}

// decodeSQSQueueValues decodes and returns sqsQueueValues from a Terraform values map.
func decodeSQSQueueValues(tfVals map[string]interface{}) (sqsQueueValues, error) {
	var v sqsQueueValues
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

// newSQSQueue creates a new SQSQueue from sqsQueueValues.
func (p *Provider) newSQSQueue(_ map[string]terraform.Resource, vals sqsQueueValues) *SQSQueue {
	v := &SQSQueue{
		provider:  p,
		region:    p.region,
		fifoQueue: vals.FifoQueue,

		// From Usage
		monthlyRequests: decimal.NewFromFloat(vals.Usage.MonthlyRequests),
		requestSizeKB:   decimal.NewFromFloat(vals.Usage.RequestSizeKB),
	}

	return v
}

// Components returns the price component queries that make up the SQSQueue.
func (v *SQSQueue) Components() []query.Component {
	components := []query.Component{v.sqsQueueComponent()}
	return components
}

func (v *SQSQueue) sqsQueueComponent() query.Component {
	// Requests-RBP for us or Requests-Tier1
	// Requests is no FIFO
	queueType := ".*Requests-[^F].*"
	if v.fifoQueue {
		queueType = ".*Requests-FIFO.*"
	}

	requests := v.requestSizeKB.Div(decimal.NewFromInt(64)).Ceil().Mul(v.monthlyRequests)

	return query.Component{
		Name:            fmt.Sprintf("Requests %s", queueType),
		MonthlyQuantity: requests,
		Details:         []string{"SQS queue", queueType},
		Usage:           true,
		Unit:            "Requests",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AWSQueueService"),
			Family:   util.StringPtr("API Request"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(queueType)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Requests"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}
