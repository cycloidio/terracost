package field

//go:generate go run github.com/dmarkham/enumer -type=Field -output=field_gen.go -linecomment

// Field represents a single column in an AWS offer file.
type Field uint8

const (
	SKU                Field = iota // SKU
	OfferTermCode                   // OfferTermCode
	RateCode                        // RateCode
	PurchaseOption                  // TermType
	PriceDescription                // PriceDescription
	EffectiveDate                   // EffectiveDate
	StartingRange                   // StartingRange
	EndingRange                     // EndingRange
	Unit                            // Unit
	PricePerUnit                    // PricePerUnit
	Currency                        // Currency
	TermLength                      // LeaseContractLength
	TermPurchaseOption              // PurchaseOption
	TermOfferingClass               // OfferingClass
	ProductFamily                   // Product Family
	ServiceCode                     // serviceCode
	Location                        // Location
	InstanceType                    // Instance Type
	StorageMedia                    // Storage Media
	VolumeType                      // Volume Type
	Tenancy                         // Tenancy
	OperatingSystem                 // Operating System
	UsageType                       // usageType
	CapacityStatus                  // CapacityStatus
	PreInstalledSW                  // Pre Installed S/W
	VolumeAPIName                   // Volume API Name

	// RDS fields
	DatabaseEngine   // Database Engine
	DatabaseEdition  // Database Edition
	LicenseModel     // License Model
	DeploymentOption // Deployment Option
)
