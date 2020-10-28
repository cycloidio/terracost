package terraform

type ProviderConfigExpression struct {
	ConstantValue string `json:"constant_value"`
}

type ProviderConfig struct {
	Name        string                              `json:"name"`
	Expressions map[string]ProviderConfigExpression `json:"expressions"`
}

type State struct {
	Values map[string]Module `json:"values"`
}

type Resource struct {
	Address      string                 `json:"address"`
	Index        int                    `json:"index"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	ProviderName string                 `json:"provider_name"`
	Values       map[string]interface{} `json:"values"`
}

type Module struct {
	Resources []Resource `json:"resources"`
}

type Configuration struct {
	ProviderConfig map[string]ProviderConfig `json:"provider_config"`
}
