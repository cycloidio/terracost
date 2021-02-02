package terraform

// ProviderConfigExpression is a single configuration variable of a ProviderConfig.
type ProviderConfigExpression struct {
	ConstantValue string   `json:"constant_value"`
	References    []string `json:"references"`
}

// ProviderConfig is configuration of a provider with the given Name.
type ProviderConfig struct {
	Name        string                              `json:"name"`
	Alias       string                              `json:"alias"`
	Expressions map[string]ProviderConfigExpression `json:"expressions"`
}

// State is a collection of resource modules.
type State struct {
	Values map[string]Module `json:"values"`
}

// Resource is a single Terraform resource definition.
type Resource struct {
	Address      string                 `json:"address"`
	Index        int                    `json:"index"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	ProviderName string                 `json:"provider_name"`
	Values       map[string]interface{} `json:"values"`
}

// Module is a collection of resources.
type Module struct {
	Resources []Resource `json:"resources"`
}

// Configuration is a Terraform plan configuration.
type Configuration struct {
	ProviderConfig map[string]ProviderConfig `json:"provider_config"`
}

// Variable is a Terraform variable declaration.
type Variable struct {
	Value string `json:"value"`
}
