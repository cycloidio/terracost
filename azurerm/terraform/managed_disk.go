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

// ManagedDisk is the entity that holds the logic to calculate price
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

// decodeManagedDiskValues decodes and returns Values from a Terraform values map.
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
	var replicationType string
	if len(p) > 1 {
		replicationType = strings.ToUpper(p[1])
	}

	var diskProductNameMap = map[string]string{
		"standard":    "Standard HDD Managed Disks",
		"standardssd": "Standard SSD Managed Disks",
		"premium":     "Premium SSD Managed Disks",
		"premiumv2":   "Azure Premium SSD v2",
		"ultrassd":    "Ultra Disks",
	}

	var diskSkuNamePrefMap = map[string]string{
		"standard":    "Standard",
		"standardssd": "Standard",
		"premium":     "Premium",
		"premiumv2":   "Premium",
		"ultrassd":    "Ultra",
	}

	productName := diskProductNameMap[strings.ToLower(diskTypePrefix)]
	skuNamePref := diskSkuNamePrefMap[strings.ToLower(diskTypePrefix)]

	if strings.ToLower(diskTypePrefix) == "ultrassd" || strings.ToLower(diskTypePrefix) == "premiumv2" {
		// Ultra & PremiumV2
		diskSizeGB := 1024
		if inst.diskSizeGB.Cmp(decimal.NewFromInt(0)) > 0 {
			diskSizeGB = int(inst.diskSizeGB.IntPart())
		}

		components = append(components, inst.managedDiskUltraComponent(inst.provider.key, inst.location, diskSizeGB, replicationType, skuNamePref, productName))

		iops := 2048
		if inst.diskIOPSReadWrite.Cmp(decimal.NewFromInt(0)) > 0 {
			iops = int(inst.diskIOPSReadWrite.IntPart())
		}
		components = append(components, inst.managedDiskUltraProvisionedIOPSComponent(inst.provider.key, inst.location, iops, replicationType, skuNamePref, productName))

		throughput := 8
		if inst.diskMBPSReadWrite.Cmp(decimal.NewFromInt(0)) > 0 {
			throughput = int(inst.diskMBPSReadWrite.IntPart())
		}
		components = append(components, inst.managedDiskUltraProvisionedThroughputComponent(inst.provider.key, inst.location, throughput, replicationType, skuNamePref, productName))
	} else {
		// standard / Premium
		diskSizeGB := 30
		if inst.diskSizeGB.Cmp(decimal.NewFromInt(0)) > 0 {
			diskSizeGB = int(inst.diskSizeGB.IntPart())
		}

		diskName := mapDiskName(diskTypePrefix, diskSizeGB)

		components = append(components, inst.managedDiskStandardPremiumComponent(inst.provider.key, inst.location, productName, diskName, replicationType))
		components = append(components, inst.managedDiskStandardPremiumOperationsComponent(inst.provider.key, inst.location, productName, diskName, replicationType, inst.monthlyDiskOperations))
	}

	return components
}

func (inst *ManagedDisk) managedDiskUltraComponent(key string, location string, size int, replicationType string, skuNamePref string, productName string) query.Component {
	return query.Component{
		Name:           fmt.Sprintf("Storage - ultra %d", size),
		HourlyQuantity: decimal.NewFromInt(int64(size)),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "skuName", Value: util.StringPtr(fmt.Sprintf("%s %s", skuNamePref, replicationType))},
				{Key: "meterName", ValueRegex: util.StringPtr("Provisioned Capacity$")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 GiB/Hour"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}

func (inst *ManagedDisk) managedDiskUltraProvisionedIOPSComponent(key string, location string, iops int, replicationType string, skuNamePref string, productName string) query.Component {
	return query.Component{
		Name:           "Provisioned IOPS",
		HourlyQuantity: decimal.NewFromInt(int64(iops)),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "skuName", Value: util.StringPtr(fmt.Sprintf("%s %s", skuNamePref, replicationType))},
				{Key: "meterName", ValueRegex: util.StringPtr("Provisioned IOPS$")},
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

func (inst *ManagedDisk) managedDiskUltraProvisionedThroughputComponent(key string, location string, iops int, replicationType string, skuNamePref string, productName string) query.Component {
	return query.Component{
		Name:           "Throughput MB/s",
		HourlyQuantity: decimal.NewFromInt(int64(iops)),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "skuName", Value: util.StringPtr(fmt.Sprintf("%s %s", skuNamePref, replicationType))},
				{Key: "meterName", ValueRegex: util.StringPtr("Provisioned Throughput \\(MBps\\)$")},
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

func (inst *ManagedDisk) managedDiskStandardPremiumOperationsComponent(key, location, productName string, diskName string, replicationType string, monthlyDiskOperations decimal.Decimal) query.Component {
	return query.Component{
		Name: "Disk operations",
		// Per 10k
		MonthlyQuantity: monthlyDiskOperations.Div(decimal.NewFromInt(10000)),
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "skuName", Value: util.StringPtr(fmt.Sprintf("%s %s", diskName, replicationType))},
				{Key: "meterName", ValueRegex: util.StringPtr(".*Disk Operations")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("10k"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}

func (inst *ManagedDisk) managedDiskStandardPremiumComponent(key, location, productName string, diskName string, replicationType string) query.Component {
	return query.Component{
		Name:            fmt.Sprintf("Storage - %s %s", diskName, replicationType),
		MonthlyQuantity: decimal.NewFromInt(1),
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "skuName", Value: util.StringPtr(fmt.Sprintf("%s %s", diskName, replicationType))},
				{Key: "meterName", ValueRegex: util.StringPtr(fmt.Sprintf("^%s (%s )?Disk(s)?$", diskName, replicationType))},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1/Month"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}

// Utils

var diskSizeMap = map[string][]struct {
	Name string
	Size int
}{
	"Standard": {
		{"S4", 32},
		{"S6", 64},
		{"S10", 128},
		{"S15", 256},
		{"S20", 512},
		{"S30", 1024},
		{"S40", 2048},
		{"S50", 4096},
		{"S60", 8192},
		{"S70", 16384},
		{"S80", 32767},
	},
	"StandardSSD": {
		{"E1", 4},
		{"E2", 8},
		{"E3", 16},
		{"E4", 32},
		{"E6", 64},
		{"E10", 128},
		{"E15", 256},
		{"E20", 512},
		{"E30", 1024},
		{"E40", 2048},
		{"E50", 4096},
		{"E60", 8192},
		{"E70", 16384},
		{"E80", 32767},
	},
	"Premium": {
		{"P1", 4},
		{"P2", 8},
		{"P3", 16},
		{"P4", 32},
		{"P6", 64},
		{"P10", 128},
		{"P15", 256},
		{"P20", 512},
		{"P30", 1024},
		{"P40", 2048},
		{"P50", 4096},
		{"P60", 8192},
		{"P70", 16384},
		{"P80", 32767},
	},
}

func mapDiskName(diskType string, requestedSize int) string {
	diskTypeMap, ok := diskSizeMap[diskType]
	if !ok {
		return ""
	}

	name := ""
	for _, v := range diskTypeMap {
		name = v.Name
		if v.Size >= requestedSize {
			break
		}
	}

	if requestedSize > diskTypeMap[len(diskTypeMap)-1].Size {
		return ""
	}

	return name
}
