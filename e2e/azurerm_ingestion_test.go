package e2e

import (
	"context"
	"database/sql"
	"testing"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/azurerm"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAzureRMIngestion(t *testing.T) {
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

	allProds, err := backend.Products().Filter(ctx, &product.Filter{Provider: util.StringPtr(azurerm.ProviderName), Service: util.StringPtr(azurerm.VirtualMachines.String())})
	require.NoError(t, err)
	assert.Len(t, allProds, 838)

	for _, prod := range allProds {
		prices, err := backend.Prices().Filter(ctx, prod.ID, nil)
		require.NoError(t, err)
		assert.Len(t, prices, 1)
	}
}
