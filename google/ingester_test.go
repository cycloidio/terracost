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
		cred    = []byte(`{"type": "service_account"}`)
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
		i, err := google.NewIngester(ctx, []byte(`{"type": "service_account"}`), google.ComputeEngine.String(), project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()), google.WithIngestionFilter(google.MinimalFilter))
		require.NoError(t, err)

		pwps := make([]*price.WithProduct, 0, 215)
		for pwp := range i.Ingest(ctx, 10) {
			pwps = append(pwps, pwp)
		}
		assert.Equal(t, 215, len(pwps))
		require.NoError(t, i.Err())
	})
	t.Run("SuccessIngestSubset", func(t *testing.T) {
		i, err := google.NewIngester(ctx, []byte(`{"type": "service_account"}`), google.ComputeEngine.String(), project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()), google.WithIngestionFilter(func(pp *price.WithProduct) bool {
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
	t.Run("ErrNotSupportedService", func(t *testing.T) {
		_, err := google.NewIngester(ctx, cred, "service", project, zone, google.WithGCPOption(option.WithEndpoint(ts.URL), option.WithoutAuthentication()))
		assert.EqualError(t, err, google.ErrNotSupportedService.Error())
	})
}
