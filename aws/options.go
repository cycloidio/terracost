package aws

import (
	"net/http"
	"time"

	"github.com/machinebox/progress"
)

//go:generate mockgen -destination=../mock/http_client.go -mock_names=HTTPClient=HTTPClient -package mock github.com/cycloidio/terracost/aws HTTPClient

// HTTPClient is an interface of a client that is able to Do HTTP requests
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Option is used to configure the Ingester.
type Option func(ing *Ingester)

// WithPricingURL sets the base AWS pricing URL, "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws" by default.
func WithPricingURL(url string) Option {
	return func(ing *Ingester) {
		ing.pricingURL = url
	}
}

// WithHTTPClient sets a custom HTTP client to be used for offer file downloads.
func WithHTTPClient(client HTTPClient) Option {
	return func(ing *Ingester) {
		ing.httpClient = client
	}
}

// WithBufferSize sets the I/O buffer size for the downloaded file, 100 MiB by default.
func WithBufferSize(size uint) Option {
	return func(ing *Ingester) {
		ing.bufferSize = size
	}
}

// WithProgress sets a channel for receiving progress updates. By default progress is not sent.
func WithProgress(progressCh chan<- progress.Progress, interval time.Duration) Option {
	return func(ing *Ingester) {
		ing.progressCh = progressCh
		ing.progressInterval = interval
	}
}

// WithIngestionFilter sets a custom IngestionFilter to control which pricing data records should be ingested.
func WithIngestionFilter(filter IngestionFilter) Option {
	return func(ing *Ingester) {
		ing.ingestionFilter = filter
	}
}
