package testutil

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// StartAzureServer starts a new test server for Azure API
// returning the data from "../testdata/azure/api"
func StartAzureServer(t *testing.T) *httptest.Server {
	t.Helper()

	rp, err := os.ReadFile("../testdata/azurerm/api/retail_prices.json")
	require.NoError(t, err)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b []byte
		switch r.URL.String() {
		case "/api/retail/prices?$filter=serviceName%20eq%20%27Virtual%20Machines%27%20and%20armRegionName%20eq%20%27francecentral%27":
			b = rp
		default:
			t.Fatalf("URL %s not handled", r.URL)
		}
		w.Write(b)
	}))
}
