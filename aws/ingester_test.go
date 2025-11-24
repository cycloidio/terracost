package aws

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/mock"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"io"
)

func TestIngester_Ingest(t *testing.T) {
	t.Run("InvalidService", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := mock.NewHTTPClient(ctrl)
		ing, err := NewIngester("InvalidService", "eu-west-3", WithHTTPClient(client))
		assert.Error(t, err)
		assert.Nil(t, ing)
	})

	t.Run("EC2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := mock.NewHTTPClient(ctrl)
		ing, err := NewIngester("AmazonEC2", "eu-west-3", WithHTTPClient(client))
		require.NoError(t, err)

		content := makeCSV([][]string{
			{"SKU", "Product Family", "serviceCode", "TermType", "Location", "Unit", "Currency", "PricePerUnit", "Tenancy", "Instance Type", "Operating System", "Volume API Name"},
			{"prod1", "Compute Instance", "AmazonEC2", "OnDemand", "EU (Paris)", "Hrs", "USD", "1.234", "Shared", "m5.xlarge", "Linux", ""},
			{"prod1", "Compute Instance", "AmazonEC2", "Reserved", "EU (Paris)", "Hrs", "USD", "0.987", "Shared", "m5.xlarge", "Linux", ""},
			{"prod2", "Storage", "AmazonEC2", "OnDemand", "EU (Paris)", "GB-Mo", "USD", "0.456", "", "", "", "gp2"},
		})
		rd := strings.NewReader(content)
		res := &http.Response{Body: ioutil.NopCloser(rd)}

		client.EXPECT().Do(gomock.Any()).Return(res, nil)

		results := ing.Ingest(context.Background(), 1)

		prod1 := &product.Product{
			Provider: ProviderName,
			SKU:      "prod1",
			Service:  "AmazonEC2",
			Family:   "Compute Instance",
			Location: "eu-west-3",
			Attributes: map[string]string{
				"InstanceType":    "m5.xlarge",
				"OperatingSystem": "Linux",
				"Tenancy":         "Shared",
			},
		}

		expected := []*price.WithProduct{
			{
				Product: prod1,
				Price: price.Price{
					Unit:       "Hrs",
					Currency:   "USD",
					Value:      decimal.RequireFromString("1.234"),
					Attributes: map[string]string{"TermType": "OnDemand"},
				},
			},
			{
				Product: prod1,
				Price: price.Price{
					Unit:       "Hrs",
					Currency:   "USD",
					Value:      decimal.RequireFromString("0.987"),
					Attributes: map[string]string{"TermType": "Reserved"},
				},
			},
			{
				Product: &product.Product{
					Provider: ProviderName,
					SKU:      "prod2",
					Service:  "AmazonEC2",
					Family:   "Storage",
					Location: "eu-west-3",
					Attributes: map[string]string{
						"VolumeAPIName": "gp2",
					},
				},
				Price: price.Price{
					Unit:       "GB-Mo",
					Currency:   "USD",
					Value:      decimal.RequireFromString("0.456"),
					Attributes: map[string]string{"TermType": "OnDemand"},
				},
			},
		}

		for _, pp := range expected {
			assert.Equal(t, pp, <-results)
		}

		_, ok := <-results
		assert.False(t, ok, "Results channel should be closed")
		require.NoError(t, ing.Err())
	})

	t.Run("LazyQuotes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := mock.NewHTTPClient(ctrl)
		ing, err := NewIngester("AmazonEC2", "us-east-1", WithHTTPClient(client))
		require.NoError(t, err)

		// Create CSV with improperly quoted fields (bare quotes) that would fail without LazyQuotes = true
		// This simulates real-world AWS pricing data that sometimes contains unescaped quotes in description fields
		content := makeCSV([][]string{
			{"SKU", "Product Family", "serviceCode", "TermType", "Location", "Unit", "Currency", "PricePerUnit", "Instance Type", "Operating System"},
			// This row has bare quotes in the Instance Type field - this would normally cause CSV parsing errors
			{`prod3`, `Compute Instance`, `AmazonEC2`, `OnDemand`, `US East (N. Virginia)`, `Hrs`, `USD`, `0.123`, `t3.micro "burstable"`, `Linux`},
			{`prod4`, `Storage`, `AmazonEC2`, `OnDemand`, `US East (N. Virginia)`, `GB-Mo`, `USD`, `0.10`, ``, ``},
		})

		rd := strings.NewReader(content)
		res := &http.Response{Body: io.NopCloser(rd)}

		client.EXPECT().Do(gomock.Any()).Return(res, nil)

		results := ing.Ingest(context.Background(), 1)

		expected := []*price.WithProduct{
			{
				Product: &product.Product{
					Provider: ProviderName,
					SKU:      "prod3",
					Service:  "AmazonEC2",
					Family:   "Compute Instance",
					Location: "us-east-1",
					Attributes: map[string]string{
						"InstanceType":    `t3.micro "burstable"`,
						"OperatingSystem": "Linux",
					},
				},
				Price: price.Price{
					Unit:       "Hrs",
					Currency:   "USD",
					Value:      decimal.RequireFromString("0.123"),
					Attributes: map[string]string{"TermType": "OnDemand"},
				},
			},
			{
				Product: &product.Product{
					Provider:   ProviderName,
					SKU:        "prod4",
					Service:    "AmazonEC2",
					Family:     "Storage",
					Location:   "us-east-1",
					Attributes: map[string]string{},
				},
				Price: price.Price{
					Unit:       "GB-Mo",
					Currency:   "USD",
					Value:      decimal.RequireFromString("0.10"),
					Attributes: map[string]string{"TermType": "OnDemand"},
				},
			},
		}

		for _, pp := range expected {
			assert.Equal(t, pp, <-results)
		}

		_, ok := <-results
		assert.False(t, ok, "Results channel should be closed")
		assert.NoError(t, ing.Err())
	})
}

func makeCSV(rows [][]string) string {
	s := "1\n2\n3\n4\n5\n" // first 5 rows are skipped
	for _, row := range rows {
		s += strings.Join(row, ",") + "\n"
	}
	return s
}
