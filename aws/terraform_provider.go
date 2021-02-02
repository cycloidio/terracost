package aws

import (
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/terraform"
)

// RegistryName is the fully qualified name under which this provider is stored in the registry.
const RegistryName = "registry.terraform.io/hashicorp/aws"

// TerraformProviderInitializer is a terraform.ProviderInitializer that initializes the default AWS provider.
var TerraformProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{ProviderName, RegistryName},
	Provider: func(values map[string]string) (terraform.Provider, error) {
		regCode := region.Code(values["region"])
		return awstf.NewProvider(ProviderName, regCode)
	},
}
