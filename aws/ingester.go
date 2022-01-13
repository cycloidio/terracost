package aws

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/machinebox/progress"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/field"
	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
)

// ProviderName uniquely identifies this provider implementation.
const ProviderName = "aws"

const (
	defaultPricingURL = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws"
	defaultBufferSize = 100 * 1024 * 1024 // 100 MiB
)

// Ingester is used to load the pricing data from AWS offer files into a database. It is one-use only and
// should be discarded after the ingestion is complete.
type Ingester struct {
	httpClient HTTPClient
	pricingURL string
	bufferSize uint

	service string
	region  string

	progressCh       chan<- progress.Progress
	progressInterval time.Duration

	ingestionFilter IngestionFilter

	err error
}

// NewIngester returns a new Ingester using the given options. Only the repositories must be provided using the
// WithRepositories function, other configuration options will use their default values.
// The service must be a valid AWS service name that is supported by Terracost, otherwise this function will
// return an error.
func NewIngester(service, region string, options ...Option) (*Ingester, error) {
	if !IsServiceSupported(service) {
		return nil, fmt.Errorf("service not supported: %s", service)
	}

	ing := &Ingester{
		httpClient:      http.DefaultClient,
		pricingURL:      defaultPricingURL,
		bufferSize:      defaultBufferSize,
		service:         service,
		region:          region,
		progressCh:      nil,
		ingestionFilter: DefaultFilter,
	}
	for _, opt := range options {
		opt(ing)
	}

	return ing, nil
}

// Ingest starts a goroutine that reads pricing data from AWS and, for the duration of the context, sends
// the results to the returned channel.
func (ing *Ingester) Ingest(ctx context.Context, chSize int) <-chan *price.WithProduct {
	results := make(chan *price.WithProduct, chSize)

	go func() {
		defer close(results)

		url := ing.pricingURL + "/" + ing.service + "/current/" + ing.region + "/index.csv"
		rc, size, err := ing.download(ctx, url)
		if err != nil {
			ing.err = err
			return
		}
		defer rc.Close()

		// Wrap the ReadCloser in a buffer and progress.Reader to be able to track read progress.
		rd := progress.NewReader(bufio.NewReaderSize(rc, int(ing.bufferSize)))

		if ing.progressCh != nil {
			// Send progress to progressCh with the specified interval.
			go func() {
				ch := progress.NewTicker(ctx, rd, size, ing.progressInterval)
				for p := range ch {
					ing.progressCh <- p
				}
				close(ing.progressCh)
			}()
		}

		csvr := csv.NewReader(rd)

		// The CSV contains rows with different numbers of columns. Namely, the metadata rows only have 2 columns, while
		// the other rows will have much more. This is needed in order to avoid errors from the reader.
		csvr.FieldsPerRecord = -1

		// Read column labels from the first non-metadata row found in the CSV. An offer file always starts with a few
		// lines key-value pairs of metadata that needs to be skipped.
		var columns map[string]int
		for {
			values, err := csvr.Read()
			if err != nil {
				ing.err = err
				return
			}
			if len(values) > 2 {
				columns = readColumnPositions(values)
				break
			}
		}

		// Read through each row in the CSV file and send a price.WithProduct on the results channel.
		for {
			select {
			case <-ctx.Done():
				ing.err = ctx.Err()
				return
			default:
			}

			row, err := csvr.Read()
			if err != nil {
				if err != io.EOF {
					ing.err = err
				}
				return
			}

			data := make(map[field.Field]string)
			for col, index := range columns {
				if f, err := field.FieldString(col); err == nil {
					data[f] = row[index]
				}
			}

			pp, err := newPriceWithProduct(data)
			if err != nil {
				ing.err = err
				return
			}

			if ing.ingestionFilter(pp) {
				results <- pp
			}
		}
	}()

	return results
}

// Err returns any error that might have happened during the ingestion.
func (ing *Ingester) Err() error {
	return ing.err
}

// download is a helper that performs an HTTP GET request and returns the body of the response. The returned
// io.ReadCloser must be manually closed after use.
func (ing *Ingester) download(ctx context.Context, url string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := ing.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, resp.ContentLength, nil
}

// readColumnPositions maps column names to their position in the CSV file.
func readColumnPositions(values []string) map[string]int {
	columns := make(map[string]int)
	for i, v := range values {
		columns[v] = i
	}
	return columns
}

// columnProductToIngest is a mapping from column title to the product.Product attribute name under which the value will
// be stored.

var columnProductToIngest = map[field.Field]string{
	field.CapacityStatus:  "CapacityStatus",
	field.Group:           "Group",
	field.InstanceType:    "InstanceType",
	field.OperatingSystem: "OperatingSystem",
	field.PreInstalledSW:  "PreInstalledSW",
	field.Tenancy:         "Tenancy",
	field.UsageType:       "UsageType",
	field.VolumeAPIName:   "VolumeAPIName",
	field.VolumeType:      "VolumeType",

	// ElastiCache
	field.CacheEngine: "CacheEngine",

	// RDS attributes
	field.DatabaseEngine:   "DatabaseEngine",
	field.DatabaseEdition:  "DatabaseEdition",
	field.DeploymentOption: "DeploymentOption",
	field.LicenseModel:     "LicenseModel",
}

// columnPriceToIngest is a mapping from column title to the price.Price attribute name under which the value will
// be stored.
var columnPriceToIngest = map[field.Field]string{
	field.StartingRange: "StartingRange",
	field.TermType:      "TermType",
}

func newPriceWithProduct(values map[field.Field]string) (*price.WithProduct, error) {
	prod := newProduct(values)

	priceVal, err := decimal.NewFromString(values[field.PricePerUnit])
	if err != nil {
		return nil, fmt.Errorf("failed to parse PricePerUnit: %w", err)
	}

	priceAttrs := map[string]string{}

	for col, attr := range columnPriceToIngest {
		if values[col] != "" {
			priceAttrs[attr] = values[col]
		}
	}

	pwp := &price.WithProduct{
		Price: price.Price{
			Unit:       values[field.Unit],
			Value:      priceVal,
			Currency:   values[field.Currency],
			Attributes: priceAttrs,
		},
		Product: prod,
	}
	return pwp, nil
}

func newProduct(values map[field.Field]string) *product.Product {
	attributes := make(map[string]string)

	for col, attr := range columnProductToIngest {
		if values[col] != "" {
			attributes[attr] = values[col]
		}
	}

	prod := &product.Product{
		Provider:   ProviderName,
		SKU:        values[field.SKU],
		Service:    values[field.ServiceCode],
		Family:     values[field.ProductFamily],
		Location:   region.NewFromName(values[field.Location]).String(),
		Attributes: attributes,
	}
	return prod
}
