package google

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/shopspring/decimal"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// ProviderName uniquely identifies this provider implementation.
const (
	ProviderName = "google"

	// The 2 groups that we need to identify to calculate the
	// machineType full price
	ram = "RAM"
	cpu = "CPU"
)

var (
	// ErrNotSupportedService reports that the service is not supported
	ErrNotSupportedService = errors.New("not supported service")

	// machineFamilies is the list of all the machine families
	// there are out there
	machineFamilies = map[string]struct{}{
		"e2":  struct{}{},
		"n2":  struct{}{},
		"n2d": struct{}{},
		"n1":  struct{}{},
		"t2d": struct{}{},
		"m2":  struct{}{},
		"m1":  struct{}{},
		"c2":  struct{}{},
		"a2":  struct{}{},
	}
)

// Ingester is the entity that will manage the ingestion process from Google
type Ingester struct {
	credentialJSON []byte

	service string
	project string
	zone    string
	region  string

	billing *cloudbilling.APIService
	compute *compute.Service

	ingestionFilter IngestionFilter

	gcpOptions []option.ClientOption

	err error
}

// NewIngester will initialize the Ingester for Google
func NewIngester(ctx context.Context, credentialJSON []byte, service, project, zone string, opts ...Option) (*Ingester, error) {
	sid, ok := services[service]
	if !ok {
		return nil, ErrNotSupportedService
	}

	ing := &Ingester{
		credentialJSON: credentialJSON,
		service:        sid,

		project: project,
		zone:    zone,
		region:  zoneToRegion(zone),

		ingestionFilter: DefaultFilter,
	}

	for _, opt := range opts {
		opt(ing)
	}

	gcpOpts := []option.ClientOption{
		option.WithCredentialsJSON(credentialJSON),
	}
	gcpOpts = append(gcpOpts, ing.gcpOptions...)

	billing, err := cloudbilling.NewService(ctx, gcpOpts...)
	if err != nil {
		return nil, err
	}

	compute, err := compute.NewService(ctx, gcpOpts...)
	if err != nil {
		return nil, err
	}

	ing.billing = billing
	ing.compute = compute

	return ing, nil
}

