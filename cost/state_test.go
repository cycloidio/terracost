package cost_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/mock"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
)

func TestNewState(t *testing.T) {
	queries := []query.Resource{
		{
			Address: "aws_instance.test1",
			Components: []query.Component{
				{
					Name:           "Compute",
					HourlyQuantity: decimal.NewFromInt(1),
					ProductFilter: &product.Filter{
						Provider: util.StringPtr("aws"),
						Service:  util.StringPtr("AmazonEC2"),
						Family:   util.StringPtr("Compute Instance"),
						Location: util.StringPtr("eu-west-3"),
						AttributeFilters: []*product.AttributeFilter{
							{Key: "instanceType", Value: util.StringPtr("t3.micro")},
						},
					},
				},
			},
		},
		{
			Address:    "aws_invalid_resource.skipped",
			Components: nil,
		},
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		productRepo := mock.NewProductRepository(ctrl)
		priceRepo := mock.NewPriceRepository(ctrl)
		backend := mock.NewBackend(ctrl)
		backend.EXPECT().Products().AnyTimes().Return(productRepo)
		backend.EXPECT().Prices().AnyTimes().Return(priceRepo)

		prod1 := &product.Product{ID: product.ID(1)}
		productRepo.EXPECT().Filter(ctx, queries[0].Components[0].ProductFilter).Return([]*product.Product{prod1}, nil)
		prc1 := &price.Price{Value: decimal.NewFromFloat(1.23), Unit: "Hrs", Currency: "USD"}
		priceRepo.EXPECT().Filter(ctx, prod1.ID, queries[0].Components[0].PriceFilter).Return([]*price.Price{prc1}, nil)

		expected := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test1": {
					Components: map[string]cost.Component{
						"Compute": {
							Rate:     cost.NewMonthly(decimal.New(89790, -2), "USD"),
							Quantity: decimal.NewFromInt(1),
						},
					},
				},
				"aws_invalid_resource.skipped": {
					Skipped: true,
				},
			},
		}

		state, err := cost.NewState(ctx, backend, queries)
		require.NoError(t, err)
		assert.Equal(t, expected, state)
	})

	t.Run("ProductRepositoryFailure", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		productRepo := mock.NewProductRepository(ctrl)
		priceRepo := mock.NewPriceRepository(ctrl)
		backend := mock.NewBackend(ctrl)
		backend.EXPECT().Products().AnyTimes().Return(productRepo)
		backend.EXPECT().Prices().AnyTimes().Return(priceRepo)

		productRepo.EXPECT().Filter(ctx, queries[0].Components[0].ProductFilter).Return(nil, errors.New("repo fail"))

		state, err := cost.NewState(ctx, backend, queries)
		require.NoError(t, err)
		assert.Error(t, state.Resources["aws_instance.test1"].Components["Compute"].Error)
	})

	t.Run("PriceRepositoryFailure", func(t *testing.T) {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		productRepo := mock.NewProductRepository(ctrl)
		priceRepo := mock.NewPriceRepository(ctrl)
		backend := mock.NewBackend(ctrl)
		backend.EXPECT().Products().AnyTimes().Return(productRepo)
		backend.EXPECT().Prices().AnyTimes().Return(priceRepo)

		prod1 := &product.Product{ID: product.ID(1)}
		productRepo.EXPECT().Filter(ctx, queries[0].Components[0].ProductFilter).Return([]*product.Product{prod1}, nil)
		priceRepo.EXPECT().Filter(ctx, prod1.ID, queries[0].Components[0].PriceFilter).Return(nil, errors.New("repo fail"))

		state, err := cost.NewState(ctx, backend, queries)
		require.NoError(t, err)
		assert.Error(t, state.Resources["aws_instance.test1"].Components["Compute"].Error)
	})
}

func TestState_Cost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		state := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test1": {
					Components: map[string]cost.Component{
						"Compute": {
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
							Quantity: decimal.NewFromInt(730),
						},
					},
				},
			},
		}

		expected := cost.NewMonthly(decimal.New(89790, -2), "USD")
		actual, err := state.Cost()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("ComponentCostMismatch", func(t *testing.T) {
		state := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test1": {
					Components: map[string]cost.Component{
						"Compute": {
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
							Quantity: decimal.NewFromInt(730),
						},
						"Storage": {
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "EUR"),
							Quantity: decimal.NewFromInt(730),
						},
					},
				},
			},
		}

		actual, err := state.Cost()
		assert.Error(t, err)
		assert.Equal(t, cost.Zero, actual)
	})
}
