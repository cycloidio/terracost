package terraform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cycloidio/terracost/azurerm/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

//To check the available meterName and SkuName for the resource
// - curl -s "https://prices.azure.com/api/retail/prices?\$filter=productName eq 'IP Addresses'" | jq '.Items[] | {skuName, meterName}' | sort -u

// PostgreSQLFlexibleServer is the entity that holds the logic to calculate price
// of the azurerm_postgresql_flexible_server
type PostgreSQLFlexibleServer struct {
	provider *Provider

	location  string
	skuName   string
	storageMB int64

	// Usage
	monthlyHours    decimal.Decimal
	backupStorageGB decimal.Decimal
}

// postgreSQLFlexibleServerValues is holds the terraform values that we need to estimate the price
type postgreSQLFlexibleServerValues struct {

	//required params
	Location string `mapstructure:"location"`
	SkuName  string `mapstructure:"sku_name"` // e.g. B_Standard_B1ms, GP_Standard_D2s_v3, MO_Standard_E4s_v3

	// optional params
	StorageMB        int64      `mapstructure:"storage_mb"` // in MB, default 32768
	HighAvailability []struct { // With HighAvailability enabled, the price will be doubled
		Mode string `mapstructure:"mode"`
	} `mapstructure:"high_availability"`
	StorageTier             string `mapstructure:"storage_tier"` // e.g. P4, P6, P10, P15, P20, P30, P40, P50, P60, P70, P80
	GeoRedudantBackupEnable bool   `mapstructure:"geo_redundant_backup_enabled"`

	// usage - with default values
	Usage struct {
		MonthlyHours    int64 `mapstructure:"monthly_hours"`
		BackupStorageGB int64 `mapstructure:"additional_backup_storage_gb"`
	} `mapstructure:"tc_usage"`
}

// decodePostgreSQLFlexibleServerValues decodes and returns postgreSQLFlexibleServerValues from a Terraform values map.
func decodePostgreSQLFlexibleServerValues(tfVals map[string]interface{}) (postgreSQLFlexibleServerValues, error) {
	var v postgreSQLFlexibleServerValues
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

// newPostgreSQLFlexibleServer initializes a new PostgreSQLFlexibleServer from the provider
func (p *Provider) newPostgreSQLFlexibleServer(vals postgreSQLFlexibleServerValues) *PostgreSQLFlexibleServer {
	inst := &PostgreSQLFlexibleServer{
		provider: p,

		location:  region.GetLocationName(vals.Location),
		skuName:   vals.SkuName, // e.g. B_Standard_B1ms, GP_Standard_D2s_v3, MO_Standard_E4s_v3
		storageMB: 32768,        // default value
		// From Usage
		monthlyHours:    decimal.NewFromInt(vals.Usage.MonthlyHours),
		backupStorageGB: decimal.NewFromInt(vals.Usage.BackupStorageGB),
	}

	// if value is set, use it. Otherwise use default defined at instantiation
	if vals.StorageMB != 0 {
		inst.storageMB = vals.StorageMB
	}

	// if its geo redundant backup enabled, we double the extra backup storage used
	if vals.GeoRedudantBackupEnable == true {
		inst.backupStorageGB = inst.backupStorageGB.Mul(decimal.NewFromInt(2))
	}
	//if high availability is enabled, the price will be doubled
	if len(vals.HighAvailability) > 0 {
		inst.monthlyHours = inst.monthlyHours.Mul(decimal.NewFromInt(2))
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *PostgreSQLFlexibleServer) Components() []query.Component {

	// Split the SKU name to extract the tier and instance type
	skuParts := strings.Split(inst.skuName, "_")
	tier := skuParts[0]

	// Use regex to find the number of cores. Example B_Standard_B1ms -> 1 cores
	nrRegex := regexp.MustCompile(`[0-9]+`)
	nrCores := nrRegex.FindString(skuParts[2])
	meterName := "vCore"
	skuName := nrCores + " vCore"

	//remove number of cores from instance type
	instanceType := nrRegex.ReplaceAllString(skuParts[2], "")

	// For some SKUs, the instance type may have a version part, e.g. "D2s_v3"
	if len(skuParts) > 3 {
		instanceType += skuParts[3]
	}
	productName := ""

	switch tier {
	// Burtstable
	case "B":
		productName = "Azure Database for PostgreSQL Flexible Server Burstable BS Series Compute"

		// For the Burstable series it stays the same for example in B_Standard_B1ms -> B1ms
		instanceType = skuParts[2] // so no need to remove the number of cores
		skuName = instanceType

		meterName = instanceType + " vCore"
		// random exception
		if instanceType == "B2s" || instanceType == "B1ms" {
			meterName = strings.ToUpper(instanceType)
			skuName = strings.ToUpper(instanceType)
		}

	// General Purpose
	case "GP":
		productName = fmt.Sprintf("Azure Database for PostgreSQL Flexible Server General Purpose %s Series Compute", instanceType)
		// random exception for Ddsv5
		if instanceType == "Ddsv5" {
			productName = fmt.Sprintf("Azure Database for PostgreSQL Flexible Server General Purpose - %s Series Compute", instanceType)
		}
	// Memory Optimized
	case "MO":
		productName = fmt.Sprintf("Azure Database for PostgreSQL Flexible Server Memory Optimized %s Series Compute", instanceType)
	}

	// compute component
	components := []query.Component{
		inst.computeComponent(inst.provider.key, inst.location, skuName, meterName, productName, inst.monthlyHours),
	}

	// storage component
	components = append(components, inst.storageComponent(inst.provider.key, inst.location, decimal.NewFromInt(inst.storageMB/1024)))

	// backup component
	components = append(components, inst.backupComponent(inst.provider.key, inst.location, inst.backupStorageGB))

	return components

}

func (inst *PostgreSQLFlexibleServer) computeComponent(key, location, skuName, meterName, productName string, monthlyHours decimal.Decimal) query.Component {

	return query.Component{
		Name:            "Compute Flexible PostgreSQL Server",
		MonthlyQuantity: monthlyHours,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Azure Database for PostgreSQL"),
			Family:   util.StringPtr("Databases"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "meterName", Value: util.StringPtr(meterName)},
				{Key: "skuName", Value: util.StringPtr(skuName)},
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

func (inst *PostgreSQLFlexibleServer) storageComponent(key, location string, monthlyGBs decimal.Decimal) query.Component {
	return query.Component{
		Name:            "Storage Flexible PostgreSQL Server",
		MonthlyQuantity: monthlyGBs,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Azure Database for PostgreSQL"),
			Family:   util.StringPtr("Databases"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr("Az DB for PostgreSQL Flexible Server Storage")},
				{Key: "meterName", Value: util.StringPtr("Storage Data Stored")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 GB/Month"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}

func (inst *PostgreSQLFlexibleServer) backupComponent(key, location string, backupStorage decimal.Decimal) query.Component {

	return query.Component{
		Name:            "Backup Storage Flexible PostgreSQL Server",
		MonthlyQuantity: backupStorage,
		Usage:           true,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Azure Database for PostgreSQL"),
			Family:   util.StringPtr("Databases"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "productName", Value: util.StringPtr("Azure Database for PostgreSQL Flexible Server Backup Storage")},
				{Key: "meterName", Value: util.StringPtr("Backup Storage LRS Data Stored")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("1 GB/Month"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
