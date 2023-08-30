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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
						},
					},
				},
			},
		}
		plan := cost.NewPlan("name", prior, nil)
		resourceDiffs := plan.ResourceDifferences()

		require.Len(t, resourceDiffs, 1)
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test1",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Prior: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
						},
					},
				},
			},
		}
		plan := cost.NewPlan("name", nil, planned)
		resourceDiffs := plan.ResourceDifferences()

		require.Len(t, resourceDiffs, 1)
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test1",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Planned: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.50), "USD"),
						},
					},
				},
				"aws_instance.test_delete": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50), "USD"),
						},
					},
				},
				"aws_instance.test_create": {
					Components: map[string]cost.Component{
						"EC2 instance hours": {
							Quantity: decimal.NewFromInt(730),
							Unit:     "Hrs",
							Rate:     cost.NewMonthly(decimal.NewFromFloat(3.21), "USD"),
						},
					},
				},
			},
		}
		plan := cost.NewPlan("name", prior, planned)
		resourceDiffs := plan.ResourceDifferences()

		require.Len(t, resourceDiffs, 3)
		assert.Contains(t, resourceDiffs, cost.ResourceDiff{
			Address: "aws_instance.test_update",
			ComponentDiffs: map[string]*cost.ComponentDiff{
				"EC2 instance hours": {
					Prior: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.50), "USD"),
					},
					Planned: &cost.Component{
						Quantity: decimal.NewFromInt(730),
						Unit:     "Hrs",
						Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50), "USD"),
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
						Rate:     cost.NewMonthly(decimal.NewFromFloat(3.21), "USD"),
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
						Rate:     cost.NewMonthly(decimal.NewFromFloat(1.23), "USD"),
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50), "USD"),
						},
					},
				},
				"aws_invalid_resource.test_skipped": {
					Skipped: true,
				},
			},
		}
		plan := cost.NewPlan("name", prior, nil)
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50), "USD"),
						},
					},
				},
				"aws_invalid_resource.test_skipped": {
					Skipped: true,
				},
			},
		}
		plan := cost.NewPlan("name", nil, planned)
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50), "USD"),
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
							Rate:     cost.NewMonthly(decimal.NewFromFloat(2.50), "USD"),
						},
					},
				},
				"aws_invalid_resource.skipped_prior_planned": {Skipped: true},
				"aws_invalid_resource.skipped_planned":       {Skipped: true},
			},
		}
		plan := cost.NewPlan("name", prior, planned)
		skipped := plan.SkippedAddresses()
		require.Len(t, skipped, 3)
		assert.Contains(t, skipped, "aws_invalid_resource.skipped_prior_planned")
		assert.Contains(t, skipped, "aws_invalid_resource.skipped_prior")
		assert.Contains(t, skipped, "aws_invalid_resource.skipped_planned")
	})
}
