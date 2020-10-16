package costestimation

import (
	"context"
	"log"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
)

//go:generate go run github.com/golang/mock/mockgen -destination=mock/ingester.go -mock_names=Ingester=Ingester -package mock github.com/cycloidio/cost-estimation Ingester
//go:generate go run github.com/golang/mock/mockgen -destination=mock/backend.go -mock_names=Backend=Backend -package mock github.com/cycloidio/cost-estimation Backend

// Ingester represents a vendor-specific mechanism to load pricing data.
type Ingester interface {
	// Ingest downloads pricing data from a cloud provider and prices with their associated products
	// on the results channel. This function is blocking.
	Ingest(ctx context.Context, results chan<- *price.WithProduct) error
}

// Backend represents a storage method used to store pricing data. It must include concrete implementations
// of all repositories.
type Backend interface {
	Product() product.Repository
	Price() price.Repository
}

// IngestPricing uses the Ingester to load the pricing data and stores it into the Backend.
func IngestPricing(ctx context.Context, backend Backend, ingester Ingester) error {
	ctx, cancel := context.WithCancel(ctx)

	// priceProducts is the channel passed to the Ingester on which the results will be sent. It has a size of 1
	// to block the Ingester as only one price.WithProduct entry can be processed concurrently.
	priceProducts := make(chan *price.WithProduct, 1)

	// done is sent to when priceProducts closes.
	done := make(chan struct{}, 1)

	// sending to errs will finish the ingestion and return the sent value to the caller.
	errs := make(chan error, 1)

	// Goroutine that processes priceProducts sent from the Ingester. It ends in 3 cases:
	// 1) the context is cancelled;
	// 2) an error happened on the backend (sending the error to the errs channel);
	// 3) the priceProducts is closed (sending to the done channel).
	go func() {
		skuProductID := make(map[string]product.ID)
		for {
			select {
			case <-ctx.Done():
				return
			case pp, ok := <-priceProducts:
				if !ok {
					done <- struct{}{}
					return
				}

				if id, ok := skuProductID[pp.Product.SKU]; ok {
					pp.Product.ID = id
				} else {
					var err error
					pp.Product.ID, err = backend.Product().Upsert(ctx, pp.Product)
					if err != nil {
						log.Println("error with product", err)
						errs <- err
						return
					}
					skuProductID[pp.Product.SKU] = pp.Product.ID
				}

				if _, err := backend.Price().Upsert(ctx, pp); err != nil {
					log.Println("error with price", err)
					errs <- err
					return
				}
			}
		}
	}()

	// Start the pricing ingestion, sending the result (error or nil) to the errs channel. This will effectively
	// finish this function.
	go func() {
		errs <- ingester.Ingest(ctx, priceProducts)
	}()

	// Wait for either the context to be cancelled or an error to be sent on the errs channel.
	for {
		select {
		case err := <-errs:
			if err != nil {
				cancel()
			} else {
				<-done
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
