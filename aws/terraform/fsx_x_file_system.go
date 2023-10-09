package terraform

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
)

// FSxFileSystem represents an EFS that can be cost-estimated.
type FSxFileSystem struct {
	provider *Provider
	region   region.Code

	storageCapacity              decimal.Decimal
	storageType                  string
	deploymentType               string
	throughputCapacity           decimal.Decimal
	automaticBackupRetentionDays decimal.Decimal

	// Used to defined if Windows/Lustre/Openzfs
	fsxType          string
	deploymentOption string

	// Usage
	backupStorage decimal.Decimal
}

// Components returns the price component queries that make up the FSxFileSystem.
func (v *FSxFileSystem) Components() []query.Component {
	components := []query.Component{v.fsxFileSystemStorageCapacityCostComponent()}

	if v.fsxType != "Lustre" {
		components = append(components, v.fsxFileSystemThroughputCapacityCostComponent())
	}

	if v.automaticBackupRetentionDays.GreaterThan(decimal.NewFromInt(0)) {
		components = append(components, v.fsxFileSystemBackupGBCostComponent())
	}

	return components
}

func (v *FSxFileSystem) fsxFileSystemThroughputCapacityCostComponent() query.Component {
	return query.Component{
		Name:            "Throughput capacity",
		MonthlyQuantity: v.throughputCapacity,
		Unit:            "MiBps-Mo",
		Details:         []string{"Throughput capacity", v.fsxType},
		Usage:           false,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonFSx"),
			Family:   util.StringPtr("Provisioned Throughput"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "Deployment_option", Value: util.StringPtr(v.deploymentOption)},
				{Key: "FileSystemType", Value: util.StringPtr(v.fsxType)},
			},
		},
	}
}

func (v *FSxFileSystem) fsxFileSystemStorageCapacityCostComponent() query.Component {

	attrFilters := []*product.AttributeFilter{
		{Key: "Deployment_option", Value: util.StringPtr(v.deploymentOption)},
		{Key: "FileSystemType", Value: util.StringPtr(v.fsxType)},
		{Key: "StorageType", Value: util.StringPtr(v.storageType)},
	}

	if v.fsxType == "Lustre" {
		f := &product.AttributeFilter{Key: "ThroughputCapacity", Value: util.StringPtr(v.throughputCapacity.String())}
		attrFilters = append(attrFilters, f)
	}

	return query.Component{
		Name:            fmt.Sprintf("%s Storage %s", v.fsxType, v.storageType),
		MonthlyQuantity: v.storageCapacity,
		Unit:            "GB-Mo",
		Details:         []string{"Storage", v.fsxType},
		Usage:           false,
		ProductFilter: &product.Filter{
			Provider:         util.StringPtr(v.provider.key),
			Service:          util.StringPtr("AmazonFSx"),
			Family:           util.StringPtr("Storage"),
			Location:         util.StringPtr(v.region.String()),
			AttributeFilters: attrFilters,
		},
	}
}

func (v *FSxFileSystem) fsxFileSystemBackupGBCostComponent() query.Component {
	deploymentOption := v.deploymentOption
	if v.fsxType == "ONTAP" {
		deploymentOption = "N/A"
	}
	if v.fsxType == "OpenZFS" {
		if v.deploymentOption == "Multi-AZ" {
			deploymentOption = "Multi-AZ"
		} else {
			deploymentOption = "N/A"
		}
	}
	if v.fsxType == "Windows" {
		if v.deploymentOption == "Multi-AZ" {
			deploymentOption = "Multi-AZ"
		} else {
			deploymentOption = "Single-AZ"
		}
	}

	return query.Component{
		Name:            fmt.Sprintf("%s Backup storage", v.fsxType),
		MonthlyQuantity: v.storageCapacity,
		Unit:            "GB-Mo",
		Details:         []string{"Storage", v.fsxType},
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonFSx"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "Deployment_option", Value: util.StringPtr(deploymentOption)},
				{Key: "FileSystemType", Value: util.StringPtr(v.fsxType)},
				{Key: "UsageType", ValueRegex: util.StringPtr(".*-BackupUsage")},
			},
		},
	}
}
