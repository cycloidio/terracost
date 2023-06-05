package terraform

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

// EFSFileSystem represents an EFS that can be cost-estimated.
type EFSFileSystem struct {
	provider                     *Provider
	region                       region.Code
	availabilityZoneName         string
	throughputMode               string
	provisionedThroughputInMibps decimal.Decimal

	// Usage
	hasLifecyclePolicy             bool
	storageGB                      decimal.Decimal
	infrequentAccessStorageGB      decimal.Decimal
	monthlyInfrequentAccessReadGB  decimal.Decimal
	monthlyInfrequentAccessWriteGB decimal.Decimal
}

// efsFileSystemValues represents the structure of Terraform values for aws_efs_file_system resource.
type efsFileSystemValues struct {
	AvailabilityZoneName string `mapstructure:"availability_zone_name"`
	LifecyclePolicy      []struct {
		TransitionToIa                  string `mapstructure:"transition_to_ia"`
		TransitionToPrimaryStorageClass string `mapstructure:"transition_to_primary_storage_class"`
	} `mapstructure:"lifecycle_policy"`
	ThroughputMode string `mapstructure:"throughput_mode"`
	// only available if ThroughputMode=provisioned
	ProvisionedThroughputInMibps float64 `mapstructure:"provisioned_throughput_in_mibps"`

	Usage struct {
		StorageGB                      float64 `mapstructure:"storage_gb"`
		InfrequentAccessStorageGB      float64 `mapstructure:"infrequent_access_storage_gb"`
		MonthlyInfrequentAccessReadGB  float64 `mapstructure:"monthly_infrequent_access_read_gb"`
		MonthlyInfrequentAccessWriteGB float64 `mapstructure:"monthly_infrequent_access_write_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeEFSFileSystemValues decodes and returns efsFileSystemValues from a Terraform values map.
func decodeEFSFileSystemValues(tfVals map[string]interface{}) (efsFileSystemValues, error) {
	var v efsFileSystemValues
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

// newEFSFileSystem creates a new EFSFileSystem from efsFileSystemValues.
func (p *Provider) newEFSFileSystem(rss map[string]terraform.Resource, vals efsFileSystemValues) *EFSFileSystem {
	v := &EFSFileSystem{
		provider:       p,
		region:         p.region,
		throughputMode: "bursting",
		// only available if ThroughputMode=provisioned
		provisionedThroughputInMibps: decimal.NewFromFloat(0),
		hasLifecyclePolicy:           false,
		// From Usage
		storageGB:                      decimal.NewFromFloat(vals.Usage.StorageGB),
		infrequentAccessStorageGB:      decimal.NewFromFloat(vals.Usage.InfrequentAccessStorageGB),
		monthlyInfrequentAccessReadGB:  decimal.NewFromFloat(vals.Usage.MonthlyInfrequentAccessReadGB),
		monthlyInfrequentAccessWriteGB: decimal.NewFromFloat(vals.Usage.MonthlyInfrequentAccessWriteGB),
	}

	if reg := region.NewFromZone(vals.AvailabilityZoneName); reg.Valid() {
		v.region = reg
		v.availabilityZoneName = vals.AvailabilityZoneName
	}

	if len(vals.LifecyclePolicy) > 0 {
		v.hasLifecyclePolicy = true
	}

	if vals.ThroughputMode != "" {
		v.throughputMode = vals.ThroughputMode
	}

	if vals.ThroughputMode == "provisioned" {
		if vals.ProvisionedThroughputInMibps > 0 {
			v.provisionedThroughputInMibps = v.calculateProvisionedThroughput(v.storageGB, decimal.NewFromFloat(vals.ProvisionedThroughputInMibps))
		}
	}

	return v
}

func (v *EFSFileSystem) calculateProvisionedThroughput(storageGB decimal.Decimal, throughput decimal.Decimal) decimal.Decimal {
	defaultThroughput := storageGB.Mul(decimal.NewFromInt(730).Div(decimal.NewFromInt(20).Mul(decimal.NewFromInt(1))))
	totalProvisionedThroughput := throughput.Mul(decimal.NewFromInt(730))
	totalBillableProvisionedThroughput := totalProvisionedThroughput.Sub(defaultThroughput).Div(decimal.NewFromInt(730))

	if totalBillableProvisionedThroughput.IsPositive() {
		return totalBillableProvisionedThroughput
	}

	return decimal.Zero
}

// Components returns the price component queries that make up the EFSFileSystem.
func (v *EFSFileSystem) Components() []query.Component {
	usagetype := ".*-TimedStorage-ByteHrs"
	if v.availabilityZoneName != "" {
		usagetype = ".*-TimedStorage-Z-ByteHrs"
	}

	components := []query.Component{v.efsFileSystemComponent(usagetype, v.storageGB)}

	if v.provisionedThroughputInMibps.GreaterThan(decimal.NewFromInt(0)) {
		components = append(components, v.provisionedThroughputComponent())
	}

	if v.hasLifecyclePolicy {
		usagetype = ".*-IATimedStorage-ByteHrs"
		if v.availabilityZoneName != "" {
			usagetype = ".*-IATimedStorage-Z-ByteHrs"
		}

		if v.infrequentAccessStorageGB.GreaterThan(decimal.NewFromInt(0)) {
			components = append(components, v.efsFileSystemComponent(usagetype, v.infrequentAccessStorageGB))
		}

		if v.monthlyInfrequentAccessReadGB.GreaterThan(decimal.NewFromInt(0)) {
			components = append(components, v.requestsComponent("Read"))
		}

		if v.monthlyInfrequentAccessWriteGB.GreaterThan(decimal.NewFromInt(0)) {
			components = append(components, v.requestsComponent("Write"))
		}

	}

	return components
}

func (v *EFSFileSystem) efsFileSystemComponent(usagetype string, storageGB decimal.Decimal) query.Component {
	return query.Component{
		Name:            fmt.Sprintf("Storage %s", usagetype),
		MonthlyQuantity: storageGB,
		Unit:            "GB",
		Details:         []string{"EFS storage", usagetype},
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonEFS"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(fmt.Sprintf("%s", usagetype))},
			},
		},
	}
}

func (v *EFSFileSystem) provisionedThroughputComponent() query.Component {
	return query.Component{
		Name:            "Provisioned throughput",
		MonthlyQuantity: v.provisionedThroughputInMibps,
		Unit:            "MBps",
		Details:         []string{"Througput"},
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonEFS"),
			Family:   util.StringPtr("Provisioned Throughput"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr("ProvisionedTP-MiBpsHrs")},
			},
		},
	}
}

func (v *EFSFileSystem) requestsComponent(accessType string) query.Component {
	var requestsGB decimal.Decimal
	if accessType == "Read" {
		requestsGB = v.monthlyInfrequentAccessReadGB
	} else {
		requestsGB = v.monthlyInfrequentAccessWriteGB
	}

	return query.Component{
		Name:            fmt.Sprintf("Requests %s", accessType),
		MonthlyQuantity: requestsGB,
		Unit:            "GB",
		Details:         []string{"Requests", "Infrequent Access", accessType},
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonEFS"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "AccessType", Value: util.StringPtr(accessType)},
				{Key: "StorageClass", Value: util.StringPtr("Infrequent Access")},
			},
		},
	}
}
