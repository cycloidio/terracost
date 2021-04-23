package e2e

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/terraform"
)

var terraformProviderInitializer = terraform.ProviderInitializer{
	MatchNames: []string{"aws", "aws-test"},
	Provider: func(config map[string]string) (terraform.Provider, error) {
		regCode := region.Code(config["region"])
		return awstf.NewProvider("aws-test", regCode)
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
				"capacitystatus":  "Used",
				"instanceType":    "t2.micro",
				"tenancy":         "Shared",
				"operatingSystem": "Linux",
				"preInstalledSw":  "NA",
			},
		},
		{
			Provider: "aws-test",
			SKU:      "TESTPROD-T2-XLARGE",
			Service:  "AmazonEC2",
			Family:   "Compute Instance",
			Location: "us-east-1",
			Attributes: map[string]string{
				"capacitystatus":  "Used",
				"instanceType":    "t2.xlarge",
				"tenancy":         "Shared",
				"operatingSystem": "Linux",
				"preInstalledSw":  "NA",
			},
		},
		{
			Provider: "aws-test",
			SKU:      "TESTPROD-STORAGE",
			Service:  "AmazonEC2",
			Family:   "Storage",
			Location: "us-east-1",
			Attributes: map[string]string{
				"volumeApiName": "gp2",
			},
		},
	}

	for _, p := range prods {
		var err error
		p.ID, err = backend.Product().Upsert(ctx, p)
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
					"purchaseOption": "on_demand",
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
					"purchaseOption": "on_demand",
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
					"purchaseOption": "on_demand",
				},
			},
		},
	}

	for _, p := range prices {
		_, err := backend.Price().Upsert(ctx, p)
		require.NoError(t, err)
	}

	t.Run("Success", func(t *testing.T) {
		f, err := os.Open("../testdata/terraform-plan.json")
		require.NoError(t, err)
		defer f.Close()

		plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, terraformProviderInitializer)
		require.NoError(t, err)

		assertDecimalEqual(t, decimal.NewFromFloat(91.2), plan.PriorCost())
		assertDecimalEqual(t, decimal.NewFromFloat(901.5), plan.PlannedCost())

		diffs := plan.ResourceDifferences()
		require.Len(t, diffs, 2)

		for _, diff := range diffs {
			switch diff.Address {
			case "aws_instance.example":
				compute := diff.ComponentDiffs["Compute"]
				require.NotNil(t, compute)
				assert.Equal(t, []string{"Linux", "on-demand", "t2.micro"}, compute.Prior.Details)
				assert.Equal(t, []string{"Linux", "on-demand", "t2.xlarge"}, compute.Planned.Details)
				assertDecimalEqual(t, decimal.NewFromFloat(87.6), compute.PriorCost())
				assertDecimalEqual(t, decimal.NewFromFloat(897.9), compute.PlannedCost())

				rootVol := diff.ComponentDiffs["Root volume: Storage"]
				require.NotNil(t, rootVol)
				assertDecimalEqual(t, decimal.NewFromFloat(3.6), rootVol.PriorCost())
				assertDecimalEqual(t, decimal.NewFromFloat(3.6), rootVol.PlannedCost())

			case "aws_lb.example":
				lb := diff.ComponentDiffs["Application Load Balancer"]
				require.NotNil(t, lb)
				assert.False(t, diff.Valid())
				assertDecimalEqual(t, decimal.NewFromFloat(0), lb.Planned.Cost())
			}
		}
	})

	t.Run("ProductNotFound", func(t *testing.T) {
		f, err := os.Open("../testdata/terraform-plan-invalid.json")
		require.NoError(t, err)
		defer f.Close()

		plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, terraformProviderInitializer)
		require.NoError(t, err)

		diffs := plan.ResourceDifferences()
		require.Len(t, diffs, 1)
		rd := diffs[0]

		rootVol := diffs[0].ComponentDiffs["Root volume: Storage"]
		require.NotNil(t, rootVol)
		assertDecimalEqual(t, decimal.NewFromFloat(3.6), rootVol.PriorCost())
		assertDecimalEqual(t, decimal.NewFromFloat(3.6), rootVol.PlannedCost())

		expected := map[string]error{
			"Compute": cost.ErrProductNotFound,
		}
		assert.Equal(t, expected, rd.Errors())
	})
}

func assertDecimalEqual(t *testing.T, expected, actual decimal.Decimal) {
	assert.Truef(t, expected.Equal(actual), "Not equal:\nexpected: %s\nactual  : %s", expected, actual)
}
