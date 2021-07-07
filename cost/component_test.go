package cost_test

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/cycloidio/terracost/cost"
)

func TestComponentDiff_PriorCost(t *testing.T) {
	t.Run("WithNilComponent", func(t *testing.T) {
		cd := cost.ComponentDiff{}
		actual := cd.PriorCost()
		assert.True(t, actual.Equal(decimal.Zero))
	})

	t.Run("WithValue", func(t *testing.T) {
		cd := cost.ComponentDiff{Prior: &cost.Component{Quantity: decimal.NewFromInt(5), Rate: cost.NewMonthly(decimal.NewFromFloat(1.5), "USD")}}
		actual := cd.PriorCost()
		assert.True(t, actual.Equal(decimal.NewFromFloat(7.5)))
	})
}

func TestComponentDiff_PlannedCost(t *testing.T) {
	t.Run("WithNilComponent", func(t *testing.T) {
		cd := cost.ComponentDiff{}
		actual := cd.PlannedCost()
		assert.True(t, actual.Equal(decimal.Zero))
	})

	t.Run("WithValue", func(t *testing.T) {
		cd := cost.ComponentDiff{Planned: &cost.Component{Quantity: decimal.NewFromInt(5), Rate: cost.NewMonthly(decimal.NewFromFloat(1.5), "USD")}}
		actual := cd.PlannedCost()
		assert.True(t, actual.Equal(decimal.NewFromFloat(7.5)))
	})
}

func TestComponentDiff_Valid(t *testing.T) {
	err := fmt.Errorf("test error")
	testcases := []struct {
		prior, planned *cost.Component
		valid          bool
	}{
		{&cost.Component{}, &cost.Component{}, true},
		{&cost.Component{}, nil, true},
		{nil, &cost.Component{}, true},
		{&cost.Component{Error: err}, &cost.Component{}, false},
		{&cost.Component{}, &cost.Component{Error: err}, false},
		{&cost.Component{Error: err}, &cost.Component{Error: err}, false},
		{&cost.Component{Error: err}, nil, false},
		{nil, &cost.Component{Error: err}, false},
		{nil, nil, true},
	}

	for i, tc := range testcases {
		cd := cost.ComponentDiff{Prior: tc.prior, Planned: tc.planned}
		assert.Equal(t, tc.valid, cd.Valid(), "case %d", i)
	}
}
