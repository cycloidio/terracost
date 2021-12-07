package terracost

import (
	"github.com/cycloidio/terracost/aws"
	"github.com/cycloidio/terracost/azurerm"
	"github.com/cycloidio/terracost/google"
	"github.com/cycloidio/terracost/terraform"
)

// defaultProviders are the currently known and supported terraform providers
var defaultProviders = []terraform.ProviderInitializer{
	aws.TerraformProviderInitializer,
	google.TerraformProviderInitializer,
	azurerm.TerraformProviderInitializer,
}

// getDefaultProviders will return the default supported providers of terracost
func getDefaultProviders() []terraform.ProviderInitializer {
	return defaultProviders
}
