package util

import (
	"net/http"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../mock/http_client.go -mock_names=HTTPClient=HTTPClient -package mock github.com/cycloidio/cost-estimation/util HTTPClient

// HTTPClient is an interface of a client that is able to Do HTTP requests
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// StringPtr returns a pointer to the passed string.
func StringPtr(s string) *string {
	return &s
}
