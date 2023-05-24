package aws

import (
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/terraform"
)

const (
	// RegistryName is the fully qualified name under which this provider is stored in the registry.
	RegistryName = "registry.terraform.io/hashicorp/aws"

	// DefaultRegion is the region used by default when none is defined on the provider
	DefaultRegion = "us-east-1"
)

// TerraformProviderInitializer is a terraform.ProviderInitializer that initializes the default AWS provider.
var TerraformProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{ProviderName, RegistryName},
	Provider: func(values map[string]interface{}) (terraform.Provider, error) {
		r, ok := values["region"]
		// If no region is defined it means it was passed via ENV variables
		// and it's not tracked on the Plan or HCL so we'll assume the
		// region to be the DefaultRegion
		if !ok {
			r = DefaultRegion
		}
		regCode := region.Code(r.(string))
		return awstf.NewProvider(ProviderName, regCode)
	},
}
