package price

import (
	"context"

	"github.com/cycloidio/terracost/product"
)

//go:generate mockgen -destination=../mock/price_repository.go -mock_names=Repository=PriceRepository -package mock github.com/cycloidio/terracost/price Repository

// Repository describes interactions with a storage system to deal with Price entries.
type Repository interface {
	// Filter returns Prices with attributes matching the product.ID and Filter.
	Filter(ctx context.Context, productID product.ID, filter *Filter) ([]*Price, error)

	// Upsert updates a Price or creates a new one if it doesn't already exist.
	Upsert(ctx context.Context, p *WithProduct) (ID, error)

	// DeleteByProductWithKeep deletes all Prices of the specified product.ID except the ones with ID in the keep slice.
	DeleteByProductWithKeep(ctx context.Context, productID product.ID, keep []ID) error
}
