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

// SecretsmanagerSecret represents an SQS queue definition that can be cost-estimated.
type SecretsmanagerSecret struct {
	provider *Provider
	region   region.Code

	// Usage
	monthlyRequests decimal.Decimal
}

type secretsmanagerSecretValues struct {
	Usage struct {
		MonthlyRequests float64 `mapstructure:"monthly_requests"`
	} `mapstructure:"tc_usage"`
}

// decodeSecretsmanagerSecretValues decodes and returns secretsmanagerSecretValues from a Terraform values map.
func decodeSecretsmanagerSecretValues(tfVals map[string]interface{}) (secretsmanagerSecretValues, error) {
	var v secretsmanagerSecretValues
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

// newSecretsmanagerSecret creates a new SecretsmanagerSecret from secretsmanagerSecretValues.
func (p *Provider) newSecretsmanagerSecret(_ map[string]terraform.Resource, vals secretsmanagerSecretValues) *SecretsmanagerSecret {
	v := &SecretsmanagerSecret{
		provider: p,
		region:   p.region,

		// From Usage
		monthlyRequests: decimal.NewFromFloat(vals.Usage.MonthlyRequests),
	}

	return v
}

// Components returns the price component queries that make up the SecretsmanagerSecret.
func (v *SecretsmanagerSecret) Components() []query.Component {
	components := []query.Component{v.secretsmanagerSecretComponent()}
	components = append(components, v.secretsmanagerSecretRequestsComponent())
	return components
}

func (v *SecretsmanagerSecret) secretsmanagerSecretComponent() query.Component {
	return query.Component{
		Name:            "Secret",
		MonthlyQuantity: decimal.NewFromInt(1),
		Details:         []string{"Secret"},
		Usage:           true,
		Unit:            "Secrets",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AWSSecretsManager"),
			Family:   util.StringPtr("Secret"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(".*-AWSSecretsManager-Secrets")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Secrets"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}

func (v *SecretsmanagerSecret) secretsmanagerSecretRequestsComponent() query.Component {
	return query.Component{
		Name:            "API Request",
		MonthlyQuantity: v.monthlyRequests,
		Details:         []string{"API Request"},
		Usage:           true,
		Unit:            "API Requests",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AWSSecretsManager"),
			Family:   util.StringPtr("API Request"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(".*-AWSSecretsManager[-]?APIRequest[s]?")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("API Requests"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}
