package terraform

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/terraform"
)

// fsxLustreFileSystemValues represents the structure of Terraform values for aws_efs_file_system resource.
type fsxLustreFileSystemValues struct {
	StorageCapacity              float64  `mapstructure:"storage_capacity"`
	StorageType                  string   `mapstructure:"storage_type"`
	DeploymentType               string   `mapstructure:"deployment_type"`
	PerUnitStorageThroughput     float64  `mapstructure:"per_unit_storage_throughput"`
	SubnetIds                    []string `mapstructure:"subnet_ids"`
	AutomaticBackupRetentionDays float64  `mapstructure:"automatic_backup_retention_days"`

	Usage struct {
		BackupStorageGB float64 `mapstructure:"backup_storage_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeFSxLustreFileSystemValues decodes and returns fsxLustreFileSystemValues from a Terraform values map.
func decodeFSxLustreFileSystemValues(tfVals map[string]interface{}) (fsxLustreFileSystemValues, error) {
	var v fsxLustreFileSystemValues
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

// newFSxLustreFileSystem creates a new FSxLustreFileSystem from fsxLustreFileSystemValues.
func (p *Provider) newFSxLustreFileSystem(rss map[string]terraform.Resource, vals fsxLustreFileSystemValues) *FSxFileSystem {
	v := &FSxFileSystem{
		provider:           p,
		region:             p.region,
		storageType:        "SSD",
		storageCapacity:    decimal.NewFromFloat(1200),
		fsxType:            "Lustre",
		throughputCapacity: decimal.NewFromFloat(0),
		deploymentOption:   "Persistent",
		// From Usage
		backupStorage: decimal.NewFromFloat(vals.Usage.BackupStorageGB),
	}

	if vals.StorageCapacity > 0 {
		v.storageCapacity = decimal.NewFromFloat(vals.StorageCapacity)
	}

	var deploymentType string
	if len(vals.DeploymentType) > 0 {
		deploymentType = vals.DeploymentType
	} else {
		deploymentType = "PERSISTENT_1"
	}

	if vals.PerUnitStorageThroughput > 0 {
		v.throughputCapacity = decimal.NewFromFloat(vals.PerUnitStorageThroughput)
	} else {
		if v.storageType == "SSD" {
			if deploymentType == "PERSISTENT_1" {
				v.throughputCapacity = decimal.NewFromFloat(50)
			} else {
				v.throughputCapacity = decimal.NewFromFloat(125)
			}
		} else if v.storageType == "HDD" {
			if deploymentType == "PERSISTENT_1" {
				v.throughputCapacity = decimal.NewFromFloat(12)
			} else {
				v.throughputCapacity = decimal.NewFromFloat(0)
			}
		}
	}

	if len(vals.StorageType) > 0 {
		v.storageType = vals.StorageType
	}

	if vals.AutomaticBackupRetentionDays > 0 {
		v.automaticBackupRetentionDays = decimal.NewFromFloat(vals.AutomaticBackupRetentionDays)
	}

	return v
}
