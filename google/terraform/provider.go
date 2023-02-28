package terraform

import (
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
)

// Provider is an implementation of the terraform.Provider, used to extract component queries from
// terraform resources.
type Provider struct {
	key    string
	region string
}

// NewProvider initializes a new Google provider with key and region
func NewProvider(key, region string) (*Provider, error) {
	return &Provider{
		key:    key,
		region: region,
	}, nil
}

// Name returns the Provider's common name.
func (p *Provider) Name() string { return p.key }

// ResourceComponents returns Component queries for a given terraform.Resource.
func (p *Provider) ResourceComponents(rss map[string]terraform.Resource, tfRes terraform.Resource) []query.Component {
	switch tfRes.Type {
	case "google_compute_instance":
		vals, err := decodeComputeInstanceValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newComputeInstance(vals).Components()
	default:
		return nil
	}
}
