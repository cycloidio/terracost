package price

import (
	"crypto/md5"
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

// WithProduct is an aggregation of a Price with a product.Product.
type WithProduct struct {
	Price
	Product *product.Product
}
