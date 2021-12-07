package azurerm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/shopspring/decimal"
)

// ProviderName is the provider that this package implements
const ProviderName = "azurerm"

var (
	// The list of all services is https://azure.microsoft.com/en-us/services/, the left side is
	// the Family and the main content is the Services
	services = map[string]struct{}{
		VirtualMachines.String(): struct{}{},
	}

	// ErrNotSupportedService reports that the service is not supported
	ErrNotSupportedService = errors.New("not supported service")
)

// Ingester is the entity that will manage the ingestion process from AzureRM
type Ingester struct {
	service string
	region  string

	client *http.Client

	ingestionFilter IngestionFilter
	endpoint        string
	endpointURL     *url.URL

	err error
}

// NewIngester returns a new Ingester for AzureRM for the specified service and region (ex: francecentral) with the
// given options
func NewIngester(ctx context.Context, service, region string, opts ...Option) (*Ingester, error) {
	if _, ok := services[service]; !ok {
		return nil, ErrNotSupportedService
	}
	ing := &Ingester{
		client:          http.DefaultClient,
		region:          region,
		service:         service,
		endpoint:        "https://prices.azure.com/",
		ingestionFilter: DefaultFilter,
	}

	for _, opt := range opts {
		opt(ing)
	}

	u, err := url.Parse(ing.endpoint)
	if err != nil {
		return nil, err
	}
	ing.endpointURL = u

	return ing, nil
}

// Ingest will initialize the process of ingesting and it'll push the price.WithProduct found
// to the returned channel
func (ing *Ingester) Ingest(ctx context.Context, chSize int) <-chan *price.WithProduct {
	results := make(chan *price.WithProduct, chSize)
	go func() {
		defer close(results)

		for rp := range ing.fetchPrices(ctx) {
			priority := "regular"
			if strings.HasSuffix(rp.MeterName, " Spot") {
				priority = "spot"
			} else if strings.HasSuffix(rp.MeterName, " Low Priority") {
				priority = "low"
			}
			prod := &product.Product{
				Provider: ProviderName,
				SKU:      rp.SkuID,
				Service:  rp.ServiceName,
				Family:   rp.ServiceFamily,
				Location: rp.ArmRegionName,
				Attributes: map[string]string{
					"arm_sku_name": rp.ArmSkuName,
					"product_name": rp.ProductName,
					"sku_name":     rp.SkuName,
					"priority":     priority,
				},
			}
			pwp := &price.WithProduct{
				Price: price.Price{
					Unit:     rp.UnitOfMeasure,
					Value:    decimal.NewFromFloat(rp.UnitPrice),
					Currency: rp.CurrencyCode,
					Attributes: map[string]string{
						"type": rp.Type,
					},
				},
				Product: prod,
			}
			if ing.ingestionFilter(pwp) {
				results <- pwp
			}
		}
	}()

	return results
}

func (ing *Ingester) fetchPrices(ctx context.Context) <-chan retailPrice {
	results := make(chan retailPrice, 100)

	go func() {
		defer close(results)
		// Docs: https://docs.microsoft.com/en-us/rest/api/cost-management/retail-prices/azure-retail-prices
		f := url.PathEscape(fmt.Sprintf("serviceName eq '%s' and armRegionName eq '%s'", ing.service, ing.region))
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?$filter=%s", ing.buildPricesURL(), f), nil)
		if err != nil {
			ing.err = err
			return
		}

		for req != nil {
			res, err := ing.client.Do(req)
			if err != nil {
				ing.err = err
				return
			}

			var rps retailPricesResponse
			err = json.NewDecoder(res.Body).Decode(&rps)
			if err != nil {
				ing.err = err
				return
			}
			res.Body.Close()

			for _, rp := range rps.Items {
				results <- rp
			}

			if rps.NextPageLink == "" {
				req = nil
			} else {
				req, _ = http.NewRequestWithContext(ctx, http.MethodGet, rps.NextPageLink, nil)
			}
		}
	}()

	return results
}

// buildPricesURL will build the prices url from the endpoint defined
// and the path to the api endpoint
func (ing *Ingester) buildPricesURL() string {
	path, _ := url.Parse("./api/retail/prices")
	return ing.endpointURL.ResolveReference(path).String()
}

// Err returns any error that might have happened during the ingestion.
func (ing *Ingester) Err() error {
	return ing.err
}
