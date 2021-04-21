package terraform

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cycloidio/terracost/query"
)

// Plan is a representation of a Terraform plan file.
type Plan struct {
	providerInitializers map[string]ProviderInitializer

	Configuration Configuration       `json:"configuration"`
	PriorState    *State              `json:"prior_state"`
	PlannedValues map[string]Module   `json:"planned_values"`
	Variables     map[string]Variable `json:"variables"`
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
	if p.PriorState == nil {
		return []query.Resource{}, nil
	}
	providers, err := p.extractProviders()
	if err != nil {
		return nil, fmt.Errorf("unable to extract prior queries: %w", err)
	}
	return p.extractQueries(p.PriorState.Values, providers), nil
}

// extractProviders returns a slice of initialized Provider instances that were found in plan's configuration.
func (p *Plan) extractProviders() (map[string]Provider, error) {
	providers := make(map[string]Provider)
	for name, provConfig := range p.Configuration.ProviderConfig {
		if pi, ok := p.providerInitializers[provConfig.Name]; ok {
			values, err := p.evaluateProviderConfigExpressions(provConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to read config of provider %q: %w", name, err)
			}
			prov, err := pi.Provider(values)
			if err != nil {
				return nil, err
			}
			providers[name] = prov
		}
	}
	return providers, nil
}

// extractQueries iterates over every resource and passes each to the corresponding Provider to get the components.
// These are used to form a slice of resource queries that are then returned back to the caller.
func (p *Plan) extractQueries(modules map[string]Module, providers map[string]Provider) []query.Resource {
	// Create a map to associate each resource with a key of the provider that
	// should be used to estimate it.
	resToProvKey := make(map[string]string)
	for _, res := range p.Configuration.RootModule.Resources {
		resToProvKey[res.Address] = res.ProviderConfigKey
	}

	result := make([]query.Resource, 0)
	for _, module := range modules {
		for _, tfres := range module.Resources {
			providerKey, ok := resToProvKey[tfres.Address]
			if !ok {
				providerKey = tfres.ProviderName
			}

			if provider, ok := providers[providerKey]; ok {
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

// evaluateProviderConfigExpressions returns evaluated values of provider's configuration block, whether a constant
// value or reference to a variable.
func (p *Plan) evaluateProviderConfigExpressions(config ProviderConfig) (map[string]string, error) {
	values := make(map[string]string)
	for name, e := range config.Expressions {
		if e.ConstantValue != "" {
			values[name] = e.ConstantValue
			continue
		}

		if len(e.References) < 1 {
			return nil, fmt.Errorf("config expression contains invalid reference")
		}

		ref := strings.Split(e.References[0], ".")
		if len(ref) < 2 {
			return nil, fmt.Errorf("config expression contains invalid reference")
		}

		varName := ref[1]
		v, ok := p.Variables[varName]
		if !ok || v.Value == "" {
			return nil, fmt.Errorf("required variable %q is not defined", varName)
		}
		values[name] = v.Value
	}
	return values, nil
}
