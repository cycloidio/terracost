package azurerm

//go:generate enumer -type=Service -output=service_string.go -linecomment=true

// Service is the type defining the services
type Service uint8

// List of all the supported services
const (
	VirtualMachines Service = iota // Virtual Machines
)
