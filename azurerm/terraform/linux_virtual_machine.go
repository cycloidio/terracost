package terraform

import (
	"fmt"
	"strings"

	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// LinuxWindowsVirtualMachine is the entity that holds the logic to calculate price
// of the google_compute_instance
type LinuxWindowsVirtualMachine struct {
	provider        *Provider
	location        string
	size            string
	ultraSSDEnabled bool

	managedDisk *ManagedDisk

	// windows params
	os          string
	licenseType string
}

// linuxVirtualMachineValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type linuxVirtualMachineValues struct {
	Size     string `mapstructure:"size"`
	Location string `mapstructure:"location"`

	OSDisk []struct {
		StorageAccountType string  `mapstructure:"storage_account_type"`
		DiskSizeGB         float64 `mapstructure:"disk_size_gb"`
	} `mapstructure:"os_disk"`

	AdditionalCapabilities []struct {
		UltraSSDEnabled bool `mapstructure:"ultra_ssd_enabled"`
		UltraSSDLRS     bool `mapstructure:"UltraSSD_LRS"`
	} `mapstructure:"additional_capabilities"`

	Usage struct {
		OSDisk struct {
			MonthlyDiskOperations float64 `mapstructure:"monthly_disk_operations"`
		} `mapstructure:"os_disk"`
	} `mapstructure:"tc_usage"`
}

// decodeLinuxWindowsVirtualMachineValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeLinuxVirtualMachineValues(tfVals map[string]interface{}) (linuxVirtualMachineValues, error) {
	var v linuxVirtualMachineValues
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

// newLinuxVirtualMachine initializes a new LinuxVirtualMachine from the provider
func (p *Provider) newLinuxVirtualMachine(vals linuxVirtualMachineValues) *LinuxWindowsVirtualMachine {
	inst := &LinuxWindowsVirtualMachine{
		provider: p,

		location: region.GetLocationName(vals.Location),
		size:     vals.Size,
		os:       "linux",
	}

	if len(vals.AdditionalCapabilities) > 0 {
		inst.ultraSSDEnabled = vals.AdditionalCapabilities[0].UltraSSDEnabled
	}

	if len(vals.OSDisk) > 0 {
		inst.managedDisk = &ManagedDisk{
			provider:           p,
			location:           region.GetLocationName(vals.Location),
			diskSizeGB:         decimal.NewFromFloat(vals.OSDisk[0].DiskSizeGB),
			storageAccountType: vals.OSDisk[0].StorageAccountType,

			// Usage
			monthlyDiskOperations: decimal.NewFromFloat(vals.Usage.OSDisk.MonthlyDiskOperations),
		}
	}
	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *LinuxWindowsVirtualMachine) Components() []query.Component {
	components := []query.Component{}

	if inst.os == "linux" {
		components = append(components, inst.linuxVirtualMachineComponent(inst.provider.key, inst.location, inst.size))
	} else {
		components = append(components, inst.windowsVirtualMachineComponent(inst.provider.key, inst.location, inst.size, inst.licenseType))
	}

	if inst.ultraSSDEnabled {
		components = append(components, inst.linuxVirtualMachineultraSSDReservationComponent(inst.provider.key, inst.location))
	}

	if inst.managedDisk != nil {
		components = append(components, inst.managedDisk.Components()...)
	}

	return components
}

func (inst *LinuxWindowsVirtualMachine) linuxVirtualMachineComponent(key, location, size string) query.Component {
	productNameRe := "Series( Linux)?$"
	if strings.HasPrefix(strings.ToLower(size), "basic_") {
		productNameRe = "Series Basic$"
	} else if !strings.HasPrefix(strings.ToLower(size), "standard_") {
		size = fmt.Sprintf("Standard_%s", size)
	}

	return query.Component{
		Name:           "Compute Linux",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Virtual Machines"),
			Family:   util.StringPtr("Compute"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", ValueRegex: util.StringPtr(productNameRe)},
				{Key: "armSkuName", Value: util.StringPtr(size)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 Hour"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}

func (inst *LinuxWindowsVirtualMachine) linuxVirtualMachineultraSSDReservationComponent(key string, location string) query.Component {
	return query.Component{
		Name:           "Ultra disk reservation vCPU",
		HourlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "skuName", Value: util.StringPtr("Ultra LRS")},
				{Key: "productName", Value: util.StringPtr("Ultra Disks")},
				{Key: "meterName", ValueRegex: util.StringPtr("Reservation per vCPU Provisioned$")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1/Hour"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
