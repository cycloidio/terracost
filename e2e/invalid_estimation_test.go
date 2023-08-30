package e2e

import (
	"context"
	"database/sql"
	"os"
	"testing"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This provide is currently UNSUPPORTED, it aims to test an UNSUPPORTED
// provider errors for both terraform and HCL files

func TestVMWareEstimation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx := context.Background()
	db, err := sql.Open("mysql", "root:terracost@tcp(172.44.0.2:3306)/terracost_test?multiStatements=true")
	require.NoError(t, err)

	backend := mysql.NewBackend(db)

	t.Run("TFPlan", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			f, err := os.Open("../testdata/invalid/terraform-empty-plan.json")
			require.NoError(t, err)
			defer f.Close()

			plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, usage.Default, terraformAWSTestProviderInitializer)
			require.Error(t, err, terraform.ErrNoQueries)
			assert.Nil(t, plan)
		})
		t.Run("UnsupportedProvider", func(t *testing.T) {
			f, err := os.Open("../testdata/invalid/terraform-unsupported-plan.json")
			require.NoError(t, err)
			defer f.Close()

			plan, err := costestimation.EstimateTerraformPlan(ctx, backend, f, usage.Default, terraformAWSTestProviderInitializer)
			require.Error(t, err, terraform.ErrNoKnownProvider)
			assert.Nil(t, plan)
		})
	})

	t.Run("HCL", func(t *testing.T) {
		t.Run("UnsupportedProvider", func(t *testing.T) {
			plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/invalid/stack-vmware", noModulePath, noForceTerragrunt, usage.Default)
			assert.Nil(t, plan)
			assert.Error(t, err, terraform.ErrNoKnownProvider)
		})
		t.Run("EmptyTerraform", func(t *testing.T) {
			plan, err := costestimation.EstimateHCL(ctx, backend, nil, "../testdata/invalid/stack-empty", noModulePath, noForceTerragrunt, usage.Default)
			assert.Nil(t, plan)
			assert.Error(t, err, terraform.ErrNoQueries)
		})
	})
}
