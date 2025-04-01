package aws

import (
	"fmt"

	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/log"
	"github.com/cycloidio/terracost/terraform"
)

const (
	// RegistryName is the fully qualified name under which this provider is stored in the registry.
	RegistryName = "registry.terraform.io/hashicorp/aws"

	// DefaultRegion is the region used by default when none is defined on the provider
	DefaultRegion = region.Code("us-east-1")
)

// TerraformProviderInitializer is a terraform.ProviderInitializer that initializes the default AWS provider.
var TerraformProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{ProviderName, RegistryName},
	Provider: func(values map[string]interface{}) (terraform.Provider, error) {
		var regCode region.Code

		r, ok := values["region"]
		if !ok {
			// If no region is defined it means it was passed via ENV variables
			// and it's not tracked on the Plan or HCL so we'll assume the
			// region to be the DefaultRegion
			regCode = DefaultRegion
			log.Logger.Info(fmt.Sprintf("AWS terraform provider region not set, defaulting to %s", DefaultRegion))
			return awstf.NewProvider(ProviderName, regCode)
		}

		switch value := r.(type) {
		case string:
			if regCode == "" {
				log.Logger.Info(fmt.Sprintf("AWS terraform provider region not set, defaulting to %s", DefaultRegion))
				return awstf.NewProvider(ProviderName, DefaultRegion)
			}

			return awstf.NewProvider(ProviderName, region.Code(value))
		default:
			return nil, fmt.Errorf("invalid region type (expected string): %T", r)
		}
	},
}
