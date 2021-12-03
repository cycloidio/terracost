package google

import "google.golang.org/api/option"

// Option is used to configure the Ingester.
type Option func(ing *Ingester)

// WithIngestionFilter sets a custom IngestionFilter to control which pricing data records should be ingested.
func WithIngestionFilter(filter IngestionFilter) Option {
	return func(ing *Ingester) {
		ing.ingestionFilter = filter
	}
}

// WithGCPOption will proxy the GCP ClientOption to the
// service initializations
func WithGCPOption(opts ...option.ClientOption) Option {
	return func(ing *Ingester) {
		ing.gcpOptions = opts
	}
}
