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

	"github.com/cycloidio/cost-estimation/mock"
	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
)

func TestIngester_Ingest(t *testing.T) {
	t.Run("EC2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := mock.NewHTTPClient(ctrl)
		ing := NewIngester("AmazonEC2", "eu-west-3", WithHTTPClient(client))

		content := makeCSV([][]string{
			{"SKU", "Product Family", "serviceCode", "TermType", "Location", "Unit", "Currency", "PricePerUnit", "Tenancy", "Instance Type", "Operating System", "Volume API Name"},
			{"prod1", "Compute Instance", "AmazonEC2", "OnDemand", "EU (Paris)", "Hrs", "USD", "1.234", "Shared", "m5.xlarge", "Linux", ""},
			{"prod1", "Compute Instance", "AmazonEC2", "Reserved", "EU (Paris)", "Hrs", "USD", "0.987", "Shared", "m5.xlarge", "Linux", ""},
			{"prod2", "Storage", "AmazonEC2", "OnDemand", "EU (Paris)", "GB-Mo", "USD", "0.456", "", "", "", "gp2"},
		})
		rd := strings.NewReader(content)
		res := &http.Response{Body: ioutil.NopCloser(rd)}

		client.EXPECT().Do(gomock.Any()).Return(res, nil)

		results := make(chan *price.WithProduct, 1)

		go func(){
			err := ing.Ingest(context.Background(), results)
			require.NoError(t, err)
		}()

		prod1 := &product.Product{
			Provider:   VendorName,
			SKU:        "prod1",
			Service:    "AmazonEC2",
			Family:     "Compute Instance",
			Location:   "eu-west-3",
			Attributes: map[string]string{
				"instanceType": "m5.xlarge",
				"operatingSystem": "Linux",
				"tenancy": "Shared",
			},
		}

		expected := []*price.WithProduct{
			{
				Product: prod1,
				Price: price.Price{
					Unit:       "Hrs",
					Currency:   "USD",
					Value:      decimal.RequireFromString("1.234"),
					Attributes: map[string]string{"purchaseOption": "on_demand"},
				},
			},
			{
				Product: prod1,
				Price: price.Price{
					Unit:       "Hrs",
					Currency:   "USD",
					Value:      decimal.RequireFromString("0.987"),
					Attributes: map[string]string{"purchaseOption": "reserved"},
				},
			},
			{
				Product: &product.Product{
					Provider:   VendorName,
					SKU:        "prod2",
					Service:    "AmazonEC2",
					Family:     "Storage",
					Location:   "eu-west-3",
					Attributes: map[string]string{
						"volumeApiName": "gp2",
					},
				},
				Price: price.Price{
					Unit:       "GB-Mo",
					Currency:   "USD",
					Value:      decimal.RequireFromString("0.456"),
					Attributes: map[string]string{"purchaseOption": "on_demand"},
				},
			},
		}

		for _, pp := range expected {
			assert.Equal(t, pp, <-results)
		}

		_, ok := <-results
		require.False(t, ok, "Results channel should be closed")
	})
}

func makeCSV(rows [][]string) string {
	s := "1\n2\n3\n4\n5\n" // first 5 rows are skipped
	for _, row := range rows {
		s += strings.Join(row, ",") + "\n"
	}
	return s
}