// Ingest will initialize the process of ingesting and it'll push the price.WithProduct found
// to the returned channel
func (ing *Ingester) Ingest(ctx context.Context, chSize int) <-chan *price.WithProduct {
	results := make(chan *price.WithProduct, chSize)

	go func() {
		defer close(results)

		// Key is the "machineFamily.(cpu|ram)" ex (e2.cpu) and value the price of it
		// it is used after the imports of the SKU as we also import the MachineTypes
		// and to calculate the prices we need this
		var (
			machinteTypePrices = make(map[string]price.Price)
		)

		for sku := range ing.fetchSKUs(ctx) {
			select {
			case <-ctx.Done():
				ing.err = ctx.Err()
				return
			default:
			}

			// If the SKU has no price we do not need it
			// IDK if this is possible but this way we
			// are sure that we'll not ingest it.
			// As it's ordered in chronological order (of sku.PricingInfo) we take
			// the last one as it would be the most current and validate
			// if it has a PricingExpression
			if sku.PricingInfo == nil || len(sku.PricingInfo) == 0 || sku.PricingInfo[len(sku.PricingInfo)-1].PricingExpression == nil {
				continue
			}

			pi := sku.PricingInfo[len(sku.PricingInfo)-1].PricingExpression

			var family, region, service, group, usage string
			if sku.Category != nil {
				family = sku.Category.ResourceFamily
				service = sku.Category.ServiceDisplayName
				group = sku.Category.ResourceGroup
				usage = sku.Category.UsageType
			}
			for _, region = range sku.ServiceRegions {
				// Google has a region named 'global' which means the price is applied to all regions
				// to make it works, we replace it with the current region
				if region == "global" {
					region = ing.region
				}

				// We want to only import the Prices/Products from the
				// region we are importing
				if region != ing.region {
					continue
				}

				prod := &product.Product{
					Provider: ProviderName,
					SKU:      sku.SkuId,
					Service:  service,
					Family:   family,
					Location: region,
					Attributes: map[string]string{
						"group": group,
						"usage": usage,
					},
				}

				// We need to set the machineFamily
				if group == cpu || group == ram {
					mf := strings.ToLower(strings.Split(sku.Description, " ")[0])
					_, ok := machineFamilies[mf]
					if ok {
						prod.Attributes["machine_family"] = mf
					}
				}
				// We check the TieredRates in case it does not have a price
				// provably it'll always have one but just in case
				if len(pi.TieredRates) == 0 {
					continue
				}
				pwp := &price.WithProduct{
					Price: price.Price{
						Unit:     pi.UsageUnit,
						Value:    decimal.NewFromFloat(float64(pi.TieredRates[0].UnitPrice.Units) + (float64(pi.TieredRates[0].UnitPrice.Nanos) / float64(1_000_000_000))),
						Currency: pi.TieredRates[0].UnitPrice.CurrencyCode,
					},
					Product: prod,
				}
				if mf, ok := prod.Attributes["machine_family"]; ok {
					machinteTypePrices[fmt.Sprintf("%s.%s", mf, group)] = pwp.Price
				}
				if ing.ingestionFilter(pwp) {
					results <- pwp
				}
			}
		}

		for mt := range ing.fetchMachineTypes(ctx) {
			mf := strings.ToLower(strings.Split(mt.Name, "-")[0])
			_, ok := machineFamilies[mf]
			if !ok {
				continue
			}

			prod := &product.Product{
				Provider: ProviderName,
				SKU:      fmt.Sprintf("machine-types-%d", mt.Id),
				Service:  "Compute Engine",
				Family:   "Compute",
				Location: ing.region,
				Attributes: map[string]string{
					"machine_type":   mt.Name,
					"group":          "MachineType",
					"cpu":            strconv.Itoa(int(mt.GuestCpus)),
					"ram":            strconv.Itoa(int(mt.MemoryMb)),
					"kind":           mt.Kind,
					"machine_family": mf,
				},
			}

			rp, ok := machinteTypePrices[fmt.Sprintf("%s.%s", mf, ram)]
			if !ok {
				continue
			}

			cp, ok := machinteTypePrices[fmt.Sprintf("%s.%s", mf, cpu)]
			if !ok {
				continue
			}

			// The mt.MemboryMb is in MB and the rp.Unit is on GiBy
			// so we have to convert it to the same unit
			rp.Value = rp.Value.Mul(decimal.NewFromInt(mt.MemoryMb / 1_000))
			cp.Value = cp.Value.Mul(decimal.NewFromInt(mt.GuestCpus))

			// The Unit for the RAM is in GiBy.h, but for the addition
			// we need the same unit which in this case would be h
			if rp.Unit == "GiBy.h" {
				rp.Unit = "h"
			} else {
				// If we cannot check the unit we skip it
				continue
			}
			err := rp.Add(cp)
			if err != nil {
				ing.err = err
				return
			}

			pwp := &price.WithProduct{
				// rp has the end result of the Sum of the 2 prices
				Price:   rp,
				Product: prod,
			}

			if ing.ingestionFilter(pwp) {
				results <- pwp
			}
		}
	}()

	return results
}

// fetchSKUs returns a channel where the Sku are beeing sent and pulls all the
// data from the Skus paginating through all the API
func (ing *Ingester) fetchSKUs(ctx context.Context) <-chan *cloudbilling.Sku {
	results := make(chan *cloudbilling.Sku, 100)

	go func() {
		defer close(results)
		// Docs: https://cloud.google.com/billing/v1/how-tos/catalog-api#getting_the_list_of_skus_for_a_service
		err := cloudbilling.NewServicesSkusService(ing.billing).List(fmt.Sprintf("services/%s", ing.service)).Pages(ctx, func(l *cloudbilling.ListSkusResponse) error {
			for _, sku := range l.Skus {
				results <- sku
			}
			return nil
		})
		if err != nil {
			ing.err = err
		}
	}()

	return results
}

// fetchMachineTypes returns a channel where the machine types are beeing sent and pulls all the
// data from the machine types paginating through all the API
func (ing *Ingester) fetchMachineTypes(ctx context.Context) <-chan *compute.MachineType {
	results := make(chan *compute.MachineType, 100)

	go func() {
		defer close(results)
		// Docs: https://cloud.google.com/compute/docs/reference/rest/v1/machineTypes/list
		err := ing.compute.MachineTypes.List(ing.project, ing.zone).Pages(ctx, func(l *compute.MachineTypeList) error {
			for _, mt := range l.Items {
				results <- mt
			}
			return nil
		})
		if err != nil {
			ing.err = err
		}
	}()

	return results
}

// Err returns any error that might have happened during the ingestion.
func (ing *Ingester) Err() error {
	return ing.err
}

// zoneToRegion will transform a europe-west1-b to europe-west1
func zoneToRegion(z string) string {
	return strings.Join(strings.Split(z, "-")[0:2], "-")
}
