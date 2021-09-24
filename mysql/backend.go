package mysql

import (
	"github.com/cycloidio/sqlr"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
)

// Backend is the MySQL implementation of the costestimation.Backend, using repositories that connect
// to a MySQL database.
type Backend struct {
	querier     sqlr.Querier
	productRepo *ProductRepository
	priceRepo   *PriceRepository
}

// NewBackend returns a new Backend with a product.Repository and a price.Repository included.
func NewBackend(querier sqlr.Querier) *Backend {
	return &Backend{
		querier:     querier,
		productRepo: NewProductRepository(querier),
		priceRepo:   NewPriceRepository(querier),
	}
}

// Products returns the product.Repository that uses the Backend's querier.
func (b *Backend) Products() product.Repository { return b.productRepo }

// Prices returns the price.Repository that uses the Backend's querier.
func (b *Backend) Prices() price.Repository { return b.priceRepo }
