package terraform

import (
	"github.com/mitchellh/mapstructure"
)

// LaunchConfiguration represents the structure of Terraform values for launch_template resource.
type launchConfigurationValues struct {
	InstanceType     string `mapstructure:"instance_type"`
	PlacementTenancy string `mapstructure:"placement_tenancy"`

	EBSOptimized     bool `mapstructure:"ebs_optimized"`
	EnableMonitoring bool `mapstructure:"enable_monitoring"`

	RootBlockDevice []struct {
		VolumeType string  `mapstructure:"volume_type"`
		VolumeSize float64 `mapstructure:"volume_size"`
		IOPS       float64 `mapstructure:"iops"`
	} `mapstructure:"root_block_device"`
}

// decodeAutoscalingGroupValues decodes and returns instanceValues from a Terraform values map.
func decodeLaunchConfigurationValues(tfVals map[string]interface{}) (launchConfigurationValues, error) {
	var v launchConfigurationValues
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
