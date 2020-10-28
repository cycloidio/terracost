package aws

import (
	awstf "github.com/cycloidio/cost-estimation/aws/terraform"
	"github.com/cycloidio/cost-estimation/terraform"
)

// NewTerraformProvider is a terraform.ProviderInitializer that is used to instantiate an AWS terraform.Provider.
func NewTerraformProvider(config terraform.ProviderConfig) terraform.Provider {
	return awstf.NewProvider(config.Name, config.Expressions["region"].ConstantValue)
}
