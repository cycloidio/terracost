package cost_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/cost-estimation/cost"
	"github.com/cycloidio/cost-estimation/mock"
	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/util"
)

func TestNewState(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	productRepo := mock.NewProductRepository(ctrl)
	priceRepo := mock.NewPriceRepository(ctrl)
	backend := mock.NewBackend(ctrl)
	backend.EXPECT().Product().AnyTimes().Return(productRepo)
	backend.EXPECT().Price().AnyTimes().Return(priceRepo)

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
	}

	prod1 := &product.Product{ID: product.ID(1)}
	productRepo.EXPECT().Filter(ctx, queries[0].Components[0].ProductFilter).Return([]*product.Product{prod1}, nil)
	prc1 := &price.Price{Value: decimal.NewFromFloat(1.23), Unit: "Hrs"}
	priceRepo.EXPECT().Filter(ctx, prod1.ID, queries[0].Components[0].PriceFilter).Return([]*price.Price{prc1}, nil)

	expected := &cost.State{
		Resources: map[string]cost.Resource{
			"aws_instance.test1": {
				Components: map[string]cost.Component{
					"Compute": {
						Rate:     decimal.New(89790, -2),
						Quantity: decimal.NewFromInt(1),
					},
				},
			},
		},
	}

	state, err := cost.NewState(ctx, backend, queries)
	require.NoError(t, err)
	assert.Equal(t, expected, state)
}

func TestState_Cost(t *testing.T) {
	state := &cost.State{
		Resources: map[string]cost.Resource{
			"aws_instance.test1": {
				Components: map[string]cost.Component{
					"Compute": {
						Rate:     decimal.NewFromFloat(1.23),
						Quantity: decimal.NewFromInt(730),
					},
				},
			},
		},
	}

	expected := decimal.NewFromFloat(897.9)
	assert.True(t, expected.Equal(state.Cost()))
}
