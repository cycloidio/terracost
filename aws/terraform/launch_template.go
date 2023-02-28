package terraform

import (
	"github.com/mitchellh/mapstructure"
)

// LaunchTemplate represents the structure of Terraform values for launch_template resource.
type launchTemplateValues struct {
	InstanceType string `mapstructure:"instance_type"`
	Placement    []struct {
		Tenancy          string `mapstructure:"yenancy"`
		AvailabilityZone string `mapstructure:"availability_zone"`
	} `mapstructure:"placement"`

	EBSOptimized bool `mapstructure:"ebs_optimized"`
	Monitoring   []struct {
		Enabled bool `mapstructure:"enabled"`
	} `mapstructure:"monitoring"`

	CreditSpecification []struct {
		CPUCredits string `mapstructure:"cpu_credits"`
	} `mapstructure:"credit_specification"`

	BlockDeviceMappings []struct {
		EBS []struct {
			VolumeType string  `mapstructure:"volume_type"`
			VolumeSize float64 `mapstructure:"volume_size"`
			IOPS       float64 `mapstructure:"iops"`
		} `mapstructure:"ebs"`
	} `mapstructure:"block_device_mappings"`
}

// decodeAutoscalingGroupValues decodes and returns instanceValues from a Terraform values map.
func decodeLaunchTemplateValues(tfVals map[string]interface{}) (launchTemplateValues, error) {
	var v launchTemplateValues
	if err := mapstructure.Decode(tfVals, &v); err != nil {
		return v, err
	}
	return v, nil
}
