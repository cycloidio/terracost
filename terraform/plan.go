package terraform

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cycloidio/terracost/query"
)

// Plan is a representation of a Terraform plan file.
type Plan struct {
	providerInitializers map[string]ProviderInitializer

	Configuration Configuration     `json:"configuration"`
	PriorState    *State            `json:"prior_state"`
	PlannedValues map[string]Module `json:"planned_values"`
}

// NewPlan returns an empty Plan.
func NewPlan(providerInitializers ...ProviderInitializer) *Plan {
	piMap := make(map[string]ProviderInitializer)
	for _, pi := range providerInitializers {
		for _, name := range pi.MatchNames {
			piMap[name] = pi
		}
	}
	plan := &Plan{providerInitializers: piMap}
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
func (p *Plan) ExtractPlannedQueries() ([]query.Resource, error) {
	providers, err := p.extractProviders()
	if err != nil {
		return nil, fmt.Errorf("unable to extract planned queries: %w", err)
	}
	return p.extractQueries(p.PlannedValues, providers), nil
}

// ExtractPriorQueries extracts a query.Resource slice from the `prior_state` part of the Plan.
func (p *Plan) ExtractPriorQueries() ([]query.Resource, error) {
	providers, err := p.extractProviders()
	if err != nil {
		return nil, fmt.Errorf("unable to extract prior queries: %w", err)
	}
	return p.extractQueries(p.PriorState.Values, providers), nil
}

// extractProviders returns a slice of initialized Provider instances that were found in plan's configuration.
func (p *Plan) extractProviders() (map[string]Provider, error) {
	providers := make(map[string]Provider)
	for alias, provConfig := range p.Configuration.ProviderConfig {
		if pi, ok := p.providerInitializers[provConfig.Name]; ok {
			var err error
			providers[alias], err = pi.Provider(provConfig)
			if err != nil {
				return nil, err
			}
		}
	}
	return providers, nil
}

// extractQueries iterates over every resource and passes each to the corresponding Provider to get the components.
// These are used to form a slice of resource queries that are then returned back to the caller.
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
