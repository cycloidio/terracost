package azurerm

// Option is used to configure the Ingester.
type Option func(ing *Ingester)

// WithIngestionFilter sets a custom IngestionFilter to control which pricing data records should be ingested.
func WithIngestionFilter(filter IngestionFilter) Option {
	return func(ing *Ingester) {
		ing.ingestionFilter = filter
	}
}

// WithEndpoint sets a custom endpoint to user for the api calls
func WithEndpoint(endpoint string) Option {
	return func(ing *Ingester) {
		ing.endpoint = endpoint
	}
}
