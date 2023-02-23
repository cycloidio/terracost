package terraform

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// Using instance struc
// type AutoscalingGroup struct {
// 	provider *Provider
// 	region   region.Code
//
// 	instanceType string
//
// 	desiredCapacity decimal.Decimal
//
// 	// tenancy describes the tenancy of an instance.
// 	// Valid values include: Shared, Dedicated, Host.
// 	// Note: only "Shared" and "Dedicated" are supported at the moment.
// 	tenancy string
// }

// autoscalingGroup represents the structure of Terraform values for autoscaling_group resource.
type autoscalingGroupValues struct {
	AvailabilityZone    string `mapstructure:"availability_zones"`
	LaunchConfiguration string `mapstructure:"launch_configuration"` // TODO set it to the right type

	LaunchTemplate []struct {
		ID      string `mapstructure:"id"` // TODO set it to the right type
		Version string `mapstructure:"version"`
	} `mapstructure:"launch_template"`
	MinSize         int64 `mapstructure:"min_size"`
	DesiredCapacity int64 `mapstructure:"desired_capacity"`
}

//
// // LaunchTemplate represents the structure of Terraform values for launch_template resource.
// type launchTemplateValues struct {
// 	InstanceType     string `mapstructure:"instance_type"`
// 	Tenancy          string `mapstructure:"tenancy"`
// 	AvailabilityZone string `mapstructure:"availability_zone"`
// 	EBSOptimized     string `mapstructure:"ebs_optimized"`
// 	EnableMonitoring bool   `mapstructure:"monitoring"`
//
// 	CreditSpecification []struct {
// 		CPUCredits string `mapstructure:"cpu_credits"`
// 	} `mapstructure:"credit_specification"`
// }

// decodeInstanceValues decodes and returns instanceValues from a Terraform values map.
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

		// TODO: need to get from LC/LT reference
		instanceType: "t3.large",
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
