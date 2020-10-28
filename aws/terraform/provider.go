package terraform

import (
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/terraform"
)

// Provider is an implementation of the terraform.Provider, used to extract component queries from
// terraform resources.
type Provider struct {
	name   string
	region string
}

// NewProvider returns a new Provider.
func NewProvider(name string, region string) *Provider {
	return &Provider{name: name, region: region}
}

func (p *Provider) ProviderName() string { return p.name }

func (p *Provider) ResourceComponents(res terraform.Resource) []query.Component {
	switch res.Type {
	case "aws_instance":
		return p.NewInstance(res.Values).Components()
	case "aws_ebs_volume":
		return p.NewVolume(res.Values).Components()
	default:
		return nil
	}
}
