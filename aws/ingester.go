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

	"github.com/cycloidio/cost-estimation/aws/field"
	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/util"
)

// VendorName uniquely identifies this vendor implementation.
const VendorName = "aws"

const (
	defaultPricingURL = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws"
	defaultBufferSize = 100 * 1024 * 1024 // 100 MiB
)

// Ingester is used to load the pricing data from AWS offer files into a database.
type Ingester struct {
	httpClient util.HTTPClient
	pricingURL string
	bufferSize uint

	service string
	region  string

	progressCh       chan<- progress.Progress
	progressInterval time.Duration
}

// NewIngester returns a new Ingester using the given options. Only the repositories must be provided using the
// WithRepositories function, other configuration options will use their default values.
func NewIngester(service, region string, options ...Option) *Ingester {
	ing := &Ingester{
		httpClient: &http.Client{},
		pricingURL: defaultPricingURL,
		bufferSize: defaultBufferSize,
		service:    service,
		region:     region,
		progressCh: nil,
	}
	for _, opt := range options {
		opt(ing)
	}

	return ing
}

// Ingest reads pricing data from AWS and sends the results to the provided channel.
func (ing *Ingester) Ingest(ctx context.Context, results chan<- *price.WithProduct) error {
	url := ing.pricingURL + "/" + ing.service + "/current/" + ing.region + "/index.csv"
	rc, size, err := ing.download(ctx, url)
	if err != nil {
		return err
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
	csvr.FieldsPerRecord = -1

	// Read column labels from the first non-metadata row found in the CSV. An offer file always starts with a few
	// lines key-value pairs of metadata that needs to be skipped.
	var columns map[string]int
	for {
		values, err := csvr.Read()
		if err != nil {
			return err
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
			return ctx.Err()
		default:
		}

		row, err := csvr.Read()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		data := make(map[field.Field]string)
		for col, index := range columns {
			if f, err := field.FieldString(col); err == nil {
				data[f] = row[index]
			}
		}

		pp, err := newPriceWithProduct(data)
		if err != nil {
			return err
		}

		results <- pp
	}

	close(results)
	return nil
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

// columnToAttribute is a mapping from column title to the product.Product attribute name under which the value will
// be stored.
var columnToAttribute = map[field.Field]string{
	field.UsageType:       "usagetype",
	field.InstanceType:    "instanceType",
	field.OperatingSystem: "operatingSystem",
	field.PreInstalledSW:  "preInstalledSw",
	field.CapacityStatus:  "capacitystatus",
	field.Tenancy:         "tenancy",
	field.VolumeAPIName:   "volumeApiName",
	field.StorageMedia:    "storageMedia",
}

// columnToPriceAttribute is a mapping from column title to the price.Price attribute name under which the value will
// be stored.
var columnToPriceAttribute = map[field.Field]string{
	field.PriceDescription:   "description",
	field.StartingRange:      "startUsageAmount",
	field.EndingRange:        "endUsageAmount",
	field.TermLength:         "termLength",
	field.TermPurchaseOption: "termPurchaseOption",
	field.TermOfferingClass:  "termOfferingClass",
	field.EffectiveDate:      "effectiveDateStart",
}

// purchaseOptions is a mapping from the values used in the CSV file to the expected values.
var purchaseOptions = map[string]string{
	"OnDemand": "on_demand",
	"Reserved": "reserved",
}

func newPriceWithProduct(values map[field.Field]string) (*price.WithProduct, error) {
	prod := newProduct(values)

	priceVal, err := decimal.NewFromString(values[field.PricePerUnit])
	if err != nil {
		return nil, fmt.Errorf("failed to parse PricePerUnit: %w", err)
	}

	priceAttrs := map[string]string{
		"purchaseOption": purchaseOptions[values[field.PurchaseOption]],
	}
	for col, attr := range columnToPriceAttribute {
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
	for col, attr := range columnToAttribute {
		if values[col] != "" {
			attributes[attr] = values[col]
		}
	}

	prod := &product.Product{
		Provider:   VendorName,
		SKU:        values[field.SKU],
		Service:    values[field.ServiceCode],
		Family:     values[field.ProductFamily],
		Location:   regionMap[values[field.Location]],
		Attributes: attributes,
	}
	return prod
}
