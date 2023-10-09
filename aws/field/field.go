package field

//go:generate enumer -type=Field -output=field_gen.go -linecomment

// Field represents a single column in an AWS offer file.
type Field uint8

// List of fields used by the AWS pricing offer file (CSV).
const (
	///// Product Attributes /////
	SKU             Field = iota // SKU
	CapacityStatus               // CapacityStatus
	Group                        // Group
	InstanceType                 // Instance Type
	Location                     // Location
	OperatingSystem              // Operating System
	PreInstalledSW               // Pre Installed S/W
	ProductFamily                // Product Family
	ServiceCode                  // serviceCode
	Tenancy                      // Tenancy
	UsageType                    // usageType
	VolumeAPIName                // Volume API Name
	VolumeType                   // Volume Type

	// EFSFileSystem
	StorageClass    // Storage Class
	AccessType      // Access Type
	ThroughputClass // Throughput Class

	// ElastiCache
	CacheEngine // Cache Engine

	// RDS fields
	DatabaseEngine   // Database Engine
	DatabaseEdition  // Database Edition
	DeploymentOption // Deployment Option
	LicenseModel     // License Model

	// FSx fields
	FileSystemType     // File system type
	StorageType        // Storage type
	ThroughputCapacity // Throughput capacity
	Deployment_option  // Deployment option

	///// Price Attributes /////
	Currency      // Currency
	PricePerUnit  // PricePerUnit
	StartingRange // StartingRange
	TermType      // TermType
	Unit          // Unit
)
