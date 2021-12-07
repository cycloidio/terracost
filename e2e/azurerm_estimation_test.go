package e2e

import (
	"context"
	"database/sql"
	"os"
	"testing"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/azurerm"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/testutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureRMEstimation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	var (
		ctx    = context.Background()
		region = "francecentral"
	)

	db, err := sql.Open("mysql", "root:terracost@tcp(172.44.0.2:3306)/terracost_test?multiStatements=true")
	require.NoError(t, err)

	ts := testutil.StartAzureServer(t)
	defer ts.Close()

	backend := mysql.NewBackend(db)
	ingester, err := azurerm.NewIngester(ctx, azurerm.VirtualMachines.String(), region, azurerm.WithIngestionFilter(azurerm.MinimalFilter), azurerm.WithEndpoint(ts.URL))
	require.NoError(t, err)

	err = costestimation.IngestPricing(ctx, backend, ingester)
	require.NoError(t, err)

	t.Run("FromPlan", func(t *testing.T) {
		f, err := os.Open("../testdata/azurerm/plan.json")
		require.NoError(t, err)
		defer f.Close()

		plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f)
		require.NoError(t, err)

		pcost, err := plan.PriorCost()
		assert.NoError(t, err)
		assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(0), ""), pcost)

		pcost, err = plan.PlannedCost()
		assert.NoError(t, err)
		assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(64.021), "USD"), pcost)
	})
	t.Run("FromHCL", func(t *testing.T) {
		plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/azurerm/stack-compute")
		require.NoError(t, err)

		assert.Nil(t, plan.Prior)

		pcost, err := plan.PlannedCost()
		assert.NoError(t, err)
		assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(64.021), "USD"), pcost)
	})
}
