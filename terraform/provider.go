package terraform

import (
	"github.com/cycloidio/terracost/query"
)

//go:generate mockgen -destination=../mock/terraform_provider.go -mock_names=Provider=TerraformProvider -package mock github.com/cycloidio/terracost/terraform Provider

// Provider represents a Terraform provider. It extracts price queries from Terraform resources.
type Provider interface {
	// Name returns the common name of this Provider.
	Name() string

	// ResourceComponents returns price component queries for the given Resource. Nil may be returned
	// which signifies a resource that is not supported by this Provider.
	// It also expects all the resources in case it needs to check the configuration of another
	// resource
	ResourceComponents(rss map[string]Resource, res Resource) []query.Component
}

// ProviderInitializer is used to initialize a Provider for each provider name that matches one of the MatchNames.
type ProviderInitializer struct {
	// MatchNames contains the names that this ProviderInitializer will match. Most providers will only
	// have one name (such as `aws`) but some might use multiple names to refer to the same provider
	// implementation (such as `google` and `google-beta`).
	MatchNames []string

	// Provider initializes a Provider instance given the values defined in the config and returns it.
	// If a provider must be ignored (related to version constraints, etc), please return nil to avoid using it.
	Provider func(values map[string]interface{}) (Provider, error)
}

// validateProviders will verify that at least one of the queries is from a known provider
// if none matches an error will be returned to stop the processing
func validateProviders(queries []query.Resource, providers map[string]Provider) error {
	for _, q := range queries {
		_, ok := providers[q.Provider]
		if ok {
			return nil
		}
	}
	return ErrNoKnownProvider
}
