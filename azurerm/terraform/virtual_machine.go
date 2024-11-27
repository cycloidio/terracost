package terraform

import (
	"strings"

	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// virtualMachineValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type virtualMachineValues struct {
	VMSize   string `mapstructure:"vm_size"`
	Location string `mapstructure:"location"`

	StorageOSDisk []struct {
		OSType          string  `mapstructure:"os_type"`
		DiskSizeGB      float64 `mapstructure:"disk_size_gb"`
		ManagedDiskType string  `mapstructure:"managed_disk_type"`
	} `mapstructure:"storage_os_disk"`

	AdditionalCapabilities []struct {
		UltraSSDEnabled bool `mapstructure:"ultra_ssd_enabled"`
	} `mapstructure:"additional_capabilities"`

	Usage struct {
		OSDisk struct {
			MonthlyDiskOperations float64 `mapstructure:"monthly_disk_operations"`
		} `mapstructure:"os_disk"`
	} `mapstructure:"tc_usage"`
}

// decodeVirtualMachineValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeVirtualMachineValues(tfVals map[string]interface{}) (virtualMachineValues, error) {
	var v virtualMachineValues
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

// newVirtualMachine initializes a new VirtualMachine from the provider
func (p *Provider) newVirtualMachine(vals virtualMachineValues) *LinuxWindowsVirtualMachine {
	inst := &LinuxWindowsVirtualMachine{
		provider: p,

		location: region.GetLocationName(vals.Location),
		size:     vals.VMSize,
		os:       "linux",
	}

	if len(vals.AdditionalCapabilities) > 0 {
		inst.ultraSSDEnabled = vals.AdditionalCapabilities[0].UltraSSDEnabled
	}

	if len(vals.StorageOSDisk) > 0 {
		inst.managedDisk = &ManagedDisk{
			provider:           p,
			location:           region.GetLocationName(vals.Location),
			diskSizeGB:         decimal.NewFromFloat(vals.StorageOSDisk[0].DiskSizeGB),
			storageAccountType: vals.StorageOSDisk[0].ManagedDiskType,

			// Usage
			monthlyDiskOperations: decimal.NewFromFloat(vals.Usage.OSDisk.MonthlyDiskOperations),
		}
		if vals.StorageOSDisk[0].OSType != "" {
			inst.os = strings.ToLower(vals.StorageOSDisk[0].OSType)
		}
	}

	return inst
}
