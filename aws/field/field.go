package field

//go:generate enumer -type=Field -output=field_gen.go -linecomment

// Field represents a single column in an AWS offer file.
type Field uint8

// List of fields used by the AWS pricing offer file (CSV).
const (
	SKU                Field = iota // SKU
	CapacityStatus                  // CapacityStatus
	Currency                        // Currency
	EffectiveDate                   // EffectiveDate
	EndingRange                     // EndingRange
	InstanceType                    // Instance Type
	Location                        // Location
	OfferTermCode                   // OfferTermCode
	OperatingSystem                 // Operating System
	PreInstalledSW                  // Pre Installed S/W
	PriceDescription                // PriceDescription
	PricePerUnit                    // PricePerUnit
	ProductFamily                   // Product Family
	PurchaseOption                  // TermType
	RateCode                        // RateCode
	ServiceCode                     // serviceCode
	StartingRange                   // StartingRange
	StorageMedia                    // Storage Media
	Tenancy                         // Tenancy
	TermLength                      // LeaseContractLength
	TermOfferingClass               // OfferingClass
	TermPurchaseOption              // PurchaseOption
	Unit                            // Unit
	UsageType                       // usageType
	VolumeAPIName                   // Volume API Name
	VolumeType                      // Volume Type

	// ElastiCache
	CacheEngine     // Cache Engine
	StorageSnapshot // Storage Snapshot

	// RDS fields
	DatabaseEngine   // Database Engine
	DatabaseEdition  // Database Edition
	DeploymentOption // Deployment Option
	LicenseModel     // License Model
)
