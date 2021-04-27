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
	PlannedValues Values              `json:"planned_values"`
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
func (p *Plan) extractQueries(values Values, providers map[string]Provider) []query.Resource {
	// Create a map to associate each resource with a Provider that
	// should be used to estimate it.
	resourceProviders := make(map[string]Provider)
	p.extractModuleConfiguration("", &p.Configuration.RootModule, providers, resourceProviders)
	return p.extractModuleQueries(&values.RootModule, resourceProviders)
}

// extractModuleConfiguration iterates over all the modules included in the plan's configuration block and
// extracts the provider that should be used for each resource. This function calls itself recursively until
// data from the entire module tree is extracted. It takes the following arguments:
//   - prefix - the current module's address. Empty string signifies the root module.
//   - module - the module's configuration block itself.
//   - providers - map of provider name to Provider.
//   - resourceProviders - used as an output of this function, it's a map of resource addresses to their assigned
//     Provider. This map should be passed empty and not nil.
func (p *Plan) extractModuleConfiguration(prefix string, module *ConfigurationModule, providers map[string]Provider, resourceProviders map[string]Provider) {
	for _, res := range module.Resources {
		key := res.ProviderConfigKey
		if strings.Contains(key, ":") {
			parts := strings.Split(key, ":")
			key = parts[len(parts)-1]
		}

		addr := res.Address
		if prefix != "" {
			addr = fmt.Sprintf("module.%s.%s", prefix, addr)
		}

		if prov, ok := providers[key]; ok {
			resourceProviders[addr] = prov
		}
	}

	for k, child := range module.ModuleCalls {
		if child.Module != nil {
			nextPrefix := k
			if prefix != "" {
				nextPrefix = fmt.Sprintf("%s.%s", prefix, k)
			}
			p.extractModuleConfiguration(nextPrefix, child.Module, providers, resourceProviders)
		}
	}
}

// extractModuleQueries iterates recursively over all the module's (and its descendants) resources. It uses the
// resourceProviders map to retrieve the correct Provider based on the resource address.
func (p *Plan) extractModuleQueries(module *Module, resourceProviders map[string]Provider) []query.Resource {
	result := make([]query.Resource, 0, len(resourceProviders))

	for _, tfres := range module.Resources {
		provider, ok := resourceProviders[tfres.Address]
		if !ok || tfres.Mode != "managed" {
			continue
		}

		comps := provider.ResourceComponents(tfres)
		q := query.Resource{
			Address:    tfres.Address,
			Components: comps,
		}
		result = append(result, q)
	}

	for _, child := range module.ChildModules {
		result = append(result, p.extractModuleQueries(child, resourceProviders)...)
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
