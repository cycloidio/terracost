package cost_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/cycloidio/cost-estimation/cost"
)

func TestComponentDiff_PriorCost(t *testing.T) {
	t.Run("WithNilComponent", func(t *testing.T) {
		cd := cost.ComponentDiff{}
		actual := cd.PriorCost()
		assert.True(t, actual.Equal(decimal.Zero))
	})

	t.Run("WithValue", func(t *testing.T) {
		cd := cost.ComponentDiff{Prior: &cost.Component{Quantity: decimal.NewFromInt(5), Rate: decimal.NewFromFloat(1.5)}}
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
		cd := cost.ComponentDiff{Planned: &cost.Component{Quantity: decimal.NewFromInt(5), Rate: decimal.NewFromFloat(1.5)}}
		actual := cd.PlannedCost()
		assert.True(t, actual.Equal(decimal.NewFromFloat(7.5)))
	})
}
