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

// KMSKey represents an SQS queue definition that can be cost-estimated.
type KMSKey struct {
	provider              *Provider
	region                region.Code
	customerMasterKeySpec string
}

type kmsKeyValues struct {
	CustomerMasterKeySpec string `mapstructure:"customer_master_key_spec"`
}

// decodeKMSKeyValues decodes and returns kmsKeyValues from a Terraform values map.
func decodeKMSKeyValues(tfVals map[string]interface{}) (kmsKeyValues, error) {
	var v kmsKeyValues
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

// newKMSKey creates a new KMSKey from kmsKeyValues.
func (p *Provider) newKMSKey(_ map[string]terraform.Resource, vals kmsKeyValues) *KMSKey {
	v := &KMSKey{
		provider:              p,
		region:                p.region,
		customerMasterKeySpec: "SYMMETRIC_DEFAULT",
	}

	if vals.CustomerMasterKeySpec != "" {
		v.customerMasterKeySpec = vals.CustomerMasterKeySpec
	}

	return v
}

// Components returns the price component queries that make up the KMSKey.
func (v *KMSKey) Components() []query.Component {
	components := []query.Component{v.kmsKeyComponent()}

	switch v.customerMasterKeySpec {
	case "RSA_2048":
		components = append(components, v.kmsKeyRequestComponent("Requests (RSA 2048)", ".*KMS-Requests-Asymmetric-RSA_2048$", ""))
	case
		"RSA_3072",
		"RSA_4096",
		"ECC_NIST_P256",
		"ECC_NIST_P384",
		"ECC_NIST_P521",
		"ECC_SECG_P256K1":
		components = append(components, v.kmsKeyRequestComponent("Requests (asymmetric)", ".*KMS-Requests-Asymmetric$", ""))
	default:
		components = append(components, v.kmsKeyRequestComponent("Requests", ".*KMS-Requests$", "API Request"))
		components = append(components, v.kmsKeyRequestComponent("ECC GenerateDataKeyPair requests", ".*KMS-Requests-GenerateDatakeyPair-ECC$", ""))
		components = append(components, v.kmsKeyRequestComponent("RSA GenerateDataKeyPair requests", ".*KMS-Requests-GenerateDatakeyPair-RSA$", ""))
	}

	return components
}

func (v *KMSKey) kmsKeyComponent() query.Component {
	return query.Component{
		Name:            "Customer master key",
		MonthlyQuantity: decimal.NewFromInt(1),
		Details:         []string{"master key"},
		Usage:           false,
		Unit:            "Keys",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("awskms"),
			Family:   util.StringPtr("Encryption Key"),
			Location: util.StringPtr(v.region.String()),
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
	}
}

func (v *KMSKey) kmsKeyRequestComponent(name string, usageType string, family string) query.Component {
	return query.Component{
		Name:            name,
		MonthlyQuantity: decimal.NewFromInt(1),
		Details:         []string{"Request"},
		Usage:           false,
		Unit:            "Requests",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("awskms"),
			Family:   util.StringPtr(family),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(usageType)},
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
