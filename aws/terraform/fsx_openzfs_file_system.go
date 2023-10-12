package terraform

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/terraform"
)

// fsxOpenzfsFileSystemValues represents the structure of Terraform values for aws_efs_file_system resource.
type fsxOpenzfsFileSystemValues struct {
	StorageCapacity              float64 `mapstructure:"storage_capacity"`
	StorageType                  string  `mapstructure:"storage_type"`
	DeploymentType               string  `mapstructure:"deployment_type"`
	ThroughputCapacity           float64 `mapstructure:"throughput_capacity"`
	AutomaticBackupRetentionDays float64 `mapstructure:"automatic_backup_retention_days"`

	DiskIopsConfiguration []struct {
		IOPS     float64 `mapstructure:"iops"`
		IOPSMode string  `mapstructure:"mode"`
	} `mapstructure:"disk_iops_configuration"`

	DataCompressionType string `mapstructure:"data_compression_type"`

	Usage struct {
		BackupStorageGB float64 `mapstructure:"backup_storage_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeFSxOpenzfsFileSystemValues decodes and returns fsxOpenzfsFileSystemValues from a Terraform values map.
func decodeFSxOpenzfsFileSystemValues(tfVals map[string]interface{}) (fsxOpenzfsFileSystemValues, error) {
	var v fsxOpenzfsFileSystemValues
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

func (v *FSxFileSystem) getOpenzfsDeployOption(deploymentType string) string {

	deploymentOption := "Multi-AZ"
	switch strings.ToLower(deploymentType) {
	case "single_az_1":
		deploymentOption = "Single-AZ"
	case "single_az_2":
		deploymentOption = "Single-AZ_2"
	}

	return deploymentOption
}

// newFSxOpenzfsFileSystem creates a new FSxOpenzfsFileSystem from fsxOpenzfsFileSystemValues.
func (p *Provider) newFSxOpenzfsFileSystem(rss map[string]terraform.Resource, vals fsxOpenzfsFileSystemValues) *FSxFileSystem {
	v := &FSxFileSystem{
		provider:           p,
		region:             p.region,
		storageType:        "SSD",
		storageCapacity:    decimal.NewFromFloat(32),
		fsxType:            "OpenZFS",
		throughputCapacity: decimal.NewFromFloat(vals.ThroughputCapacity),
		deploymentOption:   "Single-AZ",
		// From Usage
		backupStorage: decimal.NewFromFloat(vals.Usage.BackupStorageGB),
	}

	if vals.StorageCapacity > 0 {
		v.storageCapacity = decimal.NewFromFloat(vals.StorageCapacity)
	}

	if len(vals.DeploymentType) > 0 {
		v.deploymentOption = v.getOpenzfsDeployOption(vals.DeploymentType)
	}

	if len(vals.StorageType) > 0 {
		v.storageType = vals.StorageType
	}

	if vals.AutomaticBackupRetentionDays > 0 {
		v.automaticBackupRetentionDays = decimal.NewFromFloat(vals.AutomaticBackupRetentionDays)
	}

	return v
}
