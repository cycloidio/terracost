package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cycloidio/sqlr"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
)

// PriceRepository implements the price.Repository.
type PriceRepository struct {
	querier sqlr.Querier
}

// NewPriceRepository returns an implementation of price.Repository.
func NewPriceRepository(querier sqlr.Querier) *PriceRepository {
	return &PriceRepository{querier: querier}
}

type dbPrice struct {
	ID         price.ID
	ProductID  product.ID
	Hash       string
	Currency   string
	Value      decimal.Decimal
	Unit       string
	Attributes string
}

func (p *dbPrice) toDomainEntity() *price.Price {
	var attributes map[string]string
	_ = json.Unmarshal([]byte(p.Attributes), &attributes)

	return &price.Price{
		ID:         p.ID,
		Currency:   p.Currency,
		Value:      p.Value,
		Unit:       p.Unit,
		Attributes: attributes,
	}
}

func newPrice(pwp *price.WithProduct) (*dbPrice, error) {
	attributes, err := json.Marshal(pwp.Attributes)
	if err != nil {
		return nil, err
	}

	return &dbPrice{
		ProductID:  pwp.Product.ID,
		Hash:       pwp.GenerateHash(),
		Currency:   pwp.Currency,
		Value:      pwp.Value,
		Unit:       pwp.Unit,
		Attributes: string(attributes),
	}, nil
}

// Filter returns all the price.Price that belong to a given product with given product.ID and that matches the price.Filter.
func (r *PriceRepository) Filter(ctx context.Context, productID product.ID, filter *price.Filter) ([]*price.Price, error) {
	where := parsePriceFilter(filter, productID)
	q := fmt.Sprintf(`
		SELECT id, hash, product_id, currency, price, unit, attributes
		FROM pricing_product_prices
		WHERE %s
	`, where.String())

	ps := make([]*price.Price, 0)
	rows, err := r.querier.QueryContext(ctx, q, where.Parameters()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p, err := scanPrice(rows)
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

// Upsert updates a price.WithProduct if it exists or inserts it otherwise.
func (r *PriceRepository) Upsert(ctx context.Context, pwp *price.WithProduct) (price.ID, error) {
	p, err := newPrice(pwp)
	if err != nil {
		return 0, err
	}

	q := `
		INSERT INTO pricing_product_prices (product_id, hash, currency, price, unit, attributes)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id = LAST_INSERT_ID(id),
			currency = VALUES(currency),
			price = VALUES(price),
			unit = VALUES(unit),
			attributes = VALUES(attributes)
	`

	res, err := r.querier.ExecContext(ctx, q, p.ProductID, p.Hash, p.Currency, p.Value, p.Unit, p.Attributes)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return price.ID(id), nil
}

// DeleteByProductWithKeep deletes all the prices of the product with given product.ID except the ones in the keep slice.
func (r *PriceRepository) DeleteByProductWithKeep(ctx context.Context, productID product.ID, keep []price.ID) error {
	marks := make([]string, 0, len(keep))
	values := make([]interface{}, 0, len(keep)+1)
	values = append(values, productID)

	for _, v := range keep {
		marks = append(marks, "?")
		values = append(values, v)
	}

	q := fmt.Sprintf(`DELETE FROM pricing_product_prices WHERE product_id = ? AND id NOT IN (%s)`, strings.Join(marks, ","))

	_, err := r.querier.ExecContext(ctx, q, values...)
	if err != nil {
		return err
	}
	return nil
}

func scanPrice(row sqlr.Scanner) (*price.Price, error) {
	var p dbPrice
	err := row.Scan(&p.ID, &p.Hash, &p.ProductID, &p.Currency, &p.Value, &p.Unit, &p.Attributes)
	if err != nil {
		return nil, err
	}
	return p.toDomainEntity(), nil
}
