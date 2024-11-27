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

// windowsVirtualMachineValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type windowsVirtualMachineValues struct {
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

	LicenseYype string `mapstructure:"license_type"`

	Usage struct {
		OSDisk struct {
			MonthlyDiskOperations float64 `mapstructure:"monthly_disk_operations"`
		} `mapstructure:"os_disk"`
	} `mapstructure:"tc_usage"`
}

// decodeWindowsVirtualMachineValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeWindowsVirtualMachineValues(tfVals map[string]interface{}) (windowsVirtualMachineValues, error) {
	var v windowsVirtualMachineValues
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

// newWindowsVirtualMachine initializes a new WindowsVirtualMachine from the provider
func (p *Provider) newWindowsVirtualMachine(vals windowsVirtualMachineValues) *LinuxWindowsVirtualMachine {
	inst := &LinuxWindowsVirtualMachine{
		provider: p,

		location: region.GetLocationName(vals.Location),
		size:     vals.Size,
		os:       "windows",
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

func (inst *LinuxWindowsVirtualMachine) windowsVirtualMachineComponent(key, location, size string, licenseType string) query.Component {

	productNameRe := "(Series )?Windows$"
	if strings.HasPrefix(size, "Basic_") {
		productNameRe = "Basic Windows$"
	} else if !strings.HasPrefix(size, "Standard_") {
		size = fmt.Sprintf("Standard_%s", size)
	}

	priceType := "Consumption"
	// If defined, specifies that the image or disk that is being used was licensed on-premises
	if strings.ToLower(licenseType) == "windows_client" || strings.ToLower(licenseType) == "windows_server" {
		priceType = "DevTestConsumption"
	}

	return query.Component{
		Name:           "Compute Windows",
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
				{Key: "type", Value: util.StringPtr(priceType)},
			},
		},
	}
}
