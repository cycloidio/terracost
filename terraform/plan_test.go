package terraform_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/cost-estimation/mock"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/terraform"
)

func TestPlan_ExtractPlannedQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := mock.NewTerraformProvider(ctrl)

	plan := terraform.NewPlan(terraform.WithProvider("aws", func(_ terraform.ProviderConfig) terraform.Provider {
		return provider
	}))

	f, err := os.Open("testdata/plan.json")
	require.NoError(t, err)

	err = plan.Read(f)
	require.NoError(t, err)

	provider.EXPECT().ResourceComponents(gomock.Any()).DoAndReturn(func(res terraform.Resource) ([]query.Component, error) {
		assert.Equal(t, "aws_instance.example", res.Address)
		assert.Equal(t, "t2.xlarge", res.Values["instance_type"])
		return []query.Component{}, nil
	})

	queries := plan.ExtractPlannedQueries()
	require.Len(t, queries, 1)
}

func TestPlan_ExtractPriorQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := mock.NewTerraformProvider(ctrl)

	plan := terraform.NewPlan(terraform.WithProvider("aws", func(_ terraform.ProviderConfig) terraform.Provider {
		return provider
	}))

	f, err := os.Open("testdata/plan.json")
	require.NoError(t, err)

	err = plan.Read(f)
	require.NoError(t, err)

	provider.EXPECT().ResourceComponents(gomock.Any()).DoAndReturn(func(res terraform.Resource) ([]query.Component, error) {
		assert.Equal(t, "aws_instance.example", res.Address)
		assert.Equal(t, "t2.micro", res.Values["instance_type"])
		return []query.Component{}, nil
	})

	queries := plan.ExtractPriorQueries()
	require.Len(t, queries, 1)
}
