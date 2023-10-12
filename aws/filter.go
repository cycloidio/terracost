package aws

import (
	"github.com/cycloidio/terracost/price"
)

// IngestionFilter allows control over what pricing data is ingested. Given a price.WithProduct the function returns
// true if the record should be ingested, false if it should be skipped.
type IngestionFilter func(pp *price.WithProduct) bool

// DefaultFilter ingests all the records without filtering.
func DefaultFilter(_ *price.WithProduct) bool {
	return true
}

// MinimalFilter only ingests the supported records, skipping those that would never be used.
func MinimalFilter(pp *price.WithProduct) bool {
	switch pp.Product.Service {
	case "AmazonEC2":
		return minimalFilterEC2(pp)
	case "AmazonRDS":
		return minimalFilterRDS(pp)
	case "AWSELB":
		return true // is minimal already
	case "AmazonElastiCache":
		return true // is minimal already
	case "AmazonCloudWatch":
		return true // is minimal already
	case "AmazonEKS":
		return true // is minimal already
	case "AmazonEFS":
		return true // is minimal already
	case "AmazonFSx":
		return true
	default:
		return false
	}
}

// minimalFilterEC2 only ingests storage-related records as well as compute records that match supported attributes.
func minimalFilterEC2(pp *price.WithProduct) bool {
	switch pp.Product.Family {
	case "Compute Instance":
		allowedProductAttrs := map[string][]string{
			"CapacityStatus":  {"Used"},
			"OperatingSystem": {"Linux"},
			"PreInstalledSW":  {"NA"},
			"Tenancy":         {"Shared", "Dedicated"},
		}
		for k, vals := range allowedProductAttrs {
			if !isValueAllowed(pp.Product.Attributes[k], vals) {
				return false
			}
		}
		return true
	case "Storage", "System Operation", "NAT Gateway":
		return true
	default:
		return false
	}
}

// minimalFilterRDS only ingests RDS records of supported product families.
func minimalFilterRDS(pp *price.WithProduct) bool {
	switch pp.Product.Family {
	case "Database Instance", "Database Storage", "Provisioned IOPS":
		return true
	default:
		return false
	}
}

// isValueAllowed returns true if the allowed slice contains the value.
func isValueAllowed(value string, allowed []string) bool {
	for _, v := range allowed {
		if value == v {
			return true
		}
	}
	return false
}
