package mysql

import (
	"github.com/cycloidio/sqlr"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
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

func (b *Backend) Product() product.Repository { return b.productRepo }
func (b *Backend) Price() price.Repository     { return b.priceRepo }
