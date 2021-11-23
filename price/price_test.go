package price_test

import (
	"testing"

	"github.com/cycloidio/terracost/price"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	p := price.Price{
		Unit:     "Mb",
		Currency: "USD",
		Value:    decimal.NewFromInt(10),
		Attributes: map[string]string{
			"a": "1",
			"b": "2",
		},
	}
	t.Run("Success", func(t *testing.T) {
		err := p.Add(price.Price{
			Unit:     p.Unit,
			Currency: p.Currency,
			Value:    decimal.NewFromInt(20),
			Attributes: map[string]string{
				"b": "3",
				"c": "4",
			},
		})
		np := price.Price{
			Unit:     p.Unit,
			Currency: p.Currency,
			Value:    decimal.NewFromInt(30),
			Attributes: map[string]string{
				"a": "1",
				"b": "3",
				"c": "4",
			},
		}
		require.NoError(t, err)
		assert.Equal(t, np, p)
	})

	t.Run("ErrMismatchingUnit", func(t *testing.T) {
		err := p.Add(price.Price{
			Unit:     "potato",
			Currency: p.Currency,
			Value:    decimal.NewFromInt(20),
			Attributes: map[string]string{
				"b": "3",
				"c": "4",
			},
		})
		assert.EqualError(t, err, price.ErrMismatchingUnit.Error())
	})

	t.Run("ErrMismatchingCurrency", func(t *testing.T) {
		err := p.Add(price.Price{
			Unit:     p.Unit,
			Currency: "potato",
			Value:    decimal.NewFromInt(20),
			Attributes: map[string]string{
				"b": "3",
				"c": "4",
			},
		})
		assert.EqualError(t, err, price.ErrMismatchingCurrency.Error())
	})
}
