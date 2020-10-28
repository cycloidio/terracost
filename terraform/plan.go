package terraform

import (
	"encoding/json"
	"io"

	"github.com/cycloidio/cost-estimation/query"
)

// Plan is a representation of a Terraform plan file.
type Plan struct {
	providerInitializers map[string]ProviderInitializer

	Configuration Configuration     `json:"configuration"`
	PriorState    *State            `json:"prior_state"`
	PlannedValues map[string]Module `json:"planned_values"`
}

// NewPlan returns an empty Plan.
func NewPlan(opts ...Option) *Plan {
	plan := &Plan{providerInitializers: make(map[string]ProviderInitializer)}
	for _, opt := range opts {
		opt(plan)
	}
	return plan
}

// Read reads the Plan file from the provider io.Reader.
func (p *Plan) Read(r io.Reader) error {
	if err := json.NewDecoder(r).Decode(p); err != nil {
		return err
	}
	return nil
}

// ExtractPlannedQueries extracts a query.Resource slice from the `planned_values` part of the Plan.
func (p *Plan) ExtractPlannedQueries() []query.Resource {
	providers := p.extractProviders()
	return p.extractQueries(p.PlannedValues, providers)
}

// ExtractPriorQueries extracts a query.Resource slice from the `prior_state` part of the Plan.
func (p *Plan) ExtractPriorQueries() []query.Resource {
	providers := p.extractProviders()
	return p.extractQueries(p.PriorState.Values, providers)
}

func (p *Plan) extractProviders() map[string]Provider {
	providers := make(map[string]Provider)
	for alias, provConfig := range p.Configuration.ProviderConfig {
		if initializer, ok := p.providerInitializers[provConfig.Name]; ok {
			providers[alias] = initializer(provConfig)
		}
	}
	return providers
}

func (p *Plan) extractQueries(modules map[string]Module, providers map[string]Provider) []query.Resource {
	result := make([]query.Resource, 0)
	for _, module := range modules {
		for _, tfres := range module.Resources {
			if provider, ok := providers[tfres.ProviderName]; ok {
				comps := provider.ResourceComponents(tfres)
				if comps != nil {
					q := query.Resource{
						Address:    tfres.Address,
						Components: comps,
					}
					result = append(result, q)
				}
			}
		}
	}
	return result
}
