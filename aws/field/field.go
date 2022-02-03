package field

//go:generate enumer -type=Field -output=field_gen.go -linecomment

// Field represents a single column in an AWS offer file.
type Field uint8

// List of fields used by the AWS pricing offer file (CSV).
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
	EndpointType                    // Endpoint Type
	TransferType                    // Transfer Type
	Group                           // Group
	FromLocation                    // From Location
	ToLocation                      // To Location
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
