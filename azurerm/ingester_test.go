package azurerm_test

import (
	"context"
	"testing"

	"github.com/cycloidio/terracost/azurerm"
	"github.com/cycloidio/terracost/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIngest(t *testing.T) {
	var (
		ctx    = context.Background()
		region = "francecentral"
	)

	ts := testutil.StartAzureServer(t)
	defer ts.Close()
	t.Run("SuccessIngestAll", func(t *testing.T) {

		i, err := azurerm.NewIngester(ctx, azurerm.VirtualMachines.String(), region, azurerm.WithEndpoint(ts.URL))
		require.NoError(t, err)

		var count int
		for _ = range i.Ingest(ctx, 10) {
			count++
		}

		require.NoError(t, i.Err())
		assert.Equal(t, 4436, count)
	})
	t.Run("SuccessMinimal", func(t *testing.T) {

		i, err := azurerm.NewIngester(ctx, azurerm.VirtualMachines.String(), region, azurerm.WithIngestionFilter(azurerm.MinimalFilter), azurerm.WithEndpoint(ts.URL))
		require.NoError(t, err)

		var count int
		for _ = range i.Ingest(ctx, 10) {
			count++
		}

		require.NoError(t, i.Err())
		assert.Equal(t, 840, count)
	})
	t.Run("ErrNotSupportedService", func(t *testing.T) {
		_, err := azurerm.NewIngester(ctx, "invalid service", region, azurerm.WithIngestionFilter(azurerm.MinimalFilter), azurerm.WithEndpoint(ts.URL))
		assert.EqualError(t, err, azurerm.ErrNotSupportedService.Error())
	})
}
