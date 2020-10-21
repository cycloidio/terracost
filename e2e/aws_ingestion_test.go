package e2e

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	costestimation "github.com/cycloidio/cost-estimation"
	"github.com/cycloidio/cost-estimation/aws"
	"github.com/cycloidio/cost-estimation/mock"
	"github.com/cycloidio/cost-estimation/mysql"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/util"
)

func TestAWSIngestion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	httpClient := mock.NewHTTPClient(ctrl)

	db, err := sql.Open("mysql", "root:youdeploy@tcp(localhost:3306)/youdeploy_public_test?multiStatements=true")
	require.NoError(t, err)

	f, err := os.Open("testdata/AmazonEC2_eu-west-3.csv")
	require.NoError(t, err)
	defer f.Close()

	httpClient.EXPECT().Do(gomock.Any()).Return(&http.Response{Body: f}, nil)

	backend := mysql.NewBackend(db)
	ingester := aws.NewIngester("AmazonEC2-test", "eu-west-3", aws.WithHTTPClient(httpClient))

	err = costestimation.IngestPricing(ctx, backend, ingester)
	require.NoError(t, err)

	allProds, err := backend.Product().Filter(ctx, &product.Filter{Provider: util.StringPtr("aws"), Service: util.StringPtr("AmazonEC2-test")})
	require.NoError(t, err)
	assert.Len(t, allProds, 5)

	for _, prod := range allProds {
		prices, err := backend.Price().Filter(ctx, prod.ID, nil)
		require.NoError(t, err)
		assert.Len(t, prices, 1)
	}
}
