package terraform_test

import (
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/mock"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
)

func TestPlan_ExtractPlannedQueries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		provider := mock.NewTerraformProvider(ctrl)

		plan := terraform.NewPlan(terraform.ProviderInitializer{
			MatchNames: []string{"aws", "aws-test"},
			Provider: func(_ map[string]interface{}) (terraform.Provider, error) {
				return provider, nil
			},
		})

		f, err := os.Open("../testdata/aws/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		err = plan.Read(f)
		require.NoError(t, err)

		provider.EXPECT().Name().AnyTimes().Return("aws-test")
		provider.EXPECT().ResourceComponents(gomock.Any(), gomock.Any()).DoAndReturn(func(rss map[string]terraform.Resource, res terraform.Resource) ([]query.Component, error) {
			if res.Type == "aws_instance" {
				assert.Equal(t, "module.instance.aws_instance.example", res.Address)
				assert.Equal(t, "t2.xlarge", res.Values["instance_type"])
			} else {
				assert.Equal(t, "aws_lb.example", res.Address)
				assert.Equal(t, "application", res.Values["load_balancer_type"])
			}
			return []query.Component{}, nil
		}).Times(2)

		queries, err := plan.ExtractPlannedQueries()
		require.NoError(t, err)
		require.Len(t, queries, 2)
	})

	t.Run("BadProvider", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		plan := terraform.NewPlan(terraform.ProviderInitializer{
			MatchNames: []string{"aws", "aws-test"},
			Provider: func(_ map[string]interface{}) (terraform.Provider, error) {
				return nil, errors.New("bad provider")
			},
		})

		f, err := os.Open("../testdata/aws/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		err = plan.Read(f)
		require.NoError(t, err)

		_, err = plan.ExtractPlannedQueries()
		assert.Error(t, err)
	})

	t.Run("FailedResourceComponents", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		provider := mock.NewTerraformProvider(ctrl)

		plan := terraform.NewPlan(terraform.ProviderInitializer{
			MatchNames: []string{"aws", "aws-test"},
			Provider: func(_ map[string]interface{}) (terraform.Provider, error) {
				return provider, nil
			},
		})

		f, err := os.Open("../testdata/aws/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		err = plan.Read(f)
		require.NoError(t, err)

		provider.EXPECT().Name().AnyTimes().Return("aws-test")
		provider.EXPECT().ResourceComponents(gomock.Any(), gomock.Any()).DoAndReturn(func(rss map[string]terraform.Resource, res terraform.Resource) ([]query.Component, error) {
			return nil, errors.New("ResourceComponents fail")
		}).Times(2)

		queries, err := plan.ExtractPlannedQueries()
		require.NoError(t, err)
		require.Len(t, queries, 2)
		assert.Contains(t, queries, query.Resource{
			Address:  "module.instance.aws_instance.example",
			Provider: "aws-test",
			Type:     "aws_instance",
		})
	})
}

func TestPlan_ExtractPriorQueries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		provider := mock.NewTerraformProvider(ctrl)

		plan := terraform.NewPlan(terraform.ProviderInitializer{
			MatchNames: []string{"aws", "aws-test"},
			Provider: func(_ map[string]interface{}) (terraform.Provider, error) {
				return provider, nil
			},
		})

		f, err := os.Open("../testdata/aws/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		err = plan.Read(f)
		require.NoError(t, err)

		provider.EXPECT().Name().AnyTimes().Return("aws-test")
		provider.EXPECT().ResourceComponents(gomock.Any(), gomock.Any()).DoAndReturn(func(rss map[string]terraform.Resource, res terraform.Resource) ([]query.Component, error) {
			assert.Equal(t, "module.instance.aws_instance.example", res.Address)
			assert.Equal(t, "t2.micro", res.Values["instance_type"])
			return []query.Component{}, nil
		})

		queries, err := plan.ExtractPriorQueries()
		require.NoError(t, err)
		require.Len(t, queries, 1)
	})

	t.Run("BadProvider", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		plan := terraform.NewPlan(terraform.ProviderInitializer{
			MatchNames: []string{"aws", "aws-test"},
			Provider: func(_ map[string]interface{}) (terraform.Provider, error) {
				return nil, errors.New("bad provider")
			},
		})

		f, err := os.Open("../testdata/aws/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		err = plan.Read(f)
		require.NoError(t, err)

		_, err = plan.ExtractPriorQueries()
		assert.Error(t, err)
	})

	t.Run("FailedResourceComponents", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		provider := mock.NewTerraformProvider(ctrl)

		plan := terraform.NewPlan(terraform.ProviderInitializer{
			MatchNames: []string{"aws", "aws-test"},
			Provider: func(_ map[string]interface{}) (terraform.Provider, error) {
				return provider, nil
			},
		})

		f, err := os.Open("../testdata/aws/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		err = plan.Read(f)
		require.NoError(t, err)

		provider.EXPECT().Name().AnyTimes().Return("aws-test")
		provider.EXPECT().ResourceComponents(gomock.Any(), gomock.Any()).DoAndReturn(func(rss map[string]terraform.Resource, res terraform.Resource) ([]query.Component, error) {
			return nil, errors.New("ResourceComponents fail")
		})

		queries, err := plan.ExtractPriorQueries()
		require.NoError(t, err)
		require.Len(t, queries, 1)
		assert.Contains(t, queries, query.Resource{
			Address:  "module.instance.aws_instance.example",
			Provider: "aws-test",
			Type:     "aws_instance",
		})
	})
}
