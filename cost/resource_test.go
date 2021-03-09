package cost_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
