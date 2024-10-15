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

	"github.com/davecgh/go-spew/spew"
)

// ManagedDisk is the entity that holds the logic to calculate price
// of the google_compute_instance
type ManagedDisk struct {
	provider *Provider
	location string

	diskSizeGB         decimal.Decimal
	diskIOPSReadWrite  decimal.Decimal
	diskMBPSReadWrite  decimal.Decimal
	storageAccountType string

	// Usage
	monthlyDiskOperations decimal.Decimal
}

// managedDiskValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type managedDiskValues struct {
	Location           string  `mapstructure:"location"`
	DiskSizeGB         float64 `mapstructure:"disk_size_gb"`
	DiskIOPSReadWrite  float64 `mapstructure:"disk_iops_read_write"`
	DiskMBPSReadWrite  float64 `mapstructure:"disk_mbps_read_write"`
	StorageAccountType string  `mapstructure:"storage_account_type"`

	Usage struct {
		MonthlyDiskOperations float64 `mapstructure:"monthly_disk_operations"`
	} `mapstructure:"tc_usage"`
}

// decodeManagedDiskValues decodes and returns computeInstanceValues from a Terraform values map.
func decodeManagedDiskValues(tfVals map[string]interface{}) (managedDiskValues, error) {
	var v managedDiskValues
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

// newManagedDisk initializes a new ManagedDisk from the provider
func (p *Provider) newManagedDisk(vals managedDiskValues) *ManagedDisk {
	inst := &ManagedDisk{
		provider:           p,
		location:           region.GetLocationName(vals.Location),
		diskSizeGB:         decimal.NewFromFloat(vals.DiskSizeGB),
		diskIOPSReadWrite:  decimal.NewFromFloat(vals.DiskIOPSReadWrite),
		diskMBPSReadWrite:  decimal.NewFromFloat(vals.DiskMBPSReadWrite),
		storageAccountType: vals.StorageAccountType,

		// Usage
		monthlyDiskOperations: decimal.NewFromFloat(vals.Usage.MonthlyDiskOperations),
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *ManagedDisk) Components() []query.Component {
	components := []query.Component{}

	p := strings.Split(inst.storageAccountType, "_")
	diskTypePrefix := p[0]
	var storageReplicationType string
	if len(p) > 1 {
		storageReplicationType = strings.ToUpper(p[1])
	}

	// TODO
	if strings.ToLower(diskTypePrefix) == "ultrassd" {
		components = append(components, inst.managedDiskStandardPremiumOperationsComponent(inst.provider.key, inst.location, "productName", "diskName", storageReplicationType))
	} else {

		var diskProductNameMap = map[string]string{
			"standard":    "Standard HDD Managed Disks",
			"standardssd": "Standard SSD Managed Disks",
			"premium":     "Premium SSD Managed Disks",
		}

		productName, ok := diskProductNameMap[strings.ToLower(diskTypePrefix)]
		if ok {
			diskName := "foo"
			// diskName := mapDiskName(diskTypePrefix, requestedSize)
			components = append(components, inst.managedDiskStandardPremiumOperationsComponent(inst.provider.key, inst.location, productName, diskName, storageReplicationType))
		}
	}

	spew.Dump(components)

	return components
}

func (inst *ManagedDisk) managedDiskStandardPremiumOperationsComponent(key, location, productName string, diskName string, storageReplicationType string) query.Component {
	return query.Component{
		Name:            "Disk operations",
		MonthlyQuantity: decimal.NewFromInt(1),
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "product_name", Value: util.StringPtr(productName)},
				{Key: "sku_name", Value: util.StringPtr(fmt.Sprintf("%s %s", diskName, storageReplicationType))},
				{Key: "meter_name", Value: util.StringPtr("Disk Operations")},
				// {Key: "meterName", ValueRegex: regexPtr("Disk Operations$")},

			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("10k operations"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
