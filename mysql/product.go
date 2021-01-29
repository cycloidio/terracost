package mysql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cycloidio/sqlr"

	"github.com/cycloidio/terracost/product"
)

// ProductRepository implements the product.Repository.
type ProductRepository struct {
	querier sqlr.Querier
}

// NewProductRepository returns an implementation of product.Repository.
func NewProductRepository(querier sqlr.Querier) *ProductRepository {
	return &ProductRepository{querier: querier}
}

type dbProduct struct {
	ID         product.ID
	SKU        string
	Provider   string
	Service    string
	Family     string
	Location   string
	Attributes string
}

func (p *dbProduct) toDomainEntity() *product.Product {
	var attributes map[string]string
	_ = json.Unmarshal([]byte(p.Attributes), &attributes)

	return &product.Product{
		ID:         p.ID,
		SKU:        p.SKU,
		Provider:   p.Provider,
		Service:    p.Service,
		Family:     p.Family,
		Location:   p.Location,
		Attributes: attributes,
	}
}

func newProduct(p *product.Product) (*dbProduct, error) {
	attributes, err := json.Marshal(p.Attributes)
	if err != nil {
		return nil, err
	}

	return &dbProduct{
		SKU:        p.SKU,
		Provider:   p.Provider,
		Service:    p.Service,
		Family:     p.Family,
		Location:   p.Location,
		Attributes: string(attributes),
	}, nil
}

// Filter returns all the product.Product that match the given product.Filter.
func (r *ProductRepository) Filter(ctx context.Context, filter *product.Filter) ([]*product.Product, error) {
	where := parseProductFilter(filter)
	q := fmt.Sprintf(`
		SELECT id, provider, sku, service, family, location, attributes
		FROM pricing_products
		WHERE %s
	`, where.String())

	ps := make([]*product.Product, 0)
	rows, err := r.querier.QueryContext(ctx, q, where.Parameters()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ps, nil
}

// FindByVendorAndSKU returns a single product.Product of the given vendor and sku.
func (r *ProductRepository) FindByVendorAndSKU(ctx context.Context, vendor string, sku string) (*product.Product, error) {
	q := `
		SELECT id, provider, sku, service, family, location, attributes
		FROM pricing_products
		WHERE provider = ? AND sku = ?
		LIMIT 1
	`
	row := r.querier.QueryRowContext(ctx, q, vendor, sku)
	return scanProduct(row)
}

// Upsert updates a product.Product if it exists or inserts a new one otherwise.
func (r *ProductRepository) Upsert(ctx context.Context, prod *product.Product) (product.ID, error) {
	p, err := newProduct(prod)
	if err != nil {
		return 0, err
	}

	q := `
		INSERT INTO pricing_products (provider, sku, service, family, location, attributes)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			attributes = VALUES(attributes)
	`

	res, err := r.querier.ExecContext(ctx, q, p.Provider, p.SKU, p.Service, p.Family, p.Location, p.Attributes)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return product.ID(id), nil
}

func scanProduct(row sqlr.Scanner) (*product.Product, error) {
	var p dbProduct
	err := row.Scan(&p.ID, &p.Provider, &p.SKU, &p.Service, &p.Family, &p.Location, &p.Attributes)
	if err != nil {
		return nil, err
	}
	return p.toDomainEntity(), nil
}
