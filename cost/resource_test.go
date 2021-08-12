package cost_test

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/cost"
)

func TestResourceDiff_Errors(t *testing.T) {
	rd := &cost.ResourceDiff{
		Address: "complex_resource.example",
		ComponentDiffs: map[string]*cost.ComponentDiff{
			"BothNil": {
				Prior:   nil,
				Planned: nil,
			},
			"BothCorrect": {
				Prior:   &cost.Component{},
				Planned: &cost.Component{},
			},
			"BrokenPrior": {
				Prior:   &cost.Component{Error: cost.ErrProductNotFound},
				Planned: nil,
			},
			"BrokenPlanned": {
				Prior:   nil,
				Planned: &cost.Component{Error: cost.ErrPriceNotFound},
			},
			"BothBroken": {
				Prior:   &cost.Component{Error: cost.ErrProductNotFound},
				Planned: &cost.Component{Error: cost.ErrPriceNotFound},
			},
		},
	}

	expected := map[string]error{
		"BrokenPrior":   cost.ErrProductNotFound,
		"BrokenPlanned": cost.ErrPriceNotFound,
		"BothBroken":    cost.ErrProductNotFound,
	}
	assert.Equal(t, expected, rd.Errors())
}

func TestResourceDiff_Valid(t *testing.T) {
	testcases := []struct {
		label string
		cd    *cost.ComponentDiff
		valid bool
	}{
		{
			"BothNil",
			&cost.ComponentDiff{
				Prior:   nil,
				Planned: nil,
			},
			true,
		},
		{
			"BothCorrect",
			&cost.ComponentDiff{
				Prior:   &cost.Component{},
				Planned: &cost.Component{},
			},
			true,
		},
		{
			"BrokenPrior",
			&cost.ComponentDiff{
				Prior:   &cost.Component{Error: cost.ErrProductNotFound},
				Planned: nil,
			},
			false,
		},
		{
			"BrokenPlanned",
			&cost.ComponentDiff{
				Prior:   nil,
				Planned: &cost.Component{Error: cost.ErrPriceNotFound},
			},
			false,
		},
		{
			"BothBroken",
			&cost.ComponentDiff{
				Prior:   &cost.Component{Error: cost.ErrProductNotFound},
				Planned: &cost.Component{Error: cost.ErrPriceNotFound},
			},
			false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			rd := &cost.ResourceDiff{
				Address: tc.label,
				ComponentDiffs: map[string]*cost.ComponentDiff{
					tc.label: tc.cd,
				},
			}
			assert.Equal(t, tc.valid, rd.Valid())
		})
	}
}

func TestResourceDiff_Cost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rd := &cost.ResourceDiff{
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"comp1": {
					Prior: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(104.6), "USD"),
						Quantity: decimal.NewFromInt(1),
					},
					Planned: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(208.8), "USD"),
						Quantity: decimal.NewFromInt(1),
					},
				},
				"comp2": {
					Prior: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(10.5), "USD"),
						Quantity: decimal.NewFromInt(2),
					},
					Planned: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(20.7), "USD"),
						Quantity: decimal.NewFromInt(2),
					},
				},
			},
		}

		prior, err := rd.PriorCost()
		require.NoError(t, err)
		assertDecimalEqual(t, decimal.NewFromFloat(125.6), prior.Monthly())
		assert.Equal(t, "USD", prior.Currency)

		planned, err := rd.PlannedCost()
		require.NoError(t, err)
		assertDecimalEqual(t, decimal.NewFromFloat(250.2), planned.Decimal)
		assert.Equal(t, "USD", planned.Currency)
	})
	t.Run("PriorWithError", func(t *testing.T) {
		rd := &cost.ResourceDiff{
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"comp1": {
					Prior: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(104.6), "USD"),
						Quantity: decimal.NewFromInt(1),
					},
				},
				"comp2": {
					Prior: &cost.Component{
						Error: fmt.Errorf("prior error"),
					},
				},
			},
		}

		prior, err := rd.PriorCost()
		require.NoError(t, err)
		assertDecimalEqual(t, decimal.NewFromFloat(104.6), prior.Monthly())
		assert.Equal(t, "USD", prior.Currency)
	})
	t.Run("PlannedWithError", func(t *testing.T) {
		rd := &cost.ResourceDiff{
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"comp1": {
					Planned: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(208.8), "USD"),
						Quantity: decimal.NewFromInt(1),
					},
				},
				"comp2": {
					Planned: &cost.Component{
						Error: fmt.Errorf("planned error"),
					},
				},
			},
		}

		planned, err := rd.PlannedCost()
		require.NoError(t, err)
		assertDecimalEqual(t, decimal.NewFromFloat(208.8), planned.Decimal)
		assert.Equal(t, "USD", planned.Currency)
	})
	t.Run("PriorCurrencyMismatch", func(t *testing.T) {
		rd := &cost.ResourceDiff{
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"comp1": {
					Prior: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(104.6), "USD"),
						Quantity: decimal.NewFromInt(1),
					},
				},
				"comp2": {
					Prior: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(10.5), "EUR"),
						Quantity: decimal.NewFromInt(2),
					},
				},
			},
		}

		_, err := rd.PriorCost()
		require.Error(t, err)
		// Because of the iterration of the map the order of currency can be reversed.
		// We want simply to ensure that an error is returned, the order doesn't matter
		errs := []string{
			"failed calculating prior cost: currency mismatch: expected USD, got EUR",
			"failed calculating prior cost: currency mismatch: expected EUR, got USD",
		}
		assert.Contains(t, errs, err.Error())
	})
	t.Run("PlannedCurrencyMismatch", func(t *testing.T) {
		rd := &cost.ResourceDiff{
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"comp1": {
					Planned: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(208.8), "USD"),
						Quantity: decimal.NewFromInt(1),
					},
				},
				"comp2": {
					Planned: &cost.Component{
						Rate:     cost.NewMonthly(decimal.NewFromFloat(20.7), "EUR"),
						Quantity: decimal.NewFromInt(2),
					},
				},
			},
		}

		_, err := rd.PlannedCost()
		require.Error(t, err)
		// Because of the iterration of the map the order of currency can be reversed.
		// We want simply to ensure that an error is returned, the order doesn't matter
		errs := []string{
			"failed calculating planned cost: currency mismatch: expected USD, got EUR",
			"failed calculating planned cost: currency mismatch: expected EUR, got USD",
		}
		assert.Contains(t, errs, err.Error())
	})
}
