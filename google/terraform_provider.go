package google

import (
	"fmt"

	googletf "github.com/cycloidio/terracost/google/terraform"
	"github.com/cycloidio/terracost/terraform"
)

// RegistryName is the fully qualified name under which this provider is stored in the registry.
const RegistryName = "registry.terraform.io/hashicorp/google"

// TerraformProviderInitializer is a terraform.ProviderInitializer that initializes the default GCP provider.
var TerraformProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{ProviderName, RegistryName},
	Provider: func(values map[string]interface{}) (terraform.Provider, error) {
		z, ok := values["zone"]
		if !ok {
			return nil, nil
		}
		region, err := zoneToRegion(z.(string))
		if err != nil {
			return nil, fmt.Errorf("unable to get region from zone: %w", err)
		}
		return googletf.NewProvider(ProviderName, region)
	},
}
