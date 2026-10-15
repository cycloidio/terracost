package google_test

import (
	"context"
	"testing"

	"github.com/cycloidio/terracost/google"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestIngester(t *testing.T) {
	var (
		ctx     = context.Background()
		project = "proj"
		zone    = "europe-west1-b"
		cred    []byte
	)

	ts := testutil.StartGoogleServer(t)
	defer ts.Close()

	t.Run("SuccessIngestAll", func(t *testing.T) {
		i, err := google.NewIngester(ctx, cred, google.ComputeEngine.String(), project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()))
		require.NoError(t, err)

		var count int
		for _ = range i.Ingest(ctx, 10) {
			count++
		}

		require.NoError(t, i.Err())
		assert.Equal(t, 1777, count)
	})
	t.Run("SuccessMinimal", func(t *testing.T) {
		i, err := google.NewIngester(ctx, nil, google.ComputeEngine.String(), project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()), google.WithIngestionFilter(google.MinimalFilter))
		require.NoError(t, err)

		pwps := make([]*price.WithProduct, 0, 215)
		for pwp := range i.Ingest(ctx, 10) {
			pwps = append(pwps, pwp)
		}
		assert.Equal(t, 215, len(pwps))
		require.NoError(t, i.Err())
	})
	t.Run("SuccessIngestSubset", func(t *testing.T) {
		i, err := google.NewIngester(ctx, nil, google.ComputeEngine.String(), project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()), google.WithIngestionFilter(func(pp *price.WithProduct) bool {
			// This filter will import only the components of the 'e2-small'
			if pp.Product.Attributes["machine_family"] == "e2" {
				if pp.Product.Attributes["group"] == "CPU" || pp.Product.Attributes["group"] == "RAM" {
					return true
				} else if pp.Product.Attributes["machine_type"] == "e2-small" {
					return true
				}
			}
			return false
		}))
		require.NoError(t, err)

		pwps := make([]*price.WithProduct, 0, 3)
		for pwp := range i.Ingest(ctx, 10) {
			pwps = append(pwps, pwp)
		}

		require.NoError(t, i.Err())
		assert.Equal(t, 3, len(pwps))
		assert.Equal(t, "0.05441892", pwps[2].Price.Value.String())
	})
	t.Run("MachineTypeMemoryUnits", func(t *testing.T) {
		i, err := google.NewIngester(ctx, nil, google.ComputeEngine.String(), project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()), google.WithIngestionFilter(func(pp *price.WithProduct) bool {
			return pp.Product.Attributes["machine_type"] == "e2-standard-32"
		}))
		require.NoError(t, err)

		var prices []*price.WithProduct
		for pp := range i.Ingest(ctx, 10) {
			prices = append(prices, pp)
		}
		require.NoError(t, i.Err())
		require.Len(t, prices, 1)
		// The existing fixtures describe 32 CPUs and 131072 MiB (128 GiB),
		// with rates of 0.023993370 USD/CPU-hour and 0.003216090 USD/GiB-hour.
		assert.Equal(t, "1.17944736", prices[0].Price.Value.String())
		assert.Equal(t, "h", prices[0].Price.Unit)
		assert.Equal(t, "USD", prices[0].Price.Currency)
	})
	t.Run("ErrNotSupportedService", func(t *testing.T) {
		_, err := google.NewIngester(ctx, cred, "service", project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()))
		assert.EqualError(t, err, google.ErrNotSupportedService.Error())
	})
	t.Run("FailUnsupportedZone", func(t *testing.T) {
		_, err := google.NewIngester(ctx, cred, google.ComputeEngine.String(), project, "zone", google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()))
		assert.EqualError(t, err, "unable to get region from zone: invalid zone: zone")
	})
}
