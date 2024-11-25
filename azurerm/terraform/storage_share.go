package terraform

import (
	"fmt"
	"strings"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

// Some notes about the terraform resource pricing:
// It can be quite complex to understand the relationship of storage share and storage account
// This following table  allows to have an idea on the different combinations that can appear:
// https://learn.microsoft.com/en-us/azure/storage/files/storage-how-to-create-file-share?tabs=azure-portal#create-a-storage-account

// storageShare is the entity that holds the logic to calculate price
// of the azurerm_storage_share
type StorageShare struct {
	provider *Provider

	// values from storage share resource
	accessTier string
	quota      decimal.Decimal

	// values from storage accout resource
	storageAccountLocation        string
	storageAccountReplicationType string // (required) values LRS, GRS, RAGRS, ZRS, GZRS, RAGZRS
	storageAccountKind            string // (optional) values BlobStorage, BlockBlobStorage, FileStorage, Storage and StorageV2. Defaults to StorageV2.

	// default usage
	monthlyUsedStorage          decimal.Decimal
	monthlySnapshotUsedStorage  decimal.Decimal
	monthlyMetadataUsageStorage decimal.Decimal
	monthlyCoolDataRetrieval    decimal.Decimal
	monthlyWriteTransactions    decimal.Decimal
	monthlyListTransactions     decimal.Decimal
	monthlyReadTransactions     decimal.Decimal
	monthlyOtherTransactions    decimal.Decimal // delete operations are excluded
}

// storageShareValues holds the terraform values that we need to be able
// to calculate the price of the StorageShare
type storageShareValues struct {
	// required parameters
	AccessTier string `mapstructure:"access_tier"` // Values are Hot, Cool and TransactionOptimized (requires storage account with accountKind - StorageV2) or premium (requires storage account with accountKind - FileStorage)

	// optional parameters
	Quota              int64  `mapstructure:"quota"` // max share size in Gb.  Note! For premium storage Accounts corresponds to the quantity that will be charged
	StorageAccountName string `mapstructure:"storage_account_name"`

	Usage struct {
		MonthlyWriteTransactions int64 `mapstructure:"monthly_write_transactions"`
		MonthlyListTransactions  int64 `mapstructure:"monthly_list_transactions"`
		MonthlyReadTransactions  int64 `mapstructure:"monthly_read_transactions"`
		MonthlyOtherTransactions int64 `mapstructure:"monthly_other_transactions"`
	} `mapstructure:"tc_usage"`
}

// decodeStorageShareValues decodes and returns storageShareValues from a Terraform values map.
func decodeStorageShareValues(tfVals map[string]interface{}) (storageShareValues, error) {
	var v storageShareValues
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

// newStorageShare initializes a new StorageShare from the provider
func (p *Provider) newStorageShare(rss map[string]terraform.Resource, vals storageShareValues) *StorageShare {
	inst := &StorageShare{
		provider: p,
		// mandatory resource values
		quota: decimal.NewFromInt(vals.Quota),
		// optinal resources values - takes default value if empty
		accessTier:         "Standard", //
		storageAccountKind: "StorageV2",

		// Usage default values
		monthlyWriteTransactions: decimal.NewFromInt(vals.Usage.MonthlyWriteTransactions),
		monthlyListTransactions:  decimal.NewFromInt(vals.Usage.MonthlyListTransactions),
		monthlyReadTransactions:  decimal.NewFromInt(vals.Usage.MonthlyReadTransactions),
		monthlyOtherTransactions: decimal.NewFromInt(vals.Usage.MonthlyOtherTransactions),
	}
	//accessTier is optional.
	//Note! TransactionOptimized is ignored since it corresponds to the default value on the billing API standard
	// 		terraform values		billing values
	//		"Hot":                  "Hot",
	//		"Cool":                 "Cool",
	//		"TransactionOptimized": "Standard",
	//		"Premium":              "Premium",
	if vals.AccessTier != "" && vals.AccessTier != "TransactionOptimized" {
		inst.accessTier = vals.AccessTier

	}

	// retrieve values from storage account
	for _, resource := range rss {
		if resource.Type == "azurerm_storage_account" && resource.Name == vals.StorageAccountName {

			storageAccountVals, err := decodeStorageAccountValues(resource.Values)
			// if no storage account found return empty storageShare, since we won't be able to calculate price
			if err != nil || storageAccountVals == (storageAccountValues{}) {
				return &StorageShare{}
			}
			inst.storageAccountLocation = storageAccountVals.Location
			inst.storageAccountKind = storageAccountVals.AccountKind
			// RAGRS and RAGZRS should be interpreted as GRS or GZRS
			inst.storageAccountReplicationType = strings.TrimPrefix(storageAccountVals.AccountReplicationType, "RA")

			break
		}
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *StorageShare) Components() []query.Component {

	components := []query.Component{}

	// check kind of storage account -> Only these are permited FileStorage, Storage and StorageV2
	if inst.storageAccountKind != "FileStorage" && inst.storageAccountKind != "Storage" && inst.storageAccountKind != "StorageV2" {
		return components
	}

	// if Storage account type = FileStorage it requires Premium access Tier file share and only LRS and ZRS supported
	if inst.storageAccountKind == "FileStorage" && (inst.accessTier != "Premium" || !(inst.storageAccountReplicationType == "LRS" || inst.storageAccountReplicationType == "ZRS")) {
		return components
	}

	// if Storage account type = Storage V1 only GRS and LRS are supported
	if inst.storageAccountKind != "Storage" && !(inst.storageAccountReplicationType == "GRS" || inst.storageAccountReplicationType == "LRS") {
		return components
	}

	// call the different components depending on the type of Storage Account (that's how their grouped on the billing API)

	skuName := inst.accessTier + " " + inst.storageAccountReplicationType

	// alows to map component names to more easily reuse StorageShare component
	componentNameMaping := map[string]string{
		"snapshot":           "Snapshot GB/Month Used",
		"storage":            "Storage GB/Month Used",
		"cool-data":          "Cool Data Retrieval GB",
		"write-transactions": "Write Transactions per 10K",
		"list-transactions":  "List Transactions per 10K",
		"read-transactions":  "Read Transactions per 10K",
		"other-transactions": "Other Transactions per 10K",
	}
	componentsSpecs := [][]interface{}{} // in the format componentName , meterName, quantityUsed

	switch inst.storageAccountKind {
	case "Storage": // includes Storage,List/Read/Write/Other transactions price

		componentsSpecs = [][]interface{}{
			{componentNameMaping["storage"], fmt.Sprintf("%s Data Stored", inst.storageAccountReplicationType), inst.quota},
			{componentNameMaping["list-transactions"], "List Operations", inst.monthlyListTransactions},
			{componentNameMaping["read-transactions"], "Read Operations", inst.monthlyReadTransactions},
			{componentNameMaping["write-transactions"], fmt.Sprintf("%s Write Operations", inst.storageAccountReplicationType), inst.monthlyWriteTransactions},
			{componentNameMaping["other-transactions"], "Protocol Operations", inst.monthlyOtherTransactions},
		}

	case "StorageV2": // values will depend on the type of access tier of the file share
		switch inst.accessTier {

		case "Hot": //includes Storage, Snapshot, List/Read/Write/Other transactions and metadata

			meterNameOtherOperations := "Hot Other Operations"
			if inst.storageAccountReplicationType == "GZRS" {
				meterNameOtherOperations = "Hot GZRS Other Operations"
			}

			componentsSpecs = [][]interface{}{
				{componentNameMaping["storage"], fmt.Sprintf("Hot %s Data Stored", inst.storageAccountReplicationType), inst.quota},
				{componentNameMaping["metadata"], fmt.Sprintf("%s Metadata", inst.storageAccountReplicationType), inst.quota},
				{componentNameMaping["list-transactions"], fmt.Sprintf("Hot %s List Operations", inst.storageAccountReplicationType), inst.monthlyListTransactions},
				{componentNameMaping["read-transactions"], "Hot Read Operations", inst.monthlyReadTransactions},
				{componentNameMaping["write-transactions"], fmt.Sprintf("Hot %s Write Operations", inst.storageAccountReplicationType), inst.monthlyWriteTransactions},
				{componentNameMaping["other-transactions"], meterNameOtherOperations, inst.monthlyOtherTransactions},
			}

		case "Cool": //includes Storage, Snapshot, List/Read/Write/Other transactions, metadata and cold data retrieval price

			meterNameOtherOperations := "Cool Other Operations"
			meterNameCoolData := "Cool Data Retrieval"
			if inst.storageAccountReplicationType == "GZRS" {
				meterNameOtherOperations = "Cool GZRS Other Operations"
				meterNameCoolData = "Cool GZRS Data Retrieval"
			}

			componentsSpecs = [][]interface{}{
				{componentNameMaping["storage"], fmt.Sprintf("Cool %s Data Stored", inst.storageAccountReplicationType), inst.quota},
				{componentNameMaping["cool-data"], meterNameCoolData, inst.quota},
				{componentNameMaping["metadata"], fmt.Sprintf("%s Metadata", inst.storageAccountReplicationType), inst.quota},
				{componentNameMaping["list-transactions"], fmt.Sprintf("Cool %s List Operations", inst.storageAccountReplicationType), inst.monthlyListTransactions},
				{componentNameMaping["read-transactions"], "Cool Read Operations", inst.monthlyReadTransactions},
				{componentNameMaping["write-transactions"], fmt.Sprintf("Cool %s Write Operations", inst.storageAccountReplicationType), inst.monthlyWriteTransactions},
				{componentNameMaping["other-transactions"], meterNameOtherOperations, inst.monthlyOtherTransactions},
			}

		case "Standard": //includes Storage, Snapshot, List/Read/Write/Other transactions and metadata

			meterNameOtherOperations := "Protocol Operations"
			meterNameListOperations := "List Operations"
			meterNameReadOperations := "Read Operations"
			if inst.storageAccountReplicationType == "GZRS" || inst.storageAccountReplicationType == "ZRS" {
				meterNameOtherOperations = fmt.Sprintf("%s Protocol Operations", inst.storageAccountReplicationType)
				meterNameListOperations = fmt.Sprintf("%s List Operations", inst.storageAccountReplicationType)
				meterNameReadOperations = fmt.Sprintf("%s Read Operations", inst.storageAccountReplicationType)
			}

			componentsSpecs = [][]interface{}{
				{componentNameMaping["storage"], fmt.Sprintf("Cool %s Data Stored", inst.storageAccountReplicationType), inst.quota},
				{componentNameMaping["metadata"], fmt.Sprintf("%s Metadata", inst.storageAccountReplicationType), inst.quota},
				{componentNameMaping["list-transactions"], meterNameListOperations, inst.monthlyListTransactions},
				{componentNameMaping["read-transactions"], meterNameReadOperations, inst.monthlyReadTransactions},
				{componentNameMaping["write-transactions"], fmt.Sprintf("%s Write Operations", inst.storageAccountReplicationType), inst.monthlyWriteTransactions},
				{componentNameMaping["other-transactions"], meterNameOtherOperations, inst.monthlyOtherTransactions},
			}

		}
	case "FileStorage":
		componentsSpecs = [][]interface{}{
			{componentNameMaping["storage"], fmt.Sprintf("Premium %s Provisioned", inst.storageAccountReplicationType), inst.quota},
			{componentNameMaping["snapshot"], fmt.Sprintf("Premium %s Snapshots", inst.storageAccountReplicationType), inst.quota},
		}
	}

	// Iterating over componentsSpecs  to add components to component list
	for _, component := range componentsSpecs {
		components = append(components, inst.storageShareComponent(inst.provider.key, inst.storageAccountLocation, component[0].(string), skuName, component[1].(string), component[2].(decimal.Decimal)))
	}

	return components
}

func (inst *StorageShare) storageShareComponent(key, location, componentName string, skuName string, meterName string, quantityUsed decimal.Decimal) query.Component {

	// retrieve the type of pricingUnit used on the billing API depending on the componentName
	pricingUnit := "1 GB/Month"
	if strings.Contains(componentName, "Transactions") {
		pricingUnit = "10K"
	} else if strings.Contains(componentName, "Cool Data") {
		pricingUnit = "1 GB"
	}

	return query.Component{
		Name:            fmt.Sprintf(componentName),
		MonthlyQuantity: quantityUsed,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(key),
			Service:  util.StringPtr("Storage"),
			Family:   util.StringPtr("Storage"),
			Location: util.StringPtr(location),
			AttributeFilters: []*product.AttributeFilter{
				//{Key: "productName", Value: util.StringPtr(productName)},
				{Key: "meterName", Value: util.StringPtr(meterName)},
				{Key: "skuName", Value: util.StringPtr(skuName)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr(pricingUnit),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "type", Value: util.StringPtr("Consumption")},
			},
		},
	}
}
