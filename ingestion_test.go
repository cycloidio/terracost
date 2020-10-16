package costestimation

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/cost-estimation/mock"
	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
)

func TestIngestPricing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	prod := &product.Product{
		Provider:   "provider",
		SKU:        "prod1",
		Service:    "service",
		Family:     "family",
		Location:   "location",
		Attributes: map[string]string{"key": "value"},
	}
	priceProducts := []*price.WithProduct{
		{
			Product: prod,
			Price: price.Price{
				Unit:       "Hrs",
				Currency:   "USD",
				Value:      decimal.RequireFromString("1.23"),
				Attributes: map[string]string{"purchaseOption": "on_demand"},
			},
		},
		{
			Product: prod,
			Price: price.Price{
				Unit:       "Hrs",
				Currency:   "USD",
				Value:      decimal.RequireFromString("0.98"),
				Attributes: map[string]string{"purchaseOption": "reserved"},
			},
		},
	}

	productRepo := mock.NewProductRepository(ctrl)
	priceRepo := mock.NewPriceRepository(ctrl)
	backend := mock.NewBackend(ctrl)
	ingester := mock.NewIngester(ctrl)

	backend.EXPECT().Product().AnyTimes().Return(productRepo)
	backend.EXPECT().Price().AnyTimes().Return(priceRepo)

	ingester.EXPECT().Ingest(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, results chan<- *price.WithProduct) error {
		for _, pp := range priceProducts {
			results <- pp
		}
		close(results)
		return nil
	})

	productRepo.EXPECT().Upsert(gomock.Any(), prod).Return(product.ID(1), nil)
	priceRepo.EXPECT().Upsert(gomock.Any(), priceProducts[0]).Return(price.ID(1), nil)
	priceRepo.EXPECT().Upsert(gomock.Any(), priceProducts[1]).Return(price.ID(2), nil)

	err := IngestPricing(context.Background(), backend, ingester)
	require.NoError(t, err)
}
