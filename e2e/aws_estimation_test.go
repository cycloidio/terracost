package e2e

import (
	"context"
	"database/sql"
	"os"
	"testing"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/terraform"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// terraformAWSTestProviderInitializer is a testing ProviderInitializer
// pricing are directly inserted inside the database, which allows us to
// test the processing with smaller subset of data, as well as the functioning
// of MatchNames for a given provider - as data are injected using 'aws-test'
// which is also used in the tfplan & co.
var terraformAWSTestProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{"aws", "aws-test"},
	Provider: func(config map[string]interface{}) (terraform.Provider, error) {
		r, ok := config["region"]
		if !ok {
			return nil, nil
		}
		regCode := region.Code(r.(string))
		return awstf.NewProvider("aws-test", regCode)
	},
}

// terraformAWSProviderInitializer is a proper AWS provider.
// We do not want to reuse the testing terraformAWSTestProviderInitializer
// because the HCL contains actual valid AWS resources and provider
// meaning 'aws' is used an not 'aws-test'. On top of that pricing data
// from a real dump are injected to ensure better testing scenarios
var terraformAWSProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{"aws", "registry.terraform.io/hashicorp/aws"},
	Provider: func(config map[string]interface{}) (terraform.Provider, error) {
		r, ok := config["region"]
		if !ok {
			return nil, nil
		}
		regCode := region.Code(r.(string))
		return awstf.NewProvider("aws", regCode)
	},
}

