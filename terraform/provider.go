package terraform

import (
	"github.com/cycloidio/terracost/query"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../mock/terraform_provider.go -mock_names=Provider=TerraformProvider -package mock github.com/cycloidio/terracost/terraform Provider

// Provider represents a Terraform provider. It extracts price queries from Terraform resources.
type Provider interface {
	// ResourceComponents returns price component queries for the given Resource. Nil may be returned
	// which signifies a resource that is not supported by this Provider.
	ResourceComponents(res Resource) []query.Component
}

// ProviderInitializer is used to initialize a Provider for each provider name that matches one of the MatchNames.
type ProviderInitializer struct {
	// MatchNames contains the names that this ProviderInitializer will match. Most providers will only
	// have one name (such as `aws`) but some might use multiple names to refer to the same provider
	// implementation (such as `google` and `google-beta`).
	MatchNames []string

	// Provider initializes a Provider instance given the values defined in the config and returns it.
	Provider func(values map[string]string) (Provider, error)
}
