package e2e

import (
	"context"
	"database/sql"
	"os"
	"testing"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/google"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/usage"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestGoogleEstimation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	var (
		ctx     = context.Background()
		project = "proj"
		zone    = "europe-west1-b"
		cred    = []byte(`{"type": "service_account"}`)
	)

	db, err := sql.Open("mysql", "root:terracost@tcp(172.44.0.2:3306)/terracost_test?multiStatements=true")
	require.NoError(t, err)

	ts := testutil.StartGoogleServer(t)
	defer ts.Close()

	backend := mysql.NewBackend(db)
	ingester, err := google.NewIngester(ctx, cred, google.ComputeEngine.String(), project, zone, google.WithIngestionFilter(google.MinimalFilter), google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()))
	require.NoError(t, err)

	err = costestimation.IngestPricing(ctx, backend, ingester)
	require.NoError(t, err)

	t.Run("FromPlan", func(t *testing.T) {
		f, err := os.Open("../testdata/google/plan.json")
		require.NoError(t, err)
		defer f.Close()

		plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, usage.Default)
		require.NoError(t, err)

		pcost, err := plan.PriorCost()
		assert.NoError(t, err)
		assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(0), ""), pcost)

		pcost, err = plan.PlannedCost()
		assert.NoError(t, err)
		assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(39.7258116), "USD"), pcost)
	})
	t.Run("FromHCL", func(t *testing.T) {
		plans, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/google/stack-compute", noModulePath, noForceTerragrunt, usage.Default)
		require.NoError(t, err)
		require.Len(t, plans, 1)
		plan := plans[0]

		assert.Nil(t, plan.Prior)

		pcost, err := plan.PlannedCost()
		assert.NoError(t, err)
		assertCostEqual(t, cost.NewMonthly(decimal.NewFromFloat(39.7258116), "USD"), pcost)
	})
}
