package azurerm

import (
	"slices"
)

//go:generate enumer -type=Service -output=service_string.go -linecomment=true

// Service is the type defining the services
type Service uint8

// List of all the supported services
const (
	AzureBastion                Service = iota // Azure Bastion
	AzureDNS                    Service = iota // Azure DNS
	AzureDatabaseForPostgresSQL Service = iota // Azure Database for PostgreSQL
	NATGateway                  Service = iota // NAT Gateway
	Storage                     Service = iota // Storage
	VirtualMachines             Service = iota // Virtual Machines
	VirtualNetwork              Service = iota // Virtual Network
	VPNGateway                  Service = iota // VPN Gateway
)

var (
	// The list of all services is https://azure.microsoft.com/en-us/services/, the left side is
	// the Family and the main content is the Services
	services = map[string]struct{}{
		AzureBastion.String():                struct{}{},
		AzureDNS.String():                    struct{}{},
		AzureDatabaseForPostgresSQL.String(): struct{}{},
		NATGateway.String():                  struct{}{},
		Storage.String():                     struct{}{},
		VirtualMachines.String():             struct{}{},
		VPNGateway.String():                  struct{}{},
		VirtualNetwork.String():              struct{}{},
	}
)

// GetSupportedServices returns all the Azure service names that Terracost supports.
func GetSupportedServices() []string {
	svcs := make([]string, 0, len(services))
	for k := range services {
		svcs = append(svcs, k)
	}
	slices.Sort(svcs)
	return svcs
}
