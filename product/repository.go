package product

import (
	"context"
)

//go:generate go run github.com/golang/mock/mockgen -destination=../mock/product_repository.go -mock_names=Repository=ProductRepository -package mock github.com/cycloidio/cost-estimation/product Repository

// Repository describes interactions with a storage system to deal with Product entries.
type Repository interface {
	// Filter returns Products with attributes matching the Filter.
	Filter(ctx context.Context, filter *Filter) ([]*Product, error)

	// FindByVendorAndSKU finds a single Product by its vendor and SKU.
	FindByVendorAndSKU(ctx context.Context, vendor string, sku string) (*Product, error)

	// Upsert updates a Product or creates a new one if it doesn't already exist.
	Upsert(ctx context.Context, p *Product) (ID, error)
}
