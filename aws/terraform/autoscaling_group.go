package terraform

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// autoscalingGroup represents the structure of Terraform values for autoscaling_group resource.
type autoscalingGroupValues struct {
	AvailabilityZone    string `mapstructure:"availability_zones"`
	LaunchConfiguration string `mapstructure:"launch_configuration"`

	LaunchTemplate []struct {
		ID      string `mapstructure:"id"` // TODO set it to the right type
		Version string `mapstructure:"version"`
	} `mapstructure:"launch_template"`
	MinSize         int64 `mapstructure:"min_size"`
	DesiredCapacity int64 `mapstructure:"desired_capacity"`
}

// decodeAutoscalingGroupValues decodes and returns instanceValues from a Terraform values map.
func decodeAutoscalingGroupValues(tfVals map[string]interface{}) (autoscalingGroupValues, error) {
	var v autoscalingGroupValues
	if err := mapstructure.Decode(tfVals, &v); err != nil {
		return v, err
	}
	return v, nil
}

// newInstance creates a new Instance from instanceValues.
func (p *Provider) newAutoscalingGroup(vals autoscalingGroupValues) *Instance {
	inst := &Instance{
		provider: p,
		region:   p.region,
		tenancy:  "Shared",

		// Note: every Instance is estimated as a Linux without pre-installed S/W
		operatingSystem: "Linux",
		capacityStatus:  "Used",
		preInstalledSW:  "NA",
	}

	var instanceCount int64
	if vals.DesiredCapacity > 0 {
		instanceCount = vals.DesiredCapacity
	} else if vals.MinSize > 0 {
		instanceCount = vals.MinSize
	} else {
		instanceCount = 1
	}
	inst.instanceCount = decimal.NewFromInt(instanceCount)

	if vals.LaunchConfiguration != "" {
		inst.instanceType = "t3.2xlarge"
	}

	if len(vals.LaunchTemplate) > 0 {
		// lt := vals.LaunchTemplate[0]
		inst.instanceType = "t3.small"
	}

	// lcVals, err := decodeLaunchConfigurationValues(tfRes.Values)
	// if err != nil {
	// 	return nil
	// }
	// ltVals, err := decodeLaunchTemplateValues(tfRes.Values)
	// if err != nil {
	// 	return nil
	// }
	// TODO: need to get from LC/LT reference

	// TODO: need to come from LC/LT reference
	// if vals.Tenancy == "dedicated" {
	// inst.tenancy = "Dedicated"
	// }

	// volVals := volumeValues{AvailabilityZone: vals.AvailabilityZone}
	// // TODO: need to get from LC/LT reference
	// if len(vals.RootBlockDevice) > 0 {
	// 	rbd := vals.RootBlockDevice[0]
	// 	volVals.Type = rbd.VolumeType
	// 	volVals.Size = rbd.VolumeSize
	// 	volVals.IOPS = rbd.IOPS
	// }
	// inst.rootVolume = p.newVolume(volVals)

	// TODO: need to come from LC/LT reference
	// if vals.EBSOptimized {
	// 	inst.ebsOptimized = true
	// }

	// if len(vals.CreditSpecification) > 0 {
	// 	creditspec := vals.CreditSpecification[0]
	// 	if creditspec.CPUCredits == "unlimited" {
	// 		inst.cpuCredits = true
	// 	}
	// }
	//
	// if vals.EnableMonitoring {
	// 	inst.enableMonitoring = true
	// }
	return inst
}
