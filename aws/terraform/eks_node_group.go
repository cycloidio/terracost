package terraform

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/terraform"
)

// eksNodeGroup represents the structure of Terraform values for autoscaling_group resource.
type eksNodeGroupValues struct {
	ScalingConfig []struct {
		MinSize         int64 `mapstructure:"min_size"`
		DesiredCapacity int64 `mapstructure:"desired_size"`
	} `mapstructure:"scaling_config"`

	InstanceTypes []string `mapstructure:"instance_types"`
	DiskSize      float64  `mapstructure:"disk_size"`

	LaunchTemplate []struct {
		ID      string `mapstructure:"id"`
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"launch_template"`
}

// decodeEKSNodeGroupValues decodes and returns instanceValues from a Terraform values map.
func decodeEKSNodeGroupValues(tfVals map[string]interface{}) (eksNodeGroupValues, error) {
	var v eksNodeGroupValues
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

// newInstance creates a new Instance from instanceValues.
func (p *Provider) newEKSNodeGroup(rss map[string]terraform.Resource, vals eksNodeGroupValues) *Instance {

	inst := &Instance{
		provider: p,
		region:   p.region,
		tenancy:  "Shared",

		// Note: every Instance is estimated as a Linux without pre-installed S/W
		operatingSystem: "Linux",
		capacityStatus:  "Used",
		preInstalledSW:  "NA",
	}

	var defaultEKSInstanceType = "t3.medium"
	var availabilityZone = fmt.Sprintf("%s%s", p.region, "a")
	var instanceCount int64
	if len(vals.ScalingConfig) > 0 {
		if vals.ScalingConfig[0].DesiredCapacity > 0 {
			instanceCount = vals.ScalingConfig[0].DesiredCapacity
		} else if vals.ScalingConfig[0].MinSize > 0 {
			instanceCount = vals.ScalingConfig[0].MinSize
		} else {
			instanceCount = 1
		}
	}
	inst.instanceCount = decimal.NewFromInt(instanceCount)

	if len(vals.LaunchTemplate) > 0 {
		// If LT defined
		var ltref string
		if len(vals.LaunchTemplate) > 0 {
			if len(vals.LaunchTemplate[0].ID) > 0 {
				ltref = vals.LaunchTemplate[0].ID
			} else {
				ltref = vals.LaunchTemplate[0].Name
			}
		}

		lt, err := decodeLaunchTemplateValues(rss[ltref].Values)
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
	} else {
		// If no LT
		if len(vals.InstanceTypes) > 0 {
			inst.instanceType = vals.InstanceTypes[0]
		} else {
			inst.instanceType = defaultEKSInstanceType
		}

		// Set the default Linux size
		disksize := float64(20)
		if vals.DiskSize > 0 {
			disksize = vals.DiskSize
		}

		volVals := volumeValues{AvailabilityZone: availabilityZone}
		volVals.Type = "gp3"
		volVals.Size = disksize
		inst.rootVolume = p.newVolume(volVals)
	}

	// Override provider region by the one defined in LC/LT
	if reg := region.NewFromZone(availabilityZone); reg.Valid() {
		inst.region = reg
	}

	return inst
}
