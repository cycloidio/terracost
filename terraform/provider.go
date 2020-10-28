package terraform

import (
	"github.com/cycloidio/cost-estimation/query"
)

type ProviderInitializer func(config ProviderConfig) Provider

//go:generate go run github.com/golang/mock/mockgen -destination=../mock/terraform_provider.go -mock_names=Provider=TerraformProvider -package mock github.com/cycloidio/cost-estimation/terraform Provider

// Provider represents a Terraform provider. It extracts price queries from Terraform resources.
type Provider interface {
	// ProviderName returns the name of the Provider.
	ProviderName() string

	// ResourceComponents returns price component queries for the given Resource. Nil may be returned
	// which signifies a resource that is not supported by this Provider.
	ResourceComponents(res Resource) []query.Component
}
