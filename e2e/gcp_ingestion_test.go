package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	costestimation "github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/google"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestGoogleIngestion(t *testing.T) {
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

	allProds, err := backend.Products().Filter(ctx, &product.Filter{Provider: util.StringPtr(google.ProviderName), Service: util.StringPtr(google.ComputeEngine.String())})
	require.NoError(t, err)
	fmt.Println(len(allProds))
	assert.Len(t, allProds, 215)

	for _, prod := range allProds {
		prices, err := backend.Prices().Filter(ctx, prod.ID, nil)
		require.NoError(t, err)
		assert.Len(t, prices, 1)
	}
}
