package cost_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/cost"
)

func TestPlan_ResourceDifferences(t *testing.T) {
	t.Run("OnlyPrior", func(t *testing.T) {
		prior := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test1": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23)),
						},
					},
				},
			},
		}
		plan := cost.NewPlan(prior, nil)
		resourceDiffs := plan.ResourceDifferences()

		require.Len(t, resourceDiffs, 1)
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test1",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Prior: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23)),
					},
				},
			},
		})
	})

	t.Run("OnlyPlanned", func(t *testing.T) {
		planned := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test1": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23)),
						},
					},
				},
			},
		}
		plan := cost.NewPlan(nil, planned)
		resourceDiffs := plan.ResourceDifferences()

		require.Len(t, resourceDiffs, 1)
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test1",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Planned: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23)),
					},
				},
			},
		})
	})

	t.Run("PriorAndPlanned", func(t *testing.T) {
		prior := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test_update": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.50)),
						},
					},
				},
				"aws_instance.test_delete": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23)),
						},
					},
				},
			},
		}
		planned := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test_update": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50)),
						},
					},
				},
				"aws_instance.test_create": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(3.21)),
						},
					},
				},
			},
		}
		plan := cost.NewPlan(prior, planned)
		resourceDiffs := plan.ResourceDifferences()

		require.Len(t, resourceDiffs, 3)
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test_update",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Prior: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.50)),
					},
					Planned: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50)),
					},
				},
			},
		})
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test_create",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Planned: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(3.21)),
					},
				},
			},
		})
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test_delete",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Prior: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23)),
					},
				},
			},
		})
	})
}

func TestPlan_SkippedAddresses(t *testing.T) {
	t.Run("OnlyPrior", func(t *testing.T) {
		prior := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test_valid": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50)),
						},
					},
				},
				"aws_invalid_resource.test_skipped": {
					Skipped: true,
				},
			},
		}
		plan := cost.NewPlan(prior, nil)
		skipped := plan.SkippedAddresses()
		require.Len(t, skipped, 1)
		assert.Contains(t, skipped, "aws_invalid_resource.test_skipped")
	})

	t.Run("OnlyPlanned", func(t *testing.T) {
		planned := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test_valid": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50)),
						},
					},
				},
				"aws_invalid_resource.test_skipped": {
					Skipped: true,
				},
			},
		}
		plan := cost.NewPlan(nil, planned)
		skipped := plan.SkippedAddresses()
		require.Len(t, skipped, 1)
		assert.Contains(t, skipped, "aws_invalid_resource.test_skipped")
	})

	t.Run("PriorAndPlanned", func(t *testing.T) {
		prior := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test_prior": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50)),
						},
					},
				},
				"aws_invalid_resource.skipped_prior_planned": {Skipped: true},
				"aws_invalid_resource.skipped_prior":         {Skipped: true},
			},
		}
		planned := &cost.State{
			Resources: map[string]cost.Resource{
				"aws_instance.test_prior": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50)),
						},
					},
				},
				"aws_invalid_resource.skipped_prior_planned": {Skipped: true},
				"aws_invalid_resource.skipped_planned":       {Skipped: true},
			},
		}
		plan := cost.NewPlan(prior, planned)
		skipped := plan.SkippedAddresses()
		require.Len(t, skipped, 3)
		assert.Contains(t, skipped, "aws_invalid_resource.skipped_prior_planned")
		assert.Contains(t, skipped, "aws_invalid_resource.skipped_prior")
		assert.Contains(t, skipped, "aws_invalid_resource.skipped_planned")
	})
}
