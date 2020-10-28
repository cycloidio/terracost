package terraform

// Option is used to configure Plan reading process.
type Option func(*Plan)

// WithProvider associates the name with the particular ProviderInitializer. When the provider with this name is
// found in the plan file, the Provider is instantiated and used to retrieve the resources from the plan.
func WithProvider(name string, initializer ProviderInitializer) Option {
	return func(p *Plan) {
		p.providerInitializers[name] = initializer
	}
}
