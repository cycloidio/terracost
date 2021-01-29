package aws

import (
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/terraform"
)

// TerraformProviderInitializer is a terraform.ProviderInitializer that initializes the default AWS provider.
var TerraformProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{ProviderName},
	Provider: func(config terraform.ProviderConfig) (terraform.Provider, error) {
		regCode := region.Code(config.Expressions["region"].ConstantValue)
		return awstf.NewProvider(ProviderName, regCode)
	},
}
