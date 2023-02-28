package terraform

import (
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/terraform"
)

// autoscalingGroup represents the structure of Terraform values for autoscaling_group resource.
type autoscalingGroupValues struct {
	AvailabilityZones []string `mapstructure:"availability_zones"`

	LaunchConfiguration []string `mapstructure:"launch_configuration"`

	LaunchTemplate []struct {
		ID      []string `mapstructure:"id"`
		Version string   `mapstructure:"version"`
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
func (p *Provider) newAutoscalingGroup(rss map[string]terraform.Resource, vals autoscalingGroupValues) *Instance {

	inst := &Instance{
		provider: p,
		region:   p.region,
		tenancy:  "Shared",

		// Note: every Instance is estimated as a Linux without pre-installed S/W
		operatingSystem: "Linux",
		capacityStatus:  "Used",
		preInstalledSW:  "NA",
	}

	// TODO: fix vals.AvailabilityZone which is always empty
	var availabilityZone string
	if len(vals.AvailabilityZones) > 0 {
		availabilityZone = vals.AvailabilityZones[0]
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

	// ASG use LaunchConfiguration
	if len(vals.LaunchConfiguration) > 0 {
		lc, err := decodeLaunchConfigurationValues(rss[vals.LaunchConfiguration[0]].Values)
		if err != nil {
			return inst
		}
		inst.instanceType = lc.InstanceType
		if lc.PlacementTenancy == "dedicated" {
			inst.tenancy = "Dedicated"
		}

		if lc.EBSOptimized {
			inst.ebsOptimized = true
		}

		if lc.EnableMonitoring {
			inst.enableMonitoring = true
		}

		if len(lc.RootBlockDevice) > 0 {
			rbd := lc.RootBlockDevice[0]
			volVals := volumeValues{AvailabilityZone: availabilityZone}
			volVals.Type = rbd.VolumeType
			volVals.Size = rbd.VolumeSize
			volVals.IOPS = rbd.IOPS
			inst.rootVolume = p.newVolume(volVals)
		}
	}

	// ASG use LaunchTemplate
	// or
	// ASG use mixed Instance Launch Template
	// if len(vals.MixedInstanceLaunchTemplate) > 0 {
	// lt, err := decodeMixedInstancesLaunchTemplateValues(rss[vals.MixedInstanceLaunchTemplate[0]].Values)
	// }

	if len(vals.LaunchTemplate) > 0 {
		lt, err := decodeLaunchTemplateValues(rss[vals.LaunchTemplate[0].ID[0]].Values)
		if err != nil {
			return inst
		}

		inst.instanceType = lt.InstanceType
		if len(lt.Placement) > 0 {
			if lt.Placement[0].Tenancy == "dedicated" {
				inst.tenancy = "Dedicated"
			}
			if lt.Placement[0].AvailabilityZone != "" {
				availabilityZone = lt.Placement[0].AvailabilityZone
			}
		}

		if lt.EBSOptimized {
			inst.ebsOptimized = true
		}

		if len(lt.CreditSpecification) > 0 {
			creditspec := lt.CreditSpecification[0]
			if creditspec.CPUCredits == "unlimited" {
				inst.cpuCredits = true
			}
		}

		if len(lt.Monitoring) > 0 {
			monitoring := lt.Monitoring[0]
			if monitoring.Enabled {
				inst.enableMonitoring = true
			}
		}

		// We assume the first EBS defined correspond to the RootDevice
		if len(lt.BlockDeviceMappings) > 0 {
			if len(lt.BlockDeviceMappings[0].EBS) > 0 {
				rbd := lt.BlockDeviceMappings[0].EBS[0]
				volVals := volumeValues{AvailabilityZone: availabilityZone}
				volVals.Type = rbd.VolumeType
				volVals.Size = rbd.VolumeSize
				volVals.IOPS = rbd.IOPS
				inst.rootVolume = p.newVolume(volVals)
			}
		}
	}

	// Override provider region by the one defined in LC/LT
	if reg := region.NewFromZone(availabilityZone); reg.Valid() {
		inst.region = reg
	}

	return inst
}
