package terraform

import (
	"github.com/cycloidio/terracost/query"
	"github.com/mitchellh/mapstructure"
)

// StorageAccount is the entity that holds the logic to calculate price
// of the azurerm_storage_account
type StorageAccount struct {
	provider               *Provider
	name                   string
	location               string
	accountKind            string
	accountTier            string
	accessTier             string
	accountReplicationType string
}

// storageAccountValues is holds the values that we need to be able
// to calculate the price of the StorageAccount
type storageAccountValues struct {
	//required params
	Name                   string `mapstructure:"name"`
	Location               string `mapstructure:"location"`
	AccountTier            string `mapstructure:"account_tier"`             // Standard and Premium
	AccountReplicationType string `mapstructure:"account_replication_type"` //LRS, GRS, RAGRS, ZRS, GZRS and RAGZRS

	//optional params
	AccountKind string `mapstructure:"account_kind"` //BlobStorage, BlockBlobStorage, FileStorage, Storage and StorageV2. Defaults StorageV2
	AccessTier  string `mapstructure:"access_tier"`  //Hot and Cold. Default Hot
}

// decodeStorageAccountValues decodes and returns storageAccountValues from a Terraform values map.
func decodeStorageAccountValues(tfVals map[string]interface{}) (storageAccountValues, error) {
	var v storageAccountValues
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

// newStorageAccount initializes a new StorageAccount from the provider
func (p *Provider) newStorageAccount(vals storageAccountValues) *StorageAccount {
	inst := &StorageAccount{
		provider: p,
		// required terraform values
		name:                   vals.Name,
		location:               vals.Location,
		accountTier:            vals.AccountTier,
		accountReplicationType: vals.AccountReplicationType,
		//optional terraform values - take default values
		accountKind: "StorageV2",
		accessTier:  "Hot",
	}

	//Optional values
	if vals.AccountKind != "" {
		inst.accountKind = vals.AccountKind
	}
	if vals.AccessTier != "" {
		inst.accessTier = vals.AccessTier
	}

	return inst
}

// Components returns the price component empty since is only used to add details to others
func (inst *StorageAccount) Components() []query.Component {
	return []query.Component{}
}