func TestAWSEstimation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx := context.Background()
	db, err := sql.Open("mysql", "root:terracost@tcp(172.44.0.2:3306)/terracost_test?multiStatements=true")
	require.NoError(t, err)

	backend := mysql.NewBackend(db)

	prods := []*product.Product{
		{
			Provider: "aws-test",
			SKU:      "TESTPROD-T2-MICRO",
			Service:  "AmazonEC2",
			Family:   "Compute Instance",
			Location: "us-east-1",
			Attributes: map[string]string{
				"CapacityStatus":  "Used",
				"InstanceType":    "t2.micro",
				"Tenancy":         "Shared",
				"OperatingSystem": "Linux",
				"PreInstalledSW":  "NA",
			},
		},
		{
			Provider: "aws-test",
			SKU:      "TESTPROD-T2-XLARGE",
			Service:  "AmazonEC2",
			Family:   "Compute Instance",
			Location: "us-east-1",
			Attributes: map[string]string{
				"CapacityStatus":  "Used",
				"InstanceType":    "t2.xlarge",
				"Tenancy":         "Shared",
				"OperatingSystem": "Linux",
				"PreInstalledSW":  "NA",
			},
		},
		{
			Provider: "aws-test",
			SKU:      "TESTPROD-STORAGE",
			Service:  "AmazonEC2",
			Family:   "Storage",
			Location: "us-east-1",
			Attributes: map[string]string{
				"VolumeAPIName": "gp2",
			},
		},
	}

	for _, p := range prods {
		var err error
		p.ID, err = backend.Products().Upsert(ctx, p)
		require.NoError(t, err)
	}

	prices := []*price.WithProduct{
		{
			Product: prods[0],
			Price: price.Price{
				Unit:     "Hrs",
				Currency: "USD",
				Value:    decimal.NewFromFloat(0.12),
				Attributes: map[string]string{
					"TermType": "OnDemand",
				},
			},
		},
		{
			Product: prods[1],
			Price: price.Price{
				Unit:     "Hrs",
				Currency: "USD",
				Value:    decimal.NewFromFloat(1.23),
				Attributes: map[string]string{
					"TermType": "OnDemand",
				},
			},
		},
		{
			Product: prods[2],
			Price: price.Price{
				Unit:     "GB-Mo",
				Currency: "USD",
				Value:    decimal.NewFromFloat(0.45),
				Attributes: map[string]string{
					"TermType": "OnDemand",
				},
			},
		},
	}

	for _, p := range prices {
		_, err := backend.Prices().Upsert(ctx, p)
		require.NoError(t, err)
	}

	t.Run("TFPlan", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			f, err := os.Open("../testdata/aws/terraform-plan.json")
			require.NoError(t, err)
			defer f.Close()

			plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, terraformAWSTestProviderInitializer)
			require.NoError(t, err)

			pcost, err := plan.PriorCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(91.2), "USD"), pcost)

			pcost, err = plan.PlannedCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(901.5), "USD"), pcost)

			diffs := plan.ResourceDifferences()
			require.Len(t, diffs, 2)

			for _, diff := range diffs {
				switch diff.Address {
				case "aws_instance.example":
					compute := diff.ComponentDiffs["Compute"]
					require.NotNil(t, compute)
					assert.Equal(t, []string{"Linux", "on-demand", "t2.micro"}, compute.Prior.Details)
					assert.Equal(t, []string{"Linux", "on-demand", "t2.xlarge"}, compute.Planned.Details)
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(87.6), "USD"), compute.PriorCost())
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(897.9), "USD"), compute.PlannedCost())

					rootVol := diff.ComponentDiffs["Root volume: Storage"]
					require.NotNil(t, rootVol)
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(3.6), "USD"), compute.PriorCost())
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(3.6), "USD"), compute.PlannedCost())

					priorCost, err := diff.PriorCost()
					require.NoError(t, err)
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(91.2), "USD"), priorCost)

					plannedCost, err := diff.PlannedCost()
					require.NoError(t, err)
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(901.5), "USD"), plannedCost)

				case "aws_lb.example":
					lb := diff.ComponentDiffs["Application Load Balancer"]
					require.NotNil(t, lb)
					assert.False(t, diff.Valid())
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(0), ""), lb.Planned.Cost())

					priorCost, err := diff.PriorCost()
					require.NoError(t, err)
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(0), ""), priorCost)

					plannedCost, err := diff.PlannedCost()
					require.NoError(t, err)
					assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(0), ""), plannedCost)
				}
			}
		})
		t.Run("SuccessNoPrior", func(t *testing.T) {
			f, err := os.Open("../testdata/aws/terraform-noprior-plan.json")
			require.NoError(t, err)
			defer f.Close()

			plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, terraformAWSProviderInitializer)
			require.NoError(t, err)

			pcost, err := plan.PriorCost()
			assert.NoError(t, err)
			assert.Equal(t, cost.Zero, pcost)

			pcost, err = plan.PlannedCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(31.984), "USD"), pcost)

			diffs := plan.ResourceDifferences()
			require.Len(t, diffs, 1)
			require.Len(t, diffs[0].ComponentDiffs, 2)
		})

		t.Run("ProductNotFound", func(t *testing.T) {
			f, err := os.Open("../testdata/aws/terraform-plan-invalid.json")
			require.NoError(t, err)
			defer f.Close()

			plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, terraformAWSTestProviderInitializer)
			require.NoError(t, err)

			diffs := plan.ResourceDifferences()
			require.Len(t, diffs, 1)
			rd := diffs[0]

			rootVol := diffs[0].ComponentDiffs["Root volume: Storage"]
			require.NotNil(t, rootVol)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(3.6), "USD"), rootVol.PriorCost())
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(3.6), "USD"), rootVol.PlannedCost())

			expected := map[string]error{
				"Compute": cost.ErrProductNotFound,
			}
			assert.Equal(t, expected, rd.Errors())
		})
		t.Run("NoProvider", func(t *testing.T) {
			f, err := os.Open("../testdata/aws/terraform-plan-noprovider.json")
			require.NoError(t, err)
			defer f.Close()

			plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, terraformAWSTestProviderInitializer)
			require.Error(t, err, terraform.ErrNoProviders)
			require.Nil(t, plan)
		})
	})
	t.Run("HCL", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {

			plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/aws/stack-aws", terraformAWSProviderInitializer)
			require.NoError(t, err)

			assert.Nil(t, plan.Prior)

			pcost, err := plan.PlannedCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(62.892), "USD"), pcost)
		})
		t.Run("SuccessMagento", func(t *testing.T) {

			plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/aws/stack-magento", terraformAWSProviderInitializer)
			require.NoError(t, err)

			assert.Nil(t, plan.Prior)

			pcost, err := plan.PlannedCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(73.334), "USD"), pcost)
		})
		t.Run("SuccessASG", func(t *testing.T) {

			plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/aws/stack-asg", terraformAWSProviderInitializer)
			require.NoError(t, err)

			assert.Nil(t, plan.Prior)

			pcost, err := plan.PlannedCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(764.863), "USD"), pcost)
		})
		t.Run("SuccessRemote", func(t *testing.T) {

			plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/aws/stack-remote", terraformAWSProviderInitializer)
			require.NoError(t, err)

			assert.Nil(t, plan.Prior)

			pcost, err := plan.PlannedCost()
			assert.NoError(t, err)
			assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(764.863), "USD"), pcost)
		})
	})
}

func assertCostEqual(t *testing.T, expected, actual cost.Cost) {
	assert.Truef(t, expected.Equal(actual.Decimal), "Not equal:\nexpected value: %s\nactual value: %s", expected, actual)
	assert.Truef(t, expected.Currency == actual.Currency, "Not equal:\nexpected currency: %s\nactual currency: %s", expected.Currency, actual.Currency)
}
