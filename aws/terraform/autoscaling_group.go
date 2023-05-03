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

	LaunchConfiguration string `mapstructure:"launch_configuration"`

	LaunchTemplate []struct {
		ID      string `mapstructure:"id"`
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"launch_template"`

	MixedInstancesPolicy []struct {
		LaunchTemplate []struct {
			LaunchTemplateSpecification []struct {
				LaunchTemplateID   string `mapstructure:"launch_template_id"`
				LaunchTemplateName string `mapstructure:"launch_template_name"`
			} `mapstructure:"launch_template_specification"`

			Override []struct {
				InstanceType string `mapstructure:"instance_type"`
			} `mapstructure:"override"`
		} `mapstructure:"launch_template"`

		InstancesDistribution []struct {
			OnDemandBaseCapacity float64 `mapstructure:"on_demand_base_capacity"`
		} `mapstructure:"instances_distribution"`
	} `mapstructure:"mixed_instances_policy"`

	MinSize         int64 `mapstructure:"min_size"`
	DesiredCapacity int64 `mapstructure:"desired_capacity"`
}

// decodeAutoscalingGroupValues decodes and returns instanceValues from a Terraform values map.
func decodeAutoscalingGroupValues(tfVals map[string]interface{}) (autoscalingGroupValues, error) {
	var v autoscalingGroupValues
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
		lc, err := decodeLaunchConfigurationValues(rss[vals.LaunchConfiguration].Values)
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

	if len(vals.LaunchTemplate) > 0 || len(vals.MixedInstancesPolicy) > 0 {

		var ltref string
		if len(vals.LaunchTemplate) > 0 {
			if len(vals.LaunchTemplate[0].ID) > 0 {
				ltref = vals.LaunchTemplate[0].ID
			} else {
				ltref = vals.LaunchTemplate[0].Name
			}
		}

		// variable used to when instance type overrided by the ASG
		mixedOverrideInstanceType := ""
		if len(vals.MixedInstancesPolicy) > 0 {
			if len(vals.MixedInstancesPolicy[0].LaunchTemplate[0].LaunchTemplateSpecification[0].LaunchTemplateID) > 0 {
				ltref = vals.MixedInstancesPolicy[0].LaunchTemplate[0].LaunchTemplateSpecification[0].LaunchTemplateID
			} else {
				ltref = vals.MixedInstancesPolicy[0].LaunchTemplate[0].LaunchTemplateSpecification[0].LaunchTemplateName
			}

			// Logic partially implemented.
			// We do not evaluate % between spot and on-demande
			// We also assume first InstanceType override is the main one
			if len(vals.MixedInstancesPolicy[0].LaunchTemplate[0].Override) > 0 {
				mixedOverrideInstanceType = vals.MixedInstancesPolicy[0].LaunchTemplate[0].Override[0].InstanceType
			}
		}

		lt, err := decodeLaunchTemplateValues(rss[ltref].Values)
		if err != nil {
			return inst
		}

		if mixedOverrideInstanceType != "" {
			inst.instanceType = mixedOverrideInstanceType
		} else {
			inst.instanceType = lt.InstanceType
		}

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
