package price

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/product"
)

// ID represents the Price ID.
type ID uint32

// Price is a single pricing entry for a product.
type Price struct {
	ID         ID
	Unit       string
	Currency   string
	Value      decimal.Decimal
	Attributes map[string]string
}

var (
	// ErrMismatchingUnit when the unit of the 2 prices do not match when using Add
	ErrMismatchingUnit = errors.New("the unit is not the same")

	// ErrMismatchingCurrency when the currency of the 2 prices do not match when using Add
	ErrMismatchingCurrency = errors.New("the currency is not the same")
)

// GenerateHash generates the Hash field of the Price, equal to the MD5 sum of its unique values.
func (p *Price) GenerateHash() string {
	values := make([]string, 0, len(p.Attributes)+2)
	values = append(values, p.Unit, p.Currency)

	keys := make([]string, 0, len(p.Attributes))
	for k := range p.Attributes {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		values = append(values, p.Attributes[k])
	}

	data := []byte(strings.Join(values, "-"))
	return fmt.Sprintf("%x", md5.Sum(data))
}

// Add checks that the 2 prices can be added together (unit/currency) and
// then performs an addition and also joins the attributes and replacing repeated
// ones for the new ones (pr)
func (p *Price) Add(pr Price) error {
	if p.Unit != pr.Unit {
		return ErrMismatchingUnit
	} else if p.Currency != pr.Currency {
		return ErrMismatchingCurrency
	}

	p.Value = p.Value.Add(pr.Value)
	for k, v := range pr.Attributes {
		p.Attributes[k] = v
	}

	return nil
}

// WithProduct is an aggregation of a Price with a product.Product.
type WithProduct struct {
	Price
	Product *product.Product
}
