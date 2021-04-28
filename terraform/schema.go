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
	Address      string                 `json:"address"`
	Index        int                    `json:"index"`
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
	Value string `json:"value"`
}

// ConfigurationModule is used to configure a module.
type ConfigurationModule struct {
	Resources   []ConfigurationResource `json:"resources"`
	ModuleCalls map[string]struct {
		Module *ConfigurationModule `json:"module"`
	} `json:"module_calls"`
}

// ConfigurationResource is used to configure a single reosurce.
type ConfigurationResource struct {
	Address           string `json:"address"`
	ProviderConfigKey string `json:"provider_config_key"`
}
