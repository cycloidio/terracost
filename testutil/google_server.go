package testutil

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// StartGoogleServer starts a new test server for Google API
// returning the data from "../testdata/google/api"
func StartGoogleServer(t *testing.T) *httptest.Server {
	t.Helper()

	bskus, err := os.ReadFile("../testdata/google/api/skus.json")
	require.NoError(t, err)

	bmt, err := os.ReadFile("../testdata/google/api/machine_types.json")
	require.NoError(t, err)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b []byte
		switch r.URL.String() {
		case "/v1/services/6F81-5844-456A/skus?alt=json&prettyPrint=false":
			b = bskus
		case "/projects/proj/zones/europe-west1-b/machineTypes?alt=json&prettyPrint=false":
			b = bmt
		default:
			t.Fatalf("URL %s not handled", r.URL)
		}
		w.Write(b)
	}))
}
