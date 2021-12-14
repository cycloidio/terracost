package google

//go:generate enumer -type=Service -output=service_string.go -linecomment=true

// Service is the type defining the services
type Service uint8

// List of all the supported services
const (
	ComputeEngine Service = iota // Compute Engine
)

var (
	// The ID of the service was manually fetched from
	// https://cloud.google.com/billing/v1/how-tos/catalog-api#listing_public_services_from_the_catalog
	services = map[string]string{
		ComputeEngine.String(): "6F81-5844-456A",
	}
)
