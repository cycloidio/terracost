package terraform

import "encoding/json"

// ProviderConfigExpression is a single configuration variable of a ProviderConfig.
type ProviderConfigExpression struct {
	ConstantValue string   `json:"constant_value",mapstructure:"constant_value"`
	References    []string `json:"references",mapstructure:"references"`
}

// ProviderConfig is configuration of a provider with the given Name.
type ProviderConfig struct {
	Name        string                              `json:"name"`
	Alias       string                              `json:"alias"`
	Expressions map[string]ProviderConfigExpression `json:"expressions"`
}

// UnmarshalJSON handles the logic of Unmarshaling a ProviderConfig
// as we have some edge cases we want to not unmarshal as they
// are not standard/needed and would make things more complex
func (cfg *ProviderConfig) UnmarshalJSON(b []byte) error {
	var s struct {
		Name        string                 `json:"name"`
		Alias       string                 `json:"alias"`
		Expressions map[string]interface{} `json:"expressions"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	cfg.Name = s.Name
	cfg.Alias = s.Alias
	cfg.Expressions = make(map[string]ProviderConfigExpression)

	// For now we only want the ones that are structs and
	// not arrays if we need those later one we'll need
	// to change the type from map to slice
	for k, v := range s.Expressions {
		switch val := v.(type) {
		case []interface{}:
			// Ignore the [] types
			break
		case map[string]interface{}:
			// On the normal case we marshal and
			// unmarshal again the struct to let
			// json lib do the rest
			bv, err := json.Marshal(val)
			if err != nil {
				return err
			}

			var e ProviderConfigExpression
			json.Unmarshal(bv, &e)
			cfg.Expressions[k] = e
		}
	}

	return nil
}

// Values is a tree of modules and resources within.
type Values struct {
	RootModule Module `json:"root_module"`
}

// State is a collection of resource modules.
type State struct {
	Values Values `json:"values"`
}

// Resource is a single Terraform resource definition.
type Resource struct {
	Address string `json:"address"`
	// This value is 99% of the time an integer, but it can also be
	// a string, the implementation on TF side is of 'addrs.InstanceKey'
	// which can be of type 'IntKey' or 'StringKey'
	Index        interface{}            `json:"index"`
	Mode         string                 `json:"mode"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	ProviderName string                 `json:"provider_name"`
	Values       map[string]interface{} `json:"values"`
}

// Module is a collection of resources.
type Module struct {
	Address      string     `json:"address"`
	Resources    []Resource `json:"resources"`
	ChildModules []*Module  `json:"child_modules"`
}

// Configuration is a Terraform plan configuration.
type Configuration struct {
	ProviderConfig map[string]ProviderConfig `json:"provider_config"`
	RootModule     ConfigurationModule       `json:"root_module"`
}

// Variable is a Terraform variable declaration.
type Variable struct {
	Value interface{} `json:"value"`
}

// ConfigurationModule is used to configure a module.
type ConfigurationModule struct {
	Resources   []ConfigurationResource `json:"resources"`
	Variables   map[string]Variable     `json:"variables"`
	ModuleCalls map[string]struct {
		Module *ConfigurationModule `json:"module"`
	} `json:"module_calls"`
}

// ConfigurationResource is used to configure a single reosurce.
type ConfigurationResource struct {
	Address           string `json:"address"`
	ProviderConfigKey string `json:"provider_config_key"`
	// Expressions are really similar to the ProviderConfigExpression but we cannot use it (as map[string]ProviderConfigExpression)
	// as some examples do not match, some are not map[string]ProviderConfigExpression but map[string][]interface{} and some constant_value are not
	// string but other types
	Expressions map[string]interface{} `json:"expressions"`
}
